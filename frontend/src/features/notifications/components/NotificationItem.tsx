import { clsx } from 'clsx'
import type { Notification } from '../types/notification'

interface NotificationItemProps {
  notification: Notification
  detailed?: boolean
  onMarkAsRead?: () => void
  onArchive?: () => void
  onDelete?: () => void
  onAction?: () => void
  onClick?: () => void
  className?: string
}

export function NotificationItem({
  notification,
  detailed = false,
  onMarkAsRead,
  onArchive,
  onDelete,
  onAction,
  onClick,
  className,
}: NotificationItemProps) {
  const typeIcons = {
    low_stock: '📦',
    overdue_debt: '💰',
    warranty_expiring: '🛡️',
    purchase_alert: '🛒',
    sale_completed: '✅',
    payment_received: '💳',
    return_requested: '↩️',
    system_update: '⚙️',
    custom: '📌',
  }

  const typeLabels = {
    low_stock: 'انخفاض المخزون',
    overdue_debt: 'ديون متأخرة',
    warranty_expiring: 'انتهاء الضمان',
    purchase_alert: 'تنبيه مشتريات',
    sale_completed: 'عملية بيع',
    payment_received: 'استلام دفعة',
    return_requested: 'طلب مرتج',
    system_update: 'تحديث النظام',
    custom: 'إشعار',
  }

  const priorityColors = {
    low: 'border-l-muted',
    medium: 'border-l-info',
    high: 'border-l-warning',
    urgent: 'border-l-danger',
  }

  const priorityLabels = {
    low: 'منخفضة',
    medium: 'متوسطة',
    high: 'عالية',
    urgent: 'عاجلة',
  }

  const statusColors = {
    unread: 'bg-primary-5',
    read: 'bg-surface',
    archived: 'bg-muted-5',
  }

  const timeAgo = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000)

    if (seconds < 60) return 'الآن'
    if (seconds < 3600) return `منذ ${Math.floor(seconds / 60)} دقيقة`
    if (seconds < 86400) return `منذ ${Math.floor(seconds / 3600)} ساعة`
    if (seconds < 604800) return `منذ ${Math.floor(seconds / 86400)} يوم`
    return date.toLocaleDateString('ar-SA')
  }

  return (
    <div
      className={clsx(
        'bg-surface rounded-lg border border-border border-l-4 p-4 transition-colors hover:bg-muted-5 cursor-pointer',
        priorityColors[notification.priority],
        statusColors[notification.status],
        notification.status === 'unread' && 'font-medium',
        detailed && 'cursor-default',
        className
      )}
      onClick={onClick}
    >
      <div className="flex items-start gap-4">
        {/* Icon */}
        <div className="text-2xl flex-shrink-0">
          {typeIcons[notification.type]}
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2 mb-2">
            <div className="flex-1">
              <h4 className={clsx('text-text', notification.status === 'unread' && 'font-semibold')}>
                {notification.title}
              </h4>
              <div className="flex items-center gap-2 mt-1">
                <span className="text-xs text-muted">{typeLabels[notification.type]}</span>
                <span className="text-xs text-muted">•</span>
                <span className="text-xs text-muted">{timeAgo(notification.createdAt)}</span>
                <span className="text-xs text-muted">•</span>
                <span className={clsx(
                  'text-xs',
                  notification.priority === 'urgent' ? 'text-danger' :
                  notification.priority === 'high' ? 'text-warning' :
                  notification.priority === 'medium' ? 'text-info' : 'text-muted'
                )}>
                  {priorityLabels[notification.priority]}
                </span>
              </div>
            </div>
            {notification.status === 'unread' && (
              <span className="w-2 h-2 rounded-full bg-primary flex-shrink-0"></span>
            )}
          </div>

          <p className="text-sm text-muted mb-3">{notification.message}</p>

          {detailed && notification.metadata && Object.keys(notification.metadata).length > 0 && (
            <div className="bg-muted-10 rounded-lg p-3 mb-3">
              <pre className="text-xs text-muted overflow-x-auto">
                {JSON.stringify(notification.metadata, null, 2)}
              </pre>
            </div>
          )}

          {/* Actions */}
          <div className="flex items-center gap-2 flex-wrap">
            {notification.actionUrl && notification.actionLabel && (
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onAction?.()
                }}
                className="px-3 py-1 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors text-sm"
              >
                {notification.actionLabel}
              </button>
            )}

            {notification.status === 'unread' && onMarkAsRead && (
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onMarkAsRead()
                }}
                className="px-3 py-1 border border-border rounded-lg hover:bg-muted-10 transition-colors text-sm"
              >
                تحديد كمقروء
              </button>
            )}

            {onArchive && notification.status !== 'archived' && (
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onArchive()
                }}
                className="px-3 py-1 border border-border rounded-lg hover:bg-muted-10 transition-colors text-sm"
              >
                أرشفة
              </button>
            )}

            {onDelete && (
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onDelete()
                }}
                className="px-3 py-1 text-danger hover:bg-danger-10 rounded-lg transition-colors text-sm"
              >
                حذف
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
