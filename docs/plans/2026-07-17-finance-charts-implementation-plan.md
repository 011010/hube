# Finance Charts on Dashboard Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a "Expenses by category" donut chart and a "Spending over time" area chart to the Dashboard's Finance section, driven by data the existing `/api/v1/finance/summary` endpoint already returns.

**Architecture:** Pure aggregation functions in `src/utils/finance.ts` feed two hand-rolled SVG molecule components (`DonutChart`, `AreaChart`) wrapped in a new `ChartCard` molecule. The existing `FinanceWidgets` organism renders the two charts in a responsive grid between the stat cards and the recent transactions list. No backend, no API, no runtime dependencies change. Vitest is added as a devDep to unit-test the aggregation functions only (the SVG components are verified by `pnpm build` + visual inspection, by design).

**Tech Stack:** React 19, TypeScript, Tailwind 4, hand-rolled SVG (no chart library), Vitest 3 (devDep).

**Design doc:** `docs/plans/2026-07-17-finance-charts-design.md`

**Conventions:** English in all artifacts (code, comments, identifiers). No comments in code unless this plan explicitly adds them. Conventional commits. Run all commands from `frontend/` unless stated otherwise.

---

## Task 1: Add Vitest test infra

**Files:**
- Create: `frontend/vitest.config.ts`
- Modify: `frontend/package.json` (add `vitest` devDep and `test` script)

**Step 1: Add vitest devDep and the test script**

Run from `frontend/`:
```bash
pnpm add -D vitest@^3.0.0
```
Expected: `vitest` appears under `devDependencies` in `package.json`; `pnpm-lock.yaml` is updated.

If `pnpm install` reports a peer-dependency mismatch with `vite ^8`, bump to the version of vitest that tracks vite 8 (check `pnpm view vitest versions` or the vitest release notes) and re-run.

Edit `package.json` and add (inside `"scripts"`, next to `lint`):
```json
"test": "vitest run"
```

**Step 2: Create `vitest.config.ts`**

Create `frontend/vitest.config.ts`:
```ts
import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    environment: 'node',
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
  },
})
```

**Step 3: Run tests to confirm infra works (no tests yet)**

