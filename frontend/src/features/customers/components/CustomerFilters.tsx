import { Card, CardContent } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { SearchInput } from '../../../components/ui/search-input';
import { ArrowUpDown, ChevronUp, ChevronDown } from 'lucide-react';

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
  return (
    <Card>
      <CardContent>
        <div style={{ padding: '18px' }}>
          <div className="pf-search-row flex-col md:flex-row">
            <div className="min-w-0 flex-1">
              <SearchInput
                placeholder="ابحث بالاسم أو رقم الهاتف..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onClear={onClearSearch}
                size="sm"
                className="w-full md:w-[500px] lg:w-[600px]"
              />
            </div>
            <div className="pf-search-controls flex flex-wrap gap-2">
              <Button 
                variant="secondary"
                onClick={onSortName}
                className="gap-2"
              >
                <ArrowUpDown className="w-4 h-4" />
                ترتيب بالاسم
                {sortConfig.key === 'name' && (
                  sortConfig.direction === 'asc' ? <ChevronUp className="w-3.5 h-3.5" /> : 
                  sortConfig.direction === 'desc' ? <ChevronDown className="w-3.5 h-3.5" /> : null
                )}
              </Button>
              <Button 
                variant="secondary"
                onClick={onSortPurchases}
                className="gap-2"
              >
                <ArrowUpDown className="w-4 h-4" />
                ترتيب بالمشتريات
                {sortConfig.key === 'totalPurchases' && (
                  sortConfig.direction === 'asc' ? <ChevronUp className="w-3.5 h-3.5" /> : 
                  sortConfig.direction === 'desc' ? <ChevronDown className="w-3.5 h-3.5" /> : null
                )}
              </Button>
              <Button 
                variant="secondary"
                onClick={onClearSort}
                className="gap-2"
              >
                مسح الترتيب
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}