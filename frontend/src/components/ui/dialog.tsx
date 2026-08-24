import { forwardRef, type HTMLAttributes } from 'react';
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
    if (!open) return null;

    const sizes = {
      sm: 'max-w-md',
      md: 'max-w-lg',
      lg: 'max-w-2xl',
      xl: 'max-w-4xl',
      full: 'max-w-full',
    };

    return (
      <div className="fixed inset-0 z-modal flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
        <div
          ref={ref}
          className={cn(
            'bg-surface rounded-xl shadow-card border border-border w-full animate-scale-in',
            sizes[size],
            className
          )}
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
      </div>
    );
  }
);

Dialog.displayName = 'Dialog';

export { Dialog };