import { forwardRef, useState } from 'react';
import { cn } from '../../utils';
import { Search, X } from 'lucide-react';

export interface SearchInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size'> {
  containerClassName?: string;
  size?: 'sm' | 'md' | 'lg';
  showClear?: boolean;
  onClear?: () => void;
}

const SearchInput = forwardRef<HTMLInputElement, SearchInputProps>(
  ({ 
    className, 
    containerClassName,
    placeholder = 'بحث...',
    size = 'md',
    showClear = true,
    onClear,
    value,
    ...props 
  }, ref) => {
    const [isFocused, setIsFocused] = useState(false);
    const hasValue = value && value.toString().length > 0;

    const sizes = {
      sm: 'h-10 text-sm',
      md: 'h-12 text-sm',
      lg: 'h-14 text-base',
    };
    
    const handleClear = () => {
      if (onClear) {
        onClear();
      }
      // Also trigger input change if provided
      if (props.onChange) {
        const event = {
          target: { value: '' }
        } as React.ChangeEvent<HTMLInputElement>;
        props.onChange(event);
      }
    };
    
    return (
      <div className={cn('w-full', containerClassName)}>
        {/* Search Icon - Above the input */}
        <div className="flex items-center gap-2 mb-2">
          <Search
            className={cn(
              'transition-all duration-300',
              isFocused ? 'text-cyan-400 scale-110' : 'text-text-muted/50 scale-100',
              size === 'sm' ? 'w-4 h-4' : size === 'md' ? 'w-5 h-5' : 'w-6 h-6'
            )}
          />
          <span className="text-xs text-text-muted/50">{placeholder}</span>
        </div>

        {/* Input Field */}
        <div className="relative">
          <input
            ref={ref}
            type="text"
            placeholder={placeholder}
            value={value}
            className={cn(
              'search-input-custom',
              'flex w-full rounded-xl border',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:border-transparent',
              'transition-all duration-300',
              'disabled:cursor-not-allowed disabled:opacity-50',
              'placeholder:text-text-muted/40',
              'hover:border-cyan/30 hover:shadow-sm',
              sizes[size],
              'focus-visible:ring-cyan/30 focus-visible:border-cyan/50 focus-visible:shadow-lg',
              showClear && hasValue && 'pr-10', // space for clear button
              className
            )}
            style={{
              background: 'linear-gradient(135deg, rgba(30, 41, 59, 0.9) 0%, rgba(30, 41, 59, 0.7) 100%)',
              border: isFocused ? '2px solid #6366f1' : '1px solid #6366f1',
              backdropFilter: 'blur(10px)',
              color: 'var(--text-primary)',
              fontSize: '14px',
              fontWeight: '500',
              letterSpacing: '0.2px',
              boxShadow: isFocused ? '0 0 0 3px rgba(99, 102, 241, 0.3), 0 4px 20px rgba(99, 102, 241, 0.4)' : '0 2px 8px rgba(0, 0, 0, 0.1)',
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
            }}
            onFocus={(e) => {
              setIsFocused(true);
              if (props.onFocus) props.onFocus(e);
            }}
            onBlur={(e) => {
              setIsFocused(false);
              if (props.onBlur) props.onBlur(e);
            }}
            {...props}
          />

          {/* Clear Button - Right */}
          {showClear && hasValue && (
            <button
              type="button"
              onClick={handleClear}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted/40 hover:text-text-muted/70 transition-colors duration-200 pointer-events-auto z-10"
              style={{
                background: 'transparent',
                border: 'none',
                cursor: 'pointer',
                padding: '4px',
                borderRadius: '4px'
              }}
            >
              <X
                className={cn(
                  size === 'sm' ? 'w-4 h-4' : size === 'md' ? 'w-5 h-5' : 'w-6 h-6'
                )}
              />
            </button>
          )}
        </div>
      </div>
    );
  }
);

SearchInput.displayName = 'SearchInput';

export { SearchInput };