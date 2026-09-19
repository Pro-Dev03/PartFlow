import { forwardRef } from 'react';
import { cn } from '../../utils';
import { Package, Users, ShoppingCart, AlertCircle, FileText, CheckCircle } from 'lucide-react';

export interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
    variant?: 'primary' | 'secondary';
  };
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

const EmptyState = forwardRef<HTMLDivElement, EmptyStateProps>(
  ({ 
    icon, 
    title, 
    description, 
    action,
    className,
    size = 'md'
  }, ref) => {
    const sizeStyles = {
      sm: {
        minHeight: 'var(--empty-state-min-height)',
        padding: 'var(--spacing-4)',
        iconSize: 'var(--icon-size-lg)'
      },
      md: {
        minHeight: 'var(--empty-state-min-height)',
        padding: 'var(--spacing-8)',
        iconSize: 'var(--icon-size-2xl)'
      },
      lg: {
        minHeight: 'calc(var(--empty-state-min-height) * 1.5)',
        padding: 'var(--spacing-10)',
        iconSize: 'var(--icon-size-2xl)'
      }
    };

    const currentSize = sizeStyles[size];

    return (
      <div
        ref={ref}
        className={cn(
          'flex flex-col items-center justify-center text-center',
          className
        )}
        style={{
          minHeight: currentSize.minHeight,
          padding: currentSize.padding
        }}
      >
        {icon && (
          <div 
            className="mb-4 text-text-muted/50"
            style={{ 
              width: currentSize.iconSize, 
              height: currentSize.iconSize 
            }}
          >
            {icon}
          </div>
        )}
        <h3 
          className="font-semibold text-text mb-2"
          style={{ fontSize: 'var(--font-size-body)' }}
        >
          {title}
        </h3>
        {description && (
          <p 
            className="text-text-muted mb-4 max-w-md"
            style={{ fontSize: 'var(--font-size-secondary)' }}
          >
            {description}
          </p>
        )}
        {action && (
          <button
            onClick={action.onClick}
            className={cn(
              'mt-4 px-4 py-2 rounded-lg font-medium transition-colors',
              action.variant === 'primary' 
                ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
            )}
            style={{ fontSize: 'var(--font-size-secondary)' }}
          >
            {action.label}
          </button>
        )}
      </div>
    );
  }
);

EmptyState.displayName = 'EmptyState';

// Pre-built empty states for common use cases
export const EmptyStates = {
  products: (props?: Partial<EmptyStateProps>) => (
    <EmptyState
      icon={<Package />}
      title="لا توجد منتجات"
      description="أضف أول منتج إلى مخزونك لتبدأ بإدارة متجرك"
      action={{ label: 'إضافة منتج', onClick: () => {} }}
      {...props}
    />
  ),
  customers: (props?: Partial<EmptyStateProps>) => (
    <EmptyState
      icon={<Users />}
      title="لا يوجد عملاء"
      description="أضف أول عميل لتبدأ في تتبع معاملاتك"
      action={{ label: 'إضافة عميل', onClick: () => {} }}
      {...props}
    />
  ),
  sales: (props?: Partial<EmptyStateProps>) => (
    <EmptyState
      icon={<ShoppingCart />}
      title="لا توجد مبيعات"
      description="ابدأ ببيع المنتجات لتتبع أداء متجرك"
      action={{ label: 'بيع جديد', onClick: () => {} }}
      {...props}
    />
  ),
  inventory: (props?: Partial<EmptyStateProps>) => (
    <EmptyState
      icon={<Package />}
      title="المخزون فارغ"
      description="أضف قطع إلى المخزون لتبدأ في البيع"
      action={{ label: 'إضافة قطعة', onClick: () => {} }}
      {...props}
    />
  ),
  debts: (props?: Partial<EmptyStateProps>) => (
    <EmptyState
      icon={<AlertCircle />}
      title="لا توجد ديون"
      description="جميع العملاء مدفوعون بالكامل"
      {...props}
    />
  ),
  reports: (props?: Partial<EmptyStateProps>) => (
    <EmptyState
      icon={<FileText />}
      title="لا توجد بيانات للعرض"
      description="ابدأ بإضافة منتجات وإجراء مبيعات لتوليد التقارير"
      {...props}
    />
  ),
  completed: (props?: Partial<EmptyStateProps>) => (
    <EmptyState
      icon={<CheckCircle />}
      title="مكتمل"
      description="تم إنجاز جميع المهام بنجاح"
      {...props}
    />
  ),
};

export { EmptyState };