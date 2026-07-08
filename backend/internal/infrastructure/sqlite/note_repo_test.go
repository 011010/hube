package sqlite

import (
	"context"
	"testing"

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

	target := &note.Note{Title: "Other Note", Status: note.StatusDraft, Priority: note.PriorityMedium}
	if err := repo.Create(ctx, target); err != nil {
		t.Fatalf("create target note: %v", err)
	}

	source := &note.Note{
		Title:    "Source Note",
		Blocks:   `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"See [[Other Note]]"}]}]}`,
		Status:   note.StatusDraft,
		Priority: note.PriorityMedium,
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
		Title:    "Self Ref",
		Blocks:   `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"[[Self Ref]] and [[Nonexistent]]"}]}]}`,
		Status:   note.StatusDraft,
		Priority: note.PriorityMedium,
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
