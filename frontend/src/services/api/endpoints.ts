import { apiClient } from './client';

// Auth endpoints
export const authApi = {
  login: (email: string, password: string) =>
    apiClient.post('/auth/login', { email, password }),
  logout: () => apiClient.post('/auth/logout', {}),
  refreshToken: () => apiClient.post('/auth/refresh', {}),
  forgotPassword: (email: string) =>
    apiClient.post('/auth/forgot-password', { email }),
  resetPassword: (token: string, password: string) =>
    apiClient.post('/auth/reset-password', { token, password }),
};

// Dashboard endpoints
export const dashboardApi = {
  getStats: () => apiClient.get('/dashboard/stats'),
  // Aggregation endpoints (ARCHITECTURE-PRINCIPLES.md)
  getDailySalesSummary: (date?: string) => apiClient.get('/aggregations/daily-sales', { date }),
  getMonthlySalesSummary: (year?: number, month?: number) => apiClient.get('/aggregations/monthly-sales', { year, month }),
  getDailyInventorySummary: (date?: string) => apiClient.get('/aggregations/daily-inventory', { date }),
  getMonthlyInventorySummary: (year?: number, month?: number) => apiClient.get('/aggregations/monthly-inventory', { year, month }),
  getDailyDebtSummary: (date?: string) => apiClient.get('/aggregations/daily-debt', { date }),
  getMonthlyDebtSummary: (year?: number, month?: number) => apiClient.get('/aggregations/monthly-debt', { year, month }),
  getDailyProfitSummary: (date?: string) => apiClient.get('/aggregations/daily-profit', { date }),
  getMonthlyProfitSummary: (year?: number, month?: number) => apiClient.get('/aggregations/monthly-profit', { year, month }),
  getAggregationStatus: () => apiClient.get('/aggregations/status'),
  updateAggregations: (startDate?: string, endDate?: string, force?: boolean) =>
    apiClient.post('/aggregations/update', { start_date: startDate, end_date: endDate, force }),
};

// Products endpoints
export const productsApi = {
  list: (params?: { page?: number; per_page?: number; search?: string; category_id?: string }) => 
    apiClient.get('/products', params),
  get: (id: string) => apiClient.get(`/products/${id}`),
  create: (data: any) => apiClient.post('/products', data),
  update: (id: string, data: any) => apiClient.put(`/products/${id}`, data),
  delete: (id: string) => apiClient.delete(`/products/${id}`),
  archive: (id: string) => apiClient.post(`/products/${id}/archive`),
};

// Categories endpoints
export const categoriesApi = {
  list: () => apiClient.get('/categories'),
  get: (id: string) => apiClient.get(`/categories/${id}`),
  create: (data: any) => apiClient.post('/categories', data),
  update: (id: string, data: any) => apiClient.put(`/categories/${id}`, data),
  delete: (id: string) => apiClient.delete(`/categories/${id}`),
};

// Inventory endpoints
export const inventoryApi = {
  list: (params?: { page?: number; per_page?: number; search?: string; condition?: string }) =>
    apiClient.get('/inventory/items', params),
  listWithSupplier: (params?: { 
    page?: number; 
    per_page?: number; 
    search?: string; 
    condition?: string;
    supplier_id?: string;
    purchase_date_from?: string;
    purchase_date_to?: string;
    min_purchase_cost?: number;
    max_purchase_cost?: number;
  }) =>
    apiClient.get('/inventory/items-with-supplier', params),
  get: (id: string) => apiClient.get(`/inventory/items/${id}`),
  create: (data: any) => apiClient.post('/inventory/items', data),
  update: (id: string, data: any) => apiClient.put(`/inventory/items/${id}`, data),
  delete: (id: string) => apiClient.delete(`/inventory/items/${id}`),
  movements: (itemId: string) => apiClient.get(`/inventory/items/${itemId}/history`),
  createTradeIn: (data: any) => apiClient.post('/inventory/trade-ins', data),
};

// Part Types endpoints
export const partTypesApi = {
  list: () => apiClient.get('/part-types'),
  get: (id: string) => apiClient.get(`/part-types/${id}`),
  create: (data: any) => apiClient.post('/part-types', data),
  update: (id: string, data: any) => apiClient.put(`/part-types/${id}`, data),
  delete: (id: string) => apiClient.delete(`/part-types/${id}`),
  getSpecifications: (id: string) => apiClient.get(`/part-types/${id}/specifications`),
};

