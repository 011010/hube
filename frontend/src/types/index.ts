export type Priority = 'low' | 'medium' | 'high'
export type TaskStatus = 'todo' | 'in_progress' | 'done'

export interface Task {
  id: string
  title: string
  description: string
  priority: Priority
  status: TaskStatus
  due_date: string | null
  created_at: string
  updated_at: string
}

export interface CalendarEvent {
  id: string
  title: string
  description: string
  start_at: string
  end_at: string
  all_day: boolean
  color: string
  created_at: string
  updated_at: string
}

export interface App {
  id: string
  name: string
  description: string
  url: string
  icon: string
  color: string
  sort_order: number
  active: boolean
}
