package commands

import (
	"context"

	"k8s-assistant/internal/k8s"
	"k8s-assistant/internal/ollama"
)

type ToolName string

const (
	ToolNameListPods         ToolName = "list_pods"
	ToolNameListDeployments  ToolName = "list_deployments"
	ToolNameCreateDeployment ToolName = "create_deployment"
)

type Command interface {
	Name() ToolName
	Definition() ollama.Tools
	Execute(ctx context.Context, client *k8s.Client, namespace string, toolCall ollama.OllamaToolCall) (ollama.OllamaMessage, error)
}
