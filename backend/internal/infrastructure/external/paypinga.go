package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/husari/hube/internal/domain/cardtracker"
)

const (
	payPingaURLKey = "integration.paypinga_url"
	payPingaKeyKey = "integration.paypinga_key"
)

type PayPingaClient struct {
	settings SettingReader
	http     *http.Client
}

func NewPayPingaClient(settings SettingReader) *PayPingaClient {
	return &PayPingaClient{
		settings: settings,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *PayPingaClient) GetSummary(ctx context.Context) (*cardtracker.Summary, error) {
	baseURL, err := c.settings.Get(ctx, payPingaURLKey)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", payPingaURLKey, err)
	}
	apiKey, err := c.settings.Get(ctx, payPingaKeyKey)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", payPingaKeyKey, err)
	}
	if baseURL == "" || apiKey == "" {
		return nil, ErrNotConfigured
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/ct-summary", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Hube-Key", apiKey)

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
