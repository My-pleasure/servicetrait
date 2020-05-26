/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/oam-controllers/pkg/oam/util"
	"github.com/crossplane/oam-kubernetes-runtime/pkg/oam"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha2 "servicetrait/api/v1alpha2"
)

const (
	oamReconcileWait = 30 * time.Second
)

// Reconcile error strings.
const (
	errLocateWorkload    = "cannot find workload"
	errLocateResources   = "cannot find resources"
	errUpdateStatus      = "cannot apply status"
	errLocateStatefulSet = "cannot find statefulset"
	errApplyService      = "cannot apply the service"
	errGCService         = "cannot clean up stale services"
)

// ServiceTraitReconciler reconciles a ServiceTrait object
type ServiceTraitReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.oam.dev,resources=servicetraits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.oam.dev,resources=servicetraits/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.oam.dev,resources=statefulsetworkloads,verbs=get;list;
// +kubebuilder:rbac:groups=core.oam.dev,resources=statefulsetworkloads/status,verbs=get;
// +kubebuilder:rbac:groups=core.oam.dev,resources=workloaddefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *ServiceTraitReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("servicetrait", req.NamespacedName)
	log.Info("Reconcile Service Trait")

	var service corev1alpha2.ServiceTrait
	if err := r.Get(ctx, req.NamespacedName, &service); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Get the service trait", "WorkloadReference", service.Spec.WorkloadReference)

	// Fetch the workload this trait is referring to
	workload, result, err := r.fetchWorkload(ctx, &service)
	if err != nil {
		return result, err
	}

	// Fetch the child resources list from the corresponding workload
	resources, err := util.FetchWorkloadDefinition(ctx, r, workload)
	if err != nil {
		r.Log.Error(err, "Cannot find the workload child resources", "workload", workload.UnstructuredContent())
		service.Status.SetConditions(cpv1alpha1.ReconcileError(fmt.Errorf(errLocateResources)))
		return ctrl.Result{RequeueAfter: oamReconcileWait}, errors.Wrap(r.Status().Update(ctx, &service),
			errUpdateStatus)
	}

	// Create a service for the child resources we know
	svc, err := r.createService(ctx, service, resources)
	if err != nil {
		return result, err
	}

	// server side apply the service, only the fields we set are touched
	applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner(service.Name)}
	if err := r.Patch(ctx, svc, client.Apply, applyOpts...); err != nil {
		service.Status.SetConditions(cpv1alpha1.ReconcileError(errors.Wrap(err, errApplyService)))
		r.Log.Error(err, "Failed to apply a service")
		return reconcile.Result{RequeueAfter: oamReconcileWait}, errors.Wrap(r.Status().Update(ctx, &service),
			errUpdateStatus)
	}
	r.Log.Info("Successfully applied a service", "UID", svc.UID)

	// garbage collect the service that we created but not needed
	if err := r.cleanupResources(ctx, &service, &svc.UID); err != nil {
		service.Status.SetConditions(cpv1alpha1.ReconcileError(errors.Wrap(err, errGCService)))
		r.Log.Error(err, "Failed to clean up resources")
		return reconcile.Result{RequeueAfter: oamReconcileWait}, errors.Wrap(r.Status().Update(ctx, &service),
			errUpdateStatus)
	}
	service.Status.Resources = nil
	// record the new service
	service.Status.Resources = append(service.Status.Resources, cpv1alpha1.TypedReference{
		APIVersion: svc.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       svc.GetObjectKind().GroupVersionKind().Kind,
		Name:       svc.GetName(),
		UID:        svc.GetUID(),
	})

	service.Status.SetConditions(cpv1alpha1.ReconcileSuccess())

	return ctrl.Result{}, errors.Wrap(r.Status().Update(ctx, &service), errUpdateStatus)
}

func (r *ServiceTraitReconciler) createService(ctx context.Context, service corev1alpha2.ServiceTrait,
	resources []*unstructured.Unstructured) (*corev1.Service, error) {
	// Change unstructured to object
	for _, res := range resources {
		if res.GetKind() == KindStatefulSet && res.GetAPIVersion() == appsv1.SchemeGroupVersion.String() {
			r.Log.Info("Get the statefulset the trait is going to create a service for it",
				"statefulset name", res.GetName(), "UID", res.GetUID())
			// convert the unstructured to statefulset and create a service
			var ss appsv1.StatefulSet
			bts, _ := json.Marshal(res)
			if err := json.Unmarshal(bts, &ss); err != nil {
				r.Log.Error(err, "Failed to convert an unstructured obj to a statefulset")
				continue
			}
			// Create a service for the workload which this trait is referring to
			svc, err := r.renderService(ctx, &service, &ss)
			if err != nil {
				r.Log.Error(err, "Failed to render a service")
				return nil, errors.Wrap(r.Status().Update(ctx, &service),
					errUpdateStatus)
			}
			return svc, nil
		}
	}
	r.Log.Info("Cannot locate any statefulset", "total resources", len(resources))
	service.Status.SetConditions(cpv1alpha1.ReconcileError(fmt.Errorf(errLocateStatefulSet)))
	return nil, errors.Wrap(r.Status().Update(ctx, &service),
		errUpdateStatus)
}

func (r *ServiceTraitReconciler) fetchWorkload(ctx context.Context,
	oamTrait oam.Trait) (*unstructured.Unstructured, ctrl.Result, error) {
	var workload unstructured.Unstructured
	workload.SetAPIVersion(oamTrait.GetWorkloadReference().APIVersion)
	workload.SetKind(oamTrait.GetWorkloadReference().Kind)
	wn := client.ObjectKey{Name: oamTrait.GetWorkloadReference().Name, Namespace: oamTrait.GetNamespace()}
	if err := r.Get(ctx, wn, &workload); err != nil {
		oamTrait.SetConditions(cpv1alpha1.ReconcileError(errors.Wrap(err, errLocateWorkload)))
		r.Log.Error(err, "Workload not find", "kind", oamTrait.GetWorkloadReference().Kind,
			"workload name", oamTrait.GetWorkloadReference().Name)
		return nil, ctrl.Result{RequeueAfter: oamReconcileWait}, errors.Wrap(r.Status().Update(ctx, oamTrait),
			errUpdateStatus)
	}
	r.Log.Info("Get the workload the trait is pointing to", "workload name", oamTrait.GetWorkloadReference().Name,
		"UID", workload.GetUID())
	return &workload, ctrl.Result{}, nil
}

func (r *ServiceTraitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha2.ServiceTrait{}).
		Owns(&appsv1.StatefulSet{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Service{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
