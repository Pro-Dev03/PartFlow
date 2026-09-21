import { ChevronLeft, ChevronRight } from 'lucide-react';
import { Button } from './button';

interface PaginationControlsProps {
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  isLoading?: boolean;
}

export function PaginationControls({ page, pageSize, total, onPageChange, isLoading = false }: PaginationControlsProps) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const first = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const last = Math.min(page * pageSize, total);

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 border-t border-border bg-surface-elevated/60 px-4 py-3 text-xs text-text-muted" dir="rtl">
      <span className="rounded-lg border border-border-subtle bg-surface px-2.5 py-1 font-medium">عرض {first} إلى {last} من {total}</span>
      <div className="flex items-center gap-2">
        <Button type="button" variant="secondary" size="sm" className="rounded-xl px-3 shadow-sm" disabled={page <= 1 || isLoading} onClick={() => onPageChange(Math.max(1, page - 1))}>
          <ChevronRight className="ml-1 h-4 w-4" /> السابق
        </Button>
        <span className="min-w-[104px] rounded-xl border border-primary/15 bg-primary/8 px-3 py-2 text-center font-bold text-primary shadow-sm">صفحة {page} من {totalPages}</span>
        <Button type="button" variant="secondary" size="sm" className="rounded-xl px-3 shadow-sm" disabled={page >= totalPages || isLoading} onClick={() => onPageChange(Math.min(totalPages, page + 1))}>
          التالي <ChevronLeft className="mr-1 h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
