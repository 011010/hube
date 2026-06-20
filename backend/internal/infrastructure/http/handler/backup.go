package handler

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/husari/hube/internal/application/backup"
)

// BackupHandler handles on-demand backup and data export requests.
type BackupHandler struct {
	backup *backup.Service
	export *backup.ExportService
}

// NewBackupHandler creates a new BackupHandler.
func NewBackupHandler(backupSvc *backup.Service, exportSvc *backup.ExportService) *BackupHandler {
	return &BackupHandler{backup: backupSvc, export: exportSvc}
}

// CreateBackup triggers an immediate database backup.
// POST /api/v1/backup
func (h *BackupHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	if h.backup == nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Errorf("backup service not configured"))
		return
	}

	path, err := h.backup.CreateBackup(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	info, err := os.Stat(path)
	var sizeBytes int64
	if err == nil {
		sizeBytes = info.Size()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"path":       path,
		"size_bytes": sizeBytes,
	})
}

// Export builds an in-memory ZIP of all data and sends it as a file download.
// GET /api/v1/export
func (h *BackupHandler) Export(w http.ResponseWriter, r *http.Request) {
	if h.export == nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Errorf("export service not configured"))
		return
	}

	data, err := h.export.BuildZip(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	date := time.Now().UTC().Format("2006-01-02")
	filename := fmt.Sprintf("hube_export_%s.zip", date)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data) //nolint:errcheck
}
