package readiness

import (
	"fmt"
	"time"

	"github.com/deckhouse/deckhouse/dhctl/pkg/log"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/maputils"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/retry"
)

type Node struct {
	Name       string
	ExternalIp string
}

type NodeChecker interface {
	IsReady(nodeName, ip string) (bool, error)
	Name() string
}

type NodeGroupChecker struct {
	nodesToAddresses  map[string]string
	checkers          []NodeChecker
	sourceCommandName string
}

func NewChecker(allAddresses map[string]string, checkers []NodeChecker) *NodeGroupChecker {
	return &NodeGroupChecker{
		nodesToAddresses: allAddresses,
		checkers:         checkers,
	}
}

func (c *NodeGroupChecker) WithSourceCommandName(name string) *NodeGroupChecker {
	c.sourceCommandName = name
	return c
}

func (c *NodeGroupChecker) IsNodeGroupReady(excludeHostNames ...string) (bool, error) {
	nodesToCheck := maputils.ExcludeKeys(c.nodesToAddresses, excludeHostNames...)

	if len(nodesToCheck) == 0 {
		return false, fmt.Errorf("do not have control plane nodes to readiness check. passed nodesToAddresses %v, excluded nodesToAddresses %s", c.nodesToAddresses, excludeHostNames)
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

func (c *NodeGroupChecker) NodeIsReady(nodeName string) (bool, error) {
	ip, ok := c.nodesToAddresses[nodeName]
	if !ok {
		return false, fmt.Errorf("Node %s not found", nodeName)
	}

	title := fmt.Sprintf("Check control plane node %s is ready", nodeName)
	var lastErr error

	err := retry.NewLoop(title, 30, 10*time.Second).Run(func() error {
		for _, check := range c.checkers {
			err := log.Process(c.sourceCommandName, check.Name(), func() error {
				isReady, err := check.IsReady(nodeName, ip)
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
