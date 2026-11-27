import { describe, it, expect } from 'vitest'
import {
  parseNoteViews,
  applyNoteView,
  groupNotes,
  MAX_NOTE_VIEWS,
  NO_VALUE_KEY,
} from './noteViews'
import type { NoteView, NoteViewCondition, PropertyOp } from './noteViews'
import type { Note, NoteProperty } from '../types'

const prop = (
  key: string,
  type: NoteProperty['type'],
  value: NoteProperty['value'],
): NoteProperty => ({ key, type, value })

const note = (over: Partial<Note> = {}): Note => ({
  id: 'n1',
  title: 'Untitled',
  content: '',
  blocks: '',
  status: 'draft',
  folder_id: null,
  tags: [],
  properties: [],
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  ...over,
})

const view = (over: Partial<NoteView> = {}): NoteView => ({
  id: 'v1',
  name: 'View',
  filters: { properties: [] },
  sort: { field: 'updated_at', direction: 'asc' },
  layout: 'table',
  ...over,
})

const cond = (
  key: string,
  op: PropertyOp,
  value?: string | number | boolean,
): NoteViewCondition => ({ key, op, ...(value !== undefined ? { value } : {}) })

const ids = (notes: Note[]) => notes.map(n => n.id)

describe('parseNoteViews', () => {
  it('parses a valid array of views', () => {
    const raw = JSON.stringify([
      {
        id: 'v-1',
        name: 'Rating alto',
        filters: {
          status: ['draft', 'published'],
          tags: ['react'],
          folder_id: 'folder-1',
          properties: [{ key: 'rating', op: 'gte', value: 7 }],
        },
        sort: { field: 'rating', direction: 'desc' },
        layout: 'kanban',
        groupBy: 'topic',
      },
    ])
    expect(parseNoteViews(raw)).toEqual([
      {
        id: 'v-1',
        name: 'Rating alto',
        filters: {
          status: ['draft', 'published'],
          tags: ['react'],
          folder_id: 'folder-1',
          properties: [{ key: 'rating', op: 'gte', value: 7 }],
        },
        sort: { field: 'rating', direction: 'desc' },
        layout: 'kanban',
        groupBy: 'topic',
      },
    ])
  })

  it('returns [] for undefined, empty, or malformed JSON', () => {
    expect(parseNoteViews(undefined)).toEqual([])
    expect(parseNoteViews('')).toEqual([])
    expect(parseNoteViews('{"name":')).toEqual([])
    expect(parseNoteViews('not json at all')).toEqual([])
  })

  it('returns [] when the parsed JSON is not an array', () => {
    expect(parseNoteViews(JSON.stringify({ views: [] }))).toEqual([])
    expect(parseNoteViews('42')).toEqual([])
  })

  it('drops entries missing a usable name', () => {
    const raw = JSON.stringify([
      { id: 'a', name: 'Real view', filters: { properties: [] } },
      { id: 'b' },
      { id: 'c', name: '' },
      { id: 'd', name: '   ' },
      { id: 'e', name: 42 },
      { id: 'f', name: null },
    ])
    const parsed = parseNoteViews(raw)
    expect(parsed.map(v => v.name)).toEqual(['Real view'])
  })

  it('generates distinct ids for entries missing one', () => {
    const raw = JSON.stringify([
      { name: 'One', filters: { properties: [] } },
      { name: 'Two', filters: { properties: [] } },
      { id: 'kept', name: 'Three', filters: { properties: [] } },
    ])
    const parsed = parseNoteViews(raw)
    expect(parsed[0].id).not.toBe('')
    expect(parsed[0].id).not.toBe(parsed[1].id)
    expect(parsed[2].id).toBe('kept')
  })

  it('truncates to MAX_NOTE_VIEWS entries', () => {
    const raw = JSON.stringify(
      Array.from({ length: MAX_NOTE_VIEWS + 5 }, (_, i) => ({
        name: `V${i}`,
        filters: { properties: [] },
      })),
    )
    const parsed = parseNoteViews(raw)
    expect(parsed).toHaveLength(MAX_NOTE_VIEWS)
    expect(parsed[0].name).toBe('V0')
    expect(parsed[MAX_NOTE_VIEWS - 1].name).toBe('V19')
  })

  it('invalid entries do not consume truncate slots', () => {
    const entries: unknown[] = []
    for (let i = 0; i < MAX_NOTE_VIEWS + 1; i++) {
      entries.push({ noNameHere: true })
      entries.push({ name: `V${i}`, filters: { properties: [] } })
    }
    const parsed = parseNoteViews(JSON.stringify(entries))
    // 21 valid views interleaved with 21 invalid ones: the cap must be
    // reached by valid views only, and the 21st valid view must be dropped.
    expect(parsed).toHaveLength(MAX_NOTE_VIEWS)
    expect(parsed.every(v => v.name.startsWith('V'))).toBe(true)
    expect(parsed[MAX_NOTE_VIEWS - 1].name).toBe('V19')
  })

  it('deduplicates stored ids, keeping the first occurrence', () => {
    const raw = JSON.stringify([
      { id: 'dup', name: 'First', filters: { properties: [] } },
      { id: 'dup', name: 'Second', filters: { properties: [] } },
      { id: 'unique', name: 'Third', filters: { properties: [] } },
    ])
    const parsed = parseNoteViews(raw)
    expect(parsed.map(v => v.name)).toEqual(['First', 'Third'])
    expect(parsed.map(v => v.id)).toEqual(['dup', 'unique'])
  })

  it('trims names and caps them at 100 characters', () => {
    const raw = JSON.stringify([
      {
        name: `  ${'x'.repeat(150)}  `,
        filters: { properties: [] },
      },
    ])
    const parsed = parseNoteViews(raw)
    expect(parsed[0].name).toBe('x'.repeat(100))
  })

  it('drops invalid conditions from filters.properties', () => {
    const raw = JSON.stringify([
      {
        name: 'V',
        filters: {
          properties: [
            { key: 'rating', op: 'gte', value: 7 },
            { op: 'contains', value: 'x' },
            { key: '', op: 'contains', value: 'x' },
            { key: 'rating', op: 'regex', value: 'x' },
            { key: 'title', op: 'contains' },
            { key: 'title', op: 'contains', value: null },
            { key: 'title', op: 'contains', value: ['x'] },
            { key: 'flag', op: 'is_set' },
          ],
        },
      },
    ])
    const parsed = parseNoteViews(raw)
    expect(parsed[0].filters.properties).toEqual([
      { key: 'rating', op: 'gte', value: 7 },
      { key: 'flag', op: 'is_set' },
    ])
  })

  it('defaults filters.properties to [] when filters or properties are absent', () => {
    const raw = JSON.stringify([
      { name: 'No properties key', filters: {} },
      { name: 'No filters at all' },
    ])
    const parsed = parseNoteViews(raw)
    expect(parsed).toHaveLength(2)
    expect(parsed[0].filters.properties).toEqual([])
    expect(parsed[1].filters.properties).toEqual([])
  })

  it('fills defaults for missing sort and layout', () => {
    const raw = JSON.stringify([{ name: 'Minimal' }])
    const parsed = parseNoteViews(raw)
    expect(parsed[0].sort).toEqual({ field: 'updated_at', direction: 'desc' })
    expect(parsed[0].layout).toBe('table')
  })

  it('sanitizes filter fields: invalid statuses, non-string tags and folder_id are dropped', () => {
    const raw = JSON.stringify([
      {
        name: 'V',
        filters: {
          status: ['draft', 'bogus'],
          tags: ['react', 7],
          folder_id: 9,
          properties: [],
        },
      },
    ])
    const parsed = parseNoteViews(raw)
    expect(parsed[0].filters.status).toEqual(['draft'])
    expect(parsed[0].filters.tags).toEqual(['react'])
    expect(parsed[0].filters.folder_id).toBeUndefined()
  })

  it('treats an empty status or tags array as no filter', () => {
    const raw = JSON.stringify([{ name: 'V', filters: { status: [], tags: [], properties: [] } }])
    const parsed = parseNoteViews(raw)
    expect(parsed[0].filters.status).toBeUndefined()
    expect(parsed[0].filters.tags).toBeUndefined()
  })
})

