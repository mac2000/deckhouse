package control_plane

import (
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	"github.com/deckhouse/deckhouse/dhctl/pkg/operations/readiness"
)

func NewChecker(kubeCl *client.KubernetesClient, nodesWithIPs map[string]string) *readiness.NodeGroupChecker {
	proxyChecker := NewKubeProxyChecker().
		WithExternalIPs(nodesWithIPs)

	checkers := []readiness.NodeChecker{
		readiness.NewKubeNodeReadinessChecker(kubeCl),
		proxyChecker,
		NewManagerReadinessChecker(kubeCl),
	}

	nodes := make([]string, 0)
	for nodeName := range nodesWithIPs {
		nodes = append(nodes, nodeName)
	}

	return readiness.NewChecker(nodes, checkers)
}
