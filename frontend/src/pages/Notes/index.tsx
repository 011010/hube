import { useState, useMemo, useCallback } from 'react'
import { Plus, Trash2, Folder, Sparkles, Search, X, FileText, Network, PanelLeft } from 'lucide-react'
import {
  useNotes, useNote, useCreateNote, useUpdateNote, useDeleteNote,
  useFolders, useCreateFolder, useDeleteFolder, useSearchNotes, useSemanticSearch,
} from '../../hooks/useNotes'
import { useViewPreference } from '../../hooks/useViewPreference'
import { useStatusKanbanMove } from '../../hooks/useStatusKanbanMove'
import { blocksToText } from '../../utils/blocks'
import { statusLabel, statusVariant } from '../../utils/note'
import { EmptyState } from '../../components/molecules/EmptyState'
import { PageHeader } from '../../components/molecules/PageHeader'
import { ViewToggle } from '../../components/molecules/ViewToggle'
import { KanbanBoard } from '../../components/organisms/KanbanBoard'
import { DataTable } from '../../components/organisms/DataTable'
import { GraphView } from '../../components/organisms/GraphView'
import { NoteEditorPanel } from '../../components/organisms/NoteEditorPanel'
import { Button } from '../../components/atoms/Button'
import { Badge } from '../../components/atoms/Badge'
import { IconButton } from '../../components/atoms/IconButton'
import type { Note, NoteStatus } from '../../types'

type MainTab = 'editor' | 'graph'

const VALID_STATUSES: NoteStatus[] = ['draft', 'in_progress', 'published']

