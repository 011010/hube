package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	appemail "github.com/husari/hube/internal/application/email"
)

type EmailHandler struct {
	svc *appemail.Service
}

func NewEmailHandler(svc *appemail.Service) *EmailHandler {
	return &EmailHandler{svc: svc}
}

func (h *EmailHandler) SendDigest(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Errorf("email not configured (set SMTP_HOST, SMTP_USER, SMTP_PASS, SMTP_TO)"))
		return
	}
	var body struct {
		To []string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.To) == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("to is required"))
		return
	}
	if err := h.svc.SendDigest(r.Context(), appemail.DigestOptions{To: body.To}); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}
