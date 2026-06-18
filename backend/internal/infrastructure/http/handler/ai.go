package handler

import (
	"encoding/json"
	"net/http"

	appai "github.com/husari/hube/internal/application/ai"
	aidom "github.com/husari/hube/internal/domain/ai"
)

type AIHandler struct {
	anthropic aidom.Provider
	openai    aidom.Provider
	executor  *appai.HubExecutor
}

func NewAIHandler(anthropic, openai aidom.Provider, executor *appai.HubExecutor) *AIHandler {
	return &AIHandler{anthropic, openai, executor}
}

type chatRequest struct {
	Messages []aidom.ChatMessage `json:"messages"`
	Provider string              `json:"provider"` // "anthropic" | "openai" | "" (auto)
}

func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	provider := h.resolve(req.Provider)
	if provider == nil {
		http.Error(w, "No AI provider configured — set ANTHROPIC_API_KEY or OPENAI_API_KEY", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	systemPrompt := appai.BuildSystemPrompt()
	if err := provider.Chat(r.Context(), systemPrompt, req.Messages, h.executor, w); err != nil {
		b, _ := json.Marshal(aidom.SSEEvent{Type: "error", Error: err.Error()})
		w.Write(b) //nolint:errcheck
		w.Write([]byte("\n"))
	}
}

// resolve picks the provider: explicit request → fallback to configured one.
func (h *AIHandler) resolve(requested string) aidom.Provider {
	switch requested {
	case "anthropic":
		return h.anthropic
	case "openai":
		return h.openai
	default:
		// Auto: prefer Anthropic, fall back to OpenAI
		if h.anthropic != nil {
			return h.anthropic
		}
		return h.openai
	}
}
