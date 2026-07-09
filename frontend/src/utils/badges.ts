import type { Priority } from '../types'

export function priorityVariant(p: Priority): 'default' | 'warning' | 'danger' {
  return p === 'high' ? 'danger' : p === 'medium' ? 'warning' : 'default'
}

/** Capitalizes a status value, special-casing `in_progress` -> "In Progress". */
export function humanizeStatus(status: string): string {
  return status === 'in_progress' ? 'In Progress' : status.charAt(0).toUpperCase() + status.slice(1)
}
