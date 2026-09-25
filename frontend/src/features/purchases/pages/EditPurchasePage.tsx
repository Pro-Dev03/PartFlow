import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { Select } from '../../../design-system/components/select';
import { Badge } from '../../../design-system/components/badge';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { PageHeader } from '../../../design-system/components/page-header';
import { Modal } from '../../../design-system/components/modal';
import { Product, Supplier, Category } from '../../../types/models';
import {
  Search,
  Scan,
  Trash2,
  ShoppingCart,
  CheckCircle2,
  Package,
  FileText,
  Save,
  Sparkles,
  DollarSign,
  Box,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { purchasesApi, suppliersApi, productsApi, categoriesApi } from '../../../services/api/endpoints';
import { usePurchases } from '../hooks/usePurchases';
import { PurchaseItem } from '../types/purchases.types';
import { toast } from 'sonner';
import { generateSku } from '../../../utils/sku';
import { getStoreDateKey, getStoreToday, storeDateToUTCISOString } from '../../../utils/store-time';
import './edit-purchase.css';

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
  const [productPage, setProductPage] = useState(1);
  const [barcodeInput, setBarcodeInput] = useState('');
  const [invoiceNumber, setInvoiceNumber] = useState('');
  const [purchaseDate, setPurchaseDate] = useState(getStoreToday());
  const [expectedDate, setExpectedDate] = useState('');
  const [notes, setNotes] = useState('');
  const [paymentAmount, setPaymentAmount] = useState('');
  const [receiveImmediately, setReceiveImmediately] = useState(false);
  const quantityInputRefs = useRef<Record<string, HTMLInputElement | null>>({});
  const unitCostInputRefs = useRef<Record<string, HTMLInputElement | null>>({});
  const sellingPriceInputRefs = useRef<Record<string, HTMLInputElement | null>>({});
  const barcodeInputRef = useRef<HTMLInputElement | null>(null);
  
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
  const { data: purchaseData, isLoading: purchaseLoading, isError: purchaseError, refetch: refetchPurchase } = useQuery({
    queryKey: ['purchase', id],
    queryFn: () => purchasesApi.get(id || ''),
    enabled: !!id,
  });

  const { data: suppliersData, isLoading: suppliersLoading } = useQuery({
    queryKey: ['suppliers'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 100 }),
  });

  const { data: productsData } = useQuery({
    queryKey: ['products', productSearchQuery, productPage],
    queryFn: () => productsApi.list({ search: productSearchQuery, page: productPage, per_page: 10 }),
    enabled: Boolean(productSearchQuery.trim()),
  });

  const productTotal = Number(productsData?.data?.total ?? 0);
  const productTotalPages = Math.max(1, Math.ceil(productTotal / 10));

  useEffect(() => {
    setProductPage(1);
  }, [productSearchQuery]);

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
      requestAnimationFrame(() => barcodeInputRef.current?.focus());
    },
  });

  const searchedProducts = (productsData?.data?.products as Product[]) || [];
  const suppliers = (suppliersData?.data as Supplier[]) || [];
  const categories = (categoriesData?.data as Category[]) || [];

  const { updatePurchaseMutation, receivePurchaseMutation } = usePurchases();
  const purchaseRecord = purchaseData?.data?.purchase || purchaseData?.purchase || purchaseData?.data;
  const storedPurchaseItems = purchaseData?.data?.items || purchaseData?.items || purchaseRecord?.items || [];
  const purchasePaid = Number(purchaseRecord?.paid_amount || 0);
  const purchaseTotal = Number(purchaseRecord?.total_amount || 0);
  const purchaseRemaining = Math.max(0, Number(purchaseRecord?.remaining_amount ?? purchaseRecord?.remaining ?? (purchaseTotal - purchasePaid)));

  const purchaseHasUnsavedChanges = useMemo(() => {
    if (!purchaseRecord) return false;
    const normalizeItems = (list: any[]) => list.map((item) => ({
      id: item.id || '',
      product_id: item.product_id || '',
      quantity: Number(item.quantity) || 0,
      unit_cost: Number(item.unit_cost) || 0,
      selling_price: Number(item.selling_price) || 0,
      category_id: item.category_id || '',
      condition: item.condition || 'new',
    }));
    const draft = {
      supplier_id: selectedSupplier,
      invoice_number: invoiceNumber,
      purchase_date: purchaseDate,
      expected_delivery_date: expectedDate,
      notes,
      items: normalizeItems(items),
    };
    const stored = {
      supplier_id: purchaseRecord.supplier_id || '',
      invoice_number: purchaseRecord.invoice_number || '',
      purchase_date: purchaseRecord.purchase_date ? getStoreDateKey(purchaseRecord.purchase_date) ?? getStoreToday() : getStoreToday(),
      expected_delivery_date: purchaseRecord.expected_delivery_date ? getStoreDateKey(purchaseRecord.expected_delivery_date) ?? '' : '',
      notes: purchaseRecord.notes || '',
      items: normalizeItems(storedPurchaseItems),
    };
    return JSON.stringify(draft) !== JSON.stringify(stored);
  }, [purchaseRecord, storedPurchaseItems, selectedSupplier, invoiceNumber, purchaseDate, expectedDate, notes, items]);

  const paymentMutation = useMutation({
    mutationFn: (amount: number) => purchasesApi.addPayment(id || '', { amount, paymentMethod: 'cash' }),
    onSuccess: () => {
      setPaymentAmount('');
      void queryClient.invalidateQueries({ queryKey: ['purchase', id] });
      void queryClient.invalidateQueries({ queryKey: ['purchases'] });
      void queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      toast.success('تم تسجيل الدفعة وتحديث المتبقي');
    },
    onError: (error: any) => {
      if (error?.message === 'payment amount exceeds remaining balance') {
        void queryClient.invalidateQueries({ queryKey: ['purchase', id] });
        toast.error('تغيّر الرصيد المحفوظ؛ تم تحديث الفاتورة. راجع المتبقي ثم أعد إدخال الدفعة.');
        return;
      }
      toast.error(error?.message || 'تعذر تسجيل الدفعة');
    },
  });

  // Load purchase data when available
  useEffect(() => {
    const purchase = purchaseData?.data?.purchase || purchaseData?.purchase || purchaseData?.data;
    const purchaseItems = purchaseData?.data?.items || purchaseData?.items || purchase?.items;
    if (purchase) {
      setSelectedSupplier(purchase.supplier_id || '');
      setInvoiceNumber(purchase.invoice_number || '');
      setPurchaseDate(purchase.purchase_date ? getStoreDateKey(purchase.purchase_date) ?? getStoreToday() : getStoreToday());
      setExpectedDate(purchase.expected_delivery_date ? getStoreDateKey(purchase.expected_delivery_date) ?? '' : '');
      setNotes(purchase.notes || '');
      
      // Load items
      if (Array.isArray(purchaseItems)) {
        setItems(purchaseItems.map((item: any, index: number) => ({
          key: `${item.product_id}-${index}`,
          id: item.id,
          purchase_id: item.purchase_id,
          product_id: item.product_id,
          product_name: item.product_name || item.product?.name || '',
          quantity: Number(item.quantity) || 0,
          unit_cost: Number(item.unit_cost) || 0,
          selling_price: Number(item.selling_price) || 0,
          category_id: item.category_id || '',
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
      const response = await productsApi.getByBarcode(barcodeInput);
      const product = response.data?.product || response.data;

      setItems((prev) => {
        const scannedBarcode = barcodeInput.trim();
        const existingIndex = prev.findIndex((item) => item.product_id === product.id && item.barcode === scannedBarcode);
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
            barcode: scannedBarcode,
            quantity: 1,
            unit_cost: product.cost_price || 0,
            condition: 'new',
          },
        ];
      });

      setBarcodeInput('');
    } catch (error: any) {
      if (error?.status === 404) {
        setManualProductData((current) => ({ ...current, sku: generateSku(), barcode: barcodeInput.trim() }));
        setIsManualProductModalOpen(true);
        setBarcodeInput('');
        return;
      }
      console.error('Error scanning barcode:', error);
      toast.error(error?.message || 'تعذر البحث عن الباركود. تحقق من الاتصال ثم أعد المحاولة.');
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
    requestAnimationFrame(() => barcodeInputRef.current?.focus());
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
      sku: manualProductData.sku || generateSku(),
      barcode: manualProductData.barcode || undefined,
      category_id: manualProductData.category_id || undefined,
      cost_price: costPrice,
      selling_price: sellingPrice,
      min_stock_level: parseInt(manualProductData.min_stock.toString()) || 0,
      description: manualProductData.description || undefined,
    };

    createProductMutation.mutate(productData);
  }, [manualProductData, createProductMutation]);

  const handleUpdatePurchase = useCallback(async () => {
    if (!selectedSupplier) {
      toast.error('يرجى اختيار التاجر');
      return;
    }
    if (items.length === 0) {
      toast.error('يرجى إضافة عناصر للشراء');
      return;
    }

    const formData = {
      supplier_id: selectedSupplier,
      invoice_number: invoiceNumber || `PO-${Date.now()}`,
      purchase_date: storeDateToUTCISOString(purchaseDate) ?? new Date().toISOString(),
      expected_delivery_date: expectedDate ? storeDateToUTCISOString(expectedDate) ?? undefined : undefined,
      notes: notes || undefined,
      items: items.map((item) => ({
        id: item.id,
        product_id: item.product_id,
        quantity: item.quantity,
        unit_cost: item.unit_cost,
        selling_price: item.selling_price,
        category_id: item.category_id || undefined,
        condition: item.condition,
      })),
    };

    updatePurchaseMutation.mutate({ id: id || '', data: formData }, {
      onSuccess: (response) => {
        const purchaseId =
          response?.purchase?.id ||
          response?.data?.purchase?.id ||
          response?.data?.id ||
          response?.id || id;
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
          void queryClient.invalidateQueries({ queryKey: ['purchase', id] });
          void queryClient.invalidateQueries({ queryKey: ['purchases'] });
          void queryClient.refetchQueries({ queryKey: ['purchase', id], exact: true });
          toast.success('تم حفظ التعديلات. يمكنك الآن تسجيل الدفعة على الرصيد المحدّث.');
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

  if (purchaseError || !purchaseRecord) {
    return (
      <div className="mx-auto flex min-h-64 max-w-xl items-center justify-center p-4">
        <Card className="w-full">
          <CardContent className="space-y-4 p-6 text-center">
            <p className="font-semibold">تعذر تحميل بيانات الشراء</p>
            <p className="text-sm text-text-muted">تحقق من الاتصال أو حالة الفاتورة، ثم أعد المحاولة.</p>
            <div className="flex justify-center gap-2">
              <Button variant="secondary" onClick={() => navigate('/app/purchases')}>العودة للمشتريات</Button>
              <Button variant="primary" onClick={() => void refetchPurchase()} disabled={!id}>إعادة المحاولة</Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="تعديل الشراء"
        description="عدّل الفاتورة وأدخل المنتجات بسرعة باستخدام الباركود أو البحث."
        actions={
          <Button variant="secondary" onClick={() => navigate('/app/purchases')}>
            عودة للمشتريات
          </Button>
        }
      />

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <FileText className="h-4 w-4 text-cyan" />
            بيانات الفاتورة
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-0">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-5">
            <div className="xl:col-span-2">
              <label className="mb-1 block text-xs font-medium text-text-muted">المورد *</label>
              <Select
                value={selectedSupplier}
                onChange={(event) => setSelectedSupplier(event.target.value)}
                loading={suppliersLoading}
                options={[
                  { value: '', label: 'اختر المورد...' },
                  ...suppliers.map((supplier) => ({ value: supplier.id, label: supplier.name })),
                ]}
                emptyMessage="لا يوجد موردون"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-text-muted">رقم الفاتورة</label>
              <Input value={invoiceNumber} onChange={(event) => setInvoiceNumber(event.target.value)} placeholder="PO-..." />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-text-muted">تاريخ الشراء</label>
              <Input type="date" value={purchaseDate} onChange={(event) => setPurchaseDate(event.target.value)} />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-text-muted">الاستلام المتوقع</label>
              <Input type="date" value={expectedDate} onChange={(event) => setExpectedDate(event.target.value)} />
            </div>
          </div>
          {selectedSupplier && (() => {
            const supplier = suppliers.find((entry) => entry.id === selectedSupplier);
            return supplier?.phone ? <p className="mt-2 text-xs text-text-muted">هاتف المورد: {supplier.phone}</p> : null;
          })()}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <CardTitle className="flex items-center gap-2 text-base">
              <ShoppingCart className="h-4 w-4 text-cyan" />
              إدخال المنتجات
            </CardTitle>
            <Badge variant="secondary">{items.length} صنف · {totalQuantity} قطعة</Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-3 pt-0">
          <div className="grid grid-cols-1 items-end gap-2 lg:grid-cols-[minmax(240px,1fr)_minmax(240px,1fr)_auto]">
            <form onSubmit={handleBarcodeScan} className="flex min-w-0 gap-2">
              <div className="relative min-w-0 flex-1">
                <Scan className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-cyan" />
                <Input
                  ref={barcodeInputRef}
                  placeholder="امسح الباركود ثم اضغط Enter"
                  value={barcodeInput}
                  onChange={(event) => setBarcodeInput(event.target.value)}
                  className="pr-10"
                  autoFocus
                  aria-label="مسح باركود المنتج"
                />
              </div>
              <Button type="submit" variant="primary" disabled={!barcodeInput.trim()}>
                إضافة
              </Button>
            </form>

            <div className="relative min-w-0">
              <Search className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
              <Input
                placeholder="ابحث بالاسم أو الباركود..."
                value={productSearchQuery}
                onChange={(event) => setProductSearchQuery(event.target.value)}
                className="pr-10"
                aria-label="البحث عن منتج"
              />
            </div>

            <Button
              type="button"
              variant="secondary"
              onClick={() => {
                setManualProductData((current) => ({ ...current, sku: generateSku() }));
                setIsManualProductModalOpen(true);
              }}
              className="gap-2"
            >
              <Sparkles className="h-4 w-4" />
              منتج جديد
            </Button>
          </div>

          {productSearchQuery.trim() && (
            <div className="max-h-48 overflow-y-auto rounded-xl border border-border" role="region" aria-label="نتائج المنتجات">
              {searchedProducts.length ? searchedProducts.map((product) => (
                <button
                  key={product.id}
                  type="button"
                  onClick={() => handleManualAdd(product)}
                  className="flex w-full items-center justify-between gap-3 border-b border-border px-3 py-2 text-right last:border-b-0 hover:bg-surface/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan"
                >
                  <span className="min-w-0">
                    <span className="block truncate text-sm font-medium">{product.name}</span>
                    <span className="block truncate text-xs text-text-muted">
                      {product.barcode ? 'باركود: ' + product.barcode : 'بدون باركود'}
                      {product.sku ? ' · SKU: ' + product.sku : ''}
                    </span>
                  </span>
                  <span className="shrink-0 text-sm font-semibold text-cyan">
                    ₪{Number(product.cost_price || 0).toFixed(2)} <span aria-hidden="true">+</span>
                  </span>
                </button>
              )) : (
                <p className="px-3 py-4 text-center text-sm text-text-muted">لا توجد نتائج مطابقة.</p>
              )}
              {productTotalPages > 1 && (
                <div className="flex items-center justify-between border-t border-border px-2 py-1">
                  <Button type="button" variant="ghost" size="sm" disabled={productPage <= 1} onClick={() => setProductPage((page) => page - 1)}>
                    السابق <ChevronRight className="h-4 w-4" />
                  </Button>
                  <span className="text-xs text-text-muted">صفحة {productPage} من {productTotalPages}</span>
                  <Button type="button" variant="ghost" size="sm" disabled={productPage >= productTotalPages} onClick={() => setProductPage((page) => page + 1)}>
                    <ChevronLeft className="h-4 w-4" /> التالي
                  </Button>
                </div>
              )}
            </div>
          )}

          <div className="overflow-x-auto rounded-xl border border-border">
            <table className="w-full min-w-[920px] text-sm">
              <thead className="bg-surface/40 text-text-muted">
                <tr>
                  <th className="px-3 py-2 text-right font-medium">المنتج</th>
                  <th className="w-24 px-2 py-2 text-right font-medium">الكمية</th>
                  <th className="w-32 px-2 py-2 text-right font-medium">تكلفة الوحدة</th>
                  <th className="w-32 px-2 py-2 text-right font-medium">سعر البيع</th>
                  <th className="w-32 px-2 py-2 text-right font-medium">الحالة</th>
                  <th className="w-40 px-2 py-2 text-right font-medium">التصنيف</th>
                  <th className="w-28 px-2 py-2 text-right font-medium">الإجمالي</th>
                  <th className="w-14 px-2 py-2 text-center font-medium">حذف</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item, index) => (
                  <tr key={item.key} className="border-t border-border">
                    <td className="max-w-[260px] px-3 py-2">
                      <span className="block truncate font-medium" title={item.product_name}>{item.product_name}</span>
                      <span className="block truncate text-xs text-text-muted" title={item.product_id}>{item.product_id}</span>
                    </td>
                    <td className="px-2 py-2">
                      <Input
                        ref={(element) => { quantityInputRefs.current[item.key] = element; }}
                        type="number"
                        min="1"
                        step="1"
                        size="sm"
                        value={item.quantity}
                        aria-label={'كمية ' + item.product_name}
                        onChange={(event) => {
                          const value = Number.parseInt(event.target.value, 10);
                          if (Number.isInteger(value) && value > 0) handleUpdateQuantity(item.key, value);
                        }}
                        onKeyDown={(event) => {
                          if (event.key === 'Enter') {
                            event.preventDefault();
                            const next = unitCostInputRefs.current[item.key];
                            next?.focus();
                            next?.select();
                          }
                        }}
                      />
                    </td>
                    <td className="px-2 py-2">
                      <Input
                        ref={(element) => { unitCostInputRefs.current[item.key] = element; }}
                        type="number"
                        min="0"
                        step="0.01"
                        size="sm"
                        value={item.unit_cost}
                        aria-label={'تكلفة ' + item.product_name}
                        onChange={(event) => handleUpdateUnitCost(item.key, Number.parseFloat(event.target.value) || 0)}
                        onKeyDown={(event) => {
                          if (event.key === 'Enter') {
                            event.preventDefault();
                            const next = sellingPriceInputRefs.current[item.key];
                            next?.focus();
                            next?.select();
                          }
                        }}
                      />
                    </td>
                    <td className="px-2 py-2">
                      <Input
                        ref={(element) => { sellingPriceInputRefs.current[item.key] = element; }}
                        type="number"
                        min="0"
                        step="0.01"
                        size="sm"
                        value={item.selling_price ?? 0}
                        aria-label={'سعر بيع ' + item.product_name}
                        onChange={(event) => setItems((previous) => previous.map((current) => current.key === item.key ? { ...current, selling_price: Number.parseFloat(event.target.value) || 0 } : current))}
                        onKeyDown={(event) => {
                          if (event.key === 'Enter') {
                            event.preventDefault();
                            const nextItem = items[index + 1];
                            if (nextItem) {
                              const next = quantityInputRefs.current[nextItem.key];
                              next?.focus();
                              next?.select();
                            } else {
                              barcodeInputRef.current?.focus();
                            }
                          }
                        }}
                      />
                    </td>
                    <td className="px-2 py-2">
                      <Select
                        value={item.condition}
                        onChange={(event) => handleUpdateCondition(item.key, event.target.value as 'new' | 'used' | 'refurbished')}
                        options={[
                          { value: 'new', label: 'جديد' },
                          { value: 'used', label: 'مستعمل' },
                          { value: 'refurbished', label: 'مجدد' },
                        ]}
                      />
                    </td>
                    <td className="px-2 py-2">
                      <Select
                        value={item.category_id ?? ''}
                        onChange={(event) => setItems((previous) => previous.map((current) => current.key === item.key ? { ...current, category_id: event.target.value } : current))}
                        options={[
                          { value: '', label: 'بدون تصنيف' },
                          ...categories.map((category) => ({ value: category.id, label: category.name })),
                        ]}
                      />
                    </td>
                    <td className="whitespace-nowrap px-2 py-2 font-semibold">₪{(item.quantity * item.unit_cost).toFixed(2)}</td>
                    <td className="px-2 py-2 text-center">
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRemoveItem(item.key)}
                        aria-label={'حذف ' + item.product_name + ' من الشراء'}
                        title="حذف الصنف من الفاتورة"
                        className="text-red-600 hover:text-red-700"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </td>
                  </tr>
                ))}
                {items.length === 0 && (
                  <tr>
                    <td colSpan={8} className="px-3 py-10 text-center text-sm text-text-muted">
                      امسح باركود منتج أو ابحث عنه لإضافته إلى الفاتورة.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1.6fr)_minmax(320px,1fr)]">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="flex items-center gap-2 text-base">
              <FileText className="h-4 w-4 text-cyan" />
              ملاحظات
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            <textarea
              value={notes}
              onChange={(event) => setNotes(event.target.value)}
              rows={2}
              placeholder="ملاحظات على طلب الشراء..."
              className="w-full rounded-xl border border-border bg-[var(--input-bg)] px-3 py-2 text-sm text-text focus:border-cyan/50 focus:outline-none focus:ring-2 focus:ring-cyan/20"
            />
          </CardContent>
        </Card>

        <Card className="border-cyan/30 bg-cyan/5">
          <CardHeader className="pb-2">
            <CardTitle className="flex items-center gap-2 text-base">
              <Package className="h-4 w-4 text-cyan" />
              إجمالي الفاتورة
            </CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-3 gap-2 pt-0 text-center">
            <div>
              <p className="text-xs text-text-muted">الأصناف</p>
              <p className="font-semibold">{items.length}</p>
            </div>
            <div>
              <p className="text-xs text-text-muted">القطع</p>
              <p className="font-semibold">{totalQuantity}</p>
            </div>
            <div>
              <p className="text-xs text-text-muted">قيمة العناصر</p>
              <p className="font-bold text-cyan">₪{totalCost.toFixed(2)}</p>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="flex items-center gap-2 text-base">
            <DollarSign className="h-4 w-4 text-cyan" />
            الدفعات المسجلة
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 pt-0">
          <div className="grid grid-cols-2 gap-3 md:grid-cols-3">
            <div className="rounded-lg bg-surface/40 px-3 py-2">
              <span className="block text-xs text-text-muted">إجمالي الفاتورة المحفوظ</span>
              <span className="font-semibold">₪{purchaseTotal.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
            </div>
            <div className="rounded-lg bg-surface/40 px-3 py-2">
              <span className="block text-xs text-text-muted">المدفوع</span>
              <span className="font-semibold">₪{purchasePaid.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
            </div>
            <div className="rounded-lg bg-surface/40 px-3 py-2">
              <span className="block text-xs text-text-muted">المتبقي</span>
              <span className="font-semibold">₪{purchaseRemaining.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
            </div>
          </div>
          <div className="flex flex-col gap-2 sm:flex-row">
            <Input
              type="number"
              min="0.01"
              max={purchaseRemaining}
              step="0.01"
              size="sm"
              value={paymentAmount}
              onChange={(event) => setPaymentAmount(event.target.value)}
              placeholder="مبلغ الدفعة"
              aria-label="مبلغ الدفعة"
              disabled={purchaseRemaining <= 0 || purchaseHasUnsavedChanges || paymentMutation.isPending}
            />
            <Button
              type="button"
              variant="secondary"
              data-next-action
              className="shrink-0"
              onClick={() => paymentMutation.mutate(Number(paymentAmount))}
              disabled={!Number(paymentAmount) || Number(paymentAmount) <= 0 || Number(paymentAmount) > purchaseRemaining || purchaseHasUnsavedChanges || paymentMutation.isPending}
            >
              {paymentMutation.isPending ? 'جاري تسجيل الدفعة...' : 'تسجيل الدفعة'}
            </Button>
          </div>
          {purchaseHasUnsavedChanges && (
            <p className="text-xs text-amber-700 dark:text-amber-300" role="status">
              احفظ تعديلات الفاتورة أولًا ليُحدّث الرصيد قبل تسجيل الدفعة.
            </p>
          )}
        </CardContent>
      </Card>

      <div className="sticky bottom-2 z-20 flex flex-col gap-3 rounded-2xl border border-border bg-[var(--card-bg)] p-3 shadow-lg sm:flex-row sm:items-center sm:justify-between">
        <label className="flex items-start gap-2 text-sm">
          <input
            type="checkbox"
            checked={receiveImmediately}
            onChange={(event) => setReceiveImmediately(event.target.checked)}
            className="mt-0.5 h-4 w-4 rounded border-gray-300 accent-emerald-600"
          />
          <span>
            <span className="flex items-center gap-1 font-medium">
              <CheckCircle2 className="h-4 w-4 text-emerald-600" />
              استلام مباشر بعد الحفظ
            </span>
            <span className="block text-xs text-text-muted">ينشئ عناصر المخزون بعد حفظ الفاتورة.</span>
          </span>
        </label>
        <div className="flex gap-2">
          <Button variant="secondary" onClick={() => navigate('/app/purchases')}>إلغاء</Button>
          <Button
            variant="primary"
            data-next-action
            onClick={handleUpdatePurchase}
            disabled={isSubmitting}
            className="min-w-36 gap-2"
          >
            {isSubmitting ? (
              <>
                <div className="h-4 w-4 animate-spin rounded-full border-b-2 border-white" />
                جاري التحديث...
              </>
            ) : (
              <>
                <Save className="h-4 w-4" />
                {receiveImmediately ? 'تحديث واستلام' : 'حفظ التعديلات'}
              </>
            )}
          </Button>
        </div>
      </div>
      {/* Manual Product Creation Modal */}
      <Modal
        isOpen={isManualProductModalOpen}
        onClose={() => setIsManualProductModalOpen(false)}
        title="إضافة قطعة جديدة"
        variant="modern"
        size="lg"
        className="purchase-manual-product-modal"
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
                  readOnly
                  helperText="يتم توليده تلقائيًا"
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
                  min="0"
                  step="1"
                  value={manualProductData.min_stock}
                  onChange={(e) => setManualProductData({ ...manualProductData, min_stock: e.target.value })}
                  placeholder="0"
                  aria-label="الحد الأدنى للمخزون"
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
          <div className="purchase-manual-product-actions flex gap-3 justify-end pt-4">
            <Button
              variant="secondary"
              onClick={() => setIsManualProductModalOpen(false)}
              disabled={createProductMutation.isPending}
            >
              إلغاء
            </Button>
            <Button
              variant="primary"
              data-next-action
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
    </div>
  );
}
