import { useState, useRef, useEffect } from 'react';
import { Bell, Search, Moon, Sun, Globe, User, LogOut, ChevronDown } from 'lucide-react';
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
    document.documentElement.classList.toggle('dark');
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
    <header className="h-16 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between px-4">
      {/* Left side */}
      <div className="flex items-center gap-4">
        <button
          onClick={onToggleSidebar}
          className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        
        {/* Search */}
        <div className="relative hidden md:block">
          <Search className="absolute inset-y-0 start-0 flex items-center ps-3 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder={t('common.search')}
            className={cn(
              'w-64 lg:w-96 h-10 ps-10 pe-4 rounded-lg border border-gray-300',
              'bg-gray-50 dark:bg-gray-700 dark:border-gray-600',
              'focus:outline-none focus:ring-2 focus:ring-primary-500',
              'text-sm text-gray-900 dark:text-gray-100 placeholder-gray-400'
            )}
            dir="auto"
          />
        </div>
      </div>

      {/* Right side */}
      <div className="flex items-center gap-2">
        {/* Language selector */}
        <div className="relative" ref={langDropdownRef}>
          <Button 
            variant="ghost" 
            size="sm" 
            className="gap-2"
            onClick={() => setIsLangDropdownOpen(!isLangDropdownOpen)}
          >
            <Globe className="w-4 h-4" />
            <span className="hidden sm:inline">{currentLanguage.toUpperCase()}</span>
            <ChevronDown className="w-4 h-4" />
          </Button>
          
          {isLangDropdownOpen && (
            <div className="absolute top-full end-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-50">
              {languages.map((lang) => (
                <button
                  key={lang.code}
                  onClick={() => {
                    changeLanguage(lang.code);
                    setIsLangDropdownOpen(false);
                  }}
                  className={cn(
                    'w-full px-4 py-2 text-start hover:bg-gray-100 dark:hover:bg-gray-700',
                    'flex items-center gap-3 transition-colors',
                    currentLanguage === lang.code && 'bg-gray-100 dark:bg-gray-700'
                  )}
                >
                  <span className="text-xl">{lang.flag}</span>
                  <div>
                    <p className="font-medium text-sm text-gray-900 dark:text-gray-100">
                      {lang.nativeName}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      {lang.name}
                    </p>
                  </div>
                  {currentLanguage === lang.code && (
                    <span className="me-auto text-primary-600">✓</span>
                  )}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Theme toggle */}
        <Button variant="ghost" size="sm" onClick={toggleTheme}>
          {isDark ? (
            <Sun className="w-4 h-4" />
          ) : (
            <Moon className="w-4 h-4" />
          )}
        </Button>

        {/* Notifications */}
        <Button variant="ghost" size="sm" className="relative">
          <Bell className="w-4 h-4" />
          <span className="absolute top-1 end-1 w-2 h-2 bg-red-500 rounded-full" />
        </Button>

        {/* User menu */}
        <div className="flex items-center gap-2 ps-2 border-s border-gray-200 dark:border-gray-700">
          <div className="w-8 h-8 bg-primary-100 dark:bg-primary-900 rounded-full flex items-center justify-center">
            <User className="w-4 h-4 text-primary-600 dark:text-primary-400" />
          </div>
          <div className="hidden sm:block text-sm">
            <p className="font-medium text-gray-900 dark:text-gray-100">أحمد محمد</p>
            <p className="text-xs text-gray-500 dark:text-gray-400">المدير</p>
          </div>
          <Button variant="ghost" size="sm">
            <LogOut className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </header>
  );
}