// Specifications endpoints
export const specificationsApi = {
  list: () => apiClient.get('/specifications'),
  create: (data: any) => apiClient.post('/specifications', data),
};

// Type Specifications endpoints
export const typeSpecsApi = {
  link: (data: any) => apiClient.post('/type-specifications', data),
  unlink: (partTypeId: string, specId: string) => apiClient.delete(`/type-specifications/${partTypeId}/${specId}`),
};

// Item Specifications endpoints
export const itemSpecsApi = {
  get: (itemId: string) => apiClient.get(`/item-specifications/${itemId}`),
  update: (itemId: string, data: any) => apiClient.put(`/item-specifications/${itemId}`, data),
};

// Sales endpoints
export const salesApi = {
  list: (params?: { page?: number; per_page?: number; search?: string }) =>
    apiClient.get('/sales', params),
  get: (id: string) => apiClient.get(`/sales/${id}`),
  create: (data: any) => apiClient.post('/sales', data),
  update: (id: string, data: any) => apiClient.put(`/sales/${id}`, data),
  // SmartDelete - now returns SmartDeleteResult (ARCHITECTURE-PRINCIPLES.md)
  delete: (id: string) => apiClient.delete(`/sales/${id}`),
  refund: (id: string, data: any) => apiClient.post(`/sales/${id}/refund`, data),
};

// Payments endpoints (ARCHITECTURE-PRINCIPLES.md)
export const paymentsApi = {
  list: (params?: { page?: number; per_page?: number; type?: string }) =>
    apiClient.get('/payments', params),
  get: (id: string) => apiClient.get(`/payments/${id}`),
  create: (data: any) => apiClient.post('/payments', data),
  // SmartDelete - now returns SmartDeleteResult (ARCHITECTURE-PRINCIPLES.md)
  delete: (id: string) => apiClient.delete(`/payments/${id}`),
};

// Customers endpoints
export const customersApi = {
  list: (params?: { page?: number; per_page?: number; search?: string; is_active?: boolean }) => 
    apiClient.get('/customers', params),
  get: (id: string) => apiClient.get(`/customers/${id}`),
  create: (data: any) => apiClient.post('/customers', data),
  update: (id: string, data: any) => apiClient.put(`/customers/${id}`, data),
  delete: (id: string) => apiClient.delete(`/customers/${id}`),
  ledger: (id: string, params?: { page?: number; per_page?: number }) => 
    apiClient.get(`/customers/${id}/ledger`, params),
  ledgerSummary: (id: string) => 
    apiClient.get(`/customers/${id}/ledger/summary`),
};

// Debts endpoints - Note: Debts are managed under customers in the backend
export const debtsApi = {
  list: (params?: { page?: number; per_page?: number }) => 
    apiClient.get('/customers/overdue', params),
  get: (customerId: string, debtId: string) => apiClient.get(`/customers/${customerId}/debts/${debtId}`),
  recordPayment: (customerId: string, data: any) => apiClient.post(`/customers/${customerId}/debt-payments`, data),
  getDebtEntries: (customerId: string) => apiClient.get(`/customers/${customerId}/debts`),
  getDebtCollections: (customerId: string) => apiClient.get(`/customers/${customerId}/debt-collections`),
  getPendingCollections: () => apiClient.get('/debt-collections/pending'),
};

// Suppliers endpoints
export const suppliersApi = {
  list: (params?: { page?: number; per_page?: number; search?: string }) =>
    apiClient.get('/suppliers', params),
  get: (id: string) => apiClient.get(`/suppliers/${id}`),
  create: (data: any) => apiClient.post('/suppliers', data),
  update: (id: string, data: any) => apiClient.put(`/suppliers/${id}`, data),
  delete: (id: string) => apiClient.delete(`/suppliers/${id}`),
  ledger: (id: string, params?: { page?: number; per_page?: number }) => 
    apiClient.get(`/suppliers/${id}/ledger`, params),
  ledgerSummary: (id: string) => 
    apiClient.get(`/suppliers/${id}/ledger/summary`),
  getSupplierInventory: (id: string) => apiClient.get(`/suppliers/${id}/inventory`),
};

