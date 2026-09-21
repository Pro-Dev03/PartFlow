import { Card, CardContent } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { SearchInput } from '../../../design-system/components/search-input';
import { Select } from '../../../design-system/components/select';
import { Input } from '../../../design-system/components/input';
import { SortButton } from '../../../design-system/components/sort-button';
import { getButtonSize } from '../../../config/button-sizes';
import { cn } from '../../../utils';
import { Filter, RefreshCw } from 'lucide-react';
import { FilterConfig, SortConfig } from '../types/inventory.types';
import { useQuery } from '@tanstack/react-query';
import { suppliersApi, categoriesApi } from '../../../services/api/endpoints';
import { useRef, useState } from 'react';

interface InventoryFiltersProps {
  searchQuery: string;
  setSearchQuery: (value: string) => void;
  onClearSearch: () => void;
  onBarcodeScan?: (barcode: string) => Promise<boolean>;
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
  onBarcodeScan,
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

  const supplierPayload = suppliersData?.data;
  const suppliers = (
    Array.isArray(supplierPayload)
      ? supplierPayload
      : Array.isArray(supplierPayload?.suppliers)
        ? supplierPayload.suppliers
        : Array.isArray(suppliersData?.suppliers)
          ? suppliersData.suppliers
          : []
  ) as any[];
  const categories = (categoriesData?.data as any[]) || [];
  
  // Advanced filters state
  const [showAdvancedFilters, setShowAdvancedFilters] = useState(false);
  const [purchaseDateFrom, setPurchaseDateFrom] = useState('');
  const [purchaseDateTo, setPurchaseDateTo] = useState('');
  const [minPurchaseCost, setMinPurchaseCost] = useState('');
  const [maxPurchaseCost, setMaxPurchaseCost] = useState('');
  const [showSortOptions, setShowSortOptions] = useState(false);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const scanStartedAtRef = useRef<number | null>(null);
  const lastScanKeyAtRef = useRef<number | null>(null);

  const handleSearchKeyDown = async (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') {
      const value = searchQuery.trim();
      const startedAt = scanStartedAtRef.current;
      const elapsed = startedAt === null ? Number.POSITIVE_INFINITY : performance.now() - startedAt;
      const isLikelyUsbScan = Boolean(onBarcodeScan && value.length >= 6 && (elapsed <= 700 || event.key === 'Enter'));
      scanStartedAtRef.current = null;
      lastScanKeyAtRef.current = null;

      if (isLikelyUsbScan) {
        event.preventDefault();
        event.stopPropagation();
        await onBarcodeScan(value);
        requestAnimationFrame(() => {
          searchInputRef.current?.focus();
          searchInputRef.current?.select();
        });
      }
      return;
    }

