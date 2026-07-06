import { useState } from 'react'
import { Plus, Trash2, ListChecks, Circle, CheckCircle2 } from 'lucide-react'
import { useTasks, useCreateTask, useUpdateTask, useDeleteTask } from '../../hooks/useTasks'
import { useViewPreference } from '../../hooks/useViewPreference'
import { Badge } from '../../components/atoms/Badge'
import { Button } from '../../components/atoms/Button'
import { PageHeader } from '../../components/molecules/PageHeader'
import { EmptyState } from '../../components/molecules/EmptyState'
import { TaskModal } from '../../components/molecules/TaskModal'
import { ViewToggle } from '../../components/molecules/ViewToggle'
import { KanbanBoard } from '../../components/organisms/KanbanBoard'
import { DataTable } from '../../components/organisms/DataTable'
import type { Task, Priority, TaskStatus } from '../../types'

const VALID_STATUSES: TaskStatus[] = ['todo', 'in_progress', 'done']

export function TasksPage() {
  const { data: tasks = [], isLoading } = useTasks()
  const createTask = useCreateTask()
  const updateTask = useUpdateTask()
  const deleteTask = useDeleteTask()

  const [view, setView] = useViewPreference('tasks_view')
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

  const statusVariant = (s: TaskStatus) =>
    s === 'done' ? 'success' : s === 'in_progress' ? 'warning' : 'default'

  const statusLabel = (s: TaskStatus) =>
    s === 'in_progress' ? 'In Progress' : s.charAt(0).toUpperCase() + s.slice(1)

  const columns = [
    { id: 'todo', title: 'To Do', items: tasks.filter(t => t.status === 'todo') },
    { id: 'in_progress', title: 'In Progress', items: tasks.filter(t => t.status === 'in_progress') },
    { id: 'done', title: 'Done', items: tasks.filter(t => t.status === 'done') },
  ]

  const handleMove = (itemId: string, _sourceColumnId: string, targetColumnId: string) => {
    if (!VALID_STATUSES.includes(targetColumnId as TaskStatus)) return
    const task = tasks.find(t => t.id === itemId)
    if (!task || task.status === targetColumnId) return
    updateTask.mutate({ id: task.id, data: { ...task, status: targetColumnId as TaskStatus } })
  }

  const renderCard = (task: Task) => (
    <div className="space-y-2">
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-start gap-2 flex-1 min-w-0">
          <button
            onClick={() => toggleStatus(task)}
            className="shrink-0 text-text-muted hover:text-(--color-accent) transition-colors mt-0.5"
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
            className={`text-left text-sm font-medium min-w-0 ${
              task.status === 'done'
                ? 'line-through text-text-muted'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            {task.title}
          </button>
        </div>
        <button
          onClick={() => deleteTask.mutate(task.id)}
          className="text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100"
          aria-label="Delete task"
        >
          <Trash2 size={14} />
        </button>
      </div>
      {task.due_date && (
        <p className="text-xs text-text-muted">{task.due_date.slice(0, 10)}</p>
      )}
      <Badge label={task.priority} variant={priorityVariant(task.priority)} />
    </div>
  )

  const tableColumns = [
    {
      key: 'title',
      header: 'Title',
      render: (task: Task) => (
        <div>
          <button
            onClick={() => setEditTask(task)}
            className={`text-left text-sm font-medium ${
              task.status === 'done'
                ? 'line-through text-text-muted'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            {task.title}
          </button>
          {task.description && (
            <p className="text-xs text-text-muted truncate max-w-md">{task.description}</p>
          )}
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      sortValue: (task: Task) => task.status,
      render: (task: Task) => (
        <button
          type="button"
          onClick={() => toggleStatus(task)}
          className="cursor-pointer"
          aria-label={task.status === 'done' ? 'Mark as todo' : 'Mark as done'}
        >
          <Badge label={statusLabel(task.status)} variant={statusVariant(task.status)} />
        </button>
      ),
    },
    {
      key: 'priority',
      header: 'Priority',
      sortable: true,
      sortValue: (task: Task) => task.priority,
      render: (task: Task) => <Badge label={task.priority} variant={priorityVariant(task.priority)} />,
    },
    {
      key: 'due_date',
      header: 'Due date',
      sortable: true,
      sortValue: (task: Task) => task.due_date ?? '',
      render: (task: Task) => (
        <span className="text-text-secondary">{task.due_date ? task.due_date.slice(0, 10) : '-'}</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (task: Task) => (
        <button
          type="button"
          onClick={() => deleteTask.mutate(task.id)}
          className="text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100"
          aria-label="Delete task"
        >
          <Trash2 size={14} />
        </button>
      ),
    },
  ]

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Tasks"
        actions={
          <div className="flex items-center gap-2">
            <ViewToggle value={view} onChange={setView} />
            <Button onClick={() => setCreateOpen(true)} icon={<Plus size={16} />}>New task</Button>
          </div>
        }
      />

      {isLoading && <p className="text-text-muted text-sm">Loading...</p>}

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

      {tasks.length > 0 && !isLoading && (
        view === 'kanban' ? (
          <KanbanBoard
            columns={columns}
            getItemId={task => task.id}
            renderCard={renderCard}
            onMove={handleMove}
          />
        ) : (
          <DataTable
            columns={tableColumns}
            data={tasks}
            getRowKey={task => task.id}
          />
        )
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
