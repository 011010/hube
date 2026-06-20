package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApp_CRUD(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/apps"

	// Create
	resp := mustPost(t, base, `{"name":"GitHub","url":"https://github.com","icon":"github","color":"#000000","active":true}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, ok := created["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, id)

	// List — migrations seed 2 apps, so we expect at least the one we just created
	resp = mustGet(t, base)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var list []map[string]any
	mustDecode(t, resp, &list)
	assert.GreaterOrEqual(t, len(list), 1, "list should contain at least the created app")

	// Get by ID
	resp = mustGet(t, resourceURL(srv.URL, "apps", id))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var fetched map[string]any
	mustDecode(t, resp, &fetched)
	assert.Equal(t, "GitHub", fetched["name"])
	assert.Equal(t, "https://github.com", fetched["url"])

	// Partial PUT — update only icon, name and url must survive
	resp = mustPut(t, resourceURL(srv.URL, "apps", id), `{"icon":"gh-updated"}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]any
	mustDecode(t, resp, &updated)
	assert.Equal(t, "gh-updated", updated["icon"])
	assert.Equal(t, "GitHub", updated["name"])
	assert.Equal(t, "https://github.com", updated["url"])

	// Delete
	resp = mustDelete(t, resourceURL(srv.URL, "apps", id))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Get after delete — 404
	resp = mustGet(t, resourceURL(srv.URL, "apps", id))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}
