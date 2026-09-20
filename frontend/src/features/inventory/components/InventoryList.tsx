import { useState } from 'react';
import { Badge } from '../../../design-system/components/badge';
import { EmptyState } from '../../../design-system/components/empty-state';
import { Button } from '../../../design-system/components/button';
import { Table, TableHeader, TableBody, TableHead, TableRow, TableCell } from '../../../design-system/components/table';
import { Archive, Package, PackageOpen, Eye, Edit, SlidersHorizontal, Trash2, Inbox, RefreshCw, FileText, Plus, ArrowUpRight, Trash, Copy, PencilLine } from 'lucide-react';
import { ActionMenu } from '../../../design-system/components/action-menu';
import { Product, InventoryItem, ViewMode } from '../types/inventory.types';
import { formatPrice, normalizeCurrencyValue } from '../../../utils';
import { cn } from '../../../utils';
import { PaginationControls } from '../../../design-system/components/pagination-controls';
import { getLocalProductImage } from '../../../services/localProductImages';
import { getCategoryImage } from '../../../services/localCategoryImages';

interface InventoryListProps {
  viewMode: ViewMode;
  filteredProducts: Product[];
  filteredInventoryItems: InventoryItem[];
  productsLoading: boolean;
  inventoryLoading: boolean;
  inventoryStockMap?: Map<string, number>;
  searchQuery: string;
  onViewProduct: (product: Product) => void;
  onAddPurchase: (product: Product) => void;
  onEditProduct: (product: Product) => void;
  onEditMinimumStock: (product: Product) => void;
  onDeleteProduct: (productId: string) => void;
  onArchiveProduct?: (productId: string) => void;
  onDeleteInventoryItem?: (itemId: string) => void;
  onClearSearch: () => void;
  onReorderFromSupplier?: (supplierId: string, productName: string) => void;
  onViewInvoice?: (invoiceNumber: string) => void;
  onViewInventoryLedger?: (productId: string) => void;
  pagination?: { page: number; pageSize: number; total: number; onPageChange: (page: number) => void };
  layoutMode: 'cards' | 'table';
  supplierOnly?: boolean;
}

interface ActionButtonProps {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
  variant?: 'default' | 'danger';
}

function ActionButton({ icon, label, onClick, variant = 'default' }: ActionButtonProps) {
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      onClick={onClick}
      className={cn(
        'text-text-secondary hover:text-text-primary',
        variant === 'danger' && 'hover:text-danger hover:bg-danger/8'
      )}
      aria-label={label}
      title={label}
    >
      {icon}
    </Button>
  );
}

function RowActionMenu({
  product,
  onViewProduct,
  onAddPurchase,
  onEditProduct,
  onEditMinimumStock,
  onDeleteProduct,
  onArchiveProduct,
  onDeleteInventoryItem,
  onViewInventoryLedger,
}: {
  product: Product;
  onViewProduct: (product: Product) => void;
  onAddPurchase: (product: Product) => void;
  onEditProduct: (product: Product) => void;
  onEditMinimumStock: (product: Product) => void;
  onDeleteProduct: (productId: string) => void;
  onArchiveProduct?: (productId: string) => void;
  onDeleteInventoryItem?: (itemId: string) => void;
  onViewInventoryLedger?: (productId: string) => void;
}) {
  return (
    <ActionMenu
      label="خيارات المنتج"
      widthClassName="w-48"
      items={[
        { label: 'عرض', icon: Eye, onClick: () => onViewProduct(product) },
        { label: 'إضافة فاتورة', icon: Plus, onClick: () => onAddPurchase(product) },
        { label: 'تعديل', icon: PencilLine, onClick: () => onEditProduct(product) },
        { label: 'حد الأدنى', icon: SlidersHorizontal, onClick: () => onEditMinimumStock(product) },
          ...(onViewInventoryLedger ? [{ label: 'سجل الحركات', icon: FileText, onClick: () => onViewInventoryLedger(product.id) }] : []),
          ...(onArchiveProduct ? [{ label: 'أرشفة', icon: Archive, onClick: () => onArchiveProduct(product.id) }] : []),
        { label: 'حذف', icon: Trash, onClick: () => (onDeleteInventoryItem ? onDeleteInventoryItem(product.id) : onDeleteProduct(product.id)), danger: true },
      ]}
    />
  );
}

