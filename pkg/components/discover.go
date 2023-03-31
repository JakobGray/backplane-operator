package components

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	mcev1 "github.com/stolostron/backplane-operator/api/v1"
	renderer "github.com/stolostron/backplane-operator/pkg/rendering"
	"github.com/stolostron/backplane-operator/pkg/status"
	"github.com/stolostron/backplane-operator/pkg/toggle"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//	var Operator = models.MonitoredOperator{
//		Name:             "cnv",
//		Namespace:        DownstreamNamespace,
//		OperatorType:     models.OperatorTypeOlm,
//		SubscriptionName: "hco-operatorhub",
//		TimeoutSeconds:   60 * 60,
//	}
var requeuePeriod = 15 * time.Second

// NewCNVOperator creates new instance of a Container Native Virtualization installation plugin
func NewDiscoveryComponent() *component {
	return &component{}
}

// GetName reports the name of an operator this Operator manages
func (c *component) GetName() string {
	return "discovery"
}

func (c *component) GetTemplates(k8sClient client.Client, cluster string, mce *mcev1.MultiClusterEngine) ([]*unstructured.Unstructured, []error) {
	images := map[string]string{}
	return renderer.RenderChart(toggle.DiscoveryChartDir, mce, images)
}

// func (c *component) GetTemplateKeys(k8sClient client.Client, cluster string, mce *mcev1.MultiClusterEngine) ([]client.Object, error) {
// 	images := map[string]string{}
// 	return renderer.RenderChart(toggle.DiscoveryChartDir, mce, images)
// }

func (c *component) Reconcile(k8sClient client.Client, cluster string, mce *mcev1.MultiClusterEngine) (ctrl.Result, error) {
	if mce.Enabled(mcev1.Discovery) {
		return c.apply(k8sClient, cluster, mce)
	} else {
		return c.remove(k8sClient, cluster, mce)
	}
}

func (c *component) apply(k8sClient client.Client, cluster string, mce *mcev1.MultiClusterEngine) (ctrl.Result, error) {
	templates, errs := c.GetTemplates(k8sClient, cluster, mce)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Info(err.Error())
		}
		return ctrl.Result{RequeueAfter: requeuePeriod}, errs[0]
	}

	for _, template := range templates {
		result, err := r.applyTemplate(context.TODO(), backplaneConfig, template)
		if err != nil {
			return result, err
		}
	}

	return ctrl.Result{}, nil
}

func (c *component) remove(k8sClient client.Client, cluster string, mce *mcev1.MultiClusterEngine) (ctrl.Result, error) {
	templates, errs := c.GetTemplates(k8sClient, cluster, mce)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Info(err.Error())
		}
		return ctrl.Result{RequeueAfter: requeuePeriod}, nil
	}
	// Deletes all templates
	for _, template := range templates {
		result, err := r.deleteTemplate(ctx, backplaneConfig, template)
		if err != nil {
			log.Error(err, fmt.Sprintf("Failed to delete template: %s", template.GetName()))
			return result, err
		}
	}
	return ctrl.Result{}, nil
}

func (c *component) GetStatusReporter(mce *mcev1.MultiClusterEngine) status.StatusReporter {
	return toDo
}

var toDo status.StatusReporter = status.StaticStatus{
	NamespacedName: types.NamespacedName{Name: "ToDo", Namespace: ""},
	Kind:           "Component",
	Condition: mcev1.ComponentCondition{
		Type:      "Unimplemented",
		Name:      "ToDo",
		Status:    metav1.ConditionTrue,
		Reason:    status.WaitingForResourceReason,
		Kind:      "Component",
		Available: true,
		Message:   "Component status unimplemented",
	},
}

// func (r *MultiClusterEngineReconciler) ensureDiscovery(ctx context.Context, backplaneConfig *backplanev1.MultiClusterEngine) (ctrl.Result, error) {
// 	namespacedName := types.NamespacedName{Name: "discovery-operator", Namespace: backplaneConfig.Spec.TargetNamespace}
// 	r.StatusManager.RemoveComponent(toggle.DisabledStatus(namespacedName, []*unstructured.Unstructured{}))
// 	r.StatusManager.AddComponent(toggle.EnabledStatus(namespacedName))

// 	log := log.FromContext(ctx)

// 	templates, errs := renderer.RenderChart(toggle.DiscoveryChartDir, backplaneConfig, r.Images)
// 	if len(errs) > 0 {
// 		for _, err := range errs {
// 			log.Info(err.Error())
// 		}
// 		return ctrl.Result{RequeueAfter: requeuePeriod}, nil
// 	}

// 	// Applies all templates
// 	for _, template := range templates {
// 		result, err := r.applyTemplate(ctx, backplaneConfig, template)
// 		if err != nil {
// 			return result, err
// 		}
// 	}

// 	return ctrl.Result{}, nil
// }

// func (r *MultiClusterEngineReconciler) ensureNoDiscovery(ctx context.Context, backplaneConfig *backplanev1.MultiClusterEngine) (ctrl.Result, error) {
// 	log := log.FromContext(ctx)
// 	namespacedName := types.NamespacedName{Name: "discovery-operator", Namespace: backplaneConfig.Spec.TargetNamespace}

// 	// Renders all templates from charts
// 	templates, errs := renderer.RenderChart(toggle.DiscoveryChartDir, backplaneConfig, r.Images)
// 	if len(errs) > 0 {
// 		for _, err := range errs {
// 			log.Info(err.Error())
// 		}
// 		return ctrl.Result{RequeueAfter: requeuePeriod}, nil
// 	}

// 	r.StatusManager.RemoveComponent(toggle.EnabledStatus(namespacedName))
// 	r.StatusManager.AddComponent(toggle.DisabledStatus(namespacedName, []*unstructured.Unstructured{}))

// 	// Deletes all templates
// 	for _, template := range templates {
// 		result, err := r.deleteTemplate(ctx, backplaneConfig, template)
// 		if err != nil {
// 			log.Error(err, fmt.Sprintf("Failed to delete template: %s", template.GetName()))
// 			return result, err
// 		}
// 	}
// 	return ctrl.Result{}, nil
// }
