export interface ImportJob {
  id: string
  type: 'products' | 'customers' | 'suppliers' | 'inventory'
  fileName: string
  fileSize: number
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'partial'
  totalRows: number
  processedRows: number
  failedRows: number
  errors?: ImportErrorError[]
  createdAt: string
  completedAt?: string
}

export interface ImportErrorError {
  row: number
  field: string
  message: string
  value?: string
}

export interface ExportJob {
  id: string
  type: 'products' | 'customers' | 'suppliers' | 'sales' | 'purchases' | 'inventory' | 'reports'
  format: 'csv' | 'excel' | 'pdf'
  filters?: Record<string, any>
  status: 'pending' | 'processing' | 'completed' | 'failed'
  fileName: string
  fileSize?: number
  downloadUrl?: string
  createdAt: string
  completedAt?: string
}

export interface FieldMapping {
  sourceField: string
  targetField: string
  required: boolean
}

export interface ImportPreview {
  headers: string[]
  rows: string[][]
  totalRows: number
  sampleRows: string[][]
  detectedEncoding: string
  detectedDelimiter: string
}
