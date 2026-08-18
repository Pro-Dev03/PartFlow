import { Routes, Route } from 'react-router-dom'
import { DesktopLayout } from '../../layouts/DesktopLayout'
import { MobileLayout } from '../../layouts/MobileLayout'
import { Dashboard } from '../../features/dashboard'
import { Sales } from '../../features/sales'
import { Inventory } from '../../features/inventory'
import { Customers } from '../../features/customers'
import { Debts } from '../../features/debts'
import { Auth } from '../../features/auth'

export function AppRouter() {
  return (
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
  )
}
