import { Routes, Route, Navigate } from 'react-router-dom'
import { lazy, Suspense } from 'react'
import { DesktopLayout } from '../../layouts/DesktopLayout'
import { MobileLayout } from '../../layouts/MobileLayout'
import { Auth } from '../../features/auth'

// Lazy load route components for better performance
const Dashboard = lazy(() => import('../../features/dashboard').then(m => ({ default: m.Dashboard })))
const Sales = lazy(() => import('../../features/sales').then(m => ({ default: m.Sales })))
const Inventory = lazy(() => import('../../features/inventory').then(m => ({ default: m.Inventory })))
const Customers = lazy(() => import('../../features/customers').then(m => ({ default: m.Customers })))
const Debts = lazy(() => import('../../features/debts').then(m => ({ default: m.Debts })))

// Additional routes for complete feature coverage
const Expenses = lazy(() => import('../../features/expenses').then(m => ({ default: m.Expenses })))
const Reports = lazy(() => import('../../features/reports').then(m => ({ default: m.Reports })))
const Purchases = lazy(() => import('../../features/purchases').then(m => ({ default: m.Purchases })))
const Suppliers = lazy(() => import('../../features/suppliers').then(m => ({ default: m.Suppliers })))
const Settings = lazy(() => import('../../features/settings').then(m => ({ default: m.Settings })))
const Barcode = lazy(() => import('../../features/barcode').then(m => ({ default: m.Barcode })))
const ImportExport = lazy(() => import('../../features/import-export').then(m => ({ default: m.ImportExport })))

// Loading component for lazy routes
function LoadingFallback() {
  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
        <p className="mt-4 text-muted">جاري التحميل...</p>
      </div>
    </div>
  )
}

export function AppRouter() {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <Routes>
        {/* Auth Route */}
        <Route path="/auth" element={<Auth />} />
        
        {/* Desktop Routes */}
        <Route path="/" element={<DesktopLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="sales" element={<Sales />} />
          <Route path="inventory" element={<Inventory />} />
          <Route path="customers" element={<Customers />} />
          <Route path="debts" element={<Debts />} />
          <Route path="expenses" element={<Expenses />} />
          <Route path="reports" element={<Reports />} />
          <Route path="purchases" element={<Purchases />} />
          <Route path="suppliers" element={<Suppliers />} />
          <Route path="settings" element={<Settings />} />
          <Route path="barcode" element={<Barcode />} />
          <Route path="import-export" element={<ImportExport />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
        
        {/* Mobile Routes */}
        <Route path="/mobile" element={<MobileLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="sales" element={<Sales />} />
          <Route path="scan" element={<Barcode />} />
          <Route path="inventory" element={<Inventory />} />
          <Route path="more" element={<Settings />} />
          <Route path="barcode" element={<Barcode />} />
          <Route path="import-export" element={<ImportExport />} />
          <Route path="*" element={<Navigate to="/mobile/dashboard" replace />} />
        </Route>
      </Routes>
    </Suspense>
  )
}
