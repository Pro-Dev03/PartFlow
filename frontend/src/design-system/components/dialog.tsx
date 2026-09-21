import { forwardRef, useEffect, type HTMLAttributes } from 'react';
import { createPortal } from 'react-dom';
import { cn } from '../../utils/helpers';
import { X } from 'lucide-react';

export interface DialogProps extends HTMLAttributes<HTMLDivElement> {
  open: boolean;
  onClose: () => void;
  title?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
}

const Dialog = forwardRef<HTMLDivElement, DialogProps>(
  ({ className, open, onClose, title, size = 'md', children, ...props }, ref) => {
    useEffect(() => {
      if (!open) return undefined;

      const handleKeyDown = (event: KeyboardEvent) => {
        if (event.key === 'Escape') onClose();
      };

      document.addEventListener('keydown', handleKeyDown);
      return () => document.removeEventListener('keydown', handleKeyDown);
    }, [open, onClose]);

    if (!open) return null;

    const sizes = {
      sm: 'max-w-md',
      md: 'max-w-lg',
      lg: 'max-w-2xl',
      xl: 'max-w-4xl',
      full: 'max-w-full',
    };

    return createPortal(
      <div
        className="fixed inset-0 z-[10000] flex items-center justify-center overflow-y-auto bg-black/50 p-4 backdrop-blur-sm"
        role="presentation"
        onMouseDown={(event) => {
          if (event.target === event.currentTarget) onClose();
        }}
      >
        <div
          ref={ref}
          className={cn(
            'bg-surface max-h-[calc(100vh-2rem)] overflow-y-auto rounded-xl border border-border shadow-card w-full animate-scale-in',
            sizes[size],
            className
          )}
          role="dialog"
          aria-modal="true"
          {...props}
        >
          {title && (
            <div className="flex items-center justify-between p-6 border-b border-border">
              <h2 className="text-xl font-semibold text-text-primary">{title}</h2>
              <button
                onClick={onClose}
                className="text-text-secondary hover:text-text-primary transition-colors duration-normal"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
          )}
          <div className="p-6">{children}</div>
        </div>
      </div>,
      document.body
    );
  }
);

Dialog.displayName = 'Dialog';

export { Dialog };