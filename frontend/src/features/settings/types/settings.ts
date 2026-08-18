export interface OrganizationSettings {
  id: string
  name: string
  type: 'computer_store' | 'electronics' | 'repair' | 'trading'
  currency: 'ILS' | 'USD' | 'EUR' | 'GBP'
  language: 'ar' | 'he' | 'en'
  timezone: string
  logo?: string
  address?: string
  phone?: string
  email?: string
  taxId?: string
  businessHours?: {
    sunday: { open: string; close: string; closed: boolean }
    monday: { open: string; close: string; closed: boolean }
    tuesday: { open: string; close: string; closed: boolean }
    wednesday: { open: string; close: string; closed: boolean }
    thursday: { open: string; close: string; closed: boolean }
    friday: { open: string; close: string; closed: boolean }
    saturday: { open: string; close: string; closed: boolean }
  }
  createdAt: string
  updatedAt: string
}

export interface User {
  id: string
  name: string
  email: string
  phone?: string
  role: 'owner' | 'manager' | 'employee' | 'accountant'
  avatar?: string
  isActive: boolean
  lastLogin?: string
  createdAt: string
  updatedAt: string
}

export interface NotificationSettings {
  id: string
  userId: string
  email: {
    lowStock: boolean
    overdueDebts: boolean
    warrantyExpiring: boolean
    purchaseAlerts: boolean
    salesReports: boolean
    systemUpdates: boolean
  }
  push: {
    lowStock: boolean
    overdueDebts: boolean
    warrantyExpiring: boolean
    purchaseAlerts: boolean
    salesReports: boolean
    systemUpdates: boolean
  }
  inApp: {
    lowStock: boolean
    overdueDebts: boolean
    warrantyExpiring: boolean
    purchaseAlerts: boolean
    salesReports: boolean
    systemUpdates: boolean
  }
}

export interface SystemSettings {
  id: string
  lowStockThreshold: number
  defaultWarrantyDays: number
  allowNegativeStock: boolean
  requireBarcode: boolean
  autoBackup: boolean
  backupFrequency: 'daily' | 'weekly' | 'monthly'
  retentionDays: number
  barcodePrefix: string
  priceRounding: 'none' | 'nearest' | 'up' | 'down'
  decimalPlaces: number
}

export interface IntegrationSettings {
  id: string
  name: string
  type: 'payment' | 'shipping' | 'accounting' | 'notification' | 'custom'
  isEnabled: boolean
  config: Record<string, any>
  lastSync?: string
  status: 'active' | 'error' | 'disconnected'
}

export interface BackupInfo {
  id: string
  name: string
  size: number
  createdAt: string
  type: 'manual' | 'automatic'
  status: 'completed' | 'failed' | 'in_progress'
}
