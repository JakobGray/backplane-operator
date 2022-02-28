// Copyright Contributors to the Open Cluster Management project
package status

import (
	"context"

	"fmt"

	bpv1 "github.com/stolostron/backplane-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ChartDisabledStatus fulfills the StatusReporter interface for a toggleable component. It ensures all resources are removed
type ChartDisabledStatus struct {
	types.NamespacedName
	Resources []*unstructured.Unstructured
}

func (ts ChartDisabledStatus) GetName() string {
	return ts.Name
}

func (ts ChartDisabledStatus) GetNamespace() string {
	return ts.Namespace
}

func (ts ChartDisabledStatus) GetKind() string {
	return "Component"
}

// Converts this component's status to a backplane component status
func (ts ChartDisabledStatus) Status(k8sClient client.Client) bpv1.ComponentCondition {
	present := []*unstructured.Unstructured{}
	presentString := ""
	for _, u := range ts.Resources {
		err := k8sClient.Get(context.TODO(), types.NamespacedName{
			Name:      u.GetName(),
			Namespace: u.GetNamespace(),
		}, u)

		if errors.IsNotFound(err) {
			continue
		}

		if err != nil {
			return bpv1.ComponentCondition{
				Name:               ts.GetName(),
				Kind:               ts.GetKind(),
				Type:               "Unknown",
				Status:             metav1.ConditionUnknown,
				LastUpdateTime:     metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Reason:             "Error checking status",
				Message:            "Error getting resource",
				Available:          false,
			}
		}

		present = append(present, u)
		resourceName := u.GetName()
		if u.GetNamespace() != "" {
			resourceName = fmt.Sprintf("%s/%s", u.GetNamespace(), resourceName)
		}
		presentString = fmt.Sprintf("%s <%s %s>", presentString, u.GetKind(), resourceName)
	}

	if len(present) == 0 {
		// The good ending
		return bpv1.ComponentCondition{
			Name:               ts.GetName(),
			Kind:               ts.GetKind(),
			Type:               "NotPresent",
			Status:             metav1.ConditionTrue,
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "ComponentDisabled",
			Message:            "No resources present",
			Available:          true,
		}
	} else {
		conditionMessage := fmt.Sprintf("The following resources remain:%s", presentString)
		return bpv1.ComponentCondition{
			Name:               ts.GetName(),
			Kind:               ts.GetKind(),
			Type:               "Uninstalled",
			Status:             metav1.ConditionFalse,
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "ResourcesPresent",
			Message:            conditionMessage,
			Available:          false,
		}
	}
}

// ChartEnabledStatus fulfills the StatusReporter interface for a toggleable component. It ensures all resources are removed
type ChartEnabledStatus struct {
	types.NamespacedName
	Resources []*unstructured.Unstructured
}

func (ts ChartEnabledStatus) GetName() string {
	return ts.Name
}

func (ts ChartEnabledStatus) GetNamespace() string {
	return ts.Namespace
}

func (ts ChartEnabledStatus) GetKind() string {
	return "Component"
}

// Converts this component's status to a backplane component status
func (ts ChartEnabledStatus) Status(k8sClient client.Client) bpv1.ComponentCondition {
	discrepancies := []*unstructured.Unstructured{}
	presentString := ""
	for _, u := range ts.Resources {
		if u.GetKind() == "Deployment" {
			// deployCheck
			if err := deploymentCheck(k8sClient, u); err != nil {
				discrepancies = append(discrepancies, u)
			}
		} else {
			// existsCheck
			if err := existsCheck(k8sClient, u); err != nil {
				discrepancies = append(discrepancies, u)
			}
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

func existsCheck(k8sClient client.Client, u *unstructured.Unstructured) error {
	return k8sClient.Get(context.TODO(), types.NamespacedName{
		Name:      u.GetName(),
		Namespace: u.GetNamespace(),
	}, u)
}

func deploymentCheck(k8sClient client.Client, u *unstructured.Unstructured) error {
	return existsCheck(k8sClient, u)
}
