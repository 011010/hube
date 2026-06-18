import type { ButtonHTMLAttributes } from 'react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'ghost' | 'danger'
  size?: 'sm' | 'md'
}

const variants = {
  primary: 'bg-(--color-accent) hover:bg-(--color-accent-hover) text-white',
  ghost: 'bg-transparent hover:bg-gray-800 text-gray-300 dark:hover:bg-gray-800',
  danger: 'bg-red-900 hover:bg-red-800 text-red-200',
}

const sizes = {
  sm: 'px-3 py-1.5 text-sm',
  md: 'px-4 py-2 text-sm',
}

export function Button({ variant = 'primary', size = 'md', className = '', ...props }: ButtonProps) {
  return (
    <button
      className={`inline-flex items-center gap-2 rounded-lg font-medium transition-colors disabled:opacity-50 ${variants[variant]} ${sizes[size]} ${className}`}
      {...props}
    />
  )
}
