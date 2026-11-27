# Note Views (Bases-lite) Design

Date: 2026-11-27
Status: Approved

## Overview

Saved views for the Notes page: named combinations of filters, property
conditions, sort, layout, and an optional read-only grouping. This is the
Obsidian Bases concept reduced to what hube needs — no query language, no
formulas, no `.base` files. Views are evaluated in memory against the already
loaded notes list.

## Goals

1. Save, switch, edit, and delete named views.
2. Filter by status, tags, folder, and typed property conditions.
3. Sort by a built-in field or a property key.
4. Optionally group the table by a property into labeled sections.
5. Persist views per user with no new tables or endpoints.

## Non-Goals

- Kanban grouping by property, and drag-to-update (follow-up).
- Server-side filtering or a query language.
- Formulas, summaries, or computed columns.
- Sharing views between users.

## Persistence

The settings store is a flat key/value table. Views live under
`general.notes_views` as a JSON string, exactly like `general.view_preferences`.

Backend change: add `notes_views` to the settings request/response structs and
map it to the key. `PUT /settings` is full replacement, matching the existing
contract. No migration.

## View Schema

```ts
type PropertyOp =
  | 'contains' | 'not_contains' | 'equals'         // text
  | 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte'      // number
  | 'before' | 'on' | 'after'                       // date
  | 'is_true' | 'is_false'                           // boolean
  | 'includes' | 'not_includes'                       // list
  | 'is_set' | 'is_not_set'                          // any

interface NoteViewCondition {
  key: string                 // property key
  op: PropertyOp
  value?: string | number | boolean
}

interface NoteView {
  id: string
  name: string
  filters: {
    status?: NoteStatus[]
    tags?: string[]           // all must match (AND)
    folder_id?: string | null
    properties: NoteViewCondition[]
  }
  sort: {
    field: 'updated_at' | 'created_at' | 'title' | 'status' | string  // property key
    direction: 'asc' | 'desc'
  }
  layout: 'table' | 'kanban'
  groupBy?: string            // property key; table layout only
}
```

Condition semantics: the property type is taken from each note's stored value,
not from the view. An op that does not apply to a note's actual type does not
match that note. This keeps the engine total: two notes can disagree on the
type of the same key without breaking a view.

## Engine

`frontend/src/utils/noteViews.ts`, pure and unit-tested:

- `parseNoteViews(raw: string | undefined): NoteView[]` — drops malformed
  entries, caps at 20 views, trims names to 100 characters, generates ids for
  entries missing one.
- `applyNoteView(notes: Note[], view: NoteView | null): Note[]` — filter, then
  sort. Missing sort values always sort last regardless of direction.
- `groupNotes(notes: Note[], field?: string): { key: string; notes: Note[] }[]`
  — groups by a property key or `status`. Status uses the fixed
  draft/in_progress/published order. Text and list values group per distinct
  value; a note with several list values appears in each matching section.
  Notes without the property land in a final "No value" section.

Text comparison is `localeCompare`; numbers and dates compare numerically.

## Frontend Components

- A view selector in the Notes page header: "Todas las notas" plus each saved
  view, using a native `<select>` consistent with the rest of the app.
- A view editor in the existing `atoms/Modal`: name, status chips, tags,
  folder, property condition rows (type-aware inputs reusing the
  `PropertiesPanel` patterns), sort, layout, group-by, save and delete.
- When a view is active it drives the displayed notes; the search box still
  applies on top of the view.
- Grouping renders a section header above a `DataTable` per group when
  `layout` is `table` and `groupBy` is set. Read-only.
- `useNoteViews()` reads and writes `general.notes_views`, merging with the
  latest cached settings exactly like `useViewPreference` does.

## Testing

- Engine unit tests (vitest): every op, type mismatches, sort direction and
  missing-value placement, grouping multi-membership, malformed JSON parsing.
- Backend handler test: `PUT`/`GET` round-trip for `notes_views`; absent field
  clears it (full replacement contract).
- Frontend build and lint stay clean.

## Acceptance Criteria

1. Create a view "Rating alto" with `rating >= 7` and `topics includes design`,
   sorted by `rating` descending. Reload the app; the view is still there and
   the table shows only matching notes.
2. Group by a text property; sections render per value, list values appear in
   every matching section, unset notes land in "No value".
3. A malformed `notes_views` JSON in settings does not break the page; the
   selector falls back to "Todas las notas".
4. Deleting a view removes it from the selector; switching to "Todas las
   notas" restores the full list.

## Follow-Ups

1. Kanban grouped by property with drag-to-update.
2. Server-side filtering once note count makes client-side evaluation slow.
3. View templates for the clip inbox (e.g. "Unread clippings").
