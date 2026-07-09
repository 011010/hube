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
// the block editor existed) into a minimal TipTap doc so it stays visible and
// editable instead of appearing blank.
export function textToBlocks(text: string): string {
  return JSON.stringify({
    type: 'doc',
    content: [
      {
        type: 'paragraph',
        content: [{ type: 'text', text }],
      },
    ],
  })
}
