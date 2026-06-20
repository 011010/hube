package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWishlist_CRUD(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/wishlist"

	// Create
	resp := mustPost(t, base, `{"name":"Mechanical keyboard","priority":"high","status":"pending"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, ok := created["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, id)
	assert.Equal(t, "Mechanical keyboard", created["name"])

	// List
	resp = mustGet(t, base)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var list []map[string]any
	mustDecode(t, resp, &list)
	assert.Len(t, list, 1)

	// Get by ID
	resp = mustGet(t, resourceURL(srv.URL, "wishlist", id))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var fetched map[string]any
	mustDecode(t, resp, &fetched)
	assert.Equal(t, "Mechanical keyboard", fetched["name"])

	// Partial PUT — update only priority, name must survive
	resp = mustPut(t, resourceURL(srv.URL, "wishlist", id), `{"priority":"low"}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]any
	mustDecode(t, resp, &updated)
	assert.Equal(t, "low", updated["priority"])
	assert.Equal(t, "Mechanical keyboard", updated["name"])

	// Delete
	resp = mustDelete(t, resourceURL(srv.URL, "wishlist", id))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Get after delete — 404
	resp = mustGet(t, resourceURL(srv.URL, "wishlist", id))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}
