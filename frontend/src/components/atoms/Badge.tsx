interface BadgeProps {
  label: string
  variant?: 'default' | 'success' | 'warning' | 'danger'
}

const variants = {
  default: 'bg-gray-700 text-gray-300',
  success: 'bg-emerald-900 text-emerald-300',
  warning: 'bg-amber-900 text-amber-300',
  danger: 'bg-red-900 text-red-300',
}

export function Badge({ label, variant = 'default' }: BadgeProps) {
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${variants[variant]}`}>
      {label}
    </span>
  )
}
