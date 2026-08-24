import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Badge } from '../../../components/ui/badge';
import { EmptyState } from '../../../components/ui/empty-state';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { TableCard, TableCardItem, TableCardActions, ResponsiveTable } from '../../../components/ui/table-card';
import { getButtonSize } from '../../../config/button-sizes';
import { Package, PackageOpen, Eye, Edit, Trash2, Inbox } from 'lucide-react';
import { Product, InventoryItem, ViewMode } from '../types/inventory.types';

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
    return variants[condition] || { label: condition, variant: 'default' };
  };

  return (
    <>
      {/* Products Table */}
      {viewMode === 'products' && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Package className="w-5 h-5 text-primary" />
              المنتجات
            </CardTitle>
          </CardHeader>
          <CardContent>
            {productsLoading ? (
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '256px' }} role="status" aria-label="Loading">
                <div style={{ animation: 'spin 1s linear infinite', borderRadius: '50%', height: '32px', width: '32px', borderBottom: '2px solid var(--primary)' }} />
              </div>
            ) : filteredProducts.length === 0 ? (
              <EmptyState
                icon={<Inbox className="w-5 h-5" />}
                title="لا توجد منتجات"
                description="لم يتم العثور على منتجات تطابق بحثك"
                action={
                  <Button variant="primary" size={getButtonSize('inventory', 'modalAction')} onClick={onClearSearch}>
                    مسح البحث
                  </Button>
                }
              />
            ) : (
              <>
                {/* Desktop Table */}
                <div className="hidden md:block">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>الاسم</TableHead>
                        <TableHead>SKU</TableHead>
                        <TableHead>الفئة</TableHead>
                        <TableHead>الحالة</TableHead>
                        <TableHead>المخزون</TableHead>
                        <TableHead>السعر</TableHead>
                        <TableHead>الحالة</TableHead>
                        <TableHead style={{ textAlign: 'right' }}>الإجراءات</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredProducts.map((product: Product) => (
                        <TableRow key={product.id}>
                          <TableCell style={{ fontWeight: '500' }}>{product.name}</TableCell>
                          <TableCell>{product.sku}</TableCell>
                          <TableCell>{product.category}</TableCell>
                          <TableCell>
                            <Badge {...getConditionBadge(product.condition)} />
                          </TableCell>
                          <TableCell>{product.stock}</TableCell>
                          <TableCell>₪{product.price}</TableCell>
                          <TableCell>
                            <Badge variant={product.stock > 10 ? 'success' : 'warning'}>
                              {product.stock > 10 ? 'متوفر' : 'منخفض'}
                            </Badge>
                          </TableCell>
                          <TableCell style={{ textAlign: 'right' }}>
                            <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => onViewProduct(product)}
                              >
                                <Eye className="w-3.5 h-3.5" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => onEditProduct(product)}
                              >
                                <Edit className="w-3.5 h-3.5" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => onDeleteProduct(product.id)}
                              >
                                <Trash2 className="w-3.5 h-3.5" />
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>

                {/* Mobile Cards */}
                <ResponsiveTable>
                  <div className="grid gap-3">
                    {filteredProducts.map((product: Product) => (
                      <TableCard key={product.id}>
                        <TableCardItem label="الاسم" value={product.name} variant="highlight" />
                        <TableCardItem label="SKU" value={product.sku} />
                        <TableCardItem label="الفئة" value={product.category} />
                        <TableCardItem label="الحالة" value={<Badge {...getConditionBadge(product.condition)} />} />
                        <TableCardItem label="المخزون" value={product.stock} variant={product.stock > 10 ? 'success' : 'warning'} />
                        <TableCardItem label="السعر" value={`₪${product.price}`} variant="highlight" />
                        <TableCardActions>
                          <Button
                            variant="ghost"
                            onClick={() => onViewProduct(product)}
                            className="flex-1"
                          >
                            <Eye className="w-3.5 h-3.5 me-1" />
                            عرض
                          </Button>
                          <Button
                            variant="ghost"
                            onClick={() => onEditProduct(product)}
                            className="flex-1"
                          >
                            <Edit className="w-3.5 h-3.5 me-1" />
                            تعديل
                          </Button>
                          <Button
                            variant="ghost"
                            onClick={() => onDeleteProduct(product.id)}
                            className="flex-1"
                          >
                            <Trash2 className="w-3.5 h-3.5 me-1" />
                            حذف
                          </Button>
                        </TableCardActions>
                      </TableCard>
                    ))}
                  </div>
                </ResponsiveTable>
              </>
            )}
          </CardContent>
        </Card>
      )}

      {/* Inventory Items Table */}
      {viewMode === 'items' && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <PackageOpen className="w-5 h-5 text-cyan" />
              عناصر المخزون
            </CardTitle>
          </CardHeader>
          <CardContent>
            {inventoryLoading ? (
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '256px' }} role="status" aria-label="Loading">
                <div style={{ animation: 'spin 1s linear infinite', borderRadius: '50%', height: '32px', width: '32px', borderBottom: '2px solid var(--primary)' }} />
              </div>
            ) : filteredInventoryItems.length === 0 ? (
              <EmptyState
                icon={<Inbox className="w-5 h-5" />}
                title="لا توجد عناصر"
                description="لم يتم العثور على عناصر في المخزون"
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>المنتج</TableHead>
                    <TableHead>الحالة</TableHead>
                    <TableHead>السعر</TableHead>
                    <TableHead>الموقع</TableHead>
                    <TableHead>الحالة</TableHead>
                    <TableHead>تاريخ الإضافة</TableHead>
                    <TableHead style={{ textAlign: 'right' }}>الإجراءات</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredInventoryItems.map((item: InventoryItem) => (
                    <TableRow key={item.id}>
                      <TableCell style={{ fontWeight: '500' }}>{item.product_name || item.product?.name}</TableCell>
                      <TableCell>
                        <Badge {...getConditionBadge(item.condition?.toLowerCase())} />
                      </TableCell>
                      <TableCell>₪{(item.selling_price / 100).toFixed(2)}</TableCell>
                      <TableCell>{item.location || '-'}</TableCell>
                      <TableCell>
                        <Badge variant={item.status === 'AVAILABLE' ? 'success' : 'warning'}>
                          {item.status === 'AVAILABLE' ? 'متوفر' : item.status}
                        </Badge>
                      </TableCell>
                      <TableCell>{new Date(item.created_at).toLocaleDateString('ar-SA')}</TableCell>
                      <TableCell style={{ textAlign: 'right' }}>
                        <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                          <Button variant="ghost" size="icon">
                            <Eye className="w-3.5 h-3.5" />
                          </Button>
                          <Button variant="ghost" size="icon">
                            <Edit className="w-3.5 h-3.5" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}
    </>
  );
}