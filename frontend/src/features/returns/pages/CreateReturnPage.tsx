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

const getPayload = (response: any) => response?.data ?? response;

export function CreateReturnPage() {
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
  const [returnDate, setReturnDate] = useState(new Date().toISOString().slice(0, 10));
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
    queryFn: () => customersApi.list({ page: 1, per_page: 100 }),
  });

  const salesPayload = getPayload(salesData);
  const sales = Array.isArray(salesPayload) ? salesPayload : Array.isArray(salesPayload?.sales) ? salesPayload.sales : [];
  const selectedSale = sales.find((item: any) => String(item.id) === saleId);
  const customersPayload = getPayload(customersData);
  const customers = Array.isArray(customersPayload) ? customersPayload : Array.isArray(customersPayload?.data) ? customersPayload.data : [];
  const customerNames = new Map(customers.map((customer: any) => [String(customer.id), customer.name]));
  const filteredSales = sales.filter((item: any) => {
    const customerName = item.customer_name || item.customerName || customerNames.get(String(item.customer_id || item.customerId)) || '';
    const search = saleSearch.trim().toLowerCase();
    return !search || `${item.invoice_number || item.invoiceNumber || item.id} ${customerName}`.toLowerCase().includes(search);
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
      ? 'مضاف مباشرة إلى المخزون'
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
        returned_condition: condition === 'READY_FOR_SALE' ? 'NEW' : 'DEFECTIVE',
        resolution: condition === 'READY_FOR_SALE' ? 'RESTOCK' : condition === 'RETURN_TO_SUPPLIER' ? 'SUPPLIER_RETURN' : condition === 'NOT_FOR_SALE' ? 'WRITE_OFF' : 'REPAIR',
      }],
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['returns'] });
      queryClient.invalidateQueries({ queryKey: ['returns-statistics'] });
      queryClient.invalidateQueries({ queryKey: ['sales-returns-analysis'] });
      toast.success('تم إنشاء المرتجع بنجاح');
      navigate('/app/returns');
    },
    onError: (error: any) => toast.error(String(error?.message || '').toLowerCase().includes('quantity') || String(error?.message || '').includes('الكمية')
      ? 'لا يمكن إنشاء المرتجع: تم إرجاع هذه الكمية سابقًا أو لا توجد كمية متاحة للإرجاع.'
      : 'تعذر إنشاء المرتجع، تحقق من الفاتورة والكمية'),
  });

  const canSubmit = !!saleId && !!saleItemId && Number(quantity) > 0 && Number(quantity) <= availableQuantity && unitPrice >= 0;

  return (
    <div>
      <PageHeader
        title="إنشاء مرتجع جديد"
        description="اختر فاتورة البيع والعنصر المراد إرجاعه"
        actions={<Button variant="secondary" onClick={() => navigate('/app/returns')} className="gap-2"><ArrowRight className="w-4 h-4" /> العودة للمرتجعات</Button>}
      />
      <Card>
        <CardHeader><CardTitle className="flex items-center gap-2"><RotateCcw className="w-5 h-5" /> بيانات المرتجع</CardTitle></CardHeader>
        <CardContent>
          <form className="return-form-grid grid grid-cols-1 gap-4 md:grid-cols-2" onSubmit={(event) => { event.preventDefault(); if (canSubmit) createMutation.mutate(); }}>
            <div>
              <label className="mb-2 block text-sm font-medium text-text-secondary">فاتورة البيع</label>
              <SearchInput placeholder="ابحث برقم الفاتورة أو اسم العميل..." value={saleSearch} onChange={(event) => setSaleSearch(event.target.value)} onClear={() => setSaleSearch('')} size="sm" className="mb-2" />
              <Select value={saleId} onChange={(event) => { setSaleId(event.target.value); setSaleItemId(''); }} options={[{ value: '', label: salesLoading ? 'جاري تحميل الفواتير...' : 'اختر الفاتورة' }, ...filteredSales.map((item: any) => { const customerName = item.customer_name || item.customerName || customerNames.get(String(item.customer_id || item.customerId)); return { value: String(item.id), label: `${item.invoice_number || item.invoiceNumber || item.id}${customerName ? ` - ${customerName}` : ''} - ₪${Number(item.total_amount ?? item.total ?? 0).toLocaleString()}` }; })]} disabled={salesLoading} required />
              {saleId && <>
                <p className="mt-2 text-sm text-text-secondary">العميل: <strong className="text-text-primary">{saleCustomerName || 'عميل عام'}</strong></p>
                <div className="mt-3 rounded-lg border border-border bg-surface-muted p-3 text-sm">
                  <p className="text-text-secondary">إجمالي العناصر قبل الخصم: <strong className="text-text-primary">₪{Number(sale?.subtotal ?? selectedSale?.subtotal ?? sale?.total_amount ?? selectedSale?.total_amount ?? 0).toLocaleString()}</strong></p>
                  {invoiceDiscount > 0 && <p className="mt-1 text-text-secondary">الخصم: <strong className="text-success">-₪{invoiceDiscount.toLocaleString()}</strong></p>}
                  <p className="mt-1 text-text-secondary">إجمالي الفاتورة: <strong className="text-text-primary">₪{Number(sale?.total_amount ?? sale?.total ?? selectedSale?.total_amount ?? selectedSale?.total ?? 0).toLocaleString()}</strong></p>
                  <p className="mt-1 text-text-secondary">الدين على الفاتورة: <strong className={isGeneralCustomer ? 'text-text-secondary' : 'text-danger'}>{isGeneralCustomer ? 'لا يوجد دين مرتبط بهذه الفاتورة' : invoiceDebtLoading ? 'جاري التحميل...' : invoiceDebtError ? 'غير متاح' : `₪${invoiceDebtRemaining.toLocaleString()}`}</strong></p>
                  {!isGeneralCustomer && !invoiceDebtLoading && invoiceDebtError && <p className="mt-1 text-xs text-danger">تعذر التحقق من مديونية الفاتورة حاليًا. لا يمكن استخدام تعديل الدين حتى ينجح التحقق.</p>}
                  {!isGeneralCustomer && !invoiceDebtLoading && !invoiceDebtError && !invoiceDebt && <p className="mt-1 text-xs text-text-tertiary">لا يوجد دين مرتبط بهذه الفاتورة.</p>}
                </div>
              </>}
            </div>
            <div>
              <label className="mb-2 block text-sm font-medium text-text-secondary">العنصر</label>
              <Select value={saleItemId} onChange={(event) => setSaleItemId(event.target.value)} options={[{ value: '', label: saleLoading ? 'جاري تحميل العناصر...' : 'اختر العنصر' }, ...items.map((item: any) => ({ value: String(item.id || item.sale_item_id), label: `${item.product_name || item.productName || 'منتج'} - ₪${Number(item.unit_price ?? item.unitPrice ?? item.price ?? 0).toLocaleString()}` }))]} disabled={!saleId || saleLoading} required />
            </div>
            {selectedItem && (
              <div className="md:col-span-2 rounded-xl border border-border bg-surface-muted p-4">
                <p className="text-xs font-medium uppercase tracking-wide text-text-secondary">تفاصيل العنصر</p>
                <div className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                  <div>
                    <p className="text-xs text-text-secondary">المنتج</p>
                    <p className="text-sm font-semibold text-text-primary">{selectedItem.product_name || selectedItem.productName || 'منتج غير مسمى'}</p>
                  </div>
                  <div>
                    <p className="text-xs text-text-secondary">مصدر العنصر</p>
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
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">الكمية</label><Input type="number" min="1" max={availableQuantity || undefined} value={quantity} onChange={(event) => setQuantity(event.target.value)} required /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">تاريخ المرتجع</label><Input type="date" lang="en-CA" dir="ltr" value={returnDate} onChange={(event) => setReturnDate(event.target.value)} required /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">سبب المرتجع</label><Select value={reason} onChange={(event) => setReason(event.target.value)} options={[{ value: 'DEFECTIVE', label: 'منتج معطل' }, { value: 'WRONG_ITEM', label: 'منتج خاطئ' }, { value: 'CUSTOMER_CHANGED_MIND', label: 'تغيير رأي العميل' }, { value: 'DAMAGED', label: 'تالف' }, { value: 'OTHER', label: 'أخرى' }]} /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">ماذا يحدث للمنتج بعد الإرجاع؟</label><Select value={condition} onChange={(event) => setCondition(event.target.value)} options={[{ value: 'READY_FOR_SALE', label: 'إعادته للمخزون وبيعه مرة أخرى' }, { value: 'NOT_FOR_SALE', label: 'إخراجه من المخزون وعدم بيعه' }, { value: 'RETURN_TO_SUPPLIER', label: 'إرساله إلى المورد' }, { value: 'NEEDS_REPAIR', label: 'إرساله للإصلاح' }]} /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">طريقة رد المبلغ</label><Select value={refundMethod} onChange={(event) => setRefundMethod(event.target.value)} options={[{ value: 'CASH', label: 'نقدي' }, ...(invoiceDebtRemaining > 0 ? [{ value: 'DEBT_ADJUSTMENT', label: `تعديل الدين (₪${invoiceDebtRemaining.toLocaleString()})` }] : [])]} /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">ملاحظات</label><Input value={notes} onChange={(event) => setNotes(event.target.value)} placeholder="ملاحظات اختيارية" /></div>
            <div className="md:col-span-2 flex items-center justify-between gap-4 border-t border-border pt-4"><p className="text-sm text-text-secondary">قيمة استرجاع العنصر ({selectedItem?.product_name || 'العنصر'} × {Number(quantity || 0)} بعد الخصم): <strong className="text-text-primary">₪{(netUnitPrice * Number(quantity || 0)).toLocaleString()}</strong></p><Button type="submit" variant="primary" disabled={!canSubmit || createMutation.isPending}>{createMutation.isPending ? 'جاري الحفظ...' : 'حفظ المرتجع'}</Button></div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
