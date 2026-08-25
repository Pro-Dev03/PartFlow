import { useState, useEffect } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { X, LayoutDashboard, Package, ShoppingCart, Users, DollarSign, Truck, FileText, Settings, HelpCircle, ChevronRight } from 'lucide-react';
import { cn } from '../../utils';

interface MobileMenuProps {
  isOpen: boolean;
  onClose: () => void;
}

interface NavItem {
  label: string;
  path: string;
  icon: React.ElementType;
  section?: 'workspace' | 'system';
}

const navItems: NavItem[] = [
  { label: 'لوحة التحكم', path: '/app/dashboard', icon: LayoutDashboard, section: 'workspace' },
  { label: 'المخزون', path: '/app/inventory', icon: Package, section: 'workspace' },
  { label: 'نقاط البيع', path: '/app/sales', icon: ShoppingCart, section: 'workspace' },
  { label: 'العملاء', path: '/app/customers', icon: Users, section: 'workspace' },
  { label: 'الديون', path: '/app/debts', icon: DollarSign, section: 'workspace' },
  { label: 'الموردين', path: '/app/suppliers', icon: Truck, section: 'workspace' },
  { label: 'المشتريات', path: '/app/purchases', icon: Package, section: 'workspace' },
  { label: 'المصروفات', path: '/app/expenses', icon: DollarSign, section: 'workspace' },
  { label: 'المرتجعات', path: '/app/returns', icon: ShoppingCart, section: 'workspace' },
  { label: 'التقارير', path: '/app/reports', icon: FileText, section: 'workspace' },
  { label: 'الإعدادات', path: '/app/settings', icon: Settings, section: 'system' },
];

export function MobileMenu({ isOpen, onClose }: MobileMenuProps) {
  const location = useLocation();

  // Close menu on route change
  useEffect(() => {
    if (isOpen) {
      onClose();
    }
  }, [location.pathname, isOpen, onClose]);

  // Prevent body scroll when menu is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  const workspaceItems = navItems.filter(item => item.section === 'workspace');
  const systemItems = navItems.filter(item => item.section === 'system');

  if (!isOpen) return null;

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 backdrop-blur-sm z-40 md:hidden"
        onClick={onClose}
        style={{ animation: 'fadeIn 0.2s ease-out' }}
      />

      {/* Menu Panel */}
      <div
        className={cn(
          'fixed inset-y-0 right-0 w-80 bg-surface border-l border-border z-50 md:hidden mobile-menu',
          'transform transition-transform duration-300 ease-in-out'
        )}
        style={{
          background: 'linear-gradient(180deg, rgba(12, 17, 28, 0.98), rgba(7, 10, 18, 0.96))',
          backdropFilter: 'blur(20px)'
        }}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <div className="flex items-center gap-3">
            <div style={{
              width: '32px',
              height: '32px',
              borderRadius: '8px',
              display: 'grid',
              placeItems: 'center',
              color: 'var(--color-primary)',
              border: '1px solid rgba(34, 211, 238, 0.35)',
              background: 'linear-gradient(135deg, rgba(34, 211, 238, 0.12), rgba(59, 130, 246, 0.04))'
            }}>
              <span style={{ fontSize: '16px' }}>◈</span>
            </div>
            <div>
              <div style={{ color: 'var(--text-primary)', fontWeight: '700', fontSize: '14px' }}>PARTFLOW</div>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-surface-elevated rounded-lg transition-colors"
            aria-label="إغلاق القائمة"
          >
            <X className="w-5 h-5 text-text-secondary" />
          </button>
        </div>

        {/* Navigation */}
        <div className="pb-20 overflow-y-auto">
          {/* Workspace Section */}
          <div className="p-4">
            <div className="text-xs font-medium text-text-muted uppercase tracking-wider mb-3 px-2">
              مساحة العمل
            </div>
            <nav className="space-y-1">
              {workspaceItems.map((item) => {
                const Icon = item.icon;
                const isActive = location.pathname === item.path;
                return (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={cn(
                      'flex items-center gap-3 px-3 py-3 rounded-lg transition-all duration-200',
                      'hover:bg-surface-elevated',
                      isActive && 'bg-[var(--color-accent)]/10 border border-[var(--color-accent)]/20'
                    )}
                    style={{
                      color: isActive ? '#14b8a6' : '#94a3b8'
                    }}
                  >
                    <Icon className="w-5 h-5" style={{ color: isActive ? '#14b8a6' : '#65748c' }} />
                    <span className="flex-1 text-sm font-medium">{item.label}</span>
                    {isActive && <ChevronRight className="w-4 h-4 text-cyan" />}
                  </Link>
                );
              })}
            </nav>
          </div>

          {/* System Section */}
          <div className="p-4 border-t border-border">
            <div className="text-xs font-medium text-text-muted uppercase tracking-wider mb-3 px-2">
              النظام
            </div>
            <nav className="space-y-1">
              {systemItems.map((item) => {
                const Icon = item.icon;
                const isActive = location.pathname === item.path;
                return (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={cn(
                      'flex items-center gap-3 px-3 py-3 rounded-lg transition-all duration-200',
                      'hover:bg-surface-elevated',
                      isActive && 'bg-[var(--color-accent)]/10 border border-[var(--color-accent)]/20'
                    )}
                    style={{
                      color: isActive ? '#14b8a6' : '#94a3b8'
                    }}
                  >
                    <Icon className="w-5 h-5" style={{ color: isActive ? '#14b8a6' : '#65748c' }} />
                    <span className="flex-1 text-sm font-medium">{item.label}</span>
                    {isActive && <ChevronRight className="w-4 h-4 text-cyan" />}
                  </Link>
                );
              })}
            </nav>
          </div>
        </div>
      </div>

      <style>{`
        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
      `}</style>
    </>
  );
}