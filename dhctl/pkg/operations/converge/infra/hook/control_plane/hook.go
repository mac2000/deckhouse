package control_plane

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/actions/deckhouse"
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	"github.com/deckhouse/deckhouse/dhctl/pkg/log"
	"github.com/deckhouse/deckhouse/dhctl/pkg/operations/converge/infra/hook"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/input"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/retry"
)

const convergeLabel = "dhctl.deckhouse.io/node-for-converge"

type Hook struct {
	nodesNamesToCheck []string
	checkers          []hook.NodeChecker
	sourceCommandName string
	kubeCl            *client.KubernetesClient
	nodeToConverge    string
	visitedNodes      map[string]struct{}
}

func NewHook(kubeCl *client.KubernetesClient, nodesToCheckWithIPs map[string]string) *Hook {
	proxyChecker := NewKubeProxyChecker().
		WithExternalIPs(nodesToCheckWithIPs)

	checkers := []hook.NodeChecker{
		hook.NewKubeNodeReadinessChecker(kubeCl),
		proxyChecker,
		NewManagerReadinessChecker(kubeCl),
	}

	nodes := make([]string, 0)
	for nodeName := range nodesToCheckWithIPs {
		nodes = append(nodes, nodeName)
	}

	return &Hook{
		nodesNamesToCheck: nodes,
		checkers:          checkers,
		kubeCl:            kubeCl,
	}
}

func (h *Hook) WithSourceCommandName(name string) *Hook {
	h.sourceCommandName = name
	return h
}

func (h *Hook) WithNodeToConverge(nodeToConverge string) *Hook {
	h.nodeToConverge = nodeToConverge
	return h
}

func (h *Hook) convergeLabelToNode(add bool) error {
	node, err := h.kubeCl.CoreV1().Nodes().Get(context.TODO(), h.nodeToConverge, metav1.GetOptions{})
	if err != nil {
		return err
	}

	labels := node.GetLabels()

	if add {
		if _, ok := labels[convergeLabel]; ok {
			return nil
		}

		labels[convergeLabel] = ""
	} else {
		if _, ok := labels[convergeLabel]; !ok {
			return nil
		}

		delete(labels, convergeLabel)
	}

	node.SetLabels(labels)

	_, err = h.kubeCl.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})

	return err
}

func (h *Hook) BeforeAction() (bool, error) {
	runAfterAction := false
	err := log.Process(h.sourceCommandName, "Check deckhouse pod is not on converged node", func() error {
		var pod *v1.Pod
		err := retry.NewSilentLoop("Get deckhouse pod", 10, 3*time.Second).Run(func() error {
			var err error
			pod, err = deckhouse.GetRunningPod(h.kubeCl)

			return err
		})

		if err != nil {
			return fmt.Errorf("Deckhouse pod did not get: %s", err)
		}

		if pod.Spec.NodeName != h.nodeToConverge {
			runAfterAction = false
			return nil
		}

		confirm := input.NewConfirmation().
			WithMessage("Deckhouse pod located on node to converge. Do you want to move pod in another node?")

		if !confirm.Ask() {
			log.WarnLn("Skip moving deckhouse pod")
			runAfterAction = false
			return nil
		}

		title := fmt.Sprintf("Set label '%s' on converged node", convergeLabel)
		err = retry.NewLoop(title, 10, 3*time.Second).Run(func() error {
			return h.convergeLabelToNode(true)
		})

		if err != nil {
			return fmt.Errorf("Cannot set label '%s' to node: %v", convergeLabel, err)
		}

		err = retry.NewLoop("Evict deckhouse pod from node", 10, 3*time.Second).Run(func() error {
			return deckhouse.RestartPod(h.kubeCl)
		})

		if err != nil {
			return err
		}

		runAfterAction = true

		return nil
	})

	if err != nil {
		return false, err
	}

	return false, nil
}

func (h *Hook) AfterAction() error {
	title := fmt.Sprintf("Delete label '%s' from converged node", convergeLabel)
	return retry.NewLoop(title, 10, 3*time.Second).Run(func() error {
		return h.convergeLabelToNode(false)
	})
}

func (h *Hook) IsReady() error {
	err := deckhouse.WaitForReadiness(h.kubeCl)
	if err != nil {
		return err
	}

	return hook.IsAllNodesReady(h.checkers, h.nodesNamesToCheck, h.sourceCommandName, "Control plane readiness check")
}
