package components

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"text/template"

	mcev1 "github.com/stolostron/backplane-operator/api/v1"
	"github.com/stolostron/backplane-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

var componentMapping = map[string]string{
	mcev1.ManagedServiceAccount:  "templates/managed-serviceaccount",
	mcev1.ConsoleMCE:             "templates/console-mce",
	mcev1.Discovery:              "templates/discovery",
	mcev1.Hive:                   "",
	mcev1.AssistedService:        "templates/assisted-service",
	mcev1.ClusterLifecycle:       "templates/cluster-lifecycle",
	mcev1.ClusterManager:         "templates/cluster-manager",
	mcev1.ServerFoundation:       "templates/server-foundation",
	mcev1.HyperShift:             "templates/hypershift",
	mcev1.ClusterProxyAddon:      "templates/cluster-proxy-addon",
	mcev1.HypershiftLocalHosting: "",
	mcev1.LocalCluster:           "",
}

type Values struct {
	// Images is a full list of all known image addresses mapped by a snake case key name
	Images map[string]string `json:"images"`

	// Derived from MCE

	// Pull policy is the image pull policy for deployment images. It derives from the
	// multiclusterengine spec.
	PullPolicy string `json:"pullPolicy"`
	// Pull secret is an optional imagePullSecret for deployments.
	PullSecret string `json:"pullSecret"`
	// Namespace to deploy the resource into if namespace-scoped
	Namespace    string `json:"namespace"`
	ConfigSecret string `json:"configSecret"`

	// NodeSelector for mapping pods
	NodeSelector map[string]string `json:"nodeSelector"`
	// Proxy environment variables for deployments
	ProxyConfigs map[string]string   `json:"proxyConfigs"`
	ReplicaCount int                 `json:"replicaCount"`
	Tolerations  []corev1.Toleration `json:"tolerations"`

	// Derived from runtime environment

	OCPVersion           string `json:"ocpVersion"`
	ClusterIngressDomain string `json:"clusterIngressDomain"`
	HubType              string `json:"hubType"`

	Org string `json:"org" structs:"org"`
}

// GetValues returns a manifest of values for rendering templates
func GetValues(mce *mcev1.MultiClusterEngine, images map[string]string) Values {
	values := Values{
		Org:        "open-cluster-management",
		Images:     images,
		PullPolicy: string(utils.GetImagePullPolicy(mce)),
		PullSecret: mce.Spec.ImagePullSecret,
		Namespace:  mce.Spec.TargetNamespace,
		// ConfigSecret: "",

		NodeSelector: mce.Spec.NodeSelector,
		// ProxyConfigs: map[string]string{},
		ReplicaCount:         utils.DefaultReplicaCount(mce),
		Tolerations:          utils.DefaultTolerations(), // TODO: update method
		OCPVersion:           os.Getenv("ACM_HUB_OCP_VERSION"),
		ClusterIngressDomain: os.Getenv("ACM_CLUSTER_INGRESS_DOMAIN"),
		HubType:              utils.GetHubType(mce),
	}

	return values
}

//go:embed templates
var templateFS embed.FS

var val = Values{
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
			Key:      "dedicated",
			Operator: "Exists",
			Effect:   "NoSchedule",
			Value:    "test",
		},
	},
}

func PrintDiscoveryFiles() {
	files, err := getComponentFiles(mcev1.Discovery)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		println(f)
		b, err := fs.ReadFile(templateFS, f)
		if err != nil {
			panic(err)
		}
		t := template.Must(template.New(f).Parse(string(b)))
		err = t.Execute(os.Stdout, val)
		if err != nil {
			panic(err)
		}

		// println(string(b))
	}
}

// getComponentFiles returns a list of filepaths comprising the component
func getComponentFiles(component string) ([]string, error) {
	path, ok := componentMapping[component]
	if !ok {
		return []string{}, fmt.Errorf("Unknown component")
	}

	return getFilesRecursive(templateFS, path)
}

func getFilesRecursive(f fs.FS, root string) ([]string, error) {
	res := []string{}
	err := fs.WalkDir(f, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		res = append(res, path)
		return nil
	})
	return res, err
}
