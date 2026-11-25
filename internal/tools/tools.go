package tools

import (
	"context"
	"fmt"

	"k8s-assistant/internal/k8s"
	"k8s-assistant/internal/ollama"
)

// Definitions returns the tools exposed to the Ollama model.
func Definitions() []ollama.Tools {
	return []ollama.Tools{
		{
			Type: "function",
			Function: ollama.Function{
				Name:        "list_pods",
				Description: "Read-only: list pod status in the namespace selected by the CLI.",
				Parameters: ollama.FunctionParameters{
					Type:       "object",
					Properties: map[string]any{},
					Required:   []string{},
				},
			},
		},
	}
}

// Dispatch executes an allowlisted tool call and returns its Ollama tool message.
func Dispatch(ctx context.Context, namespace string, call ollama.OllamaToolCall) (ollama.OllamaMessage, error) {
	switch call.Function.Name {
	case "list_pods":
		pods, err := k8s.GetK8sPods(ctx, namespace)
		if err != nil {
			return ollama.OllamaMessage{}, fmt.Errorf("could not get Kubernetes pods: %w", err)
		}
		return ollama.OllamaMessage{
			Role:     "tool",
			Content:  fmt.Sprintf("Pods in namespace %s: %+v", namespace, pods),
			ToolName: call.Function.Name,
		}, nil
	default:
		return ollama.OllamaMessage{}, fmt.Errorf("unknown tool call: %s", call.Function.Name)
	}
}
