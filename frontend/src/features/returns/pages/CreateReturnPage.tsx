import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { returnsApi } from '../../../services/api/endpoints';
import { customersApi, salesApi } from '../../../services/api/endpoints';
import { Button } from '../../../components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { SearchInput } from '../../../components/ui/search-input';
import { PageHeader } from '../../../components/ui/page-header';
import { ArrowRight, RotateCcw } from 'lucide-react';
import { toast } from 'sonner';

const getPayload = (response: any) => response?.data ?? response;

export function CreateReturnPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [saleId, setSaleId] = useState('');
  const [saleSearch, setSaleSearch] = useState('');
  const [saleItemId, setSaleItemId] = useState('');
  const [quantity, setQuantity] = useState('1');
  const [reason, setReason] = useState('DEFECTIVE');
  const [condition, setCondition] = useState('NEEDS_INSPECTION');
  const [refundMethod, setRefundMethod] = useState('CASH');
  const [returnDate, setReturnDate] = useState(new Date().toISOString().slice(0, 10));
  const [notes, setNotes] = useState('');

  const { data: salesData, isLoading: salesLoading } = useQuery({
    queryKey: ['sales-available-for-return'],
    queryFn: () => salesApi.list({ page: 1, per_page: 100, available_for_return: true }),
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
  const customersPayload = getPayload(customersData);
  const customers = Array.isArray(customersPayload) ? customersPayload : Array.isArray(customersPayload?.data) ? customersPayload.data : [];
  const customerNames = new Map(customers.map((customer: any) => [String(customer.id), customer.name]));
  const filteredSales = sales.filter((item: any) => {
    const customerName = item.customer_name || item.customerName || customerNames.get(String(item.customer_id || item.customerId)) || '';
    const search = saleSearch.trim().toLowerCase();
    return !search || `${item.invoice_number || item.invoiceNumber || item.id} ${customerName}`.toLowerCase().includes(search);
  });
  const sale = getPayload(saleData);
  const items = Array.isArray(sale?.items) ? sale.items : [];
  const selectedItem = items.find((item: any) => String(item.id || item.sale_item_id) === saleItemId);
  const unitPrice = Number(selectedItem?.unit_price ?? selectedItem?.unitPrice ?? selectedItem?.price ?? 0);
  const availableQuantity = Number(selectedItem?.quantity ?? selectedItem?.remaining_quantity ?? 0);

  const createMutation = useMutation({
    mutationFn: () => returnsApi.create({
      sale_id: saleId,
      customer_id: sale?.customer_id || sale?.customerId || undefined,
      return_date: `${returnDate}T12:00:00Z`,
      return_type: 'PARTIAL',
      reason,
      item_condition_after_return: condition,
      refund_method: refundMethod,
      notes,
      items: [{
        sale_item_id: saleItemId,
        product_id: selectedItem?.product_id ?? selectedItem?.productId,
        quantity_returned: Number(quantity),
        unit_price: unitPrice,
        total_refund_amount: unitPrice * Number(quantity),
        returned_condition: condition === 'SELLABLE' ? 'NEW' : 'DEFECTIVE',
        resolution: condition === 'SELLABLE' ? 'RESTOCK' : condition === 'SUPPLIER_RETURN' ? 'SUPPLIER_RETURN' : 'REPAIR',
        inspection_required: condition === 'NEEDS_INSPECTION',
      }],
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['returns'] });
      queryClient.invalidateQueries({ queryKey: ['returns-statistics'] });
      queryClient.invalidateQueries({ queryKey: ['sales-returns-analysis'] });
      toast.success('تم إنشاء المرتجع بنجاح');
      navigate('/app/returns');
    },
    onError: () => toast.error('تعذر إنشاء المرتجع، تحقق من الفاتورة والكمية'),
  });

  const canSubmit = !!saleId && !!saleItemId && Number(quantity) > 0 && Number(quantity) <= availableQuantity && unitPrice >= 0;

  return (
    <div>
      <PageHeader
        eyebrow="Returns Management"
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
              {saleId && <p className="mt-2 text-sm text-text-secondary">العميل: <strong className="text-text-primary">{sale?.customer_name || sale?.customerName || customerNames.get(String(sale?.customer_id || sale?.customerId)) || 'عميل عام'}</strong></p>}
            </div>
            <div>
              <label className="mb-2 block text-sm font-medium text-text-secondary">العنصر</label>
              <Select value={saleItemId} onChange={(event) => setSaleItemId(event.target.value)} options={[{ value: '', label: saleLoading ? 'جاري تحميل العناصر...' : 'اختر العنصر' }, ...items.map((item: any) => ({ value: String(item.id || item.sale_item_id), label: `${item.product_name || item.productName || 'منتج'} - ₪${Number(item.unit_price ?? item.unitPrice ?? item.price ?? 0).toLocaleString()}` }))]} disabled={!saleId || saleLoading} required />
            </div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">الكمية</label><Input type="number" min="1" max={availableQuantity || undefined} value={quantity} onChange={(event) => setQuantity(event.target.value)} required /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">تاريخ المرتجع</label><Input type="date" lang="en-CA" dir="ltr" value={returnDate} onChange={(event) => setReturnDate(event.target.value)} required /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">سبب المرتجع</label><Select value={reason} onChange={(event) => setReason(event.target.value)} options={[{ value: 'DEFECTIVE', label: 'منتج معطل' }, { value: 'WRONG_ITEM', label: 'منتج خاطئ' }, { value: 'CUSTOMER_CHANGED_MIND', label: 'تغيير رأي العميل' }, { value: 'DAMAGED', label: 'تالف' }, { value: 'OTHER', label: 'أخرى' }]} /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">حالة المنتج بعد الإرجاع</label><Select value={condition} onChange={(event) => setCondition(event.target.value)} options={[{ value: 'SELLABLE', label: 'قابل للبيع' }, { value: 'NEEDS_INSPECTION', label: 'يحتاج فحص' }, { value: 'NEEDS_REPAIR', label: 'يحتاج إصلاح' }, { value: 'DAMAGED', label: 'تالف' }, { value: 'SUPPLIER_RETURN', label: 'إرجاع للمورد' }]} /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">طريقة رد المبلغ</label><Select value={refundMethod} onChange={(event) => setRefundMethod(event.target.value)} options={[{ value: 'CASH', label: 'نقدي' }, { value: 'CREDIT', label: 'رصيد العميل' }, { value: 'DEBT_ADJUSTMENT', label: 'تعديل الدين' }, { value: 'STORE_CREDIT', label: 'رصيد المتجر' }]} /></div>
            <div><label className="mb-2 block text-sm font-medium text-text-secondary">ملاحظات</label><Input value={notes} onChange={(event) => setNotes(event.target.value)} placeholder="ملاحظات اختيارية" /></div>
            <div className="md:col-span-2 flex items-center justify-between gap-4 border-t border-border pt-4"><p className="text-sm text-text-secondary">قيمة الاسترجاع: <strong className="text-text-primary">₪{(unitPrice * Number(quantity || 0)).toLocaleString()}</strong></p><Button type="submit" variant="primary" disabled={!canSubmit || createMutation.isPending}>{createMutation.isPending ? 'جاري الحفظ...' : 'حفظ المرتجع'}</Button></div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
