import { useState } from 'react'
import { Plus, Trash2, CheckCircle2, Circle, ListChecks } from 'lucide-react'
import { useTasks, useCreateTask, useUpdateTask, useDeleteTask } from '../../hooks/useTasks'
import { Badge } from '../../components/atoms/Badge'
import { Button } from '../../components/atoms/Button'
import { PageHeader } from '../../components/molecules/PageHeader'
import { EmptyState } from '../../components/molecules/EmptyState'
import { TaskModal } from '../../components/molecules/TaskModal'
import type { Task, Priority, TaskStatus } from '../../types'

export function TasksPage() {
  const { data: tasks = [], isLoading } = useTasks()
  const createTask = useCreateTask()
  const updateTask = useUpdateTask()
  const deleteTask = useDeleteTask()

  const [createOpen, setCreateOpen] = useState(false)
  const [editTask, setEditTask] = useState<Task | null>(null)

  const handleCreate = (data: Omit<Task, 'id' | 'created_at' | 'updated_at'>) => {
    createTask.mutate(data, { onSuccess: () => setCreateOpen(false) })
  }

  const handleUpdate = (data: Omit<Task, 'id' | 'created_at' | 'updated_at'>) => {
    if (!editTask) return
    updateTask.mutate(
      { id: editTask.id, data: { ...editTask, ...data } },
      { onSuccess: () => setEditTask(null) },
    )
  }

  const toggleStatus = (task: Task) => {
    const next: TaskStatus = task.status === 'done' ? 'todo' : 'done'
    updateTask.mutate({ id: task.id, data: { ...task, status: next } })
  }

  const priorityVariant = (p: Priority) =>
    p === 'high' ? 'danger' : p === 'medium' ? 'warning' : 'default'

  const byStatus = (status: TaskStatus) => tasks.filter(t => t.status === status)

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Tasks"
        actions={
          <Button onClick={() => setCreateOpen(true)} icon={<Plus size={16} />}>New task</Button>
        }
      />

      {isLoading && <p className="text-text-muted text-sm">Loading...</p>}

      {(['todo', 'in_progress', 'done'] as TaskStatus[]).map(status => {
        const group = byStatus(status)
        if (group.length === 0) return null
        return (
          <section key={status}>
            <h2 className="text-xs font-semibold text-text-muted uppercase tracking-widest mb-2">
              {status.replace('_', ' ')} · {group.length}
            </h2>
            <ul className="space-y-2">
              {group.map(task => (
                <li
                  key={task.id}
                  className="flex items-center gap-3 bg-surface-elevated border border-border rounded-lg px-4 py-3 group"
                >
                  <button
                    onClick={() => toggleStatus(task)}
                    className="shrink-0 text-text-muted hover:text-(--color-accent) transition-colors"
                    aria-label={task.status === 'done' ? 'Mark as todo' : 'Mark as done'}
                  >
                    {task.status === 'done' ? (
                      <CheckCircle2 size={18} className="text-emerald-400" />
                    ) : (
                      <Circle size={18} />
                    )}
                  </button>
                  <button
                    onClick={() => setEditTask(task)}
                    className={`flex-1 min-w-0 text-sm text-left transition-colors ${
                      task.status === 'done'
                        ? 'line-through text-text-muted'
                        : 'text-text-secondary hover:text-text-primary'
                    }`}
                  >
                    <span>{task.title}</span>
                    {task.due_date && (
                      <span className="ml-2 text-xs text-text-muted">{task.due_date.slice(0, 10)}</span>
                    )}
                    {task.description && (
                      <span className="block text-xs text-text-muted mt-0.5 truncate">{task.description}</span>
                    )}
                  </button>
                  <Badge label={task.priority} variant={priorityVariant(task.priority)} />
                  <button
                    onClick={() => deleteTask.mutate(task.id)}
                    className="text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100"
                    aria-label="Delete task"
                  >
                    <Trash2 size={14} />
                  </button>
                </li>
              ))}
            </ul>
          </section>
        )
      })}

      {tasks.length === 0 && !isLoading && (
        <EmptyState
          icon={<ListChecks size={44} />}
          title="No tasks yet"
          description="Create one to get started."
          action={
            <Button onClick={() => setCreateOpen(true)} icon={<Plus size={16} />}>New task</Button>
          }
        />
      )}

      <TaskModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSave={handleCreate}
        isPending={createTask.isPending}
      />

      <TaskModal
        open={Boolean(editTask)}
        onClose={() => setEditTask(null)}
        onSave={handleUpdate}
        task={editTask ?? undefined}
        isPending={updateTask.isPending}
      />
    </div>
  )
}
