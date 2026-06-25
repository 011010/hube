import type { ButtonHTMLAttributes, ReactNode } from 'react'

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon: ReactNode
  label: string
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

export function IconButton({ icon, label, variant = 'ghost', size = 'md', className = '', ...props }: IconButtonProps) {
  return (
    <button
      aria-label={label}
      title={label}
      className={`inline-flex items-center justify-center rounded-lg transition-colors ${variants[variant]} ${sizes[size]} ${className}`}
      {...props}
    >
      {icon}
    </button>
  )
}
