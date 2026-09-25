import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Activity, ArrowLeft, ArrowRight, CalendarDays, Clock, DollarSign, Eye, RefreshCw, ReceiptText, RotateCcw, Search, ShoppingCart, Undo2, UserRound } from 'lucide-react';
import { dashboardApi, purchasesApi, returnsApi, salesApi } from '../../../services/api/endpoints';
import { PageHeader } from '../../../design-system/components/page-header';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { Badge } from '../../../design-system/components/badge';
import { Input } from '../../../design-system/components/input';
import { Modal } from '../../../design-system/components/modal';
import { SalesInvoice } from '../../../components/invoice/SalesInvoice';
import { SupplierInvoiceModal } from '../../purchases/components/SupplierInvoiceModal';
import { ReturnDetailsPage } from '../../returns/pages/ReturnDetailsPage';
import { formatStoreActivityDateTime } from '../../../utils/store-time';

const PAGE_SIZE = 10;

const statusLabels: Record<string, string> = {
  completed: 'مكتمل',
  pending: 'قيد الانتظار',
  reversed: 'إلغاء الشراء',
  cancelled: 'ملغي',
  received: 'مستلم',
  ordered: 'تم الطلب',
  draft: 'مسودة',
  approved: 'معتمد',
  rejected: 'مرفوض',
};

const paymentLabels: Record<string, string> = {
  cash: 'نقدًا',
  card: 'بطاقة',
  debt: 'دين',
  credit: 'دين',
  transfer: 'تحويل بنكي',
  bank_transfer: 'تحويل بنكي',
  checks: 'شيك',
};

function getStatusLabel(status: string) {
  return statusLabels[String(status || '').toLowerCase()] || 'قيد المعالجة';
}

function getActivityIcon(type: string) {
  switch (type) {
    case 'sale': return ShoppingCart;
    case 'purchase': return DollarSign;
    case 'return': return RotateCcw;
    default: return Activity;
  }
}

