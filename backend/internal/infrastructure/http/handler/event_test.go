package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvent_CRUD(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/events"

	startAt := "2026-01-15T09:00:00Z"
	endAt := "2026-01-15T10:00:00Z"

	// Create
	body := `{"title":"Team sync","start_at":"` + startAt + `","end_at":"` + endAt + `","color":"#FF0000"}`
	resp := mustPost(t, base, body)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, ok := created["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, id)

	// List
	resp = mustGet(t, base)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var list []map[string]any
	mustDecode(t, resp, &list)
	assert.Len(t, list, 1)

	// Get by ID
	resp = mustGet(t, resourceURL(srv.URL, "events", id))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var fetched map[string]any
	mustDecode(t, resp, &fetched)
	assert.Equal(t, "Team sync", fetched["title"])

	// Verify start_at is stored correctly (present in response)
	assert.NotEmpty(t, fetched["start_at"])

	// Partial PUT — update only title, start_at must survive
	resp = mustPut(t, resourceURL(srv.URL, "events", id), `{"title":"Updated sync"}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]any
	mustDecode(t, resp, &updated)
	assert.Equal(t, "Updated sync", updated["title"])
	assert.NotEmpty(t, updated["start_at"], "start_at should be preserved after partial PUT")

	// Delete
	resp = mustDelete(t, resourceURL(srv.URL, "events", id))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Get after delete — 404
	resp = mustGet(t, resourceURL(srv.URL, "events", id))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}
