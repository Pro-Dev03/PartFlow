import { useState, useRef, useEffect } from 'react';
import { cn } from '../../utils';
import { Filter, ChevronDown, X } from 'lucide-react';

export interface FilterOption {
  id: string;
  label: string;
  icon?: React.ReactNode;
}

export interface FilterGroup {
  key: string;
  label: string;
  options: FilterOption[];
  value: string | string[] | null;
  onChange: (value: string | string[]) => void;
  multi?: boolean;
}

export interface FilterDropdownProps {
  label?: string;
  groups: FilterGroup[];
  onClearAll?: () => void;
  className?: string;
  align?: 'start' | 'end';
}

const FilterDropdown = ({
  label = 'فلتر',
  groups,
  onClearAll,
  className,
  align = 'end',
}: FilterDropdownProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  const activeCount = groups.filter(g => g.value).length;
  const hasActive = activeCount > 0;

  const handleGroupChange = (group: FilterGroup, value: string) => {
    if (group.multi) {
      const current = Array.isArray(group.value) ? group.value : [];
      const newValue = current.includes(value)
        ? current.filter(v => v !== value)
        : [...current, value];
      group.onChange(newValue.length ? newValue : null);
    } else {
      group.onChange(value);
    }
  };

  const isOptionActive = (group: FilterGroup, optionId: string): boolean => {
    if (group.multi) {
      return Array.isArray(group.value) && group.value.includes(optionId);
    }
    return group.value === optionId;
  };

  return (
    <div className={cn('relative inline-block', className)} ref={dropdownRef}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="true"
        className={cn(
          'inline-flex h-9 items-center gap-2 rounded-lg border px-3.5 text-sm font-medium',
          'border-border bg-transparent text-text-secondary',
          'transition-all duration-200 ease-out',
          'hover:border-primary hover:text-text-primary hover:bg-bg-surface-elevated',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30',
          isOpen && 'border-primary bg-bg-surface-elevated text-text-primary'
        )}
      >
        <Filter className="h-4 w-4 shrink-0" />
        <span>{label}</span>
        {hasActive && (
          <span
            className="inline-flex h-5 min-w-[20px] items-center justify-center rounded-full bg-primary/10 px-1.5 text-xs font-semibold text-primary"
            aria-label={`${activeCount} فلاتر نشطة`}
          >
            {activeCount}
          </span>
        )}
        <ChevronDown
          className={cn(
            'h-3.5 w-3.5 shrink-0 transition-transform duration-200',
            isOpen && 'rotate-180'
          )}
        />
      </button>

      {isOpen && (
        <div
          className={cn(
            'absolute z-10 mt-2 w-64 origin-top-right rounded-xl border border-border bg-bg-surface',
            'shadow-[0_8px_32px_rgba(0,0,0,0.25)]',
            'animate-in slide-in-from-top-2 duration-200 ease-out',
            align === 'end' ? 'right-0' : 'left-0'
          )}
        >
          <div className="p-3">
            {groups.map((group) => (
              <div key={group.key} className="mb-3 last:mb-0">
                <div className="mb-2 text-xs font-semibold text-text-secondary">
                  {group.label}
                </div>
                <div className="flex flex-wrap gap-2">
                  {group.options.map((option) => {
                    const isActive = isOptionActive(group, option.id);
                    return (
                      <button
                        key={option.id}
                        type="button"
                        onClick={() => handleGroupChange(group, option.id)}
                        className={cn(
                          'inline-flex items-center gap-1.5 rounded-lg border px-3 py-1.5',
                          'text-xs font-medium transition-all duration-200 ease-out',
                          'cursor-pointer',
                          isActive
                            ? 'border-primary bg-primary/10 text-primary'
                            : 'border-border text-text-secondary hover:border-primary hover:bg-bg-surface hover:text-text-primary'
                        )}
                      >
                        {option.icon}
                        {option.label}
                      </button>
                    );
                  })}
                </div>
              </div>
            ))}

            {onClearAll && hasActive && (
              <button
                type="button"
                onClick={() => {
                  onClearAll();
                  groups.forEach(g => g.onChange(g.multi ? [] : null));
                }}
                className={cn(
                  'mt-2 flex w-full items-center justify-center gap-1.5 rounded-lg border',
                  'border-border bg-bg-surface-elevated px-3 py-2 text-sm font-semibold',
                  'text-text-secondary transition-all duration-200',
                  'hover:border-danger hover:bg-danger/5 hover:text-danger cursor-pointer'
                )}
              >
                <X className="h-4 w-4" />
                <span>مسح الكل</span>
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export { FilterDropdown };
