import { AppWindow } from 'lucide-react'
import { useApps } from '../../hooks/useApps'
import { safeHref } from '../../utils/url'

export function LauncherPage() {
  const { data: apps = [], isLoading } = useApps()

  return (
    <div className="p-8 space-y-6">
      <h1 className="text-2xl font-semibold text-text-primary">Apps</h1>
      {isLoading && <p className="text-text-muted text-sm">Loading...</p>}
      {apps.length === 0 && !isLoading && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <AppWindow size={44} className="text-text-muted opacity-40 mb-4" />
          <p className="text-sm text-text-secondary">No apps configured yet.</p>
        </div>
      )}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
        {apps.map(app => (
          <a
            key={app.id}
            href={safeHref(app.url)}
            target="_blank"
            rel="noreferrer"
            className="group flex flex-col items-center gap-3 bg-surface-elevated hover:bg-surface-card border border-border hover:border-border-subtle rounded-2xl p-6 transition-all"
          >
            <span
              className="w-14 h-14 flex items-center justify-center rounded-2xl text-3xl"
              style={{ backgroundColor: `${app.color}20` }}
            >
              {app.icon}
            </span>
            <div className="text-center">
              <p className="text-sm font-semibold text-text-primary group-hover:text-(--color-accent) transition-colors">
                {app.name}
              </p>
              {app.description && (
                <p className="text-xs text-text-muted mt-0.5">{app.description}</p>
              )}
            </div>
          </a>
        ))}
      </div>
    </div>
  )
}
