package commands

import (
	"context"
	"fmt"

	"k8s-assistant/internal/k8s"
	"k8s-assistant/internal/ollama"
)

type ListDeploymentsCommand struct{}

func (ListDeploymentsCommand) Name() ToolName {
	return ToolNameListDeployments
}

func (ListDeploymentsCommand) Definition() ollama.Tools {
	return ollama.Tools{
		Type: "function",
		Function: ollama.Function{
			Name:        string(ListDeploymentsCommand{}.Name()),
			Description: "Read-only: list deployment status in the namespace selected by the CLI.",
			Parameters: ollama.FunctionParameters{
				Type:       "object",
				Properties: map[string]any{},
				Required:   []string{},
			},
		},
	}
}

func (ListDeploymentsCommand) Execute(ctx context.Context, client *k8s.Client, namespace string,
	toolCall ollama.OllamaToolCall) (ollama.OllamaMessage, error) {
	deployments, err := client.GetDeployments(ctx, namespace)
	if err != nil {
		return ollama.OllamaMessage{}, fmt.Errorf("could not get Kubernetes deployments: %w", err)
	}
	return newToolMessage(namespace, "deployments", toolCall.Id, ToolNameListDeployments, deployments)
}
