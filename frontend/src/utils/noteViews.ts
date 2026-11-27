import type { Note, NoteProperty, NoteStatus, PropertyType } from '../types'

export type PropertyOp =
  | 'contains' | 'not_contains' | 'equals'
  | 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte'
  | 'before' | 'on' | 'after'
  | 'is_true' | 'is_false'
  | 'includes' | 'not_includes'
  | 'is_set' | 'is_not_set'

export interface NoteViewCondition {
  key: string
  op: PropertyOp
  value?: string | number | boolean
}

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

export interface NoteGroup {
  key: string
  notes: Note[]
}

export const MAX_NOTE_VIEWS = 20

const MAX_VIEW_NAME_LENGTH = 100

export const NO_VALUE_KEY = 'No value'

// Shared collator: numeric: true gives natural ordering ("Note 2" < "Note 10").
const collator = new Intl.Collator(undefined, { numeric: true })

const NOTE_STATUSES: NoteStatus[] = ['draft', 'in_progress', 'published']

const STATUS_RANK: Record<NoteStatus, number> = {
  draft: 0,
  in_progress: 1,
  published: 2,
}

const PROPERTY_OPS: ReadonlySet<PropertyOp> = new Set<PropertyOp>([
  'contains', 'not_contains', 'equals',
  'eq', 'neq', 'gt', 'gte', 'lt', 'lte',
  'before', 'on', 'after',
  'is_true', 'is_false',
  'includes', 'not_includes',
  'is_set', 'is_not_set',
])

// The type of a property is whatever the note stores for that key, so the
// same view key can be a number on one note and text on another.
const OPS_BY_TYPE: Record<NoteProperty['type'], ReadonlySet<PropertyOp>> = {
  text: new Set<PropertyOp>(['contains', 'not_contains', 'equals']),
  number: new Set<PropertyOp>(['eq', 'neq', 'gt', 'gte', 'lt', 'lte']),
  date: new Set<PropertyOp>(['before', 'on', 'after']),
  boolean: new Set<PropertyOp>(['is_true', 'is_false']),
  list: new Set<PropertyOp>(['includes', 'not_includes']),
}

const VALUELESS_OPS: ReadonlySet<PropertyOp> = new Set<PropertyOp>([
  'is_set', 'is_not_set', 'is_true', 'is_false',
])

let fallbackIdCounter = 0

function generateId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  fallbackIdCounter += 1
  return `view-${Date.now().toString(36)}-${fallbackIdCounter}`
}

// crypto.randomUUID throws in non-secure contexts (plain-HTTP LAN hosting),
// so callers that need a fresh view id must use this guarded generator.
export function newViewId(): string {
  return generateId()
}

function isNoteStatus(value: unknown): value is NoteStatus {
  return value === 'draft' || value === 'in_progress' || value === 'published'
}

function normalizeCondition(raw: unknown): NoteViewCondition | null {
  if (typeof raw !== 'object' || raw === null) return null
  const { key, op, value } = raw as Record<string, unknown>
  if (typeof key !== 'string' || key.trim() === '') return null
  if (typeof op !== 'string' || !PROPERTY_OPS.has(op as PropertyOp)) return null
  const typedOp = op as PropertyOp
  if (VALUELESS_OPS.has(typedOp)) return { key, op: typedOp }
  if (
    typeof value === 'string' ||
    typeof value === 'number' ||
    typeof value === 'boolean'
  ) {
    return { key, op: typedOp, value }
  }
  return null
}

function normalizeFilters(raw: unknown): NoteViewFilters {
  const source = typeof raw === 'object' && raw !== null ? (raw as Record<string, unknown>) : {}
  const rawProperties = Array.isArray(source.properties) ? source.properties : []
  const filters: NoteViewFilters = {
    properties: rawProperties
      .map(normalizeCondition)
      .filter((c): c is NoteViewCondition => c !== null),
  }
  const status = Array.isArray(source.status) ? source.status.filter(isNoteStatus) : []
  if (status.length > 0) filters.status = status
  const tags = Array.isArray(source.tags)
    ? source.tags.filter((t): t is string => typeof t === 'string')
    : []
  if (tags.length > 0) filters.tags = tags
  if (typeof source.folder_id === 'string') filters.folder_id = source.folder_id
  return filters
}

function normalizeSort(raw: unknown): NoteView['sort'] {
  const source = typeof raw === 'object' && raw !== null ? (raw as Record<string, unknown>) : {}
  const field = typeof source.field === 'string' && source.field.trim() !== ''
    ? source.field
    : 'updated_at'
  const direction = source.direction === 'asc' ? 'asc' : 'desc'
  return { field, direction }
}

function normalizeView(entry: unknown): NoteView | null {
  if (typeof entry !== 'object' || entry === null) return null
  const source = entry as Record<string, unknown>
  const name = typeof source.name === 'string' ? source.name.trim() : ''
  if (name === '') return null
  const view: NoteView = {
    id: typeof source.id === 'string' && source.id !== '' ? source.id : generateId(),
    name: name.slice(0, MAX_VIEW_NAME_LENGTH),
    filters: normalizeFilters(source.filters),
    sort: normalizeSort(source.sort),
    layout: source.layout === 'kanban' ? 'kanban' : 'table',
  }
  if (typeof source.groupBy === 'string' && source.groupBy.trim() !== '') {
    view.groupBy = source.groupBy
  }
  return view
}

