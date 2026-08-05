import { useEffect, useRef, useState } from 'react'
import { Trash2, X } from 'lucide-react'
import { PropertiesPanel } from '../molecules/PropertiesPanel'
import { BlockEditor } from './BlockEditor'
import { IconButton } from '../atoms/IconButton'
import { textToBlocks } from '../../utils/blocks'
import type { Folder, Note, NoteStatus } from '../../types'
import type { NoteInput } from '../../hooks/useNotes'

const AUTOSAVE_DELAY_MS = 800

interface NoteDraft {
  title: string
  blocks: string
  status: NoteStatus
  tags: string[]
  folder_id: string | null
}

function toDraft(note: Note): NoteDraft {
  return {
    title: note.title,
    // Notes created before the block editor only have plain-text content.
    blocks: note.blocks.trim()
      ? note.blocks
      : (note.content.trim() ? textToBlocks(note.content) : note.blocks),
    status: note.status,
    tags: note.tags,
    folder_id: note.folder_id,
  }
}

interface NoteEditorPanelProps {
  note: Note
  folders: Folder[]
  onSave: (data: NoteInput) => void
  onDelete: () => void
  onClose: () => void
}

export function NoteEditorPanel({ note, folders, onSave, onDelete, onClose }: NoteEditorPanelProps) {
  const [draft, setDraft] = useState<NoteDraft>(() => toDraft(note))
  const [dirty, setDirty] = useState(false)
  const loadedNoteId = useRef(note.id)
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const pending = useRef<NoteInput | null>(null)
  // Kept in a ref so the unmount cleanup can flush without re-subscribing.
  const onSaveRef = useRef(onSave)
  useEffect(() => {
    onSaveRef.current = onSave
  }, [onSave])

  const flush = () => {
    if (saveTimer.current) {
      clearTimeout(saveTimer.current)
      saveTimer.current = null
    }
    if (!pending.current) return
    onSaveRef.current(pending.current)
    pending.current = null
  }

  // Reload the draft only when a different note is selected, so autosave
  // round-trips don't clobber what is being typed.
  useEffect(() => {
    if (loadedNoteId.current === note.id) return
    loadedNoteId.current = note.id
    if (saveTimer.current) clearTimeout(saveTimer.current)
    saveTimer.current = null
    pending.current = null
    setDraft(toDraft(note))
    setDirty(false)
  }, [note])

  // Unmounting (closing the panel, deleting, navigating away) must not discard
  // edits that are still inside the debounce window.
  useEffect(() => {
    return () => {
      if (saveTimer.current) clearTimeout(saveTimer.current)
      if (pending.current) onSaveRef.current(pending.current)
    }
  }, [])

  const update = (patch: Partial<NoteDraft>) => {
    const next = { ...draft, ...patch }
    setDraft(next)
    setDirty(true)
    pending.current = {
      title: next.title.trim() || 'Untitled',
      blocks: next.blocks,
      status: next.status,
      tags: next.tags,
      folder_id: next.folder_id,
    }
    if (saveTimer.current) clearTimeout(saveTimer.current)
    saveTimer.current = setTimeout(() => {
      flush()
      setDirty(false)
    }, AUTOSAVE_DELAY_MS)
  }

  return (
    <section
      aria-label="Note editor"
      className="flex flex-col min-w-0 h-full bg-surface-base"
    >
      <div className="flex items-center gap-2 pb-3 border-b border-border shrink-0">
        <input
          value={draft.title}
          onChange={(e) => update({ title: e.target.value })}
          placeholder="Untitled"
          aria-label="Note title"
          className="flex-1 min-w-0 bg-transparent text-lg font-semibold text-text-primary placeholder-text-muted focus:outline-none"
        />
        <span className="text-xs text-text-muted shrink-0" aria-live="polite">
          {dirty ? 'Saving…' : 'Saved'}
        </span>
        <IconButton
          icon={<Trash2 size={16} />}
          aria-label="Delete note"
          variant="ghost"
          onClick={() => {
            // Drop the pending autosave so unmount doesn't PUT a deleted note.
            if (saveTimer.current) clearTimeout(saveTimer.current)
            saveTimer.current = null
            pending.current = null
            onDelete()
          }}
        />
        <IconButton
          icon={<X size={16} />}
          aria-label="Close editor"
          variant="ghost"
          onClick={onClose}
        />
      </div>

      <div className="flex-1 min-h-0 flex flex-col pt-4 gap-4">
        <div className="shrink-0">
          <PropertiesPanel
            status={draft.status}
            tags={draft.tags}
            folderId={draft.folder_id}
            folders={folders}
            onChange={(patch) =>
              update({
                ...(patch.status !== undefined && { status: patch.status }),
                ...(patch.tags !== undefined && { tags: patch.tags }),
                ...(patch.folderId !== undefined && { folder_id: patch.folderId }),
              })
            }
          />
        </div>

        {/* Takes whatever height is left so tall windows aren't half empty. */}
        <div className="flex-1 min-h-0">
          <BlockEditor
            value={draft.blocks}
            onChange={(json) => update({ blocks: json })}
            placeholder="Start writing…"
          />
        </div>
      </div>
    </section>
  )
}
