package handler

import (
	"encoding/json"
	"net/http"

	appsetting "github.com/husari/hube/internal/application/setting"
)

type SettingHandler struct{ svc *appsetting.Service }

func NewSettingHandler(svc *appsetting.Service) *SettingHandler { return &SettingHandler{svc: svc} }

type settingsPayload struct {
	General      generalSettings      `json:"general"`
	Integrations integrationSettings  `json:"integrations"`
}

type generalSettings struct {
	DisplayName string `json:"display_name"`
}

type integrationSettings struct {
	MonkeyAPIURL     string `json:"monkeyapi_url"`
	MonkeyAPIKey     string `json:"monkeyapi_key"`
	MonkeyAPIEnabled bool   `json:"monkeyapi_enabled"`
	PayPingaURL      string `json:"paypinga_url"`
	PayPingaKey      string `json:"paypinga_key"`
	PayPingaEnabled  bool   `json:"paypinga_enabled"`
}

func (h *SettingHandler) Get(w http.ResponseWriter, r *http.Request) {
	all, err := h.svc.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	monkeyURL := all["integration.monkeyapi_url"]
	payURL := all["integration.paypinga_url"]
	writeJSON(w, http.StatusOK, settingsPayload{
		General: generalSettings{
			DisplayName: all["general.display_name"],
		},
		Integrations: integrationSettings{
			MonkeyAPIURL:     monkeyURL,
			MonkeyAPIKey:     all["integration.monkeyapi_key"],
			MonkeyAPIEnabled: monkeyURL != "",
			PayPingaURL:      payURL,
			PayPingaKey:      all["integration.paypinga_key"],
			PayPingaEnabled:  payURL != "",
		},
	})
}

func (h *SettingHandler) Put(w http.ResponseWriter, r *http.Request) {
	var payload settingsPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ctx := r.Context()
	pairs := map[string]string{
		"general.display_name":       payload.General.DisplayName,
		"integration.monkeyapi_url":  payload.Integrations.MonkeyAPIURL,
		"integration.monkeyapi_key":  payload.Integrations.MonkeyAPIKey,
		"integration.paypinga_url":   payload.Integrations.PayPingaURL,
		"integration.paypinga_key":   payload.Integrations.PayPingaKey,
	}
	for k, v := range pairs {
		if err := h.svc.Set(ctx, k, v); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	h.Get(w, r)
}
