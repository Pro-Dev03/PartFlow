import { Card, CardContent } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { SearchInput } from '../../../components/ui/search-input';
import { getButtonSize } from '../../../config/button-sizes';
import { cn } from '../../../utils';
import { Filter, ArrowUpDown, ChevronUp, ChevronDown, Zap, Layers } from 'lucide-react';
import { FilterConfig, SortConfig } from '../types/inventory.types';

interface InventoryFiltersProps {
  searchQuery: string;
  setSearchQuery: (value: string) => void;
  onClearSearch: () => void;
  filters: FilterConfig[];
  setFilters: (filters: FilterConfig[]) => void;
  sortConfig: SortConfig;
  setSortConfig: (config: SortConfig) => void;
  onRefresh: () => void;
  isMobile: boolean;
}

export function InventoryFilters({
  searchQuery,
  setSearchQuery,
  onClearSearch,
  filters,
  setFilters,
  sortConfig,
  setSortConfig,
  onRefresh,
  isMobile,
}: InventoryFiltersProps) {
  const handleSort = (key: string) => {
    let direction: 'asc' | 'desc' | null = 'asc';
    
    if (sortConfig.key === key) {
      if (sortConfig.direction === 'asc') {
        direction = 'desc';
      } else if (sortConfig.direction === 'desc') {
        direction = null;
      }
    }
    
    setSortConfig({ key, direction });
  };

  return (
    <Card>
      <CardContent>
        <div style={{ padding: '18px' }}>
          <div className={cn(
            "flex flex-col gap-md",
            isMobile ? "" : "md:flex-row"
          )}>
            <div className="flex-1">
              <SearchInput
                placeholder="بحث"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onClear={onClearSearch}
                size="sm"
                className={cn(isMobile ? "w-full" : "w-full md:w-[500px] lg:w-[600px]")}
              />
            </div>
            <div className={cn(
              "flex gap-2",
              isMobile ? "flex-wrap" : ""
            )}>
              <Button
                variant={filters.some(f => f.key === 'condition' && f.value === 'USED') ? 'primary' : 'secondary'}
                size={getButtonSize('inventory', 'headerActions')}
                onClick={() => {
                  if (filters.some(f => f.key === 'condition' && f.value === 'USED')) {
                    setFilters(filters.filter(f => !(f.key === 'condition' && f.value === 'USED')));
                  } else {
                    setFilters([...filters, { key: 'condition', value: 'USED' }]);
                  }
                }}
                className={cn("gap-2", isMobile ? "flex-1" : "")}
              >
                <Layers className="w-4 h-4" />
                <span>قطع مستعملة</span>
              </Button>
              <Button
                variant="secondary"
                size={getButtonSize('inventory', 'headerActions')}
                onClick={() => {
                  if (filters.some(f => f.key === 'condition' && f.value === 'new')) {
                    setFilters(filters.filter(f => !(f.key === 'condition' && f.value === 'new')));
                  } else {
                    setFilters([...filters.filter(f => f.key !== 'condition'), { key: 'condition', value: 'new' }]);
                  }
                }}
                className={cn("gap-2", isMobile ? "flex-1" : "")}
              >
                <Filter className="w-4 h-4" />
                <span>فلتر</span>
                {filters.some(f => f.key === 'condition') && (
                  <span className="text-xs" style={{
                    background: 'rgba(34, 211, 238, 0.2)',
                    color: 'var(--color-primary)',
                    padding: '2px 6px',
                    borderRadius: '4px',
                    fontWeight: '600'
                  }}>{filters.filter(f => f.key === 'condition').length}</span>
                )}
              </Button>
              <Button
                variant="secondary"
                size={getButtonSize('inventory', 'headerActions')}
                onClick={() => handleSort('name')}
                className={cn("gap-2", isMobile ? "flex-1" : "")}
              >
                <ArrowUpDown className="w-4 h-4" />
                <span>ترتيب</span>
                {sortConfig.key === 'name' && (
                  sortConfig.direction === 'asc' ? <ChevronUp className="w-3.5 h-3.5" /> :
                  sortConfig.direction === 'desc' ? <ChevronDown className="w-3.5 h-3.5" /> : null
                )}
              </Button>
              <Button
                variant="secondary"
                size={getButtonSize('inventory', 'headerActions')}
                onClick={onRefresh}
                className={cn("gap-2", isMobile ? "flex-1" : "")}
              >
                <Zap className="w-4 h-4" />
                <span>تحديث</span>
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}