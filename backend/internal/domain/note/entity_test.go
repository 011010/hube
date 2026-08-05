package note

import (
	"errors"
	"testing"
)

func TestNote_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    Note
		expected Note
	}{
		{
			name:     "defaults empty status",
			input:    Note{Title: "Test"},
			expected: Note{Title: "Test", Status: StatusDraft},
		},
		{
			name:     "preserves existing status",
			input:    Note{Title: "Test", Status: StatusPublished},
			expected: Note{Title: "Test", Status: StatusPublished},
		},
		{
			name:     "trims whitespace from title",
			input:    Note{Title: "  Test Note  ", Status: StatusDraft},
			expected: Note{Title: "Test Note", Status: StatusDraft},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Normalize()
			if tt.input.Title != tt.expected.Title {
				t.Errorf("Title = %q, want %q", tt.input.Title, tt.expected.Title)
			}
			if tt.input.Status != tt.expected.Status {
				t.Errorf("Status = %q, want %q", tt.input.Status, tt.expected.Status)
			}
		})
	}
}

func TestNote_Validate(t *testing.T) {
	tests := []struct {
		name      string
		note      Note
		wantErr   string
		wantField string
	}{
		{
			name:      "empty title rejected",
			note:      Note{Status: StatusDraft},
			wantErr:   "title is required",
			wantField: "title",
		},
		{
			name:      "invalid status rejected",
			note:      Note{Title: "Test", Status: "invalid"},
			wantErr:   "invalid status",
			wantField: "status",
		},
		{
			name: "valid note accepted",
			note: Note{Title: "Test", Status: StatusDraft},
		},
		{
			name: "in_progress status accepted",
			note: Note{Title: "Test", Status: StatusInProgress},
		},
		{
			name: "published status accepted",
			note: Note{Title: "Test", Status: StatusPublished},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.note.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Errorf("Validate() error = nil, want %q", tt.wantErr)
				return
			}
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Errorf("Validate() error type = %T, want *ValidationError", err)
				return
			}
			if valErr.Field != tt.wantField {
				t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, tt.wantField)
			}
			if valErr.Message == "" {
				t.Errorf("ValidationError.Message should not be empty")
			}
		})
	}
}
