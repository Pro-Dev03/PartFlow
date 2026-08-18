import { clsx } from 'clsx'
import type { AuditLog } from '../types/audit'

interface AuditLogTableProps {
  logs: AuditLog[]
  onViewDetails?: (log: AuditLog) => void
  className?: string
}

export function AuditLogTable({ logs, onViewDetails, className }: AuditLogTableProps) {
  const actionIcons = {
    create: '➕',
    update: '✏️',
    delete: '🗑️',
    sale: '💰',
    purchase: '🛒',
    login: '🔐',
    logout: '🚪',
    export: '📤',
    import: '📥',
  }

  const entityTypeLabels = {
    product: 'منتج',
    customer: 'عميل',
    sale: 'بيع',
    purchase: 'مشتراة',
    expense: 'مصروف',
    supplier: 'مورد',
    inventory: 'مخزون',
    settings: 'إعدادات',
    user: 'مستخدم',
  }

  const statusColors = {
    success: 'bg-success-10 text-success',
    failure: 'bg-danger-10 text-danger',
  }

  if (logs.length === 0) {
    return (
      <div className={clsx('text-center py-8 text-muted', className)}>
        لا يوجد سجل
      </div>
    )
  }

  return (
    <div className={clsx('bg-surface rounded-lg overflow-hidden', className)}>
      <table className="w-full">
        <thead className="bg-muted-10 border-b border-border">
          <tr>
            <th className="px-4 py-3 text-right text-sm font-medium text-muted">التاريخ</th>
            <th className="px-4 py-3 text-right text-sm font-medium text-muted">المستخدم</th>
            <th className="px-4 py-3 text-right text-sm font-medium text-muted">الإجراء</th>
            <th className="px-4 py-3 text-right text-sm font-medium text-muted">الكيان</th>
            <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
            <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
          </tr>
        </thead>
        <tbody>
          {logs.map((log) => (
            <tr key={log.id} className="border-b border-border hover:bg-muted-5">
              <td className="px-4 py-3 text-muted text-sm">
                {log.timestamp}
              </td>
              <td className="px-4 py-3">
                <div>
                  <p className="font-medium text-text">{log.userName}</p>
                  {log.ipAddress && (
                    <p className="text-xs text-muted">{log.ipAddress}</p>
                  )}
                </div>
              </td>
              <td className="px-4 py-3">
                <div className="flex items-center gap-2">
                  <span className="text-lg">
                    {actionIcons[log.action as keyof typeof actionIcons] || '📝'}
                  </span>
                  <span className="text-sm">{log.action}</span>
                </div>
              </td>
              <td className="px-4 py-3">
                <div>
                  <p className="text-sm font-medium text-text">
                    {entityTypeLabels[log.entityType as keyof typeof entityTypeLabels] || log.entityType}
                  </p>
                  {log.entityName && (
                    <p className="text-xs text-muted">{log.entityName}</p>
                  )}
                </div>
              </td>
              <td className="px-4 py-3">
                <span className={clsx('px-2 py-1 rounded text-xs font-medium', statusColors[log.status])}>
                  {log.status === 'success' ? 'نجح' : 'فشل'}
                </span>
              </td>
              <td className="px-4 py-3">
                <button
                  onClick={() => onViewDetails?.(log)}
                  className="text-primary hover:text-primary-600 text-sm"
                >
                  عرض التفاصيل
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
