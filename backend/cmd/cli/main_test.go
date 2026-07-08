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
