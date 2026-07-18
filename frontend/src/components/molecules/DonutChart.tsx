import { useState } from 'react'

type Slice = { label: string; value: number; color: string }

export function DonutChart({
  data,
  formatValue,
  centerLabel,
}: {
  data: Slice[]
  formatValue: (n: number) => string
  centerLabel?: string
}) {
  const [hover, setHover] = useState<number | null>(null)
  const total = data.reduce((s, d) => s + d.value, 0)
  const size = 200
  const stroke = 44
  const r = (size - stroke) / 2
  const c = 2 * Math.PI * r
  const a11y =
    data.length > 0
      ? `Expenses by category.${data
          .slice(0, 3)
          .map(s => ` ${s.label} ${formatValue(s.value)}`)
          .join(',')}.`
      : 'Expenses by category, no data.'

  if (data.length === 0 || total === 0) return null

  const segs = data.reduce<{ s: Slice; i: number; len: number; off: number }[]>(
    (acc, s, i) => {
      const len = c * (s.value / total)
      const off = acc.length === 0 ? 0 : acc[acc.length - 1].off + acc[acc.length - 1].len
      acc.push({ s, i, len, off })
      return acc
    },
    []
  )

  return (
    <div>
      <div className="flex flex-col items-center gap-4">
        <svg
          width={size}
          height={size}
          viewBox={`0 0 ${size} ${size}`}
          role="img"
          aria-label={a11y}
        >
          <g transform={`translate(${size / 2} ${size / 2}) rotate(-90)`}>
            {segs.map(({ s, i, len, off }) => (
              <circle
                key={s.label}
                r={r}
                fill="none"
                stroke={s.color}
                strokeWidth={stroke}
                strokeDasharray={`${len} ${c - len}`}
                strokeDashoffset={-off}
                opacity={hover === null || hover === i ? 1 : 0.4}
                onMouseEnter={() => setHover(i)}
                onMouseLeave={() => setHover(null)}
                style={{ transition: 'opacity 120ms' }}
              />
            ))}
          </g>
          <text
            x={size / 2}
            y={size / 2 - 2}
            textAnchor="middle"
            fill="currentColor"
            className="text-text-primary"
            style={{ fontSize: 22, fontWeight: 600 }}
          >
            {formatValue(total)}
          </text>
          <text
            x={size / 2}
            y={size / 2 + 18}
            textAnchor="middle"
            fill="currentColor"
            className="text-text-muted"
            style={{ fontSize: 11 }}
          >
            {centerLabel ?? 'Expenses'}
          </text>
        </svg>

        <ul className="w-full space-y-1.5">
          {data.map((s, i) => (
            <li
              key={s.label}
              className="flex items-center gap-2 text-sm"
              onMouseEnter={() => setHover(i)}
              onMouseLeave={() => setHover(null)}
            >
              <span
                className="w-2.5 h-2.5 rounded-full shrink-0"
                style={{ backgroundColor: s.color }}
              />
              <span className="text-text-secondary truncate flex-1">
                {s.label}
              </span>
              <span className="text-text-primary tabular-nums">
                {formatValue(s.value)}
              </span>
              <span className="text-text-muted tabular-nums w-12 text-right">
                {Math.round((s.value / total) * 100)}%
              </span>
            </li>
          ))}
        </ul>

        <p className="text-xs text-text-muted w-full">Excludes uncategorized.</p>
      </div>
    </div>
  )
}
