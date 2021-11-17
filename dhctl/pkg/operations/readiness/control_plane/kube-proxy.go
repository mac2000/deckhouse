package control_plane

import (
	"context"
	"fmt"
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	"github.com/deckhouse/deckhouse/dhctl/pkg/log"
	"github.com/deckhouse/deckhouse/dhctl/pkg/system/ssh"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type KubeProxyChecker struct {
	initParams       *client.KubernetesInitParams
	nsToCheck        string
	logCheckResult   bool
	askPassword      bool
	stopProxy        bool
	nodesExternalIPs map[string]string
}

func NewKubeProxyChecker() *KubeProxyChecker {
	return &KubeProxyChecker{
		nsToCheck: "kube-system",
		stopProxy: true,
	}
}

func (c *KubeProxyChecker) WithInitParams(p *client.KubernetesInitParams) *KubeProxyChecker {
	c.initParams = p
	return c
}

func (c *KubeProxyChecker) WithLogResult(f bool) *KubeProxyChecker {
	c.logCheckResult = f
	return c
}

func (c *KubeProxyChecker) WithAskPassword(f bool) *KubeProxyChecker {
	c.askPassword = f
	return c
}

func (c *KubeProxyChecker) WithStopProxy(f bool) *KubeProxyChecker {
	c.stopProxy = f
	return c
}

func (c *KubeProxyChecker) WithNamespaceToCheck(ns string) *KubeProxyChecker {
	c.nsToCheck = ns
	return c
}

func (c *KubeProxyChecker) WithExternalIPs(ips map[string]string) *KubeProxyChecker {
	c.nodesExternalIPs = ips
	return c
}

func (c *KubeProxyChecker) IsReady(nodeName string) (bool, error) {
	sshClient := ssh.NewClientFromFlags()

	if len(c.nodesExternalIPs) > 0 {
		ip, ok := c.nodesExternalIPs[nodeName]
		if !ok {
			return false, fmt.Errorf("Not found external ip for node %s", nodeName)
		}

		sshClient.Settings.SetAvailableHosts([]string{ip})
	}

	var err error
	sshClient, err = sshClient.Start()
	if err != nil {
		return false, err
	}

	kubeCl := client.NewKubernetesClient().WithSSHClient(sshClient)
	err = kubeCl.Init(client.AppKubernetesInitParams())
	if err != nil {
		return false, fmt.Errorf("open kubernetes connection: %v", err)
	}

	ns, err := kubeCl.CoreV1().Namespaces().Get(context.TODO(), "kube-system", v1.GetOptions{})
	if err != nil {
		return false, err
	}

	c.printNs(ns)

	if c.stopProxy {
		if kubeCl.KubeProxy != nil {
			kubeCl.KubeProxy.Stop()
		}

		if kubeCl.SSHClient != nil {
			kubeCl.SSHClient.Stop()
		}
	}

	return true, nil
}

func (c *KubeProxyChecker) Name() string {
	return "Ssh access and kube-proxy availability"
}

func (c *KubeProxyChecker) printNs(ns *apiv1.Namespace) {
	if !c.logCheckResult {
		return
	}

	yamlRepr, err := yaml.Marshal(ns)
	if err != nil {
		log.ErrorF("Ns marshal error %v\n", err)
		return
	}

	log.InfoF("Namespace '%s' info:\n%s\n", c.nsToCheck, string(yamlRepr))
}
