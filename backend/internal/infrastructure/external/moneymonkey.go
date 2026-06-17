package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/husari/hube/internal/domain/finance"
)

type MoneyMonkeyClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewMoneyMonkeyClient(baseURL, apiKey string) *MoneyMonkeyClient {
	return &MoneyMonkeyClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *MoneyMonkeyClient) GetSummary(ctx context.Context) (*finance.Summary, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/finance-summary", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Hube-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call money monkey: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("money monkey returned %d", resp.StatusCode)
	}

	var summary finance.Summary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &summary, nil
}
