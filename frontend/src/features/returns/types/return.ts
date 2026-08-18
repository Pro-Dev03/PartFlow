export interface Return {
  id: string
  saleId: string
  saleDate: string
  customerId: string
  customerName: string
  items: ReturnItem[]
  totalAmount: number
  refundAmount: number
  reason: string
  status: 'pending' | 'approved' | 'rejected' | 'completed'
  notes?: string
  requestedAt: string
  processedAt?: string
  processedBy?: string
}

export interface ReturnItem {
  productId: string
  productName: string
  quantity: number
  price: number
  totalAmount: number
  condition: 'good' | 'damaged' | 'defective'
  reason: string
}

export interface ReturnSummary {
  totalReturns: number
  pendingReturns: number
  approvedReturns: number
  rejectedReturns: number
  totalRefundAmount: number
  thisMonth: number
}
