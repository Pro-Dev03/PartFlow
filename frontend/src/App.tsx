import { lazy, Suspense, useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import { HashRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryProvider } from './app/providers/QueryProvider';
import { AppLayout, AuthLayout } from './layouts';
import { forceLogoutToLogin, markCloudVerificationPending, retrySubscriptionVerification, useAuthStore, validateSubscriptionWithCloud } from './stores/authStore';
import { ErrorBoundary } from './design-system/components/error-boundary';
import { ToastContainer } from './design-system/components/ToastContainer';
import { appRoutes, PageLoader } from './app/router';
import { LayoutProvider } from './contexts/LayoutContext';
import AIAssistantWrapper from './design-system/components/ai-assistant-wrapper';
import { InitialDataSyncModal } from './features/settings/components/InitialDataSyncModal';
import { isInitialSyncNeeded } from './hooks/useInitialDataSync';
import { authApi } from './services/api/endpoints';
import { SubscriptionVerificationScreen } from './features/auth/components/SubscriptionVerificationScreen';
import { initializeProductImages } from './services/localProductImages';
import { initializePartTypeImages } from './services/localPartTypeImages';
import { initializeCategoryImages } from './services/localCategoryImages';
import { RegionalProfileLoader } from './components/RegionalProfileLoader';
import { AlertTriangle, RefreshCw } from 'lucide-react';

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

    // The admin check is only needed when the optional automatic initial sync
    // is enabled. Avoid probing a protected endpoint on every app load when
    // there is no sync modal to open; this also prevents a stale local token
    // from producing a misleading 401 in the browser console during normal
    // offline-first use.
    const autoSyncEnabled = localStorage.getItem('partflow-auto-sync-enabled') === 'true';
    if (!autoSyncEnabled) {
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
  const { isAuthenticated, sessionVerified, cloudVerificationPending, isLoading, isPostLoginVerifying } = useAuthStore();
  const [isRetryingCloudVerification, setIsRetryingCloudVerification] = useState(false);

  useEffect(() => {
    void initializeProductImages();
    void initializePartTypeImages();
    void initializeCategoryImages();
  }, []);

  const retryCloudVerification = async () => {
    setIsRetryingCloudVerification(true);
    try {
      await retrySubscriptionVerification();
    } finally {
      setIsRetryingCloudVerification(false);
    }
  };

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
        if (!valid) return;
      });
    };
    const handleOffline = () => {
      const state = useAuthStore.getState();
      if (state.isAuthenticated || state.token || state.cloudToken) {
        markCloudVerificationPending();
      }
    };
    const handleAuthInvalidated = (event: Event) => {
      // A rejected/expired access token is recoverable through refresh or a
      // later online validation. Subscription expiry is handled only by the
      // explicit SUBSCRIPTION_EXPIRED response in validateSubscriptionWithCloud.
      markCloudVerificationPending();
    };
    const handleCloudVerificationPending = () => {
      markCloudVerificationPending();
    };
    window.addEventListener('online', validate);
    window.addEventListener('offline', handleOffline);
    window.addEventListener('partflow:auth-invalidated', handleAuthInvalidated);
    window.addEventListener('partflow:cloud-verification-pending', handleCloudVerificationPending);
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
      window.removeEventListener('partflow:auth-invalidated', handleAuthInvalidated);
      window.removeEventListener('partflow:cloud-verification-pending', handleCloudVerificationPending);
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
        {isAuthenticated && sessionVerified && <RegionalProfileLoader />}
        {isAuthenticated && sessionVerified && cloudVerificationPending && typeof document !== 'undefined' && createPortal(
          <div
            role="status"
            className="fixed inset-0 z-[100] flex min-h-screen items-center justify-center bg-slate-950/20 px-4 py-6 backdrop-blur-[2px]"
            style={{ position: 'fixed', top: 0, right: 0, bottom: 0, left: 0, width: '100vw', height: '100vh' }}
          >
            <div className="w-full max-w-md rounded-2xl border border-amber-200 bg-white p-6 text-center shadow-2xl" dir="rtl">
              <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-amber-100 text-amber-700">
                <AlertTriangle className="h-7 w-7" aria-hidden="true" />
              </div>
              <h2 className="mt-4 text-lg font-bold text-slate-900">تعذر التحقق من الاشتراك</h2>
              <p className="mt-2 text-sm leading-6 text-slate-600">
                تعذر الاتصال بخدمة الاشتراك في الوقت الحالي. أوقفنا عمليات الحفظ مؤقتًا لحماية بياناتك، وسيُعاد التحقق تلقائيًا عند عودة الاتصال.
              </p>
              <button
                type="button"
                className="mt-5 inline-flex min-h-11 items-center justify-center gap-2 rounded-lg bg-amber-600 px-5 text-sm font-bold text-white transition hover:bg-amber-700 disabled:cursor-not-allowed disabled:opacity-60"
                onClick={() => void retryCloudVerification()}
                disabled={isRetryingCloudVerification}
              >
                <RefreshCw className={`h-4 w-4 ${isRetryingCloudVerification ? 'animate-spin' : ''}`} aria-hidden="true" />
                {isRetryingCloudVerification ? 'جارٍ التحقق...' : 'إعادة التحقق الآن'}
              </button>
            </div>
          </div>,
          document.body,
        )}
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
