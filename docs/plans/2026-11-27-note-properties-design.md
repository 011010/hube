# Note Properties Design

Date: 2026-11-27
Status: Implemented on `feat/note-properties`

## Overview

Add generic, typed properties to notes. Today a note only carries the fixed
columns `status`, `priority`, `due_date`, `tags`, and `folder_id`. Anything
else (a book rating, a clipping source, an author) has nowhere to live, so it
ends up as unstructured text inside the body and cannot be filtered, sorted, or
exported as metadata.

This is the enabling slice for future work: web clipping needs a place to store
`source`, `author`, and `published`; saved views need fields to filter on; the
agent export needs structured frontmatter.

## Goals

1. Store arbitrary key/value metadata on a note with an explicit type.
2. Read and write properties through the existing note API without changing
   the note autosave flow.
3. Validate values on write so the UI and future filters can trust the type.
4. Include properties in the agent Markdown export.
5. Keep all existing notes working with no data migration.

## Non-Goals

- Saved views / Bases (filters, sorts, grouping) — next slice.
- Formula properties.
- Property definitions with option lists, descriptions, or defaults.
- Link-typed properties. They need rename cascades and title resolution; add
  them once saved views exist and the need is proven.
- Two-way sync with an Obsidian vault.

## Data Model

New migration `013_note_properties.sql`:

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

`value` is JSON-encoded. A single column keeps the schema stable while the type
discriminator tells readers how to decode it:

| Type      | JSON value           | Example                          |
| --------- | -------------------- | -------------------------------- |
| `text`    | string               | `"Kepano"`                       |
| `number`  | number               | `8`                              |
| `date`    | string `YYYY-MM-DD`  | `"2024-03-11"`                   |
| `boolean` | boolean              | `true`                           |
| `list`    | array of strings     | `["design", "tools"]`            |

Notes:

- The existing file `009_note_properties.sql` is already applied and only adds
  fixed columns. It cannot be reused; the table ships in a new migration.
- `PRIMARY KEY (note_id, key)` makes a property unique per note and gives the
  bulk-replace write a natural upsert target.
- The index on `key` is not needed by this slice but costs little and is the
  access path saved views will use.

## Domain

`backend/internal/domain/note/entity.go`:

```go
type Property struct {
    Key   string `json:"key"  db:"key"`
    Type  string `json:"type" db:"type"`
    Value any    `json:"value"`
}

const (
    PropertyText    = "text"
    PropertyNumber  = "number"
    PropertyDate    = "date"
    PropertyBoolean = "boolean"
    PropertyList    = "list"
)
```

`Note` gains `Properties []Property \`json:"properties"\``. It is loaded the same
way tags are, with a second query per note.

`ValidateProperty` lives in the domain package:

- `key` is trimmed, lowercased, and must match `^[a-z][a-z0-9_-]{0,63}$`.
- At most 50 properties per note.
- The value must decode into the declared type.
- `date` must parse as `YYYY-MM-DD`.
- `text` is capped at 10,000 characters; `list` at 100 entries of 500
  characters each.

## Repository

`note.Repository` gains:

```go
SetProperties(ctx context.Context, noteID string, props []note.Property) error
```

The implementation runs in one transaction: `DELETE FROM note_properties WHERE
note_id = ?`, then inserts the new set. This matches the editor's bulk autosave
and avoids per-key endpoints.

Reads:

- `FindByID` loads properties for the single note.
- `FindAll` and `Search` batch-load properties in one query and group them by
  note ID, so adding properties does not double the existing per-note tag
  queries in list responses.

## Application

`note.Service` gains:

```go
SetProperties(ctx context.Context, noteID string, props []note.Property) ([]note.Property, error)
```

It checks that the note exists, normalizes keys, validates every property, and
returns the stored set.

`ExportAgentMarkdown` writes properties into the YAML frontmatter after the
built-in fields. Keys are emitted as-is; values use the existing quoting helper.

## HTTP API

```
GET /api/notes/{id}/properties   -> [{key, type, value}]
PUT /api/notes/{id}/properties   -> [{key, type, value}]
```

- `PUT` replaces the full set. An empty array clears it.
- Request body: `{"properties": [...]}`.
- `400` on invalid key, unknown type, or value that does not match the type.
- `404` when the note does not exist.
- `Note` responses from `GET /api/notes` and `GET /api/notes/{id}` include
  `properties`, so the editor does not need a second fetch.

## Frontend

`PropertiesPanel` stops being hardcoded. It keeps `Status`, `Tags`, and
`Folder` as first-class controls, then renders a `Properties` section below
them:

