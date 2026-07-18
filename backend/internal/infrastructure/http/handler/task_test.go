package handler

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTask_CRUD(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/tasks"

	// Create
	resp := mustPost(t, base, `{"title":"Test task","priority":"high","status":"todo"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, ok := created["id"].(string)
	assert.True(t, ok, "id should be a string")
	assert.NotEmpty(t, id)

	// List — expect exactly one item
	resp = mustGet(t, base)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var list []map[string]any
	mustDecode(t, resp, &list)
	assert.Len(t, list, 1)

	// Get by ID
	resp = mustGet(t, resourceURL(srv.URL, "tasks", id))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var fetched map[string]any
	mustDecode(t, resp, &fetched)
	assert.Equal(t, id, fetched["id"])
	assert.Equal(t, "Test task", fetched["title"])

	// Partial PUT — only update status
	resp = mustPut(t, resourceURL(srv.URL, "tasks", id), `{"status":"done"}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]any
	mustDecode(t, resp, &updated)
	assert.Equal(t, "done", updated["status"])
	// title and priority must be preserved because handler reads existing first
	assert.Equal(t, "Test task", updated["title"])
	assert.Equal(t, "high", updated["priority"])

	// Delete
	resp = mustDelete(t, resourceURL(srv.URL, "tasks", id))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Get after delete — must be 404
	resp = mustGet(t, resourceURL(srv.URL, "tasks", id))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestTask_PartialPUT_PreservesFields(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/tasks"

	resp := mustPost(t, base, `{"title":"Important task","description":"Some details","priority":"high","status":"todo"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id := created["id"].(string)

	// PUT with only status — other fields should survive
	resp = mustPut(t, resourceURL(srv.URL, "tasks", id), `{"status":"done"}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]any
	mustDecode(t, resp, &updated)

	assert.Equal(t, "done", updated["status"])
	assert.Equal(t, "Important task", updated["title"])
	assert.Equal(t, "Some details", updated["description"])
	assert.Equal(t, "high", updated["priority"])
}

func TestTask_BodyLimit(t *testing.T) {
	srv := newTestServer(t)

	// 3 MB body — exceeds the 2 MB limit set in the router
	body := bytes.Repeat([]byte("x"), 3<<20)
	resp := mustPostRaw(t, srv.URL+"/api/v1/tasks", body)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestTask_NotFound_MasksInternalError(t *testing.T) {
	srv := newTestServer(t)

	resp := mustGet(t, resourceURL(srv.URL, "tasks", "nonexistent-id"))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var body map[string]any
	mustDecode(t, resp, &body)

	errMsg, _ := body["error"].(string)
	assert.False(t, strings.Contains(errMsg, "sql:"),
		"error message must not leak internal sql details, got: %q", errMsg)
	assert.Equal(t, "internal server error", errMsg)
}
