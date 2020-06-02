package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/crossplane/oam-controllers/pkg/oam/util"
	"github.com/crossplane/oam-kubernetes-runtime/apis/core/v1alpha2"
	"github.com/crossplane/oam-kubernetes-runtime/pkg/oam"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	workloadAPIVersion = v1alpha2.SchemeGroupVersion.String()
	appsAPIVersion     = appsv1.SchemeGroupVersion.String()
)

const (
	KindStatefulSet = "StatefulSet"
	KindDeployment  = "Deployment"
)

// Determine whether the workload is K8S native resources or oam WorkloadDefinition
func DetermineWorkloadType(ctx context.Context, log logr.Logger, r client.Reader,
	workload *unstructured.Unstructured) ([]*unstructured.Unstructured, error) {
	apiVersion := workload.GetAPIVersion()
	switch apiVersion {
	case workloadAPIVersion:
		return util.FetchWorkloadDefinition(ctx, log, r, workload)
	case appsAPIVersion:
		log.Info("workload is K8S native resources", "APIVersion", apiVersion)
		return []*unstructured.Unstructured{workload}, nil
	case "":
		return nil, errors.Errorf(fmt.Sprint("failed to get the workload apiVersion"))
	default:
		return nil, errors.Errorf(fmt.Sprint("This trait doesn't support the type", apiVersion))
	}
}

// Determine whether the resource is StatefulSet or Deployment
func DetermineResourceKind(log logr.Logger, resource *unstructured.Unstructured) (oam.Object, error) {
	kind := resource.GetKind()
	switch kind {
	case KindStatefulSet:
		var ss appsv1.StatefulSet
		bts, _ := json.Marshal(resource)
		if err := json.Unmarshal(bts, &ss); err != nil {
			log.Error(err, "Failed to convert an unstructured obj to a statefulset")
		}
		return oam.Object(&ss), nil
	case KindDeployment:
		var deploy appsv1.Deployment
		bts, _ := json.Marshal(resource)
		if err := json.Unmarshal(bts, &deploy); err != nil {
			log.Error(err, "Failed to convert an unstructured obj to a deployment")
		}
		return oam.Object(&deploy), nil
	default:
		return nil, errors.Errorf("this resources is not statefulset or deployment")
	}
}
