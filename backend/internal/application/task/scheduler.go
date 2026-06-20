package task

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/husari/hube/internal/domain/task"
)

// Scheduler watches done recurring tasks and resets them when their period elapses.
type Scheduler struct {
	repo task.Repository
}

func NewScheduler(repo task.Repository) *Scheduler {
	return &Scheduler{repo: repo}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	s.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	tasks, err := s.repo.FindRecurring(ctx)
	if err != nil {
		log.Printf("scheduler: find recurring: %v", err)
		return
	}
	now := time.Now()
	for _, t := range tasks {
		if !s.isDue(t, now) {
			continue
		}
		next := task.Task{
			ID:          uuid.New().String(),
			Title:       t.Title,
			Description: t.Description,
			Priority:    t.Priority,
			Status:      task.StatusTodo,
			ProjectID:   t.ProjectID,
			Recurrence:  t.Recurrence,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.repo.Create(ctx, &next); err != nil {
			log.Printf("scheduler: create recurrence for %q: %v", t.Title, err)
			continue
		}
		t.LastRecurredAt = &now
		if err := s.repo.Update(ctx, &t); err != nil {
			log.Printf("scheduler: update last_recurred_at for %q: %v", t.Title, err)
		}
	}
}

func (s *Scheduler) isDue(t task.Task, now time.Time) bool {
	ref := t.LastRecurredAt
	if ref == nil {
		ref = &t.UpdatedAt
	}
	switch t.Recurrence {
	case "daily":
		return now.Sub(*ref) >= 24*time.Hour
	case "weekly":
		return now.Sub(*ref) >= 7*24*time.Hour
	case "monthly":
		next := ref.AddDate(0, 1, 0)
		return now.After(next)
	}
	return false
}
