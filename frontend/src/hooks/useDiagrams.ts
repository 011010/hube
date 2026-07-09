import { createCrudHooks } from './createCrudHooks'
import { diagramsApi } from '../services/api'
import type { Diagram } from '../types'

const diagramsCrud = createCrudHooks<Diagram, Pick<Diagram, 'name'>, Partial<Diagram>>(
  'diagrams',
  diagramsApi,
)

export const useDiagrams = diagramsCrud.useList
export const useCreateDiagram = diagramsCrud.useCreate
export const useUpdateDiagram = diagramsCrud.useUpdate
export const useDeleteDiagram = diagramsCrud.useDelete
