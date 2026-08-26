import { useState, useEffect, useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Modal } from '../../../components/ui/modal';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { Badge } from '../../../components/ui/badge';
import { Card, CardContent } from '../../../components/ui/card';
import { getButtonSize } from '../../../config/button-sizes';
import {
  Search,
  Plus,
  Scan,
  Trash2,
  ShoppingCart,
  Calendar,
  Truck,
  CheckCircle2,
  AlertCircle,
  Package,
  Barcode,
  FileText,
} from 'lucide-react';
import { purchasesApi, suppliersApi, productsApi } from '../../../services/api/endpoints';
import { usePurchases } from '../hooks/usePurchases';
import { PurchaseItem } from '../types/purchases.types';

interface PurchaseModalProps {
  isOpen: boolean;
  onClose: () => void;
}

interface LineItem extends PurchaseItem {
  key: string;
}

export function PurchaseModal({ isOpen, onClose }: PurchaseModalProps) {
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

  const { data: suppliersData, isLoading: suppliersLoading } = useQuery({
    queryKey: ['suppliers'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 10 }),
    enabled: isOpen,
  });

  const { data: productsData } = useQuery({
    queryKey: ['products', productSearchQuery],
    queryFn: () => productsApi.list({ search: productSearchQuery, page: 1, per_page: 10 }),
    enabled: isOpen && activeTab === 'manual',
  });

  const searchedProducts = (productsData?.data?.products as any[]) || [];
  const suppliers = (suppliersData?.data as any[]) || [];

  const { createPurchaseMutation, receivePurchaseMutation } = usePurchases();

  const totalCost = useMemo(
    () => items.reduce((sum, item) => sum + item.quantity * item.unit_cost, 0),
    [items]
  );

  const totalQuantity = useMemo(
    () => items.reduce((sum, item) => sum + item.quantity, 0),
    [items]
  );

  useEffect(() => {
    if (isOpen) {
      setInvoiceNumber(`PO-${Date.now()}`);
      setPurchaseDate(new Date().toISOString().split('T')[0]);
      setExpectedDate('');
      setNotes('');
      setItems([]);
      setSelectedSupplier('');
      setBarcodeInput('');
      setProductSearchQuery('');
      setActiveTab('manual');
      setReceiveImmediately(false);
    }
  }, [isOpen]);

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

  const handleCreatePurchase = useCallback(async () => {
    if (!selectedSupplier) {
      alert('يرجى اختيار المورد');
      return;
    }
    if (items.length === 0) {
      alert('يرجى إضافة عناصر للشراء');
      return;
    }

    const formData = {
      supplier_id: selectedSupplier,
      invoice_number: invoiceNumber || `PO-${Date.now()}`,
      purchase_date: new Date(purchaseDate).toISOString(),
      expected_date: expectedDate ? new Date(expectedDate).toISOString() : undefined,
      notes: notes || undefined,
      items: items.map((item) => ({
        product_id: item.product_id,
        quantity: item.quantity,
        unit_cost: item.unit_cost,
        condition: item.condition,
      })),
    };

    createPurchaseMutation.mutate(formData, {
      onSuccess: (response) => {
        const purchaseId = response?.data?.id;
        if (purchaseId && receiveImmediately) {
          receivePurchaseMutation.mutate(purchaseId, {
            onSuccess: () => {
              queryClient.invalidateQueries({ queryKey: ['purchases'] });
              queryClient.invalidateQueries({ queryKey: ['inventory'] });
              queryClient.invalidateQueries({ queryKey: ['products'] });
              onClose();
            },
          });
        } else {
          queryClient.invalidateQueries({ queryKey: ['purchases'] });
          onClose();
        }
      },
    });
  }, [
    selectedSupplier,
    items,
    invoiceNumber,
    purchaseDate,
    expectedDate,
    notes,
    receiveImmediately,
    createPurchaseMutation,
    receivePurchaseMutation,
    queryClient,
    onClose,
  ]);

  const isSubmitting =
    createPurchaseMutation.isPending || receivePurchaseMutation.isPending;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="إنشاء شراء جديد"
      variant="modern"
      size="2xl"
    >
      <div className="space-y-6">
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

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <ShoppingCart className="w-5 h-5 text-cyan" />
                <h3 className="font-semibold">إضافة القطع</h3>
              </div>
              <Badge variant="secondary">
                {items.length} عنصر • {totalQuantity} قطعة
              </Badge>
            </div>

            <div className="flex gap-2 mb-4">
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
                variant={activeTab !== 'scan' ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => setActiveTab('manual')}
                className="flex-1"
              >
                <Plus className="w-4 h-4 ml-2" />
                إضافة يدوية
              </Button>
            </div>

            {activeTab === 'scan' ? (
              <form onSubmit={handleBarcodeScan} className="flex gap-2">
                <Input
                  placeholder="امسح الباركود أو اكتب الرقم..."
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
                  <div className="border border-border rounded-lg max-h-48 overflow-y-auto">
                    {searchedProducts.map((product) => (
                      <div
                        key={product.id}
                        className="p-3 hover:bg-surface/30 cursor-pointer border-b border-border last:border-b-0 flex justify-between items-center"
                        onClick={() => handleManualAdd(product)}
                      >
                        <div>
                          <div className="font-medium">{product.name}</div>
                          <div className="text-sm text-text-muted">
                            {product.barcode ? `الباركود: ${product.barcode}` : product.sku ? `SKU: ${product.sku}` : product.id}
                          </div>
                        </div>
                        <div className="text-sm font-medium text-cyan">
                          ₪{product.cost_price?.toFixed(2) || '0.00'}
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

        {items.length > 0 && (
          <Card>
            <CardContent className="p-0">
              <div className="p-4 border-b border-border">
                <div className="flex items-center justify-between">
                  <h3 className="font-semibold">العناصر المضافة ({items.length})</h3>
                  <Badge variant="secondary">الإجمالي: ₪{totalCost.toFixed(2)}</Badge>
                </div>
              </div>
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
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRemoveItem(item.key)}
                            className="text-red-600 hover:text-red-700"
                          >
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        )}

        <div>
          <label className="block text-sm font-medium text-text mb-2">ملاحظات</label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={3}
            placeholder="ملاحظات على طلب الشراء..."
            className="w-full px-4 py-3 border border-border rounded-xl focus:ring-2 focus:ring-cyan/30 focus:border-cyan/50 transition-all"
          />
        </div>

        <Card className="border-cyan/30 bg-cyan/5">
          <CardContent className="p-4">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={receiveImmediately}
                onChange={(e) => setReceiveImmediately(e.target.checked)}
                className="w-5 h-5 rounded border-gray-300 text-cyan focus:ring-cyan"
              />
              <div>
                <div className="font-medium flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-cyan" />
                  استلام مباشر
                </div>
                <div className="text-sm text-text-muted">
                  عند التفعيل، سيتم استلام البضاعة وإنشاء عناصر المخزون تلقائياً بعد إنشاء الشراء
                </div>
              </div>
            </label>
          </CardContent>
        </Card>

        <div className="flex flex-col md:flex-row justify-between items-center gap-4 p-4 bg-surface/30 rounded-lg">
          <div className="text-center md:text-right">
            <div className="text-sm text-text-muted">إجمالي الشراء</div>
            <div className="text-2xl font-bold text-cyan">₪{totalCost.toFixed(2)}</div>
            <div className="text-xs text-text-muted mt-1">
              {items.length} عنصر • {totalQuantity} قطعة
            </div>
          </div>
          <div className="flex gap-3 w-full md:w-auto">
            <Button variant="secondary" onClick={onClose} className="flex-1 md:flex-none">
              إلغاء
            </Button>
            <Button
              variant="primary"
              onClick={handleCreatePurchase}
              disabled={isSubmitting}
              className="flex-1 md:flex-none gap-2"
            >
              {isSubmitting ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white" />
                  جاري التنفيذ...
                </>
              ) : (
                <>
                  <ShoppingCart className="w-4 h-4" />
                  {receiveImmediately ? 'إنشاء واستلام' : 'إنشاء الشراء'}
                </>
              )}
            </Button>
          </div>
        </div>
      </div>
    </Modal>
  );
}
