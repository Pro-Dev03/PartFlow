import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from '../../hooks/useTranslation';
import {
  LayoutDashboard,
  ShoppingCart,
  Package,
  Users,
  DollarSign,
  Truck,
  CreditCard,
  RotateCcw,
  Shield,
  BarChart3,
  Settings,
  ChevronLeft,
  ChevronRight,
  Scan
} from 'lucide-react';
import { cn } from '../../lib/utils';

interface SidebarProps {
  isCollapsed: boolean;
  onToggle: () => void;
}

export function Sidebar({ isCollapsed, onToggle }: SidebarProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const [activeItem, setActiveItem] = useState('dashboard');

  const menuItems = [
    { id: 'dashboard', icon: LayoutDashboard, label: t('nav.dashboard'), path: '/app/dashboard' },
    { id: 'sales', icon: ShoppingCart, label: t('nav.sales'), path: '/app/sales' },
    { id: 'inventory', icon: Package, label: t('nav.inventory'), path: '/app/inventory' },
    { id: 'customers', icon: Users, label: t('nav.customers'), path: '/app/customers' },
    { id: 'debts', icon: DollarSign, label: t('nav.debts'), path: '/app/debts' },
    { id: 'suppliers', icon: Truck, label: t('nav.suppliers'), path: '/app/suppliers' },
    { id: 'purchases', icon: CreditCard, label: t('nav.purchases'), path: '/app/purchases' },
    { id: 'expenses', icon: DollarSign, label: t('nav.expenses'), path: '/app/expenses' },
    { id: 'returns', icon: RotateCcw, label: t('nav.returns'), path: '/app/returns' },
    { id: 'warranties', icon: Shield, label: t('nav.warranties'), path: '/app/warranties' },
    { id: 'reports', icon: BarChart3, label: t('nav.reports'), path: '/app/reports' },
    { id: 'settings', icon: Settings, label: t('nav.settings'), path: '/app/settings' },
  ];

  // Update active item based on current location
  useEffect(() => {
    const currentItem = menuItems.find(item => location.pathname === item.path);
    if (currentItem) {
      setActiveItem(currentItem.id);
    }
  }, [location.pathname]);

  return (
    <aside
      className={cn(
        'bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700',
        'transition-all duration-300 ease-in-out',
        isCollapsed ? 'w-16' : 'w-64',
        'border-s'
      )}
    >
      {/* Logo */}
      <div className="h-16 flex items-center justify-between px-4 border-b border-gray-200 dark:border-gray-700">
        {!isCollapsed && (
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center">
              <Package className="w-5 h-5 text-white" />
            </div>
            <span className="font-bold text-lg text-gray-900 dark:text-gray-100">PartFlow</span>
          </div>
        )}
        <button
          onClick={onToggle}
          className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        >
          {isCollapsed ? (
            <ChevronRight className="w-5 h-5 text-gray-600 dark:text-gray-400" />
          ) : (
            <ChevronLeft className="w-5 h-5 text-gray-600 dark:text-gray-400" />
          )}
        </button>
      </div>

      {/* Navigation */}
      <nav className="p-2 space-y-1 overflow-y-auto flex-1">
        {menuItems.map((item) => {
          const Icon = item.icon;
          const isActive = activeItem === item.id;
          
          return (
            <button
              key={item.id}
              onClick={() => {
                setActiveItem(item.id);
                navigate(item.path);
              }}
              className={cn(
                'w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors',
                'text-sm font-medium',
                isActive
                  ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-400'
                  : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700',
                isCollapsed && 'justify-center'
              )}
              title={isCollapsed ? item.label : undefined}
            >
              <Icon className="w-5 h-5 flex-shrink-0" />
              {!isCollapsed && <span>{item.label}</span>}
            </button>
          );
        })}
      </nav>

      {/* Quick Scan Button */}
      <div className="p-2 border-t border-gray-200 dark:border-gray-700">
        <button
          className={cn(
            'w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors',
            'text-sm font-medium bg-primary-600 text-white hover:bg-primary-700',
            isCollapsed && 'justify-center'
          )}
          title={isCollapsed ? 'Scan' : undefined}
        >
          <Scan className="w-5 h-5 flex-shrink-0" />
          {!isCollapsed && <span>Scan</span>}
        </button>
      </div>
    </aside>
  );
}