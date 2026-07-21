package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	openAIEmbedModelKey   = "integration.openai_embedding_model"
	defaultEmbeddingModel = "text-embedding-3-small"
)

type EmbeddingsClient struct {
	settings SettingReader
}

func NewEmbeddingsClient(settings SettingReader) *EmbeddingsClient {
	return &EmbeddingsClient{settings: settings}
}

type embeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (c *EmbeddingsClient) Embed(ctx context.Context, text string) ([]float32, error) {
	apiKey, err := c.settings.Get(ctx, openAIKeyKey)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", openAIKeyKey, err)
	}
	if apiKey == "" {
		return nil, ErrNotConfigured
	}

	baseURL, err := c.settings.Get(ctx, openAIBaseURLKey)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", openAIBaseURLKey, err)
	}
	if baseURL == "" {
		baseURL = defaultOpenAIBase
	}
	baseURL = strings.TrimRight(baseURL, "/")

	model, err := c.settings.Get(ctx, openAIEmbedModelKey)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", openAIEmbedModelKey, err)
	}
	if model == "" {
		model = defaultEmbeddingModel
	}

	body, _ := json.Marshal(embeddingRequest{Input: text, Model: model})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embeddings request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embeddings API returned %d", resp.StatusCode)
	}

	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embedding: %w", err)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return result.Data[0].Embedding, nil
}
