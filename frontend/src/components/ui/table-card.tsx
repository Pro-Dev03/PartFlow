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
        'bg-surface border border-border rounded-lg p-4',
        'hover:border-cyan/30 transition-colors duration-200',
        className
      )}
    >
      {children}
    </div>
  );
}

export function TableCardItem({ label, value, className, variant = 'default' }: TableCardItemProps) {
  const variantStyles = {
    default: 'text-text',
    highlight: 'text-cyan font-semibold',
    success: 'text-green font-semibold',
    warning: 'text-yellow font-semibold',
    danger: 'text-red font-semibold',
  };

  return (
    <div className={cn('flex justify-between items-start py-2 border-b border-border/50 last:border-0', className)}>
      <span className="text-small text-text-muted font-medium">{label}</span>
      <span className={cn('text-small font-medium', variantStyles[variant])}>{value}</span>
    </div>
  );
}

interface TableCardActionsProps {
  children: React.ReactNode;
  className?: string;
}

export function TableCardActions({ children, className }: TableCardActionsProps) {
  return (
    <div className={cn('flex gap-2 pt-3 mt-2 border-t border-border', className)}>
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