import { useState } from 'react';
import { AlertTriangle, ChevronLeft, ChevronRight, Eye, Package, PackageX } from 'lucide-react';
import { Badge } from '../../../components/ui/badge';
import { Button } from '../../../components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { formatPrice, normalizeCurrencyValue } from '../../../utils';
import { getLocalProductImage } from '../../../services/localProductImages';
import { getCategoryImage } from '../../../services/localCategoryImages';
import type { Product } from '../types/inventory.types';

interface StockAlertCardsProps {
  products: Product[];
  inventoryStockMap?: Map<string, number>;
  onViewProduct: (product: Product) => void;
  layoutMode: 'cards' | 'table';
}

function getStockValue(product: Product, inventoryStockMap?: Map<string, number>): number {
  const mappedStock = inventoryStockMap?.get(String(product.id));
  if (mappedStock !== undefined) return mappedStock;

  const value = Number(product.stock ?? 0);
  return Number.isFinite(value) && value > 0 ? value : 0;
}

function AlertCard({
  title,
  description,
  products,
  inventoryStockMap,
  icon: Icon,
  tone,
  onViewProduct,
  layoutMode,
}: {
  title: string;
  description: string;
  products: Product[];
  inventoryStockMap?: Map<string, number>;
  icon: typeof AlertTriangle;
  tone: 'danger' | 'warning';
  onViewProduct: (product: Product) => void;
  layoutMode: 'cards' | 'table';
}) {
  const pageSize = 3;
  const [page, setPage] = useState(1);
  const pageCount = Math.max(1, Math.ceil(products.length / pageSize));
  const visibleProducts = products.slice((page - 1) * pageSize, page * pageSize);
  const canGoPrevious = page > 1;
  const canGoNext = page < pageCount;

  return (
    <section className="rounded-[12px] border border-border bg-surface shadow-[0_8px_18px_rgba(15,23,42,0.04)]" aria-label={title}>
      <div className="flex items-start gap-3 border-b border-border px-5 py-4">
        <div className={`rounded-lg p-2 ${tone === 'danger' ? 'bg-danger/10 text-danger' : 'bg-warning/10 text-warning'}`}>
          <Icon className="h-5 w-5" />
        </div>
        <div className="min-w-0">
          <h3 className="font-semibold text-text-primary">{title}</h3>
          <p className="mt-1 text-xs text-text-tertiary">{description}</p>
        </div>
        <span className={`ms-auto rounded-full px-2.5 py-1 text-xs font-semibold ${tone === 'danger' ? 'bg-danger/10 text-danger' : 'bg-warning/10 text-warning'}`}>
          {products.length}
        </span>
      </div>

      {products.length === 0 ? (
        <div className="m-4 rounded-lg border border-dashed border-border px-3 py-5 text-center text-sm text-text-tertiary">
          لا توجد منتجات حاليا
        </div>
      ) : layoutMode === 'table' ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>المنتج</TableHead>
              <TableHead>التصنيف</TableHead>
              <TableHead className="text-center">المخزون</TableHead>
              <TableHead className="text-center">السعر</TableHead>
              <TableHead className="text-end">الإجراءات</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {visibleProducts.map((product) => (
              <TableRow key={product.id}>
                <TableCell>
                  <div className="min-w-0">
                    <div className="truncate font-semibold text-text-primary">{product.name || '-'}</div>
                    <div className="mt-0.5 text-[11px] text-text-tertiary">{product.sku || '-'}</div>
                  </div>
                </TableCell>
                <TableCell className="text-text-secondary">
                  {(product as Product & { category_name?: string }).category_name || product.category || 'بدون تصنيف'}
                </TableCell>
                <TableCell className={`text-center font-semibold ${tone === 'danger' ? 'text-danger' : 'text-warning'}`}>
                  {getStockValue(product, inventoryStockMap)}
                </TableCell>
                <TableCell className="text-center font-semibold text-primary">
                  {formatPrice(normalizeCurrencyValue(product.sellingPrice ?? product.price ?? 0))}
                </TableCell>
                <TableCell className="text-end">
                  <Button type="button" variant="ghost" size="icon" onClick={() => onViewProduct(product)} aria-label={`عرض المنتج ${product.name || ''}`} title="عرض المنتج">
                    <Eye className="h-3.5 w-3.5" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      ) : (
        <div className="grid grid-cols-1 gap-3 p-4">
            {visibleProducts.map((product) => (
            <div key={product.id} className="rounded-xl border border-border bg-surface-elevated/25 p-4">
              <div className="mb-3 flex items-center justify-center overflow-hidden rounded-lg bg-surface-muted" style={{ height: '128px', minHeight: '128px' }}>
                {product.image_url || getLocalProductImage(product.id) || (product.category_id ? getCategoryImage(product.category_id) : undefined) ? (
                  <img
                    src={product.image_url || getLocalProductImage(product.id) || (product.category_id ? getCategoryImage(product.category_id) : undefined)}
                    alt={product.name || 'المنتج'}
                    className="max-h-full max-w-full object-contain"
                    style={{ width: '100%', height: '100%' }}
                  />
                ) : <Package className="h-12 w-12 text-text-tertiary" />}
              </div>
              <div className="mb-3 flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="truncate font-semibold text-text-primary">{product.name || '-'}</div>
                  <div className="mt-1 text-[11px] text-text-tertiary">{product.sku || '-'}</div>
                </div>
                <Badge variant={tone === 'danger' ? 'danger' : 'warning'} size="sm">
                  {getStockValue(product, inventoryStockMap)}
                </Badge>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex items-center justify-between gap-2">
                  <span className="text-text-tertiary">التصنيف</span>
                  <span className="font-medium text-text-secondary">{(product as Product & { category_name?: string }).category_name || product.category || 'بدون تصنيف'}</span>
                </div>
                <div className="flex items-center justify-between gap-2">
                  <span className="text-text-tertiary">المخزون</span>
                  <span className={`font-semibold ${tone === 'danger' ? 'text-danger' : 'text-warning'}`}>{getStockValue(product, inventoryStockMap)}</span>
                </div>
                <div className="flex items-center justify-between gap-2">
                  <span className="text-text-tertiary">السعر قبل الضريبة</span>
                  <span className="font-semibold text-primary">{formatPrice(normalizeCurrencyValue(product.sellingPrice ?? product.price ?? 0))}</span>
                </div>
              </div>
              <div className="mt-3 flex justify-end border-t border-border pt-3">
                <Button type="button" variant="ghost" size="sm" onClick={() => onViewProduct(product)}>
                  <Eye className="h-3.5 w-3.5" />
                  عرض المنتج
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
      {products.length > pageSize && (
        <div className="flex items-center justify-between border-t border-border px-4 py-3">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={!canGoPrevious}
            onClick={() => setPage((current) => Math.max(1, current - 1))}
          >
            <ChevronRight className="h-4 w-4" />
            السابق
          </Button>
          <span className="text-xs text-text-tertiary">{page} من {pageCount}</span>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={!canGoNext}
            onClick={() => setPage((current) => Math.min(pageCount, current + 1))}
          >
            التالي
            <ChevronLeft className="h-4 w-4" />
          </Button>
        </div>
      )}
    </section>
  );
}

export function StockAlertCards({ products, inventoryStockMap, onViewProduct, layoutMode }: StockAlertCardsProps) {
  const outOfStockProducts = products.filter((product) => getStockValue(product, inventoryStockMap) === 0);
  const lowStockProducts = products.filter((product) => {
    const stock = getStockValue(product, inventoryStockMap);
    const minimumStock = Math.max(1, Number(product.min_stock_level) || 3);
    return stock > 0 && stock <= minimumStock;
  });

  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-1">
      <AlertCard
        title="المنتجات التي نفدت"
        description="تحتاج إلى إعادة تزويد المخزون"
        products={outOfStockProducts}
        inventoryStockMap={inventoryStockMap}
        icon={PackageX}
        tone="danger"
        onViewProduct={onViewProduct}
        layoutMode={layoutMode}
      />
      <AlertCard
        title="المنتجات ذات الكمية القليلة"
        description="اقتربت من الحد الأدنى للمخزون"
        products={lowStockProducts}
        inventoryStockMap={inventoryStockMap}
        icon={AlertTriangle}
        tone="warning"
        onViewProduct={onViewProduct}
        layoutMode={layoutMode}
      />
    </div>
  );
}
