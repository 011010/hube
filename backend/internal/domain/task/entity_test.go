package task

import (
	"errors"
	"testing"
	"time"
)

func TestTask_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    Task
		expected Task
	}{
		{
			name:     "defaults empty status and priority",
			input:    Task{Title: "Ship it"},
			expected: Task{Title: "Ship it", Status: StatusTodo, Priority: PriorityMedium},
		},
		{
			name:     "preserves existing status and priority",
			input:    Task{Title: "Ship it", Status: StatusDone, Priority: PriorityHigh},
			expected: Task{Title: "Ship it", Status: StatusDone, Priority: PriorityHigh},
		},
		{
			name:     "defaults only empty status",
			input:    Task{Title: "Ship it", Priority: PriorityLow},
			expected: Task{Title: "Ship it", Status: StatusTodo, Priority: PriorityLow},
		},
		{
			name:     "trims whitespace from title",
			input:    Task{Title: "  Ship it  ", Status: StatusTodo, Priority: PriorityMedium},
			expected: Task{Title: "Ship it", Status: StatusTodo, Priority: PriorityMedium},
		},
		{
			name:     "trims recurrence",
			input:    Task{Title: "Ship it", Recurrence: " weekly ", Status: StatusTodo, Priority: PriorityMedium},
			expected: Task{Title: "Ship it", Recurrence: RecurrenceWeekly, Status: StatusTodo, Priority: PriorityMedium},
		},
		{
			name:     "a whitespace-only title collapses to empty so Validate can reject it",
			input:    Task{Title: "   ", Status: StatusTodo, Priority: PriorityMedium},
			expected: Task{Title: "", Status: StatusTodo, Priority: PriorityMedium},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input
			got.Normalize()
			if got != tc.expected {
				t.Errorf("got %+v, want %+v", got, tc.expected)
			}
		})
	}
}

func TestTask_Validate(t *testing.T) {
	valid := func(mut func(*Task)) Task {
		tk := Task{Title: "Ship it", Status: StatusTodo, Priority: PriorityMedium}
		if mut != nil {
			mut(&tk)
		}
		return tk
	}

	tests := []struct {
		name      string
		task      Task
		wantField string // empty means the task must validate
	}{
		{name: "a fully specified task is valid", task: valid(nil)},
		{
			name: "every recurrence value the scheduler understands is accepted",
			task: valid(func(tk *Task) { tk.Recurrence = RecurrenceMonthly }),
		},
		{
			name: "no recurrence is valid",
			task: valid(func(tk *Task) { tk.Recurrence = RecurrenceNone }),
		},
		{
			name:      "empty title is rejected",
			task:      valid(func(tk *Task) { tk.Title = "" }),
			wantField: "title",
		},
		{
			name:      "unknown status is rejected",
			task:      valid(func(tk *Task) { tk.Status = "not-a-status" }),
			wantField: "status",
		},
		{
			name:      "unknown priority is rejected",
			task:      valid(func(tk *Task) { tk.Priority = "banana" }),
			wantField: "priority",
		},
		{
			// The scheduler silently skips recurrence values it does not
			// recognise, so an unchecked typo means a task that simply
			// never repeats and never reports why.
			name:      "unknown recurrence is rejected",
			task:      valid(func(tk *Task) { tk.Recurrence = "fortnightly" }),
			wantField: "recurrence",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.task.Validate()

			if tc.wantField == "" {
				if err != nil {
					t.Fatalf("expected the task to be valid, got %v", err)
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

func TestTask_ValidateAcceptsEveryDeclaredConstant(t *testing.T) {
	// Guards against a constant being added to the type without being
	// added to the corresponding Validate switch.
	for _, s := range []Status{StatusTodo, StatusInProgress, StatusDone} {
		tk := Task{Title: "Ship it", Status: s, Priority: PriorityMedium}
		if err := tk.Validate(); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
	for _, p := range []Priority{PriorityLow, PriorityMedium, PriorityHigh} {
		tk := Task{Title: "Ship it", Status: StatusTodo, Priority: p}
		if err := tk.Validate(); err != nil {
			t.Errorf("priority %q should be valid: %v", p, err)
		}
	}
	for _, r := range []string{RecurrenceNone, RecurrenceDaily, RecurrenceWeekly, RecurrenceMonthly} {
		tk := Task{Title: "Ship it", Status: StatusTodo, Priority: PriorityMedium, Recurrence: r}
		if err := tk.Validate(); err != nil {
			t.Errorf("recurrence %q should be valid: %v", r, err)
		}
	}
}

func TestTask_NormalizeThenValidateRejectsWhitespaceTitle(t *testing.T) {
	tk := Task{Title: "\t \n", Priority: PriorityMedium}
	tk.Normalize()

	var valErr *ValidationError
	if err := tk.Validate(); !errors.As(err, &valErr) || valErr.Field != "title" {
		t.Fatalf("expected a title validation error, got %v", err)
	}
}

func TestValidationError_Message(t *testing.T) {
	err := &ValidationError{Field: "title", Message: "title is required"}
	if got, want := err.Error(), "title: title is required"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Compile-time reminder that DueDate is a *time.Time here, unlike the string
// dates other entities carry. JSON decoding rejects malformed values before
// Validate ever runs, so there is nothing to check for it.
var _ = Task{DueDate: new(time.Time)}
