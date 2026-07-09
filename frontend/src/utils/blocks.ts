interface TipTapNode {
  type?: string
  text?: string
  content?: TipTapNode[]
  [key: string]: unknown
}

function walkNode(node: TipTapNode, parts: string[]) {
  if (typeof node.text === 'string') {
    parts.push(node.text)
  }
  if (Array.isArray(node.content)) {
    for (const child of node.content) {
      if (child && typeof child === 'object') {
        walkNode(child, parts)
      }
    }
  }
}

export function blocksToText(json: string): string {
  if (!json.trim()) return ''
  try {
    const doc: TipTapNode = JSON.parse(json)
    const parts: string[] = []
    walkNode(doc, parts)
    return parts.join(' ')
  } catch {
    return ''
  }
}

// textToBlocks wraps legacy plain-text content (from notes created/edited before
// the block editor existed) into a TipTap doc so it stays visible and editable
// instead of appearing blank. Each line becomes its own paragraph so multi-line
// legacy content keeps its visual structure instead of collapsing into one run-on line.
export function textToBlocks(text: string): string {
  const normalized = text.replace(/\r\n/g, '\n').replace(/\n$/, '')
  const paragraphs = normalized.split('\n').map(line =>
    line === ''
      ? { type: 'paragraph' }
      : { type: 'paragraph', content: [{ type: 'text', text: line }] }
  )
  return JSON.stringify({ type: 'doc', content: paragraphs })
}
