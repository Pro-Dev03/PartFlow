// ==================== SHARED TYPES ====================
// Centralized type definitions used across the application

// ==================== COMMON TYPES ====================

export type ProductCondition = 'new' | 'used'
export type ProductGrade = 'A' | 'B' | 'C' | 'D'
export type PaymentMethod = 'cash' | 'card' | 'transfer' | 'debt'
export type PaymentStatus = 'paid' | 'partial' | 'pending'

// ==================== CUSTOMER ====================

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

export interface CustomerTimelineEvent {
  id: string
  type: 'purchase' | 'payment' | 'debt' | 'note' | 'contact'
  date: string
  description: string
  details?: string
  amount?: number
  user?: string
}

// ==================== PRODUCT ====================

export interface Product {
  id: string
  name: string
  barcode: string
  sku?: string
  category: string
  manufacturer: string
  model?: string
  condition: ProductCondition
  grade?: ProductGrade
  cost: number
  price: number
  stock: number
  minStock?: number
  location?: string
  serialNumber?: string
  warranty?: {
    enabled: boolean
    duration?: number
    type?: string
  }
  description?: string
  images?: string[]
  supplierId?: string
  supplierName?: string
  createdAt: string
  updatedAt: string
}

export interface ProductTimelineEvent {
  id: string
  type: 'purchase' | 'sale' | 'inspection' | 'price_change' | 'stock_adjustment' | 'transfer'
  date: string
  description: string
  details?: string
  user?: string
}

export interface Inspection {
  id: string
  productId: string
  inspector: string
  date: string
  powerTest: 'passed' | 'failed' | 'skipped'
  temperatureTest: 'passed' | 'failed' | 'skipped'
  performanceTest: 'passed' | 'failed' | 'skipped'
  portsTest: 'passed' | 'failed' | 'skipped'
  visualInspection: 'passed' | 'failed' | 'skipped'
  serialVerification: 'passed' | 'failed' | 'skipped'
  overallResult: 'passed' | 'failed'
  notes?: string
  photos?: string[]
  createdAt: string
}

// ==================== SALES ====================

export interface CartItem {
  id: string
  productId: string
  productName: string
  barcode: string
  serialNumber?: string
  condition: ProductCondition
  price: number
  cost: number
  quantity: number
  availableStock: number
  discount?: number
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
  paymentMethod: PaymentMethod
  paymentStatus: PaymentStatus
  paidAmount: number
  remainingAmount: number
  notes?: string
  createdAt: Date
  createdBy: string
}

export interface PaymentMethodInfo {
  id: string
  name: string
  icon: string
  value: PaymentMethod
}

// ==================== API ====================

export interface ApiResponse<T> {
  success: boolean
  data: T
  meta?: {
    page?: number
    limit?: number
    total?: number
    totalPages?: number
  }
  error?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  meta: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
}

// ==================== FILTERS ====================

export interface FilterOption {
  label: string
  value: string
  count?: number
}

export interface SortOption {
  label: string
  value: string
  direction: 'asc' | 'desc'
}
