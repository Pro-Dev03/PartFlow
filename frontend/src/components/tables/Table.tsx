import { clsx } from 'clsx'

export interface TableColumn<T = any> {
  key: string
  label: string
  sortable?: boolean
  render?: (value: any, row: T, index: number) => React.ReactNode
  className?: string
}

interface TableProps<T = any> {
  columns: TableColumn<T>[]
  data: T[]
  onRowClick?: (row: T, index: number) => void
  onSort?: (column: string, direction: 'asc' | 'desc') => void
  sortColumn?: string
  sortDirection?: 'asc' | 'desc'
  loading?: boolean
  emptyMessage?: string
  className?: string
}

export function Table<T = any>({
  columns,
  data,
  onRowClick,
  onSort,
  sortColumn,
  sortDirection,
  loading = false,
  emptyMessage = 'لا توجد بيانات',
  className,
}: TableProps<T>) {
  const handleSort = (column: TableColumn<T>) => {
    if (!column.sortable || !onSort) return

    const newDirection =
      sortColumn === column.key && sortDirection === 'asc' ? 'desc' : 'asc'
    onSort(column.key, newDirection)
  }

  const getSortIcon = (column: TableColumn<T>) => {
    if (sortColumn !== column.key) return null
    return sortDirection === 'asc' ? '↑' : '↓'
  }

  if (loading) {
    return (
      <div className={clsx('bg-surface rounded-lg border border-border p-8', className)}>
        <div className="text-center text-muted">جاري التحميل...</div>
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className={clsx('bg-surface rounded-lg border border-border p-8', className)}>
        <div className="text-center text-muted">{emptyMessage}</div>
      </div>
    )
  }

  return (
    <div className={clsx('bg-surface rounded-lg border border-border overflow-hidden', className)}>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-muted-10 border-b border-border">
            <tr>
              {columns.map((column) => (
                <th
                  key={column.key}
                  onClick={() => handleSort(column)}
                  className={clsx(
                    'px-4 py-3 text-right text-sm font-medium text-muted whitespace-nowrap',
                    column.sortable && onSort && 'cursor-pointer hover:bg-muted-20 transition-colors',
                    column.className
                  )}
                >
                  <div className="flex items-center gap-2">
                    {column.label}
                    {getSortIcon(column)}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.map((row, rowIndex) => (
              <tr
                key={rowIndex}
                onClick={() => onRowClick?.(row, rowIndex)}
                className={clsx(
                  'border-b border-border hover:bg-muted-5 transition-colors',
                  onRowClick && 'cursor-pointer'
                )}
              >
                {columns.map((column) => (
                  <td
                    key={column.key}
                    className={clsx('px-4 py-3 whitespace-nowrap', column.className)}
                  >
                    {column.render
                      ? column.render(row[column.key], row, rowIndex)
                      : row[column.key]}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
