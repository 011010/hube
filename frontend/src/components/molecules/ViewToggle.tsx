import { LayoutGrid, Table2 } from 'lucide-react'
import { IconButton } from '../atoms/IconButton'
import { type ViewMode } from '../../hooks/useViewPreference'

export function ViewToggle({ value, onChange }: { value: ViewMode; onChange: (v: ViewMode) => void }) {
  return (
    <div role="group" aria-label="View mode" className="flex items-center bg-surface-elevated border border-border rounded-lg p-0.5">
      <IconButton
        icon={<LayoutGrid size={16} />}
        aria-label="Kanban"
        aria-pressed={value === 'kanban'}
        variant={value === 'kanban' ? 'primary' : 'ghost'}
        onClick={() => onChange('kanban')}
      />
      <IconButton
        icon={<Table2 size={16} />}
        aria-label="Table"
        aria-pressed={value === 'table'}
        variant={value === 'table' ? 'primary' : 'ghost'}
        onClick={() => onChange('table')}
      />
    </div>
  )
}
