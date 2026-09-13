import { PartFlowLogo } from '../../../components/branding/PartFlowLogo';

import { ReactNode } from 'react';
import { cn } from '../../../../utils';

interface ModernPOSLayoutProps {
  header: ReactNode;
  productsSection: ReactNode;
  cartSection: ReactNode;
  footer: ReactNode;
}

export function ModernPOSLayout({
  header,
  productsSection,
  cartSection,
  footer,
}: ModernPOSLayoutProps) {
  return (
    <div className="pos-modern-layout">
      {/* Header */}
      <header className="pos-modern-header">
        {header}
      </header>

      {/* Main Content */}
      <div className="pos-modern-body">
        {/* Products Section */}
        <main className="pos-modern-products">
          {productsSection}
        </main>

        {/* Cart Sidebar */}
        <aside className="pos-modern-cart">
          {cartSection}
        </aside>
      </div>

      {/* Footer Actions */}
      <footer className="pos-modern-footer">
        {footer}
      </footer>
    </div>
  );
}

interface ModernHeaderProps {
  title: string;
  subtitle?: string;
  cashierName?: string;
  actions?: ReactNode;
}

export function ModernHeader({
  title,
  subtitle,
  cashierName = 'الكاشير 01',
  actions,
}: ModernHeaderProps) {
  return (
    <div className="pos-modern-header-content">
      <div className="pos-modern-header-left">
        <div className="pos-modern-logo">
          <div className="pos-modern-logo-icon">
            <PartFlowLogo size={36} priority />
          </div>
        </div>
        <div className="pos-modern-header-titles">
          <h1 className="pos-modern-title">{title}</h1>
          {subtitle && <p className="pos-modern-subtitle">{subtitle}</p>}
        </div>
      </div>
      <div className="pos-modern-header-right">
        <div className="pos-modern-cashier-info">
          <span className="pos-modern-cashier-badge">{cashierName}</span>
          <span className="pos-modern-time">
            {new Intl.DateTimeFormat('ar', {
              hour: '2-digit',
              minute: '2-digit',
            }).format(new Date())}
          </span>
        </div>
        {actions && <div className="pos-modern-header-actions">{actions}</div>}
      </div>
    </div>
  );
}

interface ModernSectionProps {
  title: string;
  badge?: string | number;
  icon?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
}

export function ModernSection({
  title,
  badge,
  icon,
  actions,
  children,
  className,
}: ModernSectionProps) {
  return (
    <section className={cn('pos-modern-section', className)}>
      <div className="pos-modern-section-header">
        <div className="pos-modern-section-title-group">
          {icon && <div className="pos-modern-section-icon">{icon}</div>}
            <span className="pos-modern-section-title">{title}</span>
            {badge !== undefined && (
              <span className="pos-modern-section-badge">{badge}</span>
            )}
        </div>
        {actions && <div className="pos-modern-section-actions">{actions}</div>}
      </div>
      <div className="pos-modern-section-body">{children}</div>
    </section>
  );
}
