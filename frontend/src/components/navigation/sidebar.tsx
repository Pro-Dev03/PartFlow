import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from '../../hooks/useTranslation';
import { useUIStore } from '../../stores/uiStore';
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
  CheckCircle,
  Clock,
  AlertTriangle,
} from 'lucide-react';
import { cn } from '../../utils';

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
  const { theme } = useUIStore();
  const [activeItem, setActiveItem] = useState('dashboard');

  const menuGroups: MenuGroup[] = [
    {
      title: t('nav.main') || 'القائمة الرئيسية',
      items: [
        { id: 'dashboard', icon: LayoutDashboard, label: t('nav.dashboard') || 'لوحة التحكم', path: '/app/dashboard' },
      ]
    },
    {
      title: t('nav.operations') || 'العمليات اليومية',
      items: [
        { id: 'sales', icon: ShoppingCart, label: t('nav.pos') || 'نقطة البيع', path: '/app/sales' },
        { id: 'inventory', icon: Package, label: t('nav.inventory') || 'المخزون', path: '/app/inventory' },
        { id: 'used-parts', icon: Layers, label: 'القطع المستعملة', path: '/app/usedparts' },
        { id: 'inspections', icon: CheckCircle, label: 'الفحص', path: '/app/inspections' },
        { id: 'customers', icon: Users, label: t('nav.customers') || 'العملاء', path: '/app/customers' },
        { id: 'debts', icon: DollarSign, label: t('nav.debts') || 'الديون', path: '/app/debts' },
      ]
    },
    {
      title: 'القطع المستعملة',
      items: [
        { id: 'item-history', icon: Clock, label: 'تاريخ القطع', path: '/app/item-history' },
        { id: 'aging', icon: AlertTriangle, label: 'تقادم القطع', path: '/app/aging' },
        { id: 'seller-balances', icon: DollarSign, label: 'رصيد البائعين', path: '/app/seller-balances' },
      ]
    },
    {
      title: t('nav.management') || 'الإدارة',
      items: [
        { id: 'suppliers', icon: Truck, label: t('nav.suppliers') || 'الموردون', path: '/app/suppliers' },
        { id: 'purchases', icon: CreditCard, label: t('nav.purchases') || 'المشتريات', path: '/app/purchases' },
        { id: 'expenses', icon: DollarSign, label: t('nav.expenses') || 'المصروفات', path: '/app/expenses' },
        { id: 'returns', icon: RotateCcw, label: t('nav.returns') || 'المرتجعات', path: '/app/returns' },
        { id: 'categories', icon: Tag, label: 'التصنيفات', path: '/app/categories' },
      ]
    },
    {
      title: t('nav.system') || 'النظام',
      items: [
        { id: 'reports', icon: BarChart3, label: t('nav.reports') || 'التقارير', path: '/app/reports' },
        { id: 'settings', icon: Settings, label: t('nav.settings') || 'الإعدادات', path: '/app/settings' },
      ]
    }
  ];

  // Flatten all items for active state checking
  const allItems = menuGroups.flatMap(group => group.items);

  // Update active item based on current location
  useEffect(() => {
    const currentItem = allItems.find(item => {
      const itemPath = item.path.replace('/app', '');
      const currentPath = location.pathname.replace('/app', '');
      return currentPath === itemPath || location.pathname === item.path;
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
      style={{
        background: 'var(--bg-surface)',
        boxShadow: 'var(--shadow-lg)'
      }}
    >
      {/* Logo */}
      <div className="flex items-center justify-center px-[var(--spacing-4)] py-[var(--spacing-5)] border-b border-[var(--border-default)]" style={{
        background: 'var(--bg-surface-elevated)'
      }}>
        {!isCollapsed && (
          <div className="flex items-center gap-[var(--spacing-3)]">
            <div
              className="w-10 h-10 rounded-[var(--radius-lg)] flex items-center justify-center"
              style={{
                background: 'var(--gradient-primary)',
                border: '1px solid var(--border-default)',
                boxShadow: 'var(--shadow-glow)'
              }}
            >
              <Package className="w-5 h-5" style={{ color: 'var(--text-on-primary)' }} />
            </div>
            <div className="brand-text">
              <span className="font-bold text-lg" style={{ 
                color: 'var(--text-primary)',
                letterSpacing: '0.5px'
              }}>PARTFLOW</span>
              <p className="text-xs brand-sub" style={{ 
                color: 'var(--text-secondary)',
                letterSpacing: '0.3px'
              }}>Store Operating System</p>
            </div>
          </div>
        )}
        {isCollapsed && (
          <div
            className="w-10 h-10 rounded-[var(--radius-lg)] flex items-center justify-center"
            style={{
              background: 'var(--gradient-primary)',
              border: '1px solid var(--border-default)',
              boxShadow: 'var(--shadow-glow)'
            }}
          >
            <Package className="w-5 h-5" style={{ color: 'var(--text-on-primary)' }} />
          </div>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-[var(--spacing-3)] py-[var(--spacing-4)] space-y-[var(--spacing-4)] overflow-hidden">
        {menuGroups.map((group) => (
          <div key={group.title}>
            {!isCollapsed && (
              <div
                className="px-[var(--spacing-3)] py-[var(--spacing-2)] text-xs font-semibold uppercase tracking-wider mb-[var(--spacing-2)]"
                style={{ color: 'var(--text-tertiary)' }}
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
                      'w-full flex items-center gap-[var(--spacing-3)] px-[var(--spacing-3)] py-[var(--spacing-2)] rounded-[var(--radius-md)] transition-all duration-[var(--transition-normal)]',
                      'text-sm font-medium relative overflow-hidden',
                      isCollapsed && 'justify-center'
                    )}
                    style={{
                      background: isActive
                        ? 'var(--color-primary-15)'
                        : 'transparent',
                      color: isActive ? 'var(--color-primary)' : 'var(--text-secondary)',
                      border: isActive ? '1px solid var(--color-primary-25)' : '1px solid transparent',
                      boxShadow: isActive ? 'var(--shadow-glow-soft)' : 'none'
                    }}
                    title={isCollapsed ? item.label : undefined}
                    onMouseEnter={(e) => {
                      if (!isActive) {
                        e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                        e.currentTarget.style.color = 'var(--color-primary)';
                        e.currentTarget.style.transform = 'translateX(-4px)';
                      }
                    }}
                    onMouseLeave={(e) => {
                      if (!isActive) {
                        e.currentTarget.style.background = 'transparent';
                        e.currentTarget.style.color = 'var(--text-secondary)';
                        e.currentTarget.style.transform = 'translateX(0)';
                      }
                    }}
                  >
                    <Icon
                      className={cn(
                        'flex-shrink-0',
                        'w-5 h-5'
                      )}
                      style={{
                        color: isActive
                          ? 'var(--color-primary)'
                          : 'inherit'
                      }}
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
      <div className="p-[var(--spacing-4)] border-t border-[var(--border-default)]" style={{
        background: 'var(--bg-surface-elevated)'
      }}>
        <button
          className={cn(
            'w-full flex items-center gap-[var(--spacing-3)] px-[var(--spacing-4)] py-[var(--spacing-3)] rounded-[var(--radius-lg)] transition-all duration-[var(--transition-normal)]',
            'text-sm font-medium relative overflow-hidden',
            isCollapsed && 'justify-center'
          )}
          style={{
            background: 'var(--gradient-primary)',
            border: '1px solid var(--border-default)',
            boxShadow: 'var(--shadow-glow)',
            color: 'var(--text-on-primary)'
          }}
          title={isCollapsed ? 'مسح الباركود' : undefined}
          onMouseEnter={(e) => {
            e.currentTarget.style.transform = 'translateY(-1px)';
            e.currentTarget.style.boxShadow = 'var(--shadow-glow-strong)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.transform = 'translateY(0)';
            e.currentTarget.style.boxShadow = 'var(--shadow-glow)';
          }}
        >
          <Scan className="w-5 h-5 flex-shrink-0" style={{ color: 'var(--text-on-primary)' }} />
          {!isCollapsed && (
            <span className="nav-label" style={{ color: 'var(--text-on-primary)' }}>مسح الباركود</span>
          )}
        </button>
      </div>
    </aside>
  );
}