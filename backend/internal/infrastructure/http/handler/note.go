package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	appnote "github.com/husari/hube/internal/application/note"
	"github.com/husari/hube/internal/domain/note"
)

type NoteHandler struct {
	svc *appnote.Service
	rag *appnote.RAGService
}

func NewNoteHandler(svc *appnote.Service, rag *appnote.RAGService) *NoteHandler {
	return &NoteHandler{svc: svc, rag: rag}
}

func (h *NoteHandler) Routes() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.list)
		r.Get("/search", h.search)
		r.Post("/semantic-search", h.semanticSearch)
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

func (h *NoteHandler) semanticSearch(w http.ResponseWriter, r *http.Request) {
	if h.rag == nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Errorf("semantic search not configured (set OPENAI_API_KEY)"))
		return
	}
	var body struct {
		Q    string `json:"q"`
		TopK int    `json:"top_k"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Q == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("q is required"))
		return
	}
	if body.TopK <= 0 || body.TopK > 50 {
		body.TopK = 5
	}
	results, err := h.rag.SemanticSearch(r.Context(), body.Q, body.TopK)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *NoteHandler) create(w http.ResponseWriter, r *http.Request) {
	var n note.Note
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.svc.Create(r.Context(), &n); err != nil {
		var valErr *note.ValidationError
		if errors.As(err, &valErr) {
			writeError(w, http.StatusBadRequest, valErr)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if h.rag != nil {
		// Detach from request context so indexing isn't cancelled when the handler returns.
		go h.rag.IndexNote(context.WithoutCancel(r.Context()), &n)
	}
	writeJSON(w, http.StatusCreated, n)
}

func (h *NoteHandler) update(w http.ResponseWriter, r *http.Request) {
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
		var valErr *note.ValidationError
		if errors.As(err, &valErr) {
			writeError(w, http.StatusBadRequest, valErr)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if h.rag != nil {
		go h.rag.IndexNote(context.WithoutCancel(r.Context()), existing)
	}
	writeJSON(w, http.StatusOK, existing)
}

func (h *NoteHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
