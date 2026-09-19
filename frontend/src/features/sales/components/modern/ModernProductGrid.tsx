import { Info, Package, Plus } from 'lucide-react';
import { cn } from '../../../../utils';
import type { ProductCardProps } from '../../../../design-system/components/product-card';
import { getLocalProductImage } from '../../../../services/localProductImages';

type SearchProduct = ProductCardProps['product'] & {
  category_id?: string;
  category_name?: string;
  category_image_url?: string;
  image_url?: string;
  sales_count?: number;
};

interface ModernProductGridProps {
  products: SearchProduct[];
  onProductClick: (product: SearchProduct) => void;
  currentPage?: number;
  totalPages?: number;
  onPreviousPage?: () => void;
  onNextPage?: () => void;
  viewMode?: 'cards' | 'list';
  isLoading?: boolean;
  taxRate?: number;
  taxExempt?: boolean;
  showDetails: boolean;
  onToggleDetails: () => void;
  onAddProduct?: () => void;
  hasSearch?: boolean;
}

export function ModernProductGrid({
  products,
  onProductClick,
  currentPage = 1,
  totalPages = 1,
  onPreviousPage,
  onNextPage,
  viewMode = 'cards',
  isLoading = false,
  taxRate = 0,
  taxExempt = false,
  showDetails,
  onToggleDetails,
  onAddProduct,
  hasSearch = false,
}: ModernProductGridProps) {
  if (isLoading) {
    return (
      <div className="pos-modern-products-grid">
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="pos-modern-product-card skeleton">
            <div className="skeleton-image" />
            <div className="skeleton-text" />
            <div className="skeleton-text short" />
          </div>
        ))}
      </div>
    );
  }

  if (products.length === 0) {
    return (
      <div className="pos-modern-products-empty">
        <Package className="empty-icon" />
        <p className="empty-title">{hasSearch ? 'لا توجد نتائج مطابقة' : 'لا توجد منتجات في المخزون'}</p>
        <p className="empty-subtitle">
          {hasSearch ? 'جرّب اسمًا أو باركودًا مختلفًا' : 'أضف أول منتج لبدء البيع من نقطة البيع'}
        </p>
        {!hasSearch && onAddProduct && (
          <button type="button" className="pos-empty-add-product" onClick={onAddProduct}>
            <Plus className="h-4 w-4" />
            <span>إضافة منتج</span>
          </button>
        )}
      </div>
    );
  }

  return (
    <>
      <div className="pos-product-view-toolbar">
        <button
          type="button"
          className={`pos-product-details-toggle ${showDetails ? 'active' : ''}`}
          onClick={onToggleDetails}
          aria-pressed={showDetails}
          title={showDetails ? 'إخفاء تفاصيل المنتج' : 'عرض تفاصيل المنتج'}
        >
          <Info className="h-4 w-4" aria-hidden="true" />
          <span>{showDetails ? 'إخفاء التفاصيل' : 'عرض التفاصيل'}</span>
        </button>
      </div>
      <div className={viewMode === 'list' ? 'pos-modern-products-list' : 'pos-modern-products-grid'}>
        {products.map((product) => (
          viewMode === 'list' ? (
            <ModernProductListItem key={product.id} product={product} taxRate={taxRate} taxExempt={taxExempt} showDetails={showDetails} onClick={() => onProductClick(product)} />
          ) : (
            <ModernProductCard key={product.id} product={product} taxRate={taxRate} taxExempt={taxExempt} showDetails={showDetails} onClick={() => onProductClick(product)} />
          )
        ))}
      </div>
      {totalPages > 1 && (
        <div className="pos-modern-pagination" dir="rtl" aria-label="التنقل بين صفحات المنتجات">
          <button
            type="button"
            className="pos-modern-pagination-btn"
            onClick={onNextPage}
            disabled={currentPage >= totalPages || isLoading}
            aria-label="الصفحة التالية"
          >
            التالي
          </button>
          <span className="pos-modern-pagination-status" aria-live="polite">
            صفحة {currentPage} من {totalPages}
          </span>
          <button
            type="button"
            className="pos-modern-pagination-btn"
            onClick={onPreviousPage}
            disabled={currentPage <= 1 || isLoading}
            aria-label="الصفحة السابقة"
          >
            السابق
          </button>
        </div>
      )}
    </>
  );
}

