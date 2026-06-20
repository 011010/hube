package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appapp "github.com/husari/hube/internal/application/app"
	"github.com/husari/hube/internal/domain/app"
)

type AppHandler struct{ svc *appapp.Service }

func NewAppHandler(svc *appapp.Service) *AppHandler { return &AppHandler{svc: svc} }

func (h *AppHandler) Routes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	}
}

func (h *AppHandler) list(w http.ResponseWriter, r *http.Request) {
	apps, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, apps)
}

func (h *AppHandler) get(w http.ResponseWriter, r *http.Request) {
	a, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, nil)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *AppHandler) create(w http.ResponseWriter, r *http.Request) {
	var a app.App
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.svc.Create(r.Context(), &a); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (h *AppHandler) update(w http.ResponseWriter, r *http.Request) {
	var a app.App
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	a.ID = chi.URLParam(r, "id")
	if err := h.svc.Update(r.Context(), &a); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *AppHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
