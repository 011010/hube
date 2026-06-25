import type { SelectHTMLAttributes } from 'react'

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string
  options: { value: string; label: string }[]
  placeholder?: string
}

export function Select({ label, options, placeholder, className = '', ...props }: SelectProps) {
  return (
    <div className="flex flex-col gap-1.5">
      {label && <label className="text-sm text-(--color-text-secondary)">{label}</label>}
      <select
        className={`w-full px-3 py-2 rounded-lg bg-(--color-surface-card) border border-(--color-border) text-(--color-text-primary) text-sm focus:outline-none focus:ring-2 focus:ring-(--color-accent)/40 transition-all ${className}`}
        {...props}
      >
        {placeholder && <option value="">{placeholder}</option>}
        {options.map(opt => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
    </div>
  )
}
