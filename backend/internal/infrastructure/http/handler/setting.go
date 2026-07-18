package handler

import (
	"encoding/json"
	"net/http"

	appsetting "github.com/husari/hube/internal/application/setting"
)

type SettingHandler struct{ svc *appsetting.Service }

func NewSettingHandler(svc *appsetting.Service) *SettingHandler { return &SettingHandler{svc: svc} }

// settingsResponse is the masked settings shape returned by GET and PUT.
// Integration API keys are write-only: the response only reports whether a
// non-empty key is stored, never the key value.
type settingsResponse struct {
	General      generalSettings             `json:"general"`
	Integrations integrationSettingsResponse `json:"integrations"`
}

type generalSettings struct {
	DisplayName     string `json:"display_name"`
	ViewPreferences string `json:"view_preferences,omitempty"`
}

type integrationSettingsResponse struct {
	MonkeyAPIURL     string `json:"monkeyapi_url"`
	MonkeyAPIKeySet  bool   `json:"monkeyapi_key_set"`
	MonkeyAPIEnabled bool   `json:"monkeyapi_enabled"`
	PayPingaURL      string `json:"paypinga_url"`
	PayPingaKeySet   bool   `json:"paypinga_key_set"`
	PayPingaEnabled  bool   `json:"paypinga_enabled"`
}

// settingsRequest is the accepted PUT shape. Keys are optional: an empty key
// value means "leave the stored key unchanged".
type settingsRequest struct {
	General      generalSettings            `json:"general"`
	Integrations integrationSettingsRequest `json:"integrations"`
}

type integrationSettingsRequest struct {
	MonkeyAPIURL string `json:"monkeyapi_url"`
	MonkeyAPIKey string `json:"monkeyapi_key"`
	PayPingaURL  string `json:"paypinga_url"`
	PayPingaKey  string `json:"paypinga_key"`
}

func (h *SettingHandler) Get(w http.ResponseWriter, r *http.Request) {
	all, err := h.svc.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	monkeyURL := all["integration.monkeyapi_url"]
	payURL := all["integration.paypinga_url"]
	// GetAll masks sensitive values with a non-empty sentinel, so a non-empty
	// value here means a key is stored — without exposing the key itself.
	writeJSON(w, http.StatusOK, settingsResponse{
		General: generalSettings{
			DisplayName:     all["general.display_name"],
			ViewPreferences: all["general.view_preferences"],
		},
		Integrations: integrationSettingsResponse{
			MonkeyAPIURL:     monkeyURL,
			MonkeyAPIKeySet:  all["integration.monkeyapi_key"] != "",
			MonkeyAPIEnabled: monkeyURL != "",
			PayPingaURL:      payURL,
			PayPingaKeySet:   all["integration.paypinga_key"] != "",
			PayPingaEnabled:  payURL != "",
		},
	})
}

func (h *SettingHandler) Put(w http.ResponseWriter, r *http.Request) {
	var payload settingsRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ctx := r.Context()
	pairs := map[string]string{
		"general.display_name":      payload.General.DisplayName,
		"general.view_preferences":  payload.General.ViewPreferences,
		"integration.monkeyapi_url": payload.Integrations.MonkeyAPIURL,
		"integration.paypinga_url":  payload.Integrations.PayPingaURL,
	}
	// API keys are write-only: only overwrite the stored key when the incoming
	// value is non-empty (empty string = "leave unchanged").
	if payload.Integrations.MonkeyAPIKey != "" {
		pairs["integration.monkeyapi_key"] = payload.Integrations.MonkeyAPIKey
	}
	if payload.Integrations.PayPingaKey != "" {
		pairs["integration.paypinga_key"] = payload.Integrations.PayPingaKey
	}
	for k, v := range pairs {
		if err := h.svc.Set(ctx, k, v); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	h.Get(w, r)
}
