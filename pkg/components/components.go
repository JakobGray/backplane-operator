package components

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	bpv1 "github.com/stolostron/backplane-operator/api/v1"
	renderer "github.com/stolostron/backplane-operator/pkg/rendering"
	"github.com/stolostron/backplane-operator/pkg/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	AllComponentNames = []string{
		"hive",
		"clusterLifecycle",
		"clusterManager",
		"discovery",
		"serverFoundation",
		"managedServiceAccountPreview",
	}
)

type Component interface {
	Enable()
	Disable()

	Render() ([]*unstructured.Unstructured, error)
	Reconcile() error
	// Unapply() error
	Status() status.StatusReporter
	// NoStatus() status.StatusReporter
	// DisabledCondition() bpv1.ComponentCondition
}

// Chart implements Component
type ChartComponent struct {
	Name      string
	chartPath string

	k8sClient client.Client
	scheme    *runtime.Scheme
	images    map[string]string
	config    *bpv1.MultiClusterEngine
	Enabled   bool

	// RenderChartFunc func() ([]*unstructured.Unstructured, error)
	// EnabledStatus   func() bpv1.ComponentCondition
	// DisabledStatus  func() bpv1.ComponentCondition

	// InstallFunc func() error
	// RequirementsMet   func() bool
	// Render            func() []*unstructured.Unstructured
	// EnabledCondition  func() bpv1.ComponentCondition
	// DisabledCondition func() bpv1.ComponentCondition
}

func (c ChartComponent) IsEnabled() bool {
	return c.Enabled
}

func (c ChartComponent) Enable() {
	c.Enabled = true
}

func (c ChartComponent) Disable() {
	c.Enabled = false
}

// Render renders all the resources belonging to the component
func (c ChartComponent) Render() ([]*unstructured.Unstructured, error) {
	templates, errs := renderer.RenderChart(c.chartPath, c.config, c.images)
	if len(errs) > 0 {
		return nil, errs[0] // TODO: merge or nest errors
	}

	return templates, nil
}

// Status returns a StatusReporter that can identify whether all components have been successfully
// installed on the cluster. Status depends on whether the component is enabled.
func (c ChartComponent) Status() status.StatusReporter {
	resourceList, err := c.Render()
	if err != nil {
		panic("Couldn't render") // TODO: handle render failure
	}

	//
	resources := []*unstructured.Unstructured{}
	for _, u := range resourceList {
		resources = append(resources, newUnstructured(
			types.NamespacedName{Name: u.GetName(), Namespace: u.GetNamespace()},
			u.GroupVersionKind(),
		))
	}

	if c.Enabled {
		return status.ChartEnabledStatus{
			NamespacedName: types.NamespacedName{Name: c.Name, Namespace: ""},
			Resources:      resources,
		}
	} else {
		return status.ChartDisabledStatus{
			NamespacedName: types.NamespacedName{Name: c.Name, Namespace: ""},
			Resources:      resources,
		}
	}
}

// Reconcile instlals or uninstalls components depending on if they are enabled
func (c ChartComponent) Reconcile() error {
	if c.Enabled {
		return ensureChart(c)
	} else {
		return ensureNoChart(c)
	}
}

func ensureChart(c ChartComponent) error {
	templates, errs := renderer.RenderChart(c.chartPath, c.config, c.images)
	if len(errs) > 0 {
		// for _, err := range errs {
		// 	log.Info(err.Error())
		// }
		return errs[0] // TODO merge errs
	}

	// Applies all templates
	for _, template := range templates {
		err := applyTemplate(context.TODO(), c.k8sClient, c.scheme, c.config, template)
		if err != nil {
			return err
		}
	}
	return nil
}

func ensureNoChart(c ChartComponent) error {
	templates, errs := renderer.RenderChart(c.chartPath, c.config, c.images)
	if len(errs) > 0 {
		// for _, err := range errs {
		// 	log.Info(err.Error())
		// }
		return errs[0] // TODO merge errs
	}

	// Deletes all templates
	for _, template := range templates {
		err := deleteTemplate(context.TODO(), c.k8sClient, template)
		if err != nil {
			return err
		}
	}
	return nil
}

