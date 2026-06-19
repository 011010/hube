import { useState } from 'react'
import { useTasks, useCreateTask, useUpdateTask, useDeleteTask } from '../../hooks/useTasks'
import { Badge } from '../../components/atoms/Badge'
import { Button } from '../../components/atoms/Button'
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
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-white">Tasks</h1>
        <Button onClick={() => setCreateOpen(true)}>+ New task</Button>
      </div>

      {isLoading && <p className="text-gray-400 text-sm">Loading...</p>}

      {(['todo', 'in_progress', 'done'] as TaskStatus[]).map(status => {
        const group = byStatus(status)
        if (group.length === 0) return null
        return (
          <section key={status}>
            <h2 className="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-2">
              {status.replace('_', ' ')} · {group.length}
            </h2>
            <ul className="space-y-2">
              {group.map(task => (
                <li
                  key={task.id}
                  className="flex items-center gap-3 bg-gray-900 border border-gray-800 rounded-lg px-4 py-3 group"
                >
                  <input
                    type="checkbox"
                    checked={task.status === 'done'}
                    onChange={() => toggleStatus(task)}
                    className="w-4 h-4 accent-indigo-500 shrink-0"
                  />
                  <button
                    onClick={() => setEditTask(task)}
                    className={`flex-1 min-w-0 text-sm text-left transition-colors ${
                      task.status === 'done'
                        ? 'line-through text-gray-500'
                        : 'text-gray-200 hover:text-white'
                    }`}
                  >
                    <span>{task.title}</span>
                    {task.due_date && (
                      <span className="ml-2 text-xs text-gray-500">{task.due_date.slice(0, 10)}</span>
                    )}
                    {task.description && (
                      <span className="block text-xs text-gray-500 mt-0.5 truncate">{task.description}</span>
                    )}
                  </button>
                  <Badge label={task.priority} variant={priorityVariant(task.priority)} />
                  <button
                    onClick={() => deleteTask.mutate(task.id)}
                    className="text-gray-700 hover:text-red-400 transition-colors text-xs opacity-0 group-hover:opacity-100"
                  >
                    ✕
                  </button>
                </li>
              ))}
            </ul>
          </section>
        )
      })}

      {tasks.length === 0 && !isLoading && (
        <p className="text-gray-500 text-sm">No tasks yet. Create one to get started.</p>
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
