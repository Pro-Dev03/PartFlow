import { lazy, Suspense, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryProvider } from './app/providers/QueryProvider';
import { AppLayout, AuthLayout } from './layouts';
import { useAuthStore } from './stores/authStore';
import { ErrorBoundary } from './components/ui/error-boundary';
import { ToastContainer } from './components/ui/ToastContainer';

// Lazy load components for better performance
const LoginPage = lazy(() => import('./features/auth/pages/LoginPage').then(m => ({ default: m.LoginPage })));
const DashboardPage = lazy(() => import('./features/dashboard/pages/DashboardPage').then(m => ({ default: m.DashboardPage })));
const InventoryPage = lazy(() => import('./features/inventory/pages/InventoryPage').then(m => ({ default: m.InventoryPage })));
const POSPage = lazy(() => import('./features/sales/pages/POSPage').then(m => ({ default: m.POSPage })));
const CustomersPage = lazy(() => import('./features/customers/pages/CustomersPage').then(m => ({ default: m.CustomersPage })));
const DebtsPage = lazy(() => import('./features/debts/pages/DebtsPage').then(m => ({ default: m.DebtsPage })));
const SuppliersPage = lazy(() => import('./features/suppliers/pages/SuppliersPage').then(m => ({ default: m.SuppliersPage })));
const PurchasesPage = lazy(() => import('./features/purchases/pages/PurchasesPage').then(m => ({ default: m.PurchasesPage })));
const ExpensesPage = lazy(() => import('./features/expenses/pages/ExpensesPage').then(m => ({ default: m.ExpensesPage })));
const ReturnsPage = lazy(() => import('./features/returns/pages/ReturnsPage').then(m => ({ default: m.ReturnsPage })));
const WarrantiesPage = lazy(() => import('./features/warranties/pages/WarrantiesPage').then(m => ({ default: m.WarrantiesPage })));
const ReportsPage = lazy(() => import('./features/reports/pages/ReportsPage').then(m => ({ default: m.ReportsPage })));
const SettingsPage = lazy(() => import('./features/settings/pages/SettingsPage').then(m => ({ default: m.SettingsPage })));
const DesignSystemShowcase = lazy(() => import('./features/design-system/DesignSystemShowcase').then(m => ({ default: m.DesignSystemShowcase })));



// Loading component for lazy loaded routes
function PageLoader() {
  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
    </div>
  );
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore();
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore();
  
  if (isAuthenticated) {
    return <Navigate to="/app" replace />;
  }
  
  return <>{children}</>;
}

// Component to preload critical pages
function PagePreloader() {
  useEffect(() => {
    // Preload POS page (most used)
    import('./features/sales/pages/POSPage');
    // Preload Inventory page (second most used)
    import('./features/inventory/pages/InventoryPage');
  }, []);

  return null;
}

function App() {
  return (
    <ErrorBoundary>
      <QueryProvider>
        <Router>
          <PagePreloader />
          <Routes>
            {/* Public routes */}
            <Route
              path="/login"
              element={
                <PublicRoute>
                  <AuthLayout>
                    <Suspense fallback={<PageLoader />}>
                      <LoginPage />
                    </Suspense>
                  </AuthLayout>
                </PublicRoute>
              }
            />

            {/* Protected routes */}
            <Route
              path="/app/*"
              element={
                <ProtectedRoute>
                  <AppLayout>
                    <Routes>
                      <Route path="/" element={<Suspense fallback={<PageLoader />}><DashboardPage /></Suspense>} />
                      <Route path="/dashboard" element={<Suspense fallback={<PageLoader />}><DashboardPage /></Suspense>} />
                      <Route path="/inventory" element={<Suspense fallback={<PageLoader />}><InventoryPage /></Suspense>} />
                      <Route path="/sales" element={<Suspense fallback={<PageLoader />}><POSPage /></Suspense>} />
                      <Route path="/customers" element={<Suspense fallback={<PageLoader />}><CustomersPage /></Suspense>} />
                      <Route path="/debts" element={<Suspense fallback={<PageLoader />}><DebtsPage /></Suspense>} />
                      <Route path="/suppliers" element={<Suspense fallback={<PageLoader />}><SuppliersPage /></Suspense>} />
                      <Route path="/purchases" element={<Suspense fallback={<PageLoader />}><PurchasesPage /></Suspense>} />
                      <Route path="/expenses" element={<Suspense fallback={<PageLoader />}><ExpensesPage /></Suspense>} />
                      <Route path="/returns" element={<Suspense fallback={<PageLoader />}><ReturnsPage /></Suspense>} />
                      <Route path="/warranties" element={<Suspense fallback={<PageLoader />}><WarrantiesPage /></Suspense>} />
                      <Route path="/reports" element={<Suspense fallback={<PageLoader />}><ReportsPage /></Suspense>} />
                      <Route path="/settings" element={<Suspense fallback={<PageLoader />}><SettingsPage /></Suspense>} />
                      <Route path="/design-system" element={<Suspense fallback={<PageLoader />}><DesignSystemShowcase /></Suspense>} />
                      <Route path="*" element={<Navigate to="/app" replace />} />
                    </Routes>
                  </AppLayout>
                </ProtectedRoute>
              }
            />

            {/* Default redirect */}
            <Route path="/" element={<Navigate to="/app" replace />} />
            <Route path="*" element={<Navigate to="/app" replace />} />
          </Routes>
        </Router>
        <ToastContainer />
      </QueryProvider>
    </ErrorBoundary>
  );
}

export default App;