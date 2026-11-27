package note

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestProperty_Validate(t *testing.T) {
	tests := []struct {
		name      string
		prop      Property
		wantKey   string
		wantType  string
		wantField string
		wantErr   bool
	}{
		{name: "text ok", prop: Property{Key: "author", Type: PropertyText, Value: "Kepano"}},
		{name: "number ok", prop: Property{Key: "rating", Type: PropertyNumber, Value: float64(8)}},
		{name: "date ok", prop: Property{Key: "published", Type: PropertyDate, Value: "2024-03-11"}},
		{name: "boolean ok", prop: Property{Key: "archived", Type: PropertyBoolean, Value: true}},
		{name: "list ok", prop: Property{Key: "topics", Type: PropertyList, Value: []any{"design", "tools"}}},
		{name: "list from Go strings", prop: Property{Key: "topics", Type: PropertyList, Value: []string{"design"}}},
		{name: "normalizes key", prop: Property{Key: "  Rating ", Type: PropertyNumber, Value: float64(1)}, wantKey: "rating"},
		{name: "normalizes type", prop: Property{Key: "rating", Type: " Number ", Value: float64(1)}, wantType: "number"},
		{name: "text at limit", prop: Property{Key: "author", Type: PropertyText, Value: strings.Repeat("a", 10_000)}},
		{name: "text under limit with multibyte", prop: Property{Key: "author", Type: PropertyText, Value: strings.Repeat("é", 5_001)}},
		{name: "list at item limit", prop: Property{Key: "topics", Type: PropertyList, Value: repeatList(100, "x")}},
		{name: "list item at length limit", prop: Property{Key: "topics", Type: PropertyList, Value: repeatList(1, strings.Repeat("é", 500))}},
		{name: "empty key", prop: Property{Key: "", Type: PropertyText, Value: "x"}, wantErr: true, wantField: "properties"},
		{name: "key starts with number", prop: Property{Key: "1st", Type: PropertyText, Value: "x"}, wantErr: true, wantField: "properties"},
		{name: "key with space", prop: Property{Key: "my key", Type: PropertyText, Value: "x"}, wantErr: true, wantField: "properties"},
		{name: "unknown type", prop: Property{Key: "x", Type: "select", Value: "x"}, wantErr: true, wantField: "properties.x"},
		{name: "number as string", prop: Property{Key: "rating", Type: PropertyNumber, Value: "8"}, wantErr: true, wantField: "properties.rating"},
		{name: "number NaN", prop: Property{Key: "rating", Type: PropertyNumber, Value: math.NaN()}, wantErr: true, wantField: "properties.rating"},
		{name: "number positive infinity", prop: Property{Key: "rating", Type: PropertyNumber, Value: math.Inf(1)}, wantErr: true, wantField: "properties.rating"},
		{name: "number negative infinity", prop: Property{Key: "rating", Type: PropertyNumber, Value: math.Inf(-1)}, wantErr: true, wantField: "properties.rating"},
		{name: "bad date", prop: Property{Key: "published", Type: PropertyDate, Value: "11/03/2024"}, wantErr: true, wantField: "properties.published"},
		{name: "boolean as string", prop: Property{Key: "archived", Type: PropertyBoolean, Value: "true"}, wantErr: true, wantField: "properties.archived"},
		{name: "list with non-string", prop: Property{Key: "topics", Type: PropertyList, Value: []any{"a", float64(2)}}, wantErr: true, wantField: "properties.topics"},
		{name: "text with wrong type", prop: Property{Key: "author", Type: PropertyText, Value: float64(1)}, wantErr: true, wantField: "properties.author"},
		{name: "text over limit", prop: Property{Key: "author", Type: PropertyText, Value: strings.Repeat("a", 10_001)}, wantErr: true, wantField: "properties.author"},
		{name: "list over item limit", prop: Property{Key: "topics", Type: PropertyList, Value: repeatList(101, "x")}, wantErr: true, wantField: "properties.topics"},
		{name: "list item too long", prop: Property{Key: "topics", Type: PropertyList, Value: repeatList(1, strings.Repeat("a", 501))}, wantErr: true, wantField: "properties.topics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prop.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Validate is pure, so normalization is observed on a copy.
			normalized := tt.prop
			normalized.Normalize()
			if tt.wantKey != "" && normalized.Key != tt.wantKey {
				t.Fatalf("key = %q, want %q", normalized.Key, tt.wantKey)
			}
			if tt.wantType != "" && normalized.Type != tt.wantType {
				t.Fatalf("type = %q, want %q", normalized.Type, tt.wantType)
			}

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}

			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("Validate() error type = %T, want *ValidationError", err)
			}
			if tt.wantField != "" && valErr.Field != tt.wantField {
				t.Fatalf("ValidationError.Field = %q, want %q", valErr.Field, tt.wantField)
			}
		})
	}
}

func repeatList(n int, value string) []any {
	items := make([]any, n)
	for i := range items {
		items[i] = value
	}
	return items
}

func TestProperty_Validate_DoesNotMutateReceiver(t *testing.T) {
	p := Property{Key: "  Rating ", Type: " Number ", Value: float64(1)}
	if err := p.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
	if p.Key != "  Rating " || p.Type != " Number " {
		t.Fatalf("Validate() mutated receiver: key = %q, type = %q", p.Key, p.Type)
	}
}

func TestValidateProperties_Limit(t *testing.T) {
	t.Run("50 properties pass", func(t *testing.T) {
		props := make([]Property, 50)
		for i := range props {
			props[i] = Property{Key: fmt.Sprintf("k%d", i), Type: PropertyText, Value: "x"}
		}
		if err := ValidateProperties(props); err != nil {
			t.Fatalf("ValidateProperties() error = %v, want nil", err)
		}
	})

	t.Run("51 properties rejected", func(t *testing.T) {
		props := make([]Property, 51)
		for i := range props {
			props[i] = Property{Key: fmt.Sprintf("k%d", i), Type: PropertyText, Value: "x"}
		}
		err := ValidateProperties(props)
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("ValidateProperties() error = %v, want *ValidationError", err)
		}
		if valErr.Field != "properties" {
			t.Fatalf("ValidationError.Field = %q, want %q", valErr.Field, "properties")
		}
	})
}

func TestValidateProperties_Duplicates(t *testing.T) {
	tests := []struct {
		name  string
		props []Property
	}{
		{
			name: "exact duplicate keys",
			props: []Property{
				{Key: "rating", Type: PropertyNumber, Value: float64(1)},
				{Key: "rating", Type: PropertyNumber, Value: float64(2)},
			},
		},
		{
			name: "duplicate after normalization",
			props: []Property{
				{Key: "Rating", Type: PropertyNumber, Value: float64(1)},
				{Key: "rating", Type: PropertyNumber, Value: float64(2)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProperties(tt.props)
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("ValidateProperties() error = %v, want *ValidationError", err)
			}
			if valErr.Field != "properties.rating" {
				t.Fatalf("ValidationError.Field = %q, want %q", valErr.Field, "properties.rating")
			}
			if tt.props[0].Key != "rating" {
				t.Fatalf("props[0].Key = %q, want normalized %q", tt.props[0].Key, "rating")
			}
		})
	}
}
