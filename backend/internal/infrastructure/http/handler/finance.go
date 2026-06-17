package handler

import (
	"net/http"

	"github.com/husari/hube/internal/infrastructure/external"
)

type FinanceHandler struct{ client *external.MoneyMonkeyClient }

func NewFinanceHandler(client *external.MoneyMonkeyClient) *FinanceHandler {
	return &FinanceHandler{client: client}
}

func (h *FinanceHandler) Summary(w http.ResponseWriter, r *http.Request) {
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
