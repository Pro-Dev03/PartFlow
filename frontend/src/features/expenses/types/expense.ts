export interface Expense {
  id: string
  categoryId: string
  categoryName: string
  amount: number
  date: string
  description: string
  receipt?: string
  notes?: string
  createdAt: string
  updatedAt: string
  createdBy: string
}

export interface ExpenseCategory {
  id: string
  name: string
  description?: string
  budget?: number
  color: string
}

export interface ExpenseSummary {
  totalExpenses: number
  thisMonth: number
  byCategory: Array<{
    categoryId: string
    categoryName: string
    amount: number
    percentage: number
  }>
  monthlyTrend: Array<{
    month: string
    amount: number
  }>
}
