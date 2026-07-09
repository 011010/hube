# hube Enhancement Design

Date: 2026-07-06
Status: Approved

## Overview

Extend hube with flexible kanban/table views, a Notion-like notes experience with block editor and graph view, richer diagram capabilities, agent-friendly content export, and an extended CLI.

## Goals

1. Add persistent kanban/table view toggle for Tasks, Projects, and Notes.
2. Redesign Notes with fixed properties (status, priority, due date) and a TipTap/BlockNote block editor.
3. Provide an Obsidian-like graph view showing Notes, Tasks, and Projects as nodes with relationships as edges.
4. Improve the Network page so it supports general-purpose diagrams, templates, and richer node types.
5. Make note content easily consumable by AI agents via YAML frontmatter export.
6. Extend the CLI from read-only listings to basic CRUD and agent-friendly export.

## Backend Changes

### Settings: View Preferences

The `settings` table is a key/value store. Store `view_preferences` as a JSON string under the key `general.view_preferences`:

```json
{
  "tasks_view": "kanban|table",
  "projects_view": "kanban|table",
  "notes_view": "kanban|table"
}
```

- Default: `"kanban"` for all three.
- Updated via existing `PUT /settings` endpoint.
- The handler exposes `view_preferences` as a string field inside the `general` settings object.
- Returned via existing `GET /settings` endpoint.

### Notes: New Fields

SQLite migration adds:

| Column     | Type    | Values                              |
|------------|---------|-------------------------------------|
| status     | TEXT    | `draft`, `in_progress`, `published` |
| priority   | TEXT    | `low`, `medium`, `high`             |
| due_date   | TEXT    | ISO date or null                    |
| blocks     | TEXT    | TipTap/BlockNote JSON string        |

Domain entity and repository updated accordingly. FTS triggers continue to index `title` and a plain-text extraction of `blocks`.

### Graph Relationships

Add table `note_links`:

```sql
CREATE TABLE note_links (
  note_id TEXT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  target_note_id TEXT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  PRIMARY KEY (note_id, target_note_id)
);
```

- Populated by parsing `[[note-title-or-id]]` syntax from note blocks on save.
- Optional `note_id` on `tasks` and `projects` to link them to notes.
- Graph API endpoint returns nodes (notes, tasks, projects) and edges (note-note, note-task, note-project, project-task).

## Frontend Components

### Reusable Components

- `ViewToggle`: kanban/table switch bound to settings mutation.
- `KanbanBoard`: columns by status, draggable cards, counters.
- `DataTable`: sortable table with action column.
- `PropertiesPanel`: status, priority, due date, tags, folder selectors.
- `BlockEditor`: TipTap/BlockNote wrapper.
- `GraphView`: React Flow canvas rendering notes/tasks/projects graph.

### Tasks & Projects Pages

Replace current list/grid with:

```
PageHeader
ViewToggle
{view === 'kanban' ? <KanbanBoard /> : <DataTable />}
```

Cards in kanban allow inline status change via drag-and-drop.

### Notes Page Redesign

Three-pane layout:

1. **Sidebar**: folders + search + filters (status, priority).
2. **Note list**: title, status badge, priority, updated date.
3. **Editor area**:
   - Properties panel
   - Tabs/switch: Editor / Graph View
   - TipTap block editor

### Network Page Redesign

Rename concept from "Network Diagrams" to "Diagrams":

- Templates: blank, flowchart, mind map, architecture, ER diagram.
- Node types: process, decision, input/output, person, document, idea, image, plus existing network types.
- Toolbar: node search, labeled connectors, color/shape picker.
- Export: JSON, PNG, SVG.

## Agent-Friendly Content

Each note can be exported as Markdown with YAML frontmatter:

```markdown
---
title: "Note Title"
status: in_progress
priority: high
due_date: 2026-07-10
tags: [ai, planning]
backlinks: [note-b, note-c]
---

# Note Title

Clean markdown body...
```

Used by:
- CLI `hube export notes --format agent`
- Backend export endpoint extension
- RAG indexing context

## CLI Extensions

New commands under `backend/cmd/cli`:

- `hube note create --title "X" --status draft --priority medium`
- `hube note show <id> --format yaml|json`
- `hube note update <id> --status published`
- `hube task create "Title" --priority high --due 2026-07-10`
- `hube task done <id>`
- `hube export notes --format agent`
- `hube graph` (optional): output graph nodes/edges as JSON

## Implementation Phases

1. **View toggle + kanban/table for Tasks and Projects**
   - Backend settings migration
   - `ViewToggle`, `KanbanBoard`, `DataTable` components
   - Refactor Tasks and Projects pages

2. **Notes backend + properties**
   - Add note fields migration
   - Update domain, repo, handler, tests
   - Add properties panel to Notes UI

3. **Notes block editor**
   - Add TipTap/BlockNote dependency
   - Store blocks as JSON, render plain text for preview/FTS
   - Replace markdown textarea

4. **Graph view + backlinks**
   - Parse `[[note]]` links
   - Add `note_links` table and graph endpoint
   - Build `GraphView` component

5. **Network improvements**
   - New node types and templates
   - Toolbar redesign
   - PNG/SVG export

6. **Agent export + CLI**
   - YAML frontmatter export
   - CLI CRUD commands

## Open Questions / Notes

- Block editor library final choice: TipTap vs BlockNote. Both are valid; final decision during implementation planning.
- Drag-and-drop library: `@dnd-kit/core` is recommended for kanban.
- Graph rendering: React Flow is already a project dependency.
- Migration path: existing markdown notes will be migrated by wrapping their content in a single paragraph block.
