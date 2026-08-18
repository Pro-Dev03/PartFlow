export interface Warranty {
  id: string
  productId: string
  productName: string
  customerId?: string
  customerName?: string
  type: 'manufacturer' | 'seller' | 'extended'
  duration: number // in days
  startDate: string
  endDate: string
  status: 'active' | 'expired' | 'claimed' | 'cancelled'
  terms?: string
  notes?: string
  createdAt: string
  updatedAt: string
}

export interface WarrantyClaim {
  id: string
  warrantyId: string
  claimedAt: string
  reason: string
  status: 'pending' | 'approved' | 'rejected' | 'completed'
  resolution?: string
  notes?: string
  processedAt?: string
  processedBy?: string
}

export interface WarrantySummary {
  totalWarranties: number
  activeWarranties: number
  expiringSoon: number // within 30 days
  expiredWarranties: number
  claimedWarranties: number
  thisMonthExpiring: number
}
