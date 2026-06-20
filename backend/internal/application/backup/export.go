package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/husari/hube/internal/domain/app"
	"github.com/husari/hube/internal/domain/diagram"
	"github.com/husari/hube/internal/domain/event"
	"github.com/husari/hube/internal/domain/note"
	"github.com/husari/hube/internal/domain/project"
	"github.com/husari/hube/internal/domain/task"
	"github.com/husari/hube/internal/domain/wishlist"
)

// ExportService builds a full data export as an in-memory ZIP archive.
type ExportService struct {
	noteRepo     note.Repository
	taskRepo     task.Repository
	eventRepo    event.Repository
	appRepo      app.Repository
	wishlistRepo wishlist.Repository
	projectRepo  project.Repository
	diagramRepo  diagram.Repository
}

// NewExportService creates a new ExportService with the given domain repositories.
func NewExportService(
	noteRepo note.Repository,
	taskRepo task.Repository,
	eventRepo event.Repository,
	appRepo app.Repository,
	wishlistRepo wishlist.Repository,
	projectRepo project.Repository,
	diagramRepo diagram.Repository,
) *ExportService {
	return &ExportService{
		noteRepo:     noteRepo,
		taskRepo:     taskRepo,
		eventRepo:    eventRepo,
		appRepo:      appRepo,
		wishlistRepo: wishlistRepo,
		projectRepo:  projectRepo,
		diagramRepo:  diagramRepo,
	}
}

// BuildZip fetches all data and returns an in-memory ZIP archive.
func (s *ExportService) BuildZip(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	if err := s.addNotes(ctx, zw); err != nil {
		return nil, fmt.Errorf("export: notes: %w", err)
	}
	if err := s.addJSON(ctx, zw, "tasks.json", func() (any, error) {
		return s.taskRepo.FindAll(ctx)
	}); err != nil {
		return nil, fmt.Errorf("export: tasks: %w", err)
	}
	if err := s.addEvents(ctx, zw); err != nil {
		return nil, fmt.Errorf("export: events: %w", err)
	}
	if err := s.addJSON(ctx, zw, "apps.json", func() (any, error) {
		return s.appRepo.FindAll(ctx)
	}); err != nil {
		return nil, fmt.Errorf("export: apps: %w", err)
	}
	if err := s.addJSON(ctx, zw, "wishlist.json", func() (any, error) {
		return s.wishlistRepo.FindAll(ctx)
	}); err != nil {
		return nil, fmt.Errorf("export: wishlist: %w", err)
	}
	if err := s.addJSON(ctx, zw, "projects.json", func() (any, error) {
		return s.projectRepo.FindAll(ctx)
	}); err != nil {
		return nil, fmt.Errorf("export: projects: %w", err)
	}
	if err := s.addJSON(ctx, zw, "diagrams.json", func() (any, error) {
		return s.diagramRepo.FindAll(ctx)
	}); err != nil {
		return nil, fmt.Errorf("export: diagrams: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("export: close zip: %w", err)
	}
	return buf.Bytes(), nil
}

// addNotes writes each note as a Markdown file with YAML frontmatter.
func (s *ExportService) addNotes(ctx context.Context, zw *zip.Writer) error {
	notes, err := s.noteRepo.FindAll(ctx, nil)
	if err != nil {
		return err
	}
	for _, n := range notes {
		slug := titleSlug(n.Title)
		if slug == "" {
			slug = n.ID
		}
		fname := fmt.Sprintf("notes/%s.md", slug)
		w, err := zw.Create(fname)
		if err != nil {
			return err
		}

		tags := "[]"
		if len(n.Tags) > 0 {
			tagItems := make([]string, len(n.Tags))
			for i, t := range n.Tags {
				tagItems[i] = fmt.Sprintf("  - %s", t)
			}
			tags = "\n" + strings.Join(tagItems, "\n")
		}

		frontmatter := fmt.Sprintf("---\ntitle: %q\ntags: %s\ncreated_at: %s\n---\n\n",
			n.Title,
			tags,
			n.CreatedAt.UTC().Format(time.RFC3339),
		)
		if _, err := fmt.Fprint(w, frontmatter+n.Content); err != nil {
			return err
		}
	}
	return nil
}

// addJSON fetches data via fetch and writes it as a JSON file inside the zip.
func (s *ExportService) addJSON(_ context.Context, zw *zip.Writer, name string, fetch func() (any, error)) error {
	data, err := fetch()
	if err != nil {
		return err
	}
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// addEvents writes all events in minimal RFC 5545 iCal format.
func (s *ExportService) addEvents(ctx context.Context, zw *zip.Writer) error {
	events, err := s.eventRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	w, err := zw.Create("events.ics")
	if err != nil {
		return err
	}

	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//hube//hube//EN",
	}

	for _, e := range events {
		desc := strings.ReplaceAll(e.Description, "\n", "\\n")
		desc = strings.ReplaceAll(desc, ",", "\\,")
		desc = strings.ReplaceAll(desc, ";", "\\;")

		title := strings.ReplaceAll(e.Title, ",", "\\,")
		title = strings.ReplaceAll(title, ";", "\\;")

		lines = append(lines,
			"BEGIN:VEVENT",
			fmt.Sprintf("UID:%s@hube", e.ID),
			fmt.Sprintf("DTSTART:%s", e.StartAt.UTC().Format("20060102T150405Z")),
			fmt.Sprintf("DTEND:%s", e.EndAt.UTC().Format("20060102T150405Z")),
			fmt.Sprintf("SUMMARY:%s", title),
			fmt.Sprintf("DESCRIPTION:%s", desc),
			"END:VEVENT",
		)
	}

	lines = append(lines, "END:VCALENDAR")

	_, err = fmt.Fprint(w, strings.Join(lines, "\r\n")+"\r\n")
	return err
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// titleSlug converts a note title to a URL-safe filename slug.
func titleSlug(title string) string {
	s := strings.ToLower(title)
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}
