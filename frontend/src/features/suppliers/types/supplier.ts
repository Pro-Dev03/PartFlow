export interface Supplier {
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
  totalAmount: number
  paidAmount: number
  outstandingBalance: number
  lastPurchaseDate?: string
  isActive: boolean
}

export interface SupplierTimelineEvent {
  id: string
  type: 'purchase' | 'payment' | 'note' | 'contact'
  date: string
  description: string
  details?: string
  amount?: number
  user?: string
}

export interface SupplierStats {
  totalPurchases: number
  totalAmount: number
  averagePurchaseValue: number
  outstandingBalance: number
  paymentHistory: {
    onTime: number
    late: number
    overdue: number
  }
}
