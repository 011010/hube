package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	apptask "github.com/husari/hube/internal/application/task"
	"github.com/husari/hube/internal/domain/task"
)

type TaskHandler struct{ svc *apptask.Service }

func NewTaskHandler(svc *apptask.Service) *TaskHandler { return &TaskHandler{svc: svc} }

func (h *TaskHandler) Routes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	}
}

func (h *TaskHandler) list(w http.ResponseWriter, r *http.Request) {
	var (
		tasks []task.Task
		err   error
	)
	if pid := r.URL.Query().Get("project_id"); pid != "" {
		tasks, err = h.svc.ListByProject(r.Context(), pid)
	} else {
		tasks, err = h.svc.List(r.Context())
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) get(w http.ResponseWriter, r *http.Request) {
	t, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil || t == nil {
		writeError(w, http.StatusNotFound, nil)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *TaskHandler) create(w http.ResponseWriter, r *http.Request) {
	var t task.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.svc.Create(r.Context(), &t); err != nil {
		var valErr *task.ValidationError
		if errors.As(err, &valErr) {
			writeError(w, http.StatusBadRequest, valErr)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (h *TaskHandler) update(w http.ResponseWriter, r *http.Request) {
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
		var valErr *task.ValidationError
		if errors.As(err, &valErr) {
			writeError(w, http.StatusBadRequest, valErr)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (h *TaskHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
