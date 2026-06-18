package ai

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
