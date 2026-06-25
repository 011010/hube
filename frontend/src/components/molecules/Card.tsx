import type { HTMLAttributes } from 'react'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  padding?: 'compact' | 'normal'
  hover?: boolean
}

export function Card({
  padding = 'normal',
  hover = false,
  className = '',
  children,
  ...props
}: CardProps) {
  return (
    <div
      className={`bg-surface-elevated border border-border rounded-xl ${
        padding === 'normal' ? 'p-5' : 'p-3'
      } ${hover ? 'hover:bg-surface-card transition-colors cursor-pointer' : ''} ${className}`}
      {...props}
    >
      {children}
    </div>
  )
}
