import { Badge } from '../../../components/ui/badge';
import { EmptyState } from '../../../components/ui/empty-state';
import { Button } from '../../../components/ui/button';
import { Package, PackageOpen, Eye, Edit, SlidersHorizontal, Trash2, Inbox, RefreshCw, FileText } from 'lucide-react';
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
        "text-text-secondary hover:text-text-primary",
        variant === 'danger' && "hover:text-red hover:bg-red/8"
      )}
      aria-label={label}
    >
      {icon}
    </Button>
  );
}

function LoadingSpinner() {
  return (
    <div className="flex h-64 items-center justify-center" role="status" aria-label="Loading">
      <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
    </div>
  );
}

function getStockDisplay(stock: number | undefined): { text: string; variant: 'success' | 'warning' | 'danger' | 'secondary' | 'outline' } {
  if (stock === undefined || stock === null) {
    return { text: 'غير محدد', variant: 'secondary' };
  }
  if (stock === 0) {
    return { text: 'نفد المخزون', variant: 'danger' };
  }
  if (stock <= 10) {
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
      {/* Products Table */}
      {viewMode === 'products' && (
        <div className="rounded-xl border border-border bg-surface">
          <div className="border-b border-border px-5 py-4 flex items-center gap-2">
            <Package className="h-5 w-5 text-primary" />
            <h3 className="text-sm font-semibold text-text-primary">المنتجات</h3>
          </div>
          <div>
            {productsLoading ? (
              <LoadingSpinner />
            ) : displayProducts.length === 0 ? (
              <EmptyState
                icon={<Inbox className="h-5 w-5" />}
                title="لا توجد منتجات"
                description="لم يتم العثور على منتجات تطابق بحثك"
                action={
                  <Button
                    size="sm"
                    variant="primary"
                    onClick={onClearSearch}
                  >
                    مسح البحث
                  </Button>
                }
              />
            ) : (
              <>
                {/* Desktop Table */}
                <div className="hidden md:block">
                  <table className="w-full border-collapse">
                    <thead>
                      <tr className="border-b border-border bg-surface-elevated/50">
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">الاسم</th>
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">SKU</th>
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">التصنيف</th>
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">الحالة</th>
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">المخزون</th>
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">السعر</th>
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">الحالة</th>
                        <th className="h-12 px-4 text-end text-xs font-medium text-text-tertiary uppercase tracking-wider">الإجراءات</th>
                      </tr>
                    </thead>
                    <tbody>
                      {displayProducts.map((product: Product) => {
                        const stockVal = getStockValue(product);
                        const stockBadge = getStockDisplay(stockVal);
                        return (
                          <tr
                            key={product.id}
                            className="border-b border-border transition-colors duration-normal hover:bg-surface-elevated/30"
                          >
                            <td className="p-4 align-middle">
                              <span className="font-medium text-text-primary">{product.name || '-'}</span>
                            </td>
                            <td className="p-4 align-middle text-text-secondary">{product.sku || '-'}</td>
                            <td className="p-4 align-middle text-text-secondary">
                              {product.category_name || product.category || product.categoryName || '-'}
                            </td>
                            <td className="p-4 align-middle">
                              <Badge variant={getConditionBadge(product.condition || '').variant}>
                                {getConditionBadge(product.condition || '').label}
                              </Badge>
                            </td>
                            <td className="p-4 align-middle">
                              {stockVal !== undefined ? stockVal : '-'}
                            </td>
                            <td className="p-4 align-middle">
                              <span className="font-medium text-cyan">{getPrice(product)}</span>
                            </td>
                            <td className="p-4 align-middle">
                              <Badge variant={stockBadge.variant} className="capitalize">
                                {stockBadge.text}
                              </Badge>
                            </td>
                            <td className="p-4 align-middle">
                              <div className="flex items-center justify-end gap-1">
                                <ActionButton
                                  icon={<Eye className="h-4 w-4" />}
                                  label="عرض"
                                  onClick={() => onViewProduct(product)}
                                />
                                <ActionButton
                                  icon={<FileText className="h-4 w-4" />}
                                  label="إضافة عبر فاتورة"
                                  onClick={() => onAddPurchase(product)}
                                />
                                <ActionButton
                                  icon={<Edit className="h-4 w-4" />}
                                  label="تعديل"
                                  onClick={() => onEditProduct(product)}
                                />
                                <ActionButton
                                  icon={<SlidersHorizontal className="h-4 w-4" />}
                                  label="تعديل الحد الأدنى للمخزون"
                                  onClick={() => onEditMinimumStock(product)}
                                />
                                {onViewInventoryLedger && (
                                  <ActionButton
                                    icon={<FileText className="h-4 w-4" />}
                                    label="سجل الحركات"
                                    onClick={() => onViewInventoryLedger(product.id)}
                                  />
                                )}
                                <ActionButton
                                  icon={<Trash2 className="h-4 w-4" />}
                                  label="حذف"
                                  variant="danger"
                                  onClick={() => onDeleteProduct(product.id)}
                                />
                              </div>
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>

                {/* Mobile Cards */}
                <div className="block md:hidden">
                  <div className="grid gap-3 p-4">
                    {displayProducts.map((product: Product) => {
                      const stockVal = getStockValue(product);
                      const stockBadge = getStockDisplay(stockVal);
                      return (
                        <div
                          key={product.id}
                          className="rounded-lg border border-border bg-surface-elevated/30 p-4 transition-colors hover:border-cyan/30"
                        >
                          <div className="mb-3 flex items-start justify-between">
                            <div>
                              <span className="block font-medium text-text-primary">
                                {product.name || '-'}
                              </span>
                              <span className="text-xs text-text-tertiary">{product.sku || '-'}</span>
                            </div>
                            <Badge variant={stockBadge.variant} className="capitalize">
                              {stockBadge.text}
                            </Badge>
                          </div>
                          <div className="grid grid-cols-2 gap-y-2">
                            <div>
                              <span className="text-xs text-text-tertiary">التصنيف</span>
                              <span className="block text-sm text-text-secondary">
                                {product.category_name || product.category || product.categoryName || '-'}
                              </span>
                            </div>
                            <div>
                              <span className="text-xs text-text-tertiary">الحالة</span>
                              <Badge variant={getConditionBadge(product.condition || '').variant} className="mt-1">
                                {getConditionBadge(product.condition || '').label}
                              </Badge>
                            </div>
                            <div>
                              <span className="text-xs text-text-tertiary">المخزون</span>
                              <span className="block text-sm text-yellow font-medium">
                                {stockVal !== undefined ? stockVal : '-'}
                              </span>
                            </div>
                            <div>
                              <span className="text-xs text-text-tertiary">السعر</span>
                              <span className="block text-sm font-semibold text-cyan">{getPrice(product)}</span>
                            </div>
                          </div>
                          <div className="mt-3 flex items-center justify-end gap-2 border-t border-border pt-3">
                            <ActionButton
                              icon={<Eye className="h-4 w-4" />}
                              label="عرض"
                              onClick={() => onViewProduct(product)}
                            />
                            <ActionButton
                              icon={<FileText className="h-4 w-4" />}
                              label="إضافة عبر فاتورة"
                              onClick={() => onAddPurchase(product)}
                            />
                            <ActionButton
                              icon={<Edit className="h-4 w-4" />}
                              label="تعديل"
                              onClick={() => onEditProduct(product)}
                            />
                            <ActionButton
                              icon={<SlidersHorizontal className="h-4 w-4" />}
                              label="تعديل الحد الأدنى للمخزون"
                              onClick={() => onEditMinimumStock(product)}
                            />
                            <ActionButton
                              icon={<Trash2 className="h-4 w-4" />}
                              label="حذف"
                              variant="danger"
                              onClick={() => onDeleteProduct(product.id)}
                            />
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {/* Inventory Items Table */}
      {viewMode === 'items' && (
        <div className="rounded-xl border border-border bg-surface">
          <div className="border-b border-border px-5 py-4 flex items-center gap-2">
            <PackageOpen className="h-5 w-5 text-primary" />
            <h3 className="text-sm font-semibold text-text-primary">عناصر المخزون</h3>
          </div>
          <div>
            {inventoryLoading ? (
              <LoadingSpinner />
            ) : displayInventoryItems.length === 0 ? (
              <EmptyState
                icon={<Inbox className="h-5 w-5" />}
                title="لا توجد عناصر"
                description="لم يتم العثور على عناصر في المخزون"
              />
            ) : (
              <div className="hidden md:block">
                <table className="w-full border-collapse">
                  <thead>
                    <tr className="border-b border-border bg-surface-elevated/50">
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">المنتج</th>
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">الحالة</th>
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">المورد</th>
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">تاريخ الشراء</th>
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">سعر الشراء</th>
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">سعر البيع</th>
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">الموقع</th>
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">الحالة</th>
                      <th className="h-12 px-4 text-end text-xs font-medium text-text-tertiary uppercase tracking-wider">الإجراءات</th>
                    </tr>
                  </thead>
                  <tbody>
                    {displayInventoryItems.map((item: InventoryItem) => {
                      const condition = item.condition?.toLowerCase() || '';
                      const condBadge = getConditionBadge(condition);
                      return (
                        <tr key={item.id} className="border-b border-border transition-colors duration-normal hover:bg-surface-elevated/30">
                          <td className="p-4 align-middle">
                            <span className="font-medium text-text-primary">
                              {item.product_name || item.product?.name || '-'}
                            </span>
                          </td>
                          <td className="p-4 align-middle">
                            <Badge variant={condBadge.variant}>
                              {condBadge.label}
                            </Badge>
                          </td>
                          <td className="p-4 align-middle text-text-secondary">
                            {item.supplier_name || '-'}
                          </td>
                          <td className="p-4 align-middle text-text-secondary">
                            {item.purchase_date ? new Date(item.purchase_date).toLocaleDateString('ar-SA') : '-'}
                          </td>
                          <td className="p-4 align-middle">
                            <span className="font-medium text-text-secondary">
                              {formatPrice(normalizeCurrencyValue(item.purchase_cost ?? 0))}
                            </span>
                          </td>
                          <td className="p-4 align-middle">
                            <span className="font-medium text-cyan">
                              {formatPrice(normalizeCurrencyValue(item.selling_price ?? item.price ?? 0))}
                            </span>
                          </td>
                          <td className="p-4 align-middle text-text-secondary">{item.location || '-'}</td>
                          <td className="p-4 align-middle">
                            <Badge variant={item.status === 'AVAILABLE' ? 'success' : 'warning'}>
                              {item.status === 'AVAILABLE' ? 'متوفر' : item.status}
                            </Badge>
                          </td>
                          <td className="p-4 align-middle">
                            <div className="flex items-center justify-end gap-1">
                              <ActionButton
                                icon={<Eye className="h-4 w-4" />}
                                label="عرض"
                                onClick={() => {}}
                              />
                              {onViewInventoryLedger && (
                                <ActionButton
                                  icon={<FileText className="h-4 w-4" />}
                                  label="سجل الحركات"
                                  onClick={() => onViewInventoryLedger(item.id)}
                                />
                              )}
                              <ActionButton
                                icon={<Edit className="h-4 w-4" />}
                                label="تعديل"
                                onClick={() => {}}
                              />
                              {item.supplier_name && onReorderFromSupplier && (
                                <ActionButton
                                  icon={<RefreshCw className="h-4 w-4" />}
                                  label="إعادة الشراء"
                                  onClick={() => onReorderFromSupplier(item.supplier_id || '', item.product_name || '')}
                                />
                              )}
                              {onViewInvoice && (
                                <ActionButton
                                  icon={<FileText className="h-4 w-4" />}
                                  label="عرض الفاتورة"
                                  onClick={() => onViewInvoice(item.supplier_id || '')}
                                />
                              )}
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
}
