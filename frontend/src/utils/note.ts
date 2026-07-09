import type { NoteStatus } from '../types'
import { humanizeStatus } from './badges'

export function statusLabel(status: NoteStatus): string {
  return humanizeStatus(status)
}

export function statusVariant(status: NoteStatus): 'default' | 'success' | 'warning' {
  switch (status) {
    case 'published':
      return 'success'
    case 'in_progress':
      return 'warning'
    case 'draft':
    default:
      return 'default'
  }
}
