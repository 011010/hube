# Note Views (Bases-lite) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Named, persisted views over notes — property filters, sort, layout, and read-only property grouping — evaluated in memory.

**Architecture:** Views are a JSON array stored under the `general.notes_views` settings key (full-replacement PUT contract). A pure frontend engine (`noteViews.ts`) filters, sorts, and groups the already-loaded notes. The Notes page gains a view selector and a modal editor. No new tables, no new endpoints.

**Tech Stack:** Go/Chi/SQLite (settings field only), React 19, TypeScript, TanStack Query, vitest for the pure engine.

**Design doc:** `docs/plans/2026-11-27-note-views-design.md`

---

### Task 1: Backend settings field

**Files:**
- Modify: `backend/internal/infrastructure/http/handler/setting.go`
- Test: `backend/internal/infrastructure/http/handler/setting_test.go`

**Step 1: Write the failing test**

```go
func TestSettings_NoteViewsRoundTrip(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/settings"

	views := `[{"id":"v1","name":"Rating alto","filters":{"properties":[{"key":"rating","op":"gte","value":7}]},"sort":{"field":"rating","direction":"desc"},"layout":"table"}]`
	body := `{"general":{"notes_views":` + strconv.Quote(views) + `}}`

	put := mustPut(t, base, body)
	assert.Equal(t, http.StatusOK, put.StatusCode)
	var got map[string]any
	mustDecode(t, put, &got)
	general := got["general"].(map[string]any)
	assert.Equal(t, views, general["notes_views"])

	get := mustGet(t, base)
	var current map[string]any
	mustDecode(t, get, &current)
	g := current["general"].(map[string]any)
	assert.Equal(t, views, g["notes_views"])

	// Full replacement: PUT without the field clears it.
	put2 := mustPut(t, base, `{"general":{"display_name":"Hube"}}`)
	assert.Equal(t, http.StatusOK, put2.StatusCode)
	get2 := mustGet(t, base)
	var cleared map[string]any
	mustDecode(t, get2, &cleared)
	g2 := cleared["general"].(map[string]any)
	assert.Nil(t, g2["notes_views"])
}
```

Add `strconv` to imports if missing.

**Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/infrastructure/http/handler/ -run TestSettings_NoteViews -v
```

Expected: FAIL — `notes_views` absent from the response.

**Step 3: Write minimal implementation**

In `setting.go`, extend `generalSettings` (it serves both request and response):

```go
type generalSettings struct {
	DisplayName     string `json:"display_name"`
	ViewPreferences string `json:"view_preferences,omitempty"`
	NotesViews      string `json:"notes_views,omitempty"`
}
```

In `Get`, add to the response literal: `NotesViews: all["general.notes_views"]`.
In `Put`, add to `pairs`: `"general.notes_views": payload.General.NotesViews`.

**Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/infrastructure/http/handler/ -v && go vet ./... && gofmt -l ./internal
```

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/http/handler/setting.go backend/internal/infrastructure/http/handler/setting_test.go
git commit -m "feat(settings): store note views"
```

---

### Task 2: View model and engine (TDD)

**Files:**
- Create: `frontend/src/utils/noteViews.ts`
- Test: `frontend/src/utils/noteViews.test.ts`
- Modify: `frontend/src/hooks/useSettings.ts` (add `notes_views?: string` to the settings `general` type)

**Step 1: Write the failing tests**

Cover, at minimum:

- `parseNoteViews`: valid array parses; malformed JSON → `[]`; non-array → `[]`; entries missing a name are dropped; missing ids are generated; more than 20 views are truncated; names are trimmed and capped at 100 characters; invalid conditions are dropped from `filters.properties`; `filters.properties` defaults to `[]`.
- `applyNoteView` with `filters.status`: only listed statuses match.
- `filters.tags`: all tags must match, case-insensitively; a note missing one tag is excluded.
- `filters.folder_id`: exact match.
- Text ops: `contains`, `not_contains`, `equals` (case-insensitive).
- Number ops: `eq`, `neq`, `gt`, `gte`, `lt`, `lte` — including a condition value that is not a number (no match).
- Date ops: `before`, `on`, `after` with `YYYY-MM-DD` string comparison; non-string condition value does not match.
- Boolean ops: `is_true`, `is_false`.
- List ops: `includes`, `not_includes` (case-insensitive).
- `is_set`, `is_not_set` for every type.
- Type mismatch: `gt` against a text property does not match; `contains` against a number does not match.
- Multiple conditions are ANDed.
- Sort: `updated_at`, `title` (locale), `status` (draft < in_progress < published), and a property key — numbers numerically, dates lexicographically, text by locale, list by first item, boolean false < true. Missing values always sort last in both directions. Stable for equal keys.
- `groupNotes`: by `status` in fixed order; by a text property with distinct sections sorted by locale; by a list property a note appears in every matching section; by a number property sections sort numerically; notes without the property land in the last `No value` section; no field → a single group.

**Step 2: Run tests to verify they fail**

```bash
cd frontend && pnpm vitest run src/utils/noteViews.test.ts
```

Expected: FAIL — module does not exist.

**Step 3: Write the implementation**

Export from `frontend/src/utils/noteViews.ts`:

```ts
export type PropertyOp =
  | 'contains' | 'not_contains' | 'equals'
  | 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte'
  | 'before' | 'on' | 'after'
  | 'is_true' | 'is_false'
  | 'includes' | 'not_includes'
  | 'is_set' | 'is_not_set'

