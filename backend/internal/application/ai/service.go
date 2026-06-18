package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appapp "github.com/husari/hube/internal/application/app"
	appevent "github.com/husari/hube/internal/application/event"
	appnote "github.com/husari/hube/internal/application/note"
	appproject "github.com/husari/hube/internal/application/project"
	apptask "github.com/husari/hube/internal/application/task"
	aidom "github.com/husari/hube/internal/domain/ai"
	"github.com/husari/hube/internal/domain/note"
	"github.com/husari/hube/internal/domain/task"
)

// HubExecutor implements external.ToolExecutor using the hub's services.
type HubExecutor struct {
	tasks    *apptask.Service
	notes    *appnote.Service
	projects *appproject.Service
	events   *appevent.Service
	apps     *appapp.Service
}

func NewHubExecutor(
	tasks *apptask.Service,
	notes *appnote.Service,
	projects *appproject.Service,
	events *appevent.Service,
	apps *appapp.Service,
) *HubExecutor {
	return &HubExecutor{tasks, notes, projects, events, apps}
}

// Tools returns provider-agnostic tool definitions.
func (e *HubExecutor) Tools() []aidom.ToolDef {
	return []aidom.ToolDef{
		{
			Name:        "list_tasks",
			Description: "List tasks from the hub. Returns all tasks or those filtered by status.",
			Properties: map[string]any{
				"status": map[string]any{
					"type":        "string",
					"enum":        []string{"todo", "in_progress", "done"},
					"description": "Filter by task status",
				},
			},
		},
		{
			Name:        "create_task",
			Description: "Create a new task in the hub.",
			Properties: map[string]any{
				"title":       map[string]any{"type": "string", "description": "Task title (required)"},
				"description": map[string]any{"type": "string", "description": "Task description"},
				"priority":    map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}, "description": "Priority level"},
				"due_date":    map[string]any{"type": "string", "description": "Due date ISO 8601 (YYYY-MM-DD)"},
				"project_id":  map[string]any{"type": "string", "description": "Associate with a project ID"},
			},
			Required: []string{"title"},
		},
		{
			Name:        "update_task",
			Description: "Update an existing task's status, title, or priority.",
			Properties: map[string]any{
				"id":       map[string]any{"type": "string", "description": "Task ID (required)"},
				"title":    map[string]any{"type": "string", "description": "New title"},
				"status":   map[string]any{"type": "string", "enum": []string{"todo", "in_progress", "done"}, "description": "New status"},
				"priority": map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}, "description": "New priority"},
			},
			Required: []string{"id"},
		},
		{
			Name:        "search_notes",
			Description: "Full-text search across all notes in the hub.",
			Properties:  map[string]any{"query": map[string]any{"type": "string", "description": "Search query (required)"}},
			Required:    []string{"query"},
		},
		{
			Name:        "list_notes",
			Description: "List all notes, optionally filtered by folder.",
			Properties:  map[string]any{"folder_id": map[string]any{"type": "string", "description": "Filter by folder ID"}},
		},
		{
			Name:        "create_note",
			Description: "Create a new note in the hub.",
			Properties: map[string]any{
				"title":   map[string]any{"type": "string", "description": "Note title (required)"},
				"content": map[string]any{"type": "string", "description": "Note content in Markdown"},
			},
			Required: []string{"title"},
		},
		{
			Name:        "list_projects",
			Description: "List all projects with progress, status, and task counts.",
			Properties:  map[string]any{},
		},
		{
			Name:        "list_events",
			Description: "List calendar events. Defaults to the next 30 days if no range given.",
			Properties: map[string]any{
				"from": map[string]any{"type": "string", "description": "Start date ISO 8601 (YYYY-MM-DD). Defaults to today."},
				"to":   map[string]any{"type": "string", "description": "End date ISO 8601 (YYYY-MM-DD). Defaults to 30 days from now."},
			},
		},
		{
			Name:        "list_apps",
			Description: "List all installed apps/shortcuts in the hub launcher.",
			Properties:  map[string]any{},
		},
	}
}

// Execute dispatches a tool call to the appropriate service.
func (e *HubExecutor) Execute(ctx context.Context, name string, input json.RawMessage) (any, error) {
	var args map[string]any
	if err := json.Unmarshal(input, &args); err != nil {
		return nil, fmt.Errorf("bad tool input: %w", err)
	}
	str := func(key string) string {
		if v, ok := args[key].(string); ok {
			return v
		}
		return ""
	}
	strPtr := func(key string) *string {
		if v, ok := args[key].(string); ok {
			return &v
		}
		return nil
	}

	switch name {
	case "list_tasks":
		tasks, err := e.tasks.List(ctx)
		if err != nil {
			return nil, err
		}
		if status := str("status"); status != "" {
			filtered := tasks[:0]
			for _, t := range tasks {
				if string(t.Status) == status {
					filtered = append(filtered, t)
				}
			}
			tasks = filtered
		}
		return tasks, nil

	case "create_task":
		t := &task.Task{
			Title:       str("title"),
			Description: str("description"),
			Priority:    task.Priority(str("priority")),
			ProjectID:   strPtr("project_id"),
		}
		if t.Priority == "" {
			t.Priority = task.PriorityMedium
		}
		if d := str("due_date"); d != "" {
			if parsed, err := time.Parse("2006-01-02", d); err == nil {
				t.DueDate = &parsed
			}
		}
		if err := e.tasks.Create(ctx, t); err != nil {
			return nil, err
		}
		return t, nil

	case "update_task":
		id := str("id")
		if id == "" {
			return nil, fmt.Errorf("id is required")
		}
		t, err := e.tasks.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if v := str("title"); v != "" {
			t.Title = v
		}
		if v := str("status"); v != "" {
			t.Status = task.Status(v)
		}
		if v := str("priority"); v != "" {
			t.Priority = task.Priority(v)
		}
		if err := e.tasks.Update(ctx, t); err != nil {
			return nil, err
		}
		return t, nil

	case "search_notes":
		query := str("query")
		if query == "" {
			return nil, fmt.Errorf("query is required")
		}
		return e.notes.Search(ctx, query)

	case "list_notes":
		return e.notes.List(ctx, strPtr("folder_id"))

	case "create_note":
		if str("title") == "" {
			return nil, fmt.Errorf("title is required")
		}
		n := &note.Note{
			Title:   str("title"),
			Content: str("content"),
		}
		if err := e.notes.Create(ctx, n); err != nil {
			return nil, err
		}
		return n, nil

	case "list_projects":
		return e.projects.List(ctx)

	case "list_events":
		from := time.Now()
		to := from.Add(30 * 24 * time.Hour)
		if v := str("from"); v != "" {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				from = t
			}
		}
		if v := str("to"); v != "" {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				to = t
			}
		}
		return e.events.ListByRange(ctx, from, to)

	case "list_apps":
		return e.apps.List(ctx)

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// BuildSystemPrompt constructs the assistant system prompt with current date.
func BuildSystemPrompt() string {
	now := time.Now()
	return fmt.Sprintf(`You are hube's AI assistant — a smart, concise personal hub for Iosif.

Current date: %s (%s)

hube manages tasks, notes, projects, calendar events, and apps. You have tools to read and write data in the hub. Use them whenever the user asks about their data or wants to take an action.

Guidelines:
- Be concise and direct. No unnecessary preambles.
- When you create or update something, confirm briefly what you did.
- Format lists with markdown. Use code blocks when showing code.
- If you're unsure which project or note the user means, ask for clarification rather than guessing.
`, now.Format("Monday, January 2, 2006"), now.Format("15:04"))
}
