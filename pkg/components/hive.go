package components

import (
	"context"

	bpv1 "github.com/stolostron/backplane-operator/api/v1"
	"github.com/stolostron/backplane-operator/pkg/hive"
	"github.com/stolostron/backplane-operator/pkg/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var hiveChartPath = "pkg/templates/charts/always/hive-operator"

func NewHiveComponent(bpc *bpv1.MultiClusterEngine, images map[string]string, k8sClient client.Client, scheme *runtime.Scheme) Component {
	return MultiPartComponent{
		components: []Component{
			ChartComponent{
				Name:      "hive",
				chartPath: hiveChartPath,

				k8sClient: k8sClient,
				scheme:    scheme,
				images:    images,
				config:    bpc,
				Enabled:   true, // TODO: check bpc
			},
			FunctionRender{
				k8sClient: k8sClient,
				scheme:    scheme,
				images:    images,
				config:    bpc,
				enabled:   true, // TODO: check bpc

				RenderFunc: func(bpc *bpv1.MultiClusterEngine) (*unstructured.Unstructured, error) {
					return hive.HiveConfig(bpc), nil
				},
			},
		},
	}
}

type FunctionRender struct {
	k8sClient client.Client
	scheme    *runtime.Scheme
	images    map[string]string
	config    *bpv1.MultiClusterEngine

	RenderFunc func(bpc *bpv1.MultiClusterEngine) (*unstructured.Unstructured, error)

	enabled bool
}

func (fr FunctionRender) Enable() {
	fr.enabled = true
}

func (fr FunctionRender) Disable() {
	fr.enabled = false
}

func (fr FunctionRender) Render() ([]*unstructured.Unstructured, error) {
	template, err := fr.RenderFunc(fr.config)
	return []*unstructured.Unstructured{template}, err
}

func (fr FunctionRender) Status() status.StatusReporter {
	template, err := fr.RenderFunc(fr.config)
	if err != nil {
		panic("Couldn't render") // TODO: handle render failure
	}

	return status.IsPresent(
		types.NamespacedName{Name: "hiveconfig", Namespace: ""},
		[]*unstructured.Unstructured{template},
	)
}

func (fr FunctionRender) Reconcile() error {
	template, err := fr.RenderFunc(fr.config)
	if err != nil {
		return err
	}

	return applyTemplate(context.TODO(), fr.k8sClient, fr.scheme, fr.config, template)
}

var clcChartPath = "pkg/templates/charts/always/cluster-lifecycle"
var clcRbacPath = "pkg/templates/charts/always/rbac-aggregates"

func NewCLCComponent(bpc *bpv1.MultiClusterEngine, images map[string]string, k8sClient client.Client, scheme *runtime.Scheme) Component {
	return MultiPartComponent{
		components: []Component{
			ChartComponent{
				Name:      "cluster-lifecycle",
				chartPath: clcChartPath,

				k8sClient: k8sClient,
				scheme:    scheme,
				images:    images,
				config:    bpc,
				Enabled:   true, // TODO: check bpc
			},
			ChartComponent{
				Name:      "rbac-aggregates",
				chartPath: clcRbacPath,

				k8sClient: k8sClient,
				scheme:    scheme,
				images:    images,
				config:    bpc,
				Enabled:   true, // TODO: check bpc
			},
		},
	}
}

var serverFoundationPath = "pkg/templates/charts/always/server-foundation"

func NewServerFoundationComponent(bpc *bpv1.MultiClusterEngine, images map[string]string, k8sClient client.Client, scheme *runtime.Scheme) Component {
	return ChartComponent{
		Name:      "server-foundation",
		chartPath: serverFoundationPath,

		k8sClient: k8sClient,
		scheme:    scheme,
		images:    images,
		config:    bpc,
		Enabled:   true, // TODO: check bpc
	}
}

var clusterManagerPath = "pkg/templates/charts/always/cluster-manager"

func NewClusterManagerComponent(bpc *bpv1.MultiClusterEngine, images map[string]string, k8sClient client.Client, scheme *runtime.Scheme) Component {
	return ChartComponent{
		Name:      "cluster-manager",
		chartPath: clusterManagerPath,

		k8sClient: k8sClient,
		scheme:    scheme,
		images:    images,
		config:    bpc,
		Enabled:   true, // TODO: check bpc
	}
}
