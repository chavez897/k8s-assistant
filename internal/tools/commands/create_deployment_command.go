package commands

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"k8s-assistant/internal/k8s"
	"k8s-assistant/internal/ollama"
	"strings"
)

type CreateDeploymentCommand struct{}

func (CreateDeploymentCommand) Name() ToolName {
	return ToolNameCreateDeployment
}

func (CreateDeploymentCommand) Definition() ollama.Tools {
	return ollama.Tools{
		Type: "function",
		Function: ollama.Function{
			Name:        string(CreateDeploymentCommand{}.Name()),
			Description: "Create a new deployment in the namespace selected by the CLI.",
			Parameters: ollama.FunctionParameters{
				Type: "object",
				Properties: map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The name of the deployment.",
					},
					"image": map[string]any{
						"type":        "string",
						"description": "The container image to use for the deployment.",
					},
					"replicas": map[string]any{
						"type":        "integer",
						"description": "The number of replicas for the deployment.",
						"minimum":     1,
						"maximum":     5,
					},
				},
				Required: []string{"name", "image", "replicas"},
			},
		},
	}
}

func (CreateDeploymentCommand) Execute(ctx context.Context, client *k8s.Client, namespace string, toolCall ollama.OllamaToolCall) (ollama.OllamaMessage, error) {
	request, err := parseCreateDeploymentArguments(toolCall.Function.Arguments)
	if err != nil {
		return ollama.OllamaMessage{}, fmt.Errorf("could not parse the arguments: %w", err)
	}
	deploymentInfo, err := client.CreateDeployment(ctx, namespace, request)
	if err != nil {
		return ollama.OllamaMessage{}, fmt.Errorf("could not create Kubernetes deployment: %w", err)
	}
	return newToolMessage(namespace, "deployments", toolCall.Id, ToolNameCreateDeployment, deploymentInfo)
}

func parseCreateDeploymentArguments(args map[string]any) (k8s.CreateDeploymentRequest, error) {
	data, err := json.Marshal(args)
	if err != nil {
		return k8s.CreateDeploymentRequest{}, fmt.Errorf("could not encode deployment arguments: %w", err)
	}

	var request k8s.CreateDeploymentRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return k8s.CreateDeploymentRequest{}, fmt.Errorf("invalid deployment arguments: %w", err)
	}

	if strings.TrimSpace(request.Name) == "" {
		return k8s.CreateDeploymentRequest{}, fmt.Errorf("deployment name is required")
	}
	if strings.TrimSpace(request.Image) == "" {
		return k8s.CreateDeploymentRequest{}, fmt.Errorf("container image is required")
	}
	if request.Replicas < 1 || request.Replicas > 5 {
		return k8s.CreateDeploymentRequest{}, fmt.Errorf("replicas must be between 1 and 5")
	}

	return request, nil
}
