import { NavLink } from 'react-router-dom'
import { HubeLogo } from '../atoms/HubeLogo'

const nav = [
  { to: '/', label: 'Dashboard', icon: '⊞' },
  { to: '/launcher', label: 'Apps', icon: '◈' },
  { to: '/tasks', label: 'Tasks', icon: '✓' },
  { to: '/calendar', label: 'Calendar', icon: '◷' },
  { to: '/notes', label: 'Notes', icon: '✎' },
  { to: '/projects', label: 'Projects', icon: '◫' },
  { to: '/ai', label: 'IA', icon: '✦' },
  { to: '/wishlist', label: 'Wishlist', icon: '♡' },
  { to: '/network', label: 'Network', icon: '⊛' },
]

const navItemClass = ({ isActive }: { isActive: boolean }) =>
  `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
    isActive
      ? 'bg-(--color-accent) text-white'
      : 'text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white'
  }`

export function Sidebar() {
  return (
    <aside className="w-56 shrink-0 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 flex flex-col">
      <div className="px-5 py-5 border-b border-gray-200 dark:border-gray-800">
        <div className="flex items-center gap-2.5">
          <HubeLogo size={20} />
          <span className="text-lg font-semibold tracking-tight text-gray-900 dark:text-white">hube</span>
        </div>
      </div>
      <nav className="flex-1 px-3 py-4 space-y-0.5">
        {nav.map(({ to, label, icon }) => (
          <NavLink key={to} to={to} end={to === '/'} className={navItemClass}>
            <span className="text-base">{icon}</span>
            {label}
          </NavLink>
        ))}
      </nav>
      <div className="px-3 py-3 border-t border-gray-200 dark:border-gray-800 space-y-0.5">
        <NavLink to="/docs" className={navItemClass}>
          <span className="text-base">☰</span>
          Docs
        </NavLink>
        <NavLink to="/settings" className={navItemClass}>
          <span className="text-base">⚙</span>
          Settings
        </NavLink>
      </div>
    </aside>
  )
}
