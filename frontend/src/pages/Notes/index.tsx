import { useState, useEffect, useRef, useCallback } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import {
  useNotes, useNote, useCreateNote, useUpdateNote, useDeleteNote,
  useFolders, useCreateFolder, useDeleteFolder, useSearchNotes, useSemanticSearch,
} from '../../hooks/useNotes'

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
      <aside className="w-56 shrink-0 bg-gray-950 border-r border-gray-800 flex flex-col">
        <div className="p-3 border-b border-gray-800 flex items-center justify-between">
          <span className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Notes</span>
          <button
            onClick={handleNewNote}
            className="text-indigo-400 hover:text-indigo-300 text-lg leading-none"
            title="New note"
          >+</button>
        </div>

        {/* Search */}
        <div className="px-3 py-2 border-b border-gray-800 space-y-1.5">
          <input
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder={semanticMode ? 'Semantic search…' : 'Search…'}
            className="w-full bg-gray-900 border border-gray-700 rounded-md px-2 py-1 text-xs text-white placeholder-gray-600 focus:outline-none focus:border-indigo-500"
          />
          <button
            onClick={() => setSemanticMode(m => !m)}
            className={`text-xs px-2 py-0.5 rounded transition-colors ${semanticMode ? 'bg-indigo-700 text-indigo-200' : 'text-gray-600 hover:text-gray-400'}`}
          >
            ✦ Semantic
          </button>
        </div>

        {/* All notes */}
        <button
          onClick={() => { setSelectedFolder(undefined); setSearch('') }}
          className={`text-left px-3 py-2 text-sm transition-colors ${
            selectedFolder === undefined && !search ? 'text-white bg-gray-800' : 'text-gray-400 hover:text-white'
          }`}
        >
          All notes
        </button>

        {/* Folders */}
        <div className="flex-1 overflow-y-auto">
          <div className="px-3 py-1 flex items-center justify-between">
            <span className="text-xs text-gray-600 uppercase tracking-wider">Folders</span>
            <button onClick={() => setShowNewFolder(v => !v)} className="text-gray-600 hover:text-gray-400 text-xs">+</button>
          </div>

          {showNewFolder && (
            <form onSubmit={handleNewFolder} className="px-3 pb-2">
              <input
                autoFocus
                value={newFolderName}
                onChange={e => setNewFolderName(e.target.value)}
                placeholder="Folder name"
                className="w-full bg-gray-900 border border-gray-700 rounded px-2 py-1 text-xs text-white focus:outline-none focus:border-indigo-500"
              />
            </form>
          )}

          {folders.map(f => (
            <div key={f.id} className="group flex items-center">
              <button
                onClick={() => { setSelectedFolder(f.id); setSearch('') }}
                className={`flex-1 text-left px-3 py-1.5 text-sm transition-colors ${
                  selectedFolder === f.id ? 'text-white bg-gray-800' : 'text-gray-400 hover:text-white'
                }`}
              >
                📁 {f.name}
              </button>
              <button
                onClick={() => deleteFolder.mutate(f.id)}
                className="opacity-0 group-hover:opacity-100 pr-2 text-gray-600 hover:text-red-400 text-xs"
              >✕</button>
            </div>
          ))}
        </div>
      </aside>

      {/* Note list */}
      <div className="w-56 shrink-0 border-r border-gray-800 flex flex-col bg-gray-950">
        <div className="p-3 border-b border-gray-800">
          <span className="text-xs text-gray-500">
            {search ? `Results for "${search}"` : selectedFolder ? folders.find(f => f.id === selectedFolder)?.name : 'All notes'}
            {' '}· {displayNotes.length}
          </span>
        </div>
        <ul className="flex-1 overflow-y-auto divide-y divide-gray-800/50">
          {displayNotes.map(n => (
            <li key={n.id}>
              <button
                onClick={() => setSelectedNoteId(n.id)}
                className={`w-full text-left px-3 py-3 transition-colors ${
                  selectedNoteId === n.id ? 'bg-gray-800' : 'hover:bg-gray-900'
                }`}
              >
                <p className="text-sm text-white truncate">{n.title || 'Untitled'}</p>
                <p className="text-xs text-gray-500 mt-0.5 truncate">{n.content.slice(0, 60) || 'Empty'}</p>
                <p className="text-xs text-gray-700 mt-1">{n.updated_at.slice(0, 10)}</p>
              </button>
            </li>
          ))}
          {displayNotes.length === 0 && (
            <li className="px-3 py-4 text-xs text-gray-600">
              {search ? 'No results.' : 'No notes yet.'}
            </li>
          )}
        </ul>
      </div>

      {/* Editor */}
      <div className="flex-1 flex flex-col min-w-0 bg-gray-950">
        {selectedNoteId && activeNote ? (
          <>
            {/* Toolbar */}
            <div className="flex items-center justify-between px-6 py-3 border-b border-gray-800 shrink-0">
              <input
                value={draft.title}
                onChange={e => handleTitleChange(e.target.value)}
                className="flex-1 bg-transparent text-lg font-semibold text-white focus:outline-none placeholder-gray-600"
                placeholder="Untitled"
              />
              <div className="flex items-center gap-1 ml-4">
                {(['edit', 'split', 'preview'] as ViewMode[]).map(m => (
                  <button
                    key={m}
                    onClick={() => setViewMode(m)}
                    className={`px-2 py-1 text-xs rounded capitalize transition-colors ${
                      viewMode === m ? 'bg-indigo-600 text-white' : 'text-gray-500 hover:text-white'
                    }`}
                  >{m}</button>
                ))}
                <button
                  onClick={() => { deleteNote.mutate(selectedNoteId); setSelectedNoteId(null) }}
                  className="ml-2 text-gray-600 hover:text-red-400 text-xs transition-colors"
                >Delete</button>
              </div>
            </div>

            {/* Panes */}
            <div className="flex-1 flex overflow-hidden">
              {(viewMode === 'edit' || viewMode === 'split') && (
                <textarea
                  value={draft.content}
                  onChange={e => handleContentChange(e.target.value)}
                  className={`bg-transparent text-sm text-gray-200 p-6 resize-none focus:outline-none font-mono leading-relaxed ${
                    viewMode === 'split' ? 'w-1/2 border-r border-gray-800' : 'w-full'
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
                          <code className="bg-gray-800 text-pink-400 px-1 py-0.5 rounded text-xs" {...props}>
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
            <div className="text-center">
              <p className="text-gray-600 text-sm mb-3">Select a note or create a new one</p>
              <button
                onClick={handleNewNote}
                className="text-sm text-indigo-400 hover:text-indigo-300 transition-colors"
              >+ New note</button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
