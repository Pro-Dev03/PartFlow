import { forwardRef, type InputHTMLAttributes } from 'react';
import { Input } from './input';
import { Search, X } from 'lucide-react';
import { cn } from '../../utils';

export interface SearchInputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
  onClear?: () => void;
  isLoading?: boolean;
  placeholder?: string;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
}

const SearchInput = forwardRef<HTMLInputElement, SearchInputProps>(
  ({ 
    className, 
    onClear, 
    isLoading = false,
    placeholder = 'بحث...',
    size = 'md',
    value,
    ...props 
  }, ref) => {
    const hasValue = value && value.toString().length > 0;

    return (
      <div className="pf-search-input relative w-full max-w-full min-w-0 overflow-hidden">
        <Search
          className="pf-search-icon absolute top-1/2 end-3 -translate-y-1/2"
          style={{ 
            width: 'var(--icon-size-sm)', 
            height: 'var(--icon-size-sm)' 
          }}
        />
        <Input
          ref={ref}
          type="text"
          placeholder={placeholder}
          value={value}
          size={size}
          className={cn('pf-search-field pf-search-input-canonical w-full max-w-full min-w-0', className)}
          {...props}
        />
        {(hasValue || isLoading) && (
          <button
            type="button"
            onClick={onClear}
            aria-label="مسح البحث"
            className="pf-search-clear absolute left-2 top-1/2 inline-flex h-7 w-7 -translate-y-1/2 items-center justify-center rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface-elevated)] text-[var(--text-muted)] shadow-sm transition-colors hover:border-[var(--primary)] hover:bg-[var(--color-primary-10)] hover:text-[var(--primary)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--primary)]/30"
            disabled={isLoading}
          >
            {isLoading ? (
              <div className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-text-muted/20 border-t-text-muted" />
            ) : (
              <X 
                style={{ 
                  width: 'var(--icon-size-sm)', 
                  height: 'var(--icon-size-sm)' 
                }} 
              />
            )}
          </button>
        )}
      </div>
    );
  }
);

SearchInput.displayName = 'SearchInput';

export { SearchInput };