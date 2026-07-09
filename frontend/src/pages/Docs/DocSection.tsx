import { Fragment, type ReactNode } from 'react'

// ---------------------------------------------------------------------------
// Presentational primitives shared by every docs section
// ---------------------------------------------------------------------------

export function SectionHeading({ id, children }: { id: string; children: ReactNode }) {
  return (
    <h2 id={id} className="text-xl font-semibold text-text-primary mb-4 pt-2 scroll-mt-6">
      {children}
    </h2>
  )
}

export function SubHeading({ children }: { children: ReactNode }) {
  return (
    <h3 className="text-sm font-semibold text-text-secondary uppercase tracking-wider mt-6 mb-3">
      {children}
    </h3>
  )
}

export function P({ children }: { children: ReactNode }) {
  return (
    <p className="text-sm text-text-secondary leading-relaxed mb-3">
      {children}
    </p>
  )
}

export function Code({ children }: { children: ReactNode }) {
  return (
    <code className="bg-surface-elevated text-text-primary px-1.5 py-0.5 rounded text-[12px] font-mono">
      {children}
    </code>
  )
}

export function CodeBlock({ children }: { children: string }) {
  return (
    <pre className="bg-surface-base border border-border rounded-xl p-4 text-[12px] font-mono text-text-secondary overflow-x-auto mb-4 leading-relaxed">
      <code>{children}</code>
    </pre>
  )
}

export function Table({
  headers,
  rows,
}: {
  headers: string[]
  rows: (string | ReactNode)[][]
}) {
  return (
    <div className="overflow-x-auto mb-4">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className="border-b border-border">
            {headers.map((h) => (
              <th
                key={h}
                className="text-left py-2 pr-6 text-xs font-semibold text-text-muted uppercase tracking-wider whitespace-nowrap"
              >
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} className="border-b border-border/50 last:border-0">
              {row.map((cell, j) => (
                <td key={j} className="py-2 pr-6 text-text-secondary align-top">
                  {cell}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export function Divider() {
  return <hr className="border-border my-8" />
}

export function BulletList({ items }: { items: (string | ReactNode)[] }) {
  return (
    <ul className="space-y-1.5 mb-4">
      {items.map((item, i) => (
        <li key={i} className="flex gap-2 text-sm text-text-secondary leading-relaxed">
          <span className="text-text-muted shrink-0 mt-0.5">–</span>
          {item}
        </li>
      ))}
    </ul>
  )
}

// ---------------------------------------------------------------------------
// Generic content schema
// ---------------------------------------------------------------------------

/**
 * One renderable unit inside a subsection. Only the fields that are set are
 * rendered, in this fixed order: body -> table -> bullets -> code. Most
 * blocks only set a single field; a block may combine several when the
 * original layout required it (e.g. body + code together).
 */
export interface ContentBlock {
  body?: ReactNode[]
  table?: { headers: string[]; rows: (string | ReactNode)[][] }
  bullets?: (string | ReactNode)[]
  code?: string
}

export interface DocSubsectionData {
  /** Omit for content that sits directly under the section heading (no SubHeading rendered). */
  heading?: ReactNode
  blocks: ContentBlock[]
}

export interface DocSectionData {
  id: string
  title: ReactNode
  subsections: DocSubsectionData[]
}

/**
 * Renders a single documentation section from data. Every docs section in
 * this app follows the same shape: a heading, then a sequence of optionally
 * sub-headed blocks combining paragraphs, tables, bullet lists, and code
 * blocks — so this one component replaces what used to be ~17 near-identical
 * hand-written section components.
 */
export function DocSection({ data }: { data: DocSectionData }) {
  return (
    <section>
      <SectionHeading id={data.id}>{data.title}</SectionHeading>
      {data.subsections.map((sub, i) => (
        <div key={i}>
          {sub.heading !== undefined && <SubHeading>{sub.heading}</SubHeading>}
          {sub.blocks.map((block, j) => (
            <Fragment key={j}>
              {block.body?.map((paragraph, k) => (
                <P key={k}>{paragraph}</P>
              ))}
              {block.table && <Table headers={block.table.headers} rows={block.table.rows} />}
              {block.bullets && <BulletList items={block.bullets} />}
              {block.code !== undefined && <CodeBlock>{block.code}</CodeBlock>}
            </Fragment>
          ))}
        </div>
      ))}
    </section>
  )
}
