// Copyright Contributors to the Open Cluster Management project

package renderer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	loader "helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/fatih/structs"
	v1 "github.com/stolostron/backplane-operator/api/v1"
	"github.com/stolostron/backplane-operator/pkg/utils"
	"helm.sh/helm/v3/pkg/engine"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

const (
	AlwaysChartsDir = "pkg/templates/charts/always"
)

type Values struct {
	Global    Global    `json:"global" structs:"global"`
	HubConfig HubConfig `json:"hubconfig" structs:"hubconfig"`
	Org       string    `json:"org" structs:"org"`
}

type Global struct {
	ImageOverrides map[string]string `json:"imageOverrides" structs:"imageOverrides"`
	PullPolicy     string            `json:"pullPolicy" structs:"pullPolicy"`
	PullSecret     string            `json:"pullSecret" structs:"pullSecret"`
	Namespace      string            `json:"namespace" structs:"namespace"`
}

type HubConfig struct {
	NodeSelector         map[string]string `json:"nodeSelector" structs:"nodeSelector"`
	ProxyConfigs         map[string]string `json:"proxyConfigs" structs:"proxyConfigs"`
	ReplicaCount         int               `json:"replicaCount" structs:"replicaCount"`
	Tolerations          []Toleration      `json:"tolerations" structs:"tolerations"`
	OCPVersion           string            `json:"ocpVersion" structs:"ocpVersion"`
	ClusterIngressDomain string            `json:"clusterIngressDomain" structs:"clusterIngressDomain"`
}

type Toleration struct {
	Key               string                    `json:"Key" protobuf:"bytes,1,opt,name=key"`
	Operator          corev1.TolerationOperator `json:"Operator" protobuf:"bytes,2,opt,name=operator,casttype=TolerationOperator"`
	Value             string                    `json:"Value" protobuf:"bytes,3,opt,name=value"`
	Effect            corev1.TaintEffect        `json:"Effect" protobuf:"bytes,4,opt,name=effect,casttype=TaintEffect"`
	TolerationSeconds *int64                    `json:"TolerationSeconds" protobuf:"varint,5,opt,name=tolerationSeconds"`
}

func convertTolerations(tols []corev1.Toleration) []Toleration {
	var tolerations []Toleration
	for _, t := range tols {
		tolerations = append(tolerations, Toleration{
			Key:               t.Key,
			Operator:          t.Operator,
			Value:             t.Value,
			Effect:            t.Effect,
			TolerationSeconds: t.TolerationSeconds,
		})
	}
	return tolerations
}

// func (v *Values) MarshalJSON() ([]byte, error) {
// 	type Alias Values
// 	return json.Marshal(&struct {
// 		LastSeen int64 `json:"lastSeen"`
// 		*Alias
// 	}{
// 		LastSeen: u.LastSeen.Unix(),
// 		Alias:    (*Alias)(u),
// 	})
// }

func (val *Values) ToMap() map[string]interface{} {

	var inInterface map[string]interface{}
	inrec, _ := json.Marshal(val)
	json.Unmarshal(inrec, &inInterface)

	// iterate through inrecs
	for field, val := range inInterface {
		fmt.Println("KV Pair: ", field, val)
	}
	return inInterface
}

func (val *Values) ToMap2() map[string]interface{} {
	inInterface := structs.Map(val)
	// iterate through inrecs
	for field, val := range inInterface {
		fmt.Println("KV Pair: ", field, val)
	}
	return inInterface
}

func (val *Values) ToValues() chartutil.Values {
	inrec, _ := json.Marshal(val)
	vals, _ := chartutil.ReadValues(inrec)
	return vals
}

func (val *Values) ToMap3() map[string]interface{} {
	top := map[string]interface{}{
		"org": val.Org,
		"hubconfig": map[string]interface{}{
			"nodeSelector": val.HubConfig.NodeSelector,
			"proxyConfigs": val.HubConfig.ProxyConfigs,
			"replicaCount": val.HubConfig.ReplicaCount,
			"tolerations": map[string]interface{}{
				"Key": "1",
			},
			"ocpVersion":           val.HubConfig.OCPVersion,
			"clusterIngressDomain": val.HubConfig.ClusterIngressDomain,
		},
	}
	// "global": map[string]interface{}{
	// 	"Name":      options.Name,
	// 	"Namespace": options.Namespace,
	// 	"IsUpgrade": options.IsUpgrade,
	// 	"IsInstall": options.IsInstall,
	// 	"Revision":  options.Revision,
	// 	"Service":   "Helm",
	// },
	// }
	return top
}

func RenderCRDs(crdDir string) ([]*unstructured.Unstructured, []error) {
	var crds []*unstructured.Unstructured
	errs := []error{}

	if val, ok := os.LookupEnv("DIRECTORY_OVERRIDE"); ok {
		crdDir = path.Join(val, crdDir)
	}

	// Read CRD files
	err := filepath.Walk(crdDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		crd := &unstructured.Unstructured{}
		if info == nil || info.IsDir() {
			return nil
		}
		bytesFile, e := ioutil.ReadFile(path)
		if e != nil {
			errs = append(errs, fmt.Errorf("%s - error reading file: %v", info.Name(), err.Error()))
		}
		if err = yaml.Unmarshal(bytesFile, crd); err != nil {
			errs = append(errs, fmt.Errorf("%s - error unmarshalling file to unstructured: %v", info.Name(), err.Error()))
		}
		crds = append(crds, crd)
		return nil
	})
	if err != nil {
		return crds, errs
	}

	return crds, errs
}

