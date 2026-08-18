export interface DashboardStats {
  todaySales: number
  todayProfit: number
  inventoryValue: number
  outstandingDebts: number
  lowStockCount: number
  overdueDebtsCount: number
  todaySalesCount: number
  newCustomersCount: number
}

export interface Alert {
  id: string
  type: 'low-stock' | 'overdue-debt' | 'warranty-expiring' | 'inspection-needed'
  severity: 'low' | 'medium' | 'high' | 'critical'
  title: string
  description: string
  actionLabel?: string
  actionLink?: string
  createdAt: Date
}

export interface QuickAction {
  id: string
  label: string
  icon: string
  path: string
  color: 'primary' | 'success' | 'warning' | 'danger'
}

export interface RecentActivity {
  id: string
  type: 'sale' | 'purchase' | 'payment' | 'return' | 'inventory'
  title: string
  description: string
  amount?: number
  timestamp: Date
  user: string
}