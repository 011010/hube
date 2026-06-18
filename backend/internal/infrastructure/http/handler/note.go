package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appnote "github.com/husari/hube/internal/application/note"
	"github.com/husari/hube/internal/domain/note"
)

type NoteHandler struct{ svc *appnote.Service }

func NewNoteHandler(svc *appnote.Service) *NoteHandler { return &NoteHandler{svc: svc} }

func (h *NoteHandler) Routes() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.list)
		r.Get("/search", h.search)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	}
}

func (h *NoteHandler) list(w http.ResponseWriter, r *http.Request) {
	var folderID *string
	if f := r.URL.Query().Get("folder_id"); f != "" {
		folderID = &f
	}
	notes, err := h.svc.List(r.Context(), folderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

func (h *NoteHandler) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, []note.Note{})
		return
	}
	notes, err := h.svc.Search(r.Context(), q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

func (h *NoteHandler) get(w http.ResponseWriter, r *http.Request) {
	n, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if n == nil {
		writeError(w, http.StatusNotFound, nil)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *NoteHandler) create(w http.ResponseWriter, r *http.Request) {
	var n note.Note
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.svc.Create(r.Context(), &n); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

func (h *NoteHandler) update(w http.ResponseWriter, r *http.Request) {
	var n note.Note
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	n.ID = chi.URLParam(r, "id")
	if err := h.svc.Update(r.Context(), &n); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *NoteHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
