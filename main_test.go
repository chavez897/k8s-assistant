package main

import (
	"encoding/json/v2"
	"testing"
)

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "trims spaces", input: "  hello  ", want: "hello"},
		{name: "rejects empty input", input: "", wantErr: true},
		{name: "rejects whitespace", input: "   ", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := sanitizeInput(test.input, "question")

			if (err != nil) != test.wantErr {
				t.Fatalf("sanitizeInput() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("sanitizeInput() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestOllamaRequestPreservesToolCallFields(t *testing.T) {
	request := OllamaRequest{
		Messages: []OllamaMessage{
			{
				Role: "assistant",
				ToolCalls: []OllamaToolCall{
					{
						Type: "function",
						Function: OllamaToolCallFunction{
							Index: 0,
							Name:  "list_pods",
						},
					},
				},
			},
			{
				Role:     "tool",
				Content:  "pod data",
				ToolName: "list_pods",
			},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var got OllamaRequest
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	toolCall := got.Messages[0].ToolCalls[0]
	if toolCall.Type != "function" {
		t.Errorf("tool call type = %q, want %q", toolCall.Type, "function")
	}
	if toolCall.Function.Index != 0 {
		t.Errorf("tool call function index = %d, want 0", toolCall.Function.Index)
	}
	if got.Messages[1].ToolName != "list_pods" {
		t.Errorf("tool result name = %q, want %q", got.Messages[1].ToolName, "list_pods")
	}
}
