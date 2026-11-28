package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
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
		wantErr    string
	}{
		{
			name:       "empty title rejected",
			body:       `{"status":"draft"}`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "title is required",
		},
		{
			name:       "invalid status rejected",
			body:       `{"title":"Test","status":"invalid"}`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid status",
		},
		{
			name:       "empty status gets default",
			body:       `{"title":"Test"}`,
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := mustPost(t, base, tt.body)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantErr != "" {
				var body map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("decode error body: %v", err)
				}
				errMsg, _ := body["error"].(string)
				assert.Contains(t, errMsg, tt.wantErr)
			}
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
}

func TestNote_Create_RejectsProperties(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	resp := mustPost(t, base, `{"title":"Props","properties":[{"key":"rating","type":"number","value":8}]}`)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	errMsg, _ := errBody["error"].(string)
	assert.Contains(t, errMsg, "/notes/{id}/properties")

	resp = mustPost(t, base, `{"title":"No Props"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	props, ok := created["properties"].([]any)
	assert.True(t, ok, "expected properties array in create response, got %#v", created["properties"])
	assert.NotNil(t, props)
	assert.Empty(t, props)
}

func TestNote_Update_Validation(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	resp := mustPost(t, base, `{"title":"My note","status":"draft"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, _ := created["id"].(string)

	// Update with empty title should be rejected by validation.
	resp = mustPut(t, resourceURL(srv.URL, "notes", id), `{"title":""}`)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	errMsg, _ := body["error"].(string)
	assert.Contains(t, errMsg, "title is required")
}

func TestNote_Create_MalformedBlocksRejectedAs400(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	resp := mustPost(t, base, `{"title":"Test","status":"draft","blocks":"not json"}`)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNote_Update_MalformedBlocksRejectedAs400(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	resp := mustPost(t, base, `{"title":"Test","status":"draft"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, _ := created["id"].(string)

	resp = mustPut(t, resourceURL(srv.URL, "notes", id), `{"blocks":"not json"}`)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNote_CRUD(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	// Create
	resp := mustPost(t, base, `{"title":"My note","blocks":"{\"type\":\"doc\",\"content\":[{\"type\":\"paragraph\",\"content\":[{\"text\":\"Hello world\"}]}]}","status":"draft","tags":["go","test"]}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, ok := created["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, id)
	assert.Equal(t, "My note", created["title"])
	assert.Equal(t, "draft", created["status"])
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
	assert.Equal(t, "Hello world", fetched["content"])

	// Partial PUT — update only tags, title/status must survive
	resp = mustPut(t, resourceURL(srv.URL, "notes", id), `{"tags":["updated"]}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]any
	mustDecode(t, resp, &updated)
	assert.Equal(t, "My note", updated["title"])
	assert.Equal(t, "draft", updated["status"])

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

func TestNoteGraph(t *testing.T) {
	srv := newTestServer(t)

	// Note A — will be linked to by Note B and referenced by a project.
	resp := mustPost(t, srv.URL+"/api/v1/notes", `{"title":"Note A"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var noteA map[string]any
	mustDecode(t, resp, &noteA)
	noteAID := noteA["id"].(string)

	// Note B — links to Note A via [[Note A]].
	resp = mustPost(t, srv.URL+"/api/v1/notes", `{"title":"Note B","blocks":"{\"type\":\"doc\",\"content\":[{\"type\":\"paragraph\",\"content\":[{\"type\":\"text\",\"text\":\"See [[Note A]]\"}]}]}"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var noteB map[string]any
	mustDecode(t, resp, &noteB)
	noteBID := noteB["id"].(string)

	// Task linked to Note B.
	resp = mustPost(t, srv.URL+"/api/v1/tasks", `{"title":"Linked task","note_id":"`+noteBID+`"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Project linked to Note A.
	resp = mustPost(t, srv.URL+"/api/v1/projects", `{"name":"Linked project","note_id":"`+noteAID+`"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	resp = mustGet(t, srv.URL+"/api/v1/notes/graph")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var graph struct {
		Nodes []map[string]any `json:"nodes"`
		Edges []map[string]any `json:"edges"`
	}
	mustDecode(t, resp, &graph)

	assert.Len(t, graph.Nodes, 4, "expected 2 notes + 1 task + 1 project as nodes")
	assert.Len(t, graph.Edges, 3, "expected 1 link edge + 1 task edge + 1 project edge")

	nodeByID := make(map[string]map[string]any, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodeByID[n["id"].(string)] = n
	}

	noteANode, ok := nodeByID["note:"+noteAID]
	assert.True(t, ok, "expected note:%s node", noteAID)
	assert.Equal(t, "note", noteANode["type"])
	assert.Equal(t, "Note A", noteANode["label"])

	noteBNode, ok := nodeByID["note:"+noteBID]
	assert.True(t, ok, "expected note:%s node", noteBID)
	assert.Equal(t, "note", noteBNode["type"])

	var hasLinkEdge, hasTaskEdge, hasProjectEdge bool
	for _, e := range graph.Edges {
		src, _ := e["source"].(string)
		tgt, _ := e["target"].(string)
		typ, _ := e["type"].(string)
		switch {
		case typ == "link" && src == "note:"+noteBID && tgt == "note:"+noteAID:
			hasLinkEdge = true
		case typ == "task" && tgt == "note:"+noteBID:
			hasTaskEdge = true
		case typ == "project" && tgt == "note:"+noteAID:
			hasProjectEdge = true
		}
	}
	assert.True(t, hasLinkEdge, "expected link edge from Note B to Note A, got %+v", graph.Edges)
	assert.True(t, hasTaskEdge, "expected task edge to Note B, got %+v", graph.Edges)
	assert.True(t, hasProjectEdge, "expected project edge to Note A, got %+v", graph.Edges)
}

func TestNoteExport(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	resp := mustPost(t, base, `{"title":"Export Me","blocks":"{\"type\":\"doc\",\"content\":[{\"type\":\"paragraph\",\"content\":[{\"text\":\"Hello world\"}]}]}","status":"draft","tags":["go","test"]}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, _ := created["id"].(string)

	resp = mustGet(t, resourceURL(srv.URL, "notes", id)+"/export")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	contentType := resp.Header.Get("Content-Type")
	assert.True(t, strings.HasPrefix(contentType, "text/markdown") || strings.HasPrefix(contentType, "text/plain"),
		"expected markdown/plain content type, got %q", contentType)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	md := string(body)

	assert.True(t, strings.HasPrefix(md, "---\n"), "expected YAML frontmatter delimiter, got: %s", md)
	assert.Contains(t, md, `title: "Export Me"`)
	assert.Contains(t, md, "status: draft")
	assert.Contains(t, md, `  - "go"`)
	assert.Contains(t, md, `  - "test"`)
	assert.Contains(t, md, "Hello world")
}

func TestNoteExport_NotFound(t *testing.T) {
	srv := newTestServer(t)

	resp := mustGet(t, resourceURL(srv.URL, "notes", "does-not-exist")+"/export")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestNote_PropertiesRoundTrip(t *testing.T) {
	srv := newTestServer(t)
	resp := mustPost(t, srv.URL+"/api/v1/notes", `{"title":"Props"}`)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, _ := created["id"].(string)

	body := `{"properties":[
		{"key":" Rating ","type":" NUMBER ","value":8},
		{"key":"source","type":"text","value":"https://example.com"},
		{"key":"published","type":"date","value":"2024-03-11"},
		{"key":"topics","type":"list","value":["design","tools"]}
	]}`
	put := mustPut(t, srv.URL+"/api/v1/notes/"+id+"/properties", body)
	assert.Equal(t, http.StatusOK, put.StatusCode)
	putProps := propsByKey(t, put)
	assert.Len(t, putProps, 4)
	assert.Equal(t, "rating", putProps["rating"]["key"])
	assert.Equal(t, "number", putProps["rating"]["type"])
	assert.Equal(t, float64(8), putProps["rating"]["value"])
	assert.Equal(t, "https://example.com", putProps["source"]["value"])
	assert.Equal(t, "2024-03-11", putProps["published"]["value"])
	assert.Equal(t, []any{"design", "tools"}, putProps["topics"]["value"])

	get := mustGet(t, srv.URL+"/api/v1/notes/"+id+"/properties")
	assert.Equal(t, http.StatusOK, get.StatusCode)
	getProps := propsByKey(t, get)
	assert.Len(t, getProps, 4)
	assert.Equal(t, float64(8), getProps["rating"]["value"])
	assert.Equal(t, "https://example.com", getProps["source"]["value"])
	assert.Equal(t, "2024-03-11", getProps["published"]["value"])
	assert.Equal(t, []any{"design", "tools"}, getProps["topics"]["value"])

	noteResp := mustGet(t, srv.URL+"/api/v1/notes/"+id)
	var noteBody map[string]any
	mustDecode(t, noteResp, &noteBody)
	assert.Len(t, noteBody["properties"], 4)
}

func TestNote_PropertiesValidation(t *testing.T) {
	srv := newTestServer(t)
	resp := mustPost(t, srv.URL+"/api/v1/notes", `{"title":"Props"}`)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, _ := created["id"].(string)

	bad := mustPut(t, srv.URL+"/api/v1/notes/"+id+"/properties",
		`{"properties":[{"key":"rating","type":"number","value":"8"}]}`)
	defer bad.Body.Close()
	assert.Equal(t, http.StatusBadRequest, bad.StatusCode)

	reserved := mustPut(t, srv.URL+"/api/v1/notes/"+id+"/properties",
		`{"properties":[{"key":"title","type":"text","value":"x"}]}`)
	defer reserved.Body.Close()
	assert.Equal(t, http.StatusBadRequest, reserved.StatusCode)
}

func TestNote_PropertiesNotFound(t *testing.T) {
	srv := newTestServer(t)
	resp := mustGet(t, srv.URL+"/api/v1/notes/missing/properties")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	put := mustPut(t, srv.URL+"/api/v1/notes/missing/properties", `{"properties":[]}`)
	defer put.Body.Close()
	assert.Equal(t, http.StatusNotFound, put.StatusCode)
}

func TestNote_PropertiesClear(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	resp := mustPost(t, base, `{"title":"Props"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, _ := created["id"].(string)
	propsURL := resourceURL(srv.URL, "notes", id) + "/properties"

	set := mustPut(t, propsURL, `{"properties":[{"key":"rating","type":"number","value":8}]}`)
	assert.Equal(t, http.StatusOK, set.StatusCode)
	set.Body.Close()

	clear := mustPut(t, propsURL, `{"properties":[]}`)
	assert.Equal(t, http.StatusOK, clear.StatusCode)
	var cleared []map[string]any
	mustDecode(t, clear, &cleared)
	assert.NotNil(t, cleared, "expected [] not null from PUT response")
	assert.Empty(t, cleared)

	got := mustGet(t, propsURL)
	assert.Equal(t, http.StatusOK, got.StatusCode)
	var props []map[string]any
	mustDecode(t, got, &props)
	assert.NotNil(t, props, "expected [] not null")
	assert.Empty(t, props)
}

func TestNote_PropertiesBodyContract(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	resp := mustPost(t, base, `{"title":"Props"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, _ := created["id"].(string)
	propsURL := resourceURL(srv.URL, "notes", id) + "/properties"

	seed := mustPut(t, propsURL, `{"properties":[{"key":"rating","type":"number","value":8}]}`)
	assert.Equal(t, http.StatusOK, seed.StatusCode)
	seed.Body.Close()

	assertStoredRating := func(t *testing.T) {
		t.Helper()
		got := mustGet(t, propsURL)
		assert.Equal(t, http.StatusOK, got.StatusCode)
		var stored []map[string]any
		mustDecode(t, got, &stored)
		if assert.Len(t, stored, 1) {
			assert.Equal(t, "rating", stored[0]["key"])
			assert.Equal(t, float64(8), stored[0]["value"])
		}
	}

	t.Run("missing properties field is rejected", func(t *testing.T) {
		bad := mustPut(t, propsURL, `{}`)
		defer bad.Body.Close()
		assert.Equal(t, http.StatusBadRequest, bad.StatusCode)
		assertStoredRating(t)
	})

	t.Run("null properties is rejected", func(t *testing.T) {
		bad := mustPut(t, propsURL, `{"properties":null}`)
		defer bad.Body.Close()
		assert.Equal(t, http.StatusBadRequest, bad.StatusCode)
		assertStoredRating(t)
	})

	t.Run("empty array clears", func(t *testing.T) {
		clear := mustPut(t, propsURL, `{"properties":[]}`)
		assert.Equal(t, http.StatusOK, clear.StatusCode)
		var cleared []map[string]any
		mustDecode(t, clear, &cleared)
		assert.NotNil(t, cleared, "expected [] not null from PUT response")
		assert.Empty(t, cleared)

		got := mustGet(t, propsURL)
		assert.Equal(t, http.StatusOK, got.StatusCode)
		var stored []map[string]any
		mustDecode(t, got, &stored)
		assert.NotNil(t, stored, "expected [] not null")
		assert.Empty(t, stored)
	})

	t.Run("object instead of array is rejected", func(t *testing.T) {
		bad := mustPut(t, propsURL, `{"properties":{}}`)
		defer bad.Body.Close()
		assert.Equal(t, http.StatusBadRequest, bad.StatusCode)
	})

	t.Run("malformed json is rejected", func(t *testing.T) {
		bad := mustPut(t, propsURL, `{"properties":[`)
		defer bad.Body.Close()
		assert.Equal(t, http.StatusBadRequest, bad.StatusCode)
	})
}

func TestNote_PropertiesNotOverwrittenByNoteUpdate(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/notes"

	resp := mustPost(t, base, `{"title":"Props"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, _ := created["id"].(string)
	propsURL := resourceURL(srv.URL, "notes", id) + "/properties"

	set := mustPut(t, propsURL, `{"properties":[{"key":"rating","type":"number","value":8}]}`)
	assert.Equal(t, http.StatusOK, set.StatusCode)
	set.Body.Close()

	upd := mustPut(t, resourceURL(srv.URL, "notes", id),
		`{"title":"Renamed","properties":[{"key":"rating","type":"number","value":99}]}`)
	assert.Equal(t, http.StatusOK, upd.StatusCode)
	var updated map[string]any
	mustDecode(t, upd, &updated)
	assert.Equal(t, "Renamed", updated["title"])

	props, ok := updated["properties"].([]any)
	if assert.True(t, ok, "expected properties array in update response, got %#v", updated["properties"]) &&
		assert.Len(t, props, 1) {
		p, _ := props[0].(map[string]any)
		assert.Equal(t, float64(8), p["value"])
	}

	got := mustGet(t, propsURL)
	assert.Equal(t, http.StatusOK, got.StatusCode)
	var stored []map[string]any
	mustDecode(t, got, &stored)
	if assert.Len(t, stored, 1) {
		assert.Equal(t, float64(8), stored[0]["value"])
	}
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

// propsByKey decodes a properties response into a map keyed by property key.
func propsByKey(t *testing.T, resp *http.Response) map[string]map[string]any {
	t.Helper()
	var props []map[string]any
	mustDecode(t, resp, &props)
	byKey := make(map[string]map[string]any, len(props))
	for _, p := range props {
		key, _ := p["key"].(string)
		byKey[key] = p
	}
	return byKey
}
