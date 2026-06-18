package handler

import (
	"net/http"

	"github.com/husari/hube/internal/infrastructure/external"
)

type CardTrackerHandler struct{ client *external.PayPingaClient }

func NewCardTrackerHandler(client *external.PayPingaClient) *CardTrackerHandler {
	return &CardTrackerHandler{client: client}
}

func (h *CardTrackerHandler) Summary(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		writeJSON(w, http.StatusOK, map[string]any{"configured": false})
		return
	}

	summary, err := h.client.GetSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