export function parseNoteViews(raw: string | undefined): NoteView[] {
  if (typeof raw !== 'string' || raw === '') return []
  let parsed: unknown
  try {
    parsed = JSON.parse(raw)
  } catch {
    return []
  }
  if (!Array.isArray(parsed)) return []
  const views: NoteView[] = []
  const seenIds = new Set<string>()
  for (const entry of parsed) {
    if (views.length >= MAX_NOTE_VIEWS) break
    const view = normalizeView(entry)
    if (!view) continue
    // Duplicate stored ids would collide as React keys; keep the first.
    if (seenIds.has(view.id)) continue
    seenIds.add(view.id)
    views.push(view)
  }
  return views
}

function findProperty(note: Note, key: string): NoteProperty | undefined {
  return note.properties.find(p => p.key === key)
}

function matchCondition(note: Note, condition: NoteViewCondition): boolean {
  const property = findProperty(note, condition.key)
  // A condition on a key the note does not have matches only is_not_set.
  if (!property) return condition.op === 'is_not_set'
  if (condition.op === 'is_set') return true
  if (condition.op === 'is_not_set') return false
  // An op that does not apply to the note's actual type does not match.
  if (!OPS_BY_TYPE[property.type].has(condition.op)) return false
  switch (property.type) {
    case 'text': {
      const value = property.value
      const expected = condition.value
      if (typeof value !== 'string' || typeof expected !== 'string') return false
      const a = value.toLowerCase()
      const b = expected.toLowerCase()
      if (condition.op === 'contains') return a.includes(b)
      if (condition.op === 'not_contains') return !a.includes(b)
      return a === b
    }
    case 'number': {
      const value = property.value
      const expected = condition.value
      if (typeof value !== 'number' || typeof expected !== 'number') return false
      switch (condition.op) {
        case 'eq': return value === expected
        case 'neq': return value !== expected
        case 'gt': return value > expected
        case 'gte': return value >= expected
        case 'lt': return value < expected
        case 'lte': return value <= expected
      }
      return false
    }
    case 'date': {
      const value = property.value
      const expected = condition.value
      if (typeof value !== 'string' || typeof expected !== 'string') return false
      if (condition.op === 'before') return value < expected
      if (condition.op === 'on') return value === expected
      return value > expected
    }
    case 'boolean': {
      const value = property.value
      if (typeof value !== 'boolean') return false
      return condition.op === 'is_true' ? value : !value
    }
    case 'list': {
      const value = property.value
      const expected = condition.value
      if (!Array.isArray(value) || typeof expected !== 'string') return false
      const needle = expected.toLowerCase()
      const hit = value.some(item => typeof item === 'string' && item.toLowerCase() === needle)
      return condition.op === 'includes' ? hit : !hit
    }
  }
}

function matchesFilters(note: Note, filters: NoteViewFilters): boolean {
  const status = filters.status
  if (status && status.length > 0 && !status.includes(note.status)) return false
  if (filters.tags) {
    const present = new Set(note.tags.map(t => t.toLowerCase()))
    if (!filters.tags.every(t => present.has(t.toLowerCase()))) return false
  }
  if (filters.folder_id !== undefined && note.folder_id !== filters.folder_id) return false
  return filters.properties.every(c => matchCondition(note, c))
}

type SortComparable =
  | { kind: 'number'; value: number }
  | { kind: 'date'; value: string }
  | { kind: 'text'; value: string }

function propertyComparable(property: NoteProperty | undefined): SortComparable | undefined {
  if (!property) return undefined
  switch (property.type) {
    case 'number': {
      const value = property.value
      return typeof value === 'number' ? { kind: 'number', value } : undefined
    }
    case 'boolean': {
      const value = property.value
      return typeof value === 'boolean' ? { kind: 'number', value: value ? 1 : 0 } : undefined
    }
    case 'date': {
      const value = property.value
      return typeof value === 'string' ? { kind: 'date', value } : undefined
    }
    case 'text': {
      const value = property.value
      return typeof value === 'string' ? { kind: 'text', value } : undefined
    }
    case 'list': {
      const value = property.value
      if (!Array.isArray(value) || value.length === 0) return undefined
      return { kind: 'text', value: value[0] }
    }
  }
}

function fieldComparable(note: Note, field: string): SortComparable | undefined {
  if (field === 'status') return { kind: 'number', value: STATUS_RANK[note.status] }
  if (field === 'title') return { kind: 'text', value: note.title }
  if (field === 'updated_at' || field === 'created_at') {
    return { kind: 'date', value: note[field] }
  }
  return propertyComparable(findProperty(note, field))
}

