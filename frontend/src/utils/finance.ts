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

export type DayPoint = {
  date: string
  total: number
}

const dayKey = (iso: string) => iso.slice(0, 10)

function addDays(yyyyMmDd: string, n: number): string {
  const d = new Date(yyyyMmDd + 'T00:00:00Z')
  d.setUTCDate(d.getUTCDate() + n)
  return d.toISOString().slice(0, 10)
}

export function aggregateExpensesByDay(
  transactions: RecentTransaction[],
): DayPoint[] {
  const totals = new Map<string, number>()
  for (const t of transactions) {
    if (!isExpense(t)) continue
    const k = dayKey(t.date)
    totals.set(k, (totals.get(k) ?? 0) + t.amount)
  }
  if (totals.size === 0) return []
  const dates = Array.from(totals.keys()).sort()
  const start = dates[0]
  const end = dates[dates.length - 1]
  const out: DayPoint[] = []
  let cur = start
  while (cur <= end) {
    out.push({ date: cur, total: totals.get(cur) ?? 0 })
    cur = addDays(cur, 1)
  }
  return out
}
