import { createCrudHooks } from './createCrudHooks'
import { tasksApi } from '../services/api'
import type { Task } from '../types'

const tasksCrud = createCrudHooks<Task, Omit<Task, 'id' | 'created_at' | 'updated_at'>, Partial<Task>>(
  'tasks',
  tasksApi,
)

export const useTasks = tasksCrud.useList
export const useCreateTask = tasksCrud.useCreate
export const useUpdateTask = tasksCrud.useUpdate
export const useDeleteTask = tasksCrud.useDelete
