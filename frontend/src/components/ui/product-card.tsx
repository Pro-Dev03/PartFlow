import { type HTMLAttributes } from 'react';
import { cn } from '../../utils';
import { Package } from 'lucide-react';
import { Badge } from './badge';
import { formatPrice } from '../../utils';

export interface ProductCardProps extends HTMLAttributes<HTMLDivElement> {
  product: {
    id: string | number;
    name: string;
    sku?: string;
    barcode?: string;
    sellingPrice?: number;
    selling_price?: number;
    price?: number;
    costPrice?: number;
    cost_price?: number;
    stock?: number;
    stock_quantity?: number;
    quantity?: number;
    images?: string[];
  };
  quickAdd?: boolean;
  onClick?: (product: ProductCardProps['product']) => void;
}

const getStockValue = (p: ProductCardProps['product']): number | undefined => {
  if (p.stock !== undefined && p.stock !== null) return p.stock;
  if (p.stock_quantity !== undefined && p.stock_quantity !== null) return p.stock_quantity;
  if (p.quantity !== undefined && p.quantity !== null) return p.quantity;
  return undefined;
};

const getPriceValue = (p: ProductCardProps['product']): number => {
  if (p.sellingPrice !== undefined && p.sellingPrice !== null) {
    const val = p.sellingPrice;
    return val > 1000 ? val / 100 : val;
  }
  if (p.selling_price !== undefined && p.selling_price !== null) {
    const val = p.selling_price;
    return val > 1000 ? val / 100 : val;
  }
  if (p.price !== undefined && p.price !== null) {
    const val = p.price;
    return val > 1000 ? val / 100 : val;
  }
  return 0;
};

const getCostPriceValue = (p: ProductCardProps['product']): number | undefined => {
  if (p.costPrice !== undefined && p.costPrice !== null) {
    const val = p.costPrice;
    return val > 1000 ? val / 100 : val;
  }
  if (p.cost_price !== undefined && p.cost_price !== null) {
    const val = p.cost_price;
    return val > 1000 ? val / 100 : val;
  }
  return undefined;
};

function ProductCard({
  product,
  quickAdd = false,
  onClick,
  className,
  ...props
}: ProductCardProps) {
  const stock = getStockValue(product);
  const price = getPriceValue(product);
  const costPrice = getCostPriceValue(product);
  const sku = product.sku ?? product.barcode ?? '';

  const getStockStatus = () => {
    if (stock === undefined) {
      return <Badge variant="outline" size="sm">متاح</Badge>;
    }
    if (stock === 0) {
      return <Badge variant="danger" size="sm">نفد المخزون</Badge>;
    }
    if (stock <= 10) {
      return <Badge variant="warning" size="sm">متبقي {stock}</Badge>;
    }
    return <Badge variant="success" size="sm">متوفر</Badge>;
  };

  const handleClick = () => {
    if (onClick) onClick(product);
  };

  return (
    <div
      onClick={handleClick}
      className={cn(
        'group relative cursor-pointer rounded-xl border border-border bg-bg-surface',
        'transition-all duration-200 ease-out',
        'hover:border-primary hover:shadow-[0_4px_12px_rgba(99,102,241,0.12)]',
        'active:scale-[0.99]',
        className
      )}
      {...props}
    >
      <div className="p-3.5">
        {quickAdd ? (
          <div className="text-center">
            <div
              className="truncate text-sm font-semibold text-text-primary"
            >
              {product.name}
            </div>
            <div className="mt-1.5 text-center text-base font-bold text-primary">
              {formatPrice(price)}
            </div>
          </div>
        ) : (
          <>
            <div className="flex items-center gap-3">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-bg-surface-elevated border border-border">
                <Package className="h-5 w-5 text-text-tertiary" />
              </div>
              <div className="min-w-0 flex-1">
                <div className="truncate text-sm font-semibold text-text-primary">
                  {product.name}
                </div>
                {sku && (
                  <div className="truncate text-xs text-text-secondary">
                    {sku}
                  </div>
                )}
              </div>
            </div>

            <div className="mt-3 flex items-center justify-between">
              <div>
                <div className="text-lg font-bold text-primary">
                  {formatPrice(price)}
                </div>
                {costPrice !== undefined && costPrice > 0 && (
                  <div className="text-xs text-text-tertiary">
                    ت: {formatPrice(costPrice)}
                  </div>
                )}
              </div>
              {getStockStatus()}
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export { ProductCard };