export function ActivityPage() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [activityType, setActivityType] = useState('');
  const [search, setSearch] = useState('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [pageSize, setPageSize] = useState(PAGE_SIZE);
  const [selectedItem, setSelectedItem] = useState<any | null>(null);
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['dashboard-activity', page, pageSize, activityType, search, startDate, endDate],
    queryFn: () => dashboardApi.getActivity({
      page,
      per_page: pageSize,
      type: activityType || undefined,
      search: search.trim() || undefined,
      start_date: startDate || undefined,
      end_date: endDate || undefined,
    }),
    staleTime: 30000,
  });

  const detailQuery = useQuery({
    queryKey: ['activity-detail', selectedItem?.type, selectedItem?.id],
    queryFn: () => {
      if (selectedItem?.type === 'sale') return salesApi.get(selectedItem.id);
      if (selectedItem?.type === 'purchase') return purchasesApi.get(selectedItem.id);
      return returnsApi.getWithItems(selectedItem.id);
    },
    enabled: Boolean(selectedItem),
  });

  const activityPage = data?.data;
  const items = activityPage?.items || [];
  const totalPages = activityPage?.total_pages || 0;

  const filteredItems = items;

  const changeType = (value: string) => {
    setActivityType(value);
    setPage(1);
  };

  const closeDetails = () => setSelectedItem(null);
  const detailPayload = detailQuery.data?.data ?? detailQuery.data;
  const saleDetails = detailPayload?.sale ?? detailPayload;
  const saleItems = Array.isArray(detailPayload?.items) ? detailPayload.items : Array.isArray(saleDetails?.items) ? saleDetails.items : [];
  const purchaseDetails = detailPayload?.purchase ?? detailPayload;
  const purchaseItems = Array.isArray(detailPayload?.items) ? detailPayload.items : [];
  const returnDetails = detailPayload?.return ?? detailPayload;
  const saleInvoice = saleDetails ? {
    id: saleDetails.id,
    invoiceNumber: saleDetails.invoice_number,
    customerName: saleDetails.customer_name || saleDetails.customer?.name || 'عميل عام',
    customerPhone: saleDetails.customer_phone || saleDetails.customer?.phone,
    saleDate: saleDetails.sale_date || saleDetails.created_at,
    items: saleItems.map((item: any) => ({
      name: item.product_name || item.product?.name || item.name || 'منتج',
      sku: item.sku,
      barcode: item.barcode,
      sellingPrice: Number(item.unit_price || item.price || 0),
      quantity: Number(item.quantity || 0),
      total: Number(item.total_amount || 0),
      discountAmount: Number(item.discount_amount || 0),
      taxAmount: Number(item.tax_amount || 0),
    })),
    subtotal: Number(saleDetails.subtotal || 0),
    discountAmount: Number(saleDetails.discount_amount || 0),
    taxAmount: Number(saleDetails.tax_amount || 0),
    total: Number(saleDetails.total_amount || 0),
    paidAmount: Number(saleDetails.paid_amount || 0),
    remaining: Math.max(Number(saleDetails.total_amount || 0) - Number(saleDetails.paid_amount || 0), 0),
    paymentMethod: saleDetails.payment_method || 'cash',
    paymentStatus: saleDetails.payment_status,
    cashReceived: Number(saleDetails.cash_received || 0),
    changeAmount: Number(saleDetails.change_amount || 0),
    notes: saleDetails.notes,
  } : null;

  return (
    <div>
      <PageHeader
        title="كل النشاط"
        description="سجل المبيعات والمشتريات والمرتجعات مرتبًا من الأحدث إلى الأقدم"
        actions={(
              <div className="flex flex-wrap gap-2">
                <Button type="button" variant="secondary" onClick={() => { void refetch(); }} className="gap-2">
                  <RefreshCw className="w-4 h-4" /> تحديث
                </Button>
                <Button type="button" variant="secondary" onClick={() => navigate(-1)} className="gap-2">
                  <ArrowRight className="w-4 h-4" /> رجوع
                </Button>
              </div>
        )}
      />

      <Card variant="open" style={{ marginTop: 'var(--spacing-6)' }}>
        <CardHeader>
          <div className="flex items-center justify-between gap-3 flex-wrap">
            <CardTitle className="flex items-center gap-2">
              <Clock className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
              سجل العمليات
            </CardTitle>
            <div className="flex w-full flex-wrap items-center gap-2 md:w-auto">
              <div className="relative min-w-[220px] flex-1 md:flex-none">
                <Search className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
                <Input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="ابحث في النشاط..." className="h-10 pr-9" />
              </div>
              <select value={activityType} onChange={(event) => changeType(event.target.value)} aria-label="تصفية نوع النشاط" className="h-10 rounded-md border border-border bg-surface px-3 text-text-primary">
                <option value="">كل العمليات</option><option value="sale">المبيعات</option><option value="purchase">المشتريات</option><option value="return">المرتجعات</option>
              </select>
              <select value={pageSize} onChange={(event) => { setPageSize(Number(event.target.value)); setPage(1); }} aria-label="عدد العناصر في الصفحة" className="h-10 rounded-md border border-border bg-surface px-3 text-text-primary">
                <option value={10}>10 / صفحة</option><option value={20}>20 / صفحة</option><option value={50}>50 / صفحة</option>
              </select>
            </div>
            <div className="mt-3 flex flex-wrap items-center gap-2">
              <CalendarDays className="h-4 w-4 text-text-muted" />
              <label className="text-xs text-text-muted">من <input type="date" value={startDate} onChange={(event) => setStartDate(event.target.value)} className="mr-1 h-9 rounded-md border border-border bg-surface px-2 text-text-primary" /></label>
              <label className="text-xs text-text-muted">إلى <input type="date" value={endDate} onChange={(event) => setEndDate(event.target.value)} className="mr-1 h-9 rounded-md border border-border bg-surface px-2 text-text-primary" /></label>
              {(search || startDate || endDate) && <Button type="button" variant="ghost" size="sm" onClick={() => { setSearch(''); setStartDate(''); setEndDate(''); }}>مسح الفلاتر</Button>}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-center text-text-muted py-8">جارِ تحميل النشاط...</p>
          ) : error ? (
            <p className="text-center text-text-muted py-8">تعذر تحميل سجل النشاط</p>
          ) : filteredItems.length === 0 ? (
            <p className="text-center text-text-muted py-8">لا توجد عمليات مطابقة</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
              {filteredItems.map((item: any) => {
                const Icon = getActivityIcon(item.type);
                const normalizedStatus = String(item.status || '').toLowerCase();
                const activityDescription = item.type === 'sale' && item.seller_name
                  ? `بواسطة: ${item.seller_name}`
                  : item.description;
                const isSale = item.type === 'sale';
                const paymentLabel = paymentLabels[String(item.payment_method || '').toLowerCase()] || item.payment_method || 'غير محدد';
                return (
                  <div
                    key={`${item.type}-${item.id}`}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 'var(--spacing-3)',
                      padding: 'var(--spacing-4)',
                      borderRadius: 'var(--radius-md)',
                      border: '1px solid var(--border-default)',
                    }}
                  >
                    <div className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'var(--color-primary-10)' }}>
                      <Icon className="w-5 h-5" style={{ color: 'var(--color-primary)' }} />
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div className="flex flex-wrap items-center gap-2">
                        <p style={{ fontWeight: 'var(--font-weight-medium)', color: 'var(--text-primary)' }}>{item.title}</p>
                        {isSale && item.invoice_number && <span className="text-xs text-text-muted">فاتورة {item.invoice_number}</span>}
                      </div>
                      <p style={{ fontSize: 'var(--font-size-caption)', color: 'var(--text-secondary)' }}>{activityDescription}</p>
                      {isSale && (
                        <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-text-muted">
                          <span>العميل: {item.customer_name || 'عميل عام'}</span>
                          <span>الدفع: {paymentLabel}</span>
                          <span>المدفوع: ₪{Number(item.paid_amount || 0).toLocaleString('en-US')}</span>
                          <span>المتبقي: ₪{Number(item.remaining_amount || 0).toLocaleString('en-US')}</span>
                        </div>
                      )}
                    </div>
                    <div style={{ textAlign: 'right' }}>
                      <p style={{ fontWeight: 'var(--font-weight-medium)', color: 'var(--text-primary)' }}>₪{Number(item.amount || 0).toLocaleString('en-US')}</p>
                      <p style={{ fontSize: 'var(--font-size-caption)', color: 'var(--text-secondary)' }}>{formatStoreActivityDateTime(item.time, item.type === 'sale' ? item.sale_date : undefined, 'ar')}</p>
                    </div>
                    <Badge variant={normalizedStatus === 'completed' ? 'success' : 'warning'} size="sm">{getStatusLabel(item.status)}</Badge>
                    <Button type="button" variant="ghost" size="icon" onClick={() => setSelectedItem(item)} aria-label="عرض تفاصيل العملية" title="عرض التفاصيل والفاتورة">
                      {isSale ? <ReceiptText className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </Button>
                    {isSale && item.customer_id && <Button type="button" variant="ghost" size="icon" onClick={() => navigate(`/app/customers/${item.customer_id}/purchases`)} aria-label="فتح ملف العميل" title="مشتريات العميل"><UserRound className="h-4 w-4" /></Button>}
                    {isSale && <Button type="button" variant="ghost" size="icon" onClick={() => navigate(`/app/returns/create?sale_id=${encodeURIComponent(item.id)}`)} aria-label="إنشاء مرتجع من البيع" title="إنشاء مرتجع من هذا البيع"><Undo2 className="h-4 w-4" /></Button>}
                    <Button type="button" variant="ghost" size="sm" onClick={() => navigate(`/app/${item.type === 'sale' ? 'sales' : item.type === 'purchase' ? 'purchases' : 'returns'}`)}>فتح القسم</Button>
                  </div>
                );
              })}
            </div>
          )}

          {totalPages > 1 && (
            <div className="flex items-center justify-between gap-3" style={{ marginTop: 'var(--spacing-5)' }}>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage((current) => Math.max(1, current - 1))}
                className="gap-1"
              >
                <ArrowRight className="w-4 h-4" />
                السابق
              </Button>
              <span style={{ fontSize: 'var(--font-size-caption)', color: 'var(--text-secondary)' }}>
                صفحة {page} من {totalPages}
              </span>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
                className="gap-1"
              >
                التالي
                <ArrowLeft className="w-4 h-4" />
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {selectedItem?.type === 'purchase' ? (
        <SupplierInvoiceModal isOpen onClose={closeDetails} purchase={purchaseDetails} supplier={purchaseDetails?.supplier} items={purchaseItems} />
      ) : (
        <Modal isOpen={Boolean(selectedItem)} onClose={closeDetails} title={selectedItem ? `تفاصيل ${selectedItem.type === 'sale' ? 'المبيعات' : 'المرتجع'}` : 'التفاصيل'} size="2xl">
          {detailQuery.isLoading ? <p className="p-8 text-center text-text-muted">جاري تحميل التفاصيل...</p> : detailQuery.isError ? <p className="p-8 text-center text-text-muted">تعذر تحميل تفاصيل العملية</p> : selectedItem?.type === 'sale' && saleInvoice ? <SalesInvoice saleData={saleInvoice} onClose={closeDetails} /> : selectedItem?.type === 'return' ? <ReturnDetailsPage returnId={returnDetails?.id || selectedItem.id} embedded onClose={closeDetails} /> : null}
        </Modal>
      )}
    </div>
  );
}
