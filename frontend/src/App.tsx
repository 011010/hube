import { useState } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Menu } from 'lucide-react'
import { ThemeProvider } from './contexts/ThemeContext'
import { Sidebar } from './components/organisms/Sidebar'
import { HubeLogo } from './components/atoms/HubeLogo'
import { DashboardPage } from './pages/Dashboard'
import { LauncherPage } from './pages/Launcher'
import { TasksPage } from './pages/Tasks'
import { CalendarPage } from './pages/Calendar'
import { NotesPage } from './pages/Notes'
import { ProjectsPage } from './pages/Projects'
import { ProjectDetailPage } from './pages/Projects/Detail'
import { AIPage } from './pages/AI'
import { SettingsPage } from './pages/Settings'
import { WishlistPage } from './pages/Wishlist'
import { NetworkPage } from './pages/Network'
import { DocsPage } from './pages/Docs'
import { ErrorBoundary } from './components/ErrorBoundary'

const queryClient = new QueryClient()

export default function App() {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false)

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <BrowserRouter>
          <div className="flex h-screen overflow-hidden bg-surface-base">
            <Sidebar isOpen={isSidebarOpen} onClose={() => setIsSidebarOpen(false)} />
            <div className="flex flex-1 flex-col overflow-hidden">
              {/* Mobile top bar */}
              <div className="pt-safe flex items-center gap-3 px-4 py-3 border-b border-border md:hidden">
                <button
                  type="button"
                  onClick={() => setIsSidebarOpen(true)}
                  className="text-text-secondary hover:text-text-primary transition-colors"
                  aria-label="Open menu"
                >
                  <Menu size={22} />
                </button>
                <div className="flex items-center gap-2.5">
                  <HubeLogo size={20} />
                  <span className="text-lg font-semibold tracking-tight text-text-primary">hube</span>
                </div>
              </div>
              <main className="flex-1 overflow-y-auto animate-fadeIn">
              <ErrorBoundary>
              <Routes>
                <Route path="/" element={<DashboardPage />} />
                <Route path="/launcher" element={<LauncherPage />} />
                <Route path="/tasks" element={<TasksPage />} />
                <Route path="/calendar" element={<CalendarPage />} />
                <Route path="/notes" element={<NotesPage />} />
                <Route path="/projects" element={<ProjectsPage />} />
                <Route path="/projects/:id" element={<ProjectDetailPage />} />
                <Route path="/ai" element={<AIPage />} />
                <Route path="/wishlist" element={<WishlistPage />} />
                <Route path="/network" element={<NetworkPage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/docs" element={<DocsPage />} />
              </Routes>
              </ErrorBoundary>
              </main>
            </div>
          </div>
        </BrowserRouter>
      </ThemeProvider>
    </QueryClientProvider>
  )
}