describe('applyNoteView filtering', () => {
  it('returns the notes unchanged when the view is null', () => {
    const notes = [note({ id: 'a' }), note({ id: 'b' })]
    expect(applyNoteView(notes, null)).toEqual(notes)
  })

  it('filters by status: only listed statuses match', () => {
    const notes = [
      note({ id: 'a', status: 'draft' }),
      note({ id: 'b', status: 'in_progress' }),
      note({ id: 'c', status: 'published' }),
    ]
    const v = view({ filters: { status: ['draft', 'published'], properties: [] } })
    expect(ids(applyNoteView(notes, v))).toEqual(['a', 'c'])
  })

  it('treats an empty status array as no filter', () => {
    const notes = [
      note({ id: 'a', status: 'draft' }),
      note({ id: 'b', status: 'published' }),
    ]
    const v = view({ filters: { status: [], properties: [] } })
    expect(ids(applyNoteView(notes, v))).toEqual(['a', 'b'])
  })

  it('matches tags case-insensitively', () => {
    const notes = [
      note({ id: 'a', tags: ['React'] }),
      note({ id: 'b', tags: ['go'] }),
      note({ id: 'c', tags: [] }),
    ]
    const v = view({ filters: { tags: ['REACT'], properties: [] } })
    expect(ids(applyNoteView(notes, v))).toEqual(['a'])
  })

  it('requires every tag to be present; a note missing one tag is excluded', () => {
    const notes = [
      note({ id: 'a', tags: ['react', 'css'] }),
      note({ id: 'b', tags: ['react'] }),
    ]
    const v = view({ filters: { tags: ['react', 'css'], properties: [] } })
    expect(ids(applyNoteView(notes, v))).toEqual(['a'])
  })

  it('filters by folder_id with exact match', () => {
    const notes = [
      note({ id: 'a', folder_id: 'folder-1' }),
      note({ id: 'b', folder_id: 'folder-2' }),
      note({ id: 'c', folder_id: null }),
    ]
    const v = view({ filters: { folder_id: 'folder-1', properties: [] } })
    expect(ids(applyNoteView(notes, v))).toEqual(['a'])
  })

  it('text ops are case-insensitive: contains, not_contains, equals', () => {
    const notes = [
      note({ id: 'a', properties: [prop('name', 'text', 'React Notes')] }),
      note({ id: 'b', properties: [prop('name', 'text', 'Go notes')] }),
      note({ id: 'c', properties: [prop('name', 'text', 'react')] }),
      note({ id: 'd', properties: [] }),
    ]
    const make = (op: PropertyOp, value: string) =>
      applyNoteView(notes, view({ filters: { properties: [cond('name', op, value)] } }))
    expect(ids(make('contains', 'REACT'))).toEqual(['a', 'c'])
    expect(ids(make('not_contains', 'REACT'))).toEqual(['b'])
    expect(ids(make('equals', 'REACT'))).toEqual(['c'])
  })

  it('number ops: eq, neq, gt, gte, lt, lte', () => {
    const notes = [
      note({ id: 'a', properties: [prop('rating', 'number', 7)] }),
      note({ id: 'b', properties: [prop('rating', 'number', 3)] }),
      note({ id: 'c', properties: [prop('rating', 'number', 10)] }),
    ]
    const make = (op: PropertyOp, value: number) =>
      ids(applyNoteView(notes, view({ filters: { properties: [cond('rating', op, value)] } })))
    expect(make('eq', 7)).toEqual(['a'])
    expect(make('neq', 7)).toEqual(['b', 'c'])
    expect(make('gt', 7)).toEqual(['c'])
    expect(make('gte', 7)).toEqual(['a', 'c'])
    expect(make('lt', 7)).toEqual(['b'])
    expect(make('lte', 7)).toEqual(['a', 'b'])
  })

  it('number ops with a non-number condition value never match, not even neq', () => {
    const notes = [note({ id: 'a', properties: [prop('rating', 'number', 7)] })]
    const make = (op: PropertyOp, value: string) =>
      ids(applyNoteView(notes, view({ filters: { properties: [cond('rating', op, value)] } })))
    expect(make('gt', '5')).toEqual([])
    expect(make('neq', 'x')).toEqual([])
  })

  it('date ops compare YYYY-MM-DD strings: before, on, after', () => {
    const notes = [
      note({ id: 'a', properties: [prop('due', 'date', '2026-01-05')] }),
      note({ id: 'b', properties: [prop('due', 'date', '2026-03-10')] }),
      note({ id: 'c', properties: [prop('due', 'date', '2026-12-01')] }),
    ]
    const make = (op: PropertyOp, value: string) =>
      ids(applyNoteView(notes, view({ filters: { properties: [cond('due', op, value)] } })))
    expect(make('before', '2026-03-10')).toEqual(['a'])
    expect(make('on', '2026-03-10')).toEqual(['b'])
    expect(make('after', '2026-03-10')).toEqual(['c'])
  })

  it('date ops with a non-string condition value do not match', () => {
    const notes = [note({ id: 'a', properties: [prop('due', 'date', '2026-01-05')] })]
    const v = view({ filters: { properties: [cond('due', 'before', 20260105)] } })
    expect(ids(applyNoteView(notes, v))).toEqual([])
  })

  it('boolean ops: is_true, is_false', () => {
    const notes = [
      note({ id: 'a', properties: [prop('done', 'boolean', true)] }),
      note({ id: 'b', properties: [prop('done', 'boolean', false)] }),
      note({ id: 'c', properties: [] }),
    ]
    const make = (op: PropertyOp) =>
      ids(applyNoteView(notes, view({ filters: { properties: [cond('done', op)] } })))
    expect(make('is_true')).toEqual(['a'])
    expect(make('is_false')).toEqual(['b'])
  })

  it('list ops are case-insensitive: includes, not_includes', () => {
    const notes = [
      note({ id: 'a', properties: [prop('topics', 'list', ['Design', 'Research'])] }),
      note({ id: 'b', properties: [prop('topics', 'list', ['Backend'])] }),
      note({ id: 'c', properties: [] }),
    ]
    const make = (op: PropertyOp, value: string) =>
      ids(applyNoteView(notes, view({ filters: { properties: [cond('topics', op, value)] } })))
    expect(make('includes', 'DESIGN')).toEqual(['a'])
    expect(make('not_includes', 'DESIGN')).toEqual(['b'])
  })

  it('is_set and is_not_set work for every property type', () => {
    const full = note({
      id: 'a',
      properties: [
        prop('t', 'text', 'x'),
        prop('n', 'number', 1),
        prop('d', 'date', '2026-01-01'),
        prop('b', 'boolean', true),
        prop('l', 'list', ['x']),
      ],
    })
    const empty = note({ id: 'b', properties: [] })
    const make = (key: string, op: PropertyOp) =>
      ids(applyNoteView([full, empty], view({ filters: { properties: [cond(key, op)] } })))
    const keys = ['t', 'n', 'd', 'b', 'l']
    expect(keys.map(k => make(k, 'is_set'))).toEqual(keys.map(() => ['a']))
    expect(keys.map(k => make(k, 'is_not_set'))).toEqual(keys.map(() => ['b']))
  })

  it('a condition on a key the note does not have matches only is_not_set', () => {
    const notes = [note({ id: 'a', properties: [] })]
    const make = (op: PropertyOp) =>
      ids(applyNoteView(notes, view({ filters: { properties: [cond('ghost', op, 'x')] } })))
    expect(make('contains')).toEqual([])
    expect(make('not_contains')).toEqual([])
    expect(make('eq')).toEqual([])
    expect(make('is_set')).toEqual([])
    expect(make('is_not_set')).toEqual(['a'])
  })

  it('an op that does not apply to the property type never matches', () => {
    const notes = [
      note({ id: 'a', properties: [prop('mixed', 'text', '42')] }),
      note({ id: 'b', properties: [prop('mixed', 'number', 42)] }),
    ]
    const gt = view({ filters: { properties: [cond('mixed', 'gt', 10)] } })
    expect(ids(applyNoteView(notes, gt))).toEqual(['b'])
    const contains = view({ filters: { properties: [cond('mixed', 'contains', '4')] } })
    expect(ids(applyNoteView(notes, contains))).toEqual(['a'])
  })

  it('ANDs multiple conditions and note-field filters', () => {
    const notes = [
      note({
        id: 'a',
        status: 'published',
        tags: ['design'],
        properties: [prop('rating', 'number', 8), prop('topics', 'list', ['design'])],
      }),
      note({
        id: 'b',
        status: 'published',
        tags: ['design'],
        properties: [prop('rating', 'number', 8), prop('topics', 'list', ['backend'])],
      }),
      note({
        id: 'c',
        status: 'draft',
        tags: ['design'],
        properties: [prop('rating', 'number', 8), prop('topics', 'list', ['design'])],
      }),
    ]
    const v = view({
      filters: {
        status: ['published'],
        tags: ['design'],
        properties: [cond('rating', 'gte', 7), cond('topics', 'includes', 'design')],
      },
    })
    expect(ids(applyNoteView(notes, v))).toEqual(['a'])
  })
})

