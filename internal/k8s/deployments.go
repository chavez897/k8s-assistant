package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetDeployments(ctx context.Context, namespace string) ([]DeploymentInfo, error) {
	deployments, err := c.clientSet.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not list Kubernetes deployments: %w", err)
	}
	var deploymentInfos []DeploymentInfo
	for _, deployment := range deployments.Items {
		deploymentInfos = append(deploymentInfos, toDeploymentInfo(&deployment))
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

func (c *Client) CreateDeployment(ctx context.Context, namespace string, request CreateDeploymentRequest) (DeploymentInfo, error) {
	deploymentSpec := buildDeploymentSpec(request)
	deployment, err := c.clientSet.AppsV1().Deployments(namespace).Create(ctx, &deploymentSpec, metav1.CreateOptions{})
	if err != nil {
		return DeploymentInfo{}, fmt.Errorf("error creating the deployment: %w", err)
	}
	return toDeploymentInfo(deployment), nil
}

func toDeploymentInfo(deployment *appsv1.Deployment) DeploymentInfo {
	return DeploymentInfo{
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
}

func buildDeploymentSpec(request CreateDeploymentRequest) appsv1.Deployment {
	replicas := request.Replicas
	labels := map[string]string{"app": request.Name}
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: request.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: request.Image,
						},
					},
				},
			},
		},
	}

}

type CreateDeploymentRequest struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	Replicas int32  `json:"replicas"`
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
