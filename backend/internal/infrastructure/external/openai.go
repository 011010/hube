package external

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	aidom "github.com/husari/hube/internal/domain/ai"
)

const (
	openAIKeyKey       = "integration.openai_api_key"
	openAIBaseURLKey   = "integration.openai_base_url"
	openAIModelKey     = "integration.openai_model"
	defaultOpenAIBase  = "https://api.openai.com/v1"
	defaultOpenAIModel = "gpt-4o"
	// maxSSEBytes caps a single SSE line we'll accept; tool call argument
	// dumps can grow large but staying under 1 MiB keeps the parser bounded.
	maxSSEBytes = 1 << 20
)

var httpClient = &http.Client{Timeout: 60 * time.Second}

type OpenAIClient struct {
	settings SettingReader
}

func NewOpenAIClient(settings SettingReader) *OpenAIClient {
	return &OpenAIClient{settings: settings}
}

type openAIConfig struct {
	apiKey  string
	baseURL string
	model   string
}

func (c *OpenAIClient) configFor(ctx context.Context) (openAIConfig, error) {
	apiKey, err := c.settings.Get(ctx, openAIKeyKey)
	if err != nil {
		return openAIConfig{}, fmt.Errorf("read %s: %w", openAIKeyKey, err)
	}
	if apiKey == "" {
		return openAIConfig{}, ErrNotConfigured
	}
	baseURL, err := c.settings.Get(ctx, openAIBaseURLKey)
	if err != nil {
		return openAIConfig{}, fmt.Errorf("read %s: %w", openAIBaseURLKey, err)
	}
	if baseURL == "" {
		baseURL = defaultOpenAIBase
	}
	baseURL = strings.TrimRight(baseURL, "/")

	model, err := c.settings.Get(ctx, openAIModelKey)
	if err != nil {
		return openAIConfig{}, fmt.Errorf("read %s: %w", openAIModelKey, err)
	}
	if model == "" {
		model = defaultOpenAIModel
	}
	return openAIConfig{apiKey: apiKey, baseURL: baseURL, model: model}, nil
}

// --- wire types for OpenAI chat completions API ---

type oaiMessage struct {
	Role       string        `json:"role"`
	Content    string        `json:"content,omitempty"`
	ToolCalls  []oaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	Name       string        `json:"name,omitempty"`
}

type oaiToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type oaiTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  struct {
			Type       string         `json:"type"`
			Properties map[string]any `json:"properties"`
			Required   []string       `json:"required,omitempty"`
		} `json:"parameters"`
	} `json:"function"`
}

type oaiRequest struct {
	Model    string       `json:"model"`
	Messages []oaiMessage `json:"messages"`
	Tools    []oaiTool    `json:"tools,omitempty"`
	Stream   bool         `json:"stream"`
}

// Streaming chunk types
type pendingCall struct {
	id        string
	name      string
	arguments string
}

type oaiChunk struct {
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func toOAITools(defs []aidom.ToolDef) []oaiTool {
	out := make([]oaiTool, len(defs))
	for i, d := range defs {
		out[i].Type = "function"
		out[i].Function.Name = d.Name
		out[i].Function.Description = d.Description
		out[i].Function.Parameters.Type = "object"
		out[i].Function.Parameters.Properties = d.Properties
		out[i].Function.Parameters.Required = d.Required
	}
	return out
}

// Chat implements aidom.Provider.
func (c *OpenAIClient) Chat(
	ctx context.Context,
	systemPrompt string,
	history []aidom.ChatMessage,
	executor aidom.ToolExecutor,
	w http.ResponseWriter,
) error {
	cfg, err := c.configFor(ctx)
	if err != nil {
		return err
	}

	flusher, _ := w.(http.Flusher)

	messages := []oaiMessage{{Role: "system", Content: systemPrompt}}
	for _, m := range history {
		role := "user"
		if m.Role == aidom.RoleAssistant {
			role = "assistant"
		}
		messages = append(messages, oaiMessage{Role: role, Content: m.Content})
	}

	tools := toOAITools(executor.Tools())

	for round := 0; round < 10; round++ {
		req := oaiRequest{
			Model:    cfg.model,
			Messages: messages,
			Tools:    tools,
			Stream:   true,
		}
		body, _ := json.Marshal(req)

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
			cfg.baseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			return err
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+cfg.apiKey)

		resp, err := httpClient.Do(httpReq)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			var errBody map[string]any
			json.NewDecoder(resp.Body).Decode(&errBody) //nolint:errcheck
			resp.Body.Close()
			return fmt.Errorf("API error %d: %v", resp.StatusCode, errBody)
		}

		// Accumulate the streamed response
		var textBuilder strings.Builder
		pending := map[int]*pendingCall{}
		var finishReason string

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), maxSSEBytes)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk oaiChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) == 0 {
				continue
			}
			choice := chunk.Choices[0]

			if choice.FinishReason != "" {
				finishReason = choice.FinishReason
			}

			// Text delta
			if choice.Delta.Content != "" {
				textBuilder.WriteString(choice.Delta.Content)
				writeNDJSON(w, aidom.SSEEvent{Type: "text", Content: choice.Delta.Content})
				if flusher != nil {
					flusher.Flush()
				}
			}

			// Tool call deltas
			for _, tc := range choice.Delta.ToolCalls {
				if _, ok := pending[tc.Index]; !ok {
					pending[tc.Index] = &pendingCall{}
				}
				p := pending[tc.Index]
				if tc.ID != "" {
					p.id = tc.ID
				}
				if tc.Function.Name != "" {
					p.name = tc.Function.Name
				}
				p.arguments += tc.Function.Arguments
			}
		}
		scanErr := scanner.Err()
		resp.Body.Close()
		if scanErr != nil {
			return fmt.Errorf("read stream: %w", scanErr)
		}

		if finishReason != "tool_calls" || len(pending) == 0 {
			break
		}

		// Build the assistant message with tool_calls for history. Iterate
		// indices in the order OpenAI sent them so the conversation stays
		// consistent if a future implementation drops the contiguous-index
		// guarantee.
		var toolCalls []oaiToolCall
		for _, idx := range sortedPendingKeys(pending) {
			p := pending[idx]
			toolCalls = append(toolCalls, oaiToolCall{
				ID:   p.id,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{Name: p.name, Arguments: p.arguments},
			})
		}
		messages = append(messages, oaiMessage{
			Role:      "assistant",
			ToolCalls: toolCalls,
		})

		// Execute tools
		for _, tc := range toolCalls {
			writeNDJSON(w, aidom.SSEEvent{Type: "tool_use", Tool: tc.Function.Name})
			if flusher != nil {
				flusher.Flush()
			}

			var inputJSON json.RawMessage
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &inputJSON); err != nil {
				inputJSON = json.RawMessage(`{}`)
			}

			result, err := executor.Execute(ctx, tc.Function.Name, inputJSON)
			if err != nil {
				result = map[string]string{"error": err.Error()}
			}
			resultJSON, _ := json.Marshal(result)

			writeNDJSON(w, aidom.SSEEvent{Type: "tool_result", Tool: tc.Function.Name})
			if flusher != nil {
				flusher.Flush()
			}

			messages = append(messages, oaiMessage{
				Role:       "tool",
				Content:    string(resultJSON),
				ToolCallID: tc.ID,
			})
		}
	}

	writeNDJSON(w, aidom.SSEEvent{Type: "done"})
	if flusher != nil {
		flusher.Flush()
	}
	return nil
}

func sortedPendingKeys(m map[int]*pendingCall) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
