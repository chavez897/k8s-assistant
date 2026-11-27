package app

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"

	"k8s-assistant/internal/k8s"
	"k8s-assistant/internal/ollama"
	"k8s-assistant/internal/tools"
)

const maxConcurrentToolCalls = 4

func Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	question, k8sNamespace, err := getUserInput()
	if err != nil {
		return err
	}
	fmt.Printf("Question: %s\n", question)
	request, err := buildInitialOllamaRequest(question, k8sNamespace)
	if err != nil {
		return fmt.Errorf("could not build initial Ollama request: %w", err)
	}
	response, err := agentLoop(ctx, k8sNamespace, request)
	if err != nil {
		return fmt.Errorf("could not execute agent loop: %w", err)
	}
	fmt.Printf("Ollama Response: %s\n", response.Message.Content)
	return nil
}

func agentLoop(ctx context.Context, k8sNamespace string, request ollama.OllamaRequest) (ollama.OllamaResponse, error) {
	response, err := ollama.SendChatRequest(ctx, request)
	if err != nil {
		return ollama.OllamaResponse{}, fmt.Errorf("could not send Ollama request: %w", err)
	}

	var k8sClient *k8s.Client
	if len(response.Message.ToolCalls) > 0 {
		k8sClient, err = k8s.NewClient()
		if err != nil {
			return ollama.OllamaResponse{}, fmt.Errorf("could not create Kubernetes client: %w", err)
		}
	}

	for len(response.Message.ToolCalls) > 0 {
		toolCalls := response.Message.ToolCalls
		request.Messages = append(request.Messages, response.Message)
		toolMessages, err := executeToolCalls(ctx, k8sClient, k8sNamespace, toolCalls)
		if err != nil {
			return ollama.OllamaResponse{}, err
		}
		for _, toolMessage := range toolMessages {
			request.Messages = append(request.Messages, toolMessage)
		}
		response, err = ollama.SendChatRequest(ctx, request)
		if err != nil {
			return ollama.OllamaResponse{}, fmt.Errorf("could not send Ollama request after tool call: %w", err)
		}
	}
	return response, nil
}

func executeToolCalls(ctx context.Context, k8sClient *k8s.Client, k8sNamespace string, toolCalls []ollama.OllamaToolCall) ([]ollama.OllamaMessage, error) {
	toolMessages := make([]ollama.OllamaMessage, len(toolCalls))
	toolErrors := make([]error, len(toolCalls))
	jobs := make(chan toolJob, len(toolCalls))
	fmt.Printf("Executing %d tool calls...\n", len(toolCalls))
	for i, toolCall := range toolCalls {
		jobs <- toolJob{index: i, toolCall: toolCall}
	}
	close(jobs)

	workerCount := min(maxConcurrentToolCalls, len(toolCalls))
	var waitGroup sync.WaitGroup
	waitGroup.Add(workerCount)

	for range workerCount {
		go func() {
			defer waitGroup.Done()

			for job := range jobs {
				toolMessage, err := tools.Dispatch(ctx, k8sClient, k8sNamespace, job.toolCall)
				if err != nil {
					toolErrors[job.index] = fmt.Errorf(
						"could not execute tool %q: %w",
						job.toolCall.Function.Name,
						err,
					)
					continue
				}
				toolMessages[job.index] = toolMessage
			}
		}()
	}

	waitGroup.Wait()
	for _, err := range toolErrors {
		if err != nil {
			return nil, err
		}
	}
	return toolMessages, nil
}

func buildInitialOllamaRequest(question, k8sNamespace string) (ollama.OllamaRequest, error) {
	prompt, err := ollama.BuildPrompt(k8sNamespace, question)
	if err != nil {
		return ollama.OllamaRequest{}, fmt.Errorf("could not build prompt: %w", err)
	}
	request := ollama.BuildOllamaRequest(prompt)
	request.Tools = tools.Definitions()
	return request, nil
}

func getUserInput() (string, string, error) {
	k8sNamespacePointer := flag.String("namespace", "default", "Kubernetes namespace")
	flag.Parse()
	question, err := sanitizeInput(strings.Join(flag.Args(), " "), "question")
	if err != nil {
		return "", "", err
	}
	k8sNamespace, err := sanitizeInput(*k8sNamespacePointer, "namespace")
	if err != nil {
		return "", "", err
	}
	return question, k8sNamespace, nil
}

func sanitizeInput(input, inputName string) (string, error) {
	sanitizedInput := strings.TrimSpace(input)
	if sanitizedInput == "" {
		return "", fmt.Errorf("empty %s provided", inputName)
	}
	return sanitizedInput, nil
}

type toolJob struct {
	index    int
	toolCall ollama.OllamaToolCall
}
