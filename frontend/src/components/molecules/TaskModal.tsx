import { useState, useEffect } from 'react'
import { Modal } from '../atoms/Modal'
import { Button } from '../atoms/Button'
import { Input } from '../atoms/Input'
import { Textarea } from '../atoms/Textarea'
import { Select } from '../atoms/Select'
import type { Task, Priority, TaskStatus } from '../../types'

interface TaskForm {
  title: string
  description: string
  priority: Priority
  status: TaskStatus
  due_date: string
  recurrence: string
}

interface TaskModalProps {
  open: boolean
  onClose: () => void
  onSave: (data: Omit<Task, 'id' | 'created_at' | 'updated_at'>) => void
  task?: Task
  isPending?: boolean
}

const PRIORITIES: Priority[] = ['low', 'medium', 'high']
const STATUSES: TaskStatus[] = ['todo', 'in_progress', 'done']

const STATUS_LABEL: Record<TaskStatus, string> = {
  todo: 'To do',
  in_progress: 'In progress',
  done: 'Done',
}

const PRIORITY_COLOR: Record<Priority, string> = {
  low: 'border-border-subtle bg-surface-elevated text-text-secondary',
  medium: 'border-amber-700/50 bg-amber-950/30 text-amber-400',
  high: 'border-red-700/50 bg-red-950/30 text-red-400',
}

const PRIORITY_ACTIVE: Record<Priority, string> = {
  low: 'border-text-secondary bg-text-secondary/20 text-text-primary',
  medium: 'border-amber-400 bg-amber-900/60 text-amber-300',
  high: 'border-red-400 bg-red-900/60 text-red-300',
}

const RECURRENCE_OPTIONS = [
  { value: '', label: 'None' },
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
]

export function TaskModal({ open, onClose, onSave, task, isPending }: TaskModalProps) {
  const isEdit = Boolean(task)
  const [form, setForm] = useState<TaskForm>({
    title: '',
    description: '',
    priority: 'medium',
    status: 'todo',
    due_date: '',
    recurrence: '',
  })

  useEffect(() => {
    if (open) {
      setForm({
        title: task?.title ?? '',
        description: task?.description ?? '',
        priority: task?.priority ?? 'medium',
        status: task?.status ?? 'todo',
        due_date: task?.due_date?.slice(0, 10) ?? '',
        recurrence: task?.recurrence ?? '',
      })
    }
  }, [open, task])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.title.trim()) return
    onSave({
      title: form.title.trim(),
      description: form.description.trim(),
      priority: form.priority,
      status: form.status,
      due_date: form.due_date || null,
      project_id: task?.project_id ?? null,
      note_id: task?.note_id ?? null,
      recurrence: form.recurrence,
    })
  }

  return (
    <Modal open={open} onClose={onClose} title={isEdit ? 'Edit task' : 'New task'}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          autoFocus
          value={form.title}
          onChange={e => setForm(f => ({ ...f, title: e.target.value }))}
          placeholder="Task title"
        />

        <Textarea
          value={form.description}
          onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
          placeholder="Description (optional)"
          rows={2}
          autoResize
        />

        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-xs text-text-secondary mb-2">Priority</p>
            <div className="flex gap-1.5">
              {PRIORITIES.map(p => (
                <button
                  key={p}
                  type="button"
                  onClick={() => setForm(f => ({ ...f, priority: p }))}
                  className={`flex-1 text-xs px-2 py-1.5 rounded-md border capitalize transition-colors ${
                    form.priority === p ? PRIORITY_ACTIVE[p] : PRIORITY_COLOR[p]
                  }`}
                >
                  {p}
                </button>
              ))}
            </div>
          </div>

          <div>
            <p className="text-xs text-text-secondary mb-2">Due date</p>
            <Input
              type="date"
              value={form.due_date}
              onChange={e => setForm(f => ({ ...f, due_date: e.target.value }))}
            />
          </div>
        </div>

        {isEdit && (
          <div>
            <p className="text-xs text-text-secondary mb-2">Status</p>
            <div className="flex gap-1.5">
              {STATUSES.map(s => (
                <button
                  key={s}
                  type="button"
                  onClick={() => setForm(f => ({ ...f, status: s }))}
                  className={`flex-1 text-xs px-2 py-1.5 rounded-md border transition-colors ${
                    form.status === s
                      ? 'border-(--color-accent) bg-(--color-accent)/15 text-(--color-accent)'
                      : 'border-border bg-surface-elevated text-text-secondary'
                  }`}
                >
                  {STATUS_LABEL[s]}
                </button>
              ))}
            </div>
          </div>
        )}

        <Select
          label="Recurrence"
          value={form.recurrence}
          onChange={e => setForm(f => ({ ...f, recurrence: e.target.value }))}
          options={RECURRENCE_OPTIONS}
        />

        <div className="flex justify-end gap-2 pt-1">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={!form.title.trim() || isPending}>
            {isEdit ? 'Save changes' : 'Create task'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}
