package ollama

import "time"

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
