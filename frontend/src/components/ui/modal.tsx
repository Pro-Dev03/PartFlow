import { forwardRef, type HTMLAttributes, useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { cn } from '../../utils';
import { X } from 'lucide-react';

export interface ModalProps extends HTMLAttributes<HTMLDivElement> {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | 'full';
}

export const Modal = forwardRef<HTMLDivElement, ModalProps>(
  ({ className, isOpen, onClose, title, size = 'md', children, ...props }, ref) => {
    const modalRef = useRef<HTMLDivElement>(null);
    const previousActiveElement = useRef<HTMLElement | null>(null);

    useEffect(() => {
      if (!isOpen) return () => {};

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
    }, [isOpen, onClose]);

    if (!isOpen) {
      return null;
    }

    const sizes = {
      sm: 'max-w-sm',
      md: 'max-w-md',
      lg: 'max-w-lg',
      xl: 'max-w-xl',
      '2xl': 'max-w-2xl',
      full: 'max-w-full mx-4',
    };

    const modalContent = (
      <div 
        className="fixed inset-0 z-[9999] flex items-center justify-center p-4"
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? 'modal-title' : undefined}
        style={{ zIndex: 9999, position: 'fixed', top: 0, left: 0, right: 0, bottom: 0 }}
      >
        {/* Backdrop */}
        <div
          className="absolute inset-0 bg-black/50 backdrop-blur-sm"
          onClick={onClose}
          aria-hidden="true"
          style={{ zIndex: 9998, position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(0,0,0,0.5)' }}
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
            'relative w-full bg-white rounded-2xl shadow-lg border border-gray-200',
            'transition-all duration-200',
            'max-h-[calc(100vh-2rem)] overflow-y-auto',
            sizes[size],
            className
          )}
          tabIndex={-1}
          style={{ zIndex: 10000, position: 'relative', backgroundColor: 'white', display: 'block' }}
          {...props}
        >
          {/* Header */}
          {title && (
            <div className="flex items-center justify-between p-6 border-b border-gray-200 sticky top-0 bg-white z-10" style={{ backgroundColor: 'white' }}>
              <h3 id="modal-title" className="text-lg font-semibold text-gray-900" style={{ color: '#111827' }}>
                {title}
              </h3>
              <button
                onClick={onClose}
                className="p-2 rounded-lg hover:bg-gray-100 transition-colors text-gray-500 hover:text-gray-900"
                aria-label="إغلاق"
                style={{ color: '#6b7280' }}
              >
                <X className="w-5 h-5" />
              </button>
            </div>
          )}

          {/* Content */}
          <div className="p-6" style={{ color: '#111827' }}>
            {children}
          </div>
        </div>
      </div>
    );

    return createPortal(modalContent, document.body);
  }
);

Modal.displayName = 'Modal';