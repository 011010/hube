package main

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"simple", "My Note", "my-note"},
		{"punctuation collapses to single dash", "Hello, World!!", "hello-world"},
		{"leading/trailing spaces trimmed", "  Spaced Title  ", "spaced-title"},
		{"empty title falls back", "", "untitled"},
		{"only punctuation falls back", "!!!", "untitled"},
		{"numbers preserved", "Q3 2026 Plan", "q3-2026-plan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugify(tt.title)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestExportFilename(t *testing.T) {
	usedSlugs := map[string]bool{}

	first := exportFilename(usedSlugs, "My Note", "id-1")
	if first != "my-note.md" {
		t.Errorf("first export filename = %q, want %q", first, "my-note.md")
	}

	// Same title again (slug collision) -> must not overwrite the first file.
	second := exportFilename(usedSlugs, "My Note", "id-2")
	if second == first {
		t.Errorf("second export filename = %q, want it to differ from first %q", second, first)
	}
	if second != "my-note-id-2.md" {
		t.Errorf("second export filename = %q, want %q", second, "my-note-id-2.md")
	}

	// A normalizing collision (different raw title, same slug) must also be disambiguated.
	third := exportFilename(usedSlugs, "my note!!", "id-3")
	if third == first || third == second {
		t.Errorf("third export filename = %q, want it to differ from %q and %q", third, first, second)
	}

	// A distinct title should keep using the plain slug.
	other := exportFilename(usedSlugs, "Other Note", "id-4")
	if other != "other-note.md" {
		t.Errorf("other export filename = %q, want %q", other, "other-note.md")
	}
}
