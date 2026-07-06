export function formatDate(date: string | null | undefined): string | null {
  return date ? date.slice(0, 10) : null
}
