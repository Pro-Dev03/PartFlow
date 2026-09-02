import { Package, Plus } from 'lucide-react';
import { cn } from '../../../../utils';
import type { ProductCardProps } from '../../../../components/ui/product-card';

type SearchProduct = ProductCardProps['product'] & {
  category_id?: string;
  sales_count?: number;
};

interface ModernProductGridProps {
  products: SearchProduct[];
  onProductClick: (product: SearchProduct) => void;
  hasMore?: boolean;
  onLoadMore?: () => void;
  isLoading?: boolean;
}

export function ModernProductGrid({
  products,
  onProductClick,
  hasMore = false,
  onLoadMore,
  isLoading = false,
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
        <p className="empty-title">لا توجد منتجات</p>
        <p className="empty-subtitle">ابدأ بالبحث أو امسح الباركود</p>
      </div>
    );
  }

  return (
    <>
      <div className="pos-modern-products-grid">
        {products.map((product) => (
          <ModernProductCard
            key={product.id}
            product={product}
            onClick={() => onProductClick(product)}
          />
        ))}
      </div>
      {hasMore && onLoadMore && (
        <div className="pos-modern-load-more">
          <button className="pos-modern-load-more-btn" onClick={onLoadMore}>
            تحميل المزيد
          </button>
        </div>
      )}
    </>
  );
}

interface ModernProductCardProps {
  product: SearchProduct;
  onClick: () => void;
}

function ModernProductCard({ product, onClick }: ModernProductCardProps) {
  const price =
    product.sellingPrice ??
    product.selling_price ??
    product.price ??
    0;

  const stock = product.stock ?? 0;
  const isLowStock = stock > 0 && stock <= 3;
  const isOutOfStock = stock <= 0;

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
        {product.image_url ? (
          <img src={product.image_url} alt={product.name} />
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
        </div>
        <div className="product-card-price">
          <span className="price-value">₪{price.toLocaleString()}</span>
        </div>
      </div>
    </button>
  );
}
