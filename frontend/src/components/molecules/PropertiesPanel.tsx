import type { Folder, NoteStatus } from '../../types'

interface PropertiesPanelProps {
  status: NoteStatus
  tags: string[]
  folderId: string | null
  folders: Folder[]
  onChange: (patch: {
    status?: NoteStatus
    tags?: string[]
    folderId?: string | null
  }) => void
}

const STATUS_OPTIONS: { value: NoteStatus; label: string }[] = [
  { value: 'draft', label: 'Draft' },
  { value: 'in_progress', label: 'In progress' },
  { value: 'published', label: 'Published' },
]

export function PropertiesPanel({
  status,
  tags,
  folderId,
  folders,
  onChange,
}: PropertiesPanelProps) {
  return (
    <div className="grid grid-cols-2 gap-x-4 gap-y-3">
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
        <label htmlFor="note-folder" className="text-xs text-text-secondary">
          Folder
        </label>
        <select
          id="note-folder"
          value={folderId ?? ''}
          onChange={(e) => onChange({ folderId: e.target.value || null })}
          className="w-full px-3 py-2 rounded-lg bg-surface-card border border-border text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/40 transition-all"
        >
          <option value="">No folder</option>
          {folders.map((folder) => (
            <option key={folder.id} value={folder.id}>
              {folder.name}
            </option>
          ))}
        </select>
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
