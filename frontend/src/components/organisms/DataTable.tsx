import { useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { ArrowDown, ArrowUp, ArrowUpDown } from 'lucide-react'

export type DataTableColumn<T> =
  | {
      key: string
      header: string
      render: (item: T) => ReactNode
      sortable?: false
      sortValue?: never
    }
  | {
      key: string
      header: string
      render: (item: T) => ReactNode
      sortable: true
      sortValue: (item: T) => string | number
    }

export interface DataTableProps<T> {
  columns: DataTableColumn<T>[]
  data: T[]
  getRowKey: (item: T) => string | number
  emptyMessage?: ReactNode
}

type SortDirection = 'asc' | 'desc'

interface SortState {
  key: string
  direction: SortDirection
}

function compareValues(
  a: string | number,
  b: string | number,
  direction: SortDirection,
): number {
  if (typeof a === 'number' && typeof b === 'number') {
    return direction === 'asc' ? a - b : b - a
  }

  const aString = String(a)
  const bString = String(b)
  return direction === 'asc'
    ? aString.localeCompare(bString)
    : bString.localeCompare(aString)
}

export function DataTable<T>({
  columns,
  data,
  getRowKey,
  emptyMessage = 'No rows',
}: DataTableProps<T>) {
  const [sort, setSort] = useState<SortState | null>(null)

  const sortedData = useMemo(() => {
    if (!sort) return data

    const column = columns.find((c) => c.key === sort.key)
    if (!column || !column.sortable) return data

    return [...data].sort((a, b) =>
      compareValues(
        column.sortValue(a),
        column.sortValue(b),
        sort.direction,
      ),
    )
  }, [data, columns, sort])

  const handleSort = (column: DataTableColumn<T>) => {
    if (!column.sortable) return

    setSort((current) => {
      if (current?.key === column.key) {
        return {
          key: column.key,
          direction: current.direction === 'asc' ? 'desc' : 'asc',
        }
      }
      return { key: column.key, direction: 'asc' }
    })
  }

  return (
    <div className="overflow-hidden rounded-xl border border-border bg-surface-elevated">
      <table className="w-full border-collapse text-sm">
        <thead>
          <tr className="border-b border-border">
            {columns.map((column) => {
              const isActive = sort?.key === column.key
              const ariaSort = isActive
                ? sort.direction === 'asc'
                  ? 'ascending'
                  : 'descending'
                : undefined

              return (
                <th
                  key={column.key}
                  scope="col"
                  className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider"
                  aria-sort={ariaSort}
                >
                  {column.sortable ? (
                    <button
                      type="button"
                      onClick={() => handleSort(column)}
                      className={[
                        'flex w-full items-center gap-1.5 text-left font-[inherit]',
                        'bg-transparent p-0 text-xs font-semibold uppercase tracking-wider',
                        'text-text-secondary hover:text-text-primary',
                        'cursor-pointer select-none',
                        isActive ? 'text-text-primary' : '',
                      ].join(' ')}
                    >
                      {column.header}
                      <span className="text-text-muted">
                        {isActive ? (
                          sort.direction === 'asc' ? (
                            <ArrowUp size={14} />
                          ) : (
                            <ArrowDown size={14} />
                          )
                        ) : (
                          <ArrowUpDown size={14} />
                        )}
                      </span>
                    </button>
                  ) : (
                    <span className="text-text-muted">{column.header}</span>
                  )}
                </th>
              )
            })}
          </tr>
        </thead>
        <tbody>
          {data.length === 0 ? (
            <tr>
              <td
                colSpan={columns.length}
                className="px-4 py-8 text-center text-text-muted"
              >
                {emptyMessage}
              </td>
            </tr>
          ) : (
            sortedData.map((item) => (
              <tr
                key={getRowKey(item)}
                className="border-b border-border/50 transition-colors last:border-b-0 hover:bg-surface-card"
              >
                {columns.map((column) => (
                  <td
                    key={column.key}
                    className="px-4 py-3 text-text-secondary"
                  >
                    {column.render(item)}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  )
}