// Purchases endpoints
export const purchasesApi = {
  list: (params?: { page?: number; per_page?: number; search?: string }) =>
    apiClient.get('/purchases', params),
  get: (id: string) => apiClient.get(`/purchases/${id}`),
  create: (data: any) => apiClient.post('/purchases', data),
  update: (id: string, data: any) => apiClient.put(`/purchases/${id}`, data),
  // SmartDelete - now returns SmartDeleteResult (ARCHITECTURE-PRINCIPLES.md)
  delete: (id: string) => apiClient.delete(`/purchases/${id}`),
  receive: (id: string) => apiClient.post(`/purchases/${id}/receive`, {}),
  cancel: (id: string) => apiClient.post(`/purchases/${id}/cancel`, {}),
  reverse: (id: string, reason: string) => apiClient.post(`/purchases/${id}/reverse`, { reason }),
  // Get used items info for blocked deletion
  getUsedItemsInfo: (id: string) => apiClient.get(`/purchases/${id}/used-items`),
};

// Expenses endpoints
export const expensesApi = {
  list: (params?: { page?: number; per_page?: number; search?: string }) => 
    apiClient.get('/expenses', params),
  get: (id: string) => apiClient.get(`/expenses/${id}`),
  create: (data: any) => apiClient.post('/expenses', data),
  update: (id: string, data: any) => apiClient.put(`/expenses/${id}`, data),
  delete: (id: string) => apiClient.delete(`/expenses/${id}`),
};

// Reports endpoints
export const reportsApi = {
  sales: (params?: { start_date?: string; end_date?: string }) => 
    apiClient.get('/reports/sales', params),
  netSales: (params?: { start_date?: string; end_date?: string }) => 
    apiClient.get('/reports/net-sales', params),
  profit: (params?: { start_date?: string; end_date?: string }) => 
    apiClient.get('/reports/profit', params),
  inventory: () => 
    apiClient.get('/reports/inventory'),
  debts: () => 
    apiClient.get('/reports/debts'),
  products: () => 
    apiClient.get('/reports/products'),
  suppliers: () => 
    apiClient.get('/reports/suppliers'),
  expenses: (params?: { start_date?: string; end_date?: string }) => 
    apiClient.get('/reports/expenses', params),
  returns: (params?: { start_date?: string; end_date?: string }) => 
    apiClient.get('/reports/returns', params),
};

// Settings endpoints
export const settingsApi = {
  getUsers: (params?: { page?: number; per_page?: number }) =>
    apiClient.get('/settings/users', params),
  createUser: (data: any) => apiClient.post('/settings/users', data),
  updateUser: (id: string, data: any) => apiClient.put(`/settings/users/${id}`, data),
  deleteUser: (id: string) => apiClient.delete(`/settings/users/${id}`),
  getTaxRate: () => apiClient.get('/settings/tax-rate'),
  updateTaxRate: (taxRate: number) => apiClient.put('/settings/tax-rate', { tax_rate: taxRate }),
  getPublicSettings: () => apiClient.get('/settings/public'),
  getSetting: (key: string) => apiClient.get(`/settings/${key}`),
  updateSetting: (key: string, value: string) => apiClient.put(`/settings/${key}`, { value }),
  deleteAllData: () => apiClient.delete('/settings/database'),
};

// Barcode endpoints
export const barcodeApi = {
  scan: (barcode: string) => apiClient.post('/barcode/scan', { barcode }),
  lookup: (barcode: string) => apiClient.get(`/barcodes/${barcode}`),
  lookupProduct: (barcode: string) => apiClient.get(`/barcodes/product/${barcode}`),
  lookupBySKU: (sku: string) => apiClient.get(`/barcodes/sku/${sku}`),
};

// Inspections endpoints
export const inspectionsApi = {
  list: (params?: { page?: number; per_page?: number; search?: string }) => 
    apiClient.get('/inspections', params),
  get: (id: string) => apiClient.get(`/inspections/${id}`),
  create: (data: any) => apiClient.post('/inspections', data),
  update: (id: string, data: any) => apiClient.put(`/inspections/${id}`, data),
  delete: (id: string) => apiClient.delete(`/inspections/${id}`),
};

