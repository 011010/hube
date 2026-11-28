package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/husari/hube/internal/domain/note"
)

func TestBlocksToText(t *testing.T) {
	tests := []struct {
		name    string
		blocks  string
		want    string
		wantErr bool
	}{
		{
			name:   "empty string",
			blocks: "",
			want:   "",
		},
		{
			name:    "invalid JSON",
			blocks:  "not json",
			want:    "",
			wantErr: true,
		},
		{
			name:   "single paragraph",
			blocks: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello world"}]}]}`,
			want:   "Hello world",
		},
		{
			name:   "multiple paragraphs",
			blocks: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello world"}]},{"type":"paragraph","content":[{"type":"text","text":"Goodbye world"}]}]}`,
			want:   "Hello world Goodbye world",
		},
		{
			name:   "heading and paragraph",
			blocks: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Title"}]},{"type":"paragraph","content":[{"type":"text","text":"Hello world"}]}]}`,
			want:   "Title Hello world",
		},
		{
			name:   "nested list items",
			blocks: `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Item 1"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Item 2"}]}]}]}]}`,
			want:   "Item 1 Item 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := blocksToText(tt.blocks)
			if (err != nil) != tt.wantErr {
				t.Fatalf("blocksToText() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("blocksToText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractLinks(t *testing.T) {
	tests := []struct {
		name   string
		blocks string
		want   []string
	}{
		{
			name:   "empty string",
			blocks: "",
			want:   nil,
		},
		{
			name:   "no links",
			blocks: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello world"}]}]}`,
			want:   nil,
		},
		{
			name:   "single link",
			blocks: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"See [[Other Note]] for details"}]}]}`,
			want:   []string{"Other Note"},
		},
		{
			name:   "multiple links",
			blocks: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"See [[Note A]] and [[Note B]]"}]}]}`,
			want:   []string{"Note A", "Note B"},
		},
		{
			name:   "duplicate links deduplicated",
			blocks: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"[[Note A]] again [[Note A]]"}]}]}`,
			want:   []string{"Note A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractLinks(tt.blocks)
			if len(got) != len(tt.want) {
				t.Fatalf("extractLinks() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("extractLinks()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestNoteRepo_LinksStoredOnCreateAndUpdate(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	target := &note.Note{Title: "Other Note", Status: note.StatusDraft}
	if err := repo.Create(ctx, target); err != nil {
		t.Fatalf("create target note: %v", err)
	}

	source := &note.Note{
		Title:  "Source Note",
		Blocks: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"See [[Other Note]]"}]}]}`,
		Status: note.StatusDraft,
	}
	if err := repo.Create(ctx, source); err != nil {
		t.Fatalf("create source note: %v", err)
	}

	links, err := repo.FindAllLinks(ctx)
	if err != nil {
		t.Fatalf("find all links: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d: %+v", len(links), links)
	}
	if links[0].NoteID != source.ID || links[0].TargetNoteID != target.ID {
		t.Errorf("link = %+v, want note_id=%s target_note_id=%s", links[0], source.ID, target.ID)
	}

	// Update source note to remove the link — old link rows must be replaced.
	source.Blocks = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"No links here"}]}]}`
	if err := repo.Update(ctx, source); err != nil {
		t.Fatalf("update source note: %v", err)
	}

	links, err = repo.FindAllLinks(ctx)
	if err != nil {
		t.Fatalf("find all links after update: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links after removing link text, got %d: %+v", len(links), links)
	}
}

func TestNoteRepo_LinksIgnoreUnknownAndSelfTitles(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	source := &note.Note{
		Title:  "Self Ref",
		Blocks: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"[[Self Ref]] and [[Nonexistent]]"}]}]}`,
		Status: note.StatusDraft,
	}
	if err := repo.Create(ctx, source); err != nil {
		t.Fatalf("create source note: %v", err)
	}

	links, err := repo.FindAllLinks(ctx)
	if err != nil {
		t.Fatalf("find all links: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links (self-link and unknown title ignored), got %d: %+v", len(links), links)
	}
}