export interface NoteViewCondition { key: string; op: PropertyOp; value?: string | number | boolean }
export interface NoteViewFilters {
  status?: NoteStatus[]
  tags?: string[]
  folder_id?: string
  properties: NoteViewCondition[]
}
export interface NoteView {
  id: string
  name: string
  filters: NoteViewFilters
  sort: { field: string; direction: 'asc' | 'desc' }
  layout: 'table' | 'kanban'
  groupBy?: string
}

export const MAX_NOTE_VIEWS = 20

export function parseNoteViews(raw: string | undefined): NoteView[]
export function applyNoteView(notes: Note[], view: NoteView | null): Note[]
export function groupNotes(notes: Note[], field?: string): { key: string; notes: Note[] }[]
```

Rules:

- The property type comes from each note's stored value, never from the view.
- An op that does not apply to the note's actual type does not match that note.
- `crypto.randomUUID()` for generated ids (guard for non-secure contexts with a
  fallback counter).
- Group keys are display strings; `groupNotes` returns groups in display order
  with `No value` last.

**Step 4: Run tests to verify they pass**

```bash
cd frontend && pnpm vitest run src/utils/noteViews.test.ts && pnpm run build && pnpm run lint
```

**Step 5: Commit**

```bash
git add frontend/src/utils/noteViews.ts frontend/src/utils/noteViews.test.ts frontend/src/hooks/useSettings.ts
git commit -m "feat(notes): add saved view engine"
```

---

### Task 3: Persistence hook

**Files:**
- Create: `frontend/src/hooks/useNoteViews.ts`

**Step 1: Implement**

```ts
export function useNoteViews() {
  const { data } = useSettings()
  const update = useUpdateSettings()
  const queryClient = useQueryClient()

  const views = useMemo(() => parseNoteViews(data?.general.notes_views), [data?.general.notes_views])

  const saveViews = useCallback((next: NoteView[]) => {
    const latest = queryClient.getQueryData<Settings>(['settings'])
    if (!latest) return
    update.mutate({
      general: {
        ...latest.general,
        notes_views: JSON.stringify(next.slice(0, MAX_NOTE_VIEWS)),
      },
    })
  }, [queryClient, update])

  return {
    views,
    saveView: (view: NoteView) => saveViews(upsert(views, view)),
    deleteView: (id: string) => saveViews(views.filter((v) => v.id !== id)),
  }
}
```

Include a small local `upsert` (replace by id, else append). Merge with the
latest cached settings before writing, exactly like `useViewPreference` does,
so a view save never erases `view_preferences`.

**Step 2: Verify**

```bash
cd frontend && pnpm run build && pnpm run lint
```

**Step 3: Commit**

```bash
git add frontend/src/hooks/useNoteViews.ts
git commit -m "feat(notes): persist saved views"
```

---

### Task 4: View selector and editor

**Files:**
- Create: `frontend/src/components/organisms/NoteViewEditor.tsx`
- Modify: `frontend/src/pages/Notes/index.tsx`

**Step 1: Editor component**

`NoteViewEditor` receives `view: NoteView | null`, `knownKeys: {key, type}[]`,
`folders`, `onSave(view)`, `onDelete(id)`, `onClose()`. It renders inside the
existing `atoms/Modal`:

- Name input.
- Status checkboxes (draft/in_progress/published, empty = no filter).
- Tags via the `CommaListInput` pattern from `PropertiesPanel` (extract that
  component to `frontend/src/components/molecules/CommaListInput.tsx` and
  import it from both places — do not duplicate it).
- Folder select.
- Condition rows: property key select (from `knownKeys`), op select filtered by
  the key's dominant type, value input typed by op (reuse the input patterns
  from `PropertiesPanel`), remove button, add button.
- Sort: field select (updated_at/created_at/title/status + known keys),
  direction toggle.
- Layout: table/kanban.
- Group-by select (None + known keys), only enabled for table layout.
- Save and Delete buttons.

**Step 2: Page wiring**

In `Notes/index.tsx`:

- `const { views, saveView, deleteView } = useNoteViews()`
- `const [activeViewId, setActiveViewId] = useState<string | null>(null)`
- `const activeView = views.find((v) => v.id === activeViewId) ?? null`
- Header selector: `<select>` with "All notes" plus each view, plus an "Edit
  view…" entry that opens the modal for the active view, and a "New view…"
  action (button next to the selector).
- `displayNotes` becomes: search/folder/semantic state as today, then
  `applyNoteView(notes, activeView)`.
- `knownKeys` derived from all loaded notes: union of property keys with their
  most common type.

**Step 3: Verify**

```bash
cd frontend && pnpm run build && pnpm run lint && pnpm run test
```

**Step 4: Commit**

```bash
git add frontend/src/components/molecules/CommaListInput.tsx frontend/src/components/organisms/NoteViewEditor.tsx frontend/src/components/molecules/PropertiesPanel.tsx frontend/src/pages/Notes/index.tsx
git commit -m "feat(notes): switch and edit saved views"
```

---

### Task 5: Grouped table rendering

**Files:**
- Modify: `frontend/src/pages/Notes/index.tsx`

**Step 1: Implement**

When `activeView` is set, `layout` is `table`, and `groupBy` is set:

- `const groups = groupNotes(filtered, activeView.groupBy)`
- Render, per group: a header row (`text-[11px] font-medium uppercase tracking-wide text-text-muted` with the group key and count) followed by the existing `DataTable` for that group's notes.
- Do not change the kanban path, the search behavior, or the editor.

If `groupBy` is absent, keep the current single-table rendering.

**Step 2: Verify**

```bash
cd frontend && pnpm run build && pnpm run lint && pnpm run test
```

**Step 3: Commit**

```bash
git add frontend/src/pages/Notes/index.tsx
git commit -m "feat(notes): group notes by property"
```

---

### Task 6: Verification

**Step 1: Full suites**

```bash
cd backend && go test ./... -count=1 && go vet ./... && gofmt -l ./internal
cd ../frontend && pnpm run build && pnpm run lint && pnpm run test
```

**Step 2: Settings round-trip on a database copy**

```bash
sqlite3 <real hube.db> ".backup /tmp/hube-views-check.db"
DB_PATH=/tmp/hube-views-check.db PORT=8099 HUBE_API_TOKEN=test-token go run ./cmd/api
curl -X PUT localhost:8099/api/v1/settings -H "Authorization: Bearer test-token" \
  -d '{"general":{"notes_views":"[{\"id\":\"v1\",\"name\":\"Rating alto\",\"filters\":{\"properties\":[{\"key\":\"rating\",\"op\":\"gte\",\"value\":7}]},\"sort\":{\"field\":\"rating\",\"direction\":\"desc\"},\"layout\":\"table\"}]"}}'
curl localhost:8099/api/v1/settings -H "Authorization: Bearer test-token"
```

Expected: the view JSON round-trips byte-identically.

**Step 3: Manual acceptance**

Walk the four acceptance criteria from the design doc against the running app.

---

## Out of Scope

- Kanban grouped by property and drag-to-update.
- Server-side filtering.
- Formulas, summaries, and computed columns.
- Cross-user sharing.
