import { clsx } from 'clsx'
import type { NotificationStats as NotificationStatsType } from '../types/notification'

interface NotificationStatsProps {
  stats: NotificationStatsType
  className?: string
}

export function NotificationStats({ stats, className }: NotificationStatsProps) {
  const typeLabels = {
    low_stock: 'انخفاض المخزون',
    overdue_debt: 'ديون متأخرة',
    warranty_expiring: 'انتهاء الضمان',
    purchase_alert: 'تنبيهات المشتريات',
    sale_completed: 'عمليات البيع',
    payment_received: 'المدفوعات',
    return_requested: 'المرتجعات',
    system_update: 'تحديثات النظام',
    custom: 'أخرى',
  }

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

  const priorityLabels = {
    low: 'منخفضة',
    medium: 'متوسطة',
    high: 'عالية',
    urgent: 'عاجلة',
  }

  const priorityColors = {
    low: 'text-muted',
    medium: 'text-info',
    high: 'text-warning',
    urgent: 'text-danger',
  }

  return (
    <div className={clsx('grid grid-cols-1 md:grid-cols-3 gap-4 mb-6', className)}>
      {/* Total Stats */}
      <div className="bg-surface rounded-lg p-4 border border-border">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-medium text-text">إجمالي الإشعارات</h3>
          <span className="text-2xl font-bold text-primary">{stats.total}</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted">غير مقروء:</span>
          <span className="text-sm font-semibold text-primary">{stats.unread}</span>
        </div>
      </div>

      {/* By Type */}
      <div className="bg-surface rounded-lg p-4 border border-border">
        <h3 className="font-medium text-text mb-4">حسب النوع</h3>
        <div className="space-y-2">
          {Object.entries(stats.byType).map(([type, count]) => (
            count > 0 && (
              <div key={type} className="flex items-center justify-between text-sm">
                <div className="flex items-center gap-2">
                  <span>{typeIcons[type as keyof typeof typeIcons]}</span>
                  <span className="text-muted">{typeLabels[type as keyof typeof typeLabels]}</span>
                </div>
                <span className="font-medium text-text">{count}</span>
              </div>
            )
          ))}
        </div>
      </div>

      {/* By Priority */}
      <div className="bg-surface rounded-lg p-4 border border-border">
        <h3 className="font-medium text-text mb-4">حسب الأولوية</h3>
        <div className="space-y-2">
          {Object.entries(stats.byPriority).map(([priority, count]) => (
            count > 0 && (
              <div key={priority} className="flex items-center justify-between text-sm">
                <span className="text-muted">{priorityLabels[priority as keyof typeof priorityLabels]}</span>
                <span className={clsx('font-medium', priorityColors[priority as keyof typeof priorityColors])}>
                  {count}
                </span>
              </div>
            )
          ))}
        </div>
      </div>
    </div>
  )
}
