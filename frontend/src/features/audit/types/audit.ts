export interface AuditLog {
  id: string
  userId: string
  userName: string
  action: string
  entityType: 'product' | 'customer' | 'sale' | 'purchase' | 'expense' | 'supplier' | 'inventory' | 'settings' | 'user'
  entityId: string
  entityName?: string
  changes?: AuditChange[]
  oldValues?: Record<string, any>
  newValues?: Record<string, any>
  ipAddress?: string
  userAgent?: string
  timestamp: string
  status: 'success' | 'failure'
  errorMessage?: string
}

export interface AuditChange {
  field: string
  oldValue: any
  newValue: any
}

export interface AuditSummary {
  totalLogs: number
  todayLogs: number
  thisWeekLogs: number
  thisMonthLogs: number
  byAction: Array<{
    action: string
    count: number
  }>
  byUser: Array<{
    userId: string
    userName: string
    count: number
  }>
  byEntityType: Array<{
    entityType: string
    count: number
  }>
}
