package note

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/husari/hube/internal/domain/note"
)

type fakeEmbedder struct {
	// vectors maps input text to the embedding returned for it.
	vectors map[string][]float32
	err     error
}

func (f *fakeEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.vectors[text], nil
}

// fakeEmbeddingRepo implements only the methods SemanticSearch touches. The
// embedded interface is nil, so any other call panics loudly instead of
// silently returning a zero value.
type fakeEmbeddingRepo struct {
	note.Repository
	records []note.EmbeddingRecord
	notes   map[string]*note.Note
	listErr error
}

func (f *fakeEmbeddingRepo) FindAllEmbeddings(context.Context) ([]note.EmbeddingRecord, error) {
	return f.records, f.listErr
}

func (f *fakeEmbeddingRepo) FindByID(_ context.Context, id string) (*note.Note, error) {
	return f.notes[id], nil
}

func (f *fakeEmbeddingRepo) StoreEmbedding(context.Context, string, []byte) error { return nil }

func record(t *testing.T, id string, vec []float32) note.EmbeddingRecord {
	t.Helper()
	raw, err := json.Marshal(vec)
	if err != nil {
		t.Fatalf("marshal %s: %v", id, err)
	}
	return note.EmbeddingRecord{ID: id, Embedding: raw}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a, b []float32
		want float32
	}{
		{"identical vectors", []float32{1, 2, 3}, []float32{1, 2, 3}, 1},
		{"opposite vectors", []float32{1, 0}, []float32{-1, 0}, -1},
		{"orthogonal vectors", []float32{1, 0}, []float32{0, 1}, 0},
		{"scale invariant", []float32{1, 2}, []float32{10, 20}, 1},
		{"empty input", nil, []float32{1, 2}, 0},
		{"zero vector", []float32{0, 0}, []float32{1, 2}, 0},
		{"large magnitudes", []float32{1e4, 2e4}, []float32{1e4, 2e4}, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := cosineSimilarity(tc.a, tc.b)
			if math.Abs(float64(got-tc.want)) > 1e-6 {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSemanticSearchRanksByDescendingScore(t *testing.T) {
	query := []float32{1, 0}
	repo := &fakeEmbeddingRepo{
		records: []note.EmbeddingRecord{
			record(t, "far", []float32{0.6, 0.8}),  // cos = 0.6
			record(t, "exact", []float32{1, 0}),    // cos = 1.0
			record(t, "near", []float32{0.8, 0.6}), // cos = 0.8
		},
		notes: map[string]*note.Note{
			"far":   {ID: "far", Title: "far"},
			"exact": {ID: "exact", Title: "exact"},
			"near":  {ID: "near", Title: "near"},
		},
	}
	svc := NewRAGService(repo, &fakeEmbedder{vectors: map[string][]float32{"q": query}})

	results, err := svc.SemanticSearch(context.Background(), "q", 0)
	if err != nil {
		t.Fatalf("SemanticSearch: %v", err)
	}

	want := []string{"exact", "near", "far"}
	if len(results) != len(want) {
		t.Fatalf("got %d results, want %d", len(results), len(want))
	}
	for i, id := range want {
		if results[i].Note.ID != id {
			t.Errorf("position %d: got %q, want %q", i, results[i].Note.ID, id)
		}
	}
}

func TestSemanticSearchDropsResultsBelowThreshold(t *testing.T) {
	repo := &fakeEmbeddingRepo{
		records: []note.EmbeddingRecord{
			record(t, "relevant", []float32{1, 0}),  // cos = 1.0
			record(t, "unrelated", []float32{0, 1}), // cos = 0.0, under 0.3
		},
		notes: map[string]*note.Note{
			"relevant":  {ID: "relevant"},
			"unrelated": {ID: "unrelated"},
		},
	}
	svc := NewRAGService(repo, &fakeEmbedder{vectors: map[string][]float32{"q": {1, 0}}})

	results, err := svc.SemanticSearch(context.Background(), "q", 0)
	if err != nil {
		t.Fatalf("SemanticSearch: %v", err)
	}

	if len(results) != 1 || results[0].Note.ID != "relevant" {
		t.Fatalf("expected only the relevant note, got %+v", results)
	}
}

func TestSemanticSearchRespectsTopK(t *testing.T) {
	repo := &fakeEmbeddingRepo{
		records: []note.EmbeddingRecord{
			record(t, "a", []float32{1, 0}),
			record(t, "b", []float32{0.9, 0.1}),
			record(t, "c", []float32{0.8, 0.2}),
		},
		notes: map[string]*note.Note{
			"a": {ID: "a"}, "b": {ID: "b"}, "c": {ID: "c"},
		},
	}
	svc := NewRAGService(repo, &fakeEmbedder{vectors: map[string][]float32{"q": {1, 0}}})

	results, err := svc.SemanticSearch(context.Background(), "q", 2)
	if err != nil {
		t.Fatalf("SemanticSearch: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestSemanticSearchSkipsCorruptEmbeddings(t *testing.T) {
	repo := &fakeEmbeddingRepo{
		records: []note.EmbeddingRecord{
			{ID: "corrupt", Embedding: []byte("not json")},
			record(t, "good", []float32{1, 0}),
		},
		notes: map[string]*note.Note{"good": {ID: "good"}},
	}
	svc := NewRAGService(repo, &fakeEmbedder{vectors: map[string][]float32{"q": {1, 0}}})

	// One unreadable row must not sink the whole search.
	results, err := svc.SemanticSearch(context.Background(), "q", 0)
	if err != nil {
		t.Fatalf("SemanticSearch: %v", err)
	}
	if len(results) != 1 || results[0].Note.ID != "good" {
		t.Fatalf("expected the readable note only, got %+v", results)
	}
}

func TestSemanticSearchFailsWhenQueryCannotBeEmbedded(t *testing.T) {
	repo := &fakeEmbeddingRepo{}
	svc := NewRAGService(repo, &fakeEmbedder{err: errors.New("openai down")})

	if _, err := svc.SemanticSearch(context.Background(), "q", 0); err == nil {
		t.Fatal("expected an error, got nil")
	}
}
