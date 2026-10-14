package email

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/husari/hube/internal/domain/dateonly"
	domaintask "github.com/husari/hube/internal/domain/task"
)

// Sender delivers a composed message. *external.SMTPClient satisfies it
// implicitly. Defining the interface here (where it is consumed) keeps the
// digest logic independent of the transport.
type Sender interface {
	Send(to []string, subject, body string) error
}

// TaskLister supplies the tasks a digest is built from. *task.Service
// satisfies it implicitly.
type TaskLister interface {
	List(ctx context.Context) ([]domaintask.Task, error)
}

type Service struct {
	smtp    Sender
	taskSvc TaskLister

	// now supplies the instant the digest is built against. It is a field
	// so tests can pin it: the overdue / due today / upcoming split is
	// decided by the calendar day, which makes any test written against
	// the real clock fail when it runs near midnight.
	now func() time.Time
}

func NewService(smtp Sender, taskSvc TaskLister) *Service {
	return &Service{smtp: smtp, taskSvc: taskSvc, now: time.Now}
}

type DigestOptions struct {
	To []string
}

func (s *Service) SendDigest(ctx context.Context, opts DigestOptions) error {
	tasks, err := s.taskSvc.List(ctx)
	if err != nil {
		return fmt.Errorf("email digest: list tasks: %w", err)
	}

	now := s.now()
	// Due dates are bare calendar days, so the digest buckets by day: a task
	// due before today is overdue regardless of the hour.
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	var pending, overdue, dueToday []string

	for _, t := range tasks {
		if t.Status == domaintask.StatusDone {
			continue
		}
		due, ok := parseDueDate(t.DueDate)
		if !ok {
			// No due date (or an unreadable one) has nowhere else to go.
			pending = append(pending, fmt.Sprintf("  - [%s] %s", t.Priority, t.Title))
			continue
		}
		switch {
		case due.Before(today):
			overdue = append(overdue, fmt.Sprintf("  - [%s] %s (due %s)", t.Priority, t.Title, due.Format("Jan 2")))
		case due.Equal(today):
			dueToday = append(dueToday, fmt.Sprintf("  - [%s] %s", t.Priority, t.Title))
		default:
			pending = append(pending, fmt.Sprintf("  - [%s] %s (due %s)", t.Priority, t.Title, due.Format("Jan 2")))
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Hube Daily Digest — %s\n", now.Format("Monday, January 2, 2006")))
	sb.WriteString(strings.Repeat("─", 50) + "\n\n")

	if len(dueToday) > 0 {
		sb.WriteString(fmt.Sprintf("DUE TODAY (%d)\n", len(dueToday)))
		sb.WriteString(strings.Join(dueToday, "\n") + "\n\n")
	}
	if len(overdue) > 0 {
		sb.WriteString(fmt.Sprintf("OVERDUE (%d)\n", len(overdue)))
		sb.WriteString(strings.Join(overdue, "\n") + "\n\n")
	}
	if len(pending) > 0 {
		sb.WriteString(fmt.Sprintf("UPCOMING (%d)\n", len(pending)))
		sb.WriteString(strings.Join(pending, "\n") + "\n\n")
	}
	if len(dueToday)+len(overdue)+len(pending) == 0 {
		sb.WriteString("No pending tasks. All caught up!\n")
	}

	subject := fmt.Sprintf("Hube Digest — %s", now.Format("Jan 2"))
	return s.smtp.Send(opts.To, subject, sb.String())
}

// parseDueDate reads a task's "2006-01-02" due date. ok is false when the
// task has no due date or the stored value is not a readable day.
func parseDueDate(due *string) (time.Time, bool) {
	if due == nil {
		return time.Time{}, false
	}
	d, err := time.Parse(dateonly.Layout, *due)
	if err != nil {
		return time.Time{}, false
	}
	return d, true
}
