import type { Project, ProjectStatus } from '../types'

export function progress(p: Project): number {
  if (p.task_count === 0) return 0
  return Math.round((p.completed_count / p.task_count) * 100)
}

export function statusLabel(status: ProjectStatus): string {
  const labels: Record<ProjectStatus, string> = {
    planning: 'Planning',
    active: 'Active',
    completed: 'Completed',
    on_hold: 'On Hold',
  }
  return labels[status]
}

export function statusVariant(status: ProjectStatus): 'default' | 'accent' | 'success' | 'warning' | 'danger' {
  switch (status) {
    case 'completed':
      return 'success'
    case 'active':
      return 'accent'
    case 'on_hold':
      return 'warning'
    case 'planning':
    default:
      return 'default'
  }
}
