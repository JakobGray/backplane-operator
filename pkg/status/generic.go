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

func IsPresent(nn types.NamespacedName, u []*unstructured.Unstructured) StatusReporter {
	return UnstructuredStatus{
		NamespacedName: nn,
		Resources:      u,
	}
}

// UnstructuredStatus fulfills the StatusReporter interface
type UnstructuredStatus struct {
	types.NamespacedName
	Resources []*unstructured.Unstructured
}

func (us UnstructuredStatus) GetName() string {
	return us.Name
}

func (us UnstructuredStatus) GetNamespace() string {
	return us.Namespace
}

func (us UnstructuredStatus) GetKind() string {
	return "Unstructured"
}

// Converts this component's status to a backplane component status
func (ts UnstructuredStatus) Status(k8sClient client.Client) bpv1.ComponentCondition {
	discrepancies := []*unstructured.Unstructured{}
	presentString := ""
	for _, u := range ts.Resources {
		// existsCheck
		if err := existsCheck(k8sClient, u); err != nil {
			discrepancies = append(discrepancies, u)
		}
	}

	if len(discrepancies) == 0 {
		// The good ending
		return bpv1.ComponentCondition{
			Name:               ts.GetName(),
			Kind:               ts.GetKind(),
			Type:               "Available",
			Status:             metav1.ConditionTrue,
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "AllComponentsPresent",
			Message:            "Component installed",
			Available:          true,
		}
	} else {
		for i := range discrepancies {
			resourceName := discrepancies[i].GetName()
			if discrepancies[i].GetNamespace() != "" {
				resourceName = fmt.Sprintf("%s/%s", discrepancies[i].GetNamespace(), resourceName)
			}
			presentString = fmt.Sprintf("%s <%s %s>", presentString, discrepancies[i].GetKind(), resourceName)
		}
		message := fmt.Sprintf("The following resources are missing:%s", presentString)
		return bpv1.ComponentCondition{
			Name:               ts.GetName(),
			Kind:               ts.GetKind(),
			Type:               "Uninstalled",
			Status:             metav1.ConditionFalse,
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "ResourcesPresent",
			Message:            message,
			Available:          false,
		}
	}
}
