package email

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domaintask "github.com/husari/hube/internal/domain/task"
)

type sentMail struct {
	to      []string
	subject string
	body    string
}

type fakeSender struct {
	sent []sentMail
	err  error
}

func (f *fakeSender) Send(to []string, subject, body string) error {
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, sentMail{to: to, subject: subject, body: body})
	return nil
}

type fakeTasks struct {
	tasks []domaintask.Task
	err   error
}

func (f *fakeTasks) List(context.Context) ([]domaintask.Task, error) {
	return f.tasks, f.err
}

// refNow is the instant every digest test is built against: mid-afternoon, so
// there is room on both sides of it inside the same calendar day.
var refNow = time.Date(2026, 7, 21, 15, 0, 0, 0, time.UTC)

// day returns the calendar day offset from refNow, in the "2006-01-02" form
// the task entity carries due dates in.
func day(offset int) *string {
	d := refNow.AddDate(0, 0, offset).Format("2006-01-02")
	return &d
}

func newTask(title string, status domaintask.Status, due *string) domaintask.Task {
	return domaintask.Task{
		Title:    title,
		Status:   status,
		Priority: domaintask.PriorityMedium,
		DueDate:  due,
	}
}

// digest runs SendDigest against the given tasks and returns the body of the
// single message that was sent.
func digest(t *testing.T, tasks []domaintask.Task) (string, *fakeSender) {
	t.Helper()
	sender := &fakeSender{}
	svc := NewService(sender, &fakeTasks{tasks: tasks})
	svc.now = func() time.Time { return refNow }

	if err := svc.SendDigest(context.Background(), DigestOptions{To: []string{"me@example.com"}}); err != nil {
		t.Fatalf("SendDigest: %v", err)
	}
	if len(sender.sent) != 1 {
		t.Fatalf("expected exactly one message, got %d", len(sender.sent))
	}
	return sender.sent[0].body, sender
}

func TestDigestSkipsCompletedTasks(t *testing.T) {
	body, _ := digest(t, []domaintask.Task{
		newTask("shipped already", domaintask.StatusDone, day(-2)),
	})

	if strings.Contains(body, "shipped already") {
		t.Error("a done task leaked into the digest")
	}
	// A digest of only-done tasks is an empty digest, not a broken one.
	if !strings.Contains(body, "All caught up!") {
		t.Errorf("expected the empty-state message, got:\n%s", body)
	}
}

func TestDigestClassifiesTasksByDueDate(t *testing.T) {
	body, _ := digest(t, []domaintask.Task{
		newTask("late thing", domaintask.StatusTodo, day(-2)),
		newTask("today thing", domaintask.StatusTodo, day(0)),
		newTask("later thing", domaintask.StatusTodo, day(5)),
		newTask("someday thing", domaintask.StatusTodo, nil),
	})

	sections := map[string]string{
		"DUE TODAY (1)": "today thing",
		"OVERDUE (1)":   "late thing",
		"UPCOMING (2)":  "later thing",
	}
	for header, task := range sections {
		if !strings.Contains(body, header) {
			t.Errorf("missing section %q in:\n%s", header, body)
		}
		if !strings.Contains(body, task) {
			t.Errorf("missing task %q in:\n%s", task, body)
		}
	}
	// A task with no due date has nowhere else to go but UPCOMING.
	if !strings.Contains(body, "someday thing") {
		t.Errorf("task without a due date was dropped:\n%s", body)
	}
}

func TestDigestBucketsByCalendarDay(t *testing.T) {
	// A due date is a calendar day, not an instant: a task due today is DUE
	// TODAY for the whole day, and only a date that has already passed is
	// OVERDUE. Pinning this because the two branches are easy to swap.
	tests := []struct {
		name    string
		due     *string
		section string
	}{
		{"yesterday", day(-1), "OVERDUE (1)"},
		{"today", day(0), "DUE TODAY (1)"},
		{"tomorrow", day(1), "UPCOMING (1)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := digest(t, []domaintask.Task{
				newTask("boundary", domaintask.StatusTodo, tc.due),
			})
			if !strings.Contains(body, tc.section) {
				t.Errorf("expected %q, got:\n%s", tc.section, body)
			}
		})
	}
}

func TestDigestTreatsUnreadableDueDateAsPending(t *testing.T) {
	// A due date that is not a bare calendar day must not crash or vanish;
	// it falls back to the undated bucket rather than being dropped.
	body, _ := digest(t, []domaintask.Task{
		newTask("garbled", domaintask.StatusTodo, strPtr("not-a-date")),
	})

	if !strings.Contains(body, "garbled") || !strings.Contains(body, "UPCOMING (1)") {
		t.Errorf("expected the task in UPCOMING, got:\n%s", body)
	}
}

func strPtr(s string) *string { return &s }

func TestDigestIsBuiltAgainstTheInjectedClock(t *testing.T) {
	// Guards the wiring itself: if SendDigest went back to calling
	// time.Now directly, the header would not read the pinned date.
	body, sender := digest(t, nil)

	if !strings.Contains(body, "Tuesday, July 21, 2026") {
		t.Errorf("header should reflect the pinned clock, got:\n%s", body)
	}
	if got := sender.sent[0].subject; got != "Hube Digest — Jul 21" {
		t.Errorf("subject: got %q", got)
	}
}

func TestDigestReportsEmptyStateWhenNoTasks(t *testing.T) {
	body, _ := digest(t, nil)

	if !strings.Contains(body, "All caught up!") {
		t.Errorf("expected the empty-state message, got:\n%s", body)
	}
}

func TestDigestAddressesAndSubjectsTheMessage(t *testing.T) {
	_, sender := digest(t, nil)

	msg := sender.sent[0]
	if len(msg.to) != 1 || msg.to[0] != "me@example.com" {
		t.Errorf("recipients: got %v", msg.to)
	}
	if !strings.HasPrefix(msg.subject, "Hube Digest — ") {
		t.Errorf("subject: got %q", msg.subject)
	}
}

func TestSendDigestFailsWhenTasksCannotBeListed(t *testing.T) {
	sender := &fakeSender{}
	svc := NewService(sender, &fakeTasks{err: errors.New("db down")})

	err := svc.SendDigest(context.Background(), DigestOptions{To: []string{"me@example.com"}})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	// Nothing should be sent if we could not read the tasks.
	if len(sender.sent) != 0 {
		t.Errorf("sent %d messages despite the failure", len(sender.sent))
	}
}
