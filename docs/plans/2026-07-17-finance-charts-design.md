# Finance charts on Dashboard — design

**Date:** 2026-07-17
**Status:** Approved

## Goal

Show two charts in the Dashboard's Finance section, driven by data the Money
Monkey integration already returns:

1. **Expenses by category** — donut with total at the centre and a legend
   below (category, amount, share).
2. **Spending over time** — area chart of daily expenses across the window
   covered by `recent_transactions`.

Both charts live in the existing Finance section, between the three stat cards
and the "Recent transactions" list. No new backend endpoint, no new runtime
dependency.

## Non-goals

- Charts for other sections (Cards, Tasks, Calendar). Easy to add later with
  the same building blocks.
- Keyboard-focusable chart segments (full a11y). Out of scope for a personal
  hub; the textual legend conveys the data.
- Categorising the `"All"` catch-all. It is excluded from the chart and the
  total so it does not pollute real categories.
- Server-side aggregation. The aggregation is cheap and the dataset is small
  (one window of recent transactions).

## Decisions

| Decision | Choice | Why |
|---|---|---|
| Scope | Finance only | Matches the explicit example; ship the most valuable one first. |
| Chart types | Donut (category) + area (time) | Covers the two questions the example implies: *where does the money go?* and *is it trending up?* |
| Library | Hand-rolled SVG, zero new runtime deps | Two shapes, strong existing design system (Tailwind 4 + CSS vars), no need to bring in a library to maintain a theme. Components stay small and swappable. |
| Backend | Unchanged | `GET /api/v1/finance/summary` already returns `recent_transactions[]` with `category`, `amount`, `type`, `date`. |
| Tests | `vitest` for the two aggregation functions only | Frontend has no test infra today; pure functions are the only thing that can break silently. SVG components are verified visually. |

## Architecture

### New files

| File | Purpose |
|---|---|
| `frontend/src/utils/finance.ts` | Pure aggregation: `aggregateExpensesByCategory`, `aggregateExpensesByDay`. Also exports the chart palette. |
| `frontend/src/utils/finance.test.ts` | Vitest unit tests for both aggregations. |
| `frontend/src/components/molecules/ChartCard.tsx` | Card wrapper matching the "Recent transactions" card (surface-elevated + border + title). |
| `frontend/src/components/molecules/DonutChart.tsx` | SVG donut + legend. No external deps. |
| `frontend/src/components/molecules/AreaChart.tsx` | SVG area chart with axes, gridlines and hover tooltip. No external deps. |
| `frontend/vitest.config.ts` | Minimal Vitest config (no jsdom — pure functions only). |

### Modified files

| File | Change |
|---|---|
| `frontend/src/components/organisms/FinanceWidgets.tsx` | Insert a `grid grid-cols-1 sm:grid-cols-2 gap-4` of the two charts between the stat cards (line 28) and "Recent transactions" (line 49). Extend `FinanceSkeleton` with two pulsing rectangles in the chart grid. |
| `frontend/package.json` | Add `vitest` (and `@vitest/coverage-v8` not needed) as devDep. Add `"test": "vitest run"` script. |

### Component responsibilities

- **`ChartCard`** — presentational, no state. Props: `title: string`, `children: ReactNode`, `empty?: string`. Styling identical to the existing transactions card.
- **`DonutChart`** — Props: `data: { label: string; value: number; color: string }[]`, `formatValue: (n: number) => string`, `centerLabel?: string`. SVG `viewBox` 200×200, donut thickness ~22% of the radius, legend always below the donut. Hover: other segments dim to 40% via CSS `:hover` on the segment + sibling selector. Pure render, no animation library.
- **`AreaChart`** — Props: `data: { date: string; value: number }[]`, `formatValue`, `formatDate`, `height?: number` (default 200). SVG `viewBox`, accent fill at ~18% with a solid line on top, 3-4 horizontal gridlines, ~5 date ticks on the x-axis. Hover: vertical guide + tooltip (`surface-elevated` + `border`). Scales with `width="100%"`.
- **Aggregations** — pure, in `utils/finance.ts`. No React, no hooks. Easy to unit-test and to reuse if other surfaces need the same breakdowns.

## Data flow

```
useFinanceSummary()         (existing)
  └─ data.recent_transactions
       ├─ aggregateExpensesByCategory()  ──▶ DonutChart
       └─ aggregateExpensesByDay()       ──▶ AreaChart
```

No new API call. The hook already has a 5-minute `staleTime`; the charts
re-render when the hook refetches.

## Aggregation contracts

```ts
type CategorySlice = { category: string; total: number; share: number /* 0..1 */ }
type DayPoint     = { date: string /* YYYY-MM-DD */; total: number }

function aggregateExpensesByCategory(
  transactions: RecentTransaction[],
  palette: readonly string[],
): { label: string; value: number; color: string }[]
```

