package email

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	var pending, overdue, dueToday []string

	for _, t := range tasks {
		if t.Status == domaintask.StatusDone {
			continue
		}
		if t.DueDate != nil {
			due := *t.DueDate
			if due.Before(now) {
				overdue = append(overdue, fmt.Sprintf("  - [%s] %s (due %s)", t.Priority, t.Title, due.Format("Jan 2")))
			} else if due.Year() == now.Year() && due.YearDay() == now.YearDay() {
				dueToday = append(dueToday, fmt.Sprintf("  - [%s] %s", t.Priority, t.Title))
			} else {
				pending = append(pending, fmt.Sprintf("  - [%s] %s (due %s)", t.Priority, t.Title, due.Format("Jan 2")))
			}
		} else {
			pending = append(pending, fmt.Sprintf("  - [%s] %s", t.Priority, t.Title))
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
