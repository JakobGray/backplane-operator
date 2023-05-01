package components

import (
	mcev1 "github.com/stolostron/backplane-operator/api/v1"
	"github.com/stolostron/backplane-operator/pkg/status"
	templates "github.com/stolostron/backplane-operator/pkg/template"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var Discovery = NewGenericRenderer(templates.DiscoveryFS, mcev1.Discovery)

func init() {
	err := Discovery.Init()
	if err != nil {
		panic(err)
	}
}

type RendererReporter interface {
	GetName() string
	GetStatusReporter(config *Values) status.StatusReporter
	Render(config *Values) ([]*unstructured.Unstructured, error)
}

type ChartComponent struct {
	Renderer
}

func (cc *ChartComponent) GetName() string {
	return cc.componentName
}

func (cc *ChartComponent) Render(config *Values) ([]*unstructured.Unstructured, error) {
	if !cc.ready {
		if err := cc.Init(); err != nil {
			return nil, err
		}
	}
	return cc.Render(config)
}

func (cc *ChartComponent) GetStatusReporter(config *Values) ([]*unstructured.Unstructured, error) {
	return cc.Render(config)
}