func RenderCharts(chartDir string, backplaneConfig *v1.MultiClusterEngine, images map[string]string) ([]*unstructured.Unstructured, []error) {
	log := log.FromContext(context.Background())
	var templates []*unstructured.Unstructured
	errs := []error{}
	if val, ok := os.LookupEnv("DIRECTORY_OVERRIDE"); ok {
		chartDir = path.Join(val, chartDir)
	}
	charts, err := ioutil.ReadDir(chartDir)
	if err != nil {
		errs = append(errs, err)
	}
	for _, chart := range charts {
		chartPath := filepath.Join(chartDir, chart.Name())
		chartTemplates, errs := renderTemplates(chartPath, backplaneConfig, images)
		if len(errs) > 0 {
			for _, err := range errs {
				log.Info(err.Error())
			}
			return nil, errs
		}
		templates = append(templates, chartTemplates...)
	}
	return templates, nil
}

func RenderChart(chartPath string, backplaneConfig *v1.MultiClusterEngine, images map[string]string) ([]*unstructured.Unstructured, []error) {
	log := log.FromContext(context.Background())
	errs := []error{}
	if val, ok := os.LookupEnv("DIRECTORY_OVERRIDE"); ok {
		chartPath = path.Join(val, chartPath)
	}
	chartTemplates, errs := renderTemplates(chartPath, backplaneConfig, images)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Info(err.Error())
		}
		return nil, errs
	}
	return chartTemplates, nil

}

// RenderChartWithNamespace wraps the RenderChart function, overriding the target namespace
func RenderChartWithNamespace(chartPath string, backplaneConfig *v1.MultiClusterEngine, images map[string]string, namespace string) ([]*unstructured.Unstructured, []error) {
	mce := backplaneConfig.DeepCopy()
	mce.Spec.TargetNamespace = namespace
	return RenderChart(chartPath, mce, images)
}

func renderTemplates(chartPath string, backplaneConfig *v1.MultiClusterEngine, images map[string]string) ([]*unstructured.Unstructured, []error) {
	log := log.FromContext(context.Background())
	var templates []*unstructured.Unstructured
	errs := []error{}
	chart, err := loader.Load(chartPath)
	if err != nil {
		log.Info(fmt.Sprintf("error loading chart: %s", chart.Name()))
		return nil, append(errs, err)
	}
	valuesYaml := &Values{}
	injectValuesOverrides(valuesYaml, backplaneConfig, images)
	helmEngine := engine.Engine{
		Strict:   true,
		LintMode: false,
	}
	rawTemplates, err := helmEngine.Render(chart, chartutil.Values{"Values": valuesYaml.ToMap()})
	// rawTemplates, err := helmEngine.Render(chart, valuesYaml.ToValues())

	if err != nil {
		log.Info(fmt.Sprintf("error rendering chart: %s", chart.Name()))
		return nil, append(errs, err)
	}

	for fileName, templateFile := range rawTemplates {
		unstructured := &unstructured.Unstructured{}
		if err = yaml.Unmarshal([]byte(templateFile), unstructured); err != nil {
			return nil, append(errs, fmt.Errorf("error converting file %s to unstructured", fileName))
		}

		utils.AddBackplaneConfigLabels(unstructured, backplaneConfig.Name)

		// Add namespace to namespaced resources
		switch unstructured.GetKind() {
		case "Deployment", "ServiceAccount", "Role", "RoleBinding", "Service", "ConfigMap", "Route":
			unstructured.SetNamespace(backplaneConfig.Spec.TargetNamespace)
		}
		templates = append(templates, unstructured)
	}

	return templates, errs
}

func injectValuesOverrides(values *Values, backplaneConfig *v1.MultiClusterEngine, images map[string]string) {

	values.Global.ImageOverrides = images

	values.Global.PullPolicy = string(utils.GetImagePullPolicy(backplaneConfig))

	values.Global.Namespace = backplaneConfig.Spec.TargetNamespace

	values.Global.PullSecret = backplaneConfig.Spec.ImagePullSecret

	values.HubConfig.ReplicaCount = utils.DefaultReplicaCount(backplaneConfig)

	values.HubConfig.NodeSelector = backplaneConfig.Spec.NodeSelector

	if len(backplaneConfig.Spec.Tolerations) > 0 {
		values.HubConfig.Tolerations = convertTolerations(backplaneConfig.Spec.Tolerations)
	} else {
		values.HubConfig.Tolerations = convertTolerations(utils.DefaultTolerations())
	}

	values.Org = "open-cluster-management"

	values.HubConfig.OCPVersion = os.Getenv("ACM_HUB_OCP_VERSION")

	values.HubConfig.ClusterIngressDomain = os.Getenv("ACM_CLUSTER_INGRESS_DOMAIN")

	if utils.ProxyEnvVarsAreSet() {
		proxyVar := map[string]string{}
		proxyVar["HTTP_PROXY"] = os.Getenv("HTTP_PROXY")
		proxyVar["HTTPS_PROXY"] = os.Getenv("HTTPS_PROXY")
		proxyVar["NO_PROXY"] = os.Getenv("NO_PROXY")
		values.HubConfig.ProxyConfigs = proxyVar
	}

	// TODO: Define all overrides
}
