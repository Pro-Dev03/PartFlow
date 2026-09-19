import { Button } from '../../../design-system/components/button';
import { SearchInput } from '../../../design-system/components/search-input';
import { ArrowDownAZ, ArrowDownWideNarrow, X } from 'lucide-react';
import { SortButton } from '../../../design-system/components/sort-button';

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
  const sortOptions = [
    { key: 'name', label: 'الاسم', icon: ArrowDownAZ, onClick: onSortName },
    { key: 'totalPurchases', label: 'إجمالي المشتريات', icon: ArrowDownWideNarrow, onClick: onSortPurchases },
  ];

  return (
    <div className="rounded-[16px] border border-[var(--border-default)] bg-[var(--bg-surface)] shadow-[0_10px_30px_rgba(15,23,42,0.06)]">
      <div className="flex flex-col gap-3 px-4 py-3 md:flex-row md:items-center md:justify-between">
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
          {sortOptions.map((option) => {
            const isActive = sortConfig.key === option.key;
            const OptionIcon = option.icon;
            return (
              <SortButton
                key={option.key}
                onClick={option.onClick}
                label={option.label}
                active={isActive}
                direction={isActive ? sortConfig.direction : null}
                leadingIcon={OptionIcon}
                aria-label={`ترتيب حسب ${option.label}`}
              />
            );
          })}

          {hasActiveSort && (
            <Button variant="ghost" onClick={onClearSort} className="h-9 gap-2 rounded-lg px-3 text-[12px] font-bold text-[var(--text-secondary)] hover:bg-[var(--danger)]/10 hover:text-[var(--danger)]">
              <X className="h-4 w-4" />
              <span>مسح الترتيب</span>
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}