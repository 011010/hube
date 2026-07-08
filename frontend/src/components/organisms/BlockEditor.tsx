import { useEffect } from 'react'
import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import Underline from '@tiptap/extension-underline'
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

const TASK_LIST_CSS = `
  .ProseMirror ul[data-type="taskList"] {
    list-style: none;
    padding: 0;
  }
  .ProseMirror ul[data-type="taskList"] li {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
  }
  .ProseMirror ul[data-type="taskList"] li > label {
    flex-shrink: 0;
    margin-top: 0.25rem;
  }
  .ProseMirror ul[data-type="taskList"] li > div {
    flex: 1;
  }
  .ProseMirror ul[data-type="taskList"] input[type="checkbox"] {
    cursor: pointer;
  }
`

export function BlockEditor({ value, onChange, placeholder = 'Start writing…' }: BlockEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      Placeholder.configure({ placeholder }),
      TaskList,
      TaskItem.configure({ nested: false }),
      Underline,
    ],
    content: parseContent(value),
    onUpdate: ({ editor }) => {
      const json = editor.getJSON()
      onChange(JSON.stringify(json), blocksToText(JSON.stringify(json)))
    },
  })

  useEffect(() => {
    if (!editor) return
    const current = JSON.stringify(editor.getJSON())
    if (current !== value) {
      editor.commands.setContent(parseContent(value), false)
    }
  }, [editor, value])

  if (!editor) return null

  const ToolbarButton = ({
    onClick,
    active,
    icon,
    label,
  }: {
    onClick: () => void
    active: boolean
    icon: React.ReactNode
    label: string
  }) => (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      title={label}
      aria-pressed={active}
      className={`p-1.5 rounded-md transition-colors ${
        active
          ? 'bg-accent text-white'
          : 'text-text-muted hover:text-text-secondary hover:bg-surface-card'
      }`}
    >
      {icon}
    </button>
  )

  return (
    <div className="flex flex-col h-full border border-border rounded-xl overflow-hidden bg-surface-elevated">
      <style>{TASK_LIST_CSS}</style>
      <div className="flex items-center gap-1 px-3 py-2 border-b border-border bg-surface-card">
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
      <div className="flex-1 overflow-y-auto p-4">
        <EditorContent
          editor={editor}
          className="prose prose-invert prose-sm max-w-none h-full focus:outline-none"
        />
      </div>
    </div>
  )
}
