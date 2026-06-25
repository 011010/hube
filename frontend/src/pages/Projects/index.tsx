import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Trash2, FolderKanban } from 'lucide-react'
import { useProjects, useCreateProject, useDeleteProject } from '../../hooks/useProjects'
import { Button } from '../../components/atoms/Button'
import { Input } from '../../components/atoms/Input'
import { Textarea } from '../../components/atoms/Textarea'
import { Modal } from '../../components/atoms/Modal'
import { PageHeader } from '../../components/molecules/PageHeader'
import { EmptyState } from '../../components/molecules/EmptyState'
import type { Project, ProjectStatus } from '../../types'

const STATUS_LABEL: Record<ProjectStatus, string> = {
  planning: 'Planning',
  active: 'Active',
  completed: 'Completed',
  on_hold: 'On Hold',
}

const STATUS_COLOR: Record<ProjectStatus, string> = {
  planning: 'bg-surface-elevated text-text-muted',
  active: 'bg-(--color-accent)/10 text-(--color-accent)',
  completed: 'bg-emerald-900/30 text-emerald-300',
  on_hold: 'bg-amber-900/30 text-amber-300',
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
    <Modal open onClose={onClose} title="New project">
      <form onSubmit={submit} className="space-y-4">
        <Input
          label="Name"
          value={form.name}
          onChange={e => setForm(f => ({ ...f, name: (e.target as HTMLInputElement).value }))}
          placeholder="Project name"
          autoFocus
        />
        <Textarea
          label="Description"
          rows={2}
          value={form.description}
          onChange={e => setForm(f => ({ ...f, description: (e.target as HTMLTextAreaElement).value }))}
          placeholder="What is this project about?"
        />
        <div>
          <label className="block text-sm font-medium text-text-primary mb-2">Color</label>
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
          <Button variant="ghost" type="button" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={!form.name.trim() || create.isPending}>Create</Button>
        </div>
      </form>
    </Modal>
  )
}

export function ProjectsPage() {
  const { data: projects = [], isLoading } = useProjects()
  const deleteProject = useDeleteProject()
  const navigate = useNavigate()
  const [showNew, setShowNew] = useState(false)

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <PageHeader
        title="Projects"
        actions={
          <Button onClick={() => setShowNew(true)} icon={<Plus size={16} />}>New project</Button>
        }
      />

      {isLoading && <p className="text-text-muted text-sm">Loading…</p>}

      {!isLoading && projects.length === 0 && (
        <EmptyState
          icon={<FolderKanban size={44} />}
          title="No projects yet"
          description="Create your first project to organise tasks and track progress."
          action={
            <Button onClick={() => setShowNew(true)} icon={<Plus size={16} />}>New project</Button>
          }
        />
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {projects.map(p => {
          const pct = progress(p)
          return (
            <div
              key={p.id}
              onClick={() => navigate(`/projects/${p.id}`)}
              className="bg-surface-elevated border border-border rounded-xl p-5 cursor-pointer hover:border-border-subtle transition-colors group"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full shrink-0" style={{ backgroundColor: p.color }} />
                  <h3 className="font-medium text-text-primary text-sm leading-snug">{p.name}</h3>
                </div>
                <span className={`text-xs px-2 py-0.5 rounded-full shrink-0 ml-2 ${STATUS_COLOR[p.status]}`}>
                  {STATUS_LABEL[p.status]}
                </span>
              </div>

              {p.description && (
                <p className="text-xs text-text-muted mb-3 line-clamp-2">{p.description}</p>
              )}

              <div className="mb-1 flex items-center justify-between text-xs text-text-muted">
                <span>{p.completed_count}/{p.task_count} tasks</span>
                <span>{pct}%</span>
              </div>
              <div className="h-1.5 bg-surface-base rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full transition-all"
                  style={{ width: `${pct}%`, backgroundColor: p.color }}
                />
              </div>

              {p.due_date && (
                <p className="text-xs text-text-muted mt-3">Due {p.due_date}</p>
              )}

              <button
                onClick={e => { e.stopPropagation(); deleteProject.mutate(p.id) }}
                className="mt-3 text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100"
                aria-label="Delete project"
              >
                <Trash2 size={14} />
              </button>
            </div>
          )
        })}
      </div>

      {showNew && <NewProjectModal onClose={() => setShowNew(false)} />}
    </div>
  )
}
