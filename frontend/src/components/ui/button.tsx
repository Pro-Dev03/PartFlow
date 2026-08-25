import type { ButtonHTMLAttributes } from 'react';
import { forwardRef, useState, useEffect } from 'react';
import { cn } from '../../utils';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'success' | 'outline' | 'default' | 'warning' | 'info';
  size?: 'sm' | 'md' | 'lg' | 'icon';
  isLoading?: boolean;
  fullWidth?: boolean;
}

// القيمة الافتراضية للحجم هي sm مثل زر التوصية
const DEFAULT_SIZE: 'sm' = 'sm';

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({
    className,
    variant = 'primary',
    size = DEFAULT_SIZE,
    isLoading = false,
    disabled = false,
    fullWidth = false,
    children,
    ...props
  }, ref) => {
    const [isMobile, setIsMobile] = useState(false);
    
    useEffect(() => {
      const checkMobile = () => {
        setIsMobile(window.innerWidth < 768);
      };
      
      checkMobile();
      window.addEventListener('resize', checkMobile);
      return () => window.removeEventListener('resize', checkMobile);
    }, []);
    
    const isDisabled = disabled || isLoading;
    
    const getVariantStyle = () => {
      // تصميم احترافي للوضع الفاتح والداكن
      const baseStyle: Record<string, string | number> = {
        borderColor: 'var(--color-primary-20)',
        background: 'var(--color-primary-08)',
        color: 'var(--text-primary)',
        borderRadius: '8px',
        padding: '6px 10px',
        cursor: isDisabled ? 'not-allowed' : 'pointer',
        transition: 'all 0.15s ease',
        fontSize: '12px',
        fontWeight: '500',
        opacity: isDisabled ? 0.5 : 1,
        transform: 'translateY(0px)',
        boxShadow: 'none',
        lineHeight: '1.4',
        minHeight: '32px'
      };

      const variantStyles: Record<string, Record<string, string | number>> = {
        primary: {
          ...baseStyle,
          borderColor: 'var(--color-primary-30)',
          background: 'var(--button-primary-bg)',
          color: 'var(--button-primary-text)',
          fontWeight: '600',
          boxShadow: '0 2px 8px var(--color-primary-20)'
        },
        secondary: {
          ...baseStyle,
          borderColor: 'var(--border-default)',
          background: 'var(--button-secondary-bg)',
          color: 'var(--button-secondary-text)',
          boxShadow: '0 1px 3px rgba(0, 0, 0, 0.05)'
        },
        ghost: {
          ...baseStyle,
          borderColor: 'transparent',
          background: 'transparent',
          color: 'var(--text-primary)',
          boxShadow: 'none'
        },
        danger: {
          ...baseStyle,
          borderColor: 'var(--color-danger-30)',
          background: 'var(--color-danger-08)',
          color: 'var(--color-danger)',
          boxShadow: '0 1px 3px var(--color-danger-10)'
        },
        success: {
          ...baseStyle,
          borderColor: 'var(--color-success-30)',
          background: 'var(--color-success-08)',
          color: 'var(--color-success)',
          boxShadow: '0 1px 3px var(--color-success-10)'
        },
        outline: {
          ...baseStyle,
          background: 'transparent',
          color: 'var(--text-primary)',
          borderColor: 'var(--color-primary-20)'
        },
        default: baseStyle,
        warning: {
          ...baseStyle,
          borderColor: 'var(--color-warning-30)',
          background: 'var(--color-warning-08)',
          color: 'var(--color-warning)',
          boxShadow: '0 1px 3px var(--color-warning-10)'
        },
        info: {
          ...baseStyle,
          borderColor: 'var(--color-info-30)',
          background: 'var(--color-info-08)',
          color: 'var(--color-info)',
          boxShadow: '0 1px 3px var(--color-info-10)'
        }
      };

      return variantStyles[variant] || variantStyles.default;
    };
  
    const getSizeStyle = () => {
      // تصميم احترافي مثل زر التوصية - نفس الحجم لجميع الأزرار
      const sizes = {
        sm: {
          padding: isMobile ? '8px 12px' : '6px 10px',
          fontSize: isMobile ? '13px' : '12px',
          lineHeight: '1.4',
          minHeight: isMobile ? '36px' : '32px'
        },
        md: {
          padding: isMobile ? '10px 16px' : '8px 14px',
          fontSize: isMobile ? '14px' : '13px',
          lineHeight: '1.4',
          minHeight: isMobile ? '40px' : '36px'
        },
        lg: {
          padding: isMobile ? '12px 20px' : '10px 18px',
          fontSize: isMobile ? '15px' : '14px',
          lineHeight: '1.4',
          minHeight: isMobile ? '44px' : '40px'
        },
        icon: {
          padding: isMobile ? '8px' : '6px',
          fontSize: isMobile ? '15px' : '14px',
          width: isMobile ? '36px' : '32px',
          height: isMobile ? '36px' : '32px',
          lineHeight: '1',
          minHeight: isMobile ? '36px' : '32px'
        }
      };
      return sizes[size] || sizes.sm;
    };
    
    return (
      <button
        ref={ref}
        style={{
          ...getVariantStyle(),
          ...getSizeStyle(),
          ...(fullWidth ? { width: '100%' } : {})
        }}
        className={cn('inline-flex items-center justify-center gap-2', className)}
        disabled={isDisabled}
        aria-disabled={isDisabled}
        aria-busy={isLoading}
        onMouseEnter={(e) => {
          if (!isDisabled && variant === 'primary') {
            e.currentTarget.style.transform = 'translateY(-1px)';
            e.currentTarget.style.borderColor = 'var(--color-primary-30)';
            e.currentTarget.style.boxShadow = 'var(--shadow-glow)';
          } else if (!isDisabled && variant !== 'primary' && variant !== 'ghost') {
            e.currentTarget.style.transform = 'translateY(-1px)';
            e.currentTarget.style.borderColor = 'var(--primary)';
          } else if (!isDisabled && variant === 'ghost') {
            e.currentTarget.style.transform = 'translateY(-1px)';
            e.currentTarget.style.background = 'var(--bg-surface-elevated)';
          }
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.transform = 'translateY(0)';
          const baseStyle = getVariantStyle();
          e.currentTarget.style.borderColor = baseStyle.borderColor as string;
          e.currentTarget.style.boxShadow = (baseStyle.boxShadow as string) || 'none';
          if (variant === 'ghost') {
            e.currentTarget.style.background = 'transparent';
          }
        }}
        onMouseDown={(e) => {
          e.currentTarget.style.transform = 'translateY(0)';
        }}
        {...props}
      >
        {isLoading && (
          <svg
            className="animate-spin w-4 h-4"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        )}
        <span className={cn('transition-opacity', isLoading && 'opacity-50')}>{children}</span>
      </button>
    );
  }
);

Button.displayName = 'Button';

export { Button };
