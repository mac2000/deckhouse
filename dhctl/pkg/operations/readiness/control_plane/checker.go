package control_plane

import (
	"fmt"
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	"github.com/deckhouse/deckhouse/dhctl/pkg/log"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/maputils"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/retry"
	"time"
)

type nodeChecker interface {
	IsReady() (bool, error)
	Name() string
}

type ReadinessChecker struct {
	kubeCl    *client.KubernetesClient
	addresses map[string]string

	sourceCommandName string
}

func NewChecker(kubeCl *client.KubernetesClient, allAddresses map[string]string) *ReadinessChecker {
	return &ReadinessChecker{
		addresses: allAddresses,
		kubeCl:    kubeCl,
	}
}

func (c *ReadinessChecker) WithSourceCommandName(name string) *ReadinessChecker {
	c.sourceCommandName = name
	return c
}

func (c *ReadinessChecker) IsNodeGroupReady(excludeHostNames ...string) (bool, error) {
	nodesToCheck := maputils.ExcludeKeys(c.addresses, excludeHostNames...)

	if len(nodesToCheck) == 0 {
		return false, fmt.Errorf("do not have control plane nodes to readiness check. passed addresses %v, excluded addresses %s", c.addresses, excludeHostNames)
	}

	for nodeName := range nodesToCheck {
		ready, err := c.NodeIsReady(nodeName)
		if err != nil {
			return false, err
		}

		if !ready {
			return false, fmt.Errorf("node %s is not ready", nodeName)
		}
	}

	return true, nil
}

func (c *ReadinessChecker) NodeIsReady(nodeName string) (bool, error) {

	checkers, err := c.getCheckers(nodeName)
	if err != nil {
		return false, err
	}

	title := fmt.Sprintf("Check control plane node %s is ready", nodeName)
	var lastErr error

	err = retry.NewLoop(title, 30, 10*time.Second).Run(func() error {
		for _, check := range checkers {
			err := log.Process(c.sourceCommandName, check.Name(), func() error {
				isReady, err := check.IsReady()
				if err != nil {
					return err
				}

				if !isReady {
					return fmt.Errorf("not ready")
				}

				return err
			})

			if err != nil {
				lastErr = err
				return err
			}
		}

		return nil
	})

	if err != nil {
		return false, fmt.Errorf("Node %s is not ready. last error: %v/%v", nodeName, err, lastErr)
	}

	return true, nil
}

func (c *ReadinessChecker) getCheckers(nodeName string) ([]nodeChecker, error) {
	ip, ok := c.addresses[nodeName]
	if !ok {
		return nil, fmt.Errorf("Node %s not found", nodeName)
	}

	return []nodeChecker{
		NewKubeProxyChecker(ip),
		NewControlPlaneManagerReadinessChecker(nodeName, c.kubeCl),
	}, nil
}
