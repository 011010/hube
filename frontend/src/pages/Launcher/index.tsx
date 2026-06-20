import { useApps } from '../../hooks/useApps'
import { safeHref } from '../../utils/url'

export function LauncherPage() {
  const { data: apps = [], isLoading } = useApps()

  return (
    <div className="p-8 space-y-6">
      <h1 className="text-2xl font-semibold text-white">Apps</h1>
      {isLoading && <p className="text-gray-400 text-sm">Loading...</p>}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
        {apps.map(app => (
          <a
            key={app.id}
            href={safeHref(app.url)}
            target="_blank"
            rel="noreferrer"
            className="group flex flex-col items-center gap-3 bg-gray-900 hover:bg-gray-800 border border-gray-800 hover:border-gray-600 rounded-2xl p-6 transition-all"
          >
            <span
              className="w-14 h-14 flex items-center justify-center rounded-2xl text-3xl"
              style={{ backgroundColor: `${app.color}20` }}
            >
              {app.icon}
            </span>
            <div className="text-center">
              <p className="text-sm font-semibold text-white group-hover:text-indigo-300 transition-colors">
                {app.name}
              </p>
              {app.description && (
                <p className="text-xs text-gray-500 mt-0.5">{app.description}</p>
              )}
            </div>
          </a>
        ))}
      </div>
    </div>
  )
}
