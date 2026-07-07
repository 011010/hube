package note

import (
	"errors"
	"testing"
)

func TestNote_Normalize(t *testing.T) {
	emptyDate := ""
	tests := []struct {
		name     string
		input    Note
		expected Note
	}{
		{
			name:     "defaults empty status and priority",
			input:    Note{Title: "Test"},
			expected: Note{Title: "Test", Status: StatusDraft, Priority: PriorityMedium},
		},
		{
			name:     "preserves existing status and priority",
			input:    Note{Title: "Test", Status: StatusPublished, Priority: PriorityHigh},
			expected: Note{Title: "Test", Status: StatusPublished, Priority: PriorityHigh},
		},
		{
			name:     "defaults only empty status",
			input:    Note{Title: "Test", Priority: PriorityLow},
			expected: Note{Title: "Test", Status: StatusDraft, Priority: PriorityLow},
		},
		{
			name:     "defaults only empty priority",
			input:    Note{Title: "Test", Status: StatusInProgress},
			expected: Note{Title: "Test", Status: StatusInProgress, Priority: PriorityMedium},
		},
		{
			name:     "trims whitespace from title",
			input:    Note{Title: "  Test Note  ", Status: StatusDraft, Priority: PriorityMedium},
			expected: Note{Title: "Test Note", Status: StatusDraft, Priority: PriorityMedium},
		},
		{
			name:     "normalizes empty due_date to nil",
			input:    Note{Title: "Test", Status: StatusDraft, Priority: PriorityMedium, DueDate: &emptyDate},
			expected: Note{Title: "Test", Status: StatusDraft, Priority: PriorityMedium, DueDate: nil},
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
			if tt.input.Priority != tt.expected.Priority {
				t.Errorf("Priority = %q, want %q", tt.input.Priority, tt.expected.Priority)
			}
			if (tt.input.DueDate == nil) != (tt.expected.DueDate == nil) {
				t.Errorf("DueDate nil mismatch")
			}
			if tt.input.DueDate != nil && tt.expected.DueDate != nil && *tt.input.DueDate != *tt.expected.DueDate {
				t.Errorf("DueDate = %q, want %q", *tt.input.DueDate, *tt.expected.DueDate)
			}
		})
	}
}

func TestNote_Validate(t *testing.T) {
	validDate := "2026-07-07"
	tests := []struct {
		name    string
		note    Note
		wantErr string
		wantField string
	}{
		{
			name:    "empty title rejected",
			note:    Note{Status: StatusDraft, Priority: PriorityMedium},
			wantErr: "title is required",
			wantField: "title",
		},
		{
			name:    "invalid status rejected",
			note:    Note{Title: "Test", Status: "invalid", Priority: PriorityMedium},
			wantErr: "invalid status",
			wantField: "status",
		},
		{
			name:    "invalid priority rejected",
			note:    Note{Title: "Test", Status: StatusDraft, Priority: "invalid"},
			wantErr: "invalid priority",
			wantField: "priority",
		},
		{
			name:    "invalid due_date rejected",
			note:    Note{Title: "Test", Status: StatusDraft, Priority: PriorityMedium, DueDate: strPtr("not-a-date")},
			wantErr: "invalid due_date",
			wantField: "due_date",
		},
		{
			name: "valid note accepted",
			note: Note{Title: "Test", Status: StatusDraft, Priority: PriorityMedium, DueDate: &validDate},
		},
		{
			name: "all valid statuses accepted",
			note: Note{Title: "Test", Status: StatusInProgress, Priority: PriorityHigh},
		},
		{
			name: "all valid priorities accepted",
			note: Note{Title: "Test", Status: StatusPublished, Priority: PriorityLow},
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

func strPtr(s string) *string {
	return &s
}
