import { describe, it, expect } from 'vitest'
import { aggregateExpensesByCategory } from './finance'
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
