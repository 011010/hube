package note

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	PropertyText    = "text"
	PropertyNumber  = "number"
	PropertyDate    = "date"
	PropertyBoolean = "boolean"
	PropertyList    = "list"

	maxProperties     = 50
	maxTextLength     = 10_000
	maxListItems      = 100
	maxListItemLength = 500
)

var propertyKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)

var reservedPropertyKeys = map[string]struct{}{
	"title":      {},
	"status":     {},
	"tags":       {},
	"created_at": {},
	"updated_at": {},
}

// IsReservedPropertyKey reports whether key collides with a fixed field that
// ExportAgentMarkdown writes to the same YAML frontmatter.
func IsReservedPropertyKey(key string) bool {
	_, ok := reservedPropertyKeys[key]
	return ok
}

// Property is a typed key-value pair attached to a note. Value must match
// Type: a string for text, a finite float64 for number (the shape produced by
// JSON decoding), a YYYY-MM-DD string for date, a bool for boolean, and a
// []string or []any of strings for list.
type Property struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// Normalize lowercases and trims Key and Type.
func (p *Property) Normalize() {
	p.Key = strings.ToLower(strings.TrimSpace(p.Key))
	p.Type = strings.ToLower(strings.TrimSpace(p.Type))
}

// Validate normalizes a copy of p and reports whether its key, type, and value
// are valid. It does not mutate the receiver.
func (p Property) Validate() error {
	p.Normalize()
	if !propertyKeyPattern.MatchString(p.Key) {
		return &ValidationError{Field: "properties", Message: fmt.Sprintf("invalid property key: %q", truncateRunes(p.Key, 64))}
	}

	switch p.Type {
	case PropertyText:
		s, ok := p.Value.(string)
		if !ok {
			return typeError(p.Key, PropertyText)
		}
		if utf8.RuneCountInString(s) > maxTextLength {
			return &ValidationError{Field: "properties." + p.Key, Message: "value is too long"}
		}
	case PropertyNumber:
		n, ok := p.Value.(float64)
		if !ok || math.IsNaN(n) || math.IsInf(n, 0) {
			return typeError(p.Key, PropertyNumber)
		}
	case PropertyDate:
		s, ok := p.Value.(string)
		if !ok {
			return typeError(p.Key, PropertyDate)
		}
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return &ValidationError{Field: "properties." + p.Key, Message: "date must be YYYY-MM-DD"}
		}
	case PropertyBoolean:
		if _, ok := p.Value.(bool); !ok {
			return typeError(p.Key, PropertyBoolean)
		}
	case PropertyList:
		items, ok := StringList(p.Value)
		if !ok {
			return typeError(p.Key, PropertyList)
		}
		if len(items) > maxListItems {
			return &ValidationError{Field: "properties." + p.Key, Message: "too many list items"}
		}
		for _, item := range items {
			if utf8.RuneCountInString(item) > maxListItemLength {
				return &ValidationError{Field: "properties." + p.Key, Message: "list item is too long"}
			}
		}
	default:
		return &ValidationError{Field: "properties." + p.Key, Message: fmt.Sprintf("unknown property type: %q", truncateRunes(p.Type, 64))}
	}
	return nil
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

func typeError(key, typ string) error {
	return &ValidationError{Field: "properties." + key, Message: "value does not match type " + typ}
}

// StringList converts a list property value into []string. It accepts []string
// and []any containing only strings, the latter being the shape produced by
// JSON decoding.
func StringList(v any) ([]string, bool) {
	switch items := v.(type) {
	case []string:
		return items, true
	case []any:
		out := make([]string, 0, len(items))
		for _, item := range items {
			s, ok := item.(string)
			if !ok {
				return nil, false
			}
			out = append(out, s)
		}
		return out, true
	default:
		return nil, false
	}
}

// ValidateProperties checks the count, keys, and values of props. It normalizes
// props in place before validating.
func ValidateProperties(props []Property) error {
	if len(props) > maxProperties {
		return &ValidationError{Field: "properties", Message: "too many properties"}
	}
	seen := make(map[string]struct{}, len(props))
	for i := range props {
		props[i].Normalize()
		if IsReservedPropertyKey(props[i].Key) {
			return &ValidationError{Field: "properties." + props[i].Key, Message: "reserved property key"}
		}
		if err := props[i].Validate(); err != nil {
			return err
		}
		if _, ok := seen[props[i].Key]; ok {
			return &ValidationError{Field: "properties." + props[i].Key, Message: "duplicate property key"}
		}
		seen[props[i].Key] = struct{}{}
	}
	return nil
}
