import { useRef, useState } from 'react'

type Point = { date: string; total: number }

export function AreaChart({
  data,
  formatValue,
  formatDate,
  height = 200,
}: {
  data: Point[]
  formatValue: (n: number) => string
  formatDate: (iso: string) => string
  height?: number
}) {
  const width = 600
  const padL = 52
  const padR = 12
  const padT = 16
  const padB = 28
  const innerW = width - padL - padR
  const innerH = height - padT - padB
  const yMax = Math.max(1, ...data.map(d => d.total))
  const xStep = data.length > 1 ? innerW / (data.length - 1) : 0
  const yTicks = niceTicks(yMax, 4)
  const yTop = yTicks[yTicks.length - 1] || 1
  const xPos = (i: number) => padL + i * xStep
  const yPos = (v: number) => padT + innerH - (v / yTop) * innerH

  const linePath =
    data.length > 0
      ? `M ${xPos(0)} ${yPos(data[0].total)} ` +
        data
          .slice(1)
          .map((d, i) => `L ${xPos(i + 1)} ${yPos(d.total)}`)
          .join(' ')
      : ''
  const areaPath =
    data.length > 0
      ? `M ${xPos(0)} ${yPos(data[0].total)} ` +
        data
          .slice(1)
          .map((d, i) => `L ${xPos(i + 1)} ${yPos(d.total)}`)
          .join(' ') +
        ` L ${xPos(data.length - 1)} ${padT + innerH} L ${xPos(0)} ${padT + innerH} Z`
      : ''

  const xTickIdx = pickXTicks(data.length, 5)
  const [hover, setHover] = useState<number | null>(null)
  const svgRef = useRef<SVGSVGElement | null>(null)
  const [tip, setTip] = useState<{ x: number; y: number } | null>(null)

  const total = data.reduce((s, d) => s + d.total, 0)
  const a11y =
    data.length > 0
      ? `Daily expenses, ${formatDate(data[0].date)} to ${formatDate(
          data[data.length - 1].date,
        )}, total ${formatValue(total)}.`
      : 'Daily expenses, no data.'

  return (
    <div className="relative">
      <svg
        ref={svgRef}
        viewBox={`0 0 ${width} ${height}`}
        width="100%"
        height={height}
        role="img"
        aria-label={a11y}
        onMouseLeave={() => {
          setHover(null)
          setTip(null)
        }}
      >
        {yTicks.map(t => (
          <g key={t}>
            <line
              x1={padL}
              x2={width - padR}
              y1={yPos(t)}
              y2={yPos(t)}
              stroke="currentColor"
              className="text-text-muted"
              strokeOpacity={0.3}
            />
            <text
              x={padL - 6}
              y={yPos(t) + 3}
              textAnchor="end"
              fontSize={10}
              fill="currentColor"
              className="text-text-muted"
            >
              {formatValue(t)}
            </text>
          </g>
        ))}

        {xTickIdx.map(i => (
          <text
            key={data[i].date}
            x={xPos(i)}
            y={height - 8}
            textAnchor="middle"
            fontSize={10}
            fill="currentColor"
            className="text-text-muted"
          >
            {formatDate(data[i].date)}
          </text>
        ))}

        <path
          d={areaPath}
          fill="currentColor"
          className="text-(--color-accent)"
          fillOpacity={0.18}
        />
        <path
          d={linePath}
          fill="none"
          stroke="currentColor"
          className="text-(--color-accent)"
          strokeWidth={2}
          strokeLinejoin="round"
          strokeLinecap="round"
        />

        {data.map((d, i) => (
          <circle
            key={d.date}
            cx={xPos(i)}
            cy={yPos(d.total)}
            r={3}
            fill="currentColor"
            className="text-(--color-accent)"
            opacity={hover === i ? 1 : 0}
          />
        ))}

        <rect
          x={padL}
          y={padT}
          width={innerW}
          height={innerH}
          fill="transparent"
          onMouseMove={e => {
            const rect = svgRef.current?.getBoundingClientRect()
            if (!rect) return
            const rx = ((e.clientX - rect.left) / rect.width) * width
            if (data.length === 0) return
            const idx = Math.max(
              0,
              Math.min(data.length - 1, Math.round((rx - padL) / xStep)),
            )
            setHover(idx)
            setTip({ x: e.clientX - rect.left, y: e.clientY - rect.top })
          }}
        />

        {hover !== null && (
          <line
            x1={xPos(hover)}
            x2={xPos(hover)}
            y1={padT}
            y2={padT + innerH}
            stroke="currentColor"
            className="text-text-muted"
            strokeOpacity={0.7}
            strokeDasharray="3 3"
          />
        )}
      </svg>

      {hover !== null && tip && data[hover] && (
        <div
          className="pointer-events-none absolute bg-surface-elevated border border-border rounded-md px-2 py-1 text-xs shadow"
          style={{ left: tip.x + 8, top: Math.max(0, tip.y - 44) }}
        >
          <div className="text-text-muted">{formatDate(data[hover].date)}</div>
          <div className="text-text-primary font-medium tabular-nums">
            {formatValue(data[hover].total)}
          </div>
        </div>
      )}
    </div>
  )
}

function niceNumber(x: number): number {
  const exp = Math.floor(Math.log10(x))
  const f = x / Math.pow(10, exp)
  const nf = f < 1.5 ? 1 : f < 3 ? 2 : f < 7 ? 5 : 10
  return nf * Math.pow(10, exp)
}

function niceTicks(max: number, count: number): number[] {
  if (max <= 0) return [0]
  const step = niceNumber(max / count)
  const ticks: number[] = []
  for (let v = 0; v <= max + step / 2; v += step) {
    ticks.push(Math.round(v))
  }
  return ticks
}

function pickXTicks(n: number, want: number): number[] {
  if (n <= 0) return []
  if (n <= want) return Array.from({ length: n }, (_, i) => i)
  const step = (n - 1) / (want - 1)
  return Array.from({ length: want }, (_, i) => Math.round(i * step))
}