function ModernProductListItem({ product, taxRate, taxExempt, showDetails, onClick }: ModernProductCardProps) {
  const basePrice = product.sellingPrice ?? product.selling_price ?? product.price ?? 0;
  const price = taxExempt ? basePrice : Math.round(basePrice * (1 + taxRate / 100) * 100) / 100;
  const stock = product.stock ?? 0;
  const imageUrl = product.image_url || getLocalProductImage(String(product.id)) || product.category_image_url;

  return (
    <button type="button" className="pos-modern-product-list-item" onClick={onClick} disabled={stock <= 0}>
      <span className="pos-modern-product-list-main min-w-0">
        <span className="pos-modern-product-list-thumb">
          {imageUrl ? <img src={imageUrl} alt="" /> : <Package className="h-4 w-4" aria-hidden="true" />}
        </span>
        <span className="pos-modern-product-list-copy min-w-0">
          <span className="pos-modern-product-list-name block">{product.name}</span>
          <span className="pos-modern-product-list-category block">
            {product.category_name || 'نوع غير محدد'}
          </span>
          {showDetails && (product.sku || product.barcode) && (
            <span className="pos-modern-product-list-identifiers">
              {product.sku && <span>SKU: {product.sku}</span>}
              {product.barcode && <span>باركود: {product.barcode}</span>}
            </span>
          )}
        </span>
      </span>
      <span className="pos-modern-product-list-stock">{stock > 0 ? `${stock} متوفر` : 'نفد المخزون'}</span>
      <span className="pos-modern-product-list-price">₪{price.toLocaleString()}</span>
      <Plus className="h-4 w-4" aria-hidden="true" />
    </button>
  );
}

interface ModernProductCardProps {
  product: SearchProduct;
  taxRate: number;
  taxExempt: boolean;
  showDetails?: boolean;
  onClick: () => void;
}

function ModernProductCard({ product, taxRate, taxExempt, showDetails = false, onClick }: ModernProductCardProps) {
  const basePrice =
    product.sellingPrice ??
    product.selling_price ??
    product.price ??
    0;
  const price = taxExempt ? basePrice : Math.round(basePrice * (1 + taxRate / 100) * 100) / 100;

  const stock = product.stock ?? 0;
  const isLowStock = stock > 0 && stock <= 3;
  const isOutOfStock = stock <= 0;
  const stockStatus = isOutOfStock ? 'نفد المخزون' : isLowStock ? 'قليل' : 'متوفر';
  const imageUrl = product.image_url || getLocalProductImage(String(product.id)) || product.category_image_url;

  return (
    <button
      className={cn(
        'pos-modern-product-card',
        isOutOfStock && 'out-of-stock',
        isLowStock && 'low-stock'
      )}
      onClick={onClick}
      disabled={isOutOfStock}
    >
      <div className="product-card-image">
        {imageUrl ? (
          <img src={imageUrl} alt={product.name} />
        ) : (
          <div className="product-card-placeholder">
            <Package />
          </div>
        )}
        <div className="product-card-add">
          <Plus className="w-5 h-5" />
        </div>
      </div>
      <div className="product-card-info">
        <h3 className="product-card-name">{product.name}</h3>
        <div className="product-card-meta">
          {product.condition && (
            <span className={cn('product-card-condition', product.condition)}>
              {product.condition === 'new' ? 'جديد' : 'مستعمل'}
            </span>
          )}
          {stock > 0 && (
            <span
              className={cn(
                'product-card-stock',
                isLowStock && 'low'
              )}
            >
              {stock} متوفر
            </span>
          )}
          <span className={cn('product-card-status', isOutOfStock ? 'danger' : isLowStock ? 'warning' : 'success')}>
            <span className="product-card-status-dot" aria-hidden="true" />
            {stockStatus}
          </span>
        </div>
        {showDetails && (product.sku || product.barcode) && (
          <div className="product-card-identifiers" title="معرّفات المنتج">
            {[product.sku && `SKU: ${product.sku}`, product.barcode && `باركود: ${product.barcode}`]
              .filter(Boolean)
              .join(' · ')}
          </div>
        )}
        <div className="product-card-price">
          <span className="price-value">₪{price.toLocaleString()}</span>
        </div>
      </div>
    </button>
  );
}
