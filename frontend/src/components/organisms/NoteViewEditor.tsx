import { useState, useRef, type FormEvent } from 'react'
import { Plus, X } from 'lucide-react'
import { Modal } from '../atoms/Modal'
import { Button } from '../atoms/Button'
import { CommaListInput, FIELD_CLASS } from '../molecules/CommaListInput'
import type { Folder, NoteStatus, PropertyType } from '../../types'
import { newViewId, type KnownPropertyKey, type NoteView, type NoteViewCondition, type PropertyOp } from '../../utils/noteViews'

interface NoteViewEditorProps {
  // null while creating a brand-new view.
  view: NoteView | null
  knownKeys: KnownPropertyKey[]
  folders: Folder[]
  onSave: (view: NoteView) => void
  onDelete: (id: string) => void
  onClose: () => void
}

const LABEL_CLASS = 'text-[11px] font-medium uppercase tracking-wide text-text-muted'

const STATUS_OPTIONS: { value: NoteStatus; label: string }[] = [
  { value: 'draft', label: 'Draft' },
  { value: 'in_progress', label: 'In progress' },
  { value: 'published', label: 'Published' },
]

const LAYOUT_OPTIONS: { value: NoteView['layout']; label: string }[] = [
  { value: 'table', label: 'Table' },
  { value: 'kanban', label: 'Kanban' },
]

const OP_LABEL: Record<PropertyOp, string> = {
  contains: 'contains',
  not_contains: 'does not contain',
  equals: 'is',
  eq: '=',
  neq: '≠',
  gt: '>',
  gte: '≥',
  lt: '<',
  lte: '≤',
  before: 'before',
  on: 'on',
  after: 'after',
  is_true: 'is true',
  is_false: 'is false',
  includes: 'includes',
  not_includes: 'does not include',
  is_set: 'is set',
  is_not_set: 'is not set',
}

// Mirrors the engine's OPS_BY_TYPE; is_set/is_not_set are appended for every
// type because the engine offers them universally.
const TYPE_OPS: Record<PropertyType, PropertyOp[]> = {
  text: ['contains', 'not_contains', 'equals'],
  number: ['eq', 'neq', 'gt', 'gte', 'lt', 'lte'],
  date: ['before', 'on', 'after'],
  boolean: ['is_true', 'is_false'],
  list: ['includes', 'not_includes'],
}

const SET_OPS: PropertyOp[] = ['is_set', 'is_not_set']

const NUMBER_OPS: ReadonlySet<PropertyOp> = new Set(TYPE_OPS.number)
const DATE_OPS: ReadonlySet<PropertyOp> = new Set(TYPE_OPS.date)
const VALUELESS_OPS: ReadonlySet<PropertyOp> = new Set([...SET_OPS, 'is_true', 'is_false'])

const BUILTIN_SORT_FIELDS = ['updated_at', 'created_at', 'title', 'status']

// Which value input a condition needs. This drives rendering, save coercion
// and validation, and matches how the engine interprets each op's value.
function valueKindOf(op: PropertyOp): 'text' | 'number' | 'date' | null {
  if (VALUELESS_OPS.has(op)) return null
  if (NUMBER_OPS.has(op)) return 'number'
  if (DATE_OPS.has(op)) return 'date'
  return 'text'
}

// A condition as the form edits it: a stable local row id (so React keys
// survive removals), plus the value as raw field text, coerced (and
// validated) only on save so partial input never lands in storage.
interface DraftCondition {
  id: string
  key: string
  op: PropertyOp
  value: string
}

function draftFrom(condition: NoteViewCondition): DraftCondition {
  return {
    id: newViewId(),
    key: condition.key,
    op: condition.op,
    value: condition.value === undefined ? '' : String(condition.value),
  }
}

