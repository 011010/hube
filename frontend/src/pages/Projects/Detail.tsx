import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Plus, Trash2, Check } from 'lucide-react'
import { useProject, useUpdateProject, useProjectTasks } from '../../hooks/useProjects'
import { useCreateTask, useUpdateTask, useDeleteTask } from '../../hooks/useTasks'
import { TaskModal } from '../../components/molecules/TaskModal'
import type { Task, ProjectStatus } from '../../types'

const STATUS_OPTS: { value: ProjectStatus; label: string }[] = [
  { value: 'planning', label: 'Planning' },
  { value: 'active', label: 'Active' },
  { value: 'completed', label: 'Completed' },
  { value: 'on_hold', label: 'On Hold' },
]

const PRIORITY_DOT: Record<string, string> = {
  low: 'bg-surface-elevated',
  medium: 'bg-amber-400',
  high: 'bg-red-400',
}

const STATUS_LABEL: Record<string, string> = {
  todo: 'To do',
  in_progress: 'In progress',
  done: 'Done',
}

export function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { data: project, isLoading } = useProject(id ?? null)
  const { data: tasks = [] } = useProjectTasks(id ?? '')
  const updateProject = useUpdateProject()
  const createTask = useCreateTask()
  const updateTask = useUpdateTask()
  const deleteTask = useDeleteTask()

  function invalidateProject() {
    qc.invalidateQueries({ queryKey: ['project', id] })
  }

  const [showNewTask, setShowNewTask] = useState(false)
  const [editTask, setEditTask] = useState<Task | null>(null)

  if (isLoading || !project) {
    return <div className="p-6 text-text-muted text-sm">Loading…</div>
  }

  const pct = project.task_count === 0 ? 0 : Math.round((project.completed_count / project.task_count) * 100)

  const todoTasks = tasks.filter(t => t.status === 'todo')
  const inProgressTasks = tasks.filter(t => t.status === 'in_progress')
  const doneTasks = tasks.filter(t => t.status === 'done')

  function handleStatusChange(status: ProjectStatus) {
    updateProject.mutate({ id: project!.id, data: { ...project, status } })
  }

  function handleCreateTask(data: Omit<Task, 'id' | 'created_at' | 'updated_at'>) {
    createTask.mutate({ ...data, project_id: project!.id }, {
      onSuccess: () => { setShowNewTask(false); invalidateProject() },
    })
  }

  function handleUpdateTask(data: Omit<Task, 'id' | 'created_at' | 'updated_at'>) {
    if (!editTask) return
    updateTask.mutate({ id: editTask.id, data: { ...data, project_id: project!.id } }, {
      onSuccess: () => { setEditTask(null); invalidateProject() },
    })
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <button
        onClick={() => navigate('/projects')}
        className="text-xs text-text-muted hover:text-text-secondary mb-4 flex items-center gap-1 transition-colors"
      >
        <ArrowLeft size={14} />
        Projects
      </button>

      {/* Header */}
      <div className="bg-surface-elevated border border-border rounded-xl p-5 mb-6">
        <div className="flex items-start gap-3">
          <span className="w-4 h-4 rounded-full shrink-0 mt-0.5" style={{ backgroundColor: project.color }} />
          <div className="flex-1 min-w-0">
            <h1 className="text-xl font-semibold text-text-primary">{project.name}</h1>
            {project.description && (
              <p className="text-sm text-text-muted mt-1">{project.description}</p>
            )}
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <select
              value={project.status}
              onChange={e => handleStatusChange(e.target.value as ProjectStatus)}
              className="bg-surface-base border border-border text-sm text-text-secondary rounded-lg px-2 py-1 focus:outline-none focus:border-(--color-accent) transition-colors"
            >
              {STATUS_OPTS.map(o => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>
          </div>
        </div>

        <div className="mt-4">
          <div className="flex justify-between text-xs text-text-muted mb-1">
            <span>{project.completed_count}/{project.task_count} tasks completed</span>
            <span>{pct}%</span>
          </div>
          <div className="h-2 bg-surface-base rounded-full overflow-hidden">
            <div
              className="h-full rounded-full transition-all duration-500"
              style={{ width: `${pct}%`, backgroundColor: project.color }}
            />
          </div>
        </div>
      </div>

      {/* Tasks */}
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-medium text-text-muted uppercase tracking-wider">Tasks</h2>
        <button
          onClick={() => setShowNewTask(true)}
          className="text-sm text-(--color-accent) hover:text-(--color-accent-hover) transition-colors flex items-center gap-1"
        >
          <Plus size={14} />
          Add task
        </button>
      </div>

      {tasks.length === 0 && (
        <div className="text-center py-12 text-text-muted text-sm">
          No tasks yet.{' '}
          <button onClick={() => setShowNewTask(true)} className="text-(--color-accent) hover:text-(--color-accent-hover) underline">
            Add one.
          </button>
        </div>
      )}

      {[
        { label: 'In progress', items: inProgressTasks },
        { label: 'To do', items: todoTasks },
        { label: 'Done', items: doneTasks },
      ].map(({ label, items }) => items.length > 0 && (
        <div key={label} className="mb-5">
          <p className="text-xs text-text-muted mb-2 uppercase tracking-wider">{label}</p>
          <div className="space-y-2">
            {items.map(task => (
              <div
                key={task.id}
                className="flex items-center gap-3 bg-surface-elevated border border-border rounded-lg px-4 py-3 group hover:border-border-subtle transition-colors"
              >
                <button
                  onClick={() => updateTask.mutate(
                    { id: task.id, data: { ...task, status: task.status === 'done' ? 'todo' : 'done' } },
                    { onSuccess: invalidateProject },
                  )}
                  className={`w-4 h-4 rounded border flex-shrink-0 flex items-center justify-center transition-colors ${
                    task.status === 'done'
                      ? 'bg-emerald-600 border-emerald-600 text-white'
                      : 'border-text-muted hover:border-text-secondary'
                  }`}
                  aria-label={task.status === 'done' ? 'Mark as todo' : 'Mark as done'}
                >
                  {task.status === 'done' && <Check size={12} strokeWidth={3} />}
                </button>

                <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${PRIORITY_DOT[task.priority]}`} />

                <span
                  className={`flex-1 text-sm cursor-pointer ${task.status === 'done' ? 'line-through text-text-muted' : 'text-text-secondary'}`}
                  onClick={() => setEditTask(task)}
                >
                  {task.title}
                </span>

                {task.due_date && (
                  <span className="text-xs text-text-muted shrink-0">
                    {new Date(task.due_date).toLocaleDateString()}
                  </span>
                )}

                <span className="text-xs text-text-muted shrink-0">{STATUS_LABEL[task.status]}</span>

                <button
                  onClick={() => deleteTask.mutate(task.id)}
                  className="text-text-muted hover:text-red-400 transition-all ml-1 opacity-0 group-hover:opacity-100"
                  aria-label="Delete task"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            ))}
          </div>
        </div>
      ))}

      <TaskModal
        open={showNewTask}
        onClose={() => setShowNewTask(false)}
        onSave={handleCreateTask}
        isPending={createTask.isPending}
      />

      <TaskModal
        open={Boolean(editTask)}
        onClose={() => setEditTask(null)}
        onSave={handleUpdateTask}
        task={editTask ?? undefined}
        isPending={updateTask.isPending}
      />
    </div>
  )
}
