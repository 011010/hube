import { useState, useMemo, useCallback } from 'react'
import { Plus, Trash2, Folder, Sparkles, Search, X, FileText } from 'lucide-react'
import {
  useNotes, useCreateNote, useUpdateNote, useDeleteNote,
  useFolders, useCreateFolder, useDeleteFolder, useSearchNotes, useSemanticSearch,
} from '../../hooks/useNotes'
import { useViewPreference } from '../../hooks/useViewPreference'
import { formatDate } from '../../utils/date'
import { blocksToText } from '../../utils/blocks'
import { EmptyState } from '../../components/molecules/EmptyState'
import { PageHeader } from '../../components/molecules/PageHeader'
import { ViewToggle } from '../../components/molecules/ViewToggle'
import { PropertiesPanel } from '../../components/molecules/PropertiesPanel'
import { KanbanBoard } from '../../components/organisms/KanbanBoard'
import { DataTable } from '../../components/organisms/DataTable'
import { Modal } from '../../components/atoms/Modal'
import { Input } from '../../components/atoms/Input'
import { Button } from '../../components/atoms/Button'
import { Badge } from '../../components/atoms/Badge'
import { BlockEditor } from '../../components/organisms/BlockEditor'
import type { Note, NoteStatus, Priority } from '../../types'
import type { NoteInput } from '../../hooks/useNotes'

const VALID_STATUSES: NoteStatus[] = ['draft', 'in_progress', 'published']

interface NoteForm {
  title: string
  blocks: string
  status: NoteStatus
  priority: Priority
  due_date: string
  tags: string[]
  folder_id: string | null
}

function emptyForm(folderId?: string): NoteForm {
  return {
    title: '',
    blocks: '',
    status: 'draft',
    priority: 'medium',
    due_date: '',
    tags: [],
    folder_id: folderId ?? null,
  }
}

function noteToForm(note: Note): NoteForm {
  return {
    title: note.title,
    blocks: note.blocks,
    status: note.status,
    priority: note.priority,
    due_date: note.due_date?.slice(0, 10) ?? '',
    tags: note.tags,
    folder_id: note.folder_id,
  }
}

const statusLabel = (s: NoteStatus) =>
  s === 'in_progress' ? 'In Progress' : s.charAt(0).toUpperCase() + s.slice(1)

const priorityVariant = (p: Priority) =>
  p === 'high' ? 'danger' : p === 'medium' ? 'warning' : 'default'

const statusVariant = (s: NoteStatus) =>
  s === 'published' ? 'success' : s === 'in_progress' ? 'warning' : 'default'

interface NoteModalProps {
  open: boolean
  onClose: () => void
  onSave: (data: NoteInput) => void
  note?: Note | null
  folderId?: string
  isPending: boolean
}

