package note

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/husari/hube/internal/domain/note"
	"gopkg.in/yaml.v3"
)

type mockRepo struct {
	created          *note.Note
	updated          *note.Note
	note             *note.Note
	savedProperties  []note.Property
	setPropertiesErr error
}

func (m *mockRepo) FindAll(ctx context.Context, folderID *string) ([]note.Note, error) {
	return nil, nil
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*note.Note, error) {
	return m.note, nil
}

func (m *mockRepo) Search(ctx context.Context, query string) ([]note.Note, error) {
	return nil, nil
}

func (m *mockRepo) Create(ctx context.Context, n *note.Note) error {
	m.created = n
	return nil
}

func (m *mockRepo) Update(ctx context.Context, n *note.Note) error {
	m.updated = n
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockRepo) SetTags(ctx context.Context, noteID string, tags []string) error {
	return nil
}

func (m *mockRepo) SetProperties(ctx context.Context, noteID string, props []note.Property) error {
	m.savedProperties = props
	return m.setPropertiesErr
}

func (m *mockRepo) FindAllLinks(ctx context.Context) ([]note.Link, error) {
	return nil, nil
}

func (m *mockRepo) Graph(ctx context.Context) (*note.Graph, error) {
	return &note.Graph{}, nil
}

func TestService_Create_NormalizesAndValidates(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{Title: "Test"}
	if err := svc.Create(context.Background(), n); err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	if n.Status != note.StatusDraft {
		t.Errorf("Status = %q, want %q", n.Status, note.StatusDraft)
	}
	if repo.created != n {
		t.Error("Create did not persist the normalized note")
	}
}

func TestService_Create_InvalidStatus(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{Title: "Test", Status: "invalid"}
	err := svc.Create(context.Background(), n)
	if err == nil {
		t.Fatal("Create expected error for invalid status")
	}
	var valErr *note.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("error type = %T, want *note.ValidationError", err)
	}
	if valErr.Field != "status" {
		t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, "status")
	}
	if repo.created != nil {
		t.Error("Create should not persist an invalid note")
	}
}

func TestService_Create_EmptyTitle(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{Status: note.StatusDraft}
	err := svc.Create(context.Background(), n)
	if err == nil {
		t.Fatal("Create expected error for empty title")
	}
	var valErr *note.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("error type = %T, want *note.ValidationError", err)
	}
	if valErr.Field != "title" {
		t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, "title")
	}
	if repo.created != nil {
		t.Error("Create should not persist an invalid note")
	}
}

func TestService_Create_EmptyProperties(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{Title: "Test"}
	if err := svc.Create(context.Background(), n); err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}
	if n.Properties == nil {
		t.Fatal("expected non-nil empty properties, got nil")
	}
	if len(n.Properties) != 0 {
		t.Fatalf("expected empty properties, got %+v", n.Properties)
	}
}

func TestService_Update_EmptyProperties(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{ID: "1", Title: "Test"}
	if err := svc.Update(context.Background(), n); err != nil {
		t.Fatalf("Update returned unexpected error: %v", err)
	}
	if n.Properties == nil {
		t.Fatal("expected non-nil empty properties, got nil")
	}
	if len(n.Properties) != 0 {
		t.Fatalf("expected empty properties, got %+v", n.Properties)
	}
}

func TestService_Update_NormalizesEmptyStatus(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{ID: "1", Title: "Test"}
	if err := svc.Update(context.Background(), n); err != nil {
		t.Fatalf("Update returned unexpected error: %v", err)
	}

	if n.Status != note.StatusDraft {
		t.Errorf("Status = %q, want %q", n.Status, note.StatusDraft)
	}
	if repo.updated != n {
		t.Error("Update did not persist the normalized note")
	}
}

