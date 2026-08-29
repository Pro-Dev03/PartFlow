import { lazy, Suspense, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryProvider } from './app/providers/QueryProvider';
import { AppLayout, AuthLayout } from './layouts';
import { useAuthStore } from './stores/authStore';
import { ErrorBoundary } from './components/ui/error-boundary';
import { ToastContainer } from './components/ui/ToastContainer';
import { appRoutes, PageLoader } from './app/router';
import { LayoutProvider } from './contexts/LayoutContext';
import { isOfflineSubscriptionBlocked } from './lib/subscription-guard';
import AIAssistantWrapper from './components/ui/ai-assistant-wrapper';

// Lazy load auth pages separately
const LoginPage = lazy(() => import('./features/auth/pages/LoginPage').then(m => ({ default: m.LoginPage })));
const SubscriptionExpiredPage = lazy(() => import('./features/auth/pages/SubscriptionExpiredPage').then(m => ({ default: m.default })));

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, token } = useAuthStore();

  if (!isAuthenticated || !token) {
    return <Navigate to="/login" replace />;
  }

  if (isOfflineSubscriptionBlocked(token)) {
    return <Navigate to="/subscription-expired" replace />;
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
  const checkAuth = useAuthStore((state) => state.checkAuth);

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

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

            <Route
              path="/subscription-expired"
              element={
                <Suspense fallback={<PageLoader />}>
                  <SubscriptionExpiredPage />
                </Suspense>
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
        <AIAssistantWrapper />
      </QueryProvider>
    </ErrorBoundary>
  );
}

export default App;