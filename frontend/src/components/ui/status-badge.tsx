import type { HTMLAttributes } from 'react';
import { forwardRef } from 'react';
import { cn } from '../../utils';

export interface StatusBadgeProps extends HTMLAttributes<HTMLDivElement> {
  status: 'available' | 'low-stock' | 'out-of-stock' | 'reserved' | 'current' | 'due-soon' | 'overdue' | 'paid';
  size?: 'sm' | 'md' | 'lg';
}

const StatusBadge = forwardRef<HTMLDivElement, StatusBadgeProps>(
  ({ className, status, size = 'md', ...props }, ref) => {
    const statusConfig = {
      'available': {
        label: 'متوفر',
        className: 'bg-green/10 border border-green/20 text-green'
      },
      'low-stock': {
        label: 'مخزون منخفض',
        className: 'bg-yellow/10 border border-yellow/20 text-yellow'
      },
      'out-of-stock': {
        label: 'نفذ المخزون',
        className: 'bg-red/10 border border-red/20 text-red'
      },
      'reserved': {
        label: 'محجوز',
        className: 'bg-blue/10 border border-blue/20 text-blue'
      },
      'current': {
        label: 'حالي',
        className: 'bg-green/10 border border-green/20 text-green'
      },
      'due-soon': {
        label: 'يستحق قريباً',
        className: 'bg-yellow/10 border border-yellow/20 text-yellow'
      },
      'overdue': {
        label: 'متأخر',
        className: 'bg-red/10 border border-red/20 text-red'
      },
      'paid': {
        label: 'مدفوع',
        className: 'bg-green/10 border border-green/20 text-green'
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