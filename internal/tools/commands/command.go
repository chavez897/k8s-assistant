package commands

import (
	"context"

	"k8s-assistant/internal/k8s"
	"k8s-assistant/internal/ollama"
)

type ToolName string

const (
	ToolNameListPods        ToolName = "list_pods"
	ToolNameListDeployments ToolName = "list_deployments"
)

type Command interface {
	Name() ToolName
	Definition() ollama.Tools
	Execute(ctx context.Context, client *k8s.Client, namespace string) (ollama.OllamaMessage, error)
}
