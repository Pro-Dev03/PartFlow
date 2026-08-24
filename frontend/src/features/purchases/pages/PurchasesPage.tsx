import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { purchasesApi, suppliersApi, productsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { StatCard } from '../../../components/ui/stat-card';
import { Select } from '../../../components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { Modal } from '../../../components/ui/modal';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { getButtonSize } from '../../../config/button-sizes';
import { ItemInputMethod, type ItemInputMethodType } from '../../../components/ui/item-input-method';
import { CameraScanner } from '../../../components/ui/camera-scanner';
import {
  ShoppingCart,
  Search,
  Plus,
  Filter,
  Eye,
  Truck,
  Package,
  Calendar,
  Download,
  Printer,
  Scan,
  X,
  Check,
  AlertCircle,
  Camera
} from 'lucide-react';

export function PurchasesPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isManualAdd, setIsManualAdd] = useState(false);
  const [barcodeInput, setBarcodeInput] = useState('');
  const [selectedSupplier, setSelectedSupplier] = useState('');
  const [purchaseItems, setPurchaseItems] = useState<any[]>([]);
  const [productSearchQuery, setProductSearchQuery] = useState('');

  const { data: productsData } = useQuery({
    queryKey: ['products', productSearchQuery],
    queryFn: () => productsApi.list({ search: productSearchQuery }),
    enabled: isManualAdd && productSearchQuery.length > 0,
  });

  const searchedProducts = (productsData?.data as any[]) || [];

  const { data: purchasesData, isLoading } = useQuery({
    queryKey: ['purchases'],
    queryFn: () => purchasesApi.list({ page: 1, per_page: 100 }),
  });

  const { data: suppliersData, isLoading: suppliersLoading } = useQuery({
    queryKey: ['suppliers'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 100 }),
  });

  const suppliers = (suppliersData?.data as any[]) || [];
  const purchases = (purchasesData?.data as any[]) || [];

  // Calculate totals
  const totalCost = purchaseItems.reduce((sum, item) => sum + (item.quantity * item.unitCost), 0);

  // Handle barcode scan
  const handleBarcodeScan = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!barcodeInput.trim()) return;

    try {
      const response = await productsApi.get(`/barcode/${barcodeInput.trim()}`);
      const product = response.data;

      // Check if product already in items
      const existingItem = purchaseItems.find((item) => item.product_id === product.id);
      if (existingItem) {
        setPurchaseItems(purchaseItems.map((item) =>
          item.product_id === product.id
            ? { ...item, quantity: item.quantity + 1 }
            : item
        ));
      } else {
        setPurchaseItems([
          ...purchaseItems,
          {
            product_id: product.id,
            product_name: product.name,
            quantity: 1,
            unit_cost: product.cost_price || 0,
            condition: 'new',
          },
        ]);
      }

      setBarcodeInput('');
    } catch (error) {
      console.error('Error scanning barcode:', error);
      alert('المنتج غير موجود');
    }
  };

  // Handle manual add
  const handleManualAdd = (product: any) => {
    const existingItem = purchaseItems.find((item) => item.product_id === product.id);
    if (existingItem) {
      setPurchaseItems(purchaseItems.map((item) =>
        item.product_id === product.id
          ? { ...item, quantity: item.quantity + 1 }
          : item
      ));
    } else {
      setPurchaseItems([
        ...purchaseItems,
        {
          product_id: product.id,
          product_name: product.name,
          quantity: 1,
          unit_cost: product.cost_price || 0,
          condition: 'new',
        },
      ]);
    }
  };

  // Remove item
  const handleRemoveItem = (productId: string) => {
    setPurchaseItems(purchaseItems.filter((item) => item.product_id !== productId));
  };

  // Update item quantity
  const handleUpdateQuantity = (productId: string, quantity: number) => {
    if (quantity <= 0) {
      handleRemoveItem(productId);
      return;
    }
    setPurchaseItems(
      purchaseItems.map((item) =>
        item.product_id === productId ? { ...item, quantity } : item
      )
    );
  };

  // Create purchase mutation
  const createPurchaseMutation = useMutation({
    mutationFn: (data: any) => purchasesApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchases'] });
      setIsCreateModalOpen(false);
      setPurchaseItems([]);
      setSelectedSupplier('');
      alert('تم إنشاء الشراء بنجاح');
    },
    onError: (error) => {
      console.error('Error creating purchase:', error);
      alert('فشل إنشاء الشراء');
    },
  });

  // Handle create purchase
  const handleCreatePurchase = () => {
    if (!selectedSupplier) {
      alert('يرجى اختيار المورد');
      return;
    }
    if (purchaseItems.length === 0) {
      alert('يرجى إضافة عناصر للشراء');
      return;
    }

    createPurchaseMutation.mutate({
      supplier_id: selectedSupplier,
      invoice_number: `PO-${Date.now()}`,
      purchase_date: new Date().toISOString(),
      items: purchaseItems.map((item) => ({
        product_id: item.product_id,
        quantity: item.quantity,
        unit_cost: item.unit_cost,
        condition: item.condition,
      })),
    });
  };

  // Receive purchase mutation
  const receivePurchaseMutation = useMutation({
    mutationFn: (purchaseId: string) => purchasesApi.receive(purchaseId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchases'] });
      alert('تم استلام البضاعة بنجاح');
    },
    onError: (error) => {
      console.error('Error receiving purchase:', error);
      alert('فشل استلام البضاعة');
    },
  });

  // Handle receive purchase
  const handleReceivePurchase = (purchaseId: string) => {
    if (confirm('هل أنت متأكد من استلام البضاعة؟ سيتم إنشاء عناصر المخزون تلقائياً.')) {
      receivePurchaseMutation.mutate(purchaseId);
    }
  };

  const filteredPurchases = purchases.filter((purchase: any) => {
    const matchesSearch = 
      purchase.supplier?.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      purchase.id.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = !statusFilter || purchase.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const getStatusBadge = (status: string) => {
    const variants: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
      pending: { label: 'قيد الانتظار', variant: 'secondary' },
      ordered: { label: 'تم الطلب', variant: 'outline' },
      received: { label: 'تم الاستلام', variant: 'default' },
      cancelled: { label: 'ملغي', variant: 'destructive' },
    };
    return variants[status] || { label: status, variant: 'default' };
  };

  const handleExport = () => {
    const dataToExport = purchases.map((purchase: any) => ({
      'التاريخ': purchase.date,
      'المورد': purchase.supplier,
      'الحالة': getStatusBadge(purchase.status).label,
      'التكلفة': purchase.totalCost
    }));
    exportToCSV(dataToExport, `purchases-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    const dataToPrint = purchases.map((purchase: any) => ({
      'التاريخ': purchase.date,
      'المورد': purchase.supplier,
      'الحالة': getStatusBadge(purchase.status).label,
      'التكلفة': purchase.totalCost
    }));
    printTable(dataToPrint, ['التاريخ', 'المورد', 'الحالة', 'التكلفة'], 'تقرير المشتريات');
  };

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Purchase Management"
        title={t('purchases.title')}
        description="إدارة المشتريات والطلبات"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="primary" size={getButtonSize('customers', 'headerActions')} className="gap-2" onClick={() => setIsCreateModalOpen(true)}>
              <Plus className="w-4 h-4" />
              {t('purchases.newPurchase')}
            </Button>
            <Button variant="secondary" size={getButtonSize('customers', 'headerActions')} onClick={handleExport} className="gap-2">
              <Download className="w-4 h-4" />
              تصدير
            </Button>
            <Button variant="secondary" size={getButtonSize('customers', 'headerActions')} onClick={handlePrint} className="gap-2">
              <Printer className="w-4 h-4" />
              طباعة
            </Button>
          </div>
        }
      />

      {/* Stats Cards - Futuristic + Clean */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-md">
        <StatCard title="إجمالي المشتريات" value={purchases.length} icon={ShoppingCart} variant="featured" />
        <StatCard title="قيد الانتظار" value={purchases.filter((p: any) => p.status === 'pending').length} icon={Calendar} variant="warning" />
        <StatCard title="تم الاستلام" value={purchases.filter((p: any) => p.status === 'received').length} icon={Package} variant="success" />
        <StatCard title="إجمالي التكلفة" value={`₪${(purchases as any[]).reduce((sum: number, p: any) => sum + p.totalCost, 0).toLocaleString()}`} icon={Truck} variant="default" />
      </div>

      {/* Search and Filters */}
      <Card>
        <CardContent className="p-lg">
          <div className="flex flex-col md:flex-row gap-md">
            <div className="flex-1 relative">
              <Search className="absolute inset-y-0 end-3 w-4 h-4 text-cyan" />
              <Input
                placeholder={t('common.search')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pe-10"
              />
            </div>
            <div className="flex gap-sm">
              <Select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                options={[
                  { value: '', label: 'كل الحالات' },
                  { value: 'pending', label: 'قيد الانتظار' },
                  { value: 'ordered', label: 'تم الطلب' },
                  { value: 'received', label: 'تم الاستلام' },
                  { value: 'cancelled', label: 'ملغي' },
                ]}
                emptyMessage="لا توجد حالات"
              />
              <Button variant="secondary" className="gap-2">
                <Filter className="w-4 h-4" />
                {t('common.filter')}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Purchases Table */}
      <Card>
        <CardHeader>
          <CardTitle>قائمة المشتريات</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan" />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>رقم الطلب</TableHead>
                  <TableHead>المورد</TableHead>
                  <TableHead>القطع</TableHead>
                  <TableHead>إجمالي التكلفة</TableHead>
                  <TableHead>المدفوع</TableHead>
                  <TableHead>المتبقي</TableHead>
                  <TableHead>التاريخ المتوقع</TableHead>
                  <TableHead>الحالة</TableHead>
                  <TableHead className="text-start">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredPurchases.map((purchase: any) => {
                  const statusBadge = getStatusBadge(purchase.status);
                  return (
                    <TableRow key={purchase.id}>
                      <TableCell className="font-medium">{purchase.id}</TableCell>
                      <TableCell>{purchase.supplier?.name}</TableCell>
                      <TableCell>{purchase.items?.length || 0} قطع</TableCell>
                      <TableCell>₪{purchase.totalCost?.toLocaleString()}</TableCell>
                      <TableCell className="text-green">₪{purchase.paidAmount?.toLocaleString()}</TableCell>
                      <TableCell>₪{purchase.remainingAmount?.toLocaleString()}</TableCell>
                      <TableCell>
                        {purchase.expectedDate 
                          ? new Date(purchase.expectedDate).toLocaleDateString('ar-SA')
                          : '-'
                        }
                      </TableCell>
                      <TableCell>
                        <Badge variant={statusBadge.variant}>
                          {statusBadge.label}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-start">
                        <div className="flex gap-2">
                          {purchase.status === 'pending' && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleReceivePurchase(purchase.id)}
                              className="text-green-600 hover:text-green-700"
                            >
                              <Check className="w-4 h-4" />
                            </Button>
                          )}
                          <Button variant="ghost" size="sm">
                            <Eye className="w-4 h-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Create Purchase Modal */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="إنشاء شراء جديد"
        size="xl"
      >
        <div className="space-y-6">
          {/* Supplier Selection */}
          <div>
            <label className="block text-sm font-medium text-text mb-2">اختر المورد</label>
            <Select
              value={selectedSupplier}
              onChange={(e) => setSelectedSupplier(e.target.value)}
              loading={suppliersLoading}
              options={[
                { value: '', label: 'اختر المورد...' },
                ...suppliers.map((s) => ({ value: s.id, label: s.name })),
              ]}
              emptyMessage="لا يوجد موردين"
            />
          </div>

          {/* Add Item Section */}
          <div className="border border-border rounded-lg p-4">
            <div className="flex gap-2 mb-4">
              <Button
                variant={!isManualAdd ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => setIsManualAdd(false)}
                className="flex-1"
              >
                <Scan className="w-4 h-4 ml-2" />
                مسح الباركود
              </Button>
              <Button
                variant={isManualAdd ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => setIsManualAdd(true)}
                className="flex-1"
              >
                <Plus className="w-4 h-4 ml-2" />
                إضافة يدوية
              </Button>
            </div>

            {!isManualAdd ? (
              <form onSubmit={handleBarcodeScan} className="flex gap-2">
                <Input
                  placeholder="مسح الباركود أو اكتب الرقم..."
                  value={barcodeInput}
                  onChange={(e) => setBarcodeInput(e.target.value)}
                  className="flex-1"
                  autoFocus
                />
                <Button type="submit" variant="primary">
                  <Scan className="w-4 h-4" />
                </Button>
              </form>
            ) : (
              <div className="space-y-3">
                <div className="flex gap-2">
                  <Input
                    placeholder="ابحث عن المنتج..."
                    value={productSearchQuery}
                    onChange={(e) => setProductSearchQuery(e.target.value)}
                    className="flex-1"
                  />
                  <Button variant="primary">
                    <Search className="w-4 h-4" />
                  </Button>
                </div>
                {searchedProducts.length > 0 && (
                  <div className="scrollable-card-sm border border-border rounded-lg">
                    {searchedProducts.map((product) => (
                      <div
                        key={product.id}
                        className="p-3 hover:bg-surface/30 cursor-pointer border-b border-border last:border-b-0"
                        onClick={() => {
                          handleManualAdd(product);
                          setProductSearchQuery('');
                        }}
                      >
                        <div className="font-medium">{product.name}</div>
                        <div className="text-sm text-text-muted">₪{product.cost_price?.toFixed(2) || 0}</div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Items List */}
          {purchaseItems.length > 0 && (
            <div>
              <h4 className="text-sm font-medium text-text mb-3">العناصر المضافة</h4>
              <div className="border border-border rounded-lg overflow-hidden">
                <div className="horizontal-scroll">
                  <table className="w-full">
                  <thead className="bg-surface/30">
                    <tr>
                      <th className="text-right p-3 text-sm font-medium">المنتج</th>
                      <th className="text-right p-3 text-sm font-medium">الكمية</th>
                      <th className="text-right p-3 text-sm font-medium">سعر الوحدة</th>
                      <th className="text-right p-3 text-sm font-medium">الإجمالي</th>
                      <th className="text-right p-3 text-sm font-medium">إجراء</th>
                    </tr>
                  </thead>
                  <tbody>
                    {purchaseItems.map((item) => (
                      <tr key={item.product_id} className="border-t border-border">
                        <td className="p-3">{item.product_name}</td>
                        <td className="p-3">
                          <Input
                            type="number"
                            value={item.quantity}
                            onChange={(e) => handleUpdateQuantity(item.product_id, parseInt(e.target.value) || 0)}
                            className="w-20"
                            min="1"
                          />
                        </td>
                        <td className="p-3">₪{item.unit_cost.toFixed(2)}</td>
                        <td className="p-3">₪{(item.quantity * item.unit_cost).toFixed(2)}</td>
                        <td className="p-3">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRemoveItem(item.product_id)}
                            className="text-red-600 hover:text-red-700"
                          >
                            <X className="w-4 h-4" />
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                </div>
              </div>
            </div>
          )}

          {/* Total */}
          {purchaseItems.length > 0 && (
            <div className="flex justify-between items-center p-4 bg-surface/30 rounded-lg">
              <span className="text-sm font-medium">الإجمالي:</span>
              <span className="text-xl font-bold text-cyan">₪{totalCost.toFixed(2)}</span>
            </div>
          )}

          {/* Actions */}
          <div className="flex justify-end gap-3">
            <Button
              variant="secondary"
              onClick={() => {
                setIsCreateModalOpen(false);
                setPurchaseItems([]);
                setSelectedSupplier('');
              }}
            >
              إلغاء
            </Button>
            <Button
              variant="primary"
              onClick={handleCreatePurchase}
              disabled={createPurchaseMutation.isPending}
            >
              {createPurchaseMutation.isPending ? 'جاري الإنشاء...' : 'إنشاء الشراء'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}