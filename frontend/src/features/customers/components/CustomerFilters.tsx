import { Button } from '../../../components/ui/button';
import { SearchInput } from '../../../components/ui/search-input';
import { ArrowUpDown, ChevronUp, ChevronDown, X } from 'lucide-react';
import { cn } from '../../../utils';

interface CustomerFiltersProps {
  searchQuery: string;
  setSearchQuery: (value: string) => void;
  onClearSearch: () => void;
  sortConfig: { key: string; direction: 'asc' | 'desc' | null };
  onSortName: () => void;
  onSortPurchases: () => void;
  onClearSort: () => void;
}

export function CustomerFilters({
  searchQuery,
  setSearchQuery,
  onClearSearch,
  sortConfig,
  onSortName,
  onSortPurchases,
  onClearSort,
}: CustomerFiltersProps) {
  const hasActiveSort = Boolean(sortConfig.key);

  return (
    <div className="rounded-[12px] border border-border bg-surface shadow-[0_8px_18px_rgba(15,23,42,0.04)]">
      <div className="flex flex-col gap-3 border-b border-border px-4 py-3 md:flex-row md:items-center md:justify-between">
        <div className="min-w-0 flex-1">
          <SearchInput
            placeholder="ابحث بالاسم أو رقم الهاتف..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onClear={onClearSearch}
            size="sm"
            className="w-full"
          />
        </div>

        <div className="flex flex-wrap items-center gap-2 md:justify-end">
          <Button variant="secondary" onClick={onSortName} className="gap-2">
            <ArrowUpDown className="h-4 w-4" />
            <span>ترتيب بالاسم</span>
            {sortConfig.key === 'name' && (
              sortConfig.direction === 'asc' ? <ChevronUp className="h-3.5 w-3.5" /> : sortConfig.direction === 'desc' ? <ChevronDown className="h-3.5 w-3.5" /> : null
            )}
          </Button>

          <Button variant="secondary" onClick={onSortPurchases} className="gap-2">
            <ArrowUpDown className="h-4 w-4" />
            <span>المشتريات</span>
            {sortConfig.key === 'totalPurchases' && (
              sortConfig.direction === 'asc' ? <ChevronUp className="h-3.5 w-3.5" /> : sortConfig.direction === 'desc' ? <ChevronDown className="h-3.5 w-3.5" /> : null
            )}
          </Button>

          {hasActiveSort && (
            <Button variant="ghost" onClick={onClearSort} className="gap-2 text-text-secondary">
              <X className="h-4 w-4" />
              <span>مسح الترتيب</span>
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}