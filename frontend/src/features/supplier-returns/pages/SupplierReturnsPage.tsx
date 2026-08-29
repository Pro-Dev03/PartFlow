import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { supplierReturnsApi, purchasesApi, inventoryApi } from '../../../services/api/endpoints';
import { PageHeader } from '../../../components/ui/page-header';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { toast } from 'sonner';
import { ConfirmDialog } from '../../../components/ui/confirm-dialog';
import { Printer, Trash2, XCircle } from 'lucide-react';

export function SupplierReturnsPage() {
  const queryClient = useQueryClient();
  const [purchaseId, setPurchaseId] = useState('');
  const [purchaseItemId, setPurchaseItemId] = useState('');
  const [quantity, setQuantity] = useState('1');
  const [reason, setReason] = useState('');
  const [notes, setNotes] = useState('');
  const [showArchive, setShowArchive] = useState(false);
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
  const { data: selectedInventoryData } = useQuery({
    queryKey: ['supplier-return-available-stock', selectedItem?.product_id],
    queryFn: () => inventoryApi.list({ product_id: selectedItem.product_id, status: 'AVAILABLE', per_page: 100 }),
    enabled: Boolean(selectedItem?.product_id),
  });
  const availableItems = selectedInventoryData?.data?.items || selectedInventoryData?.data || [];
  const availableQuantity = availableItems.filter((item: any) => item.status === 'AVAILABLE').length;
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
      toast.success('تم رفض طلب الإرجاع وحفظه في الأرشيف');
    },
    onError: () => toast.error('لا يمكن رفض هذا الطلب في حالته الحالية'),
  });
  const purchases = (purchasesData?.data || []).filter((p: any) => !['reversed', 'cancelled'].includes(p.status));
  const returns = returnsData?.data || [];
  const pendingReturns = returns.filter((item: any) => item.status === 'PENDING').length;
  const completedReturns = returns.filter((item: any) => item.status === 'COMPLETED').length;
  const refundTotal = returns
    .filter((item: any) => item.status === 'COMPLETED')
    .reduce((total: number, item: any) => total + Number(item.refund_amount || 0), 0);
  const visibleReturns = returns.filter((item: any) => {
    const isArchived = ['COMPLETED', 'REJECTED', 'CANCELLED'].includes(item.status);
    const matchesView = showArchive ? isArchived : !isArchived;
    const query = searchQuery.trim().toLowerCase();
    const matchesSearch = !query || [item.return_number, item.reason, item.purchase_id, item.supplier_id]
      .some((value) => String(value || '').toLowerCase().includes(query));
    return matchesView && matchesSearch;
  });
  const printReturn = (item: any) => {
    const printWindow = window.open('', '_blank', 'width=800,height=700');
    if (!printWindow) {
      toast.error('تعذر فتح نافذة الطباعة');
      return;
    }
    const returnItem = item.items?.[0] || item.item;
    printWindow.document.write(`
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
    `);
    printWindow.document.close();
    printWindow.focus();
    printWindow.print();
  };

  return (
    <div className="space-y-6">
      <PageHeader eyebrow="Supplier Returns" title="مرتجعات الموردين" description="إرجاع البضاعة إلى المورد دون خلطها بمرتجعات العملاء" />
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <Card>
          <CardContent className="p-5">
            <p className="text-sm text-text-muted">إجمالي طلبات الإرجاع</p>
            <p className="mt-2 text-3xl font-bold">{returns.length}</p>
            <p className="mt-1 text-xs text-text-muted">كل الطلبات المسجلة</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-5">
            <p className="text-sm text-text-muted">قيد الانتظار</p>
            <p className="mt-2 text-3xl font-bold text-warning">{pendingReturns}</p>
            <p className="mt-1 text-xs text-text-muted">تحتاج إلى إكمال الإرجاع</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-5">
            <p className="text-sm text-text-muted">تمت الإعادة</p>
            <p className="mt-2 text-3xl font-bold text-success">{completedReturns}</p>
            <p className="mt-1 text-xs text-text-muted">طلبات مكتملة</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-5">
            <p className="text-sm text-text-muted">إجمالي المسترد من الموردين</p>
            <p className="mt-2 text-3xl font-bold">₪{refundTotal.toLocaleString('en-US', { maximumFractionDigits: 2 })}</p>
            <p className="mt-1 text-xs text-text-muted">للطلبات المكتملة فقط</p>
          </CardContent>
        </Card>
      </div>
      <Card>
        <CardHeader><CardTitle>طلب إرجاع جديد</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <label className="text-sm font-medium">1. فاتورة الشراء</label>
              <Select value={purchaseId} onChange={(e) => { setPurchaseId(e.target.value); setPurchaseItemId(''); }} options={[
                { value: '', label: 'اختر فاتورة تحتوي على مخزون متاح' },
                ...purchases.map((p: any) => ({ value: p.id, label: `${p.invoice_number} | ${p.supplier_name || 'مورد غير محدد'} | ${p.available_for_return || 0} قطعة متاحة` })),
              ]} />
              <p className="text-xs text-text-muted">تظهر هنا الفواتير التي يمكن إرجاع صنف منها فقط.</p>
            </div>
            <div className="space-y-2">
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
          </div>
          {selectedItem && (
            <div className="grid gap-3 rounded border border-border bg-surface-muted p-4 text-sm sm:grid-cols-2 lg:grid-cols-4">
              <div><span className="text-text-muted">الصنف</span><p className="font-medium">{selectedItem.product_name || selectedItem.product?.name || 'الصنف المحدد'}</p></div>
              <div><span className="text-text-muted">SKU/الباركود</span><p>{selectedItem.sku || selectedItem.barcode || '-'}</p></div>
              <div><span className="text-text-muted">كمية الفاتورة</span><p>{selectedItem.quantity || 0}</p></div>
              <div><span className="text-text-muted">تكلفة الوحدة</span><p>₪{Number(selectedItem.unit_cost || 0).toLocaleString('en-US')}</p></div>
              <div className="sm:col-span-2 lg:col-span-4">
                <span className="text-text-muted">المتاح فعلياً للإرجاع</span>
                <p className={availableQuantity > 0 ? 'font-semibold text-success' : 'font-semibold text-danger'}>{availableQuantity} قطعة</p>
              </div>
            </div>
          )}
          <div className="grid gap-4 md:grid-cols-3">
            <Input type="number" min="1" max={availableQuantity || undefined} value={quantity} onChange={(e) => setQuantity(e.target.value)} placeholder="الكمية المراد إرجاعها" />
            <Input value={reason} onChange={(e) => setReason(e.target.value)} placeholder="سبب الإرجاع (مطلوب)" />
            <Input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="ملاحظات (اختياري)" />
          </div>
          <Button className="w-full md:w-auto" disabled={!purchaseId || !purchaseItemId || !selectedItem || availableQuantity < 1 || Number(quantity) < 1 || Number(quantity) > availableQuantity || !reason || createMutation.isPending} onClick={() => createMutation.mutate()}>
            إنشاء طلب الإرجاع
          </Button>
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>طلبات إرجاع الموردين</CardTitle></CardHeader>
        <CardContent>
          <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex gap-2">
              <Button size="sm" variant={!showArchive ? 'primary' : 'secondary'} onClick={() => setShowArchive(false)}>
                الطلبات النشطة
              </Button>
              <Button size="sm" variant={showArchive ? 'primary' : 'secondary'} onClick={() => setShowArchive(true)}>
                الأرشيف
              </Button>
            </div>
            <Input value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} placeholder="بحث برقم المرتجع أو السبب" />
          </div>
          {isLoading ? <p>جار التحميل...</p> : visibleReturns.length === 0 ? <p className="text-text-muted">{showArchive ? 'لا توجد طلبات مؤرشفة' : 'لا توجد طلبات نشطة'}</p> : (
            <div className="space-y-3">{visibleReturns.map((item: any) => (
              <div key={item.id} className="flex flex-wrap justify-between gap-3 rounded border p-3">
                <span className="font-medium">{item.return_number}</span>
                <span>{item.reason}</span>
                <span>{item.status === 'PENDING' ? 'قيد الانتظار' : item.status === 'COMPLETED' ? 'مكتمل' : item.status}</span>
                <span>₪{Number(item.refund_amount || 0).toLocaleString('en-US')}</span>
                <Button size="sm" variant="secondary" title="طباعة طلب المرتجع" onClick={() => printReturn(item)}>
                  <Printer className="h-4 w-4" /> طباعة
                </Button>
                {item.status !== 'COMPLETED' && (
                  <div className="flex gap-2">
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
                        title="يرفض الطلب ويحفظه في الأرشيف بعد بدء الشحن أو الاستلام"
                        onClick={() => setConfirmAction({ type: 'reject', id: item.id })}
                      >
                        <XCircle className="h-4 w-4" /> رفض الطلب
                      </Button>
                    )}
                    <Button
                      size="sm"
                      variant="success"
                      title="يكمل الإرجاع ويخصم الكمية من المخزون ويسجل قيمته"
                      onClick={() => completeMutation.mutate(item.id)}
                      disabled={completeMutation.isPending}
                    >
                      إكمال الإرجاع
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
          : 'سيُحفظ الطلب في الأرشيف بحالة مرفوض بدلاً من حذفه، حتى يبقى تاريخه واضحاً.'}
        confirmText={confirmAction?.type === 'delete' ? 'حذف الطلب' : 'رفض الطلب'}
        variant="danger"
        isLoading={deleteMutation.isPending || rejectMutation.isPending}
      />
    </div>
  );
}
