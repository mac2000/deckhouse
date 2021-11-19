package hook

import (
	"context"
	"fmt"
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	"github.com/deckhouse/deckhouse/dhctl/pkg/log"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/retry"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var (
	ErrNotReady = fmt.Errorf("Not ready.")
)

type NodeChecker interface {
	IsReady(nodeName string) (bool, error)
	Name() string
}

func IsAllNodesReady(checkers []NodeChecker, nodes []string, sourceCommandName, processName string) error {
	if checkers == nil {
		return nil
	}

	if len(nodes) == 0 {
		return fmt.Errorf("Do not have nodes for %s.", processName)
	}

	return log.Process(sourceCommandName, processName, func() error {
		for _, nodeName := range nodes {
			ready, err := IsNodeReady(checkers, nodeName, sourceCommandName)
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

func IsNodeReady(checkers []NodeChecker, nodeName, sourceCommandName string) (bool, error) {
	title := fmt.Sprintf("Node %s is ready", nodeName)
	var lastErr error

	err := retry.NewLoop(title, 30, 10*time.Second).Run(func() error {
		for _, check := range checkers {
			err := log.Process(sourceCommandName, check.Name(), func() error {
				isReady, err := check.IsReady(nodeName)
				if err != nil {
					return err
				}

				if !isReady {
					return ErrNotReady
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

type KubeNodeReadinessChecker struct {
	kubeCl *client.KubernetesClient
}

func NewKubeNodeReadinessChecker(kubeCl *client.KubernetesClient) *KubeNodeReadinessChecker {
	return &KubeNodeReadinessChecker{
		kubeCl: kubeCl,
	}
}

func (c *KubeNodeReadinessChecker) IsReady(nodeName string) (bool, error) {
	node, err := c.kubeCl.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	for _, c := range node.Status.Conditions {
		if c.Type == apiv1.NodeReady {
			if c.Status == apiv1.ConditionTrue {
				return true, nil
			}
		}
	}

	return false, nil
}

func (c *KubeNodeReadinessChecker) Name() string {
	return "Kube node is ready"
}
