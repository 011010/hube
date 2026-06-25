import type { ReactNode } from 'react'

interface StatCardProps {
  label: string
  value: string | number
  subText?: string
  trend?: 'up' | 'down' | 'neutral'
  accent?: 'indigo' | 'emerald' | 'amber' | 'red' | 'sky'
}

const ACCENT_STYLES: Record<string, string> = {
  indigo: 'text-accent',
  emerald: 'text-emerald-400',
  amber: 'text-amber-400',
  red: 'text-red-400',
  sky: 'text-sky-400',
}

const TREND_ICONS: Record<string, ReactNode> = {
  up: '↑',
  down: '↓',
  neutral: '→',
}

export function StatCard({ label, value, subText, trend, accent = 'indigo' }: StatCardProps) {
  return (
    <div className="bg-surface-elevated border border-border rounded-xl p-5">
      <p className="text-sm text-text-muted">{label}</p>
      <p className={`text-3xl font-semibold mt-1 tabular-nums ${ACCENT_STYLES[accent]}`}>
        {trend && (
          <span className={`inline-block mr-1 text-base ${
            trend === 'up' ? 'text-emerald-400' : trend === 'down' ? 'text-red-400' : 'text-text-muted'
          }`}>
            {TREND_ICONS[trend]}
          </span>
        )}
        {value}
      </p>
      {subText && <p className="text-xs text-text-muted mt-1">{subText}</p>}
    </div>
  )
}