function LoadingSpinner() {
  return (
    <div className="flex h-64 items-center justify-center" role="status" aria-label="Loading">
      <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
    </div>
  );
}

function getStockDisplay(stock: number | undefined, minimumStockLevel?: number): { text: string; variant: 'success' | 'warning' | 'danger' | 'secondary' | 'outline' } {
  if (stock === undefined || stock === null) {
    return { text: 'غير محدد', variant: 'secondary' };
  }
  if (stock === 0) {
    return { text: 'نفد المخزون', variant: 'danger' };
  }
  if (Number(minimumStockLevel) > 0 && stock <= Number(minimumStockLevel)) {
    return { text: 'قليل', variant: 'warning' };
  }
  return { text: 'متوفر', variant: 'success' };
}

type ProductStockSection = 'out-of-stock' | 'low-stock' | 'available';

function getProductStockSection(stock: number | undefined, minimumStockLevel?: number): ProductStockSection {
  if (stock === 0) return 'out-of-stock';
  if (stock !== undefined && stock > 0 && stock <= Math.max(1, Number(minimumStockLevel) || 3)) {
    return 'low-stock';
  }
  return 'available';
}

function getInventoryStatusDisplay(status: unknown): { text: string; variant: 'success' | 'warning' | 'danger' | 'secondary' } {
  const labels: Record<string, { text: string; variant: 'success' | 'warning' | 'danger' | 'secondary' }> = {
    AVAILABLE: { text: 'متوفر', variant: 'success' },
    SOLD: { text: 'مباع', variant: 'info' },
    RETURNED: { text: 'مرتجع', variant: 'warning' },
    REVERSED: { text: 'إلغاء الشراء', variant: 'danger' },
    CANCELLED: { text: 'ملغى', variant: 'danger' },
    DELETED: { text: 'محذوف', variant: 'danger' },
    VOID: { text: 'ملغى', variant: 'danger' },
  };
  return labels[String(status || '').trim().toUpperCase()] || { text: 'غير محدد', variant: 'secondary' };
}

function getPrice(product: any): string {
  const value = product.sellingPrice ?? product.selling_price ?? product.price ?? 0;
  return formatPrice(normalizeCurrencyValue(value));
}

