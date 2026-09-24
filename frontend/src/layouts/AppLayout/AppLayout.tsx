import { useState, useRef, useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { Header } from '../../components/navigation/header';
import { Sidebar } from '../../components/navigation/sidebar';
import { ScrollIndicator, ScrollProgress } from '../../design-system/components/scroll-indicator';
import { useTranslation } from '../../hooks/useTranslation';
import { useUIStore } from '../../stores/uiStore';
import { useLayout } from '../../contexts/LayoutContext';
import { cn } from '../../utils';
import { ChevronUp } from 'lucide-react';

interface AppLayoutProps {
  children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  const { direction } = useTranslation();
  const { sidebarCollapsed, checkoutMode, toggleSidebar, theme } = useUIStore();
  const { fullWidth } = useLayout();
  const location = useLocation();
  const isCheckoutRoute = location.pathname === '/app/sales' || location.pathname.startsWith('/app/sales/');
  const hideNavigation = checkoutMode && isCheckoutRoute;
  const [showScrollTop, setShowScrollTop] = useState(false);
  const mainRef = useRef<HTMLDivElement | null>(null);

  // Handle scroll to show/hide scroll-to-top button
  useEffect(() => {
    const handleScroll = () => {
      if (mainRef.current) {
        const scrollTop = mainRef.current.scrollTop;
        setShowScrollTop(scrollTop > 300);
      }
    };

    const mainElement = mainRef.current;
    if (mainElement) {
      mainElement.addEventListener('scroll', handleScroll);
      return () => mainElement.removeEventListener('scroll', handleScroll);
    }
    return () => {};
  }, []);

  // Scroll to top function
  const scrollToTop = () => {
    if (mainRef.current) {
      mainRef.current.scrollTo({
        top: 0,
        behavior: 'smooth'
      });
    }
  };





  return (
    <div
      dir={direction}
      className={cn('pf-app-shell', theme === 'light' && 'light')}
      style={{
        minHeight: '100vh',
        background: theme !== 'light' ? 'var(--bg-background)' : 'var(--bg-gradient-light), var(--bg-background)'
      }}
    >
      <div className="pf-app-frame" style={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
        {!hideNavigation && <Header onToggleSidebar={toggleSidebar} />}
        <div style={{ display: 'flex', flex: 1, minWidth: 0, alignItems: 'stretch' }}>
          {!hideNavigation && (
            <Sidebar
              isCollapsed={sidebarCollapsed}
            />
          )}
          <main
            ref={mainRef}
            id="main-content"
            style={{
              flex: 1,
              minWidth: 0,
              maxWidth: hideNavigation ? '100%' : (fullWidth ? '100%' : '1500px'),
              margin: hideNavigation ? '0' : (fullWidth ? '0' : '0 auto'),
              width: '100%',
              padding: hideNavigation ? '0' : '20px 24px'
            }}
            className={cn('px-4 md:px-8 lg:px-8', hideNavigation && 'checkout-mode')}
            data-app-main="true"
          >
            {children}
          </main>
        </div>
      </div>

      {/* Scroll UX Components - using window scroll */}
      <ScrollIndicator />
      <ScrollProgress />

      {/* Scroll to top button */}
      <button
        onClick={scrollToTop}
        className={cn('scroll-to-top', showScrollTop && 'visible')}
        aria-label="Scroll to top"
      >
        <ChevronUp className="w-5 h-5" />
      </button>
    </div>
  );
}
