package commands

import (
	"encoding/json/v2"
	"fmt"

	"k8s-assistant/internal/ollama"
)

func newToolMessage(namespace, resourceName string, toolName ToolName, data any) (ollama.OllamaMessage, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ollama.OllamaMessage{}, fmt.Errorf("could not marshal %s: %w", resourceName, err)
	}
	return ollama.OllamaMessage{
		Role:     "tool",
		Content:  fmt.Sprintf("%s in namespace %s: %s", resourceName, namespace, jsonData),
		ToolName: string(toolName),
	}, nil
}
