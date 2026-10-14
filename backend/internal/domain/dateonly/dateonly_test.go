package dateonly

import "testing"

func TestCanonical(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		want   string
		wantOK bool
	}{
		{"a bare day", "2026-10-13", "2026-10-13", true},
		{"surrounding whitespace is ignored", "  2026-10-13  ", "2026-10-13", true},
		{"an RFC3339 timestamp is reduced to its day", "2025-12-31T00:00:00Z", "2025-12-31", true},
		{"an RFC3339 timestamp with an offset keeps its own calendar day", "2025-12-31T23:00:00-06:00", "2025-12-31", true},
		{"blank is not a date", "   ", "", false},
		{"a non-ISO spelling is not a date", "13/10/2026", "", false},
		{"an impossible day is not a date", "2026-13-40", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Canonical(tc.in)
			if ok != tc.wantOK || got != tc.want {
				t.Errorf("Canonical(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestCanonicalize(t *testing.T) {
	day := func(s string) *string { return &s }

	tests := []struct {
		name string
		in   *string
		want *string
	}{
		{"nil stays nil", nil, nil},
		{"blank collapses to nil", day("   "), nil},
		{"a bare day is kept", day("2026-10-13"), day("2026-10-13")},
		{"an RFC3339 timestamp is reduced", day("2026-10-13T08:30:00Z"), day("2026-10-13")},
		{"an unrecognised value is trimmed but kept", day("  13/10/2026 "), day("13/10/2026")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Canonicalize(tc.in)
			switch {
			case tc.want == nil && got != nil:
				t.Errorf("got %q, want nil", *got)
			case tc.want != nil && got == nil:
				t.Errorf("got nil, want %q", *tc.want)
			case tc.want != nil && *got != *tc.want:
				t.Errorf("got %q, want %q", *got, *tc.want)
			}
		})
	}
}

func TestValid(t *testing.T) {
	if !Valid("2026-10-13") {
		t.Error("a bare day should be valid")
	}
	for _, bad := range []string{"", "2026-10-13T00:00:00Z", "13/10/2026", "2026-13-40"} {
		if Valid(bad) {
			t.Errorf("%q should not be valid", bad)
		}
	}
}
