export type Priority = 'low' | 'medium' | 'high'
export type TaskStatus = 'todo' | 'in_progress' | 'done'

export interface Task {
  id: string
  title: string
  description: string
  priority: Priority
  status: TaskStatus
  due_date: string | null
  project_id: string | null
  recurrence: string
  created_at: string
  updated_at: string
}

export type ProjectStatus = 'planning' | 'active' | 'completed' | 'on_hold'

export interface Project {
  id: string
  name: string
  description: string
  status: ProjectStatus
  color: string
  due_date: string | null
  task_count: number
  completed_count: number
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

export interface CardTrackerCard {
  id: string
  alias: string
  bank: string
  last4: string
  pay_date: string
  limit: number
  estimated_payment: number
  balance: number
}

export interface CardTrackerSummary {
  configured: boolean
  total_debt: number
  total_payment: number
  next_pay_date: string | null
  cards: CardTrackerCard[]
}

export interface RecentTransaction {
  id: string
  amount: number
  type: 'income' | 'expense'
  category: string
  description: string
  date: string
}

export interface FinanceSummary {
  configured: boolean
  balance: number
  month_income: number
  month_expenses: number
  recent_transactions: RecentTransaction[]
}

export type NoteStatus = 'draft' | 'in_progress' | 'published'

export interface Note {
  id: string
  title: string
  content: string
  blocks: string
  status: NoteStatus
  priority: Priority
  due_date: string | null
  folder_id: string | null
  tags: string[]
  created_at: string
  updated_at: string
}

export interface Folder {
  id: string
  name: string
  parent_id: string | null
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

export interface Diagram {
  id: string
  name: string
  nodes: string
  edges: string
  created_at: string
  updated_at: string
}

export type WishlistStatus = 'pending' | 'purchased'

export interface WishlistItem {
  id: string
  name: string
  description: string
  url: string
  store: string
  target_price: number
  current_price: number
  currency: string
  priority: Priority
  status: WishlistStatus
  notes: string
  created_at: string
  updated_at: string
}
