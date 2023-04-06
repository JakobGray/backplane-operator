// Copyright Contributors to the Open Cluster Management project
package status

import (
	"fmt"

	bpv1 "github.com/stolostron/backplane-operator/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// func NewStatusFromChart(namespacedName types.NamespacedName, []*unstructured.Unstructured) StatusReporter {
// 	return MultiStatus{
// 		NamespacedName: namespacedName,
// 		reporters:      reporters,
// 	}
// }

// ReportersFromUnstructured will generate a list of reporters for each resource provided based on its
// type.
func ReportersFromUnstructured(uList []*unstructured.Unstructured) ([]StatusReporter, error) {
	reporters := []StatusReporter{}
	for _, u := range uList {
		nn := types.NamespacedName{Name: u.GetName(), Namespace: u.GetNamespace()}
		switch kind := u.GroupVersionKind().Kind; kind {
		case "Deployment":
			reporters = append(reporters, DeploymentStatus{nn})
		case "ClusterManager":
			reporters = append(reporters, ClusterManagerStatus{nn})
		default:
			reporters = append(reporters, NewPresentStatus(nn, u.GroupVersionKind()))
		}
	}
	return reporters, nil
}

func NewMultiStatus(namespacedName types.NamespacedName, reporters []StatusReporter) StatusReporter {
	return MultiStatus{
		NamespacedName: namespacedName,
		reporters:      reporters,
	}
}

// MultiStatus
type MultiStatus struct {
	types.NamespacedName

	reporters []StatusReporter
}

func (s MultiStatus) GetName() string {
	return s.Name
}

func (s MultiStatus) GetNamespace() string {
	return s.Namespace
}

func (s MultiStatus) GetKind() string {
	return "Component"
}

// Returns a condition from the combined status of all its reporters. If all reporters
// are available will return a general available condition. If one or more reporters is
// not available this will return a concatenated message of each unhealthy reporter.
func (s MultiStatus) Status(k8sClient client.Client) bpv1.ComponentCondition {
	message := ""
	for _, reporter := range s.reporters {
		condition := reporter.Status(k8sClient)
		if condition.Available {
			continue
		}
		unreadyMessage := fmt.Sprintf("%s/%s is not available: %s\n", condition.Kind, condition.Name, condition.Message)
		message += unreadyMessage
	}

	if message == "" {
		return bpv1.ComponentCondition{
			Name:      s.GetName(),
			Kind:      s.GetKind(),
			Type:      "Available",
			Status:    metav1.ConditionTrue,
			Reason:    ComponentsAvailableReason,
			Available: true,
		}
	}

	return bpv1.ComponentCondition{
		Name:               s.GetName(),
		Kind:               s.GetKind(),
		Type:               "Available",
		Status:             metav1.ConditionFalse,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             RequirementsNotMetReason,
		Message:            message,
		Available:          false,
	}
}
