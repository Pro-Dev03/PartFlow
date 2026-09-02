import { apiClient } from './client';
import { TokenManager } from '../../lib/token-manager';
import { getCloudApiUrl, getLocalApiUrl, shouldUseLocalApi } from '../../lib/config/app';
import type {
  ProductCreateRequest,
  ProductUpdateRequest,
  ProductListParams,
  CategoryCreateRequest,
  CategoryUpdateRequest,
  CustomerCreateRequest,
  CustomerUpdateRequest,
  CustomerListParams,
  SupplierCreateRequest,
  SupplierUpdateRequest,
  InventoryCreateRequest,
  InventoryUpdateRequest,
  InventoryListParams,
  SaleCreateRequest,
  DebtPayment,
  PaymentCreateRequest,
  PurchaseCreateRequest,
  ExpenseCreateRequest,
  PartTypeCreateRequest,
  PartTypeUpdateRequest,
  BarcodeLookupResponse,
} from './types';

// Auth endpoints
export const authApi = {
  login: async (email: string, password: string) => {
    const response = await apiClient.post('/auth/login', { email, password });
    return response.data ?? response;
  },
  loginWithCloud: async (email: string, password: string) => {
    if (typeof navigator !== 'undefined' && !navigator.onLine) {
      throw new Error('يلزم اتصال بالإنترنت لتسجيل الدخول.');
    }

    const response = await fetch(`${getCloudApiUrl()}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    const payload = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(payload?.error?.message || payload?.error || 'تعذر تسجيل الدخول.');
    }
    return payload?.data ?? payload;
  },
  createLocalSession: async (cloudToken: string) => {
    const response = await fetch(`${getLocalApiUrl()}/auth/cloud-session`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ cloud_token: cloudToken }),
    });
    const payload = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(payload?.error?.message || payload?.error || 'تعذر إنشاء الجلسة المحلية.');
    }
    return payload?.data ?? payload;
  },
  logout: () => apiClient.post('/auth/logout', {}),
  logoutWithCloud: (accessToken: string) => {
    if (!accessToken || typeof fetch !== 'function') {
      return Promise.resolve();
    }

    // The desktop API is local-first, so its logout handler cannot revoke a
    // cloud refresh token. Best-effort call Render directly before local
    // state is cleared; failures are intentionally ignored by the caller so
    // logout still works while offline.
    return fetch(`${getCloudApiUrl()}/auth/logout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({}),
    }).then(async (response) => {
      if (!response.ok) {
        const payload = await response.json().catch(() => ({}));
        const error: any = new Error(payload?.error?.message || payload?.error || 'Cloud logout failed');
        error.status = response.status;
        throw error;
      }
    });
  },
  checkAdminAccess: async () => {
    const cloudToken = typeof window !== 'undefined' ? localStorage.getItem('cloud_token') : null;
    if (!cloudToken) {
      throw new Error('No active cloud session');
    }

    const response = await fetch(`${getCloudApiUrl()}/auth/admin-check`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${cloudToken}`,
        'X-PartFlow-Cloud-Token': cloudToken,
      },
    });

    const payload = await (async () => {
      if (typeof response.json === 'function') {
        try {
          const data = await response.json();
          if (data && typeof data === 'object') return data;
        } catch {
          // Fall through to text parsing below.
        }
      }
      if (typeof response.text === 'function') {
        const text = await response.text();
        if (!text) return {};
        try {
          return JSON.parse(text);
        } catch {
          return {};
        }
      }
      return {};
    })();

    if (!response.ok) {
      const error: any = new Error(payload?.error?.message || payload?.error || 'Access denied');
      error.status = response.status;
      error.response = payload;
      throw error;
    }

    return payload?.data ?? payload;
  },
  refreshToken: async () => {
    const refreshToken = TokenManager.getRefreshToken();
    if (!refreshToken) {
      return Promise.reject(new Error('No refresh token available'));
    }

    const baseUrl = typeof window !== 'undefined' && shouldUseLocalApi(window.location.hostname)
      ? getLocalApiUrl()
      : getCloudApiUrl();

    const response = await fetch(`${baseUrl}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    const payload = await (async () => {
      if (typeof response.json === 'function') {
        try {
          const data = await response.json();
          if (data && typeof data === 'object') return data;
        } catch {
          // Fall through to text parsing below.
        }
      }
      if (typeof response.text === 'function') {
        const text = await response.text();
        if (!text) return {};
        try {
          return JSON.parse(text);
        } catch {
          return {};
        }
      }
      return {};
    })();

    if (!response.ok) {
      const error: any = new Error(payload?.error?.message || payload?.error || 'تعذر تحديث الجلسة.');
      error.status = response.status;
      error.response = payload;
      throw error;
    }

    return payload?.data ?? payload;
  },
  forgotPassword: (email: string) =>
    apiClient.post('/auth/forgot-password', { email }),
  resetPassword: (token: string, password: string) =>
    apiClient.post('/auth/reset-password', { token, password }),
};

