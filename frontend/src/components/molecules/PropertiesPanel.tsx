import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { Plus, X } from 'lucide-react'
import { Button } from '../atoms/Button'
import type { Folder, NoteProperty, NoteStatus, PropertyType } from '../../types'

interface PropertiesPanelProps {
  status: NoteStatus
  tags: string[]
  folderId: string | null
  folders: Folder[]
  onChange: (patch: {
    status?: NoteStatus
    tags?: string[]
    folderId?: string | null
  }) => void
  properties: NoteProperty[]
  onPropertiesChange: (properties: NoteProperty[]) => void
}

const STATUS_OPTIONS: { value: NoteStatus; label: string }[] = [
  { value: 'draft', label: 'Draft' },
  { value: 'in_progress', label: 'In progress' },
  { value: 'published', label: 'Published' },
]

const PROPERTY_TYPE_OPTIONS: { value: PropertyType; label: string }[] = [
  { value: 'text', label: 'Text' },
  { value: 'number', label: 'Number' },
  { value: 'date', label: 'Date' },
  { value: 'boolean', label: 'Boolean' },
  { value: 'list', label: 'List' },
]

const FIELD_CLASS =
  'w-full px-2.5 py-1.5 rounded-lg bg-surface-card border border-border text-text-primary text-sm placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-accent/40 transition-all'

const LABEL_CLASS = 'text-[11px] font-medium uppercase tracking-wide text-text-muted'

const KEY_PATTERN = /^[a-z][a-z0-9_-]{0,63}$/

// These collide with frontmatter fields the backend owns.
const RESERVED_KEYS = new Set(['title', 'status', 'tags', 'created_at', 'updated_at'])

function todayLocal(): string {
  const now = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
}

function defaultValueFor(type: PropertyType): NoteProperty['value'] {
  switch (type) {
    case 'number':
      return 0
    case 'date':
      return todayLocal()
    case 'boolean':
      return false
    case 'list':
      return []
    case 'text':
      return ''
  }
}

function validateNewKey(rawKey: string, properties: NoteProperty[]): string | null {
  const key = rawKey.trim().toLowerCase()
  if (!key) return 'Enter a key for the property.'
  if (!KEY_PATTERN.test(key)) {
    return 'Use lowercase letters, numbers, "-" or "_", starting with a letter (max 64 characters).'
  }
  if (RESERVED_KEYS.has(key)) return `"${key}" is reserved.`
  if (properties.some((property) => property.key === key)) {
    return `"${key}" already exists.`
  }
  return null
}

