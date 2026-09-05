import { forwardRef, type CSSProperties, type HTMLAttributes, useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { cn } from '../../utils';
import { X, Sparkles } from 'lucide-react';

export interface ModalProps extends HTMLAttributes<HTMLDivElement> {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | 'full';
  variant?: 'default' | 'elegant' | 'modern' | 'glass';
  showHeader?: boolean;
  showCloseButton?: boolean;
  headerStyle?: CSSProperties;
  autoFocus?: boolean; // Auto-focus on first input when modal opens
  enableEnterNavigation?: boolean; // Enable Enter key to move to next field
  'aria-label'?: string;
  'aria-describedby'?: string;
}

export const Modal = forwardRef<HTMLDivElement, ModalProps>(
  ({ 
    className, 
    isOpen, 
    onClose, 
    title, 
    size = 'md', 
    variant = 'modern',
    showHeader = true,
    showCloseButton = true,
    headerStyle,
    autoFocus = true,
    enableEnterNavigation = true,
    'aria-label': ariaLabel,
    'aria-describedby': ariaDescribedby,
    children, 
    ...props 
  }, ref) => {
    const modalRef = useRef<HTMLDivElement>(null);
    const previousActiveElement = useRef<HTMLElement | null>(null);
    const onCloseRef = useRef(onClose);
    const autoFocusTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    onCloseRef.current = onClose;

    useEffect(() => {
      if (!isOpen) return;

      previousActiveElement.current = document.activeElement as HTMLElement;

      // Auto-focus on first input/select after a small delay
      if (autoFocus) {
        autoFocusTimeoutRef.current = setTimeout(() => {
          const focusableElements = modalRef.current?.querySelectorAll(
            'input:not([disabled]), select:not([disabled]), textarea:not([disabled]), button:not([disabled])'
          );

          if (focusableElements && focusableElements.length > 0) {
            const firstInput = Array.from(focusableElements).find(
              el => el.tagName === 'INPUT' || el.tagName === 'SELECT' || el.tagName === 'TEXTAREA'
            ) as HTMLElement;
            if (firstInput) {
              firstInput.focus();
            }
          }
        }, 100);
      }
    }, [isOpen, autoFocus]);

    useEffect(() => {
      if (!isOpen) return;

      const handleKeyDown = (e: KeyboardEvent) => {
        if (e.key === 'Escape') {
          onCloseRef.current();
        }
        
        // Enhanced TAB navigation
        if (e.key === 'Tab') {
          const focusableElements = modalRef.current?.querySelectorAll(
            'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
          );
          if (focusableElements && focusableElements.length > 0) {
            const firstElement = focusableElements[0] as HTMLElement;
            const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;
            
            if (e.shiftKey) {
              if (document.activeElement === firstElement) {
                e.preventDefault();
                lastElement.focus();
              }
            } else {
              if (document.activeElement === lastElement) {
                e.preventDefault();
                firstElement.focus();
              }
            }
          }
        }

        // Enter key navigation for inputs
        if (enableEnterNavigation && e.key === 'Enter' && !e.shiftKey) {
          const activeElement = document.activeElement;
          if (activeElement && (
            activeElement.tagName === 'INPUT' || 
            activeElement.tagName === 'SELECT'
          )) {
            const focusableElements = modalRef.current?.querySelectorAll(
              'input:not([disabled]), select:not([disabled]), textarea:not([disabled]), button:not([disabled])'
            );
            if (focusableElements && focusableElements.length > 0) {
              const elementsArray = Array.from(focusableElements);
              const currentIndex = elementsArray.indexOf(activeElement);
              
              // Find next input/select (skip buttons)
              for (let i = currentIndex + 1; i < elementsArray.length; i++) {
                const nextElement = elementsArray[i];
                if (nextElement.tagName === 'INPUT' || 
                    nextElement.tagName === 'SELECT' || 
                    nextElement.tagName === 'TEXTAREA') {
                  e.preventDefault();
                  (nextElement as HTMLElement).focus();
                  break;
                }
              }
            }
          }
        }
      };

      document.addEventListener('keydown', handleKeyDown);
      document.body.style.overflow = 'hidden';

      return () => {
        if (autoFocusTimeoutRef.current) {
          clearTimeout(autoFocusTimeoutRef.current);
        }
        document.removeEventListener('keydown', handleKeyDown);
        document.body.style.overflow = '';
        if (previousActiveElement.current) {
          previousActiveElement.current.focus();
        }
      };
    }, [isOpen, autoFocus, enableEnterNavigation]);

    if (!isOpen) {
      return null;
    }

    const sizes = {
      sm: 'max-w-[400px]',
      md: 'max-w-[500px]',
      lg: 'max-w-[600px]',
      xl: 'max-w-[800px]',
      '2xl': 'max-w-[1000px]',
      full: 'max-w-full mx-4',
    };

    const variantStyles = {
      default: {
        background: 'var(--bg-surface)',
        border: '1px solid var(--border-default)',
        shadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(255, 255, 255, 0.05) inset',
      },
      elegant: {
        background: 'linear-gradient(135deg, #1e1e2e 0%, #2d2d44 100%)',
        border: '1px solid rgba(99, 102, 241, 0.3)',
        shadow: '0 25px 50px -12px rgba(99, 102, 241, 0.25), 0 0 0 1px rgba(99, 102, 241, 0.1) inset, 0 0 40px rgba(99, 102, 241, 0.15)',
      },
      modern: {
        background: 'var(--bg-surface)',
        border: '1px solid var(--border-primary)',
        shadow: 'rgba(0, 0, 0, 0.3) 0px 20px 60px, rgba(255, 255, 255, 0.1) 0px 0px 0px 1px inset, rgba(99, 102, 241, 0.1) 0px 0px 40px',
      },
      glass: {
        background: 'rgba(255, 255, 255, 0.7)',
        backdropFilter: 'blur(20px)',
        border: '1px solid rgba(255, 255, 255, 0.3)',
        shadow: '0 25px 50px -12px rgba(0, 0, 0, 0.15), 0 0 0 1px rgba(255, 255, 255, 0.2) inset',
      },
    };

    const currentVariant = variantStyles[variant];

    const modalContent = (
      <div 
        className="fixed inset-0 z-[9999] flex items-center justify-center p-4"
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? 'modal-title' : undefined}
        aria-label={ariaLabel || title}
        aria-describedby={ariaDescribedby}
        style={{ 
          zIndex: 9999,
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0,0,0,0.5)',
          backdropFilter: 'blur(8px)'
        }}
      >
        <div
          ref={(node) => {
            modalRef.current = node;
            if (typeof ref === 'function') {
              ref(node);
            } else if (ref) {
              ref.current = node;
            }
          }}
          className={cn(
            'relative w-full rounded-2xl overflow-hidden',
            'max-h-[calc(100vh-2rem)]',
            sizes[size],
            className
          )}
          tabIndex={-1}
          style={{
            zIndex: 10000,
            position: 'relative',
            background: currentVariant.background,
            border: currentVariant.border,
            boxShadow: currentVariant.shadow,
            WebkitOverflowScrolling: 'touch',
            transform: 'none',
            opacity: 1,
            animation: 'none'
          }}
          {...props}
        >
          {variant === 'elegant' && (
            <div 
              className="absolute inset-0 pointer-events-none opacity-30"
              style={{
                background: 'radial-gradient(circle at top right, rgba(99, 102, 241, 0.15) 0%, transparent 50%)',
              }}
            />
          )}

          {showHeader && title && (
            <div 
              className={cn(
                "flex items-center justify-between px-6 py-5",
                variant === 'elegant' ? "border-b border-white/10" : "border-b border-gray-200"
              )}
              style={{
                background: variant === 'elegant' 
                  ? 'rgba(255, 255, 255, 0.05)' 
                  : 'var(--bg-surface)',
                ...headerStyle,
              }}
            >
              <div className="flex items-center gap-3">
                {variant === 'modern' && (
                  <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: 'rgba(99, 102, 241, 0.15)' }}>
                    <Sparkles className="w-4 h-4" style={{ color: 'var(--color-primary)' }} />
                  </div>
                )}
                {variant === 'elegant' && (
                  <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: 'rgba(99, 102, 241, 0.2)' }}>
                    <Sparkles className="w-4 h-4" style={{ color: 'var(--color-primary)' }} />
                  </div>
                )}
                <h3 
                  id="modal-title" 
                  className="text-lg font-semibold"
                  style={{ 
                    color: variant === 'elegant' ? '#f1f7ff' : 'var(--text-primary)',
                    fontWeight: '600',
                    fontSize: '16px'
                  }}
                >
                  {title}
                </h3>
              </div>
              
              {showCloseButton && (
                <button
                  onClick={onClose}
                  className={cn(
                    "p-2 rounded-lg transition-all duration-200",
                    variant === 'elegant' 
                      ? "hover:bg-white/10 text-gray-400 hover:text-white hover:scale-110" 
                      : "hover:bg-gray-100 text-gray-500 hover:text-gray-900 hover:scale-110"
                  )}
                  aria-label="إغلاق"
                  style={{
                    background: 'transparent',
                    border: 'none',
                    cursor: 'pointer',
                  }}
                >
                  <X className="w-5 h-5" />
                </button>
              )}
            </div>
          )}

          <div 
            className="p-6 overflow-y-auto"
            style={{ 
              color: variant === 'elegant' ? '#e2e8f0' : 'var(--text-primary)',
              background: 'transparent',
              maxHeight: 'calc(80vh - 150px)',
              overflowX: 'hidden'
            }}
          >
            {children}
          </div>

          {variant === 'modern' && (
            <div 
              className="absolute bottom-0 left-0 right-0 h-1"
              style={{
                background: 'linear-gradient(90deg, var(--color-primary) 0%, var(--color-info) 50%, var(--color-success) 100%)',
              }}
            />
          )}
        </div>
      </div>
    );

    return createPortal(modalContent, document.body);
  }
);

Modal.displayName = 'Modal';

export default Modal;