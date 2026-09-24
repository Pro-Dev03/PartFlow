import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { returnsApi } from '../../../services/api/endpoints';
import { customersApi, debtsApi, salesApi } from '../../../services/api/endpoints';
import { Button } from '../../../design-system/components/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Input } from '../../../design-system/components/input';
import { Select } from '../../../design-system/components/select';
import { SearchInput } from '../../../design-system/components/search-input';
import { PageHeader } from '../../../design-system/components/page-header';
import { ArrowRight, RotateCcw } from 'lucide-react';
import { toast } from 'sonner';
import { getStoreToday, parseBackendTimestamp } from '../../../utils/store-time';

const getPayload = (response: any) => response?.data ?? response;

interface CreateReturnPageProps {
  embedded?: boolean;
  onClose?: () => void;
}

export function CreateReturnPage({ embedded = false, onClose }: CreateReturnPageProps = {}) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const customerId = searchParams.get('customer_id') || '';
  const queryClient = useQueryClient();
  const [saleId, setSaleId] = useState('');
  const [saleSearch, setSaleSearch] = useState('');
  const [saleItemId, setSaleItemId] = useState('');
  const [quantity, setQuantity] = useState('1');
  const [reason, setReason] = useState('DEFECTIVE');
  const [condition, setCondition] = useState('READY_FOR_SALE');
  const [refundMethod, setRefundMethod] = useState('CASH');
  const [returnDate, setReturnDate] = useState(getStoreToday());
  const [notes, setNotes] = useState('');

  const { data: salesData, isLoading: salesLoading } = useQuery({
    queryKey: ['sales-available-for-return', customerId],
    queryFn: () => salesApi.list({ page: 1, per_page: 100, available_for_return: true, ...(customerId ? { customer_id: customerId } : {}) }),
  });

  const { data: saleData, isLoading: saleLoading } = useQuery({
    queryKey: ['sale-for-return', saleId],
    queryFn: () => salesApi.get(saleId),
    enabled: !!saleId,
  });

  const { data: customersData } = useQuery({
    queryKey: ['customers-for-return'],
    queryFn: () => customersApi.list({ page: 1, per_page: 100, is_active: true }),
  });

  const salesPayload = getPayload(salesData);
  const sales = Array.isArray(salesPayload) ? salesPayload : Array.isArray(salesPayload?.sales) ? salesPayload.sales : [];
  const selectedSale = sales.find((item: any) => String(item.id) === saleId);
  const customersPayload = getPayload(customersData);
  const customers = Array.isArray(customersPayload) ? customersPayload : Array.isArray(customersPayload?.data) ? customersPayload.data : [];
  const customerNames = new Map(customers.map((customer: any) => [String(customer.id), customer.name]));
  const getSaleCustomerName = (item: any) => item.customer_name
    || item.customerName
    || item.customer?.name
    || item.customer?.full_name
    || customerNames.get(String(item.customer_id || item.customerId))
    || '';
  const availableSales = Array.from(new Map(
    sales
      .filter((item: any) => Number(item.available_for_return ?? 1) > 0)
      .map((sale: any) => [String(sale.invoice_number || sale.invoiceNumber || sale.id), sale]),
  ).values()).sort((first: any, second: any) => {
    const getSaleTimestamp = (sale: any) => {
      const invoiceNumber = String(sale.invoice_number || sale.invoiceNumber || '');
      const invoiceTimestamp = invoiceNumber.match(/^INV-(\d{14})-/)?.[1];
      if (invoiceTimestamp) {
        const parsedInvoiceDate = parseBackendTimestamp(
          `${invoiceTimestamp.slice(0, 4)}-${invoiceTimestamp.slice(4, 6)}-${invoiceTimestamp.slice(6, 8)}T${invoiceTimestamp.slice(8, 10)}:${invoiceTimestamp.slice(10, 12)}:${invoiceTimestamp.slice(12, 14)}Z`,
        )?.getTime();
        if (parsedInvoiceDate !== undefined) return parsedInvoiceDate;
      }
      const saleDate = String(sale.sale_date || sale.saleDate || '');
      const createdAt = sale.created_at || sale.createdAt || '';
      return parseBackendTimestamp(saleDate.length > 10 ? saleDate : createdAt || saleDate)?.getTime() ?? 0;
    };
    const firstDate = getSaleTimestamp(first);
    const secondDate = getSaleTimestamp(second);
    return secondDate - firstDate;
  });
  const filteredSales = availableSales.filter((item: any) => {
    const customerName = getSaleCustomerName(item);
    const search = saleSearch.trim().toLowerCase();
    return !search || `${item.invoice_number || item.invoiceNumber || item.id} ${customerName} ${item.customer_phone || item.customerPhone || item.customer?.phone || ''}`.toLowerCase().includes(search);
  });
  const sale = getPayload(saleData);
  const saleCustomerId = sale?.customer_id || sale?.customerId || selectedSale?.customer_id || selectedSale?.customerId || '';
  const saleCustomerName = sale?.customer_name || sale?.customerName || selectedSale?.customer_name || selectedSale?.customerName || customerNames.get(String(saleCustomerId)) || '';
  const isGeneralCustomer = !saleCustomerName || ['عميل عام', 'زبون عام', 'walk-in customer', 'general customer'].includes(String(saleCustomerName).trim().toLowerCase());
  const { data: invoiceDebtsData, isLoading: invoiceDebtLoading, isError: invoiceDebtError } = useQuery({
    queryKey: ['debts-for-return-invoice'],
    queryFn: () => debtsApi.list({ page: 1, per_page: 100 }),
    enabled: !!saleId && !isGeneralCustomer,
  });
  const invoiceDebtsPayload = getPayload(invoiceDebtsData);
  const invoiceDebts = Array.isArray(invoiceDebtsPayload)
    ? invoiceDebtsPayload
    : Array.isArray(invoiceDebtsPayload?.debts)
      ? invoiceDebtsPayload.debts
      : [];
  const saleInvoiceNumber = sale?.invoice_number || sale?.invoiceNumber || selectedSale?.invoice_number || selectedSale?.invoiceNumber || '';
  const invoiceDebt = invoiceDebts.find((debt: any) =>
    String(debt.sale_id || debt.saleId || '') === saleId
    || String(debt.invoice_number || debt.invoiceNumber || '') === String(saleInvoiceNumber),
  );
  const items = Array.isArray(sale?.items) ? sale.items : [];
  const selectedItem = items.find((item: any) => String(item.id || item.sale_item_id) === saleItemId);
  const unitPrice = Number(selectedItem?.unit_price ?? selectedItem?.unitPrice ?? selectedItem?.price ?? 0);
  const invoiceSubtotal = Number(sale?.subtotal ?? selectedSale?.subtotal ?? 0);
  const invoiceDiscount = Number(sale?.discount_amount ?? selectedSale?.discount_amount ?? 0);
  const itemQuantityForPricing = Number(selectedItem?.quantity || selectedItem?.remaining_quantity || selectedItem?.available_quantity || 0);
  const grossItemTotal = unitPrice * itemQuantityForPricing;
  const grossInvoiceTotal = items.reduce((sum: number, item: any) => sum + (
    Number(item.unit_price ?? item.unitPrice ?? item.price ?? 0) * Number(item.quantity || item.remaining_quantity || item.available_quantity || 0)
  ), 0);
  const invoiceTotal = Number(sale?.total_amount ?? sale?.total ?? selectedSale?.total_amount ?? selectedSale?.total ?? 0);
  const effectiveInvoiceSubtotal = invoiceSubtotal > 0 ? invoiceSubtotal : grossInvoiceTotal;
  const effectiveDiscount = invoiceDiscount > 0
    ? invoiceDiscount
    : Math.max(effectiveInvoiceSubtotal - invoiceTotal, 0);
  const allocatedItemDiscount = effectiveDiscount > 0 && effectiveInvoiceSubtotal > 0
    ? effectiveDiscount * (grossItemTotal / effectiveInvoiceSubtotal)
    : 0;
  const netUnitPrice = selectedItem && itemQuantityForPricing > 0
    ? unitPrice - (allocatedItemDiscount / itemQuantityForPricing)
    : unitPrice;
  const inventorySource = String(selectedItem?.inventory_source || '').toLowerCase();
  const inventorySourceLabel = inventorySource === 'supplier_purchase'
    ? `تم شراؤه من المورد${selectedItem?.supplier_name ? `: ${selectedItem.supplier_name}` : ''}`
    : inventorySource === 'manual_inventory'
      ? 'أُضيف يدويًا للمخزون، وليس من فاتورة مورد'
      : 'مصدر المخزون غير محدد';
  const originalQuantity = Number(selectedItem?.quantity ?? selectedItem?.original_quantity ?? selectedItem?.originalQuantity ?? 0);
  const returnedQuantity = Number(selectedItem?.returned_quantity ?? selectedItem?.returnedQuantity ?? 0);
  const availableQuantity = Math.max(0, Number(selectedItem?.remaining_quantity ?? selectedItem?.remainingQuantity ?? originalQuantity - returnedQuantity));
  const invoiceDebtRemaining = Number(invoiceDebt?.remaining_amount ?? invoiceDebt?.remainingAmount ?? 0);
  const returnQuantity = Number(quantity);
  const returnType = returnQuantity < originalQuantity
    ? 'QUANTITY_PARTIAL'
    : items.length === 1
      ? 'FULL'
      : 'PARTIAL';
  const returnedCondition = reason === 'DEFECTIVE' || reason === 'DAMAGED'
    ? 'DEFECTIVE'
    : 'NEW';
  const canReturnToSupplier = inventorySource === 'supplier_purchase'
    && Boolean(selectedItem?.supplier_id || selectedItem?.supplierId);

  useEffect(() => {
    if (!canReturnToSupplier && condition === 'RETURN_TO_SUPPLIER') {
      setCondition('READY_FOR_SALE');
    }
  }, [canReturnToSupplier, condition]);

  useEffect(() => {
    if (!invoiceDebtLoading && refundMethod === 'DEBT_ADJUSTMENT' && invoiceDebtRemaining <= 0) {
      setRefundMethod('CASH');
    }
  }, [invoiceDebtLoading, invoiceDebtRemaining, refundMethod]);

  useEffect(() => {
    if (!saleId || saleLoading) return;
    const itemIds = items.map((item: any) => String(item.id || item.sale_item_id));
    if (itemIds.length === 1) {
      setSaleItemId(itemIds[0]);
      return;
    }
    if (saleItemId && !itemIds.includes(saleItemId)) {
      setSaleItemId('');
    }
  }, [items, saleId, saleItemId, saleLoading]);

  const createMutation = useMutation({
    mutationFn: () => returnsApi.create({
      sale_id: saleId,
      customer_id: sale?.customer_id || sale?.customerId || undefined,
      return_date: `${returnDate}T12:00:00Z`,
      return_type: returnType,
      reason,
      item_condition_after_return: condition,
      refund_method: refundMethod,
      notes,
      items: [{
        sale_item_id: saleItemId,
        product_id: selectedItem?.product_id ?? selectedItem?.productId,
        inventory_item_id: selectedItem?.inventory_item_id ?? selectedItem?.inventoryItemId,
        barcode: selectedItem?.barcode ?? selectedItem?.item_barcode,
        serial_number: selectedItem?.serial_number ?? selectedItem?.serialNumber,
        quantity_returned: returnQuantity,
        unit_price: netUnitPrice,
        total_refund_amount: netUnitPrice * returnQuantity,
        returned_condition: returnedCondition,
        resolution: condition === 'READY_FOR_SALE' ? 'RESTOCK' : condition === 'RETURN_TO_SUPPLIER' ? 'SUPPLIER_RETURN' : condition === 'NOT_FOR_SALE' ? 'WRITE_OFF' : 'REPAIR',
      }],
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['returns'] });
      queryClient.invalidateQueries({ queryKey: ['returns-statistics'] });
      queryClient.invalidateQueries({ queryKey: ['sales-returns-analysis'] });
      toast.success('تم إنشاء المرتجع بنجاح');
      if (onClose) {
        onClose();
      } else {
        navigate('/app/returns');
      }
    },
    onError: (error: any) => toast.error(String(error?.message || '').toLowerCase().includes('quantity') || String(error?.message || '').includes('الكمية')
      ? 'لا يمكن إنشاء المرتجع: تم إرجاع هذه الكمية سابقًا أو لا توجد كمية متاحة للإرجاع.'
      : 'تعذر إنشاء المرتجع، تحقق من الفاتورة والكمية'),
  });

  const canSubmit = !!saleId && !!saleItemId && Number(quantity) > 0 && Number(quantity) <= availableQuantity && unitPrice >= 0;

  return (
    <div>
      {!embedded && (
        <PageHeader
          title="إنشاء مرتجع جديد"
          description="اختر فاتورة البيع والعنصر المراد إرجاعه"
          actions={<Button variant="secondary" onClick={() => navigate('/app/returns')} className="gap-2"><ArrowRight className="w-4 h-4" /> العودة للمرتجعات</Button>}
        />
      )}
      <Card className={`${embedded ? 'border-0 shadow-none ' : ''}return-form-card`}>
        <CardHeader className="return-form-card-header"><CardTitle className="return-form-card-title flex items-center gap-2"><RotateCcw className="w-5 h-5" /> بيانات المرتجع</CardTitle></CardHeader>
        <CardContent className="return-form-card-content">
          <form className="return-form-grid return-form-grid-compact grid grid-cols-1 gap-3" onSubmit={(event) => { event.preventDefault(); if (canSubmit) createMutation.mutate(); }}>
            <div className="return-sale-search-field">
              <label className="mb-2 block text-sm font-medium text-text-secondary">فاتورة البيع</label>
              <SearchInput placeholder="ابحث برقم الفاتورة أو اسم العميل..." value={saleSearch} onChange={(event) => setSaleSearch(event.target.value)} onClear={() => setSaleSearch('')} size="sm" className="mb-2" />
            </div>
            <div className="return-selection-row">
              <div>
                <label className="mb-2 block text-sm font-medium text-text-secondary">اختر الفاتورة</label>
              <Select id="return-sale-select" aria-label="اختر الفاتورة" value={saleId} onChange={(event) => { setSaleId(event.target.value); setSaleItemId(''); }} options={[{ value: '', label: salesLoading ? 'جاري تحميل الفواتير...' : saleSearch.trim() && filteredSales.length === 0 ? 'لا توجد فواتير مطابقة' : 'اختر الفاتورة' }, ...filteredSales.map((item: any) => { const customerName = getSaleCustomerName(item); return { value: String(item.id), label: `${item.invoice_number || item.invoiceNumber || item.id}${customerName ? ` - ${customerName}` : ''} - ₪${Number(item.total_amount ?? item.total ?? 0).toLocaleString()}` }; })]} disabled={salesLoading} required />
              </div>
              <div>
                <label className="mb-2 block text-sm font-medium text-text-secondary">العنصر</label>
                <Select id="return-item-select" aria-label="العنصر" value={saleItemId} onChange={(event) => setSaleItemId(event.target.value)} options={[{ value: '', label: saleLoading ? 'جاري تحميل العناصر...' : 'اختر العنصر' }, ...items.map((item: any) => ({ value: String(item.id || item.sale_item_id), label: `${item.product_name || item.productName || 'منتج'} - ₪${Number(item.unit_price ?? item.unitPrice ?? item.price ?? 0).toLocaleString()}` }))]} disabled={!saleId || saleLoading} required />
              </div>
            </div>
            <div className="return-context-row">
              {saleId && (
                <div className="return-sale-summary rounded-lg border border-border bg-surface-muted p-3 text-sm">
                  <p className="text-sm text-text-secondary">العميل: <strong className="text-text-primary">{saleCustomerName || 'عميل عام'}</strong></p>
                  <div className="mt-2 grid gap-1">
                    <p className="text-text-secondary">إجمالي العناصر قبل الخصم: <strong className="text-text-primary">₪{Number(sale?.subtotal ?? selectedSale?.subtotal ?? sale?.total_amount ?? selectedSale?.total_amount ?? 0).toLocaleString()}</strong></p>
                    {invoiceDiscount > 0 && <p className="text-text-secondary">الخصم: <strong className="text-success">-₪{invoiceDiscount.toLocaleString()}</strong></p>}
                    <p className="text-text-secondary">إجمالي الفاتورة: <strong className="text-text-primary">₪{Number(sale?.total_amount ?? sale?.total ?? selectedSale?.total_amount ?? selectedSale?.total ?? 0).toLocaleString()}</strong></p>
                    <p className="text-text-secondary">الدين على الفاتورة: <strong className={isGeneralCustomer ? 'text-text-secondary' : 'text-danger'}>{isGeneralCustomer ? 'لا يوجد دين مرتبط بهذه الفاتورة' : invoiceDebtLoading ? 'جاري التحميل...' : invoiceDebtError ? 'غير متاح' : `₪${invoiceDebtRemaining.toLocaleString()}`}</strong></p>
                    {!isGeneralCustomer && !invoiceDebtLoading && invoiceDebtError && <p className="text-xs text-danger">تعذر التحقق من مديونية الفاتورة حاليًا. لا يمكن استخدام تعديل الدين حتى ينجح التحقق.</p>}
                    {!isGeneralCustomer && !invoiceDebtLoading && !invoiceDebtError && !invoiceDebt && <p className="text-xs text-text-tertiary">لا يوجد دين مرتبط بهذه الفاتورة.</p>}
                  </div>
                </div>
              )}
              {selectedItem && (
                <div className="return-item-details rounded-xl border border-border bg-surface-muted p-3">
                  <p className="text-xs font-medium uppercase tracking-wide text-text-secondary">تفاصيل العنصر</p>
                  <div className="return-item-details-grid mt-2 grid gap-2">
                  <div>
                    <p className="text-xs text-text-secondary">المنتج</p>
                    <p className="text-sm font-semibold text-text-primary">{selectedItem.product_name || selectedItem.productName || 'منتج غير مسمى'}</p>
                  </div>
                  <div>
                    <p className="text-xs text-text-secondary">طريقة إضافة المنتج</p>
                    <p className="text-sm font-semibold text-text-primary">{inventorySourceLabel}</p>
                  </div>
                  <div>
                    <p className="text-xs text-text-secondary">الكمية في الفاتورة</p>
                    <p className="text-sm font-semibold text-text-primary">{Number(selectedItem.quantity ?? selectedItem.remaining_quantity ?? 0)}</p>
                  </div>
                  <div>
                    <p className="text-xs text-text-secondary">تم إرجاعه سابقًا</p>
                    <p className="text-sm font-semibold text-text-primary">{returnedQuantity}</p>
                  </div>
                  <div>
                    <p className="text-xs text-text-secondary">المتاح للإرجاع</p>
                    <p className={`text-sm font-semibold ${availableQuantity > 0 ? 'text-success' : 'text-danger'}`}>{availableQuantity}</p>
                  </div>
                  <div>
                    <p className="text-xs text-text-secondary">سعر الوحدة</p>
                    <p className="text-sm font-semibold text-text-primary">₪{netUnitPrice.toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-xs text-text-secondary">إجمالي العنصر</p>
                    <p className="text-sm font-semibold text-text-primary">₪{(netUnitPrice * Number(selectedItem.quantity ?? selectedItem.remaining_quantity ?? 1)).toLocaleString()}</p>
                  </div>
                  </div>
                </div>
              )}
            </div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">الكمية</label><Input type="number" min="1" max={availableQuantity || undefined} value={quantity} onChange={(event) => setQuantity(event.target.value)} required /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">تاريخ المرتجع</label><Input type="date" lang="en-CA" dir="ltr" value={returnDate} onChange={(event) => setReturnDate(event.target.value)} required /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">سبب المرتجع</label><Select id="return-reason-select" aria-label="سبب الإرجاع" value={reason} onChange={(event) => setReason(event.target.value)} options={[{ value: 'DEFECTIVE', label: 'منتج معطل' }, { value: 'WRONG_ITEM', label: 'منتج خاطئ' }, { value: 'CUSTOMER_CHANGED_MIND', label: 'تغيير رأي العميل' }, { value: 'DAMAGED', label: 'تالف' }, { value: 'OTHER', label: 'أخرى' }]} /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">ماذا يحدث للمنتج بعد الإرجاع؟</label><Select id="return-condition-select" aria-label="ماذا يحدث للمنتج بعد الإرجاع؟" value={condition} onChange={(event) => setCondition(event.target.value)} options={[{ value: 'READY_FOR_SALE', label: 'إعادته للمخزون وبيعه مرة أخرى' }, { value: 'NOT_FOR_SALE', label: 'إخراجه من المخزون وعدم بيعه' }, ...(canReturnToSupplier ? [{ value: 'RETURN_TO_SUPPLIER', label: 'إرساله إلى المورد' }] : []), { value: 'NEEDS_REPAIR', label: 'إرساله للإصلاح' }]} />{!canReturnToSupplier && <p className="mt-1 text-xs text-text-muted">إرجاع المنتج للمورد متاح فقط للأصناف المرتبطة بفاتورة شراء ومورد.</p>}</div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">طريقة رد المبلغ</label><Select id="return-refund-select" aria-label="طريقة رد المبلغ" value={refundMethod} onChange={(event) => setRefundMethod(event.target.value)} options={[{ value: 'CASH', label: 'نقدي' }, ...(invoiceDebtRemaining > 0 ? [{ value: 'DEBT_ADJUSTMENT', label: `تعديل الدين (₪${invoiceDebtRemaining.toLocaleString()})` }] : [])]} /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">ملاحظات</label><Input value={notes} onChange={(event) => setNotes(event.target.value)} placeholder="ملاحظات اختيارية" /></div>
            <div className="return-form-submit flex items-center justify-between gap-4 border-t border-border pt-4"><p className="text-sm text-text-secondary">قيمة استرجاع العنصر ({selectedItem?.product_name || 'العنصر'} × {Number(quantity || 0)} بعد الخصم): <strong className="text-text-primary">₪{(netUnitPrice * Number(quantity || 0)).toLocaleString()}</strong></p><Button type="submit" variant="primary" disabled={!canSubmit || createMutation.isPending}>{createMutation.isPending ? 'جاري الحفظ...' : 'حفظ المرتجع'}</Button></div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
