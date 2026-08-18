import { clsx } from 'clsx'

interface MobileTableColumn {
  key: string
  label: string
  render?: (value: any, row: any) => React.ReactNode
}

interface MobileTableProps {
  columns: MobileTableColumn[]
  data: any[]
  onRowClick?: (row: any) => void
  className?: string
}

export function MobileTable({ columns, data, onRowClick, className }: MobileTableProps) {
  return (
    <div className={clsx('space-y-3', className)}>
      {data.map((row, rowIndex) => (
        <div
          key={rowIndex}
          onClick={() => onRowClick?.(row)}
          className="bg-surface rounded-lg border border-border p-4 active:bg-muted-10 transition-colors"
        >
          {columns.map((column) => (
            <div key={column.key} className="mb-3 last:mb-0">
              <p className="text-xs text-muted mb-1">{column.label}</p>
              <p className="font-medium text-text">
                {column.render ? column.render(row[column.key], row) : row[column.key]}
              </p>
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}
