package handler

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNote_Create_Validation(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "empty title rejected",
			body:       `{"status":"draft","priority":"medium"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid status rejected",
			body:       `{"title":"Test","status":"invalid","priority":"medium"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid priority rejected",
			body:       `{"title":"Test","status":"draft","priority":"invalid"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid due_date rejected",
			body:       `{"title":"Test","status":"draft","priority":"medium","due_date":"not-a-date"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "empty status and priority get defaults",
			body:       `{"title":"Test"}`,
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := mustPost(t, base, tt.body)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestNote_Create_Validation_DefaultsApplied(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	resp := mustPost(t, base, `{"title":"Test"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	assert.Equal(t, "draft", created["status"])
	assert.Equal(t, "medium", created["priority"])
}

func TestNote_Update_Validation(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	resp := mustPost(t, base, `{"title":"My note","status":"draft","priority":"medium"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, _ := created["id"].(string)

	// Update with empty title should be rejected by validation.
	resp = mustPut(t, resourceURL(srv.URL, "notes", id), `{"title":""}`)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestNote_CRUD(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	// Create
	resp := mustPost(t, base, `{"title":"My note","blocks":"{\"type\":\"doc\",\"content\":[{\"type\":\"paragraph\",\"content\":[{\"text\":\"Hello world\"}]}]}","status":"draft","priority":"medium","tags":["go","test"]}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, ok := created["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, id)
	assert.Equal(t, "My note", created["title"])
	assert.Equal(t, "draft", created["status"])
	assert.Equal(t, "medium", created["priority"])
	assert.Equal(t, "Hello world", created["content"])

	// List
	resp = mustGet(t, base)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var list []map[string]any
	mustDecode(t, resp, &list)
	assert.Len(t, list, 1)

	// Get by ID
	resp = mustGet(t, resourceURL(srv.URL, "notes", id))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var fetched map[string]any
	mustDecode(t, resp, &fetched)
	assert.Equal(t, "My note", fetched["title"])
	assert.Equal(t, "draft", fetched["status"])
	assert.Equal(t, "medium", fetched["priority"])
	assert.Equal(t, "Hello world", fetched["content"])

	// Partial PUT — update only tags, title/status/priority must survive
	resp = mustPut(t, resourceURL(srv.URL, "notes", id), `{"tags":["updated"]}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]any
	mustDecode(t, resp, &updated)
	assert.Equal(t, "My note", updated["title"])
	assert.Equal(t, "draft", updated["status"])
	assert.Equal(t, "medium", updated["priority"])

	// Delete
	resp = mustDelete(t, resourceURL(srv.URL, "notes", id))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Get after delete — 404
	resp = mustGet(t, resourceURL(srv.URL, "notes", id))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestNote_FTSSearch(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	// Create a note with known content
	resp := mustPost(t, base, `{"title":"Gopher guide","blocks":"{\"type\":\"doc\",\"content\":[{\"type\":\"paragraph\",\"content\":[{\"text\":\"The Go programming language is efficient\"}]}]}"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Search by keyword in title
	searchURL := base + "/search?q=" + url.QueryEscape("Gopher")
	resp = mustGet(t, searchURL)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var results []map[string]any
	mustDecode(t, resp, &results)
	assert.NotEmpty(t, results, "search should return at least one result")
	assert.Equal(t, "Gopher guide", results[0]["title"])
}

func TestNote_FTSInjection(t *testing.T) {
	srv := newTestServer(t)

	// SQL injection attempt in FTS query — should return 200 with empty array, never 500
	injected := url.QueryEscape("' OR '1'='1")
	resp := mustGet(t, srv.URL+"/api/v1/notes/search?q="+injected)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var results []map[string]any
	mustDecode(t, resp, &results)
	assert.Empty(t, results)
}
