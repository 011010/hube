import type { ButtonHTMLAttributes, ReactNode } from 'react'

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon: ReactNode
  label?: string
  variant?: 'ghost' | 'primary'
  size?: 'sm' | 'md'
}

const variants = {
  ghost: 'text-(--color-text-muted) hover:text-(--color-text-primary) hover:bg-(--color-surface-card)',
  primary: 'text-white bg-(--color-accent) hover:bg-(--color-accent-hover)',
}

const sizes = {
  sm: 'p-1.5',
  md: 'p-2',
}

export function IconButton({ icon, label, variant = 'ghost', size = 'md', className = '', 'aria-label': ariaLabel, ...props }: IconButtonProps) {
  const accessibleLabel = label ?? ariaLabel
  return (
    <button
      {...props}
      aria-label={accessibleLabel}
      title={accessibleLabel}
      className={`inline-flex items-center justify-center rounded-lg transition-colors ${variants[variant]} ${sizes[size]} ${className}`}
    >
      {icon}
    </button>
  )
}
