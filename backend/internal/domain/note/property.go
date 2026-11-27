package note

import (
	"fmt"
	"regexp"
	"strings"
	"time"
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

type Property struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

func (p *Property) Normalize() {
	p.Key = strings.ToLower(strings.TrimSpace(p.Key))
	p.Type = strings.ToLower(strings.TrimSpace(p.Type))
}

func (p *Property) Validate() error {
	p.Normalize()
	if !propertyKeyPattern.MatchString(p.Key) {
		return &ValidationError{Field: "properties", Message: fmt.Sprintf("invalid property key: %q", p.Key)}
	}

	switch p.Type {
	case PropertyText:
		s, ok := p.Value.(string)
		if !ok {
			return typeError(p.Key, PropertyText)
		}
		if len(s) > maxTextLength {
			return &ValidationError{Field: "properties." + p.Key, Message: "value is too long"}
		}
	case PropertyNumber:
		if _, ok := p.Value.(float64); !ok {
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
			if len(item) > maxListItemLength {
				return &ValidationError{Field: "properties." + p.Key, Message: "list item is too long"}
			}
		}
	default:
		return &ValidationError{Field: "properties." + p.Key, Message: "unknown property type: " + p.Type}
	}
	return nil
}

func typeError(key, typ string) error {
	return &ValidationError{Field: "properties." + key, Message: "value does not match type " + typ}
}

// StringList accepts the two shapes the API can produce: []any after JSON
// decoding and []string when properties are built in Go.
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

func ValidateProperties(props []Property) error {
	if len(props) > maxProperties {
		return &ValidationError{Field: "properties", Message: "too many properties"}
	}
	seen := make(map[string]struct{}, len(props))
	for i := range props {
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