function CommaListInput({
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

function NumberInput({
  id,
  value,
  onChange,
  label,
}: {
  id: string
  value: number
  onChange: (value: number) => void
  label: string
}) {
  const [text, setText] = useState(String(value))
  const editing = useRef(false)

  // Keep partial input (for example "-") in the field while the user types.
  useEffect(() => {
    if (!editing.current) setText(String(value))
  }, [value])

  return (
    <input
      id={id}
      type="text"
      inputMode="decimal"
      autoComplete="off"
      value={text}
      aria-label={label}
      onFocus={() => {
        editing.current = true
      }}
      onBlur={() => {
        editing.current = false
        setText(String(value))
      }}
      onChange={(e) => {
        const raw = e.target.value
        setText(raw)
        if (raw !== '' && Number.isFinite(Number(raw))) onChange(Number(raw))
      }}
      className={FIELD_CLASS}
    />
  )
}

export function PropertiesPanel({
  status,
  tags,
  folderId,
  folders,
  onChange,
  properties,
  onPropertiesChange,
}: PropertiesPanelProps) {
  const [newKey, setNewKey] = useState('')
  const [newType, setNewType] = useState<PropertyType>('text')
  const [keyError, setKeyError] = useState<string | null>(null)

  const setValue = (key: string, value: NoteProperty['value']) => {
    onPropertiesChange(
      properties.map((property) =>
        property.key === key ? { ...property, value } : property,
      ),
    )
  }

  const removeProperty = (key: string) => {
    onPropertiesChange(properties.filter((property) => property.key !== key))
  }

  const handleAdd = (e: FormEvent) => {
    e.preventDefault()
    const error = validateNewKey(newKey, properties)
    if (error) {
      setKeyError(error)
      return
    }
    const key = newKey.trim().toLowerCase()
    onPropertiesChange([...properties, { key, type: newType, value: defaultValueFor(newType) }])
    setNewKey('')
    setNewType('text')
    setKeyError(null)
  }

  const renderValueInput = (property: NoteProperty) => {
    const inputID = `note-property-value-${property.key}`
    const ariaLabel = `Value for property ${property.key}`

    switch (property.type) {
      case 'number':
        return (
          <NumberInput
            id={inputID}
            value={typeof property.value === 'number' ? property.value : 0}
            onChange={(value) => setValue(property.key, value)}
            label={ariaLabel}
          />
        )
      case 'date':
        return (
          <input
            id={inputID}
            type="date"
            value={typeof property.value === 'string' ? property.value : ''}
            onChange={(e) =>
              setValue(
                property.key,
                e.target.value === ''
                  ? (typeof property.value === 'string' ? property.value : '')
                  : e.target.value,
              )
            }
            aria-label={ariaLabel}
            className={FIELD_CLASS}
          />
        )
      case 'boolean':
        return (
          <div className="flex h-[34px] items-center">
            <input
              id={inputID}
              type="checkbox"
              checked={property.value === true}
              onChange={(e) => setValue(property.key, e.target.checked)}
              aria-label={ariaLabel}
              className="h-4 w-4 rounded border-border bg-surface-elevated text-(--color-accent) accent-(--color-accent)"
            />
          </div>
        )
      case 'list':
        return (
          <CommaListInput
            id={inputID}
            value={Array.isArray(property.value) ? property.value : []}
            onChange={(items) => setValue(property.key, items)}
            label={ariaLabel}
            placeholder="Comma-separated"
          />
        )
      case 'text':
        return (
          <input
            id={inputID}
            type="text"
            value={typeof property.value === 'string' ? property.value : ''}
            onChange={(e) => setValue(property.key, e.target.value)}
            aria-label={ariaLabel}
            className={FIELD_CLASS}
          />
        )
    }
  }

  // One compact row instead of three stacked blocks: the writing surface is the
  // point of this panel, metadata should not eat a third of it.
  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-end gap-3">
        <div className="flex flex-col gap-1 flex-1 min-w-28">
          <label htmlFor="note-status" className={LABEL_CLASS}>
            Status
          </label>
          <select
            id="note-status"
            value={status}
            onChange={(e) => onChange({ status: e.target.value as NoteStatus })}
            className={FIELD_CLASS}
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-1 flex-1 min-w-28">
          <label htmlFor="note-folder" className={LABEL_CLASS}>
            Folder
          </label>
          <select
            id="note-folder"
            value={folderId ?? ''}
            onChange={(e) => onChange({ folderId: e.target.value || null })}
            className={FIELD_CLASS}
          >
            <option value="">No folder</option>
            {folders.map((folder) => (
              <option key={folder.id} value={folder.id}>
                {folder.name}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-1 flex-[2] min-w-36">
          <label htmlFor="note-tags" className={LABEL_CLASS}>
            Tags
          </label>
          <CommaListInput
            id="note-tags"
            value={tags}
            onChange={(tags) => onChange({ tags })}
            label="Tags"
            placeholder="Add tags…"
          />
        </div>
      </div>

      <div
        role="group"
        aria-labelledby="note-properties-heading"
        className="flex flex-col gap-2 border-t border-border pt-3"
      >
        <span id="note-properties-heading" className={LABEL_CLASS}>
          Properties
        </span>

        {properties.map((property) => {
          const inputID = `note-property-value-${property.key}`
          return (
            <div key={property.key} className="flex items-end gap-1.5">
              <div className="flex flex-col gap-1 flex-1 min-w-0">
                <div className="flex items-baseline justify-between gap-2">
                  <label htmlFor={inputID} className={LABEL_CLASS}>
                    {property.key}
                  </label>
                  <span className="text-[10px] uppercase tracking-wide text-text-muted/70">
                    {property.type}
                  </span>
                </div>
                {renderValueInput(property)}
              </div>
              <button
                type="button"
                onClick={() => removeProperty(property.key)}
                aria-label={`Remove property ${property.key}`}
                title={`Remove property ${property.key}`}
                className="p-1.5 mb-px rounded-lg text-text-muted hover:text-red-400 hover:bg-surface-card transition-colors"
              >
                <X size={14} />
              </button>
            </div>
          )
        })}

        <form onSubmit={handleAdd} className="flex flex-wrap items-end gap-2">
          <div className="flex flex-col gap-1 flex-1 min-w-30">
            <label htmlFor="note-new-property-key" className={LABEL_CLASS}>
              Key
            </label>
            <input
              id="note-new-property-key"
              type="text"
              value={newKey}
              onChange={(e) => {
                setNewKey(e.target.value)
                setKeyError(null)
              }}
              placeholder="e.g. rating"
              aria-label="New property key"
              aria-invalid={Boolean(keyError)}
              aria-describedby={keyError ? 'note-new-property-error' : undefined}
              className={FIELD_CLASS}
            />
          </div>

          <div className="flex flex-col gap-1 w-28 shrink-0">
            <label htmlFor="note-new-property-type" className={LABEL_CLASS}>
              Type
            </label>
            <select
              id="note-new-property-type"
              value={newType}
              onChange={(e) => setNewType(e.target.value as PropertyType)}
              aria-label="New property type"
              className={FIELD_CLASS}
            >
              {PROPERTY_TYPE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>

          <Button type="submit" variant="ghost" size="sm" icon={<Plus size={14} />}>
            Add property
          </Button>
        </form>

        {keyError && (
          <p id="note-new-property-error" role="alert" className="text-xs text-red-400">
            {keyError}
          </p>
        )}
      </div>
    </div>
  )
}
