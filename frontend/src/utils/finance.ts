import type { RecentTransaction } from '../types'

export const CHART_PALETTE = [
  '#6366f1',
  '#10b981',
  '#f59e0b',
  '#f43f5e',
  '#0ea5e9',
  '#8b5cf6',
  '#d946ef',
  '#14b8a6',
] as const

export type DonutSlice = {
  label: string
  value: number
  color: string
}

const isExpense = (t: RecentTransaction) =>
  t.type === 'expense' && t.category !== 'All' && t.category !== ''

export function aggregateExpensesByCategory(
  transactions: RecentTransaction[],
): DonutSlice[] {
  const totals = new Map<string, number>()
  for (const t of transactions) {
    if (!isExpense(t)) continue
    totals.set(t.category, (totals.get(t.category) ?? 0) + t.amount)
  }
  const slices = Array.from(totals, ([label, value]) => ({ label, value }))
  slices.sort((a, b) => b.value - a.value)
  return slices.map((s, i) => ({
    ...s,
    color: CHART_PALETTE[i % CHART_PALETTE.length],
  }))
}
