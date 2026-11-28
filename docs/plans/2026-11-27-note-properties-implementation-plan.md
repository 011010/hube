# Note Properties Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add typed, arbitrary key/value properties to notes so metadata like `rating`, `source`, or `published` survives outside the note body.

**Architecture:** A `note_properties` table stores one JSON-encoded value per `(note_id, key)` with a type discriminator. The domain validates values, the repository bulk-replaces them in a transaction, and the existing note endpoints return them. The frontend replaces the hardcoded `PropertiesPanel` with a generic editor and writes through a dedicated endpoint.

**Tech Stack:** Go 1.26, Chi, SQLite (`sqlx` + `modernc.org/sqlite`), React 19, TypeScript, TanStack Query, Tailwind.

**Design doc:** `docs/plans/2026-11-27-note-properties-design.md`

---

### Task 1: Domain property type and validation

**Files:**
- Create: `backend/internal/domain/note/property.go`
- Test: `backend/internal/domain/note/property_test.go`

**Step 1: Write the failing test**

```go
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
```

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/domain/note/ -run TestValidateProperty -v
```

Expected: FAIL — `undefined: Property`.

**Step 3: Write minimal implementation**

```go
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
```

Add `Properties []Property \`json:"properties"\`` to `Note` in `backend/internal/domain/note/entity.go`.

**Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/domain/note/ -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add backend/internal/domain/note/property.go backend/internal/domain/note/property_test.go backend/internal/domain/note/entity.go
git commit -m "feat(notes): add typed property domain model"
```

---

### Task 2: Migration

**Files:**
- Create: `backend/internal/infrastructure/sqlite/migrations/013_note_properties_table.sql`
- Modify: `backend/internal/infrastructure/sqlite/db.go`
- Test: `backend/internal/infrastructure/sqlite/migrations_test.go`

**Step 1: Write the failing test**

Append to `migrations_test.go`:

```go
func TestMigrations_NotePropertiesTable(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM pragma_table_info('note_properties')`); err != nil {
		t.Fatalf("inspect note_properties: %v", err)
	}
	if count < 4 {
		t.Fatalf("expected note_properties columns, got %d", count)
	}

	var index int
	if err := db.Get(&index, `SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_note_properties_key'`); err != nil {
		t.Fatalf("inspect index: %v", err)
	}
	if index != 1 {
		t.Fatalf("expected idx_note_properties_key, got %d", index)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/infrastructure/sqlite/ -run TestMigrations_NotePropertiesTable -v
```

Expected: FAIL — `no such table: note_properties`.

**Step 3: Write minimal implementation**

`013_note_properties_table.sql`:

```sql
CREATE TABLE note_properties (
  note_id TEXT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  key     TEXT NOT NULL,
  type    TEXT NOT NULL CHECK (type IN ('text', 'number', 'date', 'boolean', 'list')),
  value   TEXT NOT NULL,
  PRIMARY KEY (note_id, key)
);

CREATE INDEX idx_note_properties_key ON note_properties(key);
```

In `db.go`, add:

```go
//go:embed migrations/013_note_properties_table.sql
var notePropertiesTableSQL string
```

and append to `migrations`:

```go
{"013_note_properties_table", notePropertiesTableSQL},
```

**Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/infrastructure/sqlite/ -run TestMigrations -v
```

Expected: PASS, including the existing 009 upgrade test.

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/sqlite/migrations/013_note_properties_table.sql backend/internal/infrastructure/sqlite/db.go backend/internal/infrastructure/sqlite/migrations_test.go
git commit -m "feat(notes): add note_properties table"
```

---

### Task 3: Repository read and bulk replace

**Files:**
- Modify: `backend/internal/domain/note/repository.go`
- Modify: `backend/internal/infrastructure/sqlite/note_repo.go`
- Test: `backend/internal/infrastructure/sqlite/note_repo_test.go`

**Step 1: Write the failing test**

```go
func TestNoteRepo_Properties(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewNoteRepo(db)
	ctx := context.Background()

	n := &note.Note{Title: "Property Note", Status: note.StatusDraft}
	if err := repo.Create(ctx, n); err != nil {
		t.Fatalf("create note: %v", err)
	}

	props := []note.Property{
		{Key: "rating", Type: note.PropertyNumber, Value: float64(8)},
		{Key: "source", Type: note.PropertyText, Value: "https://example.com"},
	}
	if err := repo.SetProperties(ctx, n.ID, props); err != nil {
		t.Fatalf("set properties: %v", err)
	}

	got, err := repo.FindByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("find note: %v", err)
	}
	if len(got.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d: %+v", len(got.Properties), got.Properties)
	}
	if got.Properties[0].Key != "rating" || got.Properties[1].Key != "source" {
		t.Fatalf("unexpected properties: %+v", got.Properties)
	}

	// Replace the whole set.
	if err := repo.SetProperties(ctx, n.ID, []note.Property{
		{Key: "archived", Type: note.PropertyBoolean, Value: true},
	}); err != nil {
		t.Fatalf("replace properties: %v", err)
	}
	got, _ = repo.FindByID(ctx, n.ID)
	if len(got.Properties) != 1 || got.Properties[0].Key != "archived" {
		t.Fatalf("expected replaced set, got %+v", got.Properties)
	}

	// Clearing works.
	if err := repo.SetProperties(ctx, n.ID, nil); err != nil {
		t.Fatalf("clear properties: %v", err)
	}
	got, _ = repo.FindByID(ctx, n.ID)
	if len(got.Properties) != 0 {
		t.Fatalf("expected empty properties, got %+v", got.Properties)
	}

	// Cascade on note delete.
	if err := repo.SetProperties(ctx, n.ID, props); err != nil {
		t.Fatalf("set before delete: %v", err)
	}
	if err := repo.Delete(ctx, n.ID); err != nil {
		t.Fatalf("delete note: %v", err)
	}
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM note_properties WHERE note_id = ?`, n.ID); err != nil {
		t.Fatalf("count after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected cascade delete, got %d rows", count)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/infrastructure/sqlite/ -run TestNoteRepo_Properties -v
```

Expected: FAIL — `SetProperties` not defined, `Properties` field missing on returned note.

**Step 3: Write minimal implementation**

In `repository.go`, add to the `Repository` interface:

```go
SetProperties(ctx context.Context, noteID string, props []Property) error
```

In `note_repo.go`:

```go
type propertyRow struct {
	NoteID string `db:"note_id"`
	Key    string `db:"key"`
	Type   string `db:"type"`
	Value  string `db:"value"`
}

func (r *NoteRepo) loadProperties(ctx context.Context, noteID string) ([]note.Property, error) {
	rows := make([]propertyRow, 0)
	err := r.db.SelectContext(ctx, &rows,
		`SELECT note_id,key,type,value FROM note_properties WHERE note_id = ? ORDER BY key`, noteID)
	if err != nil {
		return nil, err
	}
	return decodeProperties(rows)
}

func decodeProperties(rows []propertyRow) ([]note.Property, error) {
	props := make([]note.Property, 0, len(rows))
	for _, row := range rows {
		var value any
		if err := json.Unmarshal([]byte(row.Value), &value); err != nil {
			return nil, fmt.Errorf("decode property %s: %w", row.Key, err)
		}
		props = append(props, note.Property{Key: row.Key, Type: row.Type, Value: value})
	}
	return props, nil
}

func (r *NoteRepo) loadPropertiesBatch(ctx context.Context, ids []string) (map[string][]note.Property, error) {
	out := make(map[string][]note.Property, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	query, args, err := sqlx.In(
		`SELECT note_id,key,type,value FROM note_properties WHERE note_id IN (?) ORDER BY note_id, key`, ids)
	if err != nil {
		return nil, err
	}
	rows := make([]propertyRow, 0)
	if err := r.db.SelectContext(ctx, &rows, r.db.Rebind(query), args...); err != nil {
		return nil, err
	}
	grouped := make(map[string][]propertyRow)
	for _, row := range rows {
		grouped[row.NoteID] = append(grouped[row.NoteID], row)
	}
	for id, group := range grouped {
		props, err := decodeProperties(group)
		if err != nil {
			return nil, err
		}
		out[id] = props
	}
	return out, nil
}

func (r *NoteRepo) SetProperties(ctx context.Context, noteID string, props []note.Property) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM note_properties WHERE note_id = ?`, noteID); err != nil {
		return err
	}
	for _, p := range props {
		encoded, err := json.Marshal(p.Value)
		if err != nil {
			return fmt.Errorf("encode property %s: %w", p.Key, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO note_properties (note_id,key,type,value) VALUES (?,?,?,?)`,
			noteID, p.Key, p.Type, string(encoded)); err != nil {
			return err
		}
	}
	return tx.Commit()
}
```

Wire reads:

- `FindByID`: after `toNote`, call `loadProperties` and assign. Treat a nil slice as empty.
- `FindAll` and `Search`: collect IDs, call `loadPropertiesBatch` once, assign `propsByNote[row.ID]` to each note.

Use a small helper so responses never marshal `null`:

```go
func withProperties(n *note.Note, props []note.Property) {
	if props == nil {
		props = []note.Property{}
	}
	n.Properties = props
}
```

**Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/infrastructure/sqlite/ -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add backend/internal/domain/note/repository.go backend/internal/infrastructure/sqlite/note_repo.go backend/internal/infrastructure/sqlite/note_repo_test.go
git commit -m "feat(notes): persist and load note properties"
```

---

### Task 4: Service write path

**Files:**
- Modify: `backend/internal/application/note/service.go`
- Test: `backend/internal/application/note/service_test.go`

**Step 1: Write the failing test**

Extend the existing `mockRepo` in `service_test.go` so it can return a note and
record property writes:

```go
type mockRepo struct {
	created         *note.Note
	updated         *note.Note
	note            *note.Note
	savedProperties []note.Property
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*note.Note, error) {
	return m.note, nil
}

func (m *mockRepo) SetProperties(ctx context.Context, noteID string, props []note.Property) error {
	m.savedProperties = props
	return nil
}
```

Then add the tests:

```go
func TestService_SetProperties(t *testing.T) {
	repo := &mockRepo{note: &note.Note{ID: "n1", Title: "Test", Status: note.StatusDraft}}
	svc := NewService(repo)

	props := []note.Property{
		{Key: " Rating ", Type: note.PropertyNumber, Value: float64(8)},
	}
	got, err := svc.SetProperties(context.Background(), "n1", props)
	if err != nil {
		t.Fatalf("SetProperties() error = %v", err)
	}
	if len(got) != 1 || got[0].Key != "rating" {
		t.Fatalf("expected normalized property, got %+v", got)
	}
	if len(repo.savedProperties) != 1 || repo.savedProperties[0].Key != "rating" {
		t.Fatalf("expected repo to receive normalized properties, got %+v", repo.savedProperties)
	}
}

func TestService_SetProperties_Validation(t *testing.T) {
	repo := &mockRepo{note: &note.Note{ID: "n1", Title: "Test", Status: note.StatusDraft}}
	svc := NewService(repo)

	_, err := svc.SetProperties(context.Background(), "n1", []note.Property{
		{Key: "rating", Type: note.PropertyNumber, Value: "8"},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestService_SetProperties_MissingNote(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	got, err := svc.SetProperties(context.Background(), "missing", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil result, got %+v", got)
	}
}
```

Adding `SetProperties` to `note.Repository` forces the mock to implement it; the
two method replacements above are the only changes it needs.

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/application/note/ -run TestService_SetProperties -v
```

Expected: FAIL — `SetProperties` not defined.

**Step 3: Write minimal implementation**

```go
// SetProperties replaces the note's property set. A missing note returns
// (nil, nil) so the HTTP layer can answer 404 without inventing an error type.
func (s *Service) SetProperties(ctx context.Context, id string, props []note.Property) ([]note.Property, error) {
	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
	}
	if err := note.ValidateProperties(props); err != nil {
		return nil, err
	}
	if props == nil {
		props = []note.Property{}
	}
	if err := s.repo.SetProperties(ctx, id, props); err != nil {
		return nil, err
	}
	return props, nil
}
```

Also add properties to `ExportAgentMarkdown`, after the tags block:

```go
for _, p := range n.Properties {
	switch p.Type {
	case note.PropertyNumber, note.PropertyBoolean:
		fmt.Fprintf(&sb, "%s: %v\n", p.Key, p.Value)
	case note.PropertyList:
		items, ok := note.StringList(p.Value)
		if !ok {
			continue
		}
		sb.WriteString(p.Key + ":\n")
		for _, item := range items {
			fmt.Fprintf(&sb, "  - %s\n", yamlQuote(item))
		}
	default:
		fmt.Fprintf(&sb, "%s: %s\n", p.Key, yamlQuote(fmt.Sprint(p.Value)))
	}
}
```

**Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/application/note/ -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add backend/internal/application/note/service.go backend/internal/application/note/service_test.go
git commit -m "feat(notes): validate and export note properties"
```

---

### Task 5: HTTP endpoints

**Files:**
- Modify: `backend/internal/infrastructure/http/handler/note.go`
- Test: `backend/internal/infrastructure/http/handler/note_test.go`

**Step 1: Write the failing test**

```go
func TestNote_PropertiesRoundTrip(t *testing.T) {
	srv := newTestServer(t)
	resp := mustPost(t, srv.URL+"/api/v1/notes", `{"title":"Props"}`)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, _ := created["id"].(string)

	body := `{"properties":[
		{"key":"rating","type":"number","value":8},
		{"key":"source","type":"text","value":"https://example.com"},
		{"key":"published","type":"date","value":"2024-03-11"},
		{"key":"topics","type":"list","value":["design","tools"]}
	]}`
	put := mustPut(t, srv.URL+"/api/v1/notes/"+id+"/properties", body)
	assert.Equal(t, http.StatusOK, put.StatusCode)
	put.Body.Close()

	get := mustGet(t, srv.URL+"/api/v1/notes/"+id+"/properties")
	assert.Equal(t, http.StatusOK, get.StatusCode)
	var props []map[string]any
	mustDecode(t, get, &props)
	assert.Len(t, props, 4)

	noteResp := mustGet(t, srv.URL+"/api/v1/notes/"+id)
	var noteBody map[string]any
	mustDecode(t, noteResp, &noteBody)
	assert.Len(t, noteBody["properties"], 4)
}

func TestNote_PropertiesValidation(t *testing.T) {
	srv := newTestServer(t)
	resp := mustPost(t, srv.URL+"/api/v1/notes", `{"title":"Props"}`)
	var created map[string]any
	mustDecode(t, resp, &created)
	id, _ := created["id"].(string)

	bad := mustPut(t, srv.URL+"/api/v1/notes/"+id+"/properties",
		`{"properties":[{"key":"rating","type":"number","value":"8"}]}`)
	defer bad.Body.Close()
	assert.Equal(t, http.StatusBadRequest, bad.StatusCode)
}

func TestNote_PropertiesNotFound(t *testing.T) {
	srv := newTestServer(t)
	resp := mustGet(t, srv.URL+"/api/v1/notes/missing/properties")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
```

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/infrastructure/http/handler/ -run TestNote_Properties -v
```

Expected: FAIL — 404 from Chi, route not registered.

**Step 3: Write minimal implementation**

In `Routes()`:

```go
r.Get("/{id}/properties", h.getProperties)
r.Put("/{id}/properties", h.putProperties)
```

Handlers:

```go
func (h *NoteHandler) getProperties(w http.ResponseWriter, r *http.Request) {
	n, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if n == nil {
		writeError(w, http.StatusNotFound, nil)
		return
	}
	props := n.Properties
	if props == nil {
		props = []note.Property{}
	}
	writeJSON(w, http.StatusOK, props)
}

func (h *NoteHandler) putProperties(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Properties []note.Property `json:"properties"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	props, err := h.svc.SetProperties(r.Context(), chi.URLParam(r, "id"), body.Properties)
	if err != nil {
		var valErr *note.ValidationError
		if errors.As(err, &valErr) {
			writeError(w, http.StatusBadRequest, valErr)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if props == nil {
		writeError(w, http.StatusNotFound, nil)
		return
	}
	writeJSON(w, http.StatusOK, props)
}
```

**Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/infrastructure/http/handler/ -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/http/handler/note.go backend/internal/infrastructure/http/handler/note_test.go
git commit -m "feat(notes): expose property endpoints"
```

---

### Task 6: Frontend types, mutation, and panel

**Files:**
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/hooks/useNotes.ts`
- Modify: `frontend/src/components/molecules/PropertiesPanel.tsx`
- Modify: `frontend/src/components/organisms/NoteEditorPanel.tsx`
- Modify: `frontend/src/pages/Notes/index.tsx`

**Step 1: Add types**

```ts
export type PropertyType = 'text' | 'number' | 'date' | 'boolean' | 'list'

export interface NoteProperty {
  key: string
  type: PropertyType
  value: string | number | boolean | string[]
}
```

Add `properties: NoteProperty[]` to `Note`.

**Step 2: Add the mutation**

In `useNotes.ts`:

```ts
export function useSetNoteProperties() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, properties }: { id: string; properties: NoteProperty[] }) =>
      http.put<NoteProperty[]>(`/notes/${id}/properties`, { properties }).then(r => r.data),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['note', variables.id] })
      qc.invalidateQueries({ queryKey: ['notes'] })
    },
  })
}
```

**Step 3: Make the panel generic**

`PropertiesPanel` receives `properties` and `onPropertiesChange`. Keep the existing status/tags/folder controls untouched. Below them:

- List of rows: key label, input bound to the value, remove button.
- `Add property` reveals a key input and type select; append with `defaultPropertyValue(type)`.
- Normalize keys in the UI with the same regex before adding: `^[a-z][a-z0-9_-]{0,63}$`.

Value rendering:

| Type | Input | Parse |
| --- | --- | --- |
| `text` | text | string |
| `number` | number | `Number(value)` |
| `date` | date | string |
| `boolean` | checkbox | boolean |
| `list` | text | `value.split(',').map(s => s.trim()).filter(Boolean)` |

**Step 4: Wire the editor**

Extend `NoteDraft` with `properties: NoteProperty[]`, and change the save payload to:

```ts
export interface NoteSavePayload {
  data: NoteInput
  properties: NoteProperty[]
}
```

`flush()` calls `onSaveRef.current({ data, properties })`. `NotesPage` handles it:

```tsx
onSave={(payload) => {
  updateNote.mutate({ id: selectedNote.id, data: payload.data })
  setProperties.mutate({ id: selectedNote.id, properties: payload.properties })
}}
```

**Step 5: Run checks**

```bash
cd frontend && pnpm run build
```

Expected: TypeScript build passes.

**Step 6: Commit**

```bash
git add frontend/src/types/index.ts frontend/src/hooks/useNotes.ts frontend/src/components/molecules/PropertiesPanel.tsx frontend/src/components/organisms/NoteEditorPanel.tsx frontend/src/pages/Notes/index.tsx
git commit -m "feat(notes): edit typed properties in the note panel"
```

---

### Task 7: Verification

**Step 1: Full backend suite**

```bash
cd backend && go test ./...
```

Expected: PASS.

**Step 2: Migration on the real database copy**

```bash
cd backend && cp ../hube.db /tmp/hube-properties-check.db
DB_PATH=/tmp/hube-properties-check.db go run ./cmd/api
```

Open a note, add `rating` (number), `source` (text), and `published` (date), reload, and confirm values persist. Confirm an old note shows no properties and does not error.

**Step 3: Agent export**

```bash
curl -s localhost:8090/api/v1/notes/<id>/export | head -20
```

Expected: frontmatter includes `rating: 8` and the other properties.

**Step 4: Final commit if verification changed anything**

```bash
git add -A backend frontend docs
git commit -m "chore(notes): verify property persistence"
```

---

## Out of Scope

- Saved views and filters over properties.
- Link-typed properties.
- Property definitions, option lists, and defaults.
- Renaming a property key across notes.
- Web clipping (next slice, writes `source`/`author`/`published`).
