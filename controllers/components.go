// Copyright Contributors to the Open Cluster Management project

package controllers

import (
	mcev1 "github.com/stolostron/backplane-operator/api/v1"
	"github.com/stolostron/backplane-operator/pkg/status"
	"github.com/stolostron/backplane-operator/pkg/toggle"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/stolostron/backplane-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

type Component interface {
	GetName() string
	GetStatusReporter(*mcev1.MultiClusterEngine, utils.ClusterConfig) status.StatusReporter
}

type ClusterConfig struct {
	ConsoleEnabled bool
	ClusterVersion string
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

var normalComponents = []Component{
	DiscoveryComponent{},
}

var deploymentGVK = schema.GroupVersionKind{
	Group:   "apps",
	Version: "v1",
	Kind:    "Deployment",
}

type DiscoveryComponent struct{}

func (c DiscoveryComponent) GetName() string {
	return mcev1.Discovery
}

func (c DiscoveryComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {
	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				status.NewNotPresentStatus(types.NamespacedName{Name: "discovery-operator", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
			})
	} else {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				toggle.EnabledStatus(types.NamespacedName{Name: "discovery-operator", Namespace: mce.Spec.TargetNamespace}),
			})
	}
}

type ConsoleComponent struct{}

func (c ConsoleComponent) GetName() string {
	return mcev1.ConsoleMCE
}

func (c ConsoleComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {
	if !cc.ConsoleEnabled {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				status.ConsoleUnavailableStatus{
					NamespacedName: types.NamespacedName{Name: "console-mce-console",
						Namespace: mce.Spec.TargetNamespace},
				},
			})
	}
	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				status.NewNotPresentStatus(types.NamespacedName{Name: "console-mce-console", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
			})
	} else {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				toggle.EnabledStatus(types.NamespacedName{Name: "console-mce-console", Namespace: mce.Spec.TargetNamespace}),
			})
	}
}

type ManagedServiceAccountComponent struct{}

func (c ManagedServiceAccountComponent) GetName() string {
	return mcev1.ManagedServiceAccount
}

func (c ManagedServiceAccountComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {
	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				status.NewNotPresentStatus(types.NamespacedName{Name: "managed-serviceaccount-addon-manager", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
			})
	} else {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				toggle.EnabledStatus(types.NamespacedName{Name: "managed-serviceaccount-addon-manager", Namespace: mce.Spec.TargetNamespace}),
			})
	}
}

type HiveComponent struct{}

func (c HiveComponent) GetName() string {
	return mcev1.Hive
}

func (c HiveComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {
	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				status.NewNotPresentStatus(types.NamespacedName{Name: "hive-operator", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
			})
	} else {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				toggle.EnabledStatus(types.NamespacedName{Name: "hive-operator", Namespace: mce.Spec.TargetNamespace}),
			})
	}
}

type AssistedServiceComponent struct{}

func (c AssistedServiceComponent) GetName() string {
	return mcev1.AssistedService
}

func (c AssistedServiceComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {
	targetNamespace := mce.Spec.TargetNamespace
	if mce.Spec.Overrides != nil && mce.Spec.Overrides.InfrastructureCustomNamespace != "" {
		targetNamespace = mce.Spec.Overrides.InfrastructureCustomNamespace
	}
	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				status.NewNotPresentStatus(types.NamespacedName{Name: "infrastructure-operator", Namespace: targetNamespace}, deploymentGVK),
			})
	} else {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				toggle.EnabledStatus(types.NamespacedName{Name: "infrastructure-operator", Namespace: targetNamespace}),
			})
	}
}

type ServerFoundationComponent struct{}

func (c ServerFoundationComponent) GetName() string {
	return mcev1.ServerFoundation
}

