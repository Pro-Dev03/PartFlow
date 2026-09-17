import { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { CalendarDays, Plus, ScanLine, Trash2, Warehouse } from 'lucide-react';
import { toast } from 'sonner';
import { Modal } from '../../../components/ui/modal';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { Button } from '../../../components/ui/button';
import { barcodeApi, customersApi, inventoryApi, partTypesApi, productsApi, suppliersApi } from '../../../services/api/endpoints';

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
  serialNumber: string;
}

const today = () => new Date().toISOString().slice(0, 10);

export function OpeningStockModal({ isOpen, onClose, onCreated, stockType = 'general' }: OpeningStockModalProps) {
  const isUsedStock = stockType === 'used';
  const [mode, setMode] = useState<OpeningMode>(isUsedStock ? 'individual' : 'quantity');
  const [usedProductMode, setUsedProductMode] = useState<'new' | 'existing'>('new');
  const [productId, setProductId] = useState('');
  const [productName, setProductName] = useState('');
  const [quantity, setQuantity] = useState('1');
  const [businessDate, setBusinessDate] = useState(today);
  const [barcode, setBarcode] = useState('');
  const [serialNumber, setSerialNumber] = useState('');
  const [batchRows, setBatchRows] = useState<BatchRow[]>([]);
  const [condition, setCondition] = useState<OpeningCondition>('NEW');
  const [partTypeId, setPartTypeId] = useState('');
  const [supplierId, setSupplierId] = useState('');
  const [customerId, setCustomerId] = useState('');
  const [purchaseCost, setPurchaseCost] = useState('0');
  const [sellingPrice, setSellingPrice] = useState('0');
  const [notes, setNotes] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isResolvingBarcode, setIsResolvingBarcode] = useState(false);

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
    setSerialNumber('');
    setBatchRows([{ id: `${Date.now()}`, barcode: '', serialNumber: '' }]);
    setCondition('NEW');
    setPartTypeId('');
    setSupplierId('');
    setCustomerId('');
    setPurchaseCost('0');
    setSellingPrice('0');
    setNotes('');
  }, [isOpen]);

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
      const serials = batchRows.map((row) => row.serialNumber.trim()).filter(Boolean);
      if (barcodes.some((value) => !value)) {
        toast.error('أدخل باركودًا لكل قطعة في الدفعة.');
        return;
      }
      if (new Set(barcodes).size !== barcodes.length || new Set(serials).size !== serials.length) {
        toast.error('لا يمكن تكرار الباركود أو الرقم التسلسلي داخل الدفعة.');
        return;
      }
    }

    setIsSubmitting(true);
    try {
      const serials = mode === 'batch'
        ? batchRows.map((row) => row.serialNumber.trim()).filter(Boolean)
        : serialNumber.trim() ? [serialNumber.trim()] : [];
      const barcodes = mode === 'batch'
        ? batchRows.map((row) => row.barcode.trim()).filter(Boolean)
        : barcode.trim() ? [barcode.trim()] : [];
      if (serials.length > 0 || barcodes.length > 0) {
        const response = await inventoryApi.list({ page: 1, per_page: 1000 });
        const items = Array.isArray(response?.data)
          ? response.data
          : Array.isArray(response?.data?.items) ? response.data.items : [];
        const existingSerials = new Set(
          items
            .map((item: any) => String(item.serial_number || '').trim().toLowerCase())
            .filter(Boolean),
        );
        const existingBarcodes = new Set(
          items
            .map((item: any) => String(item.barcode || '').trim().toLowerCase())
            .filter(Boolean),
        );
        const duplicateSerial = serials.find((value) => existingSerials.has(value.toLowerCase()));
        if (duplicateSerial) {
          throw new Error(`الرقم التسلسلي موجود مسبقًا: ${duplicateSerial}`);
        }
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
        for (const row of batchRows) {
          await inventoryApi.createOpeningStock({
            ...commonPayload,
            mode: 'individual',
            quantity: 1,
            barcode: row.barcode.trim(),
            serial_number: row.serialNumber.trim() || undefined,
          });
        }
      } else {
        await inventoryApi.createOpeningStock({
          ...commonPayload,
          mode,
          quantity: mode === 'individual' ? 1 : parsedQuantity,
          barcode: mode === 'individual' ? barcode.trim() : undefined,
          serial_number: mode === 'individual' ? serialNumber.trim() || undefined : undefined,
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
    <Modal isOpen={isOpen} onClose={onClose} title={isUsedStock ? 'إضافة مخزون مستعمل موجود' : 'إضافة المخزون الحالي'} variant="modern" size="lg">
      <form onSubmit={handleSubmit} className="space-y-5">
        <div className="rounded-xl border border-cyan/20 bg-cyan/5 p-4 text-sm text-text-secondary">
          {isUsedStock
            ? 'سجّل القطع المستعملة الموجودة لديك الآن دون إنشاء عملية شراء. يمكنك ربط القطعة بالعميل الذي جاءت منه بشكل اختياري.'
            : 'استخدم هذه النافذة لتسجيل البضاعة الموجودة لديك الآن، مثل المخزون عند بدء استخدام النظام. لن يتم إنشاء فاتورة شراء أو مديونية للمورد.'}
        </div>

        <div className="grid gap-4 md:grid-cols-2">
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

        <div className="grid gap-4 md:grid-cols-3">
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
            <div className="grid gap-4 md:grid-cols-[1fr_auto]">
              <Input label="الباركود" value={barcode} onChange={(event) => setBarcode(event.target.value)} placeholder="امسح أو أدخل الباركود" required />
              <Button type="button" variant="secondary" className="self-end" onClick={() => void resolveBarcode()} disabled={isResolvingBarcode || !barcode.trim()}>
                {isResolvingBarcode ? 'جارٍ التحقق...' : 'تحقق من الباركود'}
              </Button>
            </div>
            <Input label="الرقم التسلسلي" value={serialNumber} onChange={(event) => setSerialNumber(event.target.value)} placeholder="اختياري" />
          </div>
        )}

        {mode === 'batch' && (
          <div className="space-y-3 rounded-xl border border-border-subtle bg-surface-elevated p-4">
            <div className="flex items-center justify-between gap-3">
              <div>
                <div className="text-sm font-semibold text-text-primary">بيانات القطع</div>
                <p className="mt-1 text-xs text-text-secondary">أدخل باركودًا مختلفًا لكل قطعة، والرقم التسلسلي اختياري.</p>
              </div>
              <Button type="button" variant="secondary" size="sm" onClick={() => setBatchRows((rows) => [...rows, { id: `${Date.now()}-${rows.length}`, barcode: '', serialNumber: '' }])}>
                <Plus className="h-4 w-4" /> إضافة قطعة
              </Button>
            </div>
            {batchRows.map((row, index) => (
              <div key={row.id} className="grid gap-2 md:grid-cols-[auto_1fr_1fr_auto] md:items-end">
                <span className="pb-3 text-xs font-semibold text-text-muted">#{index + 1}</span>
                <Input label="الباركود" value={row.barcode} onChange={(event) => setBatchRows((rows) => rows.map((current) => current.id === row.id ? { ...current, barcode: event.target.value } : current))} placeholder="باركود القطعة" required />
                <Input label="الرقم التسلسلي" value={row.serialNumber} onChange={(event) => setBatchRows((rows) => rows.map((current) => current.id === row.id ? { ...current, serialNumber: event.target.value } : current))} placeholder="اختياري" />
                <Button type="button" variant="ghost" size="icon" aria-label="حذف القطعة" disabled={batchRows.length === 1} onClick={() => setBatchRows((rows) => rows.filter((current) => current.id !== row.id))}>
                  <Trash2 className="h-4 w-4 text-danger" />
                </Button>
              </div>
            ))}
          </div>
        )}

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
              label="ربط بمورد (اختياري)"
              value={supplierId}
              onChange={(event) => setSupplierId(event.target.value)}
              options={[{ value: '', label: 'بدون مورد' }, ...supplierOptions]}
              loading={suppliersLoading}
              emptyMessage="لا يوجد موردون"
            />
          )}
          <p className="-mt-2 text-xs text-text-secondary md:col-span-3">
            {isUsedStock ? 'اختر عميلًا إذا أردت تسجيل مصدر هذه القطعة.' : 'اختر موردًا إذا أردت تسجيل اسم المورد المرتبط بهذه البضاعة.'}
          </p>
          <Input label="سعر التكلفة" type="number" min="0" step="0.01" value={purchaseCost} onChange={(event) => setPurchaseCost(event.target.value)} />
          <Input label="سعر البيع" type="number" min="0" step="0.01" value={sellingPrice} onChange={(event) => setSellingPrice(event.target.value)} />
        </div>

        <Input label="ملاحظات" value={notes} onChange={(event) => setNotes(event.target.value)} placeholder="مثال: موجود قبل تشغيل النظام" />

        <div className="flex items-center justify-between rounded-xl border border-border-subtle bg-surface-elevated px-4 py-3 text-xs text-text-secondary">
          <span className="flex items-center gap-2"><Warehouse className="h-4 w-4" /> النوع: {isUsedStock ? 'مخزون مستعمل موجود' : 'مخزون موجود مسبقًا'}</span>
          <span className="flex items-center gap-2"><CalendarDays className="h-4 w-4" /> {businessDate || 'اختر التاريخ'}</span>
        </div>

        <div className="flex justify-end gap-3">
          <Button type="button" variant="outline" onClick={onClose}>إلغاء</Button>
          <Button type="submit" disabled={isSubmitting}>{isSubmitting ? 'جارٍ الحفظ...' : isUsedStock ? 'حفظ المخزون المستعمل' : 'حفظ المخزون الحالي'}</Button>
        </div>
      </form>
    </Modal>
  );
}
