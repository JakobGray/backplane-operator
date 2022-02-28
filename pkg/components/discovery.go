package components

import (
	bpv1 "github.com/stolostron/backplane-operator/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// func NewPredicateFuncs(filter func(object client.Object) bool) Funcs {
// 	return Funcs{
// 		CreateFunc: func(e event.CreateEvent) bool {
// 			return filter(e.Object)
// 		},
// 		UpdateFunc: func(e event.UpdateEvent) bool {
// 			return filter(e.ObjectNew)
// 		},
// 		DeleteFunc: func(e event.DeleteEvent) bool {
// 			return filter(e.Object)
// 		},
// 		GenericFunc: func(e event.GenericEvent) bool {
// 			return filter(e.Object)
// 		},
// 	}
// }
var discoveryChartPath = "pkg/templates/charts/toggle/discovery-operator"

func NewDiscoveryComponent(bpc *bpv1.MultiClusterEngine, images map[string]string, k8sClient client.Client, scheme *runtime.Scheme) ChartComponent {
	return ChartComponent{
		Name:      "discovery",
		chartPath: discoveryChartPath,

		k8sClient: k8sClient,
		scheme:    scheme,
		images:    images,
		config:    bpc,
		Enabled:   true, // TODO: check bpc
	}
}

// func RenderDiscovery(ctx context.Context, bpc *bpv1.MultiClusterEngine, images map[string]string) ([]*unstructured.Unstructured, []error) {
// 	log := log.FromContext(ctx)
// 	templates, errs := renderer.RenderChart(discoveryChartPath, bpc, images)
// 	if len(errs) > 0 {
// 		for _, err := range errs {
// 			log.Info(err.Error())
// 		}
// 		return nil, errs
// 	}

// 	return templates, nil
// }

// func DiscoveryEnabledStatus(ns string, resourceList []*unstructured.Unstructured) status.StatusReporter {
// 	resources := []*unstructured.Unstructured{}
// 	for _, u := range resourceList {
// 		resources = append(resources, newUnstructured(
// 			types.NamespacedName{Name: u.GetName(), Namespace: u.GetNamespace()},
// 			u.GroupVersionKind(),
// 		))
// 	}

// 	return status.ChartEnabledStatus{
// 		NamespacedName: types.NamespacedName{Name: "discovery", Namespace: ns},
// 		Resources:      resources,
// 	}
// }

// func DiscoveryDisabledStatus(ns string, resourceList []*unstructured.Unstructured) status.StatusReporter {
// 	removals := []*unstructured.Unstructured{}
// 	for _, u := range resourceList {
// 		removals = append(removals, newUnstructured(
// 			types.NamespacedName{Name: u.GetName(), Namespace: u.GetNamespace()},
// 			u.GroupVersionKind(),
// 		))
// 	}

// 	return status.ChartDisabledStatus{
// 		NamespacedName: types.NamespacedName{Name: "discovery", Namespace: ns},
// 		Resources:      removals,
// 	}
// }
