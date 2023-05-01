package cluster

import (
	"context"
	"fmt"
	"os"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/stolostron/backplane-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	ConsoleEnabled bool
	// Version of OCP
	ClusterVersion string
	ProxyConfig    map[string]string
	IngressDomain  string
	// Can be multicluster-engine or stolostron-engine
	OperatorMode string
	// OperatorNamespace string
}

func NewConfigFromEnv(k8sClient client.Client) (*Config, error) {
	c := &Config{}
	c.setOperatorMode()
	c.setProxyConfig()
	err := c.setClusterIngressDomain(k8sClient)
	if err != nil {
		return nil, err
	}
	err = c.setClusterVersion(k8sClient)
	if err != nil {
		return nil, err
	}
	err = c.setConsoleEnabled(k8sClient)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func UpdateConfig(c *Config, k8sClient client.Client) error {
	err := c.setClusterIngressDomain(k8sClient)
	if err != nil {
		return err
	}
	err = c.setClusterVersion(k8sClient)
	if err != nil {
		return err
	}
	err = c.setConsoleEnabled(k8sClient)
	if err != nil {
		return err
	}
	return nil
}

func (cc Config) setOperatorMode() {
	cc.OperatorMode = utils.OperatorMode()
}

func (cc Config) setProxyConfig() {
	// OLM handles these environment variables as a unit;
	// if at least one of them is set, all three are considered overridden
	// and the cluster-wide defaults are not used for the deployments of the subscribed Operator.
	// https://docs.openshift.com/container-platform/4.12/operators/admin/olm-configuring-proxy-support.html

	http := os.Getenv("HTTP_PROXY")
	https := os.Getenv("HTTPS_PROXY")
	noProxy := os.Getenv("NO_PROXY")

	if http != "" || https != "" || noProxy != "" {
		proxyVars := map[string]string{
			"HTTP_PROXY":  http,
			"HTTPS_PROXY": https,
			"NO_PROXY":    noProxy,
		}
		cc.ProxyConfig = proxyVars
	}
}

func (cc Config) setClusterIngressDomain(k8sClient client.Client) error {
	clusterIngress := &configv1.Ingress{}
	err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: "cluster"}, clusterIngress)
	if err != nil {
		return fmt.Errorf("get ingress: %v", err)
	}

	if clusterIngress.Spec.Domain == "" {
		return fmt.Errorf("ingress domain is empty")
	}

	cc.IngressDomain = clusterIngress.Spec.Domain
	return nil
}

func (cc Config) setClusterVersion(k8sClient client.Client) error {
	clusterVersion := &configv1.ClusterVersion{}
	err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: "version"}, clusterVersion)
	if err != nil {
		return fmt.Errorf("get clusterversion: %v", err)
	}

	if len(clusterVersion.Status.History) == 0 {
		return fmt.Errorf("cluster version history is empty")
	}

	cc.ClusterVersion = clusterVersion.Status.History[0].Version
	return nil
}

func (cc Config) setConsoleEnabled(k8sClient client.Client) error {
	// TODO: Copy code from CheckConsole()
	cc.ConsoleEnabled = true
	return nil
}