export function InventoryList({
  viewMode,
  filteredProducts,
  filteredInventoryItems,
  productsLoading,
  inventoryLoading,
  inventoryStockMap,
  onViewProduct,
  onAddPurchase,
  onEditProduct,
  onEditMinimumStock,
  onDeleteProduct,
  onArchiveProduct,
  onDeleteInventoryItem,
  onClearSearch,
  onReorderFromSupplier,
  onViewInvoice,
  onViewInventoryLedger,
  pagination,
  layoutMode,
  supplierOnly = false,
}: InventoryListProps) {
  const getConditionBadge = (condition: string) => {
    const normalized = String(condition || '').trim().toLowerCase();
    const variants: Record<string, { label: string; variant: 'default' | 'success' | 'warning' | 'danger' | 'info' | 'secondary' | 'outline' }> = {
      new: { label: 'جديد', variant: 'success' },
      used: { label: 'مستعمل', variant: 'secondary' },
      refurbished: { label: 'مجدد', variant: 'info' },
      parts_only: { label: 'قطع فقط', variant: 'danger' },
      default: { label: 'غير محدد', variant: 'outline' },
    };
    return variants[normalized] || { label: condition || 'غير محدد', variant: 'outline' };
  };

  const displayInventoryItems = filteredInventoryItems.filter((item: InventoryItem) => {
    const condition = String(item?.condition ?? '').trim().toUpperCase();
    return supplierOnly || condition !== 'USED';
  });
  const [inventoryStatusFilter, setInventoryStatusFilter] = useState<'ALL' | 'AVAILABLE' | 'SOLD' | 'REVERSED'>('ALL');
  const visibleInventoryItems = inventoryStatusFilter === 'ALL'
    ? displayInventoryItems
    : displayInventoryItems.filter((item) => String(item.status || '').toUpperCase() === inventoryStatusFilter);

  const getStockValue = (product: any): number | undefined => {
    if (product.current_quantity !== undefined && product.current_quantity !== null) {
      return Number(product.current_quantity);
    }

    const mappedStock = inventoryStockMap?.get(String(product.id));
    if (mappedStock !== undefined) {
      return mappedStock;
    }

    const status = String(product.status || '').trim().toUpperCase();
    const inactiveStatuses = new Set(['SOLD', 'RETURNED', 'REVERSED', 'CANCELLED', 'DELETED', 'VOID']);

    if (inactiveStatuses.has(status)) {
      return 0;
    }

    const value = Number(
      product.stock ??
      product.stock_quantity ??
      product.available_quantity ??
      product.current_quantity ??
      product.quantity ??
      0
    );

    if (!Number.isFinite(value) || value <= 0) {
      return status === 'AVAILABLE' ? 1 : 0;
    }

    return value;
  };

  const displayProducts = [...filteredProducts].sort((left, right) => {
    const sectionOrder: Record<ProductStockSection, number> = {
      'out-of-stock': 0,
      'low-stock': 1,
      available: 2,
    };
    const leftSection = getProductStockSection(getStockValue(left), left.min_stock_level);
    const rightSection = getProductStockSection(getStockValue(right), right.min_stock_level);
    return sectionOrder[leftSection] - sectionOrder[rightSection];
  });

  const emptyProductsState = (
    <EmptyState
      icon={<Inbox className="h-5 w-5" />}
      title="لا توجد منتجات"
      description="لم يتم العثور على منتجات تطابق بحثك"
      action={{
        label: 'مسح البحث',
        onClick: onClearSearch,
        variant: 'primary',
      }}
    />
  );

  return (
    <>
      {viewMode === 'products' && (
        <div className="rounded-[12px] border border-border bg-surface shadow-[0_8px_18px_rgba(15,23,42,0.04)]">
          <div className="flex items-center gap-2 border-b border-border px-5 py-4">
            <Package className="h-4 w-4 text-primary" />
            <h3 className="text-sm font-semibold text-text-primary">المنتجات</h3>
          </div>

          {productsLoading ? (
            <LoadingSpinner />
          ) : displayProducts.length === 0 ? (
            emptyProductsState
          ) : (
            <>
              <div className={layoutMode === 'table' ? 'block' : 'hidden'}>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[32%]">المنتج</TableHead>
                      <TableHead className="w-[15%]">الحالة</TableHead>
                      <TableHead className="w-[12%] text-center">المخزون</TableHead>
                      <TableHead className="w-[14%] text-center">السعر</TableHead>
                      <TableHead className="w-[19%]">التفاصيل</TableHead>
                      <TableHead className="w-[8%] text-end">الإجراءات</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {displayProducts.map((product: Product) => {
                      const stockVal = getStockValue(product);
                      const stockBadge = getStockDisplay(stockVal, product.min_stock_level);
                      const conditionBadge = getConditionBadge(product.condition || '');

                      return (
                        <TableRow key={product.id} className="align-middle">
                          <TableCell>
                            <div className="flex min-w-0 items-center gap-3">
                              <div className="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border bg-surface-muted">
                                {product.image_url || getLocalProductImage(product.id) || (product.category_id ? getCategoryImage(product.category_id) : undefined) ? (
                                  <img
                                    src={product.image_url || getLocalProductImage(product.id) || (product.category_id ? getCategoryImage(product.category_id) : undefined)}
                                    alt={product.name || 'المنتج'}
                                    className="h-full w-full object-contain"
                                  />
                                ) : <Package className="h-4 w-4 text-text-tertiary" />}
                              </div>
                              <div className="min-w-0">
                                <div className="truncate font-semibold text-text-primary">{product.name || '-'}</div>
                                <div className="mt-0.5 truncate text-[11px] text-text-tertiary">{product.sku || '-'}</div>
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <div className="flex flex-wrap items-center gap-1.5">
                              <Badge variant={stockBadge.variant} size="sm" className="capitalize">{stockBadge.text}</Badge>
                              <Badge variant={conditionBadge.variant} size="sm">{conditionBadge.label}</Badge>
                            </div>
                          </TableCell>
                          <TableCell className="text-center text-base font-bold text-text-primary">{stockVal !== undefined ? stockVal : '-'}</TableCell>
                          <TableCell className="text-center font-semibold text-primary">{getPrice(product)}</TableCell>
                          <TableCell>
                            <div className="space-y-0.5 text-xs text-text-secondary">
                              <div className="truncate">التصنيف: {product.category_name || product.category || product.categoryName || '-'}</div>
                              <div className="truncate text-text-tertiary">المورد: {product.supplier_name || 'غير محدد'}</div>
                            </div>
                          </TableCell>
                          <TableCell className="text-end">
                            <div className="flex items-center justify-end gap-2">
                              <Button
                                type="button"
                                variant="ghost"
                                size="icon"
                                onClick={() => onViewProduct(product)}
                                aria-label="عرض المنتج"
                                title="عرض المنتج"
                              >
                                <Eye className="h-3.5 w-3.5" />
                              </Button>
                              {onViewInventoryLedger && (
                                <Button
                                  type="button"
                                  variant="outline"
                                  size="sm"
                                  onClick={() => onViewInventoryLedger(product.id)}
                                  className="h-8 px-2.5 text-[11px]"
                                  aria-label={`سجل حركات ${product.name || 'المنتج'}`}
                                  title="سجل الحركات"
                                >
                                  <FileText className="h-3.5 w-3.5" />
                                </Button>
                              )}
                              <RowActionMenu
                                product={product}
                                onViewProduct={onViewProduct}
                                onAddPurchase={onAddPurchase}
                                onEditProduct={onEditProduct}
                                onEditMinimumStock={onEditMinimumStock}
                                onDeleteProduct={onDeleteProduct}
                                onArchiveProduct={onArchiveProduct}
                                onViewInventoryLedger={onViewInventoryLedger}
                              />
                            </div>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
                {pagination && (
                  <PaginationControls
                    page={pagination.page}
                    pageSize={pagination.pageSize}
                    total={pagination.total}
                    onPageChange={pagination.onPageChange}
                    isLoading={productsLoading}
                  />
                )}
              </div>

              <div className={layoutMode === 'cards' ? 'block' : 'hidden'}>
                <div className="grid grid-cols-1 gap-3 p-4 sm:grid-cols-2 xl:grid-cols-3">
                  {displayProducts.map((product: Product) => {
                    const stockVal = getStockValue(product);
                    const stockBadge = getStockDisplay(stockVal, product.min_stock_level);
                    const conditionBadge = getConditionBadge(product.condition || '');

                    return (
                      <article key={product.id} className="compact-product-card">
                        <div className="compact-product-card-header">
                          <div className="compact-product-media">
                            {product.image_url || getLocalProductImage(product.id) || (product.category_id ? getCategoryImage(product.category_id) : undefined) ? (
                              <img
                                src={product.image_url || getLocalProductImage(product.id) || (product.category_id ? getCategoryImage(product.category_id) : undefined)}
                                alt={product.name || 'المنتج'}
                                className="h-full w-full object-contain"
                              />
                            ) : (
                              <Package className="h-5 w-5 text-text-tertiary" />
                            )}
                          </div>
                          <div className="min-w-0 flex-1">
                            <div className="truncate text-sm font-semibold text-text-primary">{product.name || '-'}</div>
                            <div className="mt-0.5 truncate text-[11px] text-text-tertiary">{product.sku || '-'}</div>
                          </div>
                          <Badge variant={stockBadge.variant} size="sm" className="shrink-0 capitalize">{stockBadge.text}</Badge>
                        </div>

                        <div className="compact-product-metrics">
                          <div>
                            <span>المخزون</span>
                            <strong>{stockVal !== undefined ? stockVal : '-'}</strong>
                          </div>
                          <div>
                            <span>السعر</span>
                            <strong className="text-primary">{getPrice(product)}</strong>
                          </div>
                        </div>

                        <div className="compact-product-meta">
                          <span className="truncate">التصنيف: <strong>{product.category_name || product.category || product.categoryName || '-'}</strong></span>
                          <span className="truncate">المورد: <strong>{product.supplier_name || 'غير محدد'}</strong></span>
                          <Badge variant={conditionBadge.variant} size="sm">{conditionBadge.label}</Badge>
                        </div>

                        <div className="compact-product-actions">
                          <Button
                            type="button"
                            variant="secondary"
                            size="sm"
                            onClick={() => onViewProduct(product)}
                            aria-label={`عرض المنتج ${product.name || ''}`}
                            title="عرض المنتج"
                          >
                            <Eye className="h-3.5 w-3.5" />
                            عرض
                          </Button>
                          <RowActionMenu
                            product={product}
                            onViewProduct={onViewProduct}
                            onAddPurchase={onAddPurchase}
                            onEditProduct={onEditProduct}
                            onEditMinimumStock={onEditMinimumStock}
                            onDeleteProduct={onDeleteProduct}
                            onArchiveProduct={onArchiveProduct}
                            onViewInventoryLedger={onViewInventoryLedger}
                          />
                        </div>
                      </article>
                    );
                  })}
                </div>
              </div>
            </>
          )}
        </div>
      )}

      {viewMode === 'items' && (
        <div className="rounded-[12px] border border-border bg-surface shadow-[0_8px_18px_rgba(15,23,42,0.04)]">
          <div className="flex items-center gap-2 border-b border-border px-5 py-4">
            <PackageOpen className="h-4 w-4 text-primary" />
            <h3 className="text-sm font-semibold text-text-primary">{supplierOnly ? 'مشتريات الموردين' : 'عناصر المخزون الفردية'}</h3>
          </div>

          <div className="flex flex-wrap gap-2 border-b border-border px-5 py-3" role="tablist" aria-label="أقسام حالات المخزون">
            {([
              ['ALL', 'الكل'],
              ['AVAILABLE', 'المتاح'],
              ['SOLD', 'المباع'],
              ['REVERSED', 'إلغاء الشراء'],
            ] as const).map(([status, label]) => (
              <Button
                key={status}
                type="button"
                size="sm"
                variant={inventoryStatusFilter === status ? 'primary' : 'ghost'}
                role="tab"
                aria-selected={inventoryStatusFilter === status}
                onClick={() => setInventoryStatusFilter(status)}
              >
                {label}
              </Button>
            ))}
          </div>

          {inventoryLoading ? (
            <LoadingSpinner />
          ) : visibleInventoryItems.length === 0 ? (
            <EmptyState
              icon={<Inbox className="h-5 w-5" />}
              title="لا توجد قطع فردية مسجلة"
              description="الكميات الإجمالية تظهر في تبويب المنتجات. استخدم طريقة القطعة المحددة بالباركود لإظهار كل قطعة هنا."
            />
          ) : (
            <div className={layoutMode === 'table' ? 'block' : 'hidden'}>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[30%]">العنصر</TableHead>
                    <TableHead className="w-[14%]">الحالة</TableHead>
                    <TableHead className="w-[16%]">المورد</TableHead>
                    <TableHead className="w-[14%]">تاريخ الشراء</TableHead>
                    <TableHead className="w-[10%] text-center">شراء</TableHead>
                    <TableHead className="w-[10%] text-center">بيع</TableHead>
                    <TableHead className="w-[8%]">القطعة</TableHead>
                    <TableHead className="w-[8%] text-end">الإجراءات</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {visibleInventoryItems.map((item: InventoryItem) => {
                    const condition = item.condition?.toLowerCase() || '';
                    const condBadge = getConditionBadge(condition);
                    const itemStatus = getInventoryStatusDisplay(item.status);

                    return (
                      <TableRow key={item.id}>
                        <TableCell>
                            <div className="flex min-w-0 items-center gap-3">
                              <div className="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border bg-surface-muted">
                                {((item as any).image_url || (item.product as any)?.image_url || getLocalProductImage(item.product_id || item.id) || getCategoryImage((item as any).category_id || '')) ? (
                                  <img src={(item as any).image_url || (item.product as any)?.image_url || getLocalProductImage(item.product_id || item.id) || getCategoryImage((item as any).category_id || '')} alt={item.product_name || 'عنصر المخزون'} className="h-full w-full object-contain" />
                                ) : <PackageOpen className="h-4 w-4 text-text-tertiary" />}
                              </div>
                              <div className="min-w-0">
                                <div className="truncate font-semibold text-text-primary">{item.product_name || item.product?.name || '-'}</div>
                                <div className="mt-0.5 truncate text-[11px] text-text-tertiary">{item.barcode || item.product?.barcode || '-'}</div>
                              </div>
                            </div>
                        </TableCell>
                        <TableCell><div className="flex flex-wrap gap-1.5"><Badge variant={itemStatus.variant} size="sm">{itemStatus.text}</Badge><Badge variant={condBadge.variant} size="sm">{condBadge.label}</Badge></div></TableCell>
                        <TableCell className="text-text-secondary">{item.supplier_name || '-'}</TableCell>
                        <TableCell className="text-text-secondary">{item.purchase_date ? new Date(item.purchase_date).toLocaleDateString('ar-SA') : '-'}</TableCell>
                        <TableCell className="text-center font-medium text-text-secondary">{formatPrice(normalizeCurrencyValue(item.purchase_cost ?? 0))}</TableCell>
                        <TableCell className="text-center font-medium text-primary">{formatPrice(normalizeCurrencyValue(item.selling_price ?? item.price ?? 0))}</TableCell>
                        <TableCell><Badge variant={condBadge.variant} size="sm">{condBadge.label}</Badge></TableCell>
                        <TableCell className="text-end">
                          <div className="flex items-center justify-end gap-2">
                            <Button type="button" variant="ghost" size="icon" onClick={() => onViewProduct({
                              id: item.id,
                              name: item.product_name || item.product?.name || '-',
                              sku: item.product?.sku || item.sku || item.barcode || '',
                              condition: item.condition || '',
                              category_name: item.category_name || '',
                              stock: item.quantity ?? 0,
                              price: item.price ?? 0,
                              sellingPrice: item.selling_price ?? item.price ?? 0,
                              category: '',
                              categoryName: '',
                              status: item.status || 'AVAILABLE',
                              barcode: item.barcode || '',
                              supplier_id: item.supplier_id || '',
                              supplier_name: item.supplier_name || '',
                            } as Product)} aria-label="عرض العنصر" title="عرض العنصر">
                              <Eye className="h-3.5 w-3.5" />
                            </Button>
                            <RowActionMenu
                              product={{
                                id: item.id,
                                name: item.product_name || item.product?.name || '-',
                                sku: item.product?.sku || item.sku || item.barcode || '',
                                condition: item.condition || '',
                                category_name: item.category_name || '',
                                stock: item.quantity ?? 0,
                                price: item.price ?? 0,
                                sellingPrice: item.selling_price ?? item.price ?? 0,
                                category: '',
                                categoryName: '',
                                status: item.status || 'AVAILABLE',
                                barcode: item.barcode || '',
                                supplier_id: item.supplier_id || '',
                                supplier_name: item.supplier_name || '',
                              } as Product}
                              onViewProduct={onViewProduct}
                              onAddPurchase={onAddPurchase}
                              onEditProduct={onEditProduct}
                              onEditMinimumStock={onEditMinimumStock}
                              onDeleteProduct={onDeleteProduct}
                              onArchiveProduct={onArchiveProduct}
                              onDeleteInventoryItem={onDeleteInventoryItem}
                              onViewInventoryLedger={onViewInventoryLedger}
                            />
                          </div>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
              {pagination && (
                <PaginationControls
                  page={pagination.page}
                  pageSize={pagination.pageSize}
                  total={pagination.total}
                  onPageChange={pagination.onPageChange}
                  isLoading={inventoryLoading}
                />
              )}
            </div>
          )}

          <div className={layoutMode === 'cards' ? 'block' : 'hidden'}>
            {inventoryLoading ? (
              <div className="p-4"><LoadingSpinner /></div>
            ) : visibleInventoryItems.length === 0 ? (
              <EmptyState
                icon={<Inbox className="h-5 w-5" />}
                title="لا توجد قطع فردية مسجلة"
                description="الكميات الإجمالية تظهر في تبويب المنتجات. استخدم طريقة القطعة المحددة بالباركود لإظهار كل قطعة هنا."
              />
            ) : (
              <div className="grid grid-cols-1 gap-3 p-4 sm:grid-cols-2 xl:grid-cols-3">
                {visibleInventoryItems.map((item: InventoryItem) => {
                  const condBadge = getConditionBadge(item.condition?.toLowerCase() || '');
                  const itemStatus = getInventoryStatusDisplay(item.status);
                  const itemImage = (item as any).image_url || (item.product as any)?.image_url || getLocalProductImage(item.product_id || item.id) || getCategoryImage((item as any).category_id || '');
                  return (
                    <article key={item.id} className="compact-product-card">
                      <div className="compact-product-card-header">
                        <div className="compact-product-media">
                          {itemImage ? <img src={itemImage} alt={item.product_name || 'عنصر المخزون'} className="h-full w-full object-contain" /> : <PackageOpen className="h-5 w-5 text-text-tertiary" />}
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="truncate text-sm font-semibold text-text-primary">{item.product_name || item.product?.name || '-'}</div>
                          <div className="mt-0.5 truncate text-[11px] text-text-tertiary">{item.barcode || item.product?.barcode || '-'}</div>
                        </div>
                        <Badge variant={itemStatus.variant} size="sm" className="shrink-0">{itemStatus.text}</Badge>
                      </div>

                      <div className="compact-product-metrics">
                        <div><span>شراء</span><strong>{formatPrice(normalizeCurrencyValue(item.purchase_cost ?? 0))}</strong></div>
                        <div><span>بيع</span><strong className="text-primary">{formatPrice(normalizeCurrencyValue(item.selling_price ?? item.price ?? 0))}</strong></div>
                      </div>

                      <div className="compact-product-meta">
                        <span className="truncate">المورد: <strong>{item.supplier_name || '-'}</strong></span>
                        <span className="truncate">التصنيف: <strong>{item.category_name || 'بدون تصنيف'}</strong></span>
                        <span className="truncate">الشراء: <strong>{item.purchase_date ? new Date(item.purchase_date).toLocaleDateString('ar-SA') : '-'}</strong></span>
                        <Badge variant={condBadge.variant} size="sm">{condBadge.label}</Badge>
                      </div>

                      <div className="compact-product-actions">
                        <Button type="button" variant="primary" size="sm" onClick={() => onViewProduct({
                          id: item.id,
                          name: item.product_name || item.product?.name || '-',
                          sku: item.product?.sku || item.sku || item.barcode || '',
                          condition: item.condition || '',
                          category_name: item.category_name || '',
                          stock: item.quantity ?? 0,
                          price: item.price ?? 0,
                          sellingPrice: item.selling_price ?? item.price ?? 0,
                          category: '',
                          categoryName: '',
                          status: item.status || 'AVAILABLE',
                          barcode: item.barcode || '',
                        } as Product)} className="h-8 px-3 text-[11px]">
                          <Eye className="h-3.5 w-3.5" />
                          عرض
                        </Button>
                        <RowActionMenu
                          product={{
                            id: item.id,
                            name: item.product_name || item.product?.name || '-',
                            sku: item.product?.sku || item.sku || item.barcode || '',
                            condition: item.condition || '',
                            category_name: item.category_name || '',
                            stock: item.quantity ?? 0,
                            price: item.price ?? 0,
                            sellingPrice: item.selling_price ?? item.price ?? 0,
                            category: '',
                            categoryName: '',
                            status: item.status || 'AVAILABLE',
                            barcode: item.barcode || '',
                          } as Product}
                          onViewProduct={onViewProduct}
                          onAddPurchase={onAddPurchase}
                          onEditProduct={onEditProduct}
                          onEditMinimumStock={onEditMinimumStock}
                          onDeleteProduct={onDeleteProduct}
                              onArchiveProduct={onArchiveProduct}
                          onDeleteInventoryItem={onDeleteInventoryItem}
                          onViewInventoryLedger={onViewInventoryLedger}
                        />
                      </div>
                    </article>
                  );
                })}
              </div>
            )}
            {pagination && visibleInventoryItems.length > 0 && (
              <PaginationControls
                page={pagination.page}
                pageSize={pagination.pageSize}
                total={pagination.total}
                onPageChange={pagination.onPageChange}
                isLoading={inventoryLoading}
              />
            )}
          </div>
        </div>
      )}
    </>
  );
}
