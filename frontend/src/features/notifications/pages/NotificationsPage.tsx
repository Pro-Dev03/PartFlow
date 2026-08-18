import { useState } from 'react'
import { clsx } from 'clsx'
import { NotificationItem } from '../components/NotificationItem'
import { NotificationStats } from '../components/NotificationStats'
import { EmptyState } from '@/components/feedback'
import type { Notification, NotificationStats as NotificationStatsType } from '../types/notification'

type FilterType = 'all' | 'unread' | 'read' | 'archived'
type PriorityFilter = 'all' | 'low' | 'medium' | 'high' | 'urgent'

export function NotificationsPage() {
  const [filter, setFilter] = useState<FilterType>('all')
  const [priorityFilter, setPriorityFilter] = useState<PriorityFilter>('all')
  const [selectedNotification, setSelectedNotification] = useState<Notification | null>(null)

  // TODO: Fetch notifications from API
  const notifications: Notification[] = []
  const stats: NotificationStatsType = {
    total: 0,
    unread: 0,
    byType: {
      low_stock: 0,
      overdue_debt: 0,
      warranty_expiring: 0,
      purchase_alert: 0,
      sale_completed: 0,
      payment_received: 0,
      return_requested: 0,
      system_update: 0,
      custom: 0,
    },
    byPriority: {
      low: 0,
      medium: 0,
      high: 0,
      urgent: 0,
    },
  }

  const filteredNotifications = notifications.filter(notification => {
    if (filter !== 'all' && notification.status !== filter) return false
    if (priorityFilter !== 'all' && notification.priority !== priorityFilter) return false
    return true
  }).sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())

  const handleMarkAsRead = (notificationId: string) => {
    // TODO: Mark notification as read
    console.log('Mark as read:', notificationId)
  }

  const handleMarkAllAsRead = () => {
    // TODO: Mark all notifications as read
    console.log('Mark all as read')
  }

  const handleArchive = (notificationId: string) => {
    // TODO: Archive notification
    console.log('Archive:', notificationId)
  }

  const handleDelete = (notificationId: string) => {
    // TODO: Delete notification
    console.log('Delete:', notificationId)
  }

  const handleAction = (notification: Notification) => {
    if (notification.actionUrl) {
      // TODO: Navigate to action URL
      console.log('Navigate to:', notification.actionUrl)
    }
  }

  if (selectedNotification) {
    return (
      <div className="container mx-auto p-6">
        <button
          onClick={() => setSelectedNotification(null)}
          className="text-muted hover:text-text mb-4 inline-flex items-center gap-2"
        >
          ← العودة للإشعارات
        </button>
        <NotificationItem
          notification={selectedNotification}
          detailed
          onMarkAsRead={() => handleMarkAsRead(selectedNotification.id)}
          onArchive={() => handleArchive(selectedNotification.id)}
          onDelete={() => handleDelete(selectedNotification.id)}
          onAction={() => handleAction(selectedNotification)}
        />
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-text mb-2">مركز الإشعارات</h1>
            <p className="text-muted">جميع التنبيهات والتحديثات</p>
          </div>
          {stats.unread > 0 && (
            <button
              onClick={handleMarkAllAsRead}
              className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              تحديد الكل كمقروء
            </button>
          )}
        </div>
      </div>

      {/* Stats */}
      <NotificationStats stats={stats} />

      {/* Filters */}
      <div className="bg-surface rounded-lg p-4 mb-6 space-y-4">
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted self-center">الحالة:</span>
          {(['all', 'unread', 'read', 'archived'] as FilterType[]).map((status) => (
            <button
              key={status}
              onClick={() => setFilter(status)}
              className={clsx(
                'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
                filter === status
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted hover:bg-muted-80'
              )}
            >
              {status === 'all' && 'الكل'}
              {status === 'unread' && 'غير مقروء'}
              {status === 'read' && 'مقروء'}
              {status === 'archived' && 'مؤرشف'}
            </button>
          ))}
        </div>

        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted self-center">الأولوية:</span>
          {(['all', 'low', 'medium', 'high', 'urgent'] as PriorityFilter[]).map((priority) => (
            <button
              key={priority}
              onClick={() => setPriorityFilter(priority)}
              className={clsx(
                'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
                priorityFilter === priority
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted hover:bg-muted-80'
              )}
            >
              {priority === 'all' && 'الكل'}
              {priority === 'low' && 'منخفضة'}
              {priority === 'medium' && 'متوسطة'}
              {priority === 'high' && 'عالية'}
              {priority === 'urgent' && 'عاجلة'}
            </button>
          ))}
        </div>
      </div>

      {/* Notifications List */}
      {filteredNotifications.length > 0 ? (
        <div className="space-y-3">
          {filteredNotifications.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
              onMarkAsRead={() => handleMarkAsRead(notification.id)}
              onArchive={() => handleArchive(notification.id)}
              onDelete={() => handleDelete(notification.id)}
              onAction={() => handleAction(notification)}
              onClick={() => setSelectedNotification(notification)}
            />
          ))}
        </div>
      ) : (
        <EmptyState
          icon="🔔"
          title="لا توجد إشعارات"
          description="لا توجد إشعارات مطابقة للفلاتر الحالية"
        />
      )}
    </div>
  )
}
