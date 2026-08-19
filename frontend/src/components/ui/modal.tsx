import { forwardRef, type HTMLAttributes, useEffect, useRef } from 'react';
import { cn } from '../../lib/utils';
import { X } from 'lucide-react';

export interface ModalProps extends HTMLAttributes<HTMLDivElement> {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
}

export const Modal = forwardRef<HTMLDivElement, ModalProps>(
  ({ className, isOpen, onClose, title, size = 'md', children, ...props }, ref) => {
    const modalRef = useRef<HTMLDivElement>(null);
    const previousActiveElement = useRef<HTMLElement | null>(null);

    useEffect(() => {
      if (isOpen) {
        // Store the previously focused element
        previousActiveElement.current = document.activeElement as HTMLElement;
        
        // Focus the modal when it opens
        setTimeout(() => {
          modalRef.current?.focus();
        }, 100);

        // Trap focus within modal
        const handleKeyDown = (e: KeyboardEvent) => {
          if (e.key === 'Escape') {
            onClose();
          }
          if (e.key === 'Tab') {
            e.preventDefault();
            // Simple focus trap - could be enhanced with more sophisticated logic
            const focusableElements = modalRef.current?.querySelectorAll(
              'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
            );
            if (focusableElements && focusableElements.length > 0) {
              const firstElement = focusableElements[0] as HTMLElement;
              const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;
              
              if (e.shiftKey) {
                if (document.activeElement === firstElement) {
                  lastElement.focus();
                }
              } else {
                if (document.activeElement === lastElement) {
                  firstElement.focus();
                }
              }
            }
          }
        };

        document.addEventListener('keydown', handleKeyDown);
        
        // Prevent body scroll
        document.body.style.overflow = 'hidden';

        return () => {
          document.removeEventListener('keydown', handleKeyDown);
          document.body.style.overflow = '';
          
          // Restore focus to previous element when modal closes
          if (previousActiveElement.current) {
            previousActiveElement.current.focus();
          }
        };
      }
    }, [isOpen, onClose]);

    if (!isOpen) return null;

    const sizes = {
      sm: 'max-w-sm',
      md: 'max-w-md',
      lg: 'max-w-lg',
      xl: 'max-w-xl',
    };

    return (
      <div 
        className="fixed inset-0 z-50 flex items-center justify-center"
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? 'modal-title' : undefined}
      >
        {/* Backdrop */}
        <div
          className="absolute inset-0 bg-black/50 backdrop-blur-sm"
          onClick={onClose}
          aria-hidden="true"
        />

        {/* Modal */}
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
            'relative w-full bg-surface rounded-xl shadow-xl border border-border',
            'transition-all duration-200',
            sizes[size],
            className
          )}
          tabIndex={-1}
          {...props}
        >
          {/* Header */}
          {title && (
            <div className="flex items-center justify-between p-6 border-b border-border">
              <h3 id="modal-title" className="text-lg font-semibold text-text-primary">
                {title}
              </h3>
              <button
                onClick={onClose}
                className="p-2 rounded-lg hover:bg-surface-elevated transition-colors text-text-secondary hover:text-text-primary"
                aria-label="إغلاق"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
          )}

          {/* Content */}
          <div className="p-6">
            {children}
          </div>
        </div>
      </div>
    );
  }
);

Modal.displayName = 'Modal';