import { useEffect, useRef, useState } from 'react'
import { Trash2, X } from 'lucide-react'
import { PropertiesPanel } from '../molecules/PropertiesPanel'
import { BlockEditor } from './BlockEditor'
import { IconButton } from '../atoms/IconButton'
import { textToBlocks } from '../../utils/blocks'
import type { Folder, Note, NoteProperty, NoteStatus } from '../../types'
import type { NoteInput } from '../../hooks/useNotes'

const AUTOSAVE_DELAY_MS = 800

interface NoteDraft {
  title: string
  blocks: string
  status: NoteStatus
  tags: string[]
  folder_id: string | null
  properties: NoteProperty[]
}

export interface NoteSavePayload {
  data?: NoteInput
  properties?: NoteProperty[]
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
    properties: note.properties ?? [],
  }
}

interface NoteEditorPanelProps {
  note: Note
  folders: Folder[]
  saving: boolean
  saveError: string | null
  onSave: (payload: NoteSavePayload) => void
  onDelete: () => void
  onClose: () => void
}

export function NoteEditorPanel({
  note,
  folders,
  saving,
  saveError,
  onSave,
  onDelete,
  onClose,
}: NoteEditorPanelProps) {
  const [draft, setDraft] = useState<NoteDraft>(() => toDraft(note))
  const [dirty, setDirty] = useState(false)
  const loadedNoteId = useRef(note.id)
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const pending = useRef<NoteSavePayload>({})
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
    const payload = pending.current
    if (!payload.data && !payload.properties) return
    onSaveRef.current(payload)
    pending.current = {}
  }

  // Reload the draft only when a different note is selected, so autosave
  // round-trips don't clobber what is being typed.
  useEffect(() => {
    if (loadedNoteId.current === note.id) return
    loadedNoteId.current = note.id
    if (saveTimer.current) clearTimeout(saveTimer.current)
    saveTimer.current = null
    pending.current = {}
    setDraft(toDraft(note))
    setDirty(false)
  }, [note])

  // Unmounting (closing the panel, deleting, navigating away) must not discard
  // edits that are still inside the debounce window.
  useEffect(() => {
    return () => {
      if (saveTimer.current) clearTimeout(saveTimer.current)
      const payload = pending.current
      if (payload.data || payload.properties) onSaveRef.current(payload)
    }
  }, [])

  const update = (patch: Partial<NoteDraft>) => {
    const next = { ...draft, ...patch }
    setDraft(next)
    setDirty(true)

    const nextPending = { ...pending.current }
    const touchesNote =
      'title' in patch ||
      'blocks' in patch ||
      'status' in patch ||
      'tags' in patch ||
      'folder_id' in patch
    if (touchesNote) {
      nextPending.data = {
        title: next.title.trim() || 'Untitled',
        blocks: next.blocks,
        status: next.status,
        tags: next.tags,
        folder_id: next.folder_id,
      }
    }
    if ('properties' in patch) {
      nextPending.properties = next.properties
    }
    pending.current = nextPending

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
          className="flex-1 min-w-0 bg-transparent text-2xl font-semibold leading-tight text-text-primary placeholder-text-muted focus:outline-none"
        />
        <span
          className="flex items-center gap-1.5 text-xs text-text-muted shrink-0"
          aria-live="polite"
          title={saveError ?? undefined}
        >
          <span
            aria-hidden="true"
            className={`h-1.5 w-1.5 rounded-full transition-colors ${
              saveError
                ? 'bg-red-400'
                : saving || dirty
                  ? 'bg-accent'
                  : 'bg-text-muted/50'
            }`}
          />
          {saveError ? 'Save failed' : saving || dirty ? 'Saving…' : 'Saved'}
        </span>
        <IconButton
          icon={<Trash2 size={16} />}
          aria-label="Delete note"
          variant="ghost"
          onClick={() => {
            // Drop the pending autosave so unmount doesn't PUT a deleted note.
            if (saveTimer.current) clearTimeout(saveTimer.current)
            saveTimer.current = null
            pending.current = {}
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

      {saveError && (
        <p role="alert" className="text-xs text-red-400 pb-2">
          {saveError}
        </p>
      )}

      <div className="flex-1 min-h-0 flex flex-col pt-3 gap-3">
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
            properties={draft.properties}
            onPropertiesChange={(properties) => update({ properties })}
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
