package note

import (
	"fmt"
	"testing"
)

func TestValidateProperty(t *testing.T) {
	tests := []struct {
		name    string
		prop    Property
		wantKey string
		wantErr bool
	}{
		{name: "text ok", prop: Property{Key: "author", Type: PropertyText, Value: "Kepano"}},
		{name: "number ok", prop: Property{Key: "rating", Type: PropertyNumber, Value: float64(8)}},
		{name: "date ok", prop: Property{Key: "published", Type: PropertyDate, Value: "2024-03-11"}},
		{name: "boolean ok", prop: Property{Key: "archived", Type: PropertyBoolean, Value: true}},
		{name: "list ok", prop: Property{Key: "topics", Type: PropertyList, Value: []any{"design", "tools"}}},
		{name: "list from Go strings", prop: Property{Key: "topics", Type: PropertyList, Value: []string{"design"}}},
		{name: "normalizes key", prop: Property{Key: "  Rating ", Type: PropertyNumber, Value: float64(1)}, wantKey: "rating"},
		{name: "empty key", prop: Property{Key: "", Type: PropertyText, Value: "x"}, wantErr: true},
		{name: "key starts with number", prop: Property{Key: "1st", Type: PropertyText, Value: "x"}, wantErr: true},
		{name: "key with space", prop: Property{Key: "my key", Type: PropertyText, Value: "x"}, wantErr: true},
		{name: "unknown type", prop: Property{Key: "x", Type: "select", Value: "x"}, wantErr: true},
		{name: "number as string", prop: Property{Key: "rating", Type: PropertyNumber, Value: "8"}, wantErr: true},
		{name: "bad date", prop: Property{Key: "published", Type: PropertyDate, Value: "11/03/2024"}, wantErr: true},
		{name: "boolean as string", prop: Property{Key: "archived", Type: PropertyBoolean, Value: "true"}, wantErr: true},
		{name: "list with non-string", prop: Property{Key: "topics", Type: PropertyList, Value: []any{"a", float64(2)}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prop.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantKey != "" && tt.prop.Key != tt.wantKey {
				t.Fatalf("key = %q, want %q", tt.prop.Key, tt.wantKey)
			}
		})
	}
}

func TestValidateNotePropertiesLimit(t *testing.T) {
	props := make([]Property, 51)
	for i := range props {
		props[i] = Property{Key: fmt.Sprintf("k%d", i), Type: PropertyText, Value: "x"}
	}
	if err := ValidateProperties(props); err == nil {
		t.Fatal("expected error for more than 50 properties")
	}
}

func TestValidateNotePropertiesDuplicates(t *testing.T) {
	props := []Property{
		{Key: "rating", Type: PropertyNumber, Value: float64(1)},
		{Key: "rating", Type: PropertyNumber, Value: float64(2)},
	}
	if err := ValidateProperties(props); err == nil {
		t.Fatal("expected error for duplicate keys")
	}
}