export function NotesPage() {
  const [selectedFolder, setSelectedFolder] = useState<string | undefined>(undefined)
  const [selectedNoteId, setSelectedNoteId] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [semanticMode, setSemanticMode] = useState(false)
  const [newFolderName, setNewFolderName] = useState('')
  const [showNewFolder, setShowNewFolder] = useState(false)
  const [view, setView] = useViewPreference('notes_view')
  const [mainTab, setMainTab] = useState<MainTab>('editor')
  const [panelOpen, setPanelOpen] = useState(false)

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

  // Prefer the list entry so edits render without waiting for a refetch, and
  // fall back to a direct fetch for notes outside the current list (active
  // search, another folder, or a note that was just created).
  const { data: fetchedNote } = useNote(selectedNoteId)
  const selectedNote = useMemo(
    () =>
      displayNotes.find((n) => n.id === selectedNoteId) ??
      (fetchedNote?.id === selectedNoteId ? fetchedNote : null),
    [displayNotes, selectedNoteId, fetchedNote],
  )
  const showEditor = mainTab === 'editor' && selectedNote !== null

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

  const handleNewNote = () => {
    createNote.mutate(
      {
        title: 'Untitled',
        blocks: '',
        status: 'draft',
        tags: [],
        folder_id: selectedFolder ?? null,
      },
      { onSuccess: (note) => setSelectedNoteId(note.id) },
    )
  }

  const handleDelete = useCallback(
    (id: string) => {
      deleteNote.mutate(id)
      setSelectedNoteId((current) => (current === id ? null : current))
    },
    [deleteNote],
  )

  const handleMove = useStatusKanbanMove(displayNotes, VALID_STATUSES, (id, data) => updateNote.mutate({ id, data }))

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
      const preview = note.blocks.trim() ? blocksToText(note.blocks) : note.content
      return (
        <div className="space-y-2">
          <div className="flex items-start justify-between gap-2">
            <button
              type="button"
              onClick={() => setSelectedNoteId(note.id)}
              className={`text-left text-sm font-medium line-clamp-2 transition-colors ${
                note.id === selectedNoteId
                  ? 'text-accent'
                  : 'text-text-secondary hover:text-text-primary'
              }`}
            >
              {note.title || 'Untitled'}
            </button>
            <button
              type="button"
              onClick={() => handleDelete(note.id)}
              className="text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100 focus-visible:opacity-100 shrink-0"
              aria-label="Delete note"
            >
              <Trash2 size={14} />
            </button>
          </div>
          {/* Same rule as the table: next to the editor, only the name fits. */}
          {!showEditor && preview && (
            <p className="text-xs text-text-muted line-clamp-2">{preview}</p>
          )}
          {!showEditor && note.tags.length > 0 && (
            <div className="flex items-center gap-1.5 flex-wrap">
              {note.tags.map((tag) => (
                <span key={tag} className="text-xs text-text-muted bg-surface-elevated rounded px-1.5 py-0.5">
                  {tag}
                </span>
              ))}
            </div>
          )}
        </div>
      )
    },
    [handleDelete, selectedNoteId, showEditor],
  )

  const tableColumns = useMemo(() => {
    const titleColumn = {
      key: 'title',
      header: 'Title',
      render: (note: Note) => (
        <button
          type="button"
          onClick={() => setSelectedNoteId(note.id)}
          className={`text-left text-sm font-medium transition-colors break-words ${
            note.id === selectedNoteId
              ? 'text-accent'
              : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          {note.title || 'Untitled'}
        </button>
      ),
    }

    // Alongside the editor the list is a narrow companion: any column past
    // the name is unreadable in the width that is left.
    if (showEditor) return [titleColumn]

    return [
      titleColumn,
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
        key: 'tags',
        header: 'Tags',
        render: (note: Note) => (
          <span className="text-text-secondary">{note.tags.length > 0 ? note.tags.join(', ') : '-'}</span>
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
            onClick={() => handleDelete(note.id)}
            className="text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
            aria-label="Delete note"
          >
            <Trash2 size={14} />
          </button>
        ),
      },
    ]
  }, [handleDelete, selectedNoteId, showEditor])

  const folderLabel = selectedFolder ? folders.find((f) => f.id === selectedFolder)?.name : 'All notes'

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Notes panel backdrop (mobile only) */}
      {panelOpen && (
        <div
          className="fixed inset-0 z-20 bg-black/50 md:hidden"
          onClick={() => setPanelOpen(false)}
          aria-hidden="true"
        />
      )}

      {/* Folders panel */}
      <aside
        className={`fixed inset-y-0 left-0 z-30 w-56 shrink-0 bg-surface-base border-r border-border flex flex-col transition-transform duration-200 ease-out md:static md:z-auto md:translate-x-0 ${
          panelOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="p-3 border-b border-border flex items-center justify-between">
          <span className="text-xs font-semibold text-text-muted uppercase tracking-wider">Notes</span>
          <button
            onClick={handleNewNote}
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
            setPanelOpen(false)
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
                  setPanelOpen(false)
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
              <IconButton
                icon={<PanelLeft size={16} />}
                aria-label="Toggle notes panel"
                variant="ghost"
                className="md:hidden"
                onClick={() => setPanelOpen((v) => !v)}
              />
              <div role="group" aria-label="Notes main view" className="flex items-center bg-surface-elevated border border-border rounded-lg p-0.5">
                <IconButton
                  icon={<FileText size={16} />}
                  aria-label="Editor"
                  aria-pressed={mainTab === 'editor'}
                  variant={mainTab === 'editor' ? 'primary' : 'ghost'}
                  onClick={() => setMainTab('editor')}
                />
                <IconButton
                  icon={<Network size={16} />}
                  aria-label="Graph"
                  aria-pressed={mainTab === 'graph'}
                  variant={mainTab === 'graph' ? 'primary' : 'ghost'}
                  onClick={() => setMainTab('graph')}
                />
              </div>
              {mainTab === 'editor' && <ViewToggle value={view} onChange={setView} />}
              {mainTab === 'editor' && (
                <Button type="button" onClick={handleNewNote} icon={<Plus size={16} />}>
                  New note
                </Button>
              )}
            </div>
          }
        />

        {/* List on the left, selected note on the right */}
        <div className="flex-1 flex min-h-0 gap-6">
          <div
            className={`flex-1 min-w-0 overflow-y-auto ${
              showEditor ? 'hidden lg:block lg:w-80 lg:flex-none' : ''
            }`}
          >
            {mainTab === 'graph' ? (
              <GraphView />
            ) : (
              <>
                {isLoading && <p className="text-text-muted text-sm">Loading notes…</p>}

                {!isLoading && displayNotes.length === 0 && (
                  <EmptyState
                    icon={<FileText size={44} />}
                    title={search ? 'No results' : 'No notes yet'}
                    description={search ? 'Try a different search term.' : 'Create a note to get started.'}
                    action={
                      !search ? (
                        <Button type="button" onClick={handleNewNote} icon={<Plus size={16} />}>
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
                      fitContainer={showEditor}
                    />
                  ))}
              </>
            )}
          </div>

          {showEditor && selectedNote && (
            <div className="flex-1 min-w-0 lg:border-l lg:border-border lg:pl-6">
              <NoteEditorPanel
                key={selectedNote.id}
                note={selectedNote}
                folders={folders}
                onSave={(data) => updateNote.mutate({ id: selectedNote.id, data })}
                onDelete={() => handleDelete(selectedNote.id)}
                onClose={() => setSelectedNoteId(null)}
              />
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
