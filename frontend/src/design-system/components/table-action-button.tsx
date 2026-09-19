import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { Button } from './button';

interface TableActionButtonProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'aria-label' | 'title' | 'children'> {
  label: string;
  icon: ReactNode;
  variant?: 'ghost' | 'secondary' | 'danger' | 'success' | 'warning' | 'info';
}

export function TableActionButton({ label, icon, variant = 'ghost', className, ...props }: TableActionButtonProps) {
  return (
    <Button
      {...props}
      type={props.type ?? 'button'}
      variant={variant}
      size="icon"
      tableAction
      className={className}
      aria-label={label}
      title={label}
    >
      {icon}
    </Button>
  );
}
