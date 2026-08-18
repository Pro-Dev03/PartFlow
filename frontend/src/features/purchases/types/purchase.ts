export interface Purchase {
  id: string
  supplierId: string
  supplierName: string
  items: PurchaseItem[]
  totalCost: number
  paidAmount: number
  remainingAmount: number
  paymentStatus: 'paid' | 'partial' | 'pending'
  notes?: string
  createdAt: string
  updatedAt: string
  createdBy: string
}

export interface PurchaseItem {
  productId: string
  productName: string
  quantity: number
  cost: number
  totalCost: number
}

export interface PurchaseSummary {
  totalPurchases: number
  totalAmount: number
  totalPaid: number
  totalRemaining: number
  pendingAmount: number
  thisMonth: number
}
