import type { NoteStatus, Priority } from '../../types'

interface PropertiesPanelProps {
  status: NoteStatus
  priority: Priority
  dueDate: string | null
  tags: string[]
  folderId: string | null
  onChange: (patch: {
    status?: NoteStatus
    priority?: Priority
    dueDate?: string | null
    tags?: string[]
    folderId?: string | null
  }) => void
}

const STATUS_OPTIONS: { value: NoteStatus; label: string }[] = [
  { value: 'draft', label: 'Draft' },
  { value: 'in_progress', label: 'In progress' },
  { value: 'published', label: 'Published' },
]

const PRIORITY_OPTIONS: { value: Priority; label: string }[] = [
  { value: 'low', label: 'Low' },
  { value: 'medium', label: 'Medium' },
  { value: 'high', label: 'High' },
]

export function PropertiesPanel({
  status,
  priority,
  dueDate,
  tags,
  folderId,
  onChange,
}: PropertiesPanelProps) {
  return (
    <div className="grid grid-cols-2 gap-4 p-4 bg-surface-elevated border border-border rounded-xl">
      <div className="flex flex-col gap-1.5">
        <label htmlFor="note-status" className="text-xs text-text-secondary">
          Status
        </label>
        <select
          id="note-status"
          value={status}
          onChange={(e) => onChange({ status: e.target.value as NoteStatus })}
          className="w-full px-3 py-2 rounded-lg bg-surface-card border border-border text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/40 transition-all"
        >
          {STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label htmlFor="note-priority" className="text-xs text-text-secondary">
          Priority
        </label>
        <select
          id="note-priority"
          value={priority}
          onChange={(e) => onChange({ priority: e.target.value as Priority })}
          className="w-full px-3 py-2 rounded-lg bg-surface-card border border-border text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/40 transition-all"
        >
          {PRIORITY_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label htmlFor="note-due-date" className="text-xs text-text-secondary">
          Due date
        </label>
        <input
          id="note-due-date"
          type="date"
          value={dueDate ?? ''}
          onChange={(e) => onChange({ dueDate: e.target.value || null })}
          className="w-full px-3 py-2 rounded-lg bg-surface-card border border-border text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/40 transition-all"
        />
      </div>

      <div className="flex flex-col gap-1.5">
        <label htmlFor="note-folder" className="text-xs text-text-secondary">
          Folder ID
        </label>
        <input
          id="note-folder"
          type="text"
          value={folderId ?? ''}
          onChange={(e) => onChange({ folderId: e.target.value || null })}
          placeholder="Optional folder ID"
          className="w-full px-3 py-2 rounded-lg bg-surface-card border border-border text-text-primary text-sm placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-accent/40 transition-all"
        />
      </div>

      <div className="col-span-2 flex flex-col gap-1.5">
        <label htmlFor="note-tags" className="text-xs text-text-secondary">
          Tags
        </label>
        <input
          id="note-tags"
          type="text"
          value={tags.join(', ')}
          onChange={(e) =>
            onChange({
              tags: e.target.value
                .split(',')
                .map((t) => t.trim())
                .filter(Boolean),
            })
          }
          placeholder="Comma separated tags"
          className="w-full px-3 py-2 rounded-lg bg-surface-card border border-border text-text-primary text-sm placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-accent/40 transition-all"
        />
      </div>
    </div>
  )
}
