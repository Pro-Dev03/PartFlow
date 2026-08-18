export interface SalesReport {
  totalSales: number
  totalRevenue: number
  averageSaleValue: number
  topProducts: Array<{
    productId: string
    productName: string
    quantity: number
    revenue: number
  }>
  salesByDay: Array<{
    date: string
    sales: number
    revenue: number
  }>
  salesByCategory: Array<{
    category: string
    sales: number
    revenue: number
  }>
}

export interface InventoryReport {
  totalProducts: number
  totalValue: number
  lowStockProducts: number
  outOfStockProducts: number
  productsByCategory: Array<{
    category: string
    count: number
    value: number
  }>
  productsByCondition: Array<{
    condition: 'new' | 'used'
    count: number
    value: number
  }>
}

export interface CustomerReport {
  totalCustomers: number
  activeCustomers: number
  totalDebts: number
  topCustomers: Array<{
    customerId: string
    customerName: string
    totalPurchases: number
    totalSpent: number
  }>
  customerAcquisition: Array<{
    month: string
    newCustomers: number
  }>
}

export interface ExpenseReport {
  totalExpenses: number
  expensesByCategory: Array<{
    category: string
    amount: number
  }>
  expensesByMonth: Array<{
    month: string
    amount: number
  }>
}

export interface ProfitLossReport {
  totalRevenue: number
  totalExpenses: number
  grossProfit: number
  netProfit: number
  profitMargin: number
  monthlyProfitLoss: Array<{
    month: string
    revenue: number
    expenses: number
    profit: number
  }>
}

export type ReportType = 'sales' | 'inventory' | 'customers' | 'expenses' | 'profit_loss'
export type DateRange = 'today' | 'week' | 'month' | 'quarter' | 'year' | 'custom'
