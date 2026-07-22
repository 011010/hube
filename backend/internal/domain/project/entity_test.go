package project

import (
	"errors"
	"testing"
)

func TestProject_Normalize(t *testing.T) {
	empty := ""
	date := "2026-07-21"

	tests := []struct {
		name     string
		input    Project
		expected Project
	}{
		{
			name:     "defaults empty status and colour",
			input:    Project{Name: "Hube"},
			expected: Project{Name: "Hube", Status: StatusPlanning, Color: defaultColor},
		},
		{
			name:     "preserves an explicit status and colour",
			input:    Project{Name: "Hube", Status: StatusActive, Color: "#ff0000"},
			expected: Project{Name: "Hube", Status: StatusActive, Color: "#ff0000"},
		},
		{
			name:     "trims whitespace from the name",
			input:    Project{Name: "  Hube  ", Status: StatusActive, Color: "#ff0000"},
			expected: Project{Name: "Hube", Status: StatusActive, Color: "#ff0000"},
		},
		{
			name:     "a whitespace-only name collapses to empty so Validate can reject it",
			input:    Project{Name: "\t\n", Status: StatusActive, Color: "#ff0000"},
			expected: Project{Name: "", Status: StatusActive, Color: "#ff0000"},
		},
		{
			// An empty string is what an untouched date input submits; it
			// must become NULL rather than fail date parsing.
			name:     "an empty due date becomes nil",
			input:    Project{Name: "Hube", Status: StatusActive, Color: "#ff0000", DueDate: &empty},
			expected: Project{Name: "Hube", Status: StatusActive, Color: "#ff0000", DueDate: nil},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input
			got.Normalize()

			if got.Name != tc.expected.Name || got.Status != tc.expected.Status || got.Color != tc.expected.Color {
				t.Errorf("got %+v, want %+v", got, tc.expected)
			}
			if (got.DueDate == nil) != (tc.expected.DueDate == nil) {
				t.Errorf("due date: got %v, want %v", got.DueDate, tc.expected.DueDate)
			}
		})
	}

	t.Run("preserves a real due date", func(t *testing.T) {
		p := Project{Name: "Hube", DueDate: &date}
		p.Normalize()
		if p.DueDate == nil || *p.DueDate != date {
			t.Errorf("got %v, want %q", p.DueDate, date)
		}
	})
}

func TestProject_Validate(t *testing.T) {
	bad := "21/07/2026"
	good := "2026-07-21"

	valid := func(mut func(*Project)) Project {
		p := Project{Name: "Hube", Status: StatusPlanning, Color: defaultColor}
		if mut != nil {
			mut(&p)
		}
		return p
	}

	tests := []struct {
		name      string
		project   Project
		wantField string // empty means the project must validate
	}{
		{name: "a fully specified project is valid", project: valid(nil)},
		{name: "a valid due date is accepted", project: valid(func(p *Project) { p.DueDate = &good })},
		{name: "no due date is valid", project: valid(func(p *Project) { p.DueDate = nil })},
		{
			name:      "empty name is rejected",
			project:   valid(func(p *Project) { p.Name = "" }),
			wantField: "name",
		},
		{
			name:      "unknown status is rejected",
			project:   valid(func(p *Project) { p.Status = "abandoned" }),
			wantField: "status",
		},
		{
			name:      "a non ISO due date is rejected",
			project:   valid(func(p *Project) { p.DueDate = &bad }),
			wantField: "due_date",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.project.Validate()

			if tc.wantField == "" {
				if err != nil {
					t.Fatalf("expected the project to be valid, got %v", err)
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

func TestProject_ValidateAcceptsEveryDeclaredStatus(t *testing.T) {
	// Guards against a status being added to the type without being added
	// to the Validate switch.
	for _, s := range []Status{StatusPlanning, StatusActive, StatusCompleted, StatusOnHold} {
		p := Project{Name: "Hube", Status: s}
		if err := p.Validate(); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
}

func TestProject_Progress(t *testing.T) {
	// TaskCount and CompletedCount come from a single SQL aggregate
	// (COUNT of joined tasks, SUM of the done ones), so completed never
	// exceeds total. These cases are the reachable ones.
	tests := []struct {
		name      string
		total     int
		completed int
		want      int
	}{
		{"a project with no tasks reports zero, not a division by zero", 0, 0, 0},
		{"nothing done yet", 4, 0, 0},
		{"everything done", 4, 4, 100},
		{"half done", 4, 2, 50},
		{"truncates rather than rounding", 3, 2, 66},
		{"a single task is all or nothing", 1, 1, 100},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Project{TaskCount: tc.total, CompletedCount: tc.completed}
			if got := p.Progress(); got != tc.want {
				t.Errorf("got %d%%, want %d%%", got, tc.want)
			}
		})
	}
}

func TestValidationError_Message(t *testing.T) {
	err := &ValidationError{Field: "name", Message: "name is required"}
	if got, want := err.Error(), "name: name is required"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
