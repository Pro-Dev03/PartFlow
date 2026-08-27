import { Card, CardContent } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { SearchInput } from '../../../components/ui/search-input';
import { Select } from '../../../components/ui/select';
import { Input } from '../../../components/ui/input';
import { getButtonSize } from '../../../config/button-sizes';
import { cn } from '../../../utils';
import { Filter, ArrowUpDown, ChevronUp, ChevronDown, Zap } from 'lucide-react';
import { FilterConfig, SortConfig } from '../types/inventory.types';
import { useQuery } from '@tanstack/react-query';
import { suppliersApi, categoriesApi } from '../../../services/api/endpoints';
import { useState } from 'react';

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
  // Fetch suppliers for filter
  const { data: suppliersData } = useQuery({
    queryKey: ['suppliers'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 100 }),
  });

  // Fetch categories for filter
  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });

  const suppliers = (suppliersData?.data?.suppliers as any[]) || [];
  const categories = (categoriesData?.data as any[]) || [];
  
  // Advanced filters state
  const [showAdvancedFilters, setShowAdvancedFilters] = useState(false);
  const [purchaseDateFrom, setPurchaseDateFrom] = useState('');
  const [purchaseDateTo, setPurchaseDateTo] = useState('');
  const [minPurchaseCost, setMinPurchaseCost] = useState('');
  const [maxPurchaseCost, setMaxPurchaseCost] = useState('');

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

  const handleSupplierFilter = (supplierId: string) => {
    if (supplierId === '') {
      setFilters(filters.filter(f => f.key !== 'supplier_id'));
    } else {
      const existingSupplierFilter = filters.find(f => f.key === 'supplier_id');
      if (existingSupplierFilter) {
        setFilters(filters.map(f => f.key === 'supplier_id' ? { key: 'supplier_id', value: supplierId } : f));
      } else {
        setFilters([...filters, { key: 'supplier_id', value: supplierId }]);
      }
    }
  };

  const handleCategoryFilter = (categoryId: string) => {
    if (categoryId === '') {
      setFilters(filters.filter(f => f.key !== 'category_id'));
    } else {
      const existingCategoryFilter = filters.find(f => f.key === 'category_id');
      if (existingCategoryFilter) {
        setFilters(filters.map(f => f.key === 'category_id' ? { key: 'category_id', value: categoryId } : f));
      } else {
        setFilters([...filters, { key: 'category_id', value: categoryId }]);
      }
    }
  };

  const handleAdvancedFilters = () => {
    const newFilters = [...filters];
    
    // Remove existing advanced filters
    const existingKeys = ['purchase_date_from', 'purchase_date_to', 'min_purchase_cost', 'max_purchase_cost'];
    existingKeys.forEach(key => {
      const index = newFilters.findIndex(f => f.key === key);
      if (index !== -1) {
        newFilters.splice(index, 1);
      }
    });
    
    // Add new advanced filters if they have values
    if (purchaseDateFrom) {
      newFilters.push({ key: 'purchase_date_from', value: purchaseDateFrom });
    }
    if (purchaseDateTo) {
      newFilters.push({ key: 'purchase_date_to', value: purchaseDateTo });
    }
    if (minPurchaseCost) {
      newFilters.push({ key: 'min_purchase_cost', value: minPurchaseCost });
    }
    if (maxPurchaseCost) {
      newFilters.push({ key: 'max_purchase_cost', value: maxPurchaseCost });
    }
    
    setFilters(newFilters);
  };

  return (
    <Card>
      <CardContent>
        <div style={{ padding: '18px' }}>
          <div className={cn(
            "flex flex-col gap-md",
            isMobile ? "" : "md:flex-row"
          )}>
            <div className="min-w-0 flex-1">
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
              "pf-search-controls flex shrink-0 gap-2",
              isMobile ? "flex-wrap" : ""
            )}>
              {/* Category Filter */}
              <Select
                value={filters.find(f => f.key === 'category_id')?.value || ''}
                onChange={(e) => handleCategoryFilter(e.target.value)}
                options={[
                  { value: '', label: 'كل التصنيفات' },
                  ...categories.map((c) => ({ value: c.id, label: c.name })),
                ]}
                className={cn(isMobile ? "flex-1" : "w-48")}
                size="sm"
              />

              {/* Supplier Filter */}
              <Select
                value={filters.find(f => f.key === 'supplier_id')?.value || ''}
                onChange={(e) => handleSupplierFilter(e.target.value)}
                options={[
                  { value: '', label: 'كل الموردين' },
                  ...suppliers.map((s) => ({ value: s.id, label: s.name })),
                ]}
                className={cn(isMobile ? "flex-1" : "w-48")}
                size="sm"
              />
              
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
                {filters.some(f => f.key === 'condition' || f.key === 'category_id') && (
                  <span className="text-xs" style={{
                    background: 'rgba(34, 211, 238, 0.2)',
                    color: 'var(--color-primary)',
                    padding: '2px 6px',
                    borderRadius: '4px',
                    fontWeight: '600'
                  }}>{filters.filter(f => f.key === 'condition' || f.key === 'category_id').length}</span>
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
              {filters.length > 0 && (
                <Button
                  variant="secondary"
                  size={getButtonSize('inventory', 'headerActions')}
                  onClick={() => setFilters([])}
                  className={cn("gap-2", isMobile ? "flex-1" : "")}
                >
                  <Filter className="w-4 h-4" />
                  <span>مسح الفلاتر</span>
                </Button>
              )}
            </div>
          </div>
          
          {/* Advanced Filters */}
          {showAdvancedFilters && (
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mt-4 p-4 bg-surface-elevated/30 rounded-lg">
              <div>
                <label className="block text-xs font-medium text-text-tertiary mb-2">تاريخ الشراء من</label>
                <Input
                  type="date"
                  value={purchaseDateFrom}
                  onChange={(e) => setPurchaseDateFrom(e.target.value)}
                  className="w-full"
                  size="sm"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-text-tertiary mb-2">تاريخ الشراء إلى</label>
                <Input
                  type="date"
                  value={purchaseDateTo}
                  onChange={(e) => setPurchaseDateTo(e.target.value)}
                  className="w-full"
                  size="sm"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-text-tertiary mb-2">أقل سعر شراء</label>
                <Input
                  type="number"
                  placeholder="0"
                  value={minPurchaseCost}
                  onChange={(e) => setMinPurchaseCost(e.target.value)}
                  className="w-full"
                  size="sm"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-text-tertiary mb-2">أعلى سعر شراء</label>
                <Input
                  type="number"
                  placeholder="∞"
                  value={maxPurchaseCost}
                  onChange={(e) => setMaxPurchaseCost(e.target.value)}
                  className="w-full"
                  size="sm"
                />
              </div>
              <div className="md:col-span-4 flex justify-end gap-2 mt-2">
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => {
                    setPurchaseDateFrom('');
                    setPurchaseDateTo('');
                    setMinPurchaseCost('');
                    setMaxPurchaseCost('');
                    setFilters(filters.filter(f => !['purchase_date_from', 'purchase_date_to', 'min_purchase_cost', 'max_purchase_cost'].includes(f.key)));
                  }}
                >
                  مسح
                </Button>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={handleAdvancedFilters}
                >
                  تطبيق الفلاتر
                </Button>
              </div>
            </div>
          )}
          
          {/* Advanced Filters Toggle */}
          <div className="mt-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowAdvancedFilters(!showAdvancedFilters)}
              className="text-xs text-text-secondary"
            >
              {showAdvancedFilters ? 'إخفاء الفلاتر المتقدمة' : 'عرض الفلاتر المتقدمة'}
              <Filter className="w-3 h-3 ml-2" />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}