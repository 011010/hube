package handler

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	msg := "internal error"
	if err != nil {
		msg = err.Error()
	}
	writeJSON(w, status, map[string]string{"error": msg})
}
