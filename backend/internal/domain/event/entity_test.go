package event

import (
	"errors"
	"testing"
	"time"
)

var (
	start = time.Date(2026, 7, 21, 9, 0, 0, 0, time.UTC)
	end   = time.Date(2026, 7, 21, 10, 0, 0, 0, time.UTC)
)

func TestEvent_Normalize(t *testing.T) {
	tests := []struct {
		name      string
		input     Event
		wantTitle string
		wantColor string
	}{
		{
			name:      "defaults an empty colour",
			input:     Event{Title: "Standup"},
			wantTitle: "Standup",
			wantColor: defaultColor,
		},
		{
			name:      "preserves an explicit colour",
			input:     Event{Title: "Standup", Color: "#ff0000"},
			wantTitle: "Standup",
			wantColor: "#ff0000",
		},
		{
			name:      "trims whitespace from the title",
			input:     Event{Title: "  Standup  ", Color: "#ff0000"},
			wantTitle: "Standup",
			wantColor: "#ff0000",
		},
		{
			name:      "a whitespace-only title collapses to empty so Validate can reject it",
			input:     Event{Title: "\t\n", Color: "#ff0000"},
			wantTitle: "",
			wantColor: "#ff0000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input
			got.Normalize()
			if got.Title != tc.wantTitle {
				t.Errorf("title: got %q, want %q", got.Title, tc.wantTitle)
			}
			if got.Color != tc.wantColor {
				t.Errorf("colour: got %q, want %q", got.Color, tc.wantColor)
			}
		})
	}
}

func TestEvent_Validate(t *testing.T) {
	valid := func(mut func(*Event)) Event {
		e := Event{Title: "Standup", StartAt: start, EndAt: end, Color: defaultColor}
		if mut != nil {
			mut(&e)
		}
		return e
	}

	tests := []struct {
		name      string
		event     Event
		wantField string // empty means the event must validate
	}{
		{name: "a fully specified event is valid", event: valid(nil)},
		{
			// A reminder is a real use case: it marks an instant, not a
			// span, so equal timestamps must be accepted.
			name:  "a zero-length event is valid",
			event: valid(func(e *Event) { e.EndAt = e.StartAt }),
		},
		{
			name:      "empty title is rejected",
			event:     valid(func(e *Event) { e.Title = "" }),
			wantField: "title",
		},
		{
			// Omitting start_at leaves the Go zero value, which lands the
			// event in year 1 where no calendar range will ever find it.
			name:      "a missing start is rejected",
			event:     valid(func(e *Event) { e.StartAt = time.Time{} }),
			wantField: "start_at",
		},
		{
			name:      "a missing end is rejected",
			event:     valid(func(e *Event) { e.EndAt = time.Time{} }),
			wantField: "end_at",
		},
		{
			name:      "an end before the start is rejected",
			event:     valid(func(e *Event) { e.EndAt = e.StartAt.Add(-time.Hour) }),
			wantField: "end_at",
		},
		{
			name:      "an end one nanosecond before the start is rejected",
			event:     valid(func(e *Event) { e.EndAt = e.StartAt.Add(-time.Nanosecond) }),
			wantField: "end_at",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.event.Validate()

			if tc.wantField == "" {
				if err != nil {
					t.Fatalf("expected the event to be valid, got %v", err)
				}
				return
			}

			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("expected a *ValidationError, got %v", err)
			}
			if valErr.Field != tc.wantField {
				t.Errorf("field: got %q, want %q", valErr.Field, tc.wantField)
			}
		})
	}
}

func TestEvent_ValidateComparesInstantsNotWallClocks(t *testing.T) {
	// 09:30 in Buenos Aires (UTC-3) is 12:30 UTC, so this end is genuinely
	// after the start even though its wall-clock reading looks earlier.
	// Comparing formatted strings instead of instants would reject it.
	ba := time.FixedZone("ART", -3*60*60)

	e := Event{
		Title:   "Standup",
		StartAt: time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC),
		EndAt:   time.Date(2026, 7, 21, 9, 30, 0, 0, ba),
	}
	if err := e.Validate(); err != nil {
		t.Fatalf("expected the event to be valid across time zones, got %v", err)
	}
}

func TestValidationError_Message(t *testing.T) {
	err := &ValidationError{Field: "end_at", Message: "end_at must not be before start_at"}
	if got, want := err.Error(), "end_at: end_at must not be before start_at"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
