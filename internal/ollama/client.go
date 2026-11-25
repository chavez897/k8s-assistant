package ollama

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
)

const ollamaBaseURL = "http://localhost:11434"
const ollamaModel = "gemma4:e2b"
const ollamaChatEndpoint = "/api/chat"

const systemPrompt = `Answer questions about Kubernetes. For current pod information, use the provided tool.
Do not guess if the tool does not provide enough data.`

func SendChatRequest(request OllamaRequest) (OllamaResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return OllamaResponse{}, fmt.Errorf("could not marshal request body: %w", err)
	}
	requestUrl := ollamaBaseURL + ollamaChatEndpoint
	resp, err := http.Post(requestUrl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return OllamaResponse{}, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return OllamaResponse{}, fmt.Errorf("received non-OK response: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return OllamaResponse{}, fmt.Errorf("could not read response body: %w", err)
	}
	response := OllamaResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return OllamaResponse{}, fmt.Errorf("could not unmarshal response: %w", err)
	}
	return response, nil
}

func BuildOllamaRequest(question string) OllamaRequest {
	return OllamaRequest{
		Model:  ollamaModel,
		Stream: false,
		Messages: []OllamaMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: question,
			},
		},
	}
}

func BuildPrompt(namespace, question string) (string, error) {
	prompt := `
	Namespace: ` + namespace + `
	Question: ` + question
	return prompt, nil
}
