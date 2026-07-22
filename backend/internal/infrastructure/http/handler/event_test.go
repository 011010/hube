package handler

import (
	"fmt"
	"net/http"
	"net/url"
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

func TestEvent_RejectsInvalidPayloads(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/events"

	// Each of these used to return 201 and persist as written.
	tests := []struct {
		name string
		body string
	}{
		{"empty title", `{"title":"","start_at":"2026-07-21T09:00:00Z","end_at":"2026-07-21T10:00:00Z"}`},
		{"whitespace-only title", `{"title":"  ","start_at":"2026-07-21T09:00:00Z","end_at":"2026-07-21T10:00:00Z"}`},
		{"end before start", `{"title":"Standup","start_at":"2026-07-21T15:00:00Z","end_at":"2026-07-21T09:00:00Z"}`},
		{"missing start", `{"title":"Standup","end_at":"2026-07-21T10:00:00Z"}`},
		{"missing end", `{"title":"Standup","start_at":"2026-07-21T09:00:00Z"}`},
		{"no timestamps at all", `{"title":"Standup"}`},
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

func TestEvent_EveryCreatedEventIsReachableFromTheCalendar(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/events"

	// An event without timestamps used to be stored at year 1, where the
	// range query — which filters on start_at alone — could never find it.
	// It existed, but no calendar view could show or delete it.
	assert.Equal(t, http.StatusBadRequest, mustPost(t, base, `{"title":"ghost"}`).StatusCode)

	resp := mustPost(t, base, `{"title":"Standup","start_at":"2026-07-21T09:00:00Z","end_at":"2026-07-21T10:00:00Z"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var all []map[string]any
	mustDecode(t, mustGet(t, base), &all)

	from := url.QueryEscape("2026-07-01T00:00:00Z")
	to := url.QueryEscape("2026-07-31T23:59:59Z")
	var ranged []map[string]any
	mustDecode(t, mustGet(t, fmt.Sprintf("%s?from=%s&to=%s", base, from, to)), &ranged)

	assert.Len(t, all, 1)
	assert.Len(t, ranged, len(all), "every stored event must be visible in the calendar range")
}

func TestEvent_AcceptsZeroLengthEventAndDefaultsColour(t *testing.T) {
	srv := newTestServer(t)

	// A reminder marks an instant, so equal timestamps are valid. The
	// repository writes colour explicitly, so the SQL default never fires
	// and Normalize has to supply it.
	resp := mustPost(t, srv.URL+"/api/v1/events",
		`{"title":"  Reminder  ","start_at":"2026-07-21T09:00:00Z","end_at":"2026-07-21T09:00:00Z"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	mustDecode(t, resp, &created)
	assert.Equal(t, "Reminder", created["title"])
	assert.Equal(t, "#6366f1", created["color"])
}
