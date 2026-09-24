import { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { CalendarDays, Plus, ScanLine, Trash2, Warehouse } from 'lucide-react';
import { toast } from 'sonner';
import { Modal } from '../../../design-system/components/modal';
import { Input } from '../../../design-system/components/input';
import { Select } from '../../../design-system/components/select';
import { Button } from '../../../design-system/components/button';
import { barcodeApi, customersApi, inventoryApi, partTypesApi, productsApi, suppliersApi } from '../../../services/api/endpoints';
import { getStoreToday } from '../../../utils/store-time';

interface OpeningStockModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated: () => void;
  stockType?: 'general' | 'used';
}

type OpeningMode = 'quantity' | 'individual' | 'batch';
type OpeningCondition = 'NEW' | 'USED' | 'REFURBISHED' | 'DAMAGED' | 'FOR_PARTS';

interface BatchRow {
  id: string;
  barcode: string;
}

const today = () => getStoreToday();

export function OpeningStockModal({ isOpen, onClose, onCreated, stockType = 'general' }: OpeningStockModalProps) {
  const isUsedStock = stockType === 'used';
  const [mode, setMode] = useState<OpeningMode>(isUsedStock ? 'individual' : 'quantity');
  const [usedProductMode, setUsedProductMode] = useState<'new' | 'existing'>('new');
  const [productId, setProductId] = useState('');
  const [productName, setProductName] = useState('');
  const [quantity, setQuantity] = useState('1');
  const [businessDate, setBusinessDate] = useState(today);
  const [barcode, setBarcode] = useState('');
  const [batchRows, setBatchRows] = useState<BatchRow[]>([]);
  const [batchBarcodeText, setBatchBarcodeText] = useState('');
  const [condition, setCondition] = useState<OpeningCondition>('NEW');
  const [partTypeId, setPartTypeId] = useState('');
  const [supplierId, setSupplierId] = useState('');
  const [customerId, setCustomerId] = useState('');
  const [purchaseCost, setPurchaseCost] = useState('0');
  const [sellingPrice, setSellingPrice] = useState('0');
  const [notes, setNotes] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isResolvingBarcode, setIsResolvingBarcode] = useState(false);
  const [currentStep, setCurrentStep] = useState<1 | 2 | 3 | 4>(1);

  const { data: productsData, isLoading: productsLoading } = useQuery({
    queryKey: ['products', 'opening-stock-picker'],
    queryFn: () => productsApi.list({ page: 1, per_page: 1000 }),
    enabled: isOpen,
  });
  const { data: suppliersData, isLoading: suppliersLoading } = useQuery({
    queryKey: ['suppliers', 'opening-stock-picker'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 100 }),
    enabled: isOpen,
  });
  const { data: customersData, isLoading: customersLoading } = useQuery({
    queryKey: ['customers', 'opening-stock-picker'],
    queryFn: () => customersApi.list({ page: 1, per_page: 100 }),
    enabled: isOpen && isUsedStock,
  });
  const { data: partTypesData, isLoading: partTypesLoading } = useQuery({
    queryKey: ['part-types', 'opening-stock-picker'],
    queryFn: () => partTypesApi.list(),
    enabled: isOpen && isUsedStock,
  });
  const { data: usedInventoryData } = useQuery({
    queryKey: ['inventory', 'opening-stock-exclude-used'],
    queryFn: () => inventoryApi.list({ page: 1, per_page: 1000, condition: 'USED' }),
    enabled: isOpen,
  });

  const allProducts = ((productsData?.data?.products ?? []) as any[]);
  const usedInventoryItems = Array.isArray(usedInventoryData?.data)
    ? usedInventoryData.data
    : Array.isArray(usedInventoryData?.data?.items)
      ? usedInventoryData.data.items
      : [];
  const usedProductIds = new Set(
    usedInventoryItems
      .map((item: any) => item.product_id ?? item.product?.id)
      .filter(Boolean)
      .map(String),
  );
  const products = allProducts.filter((product) => {
    const condition = String(product.condition ?? '').toUpperCase();
    const sku = String(product.sku ?? '').toUpperCase();
    if (isUsedStock) return condition === 'USED' || sku.startsWith('USED-') || usedProductIds.has(String(product.id));
    return condition !== 'USED' && !sku.startsWith('USED-') && !usedProductIds.has(String(product.id));
  });
  const suppliers = ((suppliersData?.data?.suppliers ?? suppliersData?.data ?? []) as any[]);
  const customers = ((customersData?.data?.customers ?? customersData?.data ?? []) as any[]);
  const partTypes = (Array.isArray(partTypesData?.data) ? partTypesData.data : []) as any[];
  const selectedProduct = products.find((product) => String(product.id) === productId);
  const productOptions = useMemo(() => products.map((product) => ({
    value: String(product.id),
    label: `${product.name}${product.sku ? ` - ${product.sku}` : ''}`,
  })), [products]);
  const supplierOptions = useMemo(() => suppliers.map((supplier) => ({
    value: String(supplier.id),
    label: supplier.name,
  })), [suppliers]);

  useEffect(() => {
    if (!isOpen) return;
    setMode(isUsedStock ? 'individual' : 'quantity');
    setUsedProductMode('new');
    setProductId('');
    setProductName('');
    setQuantity('1');
    setBusinessDate(today());
    setBarcode('');
    setBatchRows([{ id: `${Date.now()}`, barcode: '' }]);
    setBatchBarcodeText('');
    setCondition('NEW');
    setPartTypeId('');
    setSupplierId('');
    setCustomerId('');
    setPurchaseCost('0');
    setSellingPrice('0');
    setNotes('');
    setCurrentStep(1);
  }, [isOpen]);

  const advanceStep = () => {
    if (currentStep === 1) {
      if ((!isUsedStock && !productId) || (isUsedStock && usedProductMode === 'new' && !productName.trim()) || (isUsedStock && usedProductMode === 'existing' && !productId)) {
        toast.error(isUsedStock && usedProductMode === 'new' ? 'أدخل اسم القطعة واختر نوعها.' : 'اختر القطعة أولًا.');
        return;
      }
      setCurrentStep(2);
      return;
    }
    if (currentStep === 2) {
      const parsedQuantity = Math.floor(Number(quantity));
      if ((mode !== 'batch' && (!parsedQuantity || parsedQuantity < 1)) || (mode === 'batch' && batchRows.length < 1) || (mode === 'individual' && !barcode.trim()) || (isUsedStock && !partTypeId)) {
        toast.error('أكمل بيانات الكمية أو تعريف القطعة قبل المتابعة.');
        return;
      }
      setCurrentStep(3);
      return;
    }
    if (currentStep === 3) {
      setCurrentStep(4);
      return;
    }
    void handleSubmit(new Event('submit') as unknown as React.FormEvent);
  };

  useEffect(() => {
    if (!isOpen) return;
    requestAnimationFrame(() => {
      const firstField = document.querySelector<HTMLElement>('.opening-stock-stage input:not([disabled]), .opening-stock-stage select:not([disabled])');
      firstField?.focus();
    });
  }, [currentStep, isOpen]);

  const resolveBarcode = async () => {
    const value = barcode.trim();
    if (!value) return;
    setIsResolvingBarcode(true);
    try {
      const response = await barcodeApi.resolve(value);
      const resolution: any = response.data ?? response;
      if (resolution.inventory_item?.id) {
        toast.error('هذا الباركود مرتبط بقطعة موجودة بالفعل. استخدم باركودًا جديدًا.');
        return;
      }
      if (resolution.product?.id) {
        setProductId(String(resolution.product.id));
        toast.success(`تم اختيار المنتج: ${resolution.product.name || resolution.product.sku || value}`);
      } else {
        toast.info('الباركود جديد. اختر المنتج يدويًا ثم احفظ المخزون الحالي.');
      }
    } catch {
      toast.info('لم يتم العثور على الباركود. اختر المنتج يدويًا ثم احفظه ضمن المخزون الحالي.');
    } finally {
      setIsResolvingBarcode(false);
    }
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    const parsedQuantity = Math.floor(Number(quantity));
    const batchQuantity = batchRows.length;
    if ((isUsedStock && usedProductMode === 'new' && !productName.trim()) || (!isUsedStock && !productId) || (isUsedStock && usedProductMode === 'existing' && !productId) || (mode !== 'batch' && (!parsedQuantity || parsedQuantity < 1)) || (mode === 'batch' && batchQuantity < 1)) {
      toast.error(isUsedStock && usedProductMode === 'new' ? 'أدخل اسم القطعة واختر نوعها.' : 'اختر القطعة وأدخل كمية صحيحة.');
      return;
    }
    if (mode === 'individual' && !barcode.trim()) {
      toast.error('الباركود مطلوب للقطعة الفردية.');
      return;
    }
    if (isUsedStock && !partTypeId) {
      toast.error('نوع القطعة مطلوب للقطع المستعملة.');
      return;
    }
    if (mode === 'batch') {
      const barcodes = batchRows.map((row) => row.barcode.trim());
      if (barcodes.some((value) => !value)) {
        toast.error('أدخل باركودًا لكل قطعة في الدفعة.');
        return;
      }
      if (new Set(barcodes).size !== barcodes.length) {
        toast.error('لا يمكن تكرار الباركود داخل الدفعة.');
        return;
      }
    }

    setIsSubmitting(true);
    try {
      const barcodes = mode === 'batch'
        ? batchRows.map((row) => row.barcode.trim()).filter(Boolean)
        : barcode.trim() ? [barcode.trim()] : [];
      if (barcodes.length > 0) {
        const response = await inventoryApi.list({ page: 1, per_page: 1000 });
        const items = Array.isArray(response?.data)
          ? response.data
          : Array.isArray(response?.data?.items) ? response.data.items : [];
        const existingBarcodes = new Set(
          items
            .map((item: any) => String(item.barcode || '').trim().toLowerCase())
            .filter(Boolean),
        );
        const duplicateBarcode = barcodes.find((value) => existingBarcodes.has(value.toLowerCase()));
        if (duplicateBarcode) {
          throw new Error(`الباركود موجود مسبقًا: ${duplicateBarcode}`);
        }
      }

      let resolvedProductId = productId;
      let resolvedProductPrice = Number(selectedProduct?.selling_price || selectedProduct?.sellingPrice || 0);
      if (isUsedStock && usedProductMode === 'new') {
        const productResponse = await productsApi.create({
          name: productName.trim(),
          sku: `USED-${Date.now()}`,
          cost_price: Number(purchaseCost) || 0,
          selling_price: Number(sellingPrice) || 0,
          condition: 'used',
        });
        const createdProduct = (productResponse.data as any)?.product ?? productResponse.data;
        if (!createdProduct?.id) throw new Error('تعذر إنشاء القطعة الجديدة.');
        resolvedProductId = String(createdProduct.id);
        resolvedProductPrice = Number(createdProduct.selling_price ?? createdProduct.sellingPrice ?? 0);
      }

      const commonPayload = {
        product_id: resolvedProductId,
        business_date: businessDate,
        condition,
        part_type_id: isUsedStock ? partTypeId : undefined,
        purchase_cost: Number(purchaseCost) || 0,
        selling_price: Number(sellingPrice) || resolvedProductPrice,
        supplier_id: supplierId || undefined,
        customer_id: customerId || undefined,
        notes: notes.trim() || undefined,
      };
      if (mode === 'batch') {
        if (isUsedStock) {
          await inventoryApi.createBulkUsedStock({
            product_id: resolvedProductId,
            barcodes: batchRows.map((row) => row.barcode.trim()),
            business_date: businessDate,
            part_type_id: partTypeId,
            purchase_cost: Number(purchaseCost) || 0,
            selling_price: Number(sellingPrice) || resolvedProductPrice,
            notes: notes.trim() || undefined,
          });
        } else {
          for (const row of batchRows) {
            await inventoryApi.createOpeningStock({
              ...commonPayload,
              mode: 'individual',
              quantity: 1,
              barcode: row.barcode.trim(),
            });
          }
        }
      } else {
        await inventoryApi.createOpeningStock({
          ...commonPayload,
          mode,
          quantity: mode === 'individual' ? 1 : parsedQuantity,
          barcode: mode === 'individual' ? barcode.trim() : undefined,
        });
      }
      toast.success(mode === 'batch' ? `تمت إضافة ${batchQuantity} قطع إلى المخزون.` : 'تمت إضافة المخزون الحالي دون إنشاء فاتورة شراء.');
      onCreated();
      onClose();
    } catch (error: any) {
      const message = String(error?.message || '').trim();
      const arabicMessage = String(error?.arabicMessage || '').trim();
      const normalizedMessage = `${message} ${arabicMessage}`.toLowerCase();
      let reason = arabicMessage || message || 'تعذر إضافة المخزون الحالي.';
      if (normalizedMessage.includes('serial number already exists') || normalizedMessage.includes('الرقم التسلسلي')) {
        reason = 'الرقم التسلسلي موجود مسبقًا. استخدم رقمًا مختلفًا أو استخدم مسار المرتجع للقطعة نفسها.';
      } else if (normalizedMessage.includes('barcode already exists') || normalizedMessage.includes('الباركود')) {
        reason = 'الباركود موجود مسبقًا. استخدم باركودًا مختلفًا.';
      } else if (Number(error?.status) >= 500) {
        reason = `تعذر الحفظ بسبب خطأ في الخادم (HTTP ${error.status}). تحقق من بيانات القطعة وحاول مرة أخرى.`;
      }
      toast.error(reason);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={isUsedStock ? 'إضافة مخزون مستعمل موجود' : 'إضافة المخزون الحالي'} variant="modern" size="xl" className="opening-stock-modal">
      <form
        onSubmit={handleSubmit}
        className="space-y-5"
        onKeyDown={(event) => {
          if (event.key !== 'Enter' || event.shiftKey || event.isComposing) return;
          const target = event.target as HTMLElement;
          if (target.tagName === 'TEXTAREA') return;

          const isBarcodeField = target instanceof HTMLInputElement && target.dataset.openingBarcode === 'true';
          if (isBarcodeField && mode === 'batch') {
            event.preventDefault();
            event.stopPropagation();
            advanceStep();
            return;
          }

          if (target instanceof HTMLButtonElement || target.closest('button')) {
            return;
          }

          event.preventDefault();
          event.stopPropagation();
          advanceStep();
        }}
      >
        <div className="flex items-center justify-between border-b border-border pb-3">
          <div className="flex items-center gap-2" aria-label={`المرحلة ${currentStep} من 4`}>
            {[1, 2, 3, 4].map((step) => <span key={step} className={`h-1.5 rounded-full transition-all duration-150 ${step === currentStep ? 'w-6 bg-primary' : 'w-1.5 bg-border'}`} />)}
          </div>
          <div className="flex items-center gap-2">
            {currentStep > 1 && (
              <Button type="button" variant="ghost" size="sm" onClick={() => setCurrentStep((step) => Math.max(1, step - 1) as 1 | 2 | 3 | 4)}>
                رجوع
              </Button>
            )}
            {currentStep < 4 ? (
              <Button type="button" variant="primary" size="sm" onClick={advanceStep}>
                التالي
              </Button>
            ) : (
              <Button type="button" variant="primary" size="sm" onClick={(event) => void handleSubmit(event as unknown as React.FormEvent)}>
                حفظ
              </Button>
            )}
            <Button type="button" variant="ghost" size="sm" onClick={onClose}>إلغاء</Button>
          </div>
        </div>

        {currentStep === 1 && <div className="opening-stock-stage space-y-5 rounded-2xl border border-border bg-surface p-5">
        <div className="rounded-xl border border-cyan/20 bg-cyan/5 p-4 text-sm text-text-secondary">
          {isUsedStock
            ? 'سجّل القطع المستعملة الموجودة لديك الآن دون إنشاء عملية شراء. يمكنك ربط القطعة بالعميل الذي جاءت منه بشكل اختياري.'
            : 'استخدم هذه النافذة لتسجيل البضاعة الموجودة لديك الآن، مثل المخزون عند بدء استخدام النظام. لن يتم إنشاء فاتورة شراء أو مديونية للتاجر.'}
        </div>

        <div className="grid gap-3 md:grid-cols-3">
          <Select
            label="طريقة تسجيل المخزون"
            value={mode}
            onChange={(event) => setMode(event.target.value as OpeningMode)}
            options={isUsedStock ? [
              { value: 'individual', label: 'قطعة محددة برقم وباركود' },
              { value: 'batch', label: 'دفعة متعددة بباركودات مختلفة' },
            ] : [
              { value: 'quantity', label: 'كمية إجمالية من المنتج' },
              { value: 'individual', label: 'قطعة محددة برقم وباركود' },
              { value: 'batch', label: 'دفعة متعددة بباركودات مختلفة' },
            ]}
          />
          {isUsedStock ? (
            <Select
              label="طريقة إضافة القطعة"
              value={usedProductMode}
              onChange={(event) => setUsedProductMode(event.target.value as 'new' | 'existing')}
              options={[
                { value: 'new', label: 'إضافة قطعة جديدة' },
                { value: 'existing', label: 'إضافة قطعة موجودة' },
              ]}
            />
          ) : (
            <Select
              label="المنتج"
              value={productId}
              onChange={(event) => setProductId(event.target.value)}
              options={[{ value: '', label: 'اختر المنتج...' }, ...productOptions]}
              loading={productsLoading}
              emptyMessage="لا توجد منتجات"
            />
          )}
        </div>

        {isUsedStock && (
          <div className="rounded-xl border border-border-subtle bg-surface-elevated p-4">
            {usedProductMode === 'new' ? (
              <Input
                label="اسم القطعة أو الموديل"
                value={productName}
                onChange={(event) => setProductName(event.target.value)}
                placeholder="مثال: GTX 1660 مستعمل"
                required
              />
            ) : (
              <Select
                label="القطعة الموجودة"
                value={productId}
                onChange={(event) => setProductId(event.target.value)}
                options={[{ value: '', label: 'اختر القطعة الموجودة...' }, ...productOptions]}
                loading={productsLoading}
                emptyMessage="لا توجد قطع مستعملة مسجلة"
                required
              />
            )}
          </div>
        )}
        </div>}

        {currentStep === 2 && <div className="opening-stock-stage space-y-5 rounded-2xl border border-border bg-surface p-5">
        <h3 className="text-lg font-bold text-text-primary">الكمية وتعريف القطعة</h3>
        <div className="grid gap-3 md:grid-cols-4">
          {mode !== 'batch' && (
            <Input
              label="الكمية"
              type="number"
              min="1"
              value={mode === 'individual' ? '1' : quantity}
              onChange={(event) => setQuantity(event.target.value)}
              disabled={mode === 'individual'}
              required
            />
          )}
          <Input label="تاريخ إدخال المخزون" type="date" value={businessDate} onChange={(event) => setBusinessDate(event.target.value)} required />
          <Select
            label="الحالة"
            value={condition}
            onChange={(event) => setCondition(event.target.value as OpeningCondition)}
            options={[
              { value: 'NEW', label: 'جديد' },
              { value: 'USED', label: 'مستعمل' },
              { value: 'REFURBISHED', label: 'مجدد' },
              { value: 'DAMAGED', label: 'تالف' },
              { value: 'FOR_PARTS', label: 'للقطع' },
            ]}
          />
          {isUsedStock && (
            <Select
              label="نوع القطعة"
              value={partTypeId}
              onChange={(event) => setPartTypeId(event.target.value)}
              options={[
                { value: '', label: 'اختر نوع القطعة...' },
                ...partTypes.map((partType) => ({ value: String(partType.id), label: partType.name_ar })),
              ]}
              loading={partTypesLoading}
              emptyMessage="لا توجد أنواع قطع"
              required
            />
          )}
        </div>

        {mode === 'individual' && (
          <div className="rounded-xl border border-border-subtle bg-surface-elevated p-4 space-y-4">
            <div className="flex items-center gap-2 text-sm font-semibold text-text-primary"><ScanLine className="h-4 w-4" /> بيانات القطعة المحددة</div>
            <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto]">
              <Input label="الباركود" data-opening-barcode="true" value={barcode} onChange={(event) => setBarcode(event.target.value)} placeholder="امسح أو أدخل الباركود" required />
              <Button type="button" variant="secondary" className="self-end" onClick={() => void resolveBarcode()} disabled={isResolvingBarcode || !barcode.trim()}>
                {isResolvingBarcode ? 'جارٍ التحقق...' : 'تحقق من الباركود'}
              </Button>
            </div>
          </div>
        )}

        {mode === 'batch' && (
          <div className="space-y-3 rounded-xl border border-border-subtle bg-surface-elevated p-4">
            <div className="flex items-center justify-between gap-3">
              <div>
                <div className="text-sm font-semibold text-text-primary">بيانات القطع</div>
                <p className="mt-1 text-xs text-text-secondary">أدخل باركودًا مختلفًا لكل قطعة، والرقم التسلسلي اختياري.</p>
              </div>
              <Button type="button" variant="secondary" size="sm" onClick={() => setBatchRows((rows) => [...rows, { id: `${Date.now()}-${rows.length}`, barcode: '' }])}>
                <Plus className="h-4 w-4" /> إضافة قطعة
              </Button>
            </div>
            {isUsedStock && (
              <div className="rounded-xl border border-primary/20 bg-primary/5 p-3">
                <label htmlFor="used-bulk-barcodes" className="mb-1 block text-xs font-semibold text-text-primary">
                  لصق عدة باركودات دفعة واحدة
                </label>
                <textarea
                  id="used-bulk-barcodes"
                  value={batchBarcodeText}
                  onChange={(event) => {
                    const value = event.target.value;
                    setBatchBarcodeText(value);
                    const barcodes = value.split(/[\r\n,;]+/).map((entry) => entry.trim()).filter(Boolean);
                    if (barcodes.length > 0) {
                      setBatchRows(barcodes.map((value, index) => ({ id: `${Date.now()}-${index}`, barcode: value })));
                    }
                  }}
                  placeholder="امسح أو الصق باركودًا في كل سطر"
                  rows={3}
                  className="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm"
                />
                <p className="mt-1 text-[11px] text-text-secondary">بعد اللصق اضغط «التالي». سيتم فحص التكرار وحفظ الدفعة بمعاملة واحدة.</p>
              </div>
            )}
            {batchRows.map((row, index) => (
              <div key={row.id} className="grid gap-2 md:grid-cols-[auto_minmax(0,1fr)_auto] md:items-end">
                <span className="pb-3 text-xs font-semibold text-text-muted">#{index + 1}</span>
                <Input label="الباركود" data-opening-barcode="true" value={row.barcode} onChange={(event) => setBatchRows((rows) => rows.map((current) => current.id === row.id ? { ...current, barcode: event.target.value } : current))} placeholder="باركود القطعة" required />
                <Button type="button" variant="ghost" size="icon" aria-label="حذف القطعة" disabled={batchRows.length === 1} onClick={() => setBatchRows((rows) => rows.filter((current) => current.id !== row.id))}>
                  <Trash2 className="h-4 w-4 text-danger" />
                </Button>
              </div>
            ))}
          </div>
        )}
        </div>}

        {currentStep === 3 && <div className="opening-stock-stage space-y-5 rounded-2xl border border-border bg-surface p-5">
        <h3 className="text-lg font-bold text-text-primary">التفاصيل والربط</h3>
        <div className="grid gap-4 md:grid-cols-3">
          {isUsedStock ? (
            <Select
              label="ربط بعميل (اختياري)"
              value={customerId}
              onChange={(event) => setCustomerId(event.target.value)}
              options={[{ value: '', label: 'بدون عميل' }, ...customers.map((customer) => ({ value: String(customer.id), label: customer.name }))]}
              loading={customersLoading}
              emptyMessage="لا يوجد عملاء"
            />
          ) : (
            <Select
              label="ربط بتاجر (اختياري)"
              value={supplierId}
              onChange={(event) => setSupplierId(event.target.value)}
              options={[{ value: '', label: 'بدون تاجر' }, ...supplierOptions]}
              loading={suppliersLoading}
              emptyMessage="لا يوجد تجار"
            />
          )}
          <p className="-mt-2 text-xs text-text-secondary md:col-span-3">
            {isUsedStock ? 'اختر عميلًا إذا أردت تسجيل مصدر هذه القطعة.' : 'اختر تاجرًا إذا أردت تسجيل اسم التاجر المرتبط بهذه البضاعة.'}
          </p>
          <Input label="سعر التكلفة" type="number" min="0" step="0.01" value={purchaseCost} onChange={(event) => setPurchaseCost(event.target.value)} />
          <Input label="سعر البيع" type="number" min="0" step="0.01" value={sellingPrice} onChange={(event) => setSellingPrice(event.target.value)} />
        </div>

        <Input label="ملاحظات" value={notes} onChange={(event) => setNotes(event.target.value)} placeholder="مثال: موجود قبل تشغيل النظام" />
        </div>}

        {currentStep === 4 && <div className="opening-stock-stage space-y-5 rounded-2xl border border-border bg-surface p-5">
        <h3 className="text-lg font-bold text-text-primary">مراجعة المخزون الحالي</h3>
        <div className="grid gap-3 rounded-xl border border-border-subtle bg-surface-elevated p-4 text-sm text-text-secondary">
          <div className="flex justify-between"><span>الطريقة</span><strong className="text-text-primary">{mode === 'quantity' ? 'كمية إجمالية' : mode === 'individual' ? 'قطعة فردية' : 'دفعة'}</strong></div>
          <div className="flex justify-between"><span>الكمية</span><strong className="text-text-primary">{mode === 'batch' ? batchRows.length : mode === 'individual' ? 1 : quantity}</strong></div>
          <div className="flex justify-between"><span>سعر التكلفة</span><strong className="text-text-primary">₪{Number(purchaseCost || 0).toFixed(2)}</strong></div>
          <div className="flex justify-between"><span>سعر البيع</span><strong className="text-text-primary">₪{Number(sellingPrice || 0).toFixed(2)}</strong></div>
        </div>
        <div className="flex items-center justify-between rounded-xl border border-border-subtle bg-surface-elevated px-4 py-3 text-xs text-text-secondary">
          <span className="flex items-center gap-2"><Warehouse className="h-4 w-4" /> النوع: {isUsedStock ? 'مخزون مستعمل موجود' : 'مخزون موجود مسبقًا'}</span>
          <span className="flex items-center gap-2"><CalendarDays className="h-4 w-4" /> {businessDate || 'اختر التاريخ'}</span>
        </div>

        <div className="flex justify-end gap-3">
          <Button type="submit" disabled={isSubmitting}>{isSubmitting ? 'جارٍ الحفظ...' : isUsedStock ? 'حفظ المخزون المستعمل' : 'حفظ المخزون الحالي'}</Button>
        </div>
        </div>}
      </form>
    </Modal>
  );
}