func TestService_Update_InvalidStatus(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{ID: "1", Title: "Test", Status: "invalid"}
	err := svc.Update(context.Background(), n)
	if err == nil {
		t.Fatal("Update expected error for invalid status")
	}
	var valErr *note.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("error type = %T, want *note.ValidationError", err)
	}
	if valErr.Field != "status" {
		t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, "status")
	}
	if repo.updated != nil {
		t.Error("Update should not persist an invalid note")
	}
}

func TestService_ExportAgentMarkdown(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	created := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 7, 2, 9, 30, 0, 0, time.UTC)
	n := &note.Note{
		ID:        "1",
		Title:     "My Note",
		Content:   "Hello world",
		Status:    note.StatusDraft,
		Tags:      []string{"go", "test"},
		CreatedAt: created,
		UpdatedAt: updated,
	}

	md, err := svc.ExportAgentMarkdown(context.Background(), n)
	if err != nil {
		t.Fatalf("ExportAgentMarkdown returned unexpected error: %v", err)
	}

	if !strings.HasPrefix(md, "---\n") {
		t.Fatalf("expected markdown to start with YAML frontmatter delimiter, got: %q", md)
	}
	wantFields := []string{
		`title: "My Note"`,
		"status: draft",
		"tags:",
		`  - "go"`,
		`  - "test"`,
		"created_at: 2026-07-01T12:00:00Z",
		"updated_at: 2026-07-02T09:30:00Z",
	}
	for _, want := range wantFields {
		if !strings.Contains(md, want) {
			t.Errorf("expected markdown to contain %q, got:\n%s", want, md)
		}
	}
	if !strings.HasSuffix(strings.TrimRight(md, "\n"), "Hello world") {
		t.Errorf("expected markdown body to end with note content, got:\n%s", md)
	}
	// Frontmatter must be closed before the body.
	parts := strings.SplitN(md, "---\n", 3)
	if len(parts) != 3 {
		t.Fatalf("expected exactly two --- delimiters, got markdown:\n%s", md)
	}
	if !strings.Contains(parts[2], "Hello world") {
		t.Errorf("expected body section to contain note content, got: %q", parts[2])
	}
}

func TestService_ExportAgentMarkdown_NoTags(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{
		ID:      "1",
		Title:   "Bare Note",
		Content: "",
		Status:  note.StatusDraft,
	}

	md, err := svc.ExportAgentMarkdown(context.Background(), n)
	if err != nil {
		t.Fatalf("ExportAgentMarkdown returned unexpected error: %v", err)
	}
	if !strings.Contains(md, "tags: []") {
		t.Errorf("expected tags: [] for empty tags, got:\n%s", md)
	}
}

func TestService_ExportAgentMarkdown_EscapesTagsAgainstYAMLInjection(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{
		ID:      "1",
		Title:   "Injection Note",
		Content: "body",
		Status:  note.StatusDraft,
		// A tag containing a raw newline followed by YAML syntax used to be
		// written unescaped, letting it break out of the tags list and inject
		// a sibling "injected" key into the frontmatter.
		Tags: []string{"safe", "evil\ninjected: true"},
	}

	md, err := svc.ExportAgentMarkdown(context.Background(), n)
	if err != nil {
		t.Fatalf("ExportAgentMarkdown returned unexpected error: %v", err)
	}

	parts := strings.SplitN(md, "---\n", 3)
	if len(parts) != 3 {
		t.Fatalf("expected exactly two --- delimiters, got markdown:\n%s", md)
	}
	frontmatter := parts[1]

	var doc map[string]any
	if err := yaml.Unmarshal([]byte(frontmatter), &doc); err != nil {
		t.Fatalf("frontmatter is not valid YAML (tag injection broke the document): %v\nfrontmatter:\n%s", err, frontmatter)
	}

	if _, injected := doc["injected"]; injected {
		t.Errorf("tag value injected a sibling top-level YAML key, got parsed doc: %#v", doc)
	}

	tags, ok := doc["tags"].([]any)
	if !ok || len(tags) != 2 {
		t.Fatalf("expected tags to parse as a 2-element list, got: %#v", doc["tags"])
	}
	// yamlQuote escapes the embedded raw line break as \n, so the tag stays a
	// single scalar and round-trips byte-identically with no injected key.
	if tags[0] != "safe" || tags[1] != "evil\ninjected: true" {
		t.Errorf("expected tags to round-trip exactly, got: %#v", tags)
	}
}

