import { Routes, Route } from 'react-router-dom'
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
        <Route path="/auth" element={<Auth />} />
        <Route path="/" element={<DesktopLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="sales" element={<Sales />} />
          <Route path="inventory" element={<Inventory />} />
          <Route path="customers" element={<Customers />} />
          <Route path="debts" element={<Debts />} />
        </Route>
      </Routes>
    </Suspense>
  )
}
