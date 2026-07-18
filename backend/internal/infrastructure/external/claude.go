package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	aidom "github.com/husari/hube/internal/domain/ai"
)

const (
	claudeKeyKey = "integration.claude_api_key"
	claudeModel  = anthropic.ModelClaudeSonnet4_6
)

// ClaudeClient lazily builds the Anthropic SDK client the first time it sees
// each API key. New keys (or initial configuration after first boot) take
// effect on the next /ai/chat request without restarting the API.
type ClaudeClient struct {
	settings SettingReader
	client   anthropic.Client
	lastKey  string
}

func NewClaudeClient(settings SettingReader) *ClaudeClient {
	return &ClaudeClient{settings: settings}
}

func (c *ClaudeClient) clientFor(ctx context.Context) (anthropic.Client, error) {
	key, err := c.settings.Get(ctx, claudeKeyKey)
	if err != nil {
		return anthropic.Client{}, fmt.Errorf("read %s: %w", claudeKeyKey, err)
	}
	if key == "" {
		return anthropic.Client{}, ErrNotConfigured
	}
	if c.lastKey != key {
		c.client = anthropic.NewClient(option.WithAPIKey(key))
		c.lastKey = key
	}
	return c.client, nil
}

func toAnthropicTools(defs []aidom.ToolDef) []anthropic.ToolUnionParam {
	out := make([]anthropic.ToolUnionParam, len(defs))
	for i, d := range defs {
		out[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        d.Name,
				Description: anthropic.String(d.Description),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: d.Properties,
					Required:   d.Required,
				},
			},
		}
	}
	return out
}

// writeNDJSON writes one JSON line to the response.
func writeNDJSON(w http.ResponseWriter, event aidom.SSEEvent) {
	b, _ := json.Marshal(event)
	fmt.Fprintf(w, "%s\n", b)
}

// Chat runs the full agentic loop — streaming text to w and executing tools as needed.
// Implements aidom.Provider.
func (c *ClaudeClient) Chat(
	ctx context.Context,
	systemPrompt string,
	history []aidom.ChatMessage,
	executor aidom.ToolExecutor,
	w http.ResponseWriter,
) error {
	client, err := c.clientFor(ctx)
	if err != nil {
		return err
	}

	flusher, _ := w.(http.Flusher)

	messages := make([]anthropic.MessageParam, 0, len(history))
	for _, m := range history {
		block := anthropic.NewTextBlock(m.Content)
		if m.Role == aidom.RoleUser {
			messages = append(messages, anthropic.NewUserMessage(block))
		} else {
			messages = append(messages, anthropic.NewAssistantMessage(block))
		}
	}

	tools := toAnthropicTools(executor.Tools())

	for round := 0; round < 10; round++ {
		stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
			Model:     claudeModel,
			MaxTokens: 4096,
			System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
			Messages:  messages,
			Tools:     tools,
		})

		var acc anthropic.Message
		for stream.Next() {
			event := stream.Current()
			if err := acc.Accumulate(event); err != nil {
				return err
			}
			switch v := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				if d, ok := v.Delta.AsAny().(anthropic.TextDelta); ok {
					writeNDJSON(w, aidom.SSEEvent{Type: "text", Content: d.Text})
					if flusher != nil {
						flusher.Flush()
					}
				}
			}
		}
		if err := stream.Err(); err != nil {
			return err
		}

		// Build assistant message from accumulated content
		var assistantBlocks []anthropic.ContentBlockParamUnion
		for _, block := range acc.Content {
			switch b := block.AsAny().(type) {
			case anthropic.TextBlock:
				assistantBlocks = append(assistantBlocks, anthropic.NewTextBlock(b.Text))
			case anthropic.ToolUseBlock:
				assistantBlocks = append(assistantBlocks, anthropic.NewToolUseBlock(b.ID, b.Input, b.Name))
			}
		}
		messages = append(messages, anthropic.NewAssistantMessage(assistantBlocks...))

		if acc.StopReason != anthropic.StopReasonToolUse {
			break
		}

		// Execute tools and collect results
		var toolResultBlocks []anthropic.ContentBlockParamUnion
		for _, block := range acc.Content {
			tu, ok := block.AsAny().(anthropic.ToolUseBlock)
			if !ok {
				continue
			}

			writeNDJSON(w, aidom.SSEEvent{Type: "tool_use", Tool: tu.Name})
			if flusher != nil {
				flusher.Flush()
			}

			result, err := executor.Execute(ctx, tu.Name, tu.Input)
			if err != nil {
				result = map[string]string{"error": err.Error()}
			}
			resultJSON, _ := json.Marshal(result)

			writeNDJSON(w, aidom.SSEEvent{Type: "tool_result", Tool: tu.Name})
			if flusher != nil {
				flusher.Flush()
			}

			toolResultBlocks = append(toolResultBlocks, anthropic.NewToolResultBlock(tu.ID, string(resultJSON), false))
		}

		messages = append(messages, anthropic.NewUserMessage(toolResultBlocks...))
	}

	writeNDJSON(w, aidom.SSEEvent{Type: "done"})
	if flusher != nil {
		flusher.Flush()
	}
	return nil
}
