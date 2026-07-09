import type { TaskStatus } from '../types'
import { humanizeStatus } from './badges'

export function statusLabel(status: TaskStatus): string {
  return humanizeStatus(status)
}

export function statusVariant(status: TaskStatus): 'default' | 'success' | 'warning' {
  switch (status) {
    case 'done':
      return 'success'
    case 'in_progress':
      return 'warning'
    case 'todo':
    default:
      return 'default'
  }
}
