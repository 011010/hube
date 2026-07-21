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

func TestProject_RejectsInvalidPayloads(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/projects"

	tests := []struct {
		name string
		body string
	}{
		{"empty name", `{"name":""}`},
		{"whitespace-only name", `{"name":"   "}`},
		{"unknown status", `{"name":"Hube","status":"abandoned"}`},
		{"non ISO due date", `{"name":"Hube","due_date":"21/07/2026"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := mustPost(t, base, tc.body)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}

	resp := mustGet(t, base)
	var list []map[string]any
	mustDecode(t, resp, &list)
	assert.Empty(t, list)
}

func TestProject_NormalizesDefaultsOnCreate(t *testing.T) {
	srv := newTestServer(t)

	// The defaulting used to live in the application service; it now lives
	// in the entity, and this pins that the behaviour survived the move.
	resp := mustPost(t, srv.URL+"/api/v1/projects", `{"name":"  Hube  "}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	mustDecode(t, resp, &created)
	assert.Equal(t, "Hube", created["name"])
	assert.Equal(t, "planning", created["status"])
	assert.Equal(t, "#6366f1", created["color"])
}

func TestProject_EmptyDueDateBecomesNull(t *testing.T) {
	srv := newTestServer(t)

	// An untouched date input submits "", which must not fail date parsing.
	resp := mustPost(t, srv.URL+"/api/v1/projects", `{"name":"Hube","due_date":""}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	mustDecode(t, resp, &created)
	assert.Nil(t, created["due_date"])
}
