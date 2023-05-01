package components

import (
	mcev1 "github.com/stolostron/backplane-operator/api/v1"
	"github.com/stolostron/backplane-operator/pkg/cluster"
	"github.com/stolostron/backplane-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

type Values struct {
	Org string `json:"org"`

	// Images is a full list of all known image addresses mapped by a snake case key name
	Images map[string]string `json:"images"`

	// Derived from MCE
	PullPolicy   string              `json:"pullPolicy"`
	PullSecret   string              `json:"pullSecret"`
	Namespace    string              `json:"namespace"`
	ConfigSecret string              `json:"configSecret"`
	NodeSelector map[string]string   `json:"nodeSelector"`
	ReplicaCount int                 `json:"replicaCount"`
	Tolerations  []corev1.Toleration `json:"tolerations"`

	// Derived from runtime environment
	ProxyConfigs         map[string]string `json:"proxyConfigs"`
	OCPVersion           string            `json:"ocpVersion"`
	ClusterIngressDomain string            `json:"clusterIngressDomain"`

	// Derived from MCE and environment
	HubType string `json:"hubType"`
}

// GetValues returns a manifest of values for rendering templates
func GetValues(mce *mcev1.MultiClusterEngine, cc cluster.Config, images map[string]string) Values {
	return Values{
		Org:    "open-cluster-management",
		Images: images,

		PullPolicy: string(utils.GetImagePullPolicy(mce)),
		PullSecret: mce.Spec.ImagePullSecret,
		Namespace:  mce.Spec.TargetNamespace,
		// ConfigSecret: "",
		NodeSelector: mce.Spec.NodeSelector,
		ReplicaCount: utils.DefaultReplicaCount(mce),
		Tolerations:  utils.GetDefaultTolerations(mce),

		ProxyConfigs:         cc.ProxyConfig,
		OCPVersion:           cc.ClusterVersion,
		ClusterIngressDomain: cc.IngressDomain,

		HubType: utils.GetHubType(mce), //TODO: use clusterconfig + mce
	}
}

var sec int64 = 100
var dummyVal = Values{
	Namespace:  "multicluster-engine",
	Org:        "open-cluster-management",
	PullPolicy: "IfNotPresent",
	Images: map[string]string{
		"discovery_operator": "quay.io/jakobgray/discovery-operator:latest",
	},
	ProxyConfigs: map[string]string{
		"HTTP_PROXY":  "test1",
		"HTTPS_PROXY": "test2",
		"NO_PROXY":    "test3",
	},
	PullSecret: "testpullsecret",
	NodeSelector: map[string]string{
		"select": "test",
	},
	Tolerations: []corev1.Toleration{
		{
			Key:               "dedicated",
			Operator:          "Exists",
			Effect:            "NoSchedule",
			Value:             "test",
			TolerationSeconds: &sec,
		},
	},
}
