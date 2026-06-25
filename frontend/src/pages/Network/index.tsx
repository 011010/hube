import { useState, useCallback } from 'react'
import { Plus, Trash2, Download, Save, ArrowLeft } from 'lucide-react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  addEdge,
  useNodesState,
  useEdgesState,
  type Connection,
  type Node,
  type Edge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { useDiagrams, useCreateDiagram, useUpdateDiagram, useDeleteDiagram } from '../../hooks/useDiagrams'
import { Button } from '../../components/atoms/Button'
import { PageHeader } from '../../components/molecules/PageHeader'
import type { Diagram } from '../../types'

let nodeId = 1
function nextId() { return `n${nodeId++}` }

const NODE_TYPES = ['Server', 'Router', 'Switch', 'Firewall', 'Client', 'Database', 'Cloud'] as const

function DiagramEditor({
  diagram,
  onSave,
  isPending,
}: {
  diagram: Diagram
  onSave: (nodes: Node[], edges: Edge[]) => void
  isPending: boolean
}) {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>(
    JSON.parse(diagram.nodes || '[]'),
  )
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>(
    JSON.parse(diagram.edges || '[]'),
  )
  const [label, setLabel] = useState('Node')
  const [nodeType, setNodeType] = useState<typeof NODE_TYPES[number]>('Server')

  const onConnect = useCallback(
    (params: Connection) => setEdges(eds => addEdge(params, eds)),
    [setEdges],
  )

  const addNode = () => {
    const id = nextId()
    setNodes(ns => [
      ...ns,
      {
        id,
        position: { x: 100 + Math.random() * 300, y: 100 + Math.random() * 200 },
        data: { label: `${nodeType}: ${label}` },
        style: {
          background: 'var(--color-surface-elevated)',
          border: '1px solid var(--color-border)',
          color: 'var(--color-text-primary)',
          borderRadius: 8,
          padding: '8px 14px',
          fontSize: 13,
        },
      },
    ])
  }

  const exportJSON = () => {
    const data = JSON.stringify({ nodes, edges }, null, 2)
    const blob = new Blob([data], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${diagram.name}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border bg-surface-elevated shrink-0">
        <select
          value={nodeType}
          onChange={e => setNodeType(e.target.value as typeof NODE_TYPES[number])}
          className="bg-surface-base border border-border rounded px-2 py-1.5 text-sm text-text-primary"
        >
          {NODE_TYPES.map(t => <option key={t}>{t}</option>)}
        </select>
        <input
          value={label}
          onChange={e => setLabel(e.target.value)}
          placeholder="Label"
          className="bg-surface-base border border-border rounded px-2 py-1.5 text-sm text-text-primary w-32 placeholder-text-muted focus:outline-none focus:border-(--color-accent) transition-colors"
        />
        <Button size="sm" icon={<Plus size={14} />} onClick={addNode}>Add node</Button>
        <div className="flex-1" />
        <Button size="sm" variant="ghost" icon={<Download size={14} />} onClick={exportJSON}>Export JSON</Button>
        <Button size="sm" icon={<Save size={14} />} onClick={() => onSave(nodes, edges)} disabled={isPending}>
          {isPending ? 'Saving…' : 'Save'}
        </Button>
      </div>

      <div className="flex-1">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          fitView
          colorMode="dark"
        >
          <Background color="var(--color-border)" gap={20} />
          <Controls />
          <MiniMap nodeColor="var(--color-surface-elevated)" maskColor="rgba(0,0,0,0.6)" />
        </ReactFlow>
      </div>
    </div>
  )
}

export function NetworkPage() {
  const { data: diagrams = [], isLoading } = useDiagrams()
  const create = useCreateDiagram()
  const update = useUpdateDiagram()
  const remove = useDeleteDiagram()

  const [selected, setSelected] = useState<Diagram | null>(null)
  const [newName, setNewName] = useState('')

  const handleCreate = () => {
    const name = newName.trim() || 'Untitled diagram'
    create.mutate({ name }, {
      onSuccess: d => {
        setNewName('')
        setSelected(d)
      },
    })
  }

  const handleSave = (nodes: Node[], edges: Edge[]) => {
    if (!selected) return
    update.mutate(
      { id: selected.id, data: { nodes: JSON.stringify(nodes), edges: JSON.stringify(edges) } },
      { onSuccess: d => setSelected(d) },
    )
  }

  if (selected) {
    return (
      <div className="flex flex-col h-screen">
        <div className="flex items-center gap-3 px-4 py-3 bg-surface-elevated border-b border-border shrink-0">
          <button
            onClick={() => setSelected(null)}
            className="text-text-muted hover:text-text-primary text-sm transition-colors flex items-center gap-1"
          >
            <ArrowLeft size={14} />
            Diagrams
          </button>
          <span className="text-text-primary font-medium text-sm">{selected.name}</span>
        </div>
        <DiagramEditor
          diagram={selected}
          onSave={handleSave}
          isPending={update.isPending}
        />
      </div>
    )
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader title="Network Diagrams" />

      <div className="flex gap-2">
        <input
          value={newName}
          onChange={e => setNewName(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && handleCreate()}
          placeholder="New diagram name…"
          className="bg-surface-elevated border border-border rounded-lg px-3 py-2 text-sm text-text-primary placeholder-text-muted focus:outline-none focus:border-(--color-accent) transition-colors w-64"
        />
        <Button icon={<Plus size={16} />} onClick={handleCreate} disabled={create.isPending}>Create</Button>
      </div>

      {isLoading && <p className="text-text-muted text-sm">Loading…</p>}

      {diagrams.length === 0 && !isLoading && (
        <p className="text-text-muted text-sm">No diagrams yet. Create one to get started.</p>
      )}

      <ul className="space-y-2">
        {diagrams.map(d => (
          <li
            key={d.id}
            className="flex items-center gap-3 bg-surface-elevated border border-border rounded-lg px-4 py-3 group"
          >
            <button
              onClick={() => setSelected(d)}
              className="flex-1 text-left text-sm text-text-secondary hover:text-text-primary transition-colors"
            >
              <span className="font-medium">{d.name}</span>
              <span className="ml-3 text-xs text-text-muted">
                {JSON.parse(d.nodes || '[]').length} nodes
              </span>
            </button>
            <button
              onClick={() => remove.mutate(d.id)}
              className="text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100"
              aria-label="Delete diagram"
            >
              <Trash2 size={14} />
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
