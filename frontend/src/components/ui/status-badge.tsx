import type { HTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../lib/utils';

export interface StatusBadgeProps extends HTMLAttributes<HTMLDivElement> {
  status: 'available' | 'low-stock' | 'out-of-stock' | 'reserved' | 'current' | 'due-soon' | 'overdue' | 'paid';
  size?: 'sm' | 'md' | 'lg';
}

const StatusBadge = forwardRef<HTMLDivElement, StatusBadgeProps>(
  ({ className, status, size = 'md', ...props }, ref) => {
    const statusConfig = {
      'available': {
        label: 'متوفر',
        className: 'bg-success/10 border border-success/20 text-success'
      },
      'low-stock': {
        label: 'مخزون منخفض',
        className: 'bg-warning/10 border border-warning/20 text-warning'
      },
      'out-of-stock': {
        label: 'نفذ المخزون',
        className: 'bg-danger/10 border border-danger/20 text-danger'
      },
      'reserved': {
        label: 'محجوز',
        className: 'bg-info/10 border border-info/20 text-info'
      },
      'current': {
        label: 'حالي',
        className: 'bg-success/10 border border-success/20 text-success'
      },
      'due-soon': {
        label: 'يستحق قريباً',
        className: 'bg-warning/10 border border-warning/20 text-warning'
      },
      'overdue': {
        label: 'متأخر',
        className: 'bg-danger/10 border border-danger/20 text-danger'
      },
      'paid': {
        label: 'مدفوع',
        className: 'bg-success/10 border border-success/20 text-success'
      }
    };
    
    const config = statusConfig[status];
    const sizes = {
      sm: 'px-2 py-0.5 text-xs',
      md: 'px-2.5 py-1 text-sm',
      lg: 'px-3 py-1.5 text-base',
    };
    
    return (
      <div
        ref={ref}
        className={cn(
          'inline-flex items-center rounded-md font-medium transition-colors',
          config.className,
          sizes[size],
          className
        )}
        {...props}
      >
        {config.label}
      </div>
    );
  }
);

StatusBadge.displayName = 'StatusBadge';

export { StatusBadge };