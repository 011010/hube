import { describe, it, expect } from 'vitest'
import { aggregateExpensesByCategory, aggregateExpensesByDay } from './finance'
import type { RecentTransaction } from '../types'

const tx = (over: Partial<RecentTransaction>): RecentTransaction => ({
  id: 'x',
  amount: 0,
  type: 'expense',
  category: 'Food',
  description: '',
  date: '2026-07-10T00:00:00Z',
  ...over,
})

describe('aggregateExpensesByCategory', () => {
  it('groups expenses, drops income and "All", sorts desc, assigns palette', () => {
    const r = aggregateExpensesByCategory([
      tx({ category: 'Food', amount: 100 }),
      tx({ category: 'Food', amount: 50 }),
      tx({ category: 'Transport', amount: 80 }),
      tx({ category: 'All', amount: 999 }),
      tx({ type: 'income', amount: 500, category: 'Salary' }),
      tx({ category: '', amount: 123 }),
    ])
    expect(r.map(s => s.label)).toEqual(['Food', 'Transport'])
    expect(r.map(s => s.value)).toEqual([150, 80])
    expect(r[0].color).toBe('#6366f1')
    expect(r[1].color).toBe('#10b981')
  })

  it('empty input returns []', () => {
    expect(aggregateExpensesByCategory([])).toEqual([])
  })
})

describe('aggregateExpensesByDay', () => {
  it('groups by day, fills missing days with 0, sorts asc', () => {
    const r = aggregateExpensesByDay([
      tx({ date: '2026-07-10T10:00:00Z', amount: 30 }),
      tx({ date: '2026-07-12T10:00:00Z', amount: 50 }),
      tx({ date: '2026-07-10T22:00:00Z', amount: 20 }),
    ])
    expect(r).toEqual([
      { date: '2026-07-10', total: 50 },
      { date: '2026-07-11', total: 0 },
      { date: '2026-07-12', total: 50 },
    ])
  })

  it('single day returns one point', () => {
    const r = aggregateExpensesByDay([tx({ date: '2026-07-10', amount: 10 })])
    expect(r).toEqual([{ date: '2026-07-10', total: 10 }])
  })

  it('empty input returns []', () => {
    expect(aggregateExpensesByDay([])).toEqual([])
  })

  it('drops income and "All" and empty category', () => {
    const r = aggregateExpensesByDay([
      tx({ date: '2026-07-10', amount: 10 }),
      tx({ type: 'income', date: '2026-07-10', amount: 999, category: 'Salary' }),
      tx({ category: 'All', date: '2026-07-10', amount: 999 }),
      tx({ category: '', date: '2026-07-10', amount: 999 }),
    ])
    expect(r).toEqual([{ date: '2026-07-10', total: 10 }])
  })
})
