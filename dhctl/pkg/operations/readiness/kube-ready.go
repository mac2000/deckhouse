package readiness

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
)

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