func TestService_SetProperties(t *testing.T) {
	repo := &mockRepo{note: &note.Note{ID: "n1", Title: "Test", Status: note.StatusDraft}}
	svc := NewService(repo)

	props := []note.Property{
		{Key: " Rating ", Type: note.PropertyNumber, Value: float64(8)},
	}
	got, err := svc.SetProperties(context.Background(), "n1", props)
	if err != nil {
		t.Fatalf("SetProperties() error = %v", err)
	}
	if len(got) != 1 || got[0].Key != "rating" {
		t.Fatalf("expected normalized property, got %+v", got)
	}
	if len(repo.savedProperties) != 1 || repo.savedProperties[0].Key != "rating" {
		t.Fatalf("expected repo to receive normalized properties, got %+v", repo.savedProperties)
	}
}

func TestService_SetProperties_Validation(t *testing.T) {
	repo := &mockRepo{note: &note.Note{ID: "n1", Title: "Test", Status: note.StatusDraft}}
	svc := NewService(repo)

	_, err := svc.SetProperties(context.Background(), "n1", []note.Property{
		{Key: "rating", Type: note.PropertyNumber, Value: "8"},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var valErr *note.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("error type = %T, want *note.ValidationError", err)
	}
	if valErr.Field != "properties.rating" {
		t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, "properties.rating")
	}
	if repo.savedProperties != nil {
		t.Errorf("SetProperties should not persist invalid properties, got %+v", repo.savedProperties)
	}
}

func TestService_SetProperties_NilBecomesEmpty(t *testing.T) {
	repo := &mockRepo{note: &note.Note{ID: "n1", Title: "Test", Status: note.StatusDraft}}
	svc := NewService(repo)

	got, err := svc.SetProperties(context.Background(), "n1", nil)
	if err != nil {
		t.Fatalf("SetProperties() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil empty property slice, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty property slice, got %+v", got)
	}
	if repo.savedProperties == nil {
		t.Fatal("expected repo to receive non-nil empty property slice, got nil")
	}
	if len(repo.savedProperties) != 0 {
		t.Fatalf("expected repo to receive empty property slice, got %+v", repo.savedProperties)
	}
}

func TestService_SetProperties_MissingNote(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	got, err := svc.SetProperties(context.Background(), "missing", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil result, got %+v", got)
	}
}

func TestService_SetProperties_MissingNoteRace(t *testing.T) {
	repo := &mockRepo{
		note:             &note.Note{ID: "n1", Title: "Test", Status: note.StatusDraft},
		setPropertiesErr: sql.ErrNoRows,
	}
	svc := NewService(repo)

	got, err := svc.SetProperties(context.Background(), "n1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil result when the note disappears mid-write, got %+v", got)
	}
}

