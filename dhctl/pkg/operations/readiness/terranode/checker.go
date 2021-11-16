package terranode

import (
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	"github.com/deckhouse/deckhouse/dhctl/pkg/operations/readiness"
)

func NewChecker(kubeCl *client.KubernetesClient, allAddresses map[string]string) *readiness.NodeGroupChecker {
	checkers := []readiness.NodeChecker{
		readiness.NewKubeNodeReadinessChecker(kubeCl),
	}

	return readiness.NewChecker(allAddresses, checkers)
}
