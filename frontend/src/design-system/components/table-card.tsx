import { cn } from '../../utils';

interface TableCardProps {
  children: React.ReactNode;
  className?: string;
}

interface TableCardItemProps {
  label: string;
  value: React.ReactNode;
  className?: string;
  variant?: 'default' | 'highlight' | 'success' | 'warning' | 'danger';
}

export function TableCard({ children, className }: TableCardProps) {
  return (
    <div
      className={cn(
        'rounded-[18px] border border-[rgba(148,163,184,0.2)] bg-[var(--card-bg)] p-4 shadow-[0_12px_24px_rgba(15,23,42,0.06)]',
        'transition-all duration-200 hover:-translate-y-0.5 hover:border-[var(--primary)]/30 hover:shadow-[0_18px_28px_rgba(59,130,246,0.08)]',
        className
      )}
    >
      {children}
    </div>
  );
}

export function TableCardItem({ label, value, className, variant = 'default' }: TableCardItemProps) {
  const variantStyles = {
    default: 'text-[var(--text-primary)]',
    highlight: 'text-[var(--primary)] font-semibold',
    success: 'text-[var(--success)] font-semibold',
    warning: 'text-[var(--warning)] font-semibold',
    danger: 'text-[var(--danger)] font-semibold',
  };

  return (
    <div className={cn('flex items-start justify-between gap-3 border-b border-[rgba(148,163,184,0.14)] py-2.5 last:border-0', className)}>
      <span className="text-[11px] font-bold tracking-[0.08em] text-[var(--text-secondary)] uppercase">{label}</span>
      <span className={cn('text-sm font-medium text-left', variantStyles[variant])}>{value}</span>
    </div>
  );
}

interface TableCardActionsProps {
  children: React.ReactNode;
  className?: string;
}

export function TableCardActions({ children, className }: TableCardActionsProps) {
  return (
    <div className={cn('mt-3 flex gap-2 border-t border-[rgba(148,163,184,0.14)] pt-3', className)}>
      {children}
    </div>
  );
}

interface ResponsiveTableProps {
  children: React.ReactNode;
  className?: string;
}

export function ResponsiveTable({ children, className }: ResponsiveTableProps) {
  return (
    <div className={cn('block md:hidden', className)}>
      {children}
    </div>
  );
}