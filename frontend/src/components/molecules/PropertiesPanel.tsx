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

const FIELD_CLASS =
  'w-full px-2.5 py-1.5 rounded-lg bg-surface-card border border-border text-text-primary text-sm placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-accent/40 transition-all'

const LABEL_CLASS = 'text-[11px] font-medium uppercase tracking-wide text-text-muted'

export function PropertiesPanel({
  status,
  tags,
  folderId,
  folders,
  onChange,
}: PropertiesPanelProps) {
  // One compact row instead of three stacked blocks: the writing surface is the
  // point of this panel, metadata should not eat a third of it.
  return (
    <div className="flex flex-wrap items-end gap-3">
      <div className="flex flex-col gap-1 flex-1 min-w-28">
        <label htmlFor="note-status" className={LABEL_CLASS}>
          Status
        </label>
        <select
          id="note-status"
          value={status}
          onChange={(e) => onChange({ status: e.target.value as NoteStatus })}
          className={FIELD_CLASS}
        >
          {STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1 flex-1 min-w-28">
        <label htmlFor="note-folder" className={LABEL_CLASS}>
          Folder
        </label>
        <select
          id="note-folder"
          value={folderId ?? ''}
          onChange={(e) => onChange({ folderId: e.target.value || null })}
          className={FIELD_CLASS}
        >
          <option value="">No folder</option>
          {folders.map((folder) => (
            <option key={folder.id} value={folder.id}>
              {folder.name}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1 flex-[2] min-w-36">
        <label htmlFor="note-tags" className={LABEL_CLASS}>
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
          placeholder="Add tags…"
          className={FIELD_CLASS}
        />
      </div>
    </div>
  )
}
