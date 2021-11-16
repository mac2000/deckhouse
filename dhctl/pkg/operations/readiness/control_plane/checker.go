package control_plane

import (
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	"github.com/deckhouse/deckhouse/dhctl/pkg/operations/readiness"
)

func NewChecker(kubeCl *client.KubernetesClient, allAddresses map[string]string) *readiness.NodeGroupChecker {
	checkers := []readiness.NodeChecker{
		readiness.NewKubeNodeReadinessChecker(kubeCl),
		NewKubeProxyChecker(),
		NewManagerReadinessChecker(kubeCl),
	}

	return readiness.NewChecker(allAddresses, checkers)
}
