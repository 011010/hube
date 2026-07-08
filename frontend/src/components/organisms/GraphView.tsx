import { useMemo } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  type Node,
  type Edge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { useNoteGraph, type GraphNodeType, type GraphEdgeType } from '../../hooks/useNotes'

const NODE_COLORS: Record<GraphNodeType, string> = {
  note: 'var(--color-accent)',
  task: '#f59e0b',
  project: '#10b981',
}

const EDGE_COLORS: Record<GraphEdgeType, string> = {
  link: 'var(--color-border)',
  task: '#f59e0b',
  project: '#10b981',
}

function nodeType(id: string): GraphNodeType {
  return (id.split(':')[0] as GraphNodeType) ?? 'note'
}

function nodeStyle(type: GraphNodeType) {
  const color = NODE_COLORS[type] ?? 'var(--color-border)'
  return {
    background: 'var(--color-surface-elevated)',
    border: `1.5px solid ${color}`,
    color: 'var(--color-text-primary)',
    borderRadius: 8,
    padding: '8px 14px',
    fontSize: 12,
  }
}

export function GraphView() {
  const { data, isLoading } = useNoteGraph()

  const { nodes, edges } = useMemo(() => {
    const graphNodes = data?.nodes ?? []
    const graphEdges = data?.edges ?? []
    const columns = Math.max(1, Math.ceil(Math.sqrt(graphNodes.length)))
    const spacingX = 220
    const spacingY = 140

    const nodes: Node[] = graphNodes.map((n, i) => ({
      id: n.id,
      position: { x: (i % columns) * spacingX, y: Math.floor(i / columns) * spacingY },
      data: { label: n.label },
      style: nodeStyle(n.type),
    }))

    const edges: Edge[] = graphEdges.map((e, i) => ({
      id: `${e.source}->${e.target}-${i}`,
      source: e.source,
      target: e.target,
      style: { stroke: EDGE_COLORS[e.type] ?? 'var(--color-border)' },
      animated: e.type === 'link',
    }))

    return { nodes, edges }
  }, [data])

  if (isLoading) {
    return <p className="text-text-muted text-sm p-4">Loading graph…</p>
  }

  if (nodes.length === 0) {
    return (
      <p className="text-text-muted text-sm p-4">
        No linked notes yet. Use <code>[[Note Title]]</code> inside a note, or link a task/project to a note, to see it here.
      </p>
    )
  }

  return (
    <div className="h-full w-full">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        fitView
        colorMode="dark"
        nodesConnectable={false}
      >
        <Background color="var(--color-border)" gap={20} />
        <Controls />
        <MiniMap
          nodeColor={(n) => NODE_COLORS[nodeType(n.id)] ?? 'var(--color-surface-elevated)'}
          maskColor="rgba(0,0,0,0.6)"
        />
      </ReactFlow>
    </div>
  )
}
