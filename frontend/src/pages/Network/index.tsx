import { useState, useCallback, useRef } from 'react'
import { Plus, Trash2, Download, Save, ArrowLeft, Image, FileCode2 } from 'lucide-react'
import { toPng, toSvg } from 'html-to-image'
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

const NODE_TYPES = [
  'process', 'decision', 'input_output', 'person', 'document', 'idea', 'image',
  'server', 'router', 'switch', 'firewall', 'client', 'database', 'cloud',
] as const

type NodeType = typeof NODE_TYPES[number]

function formatNodeType(type: NodeType) {
  return type.replace('_', ' ').replace(/^\w/, c => c.toUpperCase())
}

const BASE_NODE_STYLE = {
  color: 'var(--color-text-primary)',
  padding: '8px 14px',
  fontSize: 13,
} as const

const NODE_TYPE_STYLES: Record<NodeType, Record<string, string | number>> = {
  process: { background: '#1e3a5f', border: '1px solid #3b82f6', borderRadius: 6 },
  decision: {
    background: '#4a1942',
    border: '1px solid #ec4899',
    clipPath: 'polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%)',
    padding: '18px 28px',
  },
  input_output: {
    background: '#1f3d2e',
    border: '1px solid #22c55e',
    clipPath: 'polygon(12% 0%, 100% 0%, 88% 100%, 0% 100%)',
  },
  person: { background: '#3a2e1f', border: '1px solid #f59e0b', borderRadius: '50%' },
  document: { background: '#2a2a3d', border: '1px solid #8b5cf6', borderRadius: '4px 4px 16px 16px' },
  idea: { background: '#3d3a1f', border: '1px solid #eab308', borderRadius: '50%' },
  image: { background: '#1f2937', border: '1px dashed #6b7280', borderRadius: 4 },
  server: { background: 'var(--color-surface-elevated)', border: '1px solid var(--color-border)', borderRadius: 8 },
  router: { background: 'var(--color-surface-elevated)', border: '1px solid #38bdf8', borderRadius: 8 },
  switch: { background: 'var(--color-surface-elevated)', border: '1px solid #a78bfa', borderRadius: 8 },
  firewall: { background: 'var(--color-surface-elevated)', border: '1px solid #f87171', borderRadius: 8 },
  client: { background: 'var(--color-surface-elevated)', border: '1px solid #34d399', borderRadius: 8 },
  database: { background: 'var(--color-surface-elevated)', border: '1px solid #fbbf24', borderRadius: 8 },
  cloud: { background: 'var(--color-surface-elevated)', border: '1px solid #94a3b8', borderRadius: 8 },
}

function nodeStyle(type: NodeType) {
  return { ...BASE_NODE_STYLE, ...NODE_TYPE_STYLES[type] }
}

function templateNode(id: string, type: NodeType, label: string, x: number, y: number): Node {
  return {
    id,
    position: { x, y },
    data: { label: `${formatNodeType(type)}: ${label}` },
    style: nodeStyle(type),
  }
}

const TEMPLATES: Record<'blank' | 'flowchart' | 'mindmap' | 'architecture' | 'er', { nodes: Node[]; edges: Edge[] }> = {
  blank: { nodes: [], edges: [] },
  flowchart: {
    nodes: [
      templateNode('t1', 'server', 'Start', 100, 40),
      templateNode('t2', 'server', 'Do work', 100, 160),
      templateNode('t3', 'server', 'OK?', 100, 280),
      templateNode('t4', 'server', 'End', 100, 400),
    ],
    edges: [
      { id: 'te1', source: 't1', target: 't2' },
      { id: 'te2', source: 't2', target: 't3' },
      { id: 'te3', source: 't3', target: 't4', label: 'Yes' },
      { id: 'te4', source: 't3', target: 't2', label: 'No' },
    ],
  },
  mindmap: {
    nodes: [
      templateNode('t1', 'server', 'Main idea', 250, 200),
      templateNode('t2', 'server', 'Branch 1', 50, 60),
      templateNode('t3', 'server', 'Branch 2', 450, 60),
      templateNode('t4', 'server', 'Branch 3', 250, 380),
    ],
    edges: [
      { id: 'te1', source: 't1', target: 't2' },
      { id: 'te2', source: 't1', target: 't3' },
      { id: 'te3', source: 't1', target: 't4' },
    ],
  },
  architecture: {
    nodes: [
      templateNode('t1', 'client', 'Client', 50, 200),
      templateNode('t2', 'firewall', 'Firewall', 220, 200),
      templateNode('t3', 'server', 'API Server', 390, 200),
      templateNode('t4', 'database', 'Database', 560, 100),
      templateNode('t5', 'cloud', 'Cloud Storage', 560, 300),
    ],
    edges: [
      { id: 'te1', source: 't1', target: 't2' },
      { id: 'te2', source: 't2', target: 't3' },
      { id: 'te3', source: 't3', target: 't4' },
      { id: 'te4', source: 't3', target: 't5' },
    ],
  },
  er: {
    nodes: [
      templateNode('t1', 'database', 'User', 80, 200),
      templateNode('t2', 'database', 'Order', 300, 200),
      templateNode('t3', 'database', 'Product', 520, 200),
    ],
    edges: [
      { id: 'te1', source: 't1', target: 't2', label: 'has many' },
      { id: 'te2', source: 't2', target: 't3', label: 'contains' },
    ],
  },
}

