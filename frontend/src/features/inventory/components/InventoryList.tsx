import { useState } from 'react';
import { Badge } from '../../../components/ui/badge';
import { EmptyState } from '../../../components/ui/empty-state';
import { Button } from '../../../components/ui/button';
import { Table, TableHeader, TableBody, TableHead, TableRow, TableCell } from '../../../components/ui/table';
import { Package, PackageOpen, Eye, Edit, SlidersHorizontal, Trash2, Inbox, RefreshCw, FileText, Plus, ArrowUpRight, Trash, Copy, PencilLine } from 'lucide-react';
import { ActionMenu } from '../../../components/ui/action-menu';
import { Product, InventoryItem, ViewMode } from '../types/inventory.types';
import { formatPrice, normalizeCurrencyValue } from '../../../utils';
import { cn } from '../../../utils';
import { PaginationControls } from '../../../components/ui/pagination-controls';
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
  onDeleteInventoryItem,
  onViewInventoryLedger,
}: {
  product: Product;
  onViewProduct: (product: Product) => void;
  onAddPurchase: (product: Product) => void;
  onEditProduct: (product: Product) => void;
  onEditMinimumStock: (product: Product) => void;
  onDeleteProduct: (productId: string) => void;
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
          ) : (
            <div className={layoutMode === 'table' ? 'block' : 'hidden'}>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[72px]">الصورة</TableHead>
                    <TableHead className="w-[24%]">الاسم</TableHead>
                    <TableHead className="w-[12%]">SKU</TableHead>
                    <TableHead className="w-[14%]">التصنيف</TableHead>
                    <TableHead className="w-[12%]">الحالة</TableHead>
                    <TableHead className="w-[10%] text-center">المخزون</TableHead>
                    <TableHead className="w-[12%] text-center">السعر قبل الضريبة</TableHead>
                    <TableHead className="w-[12%]">المؤشر</TableHead>
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
                          <div style={{ width: '48px', height: '48px', overflow: 'hidden', borderRadius: '6px' }}>
                            {product.image_url ? (
                              <img src={product.image_url} alt={product.name || 'المنتج'} style={{ width: '48px', height: '48px', maxWidth: '48px', maxHeight: '48px', objectFit: 'contain', display: 'block' }} />
                            ) : <Package className="h-8 w-8 text-text-tertiary" />}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="min-w-0">
                            <div className="max-w-[220px] truncate font-semibold text-text-primary">{product.name || '-'}</div>
                            <div className="mt-0.5 text-[11px] text-text-tertiary">{product.barcode || product.sku || '-'}</div>
                          </div>
                        </TableCell>
                        <TableCell className="font-medium text-text-secondary">{product.sku || '-'}</TableCell>
                        <TableCell className="text-text-secondary">{product.category_name || product.category || product.categoryName || '-'}</TableCell>
                        <TableCell>
                          <Badge variant={conditionBadge.variant} size="sm">{conditionBadge.label}</Badge>
                        </TableCell>
                        <TableCell className="text-center font-semibold text-text-primary">{stockVal !== undefined ? stockVal : '-'}</TableCell>
                        <TableCell className="text-center font-semibold text-primary">{getPrice(product)}</TableCell>
                        <TableCell>
                          <Badge variant={stockBadge.variant} size="sm" className="capitalize">{stockBadge.text}</Badge>
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
          )}

          <div className={layoutMode === 'cards' ? 'block' : 'hidden'}>
            {productsLoading ? (
              <div className="p-4"><LoadingSpinner /></div>
            ) : displayProducts.length === 0 ? (
              <EmptyState
                icon={<Inbox className="h-5 w-5" />}
                title="لا توجد منتجات"
                description="لم يتم العثور على منتجات تطابق بحثك"
                action={{ label: 'مسح البحث', onClick: onClearSearch, variant: 'primary' }}
              />
            ) : (
              <div className="grid grid-cols-1 gap-3 p-4 sm:grid-cols-2 xl:grid-cols-3">
                {displayProducts.map((product: Product) => {
                  const stockVal = getStockValue(product);
                  const stockBadge = getStockDisplay(stockVal, product.min_stock_level);
                  const conditionBadge = getConditionBadge(product.condition || '');

                  return (
                    <div key={product.id} className="rounded-xl border border-border bg-surface-elevated/25 p-4">
                      <div className="mb-3 flex items-center justify-center overflow-hidden rounded-lg bg-surface-muted" style={{ height: '144px', minHeight: '144px' }}>
                        {product.image_url ? (
                          <img src={product.image_url} alt={product.name || 'المنتج'} className="max-h-full max-w-full object-contain" style={{ width: '100%', height: '100%' }} />
                        ) : (
                          <Package className="h-12 w-12 text-text-tertiary" />
                        )}
                      </div>
                      <div className="mb-3 flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <div className="truncate font-semibold text-text-primary">{product.name || '-'}</div>
                          <div className="mt-1 text-[11px] text-text-tertiary">{product.sku || '-'}</div>
                        </div>
                        <Badge variant={stockBadge.variant} size="sm" className="capitalize">{stockBadge.text}</Badge>
                      </div>

                      <div className="space-y-2 text-sm">
                        <div className="flex items-center justify-between gap-2">
                          <span className="text-text-tertiary">التصنيف</span>
                          <span className="font-medium text-text-secondary">{product.category_name || product.category || product.categoryName || '-'}</span>
                        </div>
                        <div className="flex items-center justify-between gap-2">
                          <span className="text-text-tertiary">المورد</span>
                          <span className="font-medium text-text-secondary">{product.supplier_name || 'غير محدد'}</span>
                        </div>
                        <div className="flex items-center justify-between gap-2">
                          <span className="text-text-tertiary">الحالة</span>
                          <Badge variant={conditionBadge.variant} size="sm">{conditionBadge.label}</Badge>
                        </div>
                        <div className="flex items-center justify-between gap-2">
                          <span className="text-text-tertiary">المخزون</span>
                          <span className="font-semibold text-text-primary">{stockVal !== undefined ? stockVal : '-'}</span>
                        </div>
                        <div className="flex items-center justify-between gap-2">
                          <span className="text-text-tertiary">السعر قبل الضريبة</span>
                          <span className="font-semibold text-primary">{getPrice(product)}</span>
                        </div>
                      </div>

                      <div className="mt-3 flex items-center justify-between gap-2 border-t border-border pt-3">
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
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
                          onViewInventoryLedger={onViewInventoryLedger}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
            {pagination && displayProducts.length > 0 && (
              <PaginationControls
                page={pagination.page}
                pageSize={pagination.pageSize}
                total={pagination.total}
                onPageChange={pagination.onPageChange}
                isLoading={productsLoading}
              />
            )}
          </div>
        </div>
      )}

      {viewMode === 'items' && (
        <div className="rounded-[12px] border border-border bg-surface shadow-[0_8px_18px_rgba(15,23,42,0.04)]">
          <div className="flex items-center gap-2 border-b border-border px-5 py-4">
            <PackageOpen className="h-4 w-4 text-primary" />
            <h3 className="text-sm font-semibold text-text-primary">{supplierOnly ? 'مشتريات الموردين' : 'عناصر المخزون'}</h3>
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
            <EmptyState icon={<Inbox className="h-5 w-5" />} title="لا توجد عناصر" description="لم يتم العثور على عناصر في المخزون" />
          ) : (
            <div className={layoutMode === 'table' ? 'block' : 'hidden'}>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[72px]">الصورة</TableHead>
                    <TableHead className="w-[17%]">المنتج</TableHead>
                    <TableHead className="w-[10%]">الحالة</TableHead>
                    <TableHead className="w-[12%]">المورد</TableHead>
                    <TableHead className="w-[12%]">تاريخ الشراء</TableHead>
                    <TableHead className="w-[12%] text-center">شراء</TableHead>
                    <TableHead className="w-[12%] text-center">بيع</TableHead>
                    <TableHead className="w-[10%]">حالة القطعة</TableHead>
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
                          <div style={{ width: '48px', height: '48px', overflow: 'hidden', borderRadius: '6px' }}>
                            {((item as any).image_url || (item.product as any)?.image_url || getLocalProductImage(item.product_id || item.id) || getCategoryImage((item as any).category_id || '')) ? (
                              <img src={(item as any).image_url || (item.product as any)?.image_url || getLocalProductImage(item.product_id || item.id) || getCategoryImage((item as any).category_id || '')} alt={item.product_name || 'عنصر المخزون'} style={{ width: '48px', height: '48px', maxWidth: '48px', maxHeight: '48px', objectFit: 'contain', display: 'block' }} />
                            ) : <PackageOpen className="h-8 w-8 text-text-tertiary" />}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="min-w-0">
                            <div className="truncate font-semibold text-text-primary">{item.product_name || item.product?.name || '-'}</div>
                            <div className="mt-0.5 text-[11px] text-text-tertiary">{item.barcode || item.product?.barcode || '-'}</div>
                            <div className="mt-0.5 text-[11px] text-text-tertiary">التصنيف: {item.category_name || 'بدون تصنيف'}</div>
                          </div>
                        </TableCell>
                        <TableCell><Badge variant={itemStatus.variant} size="sm">{itemStatus.text}</Badge></TableCell>
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
              <EmptyState icon={<Inbox className="h-5 w-5" />} title="لا توجد عناصر" description="لم يتم العثور على عناصر في المخزون" />
            ) : (
              <div className="grid grid-cols-1 gap-3 p-4 sm:grid-cols-2 xl:grid-cols-3">
                {visibleInventoryItems.map((item: InventoryItem) => {
                  const condBadge = getConditionBadge(item.condition?.toLowerCase() || '');
                  return (
                    <div key={item.id} className="rounded-xl border border-border bg-surface-elevated/25 p-4">
                      <div className="mb-3 flex items-center justify-center overflow-hidden rounded-lg bg-surface-muted" style={{ height: '144px', minHeight: '144px' }}>
                        {((item as any).image_url || (item.product as any)?.image_url || getLocalProductImage(item.product_id || item.id) || getCategoryImage((item as any).category_id || '')) ? (
                          <img src={(item as any).image_url || (item.product as any)?.image_url || getLocalProductImage(item.product_id || item.id) || getCategoryImage((item as any).category_id || '')} alt={item.product_name || 'عنصر المخزون'} className="max-h-full max-w-full object-contain" style={{ width: '100%', height: '100%' }} />
                        ) : (
                          <PackageOpen className="h-12 w-12 text-text-tertiary" />
                        )}
                      </div>
                      <div className="mb-3 flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <div className="truncate font-semibold text-text-primary">{item.product_name || item.product?.name || '-'}</div>
                          <div className="mt-1 text-[11px] text-text-tertiary">{item.barcode || item.product?.barcode || '-'}</div>
                          <div className="mt-0.5 text-[11px] text-text-tertiary">التصنيف: {item.category_name || 'بدون تصنيف'}</div>
                        </div>
                        {(() => {
                          const itemStatus = getInventoryStatusDisplay(item.status);
                          return <Badge variant={itemStatus.variant} size="sm">{itemStatus.text}</Badge>;
                        })()}
                      </div>

                      <div className="space-y-2 text-sm">
                        <div className="flex items-center justify-between gap-2"><span className="text-text-tertiary">الحالة</span><Badge variant={condBadge.variant} size="sm">{condBadge.label}</Badge></div>
                        <div className="flex items-center justify-between gap-2"><span className="text-text-tertiary">المورد</span><span className="font-medium text-text-secondary">{item.supplier_name || '-'}</span></div>
                        <div className="flex items-center justify-between gap-2"><span className="text-text-tertiary">شراء</span><span className="font-medium text-text-secondary">{formatPrice(normalizeCurrencyValue(item.purchase_cost ?? 0))}</span></div>
                        <div className="flex items-center justify-between gap-2"><span className="text-text-tertiary">بيع</span><span className="font-medium text-primary">{formatPrice(normalizeCurrencyValue(item.selling_price ?? item.price ?? 0))}</span></div>
                      </div>

                      <div className="mt-3 flex items-center justify-between gap-2 border-t border-border pt-3">
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
                          onDeleteInventoryItem={onDeleteInventoryItem}
                          onViewInventoryLedger={onViewInventoryLedger}
                        />
                      </div>
                    </div>
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
