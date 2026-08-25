import { Badge } from '../../../components/ui/badge';
import { EmptyState } from '../../../components/ui/empty-state';
import { Package, PackageOpen, Eye, Edit, Trash2, Inbox } from 'lucide-react';
import { Product, InventoryItem, ViewMode } from '../types/inventory.types';
import { formatPrice, formatPriceFromCents } from '../../../utils';

interface InventoryListProps {
  viewMode: ViewMode;
  filteredProducts: Product[];
  filteredInventoryItems: InventoryItem[];
  productsLoading: boolean;
  inventoryLoading: boolean;
  searchQuery: string;
  onViewProduct: (product: Product) => void;
  onEditProduct: (product: Product) => void;
  onDeleteProduct: (productId: string) => void;
  onClearSearch: () => void;
}

interface ActionButtonProps {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
  variant?: 'default' | 'danger';
}

function ActionButton({ icon, label, onClick, variant = 'default' }: ActionButtonProps) {
  const baseClasses = 'flex h-8 w-8 items-center justify-center rounded-lg text-text-secondary opacity-60 hover:opacity-100 hover:bg-surface-elevated hover:text-text-primary transition-all duration-150';
  const variantClasses = variant === 'danger'
    ? 'hover:text-red hover:bg-red/8'
    : baseClasses;

  return (
    <button
      type="button"
      onClick={onClick}
      className={variant === 'danger' ? variantClasses : baseClasses}
      aria-label={label}
    >
      {icon}
    </button>
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
  if (product.sellingPrice !== undefined && product.sellingPrice !== null) {
    const val = product.sellingPrice;
    return val > 1000 ? formatPriceFromCents(val) : formatPrice(val);
  }
  if (product.selling_price !== undefined && product.selling_price !== null) {
    const val = product.selling_price;
    return val > 1000 ? formatPriceFromCents(val) : formatPrice(val);
  }
  if (product.price !== undefined && product.price !== null) {
    const val = product.price;
    return val > 1000 ? formatPriceFromCents(val) : formatPrice(val);
  }
  return formatPrice(0);
}

export function InventoryList({
  viewMode,
  filteredProducts,
  filteredInventoryItems,
  productsLoading,
  inventoryLoading,
  searchQuery,
  onViewProduct,
  onEditProduct,
  onDeleteProduct,
  onClearSearch,
}: InventoryListProps) {
  const getConditionBadge = (condition: string) => {
    const variants: Record<string, { label: string; variant: 'default' | 'success' | 'warning' | 'danger' | 'info' | 'secondary' | 'outline' }> = {
      new: { label: 'جديد', variant: 'success' },
      used: { label: 'مستعمل', variant: 'secondary' },
      refurbished: { label: 'مجدد', variant: 'info' },
      parts_only: { label: 'قطع فقط', variant: 'danger' },
    };
    return variants[condition] || { label: condition || 'افتراضي', variant: 'outline' };
  };

  const getStockValue = (product: any): number | undefined => {
    if (product.stock !== undefined && product.stock !== null) return product.stock;
    if (product.stock_quantity !== undefined && product.stock_quantity !== null) return product.stock_quantity;
    if (product.quantity !== undefined && product.quantity !== null) return product.quantity;
    return undefined;
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
            ) : filteredProducts.length === 0 ? (
              <EmptyState
                icon={<Inbox className="h-5 w-5" />}
                title="لا توجد منتجات"
                description="لم يتم العثور على منتجات تطابق بحثك"
                action={
                  <button
                    type="button"
                    onClick={onClearSearch}
                    className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white hover:bg-primary/90 transition-colors"
                  >
                    مسح البحث
                  </button>
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
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">الفئة</th>
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">الحالة</th>
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">المخزون</th>
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">السعر</th>
                        <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">الحالة</th>
                        <th className="h-12 px-4 text-end text-xs font-medium text-text-tertiary uppercase tracking-wider">الإجراءات</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredProducts.map((product: Product) => {
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
                              {product.category || product.category_name || product.categoryName || '-'}
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
                                  icon={<Edit className="h-4 w-4" />}
                                  label="تعديل"
                                  onClick={() => onEditProduct(product)}
                                />
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
                    {filteredProducts.map((product: Product) => {
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
                              <span className="text-xs text-text-tertiary">الفئة</span>
                              <span className="block text-sm text-text-secondary">
                                {product.category || product.category_name || product.categoryName || '-'}
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
                              icon={<Edit className="h-4 w-4" />}
                              label="تعديل"
                              onClick={() => onEditProduct(product)}
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
            ) : filteredInventoryItems.length === 0 ? (
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
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">السعر</th>
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">الموقع</th>
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">الحالة</th>
                      <th className="h-12 px-4 text-start text-xs font-medium text-text-tertiary uppercase tracking-wider">تاريخ الإضافة</th>
                      <th className="h-12 px-4 text-end text-xs font-medium text-text-tertiary uppercase tracking-wider">الإجراءات</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredInventoryItems.map((item: InventoryItem) => {
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
                          <td className="p-4 align-middle">
                            <span className="font-medium text-cyan">
                              {formatPriceFromCents(item.selling_price)}
                            </span>
                          </td>
                          <td className="p-4 align-middle text-text-secondary">{item.location || '-'}</td>
                          <td className="p-4 align-middle">
                            <Badge variant={item.status === 'AVAILABLE' ? 'success' : 'warning'}>
                              {item.status === 'AVAILABLE' ? 'متوفر' : item.status}
                            </Badge>
                          </td>
                          <td className="p-4 align-middle text-text-secondary">
                            {new Date(item.created_at).toLocaleDateString('ar-SA')}
                          </td>
                          <td className="p-4 align-middle">
                            <div className="flex items-center justify-end gap-1">
                              <ActionButton
                                icon={<Eye className="h-4 w-4" />}
                                label="عرض"
                                onClick={() => {}}
                              />
                              <ActionButton
                                icon={<Edit className="h-4 w-4" />}
                                label="تعديل"
                                onClick={() => {}}
                              />
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