// Dashboard endpoints
export const dashboardApi = {
  getStats: () => apiClient.get('/dashboard/stats', undefined, false),
  getLowStockItems: () => apiClient.get('/dashboard/low-stock-items', undefined, false),
  getOverdueDebts: () => apiClient.get('/dashboard/overdue-debts', undefined, false),
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
  list: (params?: ProductListParams) => 
    apiClient.get('/products', params),
  get: (id: string) => apiClient.get(`/products/${id}`),
  create: (data: ProductCreateRequest) => apiClient.post('/products', data),
  update: (id: string, data: ProductUpdateRequest) => apiClient.put(`/products/${id}`, data),
  delete: (id: string) => apiClient.delete(`/products/${id}`),
  archive: (id: string) => apiClient.post(`/products/${id}/archive`),
};

// Categories endpoints
export const categoriesApi = {
  list: () => apiClient.get('/categories'),
  get: (id: string) => apiClient.get(`/categories/${id}`),
  create: (data: CategoryCreateRequest) => apiClient.post('/categories', data),
  update: (id: string, data: CategoryUpdateRequest) => apiClient.put(`/categories/${id}`, data),
  delete: (id: string) => apiClient.delete(`/categories/${id}`),
};

// Inventory endpoints
export const inventoryApi = {
  list: (params?: InventoryListParams) =>
    apiClient.get('/inventory/items', params),
  listWithSupplier: (params?: InventoryListParams & { 
    exclude_condition?: string;
    supplier_id?: string;
    purchase_date_from?: string;
    purchase_date_to?: string;
    min_purchase_cost?: number;
    max_purchase_cost?: number;
  }) =>
    apiClient.get('/inventory/items-with-supplier', params),
  get: (id: string) => apiClient.get(`/inventory/items/${id}`),
  create: (data: InventoryCreateRequest) => apiClient.post('/inventory/items', data),
  update: (id: string, data: InventoryUpdateRequest) => apiClient.put(`/inventory/items/${id}`, data),
  updateStatus: (id: string, status: string) => apiClient.patch(`/inventory/items/${id}/status`, { status }),
  delete: (id: string) => apiClient.delete(`/inventory/items/${id}`),
  movements: <T = unknown>(itemId: string) => apiClient.get<T>(`/inventory/items/${itemId}/history`),
  createTradeIn: (data: InventoryCreateRequest) => apiClient.post('/inventory/trade-ins', data),
};

