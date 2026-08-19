import { useState, useRef, useEffect } from 'react';
import { Bell, Search, Moon, Sun, Globe, User, LogOut, ChevronDown, Menu } from 'lucide-react';
import { useTranslation } from '../../hooks/useTranslation';
import { Button } from '../ui/button';
import { cn } from '../../lib/utils';

interface HeaderProps {
  onToggleSidebar: () => void;
}

export function Header({ onToggleSidebar }: HeaderProps) {
  const { t, currentLanguage, languages, changeLanguage } = useTranslation();
  const [isDark, setIsDark] = useState(false);
  const [isLangDropdownOpen, setIsLangDropdownOpen] = useState(false);
  const langDropdownRef = useRef<HTMLDivElement>(null);

  const toggleTheme = () => {
    setIsDark(!isDark);
    document.documentElement.classList.toggle('light');
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
    <header className="h-16 bg-surface border-b border-border flex items-center justify-between px-lg">
      {/* Left side */}
      <div className="flex items-center gap-md">
        <Button
          variant="ghost"
          size="sm"
          onClick={onToggleSidebar}
          className="p-2"
          aria-label="Toggle sidebar"
        >
          <Menu className="w-5 h-5 flip-rtl" />
        </Button>
        
        {/* Search */}
        <div className="relative hidden md:block">
          <Search className="absolute inset-y-0 start-0 flex items-center ps-3 w-4 h-4 text-text-muted" />
          <input
            type="text"
            placeholder={t('common.search') || "بحث المنتجات، العملاء، الفواتير..."}
            className={cn(
              'w-64 lg:w-96 h-10 ps-10 pe-4 rounded-sm border border-border',
              'bg-surface/75',
              'focus:outline-none focus:ring-2 focus:ring-cyan focus:border-transparent',
              'text-body text-text placeholder:text-text-muted',
              'transition-all duration-normal'
            )}
            dir="auto"
            aria-label="Search"
          />
        </div>
      </div>

      {/* Right side */}
      <div className="flex items-center gap-sm">
        {/* Language selector */}
        <div className="relative" ref={langDropdownRef}>
          <Button 
            variant="ghost" 
            size="sm" 
            className="gap-2"
            onClick={() => setIsLangDropdownOpen(!isLangDropdownOpen)}
            aria-label="Change language"
            aria-expanded={isLangDropdownOpen}
          >
            <Globe className="w-4 h-4" />
            <span className="hidden sm:inline">{currentLanguage.toUpperCase()}</span>
            <ChevronDown className="w-4 h-4" />
          </Button>
          
          {isLangDropdownOpen && (
            <div 
              className="absolute top-full end-0 mt-2 w-48 bg-surface border border-border rounded-sm shadow-card z-dropdown"
              role="menu"
            >
              {languages.map((lang) => (
                <button
                  key={lang.code}
                  onClick={() => {
                    changeLanguage(lang.code);
                    setIsLangDropdownOpen(false);
                  }}
                  className={cn(
                    'w-full px-4 py-2 text-start hover:bg-surface/50',
                    'flex items-center gap-3 transition-colors duration-normal',
                    currentLanguage === lang.code && 'bg-surface/50'
                  )}
                  role="menuitem"
                >
                  <span className="text-xl">{lang.flag}</span>
                  <div>
                    <p className="font-medium text-small text-text">
                      {lang.nativeName}
                    </p>
                    <p className="text-tiny text-text-muted">
                      {lang.name}
                    </p>
                  </div>
                  {currentLanguage === lang.code && (
                    <span className="me-auto text-cyan">✓</span>
                  )}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Theme toggle */}
        <Button 
          variant="ghost" 
          size="sm" 
          onClick={toggleTheme}
          aria-label="Toggle theme"
        >
          {isDark ? (
            <Sun className="w-4 h-4" />
          ) : (
            <Moon className="w-4 h-4" />
          )}
        </Button>

        {/* Notifications */}
        <Button 
          variant="ghost" 
          size="sm" 
          className="relative"
          aria-label="Notifications"
        >
          <Bell className="w-4 h-4" />
          <span className="absolute top-1 end-1 w-2 h-2 bg-red rounded-full" aria-hidden="true" />
        </Button>

        {/* User menu */}
        <div className="flex items-center gap-2 ps-2 border-s border-border">
          <div className="w-8 h-8 bg-cyan/10 rounded-full flex items-center justify-center">
            <User className="w-4 h-4 text-cyan" />
          </div>
          <div className="hidden sm:block text-small">
            <p className="font-medium text-text">أحمد محمد</p>
            <p className="text-tiny text-text-muted">المدير</p>
          </div>
          <Button 
            variant="ghost" 
            size="sm"
            aria-label="Logout"
          >
            <LogOut className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </header>
  );
}