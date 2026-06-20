package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProject_CRUD(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/projects"

	// Create
	resp := mustPost(t, base, `{"name":"Backend rewrite","description":"Full Go rewrite","status":"planning","color":"#4A90E2"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, ok := created["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, id)
	assert.Equal(t, "Backend rewrite", created["name"])

	// List
	resp = mustGet(t, base)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var list []map[string]any
	mustDecode(t, resp, &list)
	assert.Len(t, list, 1)

	// Get by ID
	resp = mustGet(t, resourceURL(srv.URL, "projects", id))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var fetched map[string]any
	mustDecode(t, resp, &fetched)
	assert.Equal(t, "Backend rewrite", fetched["name"])

	// Partial PUT — update only status, name must survive
	resp = mustPut(t, resourceURL(srv.URL, "projects", id), `{"status":"active"}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]any
	mustDecode(t, resp, &updated)
	assert.Equal(t, "active", updated["status"])
	assert.Equal(t, "Backend rewrite", updated["name"])

	// Delete
	resp = mustDelete(t, resourceURL(srv.URL, "projects", id))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Get after delete — 404 (project handler returns 404 when nil)
	resp = mustGet(t, resourceURL(srv.URL, "projects", id))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}
