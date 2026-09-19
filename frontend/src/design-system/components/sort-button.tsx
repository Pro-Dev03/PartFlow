import type { ComponentType } from 'react';
import { ArrowUpDown, ChevronDown, ChevronUp } from 'lucide-react';
import { Button, type ButtonProps } from './button';
import { cn } from '../../utils';

type SortDirection = 'asc' | 'desc' | null;

interface SortButtonProps extends Omit<ButtonProps, 'children'> {
  label: string;
  direction?: SortDirection;
  active?: boolean;
  leadingIcon?: ComponentType<{ className?: string }>;
}

export function SortButton({
  label,
  direction = null,
  active = false,
  leadingIcon: LeadingIcon,
  className,
  ...props
}: SortButtonProps) {
  const DirectionIcon = direction === 'asc' ? ChevronUp : direction === 'desc' ? ChevronDown : ArrowUpDown;

  return (
    <Button
      {...props}
      variant={active ? 'primary' : 'outline'}
      className={cn(
        'h-[var(--button-height-sm)] min-w-[142px] justify-between rounded-[8px] px-[var(--button-padding-sm)] text-[var(--button-font-size-sm)]',
        'transition-all duration-200 hover:-translate-y-px hover:shadow-[0_6px_14px_rgba(37,99,235,0.1)]',
        active && 'shadow-[0_7px_16px_rgba(37,99,235,0.2)]',
        className
      )}
      aria-pressed={active}
      title={`ترتيب حسب ${label}`}
    >
      <span className="flex items-center gap-2">
        {LeadingIcon ? <LeadingIcon className="h-4 w-4" /> : <ArrowUpDown className="h-4 w-4" />}
        <span>{label}</span>
      </span>
      <DirectionIcon className={cn('h-4 w-4', !active && 'opacity-60')} />
    </Button>
  );
}
