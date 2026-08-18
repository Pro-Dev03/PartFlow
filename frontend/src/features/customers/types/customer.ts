export interface Customer {
  id: string
  name: string
  phone: string
  email?: string
  address?: string
  city?: string
  notes?: string
  createdAt: string
  updatedAt: string
  totalPurchases: number
  totalSpent: number
  outstandingBalance: number
  lastPurchaseDate?: string
  isActive: boolean
}

export interface CustomerTimelineEvent {
  id: string
  type: 'purchase' | 'payment' | 'debt' | 'note' | 'contact'
  date: string
  description: string
  details?: string
  amount?: number
  user?: string
}

export interface CustomerStats {
  totalPurchases: number
  totalSpent: number
  averagePurchaseValue: number
  outstandingBalance: number
  paymentHistory: {
    onTime: number
    late: number
    overdue: number
  }
}
