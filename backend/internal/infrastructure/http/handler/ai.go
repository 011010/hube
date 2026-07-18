package handler

import (
	"context"
	"encoding/json"
	"net/http"

	appai "github.com/husari/hube/internal/application/ai"
	appsetting "github.com/husari/hube/internal/application/setting"
	"github.com/husari/hube/internal/domain/ai"
)

const (
	aiProviderAnthropic = "anthropic"
	aiProviderOpenAI    = "openai"
	aiNoProviderMsg     = "No AI provider configured — set ANTHROPIC_API_KEY or OPENAI_API_KEY"
)

type AIHandler struct {
	settings  *appsetting.Service
	anthropic ai.Provider
	openai    ai.Provider
	executor  *appai.HubExecutor
}

func NewAIHandler(settings *appsetting.Service, anthropic, openai ai.Provider, executor *appai.HubExecutor) *AIHandler {
	return &AIHandler{settings: settings, anthropic: anthropic, openai: openai, executor: executor}
}

type chatRequest struct {
	Messages []ai.ChatMessage `json:"messages"`
	Provider string           `json:"provider"` // "anthropic" | "openai" | "" (auto)
}

func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	provider, ok := h.resolve(r.Context(), req.Provider)
	if !ok {
		http.Error(w, aiNoProviderMsg, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	systemPrompt := appai.BuildSystemPrompt()
	if err := provider.Chat(r.Context(), systemPrompt, req.Messages, h.executor, w); err != nil {
		b, _ := json.Marshal(ai.SSEEvent{Type: "error", Error: err.Error()})
		w.Write(b) //nolint:errcheck
		w.Write([]byte("\n"))
	}
}

// resolve picks a provider whose API key is present in settings. For an
// explicit provider request, the matching key must be set. For "auto",
// anthropic is preferred; openai is the fallback. The bool reports whether
// any usable provider was found — the caller returns 503 when it is false.
func (h *AIHandler) resolve(ctx context.Context, requested string) (ai.Provider, bool) {
	claudeConfigured := h.hasKey(ctx, "integration.claude_api_key")
	openaiConfigured := h.hasKey(ctx, "integration.openai_api_key")

	switch requested {
	case aiProviderAnthropic:
		if claudeConfigured {
			return h.anthropic, true
		}
		return nil, false
	case aiProviderOpenAI:
		if openaiConfigured {
			return h.openai, true
		}
		return nil, false
	default:
		if claudeConfigured {
			return h.anthropic, true
		}
		if openaiConfigured {
			return h.openai, true
		}
		return nil, false
	}
}

func (h *AIHandler) hasKey(ctx context.Context, key string) bool {
	v, err := h.settings.Get(ctx, key)
	if err != nil {
		return false
	}
	return v != ""
}