func ConfigureComponents(bpc *bpv1.MultiClusterEngine, images map[string]string, k8sClient client.Client, scheme *runtime.Scheme) []Component {
	componentMap := map[string]ChartComponent{
		"discovery": NewDiscoveryComponent(bpc, images, k8sClient, scheme),
	}
	configured := []Component{}

	var sb strings.Builder
	for name, comp := range componentMap {
		match := false
		for _, c := range bpc.Spec.Components {
			if c.Name == name {
				match = true
				if c.Enabled {
					sb.WriteString(fmt.Sprintf("%s: %s\n", name, "Enabled"))
					comp.Enabled = true
				} else {
					sb.WriteString(fmt.Sprintf("%s: %s\n", name, "Disabled"))
					comp.Enabled = false
				}
				break
			}
		}
		if !match {
			sb.WriteString(fmt.Sprintf("%s: %s\n", name, "Unknown"))
		}
		configured = append(configured, comp)

	}
	fmt.Println(sb.String())
	return configured
}

func applyTemplate(ctx context.Context, k8sClient client.Client, scheme *runtime.Scheme, backplaneConfig *bpv1.MultiClusterEngine, template *unstructured.Unstructured) error {
	// Set owner reference.
	err := ctrl.SetControllerReference(backplaneConfig, template, scheme)
	if err != nil {
		return errors.Wrapf(err, "Error setting controller reference on resource %s", template.GetName())
	}

	// Apply the object data.
	force := true
	err = k8sClient.Patch(ctx, template, client.Apply, &client.PatchOptions{Force: &force, FieldManager: "backplane-operator"})
	if err != nil {
		return errors.Wrapf(err, "error applying object Name: %s Kind: %s", template.GetName(), template.GetKind())
	}

	return nil
}

func deleteTemplate(ctx context.Context, k8sClient client.Client, template *unstructured.Unstructured) error {
	err := k8sClient.Delete(ctx, template)
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func newUnstructured(nn types.NamespacedName, gvk schema.GroupVersionKind) *unstructured.Unstructured {
	u := unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	u.SetName(nn.Name)
	u.SetNamespace((nn.Namespace))
	return &u
}

type MultiPartComponent struct {
	components []Component

	Enabled bool
}

func (mpc MultiPartComponent) Enable() {
	mpc.Enabled = true
	for i := range mpc.components {
		mpc.components[i].Enable()
	}
}

func (mpc MultiPartComponent) Disable() {
	mpc.Enabled = false
	for i := range mpc.components {
		mpc.components[i].Disable()
	}
}

func (mpc MultiPartComponent) Render() ([]*unstructured.Unstructured, error) {
	templates := []*unstructured.Unstructured{}
	for _, c := range mpc.components {
		t, err := c.Render()
		if err != nil {
			return nil, err
		}
		templates = append(templates, t...)
	}
	return templates, nil
}

func (mpc MultiPartComponent) Status() status.StatusReporter {
	return mpc.components[0].Status()
}

func (mpc MultiPartComponent) Reconcile() error {
	for _, c := range mpc.components {
		err := c.Reconcile()
		if err != nil {
			return err
		}
	}
	return nil
}

type ConditionalComponent struct {
	k8sClient   client.Client
	requirement func(client.Client) bool
	component   Component
}

func (cc ConditionalComponent) Enable() {
	cc.component.Enable()
}

func (cc ConditionalComponent) Disable() {
	cc.component.Disable()
}

func (cc ConditionalComponent) Render() ([]*unstructured.Unstructured, error) {
	return cc.component.Render()
}

func (cc ConditionalComponent) Status() status.StatusReporter {
	return cc.component.Status()
}

func (cc ConditionalComponent) Reconcile() error {
	if !cc.requirement(cc.k8sClient) {
		return errors.Errorf("Requirements not met")
	}
	return cc.component.Reconcile()
}
