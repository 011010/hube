# hube Enhancement Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add persistent kanban/table views to Tasks, Projects, and Notes; redesign Notes with Notion-like properties and a TipTap block editor; add an Obsidian-style graph view; improve Network diagrams with templates and richer node types; export agent-friendly YAML frontmatter notes; and extend the CLI with CRUD and export commands.

**Architecture:** Reusable frontend components (`ViewToggle`, `KanbanBoard`, `DataTable`, `PropertiesPanel`, `BlockEditor`, `GraphView`) are composed by Tasks, Projects, and Notes pages. Backend stores view preferences in settings, adds status/priority/due_date/blocks to notes, and builds a graph from explicit links and foreign-key relationships. CLI talks to the HTTP API and outputs tables, JSON, and agent-friendly Markdown.

**Tech Stack:** Go 1.26.4 (Chi, sqlx, SQLite), React 19 + TypeScript + Vite + Tailwind CSS v4, TipTap for block editing, @dnd-kit for drag-and-drop, React Flow for graph rendering, Cobra for CLI.

**PR:** [#4](https://github.com/011010/hube/pull/4) — `feat/hube-enhancement` → `main`.

---

## Phase 1: Persistent View Toggle + Kanban/Table for Tasks and Projects

### Task 1: Add view_preferences to settings backend

**Files:**
- Modify: `backend/internal/infrastructure/http/handler/setting.go`
- Test: `backend/internal/infrastructure/http/handler/setting_test.go` (create if missing)

**Context:** The `settings` table is a key/value store (`key`, `value`). Settings are grouped by dotted keys like `general.display_name` and `integration.monkeyapi_url`. We add `view_preferences` as a JSON string under the key `general.view_preferences`.

**Step 1: Update handler payload**

Add `ViewPreferences string` to the `generalSettings` struct in `backend/internal/infrastructure/http/handler/setting.go` with `json:"view_preferences,omitempty"`.

**Step 2: Wire GET and PUT**

`GET /settings` should return the stored `general.view_preferences` value. `PUT /settings` should persist `general.view_preferences` from the payload.

**Step 3: Add handler test**

Create `backend/internal/infrastructure/http/handler/setting_test.go` with `TestSettings_ViewPreferences`:

- PUT `{"general":{"display_name":"Hube","view_preferences":"{\"tasks_view\":\"kanban\",\"projects_view\":\"table\",\"notes_view\":\"kanban\"}"}}`
- Assert the response echoes `view_preferences`.
- GET `/settings` and assert `view_preferences` was persisted.

**Step 4: Run tests**

Run: `cd backend && go test ./internal/infrastructure/http/handler -run TestSettings_ViewPreferences -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/http/handler/setting.go backend/internal/infrastructure/http/handler/setting_test.go
git commit -m "feat(settings): add view_preferences to general settings"
```

---

### Task 2: Add ViewToggle component

**Files:**
- Create: `frontend/src/components/molecules/ViewToggle.tsx`
- Test: manual + TypeScript compile

**Step 1: Implement ViewToggle**

```tsx
import { LayoutGrid, Table2 } from 'lucide-react'
import { IconButton } from '../atoms/IconButton'

export type ViewMode = 'kanban' | 'table'

export function ViewToggle({ value, onChange }: { value: ViewMode; onChange: (v: ViewMode) => void }) {
  return (
    <div className="flex items-center bg-surface-elevated border border-border rounded-lg p-0.5">
      <IconButton
        icon={<LayoutGrid size={16} />}
        aria-label="Kanban"
        variant={value === 'kanban' ? 'primary' : 'ghost'}
        onClick={() => onChange('kanban')}
      />
      <IconButton
        icon={<Table2 size={16} />}
        aria-label="Table"
        variant={value === 'table' ? 'primary' : 'ghost'}
        onClick={() => onChange('table')}
      />
    </div>
  )
}
```

**Step 2: Update Button/IconButton variants if needed**

Ensure `IconButton` accepts a `primary` variant. Modify `frontend/src/components/atoms/IconButton.tsx` if it does not.

**Step 3: TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: no errors

**Step 4: Commit**

```bash
git add frontend/src/components/molecules/ViewToggle.tsx
git commit -m "feat(ui): add ViewToggle molecule"
```

---

### Task 3: Add useViewPreference hook

**Files:**
- Create: `frontend/src/hooks/useViewPreference.ts`

**Step 1: Update Settings type**

Modify `frontend/src/hooks/useSettings.ts` to add `view_preferences?: string` to `GeneralSettings`.

**Step 2: Implement hook**

```tsx
import { useSettings, useUpdateSettings } from './useSettings'

export type PageViewKey = 'tasks_view' | 'projects_view' | 'notes_view'

function parsePrefs(raw?: string): Record<PageViewKey, 'kanban' | 'table'> {
  try {
    return JSON.parse(raw ?? '{}')
  } catch {
    return {} as Record<PageViewKey, 'kanban' | 'table'>
  }
}

export function useViewPreference(key: PageViewKey) {
  const { data } = useSettings()
  const update = useUpdateSettings()

  const prefs = parsePrefs(data?.general.view_preferences)
  const value = prefs[key] ?? 'kanban'

  function setValue(v: 'kanban' | 'table') {
    if (!data) return
    const next = { ...prefs, [key]: v }
    update.mutate({
      ...data,
      general: { ...data.general, view_preferences: JSON.stringify(next) },
    })
  }

  return [value, setValue] as const
}
```

**Step 3: TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: no errors

**Step 4: Commit**

```bash
git add frontend/src/hooks/useViewPreference.ts frontend/src/hooks/useSettings.ts
git commit -m "feat(settings): add useViewPreference hook"
```

---

### Task 4: Add KanbanBoard component

**Files:**
- Create: `frontend/src/components/organisms/KanbanBoard.tsx`
- Install: `cd frontend && pnpm add @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities`

**Step 1: Install dependencies**

Run: `cd frontend && pnpm add @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities`
Expected: packages installed

**Step 2: Implement KanbanBoard**

Generic component: columns by `status`, cards as render prop, on status change callback.

```tsx
import { DndContext, DragOverlay, PointerSensor, useSensor, useSensors, type DragEndEvent } from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy, useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { useState } from 'react'

export interface KanbanColumn<T> {
  id: string
  title: string
  items: T[]
}

export interface KanbanBoardProps<T> {
  columns: KanbanColumn<T>[]
  renderCard: (item: T) => React.ReactNode
  getItemId: (item: T) => string
  onMove: (itemId: string, sourceColumnId: string, targetColumnId: string) => void
}

export function KanbanBoard<T>(props: KanbanBoardProps<T>) { /* ... */ }
```

**Step 3: TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: no errors

**Step 4: Commit**

```bash
git add frontend/src/components/organisms/KanbanBoard.tsx frontend/package.json frontend/pnpm-lock.yaml
git commit -m "feat(ui): add KanbanBoard organism with dnd-kit"
```

---

### Task 5: Add DataTable component

**Files:**
- Create: `frontend/src/components/organisms/DataTable.tsx`

**Step 1: Implement DataTable**

Generic table with sortable headers. Simplified version for now.

```tsx
export interface DataTableColumn<T> {
  key: string
  header: string
  render: (item: T) => React.ReactNode
}

export interface DataTableProps<T> {
  columns: DataTableColumn<T>[]
  data: T[]
}

export function DataTable<T>({ columns, data }: DataTableProps<T>) { /* ... */ }
```

**Step 2: TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: no errors

**Step 3: Commit**

```bash
git add frontend/src/components/organisms/DataTable.tsx
git commit -m "feat(ui): add DataTable organism"
```

---

### Task 6: Refactor TasksPage with kanban/table views

**Files:**
- Modify: `frontend/src/pages/Tasks/index.tsx`
- Modify: `frontend/src/hooks/useTasks.ts` (ensure update accepts status change)

**Step 1: Use ViewToggle + useViewPreference**

```tsx
const [view, setView] = useViewPreference('tasks_view')
```

**Step 2: Build kanban columns**

```tsx
const columns: KanbanColumn<Task>[] = [
  { id: 'todo', title: 'To Do', items: tasks.filter(t => t.status === 'todo') },
  { id: 'in_progress', title: 'In Progress', items: tasks.filter(t => t.status === 'in_progress') },
  { id: 'done', title: 'Done', items: tasks.filter(t => t.status === 'done') },
]
```

**Step 3: Render kanban or table**

Replace grouped list with conditional render.

**Step 4: Build table**

Run: `cd frontend && npx tsc --noEmit`
Expected: no errors

**Step 5: Commit**

```bash
git add frontend/src/pages/Tasks/index.tsx
git commit -m "feat(tasks): add kanban and table views"
```

---

### Task 7: Refactor ProjectsPage with kanban/table views

**Files:**
- Modify: `frontend/src/pages/Projects/index.tsx`

**Step 1: Use ViewToggle + useViewPreference**

```tsx
const [view, setView] = useViewPreference('projects_view')
```

**Step 2: Build kanban columns by project status**

`planning`, `active`, `completed`, `on_hold`.

**Step 3: Render kanban or table**

Replace grid with conditional render.

**Step 4: TypeScript check + build**

Run: `cd frontend && npx tsc --noEmit && npx vite build`
Expected: no errors

**Step 5: Commit**

```bash
git add frontend/src/pages/Projects/index.tsx
git commit -m "feat(projects): add kanban and table views"
```

---

## Phase 2: Notes Backend — Properties and Block Storage

### Task 8: Add notes migration for status/priority/due_date/blocks

**Files:**
- Create: `backend/internal/infrastructure/sqlite/migrations/009_note_properties.sql`
- Modify: `backend/internal/infrastructure/sqlite/db.go` (register migration)
- Create: `backend/internal/infrastructure/sqlite/migrations_test.go`

**Step 1: Write migration**

Create `backend/internal/infrastructure/sqlite/migrations/009_note_properties.sql`:

```sql
ALTER TABLE notes ADD COLUMN status TEXT NOT NULL DEFAULT 'draft';
ALTER TABLE notes ADD COLUMN priority TEXT NOT NULL DEFAULT 'medium';
ALTER TABLE notes ADD COLUMN due_date TEXT;
ALTER TABLE notes ADD COLUMN blocks TEXT NOT NULL DEFAULT '';
```

**Step 2: Register migration and expose `RunMigrations`**

In `backend/internal/infrastructure/sqlite/db.go`:
- Add `//go:embed migrations/009_note_properties.sql` and a `var notePropertiesSQL string`.
- Append `{"009_note_properties", notePropertiesSQL}` to the `migrations` slice.
- Extract the migration loop into an exported `RunMigrations(db *sqlx.DB) error` helper so tests can run migrations directly.

**Step 3: Add migration tests**

Create `backend/internal/infrastructure/sqlite/migrations_test.go` with:
- `TestMigrations_NoteProperties`: open a fresh DB and assert each of the four new columns exists.
- `TestMigrations_NotePropertiesUpgrade`: create a pre-migration `notes` table, insert a row, run `RunMigrations`, and assert the row receives the correct defaults (`status='draft'`, `priority='medium'`, `blocks=''`, `due_date=NULL`).

**Step 4: Verify**

Run: `cd backend && go test ./internal/infrastructure/sqlite -v`
Expected: PASS

Also run existing note tests: `cd backend && go test ./internal/infrastructure/http/handler -run TestNote -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/sqlite/migrations/009_note_properties.sql backend/internal/infrastructure/sqlite/db.go backend/internal/infrastructure/sqlite/migrations_test.go
git commit -m "feat(notes): add status priority due_date blocks columns"
```

**Note:** Do NOT modify the FTS triggers in `002_notes.sql`. The existing `content` column will continue to hold the plain-text extraction of `blocks`, and FTS will keep indexing `title` and `content`. Plain-text extraction is implemented in Task 10.

---

### Task 9: Update Note domain entity

**Files:**
- Modify: `backend/internal/domain/note/entity.go`

**Step 1: Add fields**

```go
type Note struct {
	ID        string    `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	Blocks    string    `json:"blocks" db:"blocks"`
	Status    string    `json:"status" db:"status"`
	Priority  string    `json:"priority" db:"priority"`
	DueDate   *string   `json:"due_date" db:"due_date"`
	FolderID  *string   `json:"folder_id" db:"folder_id"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

**Step 2: Commit**

```bash
git add backend/internal/domain/note/entity.go
git commit -m "feat(notes): extend Note entity with properties"
```

---

### Task 10: Update NoteRepo queries

**Files:**
- Modify: `backend/internal/infrastructure/sqlite/note_repo.go`

**Step 1: Update noteRow and queries**

Include `status`, `priority`, `due_date`, `blocks` in SELECT, INSERT, UPDATE.

**Step 2: Add plain-text extraction helper**

```go
func blocksToText(blocks string) string { /* parse TipTap JSON, return plain text */ }
```

On `Create` and `Update`, extract plain text from `blocks` and store it in the existing `content` column. The FTS triggers in `002_notes.sql` continue to index `title` and `content`, so no trigger changes are needed.

**Step 3: Update tests**

Modify `backend/internal/infrastructure/http/handler/note_test.go` to include new fields.

**Step 4: Run tests**

Run: `cd backend && go test ./...`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/sqlite/note_repo.go backend/internal/infrastructure/http/handler/note_test.go
git commit -m "feat(notes): repository supports new fields"
```

---

### Task 11: Add note validation

**Files:**
- Modify: `backend/internal/application/note/service.go`

**Step 1: Validate status/priority values**

```go
var validNoteStatuses = map[string]bool{"draft": true, "in_progress": true, "published": true}
var validPriorities = map[string]bool{"low": true, "medium": true, "high": true}
```

**Step 2: Default empty status/priority**

**Step 3: Run tests**

Run: `cd backend && go test ./internal/application/note/...`
Expected: PASS

**Step 4: Commit**

```bash
git add backend/internal/application/note/service.go
git commit -m "feat(notes): validate status and priority"
```

---

## Phase 3: Notes Frontend — Properties Panel and Block Editor

### Task 12: Install TipTap dependencies

**Files:**
- Modify: `frontend/package.json`

**Step 1: Install packages**

Run: `cd frontend && pnpm add @tiptap/react @tiptap/starter-kit @tiptap/extension-placeholder @tiptap/extension-task-list @tiptap/extension-task-item @tiptap/extension-underline`

**Step 2: Commit lockfile**

```bash
git add frontend/package.json frontend/pnpm-lock.yaml
git commit -m "chore(deps): add tiptap editor packages"
```

---

### Task 13: Update frontend Note type

**Files:**
- Modify: `frontend/src/types/index.ts`

**Step 1: Add fields**

```ts
export type NoteStatus = 'draft' | 'in_progress' | 'published'

export interface Note {
  id: string
  title: string
  content: string
  blocks: string
  status: NoteStatus
  priority: Priority
  due_date: string | null
  folder_id: string | null
  tags: string[]
  created_at: string
  updated_at: string
}
```

**Step 2: Commit**

```bash
git add frontend/src/types/index.ts
git commit -m "feat(types): add Note properties and blocks"
```

---

### Task 14: Add PropertiesPanel component

**Files:**
- Create: `frontend/src/components/molecules/PropertiesPanel.tsx`

**Step 1: Implement panel**

Inputs for status, priority, due_date, tags, folder. Emits onChange for each field.

**Step 2: Commit**

```bash
git add frontend/src/components/molecules/PropertiesPanel.tsx
git commit -m "feat(ui): add PropertiesPanel for notes"
```

---

### Task 15: Add BlockEditor component

**Files:**
- Create: `frontend/src/components/organisms/BlockEditor.tsx`
- Create: `frontend/src/utils/blocks.ts` (plain-text extraction)

**Step 1: Implement editor**

```tsx
import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'

export function BlockEditor({ value, onChange }: { value: string; onChange: (json: string, plainText: string) => void }) { /* ... */ }
```

**Step 2: Add plain-text extraction**

```ts
export function blocksToText(json: string): string { /* walk TipTap doc */ }
```

**Step 3: TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: no errors

**Step 4: Commit**

```bash
git add frontend/src/components/organisms/BlockEditor.tsx frontend/src/utils/blocks.ts
git commit -m "feat(notes): add TipTap block editor"
```

---

### Task 16: Refactor NotesPage with properties, editor, and kanban/table

**Files:**
- Modify: `frontend/src/pages/Notes/index.tsx`
- Modify: `frontend/src/hooks/useNotes.ts` to send new fields

**Step 1: Update hooks**

Ensure create/update payloads include `status`, `priority`, `due_date`, `blocks`.

**Step 2: Add view toggle and kanban/table to note list**

Kanban columns: `draft`, `in_progress`, `published`.
Table columns: title, status, priority, due date, updated.

**Step 3: Integrate PropertiesPanel and BlockEditor**

Replace markdown textarea with BlockEditor; add PropertiesPanel above it.

**Step 4: Preview plain text**

Use `blocksToText` for note list preview.

**Step 5: Build**

Run: `cd frontend && npx tsc --noEmit && npx vite build`
Expected: no errors

**Step 6: Commit**

```bash
git add frontend/src/pages/Notes/index.tsx frontend/src/hooks/useNotes.ts
git commit -m "feat(notes): notion-like editor with properties and kanban/table"
```

---

## Phase 4: Graph View + Backlinks

**Status: ✅ Done** — commits `1078d0e`, `c1a3749`, `62ab0ba`, `501f513` on `feat/hube-enhancement`.

### Task 17: Add note_links table and parsing

**Status:** ✅ Done (`1078d0e`)

**Files:**
- Create: `backend/internal/infrastructure/sqlite/migrations/010_note_links.sql` (worktree was only up to `009`, so this landed as `010`, not `011` as originally planned)
- Modify: `backend/internal/infrastructure/sqlite/note_repo.go`
- Modify: `backend/internal/domain/note/repository.go`

**Step 1: Create migration**

```sql
CREATE TABLE note_links (
  note_id TEXT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  target_note_id TEXT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  PRIMARY KEY (note_id, target_note_id)
);
```

**Step 2: Parse [[...]] syntax**

```go
func extractLinks(blocks string) []string { /* regex \[\[(.+?)\]\] */ }
```

**Step 3: Store links on create/update**

In repo, after saving note, delete old links and insert new ones.

**Step 4: Run tests**

Run: `cd backend && go test ./...`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/sqlite/migrations/011_note_links.sql backend/internal/infrastructure/sqlite/note_repo.go backend/internal/domain/note/repository.go
git commit -m "feat(notes): store note-to-note links"
```

---

### Task 18: Add task/project note_id fields

**Status:** ✅ Done (`c1a3749`)

**Files:**
- Create: `backend/internal/infrastructure/sqlite/migrations/011_task_project_note_id.sql` (follows on from `010_note_links.sql`, not `012` as originally planned)
- Modify: `backend/internal/domain/task/entity.go`
- Modify: `backend/internal/domain/project/entity.go`
- Modify repos (`task_repo.go`, `project_repo.go`) — handlers needed no changes, they already do full struct decode/encode

**Step 1: Migration**

```sql
ALTER TABLE tasks ADD COLUMN note_id TEXT REFERENCES notes(id) ON DELETE SET NULL;
ALTER TABLE projects ADD COLUMN note_id TEXT REFERENCES notes(id) ON DELETE SET NULL;
```

**Step 2: Update domain and repo**

Add `note_id` to entities and queries.

**Step 3: Update frontend types**

Add `note_id` to Task and Project types.

**Step 4: Run tests**

Run: `cd backend && go test ./...`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/sqlite/migrations/012_task_project_note_id.sql backend/internal/domain/task/entity.go backend/internal/domain/project/entity.go backend/internal/infrastructure/sqlite/task_repo.go backend/internal/infrastructure/sqlite/project_repo.go backend/internal/infrastructure/http/handler/task.go backend/internal/infrastructure/http/handler/project.go frontend/src/types/index.ts
git commit -m "feat(tasks,projects): add optional note_id link"
```

---

### Task 19: Add graph API endpoint

**Status:** ✅ Done (`62ab0ba`)

**Files:**
- Modify: `backend/internal/infrastructure/http/handler/note.go`
- Modify: `backend/internal/infrastructure/sqlite/note_repo.go`
- Modify: `backend/internal/domain/note/repository.go`

**Step 1: Add Graph method to repo**

Returns nodes and edges from notes, tasks, projects, note_links, and foreign keys.

**Step 2: Add route**

```go
r.Get("/graph", h.graph)
```

**Step 3: Handler returns JSON**

```json
{
  "nodes": [{"id": "note:1", "label": "Note", "type": "note"}],
  "edges": [{"source": "note:1", "target": "note:2", "type": "link"}]
}
```

**Step 4: Test endpoint**

Run: `cd backend && go test ./internal/infrastructure/http/handler -run TestNoteGraph -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/http/handler/note.go backend/internal/infrastructure/sqlite/note_repo.go backend/internal/domain/note/repository.go backend/internal/infrastructure/http/handler/note_test.go
git commit -m "feat(notes): add graph endpoint"
```

---

### Task 20: Add GraphView component and integrate into NotesPage

**Status:** ✅ Done (`501f513`) — `@xyflow/react` was already a dependency (used by the Diagrams page), so no new package install was needed. The Editor/Graph toggle is a local page-level tab kept separate from the persisted `ViewToggle`/`useViewPreference`.

**Files:**
- Create: `frontend/src/components/organisms/GraphView.tsx`
- Modify: `frontend/src/pages/Notes/index.tsx`

**Step 1: Fetch graph data**

Add `useNoteGraph` hook in `frontend/src/hooks/useNotes.ts`.

**Step 2: Render with React Flow**

Use different node colors by type.

**Step 3: Add Editor/Graph tabs to NotesPage**

**Step 4: Build**

Run: `cd frontend && npx tsc --noEmit && npx vite build`
Expected: no errors

**Step 5: Commit**

```bash
git add frontend/src/components/organisms/GraphView.tsx frontend/src/hooks/useNotes.ts frontend/src/pages/Notes/index.tsx
git commit -m "feat(notes): add Obsidian-style graph view"
```

---

## Phase 5: Network Improvements

**Status: ✅ Done** — commits `25b01a5`, `562a537`, `811d966` on `feat/hube-enhancement`. Route path `/network` and folder `pages/Network` were kept as-is (only the sidebar/page-title label changed to "Diagrams") to avoid a breaking change to existing links.

### Task 21: Rename Network to Diagrams and add templates

**Status:** ✅ Done (`25b01a5`)

**Files:**
- Modify: `frontend/src/pages/Network/index.tsx`
- Modify: `frontend/src/components/organisms/Sidebar.tsx` (label)
- Modify: `frontend/src/App.tsx` (route label if needed)

**Step 1: Update sidebar label**

From "Network" to "Diagrams".

**Step 2: Add templates**

```ts
const TEMPLATES = {
  blank: { nodes: [], edges: [] },
  flowchart: { /* starter nodes */ },
  mindmap: { /* starter nodes */ },
  architecture: { /* starter nodes */ },
  er: { /* starter nodes */ },
}
```

**Step 3: Allow template selection on create**

**Step 4: Commit**

```bash
git add frontend/src/pages/Network/index.tsx frontend/src/components/organisms/Sidebar.tsx
git commit -m "feat(diagrams): rename to diagrams and add templates"
```

---

### Task 22: Add richer node types and shapes

**Status:** ✅ Done (`562a537`) — `NODE_TYPES` converted from a capitalized string union to lowercase/snake_case, keeping the 7 existing network-device types plus 7 new general-purpose types. Shapes/colors share one `nodeStyle()`/`NODE_TYPE_STYLES` map (used by both templates and per-type rendering) instead of a custom React Flow `nodeTypes` registry, using CSS `clipPath`/`borderRadius` for distinct shapes.

**Files:**
- Modify: `frontend/src/pages/Network/index.tsx`

**Step 1: Extend NODE_TYPES**

```ts
const NODE_TYPES = [
  'process', 'decision', 'input_output', 'person', 'document', 'idea', 'image',
  'server', 'router', 'switch', 'firewall', 'client', 'database', 'cloud'
] as const
```

**Step 2: Render shapes/colors by type**

Use `node.type` to choose style.

**Step 3: Add labeled connectors**

Allow edge label input before connecting.

**Step 4: Commit**

```bash
git add frontend/src/pages/Network/index.tsx
git commit -m "feat(diagrams): add rich node types and labeled edges"
```

---

### Task 23: Add PNG/SVG export

**Status:** ✅ Done (`811d966`) — exports capture the live viewport (zoom/pan), not an auto-fit view; revisit if that becomes an issue.

**Files:**
- Modify: `frontend/src/pages/Network/index.tsx`
- Install: `cd frontend && pnpm add html-to-image`

**Step 1: Install dependency**

Run: `cd frontend && pnpm add html-to-image`

**Step 2: Implement export buttons**

Use `toPng` / `toSvg` from `html-to-image` on the ReactFlow container.

**Step 3: Build**

Run: `cd frontend && npx tsc --noEmit && npx vite build`
Expected: no errors

**Step 4: Commit**

```bash
git add frontend/src/pages/Network/index.tsx frontend/package.json frontend/pnpm-lock.yaml
git commit -m "feat(diagrams): export to png and svg"
```

---

## Phase 6: Agent-Friendly Export + CLI

**Status: ✅ Done** — commits `a32738d`, `fe17e59` on `feat/hube-enhancement`. Also fixed a real bug found along the way: the CLI's `apiURL` defaulted to port `8090`, but the backend (`cmd/api/main.go`) and Vite dev proxy both use `8080` — corrected in the Task 25 commit.

### Task 24: Add note-to-agent-markdown export

**Status:** ✅ Done (`a32738d`) — `ExportAgentMarkdown` reuses the existing `note.Content` (already plain text via `blocksToText`), no re-parsing of TipTap JSON. Route returns raw `text/markdown`, not JSON-wrapped.

**Files:**
- Modify: `backend/internal/application/note/service.go`
- Modify: `backend/internal/infrastructure/http/handler/note.go`

**Step 1: Add export method**

```go
func (s *Service) ExportAgentMarkdown(ctx context.Context, note *Note) (string, error) { /* YAML frontmatter + markdown body */ }
```

**Step 2: Add route**

```go
r.Get("/{id}/export", h.exportAgent)
```

**Step 3: Test**

Run: `cd backend && go test ./internal/application/note/...`
Expected: PASS

**Step 4: Commit**

```bash
git add backend/internal/application/note/service.go backend/internal/infrastructure/http/handler/note.go
git commit -m "feat(notes): export note as agent-friendly markdown"
```

---

### Task 25: Extend CLI with CRUD commands

**Status:** ✅ Done (`fe17e59`) — added `note create/show/update`, `task create/done`, `export notes [--out <dir>]` as new singular commands alongside the existing plural `notes`/`tasks` list commands; added `postJSON`/`putJSON`/`getRaw` helpers (none existed) and a tested `slugify` helper.

**Files:**
- Modify: `backend/cmd/cli/main.go`

**Step 1: Add note CRUD**

```go
func noteCreateCmd() *cobra.Command { /* ... */ }
func noteShowCmd() *cobra.Command { /* ... */ }
func noteUpdateCmd() *cobra.Command { /* ... */ }
```

**Step 2: Add task CRUD**

```go
func taskCreateCmd() *cobra.Command { /* ... */ }
func taskDoneCmd() *cobra.Command { /* ... */ }
```

**Step 3: Add export command**

```go
func exportNotesCmd() *cobra.Command { /* ... */ }
```

**Step 4: Build CLI**

Run: `cd backend && go build -o bin/hube ./cmd/cli`
Expected: binary built

**Step 5: Commit**

```bash
git add backend/cmd/cli/main.go
git commit -m "feat(cli): add note and task CRUD plus export"
```

---

## Final Verification

**Status: ✅ Done** — commit `227bf8b` on `feat/hube-enhancement`. All 25 feature tasks (1-25) plus this verification pass are complete on the branch.

### Task 26: Full backend test suite

**Status:** ✅ PASS — `go test ./...`, `go vet ./...` clean. `gofmt -l .` flagged a formatting issue in a new test file (`domain/note/entity_test.go`), fixed in `227bf8b`; 4 other pre-existing `gofmt` findings on `main` (unrelated files) were left untouched.

Run: `cd backend && go test ./...`
Expected: all tests pass

### Task 27: Full frontend build

**Status:** ✅ PASS — `tsc --noEmit` clean, `vite build` succeeds (routine >500kB chunk-size advisory only), `eslint .` shows only the 4 known pre-existing errors, no new ones from this branch.

Run: `cd frontend && npx tsc --noEmit && npx vite build`
Expected: no TypeScript errors, build succeeds

### Task 28: End-to-end smoke test

**Status:** ✅ PASS — all four flows confirmed against the running app (backend + Vite dev server, browser-driven).

Run: `cd backend && go run ./cmd/api &`
Run: `cd frontend && pnpm dev`
Open `http://localhost:5173` and verify:
- Tasks kanban/table toggle persists after refresh
- Projects kanban/table toggle persists after refresh
- Notes properties save, editor renders blocks, graph view shows nodes
- Diagrams templates create starter nodes, export works

#### Follow-up bugs found during smoke test — both fixed
1. ~~`frontend/src/pages/Network/index.tsx`: the `flowchart` and `mindmap` templates hardcode node type `'server'`...~~ **Fixed** (`b2da899`) — `flowchart` now uses `process`/`decision`, `mindmap` uses `idea`.
2. ~~`frontend/src/components/organisms/BlockEditor.tsx`: explicitly registers `@tiptap/extension-underline`...~~ **Fixed** (`4a4bae0`) — removed the redundant registration and the now-unused `@tiptap/extension-underline` dependency; `StarterKit` (TipTap v3) already bundles it.

### Judgment Day — pre-PR dual review

**JUDGMENT: APPROVED ✅** — 2 blind judges reviewed the full branch diff (main...feat/hube-enhancement) before opening the PR. 3 rounds; 10 issues fixed total, 0 confirmed CRITICALs or real WARNINGs remaining.

**Round 1** (against HEAD `4a4bae0`) found 2 build-breaking CRITICALs plus 6 real/suspect issues, all fixed in commits `b3b4fba`..`17c1e43`:
- `TaskModal.tsx` missing new required `Task.note_id` field → broke `tsc -b`
- `BlockEditor.tsx` `setContent(x, false)` — wrong arg shape for TipTap v3 → broke `tsc -b`
- Notes created before this change (`blocks=''`, real text only in `content`) opened as blank in the new editor — added a `content`→`blocks` fallback (`textToBlocks`)
- `hube export notes` silently overwrote files on title-slug collisions
- Unescaped tags in `ExportAgentMarkdown` allowed YAML frontmatter injection
- Malformed `blocks` JSON returned 500 instead of 400
- Free-text "Folder ID" input could violate the notes→folders FK, causing a 500
- Settings full-replace PUT could silently revert a freshly-toggled kanban/table preference

**Round 2** (against the Round 1 fixes) found the `textToBlocks` fallback rendered multi-paragraph legacy content as one run-on line (`\n` isn't a visual break in ProseMirror) — fixed in `fc83e0d` (split into per-line paragraphs).

**Round 3** (against that fix) found two edge cases in the same function — fixed directly in `6b60e1a`: CRLF not normalized (stray `\r` in line text), and a trailing `\n` producing a spurious empty paragraph.

Full findings, verdict tables, and reasoning: Engram topic `sdd/hube-enhancement/apply-progress`, project `hube`.

---

## Notes for Implementer

- Frontend has no test runner; rely on `tsc --noEmit` and `vite build` for verification.
- Backend uses standard Go testing; write/update handler tests for every new endpoint.
- Keep migrations additive and backward-compatible.
- Use existing atomic design components when possible.
- Follow existing naming conventions and theme token usage.
