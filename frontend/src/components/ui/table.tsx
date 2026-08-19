import { cn } from '../../lib/utils';

interface TableProps {
  className?: string;
  children: React.ReactNode;
}

const Table = ({ className, children }: TableProps) => (
  <div className="w-full overflow-auto" role="region" aria-label="جدول البيانات">
    <table className={cn('w-full caption-bottom text-small', className)}>{children}</table>
  </div>
);

const TableHeader = ({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) => (
  <thead className={cn('[&_tr]:border-b', className)} {...props} />
);

const TableBody = ({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) => (
  <tbody className={cn('[&_tr:last-child]:border-0', className)} {...props} />
);

const TableFooter = ({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) => (
  <tfoot className={cn('border-t bg-surface-2 font-medium [&_tr]:last:border-b-0', className)} {...props} />
);

const TableRow = ({ className, ...props }: React.HTMLAttributes<HTMLTableRowElement>) => (
  <tr className={cn('border-b border-border transition-colors hover:bg-surface-2', className)} {...props} />
);

const TableHead = ({ className, ...props }: React.HTMLAttributes<HTMLTableCellElement>) => (
  <th 
    className={cn('h-12 px-4 text-start align-middle font-medium text-text [&:has([role=checkbox])]:pr-0', className)} 
    scope="col"
    {...props} 
  />
);

const TableCell = ({ className, ...props }: React.TdHTMLAttributes<HTMLTableCellElement>) => (
  <td className={cn('p-4 align-middle text-text-muted [&:has([role=checkbox])]:pr-0', className)} {...props} />
);

const TableCaption = ({ className, ...props }: React.HTMLAttributes<HTMLTableCaptionElement>) => (
  <caption className={cn('mt-4 text-small text-text-muted', className)} {...props} />
);

export { Table, TableHeader, TableBody, TableFooter, TableHead, TableRow, TableCell, TableCaption };