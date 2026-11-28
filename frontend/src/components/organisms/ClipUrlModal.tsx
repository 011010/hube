import { useState, type FormEvent } from 'react'
import { Modal } from '../atoms/Modal'
import { Button } from '../atoms/Button'
import { FIELD_CLASS, LABEL_CLASS } from '../molecules/fieldStyles'
import { useClipNote } from '../../hooks/useNotes'
import type { Note } from '../../types'

// The mutation lives inside the modal on purpose: every outcome except the
// two navigations (fresh note, existing note) is purely presentational —
// pending state, the 409 "already clipped" line, inline server errors — so
// the page never needs clip state of its own. The parent mounts this modal
// only while open, so every open starts a fresh mutation, and unmounting
// after a success discards its state without any reset juggling.
interface ClipUrlModalProps {
  // Destination folder: sent as the request's folder_id and mirrored by the
  // hint. The name alone could not keep the hint honest, so both come in.
  folderId?: string | null
  // Hint label for folderId; "All notes" is shown when it is absent.
  folderName?: string
  onClipped: (note: Note) => void
  onOpenExisting: (id: string) => void
  onClose: () => void
}

type ClipFailure =
  | { kind: 'existing'; existingId: string }
  | { kind: 'error'; message: string }

const GENERIC_CLIP_ERROR = 'Clip failed'
// Server error strings are echoed back, but never unbounded.
const MAX_CLIP_ERROR_LENGTH = 200

// The backend speaks through axios: the status and body ride on
// err.response. A 409 carries the existing note's id; anything else collapses
// into the server's message (trimmed to one bounded line) or a generic one.
function clipFailureOf(err: unknown): ClipFailure {
  const e = err as {
    response?: { status?: number; data?: { error?: unknown; existing_id?: unknown } }
    message?: string
  }
  const existingId = e.response?.data?.existing_id
  if (e.response?.status === 409 && typeof existingId === 'string' && existingId) {
    return { kind: 'existing', existingId }
  }
  const raw = e.response?.data?.error
  const message = typeof raw === 'string' ? raw : e.message
  const text = (message ?? '').trim().slice(0, MAX_CLIP_ERROR_LENGTH)
  return { kind: 'error', message: text || GENERIC_CLIP_ERROR }
}

export function ClipUrlModal({ folderId, folderName, onClipped, onOpenExisting, onClose }: ClipUrlModalProps) {
  const clip = useClipNote()
  const [url, setUrl] = useState('')

  // The input keeps its value across failures so a 409 does not force a
  // reopen to clip a different URL; editing clears the stale alert (but
  // never mid-flight, where reset would flip the pending UI back on).
  const handleUrlChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUrl(e.target.value)
    if (!clip.isPending) clip.reset()
  }

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    const target = url.trim()
    if (!target || clip.isPending) return
    clip.mutate(
      { url: target, folder_id: folderId ?? null },
      { onSuccess: (note) => onClipped(note) },
    )
  }

  const failure = clip.error ? clipFailureOf(clip.error) : null

  return (
    <Modal open onClose={onClose} title="Clip article">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label htmlFor="clip-article-url" className={LABEL_CLASS}>
            Article URL
          </label>
          <input
            id="clip-article-url"
            autoFocus
            type="url"
            value={url}
            onChange={handleUrlChange}
            placeholder="https://…"
            className={FIELD_CLASS}
          />
        </div>

        <p className="text-xs text-text-muted">Saves to: {folderName ?? 'All notes'}</p>

        {failure?.kind === 'existing' && (
          <div role="alert" className="flex items-center gap-2 text-xs text-red-400">
            <span>Already clipped.</span>
            <button
              type="button"
              onClick={() => onOpenExisting(failure.existingId)}
              className="text-accent underline underline-offset-2 hover:text-accent-hover transition-colors"
            >
              Open existing note
            </button>
          </div>
        )}

        {failure?.kind === 'error' && (
          <p role="alert" className="text-xs text-red-400">
            {failure.message}
          </p>
        )}

        <div className="flex justify-end gap-2 pt-1">
          <Button type="button" variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={!url.trim() || clip.isPending}>
            {clip.isPending ? 'Clipping…' : 'Clip'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}
