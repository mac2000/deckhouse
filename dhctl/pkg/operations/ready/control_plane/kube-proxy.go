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
	ip             string
	initParams     *client.KubernetesInitParams
	nsToCheck      string
	logCheckResult bool
	askPassword    bool
}

func NewKubeProxyChecker(ip string) *KubeProxyChecker {
	return &KubeProxyChecker{
		ip:        ip,
		nsToCheck: "kube-system",
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

func (c *KubeProxyChecker) WithNamespaceToCheck(ns string) *KubeProxyChecker {
	c.nsToCheck = ns
	return c
}

func (c *KubeProxyChecker) IsReady() (bool, error) {
	sshClient := ssh.NewClientFromFlags()
	if c.ip != "" {
		sshClient.Settings.SetAvailableHosts([]string{c.ip})
	}

	var err error
	sshClient, err = sshClient.Start()
	if err != nil {
		return false, err
	}

	defer func() {
		sshClient.Stop()
	}()

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

	log.InfoF("Namespace %s info \n: %s\n", string(yamlRepr))
}
