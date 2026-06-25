import { useState, useEffect, useRef, useCallback } from 'react'
import { Plus, Trash2, Folder, Sparkles, Search, X, FileText } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import {
  useNotes, useNote, useCreateNote, useUpdateNote, useDeleteNote,
  useFolders, useCreateFolder, useDeleteFolder, useSearchNotes, useSemanticSearch,
} from '../../hooks/useNotes'
import { EmptyState } from '../../components/molecules/EmptyState'

type ViewMode = 'edit' | 'preview' | 'split'

export function NotesPage() {
  const [selectedFolder, setSelectedFolder] = useState<string | undefined>(undefined)
  const [selectedNoteId, setSelectedNoteId] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [semanticMode, setSemanticMode] = useState(false)
  const [viewMode, setViewMode] = useState<ViewMode>('split')
  const [newFolderName, setNewFolderName] = useState('')
  const [showNewFolder, setShowNewFolder] = useState(false)

  const { data: folders = [] } = useFolders()
  const { data: notes = [] } = useNotes(selectedFolder)
  const { data: searchResults } = useSearchNotes(semanticMode ? '' : search)
  const { data: semanticResults } = useSemanticSearch(semanticMode ? search : '')
  const { data: activeNote } = useNote(selectedNoteId)
  const createNote = useCreateNote()
  const updateNote = useUpdateNote()
  const deleteNote = useDeleteNote()
  const createFolder = useCreateFolder()
  const deleteFolder = useDeleteFolder()

  const [draft, setDraft] = useState<{ title: string; content: string }>({ title: '', content: '' })
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Sync draft when active note changes
  useEffect(() => {
    if (activeNote) {
      setDraft({ title: activeNote.title, content: activeNote.content })
    }
  }, [activeNote?.id])

  // Auto-save with debounce
  const scheduleSave = useCallback((title: string, content: string) => {
    if (!selectedNoteId) return
    if (saveTimer.current) clearTimeout(saveTimer.current)
    saveTimer.current = setTimeout(() => {
      updateNote.mutate({ id: selectedNoteId, data: { title, content } })
    }, 800)
  }, [selectedNoteId, updateNote])

  const handleTitleChange = (title: string) => {
    setDraft(d => ({ ...d, title }))
    scheduleSave(title, draft.content)
  }

  const handleContentChange = (content: string) => {
    setDraft(d => ({ ...d, content }))
    scheduleSave(draft.title, content)
  }

  const handleNewNote = () => {
    createNote.mutate(
      { title: 'Untitled', content: '', folder_id: selectedFolder ?? null, tags: [] },
      { onSuccess: note => setSelectedNoteId(note.id) },
    )
  }

  const handleNewFolder = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newFolderName.trim()) return
    createFolder.mutate({ name: newFolderName.trim() }, {
      onSuccess: () => { setNewFolderName(''); setShowNewFolder(false) },
    })
  }

  const displayNotes = search.length > 1
    ? semanticMode
      ? (semanticResults ?? []).map(r => r.note)
      : (searchResults ?? [])
    : notes

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Sidebar */}
      <aside className="w-56 shrink-0 bg-surface-base border-r border-border flex flex-col">
        <div className="p-3 border-b border-border flex items-center justify-between">
          <span className="text-xs font-semibold text-text-muted uppercase tracking-wider">Notes</span>
          <button
            onClick={handleNewNote}
            className="text-text-muted hover:text-(--color-accent) transition-colors"
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
              onChange={e => setSearch(e.target.value)}
              placeholder={semanticMode ? 'Semantic search…' : 'Search…'}
              className="w-full bg-surface-elevated border border-border rounded-md pl-8 pr-2 py-1.5 text-xs text-text-primary placeholder-text-muted focus:outline-none focus:border-(--color-accent) transition-colors"
            />
          </div>
          <button
            onClick={() => setSemanticMode(m => !m)}
            className={`flex items-center gap-1.5 text-xs px-2 py-1 rounded transition-colors ${
              semanticMode
                ? 'bg-(--color-accent)/10 text-(--color-accent)'
                : 'text-text-muted hover:text-text-secondary'
            }`}
          >
            <Sparkles size={12} />
            Semantic
          </button>
        </div>

        {/* All notes */}
        <button
          onClick={() => { setSelectedFolder(undefined); setSearch('') }}
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
              onClick={() => setShowNewFolder(v => !v)}
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
                onChange={e => setNewFolderName(e.target.value)}
                placeholder="Folder name"
                className="w-full bg-surface-elevated border border-border rounded px-2 py-1.5 text-xs text-text-primary placeholder-text-muted focus:outline-none focus:border-(--color-accent) transition-colors"
              />
            </form>
          )}

          {folders.map(f => (
            <div key={f.id} className="group flex items-center">
              <button
                onClick={() => { setSelectedFolder(f.id); setSearch('') }}
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

      {/* Note list */}
      <div className="w-56 shrink-0 border-r border-border flex flex-col bg-surface-base">
        <div className="p-3 border-b border-border">
          <span className="text-xs text-text-muted">
            {search ? `Results for "${search}"` : selectedFolder ? folders.find(f => f.id === selectedFolder)?.name : 'All notes'}
            {' '}· {displayNotes.length}
          </span>
        </div>
        <ul className="flex-1 overflow-y-auto divide-y divide-border/50">
          {displayNotes.map(n => (
            <li key={n.id}>
              <button
                onClick={() => setSelectedNoteId(n.id)}
                className={`w-full text-left px-3 py-3 transition-colors ${
                  selectedNoteId === n.id ? 'bg-surface-elevated' : 'hover:bg-surface-elevated/50'
                }`}
              >
                <p className="text-sm text-text-primary truncate">{n.title || 'Untitled'}</p>
                <p className="text-xs text-text-muted mt-0.5 truncate">{n.content.slice(0, 60) || 'Empty'}</p>
                <p className="text-xs text-text-muted/60 mt-1">{n.updated_at.slice(0, 10)}</p>
              </button>
            </li>
          ))}
          {displayNotes.length === 0 && (
            <li className="px-3 py-4">
              <EmptyState
                icon={<FileText size={40} />}
                title={search ? 'No results' : 'No notes yet'}
                description={search ? 'Try a different search term.' : 'Create a note to get started.'}
                action={
                  !search ? (
                    <button onClick={handleNewNote} className="text-sm text-(--color-accent) hover:text-(--color-accent-hover) transition-colors">
                      + New note
                    </button>
                  ) : undefined
                }
              />
            </li>
          )}
        </ul>
      </div>

      {/* Editor */}
      <div className="flex-1 flex flex-col min-w-0 bg-surface-base">
        {selectedNoteId && activeNote ? (
          <>
            {/* Toolbar */}
            <div className="flex items-center justify-between px-6 py-3 border-b border-border shrink-0">
              <input
                value={draft.title}
                onChange={e => handleTitleChange(e.target.value)}
                className="flex-1 bg-transparent text-lg font-semibold text-text-primary focus:outline-none placeholder-text-muted"
                placeholder="Untitled"
              />
              <div className="flex items-center gap-1 ml-4">
                {(['edit', 'split', 'preview'] as ViewMode[]).map(m => (
                  <button
                    key={m}
                    onClick={() => setViewMode(m)}
                    className={`px-2 py-1 text-xs rounded capitalize transition-colors ${
                      viewMode === m
                        ? 'bg-(--color-accent) text-white'
                        : 'text-text-muted hover:text-text-secondary'
                    }`}
                  >{m}</button>
                ))}
                <button
                  onClick={() => { deleteNote.mutate(selectedNoteId); setSelectedNoteId(null) }}
                  className="ml-2 text-text-muted hover:text-red-400 transition-colors"
                  aria-label="Delete note"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </div>

            {/* Panes */}
            <div className="flex-1 flex overflow-hidden">
              {(viewMode === 'edit' || viewMode === 'split') && (
                <textarea
                  value={draft.content}
                  onChange={e => handleContentChange(e.target.value)}
                  className={`bg-transparent text-sm text-text-secondary p-6 resize-none focus:outline-none font-mono leading-relaxed placeholder-text-muted ${
                    viewMode === 'split' ? 'w-1/2 border-r border-border' : 'w-full'
                  }`}
                  placeholder="Start writing in Markdown..."
                />
              )}
              {(viewMode === 'preview' || viewMode === 'split') && (
                <div className={`overflow-y-auto p-6 prose prose-invert prose-sm max-w-none ${
                  viewMode === 'split' ? 'w-1/2' : 'w-full'
                }`}>
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    components={{
                      code({ className, children, ...props }) {
                        const match = /language-(\w+)/.exec(className || '')
                        const isBlock = String(children).includes('\n')
                        return match || isBlock ? (
                          <SyntaxHighlighter
                            style={oneDark}
                            language={match?.[1] ?? 'text'}
                            PreTag="div"
                            customStyle={{ borderRadius: '0.5rem', fontSize: '0.8rem' }}
                          >
                            {String(children).replace(/\n$/, '')}
                          </SyntaxHighlighter>
                        ) : (
                          <code className="bg-surface-elevated text-pink-400 px-1 py-0.5 rounded text-xs" {...props}>
                            {children}
                          </code>
                        )
                      },
                    }}
                  >
                    {draft.content || '*Nothing to preview*'}
                  </ReactMarkdown>
                </div>
              )}
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <EmptyState
              icon={<FileText size={44} />}
              title="Select a note"
              description="Choose a note from the list or create a new one."
              action={
                <button
                  onClick={handleNewNote}
                  className="text-sm text-(--color-accent) hover:text-(--color-accent-hover) transition-colors"
                >
                  + New note
                </button>
              }
            />
          </div>
        )}
      </div>
    </div>
  )
}
