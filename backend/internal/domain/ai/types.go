package ai

import (
	"context"
	"encoding/json"
	"net/http"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type ChatMessage struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// SSEEvent is written to the streaming response as newline-delimited JSON.
type SSEEvent struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Tool    string `json:"tool,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ToolDef is the provider-agnostic definition of a callable tool.
type ToolDef struct {
	Name        string
	Description string
	Properties  map[string]any
	Required    []string
}

// ToolExecutor executes named tools and provides their definitions.
type ToolExecutor interface {
	Execute(ctx context.Context, name string, input json.RawMessage) (any, error)
	Tools() []ToolDef
}

// Provider is implemented by any LLM backend (Anthropic, OpenAI-compatible, …).
type Provider interface {
	Chat(ctx context.Context, systemPrompt string, history []ChatMessage, executor ToolExecutor, w http.ResponseWriter) error
}
