package sqlite

import (
	"testing"
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
