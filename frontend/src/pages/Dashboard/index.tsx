import { ArrowUpRight } from 'lucide-react'
import { useTasks } from '../../hooks/useTasks'
import { safeHref } from '../../utils/url'
import { useApps } from '../../hooks/useApps'
import { Badge } from '../../components/atoms/Badge'
import { FinanceWidgets } from '../../components/organisms/FinanceWidgets'
import { CardTrackerWidget } from '../../components/organisms/CardTrackerWidget'
import { StatCard } from '../../components/molecules/StatCard'
import { Card } from '../../components/molecules/Card'
import { PageHeader } from '../../components/molecules/PageHeader'

export function DashboardPage() {
  const { data: tasks = [] } = useTasks()
  const { data: apps = [] } = useApps()

  const pending = tasks.filter(t => t.status !== 'done')
  const done = tasks.filter(t => t.status === 'done')

  return (
    <div className="p-8 space-y-8">
      <PageHeader title="Dashboard" />

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard label="Pending tasks" value={pending.length} />
        <StatCard label="Completed" value={done.length} />
        <StatCard label="Apps" value={apps.length} />
      </div>

      <section>
        <h2 className="text-sm font-medium text-text-muted uppercase tracking-wider mb-3">Finance</h2>
        <FinanceWidgets />
      </section>

      <section>
        <h2 className="text-sm font-medium text-text-muted uppercase tracking-wider mb-3">Cards</h2>
        <CardTrackerWidget />
      </section>

      <section>
        <h2 className="text-sm font-medium text-text-muted uppercase tracking-wider mb-3">Recent tasks</h2>
        <div className="space-y-2">
          {pending.slice(0, 5).map(task => (
            <Card key={task.id} padding="compact" className="flex items-center gap-3">
              <Badge
                label={task.priority}
                variant={task.priority === 'high' ? 'danger' : task.priority === 'medium' ? 'warning' : 'default'}
              />
              <span className="text-sm text-text-secondary">{task.title}</span>
            </Card>
          ))}
          {pending.length === 0 && (
            <p className="text-sm text-text-muted">No pending tasks.</p>
          )}
        </div>
      </section>

      <section>
        <h2 className="text-sm font-medium text-text-muted uppercase tracking-wider mb-3">Quick access</h2>
        <div className="flex gap-3 flex-wrap">
          {apps.map(app => (
            <a
              key={app.id}
              href={safeHref(app.url)}
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-2 bg-surface-elevated hover:bg-surface-card border border-border rounded-xl px-4 py-3 transition-colors group"
            >
              <span className="text-xl">{app.icon}</span>
              <span className="text-sm font-medium text-text-primary">{app.name}</span>
              <ArrowUpRight size={14} className="text-text-muted opacity-0 group-hover:opacity-100 transition-opacity shrink-0" />
            </a>
          ))}
        </div>
      </section>
    </div>
  )
}
