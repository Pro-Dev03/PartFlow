import { useState } from 'react';
import { SearchInput } from '../../../components/ui/search-input';
import { Badge } from '../../../components/ui/badge';
import { Button } from '../../../components/ui/button';
import { ProductCard } from '../../../components/ui/product-card';
import { FilterDropdown } from '../../../components/ui/filter-dropdown';
import { Package } from 'lucide-react';
import { useTranslation } from '../../../hooks/useTranslation';
import { cn } from '../../../utils';
import type { ProductCardProps } from '../../../components/ui/product-card';

type SearchProduct = ProductCardProps['product'] & {
  category_id?: string;
  sales_count?: number;
};

interface SearchCategory {
  id: string;
  name: string;
}

interface ProductSearchProps {
  products: SearchProduct[];
  searchQuery: string;
  setSearchQuery: (value: string) => void;
  onClearSearch: () => void;
  quickAddMode: boolean;
  onProductClick: (product: SearchProduct) => void;
  selectedCategory: string | null;
  categories: SearchCategory[];
  hasMore?: boolean;
  onLoadMore?: () => void;
}

export function ProductSearch({
  products,
  searchQuery,
  setSearchQuery,
  onClearSearch,
  quickAddMode,
  onProductClick,
  selectedCategory,
  categories,
  hasMore = false,
  onLoadMore,
}: ProductSearchProps) {
  const { t } = useTranslation();

  const [displayedCount, setDisplayedCount] = useState(16);
  const [priceRange, setPriceRange] = useState<'all' | 'low' | 'medium' | 'high'>('all');

  const PRODUCTS_PER_PAGE = 16;

  const getPriceValue = (product: SearchProduct): number => {
    if (product.sellingPrice !== undefined && product.sellingPrice !== null) {
      return product.sellingPrice;
    }
    if (product.selling_price !== undefined && product.selling_price !== null) {
      return product.selling_price;
    }
    if (product.price !== undefined && product.price !== null) {
      return product.price;
    }
    return 0;
  };

  const filteredProducts = (products || [])
    .filter((product) => {
      const matchesSearch =
        product.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        product.sku?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        product.barcode?.toLowerCase().includes(searchQuery.toLowerCase());

      const matchesCategory = !selectedCategory || product.category_id === selectedCategory;

      let matchesPrice = true;
      if (priceRange !== 'all') {
        const price = getPriceValue(product);
        if (priceRange === 'low') matchesPrice = price < 100;
        else if (priceRange === 'medium') matchesPrice = price >= 100 && price < 500;
        else if (priceRange === 'high') matchesPrice = price >= 500;
      }

      return matchesSearch && matchesCategory && matchesPrice;
    })
    .sort((a, b) => {
      const aSales = a.sales_count || 0;
      const bSales = b.sales_count || 0;
      return bSales - aSales;
    });

  const displayedProducts = filteredProducts.slice(0, displayedCount);
  const hasMoreProducts = filteredProducts.length > displayedCount;

  const handleLoadMore = () => {
    setDisplayedCount(prev => prev + PRODUCTS_PER_PAGE);
  };

  const handleSearchChange = (value: string) => {
    setSearchQuery(value);
    setDisplayedCount(PRODUCTS_PER_PAGE);
  };

  const sectionTitle = selectedCategory
    ? `منتجات ${categories?.find((c) => c.id === selectedCategory)?.name || 'الفئة المحددة'}`
    : t('sales.recentProducts');

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-text-primary">
          {sectionTitle}
        </h3>
        <Badge variant="outline" size="sm" className="text-xs">
          عرض {displayedProducts.length} من {filteredProducts.length}
          {filteredProducts.length > 0 && ' ⭐'}
        </Badge>
      </div>

      {selectedCategory && (
        <div className="text-xs font-medium text-text-tertiary">
          ⭐ مرتبة حسب الأكثر مبيعاً
        </div>
      )}

      <div className="pf-search-row flex items-center gap-2">
        <div className="min-w-0 flex-1">
          <SearchInput
            placeholder="بحث بالاسم، الباركود، SKU..."
            value={searchQuery}
            onChange={(e) => handleSearchChange(e.target.value)}
            onClear={onClearSearch}
            size="sm"
            className="w-full pos-product-search-field"
          />
        </div>

        <FilterDropdown
          label="فلتر"
          groups={[
            {
              key: 'price',
              label: 'نطاق السعر',
              value: priceRange,
              onChange: (v) => {
                if (typeof v === 'string' && ['all', 'low', 'medium', 'high'].includes(v)) {
                  setPriceRange(v as typeof priceRange);
                }
                setDisplayedCount(PRODUCTS_PER_PAGE);
              },
              options: [
                { id: 'all', label: 'الكل' },
                { id: 'low', label: 'أقل من 100' },
                { id: 'medium', label: '100–500' },
                { id: 'high', label: '500+' },
              ],
            },
          ]}
        />
      </div>

      {displayedProducts.length > 0 ? (
        <div
          className={cn(
            'grid gap-3',
            quickAddMode
              ? 'grid-cols-2 sm:grid-cols-3 md:grid-cols-4'
              : 'grid-cols-2 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-5'
          )}
        >
          {displayedProducts.map((product) => (
            <ProductCard
              key={product.id}
              product={{
                ...product,
                costPrice: product.cost_price || product.costPrice || 0,
              }}
              quickAdd={quickAddMode}
              onClick={onProductClick}
            />
          ))}
        </div>
      ) : (
        <div className="py-12 text-center">
          <Package className="mx-auto mb-4 h-12 w-12 text-text-tertiary/30" />
          <p className="mb-1 text-sm text-text-secondary">لا توجد منتجات مطابقة</p>
          <p className="text-xs text-text-tertiary">جرب تغيير معايير البحث أو الفلتر</p>
        </div>
      )}

      {(hasMoreProducts || hasMore) && (
        <div className="flex justify-center">
          <Button
            variant="secondary"
            size="sm"
            onClick={hasMore ? onLoadMore : handleLoadMore}
            className="rounded-lg px-6 text-sm font-semibold"
          >
            تحميل المزيد ({Math.min(PRODUCTS_PER_PAGE, filteredProducts.length - displayedCount)})
          </Button>
        </div>
      )}
    </div>
  );
}
