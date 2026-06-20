const SAFE_PROTOCOLS = new Set(['http:', 'https:', 'mailto:'])

export function safeHref(url: string | undefined): string {
  if (!url) return '#'
  try {
    const { protocol } = new URL(url)
    return SAFE_PROTOCOLS.has(protocol) ? url : '#'
  } catch {
    return '#'
  }
}