func TestService_ExportAgentMarkdown_Properties(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{
		ID:      "1",
		Title:   "Properties Note",
		Content: "Body with properties",
		Status:  note.StatusDraft,
		Properties: []note.Property{
			{Key: "archived", Type: note.PropertyBoolean, Value: true},
			{Key: "published", Type: note.PropertyDate, Value: "2024-03-11"},
			{Key: "rating", Type: note.PropertyNumber, Value: float64(8)},
			{Key: "source", Type: note.PropertyText, Value: "https://example.com"},
			{Key: "topics", Type: note.PropertyList, Value: []string{"design", "tools"}},
		},
		CreatedAt: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 2, 9, 30, 0, 0, time.UTC),
	}

	md, err := svc.ExportAgentMarkdown(context.Background(), n)
	if err != nil {
		t.Fatalf("ExportAgentMarkdown returned unexpected error: %v", err)
	}

	parts := strings.SplitN(md, "---\n", 3)
	if len(parts) != 3 {
		t.Fatalf("expected exactly two --- delimiters, got markdown:\n%s", md)
	}
	frontmatter := parts[1]

	rawLines := []string{
		`"archived": true`,
		`"published": "2024-03-11"`,
		`"rating": 8`,
		`"source": "https://example.com"`,
		`"topics":`,
		`  - "design"`,
		`  - "tools"`,
	}
	for _, want := range rawLines {
		if !strings.Contains(frontmatter, want) {
			t.Errorf("expected frontmatter to contain %q, got:\n%s", want, frontmatter)
		}
	}

	tagsPos := strings.Index(frontmatter, "tags:")
	archivedPos := strings.Index(frontmatter, `"archived":`)
	createdPos := strings.Index(frontmatter, "created_at:")
	if tagsPos < 0 || archivedPos < 0 || createdPos < 0 {
		t.Fatalf("expected tags, properties and created_at in frontmatter, got:\n%s", frontmatter)
	}
	if !(tagsPos < archivedPos && archivedPos < createdPos) {
		t.Errorf("expected tags before properties before created_at, got tags=%d archived=%d created_at=%d:\n%s",
			tagsPos, archivedPos, createdPos, frontmatter)
	}

	var doc map[string]any
	if err := yaml.Unmarshal([]byte(frontmatter), &doc); err != nil {
		t.Fatalf("frontmatter is not valid YAML: %v\nfrontmatter:\n%s", err, frontmatter)
	}

	archived, ok := doc["archived"].(bool)
	if !ok || !archived {
		t.Errorf("archived = %#v, want true", doc["archived"])
	}
	published, ok := doc["published"].(string)
	if !ok || published != "2024-03-11" {
		t.Errorf("published = %#v, want %q", doc["published"], "2024-03-11")
	}
	rating, ok := doc["rating"].(int)
	if !ok || rating != 8 {
		t.Errorf("rating = %#v, want 8", doc["rating"])
	}
	source, ok := doc["source"].(string)
	if !ok || source != "https://example.com" {
		t.Errorf("source = %#v, want %q", doc["source"], "https://example.com")
	}
	topics, ok := doc["topics"].([]any)
	if !ok {
		t.Fatalf("topics = %#v, want list", doc["topics"])
	}
	if len(topics) != 2 || topics[0] != "design" || topics[1] != "tools" {
		t.Errorf("topics = %#v, want [design tools]", topics)
	}

	if !strings.HasSuffix(strings.TrimRight(md, "\n"), "Body with properties") {
		t.Errorf("expected markdown body to end with note content, got:\n%s", md)
	}
}

func TestService_ExportAgentMarkdown_PropertyEscaping(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{
		ID:      "1",
		Title:   "Escaping Note",
		Content: "body",
		Status:  note.StatusDraft,
		Properties: []note.Property{
			{Key: "notes", Type: note.PropertyText, Value: "line1\nline2"},
			{Key: "ctrl", Type: note.PropertyText, Value: "a\x00b"},
		},
	}

	md, err := svc.ExportAgentMarkdown(context.Background(), n)
	if err != nil {
		t.Fatalf("ExportAgentMarkdown returned unexpected error: %v", err)
	}

	parts := strings.SplitN(md, "---\n", 3)
	if len(parts) != 3 {
		t.Fatalf("expected exactly two --- delimiters, got markdown:\n%s", md)
	}
	frontmatter := parts[1]

	// Newlines and control characters must be escaped to two-character
	// sequences so each scalar stays on one line.
	rawLines := []string{
		`"notes": "line1\nline2"`,
		`"ctrl": "a\x00b"`,
	}
	for _, want := range rawLines {
		if !strings.Contains(frontmatter, want) {
			t.Errorf("expected frontmatter to contain %q, got:\n%s", want, frontmatter)
		}
	}

	var doc map[string]any
	if err := yaml.Unmarshal([]byte(frontmatter), &doc); err != nil {
		t.Fatalf("frontmatter is not valid YAML: %v\nfrontmatter:\n%s", err, frontmatter)
	}

	notes, ok := doc["notes"].(string)
	if !ok || notes != "line1\nline2" {
		t.Errorf("notes = %#v, want %q", doc["notes"], "line1\nline2")
	}
	ctrl, ok := doc["ctrl"].(string)
	if !ok || ctrl != "a\x00b" {
		t.Errorf("ctrl = %#v, want %q", doc["ctrl"], "a\x00b")
	}
}

