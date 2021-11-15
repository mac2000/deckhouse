package control_plane

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
)

type ControlPlaneManagerReadinessChecker struct {
	hostName string
	kubeCl   *client.KubernetesClient
}

func NewControlPlaneManagerReadinessChecker(hostName string, kubeCl *client.KubernetesClient) *ControlPlaneManagerReadinessChecker {
	return &ControlPlaneManagerReadinessChecker{
		hostName: hostName,
		kubeCl:   kubeCl,
	}
}

func (c *ControlPlaneManagerReadinessChecker) IsReady() (bool, error) {
	cpmPodsList, err := c.kubeCl.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=d8-control-plane-manager",
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", c.hostName),
	})

	if err != nil {
		return false, err
	}

	if len(cpmPodsList.Items) == 0 {
		return false, fmt.Errorf("Not found control plane manage pod")
	}

	if len(cpmPodsList.Items) > 1 {
		return false, fmt.Errorf("Found multiple control plane manager pods for one node")
	}

	cpmPod := cpmPodsList.Items[0]
	podName := cpmPod.GetName()
	phase := cpmPod.Status.Phase

	if cpmPod.Status.Phase != corev1.PodRunning {
		return false, fmt.Errorf("Control plane manager pod %s is not running (%s)", podName, phase)
	}

	for _, status := range cpmPod.Status.ContainerStatuses {
		if status.Name != "control-plane-manager" {
			continue
		}

		return status.Ready, nil
	}

	return false, fmt.Errorf("Not found control-plane-manager container in pod %s", podName)
}

func (c *ControlPlaneManagerReadinessChecker) Name() string {
	return "Control plane manager readiness"
}
