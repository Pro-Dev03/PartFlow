import { useState, useEffect, useMemo, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { Badge } from '../../../components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { PageHeader } from '../../../components/ui/page-header';
import { Modal } from '../../../components/ui/modal';
import { ConfirmDialog } from '../../../components/ui/confirm-dialog';
import { Product, Supplier, Category } from '../../../types/models';
import {
  Search,
  Scan,
  Trash2,
  ShoppingCart,
  Truck,
  CheckCircle2,
  Package,
  FileText,
  Save,
  Sparkles,
  DollarSign,
  Box,
  Edit,
} from 'lucide-react';
import { purchasesApi, suppliersApi, productsApi, categoriesApi } from '../../../services/api/endpoints';
import { usePurchases } from '../hooks/usePurchases';
import { PurchaseItem } from '../types/purchases.types';
import { toast } from 'sonner';

interface LineItem extends PurchaseItem {
  key: string;
}

export function EditPurchasePage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const queryClient = useQueryClient();
  const [selectedSupplier, setSelectedSupplier] = useState('');
  const [items, setItems] = useState<LineItem[]>([]);
  const [productSearchQuery, setProductSearchQuery] = useState('');
  const [barcodeInput, setBarcodeInput] = useState('');
  const [activeTab, setActiveTab] = useState<'scan' | 'manual'>('manual');
  const [invoiceNumber, setInvoiceNumber] = useState('');
  const [purchaseDate, setPurchaseDate] = useState(new Date().toISOString().split('T')[0]);
  const [expectedDate, setExpectedDate] = useState('');
  const [notes, setNotes] = useState('');
  const [receiveImmediately, setReceiveImmediately] = useState(false);
  const [productToDelete, setProductToDelete] = useState<Product | null>(null);
  
  // Manual product addition states
  const [isManualProductModalOpen, setIsManualProductModalOpen] = useState(false);
  const [manualProductData, setManualProductData] = useState({
    name: '',
    sku: '',
    barcode: '',
    category_id: '',
    cost_price: '',
    selling_price: '',
    min_stock: '',
    description: '',
  });

  // Fetch purchase data
  const { data: purchaseData, isLoading: purchaseLoading } = useQuery({
    queryKey: ['purchase', id],
    queryFn: () => purchasesApi.get(id || ''),
    enabled: !!id,
  });

  const { data: suppliersData, isLoading: suppliersLoading } = useQuery({
    queryKey: ['suppliers'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 10 }),
  });

  const { data: productsData } = useQuery({
    queryKey: ['products', productSearchQuery],
    queryFn: () => productsApi.list({ search: productSearchQuery, page: 1, per_page: 20 }),
  });

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });

  const createProductMutation = useMutation({
    mutationFn: (data: any) => productsApi.create(data),
    onSuccess: (response) => {
      const newProduct = response.data;
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      
      // Automatically add the new product to the purchase items
      setItems((prev) => [
        ...prev,
        {
          key: `${newProduct.id}-${Date.now()}`,
          product_id: newProduct.id,
          product_name: newProduct.name,
          quantity: 1,
          unit_cost: newProduct.cost_price || 0,
          condition: 'new',
        },
      ]);
      
      setIsManualProductModalOpen(false);
      setManualProductData({
        name: '',
        sku: '',
        barcode: '',
        category_id: '',
        cost_price: '',
        selling_price: '',
        min_stock: '',
        description: '',
      });
    },
  });

  const deleteProductMutation = useMutation({
    mutationFn: (productId: string) => productsApi.delete(productId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      toast.success('تم حذف المنتج بنجاح');
    },
    onError: (error) => {
      console.error('Error deleting product:', error);
      toast.error('حدث خطأ أثناء حذف المنتج');
    },
  });

  const searchedProducts = (productsData?.data?.products as Product[]) || [];
  const suppliers = (suppliersData?.data as Supplier[]) || [];
  const categories = (categoriesData?.data as Category[]) || [];

  const { updatePurchaseMutation, receivePurchaseMutation } = usePurchases();

  // Load purchase data when available
  useEffect(() => {
    if (purchaseData?.data) {
      const purchase = purchaseData.data;
      setSelectedSupplier(purchase.supplier_id || '');
      setInvoiceNumber(purchase.invoice_number || '');
      setPurchaseDate(purchase.purchase_date ? new Date(purchase.purchase_date).toISOString().split('T')[0] : new Date().toISOString().split('T')[0]);
      setExpectedDate(purchase.expected_delivery_date ? new Date(purchase.expected_delivery_date).toISOString().split('T')[0] : '');
      setNotes(purchase.notes || '');
      
      // Load items
      if (purchase.items && Array.isArray(purchase.items)) {
        setItems(purchase.items.map((item: any, index: number) => ({
          key: `${item.product_id}-${index}`,
          product_id: item.product_id,
          product_name: item.product_name || item.product?.name || '',
          quantity: Number(item.quantity) || 0,
          unit_cost: Number(item.unit_cost) || 0,
          condition: item.condition || 'new',
        })));
      }
    }
  }, [purchaseData]);

  const totalCost = useMemo(
    () => items.reduce((sum, item) => sum + item.quantity * item.unit_cost, 0),
    [items]
  );

  const totalQuantity = useMemo(
    () => items.reduce((sum, item) => sum + item.quantity, 0),
    [items]
  );

  const handleBarcodeScan = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    if (!barcodeInput.trim()) return;

    try {
      const response = await productsApi.get(`/barcode/${barcodeInput.trim()}`);
      const product = response.data;

      setItems((prev) => {
        const existingIndex = prev.findIndex((item) => item.product_id === product.id);
        if (existingIndex >= 0) {
          return prev.map((item, idx) =>
            idx === existingIndex ? { ...item, quantity: item.quantity + 1 } : item
          );
        }
        return [
          ...prev,
          {
            key: `${product.id}-${Date.now()}`,
            product_id: product.id,
            product_name: product.name,
            quantity: 1,
            unit_cost: product.cost_price || 0,
            condition: 'new',
          },
        ];
      });

      setBarcodeInput('');
    } catch (error) {
      console.error('Error scanning barcode:', error);
    }
  }, [barcodeInput]);

  const handleManualAdd = useCallback((product: any) => {
    setItems((prev) => {
      const existingIndex = prev.findIndex((item) => item.product_id === product.id);
      if (existingIndex >= 0) {
        return prev.map((item, idx) =>
          idx === existingIndex ? { ...item, quantity: item.quantity + 1 } : item
        );
      }
      return [
        ...prev,
        {
          key: `${product.id}-${Date.now()}`,
          product_id: product.id,
          product_name: product.name,
          quantity: 1,
          unit_cost: product.cost_price || 0,
          condition: 'new',
        },
      ];
    });
    setProductSearchQuery('');
  }, []);

  const handleRemoveItem = useCallback((key: string) => {
    setItems((prev) => prev.filter((item) => item.key !== key));
  }, []);

  const handleUpdateQuantity = useCallback((key: string, quantity: number) => {
    setItems((prev) => {
      if (quantity <= 0) return prev.filter((item) => item.key !== key);
      return prev.map((item) => (item.key === key ? { ...item, quantity } : item));
    });
  }, []);

  const handleUpdateUnitCost = useCallback((key: string, unit_cost: number) => {
    setItems((prev) =>
      prev.map((item) => (item.key === key ? { ...item, unit_cost: unit_cost || 0 } : item))
    );
  }, []);

  const handleUpdateCondition = useCallback(
    (key: string, condition: 'new' | 'used' | 'refurbished') => {
      setItems((prev) =>
        prev.map((item) => (item.key === key ? { ...item, condition } : item))
      );
    },
    []
  );

  const handleManualProductCreate = useCallback(async () => {
    if (!manualProductData.name.trim()) {
      toast.error('يرجى إدخال اسم المنتج');
      return;
    }
    const costPrice = parseFloat(manualProductData.cost_price.toString());
    if (!costPrice || costPrice <= 0) {
      toast.error('يرجى إدخال سعر التكلفة');
      return;
    }
    const sellingPrice = parseFloat(manualProductData.selling_price.toString());
    if (!sellingPrice || sellingPrice <= 0) {
      toast.error('يرجى إدخال سعر البيع');
      return;
    }

    const productData = {
      name: manualProductData.name,
      sku: manualProductData.sku || `SKU-${Date.now()}`,
      barcode: manualProductData.barcode || undefined,
      category_id: manualProductData.category_id || undefined,
      cost_price: costPrice,
      selling_price: sellingPrice,
      min_stock: parseInt(manualProductData.min_stock.toString()) || 0,
      description: manualProductData.description || undefined,
    };

    createProductMutation.mutate(productData);
  }, [manualProductData, createProductMutation]);

  const handleUpdatePurchase = useCallback(async () => {
    if (!selectedSupplier) {
      toast.error('يرجى اختيار المورد');
      return;
    }
    if (items.length === 0) {
      toast.error('يرجى إضافة عناصر للشراء');
      return;
    }

    const formData = {
      supplier_id: selectedSupplier,
      invoice_number: invoiceNumber || `PO-${Date.now()}`,
      purchase_date: new Date(purchaseDate).toISOString(),
      expected_delivery_date: expectedDate ? new Date(expectedDate).toISOString() : undefined,
      notes: notes || undefined,
      items: items.map((item) => ({
        product_id: item.product_id,
        quantity: item.quantity,
        unit_cost: item.unit_cost,
        condition: item.condition,
      })),
    };

    updatePurchaseMutation.mutate({ id: id || '', data: formData }, {
      onSuccess: (response) => {
        const purchaseId = response?.data?.id;
        if (purchaseId && receiveImmediately) {
          receivePurchaseMutation.mutate(purchaseId, {
            onSuccess: () => {
              queryClient.invalidateQueries({ queryKey: ['purchases'] });
              queryClient.invalidateQueries({ queryKey: ['inventory'] });
              queryClient.invalidateQueries({ queryKey: ['products'] });
              navigate('/app/purchases');
            },
          });
        } else {
          queryClient.invalidateQueries({ queryKey: ['purchases'] });
          navigate('/app/purchases');
        }
      },
    });
  }, [
    id,
    selectedSupplier,
    items,
    invoiceNumber,
    purchaseDate,
    expectedDate,
    notes,
    receiveImmediately,
    updatePurchaseMutation,
    receivePurchaseMutation,
    queryClient,
    navigate,
  ]);

  const isSubmitting =
    updatePurchaseMutation.isPending || receivePurchaseMutation.isPending;

  if (purchaseLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <PageHeader
        eyebrow="Purchase Management"
        title="تعديل الشراء"
        description="تعديل طلب شراء موجود"
        actions={
          <Button variant="secondary" onClick={() => navigate('/app/purchases')}>
            عودة للمشتريات
          </Button>
        }
      />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - Main Form */}
        <div className="lg:col-span-2 space-y-6">
          {/* Supplier and Invoice Info */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileText className="w-5 h-5 text-cyan" />
                معلومات الفاتورة
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-text mb-2">المورد *</label>
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
                <div>
                  <label className="block text-sm font-medium text-text mb-2">رقم الفاتورة</label>
                  <Input
                    value={invoiceNumber}
                    onChange={(e) => setInvoiceNumber(e.target.value)}
                    placeholder="PO-..."
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-text mb-2">تاريخ الشراء</label>
                  <Input
                    type="date"
                    value={purchaseDate}
                    onChange={(e) => setPurchaseDate(e.target.value)}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-text mb-2">تاريخ الاستلام المتوقع</label>
                  <Input
                    type="date"
                    value={expectedDate}
                    onChange={(e) => setExpectedDate(e.target.value)}
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Items Section */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <ShoppingCart className="w-5 h-5 text-cyan" />
                  إضافة القطع
                </CardTitle>
                <Badge variant="secondary">
                  {items.length} عنصر • {totalQuantity} قطعة
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Tab Selection */}
              <div className="flex gap-2">
                <Button
                  variant={activeTab === 'scan' ? 'primary' : 'secondary'}
                  size="sm"
                  onClick={() => setActiveTab('scan')}
                  className="flex-1"
                >
                  <Scan className="w-4 h-4 ml-2" />
                  مسح الباركود
                </Button>

                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => setIsManualProductModalOpen(true)}
                  className="flex items-center gap-2"
                >
                  <Sparkles className="w-4 h-4" />
                  إضافة قطعة جديدة
                </Button>
              </div>

              {/* Scan Tab */}
              {activeTab === 'scan' ? (
                <form onSubmit={handleBarcodeScan} className="pf-barcode-row">
                  <Input
                    placeholder="امسح الباركود أو اكتب الرقم..."
                    value={barcodeInput}
                    onChange={(e) => setBarcodeInput(e.target.value)}
                    className="min-w-0 flex-1"
                    autoFocus
                  />
                  <Button type="submit" variant="primary" className="pf-barcode-submit">
                    <Scan className="w-4 h-4" />
                  </Button>
                </form>
              ) : (
                /* Manual Tab */
                <div className="space-y-3">
                  <div className="flex gap-2">
                    <div className="relative flex-1">
                      <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                      <Input
                        placeholder="ابحث عن المنتج بالاسم أو الباركود..."
                        value={productSearchQuery}
                        onChange={(e) => setProductSearchQuery(e.target.value)}
                        className="pr-10"
                        autoFocus
                      />
                    </div>
                  </div>
                  {searchedProducts.length > 0 ? (
                    <div className="border border-border rounded-lg max-h-64 overflow-y-auto">
                      {searchedProducts.map((product) => (
                        <div
                          key={product.id}
                          className="p-4 hover:bg-surface/30 cursor-pointer border-b border-border last:border-b-0 flex justify-between items-center"
                          onClick={() => handleManualAdd(product)}
                        >
                          <div className="flex-1">
                            <div className="font-medium">{product.name}</div>
                            <div className="text-sm text-text-muted">
                              {product.barcode ? `الباركود: ${product.barcode}` : product.sku ? `SKU: ${product.sku}` : product.id}
                            </div>
                          </div>
                          <div className="flex items-center gap-4">
                            <div className="text-sm font-medium text-cyan">
                              ₪{product.cost_price?.toFixed(2) || '0.00'}
                            </div>
                            <div className="flex gap-2" onClick={(e) => e.stopPropagation()}>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => {
                                  navigate(`/app/inventory`, { state: { editProduct: product } });
                                }}
                                className="text-blue-600 hover:text-blue-700"
                              >
                                <Edit className="w-4 h-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => {
                                  if (product.sales_count > 0) {
                                    toast.error('لا يمكن حذف هذا المنتج لأنه تم بيعه مسبقاً. يمكن حذف المنتجات التي لم يتم بيعها فقط.');
                                    return;
                                  }
                                  setProductToDelete(product);
                                }}
                                className={product.sales_count > 0 ? "text-gray-400 cursor-not-allowed" : "text-red-600 hover:text-red-700"}
                                title={product.sales_count > 0 ? "لا يمكن حذف منتج تم بيعه" : "حذف المنتج نهائياً"}
                                disabled={product.sales_count > 0}
                              >
                                <Trash2 className="w-4 h-4" />
                              </Button>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-center py-8 text-text-muted">
                      {productSearchQuery ? 'لا توجد نتائج للبحث' : 'ابحث عن منتج للإضافة'}
                    </div>
                  )}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Items Table */}
          {items.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>العناصر المضافة ({items.length})</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead className="bg-surface/30">
                      <tr>
                        <th className="text-right p-3 text-sm font-medium">المنتج</th>
                        <th className="text-right p-3 text-sm font-medium">الحالة</th>
                        <th className="text-right p-3 text-sm font-medium">الكمية</th>
                        <th className="text-right p-3 text-sm font-medium">سعر الوحدة</th>
                        <th className="text-right p-3 text-sm font-medium">الإجمالي</th>
                        <th className="text-right p-3 text-sm font-medium">إجراء</th>
                      </tr>
                    </thead>
                    <tbody>
                      {items.map((item) => (
                        <tr key={item.key} className="border-t border-border">
                          <td className="p-3">
                            <div>
                              <div className="font-medium">{item.product_name}</div>
                              <div className="text-xs text-text-muted">{item.product_id}</div>
                            </div>
                          </td>
                          <td className="p-3">
                            <Select
                              value={item.condition}
                              onChange={(e) => handleUpdateCondition(item.key, e.target.value as 'new' | 'used' | 'refurbished')}
                              options={[
                                { value: 'new', label: 'جديد' },
                                { value: 'used', label: 'مستعمل' },
                                { value: 'refurbished', label: 'مجدد' },
                              ]}
                              className="w-28"
                            />
                          </td>
                          <td className="p-3">
                            <Input
                              type="number"
                              value={item.quantity}
                              onChange={(e) => handleUpdateQuantity(item.key, parseInt(e.target.value) || 0)}
                              className="w-20"
                              min="1"
                            />
                          </td>
                          <td className="p-3">
                            <Input
                              type="number"
                              value={item.unit_cost}
                              onChange={(e) => handleUpdateUnitCost(item.key, parseFloat(e.target.value) || 0)}
                              className="w-28"
                              min="0"
                              step="0.01"
                            />
                          </td>
                          <td className="p-3 font-medium">
                            ₪{(item.quantity * item.unit_cost).toFixed(2)}
                          </td>
                          <td className="p-3">
                            <div className="flex gap-2">
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => {
                                  // Edit functionality - could open a modal to edit item details
                                  toast.info('تعديل العنصر - يمكن إضافة modal لتعديل التفاصيل');
                                }}
                                className="text-blue-600 hover:text-blue-700"
                              >
                                <Edit className="w-4 h-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleRemoveItem(item.key)}
                                className="text-red-600 hover:text-red-700"
                              >
                                <Trash2 className="w-4 h-4" />
                              </Button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Notes */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileText className="w-5 h-5 text-cyan" />
                ملاحظات
              </CardTitle>
            </CardHeader>
            <CardContent>
              <textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={4}
                placeholder="ملاحظات على طلب الشراء..."
                className="w-full px-4 py-3 border border-border rounded-xl focus:ring-2 focus:ring-cyan/30 focus:border-cyan/50 transition-all"
              />
            </CardContent>
          </Card>
        </div>

        {/* Right Column - Summary and Actions */}
        <div className="space-y-6">
          {/* Quick Stats */}
          <Card className="border-cyan/30 bg-cyan/5">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Package className="w-5 h-5 text-cyan" />
                ملخص الطلب
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-text-muted">عدد العناصر</span>
                <span className="font-semibold">{items.length}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-text-muted">إجمالي القطع</span>
                <span className="font-semibold">{totalQuantity}</span>
              </div>
              <div className="border-t border-border pt-4">
                <div className="flex justify-between items-center">
                  <span className="text-text-muted">إجمالي التكلفة</span>
                  <span className="text-2xl font-bold text-cyan">₪{totalCost.toFixed(2)}</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Receive Immediately */}
          <Card className="border-green-30 bg-green-5">
            <CardContent className="p-4">
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={receiveImmediately}
                  onChange={(e) => setReceiveImmediately(e.target.checked)}
                  className="w-5 h-5 rounded border-gray-300 text-green focus:ring-green"
                />
                <div>
                  <div className="font-medium flex items-center gap-2">
                    <CheckCircle2 className="w-4 h-4 text-green" />
                    استلام مباشر
                  </div>
                  <div className="text-sm text-text-muted">
                    عند التفعيل، سيتم استلام البضاعة وإنشاء عناصر المخزون تلقائياً
                  </div>
                </div>
              </label>
            </CardContent>
          </Card>

          {/* Actions */}
          <Card>
            <CardContent className="p-4 space-y-3">
              <Button
                variant="primary"
                onClick={handleUpdatePurchase}
                disabled={isSubmitting}
                className="w-full gap-2"
                size="lg"
              >
                {isSubmitting ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white" />
                    جاري التحديث...
                  </>
                ) : (
                  <>
                    <Save className="w-4 h-4" />
                    {receiveImmediately ? 'تحديث واستلام' : 'تحديث الشراء'}
                  </>
                )}
              </Button>
              <Button
                variant="secondary"
                onClick={() => navigate('/app/purchases')}
                className="w-full"
              >
                إلغاء
              </Button>
            </CardContent>
          </Card>

          {/* Supplier Info */}
          {selectedSupplier && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-sm">
                  <Truck className="w-4 h-4 text-cyan" />
                  معلومات المورد
                </CardTitle>
              </CardHeader>
              <CardContent>
                {(() => {
                  const supplier = suppliers.find(s => s.id === selectedSupplier);
                  if (!supplier) return null;
                  return (
                    <div className="space-y-2 text-sm">
                      <div><span className="text-text-muted">الاسم:</span> {supplier.name}</div>
                      {supplier.phone && <div><span className="text-text-muted">الهاتف:</span> {supplier.phone}</div>}
                      {supplier.email && <div><span className="text-text-muted">البريد:</span> {supplier.email}</div>}
                      {supplier.address && <div><span className="text-text-muted">العنوان:</span> {supplier.address}</div>}
                    </div>
                  );
                })()}
              </CardContent>
            </Card>
          )}
        </div>
      </div>

      {/* Manual Product Creation Modal */}
      <Modal
        isOpen={isManualProductModalOpen}
        onClose={() => setIsManualProductModalOpen(false)}
        title="إضافة قطعة جديدة"
        variant="modern"
        size="lg"
      >
        <div className="space-y-4">
          {/* Basic Information */}
          <div style={{ 
            marginBottom: '20px',
            paddingBottom: '20px',
            borderBottom: '1px solid var(--border-subtle)'
          }}>
            <div style={{ 
              display: 'flex', 
              alignItems: 'center', 
              gap: '10px',
              marginBottom: '16px',
              padding: '10px 14px',
              background: 'var(--bg-surface-elevated)',
              borderRadius: '12px',
              border: '1px solid var(--border-subtle)'
            }}>
              <div style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                background: 'var(--primary)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                boxShadow: '0 4px 12px rgba(99, 102, 241, 0.3)'
              }}>
                <Sparkles className="w-4 h-4" style={{ color: 'var(--text-on-primary)' }} />
              </div>
              <h4 style={{ 
                fontSize: '13px', 
                fontWeight: '600', 
                color: 'var(--text-primary)',
                margin: 0,
                letterSpacing: '0.2px'
              }}>
                المعلومات الأساسية
              </h4>
            </div>
            
            <div style={{ 
              display: 'grid', 
              gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', 
              gap: '16px' 
            }}>
              <div>
                <label style={{ 
                  fontSize: '12px', 
                  fontWeight: '600', 
                  color: 'var(--text-secondary)',
                  marginBottom: '8px',
                  display: 'block',
                  letterSpacing: '0.2px'
                }}>
                  اسم المنتج
                  <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                </label>
                <Input 
                  value={manualProductData.name}
                  onChange={(e) => setManualProductData({ ...manualProductData, name: e.target.value })}
                  placeholder="أدخل اسم المنتج"
                />
              </div>
              <div>
                <label style={{ 
                  fontSize: '12px', 
                  fontWeight: '600', 
                  color: 'var(--text-secondary)',
                  marginBottom: '8px',
                  display: 'block',
                  letterSpacing: '0.2px'
                }}>
                  SKU
                </label>
                <Input 
                  value={manualProductData.sku}
                  onChange={(e) => setManualProductData({ ...manualProductData, sku: e.target.value })}
                  placeholder="SKU-..."
                />
              </div>
              <div>
                <label style={{ 
                  fontSize: '12px', 
                  fontWeight: '600', 
                  color: 'var(--text-secondary)',
                  marginBottom: '8px',
                  display: 'block',
                  letterSpacing: '0.2px'
                }}>
                  الباركود
                </label>
                <Input 
                  value={manualProductData.barcode}
                  onChange={(e) => setManualProductData({ ...manualProductData, barcode: e.target.value })}
                  placeholder="أدخل الباركود"
                />
              </div>
              <div>
                <label style={{ 
                  fontSize: '12px', 
                  fontWeight: '600', 
                  color: 'var(--text-secondary)',
                  marginBottom: '8px',
                  display: 'block',
                  letterSpacing: '0.2px'
                }}>
                  التصنيف
                </label>
                <Select
                  value={manualProductData.category_id}
                  onChange={(e) => setManualProductData({ ...manualProductData, category_id: e.target.value })}
                  options={[
                    { value: '', label: 'اختر التصنيف...' },
                    ...categories.map((c) => ({ value: c.id, label: c.name })),
                  ]}
                  emptyMessage="لا يوجد تصنيفات"
                />
              </div>
            </div>
          </div>

          {/* Pricing Information */}
          <div style={{ 
            marginBottom: '20px',
            paddingBottom: '20px',
            borderBottom: '1px solid var(--border-subtle)'
          }}>
            <div style={{ 
              display: 'flex', 
              alignItems: 'center', 
              gap: '10px',
              marginBottom: '16px',
              padding: '10px 14px',
              background: 'var(--bg-surface-elevated)',
              borderRadius: '12px',
              border: '1px solid var(--border-subtle)'
            }}>
              <div style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                background: 'var(--primary)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                boxShadow: '0 4px 12px rgba(99, 102, 241, 0.3)'
              }}>
                <DollarSign className="w-4 h-4" style={{ color: 'var(--text-on-primary)' }} />
              </div>
              <h4 style={{ 
                fontSize: '13px', 
                fontWeight: '600', 
                color: 'var(--text-primary)',
                margin: 0,
                letterSpacing: '0.2px'
              }}>
                معلومات التسعير
              </h4>
            </div>
            
            <div style={{ 
              display: 'grid', 
              gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', 
              gap: '16px' 
            }}>
              <div>
                <label style={{ 
                  fontSize: '12px', 
                  fontWeight: '600', 
                  color: 'var(--text-secondary)',
                  marginBottom: '8px',
                  display: 'block',
                  letterSpacing: '0.2px'
                }}>
                  سعر التكلفة
                  <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                </label>
                <Input
                  type="number"
                  value={manualProductData.cost_price}
                  onChange={(e) => setManualProductData({ ...manualProductData, cost_price: e.target.value })}
                  placeholder="0.00"
                />
              </div>
              <div>
                <label style={{ 
                  fontSize: '12px', 
                  fontWeight: '600', 
                  color: 'var(--text-secondary)',
                  marginBottom: '8px',
                  display: 'block',
                  letterSpacing: '0.2px'
                }}>
                  سعر البيع
                  <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                </label>
                <Input
                  type="number"
                  value={manualProductData.selling_price}
                  onChange={(e) => setManualProductData({ ...manualProductData, selling_price: e.target.value })}
                  placeholder="0.00"
                />
              </div>
              <div>
                <label style={{ 
                  fontSize: '12px', 
                  fontWeight: '600', 
                  color: 'var(--text-secondary)',
                  marginBottom: '8px',
                  display: 'block',
                  letterSpacing: '0.2px'
                }}>
                  الحد الأدنى للمخزون
                </label>
                <Input
                  type="number"
                  value={manualProductData.min_stock}
                  onChange={(e) => setManualProductData({ ...manualProductData, min_stock: e.target.value })}
                  placeholder="0"
                />
              </div>
            </div>
          </div>

          {/* Additional Information */}
          <div>
            <div style={{ 
              display: 'flex', 
              alignItems: 'center', 
              gap: '10px',
              marginBottom: '16px',
              padding: '10px 14px',
              background: 'var(--bg-surface-elevated)',
              borderRadius: '12px',
              border: '1px solid var(--border-subtle)'
            }}>
              <div style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                background: 'var(--primary)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                boxShadow: '0 4px 12px rgba(99, 102, 241, 0.3)'
              }}>
                <Box className="w-4 h-4" style={{ color: 'var(--text-on-primary)' }} />
              </div>
              <h4 style={{ 
                fontSize: '13px', 
                fontWeight: '600', 
                color: 'var(--text-primary)',
                margin: 0,
                letterSpacing: '0.2px'
              }}>
                معلومات إضافية
              </h4>
            </div>
            
            <div>
              <label style={{ 
                fontSize: '12px', 
                fontWeight: '600', 
                color: 'var(--text-secondary)',
                marginBottom: '8px',
                display: 'block',
                letterSpacing: '0.2px'
              }}>
                الوصف
              </label>
              <Input 
                value={manualProductData.description}
                onChange={(e) => setManualProductData({ ...manualProductData, description: e.target.value })}
                placeholder="أدخل وصف المنتج"
              />
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-3 justify-end pt-4">
            <Button
              variant="secondary"
              onClick={() => setIsManualProductModalOpen(false)}
              disabled={createProductMutation.isPending}
            >
              إلغاء
            </Button>
            <Button
              variant="primary"
              onClick={handleManualProductCreate}
              disabled={createProductMutation.isPending}
              className="gap-2"
            >
              {createProductMutation.isPending ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white" />
                  جاري الإضافة...
                </>
              ) : (
                <>
                  <Sparkles className="w-4 h-4" />
                  إضافة للشراء والمخزون
                </>
              )}
            </Button>
          </div>
        </div>
      </Modal>
      <ConfirmDialog
        isOpen={productToDelete !== null}
        onClose={() => setProductToDelete(null)}
        onConfirm={() => {
          if (productToDelete) {
            deleteProductMutation.mutate(productToDelete.id);
          }
          setProductToDelete(null);
        }}
        title="حذف المنتج"
        message="هل أنت متأكد من حذف هذا المنتج نهائياً؟ هذا الإجراء لا يمكن التراجع عنه."
        confirmText="حذف المنتج"
        isLoading={deleteProductMutation.isPending}
        variant="danger"
      />
    </div>
  );
}