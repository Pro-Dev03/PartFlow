import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { supplierReturnsApi, purchasesApi } from '../../../services/api/endpoints';
import { PageHeader } from '../../../design-system/components/page-header';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { Select } from '../../../design-system/components/select';
import { StatCard } from '../../../design-system/components/stat-card';
import { toast } from 'sonner';
import { printHtmlDocument } from '../../../services/documents/print-html';
import { ConfirmDialog } from '../../../design-system/components/confirm-dialog';
import {
  Banknote,
  CheckCircle2,
  Clock3,
  PackageCheck,
  Printer,
  Search,
  Trash2,
  XCircle,
  ArrowRight,
} from 'lucide-react';
import { formatDateTime } from '../../../utils/format';

const returnReasonLabels: Record<string, string> = {
  DAMAGED: 'منتج تالف',
  DEFECTIVE: 'منتج معيب',
  WRONG_ITEM: 'صنف غير صحيح',
  CUSTOMER_CHANGED_MIND: 'تغيير رأي العميل',
  WARRANTY: 'ضمان',
};

const statusLabels: Record<string, string> = {
  PENDING: 'قيد الانتظار',
  SHIPPED: 'تم الشحن',
  RECEIVED: 'تم استلامه لدى المورد',
  COMPLETED: 'مكتمل',
  REJECTED: 'مرفوض',
  NEEDS_SOURCE_DATA: 'يحتاج بيانات المصدر',
  RESOLVED: 'مصدر البيانات مكتمل',
  NEEDS_SOURCE: 'يحتاج بيانات المصدر',
};

const readableReturnReason = (value?: string) => returnReasonLabels[String(value || '').toUpperCase()] || value || 'غير محدد';
const readableStatus = (value?: string) => statusLabels[String(value || '').toUpperCase()] || value || 'غير محدد';

