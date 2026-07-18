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
