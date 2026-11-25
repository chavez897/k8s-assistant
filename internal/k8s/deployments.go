package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetDeployments(ctx context.Context, namespace string) []DeploymentInfo {
	fmt.Printf("Fetching deployments in namespace: %s\n", namespace)
	deployments, err := c.clientSet.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Errorf("could not list Kubernetes deployments: %w", err)
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
	fmt.Printf("Deployments %+v\n", deploymentInfos)
	return deploymentInfos
}

func getContainerImages(containers []corev1.Container) []string {
	var images []string
	for _, container := range containers {
		images = append(images, container.Image)
	}
	return images
}

type DeploymentInfo struct {
	Name                string
	Namespace           string
	DesiredReplicas     *int32
	ReadyReplicas       int32
	AvailableReplicas   int32
	UpdatedReplicas     int32
	UnavailableReplicas int32
	ContainerImages     []string
	Labels              map[string]string
	Selector            map[string]string
	Strategy            string
	Conditions          []string
}