function NoteModal({ open, onClose, onSave, note, folderId, isPending }: NoteModalProps) {
  const isEdit = Boolean(note)
  const [form, setForm] = useState<NoteForm>(() =>
    note ? noteToForm(note) : emptyForm(folderId),
  )

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSave({
      title: form.title.trim() || 'Untitled',
      blocks: form.blocks,
      status: form.status,
      priority: form.priority,
      due_date: form.due_date || null,
      tags: form.tags,
      folder_id: form.folder_id || null,
    })
  }

  return (
    <Modal open={open} onClose={onClose} title={isEdit ? 'Edit note' : 'New note'}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          autoFocus
          value={form.title}
          onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
          placeholder="Note title"
        />

        <PropertiesPanel
          status={form.status}
          priority={form.priority}
          dueDate={form.due_date}
          tags={form.tags}
          folderId={form.folder_id}
          onChange={(patch) =>
            setForm((f) => ({
              ...f,
              ...(patch.status !== undefined && { status: patch.status }),
              ...(patch.priority !== undefined && { priority: patch.priority }),
              ...(patch.dueDate !== undefined && { due_date: patch.dueDate ?? '' }),
              ...(patch.tags !== undefined && { tags: patch.tags }),
              ...(patch.folderId !== undefined && { folder_id: patch.folderId }),
            }))
          }
        />

        <div className="h-64">
          <BlockEditor
            value={form.blocks}
            onChange={(json) => setForm((f) => ({ ...f, blocks: json }))}
            placeholder="Start writing…"
          />
        </div>

        <div className="flex justify-end gap-2 pt-1">
          <Button type="button" variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={isPending}>
            {isEdit ? 'Save changes' : 'Create note'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}

export function NotesPage() {
  const [selectedFolder, setSelectedFolder] = useState<string | undefined>(undefined)
  const [search, setSearch] = useState('')
  const [semanticMode, setSemanticMode] = useState(false)
  const [newFolderName, setNewFolderName] = useState('')
  const [showNewFolder, setShowNewFolder] = useState(false)
  const [view, setView] = useViewPreference('notes_view')
  const [createOpen, setCreateOpen] = useState(false)
  const [editNote, setEditNote] = useState<Note | null>(null)

  const { data: folders = [] } = useFolders()
  const { data: notes = [], isLoading } = useNotes(selectedFolder)
  const { data: searchResults } = useSearchNotes(semanticMode ? '' : search)
  const { data: semanticResults } = useSemanticSearch(semanticMode ? search : '')
  const createNote = useCreateNote()
  const updateNote = useUpdateNote()
  const deleteNote = useDeleteNote()
  const createFolder = useCreateFolder()
  const deleteFolder = useDeleteFolder()

  const displayNotes = useMemo(() => {
    if (search.length > 1) {
      return semanticMode
        ? (semanticResults ?? []).map((r) => r.note)
        : (searchResults ?? [])
    }
    return notes
  }, [search, semanticMode, semanticResults, searchResults, notes])

  const handleNewFolder = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newFolderName.trim()) return
    createFolder.mutate(
      { name: newFolderName.trim() },
      {
        onSuccess: () => {
          setNewFolderName('')
          setShowNewFolder(false)
        },
      },
    )
  }

  const handleCreate = (data: NoteInput) => {
    createNote.mutate(data, { onSuccess: () => setCreateOpen(false) })
  }

  const handleUpdate = (data: NoteInput) => {
    if (!editNote) return
    updateNote.mutate({ id: editNote.id, data }, { onSuccess: () => setEditNote(null) })
  }

  const handleMove = useCallback(
    (itemId: string, _sourceColumnId: string, targetColumnId: string) => {
      if (!VALID_STATUSES.includes(targetColumnId as NoteStatus)) return
      const note = displayNotes.find((n) => n.id === itemId)
      if (!note || note.status === targetColumnId) return
      updateNote.mutate({ id: note.id, data: { status: targetColumnId as NoteStatus } })
    },
    [displayNotes, updateNote],
  )

  const columns = useMemo(
    () => [
      { id: 'draft', title: 'Draft', items: displayNotes.filter((n) => n.status === 'draft') },
      { id: 'in_progress', title: 'In Progress', items: displayNotes.filter((n) => n.status === 'in_progress') },
      { id: 'published', title: 'Published', items: displayNotes.filter((n) => n.status === 'published') },
    ],
    [displayNotes],
  )

  const renderCard = useCallback(
    (note: Note) => {
      const preview = blocksToText(note.blocks)
      const dueDate = formatDate(note.due_date)
      return (
        <div className="space-y-2">
          <div className="flex items-start justify-between gap-2">
            <button
              type="button"
              onClick={() => setEditNote(note)}
              className="text-left text-sm font-medium text-text-secondary hover:text-text-primary line-clamp-2"
            >
              {note.title || 'Untitled'}
            </button>
            <button
              type="button"
              onClick={() => deleteNote.mutate(note.id)}
              className="text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100 focus-visible:opacity-100 shrink-0"
              aria-label="Delete note"
            >
              <Trash2 size={14} />
            </button>
          </div>
          {preview && <p className="text-xs text-text-muted line-clamp-2">{preview}</p>}
          <div className="flex items-center gap-2 flex-wrap">
            <Badge label={note.priority} variant={priorityVariant(note.priority)} />
            {dueDate && <span className="text-xs text-text-muted">{dueDate}</span>}
          </div>
        </div>
      )
    },
    [deleteNote],
  )

  const tableColumns = useMemo(
    () => [
      {
        key: 'title',
        header: 'Title',
        render: (note: Note) => (
          <button
            type="button"
            onClick={() => setEditNote(note)}
            className="text-left text-sm font-medium text-text-secondary hover:text-text-primary"
          >
            {note.title || 'Untitled'}
          </button>
        ),
      },
      {
        key: 'status',
        header: 'Status',
        sortable: true,
        sortValue: (note: Note) => note.status,
        render: (note: Note) => (
          <Badge label={statusLabel(note.status)} variant={statusVariant(note.status)} />
        ),
      },
      {
        key: 'priority',
        header: 'Priority',
        sortable: true,
        sortValue: (note: Note) => note.priority,
        render: (note: Note) => <Badge label={note.priority} variant={priorityVariant(note.priority)} />,
      },
      {
        key: 'due_date',
        header: 'Due date',
        sortable: true,
        sortValue: (note: Note) => note.due_date ?? '',
        render: (note: Note) => (
          <span className="text-text-secondary">{formatDate(note.due_date) ?? '-'}</span>
        ),
      },
      {
        key: 'updated_at',
        header: 'Updated',
        sortable: true,
        sortValue: (note: Note) => note.updated_at,
        render: (note: Note) => (
          <span className="text-text-secondary">{note.updated_at.slice(0, 10)}</span>
        ),
      },
      {
        key: 'actions',
        header: '',
        render: (note: Note) => (
          <button
            type="button"
            onClick={() => deleteNote.mutate(note.id)}
            className="text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
            aria-label="Delete note"
          >
            <Trash2 size={14} />
          </button>
        ),
      },
    ],
    [deleteNote],
  )

  const isModalOpen = createOpen || editNote !== null
  const folderLabel = selectedFolder ? folders.find((f) => f.id === selectedFolder)?.name : 'All notes'

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Sidebar */}
      <aside className="w-56 shrink-0 bg-surface-base border-r border-border flex flex-col">
        <div className="p-3 border-b border-border flex items-center justify-between">
          <span className="text-xs font-semibold text-text-muted uppercase tracking-wider">Notes</span>
          <button
            onClick={() => setCreateOpen(true)}
            className="text-text-muted hover:text-accent transition-colors"
            title="New note"
          >
            <Plus size={16} />
          </button>
        </div>

        {/* Search */}
        <div className="px-3 py-2 border-b border-border space-y-1.5">
          <div className="relative">
            <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-text-muted pointer-events-none" />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={semanticMode ? 'Semantic search…' : 'Search…'}
              className="w-full bg-surface-elevated border border-border rounded-md pl-8 pr-2 py-1.5 text-xs text-text-primary placeholder-text-muted focus:outline-none focus:border-accent transition-colors"
            />
          </div>
          <button
            onClick={() => setSemanticMode((m) => !m)}
            className={`flex items-center gap-1.5 text-xs px-2 py-1 rounded transition-colors ${
              semanticMode
                ? 'bg-accent/10 text-accent'
                : 'text-text-muted hover:text-text-secondary'
            }`}
          >
            <Sparkles size={12} />
            Semantic
          </button>
        </div>

        {/* All notes */}
        <button
          onClick={() => {
            setSelectedFolder(undefined)
            setSearch('')
          }}
          className={`text-left px-3 py-2 text-sm transition-colors ${
            selectedFolder === undefined && !search
              ? 'text-text-primary bg-surface-elevated'
              : 'text-text-muted hover:text-text-secondary'
          }`}
        >
          All notes
        </button>

        {/* Folders */}
        <div className="flex-1 overflow-y-auto">
          <div className="px-3 py-1 flex items-center justify-between">
            <span className="text-xs text-text-muted uppercase tracking-wider">Folders</span>
            <button
              onClick={() => setShowNewFolder((v) => !v)}
              className="text-text-muted hover:text-text-secondary transition-colors"
            >
              <Plus size={14} />
            </button>
          </div>

          {showNewFolder && (
            <form onSubmit={handleNewFolder} className="px-3 pb-2">
              <input
                autoFocus
                value={newFolderName}
                onChange={(e) => setNewFolderName(e.target.value)}
                placeholder="Folder name"
                className="w-full bg-surface-elevated border border-border rounded px-2 py-1.5 text-xs text-text-primary placeholder-text-muted focus:outline-none focus:border-accent transition-colors"
              />
            </form>
          )}

          {folders.map((f) => (
            <div key={f.id} className="group flex items-center">
              <button
                onClick={() => {
                  setSelectedFolder(f.id)
                  setSearch('')
                }}
                className={`flex-1 text-left px-3 py-1.5 text-sm transition-colors flex items-center gap-2 ${
                  selectedFolder === f.id
                    ? 'text-text-primary bg-surface-elevated'
                    : 'text-text-muted hover:text-text-secondary'
                }`}
              >
                <Folder size={14} className="shrink-0" />
                {f.name}
              </button>
              <button
                onClick={() => deleteFolder.mutate(f.id)}
                className="opacity-0 group-hover:opacity-100 pr-2 text-text-muted hover:text-red-400 transition-colors"
                aria-label="Delete folder"
              >
                <X size={14} />
              </button>
            </div>
          ))}
        </div>
      </aside>

      {/* Main area */}
      <main className="flex-1 flex flex-col min-w-0 bg-surface-base p-6 overflow-hidden">
        <PageHeader
          title={search ? `Results for "${search}"` : folderLabel ?? 'All notes'}
          description={`${displayNotes.length} note${displayNotes.length === 1 ? '' : 's'}`}
          actions={
            <div className="flex items-center gap-2">
              <ViewToggle value={view} onChange={setView} />
              <Button type="button" onClick={() => setCreateOpen(true)} icon={<Plus size={16} />}>
                New note
              </Button>
            </div>
          }
        />

        <div className="flex-1 overflow-y-auto min-h-0">
          {isLoading && <p className="text-text-muted text-sm">Loading notes…</p>}

          {!isLoading && displayNotes.length === 0 && (
            <EmptyState
              icon={<FileText size={44} />}
              title={search ? 'No results' : 'No notes yet'}
              description={search ? 'Try a different search term.' : 'Create a note to get started.'}
              action={
                !search ? (
                  <Button type="button" onClick={() => setCreateOpen(true)} icon={<Plus size={16} />}>
                    New note
                  </Button>
                ) : undefined
              }
            />
          )}

          {!isLoading && displayNotes.length > 0 &&
            (view === 'kanban' ? (
              <KanbanBoard
                columns={columns}
                getItemId={(note) => note.id}
                renderCard={renderCard}
                onMove={handleMove}
              />
            ) : (
              <DataTable
                columns={tableColumns}
                data={displayNotes}
                getRowKey={(note) => note.id}
              />
            ))}
        </div>
      </main>

      {/* Create / Edit modal */}
      <NoteModal
        key={editNote?.id ?? 'new'}
        open={isModalOpen}
        onClose={() => {
          setCreateOpen(false)
          setEditNote(null)
        }}
        onSave={editNote ? handleUpdate : handleCreate}
        note={editNote}
        folderId={selectedFolder}
        isPending={createNote.isPending || updateNote.isPending}
      />
    </div>
  )
}
