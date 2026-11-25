package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetDeployments(ctx context.Context, namespace string) ([]DeploymentInfo, error) {
	fmt.Printf("Fetching deployments in namespace: %s\n", namespace)
	deployments, err := c.clientSet.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not list Kubernetes deployments: %w", err)
	}
	var deploymentInfos []DeploymentInfo
	for _, deployment := range deployments.Items {
		deploymentInfo := DeploymentInfo{
			Name:                deployment.Name,
			Namespace:           deployment.Namespace,
			DesiredReplicas:     deployment.Spec.Replicas,
			ReadyReplicas:       deployment.Status.ReadyReplicas,
			AvailableReplicas:   deployment.Status.AvailableReplicas,
			UpdatedReplicas:     deployment.Status.UpdatedReplicas,
			UnavailableReplicas: deployment.Status.UnavailableReplicas,
			ContainerImages:     getContainerImages(deployment.Spec.Template.Spec.Containers),
			Labels:              deployment.Labels,
			Selector:            deployment.Spec.Selector.MatchLabels,
			Strategy:            string(deployment.Spec.Strategy.Type),
		}
		deploymentInfos = append(deploymentInfos, deploymentInfo)
	}
	return deploymentInfos, nil
}

func getContainerImages(containers []corev1.Container) []string {
	var images []string
	for _, container := range containers {
		images = append(images, container.Image)
	}
	return images
}

type DeploymentInfo struct {
	Name                string            `json:"name"`
	Namespace           string            `json:"namespace"`
	DesiredReplicas     *int32            `json:"desiredReplicas"`
	ReadyReplicas       int32             `json:"readyReplicas"`
	AvailableReplicas   int32             `json:"availableReplicas"`
	UpdatedReplicas     int32             `json:"updatedReplicas"`
	UnavailableReplicas int32             `json:"unavailableReplicas"`
	ContainerImages     []string          `json:"containerImages"`
	Labels              map[string]string `json:"labels"`
	Selector            map[string]string `json:"selector"`
	Strategy            string            `json:"strategy"`
	Conditions          []string          `json:"conditions"`
}
