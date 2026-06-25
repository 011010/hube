import { useEffect, useRef, type TextareaHTMLAttributes } from 'react'

interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string
  error?: string
  autoResize?: boolean
}

export function Textarea({ label, error, autoResize, className = '', ...props }: TextareaProps) {
  const ref = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    if (!autoResize || !ref.current) return
    const el = ref.current
    const resize = () => {
      el.style.height = 'auto'
      el.style.height = `${el.scrollHeight}px`
    }
    resize()
    el.addEventListener('input', resize)
    return () => el.removeEventListener('input', resize)
  }, [autoResize])

  return (
    <div className="flex flex-col gap-1.5">
      {label && <label className="text-sm text-(--color-text-secondary)">{label}</label>}
      <textarea
        ref={ref}
        className={`w-full px-3 py-2 rounded-lg bg-(--color-surface-card) border border-(--color-border) text-(--color-text-primary) text-sm placeholder:text-(--color-text-muted) focus:outline-none focus:ring-2 focus:ring-(--color-accent)/40 transition-all resize-vertical ${error ? 'border-red-500' : ''} ${className}`}
        {...props}
      />
      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  )
}
