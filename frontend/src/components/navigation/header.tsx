import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Bell, Moon, Sun, Globe, User, LogOut, Menu, ShoppingCart, Plus, CreditCard, Users } from 'lucide-react';
import { useTranslation } from '../../hooks/useTranslation';
import { SearchInput } from '../ui/search-input';
import { Button } from '../ui/button';
import { IconButton } from '../ui/icon-button';
import { cn } from '../../utils';
import { useAuthStore } from '../../stores/authStore';
import { useUIStore } from '../../stores/uiStore';

interface HeaderProps {
  onToggleSidebar: () => void;
}

export function Header({ onToggleSidebar }: HeaderProps) {
  const { t, currentLanguage, languages, changeLanguage } = useTranslation();
  const navigate = useNavigate();
  const logout = useAuthStore((state) => state.logout);
  const user = useAuthStore((state) => state.user);
  const { theme, setTheme } = useUIStore();
  const [isLangDropdownOpen, setIsLangDropdownOpen] = useState(false);
  const langDropdownRef = useRef<HTMLDivElement>(null);
  const [headerSearchQuery, setHeaderSearchQuery] = useState('');

  const handleClearHeaderSearch = () => {
    setHeaderSearchQuery('');
  };

  // Initialize theme from store
  useEffect(() => {
    const isDark = theme !== 'light';
    if (isDark) {
      document.documentElement.classList.remove('light');
      document.body.classList.remove('light');
    } else {
      document.documentElement.classList.add('light');
      document.body.classList.add('light');
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

          {/* Search */}
          <div className="hidden md:block">
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
              title={t('dashboard.addProduct')}
              variant="success"
            />
            <IconButton
              onClick={() => navigate('/app/customers')}
              icon={<Users style={{ width: '16px', height: '16px' }} />}
              title={t('dashboard.addCustomer')}
            />
            <IconButton
              onClick={() => navigate('/app/debts')}
              icon={<CreditCard style={{ width: '16px', height: '16px' }} />}
              title={t('dashboard.recordPayment')}
              variant="danger"
            />
          </div>

          {/* Separator */}
          <div className="hidden lg:block w-px h-4" style={{ background: 'var(--border-default)' }} />

           {/* System Actions */}
           <div className="flex items-center gap-1">
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
             <IconButton
               icon={<Bell style={{ width: '16px', height: '16px' }} />}
               title="Notifications"
             >
               <span className="absolute top-1 right-1 w-2 h-2 rounded-full" style={{ 
                 background: 'var(--color-danger)'
               }} />
             </IconButton>
           </div>

           {/* Separator */}
           <div className="w-px h-4" style={{ background: 'var(--border-default)' }} />

           {/* User Section */}
           <div className="flex items-center gap-2">
             <div className="flex items-center justify-center" style={{
               width: '32px',
               height: '32px',
               borderRadius: '8px',
               background: 'rgba(99, 102, 241, 0.08)',
               border: '1px solid rgba(99, 102, 241, 0.15)',
               color: 'var(--color-primary)',
               boxShadow: '0 1px 2px rgba(99, 102, 241, 0.05)'
             }}>
               <User style={{ width: '16px', height: '16px' }} />
             </div>
             <div className="hidden sm:block">
               <p className="text-xs font-medium" style={{ color: 'var(--text-primary)' }}>
                 {user?.first_name || 'Admin'}
               </p>
             </div>
             <IconButton
               onClick={handleLogout}
               icon={<LogOut style={{ width: '16px', height: '16px' }} />}
               title="Logout"
               variant="danger"
             />
           </div>
        </div>
      </header>
    </>
  );
}