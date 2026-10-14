// Due dates travel as bare calendar days ("YYYY-MM-DD"). Returning the day
// verbatim avoids `new Date()` reinterpreting it as UTC midnight, which renders
// the previous day on machines west of UTC.
const DAY = /^\d{4}-\d{2}-\d{2}$/

export function formatDate(date: string | null | undefined): string | null {
  if (!date) return null
  const day = date.slice(0, 10)
  return DAY.test(day) ? day : null
}