func TestService_ExportAgentMarkdown_EmptyList(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{
		ID:      "1",
		Title:   "Empty List Note",
		Content: "body",
		Status:  note.StatusDraft,
		Properties: []note.Property{
			{Key: "tags_prop", Type: note.PropertyList, Value: []string{}},
		},
	}

	md, err := svc.ExportAgentMarkdown(context.Background(), n)
	if err != nil {
		t.Fatalf("ExportAgentMarkdown returned unexpected error: %v", err)
	}

	parts := strings.SplitN(md, "---\n", 3)
	if len(parts) != 3 {
		t.Fatalf("expected exactly two --- delimiters, got markdown:\n%s", md)
	}
	frontmatter := parts[1]

	if !strings.Contains(frontmatter, `"tags_prop": []`) {
		t.Errorf("expected empty list to render as [], got:\n%s", frontmatter)
	}

	var doc map[string]any
	if err := yaml.Unmarshal([]byte(frontmatter), &doc); err != nil {
		t.Fatalf("frontmatter is not valid YAML: %v\nfrontmatter:\n%s", err, frontmatter)
	}
	items, ok := doc["tags_prop"].([]any)
	if !ok {
		t.Fatalf("tags_prop = %#v, want list", doc["tags_prop"])
	}
	if items == nil {
		t.Fatal("expected non-nil empty list, got nil")
	}
	if len(items) != 0 {
		t.Fatalf("expected empty list, got %#v", items)
	}
}

func TestService_ExportAgentMarkdown_InvalidList(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	n := &note.Note{
		ID:      "1",
		Title:   "Invalid List Note",
		Content: "body",
		Status:  note.StatusDraft,
		Properties: []note.Property{
			{Key: "topics", Type: note.PropertyList, Value: "not-a-list"},
		},
	}

	md, err := svc.ExportAgentMarkdown(context.Background(), n)
	if err == nil {
		t.Fatal("expected an error for a non-string-list property value")
	}
	if !strings.Contains(err.Error(), "topics") {
		t.Errorf("expected error to name the offending property %q, got: %v", "topics", err)
	}
	if md != "" {
		t.Errorf("expected empty markdown on error, got:\n%s", md)
	}
}

func TestService_ExportAgentMarkdown_ReservedKey(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	// Built directly, bypassing the write path, to simulate a corrupt or
	// hand-edited row that still reaches the exporter.
	n := &note.Note{
		ID:      "1",
		Title:   "Reserved Key Note",
		Content: "body",
		Status:  note.StatusDraft,
		Properties: []note.Property{
			{Key: "title", Type: note.PropertyText, Value: "x"},
		},
	}

	md, err := svc.ExportAgentMarkdown(context.Background(), n)
	if err == nil {
		t.Fatal("expected an error for a reserved property key")
	}
	if !strings.Contains(err.Error(), "reserved key") {
		t.Errorf("expected error to mention reserved key, got: %v", err)
	}
	if md != "" {
		t.Errorf("expected empty markdown on error, got:\n%s", md)
	}
}
