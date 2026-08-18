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
  getStats: () => apiClient.get('/dashboard/stats', {}),
};

// Products endpoints
export const productsApi = {
  list: () => apiClient.get('/products', {}),
  get: (id: string) => apiClient.get(`/products/${id}`, {}),
  create: (data: any) => apiClient.post('/products', data),
  update: (id: string, data: any) => apiClient.put(`/products/${id}`, data),
  delete: (id: string) => apiClient.delete(`/products/${id}`),
};

// Inventory endpoints
export const inventoryApi = {
  list: () => apiClient.get('/inventory', {}),
  get: (id: string) => apiClient.get(`/inventory/${id}`, {}),
  create: (data: any) => apiClient.post('/inventory', data),
  update: (id: string, data: any) => apiClient.put(`/inventory/${id}`, data),
  delete: (id: string) => apiClient.delete(`/inventory/${id}`),
  movements: (itemId: string) => apiClient.get(`/inventory/${itemId}/movements`, {}),
};

// Sales endpoints
export const salesApi = {
  list: () => apiClient.get('/sales', {}),
  get: (id: string) => apiClient.get(`/sales/${id}`, {}),
  create: (data: any) => apiClient.post('/sales', data),
  update: (id: string, data: any) => apiClient.put(`/sales/${id}`, data),
  delete: (id: string) => apiClient.delete(`/sales/${id}`),
  refund: (id: string, data: any) => apiClient.post(`/sales/${id}/refund`, data),
};

// Customers endpoints
export const customersApi = {
  list: () => apiClient.get('/customers', {}),
  get: (id: string) => apiClient.get(`/customers/${id}`, {}),
  create: (data: any) => apiClient.post('/customers', data),
  update: (id: string, data: any) => apiClient.put(`/customers/${id}`, data),
  delete: (id: string) => apiClient.delete(`/customers/${id}`),
  ledger: (id: string) => apiClient.get(`/customers/${id}/ledger`, {}),
};

// Debts endpoints - Note: Debts are managed under customers in the backend
export const debtsApi = {
  list: () => apiClient.get('/customers/overdue', {}),
  get: (customerId: string, debtId: string) => apiClient.get(`/customers/${customerId}/debts/${debtId}`, {}),
  recordPayment: (customerId: string, data: any) => apiClient.post(`/customers/${customerId}/debt-payments`, data),
  getDebtEntries: (customerId: string) => apiClient.get(`/customers/${customerId}/debts`, {}),
  getDebtCollections: (customerId: string) => apiClient.get(`/customers/${customerId}/debt-collections`, {}),
  getPendingCollections: () => apiClient.get('/debt-collections/pending', {}),
};

// Suppliers endpoints
export const suppliersApi = {
  list: () => apiClient.get('/suppliers', {}),
  get: (id: string) => apiClient.get(`/suppliers/${id}`, {}),
  create: (data: any) => apiClient.post('/suppliers', data),
  update: (id: string, data: any) => apiClient.put(`/suppliers/${id}`, data),
  delete: (id: string) => apiClient.delete(`/suppliers/${id}`),
  ledger: (id: string) => apiClient.get(`/suppliers/${id}/ledger`, {}),
};

// Purchases endpoints
export const purchasesApi = {
  list: () => apiClient.get('/purchases', {}),
  get: (id: string) => apiClient.get(`/purchases/${id}`, {}),
  create: (data: any) => apiClient.post('/purchases', data),
  update: (id: string, data: any) => apiClient.put(`/purchases/${id}`, data),
  delete: (id: string) => apiClient.delete(`/purchases/${id}`),
};

// Expenses endpoints
export const expensesApi = {
  list: () => apiClient.get('/expenses', {}),
  get: (id: string) => apiClient.get(`/expenses/${id}`, {}),
  create: (data: any) => apiClient.post('/expenses', data),
  update: (id: string, data: any) => apiClient.put(`/expenses/${id}`, data),
  delete: (id: string) => apiClient.delete(`/expenses/${id}`),
};

// Returns endpoints
export const returnsApi = {
  list: () => apiClient.get('/returns', {}),
  get: (id: string) => apiClient.get(`/returns/${id}`, {}),
  create: (data: any) => apiClient.post('/returns', data),
  update: (id: string, data: any) => apiClient.put(`/returns/${id}`, data),
};

// Warranties endpoints
export const warrantiesApi = {
  list: () => apiClient.get('/warranties', {}),
  get: (id: string) => apiClient.get(`/warranties/${id}`, {}),
  create: (data: any) => apiClient.post('/warranties', data),
  update: (id: string, data: any) => apiClient.put(`/warranties/${id}`, data),
};

// Reports endpoints
export const reportsApi = {
  sales: () => apiClient.get('/reports/sales', {}),
  profit: () => apiClient.get('/reports/profit', {}),
  inventory: () => apiClient.get('/reports/inventory', {}),
  debts: () => apiClient.get('/reports/debts', {}),
  products: () => apiClient.get('/reports/products', {}),
  suppliers: () => apiClient.get('/reports/suppliers', {}),
  expenses: () => apiClient.get('/reports/expenses', {}),
  returns: () => apiClient.get('/reports/returns', {}),
  warranty: () => apiClient.get('/reports/warranty', {}),
};

// Settings endpoints
export const settingsApi = {
  getOrganization: () => apiClient.get('/settings/organization', {}),
  updateOrganization: (data: any) => apiClient.put('/settings/organization', data),
  getUsers: () => apiClient.get('/settings/users', {}),
  createUser: (data: any) => apiClient.post('/settings/users', data),
  updateUser: (id: string, data: any) => apiClient.put(`/settings/users/${id}`, data),
  deleteUser: (id: string) => apiClient.delete(`/settings/users/${id}`),
};

// Barcode endpoints
export const barcodeApi = {
  scan: (barcode: string) => apiClient.post('/barcode/scan', { barcode }),
  lookup: (barcode: string) => apiClient.get(`/barcode/lookup/${barcode}`, {}),
};

// Inspections endpoints
export const inspectionsApi = {
  list: () => apiClient.get('/inspections', {}),
  get: (id: string) => apiClient.get(`/inspections/${id}`, {}),
  create: (data: any) => apiClient.post('/inspections', data),
  update: (id: string, data: any) => apiClient.put(`/inspections/${id}`, data),
  delete: (id: string) => apiClient.delete(`/inspections/${id}`),
};

// Notifications endpoints
export const notificationsApi = {
  list: () => apiClient.get('/notifications', {}),
  markAsRead: (id: string) => apiClient.put(`/notifications/${id}/read`, {}),
  markAllAsRead: () => apiClient.put('/notifications/read-all', {}),
  getUnreadCount: () => apiClient.get('/notifications/unread-count', {}),
  updatePreferences: (data: any) => apiClient.put('/notifications/preferences', data),
};

// Global search endpoint
export const searchApi = {
  global: (query: string) => apiClient.get(`/search?q=${encodeURIComponent(query)}`, {}),
};