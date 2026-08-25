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
};

// Products endpoints
export const productsApi = {
  list: (params?: { page?: number; per_page?: number; search?: string; category_id?: string }) => 
    apiClient.get('/products', params),
  get: (id: string) => apiClient.get(`/products/${id}`),
  create: (data: any) => apiClient.post('/products', data),
  update: (id: string, data: any) => apiClient.put(`/products/${id}`, data),
  delete: (id: string) => apiClient.delete(`/products/${id}`),
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
  delete: (id: string) => apiClient.delete(`/sales/${id}`),
  refund: (id: string, data: any) => apiClient.post(`/sales/${id}/refund`, data),
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
  delete: (id: string) => apiClient.delete(`/purchases/${id}`),
  receive: (id: string) => apiClient.post(`/purchases/${id}/receive`, {}),
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

// Returns endpoints
export const returnsApi = {
  list: (params?: { page?: number; per_page?: number; search?: string }) => 
    apiClient.get('/returns', params),
  get: (id: string) => apiClient.get(`/returns/${id}`),
  create: (data: any) => apiClient.post('/returns', data),
  update: (id: string, data: any) => apiClient.put(`/returns/${id}`, data),
};

// Reports endpoints
export const reportsApi = {
  sales: (params?: { start_date?: string; end_date?: string }) => 
    apiClient.get('/reports/sales', params),
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