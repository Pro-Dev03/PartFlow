import { cn } from '../../utils';

interface TableProps {
  className?: string;
  children: React.ReactNode;
}

const Table = ({ className, children }: TableProps) => (
  <div
    className={cn(
      'pf-data-table w-full overflow-x-auto rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-surface)]'
    )}
    role="region"
    aria-label="جدول البيانات"
  >
    <table className={cn('w-full min-w-full border-collapse text-sm', className)}>{children}</table>
  </div>
);

const TableHeader = ({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) => (
  <thead
    className={cn(
      'pf-data-table-header bg-[var(--bg-surface-muted)]',
      className
    )}
    {...props}
  />
);

const TableBody = ({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) => (
  <tbody className={cn('bg-[var(--bg-surface)]', className)} {...props} />
);

const TableFooter = ({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) => (
  <tfoot className={cn('border-t border-[var(--border-subtle)] bg-[var(--bg-surface-muted)] font-medium', className)} {...props} />
);

const TableRow = ({ className, ...props }: React.HTMLAttributes<HTMLTableRowElement>) => (
  <tr
    className={cn(
      'border-b border-[var(--border-subtle)] last:border-b-0 transition-colors duration-150',
      'pf-data-table-row hover:bg-[var(--bg-surface-muted)]',
      className
    )}
    {...props}
  />
);

const TableHead = ({ className, ...props }: React.HTMLAttributes<HTMLTableCellElement>) => (
  <th
    className={cn(
      'h-11 px-3 text-start align-middle text-[11px] font-semibold tracking-[0.02em] text-[var(--text-secondary)]',
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
      'px-3 py-2.5 align-middle text-[13px] text-[var(--text-primary)] transition-colors duration-150',
      '[&:has([role=checkbox])]:pr-0',
      className
    )}
    {...props}
  />
);

const TableCaption = ({ className, ...props }: React.HTMLAttributes<HTMLTableCaptionElement>) => (
  <caption className={cn('mt-3 text-xs text-[var(--text-muted)]', className)} {...props} />
);

export { Table, TableHeader, TableBody, TableFooter, TableHead, TableRow, TableCell, TableCaption };