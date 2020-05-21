package service

import (
	"context"
	"github.com/crossplane/oam-kubernetes-runtime/pkg/oam"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
)

// Reconcile error strings.
const (
	errNotServiceTrait = "object is not a service trait"
)

const LabelKey = "workload.oam.crossplane.io"

var (
	serviceKind =reflect.TypeOf(corev1.Service{}).Name()
	serviceAPIVersion = corev1.SchemeGroupVersion.String()
)

// ServiceInjector adds a Service object for the first Port on the first
// Container for the first Deployment observed in a workload translation.
func ServiceInjector(ctx context.Context, t oam.Trait, objs []oam.Object) ([]oam.Object, error) {
	if objs == nil {
		return nil, nil
	}

	for _, o := range objs {
		set, ok := o.(*appsv1.StatefulSet)
		if !ok {
			continue
		}

		// We don't add a Service if there are no containers for the StatefulSet.
		// This should never happen in practice.
		if len(set.Spec.Template.Spec.Containers) < 1 {
			continue
		}
		svc := &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       serviceKind,
				APIVersion: serviceAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: set.GetName(),
				Namespace: set.GetNamespace(),
				Labels: map[string]string{
					LabelKey: string(t.GetUID()),
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: set.Spec.Selector.MatchLabels,
				Ports: []corev1.ServicePort{},
				Type: corev1.ServiceTypeClusterIP,
			},
		}

		// We only add a single Service for the StatefulSet, even if multiple
		// ports or no ports are defined on the first container. This is to
		// exclude the need for implementing garbage collection in the
		// short-term in the case that ports are modified after creation.
		if len(set.Spec.Template.Spec.Containers[0].Ports) > 0 {
			svc.Spec.Ports = []corev1.ServicePort{
				{
					Name:       set.GetName(),
					Port:       set.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort,
					TargetPort: intstr.FromInt(int(set.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)),
				},
			}
		}
		objs = append(objs, svc)
		break
	}

	return objs, nil
}