// Returns endpoints (ENHANCED-RETURNS-SYSTEM.md)
export const returnsApi = {
  list: (params?: { 
    page?: number; 
    per_page?: number; 
    customer_id?: string;
    sale_id?: string;
    status?: string;
    return_type?: string;
    refund_method?: string;
    start_date?: string;
    end_date?: string;
    search?: string;
    sort_by?: string;
    sort_order?: string;
  }) => 
    apiClient.get('/returns', params),
  get: (id: string) => apiClient.get(`/returns/${id}`),
  getWithItems: (id: string) => apiClient.get(`/returns/${id}/with-items`),
  create: (data: any) => apiClient.post('/returns', data),
  update: (id: string, data: any) => apiClient.put(`/returns/${id}`, data),
  delete: (id: string) => apiClient.delete(`/returns/${id}`),
  approve: (id: string) => apiClient.post(`/returns/${id}/approve`),
  reject: (id: string) => apiClient.post(`/returns/${id}/reject`),
  processRefund: (id: string) => apiClient.post(`/returns/${id}/refund`),
  complete: (id: string) => apiClient.post(`/returns/${id}/complete`),
  getBySale: (saleId: string) => apiClient.get(`/returns/sale/${saleId}`),
  getByCustomer: (customerId: string) => apiClient.get(`/returns/customer/${customerId}`),
  getPending: () => apiClient.get('/returns/pending'),
  getStatistics: () => apiClient.get('/returns/statistics'),
  getMonthlyAnalysis: () => apiClient.get('/returns/analysis/monthly'),
  getSalesReturnsAnalysis: () => apiClient.get('/returns/analysis/sales-returns'),
  addItem: (returnId: string, data: any) => apiClient.post(`/returns/${returnId}/items`, data),
  updateItem: (itemId: string, data: any) => apiClient.put(`/returns/items/${itemId}`, data),
  deleteItem: (itemId: string) => apiClient.delete(`/returns/items/${itemId}`),
  processInspection: (itemId: string, data: any) => apiClient.post(`/returns/items/${itemId}/inspection`, data),
  validateQuantity: (saleItemId: string, quantity: number) => 
    apiClient.get(`/returns/validate/${saleItemId}`, { quantity }),
  getSummary: () => apiClient.get('/returns/summary'),
  reverse: (id: string) => apiClient.post(`/returns/${id}/reverse`),
};

// Acquisitions endpoints (USED-PARTS-ACQUISITION.md)
export const acquisitionsApi = {
  list: (params?: { 
    page?: number; 
    per_page?: number; 
    type?: string; 
    status?: string; 
    payment_status?: string;
    supplier_id?: string;
    customer_id?: string;
    search?: string;
    sort_by?: string;
    sort_order?: string;
  }) => 
    apiClient.get('/acquisitions', params),
  get: (id: string) => apiClient.get(`/acquisitions/${id}`),
  create: (data: any) => apiClient.post('/acquisitions', data),
  updateStatus: (id: string, status: string) => 
    apiClient.put(`/acquisitions/${id}/status`, { status }),
  createPayment: (id: string, data: any) => 
    apiClient.post(`/acquisitions/${id}/payments`, data),
  getAging: (alertLevel?: string) => 
    apiClient.get('/acquisitions/aging', { alert_level: alertLevel }),
  getSellerBalances: () => 
    apiClient.get('/acquisitions/seller-balances'),
  addRepairCost: (itemId: string, data: any) => 
    apiClient.post(`/acquisitions/items/${itemId}/repair-cost`, data),
  getItemHistory: (itemId: string) => 
    apiClient.get(`/inventory/${itemId}/history`),
};

// Notifications endpoints
export const notificationsApi = {
  list: (params?: { page?: number; per_page?: number }) => 
    apiClient.get('/notifications', params),
  markAsRead: (id: string) => apiClient.put(`/notifications/${id}/read`, {}),
  markAllAsRead: () => apiClient.put('/notifications/read-all', {}),
  getUnreadCount: () => apiClient.get('/notifications/unread-count'),
  updatePreferences: (data: any) => apiClient.put('/notifications/preferences', data),
};

// Global search endpoint
export const searchApi = {
  global: (query: string) => apiClient.get(`/search?q=${encodeURIComponent(query)}`),
};

// General API client for custom endpoints
export const api = {
  get: (endpoint: string, params?: any) => apiClient.get(endpoint, params),
  post: (endpoint: string, data?: any) => apiClient.post(endpoint, data),
  put: (endpoint: string, data?: any) => apiClient.put(endpoint, data),
  delete: (endpoint: string) => apiClient.delete(endpoint),
};