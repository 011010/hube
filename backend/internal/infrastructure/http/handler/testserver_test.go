package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	appapp "github.com/husari/hube/internal/application/app"
	appdiagram "github.com/husari/hube/internal/application/diagram"
	appevent "github.com/husari/hube/internal/application/event"
	appfolder "github.com/husari/hube/internal/application/folder"
	appnote "github.com/husari/hube/internal/application/note"
	appproject "github.com/husari/hube/internal/application/project"
	appsetting "github.com/husari/hube/internal/application/setting"
	apptask "github.com/husari/hube/internal/application/task"
	appwishlist "github.com/husari/hube/internal/application/wishlist"
	"github.com/husari/hube/internal/infrastructure/http/middleware"
	"github.com/husari/hube/internal/infrastructure/sqlite"
)

// testToken is the bearer token the test router requires; the must* helpers
// attach it to every request, mirroring the production auth middleware.
const testToken = "test-token"

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}

	taskRepo := sqlite.NewTaskRepo(db)
	eventRepo := sqlite.NewEventRepo(db)
	appRepo := sqlite.NewAppRepo(db)
	noteRepo := sqlite.NewNoteRepo(db)
	folderRepo := sqlite.NewFolderRepo(db)
	projectRepo := sqlite.NewProjectRepo(db)
	wishlistRepo := sqlite.NewWishlistRepo(db)
	diagramRepo := sqlite.NewDiagramRepo(db)
	settingRepo := sqlite.NewSettingRepo(db)

	taskSvc := apptask.NewService(taskRepo)
	eventSvc := appevent.NewService(eventRepo)
	appSvc := appapp.NewService(appRepo)
	noteSvc := appnote.NewService(noteRepo)
	folderSvc := appfolder.NewService(folderRepo)
	projectSvc := appproject.NewService(projectRepo)
	wishlistSvc := appwishlist.NewService(wishlistRepo)
	diagramSvc := appdiagram.NewService(diagramRepo)
	settingSvc := appsetting.NewService(settingRepo)

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.Body = http.MaxBytesReader(w, req.Body, 2<<20) // 2 MB
			next.ServeHTTP(w, req)
		})
	})
	r.Use(middleware.Auth(testToken))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/tasks", NewTaskHandler(taskSvc).Routes())
		r.Route("/events", NewEventHandler(eventSvc).Routes())
		r.Route("/apps", NewAppHandler(appSvc).Routes())
		r.Route("/notes", NewNoteHandler(noteSvc, nil).Routes())
		r.Route("/folders", NewFolderHandler(folderSvc).Routes())
		r.Route("/projects", NewProjectHandler(projectSvc).Routes())
		r.Route("/wishlist", NewWishlistHandler(wishlistSvc).Routes())
		r.Route("/diagrams", NewDiagramHandler(diagramSvc).Routes())
		sh := NewSettingHandler(settingSvc)
		r.Get("/settings", sh.Get)
		r.Put("/settings", sh.Put)
	})

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(r)
	t.Cleanup(func() {
		srv.Close()
		db.Close()
	})
	return srv
}

// mustDo sends req with the test bearer token attached and returns the response.
func mustDo(t *testing.T, req *http.Request) *http.Response {
	t.Helper()
	req.Header.Set("Authorization", "Bearer "+testToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", req.Method, req.URL, err)
	}
	return resp
}

// mustPost sends a POST request with a JSON body and returns the response.
func mustPost(t *testing.T, url, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body)) //nolint:noctx
	if err != nil {
		t.Fatalf("new POST request %s: %v", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	return mustDo(t, req)
}

// mustGet sends a GET request and returns the response.
func mustGet(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil) //nolint:noctx
	if err != nil {
		t.Fatalf("new GET request %s: %v", url, err)
	}
	return mustDo(t, req)
}

// mustPut sends a PUT request with a JSON body and returns the response.
func mustPut(t *testing.T, url, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new PUT request %s: %v", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	return mustDo(t, req)
}

// mustDelete sends a DELETE request and returns the response.
func mustDelete(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		t.Fatalf("new DELETE request %s: %v", url, err)
	}
	return mustDo(t, req)
}

// mustDecode decodes the JSON body of a response into v, then closes the body.
func mustDecode(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}

// mustPostRaw sends a POST with raw bytes (used for body-limit tests).
func mustPostRaw(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		t.Fatalf("new POST request %s: %v", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	return mustDo(t, req)
}

// resourceURL builds /api/v1/<resource>/<id>.
func resourceURL(base, resource, id string) string {
	return fmt.Sprintf("%s/api/v1/%s/%s", base, resource, id)
}
