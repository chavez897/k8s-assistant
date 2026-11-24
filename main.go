package main

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const OLLAMA_BASE_URL = "http://localhost:11434"
const OLLAMA_MODEL = "gemma4:e2b"
const OLLAMA_CHAT_ENDPOINT = "/api/chat"

const SYSTEM_PROMPT = `Answer questions about Kubernetes. For current pod information, use the provided tool. 
Do not guess if the tool does not provide enough data.`

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	k8sNamespacePointer := flag.String("namespace", "default", "Kubernetes namespace")
	flag.Parse()
	question, err := sanitizeInput(strings.Join(flag.Args(), " "), "question")
	if err != nil {
		return err
	}
	k8sNamespace, err := sanitizeInput(*k8sNamespacePointer, "namespace")
	if err != nil {
		return err
	}
	fmt.Printf("Question: %s\n", question)
	prompt, err := buildPrompt(k8sNamespace, question)
	if err != nil {
		return fmt.Errorf("could not build prompt: %w", err)
	}
	request := buildOllamaRequest(prompt)
	response, err := ollamaRequest(request)
	if err != nil {
		return fmt.Errorf("could not send Ollama request: %w", err)
	}
	if len(response.Message.ToolCalls) > 0 {
		request.Messages = append(request.Messages, response.Message)
		for _, toolCall := range response.Message.ToolCalls {
			if toolCall.Function.Name == "list_pods" {
				pods, err := getK8sPods(ctx, k8sNamespace)
				if err != nil {
					return fmt.Errorf("could not get Kubernetes pods: %w", err)
				}
				toolMessage := OllamaMessage{
					Role:     "tool",
					Content:  fmt.Sprintf("Pods in namespace %s: %+v", k8sNamespace, pods),
					ToolName: toolCall.Function.Name,
				}
				request.Messages = append(request.Messages, toolMessage)
			} else {
				return fmt.Errorf("unknown tool call: %s", toolCall.Function.Name)
			}
		}
		response, err = ollamaRequest(request)
		if err != nil {
			return fmt.Errorf("could not send Ollama request after tool call: %w", err)
		}
	}
	fmt.Printf("Ollama Response: %s\n", response.Message.Content)
	return nil
}

func sanitizeInput(input, inputName string) (string, error) {
	sanitizedInput := strings.TrimSpace(input)
	if sanitizedInput == "" {
		return "", fmt.Errorf("empty %s provided", inputName)
	}
	return sanitizedInput, nil
}

func getK8sPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	fmt.Printf("Fetching pods in namespace: %s\n", namespace)
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes config: %w", err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create Kubernetes client: %w", err)
	}
	pods, err := clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not list Kubernetes pods: %w", err)
	}
	var result []PodInfo
	for _, pod := range pods.Items {
		if len(pod.Status.ContainerStatuses) == 0 {
			result = append(result, PodInfo{
				Name:          pod.Name,
				CurrentStatus: string(pod.Status.Phase),
			})
		}
		for _, containerStatus := range pod.Status.ContainerStatuses {
			podInfo := PodInfo{
				Name:          pod.Name,
				CurrentStatus: string(pod.Status.Phase),
				ContainerName: containerStatus.Name,
				RestartCount:  containerStatus.RestartCount,
			}
			if containerStatus.State.Terminated != nil {
				podInfo.Reason = containerStatus.State.Terminated.Reason
			} else if containerStatus.State.Waiting != nil {
				podInfo.Reason = containerStatus.State.Waiting.Reason
			}
			result = append(result, podInfo)
		}
	}
	return result, nil
}

func buildPrompt(namespace, question string) (string, error) {
	prompt := `
	Namespace: ` + namespace + `
	Question: ` + question
	return prompt, nil
}

func buildOllamaRequest(question string) OllamaRequest {
	return OllamaRequest{
		Model:  OLLAMA_MODEL,
		Stream: false,
		Messages: []OllamaMessage{
			{
				Role:    "system",
				Content: SYSTEM_PROMPT,
			},
			{
				Role:    "user",
				Content: question,
			},
		},
		Tools: buildTools(),
	}
}

func ollamaRequest(request OllamaRequest) (OllamaResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return OllamaResponse{}, fmt.Errorf("could not marshal request body: %w", err)
	}
	requestUrl := OLLAMA_BASE_URL + OLLAMA_CHAT_ENDPOINT
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

func buildTools() []Tools {
	tools := []Tools{
		{
			Type: "function",
			Function: Function{
				Name:        "list_pods",
				Description: "Read-only: list pod status in the namespace selected by the CLI.",
				Parameters: FunctionParameters{
					Type:       "object",
					Properties: map[string]any{},
					Required:   []string{},
				},
			},
		},
	}
	return tools
}

type OllamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []OllamaToolCall `json:"tool_calls"`
	ToolName  string           `json:"tool_name,omitempty"`
}
type OllamaRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Tools    []Tools         `json:"tools"`
}
type PodInfo struct {
	Name          string `json:"name"`
	CurrentStatus string `json:"currentStatus"`
	ContainerName string `json:"containerName"`
	RestartCount  int32  `json:"restartCount"`
	Reason        string `json:"reason"`
}

type OllamaResponse struct {
	Model                 string        `json:"model"`
	CreatedAt             time.Time     `json:"created_at"`
	Message               OllamaMessage `json:"message"`
	Done                  bool          `json:"done"`
	TotalDuration         int64         `json:"total_duration"`
	LoadDuration          int64         `json:"load_duration"`
	PromptEvalCount       int64         `json:"prompt_eval_count"`
	PromptEvalCachedCount int64         `json:"prompt_eval_cached_count"`
	PromptEvalDuration    int64         `json:"prompt_eval_duration"`
	EvalCount             int64         `json:"eval_count"`
	EvalDuration          int64         `json:"eval_duration"`
}

type OllamaToolCall struct {
	Id       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function OllamaToolCallFunction `json:"function"`
}

type OllamaToolCallFunction struct {
	Index     int            `json:"index"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type Tools struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Parameters  FunctionParameters `json:"parameters"`
}

type FunctionParameters struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	Required   []string       `json:"required"`
}
