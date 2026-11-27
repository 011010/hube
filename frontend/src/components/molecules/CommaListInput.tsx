import { useEffect, useRef, useState } from 'react'

// Shared styling for the compact form fields used by the properties panel and
// the note view editor, so both surfaces stay visually identical.
export const FIELD_CLASS =
  'w-full px-2.5 py-1.5 rounded-lg bg-surface-card border border-border text-text-primary text-sm placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-accent/40 transition-all'

// Text field backed by a comma-separated buffer that parses into a trimmed,
// non-empty string array. The raw text survives while the field is focused so
// the parent's parsed value does not rewrite the buffer mid-keystroke.
export function CommaListInput({
  id,
  value,
  onChange,
  label,
  placeholder,
}: {
  id: string
  value: string[]
  onChange: (items: string[]) => void
  label: string
  placeholder: string
}) {
  const [text, setText] = useState(value.join(', '))
  const editing = useRef(false)

  // Resync only when the field is not being edited, so the parent's parsed
  // value does not rewrite the text mid-keystroke.
  useEffect(() => {
    if (!editing.current) setText(value.join(', '))
  }, [value])

  return (
    <input
      id={id}
      type="text"
      value={text}
      aria-label={label}
      onFocus={() => {
        editing.current = true
      }}
      onBlur={() => {
        editing.current = false
        setText(value.join(', '))
      }}
      onChange={(e) => {
        setText(e.target.value)
        onChange(
          e.target.value
            .split(',')
            .map((item) => item.trim())
            .filter(Boolean),
        )
      }}
      placeholder={placeholder}
      className={FIELD_CLASS}
    />
  )
}