// Part Types endpoints
export const partTypesApi = {
  list: () => apiClient.get('/part-types'),
  get: (id: string) => apiClient.get(`/part-types/${id}`),
  create: (data: PartTypeCreateRequest) => apiClient.post('/part-types', data),
  update: (id: string, data: PartTypeUpdateRequest) => apiClient.put(`/part-types/${id}`, data),
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
  list: (params?: PaginationParams & { search?: string; available_for_return?: boolean }) =>
    apiClient.get('/sales', params),
  get: (id: string) => apiClient.get(`/sales/${id}`),
  create: (data: SaleCreateRequest) => apiClient.post('/sales', data),
  listHeld: () => apiClient.get('/sales/held'),
  hold: (items: unknown[]) => apiClient.post('/sales/held', { items }),
  deleteHeld: (id: string) => apiClient.delete(`/sales/held/${id}`),
  update: (id: string, data: Partial<SaleCreateRequest>) => apiClient.put(`/sales/${id}`, data),
  // SmartDelete - now returns SmartDeleteResult (ARCHITECTURE-PRINCIPLES.md)
  delete: (id: string) => apiClient.delete(`/sales/${id}`),
  refund: (id: string, data: { reason?: string; refund_amount?: number }) => apiClient.post(`/sales/${id}/refund`, data),
};

// Payments endpoints (ARCHITECTURE-PRINCIPLES.md)
export const paymentsApi = {
  list: (params?: PaginationParams & { type?: string }) =>
    apiClient.get('/payments', params),
  get: (id: string) => apiClient.get(`/payments/${id}`),
  create: (data: PaymentCreateRequest) => apiClient.post('/payments', data),
  // SmartDelete - now returns SmartDeleteResult (ARCHITECTURE-PRINCIPLES.md)
  delete: (id: string) => apiClient.delete(`/payments/${id}`),
};

// Customers endpoints
export const customersApi = {
  list: (params?: CustomerListParams) => 
    apiClient.get('/customers', params),
  get: (id: string) => apiClient.get(`/customers/${id}`),
  create: (data: CustomerCreateRequest) => apiClient.post('/customers', data),
  update: (id: string, data: CustomerUpdateRequest) => apiClient.put(`/customers/${id}`, data),
  delete: (id: string) => apiClient.delete(`/customers/${id}`),
  ledger: (id: string, params?: PaginationParams) => 
    apiClient.get(`/customers/${id}/ledger`, params),
  ledgerSummary: (id: string) => 
    apiClient.get(`/customers/${id}/ledger/summary`),
  getFinancialTimeline: (id: string) => 
    apiClient.get(`/customers/${id}/financial-timeline`),
};

// Debts endpoints - Note: Debts are managed under customers in the backend
export const debtsApi = {
  list: (params?: PaginationParams) => 
    apiClient.get('/debts', params, false),
  get: (customerId: string, debtId: string) => apiClient.get(`/customers/${customerId}/debts/${debtId}`),
  recordPayment: (customerId: string, data: DebtPayment) => apiClient.post(`/customers/${customerId}/debt-payments`, data),
  getDebtEntries: (customerId: string) => apiClient.get(`/customers/${customerId}/debts`),
  getDebtCollections: (customerId: string) => apiClient.get(`/customers/${customerId}/debt-collections`),
  getPendingCollections: () => apiClient.get('/debt-collections/pending'),
};

// Suppliers endpoints
export const suppliersApi = {
  list: (params?: PaginationParams & { search?: string; is_active?: boolean }) =>
    apiClient.get('/suppliers', params),
  get: (id: string) => apiClient.get(`/suppliers/${id}`),
  create: (data: SupplierCreateRequest) => apiClient.post('/suppliers', data),
  update: (id: string, data: SupplierUpdateRequest) => apiClient.put(`/suppliers/${id}`, data),
  delete: (id: string) => apiClient.delete(`/suppliers/${id}`),
  ledger: (id: string, params?: PaginationParams) => 
    apiClient.get(`/suppliers/${id}/ledger`, params),
  ledgerSummary: (id: string) => 
    apiClient.get(`/suppliers/${id}/ledger/summary`),
  getSupplierInventory: (id: string) => apiClient.get(`/suppliers/${id}/inventory`),
};

