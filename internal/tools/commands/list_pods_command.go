package commands

import (
	"context"
	"fmt"

	"k8s-assistant/internal/k8s"
	"k8s-assistant/internal/ollama"
)

type ListPodsCommand struct{}

func (ListPodsCommand) Name() ToolName {
	return ToolNameListPods
}

func (ListPodsCommand) Definition() ollama.Tools {
	return ollama.Tools{
		Type: "function",
		Function: ollama.Function{
			Name:        string(ListPodsCommand{}.Name()),
			Description: "Read-only: list pod status in the namespace selected by the CLI.",
			Parameters: ollama.FunctionParameters{
				Type:       "object",
				Properties: map[string]any{},
				Required:   []string{},
			},
		},
	}
}

func (ListPodsCommand) Execute(ctx context.Context, client *k8s.Client, namespace string,
	toolCall ollama.OllamaToolCall) (ollama.OllamaMessage, error) {
	pods, err := client.GetPods(ctx, namespace)
	if err != nil {
		return ollama.OllamaMessage{}, fmt.Errorf("could not get Kubernetes pods: %w", err)
	}
	return newToolMessage(namespace, "pods", toolCall.Id, ToolNameListPods, pods)
}
