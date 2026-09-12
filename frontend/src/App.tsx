import { lazy, Suspense, useEffect, useState } from 'react';
import { HashRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryProvider } from './app/providers/QueryProvider';
import { AppLayout, AuthLayout } from './layouts';
import { useAuthStore, validateSubscriptionWithCloud } from './stores/authStore';
import { ErrorBoundary } from './components/ui/error-boundary';
import { ToastContainer } from './components/ui/ToastContainer';
import { appRoutes, PageLoader } from './app/router';
import { LayoutProvider } from './contexts/LayoutContext';
import AIAssistantWrapper from './components/ui/ai-assistant-wrapper';
import { InitialDataSyncModal } from './features/settings/components/InitialDataSyncModal';
import { isInitialSyncNeeded } from './hooks/useInitialDataSync';
import { authApi } from './services/api/endpoints';
import { SubscriptionVerificationScreen } from './features/auth/components/SubscriptionVerificationScreen';

// Lazy load auth pages separately
const LoginPage = lazy(() => import('./features/auth/pages/LoginPage').then(m => ({ default: m.LoginPage })));
const SubscriptionExpiredPage = lazy(() => import('./features/auth/pages/SubscriptionExpiredPage').then(m => ({ default: m.default })));

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

// Hydrate the local SQLite business database once after cloud authentication.
// Subscription/authentication remain cloud-authoritative; only operational
// rows are downloaded into the local database by this flow.
function InitialSyncController() {
  const { isAuthenticated, sessionVerified, user } = useAuthStore();
  const [isOpen, setIsOpen] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    let mounted = true;
    if (!isAuthenticated || !sessionVerified) {
      setIsAdmin(false);
      setIsOpen(false);
      return () => { mounted = false; };
    }

    // Operational cloud snapshots are global in the current schema. Only the
    // configured administrator may download one until tenant isolation exists.
    void authApi.checkAdminAccess()
      .then((result) => {
        if (mounted) {
          setIsAdmin(Boolean(result?.is_admin));
        }
      })
      .catch(() => { if (mounted) setIsAdmin(false); });

    return () => { mounted = false; };
  }, [isAuthenticated, sessionVerified]);

  useEffect(() => {
    const autoSyncEnabled = localStorage.getItem('partflow-auto-sync-enabled') === 'true';

    if (autoSyncEnabled && isAdmin && isAuthenticated && sessionVerified && navigator.onLine && isInitialSyncNeeded(user?.id)) {
      setIsOpen(true);
    } else if (!isAuthenticated || !sessionVerified || !isAdmin) {
      setIsOpen(false);
    }
  }, [isAdmin, isAuthenticated, sessionVerified, user?.id]);

  return (
    <InitialDataSyncModal
      isOpen={isOpen}
      userId={user?.id}
      onComplete={() => setIsOpen(false)}
    />
  );
}

function App() {
  const checkAuth = useAuthStore((state) => state.checkAuth);
  const { isAuthenticated, sessionVerified, isLoading, isPostLoginVerifying } = useAuthStore();

  // HashRouter is required by the packaged Electron build, but a normal
  // browser can still open a deep link such as /app/sales directly. Normalize
  // that path into the hash before React Router reads it; otherwise HashRouter
  // silently falls back to the dashboard and POS smoke tests look broken.
  useEffect(() => {
    if (!window.location.pathname.startsWith('/app')) {
      return;
    }
    const hashRoute = window.location.hash.replace(/^#/, '').split('?')[0];
    if (hashRoute && hashRoute !== '/app' && hashRoute !== '/') {
      return;
    }
    const route = `${window.location.pathname}${window.location.search}`;
    window.location.hash = route;
  }, []);

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  // Keep cloud subscription authority active even while business data remains
  // local. A browser/Electron "online" event is only a trigger; the cloud
  // response is the actual authority.
  useEffect(() => {
    const validate = () => {
      void validateSubscriptionWithCloud().then((valid) => {
        if (!valid && useAuthStore.getState().isAuthenticated) {
          useAuthStore.getState().logout();
          window.location.hash = '#/login';
        }
      });
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

  if (isPostLoginVerifying) {
    return (
      <ErrorBoundary>
        <QueryProvider>
          <SubscriptionVerificationScreen />
        </QueryProvider>
      </ErrorBoundary>
    );
  }

  if (isLoading) {
    return (
      <ErrorBoundary>
        <QueryProvider>
          <PageLoader />
        </QueryProvider>
      </ErrorBoundary>
    );
  }

  return (
    <ErrorBoundary>
      <QueryProvider>
        <InitialSyncController />
        <Router>
          <PagePreloader />
          <Routes>
            {/* Public routes */}
            <Route
              path="/login"
              element={
                !isAuthenticated || !sessionVerified ? (
                  <AuthLayout>
                    <Suspense fallback={<PageLoader />}>
                      <LoginPage />
                    </Suspense>
                  </AuthLayout>
                ) : isPostLoginVerifying ? (
                  <SubscriptionVerificationScreen />
                ) : (
                  <Navigate to="/app" replace />
                )
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
                isAuthenticated && sessionVerified ? (
                  <LayoutProvider>
                    <AppLayout>
                      <Suspense fallback={<PageLoader />}>
                        {appRoutes}
                      </Suspense>
                      <AIAssistantWrapper />
                    </AppLayout>
                  </LayoutProvider>
                ) : (
                  <Navigate to="/login" replace />
                )
              }
            />

            {/* Default redirect */}
            <Route 
              path="/" 
              element={
                <Navigate to={isAuthenticated && sessionVerified ? '/app' : '/login'} replace />
              } 
            />
            <Route 
              path="*" 
              element={
                <Navigate to={isAuthenticated && sessionVerified ? '/app' : '/login'} replace />
              } 
            />
          </Routes>
        </Router>
        <ToastContainer />
      </QueryProvider>
    </ErrorBoundary>
  );
}

export default App;
