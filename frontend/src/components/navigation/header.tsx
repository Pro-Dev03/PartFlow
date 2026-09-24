import { useState, useRef, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Moon, Sun, Globe, LogOut, Menu, RefreshCw, ShoppingCart, Plus, Users, LayoutDashboard, Package } from 'lucide-react';
import { useTranslation } from '../../hooks/useTranslation';
import { SearchInput } from '../../design-system/components/search-input';
import { Button } from '../../design-system/components/button';
import { IconButton } from '../../design-system/components/icon-button';
import { useAuthStore } from '../../stores/authStore';
import { useUIStore } from '../../stores/uiStore';
import { getConnectionMode } from '../../lib/config/app';

interface HeaderProps {
  onToggleSidebar: () => void;
  sidebarOpen?: boolean;
}

const sectionNames: Array<[string, string]> = [
  ['/app/dashboard', 'لوحة التحكم'],
  ['/app/activity', 'النشاط'],
  ['/app/sales', 'نقطة البيع'],
  ['/app/inventory', 'المخزون'],
  ['/app/customers', 'الزبائن'],
  ['/app/debts', 'الديون'],
  ['/app/suppliers', 'التجار'],
  ['/app/purchases', 'المشتريات'],
  ['/app/expenses', 'المصروفات'],
  ['/app/returns', 'المرتجعات'],
  ['/app/supplier-returns', 'مرتجعات التجار'],
  ['/app/reports', 'التقارير'],
  ['/app/categories', 'التصنيفات'],
  ['/app/audit', 'سجل النظام'],
  ['/app/settings', 'الإعدادات'],
];

