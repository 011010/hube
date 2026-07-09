import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard,
  Grid3X3,
  ListChecks,
  Calendar,
  FileText,
  FolderKanban,
  Brain,
  Heart,
  Network,
  BookOpen,
  Settings,
} from 'lucide-react'
import { HubeLogo } from '../atoms/HubeLogo'
import type { LucideIcon } from 'lucide-react'

interface NavItem {
  to: string
  label: string
  icon: LucideIcon
}

const mainNav: NavItem[] = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/launcher', label: 'Apps', icon: Grid3X3 },
  { to: '/tasks', label: 'Tasks', icon: ListChecks },
  { to: '/calendar', label: 'Calendar', icon: Calendar },
  { to: '/notes', label: 'Notes', icon: FileText },
  { to: '/projects', label: 'Projects', icon: FolderKanban },
  { to: '/ai', label: 'IA', icon: Brain },
  { to: '/wishlist', label: 'Wishlist', icon: Heart },
  { to: '/network', label: 'Diagrams', icon: Network },
]

const bottomNav: NavItem[] = [
  { to: '/docs', label: 'Docs', icon: BookOpen },
  { to: '/settings', label: 'Settings', icon: Settings },
]

const navItemClass = ({ isActive }: { isActive: boolean }) =>
  `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all duration-150 ${
    isActive
      ? 'bg-(--color-accent) text-white font-medium shadow-sm'
      : 'text-text-secondary hover:bg-surface-elevated hover:text-text-primary'
  }`

interface SidebarProps {
  isOpen?: boolean
  onClose?: () => void
}

export function Sidebar({ isOpen = false, onClose }: SidebarProps) {
  return (
    <>
      {/* Mobile backdrop */}
      {isOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 md:hidden"
          onClick={onClose}
          aria-hidden="true"
        />
      )}

      <aside
        className={`fixed inset-y-0 left-0 z-50 w-56 shrink-0 bg-surface-base border-r border-border flex flex-col transition-transform duration-200 ease-out md:static md:z-auto md:translate-x-0 ${
          isOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        {/* Safe-area spacer for mobile drawer (avoids status bar overlap) */}
        <div className="h-[env(safe-area-inset-top)] md:hidden" />

        {/* Logo header */}
        <div className="px-5 py-5 border-b border-border">
          <div className="flex items-center gap-2.5">
            <HubeLogo size={20} />
            <span className="text-lg font-semibold tracking-tight text-text-primary">hube</span>
          </div>
        </div>

        {/* Main navigation */}
        <nav className="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
          {mainNav.map(({ to, label, icon: Icon }) => (
            <NavLink key={to} to={to} end={to === '/'} className={navItemClass} onClick={onClose}>
              <Icon size={18} className="shrink-0" />
              <span className="truncate">{label}</span>
            </NavLink>
          ))}
        </nav>

        {/* Bottom navigation (collapse-ready wrapper) */}
        <div className="px-3 py-3 border-t border-border space-y-0.5">
          {bottomNav.map(({ to, label, icon: Icon }) => (
            <NavLink key={to} to={to} className={navItemClass} onClick={onClose}>
              <Icon size={18} className="shrink-0" />
              <span className="truncate">{label}</span>
            </NavLink>
          ))}
        </div>
      </aside>
    </>
  )
}
