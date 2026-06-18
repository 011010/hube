package handler

import (
	"encoding/json"
	"net/http"

	appai "github.com/husari/hube/internal/application/ai"
	aidom "github.com/husari/hube/internal/domain/ai"
	"github.com/husari/hube/internal/infrastructure/external"
)

type AIHandler struct {
	claude   *external.ClaudeClient
	executor *appai.HubExecutor
}

func NewAIHandler(claude *external.ClaudeClient, executor *appai.HubExecutor) *AIHandler {
	return &AIHandler{claude, executor}
}

type chatRequest struct {
	Messages []aidom.ChatMessage `json:"messages"`
}

func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	if h.claude == nil {
		http.Error(w, "Claude API key not configured — set ANTHROPIC_API_KEY", http.StatusServiceUnavailable)
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	systemPrompt := appai.BuildSystemPrompt()
	if err := h.claude.Chat(r.Context(), systemPrompt, req.Messages, h.executor, w); err != nil {
		// Stream the error as a final event (headers already sent)
		b, _ := json.Marshal(aidom.SSEEvent{Type: "error", Error: err.Error()})
		w.Write(b) //nolint:errcheck
		w.Write([]byte("\n"))
	}
}
