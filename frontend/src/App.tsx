import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ThemeProvider } from './contexts/ThemeContext'
import { Sidebar } from './components/organisms/Sidebar'
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
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <BrowserRouter>
          <div className="flex h-screen overflow-hidden bg-gray-50 dark:bg-gray-950">
            <Sidebar />
            <main className="flex-1 overflow-y-auto">
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
        </BrowserRouter>
      </ThemeProvider>
    </QueryClientProvider>
  )
}
