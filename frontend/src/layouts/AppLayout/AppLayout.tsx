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
  const sidebarCollapsed = useUIStore((state) => state.sidebarCollapsed);
  const checkoutMode = useUIStore((state) => state.checkoutMode);
  const toggleSidebar = useUIStore((state) => state.toggleSidebar);
  const { fullWidth } = useLayout();
  const location = useLocation();
  const isCheckoutRoute = location.pathname === '/app/sales' || location.pathname.startsWith('/app/sales/');
  const isPosRoute = isCheckoutRoute;
  const hideNavigation = checkoutMode && isCheckoutRoute;
  const [showScrollTop, setShowScrollTop] = useState(false);
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false);
  const mainRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    setMobileSidebarOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    const main = mainRef.current;
    if (!main || isPosRoute || window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

    const revealSelector = '.pf-page-header, .pf-card:not(.pf-card-open), .pf-data-table, .unified-stat-card, .table-container';
    const observedElements = new Set<HTMLElement>();
    let revealObserver: IntersectionObserver | null = null;
    if (typeof IntersectionObserver !== 'undefined') {
      revealObserver = new IntersectionObserver((entries) => {
          entries.forEach((entry) => {
            if (!entry.isIntersecting) return;
            entry.target.classList.add('pf-revealed');
            revealObserver?.unobserve(entry.target);
          });
        }, { root: main, threshold: 0.08, rootMargin: '0px 0px -24px 0px' });
    }

    const reveal = (element: Element) => {
      if (!(element instanceof HTMLElement) || observedElements.has(element) || observedElements.size >= 80) return;
      observedElements.add(element);
      element.classList.add('pf-scroll-reveal');
      element.style.setProperty('--pf-reveal-delay', `${Math.min((observedElements.size - 1) * 35, 210)}ms`);
      if (revealObserver) revealObserver.observe(element);
      else element.classList.add('pf-revealed');
    };

    const scan = (node: ParentNode) => {
      if (node instanceof Element && node.matches(revealSelector)) reveal(node);
      node.querySelectorAll?.(revealSelector).forEach(reveal);
    };

    scan(main);
    const mutationObserver = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => mutation.addedNodes.forEach((node) => {
        if (node instanceof Element) scan(node);
      }));
    });
    mutationObserver.observe(main, { childList: true, subtree: true });

    return () => {
      mutationObserver.disconnect();
      revealObserver?.disconnect();
      observedElements.forEach((element) => {
        element.classList.remove('pf-scroll-reveal', 'pf-revealed');
        element.style.removeProperty('--pf-reveal-delay');
      });
    };
  }, [isPosRoute, location.pathname]);

  useEffect(() => {
    if (!mobileSidebarOpen) return;
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setMobileSidebarOpen(false);
    };
    window.addEventListener('keydown', closeOnEscape);
    return () => window.removeEventListener('keydown', closeOnEscape);
  }, [mobileSidebarOpen]);

  const handleToggleSidebar = () => {
    if (window.matchMedia('(max-width: 700px)').matches) {
      setMobileSidebarOpen((open) => !open);
    } else {
      toggleSidebar();
    }
  };

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
      className={cn('pf-app-shell', isPosRoute && 'pos-viewport-shell', hideNavigation && 'checkout-mode-shell')}
      style={{
        width: isPosRoute ? '100%' : undefined,
        minHeight: isPosRoute ? 0 : '100vh',
        height: isPosRoute ? '100dvh' : undefined,
        overflow: isPosRoute ? 'hidden' : undefined,
        margin: isPosRoute ? 0 : undefined,
        padding: isPosRoute ? 0 : undefined,
        boxSizing: isPosRoute ? 'border-box' : undefined
      }}
    >
      <div
        className="pf-app-frame"
        style={{
          display: 'flex',
          flexDirection: 'column',
          minHeight: isPosRoute ? 0 : '100vh',
          height: isPosRoute ? '100%' : undefined,
          overflow: isPosRoute ? 'hidden' : undefined,
        }}
      >
        {!hideNavigation && <Header onToggleSidebar={handleToggleSidebar} sidebarOpen={mobileSidebarOpen} />}
        <div
          style={{
            display: 'flex',
            flex: 1,
            minWidth: 0,
            minHeight: 0,
            alignItems: 'stretch',
            overflow: isPosRoute ? 'hidden' : undefined,
          }}
        >
          {!hideNavigation && (
            <>
              {mobileSidebarOpen && (
                <button
                  type="button"
                  className="pf-mobile-nav-backdrop"
                  aria-label="إغلاق القائمة"
                  onClick={() => setMobileSidebarOpen(false)}
                />
              )}
              <Sidebar isCollapsed={sidebarCollapsed && !mobileSidebarOpen} mobileOpen={mobileSidebarOpen} />
            </>
          )}
          <main
            ref={mainRef}
            id="main-content"
            style={{
              flex: 1,
              minWidth: 0,
              minHeight: 0,
              height: isPosRoute ? '100%' : undefined,
              overflow: isPosRoute ? 'hidden' : undefined,
              maxWidth: isPosRoute || fullWidth ? '100%' : '1500px',
              margin: isPosRoute || fullWidth ? '0' : '0 auto',
              width: '100%',
              padding: isPosRoute ? '0' : '20px 24px',
              boxSizing: 'border-box',
            }}
            className={cn('px-4 md:px-8 lg:px-8', isPosRoute && 'pos-viewport-main', hideNavigation && 'checkout-mode')}
            data-app-main="true"
          >
            {children}
          </main>
        </div>
      </div>

      {!isPosRoute && (
        <>
          <ScrollIndicator />
          <ScrollProgress />
          <button
            onClick={scrollToTop}
            className={cn('scroll-to-top', showScrollTop && 'visible')}
            aria-label="Scroll to top"
          >
            <ChevronUp className="w-5 h-5" />
          </button>
        </>
      )}
    </div>
  );
}
