import { forwardRef, type InputHTMLAttributes } from 'react'

// Omit `prefix` to avoid conflict with HTML 'prefix' attribute
interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'prefix'> {
  label?: string
  error?: string
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  function Input({ label, error, className = '', ...props }, ref) {
    return (
      <div className="flex flex-col gap-1.5">
        {label && <label className="text-sm text-(--color-text-secondary)">{label}</label>}
        <input
          ref={ref}
          className={`w-full px-3 py-2 rounded-lg bg-(--color-surface-card) border border-(--color-border) text-(--color-text-primary) text-sm placeholder:text-(--color-text-muted) focus:outline-none focus:ring-2 focus:ring-(--color-accent)/40 transition-all ${error ? 'border-red-500' : ''} ${className}`}
          {...props}
        />
        {error && <p className="text-xs text-red-400">{error}</p>}
      </div>
    )
  }
)
