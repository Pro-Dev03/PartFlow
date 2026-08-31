import { lazy, Suspense, useEffect } from 'react';
import { HashRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryProvider } from './app/providers/QueryProvider';
import { AppLayout, AuthLayout } from './layouts';
import { useAuthStore, validateSubscriptionWithCloud } from './stores/authStore';
import { ErrorBoundary } from './components/ui/error-boundary';
import { ToastContainer } from './components/ui/ToastContainer';
import { appRoutes, PageLoader } from './app/router';
import { LayoutProvider } from './contexts/LayoutContext';
import AIAssistantWrapper from './components/ui/ai-assistant-wrapper';

// Lazy load auth pages separately
const LoginPage = lazy(() => import('./features/auth/pages/LoginPage').then(m => ({ default: m.LoginPage })));
const SubscriptionExpiredPage = lazy(() => import('./features/auth/pages/SubscriptionExpiredPage').then(m => ({ default: m.default })));

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, token, sessionVerified } = useAuthStore();

  if (!isAuthenticated || !token || !sessionVerified || !navigator.onLine) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, sessionVerified } = useAuthStore();
  
  if (isAuthenticated && sessionVerified) {
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

  // Keep cloud subscription authority active even while business data remains
  // local. A browser/Electron "online" event is only a trigger; the cloud
  // response is the actual authority.
  useEffect(() => {
    const validate = () => {
      void validateSubscriptionWithCloud();
    };
    const handleOffline = () => {
      const state = useAuthStore.getState();
      if (state.isAuthenticated) {
        state.logout();
      }
      window.location.hash = '#/login';
    };
    window.addEventListener('online', validate);
    window.addEventListener('offline', handleOffline);
    // The cloud is the subscription authority. If connectivity disappears, the
    // protected route stops new operations until it returns.
    const interval = window.setInterval(() => {
      const state = useAuthStore.getState();
      if (!state.isAuthenticated || !state.token) return;
      if (navigator.onLine) validate();
    }, 5 * 60 * 1000);

    return () => {
      window.removeEventListener('online', validate);
      window.removeEventListener('offline', handleOffline);
      window.clearInterval(interval);
    };
  }, []);

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
