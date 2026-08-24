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
        base: 'var(--color-primary)',
        bg: 'rgba(99, 102, 241, 0.08)',
        border: 'rgba(99, 102, 241, 0.15)',
        shadow: '0 1px 2px rgba(99, 102, 241, 0.05)',
        hoverBg: 'rgba(99, 102, 241, 0.12)',
        hoverBorder: 'rgba(99, 102, 241, 0.25)',
        hoverShadow: '0 2px 4px rgba(99, 102, 241, 0.1)',
      },
      primary: {
        base: '#818cf8',
        bg: 'rgba(99, 102, 241, 0.12)',
        border: 'rgba(99, 102, 241, 0.25)',
        shadow: '0 2px 4px rgba(99, 102, 241, 0.1)',
        hoverBg: 'rgba(99, 102, 241, 0.18)',
        hoverBorder: 'rgba(99, 102, 241, 0.35)',
        hoverShadow: '0 4px 8px rgba(99, 102, 241, 0.15)',
      },
      success: {
        base: 'var(--color-success)',
        bg: 'rgba(16, 185, 129, 0.08)',
        border: 'rgba(16, 185, 129, 0.15)',
        shadow: '0 1px 2px rgba(16, 185, 129, 0.05)',
        hoverBg: 'rgba(16, 185, 129, 0.12)',
        hoverBorder: 'rgba(16, 185, 129, 0.25)',
        hoverShadow: '0 2px 4px rgba(16, 185, 129, 0.1)',
      },
      danger: {
        base: 'var(--color-danger)',
        bg: 'rgba(239, 68, 68, 0.08)',
        border: 'rgba(239, 68, 68, 0.15)',
        shadow: '0 1px 2px rgba(239, 68, 68, 0.05)',
        hoverBg: 'rgba(239, 68, 68, 0.12)',
        hoverBorder: 'rgba(239, 68, 68, 0.25)',
        hoverShadow: '0 2px 4px rgba(239, 68, 68, 0.1)',
      },
    };

    const style = variants[variant];

    return (
      <button
        ref={ref}
        title={title}
        className={cn(
          'flex items-center justify-center transition-all duration-200 rounded-[8px]',
          className
        )}
        style={{
          width: '32px',
          height: '32px',
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