// Purchases endpoints
export const purchasesApi = {
  list: (params?: PaginationParams & { search?: string }) =>
    apiClient.get('/purchases', params),
  get: (id: string) => apiClient.get(`/purchases/${id}`),
  create: (data: PurchaseCreateRequest) => apiClient.post('/purchases', data),
  update: (id: string, data: Partial<PurchaseCreateRequest>) => apiClient.put(`/purchases/${id}`, data),
  // SmartDelete - now returns SmartDeleteResult (ARCHITECTURE-PRINCIPLES.md)
  delete: (id: string) => apiClient.delete(`/purchases/${id}`),
  receive: (id: string) => apiClient.post(`/purchases/${id}/receive`, {}),
  addPayment: (id: string, data: { amount: number; paymentMethod: string }) =>
    apiClient.post(`/purchases/${id}/payment`, data),
  cancel: (id: string) => apiClient.post(`/purchases/${id}/cancel`, {}),
  reverse: (id: string, reason: string) => apiClient.post(`/purchases/${id}/reverse`, { reason }),
  // Get used items info for blocked deletion
  getUsedItemsInfo: (id: string) => apiClient.get(`/purchases/${id}/used-items`),
};

// Expenses endpoints
export const expensesApi = {
  list: (params?: PaginationParams & { search?: string }) => 
    apiClient.get('/expenses', params),
  get: (id: string) => apiClient.get(`/expenses/${id}`),
  create: (data: ExpenseCreateRequest) => apiClient.post('/expenses', data),
  update: (id: string, data: Partial<ExpenseCreateRequest>) => apiClient.put(`/expenses/${id}`, data),
  delete: (id: string) => apiClient.delete(`/expenses/${id}`),
};

// Reports endpoints
export const reportsApi = {
  sales: (params?: { start_date?: string; end_date?: string }) => 
    apiClient.get('/reports/sales', params, false),
  netSales: (params?: { start_date?: string; end_date?: string }) => 
    apiClient.get('/reports/net-sales', params, false),
  profit: (params?: { start_date?: string; end_date?: string }) => 
    apiClient.get('/reports/profit', params, false),
  inventory: () => 
    apiClient.get('/reports/inventory', undefined, false),
  debts: () => 
    apiClient.get('/reports/debts', undefined, false),
  products: () => 
    apiClient.get('/reports/products', undefined, false),
  suppliers: () => 
    apiClient.get('/reports/suppliers', undefined, false),
  purchases: (params?: { start_date?: string; end_date?: string }) =>
    apiClient.get('/reports/purchases', params, false),
  expenses: (params?: { start_date?: string; end_date?: string }) => 
    apiClient.get('/reports/expenses', params, false),
  returns: (params?: { start_date?: string; end_date?: string }) => 
    apiClient.get('/reports/returns', params, false),
  returnsAnalysis: () =>
    apiClient.get('/returns/analysis/monthly', undefined, false),
};

