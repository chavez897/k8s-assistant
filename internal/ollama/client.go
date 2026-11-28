package ollama

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"time"
)

const ollamaBaseURL = "http://localhost:11434"
const ollamaModel = "gemma4:e2b"
const ollamaChatEndpoint = "/api/chat"

const systemPrompt = `Answer questions about Kubernetes. For current pod information, use the provided tool.
Do not guess if the tool does not provide enough data.`

var httpClient = &http.Client{
	Timeout: 2 * time.Minute,
}

func SendChatRequest(ctx context.Context, request OllamaRequest) (OllamaResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return OllamaResponse{}, fmt.Errorf("could not marshal request body: %w", err)
	}
	requestUrl := ollamaBaseURL + ollamaChatEndpoint
	req, err := http.NewRequestWithContext(ctx, "POST", requestUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return OllamaResponse{}, fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
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