    if (event.key.length === 1) {
      const now = performance.now();
      const gap = lastScanKeyAtRef.current === null ? Number.POSITIVE_INFINITY : now - lastScanKeyAtRef.current;
      if (gap > 100 || scanStartedAtRef.current === null) {
        scanStartedAtRef.current = now;
      }
      lastScanKeyAtRef.current = now;
    }
  };

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

  const activeFilterCount = filters.length;
  const conditionFilterActive = filters.some((filter) => filter.key === 'condition' && filter.value === 'new');
  const sortLabels: Record<string, string> = {
    name: 'الاسم',
    price: 'السعر',
    sellingPrice: 'السعر',
    stock: 'المخزون',
    current_quantity: 'المخزون',
    purchase_date: 'التاريخ',
  };
  const sortLabel = sortLabels[sortConfig.key] || 'الاسم';

  return (
    <Card>
      <CardContent>
        <div className="p-4 md:p-[18px]">
          <div className={cn('flex flex-col gap-3', !isMobile && 'md:flex-row md:items-center')}>
            <div className="min-w-0 flex-1">
              <SearchInput
                ref={searchInputRef}
                placeholder="ابحث عن منتج أو امسح الباركود"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyDown={(event) => { void handleSearchKeyDown(event); }}
                onClear={onClearSearch}
                size="sm"
                className="w-full pf-inventory-search-input"
              />
            </div>

            <div className={cn('flex min-w-0 flex-wrap items-center gap-2', isMobile ? 'w-full' : 'shrink-0')}>
              <Select
                value={filters.find(f => f.key === 'category_id')?.value || ''}
                onChange={(e) => handleCategoryFilter(e.target.value)}
                options={[{ value: '', label: 'كل التصنيفات' }, ...categories.map((c) => ({ value: c.id, label: c.name }))]}
                className={cn('pf-inventory-filter-select', isMobile ? 'min-w-0 flex-1' : 'w-36')}
                size="sm"
              />
              <Select
                value={filters.find(f => f.key === 'supplier_id')?.value || ''}
                onChange={(e) => handleSupplierFilter(e.target.value)}
                options={[{ value: '', label: 'كل التجار' }, ...suppliers.map((s) => ({ value: s.id, label: s.name }))]}
                className={cn('pf-inventory-filter-select', isMobile ? 'min-w-0 flex-1' : 'w-36')}
                size="sm"
              />

              <div className="relative">
                <SortButton
                  onClick={() => setShowSortOptions((open) => !open)}
                  label={`ترتيب: ${sortLabel}`}
                  active={Boolean(sortConfig.key && sortConfig.direction)}
                  direction={sortConfig.key ? sortConfig.direction : null}
                  className="min-w-0 w-auto"
                  aria-label="ترتيب نتائج المخزون"
                />
                {showSortOptions && (
                  <div className="pf-sort-menu absolute end-0 top-full z-30 mt-2 min-w-40 rounded-xl p-1.5" role="menu">
                    {[
                      ['name', 'الاسم'],
                      ['price', 'السعر'],
                      ['stock', 'المخزون'],
                      ['purchase_date', 'التاريخ'],
                    ].map(([key, label]) => (
                      <Button
                        key={key}
                        type="button"
                        variant={sortConfig.key === key ? 'primary' : 'ghost'}
                        size="sm"
                        className="pf-sort-option w-full justify-start text-xs"
                        onClick={() => { handleSort(key); setShowSortOptions(false); }}
                        role="menuitem"
                      >
                        {label}
                      </Button>
                    ))}
                  </div>
                )}
              </div>

              <Button
                type="button"
                variant={showAdvancedFilters || activeFilterCount > 0 ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => setShowAdvancedFilters((open) => !open)}
                className="pf-inventory-filter-trigger gap-1.5"
              >
                <Filter className="h-3.5 w-3.5" />
                الفلاتر
                {activeFilterCount > 0 && <span className="rounded-full bg-white/20 px-1.5 text-[10px]">{activeFilterCount}</span>}
              </Button>

              {activeFilterCount > 0 && (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setFilters([]);
                    setPurchaseDateFrom('');
                    setPurchaseDateTo('');
                    setMinPurchaseCost('');
                    setMaxPurchaseCost('');
                  }}
                  className="gap-1 text-xs text-text-secondary"
                >
                  مسح الفلاتر <span aria-hidden="true">×</span>
                </Button>
              )}

              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={onRefresh}
                aria-label="تحديث المخزون"
                title="تحديث المخزون"
                className="h-8 w-8"
              >
                <RefreshCw className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>

          {showAdvancedFilters && (
            <div className="pf-inventory-advanced-panel mt-3 grid grid-cols-1 gap-3 rounded-xl p-3 sm:grid-cols-2 lg:grid-cols-5">
              <div>
                <label className="mb-1.5 block text-xs font-medium text-text-tertiary">تاريخ الشراء من</label>
                <Input type="date" value={purchaseDateFrom} onChange={(e) => setPurchaseDateFrom(e.target.value)} className="w-full" size="sm" />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium text-text-tertiary">تاريخ الشراء إلى</label>
                <Input type="date" value={purchaseDateTo} onChange={(e) => setPurchaseDateTo(e.target.value)} className="w-full" size="sm" />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium text-text-tertiary">أقل سعر شراء</label>
                <Input type="number" placeholder="0" value={minPurchaseCost} onChange={(e) => setMinPurchaseCost(e.target.value)} className="w-full" size="sm" />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium text-text-tertiary">أعلى سعر شراء</label>
                <Input type="number" placeholder="∞" value={maxPurchaseCost} onChange={(e) => setMaxPurchaseCost(e.target.value)} className="w-full" size="sm" />
              </div>
              <div className="flex items-end gap-2 sm:col-span-2 lg:col-span-1">
                <Button
                  type="button"
                  variant={conditionFilterActive ? 'primary' : 'secondary'}
                  size="sm"
                  onClick={() => {
                    if (conditionFilterActive) setFilters(filters.filter(f => !(f.key === 'condition' && f.value === 'new')));
                    else setFilters([...filters.filter(f => f.key !== 'condition'), { key: 'condition', value: 'new' }]);
                  }}
                  className="w-full"
                >
                  منتجات جديدة
                </Button>
              </div>
              <div className="flex justify-end gap-2 sm:col-span-2 lg:col-span-5">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setPurchaseDateFrom('');
                    setPurchaseDateTo('');
                    setMinPurchaseCost('');
                    setMaxPurchaseCost('');
                    setFilters(filters.filter(f => !['purchase_date_from', 'purchase_date_to', 'min_purchase_cost', 'max_purchase_cost'].includes(f.key)));
                  }}
                >
                  مسح الإضافية
                </Button>
                <Button type="button" variant="primary" size="sm" onClick={handleAdvancedFilters}>تطبيق الفلاتر</Button>
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}