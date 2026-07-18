package handler

import (
	"io"
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

func TestSettings_ViewPreferences_OmittedOnPut(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/settings"

	// Seed a value first.
	resp := mustPut(t, base, `{"general":{"display_name":"Hube","view_preferences":"{\"tasks_view\":\"kanban\"}"}}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// PUT settings without view_preferences — full replacement erases the previous value.
	resp = mustPut(t, base, `{"general":{"display_name":"NoPrefs"}}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var putResult map[string]any
	mustDecode(t, resp, &putResult)
	general, ok := putResult["general"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "NoPrefs", general["display_name"])
	assert.Nil(t, general["view_preferences"], "view_preferences should be omitted when not provided")

	// GET settings and verify view_preferences is no longer present.
	resp = mustGet(t, base)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var fetched map[string]any
	mustDecode(t, resp, &fetched)
	general, ok = fetched["general"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "NoPrefs", general["display_name"])
	assert.Nil(t, general["view_preferences"], "view_preferences should be erased after PUT without it")
}

func TestSettings_KeysNeverReturned(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/settings"

	// Store integration keys via PUT.
	resp := mustPut(t, base, `{"integrations":{"monkeyapi_url":"https://mm.example","monkeyapi_key":"secret-monkey","paypinga_url":"https://pp.example","paypinga_key":"secret-pay"}}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var putResult map[string]any
	mustDecode(t, resp, &putResult)
	integrations, ok := putResult["integrations"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, true, integrations["monkeyapi_key_set"])
	assert.Equal(t, true, integrations["paypinga_key_set"])
	assert.Equal(t, true, integrations["monkeyapi_enabled"])
	assert.Equal(t, true, integrations["paypinga_enabled"])
	_, leaked := integrations["monkeyapi_key"]
	assert.False(t, leaked, "PUT response must not contain monkeyapi_key")
	_, leaked = integrations["paypinga_key"]
	assert.False(t, leaked, "PUT response must not contain paypinga_key")

	// GET must report key_set booleans and never leak the raw key values.
	resp = mustGet(t, base)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.NoError(t, err)
	assert.NotContains(t, string(body), "secret-monkey")
	assert.NotContains(t, string(body), "secret-pay")
	assert.Contains(t, string(body), `"monkeyapi_key_set":true`)
	assert.Contains(t, string(body), `"paypinga_key_set":true`)
}

func TestSettings_EmptyKeyPreservesStored(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/settings"

	// Store a key first.
	resp := mustPut(t, base, `{"integrations":{"monkeyapi_key":"secret-monkey"}}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// PUT with an empty key must leave the stored key unchanged.
	resp = mustPut(t, base, `{"general":{"display_name":"Hube"},"integrations":{"monkeyapi_key":""}}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var result map[string]any
	mustDecode(t, resp, &result)
	integrations, ok := result["integrations"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, true, integrations["monkeyapi_key_set"], "empty key in PUT must preserve the stored key")

	// PUT with a new key must overwrite the stored one.
	resp = mustPut(t, base, `{"integrations":{"monkeyapi_key":"new-secret"}}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mustDecode(t, resp, &result)
	integrations, ok = result["integrations"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, true, integrations["monkeyapi_key_set"], "key_set must stay true after updating the key")
}
