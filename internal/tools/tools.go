package tools

import (
	"context"
	"encoding/json/v2"
	"fmt"

	"k8s-assistant/internal/k8s"
	"k8s-assistant/internal/ollama"
)

type ToolName string

const (
	ToolNameListPods        ToolName = "list_pods"
	ToolNameListDeployments ToolName = "list_deployments"
)

func Definitions() []ollama.Tools {
	return []ollama.Tools{
		{
			Type: "function",
			Function: ollama.Function{
				Name:        string(ToolNameListPods),
				Description: "Read-only: list pod status in the namespace selected by the CLI.",
				Parameters: ollama.FunctionParameters{
					Type:       "object",
					Properties: map[string]any{},
					Required:   []string{},
				},
			},
		},
		{
			Type: "function",
			Function: ollama.Function{
				Name:        string(ToolNameListDeployments),
				Description: "Read-only: list deployment status in the namespace selected by the CLI.",
				Parameters: ollama.FunctionParameters{
					Type:       "object",
					Properties: map[string]any{},
					Required:   []string{},
				},
			},
		},
	}
}

func Dispatch(ctx context.Context, client *k8s.Client, namespace string, call ollama.OllamaToolCall) (ollama.OllamaMessage, error) {
	switch call.Function.Name {
	case string(ToolNameListPods):
		pods, err := client.GetPods(ctx, namespace)
		if err != nil {
			return ollama.OllamaMessage{}, fmt.Errorf("could not get Kubernetes pods: %w", err)
		}
		jsonData, err := json.Marshal(pods)
		if err != nil {
			return ollama.OllamaMessage{}, fmt.Errorf("could not marshal pods: %w", err)
		}
		return ollama.OllamaMessage{
			Role:     "tool",
			Content:  fmt.Sprintf("Pods in namespace %s: %s", namespace, jsonData),
			ToolName: call.Function.Name,
		}, nil
	case string(ToolNameListDeployments):
		deployments, err := client.GetDeployments(ctx, namespace)
		if err != nil {
			return ollama.OllamaMessage{}, fmt.Errorf("could not get Kubernetes deployments: %w", err)
		}
		jsonData, err := json.Marshal(deployments)
		if err != nil {
			return ollama.OllamaMessage{}, fmt.Errorf("could not marshal deployments: %w", err)
		}
		return ollama.OllamaMessage{
			Role:     "tool",
			Content:  fmt.Sprintf("Deployments in namespace %s: %s", namespace, jsonData),
			ToolName: call.Function.Name,
		}, nil
	default:
		return ollama.OllamaMessage{}, fmt.Errorf("unknown tool call: %s", call.Function.Name)
	}
}
