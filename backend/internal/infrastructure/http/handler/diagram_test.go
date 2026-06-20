package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiagram_CRUD(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/diagrams"

	// nodes and edges are TEXT columns (JSON stored as a string).
	// We encode them as proper JSON strings so the body is valid.
	nodesJSON := `[{"id":"n1","type":"box","label":"Start"}]`
	edgesJSON := `[{"id":"e1","source":"n1","target":"n2"}]`

	type createBody struct {
		Name  string `json:"name"`
		Nodes string `json:"nodes"`
		Edges string `json:"edges"`
	}
	raw, err := json.Marshal(createBody{Name: "Flow diagram", Nodes: nodesJSON, Edges: edgesJSON})
	if err != nil {
		t.Fatalf("marshal create body: %v", err)
	}

	resp := mustPostRaw(t, base, raw)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, ok := created["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, id)
	assert.Equal(t, "Flow diagram", created["name"])

	// List
	resp = mustGet(t, base)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var list []map[string]any
	mustDecode(t, resp, &list)
	assert.Len(t, list, 1)

	// Get by ID — verify nodes are persisted
	resp = mustGet(t, resourceURL(srv.URL, "diagrams", id))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var fetched map[string]any
	mustDecode(t, resp, &fetched)
	assert.Equal(t, "Flow diagram", fetched["name"])
	assert.NotEmpty(t, fetched["nodes"], "nodes should be persisted")

	// Partial PUT — update only name, nodes must survive
	resp = mustPut(t, resourceURL(srv.URL, "diagrams", id), `{"name":"Renamed diagram"}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]any
	mustDecode(t, resp, &updated)
	assert.Equal(t, "Renamed diagram", updated["name"])
	assert.NotEmpty(t, updated["nodes"], "nodes should survive a name-only PUT")

	// Delete
	resp = mustDelete(t, resourceURL(srv.URL, "diagrams", id))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Get after delete — 404
	resp = mustGet(t, resourceURL(srv.URL, "diagrams", id))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}
