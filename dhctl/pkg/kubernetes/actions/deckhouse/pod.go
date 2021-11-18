package deckhouse

import (
	"context"
	"fmt"
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPod(kubeCl *client.KubernetesClient) (*v1.Pod, error) {
	pods, err := kubeCl.CoreV1().Pods("d8-system").List(context.TODO(), metav1.ListOptions{LabelSelector: "app=deckhouse"})
	if err != nil {
		return nil, ErrListPods
	}

	if len(pods.Items) != 1 {
		return nil, ErrListPods
	}

	pod := pods.Items[0]

	return &pod, nil
}

func GetRunningPod(kubeCl *client.KubernetesClient) (*v1.Pod, error) {
	pod, err := GetPod(kubeCl)
	if err != nil {
		return nil, err
	}

	phase := pod.Status.Phase
	message := fmt.Sprintf("Deckhouse pod found: %s (%s)", pod.Name, pod.Status.Phase)

	if phase != v1.PodRunning {
		return nil, fmt.Errorf(message)
	}

	return pod, nil
}

func RestartPod(kubeCl *client.KubernetesClient) error {
	pod, err := GetPod(kubeCl)
	if err != nil {
		return err
	}

	return kubeCl.CoreV1().Pods("d8-system").Delete(context.TODO(), pod.GetName(), metav1.DeleteOptions{})
}