type TemplateKey = keyof typeof TEMPLATES

const TEMPLATE_LABELS: Record<TemplateKey, string> = {
  blank: 'Blank',
  flowchart: 'Flowchart',
  mindmap: 'Mind map',
  architecture: 'Architecture',
  er: 'Entity-relationship',
}

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
  const [nodeType, setNodeType] = useState<NodeType>('process')
  const [edgeLabel, setEdgeLabel] = useState('')
  const flowRef = useRef<HTMLDivElement>(null)

  const onConnect = useCallback(
    (params: Connection) =>
      setEdges(eds => addEdge(edgeLabel ? { ...params, label: edgeLabel } : params, eds)),
    [setEdges, edgeLabel],
  )

  const addNode = () => {
    const id = nextId()
    setNodes(ns => [
      ...ns,
      {
        id,
        position: { x: 100 + Math.random() * 300, y: 100 + Math.random() * 200 },
        data: { label: `${formatNodeType(nodeType)}: ${label}` },
        style: nodeStyle(nodeType),
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

  const downloadDataUrl = (dataUrl: string, extension: string) => {
    const a = document.createElement('a')
    a.href = dataUrl
    a.download = `${diagram.name}.${extension}`
    a.click()
  }

  const exportPng = () => {
    if (!flowRef.current) return
    toPng(flowRef.current, { backgroundColor: '#0f0f13' }).then(dataUrl => downloadDataUrl(dataUrl, 'png'))
  }

  const exportSvg = () => {
    if (!flowRef.current) return
    toSvg(flowRef.current, { backgroundColor: '#0f0f13' }).then(dataUrl => downloadDataUrl(dataUrl, 'svg'))
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border bg-surface-elevated shrink-0">
        <select
          value={nodeType}
          onChange={e => setNodeType(e.target.value as NodeType)}
          className="bg-surface-base border border-border rounded px-2 py-1.5 text-sm text-text-primary"
        >
          {NODE_TYPES.map(t => <option key={t} value={t}>{formatNodeType(t)}</option>)}
        </select>
        <input
          value={label}
          onChange={e => setLabel(e.target.value)}
          placeholder="Label"
          className="bg-surface-base border border-border rounded px-2 py-1.5 text-sm text-text-primary w-32 placeholder-text-muted focus:outline-none focus:border-(--color-accent) transition-colors"
        />
        <Button size="sm" icon={<Plus size={14} />} onClick={addNode}>Add node</Button>
        <input
          value={edgeLabel}
          onChange={e => setEdgeLabel(e.target.value)}
          placeholder="Edge label"
          className="bg-surface-base border border-border rounded px-2 py-1.5 text-sm text-text-primary w-28 placeholder-text-muted focus:outline-none focus:border-(--color-accent) transition-colors"
        />
        <div className="flex-1" />
        <Button size="sm" variant="ghost" icon={<Download size={14} />} onClick={exportJSON}>Export JSON</Button>
        <Button size="sm" variant="ghost" icon={<Image size={14} />} onClick={exportPng}>PNG</Button>
        <Button size="sm" variant="ghost" icon={<FileCode2 size={14} />} onClick={exportSvg}>SVG</Button>
        <Button size="sm" icon={<Save size={14} />} onClick={() => onSave(nodes, edges)} disabled={isPending}>
          {isPending ? 'Saving…' : 'Save'}
        </Button>
      </div>

      <div className="flex-1" ref={flowRef}>
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
  const [template, setTemplate] = useState<TemplateKey>('blank')

  const handleCreate = () => {
    const name = newName.trim() || 'Untitled diagram'
    create.mutate({ name }, {
      onSuccess: d => {
        setNewName('')
        const { nodes, edges } = TEMPLATES[template]
        if (nodes.length === 0) {
          setSelected(d)
          return
        }
        update.mutate(
          { id: d.id, data: { nodes: JSON.stringify(nodes), edges: JSON.stringify(edges) } },
          { onSuccess: setSelected },
        )
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
      <PageHeader title="Diagrams" />

      <div className="flex gap-2">
        <input
          value={newName}
          onChange={e => setNewName(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && handleCreate()}
          placeholder="New diagram name…"
          className="bg-surface-elevated border border-border rounded-lg px-3 py-2 text-sm text-text-primary placeholder-text-muted focus:outline-none focus:border-(--color-accent) transition-colors w-64"
        />
        <select
          value={template}
          onChange={e => setTemplate(e.target.value as TemplateKey)}
          className="bg-surface-elevated border border-border rounded-lg px-3 py-2 text-sm text-text-primary"
        >
          {(Object.keys(TEMPLATE_LABELS) as TemplateKey[]).map(key => (
            <option key={key} value={key}>{TEMPLATE_LABELS[key]}</option>
          ))}
        </select>
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
