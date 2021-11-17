package readiness

import (
	"fmt"
	"time"

	"github.com/deckhouse/deckhouse/dhctl/pkg/log"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/retry"
)

var (
	ErrNotReady = fmt.Errorf("Not ready.")
)

type NodeChecker interface {
	IsReady(nodeName string) (bool, error)
	Name() string
}

type NodeGroupChecker struct {
	nodesNamesToCheck []string
	checkers          []NodeChecker
	sourceCommandName string
}

func NewEmptyChecker() *NodeGroupChecker {
	return NewChecker(nil, nil)
}

func NewChecker(nodesNamesToCheck []string, checkers []NodeChecker) *NodeGroupChecker {
	return &NodeGroupChecker{
		nodesNamesToCheck: nodesNamesToCheck,
		checkers:          checkers,
	}
}

func (c *NodeGroupChecker) WithSourceCommandName(name string) *NodeGroupChecker {
	c.sourceCommandName = name
	return c
}

func (c *NodeGroupChecker) IsReady() error {
	if c.checkers == nil {
		return nil
	}

	if len(c.nodesNamesToCheck) == 0 {
		return fmt.Errorf("Do not have control plane nodes to readiness check.")
	}

	return log.Process(c.sourceCommandName, "Control plane readiness check", func() error {
		for _, nodeName := range c.nodesNamesToCheck {
			ready, err := c.NodeIsReady(nodeName)
			if err != nil {
				return err
			}

			if !ready {
				return ErrNotReady
			}
		}

		return nil
	})

}

func (c *NodeGroupChecker) NodeIsReady(nodeName string) (bool, error) {
	title := fmt.Sprintf("Node %s is ready", nodeName)
	var lastErr error

	err := retry.NewLoop(title, 30, 10*time.Second).Run(func() error {
		for _, check := range c.checkers {
			err := log.Process(c.sourceCommandName, check.Name(), func() error {
				isReady, err := check.IsReady(nodeName)
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
