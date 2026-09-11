import { useState } from 'react';
import { Badge } from '../../../components/ui/badge';
import { EmptyState } from '../../../components/ui/empty-state';
import { Button } from '../../../components/ui/button';
import { Table, TableHeader, TableBody, TableHead, TableRow, TableCell } from '../../../components/ui/table';
import { Package, PackageOpen, Eye, Edit, SlidersHorizontal, Trash2, Inbox, RefreshCw, FileText, MoreHorizontal, Plus, ArrowUpRight, Trash, Copy, PencilLine } from 'lucide-react';
import { Product, InventoryItem, ViewMode } from '../types/inventory.types';
import { formatPrice, normalizeCurrencyValue } from '../../../utils';
import { cn } from '../../../utils';

interface InventoryListProps {
  viewMode: ViewMode;
  filteredProducts: Product[];
  filteredInventoryItems: InventoryItem[];
  productsLoading: boolean;
  inventoryLoading: boolean;
  searchQuery: string;
  onViewProduct: (product: Product) => void;
  onAddPurchase: (product: Product) => void;
  onEditProduct: (product: Product) => void;
  onEditMinimumStock: (product: Product) => void;
  onDeleteProduct: (productId: string) => void;
  onClearSearch: () => void;
  onReorderFromSupplier?: (supplierId: string, productName: string) => void;
  onViewInvoice?: (invoiceNumber: string) => void;
  onViewInventoryLedger?: (productId: string) => void;
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
  onViewInventoryLedger,
}: {
  product: Product;
  onViewProduct: (product: Product) => void;
  onAddPurchase: (product: Product) => void;
  onEditProduct: (product: Product) => void;
  onEditMinimumStock: (product: Product) => void;
  onDeleteProduct: (productId: string) => void;
  onViewInventoryLedger?: (productId: string) => void;
}) {
  const [open, setOpen] = useState(false);

  const actions = [
    { label: 'عرض', icon: Eye, onClick: () => onViewProduct(product) },
    { label: 'إضافة فاتورة', icon: Plus, onClick: () => onAddPurchase(product) },
    { label: 'تعديل', icon: PencilLine, onClick: () => onEditProduct(product) },
    { label: 'حد الأدنى', icon: SlidersHorizontal, onClick: () => onEditMinimumStock(product) },
    ...(onViewInventoryLedger ? [{ label: 'سجل الحركات', icon: FileText, onClick: () => onViewInventoryLedger(product.id) }] : []),
    { label: 'حذف', icon: Trash, onClick: () => onDeleteProduct(product.id), danger: true },
  ];

  return (
    <div className="relative">
      <Button
        type="button"
        variant="ghost"
        size="icon"
        onClick={() => setOpen((prev) => !prev)}
        className="text-text-secondary hover:text-text-primary"
        aria-label="خيارات المنتج"
        title="خيارات المنتج"
      >
        <MoreHorizontal className="h-4 w-4" />
      </Button>

      {open && (
        <div className="absolute left-0 top-full z-20 mt-2 w-44 rounded-xl border border-border bg-surface-elevated p-1 shadow-[0_16px_36px_rgba(15,23,42,0.22)]">
          {actions.map(({ label, icon: Icon, onClick, danger }) => (
            <button
              key={label}
              type="button"
              onClick={() => {
                onClick();
                setOpen(false);
              }}
              className={cn(
                'flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm transition-colors',
                danger ? 'text-danger hover:bg-danger/8' : 'text-text-primary hover:bg-surface'
              )}
            >
              <Icon className="h-4 w-4" />
              <span>{label}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

function LoadingSpinner() {
  return (
    <div className="flex h-64 items-center justify-center" role="status" aria-label="Loading">
      <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
    </div>
  );
}

const DEFAULT_LOW_STOCK_THRESHOLD = 3;

function getStockDisplay(stock: number | undefined, minimumStockLevel?: number): { text: string; variant: 'success' | 'warning' | 'danger' | 'secondary' | 'outline' } {
  if (stock === undefined || stock === null) {
    return { text: 'غير محدد', variant: 'secondary' };
  }
  if (stock === 0) {
    return { text: 'نفد المخزون', variant: 'danger' };
  }
  const threshold = Number(minimumStockLevel) > 0 ? Number(minimumStockLevel) : DEFAULT_LOW_STOCK_THRESHOLD;
  if (stock <= threshold) {
    return { text: 'قليل', variant: 'warning' };
  }
  return { text: 'متوفر', variant: 'success' };
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
  onViewProduct,
  onAddPurchase,
  onEditProduct,
  onEditMinimumStock,
  onDeleteProduct,
  onClearSearch,
  onReorderFromSupplier,
  onViewInvoice,
  onViewInventoryLedger,
}: InventoryListProps) {
  const getConditionBadge = (condition: string) => {
    const normalized = String(condition || '').trim().toLowerCase();
    const variants: Record<string, { label: string; variant: 'default' | 'success' | 'warning' | 'danger' | 'info' | 'secondary' | 'outline' }> = {
      new: { label: 'جديد', variant: 'success' },
      used: { label: 'مستعمل', variant: 'secondary' },
      refurbished: { label: 'مجدد', variant: 'info' },
      parts_only: { label: 'قطع فقط', variant: 'danger' },
      default: { label: 'افتراضي', variant: 'outline' },
    };
    return variants[normalized] || { label: condition || 'افتراضي', variant: 'outline' };
  };

  const displayProducts = filteredProducts;
  const displayInventoryItems = filteredInventoryItems.filter((item: InventoryItem) => {
    const condition = String(item?.condition ?? '').trim().toUpperCase();
    return condition !== 'USED';
  });

  const getStockValue = (product: any): number | undefined => {
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
            <div className="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow>
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
                              variant="primary"
                              size="sm"
                              onClick={() => onViewProduct(product)}
                              className="h-8 px-2.5 text-[11px]"
                              aria-label="عرض المنتج"
                              title="عرض المنتج"
                            >
                              <Eye className="h-3.5 w-3.5" />
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
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          )}

          <div className="block md:hidden">
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
              <div className="grid gap-3 p-4">
                {displayProducts.map((product: Product) => {
                  const stockVal = getStockValue(product);
                  const stockBadge = getStockDisplay(stockVal, product.min_stock_level);
                  const conditionBadge = getConditionBadge(product.condition || '');

                  return (
                    <div key={product.id} className="rounded-xl border border-border bg-surface-elevated/25 p-4">
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
                          variant="primary"
                          size="sm"
                          onClick={() => onViewProduct(product)}
                          className="h-8 px-3 text-[11px]"
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
          </div>
        </div>
      )}

      {viewMode === 'items' && (
        <div className="rounded-[12px] border border-border bg-surface shadow-[0_8px_18px_rgba(15,23,42,0.04)]">
          <div className="flex items-center gap-2 border-b border-border px-5 py-4">
            <PackageOpen className="h-4 w-4 text-primary" />
            <h3 className="text-sm font-semibold text-text-primary">عناصر المخزون</h3>
          </div>

          {inventoryLoading ? (
            <LoadingSpinner />
          ) : displayInventoryItems.length === 0 ? (
            <EmptyState icon={<Inbox className="h-5 w-5" />} title="لا توجد عناصر" description="لم يتم العثور على عناصر في المخزون" />
          ) : (
            <div className="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[17%]">المنتج</TableHead>
                    <TableHead className="w-[10%]">الحالة</TableHead>
                    <TableHead className="w-[12%]">المورد</TableHead>
                    <TableHead className="w-[12%]">تاريخ الشراء</TableHead>
                    <TableHead className="w-[12%] text-center">شراء</TableHead>
                    <TableHead className="w-[12%] text-center">بيع</TableHead>
                    <TableHead className="w-[10%]">الموقع</TableHead>
                    <TableHead className="w-[10%]">مؤشر</TableHead>
                    <TableHead className="w-[8%] text-end">الإجراءات</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {displayInventoryItems.map((item: InventoryItem) => {
                    const condition = item.condition?.toLowerCase() || '';
                    const condBadge = getConditionBadge(condition);
                    const itemStatus = item.status === 'AVAILABLE' ? 'متوفر' : item.status || 'غير محدد';
                    const itemStatusVariant = item.status === 'AVAILABLE' ? 'success' : 'warning';

                    return (
                      <TableRow key={item.id}>
                        <TableCell>
                          <div className="min-w-0">
                            <div className="truncate font-semibold text-text-primary">{item.product_name || item.product?.name || '-'}</div>
                            <div className="mt-0.5 text-[11px] text-text-tertiary">{item.barcode || item.product?.barcode || '-'}</div>
                          </div>
                        </TableCell>
                        <TableCell><Badge variant={condBadge.variant} size="sm">{condBadge.label}</Badge></TableCell>
                        <TableCell className="text-text-secondary">{item.supplier_name || '-'}</TableCell>
                        <TableCell className="text-text-secondary">{item.purchase_date ? new Date(item.purchase_date).toLocaleDateString('ar-SA') : '-'}</TableCell>
                        <TableCell className="text-center font-medium text-text-secondary">{formatPrice(normalizeCurrencyValue(item.purchase_cost ?? 0))}</TableCell>
                        <TableCell className="text-center font-medium text-primary">{formatPrice(normalizeCurrencyValue(item.selling_price ?? item.price ?? 0))}</TableCell>
                        <TableCell className="text-text-secondary">{item.location || '-'}</TableCell>
                        <TableCell><Badge variant={itemStatusVariant} size="sm">{itemStatus}</Badge></TableCell>
                        <TableCell className="text-end">
                          <div className="flex items-center justify-end gap-2">
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
                            } as Product)} className="h-8 px-2.5 text-[11px]" aria-label="عرض العنصر" title="عرض العنصر">
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
                              } as Product}
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
            </div>
          )}

          <div className="block md:hidden">
            {inventoryLoading ? (
              <div className="p-4"><LoadingSpinner /></div>
            ) : displayInventoryItems.length === 0 ? (
              <EmptyState icon={<Inbox className="h-5 w-5" />} title="لا توجد عناصر" description="لم يتم العثور على عناصر في المخزون" />
            ) : (
              <div className="grid gap-3 p-4">
                {displayInventoryItems.map((item: InventoryItem) => {
                  const condBadge = getConditionBadge(item.condition?.toLowerCase() || '');
                  return (
                    <div key={item.id} className="rounded-xl border border-border bg-surface-elevated/25 p-4">
                      <div className="mb-3 flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <div className="truncate font-semibold text-text-primary">{item.product_name || item.product?.name || '-'}</div>
                          <div className="mt-1 text-[11px] text-text-tertiary">{item.barcode || item.product?.barcode || '-'}</div>
                        </div>
                        <Badge variant={item.status === 'AVAILABLE' ? 'success' : 'warning'} size="sm">{item.status === 'AVAILABLE' ? 'متوفر' : item.status || 'غير محدد'}</Badge>
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
                          onViewInventoryLedger={onViewInventoryLedger}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
}
