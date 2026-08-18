import { lazy } from 'react'

// Lazy load route components for better performance
// This is a template - adjust paths based on actual feature structure
export const LazyRoutes = {
  // Dashboard
  Dashboard: lazy(() => import('../../features/dashboard')),

  // Products & Inventory
  Inventory: lazy(() => import('../../features/inventory/pages/InventoryPage')),
  ProductDetail: lazy(() => import('../../features/inventory/pages/ProductPage')),
  UsedItems: lazy(() => import('../../features/inventory/pages/UsedItemsPage')),

  // Sales
  Sales: lazy(() => import('../../features/sales')),

  // Customers
  Customers: lazy(() => import('../../features/customers/pages/CustomerPage')),

  // Debts
  Debts: lazy(() => import('../../features/debts/pages/DebtsPage')),

  // Expenses
  Expenses: lazy(() => import('../../features/expenses/pages/ExpensesPage')),

  // Reports
  Reports: lazy(() => import('../../features/reports/pages/ReportsPage')),

  // Import/Export
  ImportExport: lazy(() => import('../../features/import-export/pages/ImportExportPage')),

  // Audit
  Audit: lazy(() => import('../../features/audit/pages/AuditPage')),

  // Barcode
  Barcode: lazy(() => import('../../features/barcode')),
}

// Preload critical routes
export function preloadCriticalRoutes() {
  // Preload dashboard immediately
  LazyRoutes.Dashboard()

  // Preload other routes after a delay
  setTimeout(() => {
    LazyRoutes.Inventory()
    LazyRoutes.Sales()
    LazyRoutes.Customers()
  }, 2000)
}

// Preload route on hover
export function preloadRoute(routeName: keyof typeof LazyRoutes) {
  LazyRoutes[routeName]()
}