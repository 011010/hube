import { useTasks } from '../../hooks/useTasks'
import { useApps } from '../../hooks/useApps'
import { Badge } from '../../components/atoms/Badge'
import { FinanceWidgets } from '../../components/organisms/FinanceWidgets'

export function DashboardPage() {
  const { data: tasks = [] } = useTasks()
  const { data: apps = [] } = useApps()

  const pending = tasks.filter(t => t.status !== 'done')
  const done = tasks.filter(t => t.status === 'done')

  return (
    <div className="p-8 space-y-8">
      <h1 className="text-2xl font-semibold text-white">Dashboard</h1>

      <div className="grid grid-cols-3 gap-4">
        <StatCard label="Pending tasks" value={pending.length} />
        <StatCard label="Completed" value={done.length} />
        <StatCard label="Apps" value={apps.length} />
      </div>

      <section>
        <h2 className="text-sm font-medium text-gray-400 uppercase tracking-wider mb-3">Finance</h2>
        <FinanceWidgets />
      </section>

      <section>
        <h2 className="text-sm font-medium text-gray-400 uppercase tracking-wider mb-3">Recent tasks</h2>
        <div className="space-y-2">
          {pending.slice(0, 5).map(task => (
            <div key={task.id} className="flex items-center gap-3 bg-gray-900 rounded-lg px-4 py-3">
              <Badge
                label={task.priority}
                variant={task.priority === 'high' ? 'danger' : task.priority === 'medium' ? 'warning' : 'default'}
              />
              <span className="text-sm text-gray-200">{task.title}</span>
            </div>
          ))}
          {pending.length === 0 && (
            <p className="text-sm text-gray-500">No pending tasks.</p>
          )}
        </div>
      </section>

      <section>
        <h2 className="text-sm font-medium text-gray-400 uppercase tracking-wider mb-3">Quick access</h2>
        <div className="flex gap-3 flex-wrap">
          {apps.map(app => (
            <a
              key={app.id}
              href={app.url}
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-2 bg-gray-900 hover:bg-gray-800 border border-gray-700 rounded-xl px-4 py-3 transition-colors"
            >
              <span className="text-xl">{app.icon}</span>
              <span className="text-sm font-medium text-white">{app.name}</span>
            </a>
          ))}
        </div>
      </section>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="bg-gray-900 border border-gray-800 rounded-xl p-5">
      <p className="text-sm text-gray-400">{label}</p>
      <p className="text-3xl font-semibold text-white mt-1">{value}</p>
    </div>
  )
}
