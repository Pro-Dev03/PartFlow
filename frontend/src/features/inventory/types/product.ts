export interface Product {
  id: string
  name: string
  barcode: string
  sku?: string
  category: string
  manufacturer: string
  model?: string
  condition: 'new' | 'used'
  grade?: 'A' | 'B' | 'C' | 'D'
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
