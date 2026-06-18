package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appfolder "github.com/husari/hube/internal/application/folder"
	"github.com/husari/hube/internal/domain/folder"
)

type FolderHandler struct{ svc *appfolder.Service }

func NewFolderHandler(svc *appfolder.Service) *FolderHandler { return &FolderHandler{svc: svc} }

func (h *FolderHandler) Routes() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	}
}

func (h *FolderHandler) list(w http.ResponseWriter, r *http.Request) {
	folders, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, folders)
}

func (h *FolderHandler) create(w http.ResponseWriter, r *http.Request) {
	var f folder.Folder
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.svc.Create(r.Context(), &f); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

func (h *FolderHandler) update(w http.ResponseWriter, r *http.Request) {
	var f folder.Folder
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	f.ID = chi.URLParam(r, "id")
	if err := h.svc.Update(r.Context(), &f); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *FolderHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
