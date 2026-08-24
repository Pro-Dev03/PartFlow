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
        { id: 'customers', icon: Users, label: t('nav.customers') || 'العملاء', path: '/app/customers' },
        { id: 'debts', icon: DollarSign, label: t('nav.debts') || 'الديون', path: '/app/debts' },
      ]
    },
    {
      title: t('nav.management') || 'الإدارة',
      items: [
        { id: 'suppliers', icon: Truck, label: t('nav.suppliers') || 'الموردون', path: '/app/suppliers' },
        { id: 'purchases', icon: CreditCard, label: t('nav.purchases') || 'المشتريات', path: '/app/purchases' },
        { id: 'expenses', icon: DollarSign, label: t('nav.expenses') || 'المصروفات', path: '/app/expenses' },
        { id: 'returns', icon: RotateCcw, label: t('nav.returns') || 'المرتجعات', path: '/app/returns' },
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
      // Handle both /app/path and /path formats
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
        'border-r border-[var(--border-default)]',
        'backdrop-blur-xl',
        'transition-all duration-[var(--transition-normal)] ease-[var(--transition-timing-default)]',
        'hover:shadow-lg',
        'shrink-0',
        'relative',
        isCollapsed ? 'w-[var(--sidebar-width-collapsed)]' : 'w-[var(--sidebar-width-expanded)]'
      )}
      style={{
        background: 'linear-gradient(180deg, rgba(17, 24, 39, 0.95) 0%, rgba(17, 24, 39, 0.85) 100%)',
        boxShadow: '4px 0 24px rgba(0, 0, 0, 0.15)'
      }}
    >
      {/* Logo */}
      <div className="flex items-center justify-center px-[var(--spacing-md)] py-[var(--spacing-lg)] border-b border-[var(--border-default)]" style={{
        background: 'linear-gradient(90deg, rgba(99, 102, 241, 0.05) 0%, transparent 100%)'
      }}>
        {!isCollapsed && (
          <div className="flex items-center gap-[var(--spacing-sm)]">
            <div
              className="w-10 h-10 rounded-[var(--radius-md)] flex items-center justify-center"
              style={{
                background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.2) 0%, rgba(139, 92, 246, 0.2) 100%)',
                border: '1px solid rgba(99, 102, 241, 0.3)',
                boxShadow: '0 4px 12px rgba(99, 102, 241, 0.15)'
              }}
            >
              <Package className="w-5 h-5" style={{ color: '#818cf8' }} />
            </div>
            <div className="brand-text">
              <span className="font-bold text-lg" style={{ 
                color: '#fff',
                letterSpacing: '0.5px',
                textShadow: '0 2px 4px rgba(0, 0, 0, 0.1)'
              }}>PARTFLOW</span>
              <p className="text-xs brand-sub" style={{ 
                color: 'rgba(148, 163, 184, 0.8)',
                letterSpacing: '0.3px'
              }}>Store Operating System</p>
            </div>
          </div>
        )}
        {isCollapsed && (
          <div
            className="w-10 h-10 rounded-[var(--radius-md)] flex items-center justify-center"
            style={{
              background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.2) 0%, rgba(139, 92, 246, 0.2) 100%)',
              border: '1px solid rgba(99, 102, 241, 0.3)',
              boxShadow: '0 4px 12px rgba(99, 102, 241, 0.15)'
            }}
          >
            <Package className="w-5 h-5" style={{ color: '#818cf8' }} />
          </div>
        )}
      </div>

        {/* Navigation */}
        <nav className="flex-1 px-[var(--spacing-sm)] py-[var(--spacing-md)] space-y-[var(--spacing-lg)] overflow-hidden">
          {menuGroups.map((group) => (
            <div key={group.title}>
              {!isCollapsed && (
                <div
                  className="px-[var(--spacing-sm)] py-[var(--spacing-xs)] text-xs font-semibold uppercase tracking-wider mb-[var(--spacing-sm)]"
                  style={{ color: 'var(--text-tertiary)' }}
                >
                  {group.title}
                </div>
              )}
              <div className="space-y-[var(--spacing-xs)]">
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
                        'w-full flex items-center gap-[var(--spacing-sm)] px-[var(--spacing-sm)] py-[var(--spacing-sm)] rounded-[var(--radius-sm)] transition-all duration-[var(--transition-normal)]',
                        'text-sm font-medium relative overflow-hidden',
                        isCollapsed && 'justify-center'
                      )}
                      style={{
                        background: isActive
                          ? 'linear-gradient(135deg, rgba(99, 102, 241, 0.25) 0%, rgba(139, 92, 246, 0.25) 100%)'
                          : 'transparent',
                        color: isActive ? '#fff' : 'rgba(148, 163, 184, 0.8)',
                        border: isActive ? '1px solid rgba(99, 102, 241, 0.4)' : '1px solid transparent',
                        boxShadow: isActive ? '0 4px 12px rgba(99, 102, 241, 0.2)' : 'none'
                      }}
                      title={isCollapsed ? item.label : undefined}
                      onMouseEnter={(e) => {
                        if (!isActive) {
                          e.currentTarget.style.background = 'rgba(99, 102, 241, 0.1)';
                          e.currentTarget.style.color = '#818cf8';
                          e.currentTarget.style.transform = 'translateX(-4px)';
                        }
                      }}
                      onMouseLeave={(e) => {
                        if (!isActive) {
                          e.currentTarget.style.background = 'transparent';
                          e.currentTarget.style.color = 'rgba(148, 163, 184, 0.8)';
                          e.currentTarget.style.transform = 'translateX(0)';
                        }
                      }}
                    >
                      <Icon
                        className={cn(
                          'flex-shrink-0',
                          'w-5 h-5' // worktrack: 20px
                        )}
                        style={{
                          color: isActive
                            ? '#818cf8'
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
        <div className="p-[var(--spacing-md)] border-t border-[var(--border-default)]" style={{
          background: 'linear-gradient(90deg, rgba(99, 102, 241, 0.05) 0%, transparent 100%)'
        }}>
          <button
            className={cn(
              'w-full flex items-center gap-[var(--spacing-sm)] px-[var(--spacing-md)] py-[var(--spacing-sm)] rounded-[var(--radius-md)] transition-all duration-[var(--transition-normal)]',
              'text-sm font-medium text-white relative overflow-hidden',
              isCollapsed && 'justify-center'
            )}
            style={{
              background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.3) 0%, rgba(139, 92, 246, 0.3) 100%)',
              border: '1px solid rgba(99, 102, 241, 0.4)',
              boxShadow: '0 4px 12px rgba(99, 102, 241, 0.2)'
            }}
            title={isCollapsed ? 'مسح الباركود' : undefined}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'linear-gradient(135deg, rgba(99, 102, 241, 0.4) 0%, rgba(139, 92, 246, 0.4) 100%)';
              e.currentTarget.style.boxShadow = '0 6px 16px rgba(99, 102, 241, 0.3)';
              e.currentTarget.style.transform = 'translateY(-2px)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'linear-gradient(135deg, rgba(99, 102, 241, 0.3) 0%, rgba(139, 92, 246, 0.3) 100%)';
              e.currentTarget.style.boxShadow = '0 4px 12px rgba(99, 102, 241, 0.2)';
              e.currentTarget.style.transform = 'translateY(0)';
            }}
          >
            <Scan className="w-5 h-5 flex-shrink-0" style={{ color: '#818cf8' }} />
            {!isCollapsed && (
              <span className="nav-label" style={{ color: '#fff' }}>مسح الباركود</span>
            )}
          </button>
        </div>
      </aside>
  );
}