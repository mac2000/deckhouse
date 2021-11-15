package control_plane

import (
	"fmt"
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
)

type ControlPlaneManagerReadinessChecker struct {
	hostName string
	kubeCl   *client.KubeClient
}

func NewControlPlaneManagerReadinessChecker(hostName string, kubeCl *client.KubeClient) *ControlPlaneManagerReadinessChecker {
	return &ControlPlaneManagerReadinessChecker{
		hostName: hostName,
		kubeCl:   kubeCl,
	}
}

func (c *ControlPlaneManagerReadinessChecker) IsReady() (bool, error) {
	return false, fmt.Errorf("Not implemented")
}

func (c *ControlPlaneManagerReadinessChecker) Name() string {
	return "Ssh access and kube-proxy availability"
}
