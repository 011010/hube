import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useProjects, useCreateProject, useDeleteProject } from '../../hooks/useProjects'
import type { Project, ProjectStatus } from '../../types'

const STATUS_LABEL: Record<ProjectStatus, string> = {
  planning: 'Planning',
  active: 'Active',
  completed: 'Completed',
  on_hold: 'On Hold',
}

const STATUS_COLOR: Record<ProjectStatus, string> = {
  planning: 'bg-gray-700 text-gray-300',
  active: 'bg-indigo-900 text-indigo-300',
  completed: 'bg-green-900 text-green-300',
  on_hold: 'bg-amber-900 text-amber-300',
}

function progress(p: Project) {
  if (p.task_count === 0) return 0
  return Math.round((p.completed_count / p.task_count) * 100)
}

const COLORS = ['#6366f1', '#8b5cf6', '#ec4899', '#f97316', '#14b8a6', '#3b82f6', '#ef4444', '#22c55e']

function NewProjectModal({ onClose }: { onClose: () => void }) {
  const create = useCreateProject()
  const [form, setForm] = useState({ name: '', description: '', color: COLORS[0], status: 'planning' as ProjectStatus })

  function submit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.name.trim()) return
    create.mutate(form, { onSuccess: onClose })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="bg-gray-900 border border-gray-700 rounded-xl w-full max-w-md p-6 shadow-xl">
        <h2 className="text-lg font-semibold text-white mb-5">New project</h2>
        <form onSubmit={submit} className="space-y-4">
          <div>
            <label className="block text-xs text-gray-400 mb-1">Name</label>
            <input
              autoFocus
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
              value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              placeholder="Project name"
            />
          </div>
          <div>
            <label className="block text-xs text-gray-400 mb-1">Description</label>
            <textarea
              rows={2}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500 resize-none"
              value={form.description}
              onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              placeholder="What is this project about?"
            />
          </div>
          <div>
            <label className="block text-xs text-gray-400 mb-2">Color</label>
            <div className="flex gap-2">
              {COLORS.map(c => (
                <button
                  key={c}
                  type="button"
                  onClick={() => setForm(f => ({ ...f, color: c }))}
                  className={`w-7 h-7 rounded-full border-2 transition-transform ${form.color === c ? 'border-white scale-110' : 'border-transparent'}`}
                  style={{ backgroundColor: c }}
                />
              ))}
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm text-gray-400 hover:text-white transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!form.name.trim() || create.isPending}
              className="px-4 py-2 text-sm bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white rounded-lg transition-colors"
            >
              Create
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export function ProjectsPage() {
  const { data: projects = [], isLoading } = useProjects()
  const deleteProject = useDeleteProject()
  const navigate = useNavigate()
  const [showNew, setShowNew] = useState(false)

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold text-white">Projects</h1>
        <button
          onClick={() => setShowNew(true)}
          className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-sm rounded-lg transition-colors"
        >
          + New project
        </button>
      </div>

      {isLoading && <p className="text-gray-500 text-sm">Loading…</p>}

      {!isLoading && projects.length === 0 && (
        <div className="text-center py-20 text-gray-600">
          <p className="text-4xl mb-3">◻</p>
          <p className="text-sm">No projects yet. Create your first one.</p>
        </div>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {projects.map(p => {
          const pct = progress(p)
          return (
            <div
              key={p.id}
              onClick={() => navigate(`/projects/${p.id}`)}
              className="bg-gray-900 border border-gray-800 rounded-xl p-5 cursor-pointer hover:border-gray-600 transition-colors group"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full shrink-0" style={{ backgroundColor: p.color }} />
                  <h3 className="font-medium text-white text-sm leading-snug">{p.name}</h3>
                </div>
                <span className={`text-xs px-2 py-0.5 rounded-full shrink-0 ml-2 ${STATUS_COLOR[p.status]}`}>
                  {STATUS_LABEL[p.status]}
                </span>
              </div>

              {p.description && (
                <p className="text-xs text-gray-500 mb-3 line-clamp-2">{p.description}</p>
              )}

              <div className="mb-1 flex items-center justify-between text-xs text-gray-500">
                <span>{p.completed_count}/{p.task_count} tasks</span>
                <span>{pct}%</span>
              </div>
              <div className="h-1.5 bg-gray-800 rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full transition-all"
                  style={{ width: `${pct}%`, backgroundColor: p.color }}
                />
              </div>

              {p.due_date && (
                <p className="text-xs text-gray-600 mt-3">Due {p.due_date}</p>
              )}

              <button
                onClick={e => { e.stopPropagation(); deleteProject.mutate(p.id) }}
                className="mt-3 text-xs text-gray-700 hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100"
              >
                Delete
              </button>
            </div>
          )
        })}
      </div>

      {showNew && <NewProjectModal onClose={() => setShowNew(false)} />}
    </div>
  )
}