- One row per property: key, typed value input, remove button.
- `Add property` opens a small inline form with a key input and a type select.
- Value input switches on type: text input, number input, date input, checkbox,
  and a comma-separated input for lists.
- Writes are debounced with the same 800 ms autosave used by the editor and go
  to `PUT /notes/{id}/properties`.
- Values render read-only in the note card preview as a compact list.

Types exposed in the UI: text, number, date, boolean, list.

## Testing

- Domain: table test for `ValidateProperty` covering each type, bad keys, and
  type mismatches.
- Repository: create a note, set properties, read them back, replace the set,
  clear it, and confirm `ON DELETE CASCADE` removes rows.
- Handler: `PUT` then `GET` round trip; `400` and `404` cases.
- Export: a note with properties produces valid frontmatter.
- Migration test: an existing database at migration 012 upgrades cleanly and
  existing notes have an empty property list.

## Acceptance Criteria

1. Add `rating: 8` (number) to a note, reload, and the value persists.
2. Add `source` (text) and `published` (date); both render with the right input.
3. Remove a property and it is gone after reload.
4. Adding a property with an invalid key or a mismatched value returns `400`.
5. Existing notes open unchanged and report `properties: []`.
6. `GET /api/notes/{id}/export` includes the properties in frontmatter.

## Follow-Ups

1. Saved views (Bases-lite): filter, sort, and group by properties.
2. Clip endpoint that writes `source`, `author`, `published`, and `domain`.
3. Backlinks panel and `[[` autocomplete.
4. JSON Canvas import/export for the existing diagram feature.

## Implementation Notes (final)

These decisions were made during implementation and supersede the sections
above where they differ.

### API contract

- `PUT /notes/{id}/properties` requires an explicit JSON array. `[]` clears the
  set; `{}` and `{"properties":null}` return `400`. This prevents a client typo
  from silently wiping metadata.
- `POST /notes` rejects a non-empty `properties` array with `400`; properties
  are written only through the properties endpoint.
- `PUT /notes/{id}` ignores `properties` in the body and returns the stored set.
  It never persists or echoes a property payload.
- Every note response includes `properties`, always as `[]` rather than `null`.

### Validation

- Keys are trimmed and lowercased, then checked against
  `^[a-z][a-z0-9_-]{0,63}$`; duplicates are detected after normalization.
- Reserved keys `title`, `status`, `tags`, `created_at`, and `updated_at` are
  rejected because they collide with the fixed YAML frontmatter keys.
- Limits: 50 properties per note, 10,000 runes of text, 100 list items of 500
  runes each. Number values must be finite `float64`; NaN and ±Inf are rejected.
- `Property.Validate` has a value receiver. `ValidateProperties` normalizes the
  caller's slice in place and then validates it.

### Persistence

- The migration file is `013_note_properties_table.sql`; the earlier `009` only
  adds fixed note columns.
- Reads use a separate `propertyRow` and decode each JSON value into `any`.
- `FindAll` and `Search` batch-load properties in chunks of 500 note IDs, one
  query per chunk.
- `SetProperties` runs in a transaction: touch `notes.updated_at`, delete the
  old set, insert the new one. A missing note returns `sql.ErrNoRows`, which the
  service maps to the same `(nil, nil)` 404 contract. A mid-insert failure rolls
  back the delete and preserves the previous set.
- An upgrade test applies migrations 001–012, inserts a note, then runs 013 and
  asserts that the note survives with an empty property set and that re-running
  the migrator is a no-op.

### Export

- Property keys are quoted in the YAML frontmatter, and every scalar uses
  `strconv.Quote`, so newlines and control characters round-trip losslessly.
- Empty lists emit `key: []`.
- A reserved key or a non-list value for a list property makes the export fail
  instead of emitting invalid YAML.

### Frontend

- `PropertiesPanel` keeps status, folder, and tags as first-class controls and
  adds a typed property editor below them.
- Comma-separated list and tag inputs keep a local text buffer so commas can be
  typed; the value is parsed on each keystroke and normalized on blur.
- The number input uses a text buffer with `inputMode="decimal"` so negative and
  decimal values can be typed.
- Autosave writes only the resource that changed: `data` for note fields,
  `properties` for the property set.
- Save failures appear in a page-level alert and in the editor indicator; the
  draft is not discarded.

### Known limitations

- Bulk replace is last-writer-wins; there is no ETag or version check.
- A failed autosave is surfaced but not retried automatically; the draft keeps
  the unsaved values until the panel is closed.
- A stored list item containing a comma cannot be edited faithfully in the
  comma-separated input.
- Corrupt property rows from manual database edits make reads and exports fail
  instead of being skipped.
- The frontend logic has no component tests; backend coverage is the contract.

