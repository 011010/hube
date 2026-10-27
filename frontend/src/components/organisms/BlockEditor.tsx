import { useEffect, useRef, type MouseEvent, type ReactNode } from 'react'
import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import { Bold, Italic, Underline as UnderlineIcon, List, ListOrdered, CheckSquare } from 'lucide-react'
import { blocksToText } from '../../utils/blocks'

interface BlockEditorProps {
  value: string
  onChange: (json: string, plainText: string) => void
  placeholder?: string
}

function parseContent(value: string) {
  if (!value.trim()) return { type: 'doc', content: [] }
  try {
    return JSON.parse(value)
  } catch {
    return { type: 'doc', content: [] }
  }
}

interface ToolbarButtonProps {
  onClick: () => void
  active: boolean
  icon: ReactNode
  label: string
}

function ToolbarButton({ onClick, active, icon, label }: ToolbarButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      title={label}
      aria-pressed={active}
      className={`p-1.5 rounded-md transition-colors ${
        active
          ? 'bg-accent text-white'
          : 'text-text-muted hover:text-text-secondary hover:bg-surface-elevated'
      }`}
    >
      {icon}
    </button>
  )
}

export function BlockEditor({ value, onChange, placeholder = 'Start writing…' }: BlockEditorProps) {
  const surfaceRef = useRef<HTMLDivElement>(null)
  const editor = useEditor({
    extensions: [
      StarterKit,
      Placeholder.configure({ placeholder }),
      TaskList,
      TaskItem.configure({ nested: false }),
    ],
    content: parseContent(value),
    editorProps: {
      attributes: {
        'aria-label': 'Note content',
      },
    },
    onUpdate: ({ editor }) => {
      const json = editor.getJSON()
      onChange(JSON.stringify(json), blocksToText(JSON.stringify(json)))
    },
  })

  useEffect(() => {
    if (!editor) return
    const current = JSON.stringify(editor.getJSON())
    if (current !== value) {
      editor.commands.setContent(parseContent(value), { emitUpdate: false })
    }
  }, [editor, value])

  // The blank space below the last line should behave like the document:
  // clicking it puts the caret at the end instead of doing nothing.
  const handleSurfaceClick = (event: MouseEvent<HTMLDivElement>) => {
    if (event.target !== surfaceRef.current) return
    editor?.chain().focus('end').run()
  }

  if (!editor) return null

  return (
    <div className="flex flex-col h-full min-h-0 rounded-xl border border-border bg-surface-elevated overflow-hidden transition-colors focus-within:border-accent/60">
      <div className="flex items-center gap-1 px-2.5 py-1.5 border-b border-border bg-surface-card/60 shrink-0">
        <ToolbarButton
          label="Bold"
          icon={<Bold size={16} />}
          active={editor.isActive('bold')}
          onClick={() => editor.chain().focus().toggleBold().run()}
        />
        <ToolbarButton
          label="Italic"
          icon={<Italic size={16} />}
          active={editor.isActive('italic')}
          onClick={() => editor.chain().focus().toggleItalic().run()}
        />
        <ToolbarButton
          label="Underline"
          icon={<UnderlineIcon size={16} />}
          active={editor.isActive('underline')}
          onClick={() => editor.chain().focus().toggleUnderline().run()}
        />
        <span className="w-px h-5 bg-border mx-1" />
        <ToolbarButton
          label="Bullet list"
          icon={<List size={16} />}
          active={editor.isActive('bulletList')}
          onClick={() => editor.chain().focus().toggleBulletList().run()}
        />
        <ToolbarButton
          label="Ordered list"
          icon={<ListOrdered size={16} />}
          active={editor.isActive('orderedList')}
          onClick={() => editor.chain().focus().toggleOrderedList().run()}
        />
        <ToolbarButton
          label="Task list"
          icon={<CheckSquare size={16} />}
          active={editor.isActive('taskList')}
          onClick={() => editor.chain().focus().toggleTaskList().run()}
        />
      </div>
      <div
        ref={surfaceRef}
        onClick={handleSurfaceClick}
        className="note-editor flex-1 min-h-0 overflow-y-auto"
      >
        <EditorContent editor={editor} className="h-full" />
      </div>
    </div>
  )
}
