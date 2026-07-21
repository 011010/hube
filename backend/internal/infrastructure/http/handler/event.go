package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	appevent "github.com/husari/hube/internal/application/event"
	"github.com/husari/hube/internal/domain/event"
)

type EventHandler struct{ svc *appevent.Service }

func NewEventHandler(svc *appevent.Service) *EventHandler { return &EventHandler{svc: svc} }

func (h *EventHandler) Routes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	}
}

func (h *EventHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fromStr, toStr := q.Get("from"), q.Get("to")

	if fromStr != "" && toStr != "" {
		from, err1 := time.Parse(time.RFC3339, fromStr)
		to, err2 := time.Parse(time.RFC3339, toStr)
		if err1 != nil || err2 != nil {
			writeError(w, http.StatusBadRequest, nil)
			return
		}
		events, err := h.svc.ListByRange(r.Context(), from, to)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, events)
		return
	}

	events, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (h *EventHandler) get(w http.ResponseWriter, r *http.Request) {
	e, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, nil)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (h *EventHandler) create(w http.ResponseWriter, r *http.Request) {
	var e event.Event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.svc.Create(r.Context(), &e); err != nil {
		var valErr *event.ValidationError
		if errors.As(err, &valErr) {
			writeError(w, http.StatusBadRequest, valErr)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (h *EventHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	existing, err := h.svc.Get(r.Context(), id)
	if err != nil || existing == nil {
		writeError(w, http.StatusNotFound, nil)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(existing); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	existing.ID = id
	if err := h.svc.Update(r.Context(), existing); err != nil {
		var valErr *event.ValidationError
		if errors.As(err, &valErr) {
			writeError(w, http.StatusBadRequest, valErr)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (h *EventHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
