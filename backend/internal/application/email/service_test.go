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

func at(d time.Duration) *time.Time {
	t := time.Now().Add(d)
	return &t
}

func newTask(title string, status domaintask.Status, due *time.Time) domaintask.Task {
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
		newTask("shipped already", domaintask.StatusDone, at(-48*time.Hour)),
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
		newTask("late thing", domaintask.StatusTodo, at(-48*time.Hour)),
		newTask("today thing", domaintask.StatusTodo, at(2*time.Hour)),
		newTask("later thing", domaintask.StatusTodo, at(120*time.Hour)),
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

func TestDigestCountsEarlierTodayAsOverdue(t *testing.T) {
	// The cutoff is the current instant, not the calendar day: a task due
	// at 09:00 is overdue by 15:00. DUE TODAY means "still ahead of you
	// today". Pinning this because the two branches are easy to swap.
	body, _ := digest(t, []domaintask.Task{
		newTask("this morning", domaintask.StatusTodo, at(-1*time.Hour)),
	})

	if !strings.Contains(body, "OVERDUE (1)") {
		t.Errorf("expected OVERDUE, got:\n%s", body)
	}
	if strings.Contains(body, "DUE TODAY") {
		t.Errorf("a task already past its time must not be DUE TODAY:\n%s", body)
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