// Settings endpoints
export const settingsApi = {
  syncCloudData: () =>
    apiClient.post('/settings/sync', {}),
  syncLocalDataToCloud: () =>
    apiClient.post('/settings/sync/push', {}),
  getUsers: (params?: { page?: number; per_page?: number }) =>
    apiClient.get('/users', params),
  getSubscribers: (params?: { page?: number; per_page?: number; search?: string; is_active?: boolean }) =>
    apiClient.get('/users/subscriptions', params),
  getSubscriptionSummary: () =>
    apiClient.get('/users/subscription-summary'),
  updateSubscriptionStatus: (id: string, payload: { subscription_status: string; subscription_expires_at?: string | null }) =>
    apiClient.put(`/users/${id}/subscription`, payload),
  renewSubscription: (id: string, days: number) =>
    apiClient.post(`/users/${id}/subscription/renew`, { days }),
  createUser: (data: any) => apiClient.post('/users', data),
  updateUser: (id: string, data: any) => apiClient.put(`/users/${id}`, data),
  deleteUser: (id: string) => apiClient.delete(`/users/${id}`),
  getTaxRate: () => apiClient.get('/settings/tax-rate'),
  updateTaxRate: (taxRate: number) => apiClient.put('/settings/tax-rate', { tax_rate: taxRate }),
  getPublicSettings: () => apiClient.get('/settings/public'),
  getSetting: (key: string) => apiClient.get(`/settings/${key}`),
  updateSetting: (key: string, value: string) => apiClient.put(`/settings/${key}`, { value }),
  // The desktop application owns the operational SQLite database. Never expose
  // a client-side option that can target the cloud database for deletion.
  deleteAllData: (confirmation: string) =>
    apiClient.delete('/settings/database', { confirmation_token: confirmation, target: 'offline' }),
  deleteCloudData: async (confirmation: string) => {
    const cloudToken = typeof window !== 'undefined' ? localStorage.getItem('cloud_token') : null;
    if (!cloudToken) throw new Error('No active cloud session');

    const response = await fetch(
      `${getCloudApiUrl()}/settings/database?confirmation_token=${encodeURIComponent(confirmation)}`,
      {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${cloudToken}`,
          'X-PartFlow-Cloud-Token': cloudToken,
        },
      },
    );
    const payload = await response.json().catch(() => ({}));
    if (!response.ok) {
      const error: any = new Error(payload?.error?.message || payload?.error || 'Cloud data deletion failed');
      error.status = response.status;
      error.response = payload;
      throw error;
    }
    return payload?.data ?? payload;
  },
};

// Sync endpoints
export const syncApi = {
  getInitialData: () => apiClient.get('/sync/initial-data'),
};

// Barcode endpoints
export const barcodeApi = {
  // Barcode lookup is read-only; use the backend's canonical product lookup
  // route instead of the removed /barcode/scan endpoint.
  scan: (barcode: string) => apiClient.get(`/barcodes/product/${encodeURIComponent(barcode)}`),
  lookup: (barcode: string) => apiClient.get(`/barcodes/${barcode}`),
  lookupProduct: (barcode: string) => apiClient.get<BarcodeLookupResponse>(`/barcodes/product/${barcode}`),
  lookupBySKU: (sku: string) => apiClient.get(`/barcodes/sku/${sku}`),
};

// Inspections endpoints
export const inspectionsApi = {
  list: (params?: { page?: number; per_page?: number; search?: string; status?: string }) => 
    apiClient.get('/inspections', params),
  get: (id: string) => apiClient.get(`/inspections/${id}`),
  create: (data: any) => apiClient.post('/inspections', data),
  update: (id: string, data: any) => apiClient.put(`/inspections/${id}`, data),
  pass: (id: string) => apiClient.post(`/inspections/${id}/pass`, {}),
  fail: (id: string) => apiClient.post(`/inspections/${id}/fail`, {}),
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
  updateItem: (returnId: string, itemId: string, data: any) => apiClient.put(`/returns/${returnId}/items/${itemId}`, data),
  deleteItem: (returnId: string, itemId: string) => apiClient.delete(`/returns/${returnId}/items/${itemId}`),
  processInspection: (itemId: string, data: any) => apiClient.post(`/returns/items/${itemId}/inspection`, data),
  validateQuantity: (saleItemId: string, quantity: number) => 
    apiClient.get(`/returns/validate/${saleItemId}`, { quantity }),
  getSummary: () => apiClient.get('/returns/summary'),
  reverse: (id: string) => apiClient.post(`/returns/${id}/reverse`),
};

export const supplierReturnsApi = {
  list: (status?: string) => apiClient.get('/supplier-returns', status ? { status } : undefined),
  create: (data: { purchase_id: string; reason: string; notes?: string }) =>
    apiClient.post('/supplier-returns', data),
  addItem: (id: string, data: { purchase_item_id: string; quantity: number }) =>
    apiClient.post(`/supplier-returns/${id}/items`, data),
  delete: (id: string) => apiClient.delete(`/supplier-returns/${id}`),
  reject: (id: string) => apiClient.post(`/supplier-returns/${id}/reject`, {}),
  complete: (id: string) => apiClient.post(`/supplier-returns/${id}/complete`),
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
    apiClient.get(`/acquisitions/items/${itemId}/history`),
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
