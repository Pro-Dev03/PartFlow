export interface Debt {
  id: string
  customerId: string
  customerName: string
  customerPhone: string
  amount: number
  paidAmount: number
  remainingAmount: number
  dueDate: string
  status: 'pending' | 'partial' | 'paid' | 'overdue'
  saleId?: string
  notes?: string
  createdAt: string
  updatedAt: string
}

export interface DebtPayment {
  id: string
  debtId: string
  amount: number
  method: 'cash' | 'card' | 'transfer'
  reference?: string
  notes?: string
  createdAt: string
  createdBy: string
}

export interface DebtSummary {
  totalDebts: number
  totalAmount: number
  totalPaid: number
  totalRemaining: number
  overdueAmount: number
  overdueCount: number
  dueThisWeek: number
  dueThisMonth: number
}
