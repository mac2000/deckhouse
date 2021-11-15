package control_plane

import (
	"fmt"
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	"github.com/deckhouse/deckhouse/dhctl/pkg/log"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/maputils"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/retry"
	"time"
)

type checker interface {
	IsReady() (bool, error)
	Name() string
}

type ControlPlaneNodesReadinessController struct {
	kubeCl    *client.KubeClient
	addresses map[string]string

	sourceCommandName string
}

func NewControlPlaneChecker(kubeCl *client.KubeClient, allAddresses map[string]string) *ControlPlaneNodesReadinessController {
	return &ControlPlaneNodesReadinessController{
		addresses: allAddresses,
		kubeCl:    kubeCl,
	}
}

func (c *ControlPlaneNodesReadinessController) WithSourceCommandName(name string) *ControlPlaneNodesReadinessController {
	c.sourceCommandName = name

	return c
}

func (c *ControlPlaneNodesReadinessController) IsReady(excludeHostNames ...string) (bool, error) {
	nodesToCheck := maputils.ExcludeKeys(c.addresses, excludeHostNames...)

	if len(nodesToCheck) == 0 {
		return false, fmt.Errorf("do not have control plane nodes to readiness check. passed addresses %v, excluded addresses %s", c.addresses, excludeHostNames)
	}

	for nodeHostName, ip := range nodesToCheck {
		checkers := []checker{
			NewKubeProxyChecker(ip),
			NewControlPlaneManagerReadinessChecker(nodeHostName, c.kubeCl),
		}

		title := fmt.Sprintf("Check control plane node %s is ready", nodeHostName)
		var lastErr error

		err := retry.NewLoop(title, 30, 10*time.Second).Run(func() error {
			for _, check := range checkers {
				err := c.runChecker(check)
				if err != nil {
					lastErr = err
					return err
				}
			}

			return nil
		})

		if err != nil {
			return false, fmt.Errorf("Node %s is not ready. last error: %v/%v", err, lastErr)
		}

	}

	return true, nil
}

func (c *ControlPlaneNodesReadinessController) runChecker(check checker) error {
	return log.Process(c.sourceCommandName, check.Name(), func() error {
		isReady, err := check.IsReady()
		if err != nil {
			return err
		}

		if !isReady {
			return fmt.Errorf("not ready")
		}

		return err
	})
}
