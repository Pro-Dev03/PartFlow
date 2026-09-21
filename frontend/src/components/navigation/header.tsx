import { useState, useRef, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Bell, Moon, Sun, Globe, User, LogOut, Menu, ShoppingCart, Plus, Users, LayoutDashboard, Package, Wifi, HardDrive } from 'lucide-react';
import { useTranslation } from '../../hooks/useTranslation';
import { SearchInput } from '../../design-system/components/search-input';
import { Button } from '../../design-system/components/button';
import { IconButton } from '../../design-system/components/icon-button';
import { useAuthStore } from '../../stores/authStore';
import { useUIStore } from '../../stores/uiStore';
import { getConnectionMode } from '../../lib/config/app';

interface HeaderProps {
  onToggleSidebar: () => void;
}

export function Header({ onToggleSidebar }: HeaderProps) {
  const { t, languages, changeLanguage } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const logout = useAuthStore((state) => state.logout);
  const user = useAuthStore((state) => state.user);
  const { theme, setTheme } = useUIStore();
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

  // Initialize theme from store
  useEffect(() => {
    const isDark = theme !== 'light';
    if (isDark) {
      document.documentElement.classList.remove('light');
      document.body.classList.remove('light');
      document.documentElement.classList.add('dark');
      document.body.classList.add('dark');
    } else {
      document.documentElement.classList.add('light');
      document.body.classList.add('light');
      document.documentElement.classList.remove('dark');
      document.body.classList.remove('dark');
    }
  }, [theme]);

  const toggleTheme = () => {
    const newTheme = theme === 'light' ? 'dark' : 'light';
    setTheme(newTheme);
    localStorage.setItem('theme', newTheme);
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
  const isLightTheme = theme === 'light';

  return (
    <>
      <header className="h-16 flex items-center justify-between px-lg" style={{
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
            aria-label="Toggle sidebar"
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

          {/* Search */}
          <div className="hidden xl:block">
            <SearchInput
              placeholder={t('common.search') || "بحث المنتجات، العملاء، الفواتير..."}
              value={headerSearchQuery}
              onChange={(e) => setHeaderSearchQuery(e.target.value)}
              onClear={handleClearHeaderSearch}
              className="w-80 lg:w-[480px]"
              size="sm"
            />
          </div>
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
               className="flex items-center gap-2 rounded-full border px-2.5 py-1.5 transition-all duration-300"
               style={{
                 background: 'rgba(16, 185, 129, 0.14)',
                 borderColor: 'rgba(16, 185, 129, 0.38)',
                 boxShadow: '0 0 0 1px rgba(16,185,129,0.15), 0 0 18px rgba(16,185,129,0.6), 0 0 30px rgba(16,185,129,0.42)'
               }}
             >
               <span
                 className="relative flex items-center justify-center rounded-full"
                 style={{
                   width: '26px',
                   height: '26px',
                   background: 'radial-gradient(circle, rgba(52,211,153,1) 0%, rgba(16,185,129,1) 42%, rgba(5,150,105,1) 100%)',
                   boxShadow: '0 0 12px rgba(52,211,153,0.9), 0 0 22px rgba(16,185,129,0.8), inset 0 0 10px rgba(255,255,255,0.4)',
                   animation: 'pulse 1.8s ease-in-out infinite',
                 }}
               >
                 <svg
                   xmlns="http://www.w3.org/2000/svg"
                   width="24"
                   height="24"
                   viewBox="0 0 24 24"
                   fill="none"
                   stroke="currentColor"
                   strokeWidth="2"
                   strokeLinecap="round"
                   strokeLinejoin="round"
                   className="relative h-3.5 w-3.5"
                   style={{ color: '#ecfdf5' }}
                   aria-hidden="true"
                 >
                   <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
                   <circle cx="12" cy="7" r="4" />
                 </svg>
               </span>
               <span
                 className="text-[11px] font-bold tracking-[0.08em]"
                 style={{
                   color: isLightTheme ? '#0f172a' : '#ffffff',
                   textShadow: isLightTheme ? 'none' : '0 0 10px rgba(255,255,255,0.65)',
                   whiteSpace: 'nowrap',
                   opacity: 1,
                 }}
               >
                 {isCloudConnection ? 'سحابي' : 'محلي'}
               </span>
             </div>
           </div>
        </div>
      </header>
    </>
  );
}
