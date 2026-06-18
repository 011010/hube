package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appproject "github.com/husari/hube/internal/application/project"
	"github.com/husari/hube/internal/domain/project"
)

type ProjectHandler struct{ svc *appproject.Service }

func NewProjectHandler(svc *appproject.Service) *ProjectHandler { return &ProjectHandler{svc: svc} }

func (h *ProjectHandler) Routes() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	}
}

func (h *ProjectHandler) list(w http.ResponseWriter, r *http.Request) {
	projects, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if p == nil {
		writeError(w, http.StatusNotFound, nil)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProjectHandler) create(w http.ResponseWriter, r *http.Request) {
	var p project.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.svc.Create(r.Context(), &p); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *ProjectHandler) update(w http.ResponseWriter, r *http.Request) {
	var p project.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	p.ID = chi.URLParam(r, "id")
	if err := h.svc.Update(r.Context(), &p); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProjectHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
