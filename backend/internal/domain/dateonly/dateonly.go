// Package dateonly holds the single representation of a calendar day used
// across the domain: the bare "2006-01-02" string the UI and the API speak.
//
// A due date is a day, not an instant. Carrying it as a string keeps the wire
// format stable and sidesteps the timezone drift that comes from round-tripping
// a day through time.Time: parsing "2026-10-13" as UTC midnight and rendering
// it back on a machine west of UTC lands on the previous day.
package dateonly

import (
	"strings"
	"time"
)

// Layout is the wire and storage format for a calendar day.
const Layout = "2006-01-02"

// Canonical reduces s to a bare calendar day. It accepts the day itself or a
// full RFC3339 timestamp (the documented ISO 8601 form), which is reduced to
// its date. ok is false when s is neither.
func Canonical(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if _, err := time.Parse(Layout, s); err == nil {
		return s, true
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Format(Layout), true
	}
	return "", false
}

// Canonicalize turns an optional date into its canonical form: nil or blank
// becomes nil, a recognised day or timestamp becomes the bare day, and an
// unrecognised value is returned trimmed so Validate can reject it.
func Canonicalize(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	if day, ok := Canonical(trimmed); ok {
		return &day
	}
	return &trimmed
}

// Valid reports whether s is already a bare calendar day.
func Valid(s string) bool {
	_, err := time.Parse(Layout, s)
	return err == nil
}
