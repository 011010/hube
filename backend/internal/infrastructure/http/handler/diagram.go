package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appdiagram "github.com/husari/hube/internal/application/diagram"
	"github.com/husari/hube/internal/domain/diagram"
)

type DiagramHandler struct{ svc *appdiagram.Service }

func NewDiagramHandler(svc *appdiagram.Service) *DiagramHandler {
	return &DiagramHandler{svc: svc}
}

func (h *DiagramHandler) Routes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	}
}

func (h *DiagramHandler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *DiagramHandler) get(w http.ResponseWriter, r *http.Request) {
	d, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, nil)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *DiagramHandler) create(w http.ResponseWriter, r *http.Request) {
	var d diagram.Diagram
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.svc.Create(r.Context(), &d); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

func (h *DiagramHandler) update(w http.ResponseWriter, r *http.Request) {
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
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (h *DiagramHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
