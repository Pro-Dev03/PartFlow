export interface CartItem {
  id: string
  productId: string
  productName: string
  barcode: string
  serialNumber?: string
  condition: 'new' | 'used'
  price: number
  cost: number
  quantity: number
  availableStock: number
}

export interface Sale {
  id: string
  customerId?: string
  customerName?: string
  items: CartItem[]
  subtotal: number
  tax: number
  discount: number
  total: number
  paymentMethod: 'cash' | 'card' | 'transfer' | 'debt'
  paymentStatus: 'paid' | 'partial' | 'pending'
  paidAmount: number
  remainingAmount: number
  notes?: string
  createdAt: Date
  createdBy: string
}

export interface Customer {
  id: string
  name: string
  phone: string
  email?: string
  address?: string
  outstandingBalance: number
  totalPurchases: number
}

export interface PaymentMethod {
  id: string
  name: string
  icon: string
  value: 'cash' | 'card' | 'transfer' | 'debt'
}