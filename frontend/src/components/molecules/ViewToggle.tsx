import { LayoutGrid, Table2 } from 'lucide-react'
import { IconButton } from '../atoms/IconButton'

export type ViewMode = 'kanban' | 'table'

export function ViewToggle({ value, onChange }: { value: ViewMode; onChange: (v: ViewMode) => void }) {
  return (
    <div className="flex items-center bg-surface-elevated border border-border rounded-lg p-0.5">
      <IconButton
        icon={<LayoutGrid size={16} />}
        aria-label="Kanban"
        variant={value === 'kanban' ? 'primary' : 'ghost'}
        onClick={() => onChange('kanban')}
      />
      <IconButton
        icon={<Table2 size={16} />}
        aria-label="Table"
        variant={value === 'table' ? 'primary' : 'ghost'}
        onClick={() => onChange('table')}
      />
    </div>
  )
}
