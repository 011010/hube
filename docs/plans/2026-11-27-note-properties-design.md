# Note Properties Design

Date: 2026-11-27
Status: Approved

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
