package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/husari/hube/internal/domain/cardtracker"
)

type PayPingaClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewPayPingaClient(baseURL, apiKey string) *PayPingaClient {
	return &PayPingaClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *PayPingaClient) GetSummary(ctx context.Context) (*cardtracker.Summary, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/ct-summary", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Hube-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call paypinga: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paypinga returned %d", resp.StatusCode)
	}

	var summary cardtracker.Summary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &summary, nil
}
