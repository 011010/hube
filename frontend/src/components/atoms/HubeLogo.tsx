const NODES: [number, number][] = [
  [16,    6],
  [24.66, 11],
  [24.66, 21],
  [16,    26],
  [7.34,  21],
  [7.34,  11],
]

export function HubeLogo({ size = 20 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      {NODES.map(([x, y], i) => (
        <line
          key={i}
          x1="16" y1="16"
          x2={x}  y2={y}
          stroke="var(--color-accent)"
          strokeWidth="1.2"
          strokeLinecap="round"
          opacity="0.45"
        />
      ))}
      {NODES.map(([cx, cy], i) => (
        <circle
          key={i}
          cx={cx}
          cy={cy}
          r="2"
          fill="var(--color-accent)"
          opacity="0.65"
        />
      ))}
      <circle cx="16" cy="16" r="4.5" fill="var(--color-accent)" />
    </svg>
  )
}
