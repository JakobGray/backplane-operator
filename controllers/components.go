package controllers

import (
	"context"

	mcev1 "github.com/stolostron/backplane-operator/api/v1"
	"github.com/stolostron/backplane-operator/pkg/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateStatus(ctx context.Context, mce *mcev1.MultiClusterEngine) []mcev1.ComponentCondition {
	return []mcev1.ComponentCondition{}
}

type Component struct {
	statusreporter status.StatusReporter
}

func (c *Component) GetStatus(c bpv1.MultiClusterEngineCondition, k8sClient client.Client) {
	c.statusreporter.Status(k8sClient)
	sm.Conditions = setCondition(sm.Conditions, c)
}


type DiscoveryComponent {
	
}