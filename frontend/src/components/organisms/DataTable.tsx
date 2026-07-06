import { useMemo, useState } from 'react'
import type { ReactNode, KeyboardEvent } from 'react'
import { ArrowDown, ArrowUp } from 'lucide-react'

export interface DataTableColumn<T> {
  key: string
  header: string
  render: (item: T) => ReactNode
  sortable?: boolean
}

export interface DataTableProps<T> {
  columns: DataTableColumn<T>[]
  data: T[]
}

type SortDirection = 'asc' | 'desc'

interface SortState {
  key: string
  direction: SortDirection
}

function renderValue<T>(column: DataTableColumn<T>, item: T): string {
  return String(column.render(item))
}

export function DataTable<T>({ columns, data }: DataTableProps<T>) {
  const [sort, setSort] = useState<SortState | null>(null)

  const sortedData = useMemo(() => {
    if (!sort) return data

    const column = columns.find((c) => c.key === sort.key)
    if (!column) return data

    const sorted = [...data].sort((a, b) => {
      const aValue = renderValue(column, a)
      const bValue = renderValue(column, b)
      return sort.direction === 'asc'
        ? aValue.localeCompare(bValue)
        : bValue.localeCompare(aValue)
    })

    return sorted
  }, [data, columns, sort])

  const handleHeaderClick = (column: DataTableColumn<T>) => {
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

  const handleHeaderKeyDown = (
    event: KeyboardEvent<HTMLTableCellElement>,
    column: DataTableColumn<T>,
  ) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      handleHeaderClick(column)
    }
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
                  className={[
                    'px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider',
                    column.sortable
                      ? 'cursor-pointer select-none hover:text-text-primary'
                      : '',
                  ].join(' ')}
                  aria-sort={ariaSort}
                  {...(column.sortable
                    ? {
                        role: 'button',
                        tabIndex: 0,
                        onClick: () => handleHeaderClick(column),
                        onKeyDown: (event) =>
                          handleHeaderKeyDown(event, column),
                      }
                    : {})}
                >
                  <span
                    className={[
                      'flex items-center gap-1.5',
                      column.sortable
                        ? 'text-text-secondary'
                        : 'text-text-muted',
                      isActive ? 'text-text-primary' : '',
                    ].join(' ')}
                  >
                    {column.header}
                    {isActive && (
                      <span className="text-text-muted">
                        {sort.direction === 'asc' ? (
                          <ArrowUp size={14} />
                        ) : (
                          <ArrowDown size={14} />
                        )}
                      </span>
                    )}
                  </span>
                </th>
              )
            })}
          </tr>
        </thead>
        <tbody>
          {sortedData.map((item, rowIndex) => (
            <tr
              key={rowIndex}
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
          ))}
        </tbody>
      </table>
    </div>
  )
}
