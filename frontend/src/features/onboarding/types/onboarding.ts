export interface OnboardingData {
  step: number
  organization: {
    name: string
    type: 'computer_store' | 'electronics' | 'repair' | 'trading'
    currency: 'ILS' | 'USD' | 'EUR' | 'GBP'
    language: 'ar' | 'he' | 'en'
  }
  user: {
    name: string
    email: string
    password: string
  }
  preferences: {
    businessHours: {
      sunday: { open: string; close: string; closed: boolean }
      monday: { open: string; close: string; closed: boolean }
      tuesday: { open: string; close: string; closed: boolean }
      wednesday: { open: string; close: string; closed: boolean }
      thursday: { open: string; close: string; closed: boolean }
      friday: { open: string; close: string; closed: boolean }
      saturday: { open: string; close: string; closed: boolean }
    }
    lowStockThreshold: number
    defaultWarrantyDays: number
  }
  completed: boolean
}

export interface OnboardingStep {
  id: number
  title: string
  description: string
  component: string
}