function compareComparable(a: SortComparable, b: SortComparable): number {
  if (a.kind === 'number' && b.kind === 'number') return a.value - b.value
  if (a.kind === 'text' && b.kind === 'text') return collator.compare(a.value, b.value)
  if (a.kind === 'date' && b.kind === 'date') {
    // created_at/updated_at can mix "...T00:00:00Z" with "...T00:00:00.5Z",
    // which plain string comparison misorders ("." < "Z"). Compare as epochs
    // when both sides parse; property dates ("YYYY-MM-DD") parse too.
    const ta = Date.parse(a.value)
    const tb = Date.parse(b.value)
    if (Number.isFinite(ta) && Number.isFinite(tb)) return ta - tb
    return a.value < b.value ? -1 : a.value > b.value ? 1 : 0
  }
  // Mixed types on the same key: keep the order total and deterministic.
  return collator.compare(a.kind, b.kind) || collator.compare(String(a.value), String(b.value))
}

function sortNotes(notes: Note[], sort: NoteView['sort']): Note[] {
  const copy = [...notes]
  copy.sort((a, b) => {
    const ca = fieldComparable(a, sort.field)
    const cb = fieldComparable(b, sort.field)
    // Missing values always sort last, regardless of direction.
    if (ca === undefined && cb === undefined) return 0
    if (ca === undefined) return 1
    if (cb === undefined) return -1
    const base = compareComparable(ca, cb)
    return sort.direction === 'asc' ? base : base === 0 ? 0 : -base
  })
  return copy
}

export function applyNoteView(notes: Note[], view: NoteView | null): Note[] {
  if (!view) return notes
  const filtered = notes.filter(n => matchesFilters(n, view.filters))
  return sortNotes(filtered, view.sort)
}

// Display keys the note's property maps to, or null when it has no value.
function groupValues(property: NoteProperty): string[] | null {
  switch (property.type) {
    case 'list': {
      const value = property.value
      if (!Array.isArray(value) || value.length === 0) return null
      return [...new Set(value)]
    }
    case 'number': {
      const value = property.value
      return typeof value === 'number' ? [String(value)] : null
    }
    case 'boolean': {
      const value = property.value
      return typeof value === 'boolean' ? [String(value)] : null
    }
    case 'text':
    case 'date': {
      const value = property.value
      return typeof value === 'string' ? [value] : null
    }
  }
}

interface GroupBucket {
  key: string
  type: NoteProperty['type']
  notes: Note[]
}

function compareBuckets(a: GroupBucket, b: GroupBucket): number {
  if (a.type === 'number' && b.type === 'number') {
    return Number(a.key) - Number(b.key)
  }
  if (a.type === 'boolean' && b.type === 'boolean') {
    return (a.key === 'true' ? 1 : 0) - (b.key === 'true' ? 1 : 0)
  }
  if (a.type === 'date' && b.type === 'date') {
    return a.key < b.key ? -1 : a.key > b.key ? 1 : 0
  }
  return collator.compare(a.key, b.key)
}

export function groupNotes(notes: Note[], field?: string): NoteGroup[] {
  if (notes.length === 0) return []
  if (!field || field === '') return [{ key: '', notes }]
  if (field === 'status') {
    return NOTE_STATUSES
      .map(status => ({
        key: status,
        notes: notes.filter(n => n.status === status),
      }))
      .filter(g => g.notes.length > 0)
  }
  const buckets = new Map<string, GroupBucket>()
  const noValue: Note[] = []
  for (const note of notes) {
    const property = findProperty(note, field)
    if (!property) {
      noValue.push(note)
      continue
    }
    const keys = groupValues(property)
    if (!keys) {
      noValue.push(note)
      continue
    }
    for (const key of keys) {
      const bucket = buckets.get(key)
      if (bucket) bucket.notes.push(note)
      else buckets.set(key, { key, type: property.type, notes: [note] })
    }
  }
  const groups: NoteGroup[] = [...buckets.values()]
    .sort(compareBuckets)
    .map(b => ({ key: b.key, notes: b.notes }))
  if (noValue.length > 0) groups.push({ key: NO_VALUE_KEY, notes: noValue })
  return groups
}

export interface KnownPropertyKey {
  key: string
  type: PropertyType
}

// Union of property keys across notes, each typed by its dominant type (the
// most common type across notes). Count ties break alphabetically by type
// name, so the result never depends on note order. Keys sort with the module
// collator (natural numeric order).
export function knownPropertyKeys(notes: Note[]): KnownPropertyKey[] {
  const countsByType = new Map<string, Map<PropertyType, number>>()
  for (const note of notes) {
    for (const property of note.properties) {
      const counts = countsByType.get(property.key) ?? new Map<PropertyType, number>()
      counts.set(property.type, (counts.get(property.type) ?? 0) + 1)
      countsByType.set(property.key, counts)
    }
  }
  return [...countsByType.entries()]
    .map(([key, counts]) => ({
      key,
      type: [...counts.entries()]
        .sort((a, b) => b[1] - a[1] || collator.compare(a[0], b[0]))[0][0],
    }))
    .sort((a, b) => collator.compare(a.key, b.key))
}
