// Copyright Contributors to the Open Cluster Management project
package status

import (
	"context"

	operatorv1 "github.com/openshift/api/operator/v1"
	bpv1 "github.com/stolostron/backplane-operator/api/v1"
	"github.com/stolostron/backplane-operator/pkg/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PluginStatus fulfills the StatusReporter interface for deployments
type PluginStatus struct {
	mceEnabled bool
}

func (s PluginStatus) GetName() string {
	return "ConsolePlugin"
}

func (s PluginStatus) GetNamespace() string {
	return ""
}

func (s PluginStatus) GetKind() string {
	return "Console"
}

// Converts a deployment's status to a backplane component status
func (s PluginStatus) Status(k8sClient client.Client) bpv1.ComponentCondition {
	console := &operatorv1.Console{}
	// If trying to check this resource from the CLI run - `oc get consoles.operator.openshift.io cluster`.
	// The default `console` is not the correct resource
	err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: "cluster"}, console)

	if err != nil && !apierrors.IsNotFound(err) {
		return unknownStatus(s.GetName(), s.GetKind())
	} else if apierrors.IsNotFound(err) {
		return unknownStatus(s.GetName(), s.GetKind())
	}

	return mapPlugin(console, s.mceEnabled)
}

func mapPlugin(console *operatorv1.Console, enabled bool) bpv1.ComponentCondition {
	if console.Spec.Plugins == nil || !utils.Contains(console.Spec.Plugins, "mce") {
		return bpv1.ComponentCondition{
			Name:      "ConsolePlugin",
			Kind:      "Console",
			Type:      "Configured",
			Status:    metav1.ConditionFalse,
			Reason:    WaitingForResourceReason,
			Message:   "Console plugin has not been configured yet",
			Available: false,
		}
	}

	return bpv1.ComponentCondition{
		Name:      "ConsolePlugin",
		Kind:      "Console",
		Type:      "Configured",
		Status:    metav1.ConditionTrue,
		Reason:    DeploySuccessReason,
		Available: true,
	}
}
