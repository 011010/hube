package note

import (
	"strings"
	"testing"
)

func TestNote_Normalize(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Normalize()
			if tt.input.Status != tt.expected.Status {
				t.Errorf("Status = %q, want %q", tt.input.Status, tt.expected.Status)
			}
			if tt.input.Priority != tt.expected.Priority {
				t.Errorf("Priority = %q, want %q", tt.input.Priority, tt.expected.Priority)
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
	}{
		{
			name:    "empty title rejected",
			note:    Note{Status: StatusDraft, Priority: PriorityMedium},
			wantErr: "title is required",
		},
		{
			name:    "invalid status rejected",
			note:    Note{Title: "Test", Status: "invalid", Priority: PriorityMedium},
			wantErr: "invalid status: invalid",
		},
		{
			name:    "invalid priority rejected",
			note:    Note{Title: "Test", Status: StatusDraft, Priority: "invalid"},
			wantErr: "invalid priority: invalid",
		},
		{
			name:    "invalid due_date rejected",
			note:    Note{Title: "Test", Status: StatusDraft, Priority: PriorityMedium, DueDate: strPtr("not-a-date")},
			wantErr: "invalid due_date",
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
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
