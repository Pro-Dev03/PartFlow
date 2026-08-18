// API Endpoints
export const endpoints = {
  // Health
  health: '/health',
  ready: '/ready',
  alive: '/alive',

  // Auth
  auth: {
    register: '/api/v1/auth/register',
    login: '/api/v1/auth/login',
    refresh: '/api/v1/auth/refresh',
    logout: '/api/v1/auth/logout',
    changePassword: '/api/v1/auth/change-password',
    me: '/api/v1/users/me',
  },

  // Products
  products: {
    list: '/api/v1/products',
    create: '/api/v1/products',
    get: (id: string) => `/api/v1/products/${id}`,
    update: (id: string) => `/api/v1/products/${id}`,
    delete: (id: string) => `/api/v1/products/${id}`,
    archive: (id: string) => `/api/v1/products/${id}/archive`,
    addBarcode: (id: string) => `/api/v1/products/${id}/barcode`,
    getStock: (id: string) => `/api/v1/products/${id}/stock`,
    search: '/api/v1/products/search',
    getByBarcode: (barcode: string) => `/api/v1/products/barcode/${barcode}`,
  },

  // Categories
  categories: {
    list: '/api/v1/categories',
    create: '/api/v1/categories',
    get: (id: string) => `/api/v1/categories/${id}`,
    update: (id: string) => `/api/v1/categories/${id}`,
    delete: (id: string) => `/api/v1/categories/${id}`,
  },

  // Brands
  brands: {
    list: '/api/v1/brands',
    create: '/api/v1/brands',
    get: (id: string) => `/api/v1/brands/${id}`,
    update: (id: string) => `/api/v1/brands/${id}`,
    delete: (id: string) => `/api/v1/brands/${id}`,
  },

  // Inventory
  inventory: {
    items: {
      list: '/api/v1/inventory/items',
      create: '/api/v1/inventory/items',
      get: (id: string) => `/api/v1/inventory/items/${id}`,
      updateStatus: (id: string) => `/api/v1/inventory/items/${id}/status`,
      receive: (id: string) => `/api/v1/inventory/items/${id}/receive`,
      adjust: (id: string) => `/api/v1/inventory/items/${id}/adjust`,
      transfer: (id: string) => `/api/v1/inventory/items/${id}/transfer`,
      history: (id: string) => `/api/v1/inventory/items/${id}/history`,
    },
    barcode: (code: string) => `/api/v1/inventory/barcode/${code}`,
    locations: {
      list: '/api/v1/locations',
      create: '/api/v1/locations',
      get: (id: string) => `/api/v1/locations/${id}`,
    },
    reservations: {
      create: '/api/v1/reservations',
      release: (id: string) => `/api/v1/reservations/${id}/release`,
    },
  },

  // Customers
  customers: {
    list: '/api/v1/customers',
    create: '/api/v1/customers',
    get: (id: string) => `/api/v1/customers/${id}`,
    update: (id: string) => `/api/v1/customers/${id}`,
    delete: (id: string) => `/api/v1/customers/${id}`,
    ledger: (id: string) => `/api/v1/customers/${id}/ledger`,
    payments: (id: string) => `/api/v1/customers/${id}/payments`,
    debtSummary: (id: string) => `/api/v1/customers/${id}/debt-summary`,
    creditLimit: (id: string) => `/api/v1/customers/${id}/credit-limit`,
    overdue: '/api/v1/customers/overdue',
  },

  // Sales
  sales: {
    list: '/api/v1/sales',
    create: '/api/v1/sales',
    get: (id: string) => `/api/v1/sales/${id}`,
    payment: (id: string) => `/api/v1/sales/${id}/payment`,
    cancel: (id: string) => `/api/v1/sales/${id}/cancel`,
    summary: '/api/v1/sales/summary',
    topProducts: '/api/v1/sales/top-products',
  },

  // Payments
  payments: {
    list: '/api/v1/payments',
    create: '/api/v1/payments',
    get: (id: string) => `/api/v1/payments/${id}`,
    update: (id: string) => `/api/v1/payments/${id}`,
    delete: (id: string) => `/api/v1/payments/${id}`,
    complete: (id: string) => `/api/v1/payments/${id}/complete`,
    cancel: (id: string) => `/api/v1/payments/${id}/cancel`,
    summary: '/api/v1/payments/summary',
  },

  // Suppliers
  suppliers: {
    list: '/api/v1/suppliers',
    create: '/api/v1/suppliers',
    get: (id: string) => `/api/v1/suppliers/${id}`,
    update: (id: string) => `/api/v1/suppliers/${id}`,
    delete: (id: string) => `/api/v1/suppliers/${id}`,
    ledger: (id: string) => `/api/v1/suppliers/${id}/ledger`,
    payments: (id: string) => `/api/v1/suppliers/${id}/payments`,
    debtSummary: (id: string) => `/api/v1/suppliers/${id}/debt-summary`,
    creditLimit: (id: string) => `/api/v1/suppliers/${id}/credit-limit`,
    overdue: '/api/v1/suppliers/overdue',
  },

  // Purchases
  purchases: {
    list: '/api/v1/purchases',
    create: '/api/v1/purchases',
    get: (id: string) => `/api/v1/purchases/${id}`,
    update: (id: string) => `/api/v1/purchases/${id}`,
    delete: (id: string) => `/api/v1/purchases/${id}`,
    receive: (id: string) => `/api/v1/purchases/${id}/receive`,
    cancel: (id: string) => `/api/v1/purchases/${id}/cancel`,
    payment: (id: string) => `/api/v1/purchases/${id}/payment`,
    items: {
      create: (purchaseId: string) => `/api/v1/purchases/${purchaseId}/items`,
      update: (_purchaseId: string, itemId: string) => `/api/v1/purchases/items/${itemId}`,
      delete: (_purchaseId: string, itemId: string) => `/api/v1/purchases/items/${itemId}`,
    },
  },

  // Expenses
  expenses: {
    list: '/api/v1/expenses',
    create: '/api/v1/expenses',
    get: (id: string) => `/api/v1/expenses/${id}`,
    update: (id: string) => `/api/v1/expenses/${id}`,
    delete: (id: string) => `/api/v1/expenses/${id}`,
    approve: (id: string) => `/api/v1/expenses/${id}/approve`,
    reject: (id: string) => `/api/v1/expenses/${id}/reject`,
    summary: '/api/v1/expenses/summary',
    categories: '/api/v1/expenses/categories',
  },

  // Returns
  returns: {
    list: '/api/v1/returns',
    create: '/api/v1/returns',
    get: (id: string) => `/api/v1/returns/${id}`,
    update: (id: string) => `/api/v1/returns/${id}`,
    delete: (id: string) => `/api/v1/returns/${id}`,
    approve: (id: string) => `/api/v1/returns/${id}/approve`,
    reject: (id: string) => `/api/v1/returns/${id}/reject`,
    refund: (id: string) => `/api/v1/returns/${id}/refund`,
    items: {
      create: (returnId: string) => `/api/v1/returns/${returnId}/items`,
      update: (_returnId: string, itemId: string) => `/api/v1/returns/items/${itemId}`,
      delete: (_returnId: string, itemId: string) => `/api/v1/returns/items/${itemId}`,
    },
  },

  // Warranties
  warranties: {
    list: '/api/v1/warranties',
    create: '/api/v1/warranties',
    get: (id: string) => `/api/v1/warranties/${id}`,
    update: (id: string) => `/api/v1/warranties/${id}`,
    delete: (id: string) => `/api/v1/warranties/${id}`,
    expiringSoon: '/api/v1/warranties/expiring-soon',
    claims: {
      create: (id: string) => `/api/v1/warranties/${id}/claims`,
      get: (id: string) => `/api/v1/warranties/claims/${id}`,
      list: '/api/v1/warranties/claims',
      update: (id: string) => `/api/v1/warranties/claims/${id}`,
      delete: (id: string) => `/api/v1/warranties/claims/${id}`,
      approve: (id: string) => `/api/v1/warranties/claims/${id}/approve`,
      reject: (id: string) => `/api/v1/warranties/claims/${id}/reject`,
      complete: (id: string) => `/api/v1/warranties/claims/${id}/complete`,
    },
  },

  // Inspections
  inspections: {
    list: '/api/v1/inspections',
    create: '/api/v1/inspections',
    get: (id: string) => `/api/v1/inspections/${id}`,
    update: (id: string) => `/api/v1/inspections/${id}`,
    delete: (id: string) => `/api/v1/inspections/${id}`,
  },

  // Reports
  reports: {
    list: '/api/v1/reports',
    create: '/api/v1/reports',
    get: (id: string) => `/api/v1/reports/${id}`,
    delete: (id: string) => `/api/v1/reports/${id}`,
    sales: '/api/v1/reports/sales',
    profit: '/api/v1/reports/profit',
    inventory: '/api/v1/reports/inventory',
    debts: '/api/v1/reports/debts',
    purchases: '/api/v1/reports/purchases',
    expenses: '/api/v1/reports/expenses',
  },

  // Notifications
  notifications: {
    list: '/api/v1/notifications',
    create: '/api/v1/notifications',
    get: (id: string) => `/api/v1/notifications/${id}`,
    markAsRead: (id: string) => `/api/v1/notifications/${id}/read`,
    markAllAsRead: '/api/v1/notifications/read-all',
    delete: (id: string) => `/api/v1/notifications/${id}`,
    unreadCount: '/api/v1/notifications/unread-count',
    preferences: {
      get: '/api/v1/notifications/preferences',
      update: '/api/v1/notifications/preferences',
    },
  },

  // Dashboard
  dashboard: {
    stats: '/api/v1/dashboard/stats',
  },

  // Audit
  audit: {
    create: '/api/v1/audit',
    get: (id: string) => `/api/v1/audit/${id}`,
    list: '/api/v1/audit',
    summary: '/api/v1/audit/summary',
    stats: '/api/v1/audit/stats',
    export: '/api/v1/audit/export',
    user: (userId: string) => `/api/v1/audit/user/${userId}`,
    entity: (entityId: string) => `/api/v1/audit/entity/${entityId}`,
  },

  // Barcodes
  barcodes: {
    get: (code: string) => `/api/v1/barcodes/${code}`,
    generate: '/api/v1/barcodes/generate',
  },

  // Search
  search: '/api/v1/search',
} as const;