export function NoteViewEditor({ view, knownKeys, folders, onSave, onDelete, onClose }: NoteViewEditorProps) {
  // The parent mounts this editor only while it is open, so initial state
  // doubles as the seed. null means a brand-new view with the defaults.
  const [name, setName] = useState(view?.name ?? 'New view')
  const [statuses, setStatuses] = useState<NoteStatus[]>(view?.filters.status ?? [])
  const [tags, setTags] = useState<string[]>(view?.filters.tags ?? [])
  const [folderId, setFolderId] = useState(view?.filters.folder_id ?? '')
  const [conditions, setConditions] = useState<DraftCondition[]>(
    (view?.filters.properties ?? []).map(draftFrom),
  )
  const [sortField, setSortField] = useState(view?.sort.field ?? 'updated_at')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>(view?.sort.direction ?? 'desc')
  const [layout, setLayout] = useState<NoteView['layout']>(view?.layout ?? 'table')
  const [groupBy, setGroupBy] = useState(view?.groupBy ?? '')
  const [error, setError] = useState<string | null>(null)

  const isNew = view === null

  // A fresh view opens with the default name selected, but only until the
  // name is edited once; later refocuses must place the caret instead of
  // wiping what the user typed. False from the start for existing views.
  const canSelectName = useRef(isNew)
  const typeByKey = new Map(knownKeys.map((known) => [known.key, known.type] as const))

  const updateCondition = (index: number, patch: Partial<DraftCondition>) => {
    setConditions((current) =>
      current.map((condition, i) => (i === index ? { ...condition, ...patch } : condition)),
    )
  }

  const changeKey = (index: number, key: string) => {
    const type = typeByKey.get(key)
    // Key options only contain known keys, so a missing type cannot happen;
    // guard anyway rather than seed a row with a bogus default op.
    if (!type) return
    // The op list is type-driven: a new key starts from the type's default
    // op and an empty value buffer.
    updateCondition(index, { key, op: TYPE_OPS[type][0], value: '' })
  }

  const changeOp = (index: number, op: PropertyOp) => {
    setConditions((current) =>
      current.map((condition, i) => {
        if (i !== index) return condition
        // Keep the typed value when the input kind is unchanged (contains ->
        // equals); otherwise start over with an empty value.
        return valueKindOf(condition.op) === valueKindOf(op)
          ? { ...condition, op }
          : { ...condition, op, value: '' }
      }),
    )
  }

  const addCondition = () => {
    const first = knownKeys[0]
    if (!first) return
    setConditions((current) => [
      ...current,
      { id: newViewId(), key: first.key, op: TYPE_OPS[first.type][0], value: '' },
    ])
  }

  // Ops offered for a key. A key no loaded note has anymore can only ever
  // match the set ops in the engine, so those are the only new choices; the
  // stored op stays listed either way so saving does not silently rewrite it.
  const opsFor = (condition: DraftCondition): PropertyOp[] => {
    const type = typeByKey.get(condition.key)
    const base: PropertyOp[] = type ? [...TYPE_OPS[type], ...SET_OPS] : [...SET_OPS]
    return base.includes(condition.op) ? base : [condition.op, ...base]
  }

  const handleDelete = () => {
    if (!view) return
    onDelete(view.id)
    onClose()
  }

  const handleSave = (e: FormEvent) => {
    e.preventDefault()
    const trimmedName = name.trim()
    if (trimmedName === '') {
      setError('Enter a name for the view.')
      return
    }
    const properties: NoteViewCondition[] = []
    for (let index = 0; index < conditions.length; index += 1) {
      const condition = conditions[index]
      const kind = valueKindOf(condition.op)
      if (kind === null) {
        properties.push({ key: condition.key, op: condition.op })
        continue
      }
      if (condition.value.trim() === '') {
        setError(`Enter a value for "${condition.key}".`)
        return
      }
      if (kind === 'number' && !Number.isFinite(Number(condition.value))) {
        setError(`The "${condition.key}" condition needs a number value.`)
        return
      }
      properties.push({
        key: condition.key,
        op: condition.op,
        value: kind === 'number' ? Number(condition.value) : condition.value.trim(),
      })
    }
    setError(null)
    onSave({
      id: view?.id ?? newViewId(),
      name: trimmedName,
      filters: {
        properties,
        ...(statuses.length > 0 ? { status: statuses } : {}),
        ...(tags.length > 0 ? { tags } : {}),
        ...(folderId !== '' ? { folder_id: folderId } : {}),
      },
      sort: { field: sortField, direction: sortDirection },
      layout,
      // Kept even on the kanban layout (which groups by status): the value
      // stays in the view for the table grouping renderer to consume.
      ...(groupBy !== '' ? { groupBy } : {}),
    })
    onClose()
  }

  const renderConditionValue = (condition: DraftCondition, index: number) => {
    const kind = valueKindOf(condition.op)
    if (kind === null) return null
    const ariaLabel = `Condition ${index + 1} value`
    if (kind === 'date') {
      return (
        <input
          type="date"
          value={condition.value}
          onChange={(e) => updateCondition(index, { value: e.target.value })}
          aria-label={ariaLabel}
          className={FIELD_CLASS}
        />
      )
    }
    return (
      <input
        type="text"
        inputMode={kind === 'number' ? 'decimal' : undefined}
        autoComplete="off"
        value={condition.value}
        onChange={(e) => updateCondition(index, { value: e.target.value })}
        placeholder={kind === 'number' ? 'e.g. 5' : 'Value'}
        aria-label={ariaLabel}
        className={FIELD_CLASS}
      />
    )
  }

  return (
    <Modal open onClose={onClose} title={isNew ? 'New view' : 'Edit view'}>
      <form onSubmit={handleSave} className="flex flex-col gap-4">
        <div className="flex flex-col gap-1">
          <label htmlFor="note-view-name" className={LABEL_CLASS}>
            Name
          </label>
          <input
            id="note-view-name"
            type="text"
            value={name}
            onChange={(e) => {
              canSelectName.current = false
              setName(e.target.value)
            }}
            autoFocus={isNew}
            onFocus={(e) => {
              if (canSelectName.current) e.currentTarget.select()
            }}
            className={FIELD_CLASS}
          />
        </div>

        <div role="group" aria-labelledby="note-view-status-label" className="flex flex-col gap-1">
          <span id="note-view-status-label" className={LABEL_CLASS}>
            Status
          </span>
          <div className="flex flex-wrap gap-3">
            {STATUS_OPTIONS.map((option) => (
              <label key={option.value} className="flex items-center gap-1.5 text-sm text-text-secondary">
                <input
                  type="checkbox"
                  checked={statuses.includes(option.value)}
                  onChange={(e) =>
                    setStatuses((current) =>
                      e.target.checked
                        ? [...current, option.value]
                        : current.filter((s) => s !== option.value),
                    )
                  }
                  className="h-4 w-4 rounded border-border bg-surface-elevated text-(--color-accent) accent-(--color-accent)"
                />
                {option.label}
              </label>
            ))}
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div className="flex flex-col gap-1">
            <label htmlFor="note-view-folder" className={LABEL_CLASS}>
              Folder
            </label>
            <select
              id="note-view-folder"
              value={folderId}
              onChange={(e) => setFolderId(e.target.value)}
              className={FIELD_CLASS}
            >
              <option value="">All folders</option>
              {folders.map((folder) => (
                <option key={folder.id} value={folder.id}>
                  {folder.name}
                </option>
              ))}
              {/* A folder deleted after the view was saved stays listed so
                  the value is visible until the user changes it. */}
              {folderId !== '' && !folders.some((f) => f.id === folderId) && (
                <option value={folderId}>Unknown folder</option>
              )}
            </select>
          </div>

          <div className="flex flex-col gap-1">
            <label htmlFor="note-view-tags" className={LABEL_CLASS}>
              Tags
            </label>
            <CommaListInput
              id="note-view-tags"
              value={tags}
              onChange={setTags}
              label="View tags"
              placeholder="Add tags…"
            />
          </div>
        </div>

        <div role="group" aria-labelledby="note-view-conditions-label" className="flex flex-col gap-2">
          <span id="note-view-conditions-label" className={LABEL_CLASS}>
            Conditions
          </span>

          {conditions.length === 0 && (
            <p className="text-xs text-text-muted">Every note passes until a condition is added.</p>
          )}

          {conditions.map((condition, index) => (
            <div key={condition.id} className="flex flex-col gap-1.5 rounded-lg border border-border p-2">
              <div className="flex items-center gap-1.5">
                <select
                  aria-label={`Condition ${index + 1} property`}
                  value={condition.key}
                  onChange={(e) => changeKey(index, e.target.value)}
                  className={`${FIELD_CLASS} min-w-0 flex-1`}
                >
                  {/* Keys already referenced stay selectable even if no
                      loaded note has them anymore, so the row can be
                      inspected and removed. */}
                  {condition.key !== '' && !typeByKey.has(condition.key) && (
                    <option value={condition.key}>{condition.key}</option>
                  )}
                  {knownKeys.map((known) => (
                    <option key={known.key} value={known.key}>
                      {known.key}
                    </option>
                  ))}
                </select>
                <div className="w-36 shrink-0">
                  <select
                    aria-label={`Condition ${index + 1} operator`}
                    value={condition.op}
                    onChange={(e) => changeOp(index, e.target.value as PropertyOp)}
                    className={FIELD_CLASS}
                  >
                    {opsFor(condition).map((op) => (
                      <option key={op} value={op}>
                        {OP_LABEL[op]}
                      </option>
                    ))}
                  </select>
                </div>
                <button
                  type="button"
                  onClick={() => setConditions((current) => current.filter((_, i) => i !== index))}
                  aria-label={`Remove condition ${index + 1}`}
                  title="Remove condition"
                  className="shrink-0 rounded-lg p-1.5 text-text-muted transition-colors hover:bg-surface-card hover:text-red-400"
                >
                  <X size={14} />
                </button>
              </div>
              {renderConditionValue(condition, index)}
            </div>
          ))}

          <Button
            type="button"
            variant="ghost"
            size="sm"
            icon={<Plus size={14} />}
            disabled={knownKeys.length === 0}
            onClick={addCondition}
            title={knownKeys.length === 0 ? 'No note properties exist yet' : undefined}
          >
            Add condition
          </Button>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="flex flex-col gap-1">
            <label htmlFor="note-view-sort-field" className={LABEL_CLASS}>
              Sort by
            </label>
            <select
              id="note-view-sort-field"
              value={sortField}
              onChange={(e) => setSortField(e.target.value)}
              className={FIELD_CLASS}
            >
              <option value="updated_at">Updated</option>
              <option value="created_at">Created</option>
              <option value="title">Title</option>
              <option value="status">Status</option>
              {knownKeys.map((known) => (
                <option key={known.key} value={known.key}>
                  {known.key}
                </option>
              ))}
              {/* A stale sort key stays listed so saving does not rewrite it. */}
              {!BUILTIN_SORT_FIELDS.includes(sortField) && !typeByKey.has(sortField) && (
                <option value={sortField}>{sortField}</option>
              )}
            </select>
          </div>

          <div className="flex flex-col gap-1">
            <label htmlFor="note-view-sort-direction" className={LABEL_CLASS}>
              Direction
            </label>
            <select
              id="note-view-sort-direction"
              value={sortDirection}
              onChange={(e) => setSortDirection(e.target.value as 'asc' | 'desc')}
              className={FIELD_CLASS}
            >
              <option value="asc">Ascending</option>
              <option value="desc">Descending</option>
            </select>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div role="group" aria-labelledby="note-view-layout-label" className="flex flex-col gap-1">
            <span id="note-view-layout-label" className={LABEL_CLASS}>
              Layout
            </span>
            <div className="flex gap-1.5">
              {LAYOUT_OPTIONS.map((option) => (
                <button
                  key={option.value}
                  type="button"
                  aria-pressed={layout === option.value}
                  onClick={() => setLayout(option.value)}
                  className={`flex-1 rounded-md border px-2 py-1.5 text-xs transition-colors ${
                    layout === option.value
                      ? 'border-accent bg-accent/15 text-accent'
                      : 'border-border bg-surface-card text-text-secondary'
                  }`}
                >
                  {option.label}
                </button>
              ))}
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <label htmlFor="note-view-group-by" className={LABEL_CLASS}>
              Group by
            </label>
            <select
              id="note-view-group-by"
              value={groupBy}
              disabled={layout !== 'table'}
              onChange={(e) => setGroupBy(e.target.value)}
              className={FIELD_CLASS}
            >
              <option value="">None</option>
              {knownKeys.map((known) => (
                <option key={known.key} value={known.key}>
                  {known.key}
                </option>
              ))}
              {/* A stale group key stays listed so saving does not rewrite it. */}
              {groupBy !== '' && !typeByKey.has(groupBy) && <option value={groupBy}>{groupBy}</option>}
            </select>
          </div>
        </div>

        {error && (
          <p role="alert" className="text-xs text-red-400">
            {error}
          </p>
        )}

        <div className="flex items-center justify-between gap-2">
          {isNew ? <span aria-hidden="true" /> : (
            <Button type="button" variant="danger" size="sm" onClick={handleDelete}>
              Delete
            </Button>
          )}
          <div className="flex gap-2">
            <Button type="button" variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit">{isNew ? 'Create view' : 'Save changes'}</Button>
          </div>
        </div>
      </form>
    </Modal>
  )
}
