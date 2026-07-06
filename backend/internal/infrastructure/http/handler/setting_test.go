package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettings_ViewPreferences(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/settings"

	// PUT settings with view_preferences as a JSON string.
	resp := mustPut(t, base, `{"general":{"display_name":"Hube","view_preferences":"{\"tasks_view\":\"kanban\",\"projects_view\":\"table\",\"notes_view\":\"kanban\"}"}}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var putResult map[string]any
	mustDecode(t, resp, &putResult)
	general, ok := putResult["general"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "Hube", general["display_name"])
	assert.Equal(t, `{"tasks_view":"kanban","projects_view":"table","notes_view":"kanban"}`, general["view_preferences"])

	// GET settings and verify view_preferences persists.
	resp = mustGet(t, base)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var fetched map[string]any
	mustDecode(t, resp, &fetched)
	general, ok = fetched["general"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "Hube", general["display_name"])
	assert.Equal(t, `{"tasks_view":"kanban","projects_view":"table","notes_view":"kanban"}`, general["view_preferences"])
}
