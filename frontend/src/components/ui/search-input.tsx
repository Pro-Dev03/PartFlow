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
      if (props.onChange) {
        const event = {
          target: { value: '' }
        } as React.ChangeEvent<HTMLInputElement>;
        props.onChange(event);
      }
    };
    
    return (
      <div className={cn('w-full', containerClassName)}>
        <div className="relative">
          <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
            <Search
              className={cn(
                'h-4 w-4 transition-all duration-300',
                isFocused ? 'text-primary scale-110' : 'text-text-muted/50',
                size === 'sm' ? 'h-4 w-4' : size === 'md' ? 'h-5 w-5' : 'h-6 w-6'
              )}
            />
          </div>
          <input
            ref={ref}
            type="text"
            placeholder={placeholder}
            value={value}
            className={cn(
              'search-input-custom',
              'relative w-full text-sm font-medium',
              'disabled:cursor-not-allowed disabled:opacity-50',
              'placeholder:text-text-muted/40',
              sizes[size],
              showClear && hasValue && 'pe-10',
              className
            )}
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

          {showClear && hasValue && (
            <button
              type="button"
              onClick={handleClear}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted/40 hover:text-text-muted/70 transition-colors duration-200 z-10"
            >
              <X
                className={cn(
                  'shrink-0',
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