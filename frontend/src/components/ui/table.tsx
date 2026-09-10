import { cn } from '../../utils';

interface TableProps {
  className?: string;
  children: React.ReactNode;
}

const Table = ({ className, children }: TableProps) => (
  <div
    className={cn(
      'w-full overflow-hidden rounded-[10px] border border-[var(--table-border)] bg-[var(--card-bg)]',
      className
    )}
    role="region"
    aria-label="جدول البيانات"
  >
    <table className={cn('w-full caption-bottom border-collapse text-sm', className)}>{children}</table>
  </div>
);

const TableHeader = ({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) => (
  <thead
    className={cn(
      'bg-[var(--bg-surface-elevated)]',
      className
    )}
    {...props}
  />
);

const TableBody = ({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) => (
  <tbody className={cn('bg-[var(--card-bg)]', className)} {...props} />
);

const TableFooter = ({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) => (
  <tfoot className={cn('border-t border-[var(--table-border)] bg-[rgba(148,163,184,0.03)] font-medium', className)} {...props} />
);

const TableRow = ({ className, ...props }: React.HTMLAttributes<HTMLTableRowElement>) => (
  <tr
    className={cn(
      'border-b border-[var(--border-subtle)] transition-colors duration-150',
      'hover:bg-[var(--bg-surface-elevated)]',
      className
    )}
    {...props}
  />
);

const TableHead = ({ className, ...props }: React.HTMLAttributes<HTMLTableCellElement>) => (
  <th
    className={cn(
      'h-10 px-4 text-start align-middle text-[10px] font-semibold tracking-[0.08em] text-[var(--text-secondary)] uppercase',
      'transition-colors duration-150',
      '[&:has([role=checkbox])]:pr-0',
      className
    )}
    scope="col"
    {...props}
  />
);

const TableCell = ({ className, ...props }: React.TdHTMLAttributes<HTMLTableCellElement>) => (
  <td
    className={cn(
      'px-4 py-2.5 align-middle text-[13px] text-[var(--text-primary)] transition-colors duration-150',
      '[&:has([role=checkbox])]:pr-0',
      className
    )}
    {...props}
  />
);

const TableCaption = ({ className, ...props }: React.HTMLAttributes<HTMLTableCaptionElement>) => (
  <caption className={cn('mt-4 text-xs text-[var(--text-muted)]', className)} {...props} />
);

export { Table, TableHeader, TableBody, TableFooter, TableHead, TableRow, TableCell, TableCaption };