export function Header({ onToggleSidebar, sidebarOpen = false }: HeaderProps) {
  const { t, languages, changeLanguage } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const logout = useAuthStore((state) => state.logout);
  const user = useAuthStore((state) => state.user);
  const theme = useUIStore((state) => state.theme);
  const setTheme = useUIStore((state) => state.setTheme);
  const [isLangDropdownOpen, setIsLangDropdownOpen] = useState(false);
  const langDropdownRef = useRef<HTMLDivElement>(null);
  const [headerSearchQuery, setHeaderSearchQuery] = useState('');
  const [connectionMode, setConnectionMode] = useState<'local' | 'cloud'>(getConnectionMode());
  const primaryNavigation = [
    { label: 'لوحة التحكم', path: '/app/dashboard', icon: LayoutDashboard },
    { label: 'نقطة البيع', path: '/app/sales', icon: ShoppingCart },
    { label: 'الزبائن', path: '/app/customers', icon: Users },
    { label: 'المخزون', path: '/app/inventory', icon: Package },
  ];

  const handleClearHeaderSearch = () => {
    setHeaderSearchQuery('');
  };

  const toggleTheme = () => {
    const newTheme = theme === 'light' ? 'dark' : 'light';
    setTheme(newTheme);
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  useEffect(() => {
    const syncConnectionMode = () => setConnectionMode(getConnectionMode());
    syncConnectionMode();

    const handleConnectionModeChanged = () => syncConnectionMode();
    window.addEventListener('partflow:connection-mode-changed', handleConnectionModeChanged);
    window.addEventListener('partflow:cloud-api-url-changed', handleConnectionModeChanged);

    return () => {
      window.removeEventListener('partflow:connection-mode-changed', handleConnectionModeChanged);
      window.removeEventListener('partflow:cloud-api-url-changed', handleConnectionModeChanged);
    };
  }, []);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (langDropdownRef.current && !langDropdownRef.current.contains(event.target as Node)) {
        setIsLangDropdownOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const isCloudConnection = connectionMode === 'cloud';
  const currentSection = sectionNames.find(([path]) => location.pathname === path || location.pathname.startsWith(`${path}/`))?.[1] || 'لوحة التحكم';

  const handleHeaderSearch = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const query = headerSearchQuery.trim();
    if (query) navigate(`/app/inventory?search=${encodeURIComponent(query)}`);
  };

  return (
    <>
      <header className="pf-app-header h-16 flex items-center justify-between px-lg" style={{
        background: 'var(--topbar-bg)', // Updated to use design system variable
        borderBottom: '1px solid var(--topbar-border)', // Updated to use design system variable
        boxShadow: 'var(--shadow-sm)' // Added subtle shadow from design system
      }}>
        {/* Left side */}
        <div className="flex items-center gap-md">
          <Button
            variant="ghost"
            size="icon"
            onClick={onToggleSidebar}
            aria-label="فتح أو إغلاق القائمة"
            aria-controls="app-sidebar"
            aria-expanded={sidebarOpen}
          >
            <Menu className="w-4 h-4 flip-rtl" />
          </Button>

          <nav className="header-primary-nav flex shrink-0 items-center gap-1" aria-label="التنقل الرئيسي">
            {primaryNavigation.map(({ label, path, icon: Icon }) => {
              const isActive = location.pathname === path || location.pathname.startsWith(`${path}/`);
              return (
                <Button
                  key={path}
                  type="button"
                  variant={isActive ? 'primary' : 'ghost'}
                  size="sm"
                  className="gap-1.5"
                  aria-current={isActive ? 'page' : undefined}
                  onClick={() => navigate(path)}
                >
                  <Icon className="h-4 w-4" />
                  <span>{label}</span>
                </Button>
              );
            })}
          </nav>

          <span className="pf-header-location" aria-label={`القسم الحالي: ${currentSection}`}>
            <span>مساحة العمل</span>
            <strong>{currentSection}</strong>
          </span>

          {/* Search */}
          <form className="pf-header-search" role="search" onSubmit={handleHeaderSearch}>
            <SearchInput
              placeholder="بحث في المخزون..."
              value={headerSearchQuery}
              onChange={(e) => setHeaderSearchQuery(e.target.value)}
              onClear={handleClearHeaderSearch}
              className="w-full"
              size="sm"
            />
          </form>
        </div>

        {/* Right side */}
        <div className="flex items-center gap-2">
          {/* Quick Actions - Soft UI Icons */}
          <div className="hidden lg:flex items-center gap-1">
            <IconButton
              onClick={() => navigate('/app/sales')}
              icon={<ShoppingCart style={{ width: '16px', height: '16px' }} />}
              title={t('dashboard.newSale')}
            />
            <IconButton
              onClick={() => navigate('/app/inventory')}
              icon={<Plus style={{ width: '16px', height: '16px' }} />}
              title="إضافة مخزون"
              variant="success"
            />
            <IconButton
              onClick={() => navigate('/app/customers')}
              icon={<Users style={{ width: '16px', height: '16px' }} />}
              title={t('dashboard.addCustomer')}
            />
          </div>

          {/* Separator */}
          <div className="hidden lg:block w-px h-4" style={{ background: 'var(--border-default)' }} />

           {/* System Actions */}
           <div className="relative flex items-center gap-1">
             {/* Language */}
             <div className="relative" ref={langDropdownRef}>
               <IconButton
                 onClick={() => setIsLangDropdownOpen(!isLangDropdownOpen)}
                 icon={<Globe style={{ width: '16px', height: '16px' }} />}
                 title="Change language"
                 aria-label="Change language"
               />

               {isLangDropdownOpen && (
                 <div
                   className="absolute top-full end-0 mt-2 w-48 rounded-lg shadow-lg z-dropdown overflow-hidden"
                   style={{
                     background: 'var(--bg-surface-elevated)',
                     border: '1px solid var(--border-default)'
                   }}
                   role="menu"
                 >
                   {languages.map((lang) => (
                     <button
                       key={lang.code}
                       onClick={() => {
                         changeLanguage(lang.code);
                         setIsLangDropdownOpen(false);
                       }}
                       className="w-full px-3 py-2 text-start hover:bg-white/5 transition-colors flex items-center gap-2"
                       role="menuitem"
                       style={{ color: 'var(--text-primary)' }}
                     >
                       <span>{lang.flag}</span>
                       <span className="text-sm">{lang.nativeName}</span>
                     </button>
                   ))}
                 </div>
               )}
             </div>

             <IconButton
               onClick={() => window.location.reload()}
               icon={<RefreshCw style={{ width: '16px', height: '16px' }} />}
               title={t('common.refresh')}
               aria-label={t('common.refresh')}
             />

             {/* Theme */}
             <IconButton
               onClick={toggleTheme}
               icon={theme !== 'light' ? (
                 <Sun style={{ width: '16px', height: '16px' }} />
               ) : (
                 <Moon style={{ width: '16px', height: '16px' }} />
               )}
               title="Toggle theme"
             />

             {/* Notifications */}
           </div>

           {/* Separator */}
           <div className="w-px h-4" style={{ background: 'var(--border-default)' }} />

           {/* User Section */}
           <div className="flex items-center gap-2">
             <div className="hidden sm:block">
               <p className="text-xs font-medium" style={{ color: 'var(--text-primary)' }}>
                 {user?.name || 'Admin'}
               </p>
             </div>
             <IconButton
               onClick={handleLogout}
               icon={<LogOut style={{ width: '16px', height: '16px' }} />}
               title="Logout"
               variant="danger"
             />
             <div
               className="pf-connection-status flex items-center gap-2 rounded-full border px-2.5 py-1.5 transition-all duration-300"
               title={isCloudConnection ? 'الوضع السحابي' : 'الوضع المحلي'}
             >
               <span className="pf-connection-dot" aria-hidden="true" />
               <span>{isCloudConnection ? 'سحابي' : 'محلي'}</span>
             </div>
           </div>
        </div>
      </header>
    </>
  );
}