func TestNoteRepo_Create_MalformedBlocksReturnsValidationError(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	n := &note.Note{
		Title:  "Bad Blocks",
		Blocks: "not json",
		Status: note.StatusDraft,
	}

	err = repo.Create(ctx, n)
	if err == nil {
		t.Fatal("expected an error for malformed blocks, got nil")
	}
	var valErr *note.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected error to be (or wrap) a *note.ValidationError so the HTTP handler returns 400, got: %T: %v", err, err)
	}
}

func assertProperties(t *testing.T, got []note.Property, want []note.Property) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("property count = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Key != want[i].Key || got[i].Type != want[i].Type {
			t.Fatalf("property[%d] = {%s,%s}, want {%s,%s}", i, got[i].Key, got[i].Type, want[i].Key, want[i].Type)
		}
		if !reflect.DeepEqual(got[i].Value, want[i].Value) {
			t.Fatalf("property %q value = %#v, want %#v", want[i].Key, got[i].Value, want[i].Value)
		}
	}
}

func TestNoteRepo_Properties(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	n := &note.Note{Title: "Property Note", Status: note.StatusDraft}
	if err := repo.Create(ctx, n); err != nil {
		t.Fatalf("create note: %v", err)
	}

	props := []note.Property{
		{Key: "archived", Type: note.PropertyBoolean, Value: true},
		{Key: "published", Type: note.PropertyDate, Value: "2024-03-11"},
		{Key: "rating", Type: note.PropertyNumber, Value: float64(8)},
		{Key: "source", Type: note.PropertyText, Value: "https://example.com"},
		{Key: "topics", Type: note.PropertyList, Value: []string{"design", "tools"}},
	}
	if err := repo.SetProperties(ctx, n.ID, props); err != nil {
		t.Fatalf("set properties: %v", err)
	}

	want := []note.Property{
		{Key: "archived", Type: note.PropertyBoolean, Value: true},
		{Key: "published", Type: note.PropertyDate, Value: "2024-03-11"},
		{Key: "rating", Type: note.PropertyNumber, Value: float64(8)},
		{Key: "source", Type: note.PropertyText, Value: "https://example.com"},
		{Key: "topics", Type: note.PropertyList, Value: []any{"design", "tools"}},
	}

	got, err := repo.FindByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("find note: %v", err)
	}
	assertProperties(t, got.Properties, want)

	// Replace the whole set.
	if err := repo.SetProperties(ctx, n.ID, []note.Property{
		{Key: "archived", Type: note.PropertyBoolean, Value: true},
	}); err != nil {
		t.Fatalf("replace properties: %v", err)
	}
	got, err = repo.FindByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("find note after replace: %v", err)
	}
	assertProperties(t, got.Properties, []note.Property{
		{Key: "archived", Type: note.PropertyBoolean, Value: true},
	})

	// A failed replace must leave the previous set intact.
	if err := repo.SetProperties(ctx, n.ID, []note.Property{
		{Key: "dup", Type: note.PropertyText, Value: "one"},
		{Key: "dup", Type: note.PropertyText, Value: "two"},
	}); err == nil {
		t.Fatal("expected duplicate key insert to fail")
	}
	got, err = repo.FindByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("find note after failed replace: %v", err)
	}
	assertProperties(t, got.Properties, []note.Property{
		{Key: "archived", Type: note.PropertyBoolean, Value: true},
	})

	// Clearing works.
	if err := repo.SetProperties(ctx, n.ID, nil); err != nil {
		t.Fatalf("clear properties: %v", err)
	}
	got, err = repo.FindByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("find note after clear: %v", err)
	}
	assertProperties(t, got.Properties, nil)

	// Cascade on note delete.
	if err := repo.SetProperties(ctx, n.ID, props); err != nil {
		t.Fatalf("set before delete: %v", err)
	}
	if err := repo.Delete(ctx, n.ID); err != nil {
		t.Fatalf("delete note: %v", err)
	}
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM note_properties WHERE note_id = ?`, n.ID); err != nil {
		t.Fatalf("count after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected cascade delete, got %d rows", count)
	}
}

func TestNoteRepo_Properties_TouchesNote(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	n := &note.Note{Title: "Touch Note", Status: note.StatusDraft}
	if err := repo.Create(ctx, n); err != nil {
		t.Fatalf("create note: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`UPDATE notes SET updated_at = '2000-01-01 00:00:00' WHERE id = ?`, n.ID); err != nil {
		t.Fatalf("force old updated_at: %v", err)
	}

	if err := repo.SetProperties(ctx, n.ID, []note.Property{
		{Key: "rating", Type: note.PropertyNumber, Value: float64(1)},
	}); err != nil {
		t.Fatalf("set properties: %v", err)
	}

	got, err := repo.FindByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("find note: %v", err)
	}
	if !got.UpdatedAt.After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("UpdatedAt = %v, want after 2000-01-01T00:00:00Z", got.UpdatedAt)
	}
}

func TestNoteRepo_Properties_MissingNote(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	err = repo.SetProperties(ctx, "missing", nil)
	if err == nil {
		t.Fatal("expected an error for a missing note")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected error to wrap sql.ErrNoRows, got: %v", err)
	}
}

func TestNoteRepo_Properties_ListAndSearch(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	empty, err := repo.FindAll(ctx, nil)
	if err != nil {
		t.Fatalf("find all on empty db: %v", err)
	}
	if empty == nil {
		t.Fatal("find all on empty db returned nil, want non-nil empty slice")
	}
	if len(empty) != 0 {
		t.Fatalf("find all on empty db returned %d notes: %+v", len(empty), empty)
	}

	makeNote := func(title string) *note.Note {
		n := &note.Note{Title: title, Status: note.StatusDraft}
		if err := repo.Create(ctx, n); err != nil {
			t.Fatalf("create %s: %v", title, err)
		}
		return n
	}

	alpha := makeNote("Alpha")
	beta := makeNote("Beta")
	makeNote("Gamma")

	alphaProps := []note.Property{
		{Key: "rating", Type: note.PropertyNumber, Value: float64(8)},
		{Key: "source", Type: note.PropertyText, Value: "https://example.com"},
	}
	if err := repo.SetProperties(ctx, alpha.ID, alphaProps); err != nil {
		t.Fatalf("set alpha properties: %v", err)
	}
	if err := repo.SetProperties(ctx, beta.ID, []note.Property{
		{Key: "archived", Type: note.PropertyBoolean, Value: true},
	}); err != nil {
		t.Fatalf("set beta properties: %v", err)
	}

	all, err := repo.FindAll(ctx, nil)
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	byTitle := make(map[string]note.Note, len(all))
	for _, n := range all {
		byTitle[n.Title] = n
	}

	gotAlpha, ok := byTitle["Alpha"]
	if !ok {
		t.Fatalf("alpha missing from FindAll: %+v", all)
	}
	assertProperties(t, gotAlpha.Properties, []note.Property{
		{Key: "rating", Type: note.PropertyNumber, Value: float64(8)},
		{Key: "source", Type: note.PropertyText, Value: "https://example.com"},
	})

	gotBeta, ok := byTitle["Beta"]
	if !ok {
		t.Fatalf("beta missing from FindAll: %+v", all)
	}
	assertProperties(t, gotBeta.Properties, []note.Property{
		{Key: "archived", Type: note.PropertyBoolean, Value: true},
	})

	gotGamma, ok := byTitle["Gamma"]
	if !ok {
		t.Fatalf("gamma missing from FindAll: %+v", all)
	}
	if gotGamma.Properties == nil {
		t.Fatal("gamma properties is nil, want non-nil empty slice")
	}
	if len(gotGamma.Properties) != 0 {
		t.Fatalf("gamma properties = %+v, want empty", gotGamma.Properties)
	}

	found, err := repo.Search(ctx, "Alpha")
	if err != nil {
		t.Fatalf("search alpha: %v", err)
	}
	if len(found) != 1 || found[0].Title != "Alpha" {
		t.Fatalf("search alpha = %+v, want one Alpha note", found)
	}
	assertProperties(t, found[0].Properties, []note.Property{
		{Key: "rating", Type: note.PropertyNumber, Value: float64(8)},
		{Key: "source", Type: note.PropertyText, Value: "https://example.com"},
	})

	found, err = repo.Search(ctx, "Gamma")
	if err != nil {
		t.Fatalf("search gamma: %v", err)
	}
	if len(found) != 1 || found[0].Title != "Gamma" {
		t.Fatalf("search gamma = %+v, want one Gamma note", found)
	}
	if found[0].Properties == nil {
		t.Fatal("search gamma properties is nil, want non-nil empty slice")
	}
	if len(found[0].Properties) != 0 {
		t.Fatalf("search gamma properties = %+v, want empty", found[0].Properties)
	}
}

func TestNoteRepo_Properties_BatchChunks(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	const noteCount = 501
	wantIdx := make(map[string]int, noteCount)
	for i := 0; i < noteCount; i++ {
		title := fmt.Sprintf("Chunk %03d", i)
		n := &note.Note{Title: title, Status: note.StatusDraft}
		if err := repo.Create(ctx, n); err != nil {
			t.Fatalf("create %s: %v", title, err)
		}
		if err := repo.SetProperties(ctx, n.ID, []note.Property{
			{Key: "idx", Type: note.PropertyNumber, Value: float64(i)},
		}); err != nil {
			t.Fatalf("set properties for %s: %v", title, err)
		}
		wantIdx[title] = i
	}

	notes, err := repo.FindAll(ctx, nil)
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	if len(notes) != noteCount {
		t.Fatalf("find all returned %d notes, want %d", len(notes), noteCount)
	}

	seen := make(map[string]bool, noteCount)
	for _, n := range notes {
		want, ok := wantIdx[n.Title]
		if !ok {
			t.Fatalf("unexpected note title %q", n.Title)
		}
		seen[n.Title] = true
		if len(n.Properties) != 1 || n.Properties[0].Key != "idx" {
			t.Fatalf("note %q properties = %+v, want single idx property", n.Title, n.Properties)
		}
		if n.Properties[0].Type != note.PropertyNumber {
			t.Fatalf("note %q idx type = %q, want %q", n.Title, n.Properties[0].Type, note.PropertyNumber)
		}
		got, ok := n.Properties[0].Value.(float64)
		if !ok {
			t.Fatalf("note %q idx value = %#v, want float64", n.Title, n.Properties[0].Value)
		}
		if got != float64(want) {
			t.Fatalf("note %q idx = %v, want %d (cross-note leakage)", n.Title, got, want)
		}
	}
	if len(seen) != noteCount {
		t.Fatalf("find all returned %d distinct notes, want %d", len(seen), noteCount)
	}
}

func TestNoteRepo_FindNoteIDByProperty(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	alpha := &note.Note{Title: "Alpha", Status: note.StatusDraft}
	if err := repo.Create(ctx, alpha); err != nil {
		t.Fatalf("create alpha: %v", err)
	}
	beta := &note.Note{Title: "Beta", Status: note.StatusDraft}
	if err := repo.Create(ctx, beta); err != nil {
		t.Fatalf("create beta: %v", err)
	}

	if err := repo.SetProperties(ctx, alpha.ID, []note.Property{
		{Key: "source", Type: note.PropertyText, Value: "https://example.com/alpha"},
	}); err != nil {
		t.Fatalf("set alpha properties: %v", err)
	}
	if err := repo.SetProperties(ctx, beta.ID, []note.Property{
		{Key: "source", Type: note.PropertyText, Value: "https://example.com/beta"},
		{Key: "rating", Type: note.PropertyNumber, Value: float64(8)},
	}); err != nil {
		t.Fatalf("set beta properties: %v", err)
	}

	// Exact key plus JSON-encoded value returns the note holding it.
	got, err := repo.FindNoteIDByProperty(ctx, "source", `"https://example.com/alpha"`)
	if err != nil {
		t.Fatalf("find by alpha source: %v", err)
	}
	if got != alpha.ID {
		t.Fatalf("find by alpha source = %q, want alpha id %q", got, alpha.ID)
	}

	got, err = repo.FindNoteIDByProperty(ctx, "source", `"https://example.com/beta"`)
	if err != nil {
		t.Fatalf("find by beta source: %v", err)
	}
	if got != beta.ID {
		t.Fatalf("find by beta source = %q, want beta id %q", got, beta.ID)
	}

	// Non-text values are stored as their JSON form too.
	got, err = repo.FindNoteIDByProperty(ctx, "rating", "8")
	if err != nil {
		t.Fatalf("find by rating: %v", err)
	}
	if got != beta.ID {
		t.Fatalf("find by rating = %q, want beta id %q", got, beta.ID)
	}

	// The same value stored under a different key misses.
	got, err = repo.FindNoteIDByProperty(ctx, "rating", `"https://example.com/alpha"`)
	if err != nil {
		t.Fatalf("find by wrong key: %v", err)
	}
	if got != "" {
		t.Fatalf("find by wrong key = %q, want empty", got)
	}

	// The stored form of a text value is the JSON string: the raw,
	// unencoded text must not match.
	got, err = repo.FindNoteIDByProperty(ctx, "source", "https://example.com/alpha")
	if err != nil {
		t.Fatalf("find by raw value: %v", err)
	}
	if got != "" {
		t.Fatalf("find by raw value = %q, want empty (stored form is JSON-encoded)", got)
	}

	// A missing key misses without error.
	got, err = repo.FindNoteIDByProperty(ctx, "missing", `"https://example.com/alpha"`)
	if err != nil {
		t.Fatalf("find by missing key: %v", err)
	}
	if got != "" {
		t.Fatalf("find by missing key = %q, want empty", got)
	}

	// A missing value under an existing key misses without error.
	got, err = repo.FindNoteIDByProperty(ctx, "source", `"https://example.com/nowhere"`)
	if err != nil {
		t.Fatalf("find by missing value: %v", err)
	}
	if got != "" {
		t.Fatalf("find by missing value = %q, want empty", got)
	}
}

// FindNoteIDByProperty must stay deterministic when two notes hold the
// exact same (key, value): "first" means the lexicographically smallest
// note id — stable by id, not whichever matching row a scan happens to
// visit first.
func TestNoteRepo_FindNoteIDByProperty_TieBreak(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	// Like a page clipped twice: two notes can carry the same source URL.
	const url = "https://example.com/clipped-twice"
	one := &note.Note{Title: "Clipped Once", Status: note.StatusDraft}
	if err := repo.Create(ctx, one); err != nil {
		t.Fatalf("create first note: %v", err)
	}
	two := &note.Note{Title: "Clipped Again", Status: note.StatusDraft}
	if err := repo.Create(ctx, two); err != nil {
		t.Fatalf("create second note: %v", err)
	}

	// Both notes hold source = url. Store the larger id's row first:
	// without the ORDER BY, SQLite returns matching rows in insertion
	// order, so the expected id only wins through the id ordering, not
	// through storage luck.
	want := min(one.ID, two.ID)
	for _, id := range []string{max(one.ID, two.ID), want} {
		if err := repo.SetProperties(ctx, id, []note.Property{
			{Key: "source", Type: note.PropertyText, Value: url},
		}); err != nil {
			t.Fatalf("set source on %s: %v", id, err)
		}
	}

	// The value is the JSON-encoded stored form of the text.
	got, err := repo.FindNoteIDByProperty(ctx, "source", strconv.Quote(url))
	if err != nil {
		t.Fatalf("find by shared source: %v", err)
	}
	if got != want {
		t.Fatalf("find by shared source = %q, want lexicographically smaller id %q", got, want)
	}
}

func TestNoteRepo_Update_MalformedBlocksReturnsValidationError(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	n := &note.Note{Title: "Ok Note", Status: note.StatusDraft}
	if err := repo.Create(ctx, n); err != nil {
		t.Fatalf("create note: %v", err)
	}

	n.Blocks = "not json"
	err = repo.Update(ctx, n)
	if err == nil {
		t.Fatal("expected an error for malformed blocks, got nil")
	}
	var valErr *note.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected error to be (or wrap) a *note.ValidationError so the HTTP handler returns 400, got: %T: %v", err, err)
	}
}
