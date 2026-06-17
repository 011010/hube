import { useState } from 'react'
import { useTasks, useCreateTask, useUpdateTask, useDeleteTask } from '../../hooks/useTasks'
import { Badge } from '../../components/atoms/Badge'
import { Button } from '../../components/atoms/Button'
import type { Task, Priority, TaskStatus } from '../../types'

export function TasksPage() {
  const { data: tasks = [], isLoading } = useTasks()
  const createTask = useCreateTask()
  const updateTask = useUpdateTask()
  const deleteTask = useDeleteTask()
  const [newTitle, setNewTitle] = useState('')

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTitle.trim()) return
    createTask.mutate({
      title: newTitle.trim(),
      description: '',
      priority: 'medium',
      status: 'todo',
      due_date: null,
    })
    setNewTitle('')
  }

  const toggleStatus = (task: Task) => {
    const next: TaskStatus = task.status === 'done' ? 'todo' : 'done'
    updateTask.mutate({ id: task.id, data: { status: next } })
  }

  const priorityVariant = (p: Priority) =>
    p === 'high' ? 'danger' : p === 'medium' ? 'warning' : 'default'

  const byStatus = (status: TaskStatus) => tasks.filter(t => t.status === status)

  return (
    <div className="p-8 space-y-6">
      <h1 className="text-2xl font-semibold text-white">Tasks</h1>

      <form onSubmit={handleCreate} className="flex gap-3">
        <input
          value={newTitle}
          onChange={e => setNewTitle(e.target.value)}
          placeholder="New task..."
          className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-4 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500"
        />
        <Button type="submit" disabled={createTask.isPending}>Add</Button>
      </form>

      {isLoading && <p className="text-gray-400 text-sm">Loading...</p>}

      {(['todo', 'in_progress', 'done'] as TaskStatus[]).map(status => {
        const group = byStatus(status)
        if (group.length === 0) return null
        return (
          <section key={status}>
            <h2 className="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-2">
              {status.replace('_', ' ')}
            </h2>
            <ul className="space-y-2">
              {group.map(task => (
                <li key={task.id} className="flex items-center gap-3 bg-gray-900 border border-gray-800 rounded-lg px-4 py-3">
                  <input
                    type="checkbox"
                    checked={task.status === 'done'}
                    onChange={() => toggleStatus(task)}
                    className="w-4 h-4 accent-indigo-500"
                  />
                  <span className={`flex-1 text-sm ${task.status === 'done' ? 'line-through text-gray-500' : 'text-gray-200'}`}>
                    {task.title}
                  </span>
                  <Badge label={task.priority} variant={priorityVariant(task.priority)} />
                  <button
                    onClick={() => deleteTask.mutate(task.id)}
                    className="text-gray-600 hover:text-red-400 transition-colors text-xs"
                  >
                    ✕
                  </button>
                </li>
              ))}
            </ul>
          </section>
        )
      })}
    </div>
  )
}
