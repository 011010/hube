import { useState, useEffect } from 'react'
import { CheckCircle, XCircle, Info, AlertTriangle, X } from 'lucide-react'

interface ToastProps {
  message: string
  variant?: 'success' | 'error' | 'info' | 'warning'
  duration?: number
  onDismiss: () => void
}

const ICONS = {
  success: CheckCircle,
  error: XCircle,
  info: Info,
  warning: AlertTriangle,
}

const STYLES = {
  success: 'bg-emerald-900/60 border-emerald-700 text-emerald-200',
  error: 'bg-red-900/60 border-red-700 text-red-200',
  info: 'bg-sky-900/60 border-sky-700 text-sky-200',
  warning: 'bg-amber-900/60 border-amber-700 text-amber-200',
}

export function Toast({ message, variant = 'info', duration = 4000, onDismiss }: ToastProps) {
  const [visible, setVisible] = useState(true)

  useEffect(() => {
    const timer = setTimeout(() => {
      setVisible(false)
      setTimeout(onDismiss, 200)
    }, duration)
    return () => clearTimeout(timer)
  }, [duration, onDismiss])

  const Icon = ICONS[variant]

  return (
    <div
      className={`fixed bottom-4 right-4 z-50 flex items-center gap-3 px-4 py-3 rounded-lg border shadow-lg transition-all ${
        STYLES[variant]
      } ${visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-2'}`}
    >
      <Icon size={18} className="shrink-0" />
      <span className="text-sm">{message}</span>
      <button onClick={onDismiss} className="ml-2 opacity-60 hover:opacity-100 transition-opacity">
        <X size={14} />
      </button>
    </div>
  )
}