func (c ServerFoundationComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {
	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				status.NewNotPresentStatus(types.NamespacedName{Name: "ocm-controller", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
				status.NewNotPresentStatus(types.NamespacedName{Name: "ocm-proxyserver", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
				status.NewNotPresentStatus(types.NamespacedName{Name: "ocm-webhook", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
			})
	} else {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				toggle.EnabledStatus(types.NamespacedName{Name: "ocm-controller", Namespace: mce.Spec.TargetNamespace}),
				toggle.EnabledStatus(types.NamespacedName{Name: "ocm-proxyserver", Namespace: mce.Spec.TargetNamespace}),
				toggle.EnabledStatus(types.NamespacedName{Name: "ocm-webhook", Namespace: mce.Spec.TargetNamespace}),
			})
	}
}

type ClusterLifecycleComponent struct{}

func (c ClusterLifecycleComponent) GetName() string {
	return mcev1.ClusterLifecycle
}

func (c ClusterLifecycleComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {
	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				status.NewNotPresentStatus(types.NamespacedName{Name: "cluster-curator-controller", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
				status.NewNotPresentStatus(types.NamespacedName{Name: "clusterclaims-controller", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
				status.NewNotPresentStatus(types.NamespacedName{Name: "provider-credential-controller", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
				status.NewNotPresentStatus(types.NamespacedName{Name: "clusterlifecycle-state-metrics-v2", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
				status.NewNotPresentStatus(types.NamespacedName{Name: "cluster-image-set-controller", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
			})
	} else {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				toggle.EnabledStatus(types.NamespacedName{Name: "cluster-curator-controller", Namespace: mce.Spec.TargetNamespace}),
				toggle.EnabledStatus(types.NamespacedName{Name: "clusterclaims-controller", Namespace: mce.Spec.TargetNamespace}),
				toggle.EnabledStatus(types.NamespacedName{Name: "provider-credential-controller", Namespace: mce.Spec.TargetNamespace}),
				toggle.EnabledStatus(types.NamespacedName{Name: "clusterlifecycle-state-metrics-v2", Namespace: mce.Spec.TargetNamespace}),
				toggle.EnabledStatus(types.NamespacedName{Name: "cluster-image-set-controller", Namespace: mce.Spec.TargetNamespace}),
			})
	}
}

type ClusterManagerComponent struct{}

func (c ClusterManagerComponent) GetName() string {
	return mcev1.ClusterManager
}

func (c ClusterManagerComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {
	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				status.NewNotPresentStatus(types.NamespacedName{Name: "cluster-manager", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
				status.NewNotPresentStatus(types.NamespacedName{Name: "cluster-manager"},
					schema.GroupVersionKind{
						Group:   "operator.open-cluster-management.io",
						Version: "v1",
						Kind:    "ClusterManager",
					}),
			})
	} else {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				toggle.EnabledStatus(types.NamespacedName{Name: "cluster-manager", Namespace: mce.Spec.TargetNamespace}),
				status.ClusterManagerStatus{NamespacedName: types.NamespacedName{Name: "cluster-manager"}},
			})
	}
}

type HypershiftComponent struct{}

func (c HypershiftComponent) GetName() string {
	return mcev1.HyperShift
}

func (c HypershiftComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {
	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				status.NewNotPresentStatus(types.NamespacedName{Name: "hypershift-addon-manager", Namespace: mce.Spec.TargetNamespace}, deploymentGVK),
				status.NewNotPresentStatus(types.NamespacedName{Name: "hypershift-addon"}, clusterManagementAddOnGVK),
			})
	} else {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				toggle.EnabledStatus(types.NamespacedName{Name: "hypershift-addon-manager", Namespace: mce.Spec.TargetNamespace}),
				status.NewPresentStatus(types.NamespacedName{Name: "hypershift-addon"}, clusterManagementAddOnGVK),
			})
	}
}

type HypershiftLocalHostingComponent struct{}

func (c HypershiftLocalHostingComponent) GetName() string {
	return mcev1.HypershiftLocalHosting
}

func (c HypershiftLocalHostingComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {

	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		return status.NewDisabledStatus(types.NamespacedName{Name: c.GetName()}, "Component is disabled", []*unstructured.Unstructured{})
	}

	if !mce.Enabled(mcev1.HyperShift) {
		return status.NewDisabledStatus(types.NamespacedName{Name: c.GetName()}, "Local hosting only available when hypershift is enabled", []*unstructured.Unstructured{})
	}
	if !mce.Enabled(mcev1.LocalCluster) {
		return status.NewDisabledStatus(types.NamespacedName{Name: c.GetName()}, "Local hosting only available when local-cluster is enabled", []*unstructured.Unstructured{})
	}

	return status.NewMultiStatus(
		types.NamespacedName{Name: c.GetName()},
		[]status.StatusReporter{
			status.ManagedClusterAddOnStatus{NamespacedName: types.NamespacedName{Name: "hypershift-addon", Namespace: "local-cluster"}},
		})
}

type ClusterProxyAddonComponent struct{}

func (c ClusterProxyAddonComponent) GetName() string {
	return mcev1.ClusterProxyAddon
}

func (c ClusterProxyAddonComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {
	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		return status.NewDisabledStatus(types.NamespacedName{Name: c.GetName()}, "Component is disabled", []*unstructured.Unstructured{})
	} else {
		return status.NewMultiStatus(
			types.NamespacedName{Name: c.GetName()},
			[]status.StatusReporter{
				toggle.EnabledStatus(types.NamespacedName{Name: "cluster-proxy-addon-manager", Namespace: mce.Spec.TargetNamespace}),
				toggle.EnabledStatus(types.NamespacedName{Name: "cluster-proxy-addon-user", Namespace: mce.Spec.TargetNamespace}),
			})
	}
}

type LocalClusterComponent struct{}

func (c LocalClusterComponent) GetName() string {
	return mcev1.LocalCluster
}

func (c LocalClusterComponent) GetStatusReporter(mce *mcev1.MultiClusterEngine, cc utils.ClusterConfig) status.StatusReporter {
	if mce.GetDeletionTimestamp() != nil || !mce.Enabled(c.GetName()) {
		nsn := types.NamespacedName{Name: "local-cluster", Namespace: mce.Spec.TargetNamespace}
		return status.LocalClusterStatus{
			NamespacedName: nsn,
			Enabled:        false,
		}
	} else {
		nsn := types.NamespacedName{Name: "local-cluster", Namespace: mce.Spec.TargetNamespace}
		return status.LocalClusterStatus{
			NamespacedName: nsn,
			Enabled:        true,
		}

	}
}
