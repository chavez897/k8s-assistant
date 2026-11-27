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
	prompt, err := ollama.BuildPrompt(k8sNamespace, question)
	if err != nil {
		return fmt.Errorf("could not build prompt: %w", err)
	}
	request := ollama.BuildOllamaRequest(prompt)
	request.Tools = tools.Definitions()
	response, err := ollama.SendChatRequest(request)
	if err != nil {
		return fmt.Errorf("could not send Ollama request: %w", err)
	}
	var k8sClient *k8s.Client
	if len(response.Message.ToolCalls) > 0 {
		k8sClient, err = k8s.NewClient()
		if err != nil {
			return fmt.Errorf("could not create Kubernetes client: %w", err)
		}
	}

	for len(response.Message.ToolCalls) > 0 {
		toolCalls := response.Message.ToolCalls
		toolMessages := make([]ollama.OllamaMessage, len(toolCalls))
		toolErrors := make([]error, len(toolCalls))
		request.Messages = append(request.Messages, response.Message)
		jobs := make(chan toolJob, len(toolCalls))
		fmt.Printf("Executing %d tool calls...\n", len(toolCalls))
		for i, toolCall := range toolCalls {
			jobs <- toolJob{i, toolCall}
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
				return err
			}
		}
		for _, toolMessage := range toolMessages {
			request.Messages = append(request.Messages, toolMessage)
		}
		response, err = ollama.SendChatRequest(request)
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

type toolJob struct {
	index    int
	toolCall ollama.OllamaToolCall
}
