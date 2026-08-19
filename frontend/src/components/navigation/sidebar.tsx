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
  Settings,
  ChevronLeft,
  ChevronRight,
  Scan,
  Heading,
  BarChart3,
  Palette
} from 'lucide-react';
import { cn } from '../../lib/utils';

interface SidebarProps {
  isCollapsed: boolean;
  onToggle: () => void;
}

interface MenuItem {
  id: string;
  icon: any;
  label: string;
  path: string;
}

interface MenuGroup {
  title: string;
  items: MenuItem[];
}

export function Sidebar({ isCollapsed, onToggle }: SidebarProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const [activeItem, setActiveItem] = useState('dashboard');

  const menuGroups: MenuGroup[] = [
    {
      title: t('nav.main') || 'الرئيسية',
      items: [
        { id: 'dashboard', icon: LayoutDashboard, label: t('nav.dashboard') || 'لوحة التحكم', path: '/app/dashboard' },
      ]
    },
    {
      title: t('nav.sales') || 'المبيعات',
      items: [
        { id: 'pos', icon: ShoppingCart, label: t('nav.pos') || 'نقطة البيع', path: '/app/sales' },
        { id: 'sales', icon: BarChart3, label: t('nav.sales') || 'المبيعات', path: '/app/sales' },
        { id: 'returns', icon: RotateCcw, label: t('nav.returns') || 'المرتجعات', path: '/app/returns' },
      ]
    },
    {
      title: t('nav.inventory') || 'المخزون',
      items: [
        { id: 'inventory', icon: Package, label: t('nav.inventory') || 'المخزون', path: '/app/inventory' },
        { id: 'categories', icon: Heading, label: t('nav.categories') || 'الفئات', path: '/app/inventory' },
      ]
    },
    {
      title: t('nav.customers') || 'العملاء',
      items: [
        { id: 'customers', icon: Users, label: t('nav.customers') || 'العملاء', path: '/app/customers' },
        { id: 'debts', icon: DollarSign, label: t('nav.debts') || 'الديون', path: '/app/debts' },
      ]
    },
    {
      title: t('nav.purchasing') || 'المشتريات',
      items: [
        { id: 'purchases', icon: CreditCard, label: t('nav.purchases') || 'المشتريات', path: '/app/purchases' },
        { id: 'suppliers', icon: Truck, label: t('nav.suppliers') || 'الموردين', path: '/app/suppliers' },
      ]
    },
    {
      title: t('nav.finance') || 'المالية',
      items: [
        { id: 'expenses', icon: DollarSign, label: t('nav.expenses') || 'المصروفات', path: '/app/expenses' },
        { id: 'reports', icon: BarChart3, label: t('nav.reports') || 'التقارير', path: '/app/reports' },
      ]
    },
    {
      title: t('nav.system') || 'النظام',
      items: [
        { id: 'warranties', icon: Shield, label: t('nav.warranties') || 'الضمانات', path: '/app/warranties' },
        { id: 'settings', icon: Settings, label: t('nav.settings') || 'الإعدادات', path: '/app/settings' },
        { id: 'design-system', icon: Palette, label: 'Design System', path: '/app/design-system' },
      ]
    }
  ];

  // Flatten all items for active state checking
  const allItems = menuGroups.flatMap(group => group.items);

  // Update active item based on current location
  useEffect(() => {
    const currentItem = allItems.find(item => location.pathname === item.path);
    if (currentItem) {
      setActiveItem(currentItem.id);
    }
  }, [location.pathname, allItems]);

  return (
    <aside
      className={cn(
        'border-r border-border bg-sidebar-gradient backdrop-blur-lg',
        'transition-all duration-300 ease-in-out',
        isCollapsed ? 'w-16' : 'w-64',
        'flex flex-col'
      )}
    >
      {/* Logo */}
      <div className="flex items-center justify-between px-xl py-6 border-b border-border">
        {!isCollapsed && (
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-sm flex items-center justify-center text-cyan border border-cyan/35 bg-logo-gradient shadow-glow-strong">
              <Package className="w-5 h-5" />
            </div>
            <div>
              <span className="font-extrabold text-lg text-text tracking-wide">PARTFLOW</span>
              <small className="block text-tiny text-text-muted tracking-widest uppercase">Store Operating System</small>
            </div>
          </div>
        )}
        <button
          onClick={onToggle}
          className="p-2 rounded-sm hover:bg-surface/50 transition-all duration-normal"
        >
          {isCollapsed ? (
            <ChevronRight className="w-5 h-5 text-text-muted" />
          ) : (
            <ChevronLeft className="w-5 h-5 text-text-muted" />
          )}
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto px-md py-0">
        {menuGroups.map((group) => (
          <div key={group.title} className="mb-0">
            {!isCollapsed && (
              <div className="px-sm py-2 text-tiny font-semibold text-text-muted uppercase tracking-wider">
                {group.title}
              </div>
            )}
            <div className="grid gap-1">
              {group.items.map((item) => {
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
                      'flex items-center gap-3 px-sm py-3 rounded-sm transition-all duration-normal',
                      'text-body font-medium',
                      isActive
                        ? 'text-cyan-100 bg-nav-active-gradient border border-cyan/14 shadow-nav-active'
                        : 'text-text-muted hover:bg-surface/6 hover:text-text hover:translate-x-0.5',
                      isCollapsed && 'justify-center'
                    )}
                    title={isCollapsed ? item.label : undefined}
                  >
                    <Icon className={cn('w-4.5 h-4.5 flex-shrink-0 text-center', isActive ? 'text-cyan' : 'text-text-muted')} />
                    {!isCollapsed && <span>{item.label}</span>}
                  </button>
                );
              })}
            </div>
          </div>
        ))}
      </nav>

      {/* Quick Scan Button */}
      <div className="p-md border-t border-border">
        <button
          className={cn(
            'flex items-center gap-3 px-sm py-3 rounded-sm transition-all duration-normal',
            'text-body font-medium border border-cyan/35 bg-button-primary-gradient text-cyan-100 hover:-translate-y-px hover:border-cyan/30 shadow-glow',
            isCollapsed && 'justify-center'
          )}
          title={isCollapsed ? 'Scan' : undefined}
        >
          <Scan className="w-4.5 h-4.5 flex-shrink-0" />
          {!isCollapsed && <span>مسح الباركود</span>}
        </button>
      </div>
    </aside>
  );
}