package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	fmt.Printf("Fetching pods in namespace: %s\n", namespace)
	pods, err := c.clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not list Kubernetes pods: %w", err)
	}
	var result []PodInfo
	for _, pod := range pods.Items {
		if len(pod.Status.ContainerStatuses) == 0 {
			result = append(result, PodInfo{
				Name:          pod.Name,
				CurrentStatus: string(pod.Status.Phase),
			})
		}
		for _, containerStatus := range pod.Status.ContainerStatuses {
			podInfo := PodInfo{
				Name:          pod.Name,
				CurrentStatus: string(pod.Status.Phase),
				ContainerName: containerStatus.Name,
				RestartCount:  containerStatus.RestartCount,
			}
			if containerStatus.State.Terminated != nil {
				podInfo.Reason = containerStatus.State.Terminated.Reason
			} else if containerStatus.State.Waiting != nil {
				podInfo.Reason = containerStatus.State.Waiting.Reason
			}
			result = append(result, podInfo)
		}
	}
	return result, nil
}

type PodInfo struct {
	Name          string `json:"name"`
	CurrentStatus string `json:"currentStatus"`
	ContainerName string `json:"containerName"`
	RestartCount  int32  `json:"restartCount"`
	Reason        string `json:"reason"`
}
