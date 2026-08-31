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
  Settings,
  Scan,
  BarChart3,
  Layers,
  Tag,
  FileText,
  UserCheck,
} from 'lucide-react';
import { cn } from '../../utils';
import type { LucideIcon } from 'lucide-react';

interface SidebarProps {
  isCollapsed: boolean;
}

interface MenuItem {
  id: string;
  icon: LucideIcon;
  label: string;
  path: string;
}

interface MenuGroup {
  title: string;
  items: MenuItem[];
}

export function Sidebar({ isCollapsed }: SidebarProps) {
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
      title: 'البيع',
      items: [
        { id: 'sales', icon: ShoppingCart, label: t('nav.pos') || 'نقطة البيع', path: '/app/sales' },
        { id: 'customers', icon: Users, label: t('nav.customers') || 'العملاء', path: '/app/customers' },
        { id: 'debts', icon: DollarSign, label: t('nav.debts') || 'الديون', path: '/app/debts' },
      ]
    },
    {
      title: 'المخزون',
      items: [
        { id: 'inventory', icon: Package, label: t('nav.inventory') || 'المنتجات', path: '/app/inventory' },
        { id: 'used-parts', icon: Layers, label: 'مخزون القطع المستعملة', path: '/app/usedparts' },
        { id: 'seller-balances', icon: UserCheck, label: 'رصيد البائعين', path: '/app/seller-balances' },
      ]
    },
    {
      title: 'المشتريات',
      items: [
        { id: 'purchases', icon: CreditCard, label: t('nav.purchases') || 'المشتريات', path: '/app/purchases' },
        { id: 'suppliers', icon: Truck, label: t('nav.suppliers') || 'الموردون', path: '/app/suppliers' },
      ]
    },
    {
      title: 'المال',
      items: [
        { id: 'expenses', icon: DollarSign, label: t('nav.expenses') || 'المصروفات', path: '/app/expenses' },
        { id: 'returns', icon: RotateCcw, label: t('nav.returns') || 'المرتجعات', path: '/app/returns' },
        { id: 'supplier-returns', icon: RotateCcw, label: 'مرتجعات الموردين', path: '/app/supplier-returns' },
        { id: 'return-details', icon: FileText, label: 'تفاصيل المرتجعات', path: '/app/return-details' },
        { id: 'reports', icon: BarChart3, label: t('nav.reports') || 'التقارير', path: '/app/reports' },
      ]
    },
    {
      title: t('nav.system') || 'النظام',
      items: [
        { id: 'categories', icon: Tag, label: 'التصنيفات', path: '/app/categories' },
        { id: 'part-types', icon: FileText, label: 'أنواع القطع', path: '/app/part-types' },
        { id: 'settings', icon: Settings, label: t('nav.settings') || 'الإعدادات', path: '/app/settings' },
      ]
    }
  ];

  // Flatten all items for active state checking
  const allItems = menuGroups.flatMap(group => group.items);

  // Update active item based on current location
  useEffect(() => {
    const currentItem = allItems.find(item => {
      // Match exact path or path with params
      return location.pathname === item.path || 
             location.pathname.startsWith(item.path + '/');
    });
    if (currentItem) {
      setActiveItem(currentItem.id);
    }
  }, [location.pathname, allItems]);

  return (
    <aside
      className={cn(
        'flex flex-col sidebar',
        'border-l border-[var(--border-default)]',
        'backdrop-blur-xl',
        'transition-all duration-[var(--transition-normal)] ease-[var(--ease-out)]',
        'hover:shadow-lg',
        'shrink-0',
        'relative',
        isCollapsed ? 'w-[72px]' : 'w-[260px]'
      )}
    >
      {/* Logo */}
      <div className="flex items-center justify-center border-b border-[var(--border-default)] bg-[var(--bg-surface-elevated)] px-[var(--spacing-4)] py-[var(--spacing-5)]">
        {!isCollapsed && (
          <div className="flex items-center gap-[var(--spacing-3)]">
            <div
              className="flex h-10 w-10 shrink-0 items-center justify-center rounded-[var(--radius-lg)] border shadow-[var(--shadow-glow)]"
              style={{
                background: 'var(--color-primary-08)',
                borderColor: 'var(--color-primary-25)',
                boxShadow: '0 4px 14px var(--color-primary-15)',
                width: '40px',
                minWidth: '40px',
                height: '40px',
                minHeight: '40px',
              }}
            >
              <img
                src="./favicon.svg?v=3"
                alt=""
                aria-hidden="true"
                className="block h-5 w-5 shrink-0"
                style={{ width: '20px', height: '20px' }}
              />
            </div>
            <div className="brand-text">
              <span className="text-lg font-bold tracking-[0.5px] text-[var(--text-primary)]">PARTFLOW</span>
              <p className="brand-sub text-xs tracking-[0.3px] text-[var(--text-secondary)]">Store Operating System</p>
            </div>
          </div>
        )}
        {isCollapsed && (
          <div
            className="flex h-10 w-10 shrink-0 items-center justify-center rounded-[var(--radius-lg)] border shadow-[var(--shadow-glow)]"
            style={{
              background: 'var(--color-primary-08)',
              borderColor: 'var(--color-primary-25)',
              boxShadow: '0 4px 14px var(--color-primary-15)',
              width: '40px',
              minWidth: '40px',
              height: '40px',
              minHeight: '40px',
            }}
          >
            <img
              src="/favicon.svg?v=3"
              alt=""
              aria-hidden="true"
              className="block h-5 w-5 shrink-0"
              style={{ width: '20px', height: '20px' }}
            />
          </div>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-[var(--spacing-3)] py-[var(--spacing-4)] space-y-[var(--spacing-4)] overflow-hidden">
        {menuGroups.map((group) => (
          <div key={group.title}>
            {!isCollapsed && (
              <div
                className="mb-[var(--spacing-2)] px-[var(--spacing-3)] py-[var(--spacing-2)] text-xs font-semibold uppercase tracking-wider text-[var(--text-tertiary)]"
              >
                {group.title}
              </div>
            )}
            <div className="space-y-[var(--spacing-1)]">
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
                      'sidebar-item w-full flex items-center gap-[var(--spacing-3)] px-[var(--spacing-3)] py-[var(--spacing-2)] rounded-[var(--radius-md)]',
                      'text-sm font-medium relative overflow-hidden',
                      isCollapsed && 'justify-center',
                      isActive && 'active'
                    )}
                    title={isCollapsed ? item.label : undefined}
                  >
                    <Icon
                      className={cn(
                        'flex-shrink-0',
                        'w-5 h-5'
                      )}
                    />
                    {!isCollapsed && (
                      <span className="truncate nav-label">{item.label}</span>
                    )}
                  </button>
                );
              })}
            </div>
          </div>
        ))}
      </nav>

      {/* Quick Scan Button */}
      <div className="border-t border-[var(--border-default)] bg-[var(--bg-surface-elevated)] p-[var(--spacing-4)]">
        <button
          className={cn(
            'sidebar-scan-button w-full flex items-center gap-[var(--spacing-3)] rounded-[var(--radius-lg)] border border-[var(--border-default)] px-[var(--spacing-4)] py-[var(--spacing-3)] text-sm font-medium shadow-[var(--shadow-glow)] transition-all duration-[var(--transition-normal)]',
            isCollapsed && 'justify-center'
          )}
          title={isCollapsed ? 'مسح الباركود' : undefined}
          onMouseEnter={(e) => {
            e.currentTarget.classList.add('sidebar-scan-button-hover');
          }}
          onMouseLeave={(e) => {
            e.currentTarget.classList.remove('sidebar-scan-button-hover');
          }}
        >
          <Scan className="h-5 w-5 flex-shrink-0" style={{ color: '#ffffff' }} />
          {!isCollapsed && (
            <span className="nav-label text-[var(--text-on-primary)]">مسح الباركود</span>
          )}
        </button>
      </div>
    </aside>
  );
}