```bash
cd frontend && pnpm test
```
Expected: `No test files found, exiting with code 0` (or vitest's equivalent "no tests"). Exit 0. This proves vitest is wired correctly.

**Step 4: Run build and lint to confirm nothing else broke**

```bash
cd frontend && pnpm run build && pnpm run lint
```
Expected: both pass.

**Step 5: Commit**

```bash
cd /Users/ios.husari.orozco/Documents/projects/hube && git add frontend/package.json frontend/pnpm-lock.yaml frontend/vitest.config.ts && git commit -m "chore(frontend): add vitest for finance aggregation tests"
```

---

## Task 2: `aggregateExpensesByCategory` (TDD)

**Files:**
- Create: `frontend/src/utils/finance.ts`
- Create: `frontend/src/utils/finance.test.ts`

**Step 1: Write the failing test**

Create `frontend/src/utils/finance.test.ts`:
```ts
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
```

**Step 2: Run the test and confirm it fails**

```bash
cd frontend && pnpm test
```
Expected: FAIL — `aggregateExpensesByCategory` is not exported from `./finance`.

**Step 3: Implement the function (and the palette)**

Create `frontend/src/utils/finance.ts`:
```ts
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
```

**Step 4: Run the test and confirm it passes**

```bash
cd frontend && pnpm test
```
Expected: PASS — 2 tests pass.

**Step 5: Commit**

```bash
cd /Users/ios.husari.orozco/Documents/projects/hube && git add frontend/src/utils/finance.ts frontend/src/utils/finance.test.ts && git commit -m "feat(frontend): aggregate expenses by category"
```

---

## Task 3: `aggregateExpensesByDay` (TDD)

**Files:**
- Modify: `frontend/src/utils/finance.ts`
- Modify: `frontend/src/utils/finance.test.ts`

**Step 1: Write the failing tests**

Append to `frontend/src/utils/finance.test.ts`:
```ts
import { aggregateExpensesByDay } from './finance'

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
```

**Step 2: Run the tests and confirm they fail**

```bash
cd frontend && pnpm test
```
Expected: FAIL — `aggregateExpensesByDay` is not exported from `./finance`.

**Step 3: Implement the function**

Append to `frontend/src/utils/finance.ts`:
```ts
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
```

**Step 4: Run the tests and confirm they all pass**

```bash
cd frontend && pnpm test
```
Expected: PASS — 6 tests pass (2 from task 2 + 4 from this task).

**Step 5: Commit**

```bash
cd /Users/ios.husari.orozco/Documents/projects/hube && git add frontend/src/utils/finance.ts frontend/src/utils/finance.test.ts && git commit -m "feat(frontend): aggregate expenses by day, fill gaps with zero"
```

---

## Task 4: `ChartCard` molecule

**Files:**
- Create: `frontend/src/components/molecules/ChartCard.tsx`

`ChartCard` is a presentational card matching the "Recent transactions" card (surface-elevated + border + uppercase muted title). It is **not** unit-tested (visual component; verified by `pnpm build`).

**Step 1: Create the component**

Create `frontend/src/components/molecules/ChartCard.tsx`:
```tsx
import type { ReactNode } from 'react'

export function ChartCard({
  title,
  children,
  empty,
}: {
  title: string
  children: ReactNode
  empty?: string
}) {
  return (
    <div className="bg-surface-elevated border border-border rounded-xl overflow-hidden">
      <div className="px-5 py-3 border-b border-border">
        <span className="text-xs font-medium text-text-secondary uppercase tracking-wider">
          {title}
        </span>
      </div>
      <div className="px-5 py-4">
        {empty ? (
          <p className="text-sm text-text-muted text-center py-8">{empty}</p>
        ) : (
          children
        )}
      </div>
    </div>
  )
}
```

**Step 2: Run build and lint to verify**

```bash
cd frontend && pnpm run build && pnpm run lint
```
Expected: both pass.

**Step 3: Commit**

```bash
cd /Users/ios.husari.orozco/Documents/projects/hube && git add frontend/src/components/molecules/ChartCard.tsx && git commit -m "feat(frontend): add ChartCard molecule"
```

---

## Task 5: `DonutChart` molecule

**Files:**
- Create: `frontend/src/components/molecules/DonutChart.tsx`

`DonutChart` renders a donut (segments via `stroke-dasharray` on a single rotated circle) with the total at the centre and a legend list below. Hover dims other segments. Not unit-tested (visual SVG; verified by `pnpm build`).

**Step 1: Create the component**

Create `frontend/src/components/molecules/DonutChart.tsx`:
```tsx
import { useState } from 'react'

type Slice = { label: string; value: number; color: string }

export function DonutChart({
  data,
  formatValue,
  centerLabel,
}: {
  data: Slice[]
  formatValue: (n: number) => string
  centerLabel?: string
}) {
  const [hover, setHover] = useState<number | null>(null)
  const total = data.reduce((s, d) => s + d.value, 0)
  const size = 200
  const stroke = 44
  const r = (size - stroke) / 2
  const c = 2 * Math.PI * r
  let offset = 0
  const a11y =
    data.length > 0
      ? `Expenses by category.${data
          .slice(0, 3)
          .map(s => ` ${s.label} ${formatValue(s.value)}`)
          .join(',')}.`
      : 'Expenses by category, no data.'

  if (data.length === 0 || total === 0) return null

  return (
    <div>
      <div className="flex flex-col items-center gap-4">
        <svg
          width={size}
          height={size}
          viewBox={`0 0 ${size} ${size}`}
          role="img"
          aria-label={a11y}
        >
          <g transform={`translate(${size / 2} ${size / 2}) rotate(-90)`}>
            {data.map((s, i) => {
              const frac = s.value / total
              const len = c * frac
              const seg = (
                <circle
                  key={s.label}
                  r={r}
                  fill="none"
                  stroke={s.color}
                  strokeWidth={stroke}
                  strokeDasharray={`${len} ${c - len}`}
                  strokeDashoffset={-offset}
                  opacity={hover === null || hover === i ? 1 : 0.4}
                  onMouseEnter={() => setHover(i)}
                  onMouseLeave={() => setHover(null)}
                  style={{ transition: 'opacity 120ms' }}
                />
              )
              offset += len
              return seg
            })}
          </g>
          <text
            x={size / 2}
            y={size / 2 - 2}
            textAnchor="middle"
            fill="currentColor"
            className="text-text-primary"
            style={{ fontSize: 22, fontWeight: 600 }}
          >
            {formatValue(total)}
          </text>
          <text
            x={size / 2}
            y={size / 2 + 18}
            textAnchor="middle"
            fill="currentColor"
            className="text-text-muted"
            style={{ fontSize: 11 }}
          >
            {centerLabel ?? 'Expenses'}
          </text>
        </svg>

        <ul className="w-full space-y-1.5">
          {data.map((s, i) => (
            <li
              key={s.label}
              className="flex items-center gap-2 text-sm"
              onMouseEnter={() => setHover(i)}
              onMouseLeave={() => setHover(null)}
            >
              <span
                className="w-2.5 h-2.5 rounded-full shrink-0"
                style={{ backgroundColor: s.color }}
              />
              <span className="text-text-secondary truncate flex-1">
                {s.label}
              </span>
              <span className="text-text-primary tabular-nums">
                {formatValue(s.value)}
              </span>
              <span className="text-text-muted tabular-nums w-12 text-right">
                {Math.round((s.value / total) * 100)}%
              </span>
            </li>
          ))}
        </ul>

        <p className="text-xs text-text-muted w-full">Excludes uncategorized.</p>
      </div>
    </div>
  )
}
```

**Step 2: Run build and lint**

```bash
cd frontend && pnpm run build && pnpm run lint
```
Expected: both pass. If the TypeScript build complains about an implicit `any` in the slice map (e.g. `data.slice(0, 3).map(...)` typing), it should be fine because `data: Slice[]`; if not, the implementer tightens the type. No known issues.

**Step 3: Commit**

```bash
cd /Users/ios.husari.orozco/Documents/projects/hube && git add frontend/src/components/molecules/DonutChart.tsx && git commit -m "feat(frontend): add DonutChart SVG component"
```

---

## Task 6: `AreaChart` molecule

**Files:**
- Create: `frontend/src/components/molecules/AreaChart.tsx`

`AreaChart` renders an SVG area + line with axes, gridlines, ~5 x-ticks, and a hover tooltip. Not unit-tested (visual SVG; verified by `pnpm build`).

**Step 1: Create the component**

Create `frontend/src/components/molecules/AreaChart.tsx`:
```tsx
import { useRef, useState } from 'react'

type Point = { date: string; total: number }

export function AreaChart({
  data,
  formatValue,
  formatDate,
  height = 200,
}: {
  data: Point[]
  formatValue: (n: number) => string
  formatDate: (iso: string) => string
  height?: number
}) {
  const width = 600
  const padL = 52
  const padR = 12
  const padT = 16
  const padB = 28
  const innerW = width - padL - padR
  const innerH = height - padT - padB
  const yMax = Math.max(1, ...data.map(d => d.total))
  const xStep = data.length > 1 ? innerW / (data.length - 1) : 0
  const yTicks = niceTicks(yMax, 4)
  const yTop = yTicks[yTicks.length - 1] || 1
  const xPos = (i: number) => padL + i * xStep
  const yPos = (v: number) => padT + innerH - (v / yTop) * innerH

  const linePath =
    data.length > 0
      ? `M ${xPos(0)} ${yPos(data[0].total)} ` +
        data
          .slice(1)
          .map((d, i) => `L ${xPos(i + 1)} ${yPos(d.total)}`)
          .join(' ')
      : ''
  const areaPath =
    data.length > 0
      ? `M ${xPos(0)} ${yPos(data[0].total)} ` +
        data
          .slice(1)
          .map((d, i) => `L ${xPos(i + 1)} ${yPos(d.total)}`)
          .join(' ') +
        ` L ${xPos(data.length - 1)} ${padT + innerH} L ${xPos(0)} ${padT + innerH} Z`
      : ''

  const xTickIdx = pickXTicks(data.length, 5)
  const [hover, setHover] = useState<number | null>(null)
  const svgRef = useRef<SVGSVGElement | null>(null)
  const [tip, setTip] = useState<{ x: number; y: number } | null>(null)

  const total = data.reduce((s, d) => s + d.total, 0)
  const a11y =
    data.length > 0
      ? `Daily expenses, ${formatDate(data[0].date)} to ${formatDate(
          data[data.length - 1].date,
        )}, total ${formatValue(total)}.`
      : 'Daily expenses, no data.'

  return (
    <div className="relative">
      <svg
        ref={svgRef}
        viewBox={`0 0 ${width} ${height}`}
        width="100%"
        height={height}
        role="img"
        aria-label={a11y}
        onMouseLeave={() => {
          setHover(null)
          setTip(null)
        }}
      >
        {yTicks.map(t => (
          <g key={t}>
            <line
              x1={padL}
              x2={width - padR}
              y1={yPos(t)}
              y2={yPos(t)}
              stroke="currentColor"
              className="text-text-muted"
              strokeOpacity={0.3}
            />
            <text
              x={padL - 6}
              y={yPos(t) + 3}
              textAnchor="end"
              fontSize={10}
              fill="currentColor"
              className="text-text-muted"
            >
              {formatValue(t)}
            </text>
          </g>
        ))}

        {xTickIdx.map(i => (
          <text
            key={data[i].date}
            x={xPos(i)}
            y={height - 8}
            textAnchor="middle"
            fontSize={10}
            fill="currentColor"
            className="text-text-muted"
          >
            {formatDate(data[i].date)}
          </text>
        ))}

        <path
          d={areaPath}
          fill="currentColor"
          className="text-(--color-accent)"
          fillOpacity={0.18}
        />
        <path
          d={linePath}
          fill="none"
          stroke="currentColor"
          className="text-(--color-accent)"
          strokeWidth={2}
          strokeLinejoin="round"
          strokeLinecap="round"
        />

        {data.map((d, i) => (
          <circle
            key={d.date}
            cx={xPos(i)}
            cy={yPos(d.total)}
            r={3}
            fill="currentColor"
            className="text-(--color-accent)"
            opacity={hover === i ? 1 : 0}
          />
        ))}

        <rect
          x={padL}
          y={padT}
          width={innerW}
          height={innerH}
          fill="transparent"
          onMouseMove={e => {
            const rect = svgRef.current?.getBoundingClientRect()
            if (!rect) return
            const rx = ((e.clientX - rect.left) / rect.width) * width
            if (data.length === 0) return
            const idx = Math.max(
              0,
              Math.min(data.length - 1, Math.round((rx - padL) / xStep)),
            )
            setHover(idx)
            setTip({ x: e.clientX - rect.left, y: e.clientY - rect.top })
          }}
        />

        {hover !== null && (
          <line
            x1={xPos(hover)}
            x2={xPos(hover)}
            y1={padT}
            y2={padT + innerH}
            stroke="currentColor"
            className="text-text-muted"
            strokeOpacity={0.7}
            strokeDasharray="3 3"
          />
        )}
      </svg>

      {hover !== null && tip && data[hover] && (
        <div
          className="pointer-events-none absolute bg-surface-elevated border border-border rounded-md px-2 py-1 text-xs shadow"
          style={{ left: tip.x + 8, top: Math.max(0, tip.y - 44) }}
        >
          <div className="text-text-muted">{formatDate(data[hover].date)}</div>
          <div className="text-text-primary font-medium tabular-nums">
            {formatValue(data[hover].total)}
          </div>
        </div>
      )}
    </div>
  )
}

function niceNumber(x: number): number {
  const exp = Math.floor(Math.log10(x))
  const f = x / Math.pow(10, exp)
  const nf = f < 1.5 ? 1 : f < 3 ? 2 : f < 7 ? 5 : 10
  return nf * Math.pow(10, exp)
}

function niceTicks(max: number, count: number): number[] {
  if (max <= 0) return [0]
  const step = niceNumber(max / count)
  const ticks: number[] = []
  for (let v = 0; v <= max + step / 2; v += step) {
    ticks.push(Math.round(v))
  }
  return ticks
}

function pickXTicks(n: number, want: number): number[] {
  if (n <= 0) return []
  if (n <= want) return Array.from({ length: n }, (_, i) => i)
  const step = (n - 1) / (want - 1)
  return Array.from({ length: want }, (_, i) => Math.round(i * step))
}
```

**Step 2: Run build and lint**

```bash
cd frontend && pnpm run build && pnpm run lint
```
Expected: both pass.

**Step 3: Commit**

```bash
cd /Users/ios.husari.orozco/Documents/projects/hube && git add frontend/src/components/molecules/AreaChart.tsx && git commit -m "feat(frontend): add AreaChart SVG component"
```

---

## Task 7: Integrate charts into `FinanceWidgets`

**Files:**
- Modify: `frontend/src/components/organisms/FinanceWidgets.tsx`

Add the chart grid between the stat cards and "Recent transactions", extend the skeleton, and wire the aggregations.

**Step 1: Replace the file with the complete updated version**

Write the complete `frontend/src/components/organisms/FinanceWidgets.tsx`:
```tsx
import { useMemo } from 'react'
import { useFinanceSummary } from '../../hooks/useFinance'
import {
  aggregateExpensesByCategory,
  aggregateExpensesByDay,
} from '../../utils/finance'
import { ChartCard } from '../molecules/ChartCard'
import { DonutChart } from '../molecules/DonutChart'
import { AreaChart } from '../molecules/AreaChart'

function fmt(amount: number) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatShortDate(iso: string) {
  return new Date(iso + 'T00:00:00Z').toLocaleDateString('en-US', {
    month: 'short',
    day: '2-digit',
    timeZone: 'UTC',
  })
}

export function FinanceWidgets() {
  const { data, isLoading, isError } = useFinanceSummary()

  const byCategory = useMemo(
    () => (data ? aggregateExpensesByCategory(data.recent_transactions) : []),
    [data],
  )
  const byDay = useMemo(
    () => (data ? aggregateExpensesByDay(data.recent_transactions) : []),
    [data],
  )

  if (isLoading) return <FinanceSkeleton />

  if (isError || !data || data.configured === false) {
    return (
      <div className="bg-surface-elevated border border-border rounded-xl px-5 py-4 text-sm text-text-muted">
        Money Monkey not connected.{' '}
        <span className="text-text-secondary">
          Set MONKEYAPI_URL and MONKEYAPI_KEY to enable.
        </span>
      </div>
    )
  }

  const savingsRate =
    data.month_income > 0
      ? Math.round(
          ((data.month_income - data.month_expenses) / data.month_income) * 100,
        )
      : 0

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <FinanceCard
          label="Balance"
          value={fmt(data.balance)}
          sub={data.balance >= 0 ? 'positive' : 'negative'}
          accent={data.balance >= 0 ? 'emerald' : 'red'}
        />
        <FinanceCard
          label="Income this month"
          value={fmt(data.month_income)}
          accent="indigo"
        />
        <FinanceCard
          label="Expenses this month"
          value={fmt(data.month_expenses)}
          sub={`${savingsRate}% savings rate`}
          accent="amber"
        />
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <ChartCard
          title="Expenses by category"
          empty={byCategory.length === 0 ? 'No expenses in this period' : undefined}
        >
          <DonutChart data={byCategory} formatValue={fmt} centerLabel="Expenses" />
        </ChartCard>
        <ChartCard
          title="Spending over time"
          empty={byDay.length === 0 ? 'No expenses in this period' : undefined}
        >
          <AreaChart
            data={byDay}
            formatValue={fmt}
            formatDate={formatShortDate}
          />
        </ChartCard>
      </div>

      {data.recent_transactions?.length > 0 && (
        <div className="bg-surface-elevated border border-border rounded-xl overflow-hidden">
          <div className="px-5 py-3 border-b border-border">
            <span className="text-xs font-medium text-text-secondary uppercase tracking-wider">
              Recent transactions
            </span>
          </div>
          <ul className="divide-y divide-border">
            {data.recent_transactions.map(tx => (
              <li key={tx.id} className="flex items-center justify-between px-5 py-3">
                <div className="flex items-center gap-3">
                  <span
                    className={`w-1.5 h-1.5 rounded-full ${
                      tx.type === 'income' ? 'bg-emerald-400' : 'bg-red-400'
                    }`}
                  />
                  <div>
                    <p className="text-sm text-text-primary">
                      {tx.description || (tx.category !== 'All' ? tx.category : 'Expense')}
                    </p>
                    <p className="text-xs text-text-muted">
                      {tx.category !== 'All' ? tx.category : ''}
                      {tx.category !== 'All' ? ' · ' : ''}
                      {tx.date.slice(0, 10)}
                    </p>
                  </div>
                </div>
                <span
                  className={`text-sm font-medium tabular-nums ${
                    tx.type === 'income' ? 'text-emerald-400' : 'text-red-400'
                  }`}
                >
                  {tx.type === 'income' ? '+' : '-'}
                  {fmt(tx.amount)}
                </span>
              </li>
            ))}
          </ul>
          <div className="px-5 py-3 border-t border-border">
            <a
              href="https://money-monkey-pwa.vercel.app"
              target="_blank"
              rel="noreferrer"
              className="text-xs text-(--color-accent) hover:text-(--color-accent-hover) transition-colors"
            >
              Open Money Monkey →
            </a>
          </div>
        </div>
      )}
    </div>
  )
}

function FinanceCard({
  label,
  value,
  sub,
  accent = 'indigo',
}: {
  label: string
  value: string
  sub?: string
  accent?: 'emerald' | 'red' | 'indigo' | 'amber'
}) {
  const colors = {
    emerald: 'text-emerald-400',
    red: 'text-red-400',
    indigo: 'text-indigo-400',
    amber: 'text-amber-400',
  }
  return (
    <div className="bg-surface-elevated border border-border rounded-xl px-5 py-4">
      <p className="text-xs text-text-muted mb-1">{label}</p>
      <p className={`text-2xl font-semibold tabular-nums ${colors[accent]}`}>
        {value}
      </p>
      {sub && <p className="text-xs text-text-muted mt-1">{sub}</p>}
    </div>
  )
}

function FinanceSkeleton() {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 animate-pulse">
        {[0, 1, 2].map(i => (
          <div
            key={i}
            className="bg-surface-elevated border border-border rounded-xl h-20"
          />
        ))}
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 animate-pulse">
        <div className="bg-surface-elevated border border-border rounded-xl h-64" />
        <div className="bg-surface-elevated border border-border rounded-xl h-64" />
      </div>
    </div>
  )
}
```

**Step 2: Run build, lint, and tests**

```bash
cd frontend && pnpm run build && pnpm run lint && pnpm test
```
Expected: build passes, lint passes, all 6 tests pass.

If `pnpm run build` reports a Tailwind class issue (e.g. `text-(--color-accent)` not found in the JIT/AOT scan because it appears only inside a template-string class concatenation), it means the literal string isn't present at build time. All `text-(--color-accent)` uses in this plan are full literal strings in the source, so Tailwind's scanner should pick them up. If it does not, the implementer moves the affected `className` to a constant in module scope or uses the equivalent `text-[var(--color-accent)]` arbitrary value. (The codebase already uses `text-(--color-accent)` and `bg-(--color-accent)` in many files — this pattern is proven to work.)

**Step 3: Commit**

```bash
cd /Users/ios.husari.orozco/Documents/projects/hube && git add frontend/src/components/organisms/FinanceWidgets.tsx && git commit -m "feat(frontend): show category and trend charts in FinanceWidgets"
```

---

## Task 8: Final verification

**Files:** none (verification only)

**Step 1: Backend still green**

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```
Expected: all pass. (The frontend work touched no backend code; this is a safety net.)

**Step 2: Frontend full check**

```bash
cd frontend && pnpm test && pnpm run lint && pnpm run build
```
Expected: 6 tests pass, lint passes, build passes.

**Step 3: Visual smoke check (manual)**

- Rebuild the `web` container in Tailscale if you want to see it live:
  ```bash
  docker compose -f docker-compose.tailscale.yml --env-file .env.tailscale build web
  docker compose -f docker-compose.tailscale.yml --env-file .env.tailscale up -d
  ```
- Hard-reload the browser (`Cmd+Shift+R`).
- The Dashboard Finance section should now show the stat cards, then two charts side by side on `sm+` widths (stacked on mobile), then the recent transactions list. The donut shows expenses grouped by category with the total in the centre and a legend. The area chart shows daily expenses with a hover tooltip. If Money Monkey has no expenses, the chart cards show "No expenses in this period".

**Step 4: Commit any final tweaks (if needed)**

If you made any small visual tweaks (spacing, colours) during the smoke check, commit them now:
```bash
cd /Users/ios.husari.orozco/Documents/projects/hube && git add -A && git commit -m "style(frontend): finance chart visual tweaks"
```

If no tweaks were needed, skip this step — the work is already committed per-task.

---

## Done

The Dashboard's Finance section now shows expenses by category (donut) and spending over time (area) using the data the API already returns. Zero backend changes, zero new runtime dependencies (only `vitest` as a devDep for the aggregation tests).
