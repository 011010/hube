import { useCallback, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus, Trash2, FolderKanban } from 'lucide-react'
import { useProjects, useCreateProject, useDeleteProject, useUpdateProject } from '../../hooks/useProjects'
import { useViewPreference } from '../../hooks/useViewPreference'
import { formatDate } from '../../utils/date'
import { progress, statusLabel, statusVariant } from '../../utils/project'
import { Badge } from '../../components/atoms/Badge'
import { Button } from '../../components/atoms/Button'
import { Input } from '../../components/atoms/Input'
import { Textarea } from '../../components/atoms/Textarea'
import { Modal } from '../../components/atoms/Modal'
import { PageHeader } from '../../components/molecules/PageHeader'
import { EmptyState } from '../../components/molecules/EmptyState'
import { ViewToggle } from '../../components/molecules/ViewToggle'
import { KanbanBoard } from '../../components/organisms/KanbanBoard'
import { DataTable } from '../../components/organisms/DataTable'
import type { Project, ProjectStatus } from '../../types'

const VALID_STATUSES: ProjectStatus[] = ['planning', 'active', 'completed', 'on_hold']

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
  const updateProject = useUpdateProject()
  const deleteProject = useDeleteProject()

  const [view, setView] = useViewPreference('projects_view')
  const [showNew, setShowNew] = useState(false)

  const handleMove = useCallback((itemId: string, _sourceColumnId: string, targetColumnId: string) => {
    const project = projects.find(p => p.id === itemId)
    if (!project || !VALID_STATUSES.includes(targetColumnId as ProjectStatus)) return
    updateProject.mutate({ id: project.id, data: { status: targetColumnId as ProjectStatus } })
  }, [projects, updateProject])

  const columns = useMemo(
    () => [
      { id: 'planning', title: 'Planning', items: projects.filter(p => p.status === 'planning') },
      { id: 'active', title: 'Active', items: projects.filter(p => p.status === 'active') },
      { id: 'completed', title: 'Completed', items: projects.filter(p => p.status === 'completed') },
      { id: 'on_hold', title: 'On Hold', items: projects.filter(p => p.status === 'on_hold') },
    ],
    [projects],
  )

  const renderCard = useCallback((project: Project) => {
    const pct = progress(project)
    const dueDate = formatDate(project.due_date)
    return (
      <div className="space-y-2">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2 flex-1 min-w-0">
            <span
              className="w-3 h-3 rounded-full shrink-0"
              style={{ backgroundColor: project.color }}
            />
            <Link
              to={`/projects/${project.id}`}
              className="text-left text-sm font-medium text-text-primary leading-snug truncate hover:text-text-secondary flex-1 min-w-0 no-underline"
            >
              {project.name}
            </Link>
          </div>
          <Badge label={statusLabel(project.status)} variant={statusVariant(project.status)} />
        </div>

        <div className="mb-1 flex items-center justify-between text-xs text-text-muted">
          <span>{project.completed_count}/{project.task_count} tasks</span>
          <span>{pct}%</span>
        </div>
        <div className="h-1.5 bg-surface-base rounded-full overflow-hidden">
          <div
            className="h-full rounded-full transition-all"
            style={{ width: `${pct}%`, backgroundColor: project.color }}
          />
        </div>

        <div className="flex items-center justify-between min-h-[20px]">
          {dueDate && <p className="text-xs text-text-muted">Due {dueDate}</p>}
          <button
            type="button"
            onClick={e => { e.stopPropagation(); deleteProject.mutate(project.id) }}
            className="text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100 focus-visible:opacity-100 ml-auto"
            aria-label="Delete project"
          >
            <Trash2 size={14} />
          </button>
        </div>
      </div>
    )
  }, [deleteProject])

  const tableColumns = useMemo(
    () => [
      {
        key: 'name',
        header: 'Name',
        render: (project: Project) => (
          <div>
            <Link
              to={`/projects/${project.id}`}
              className="text-left text-sm font-medium text-text-secondary hover:text-text-primary no-underline"
            >
              {project.name}
            </Link>
            {project.description && (
              <p className="text-xs text-text-muted truncate max-w-md">{project.description}</p>
            )}
          </div>
        ),
      },
      {
        key: 'status',
        header: 'Status',
        sortable: true,
        sortValue: (project: Project) => project.status,
        render: (project: Project) => (
          <Badge label={statusLabel(project.status)} variant={statusVariant(project.status)} />
        ),
      },
      {
        key: 'progress',
        header: 'Progress',
        sortable: true,
        sortValue: (project: Project) => progress(project),
        render: (project: Project) => {
          const pct = progress(project)
          return <span className="text-text-secondary">{pct}%</span>
        },
      },
      {
        key: 'due_date',
        header: 'Due date',
        sortable: true,
        sortValue: (project: Project) => project.due_date ?? '',
        render: (project: Project) => {
          const dueDate = formatDate(project.due_date)
          return <span className="text-text-secondary">{dueDate ?? '-'}</span>
        },
      },
      {
        key: 'actions',
        header: '',
        render: (project: Project) => (
          <button
            type="button"
            onClick={() => deleteProject.mutate(project.id)}
            className="text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
            aria-label="Delete project"
          >
            <Trash2 size={14} />
          </button>
        ),
      },
    ],
    [deleteProject],
  )

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Projects"
        actions={
          <div className="flex items-center gap-2">
            <ViewToggle value={view} onChange={setView} />
            <Button onClick={() => setShowNew(true)} icon={<Plus size={16} />}>New project</Button>
          </div>
        }
      />

      {isLoading && <p className="text-text-muted text-sm">Loading…</p>}

      {projects.length === 0 && !isLoading && (
        <EmptyState
          icon={<FolderKanban size={44} />}
          title="No projects yet"
          description="Create your first project to organise tasks and track progress."
          action={
            <Button onClick={() => setShowNew(true)} icon={<Plus size={16} />}>New project</Button>
          }
        />
      )}

      {projects.length > 0 && !isLoading && (
        view === 'kanban' ? (
          <KanbanBoard
            columns={columns}
            getItemId={project => project.id}
            renderCard={renderCard}
            onMove={handleMove}
          />
        ) : (
          <DataTable
            columns={tableColumns}
            data={projects}
            getRowKey={project => project.id}
          />
        )
      )}

      {showNew && <NewProjectModal onClose={() => setShowNew(false)} />}
    </div>
  )
}
