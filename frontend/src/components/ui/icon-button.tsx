import { forwardRef } from 'react';
import { cn } from '../../utils';

interface IconButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  icon: React.ReactNode;
  title?: string;
  variant?: 'default' | 'primary' | 'success' | 'danger';
}

const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(
  ({ icon, title, variant = 'default', className, ...props }, ref) => {
    const variants = {
      default: {
        base: 'var(--text-secondary)',
        bg: 'var(--bg-surface-elevated)',
        border: 'var(--border-default)',
        shadow: 'var(--shadow-sm)',
        hoverBg: 'var(--bg-surface-3)',
        hoverBorder: 'var(--border-primary)',
        hoverShadow: 'var(--shadow-md)',
      },
      primary: {
        base: 'var(--color-primary)',
        bg: 'var(--color-primary-10)',
        border: 'var(--color-primary-25)',
        shadow: 'var(--shadow-sm)',
        hoverBg: 'var(--color-primary-20)',
        hoverBorder: 'var(--color-primary-40)',
        hoverShadow: 'var(--shadow-md)',
      },
      success: {
        base: 'var(--color-success)',
        bg: 'var(--color-success-10)',
        border: 'var(--color-success-20)',
        shadow: 'var(--shadow-sm)',
        hoverBg: 'var(--color-success-20)',
        hoverBorder: 'var(--color-success-30)',
        hoverShadow: 'var(--shadow-md)',
      },
      danger: {
        base: 'var(--color-danger)',
        bg: 'var(--color-danger-10)',
        border: 'var(--color-danger-20)',
        shadow: 'var(--shadow-sm)',
        hoverBg: 'var(--color-danger-20)',
        hoverBorder: 'var(--color-danger-30)',
        hoverShadow: 'var(--shadow-md)',
      },
    };

    const style = variants[variant];

    return (
      <button
        ref={ref}
        title={title}
        className={cn(
          'flex items-center justify-center transition-all duration-150 rounded-[10px] focus:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed',
          className
        )}
        style={{
          width: '36px',
          height: '36px',
          background: style.bg,
          border: `1px solid ${style.border}`,
          color: style.base,
          boxShadow: style.shadow,
          padding: '0',
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.background = style.hoverBg;
          e.currentTarget.style.borderColor = style.hoverBorder;
          e.currentTarget.style.boxShadow = style.hoverShadow;
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.background = style.bg;
          e.currentTarget.style.borderColor = style.border;
          e.currentTarget.style.boxShadow = style.shadow;
        }}
        {...props}
      >
        {icon}
      </button>
    );
  }
);

IconButton.displayName = 'IconButton';

export { IconButton };
