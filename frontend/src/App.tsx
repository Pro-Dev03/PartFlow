import { lazy, Suspense, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryProvider } from './app/providers/QueryProvider';
import { AppLayout, AuthLayout } from './layouts';
import { useAuthStore } from './stores/authStore';
import { ErrorBoundary } from './components/ui/error-boundary';
import { ToastContainer } from './components/ui/ToastContainer';
import { appRoutes, PageLoader } from './app/router';
import { LayoutProvider } from './contexts/LayoutContext';

// Lazy load login page separately
const LoginPage = lazy(() => import('./features/auth/pages/LoginPage').then(m => ({ default: m.LoginPage })));

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
                  <LayoutProvider>
                    <AppLayout>
                      <Suspense fallback={<PageLoader />}>
                        {appRoutes}
                      </Suspense>
                    </AppLayout>
                  </LayoutProvider>
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