- Filter `type === "expense"`.
- Drop entries with `category === "All"` (or `""`).
- Group by `category`, sum `amount`.
- Sort desc by `total`.
- Assign `color` by sorted index modulo the palette.
- Return shape ready for `DonutChart` (already in label/value form; the chart
  doesn't need `share` — the legend computes it from `value / total`).

```ts
function aggregateExpensesByDay(
  transactions: RecentTransaction[],
): DayPoint[]
```

- Filter `type === "expense"` (drop income and `"All"`).
- Group by `date.slice(0, 10)`, sum `amount`.
- Determine range as `[min(date), max(date)]`.
- Fill missing days with `total: 0`.
- Sort asc by `date`.
- Return one point per day in the range.

## Visual

### Donut

- Diameter 200, stroke 44, centred. Segments use `stroke-dasharray` on a
  single circle (one path per segment) — the simplest, most robust approach.
  `stroke-linecap="butt"` so segments butt cleanly.
- Centre text: large `total` formatted (e.g. `$8,040`) and a small `Expenses`
  label below it.
- Legend below the donut: vertical list, one row per category, dot (8px) in
  the segment colour, category name, right-aligned amount and share %.
- Hover: a `:hover` rule on each segment path that sets opacity on its
  siblings (using `group-hover` and `peer`-style attributes) — pure CSS.
- Empty state: a single line "No expenses in this period" in `text-text-muted`.

### Area

- Width 100%, height 200, padding 24/16/28/36 (top/right/bottom/left) for
  axis labels.
- Y axis: 4 ticks, formatted currency, short (`$1k`, `$2k`...) using
  `Intl.NumberFormat` with `notation: "compact"`.
- X axis: 5 date ticks (`formatDate` returns e.g. `Jul 02`), rotated 0°.
- Gridlines: horizontal at each Y tick, `var(--border)` at 60% opacity, 1px.
- Area: `path` with `fill="var(--color-accent)" fill-opacity="0.18"`.
- Line: `path` with `stroke="var(--color-accent)" stroke-width="2" fill="none"`.
- Hover: a transparent wide `rect` captures pointer events; on `mousemove`,
  find the closest day, draw a vertical guide and a tooltip. Tooltip is a
  small absolutely-positioned div in `surface-elevated` + `border` + small
  text, follows the pointer with a small offset, clamped inside the chart.
- Empty state: same as donut.

### Palette (chart)

```ts
export const CHART_PALETTE = [
  '#6366f1', // indigo
  '#10b981', // emerald
  '#f59e0b', // amber
  '#f43f5e', // rose
  '#0ea5e9', // sky
  '#8b5cf6', // violet
  '#d946ef', // fuchsia
  '#14b8a6', // teal
] as const
```

Chosen for sufficient contrast on the dark theme. Cycles if there are more
than 8 categories.

## States

| State | Behaviour |
|---|---|
| `data.configured === false` or error | Existing "Money Monkey not connected" block already replaces the whole widget. Charts are not rendered. |
| Loading | `FinanceSkeleton` extended: two pulsing rectangles (`h-48`) in the chart grid. |
| Connected, but `recent_transactions` has no expenses | Each `ChartCard` shows its title and a muted "No expenses in this period". |
| Connected, with data | Charts render. |

## Tests

`vitest` (devDep only). Config in `vitest.config.ts` — `environment: "node"`
(no jsdom needed), include `src/**/*.{test,spec}.{ts,tsx}`.

`utils/finance.test.ts` covers:

- `aggregateExpensesByCategory`:
  - groups expenses, drops income and `"All"`
  - sorts desc by `total`
  - assigns palette colours by sorted index
  - empty input → `[]`
- `aggregateExpensesByDay`:
  - groups by `date.slice(0, 10)`
  - fills missing days with `0` in the `[min..max]` range
  - sorts asc
  - single-day input → one point
  - empty input → `[]`

SVG components are not unit-tested (no jsdom, no RTL — the cost outweighs
the benefit for two static visual components). They are verified manually
against the design and the Tailscale deploy.

## Responsive and a11y

- The chart grid is `grid grid-cols-1 sm:grid-cols-2 gap-4` — stacked on
  mobile, side by side from `sm` up.
- Donut is always rendered with the legend below it (no left/right variant),
  which keeps it clean at any width and avoids responsive branching.
- SVGs use `viewBox` + `width="100%"` for scaling; the area chart has a
  fixed `height` (~200) and the donut a fixed 200×200 box.
- A11y (basic):
  - Each chart `<svg>` has `role="img"` and a concise `aria-label`
    summarising the top items, e.g.
    `aria-label="Expenses by category. Top: Card Payments $2,211, Debts $1,100, Healthcare $850"`.
  - The legend is a semantic `<ul>` of plain text — fully readable by screen
    readers without any special interaction.
  - The hover tooltip is an enhancement, not the source of truth.
- Keyboard a11y for chart segments is **out of scope** (see Non-goals).

## Files summary

**New**
- `frontend/src/utils/finance.ts`
- `frontend/src/utils/finance.test.ts`
- `frontend/src/components/molecules/ChartCard.tsx`
- `frontend/src/components/molecules/DonutChart.tsx`
- `frontend/src/components/molecules/AreaChart.tsx`
- `frontend/vitest.config.ts`

**Modified**
- `frontend/src/components/organisms/FinanceWidgets.tsx`
- `frontend/package.json` (vitest devDep + `test` script)

**0 backend changes. 0 API changes. 0 new runtime dependencies.**
