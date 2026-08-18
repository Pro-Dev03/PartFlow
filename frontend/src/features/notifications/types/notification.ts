export interface Notification {
  id: string
  userId: string
  type: 'low_stock' | 'overdue_debt' | 'warranty_expiring' | 'purchase_alert' | 'sale_completed' | 'payment_received' | 'return_requested' | 'system_update' | 'custom'
  title: string
  message: string
  priority: 'low' | 'medium' | 'high' | 'urgent'
  status: 'unread' | 'read' | 'archived'
  actionUrl?: string
  actionLabel?: string
  metadata?: Record<string, any>
  createdAt: string
  readAt?: string
}

export interface NotificationStats {
  total: number
  unread: number
  byType: {
    low_stock: number
    overdue_debt: number
    warranty_expiring: number
    purchase_alert: number
    sale_completed: number
    payment_received: number
    return_requested: number
    system_update: number
    custom: number
  }
  byPriority: {
    low: number
    medium: number
    high: number
    urgent: number
  }
}

export interface NotificationPreferences {
  email: {
    enabled: boolean
    lowStock: boolean
    overdueDebts: boolean
    warrantyExpiring: boolean
    purchaseAlerts: boolean
    salesReports: boolean
    systemUpdates: boolean
  }
  push: {
    enabled: boolean
    lowStock: boolean
    overdueDebts: boolean
    warrantyExpiring: boolean
    purchaseAlerts: boolean
    salesReports: boolean
    systemUpdates: boolean
  }
  inApp: {
    enabled: boolean
    lowStock: boolean
    overdueDebts: boolean
    warrantyExpiring: boolean
    purchaseAlerts: boolean
    salesReports: boolean
    systemUpdates: boolean
  }
}
