import { lazy } from 'react';
import { Routes, Route } from 'react-router-dom';

// Lazy load components for better performance
const DashboardPage = lazy(() => import('../../features/dashboard/pages/DashboardPage').then(m => ({ default: m.DashboardPage })));
const InventoryPage = lazy(() => import('../../features/inventory/pages/InventoryPage').then(m => ({ default: m.InventoryPage })));
const POSPage = lazy(() => import('../../features/sales/pages/POSPage').then(m => ({ default: m.POSPage })));
const CustomersPage = lazy(() => import('../../features/customers/pages/CustomersPage').then(m => ({ default: m.CustomersPage })));
const DebtsPage = lazy(() => import('../../features/debts/pages/DebtsPage').then(m => ({ default: m.DebtsPage })));
const SuppliersPage = lazy(() => import('../../features/suppliers/pages/SuppliersPage').then(m => ({ default: m.SuppliersPage })));
const PurchasesPage = lazy(() => import('../../features/purchases/pages/PurchasesPage').then(m => ({ default: m.PurchasesPage })));
const CreatePurchasePage = lazy(() => import('../../features/purchases/pages/CreatePurchasePage').then(m => ({ default: m.CreatePurchasePage })));
const EditPurchasePage = lazy(() => import('../../features/purchases/pages/EditPurchasePage').then(m => ({ default: m.EditPurchasePage })));
const ExpensesPage = lazy(() => import('../../features/expenses/pages/ExpensesPage').then(m => ({ default: m.ExpensesPage })));
const ReturnsPage = lazy(() => import('../../features/returns/pages/ReturnsPage').then(m => ({ default: m.ReturnsPage })));
const ReportsPage = lazy(() => import('../../features/reports/pages/ReportsPage').then(m => ({ default: m.ReportsPage })));
const SettingsPage = lazy(() => import('../../features/settings/pages/SettingsPage').then(m => ({ default: m.SettingsPage })));

const PartTypesPage = lazy(() => import('../../features/parttypes/pages/PartTypesPage').then(m => ({ default: m.PartTypesPage })));
const UsedPartsPage = lazy(() => import('../../features/usedparts/pages/UsedPartsPage').then(m => ({ default: m.UsedPartsPage })));
const InspectionsPage = lazy(() => import('../../features/inspections/pages/InspectionsPage').then(m => ({ default: m.InspectionsPage })));
const ItemHistoryPage = lazy(() => import('../../features/item-history/pages/ItemHistoryPage').then(m => ({ default: m.ItemHistoryPage })));
const AgingPage = lazy(() => import('../../features/aging/pages/AgingPage').then(m => ({ default: m.AgingPage })));
const SellerBalancesPage = lazy(() => import('../../features/seller-balances/pages/SellerBalancesPage').then(m => ({ default: m.SellerBalancesPage })));
const ReturnDetailsPage = lazy(() => import('../../features/returns/pages/ReturnDetailsPage').then(m => ({ default: m.ReturnDetailsPage })));
const CategoriesPage = lazy(() => import('../../features/categories/pages/CategoriesPage').then(m => ({ default: m.CategoriesPage })));

// Loading component for lazy loaded routes
export function PageLoader() {
  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-cyan"></div>
    </div>
  );
}

// Centralized routes configuration
export const appRoutes = (
  <Routes>
    {/* Protected routes */}
    <Route path="/" element={<DashboardPage />} />
    <Route path="/dashboard" element={<DashboardPage />} />
    <Route path="/sales" element={<POSPage />} />
    <Route path="/inventory" element={<InventoryPage />} />
    <Route path="/usedparts" element={<UsedPartsPage />} />
    <Route path="/inspections" element={<InspectionsPage />} />
    <Route path="/item-history/:itemId" element={<ItemHistoryPage />} />
    <Route path="/aging" element={<AgingPage />} />
    <Route path="/seller-balances" element={<SellerBalancesPage />} />
    <Route path="/customers" element={<CustomersPage />} />
    <Route path="/debts" element={<DebtsPage />} />
    <Route path="/suppliers" element={<SuppliersPage />} />
    <Route path="/app/suppliers" element={<SuppliersPage />} />
    <Route path="/purchases" element={<PurchasesPage />} />
    <Route path="/app/purchases" element={<PurchasesPage />} />
    <Route path="/purchases/create" element={<CreatePurchasePage />} />
    <Route path="/app/purchases/create" element={<CreatePurchasePage />} />
    <Route path="/purchases/edit/:id" element={<EditPurchasePage />} />
    <Route path="/app/purchases/edit/:id" element={<EditPurchasePage />} />
    <Route path="/expenses" element={<ExpensesPage />} />
    <Route path="/returns" element={<ReturnsPage />} />
    <Route path="/app/returns" element={<ReturnsPage />} />
    <Route path="/app/returns/:id" element={<ReturnDetailsPage />} />
    <Route path="/reports" element={<ReportsPage />} />
    <Route path="/settings" element={<SettingsPage />} />
    <Route path="/parttypes" element={<PartTypesPage />} />
    <Route path="/app/categories" element={<CategoriesPage />} />
    <Route path="/categories" element={<CategoriesPage />} />
    
    {/* Catch all - redirect to dashboard */}
    <Route path="*" element={<DashboardPage />} />
  </Routes>
);