describe('applyNoteView sorting', () => {
  it('sorts by updated_at asc and desc', () => {
    const notes = [
      note({ id: 'c', updated_at: '2026-03-01T00:00:00Z' }),
      note({ id: 'a', updated_at: '2026-01-01T00:00:00Z' }),
      note({ id: 'b', updated_at: '2026-02-01T00:00:00Z' }),
    ]
    const asc = view({ sort: { field: 'updated_at', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['a', 'b', 'c'])
    const desc = view({ sort: { field: 'updated_at', direction: 'desc' } })
    expect(ids(applyNoteView(notes, desc))).toEqual(['c', 'b', 'a'])
  })

  it('sorts by title using locale comparison', () => {
    const notes = [
      note({ id: 'a', title: 'Cherry' }),
      note({ id: 'b', title: 'árbol' }),
      note({ id: 'c', title: 'Banana' }),
    ]
    const asc = view({ sort: { field: 'title', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['b', 'c', 'a'])
  })

  it('sorts titles with natural numeric ordering', () => {
    const notes = [
      note({ id: 'a', title: 'Note 10' }),
      note({ id: 'b', title: 'Note 2' }),
      note({ id: 'c', title: 'Note 1' }),
    ]
    const asc = view({ sort: { field: 'title', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['c', 'b', 'a'])
  })

  it('sorts timestamps of mixed precision chronologically', () => {
    const notes = [
      note({ id: 'a', created_at: '2026-01-01T00:00:00.5Z' }),
      note({ id: 'b', created_at: '2026-01-01T00:00:00Z' }),
      note({ id: 'c', created_at: '2026-01-01T00:00:01Z' }),
    ]
    const asc = view({ sort: { field: 'created_at', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['b', 'a', 'c'])
  })

  it('does not mutate the input array when sorting', () => {
    const notes = [
      note({ id: 'b', updated_at: '2026-02-01T00:00:00Z' }),
      note({ id: 'a', updated_at: '2026-01-01T00:00:00Z' }),
      note({ id: 'c', updated_at: '2026-03-01T00:00:00Z' }),
    ]
    const before = ids(notes)
    const result = applyNoteView(notes, view({ sort: { field: 'updated_at', direction: 'desc' } }))
    expect(ids(result)).toEqual(['c', 'b', 'a'])
    expect(result).not.toBe(notes)
    expect(ids(notes)).toEqual(before)
  })

  it('sorts by status: draft < in_progress < published', () => {
    const notes = [
      note({ id: 'a', status: 'published' }),
      note({ id: 'b', status: 'in_progress' }),
      note({ id: 'c', status: 'draft' }),
    ]
    const asc = view({ sort: { field: 'status', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['c', 'b', 'a'])
    const desc = view({ sort: { field: 'status', direction: 'desc' } })
    expect(ids(applyNoteView(notes, desc))).toEqual(['a', 'b', 'c'])
  })

  it('sorts a number property numerically, not lexicographically', () => {
    const notes = [
      note({ id: 'two', properties: [prop('rating', 'number', 2)] }),
      note({ id: 'ten', properties: [prop('rating', 'number', 10)] }),
      note({ id: 'one', properties: [prop('rating', 'number', 1)] }),
    ]
    const asc = view({ sort: { field: 'rating', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['one', 'two', 'ten'])
    const desc = view({ sort: { field: 'rating', direction: 'desc' } })
    expect(ids(applyNoteView(notes, desc))).toEqual(['ten', 'two', 'one'])
  })

  it('sorts a date property lexicographically', () => {
    const notes = [
      note({ id: 'a', properties: [prop('due', 'date', '2026-03-01')] }),
      note({ id: 'b', properties: [prop('due', 'date', '2026-01-10')] }),
      note({ id: 'c', properties: [prop('due', 'date', '2026-12-01')] }),
    ]
    const asc = view({ sort: { field: 'due', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['b', 'a', 'c'])
  })

  it('sorts a text property by locale', () => {
    const notes = [
      note({ id: 'a', properties: [prop('topic', 'text', 'Banana')] }),
      note({ id: 'b', properties: [prop('topic', 'text', 'apple')] }),
      note({ id: 'c', properties: [prop('topic', 'text', 'cherry')] }),
    ]
    const asc = view({ sort: { field: 'topic', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['b', 'a', 'c'])
  })

  it('sorts a list property by its first item', () => {
    const notes = [
      note({ id: 'a', properties: [prop('topics', 'list', ['Zed', 'Alpha'])] }),
      note({ id: 'b', properties: [prop('topics', 'list', ['Beta'])] }),
      note({ id: 'c', properties: [prop('topics', 'list', ['Gamma'])] }),
    ]
    const asc = view({ sort: { field: 'topics', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['b', 'c', 'a'])
    const desc = view({ sort: { field: 'topics', direction: 'desc' } })
    expect(ids(applyNoteView(notes, desc))).toEqual(['a', 'c', 'b'])
  })

  it('sorts a boolean property false < true', () => {
    const notes = [
      note({ id: 'a', properties: [prop('done', 'boolean', true)] }),
      note({ id: 'b', properties: [prop('done', 'boolean', false)] }),
      note({ id: 'c', properties: [prop('done', 'boolean', true)] }),
    ]
    const asc = view({ sort: { field: 'done', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['b', 'a', 'c'])
    const desc = view({ sort: { field: 'done', direction: 'desc' } })
    expect(ids(applyNoteView(notes, desc))).toEqual(['a', 'c', 'b'])
  })

  it('sorts notes missing the sort value last, in both directions', () => {
    const notes = [
      note({ id: 'a', properties: [prop('rating', 'number', 5)] }),
      note({ id: 'b', properties: [] }),
      note({ id: 'c', properties: [prop('rating', 'number', 1)] }),
      note({ id: 'd', properties: [prop('other', 'text', 'x')] }),
    ]
    const asc = view({ sort: { field: 'rating', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['c', 'a', 'b', 'd'])
    const desc = view({ sort: { field: 'rating', direction: 'desc' } })
    expect(ids(applyNoteView(notes, desc))).toEqual(['a', 'c', 'b', 'd'])
  })

  it('keeps input order for equal sort keys', () => {
    const notes = [
      note({ id: 'a', title: 'First', updated_at: '2026-01-01T00:00:00Z' }),
      note({ id: 'b', title: 'Second', updated_at: '2026-01-01T00:00:00Z' }),
      note({ id: 'c', title: 'Third', updated_at: '2026-01-01T00:00:00Z' }),
    ]
    const asc = view({ sort: { field: 'updated_at', direction: 'asc' } })
    expect(ids(applyNoteView(notes, asc))).toEqual(['a', 'b', 'c'])
    const desc = view({ sort: { field: 'updated_at', direction: 'desc' } })
    expect(ids(applyNoteView(notes, desc))).toEqual(['a', 'b', 'c'])
  })
})

describe('groupNotes', () => {
  it('groups by status in fixed order, skipping empty statuses', () => {
    const notes = [
      note({ id: 'p', status: 'published' }),
      note({ id: 'd1', status: 'draft' }),
      note({ id: 'd2', status: 'draft' }),
    ]
    const groups = groupNotes(notes, 'status')
    expect(groups.map(g => g.key)).toEqual(['draft', 'published'])
    expect(groups[0].notes.map(n => n.id)).toEqual(['d1', 'd2'])
    expect(groups[1].notes.map(n => n.id)).toEqual(['p'])
  })

  it('groups by a text property into distinct sections sorted by locale', () => {
    const notes = [
      note({ id: 'a', properties: [prop('topic', 'text', 'Banana')] }),
      note({ id: 'b', properties: [prop('topic', 'text', 'apple')] }),
      note({ id: 'c', properties: [prop('topic', 'text', 'cherry')] }),
      note({ id: 'd', properties: [prop('topic', 'text', 'Banana')] }),
    ]
    const groups = groupNotes(notes, 'topic')
    expect(groups.map(g => g.key)).toEqual(['apple', 'Banana', 'cherry'])
    const banana = groups.find(g => g.key === 'Banana')
    expect(banana?.notes.map(n => n.id)).toEqual(['a', 'd'])
  })

  it('groups by a list property with the note in every matching section', () => {
    const notes = [
      note({ id: 'a', properties: [prop('topics', 'list', ['Design', 'Research'])] }),
      note({ id: 'b', properties: [prop('topics', 'list', ['Research'])] }),
      note({ id: 'c', properties: [prop('topics', 'list', ['Backend'])] }),
    ]
    const groups = groupNotes(notes, 'topics')
    expect(groups.map(g => g.key)).toEqual(['Backend', 'Design', 'Research'])
    const research = groups.find(g => g.key === 'Research')
    expect(research?.notes.map(n => n.id)).toEqual(['a', 'b'])
    const design = groups.find(g => g.key === 'Design')
    expect(design?.notes.map(n => n.id)).toEqual(['a'])
  })

  it('groups by a number property with sections sorted numerically', () => {
    const notes = [
      note({ id: 'a', properties: [prop('rating', 'number', 10)] }),
      note({ id: 'b', properties: [prop('rating', 'number', 2)] }),
    ]
    const groups = groupNotes(notes, 'rating')
    expect(groups.map(g => g.key)).toEqual(['2', '10'])
    expect(groups[0].notes.map(n => n.id)).toEqual(['b'])
  })

  it('puts notes without the property in a last "No value" section', () => {
    const notes = [
      note({ id: 'a', properties: [prop('topic', 'text', 'zebra')] }),
      note({ id: 'b', properties: [] }),
      note({ id: 'c', properties: [prop('topic', 'text', 'Apple')] }),
    ]
    const groups = groupNotes(notes, 'topic')
    expect(groups.map(g => g.key)).toEqual(['Apple', 'zebra', 'No value'])
    expect(groups[groups.length - 1].notes.map(n => n.id)).toEqual(['b'])
  })

  it('treats an empty list value as no value', () => {
    const notes = [
      note({ id: 'a', properties: [prop('topics', 'list', ['Design'])] }),
      note({ id: 'b', properties: [prop('topics', 'list', [])] }),
    ]
    const groups = groupNotes(notes, 'topics')
    expect(groups.map(g => g.key)).toEqual(['Design', 'No value'])
    expect(groups[1].notes.map(n => n.id)).toEqual(['b'])
  })

  it('exports NO_VALUE_KEY and uses it for the no-value section', () => {
    expect(NO_VALUE_KEY).toBe('No value')
    const notes = [
      note({ id: 'a', properties: [prop('topic', 'text', 'x')] }),
      note({ id: 'b', properties: [] }),
    ]
    const groups = groupNotes(notes, 'topic')
    expect(groups[groups.length - 1].key).toBe(NO_VALUE_KEY)
  })

  it('returns a single group when no field is given', () => {
    const notes = [note({ id: 'a' }), note({ id: 'b' })]
    expect(groupNotes(notes)).toEqual([{ key: '', notes }])
    expect(groupNotes(notes, '')).toEqual([{ key: '', notes }])
  })

  it('returns [] for empty input', () => {
    expect(groupNotes([], 'status')).toEqual([])
    expect(groupNotes([])).toEqual([])
  })
})