export function SupplierReturnsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [purchaseId, setPurchaseId] = useState('');
  const [purchaseItemId, setPurchaseItemId] = useState('');
  const [quantity, setQuantity] = useState('1');
  const [reason, setReason] = useState('');
  const [notes, setNotes] = useState('');
  const [showClosed, setShowClosed] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [confirmAction, setConfirmAction] = useState<{ type: 'delete' | 'reject'; id: string } | null>(null);
  const { data: returnsData, isLoading } = useQuery({
    queryKey: ['supplier-returns'],
    queryFn: () => supplierReturnsApi.list(),
  });
  const { data: purchasesData } = useQuery({
    queryKey: ['purchases', 'supplier-return-options'],
    queryFn: () => purchasesApi.list({ page: 1, per_page: 100, available_for_return: true }),
  });
  const { data: purchaseDetailsData } = useQuery({
    queryKey: ['purchase', 'supplier-return', purchaseId],
    queryFn: () => purchasesApi.get(purchaseId),
    enabled: Boolean(purchaseId),
  });
  const purchaseItems = purchaseDetailsData?.data?.items || purchaseDetailsData?.items || [];
  const selectedItem = purchaseItems.find((item: any) => item.id === purchaseItemId);
  const availableQuantity = Number(selectedItem?.available_for_return || 0);
  const createMutation = useMutation({
    mutationFn: async () => {
      const created = await supplierReturnsApi.create({ purchase_id: purchaseId, reason, notes });
      const returnId = created?.data?.id || created?.id;
      await supplierReturnsApi.addItem(returnId, {
        purchase_item_id: purchaseItemId,
        quantity: Number(quantity),
      });
      return created;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['supplier-returns'] });
      setPurchaseId('');
      setPurchaseItemId('');
      setQuantity('1');
      setReason('');
      setNotes('');
      toast.success('تم إنشاء طلب إرجاع للمورد');
    },
    onError: () => toast.error('تعذر إنشاء طلب إرجاع للمورد'),
  });
  const completeMutation = useMutation({
    mutationFn: (id: string) => supplierReturnsApi.complete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['supplier-returns'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['purchases', 'supplier-return-options'] });
      toast.success('تم إكمال إرجاع المورد وتحديث المخزون');
    },
    onError: () => toast.error('تعذر إكمال إرجاع المورد'),
  });
  const deleteMutation = useMutation({
    mutationFn: (id: string) => supplierReturnsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['supplier-returns'] });
      setConfirmAction(null);
      toast.success('تم حذف طلب الإرجاع غير المعالج');
    },
    onError: () => toast.error('لا يمكن حذف هذا الطلب؛ قد يكون بدأ معالجته أو يحتوي أصنافاً'),
  });
  const rejectMutation = useMutation({
    mutationFn: (id: string) => supplierReturnsApi.reject(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['supplier-returns'] });
      setConfirmAction(null);
      toast.success('تم رفض طلب الإرجاع');
    },
    onError: () => toast.error('لا يمكن رفض هذا الطلب في حالته الحالية'),
  });
  const purchases = Array.from(new Map(
    (purchasesData?.data || [])
      .filter((p: any) => !['reversed', 'cancelled'].includes(String(p.status || '').toLowerCase()))
      .map((purchase: any) => [purchase.id, purchase]),
  ).values());
  const returns = returnsData?.data || [];
  const pendingReturns = returns.filter((item: any) => ['PENDING', 'SHIPPED', 'RECEIVED', 'NEEDS_SOURCE_DATA'].includes(item.status)).length;
  const completedReturns = returns.filter((item: any) => item.status === 'COMPLETED').length;
  const refundTotal = returns
    .filter((item: any) => item.status === 'COMPLETED')
    .reduce((total: number, item: any) => total + Number(item.refund_amount || 0), 0);
  const visibleReturns = returns.filter((item: any) => {
    const isClosed = ['COMPLETED', 'REJECTED', 'CANCELLED'].includes(item.status);
    const matchesView = showClosed ? isClosed : !isClosed;
    const query = searchQuery.trim().toLowerCase();
    const matchesSearch = !query || [item.return_number, item.reason, item.purchase_id, item.supplier_id]
      .some((value) => String(value || '').toLowerCase().includes(query));
    return matchesView && matchesSearch;
  });
  const printReturn = (item: any) => {
    const returnItem = item.items?.[0] || item.item;
    void printHtmlDocument(`
      <html dir="rtl" lang="ar"><head><title>مرتجع مورد ${item.return_number}</title>
      <style>body{font-family:Arial,sans-serif;padding:32px;color:#111}h1{margin-bottom:24px}p{margin:10px 0}.line{border-bottom:1px solid #ddd;padding:12px 0}</style>
      </head><body>
      <h1>طلب مرتجع مورد</h1>
      <p><strong>رقم المرتجع:</strong> ${item.return_number}</p>
      <p><strong>رقم الشراء:</strong> ${item.purchase_id}</p>
      <p><strong>الحالة:</strong> ${item.status}</p>
      <p><strong>السبب:</strong> ${item.reason}</p>
      ${returnItem ? `<p><strong>الصنف:</strong> ${returnItem.product_name || returnItem.product?.name || '-'} | الكمية: ${returnItem.quantity || 0} | تكلفة الوحدة: ₪${Number(returnItem.unit_cost || 0).toLocaleString('en-US')}</p>` : ''}
      <p><strong>قيمة الاسترداد:</strong> ₪${Number(item.refund_amount || 0).toLocaleString('en-US')}</p>
      <p class="line"><strong>ملاحظة:</strong> ${item.notes || '-'}</p>
      </body></html>
    `).catch(() => toast.error('تعذر فتح نافذة الطباعة'));
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="مرتجعات التجار"
        description="إرجاع البضاعة إلى التاجر دون خلطها بمرتجعات العملاء"
        actions={
          <Button variant="secondary" className="gap-2" onClick={() => navigate('/app/returns')}>
            <ArrowRight className="h-4 w-4" />
            العودة للمرتجعات
          </Button>
        }
      />
      <div className="unified-stats-grid supplier-stats grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-md">
        <StatCard
          title="إجمالي طلبات الإرجاع"
          value={returns.length}
          icon={PackageCheck}
          variant="featured"
          size="sm"
        />
        <StatCard
          title="قيد الانتظار"
          value={pendingReturns}
          icon={Clock3}
          variant="default"
          size="sm"
        />
        <StatCard
          title="تمت الإعادة"
          value={completedReturns}
          icon={CheckCircle2}
          variant="default"
          size="sm"
        />
        <StatCard
          title="إجمالي المسترد من الموردين"
          value={`₪${refundTotal.toLocaleString('en-US', { maximumFractionDigits: 2 })}`}
          icon={Banknote}
          variant="default"
          size="sm"
        />
      </div>
      <Card>
        <CardHeader className="border-b border-border px-4 py-3" style={{ marginBottom: 0, padding: '14px 16px' }}>
            <div>
              <CardTitle>طلب إرجاع جديد</CardTitle>
              <p className="mt-1 text-xs text-text-muted">اختر الفاتورة والصنف ثم أدخل تفاصيل الإرجاع.</p>
            </div>
            <PackageCheck className="h-5 w-5 text-primary" />
          </CardHeader>
          <CardContent className="space-y-4" style={{ padding: '16px' }}>
            <div className="supplier-return-form-grid">
            <div className="supplier-return-invoice-field space-y-2">
              <label className="text-sm font-medium">1. فاتورة الشراء</label>
              <Select value={purchaseId} onChange={(e) => { setPurchaseId(e.target.value); setPurchaseItemId(''); }} options={[
                { value: '', label: 'اختر فاتورة تحتوي على مخزون متاح' },
                ...purchases.map((p: any) => ({ value: p.id, label: `${p.invoice_number} | ${p.supplier_name || 'مورد غير محدد'} | ${p.available_for_return || 0} قطعة متاحة` })),
              ]} />
              <p className="text-xs text-text-muted">تظهر هنا الفواتير التي يمكن إرجاع صنف منها فقط.</p>
            </div>
            <div className="supplier-return-item-field space-y-2">
              <label className="text-sm font-medium">2. الصنف المراد إرجاعه</label>
              <Select value={purchaseItemId} onChange={(e) => setPurchaseItemId(e.target.value)} options={[
                { value: '', label: purchaseId ? 'اختر الصنف من هذه الفاتورة' : 'اختر الفاتورة أولاً' },
                ...purchaseItems.map((item: any) => ({
                  value: item.id,
                  label: `${item.product_name || item.product?.name || 'صنف غير مسمى'} | فاتورة: ${item.quantity} | تكلفة: ₪${Number(item.unit_cost || 0).toLocaleString('en-US')}`,
                })),
              ]} disabled={!purchaseId || purchaseItems.length === 0} />
              <p className="text-xs text-text-muted">بعد اختيار الصنف ستظهر الكمية الموجودة فعلياً.</p>
            </div>
          {selectedItem && (
            <div className="supplier-return-selected-item grid gap-3 rounded border border-border bg-surface-muted p-4 text-sm sm:grid-cols-2 lg:grid-cols-4">
              <div><span className="text-text-muted">الصنف</span><p className="font-medium">{selectedItem.product_name || selectedItem.product?.name || 'الصنف المحدد'}</p></div>
              <div><span className="text-text-muted">الباركود</span><p>{selectedItem.barcode || '-'}</p></div>
              <div><span className="text-text-muted">SKU داخلي</span><p>{selectedItem.sku || '-'}</p></div>
              <div><span className="text-text-muted">الكمية المشتراة</span><p>{selectedItem.quantity || 0}</p></div>
              <div><span className="text-text-muted">الكمية المستلمة</span><p>{selectedItem.received_quantity || 0}</p></div>
              <div><span className="text-text-muted">الكمية المرتجعة</span><p>{selectedItem.returned_quantity || 0}</p></div>
              <div><span className="text-text-muted">تكلفة الوحدة</span><p>₪{Number(selectedItem.unit_cost || 0).toLocaleString('en-US')}</p></div>
              <div className="sm:col-span-2 lg:col-span-4">
                <span className="text-text-muted">المتاح للإرجاع</span>
                <p className={availableQuantity > 0 ? 'font-semibold text-success' : 'font-semibold text-danger'}>{availableQuantity} قطعة</p>
              </div>
            </div>
          )}
          <div className="supplier-return-quantity-field">
            <Input type="number" min="1" max={availableQuantity || undefined} value={quantity} onChange={(e) => setQuantity(e.target.value)} placeholder="الكمية المراد إرجاعها" />
          </div>
          <div className="supplier-return-reason-field">
            <Input value={reason} onChange={(e) => setReason(e.target.value)} placeholder="سبب الإرجاع (مطلوب)" />
          </div>
          <div className="supplier-return-notes-field">
            <Input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="ملاحظات (اختياري)" />
          </div>
          </div>
          <Button size="sm" className="w-full md:w-auto" disabled={!purchaseId || !purchaseItemId || !selectedItem || availableQuantity < 1 || Number(quantity) < 1 || Number(quantity) > availableQuantity || !reason || createMutation.isPending} onClick={() => createMutation.mutate()}>
            إنشاء طلب الإرجاع
          </Button>
          </CardContent>
        </Card>
      <Card style={{ padding: 0 }}>
        <CardHeader className="border-b border-border px-4 py-3" style={{ marginBottom: 0, padding: '14px 16px' }}>
          <div>
            <CardTitle>طلبات إرجاع التجار</CardTitle>
            <p className="mt-1 text-xs text-text-muted">تابع الطلبات المفتوحة أو راجع الطلبات المنتهية.</p>
          </div>
          <span className="rounded-full bg-surface-muted px-2.5 py-1 text-xs text-text-muted">{visibleReturns.length} نتيجة</span>
        </CardHeader>
        <CardContent style={{ padding: '16px' }}>
          <div className="mb-4 flex flex-col gap-2 rounded-xl border border-border bg-surface-muted/50 p-2 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex gap-2">
              <Button size="sm" variant={!showClosed ? 'primary' : 'secondary'} onClick={() => setShowClosed(false)}>
                الطلبات النشطة
              </Button>
              <Button size="sm" variant={showClosed ? 'primary' : 'secondary'} onClick={() => setShowClosed(true)}>
                الطلبات المنتهية
              </Button>
            </div>
            <div className="relative w-full sm:max-w-xs">
              <Search className="pointer-events-none absolute right-3 top-1/2 z-10 h-4 w-4 -translate-y-1/2 text-text-muted" />
              <Input className="pr-9" value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} placeholder="بحث برقم المرتجع أو السبب" />
            </div>
          </div>
          {isLoading ? <p>جار التحميل...</p> : visibleReturns.length === 0 ? <p className="text-text-muted">{showClosed ? 'لا توجد طلبات منتهية' : 'لا توجد طلبات نشطة'}</p> : (
            <div className="space-y-2">{visibleReturns.map((item: any) => (
              <div key={item.id} className="supplier-return-list-row rounded-xl border border-border bg-surface-muted/30 p-3">
                <span className="supplier-return-number text-sm font-semibold text-text">رقم الطلب: {item.return_number}</span>
                <span className="supplier-return-product text-sm font-semibold text-text">اسم القطعة: {item.product_name || item.product?.name || '-'}</span>
                <span className="supplier-return-reason text-xs text-text-muted">السبب: {readableReturnReason(item.reason)}</span>
                <span className="supplier-return-status text-xs text-text-muted">الحالة: {item.needs_source_resolution || item.status === 'NEEDS_SOURCE_DATA' ? 'يحتاج بيانات المصدر' : readableStatus(item.status)}</span>
                <span className="supplier-return-amount text-sm font-semibold text-info">₪{Number(item.refund_amount || 0).toLocaleString('en-US')}</span>
                {item.customer_return_id && item.customer_return_id !== '00000000-0000-0000-0000-000000000000' && (
                  <div className={`supplier-return-source-grid rounded-lg border p-3 text-xs ${item.needs_source_resolution || item.status === 'NEEDS_SOURCE_DATA' ? 'border-warning/40 bg-warning/10' : 'border-border bg-surface-muted'}`}>
                    <strong className="supplier-return-source-title">{item.needs_source_resolution || item.status === 'NEEDS_SOURCE_DATA' ? 'مرتجع عميل يحتاج بيانات الشراء أو المورد' : 'مصدر الطلب: مرتجع عميل'}</strong>
                    <span>اسم القطعة: {item.product_name || item.product?.name || '-'}</span>
                    <span data-testid="unresolved-customer-return-id">معرّف مرتجع العميل: {item.customer_return_id}</span>
                    <span>معرّف البيع: {item.sale_id || '-'}</span>
                    <span>المورد: {item.supplier_name || item.supplier_id || '-'}</span>
                    <span>معرّف الشراء: {item.purchase_id || '-'}</span>
                    <span>معرّف قطعة المخزون: {item.inventory_item_id || '-'}</span>
                    <span>الباركود: {item.barcode || '-'}</span>
                    <span>الرقم التسلسلي: {item.serial_number || '-'}</span>
                    <span>الكمية: {item.quantity || 0}</span>
                    <span>تكلفة الشراء: ₪{Number(item.purchase_cost || 0).toLocaleString('en-US')}</span>
                    <span>سبب الإرجاع: {readableReturnReason(item.return_reason || item.reason)}</span>
                    <span>تاريخ الإرجاع: {item.return_date ? formatDateTime(item.return_date, 'ar-EG') : 'غير محدد'}</span>
                    <span>حالة المصدر: {readableStatus(item.source_status || item.status)}</span>
                    <span>مصدر الطلب: {item.source === 'Customer Return' ? 'مرتجع عميل' : item.source || 'غير محدد'}</span>
                  </div>
                )}
                <Button className="supplier-return-print" size="sm" variant="secondary" title="طباعة طلب المرتجع" onClick={() => printReturn(item)}>
                  <Printer className="h-4 w-4" /> طباعة
                </Button>
                {item.status !== 'COMPLETED' && (
                  <div className="supplier-return-actions flex gap-2">
                    {item.status === 'PENDING' ? (
                      <Button
                        size="sm"
                        variant="danger"
                        title="يحذف الطلب فقط إذا لم تتم معالجته ولم يُرسل للمورد"
                        onClick={() => setConfirmAction({ type: 'delete', id: item.id })}
                      >
                        <Trash2 className="h-4 w-4" /> حذف الطلب
                      </Button>
                    ) : (
                      <Button
                        size="sm"
                        variant="danger"
                        title="يرفض الطلب مع حفظ حالته في سجل المرتجعات"
                        onClick={() => setConfirmAction({ type: 'reject', id: item.id })}
                      >
                        <XCircle className="h-4 w-4" /> رفض الطلب
                      </Button>
                    )}
                    <Button
                      size="sm"
                      variant="success"
                      title={item.needs_source_resolution || item.status === 'NEEDS_SOURCE_DATA'
                        ? 'لا يمكن الإكمال قبل تحديد فاتورة الشراء والمورد وقطعة المخزون'
                        : 'يكمل الإرجاع ويخصم الكمية من المخزون ويسجل قيمته'}
                      onClick={() => completeMutation.mutate(item.id)}
                      disabled={completeMutation.isPending || item.needs_source_resolution || item.status === 'NEEDS_SOURCE_DATA'}
                    >
                      {item.needs_source_resolution || item.status === 'NEEDS_SOURCE_DATA' ? 'بانتظار بيانات المصدر' : 'إكمال الإرجاع'}
                    </Button>
                  </div>
                )}
              </div>
            ))}</div>
          )}
        </CardContent>
      </Card>
      <ConfirmDialog
        isOpen={Boolean(confirmAction)}
        onClose={() => setConfirmAction(null)}
        onConfirm={() => {
          if (!confirmAction) return;
          if (confirmAction.type === 'delete') deleteMutation.mutate(confirmAction.id);
          else rejectMutation.mutate(confirmAction.id);
        }}
        title={confirmAction?.type === 'delete' ? 'حذف طلب الإرجاع' : 'رفض طلب الإرجاع'}
        message={confirmAction?.type === 'delete'
          ? 'سيُحذف الطلب نهائياً لأنه لم يُعالج بعد. لا تستخدم هذا الخيار إذا أُرسلت البضاعة للمورد.'
          : 'سيتم تغيير حالة الطلب إلى مرفوض مع إبقاء سجل العملية للرجوع إليه.'}
        confirmText={confirmAction?.type === 'delete' ? 'حذف الطلب' : 'رفض الطلب'}
        variant="danger"
        isLoading={deleteMutation.isPending || rejectMutation.isPending}
      />
    </div>
  );
}
