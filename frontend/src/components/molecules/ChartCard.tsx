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
