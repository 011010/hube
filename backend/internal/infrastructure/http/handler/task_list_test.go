package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTask_ListFiltersByStatus(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/tasks"

	// The CLI's `tasks --status` sends ?status=; the API used to ignore it
	// and return every task.
	resp := mustPost(t, base, `{"title":"open one"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	resp = mustPost(t, base, `{"title":"done one"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id := created["id"].(string)

	resp = mustPut(t, resourceURL(srv.URL, "tasks", id), `{"status":"done"}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = mustGet(t, base+"?status=done")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var done []map[string]any
	mustDecode(t, resp, &done)
	assert.Len(t, done, 1)
	assert.Equal(t, "done one", done[0]["title"])
}
