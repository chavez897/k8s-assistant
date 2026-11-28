package tools

import (
	"context"
	"fmt"

	"k8s-assistant/internal/tools/commands"

	"k8s-assistant/internal/k8s"
	"k8s-assistant/internal/ollama"
)

var commandRegistry = []commands.Command{
	commands.ListPodsCommand{},
	commands.ListDeploymentsCommand{},
	commands.CreateDeploymentCommand{},
}

func Definitions() []ollama.Tools {
	definitions := make([]ollama.Tools, 0, len(commandRegistry))
	for _, command := range commandRegistry {
		definitions = append(definitions, command.Definition())
	}
	return definitions
}

func Dispatch(ctx context.Context, client *k8s.Client, namespace string, toolCall ollama.OllamaToolCall) (ollama.OllamaMessage, error) {
	for _, command := range commandRegistry {
		if command.Name() == commands.ToolName(toolCall.Function.Name) {
			return command.Execute(ctx, client, namespace, toolCall)
		}
	}
	return ollama.OllamaMessage{}, fmt.Errorf("unknown tool call: %s", toolCall.Function.Name)
}
