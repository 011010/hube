package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/husari/hube/internal/domain/finance"
)

const (
	monkeyURLKey = "integration.monkeyapi_url"
	monkeyKeyKey = "integration.monkeyapi_key"
)

type MoneyMonkeyClient struct {
	settings SettingReader
	http     *http.Client
}

func NewMoneyMonkeyClient(settings SettingReader) *MoneyMonkeyClient {
	return &MoneyMonkeyClient{
		settings: settings,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *MoneyMonkeyClient) GetSummary(ctx context.Context) (*finance.Summary, error) {
	baseURL, err := c.settings.Get(ctx, monkeyURLKey)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", monkeyURLKey, err)
	}
	apiKey, err := c.settings.Get(ctx, monkeyKeyKey)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", monkeyKeyKey, err)
	}
	if baseURL == "" || apiKey == "" {
		return nil, ErrNotConfigured
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/finance-summary", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Hube-Key", apiKey)

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
