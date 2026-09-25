import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { AlertTriangle, ArrowRight, Calendar, CalendarDays, CircleDollarSign, DollarSign, Eye, FileText, Plus, ReceiptText, Search } from 'lucide-react';
import { PageHeader } from '../../../design-system/components/page-header';
import { Button } from '../../../design-system/components/button';
import { Card, CardContent } from '../../../design-system/components/card';
import { StatCard } from '../../../design-system/components/stat-card';
import { Input } from '../../../design-system/components/input';
import { Badge } from '../../../design-system/components/badge';
import { Modal } from '../../../design-system/components/modal';
import { PaginationControls } from '../../../design-system/components/pagination-controls';
import { salesApi, customersApi, debtsApi } from '../../../services/api/endpoints';
import { useDebounce } from '../../../hooks/useDebounce';
import { SalesInvoice } from '../../../components/invoice/SalesInvoice';
import { formatStoreDate } from '../../../utils/store-time';

const formatMoney = (value: unknown) => `₪${Number(value || 0).toLocaleString('en-US')}`;

const formatDate = (value?: string) => {
  if (!value) return 'بدون تاريخ';
  const formatted = formatStoreDate(value, 'ar-SA');
  return formatted === 'غير محدد' ? value : formatted;
};

export function CustomerPurchasesPage() {
  const navigate = useNavigate();
  const { customerId } = useParams();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [selectedSaleId, setSelectedSaleId] = useState<string | null>(null);
  const pageSize = 20;
  const debouncedSearch = useDebounce(search, 250);

  const customerQuery = useQuery({
    queryKey: ['customer', customerId],
    queryFn: () => customersApi.get(customerId || ''),
    enabled: Boolean(customerId),
  });

  const ledgerQuery = useQuery({
    queryKey: ['customer-purchases-ledger', customerId],
    queryFn: () => customersApi.ledger(customerId || '', { page: 1, per_page: 100 }),
    enabled: Boolean(customerId),
  });

  const debtsQuery = useQuery({
    queryKey: ['customer-purchases-debts', customerId],
    queryFn: () => debtsApi.getDebtEntries(customerId || ''),
    enabled: Boolean(customerId),
  });

  const salesQuery = useQuery({
    queryKey: ['customer-purchases', customerId, page, pageSize, debouncedSearch],
    queryFn: () => salesApi.list({
      customer_id: customerId,
      page,
      per_page: pageSize,
      search: debouncedSearch.trim() || undefined,
    }),
    enabled: Boolean(customerId),
  });

  const saleDetailsQuery = useQuery({
    queryKey: ['customer-purchase-invoice', selectedSaleId],
    queryFn: () => salesApi.get(selectedSaleId || ''),
    enabled: Boolean(selectedSaleId),
  });

  const customer = customerQuery.data?.data ?? customerQuery.data;
  const saleDetailsPayload = saleDetailsQuery.data?.data ?? saleDetailsQuery.data;
  const saleDetails = saleDetailsPayload?.sale ?? saleDetailsPayload;
  const saleItems = Array.isArray(saleDetailsPayload?.items)
    ? saleDetailsPayload.items
    : Array.isArray(saleDetails?.items)
      ? saleDetails.items
      : [];
  const invoiceData = saleDetails ? {
    id: saleDetails.id,
    invoiceNumber: saleDetails.invoice_number,
    customerName: customer?.name || '',
    customerPhone: customer?.phone,
    saleDate: saleDetails.sale_date || saleDetails.created_at,
    items: saleItems.map((item: any) => ({
      name: item.product_name || item.productName || item.product?.name || item.name || 'منتج',
      sku: item.sku,
      barcode: item.barcode,
      condition: item.condition || 'NEW',
      sellingPrice: Number(item.unit_price || 0),
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
    paymentMethod: saleDetails.payment_method === 'debt' ? 'credit' : saleDetails.payment_method || 'cash',
    paymentStatus: saleDetails.payment_status,
    cashReceived: Number(saleDetails.cash_received || 0),
    changeAmount: Number(saleDetails.change_amount || 0),
    notes: saleDetails.notes,
  } : null;
  const payload = salesQuery.data?.data ?? salesQuery.data;
  const sales = Array.isArray(payload) ? payload : (payload?.sales ?? []);
  const ledgerPayload = ledgerQuery.data?.data ?? ledgerQuery.data;
  const ledgerEntries = Array.isArray(ledgerPayload) ? ledgerPayload : [];
  const lastLedgerEntry = ledgerEntries[ledgerEntries.length - 1];
  const debtsPayload = debtsQuery.data?.data ?? debtsQuery.data;
  const customerDebts = Array.isArray(debtsPayload) ? debtsPayload : (debtsPayload?.debts ?? []);
  const total = Number(payload?.total ?? sales.length);
  const salesTotal = sales.reduce((sum: number, sale: any) => sum + Number(sale.total_amount ?? 0), 0);
  const customerTotalPurchases = Number(customer?.totalPurchases ?? customer?.total_purchases ?? 0);
  const totalPurchases = customerTotalPurchases > 0 ? customerTotalPurchases : salesTotal;
  const outstanding = Number(lastLedgerEntry?.balance ?? customer?.outstanding ?? customer?.current_balance ?? 0);
  const totalPaid = Math.max(totalPurchases - outstanding, 0);
  const pageError = customerQuery.isError || ledgerQuery.isError || debtsQuery.isError || salesQuery.isError;
  const retryPageData = () => {
    void Promise.all([
      customerQuery.refetch(),
      ledgerQuery.refetch(),
      debtsQuery.refetch(),
      salesQuery.refetch(),
    ]);
  };

  if (!customerId || pageError) {
    return (
      <div>
        <PageHeader title="مشتريات العميل" description="سجل الفواتير والمدفوعات والرصيد المستحق" />
        <Card>
          <CardContent className="flex flex-col items-center gap-3 p-8 text-center">
            <AlertTriangle className="h-8 w-8 text-red-600" />
            <p className="font-semibold text-text-primary">
              {!customerId ? 'رابط العميل غير مكتمل' : 'تعذر تحميل بيانات مشتريات العميل'}
            </p>
            <p className="text-sm text-text-muted">تحقق من الاتصال ثم أعد المحاولة لتحميل بيانات العميل والفواتير والرصيد.</p>
            <Button type="button" variant="secondary" onClick={retryPageData} disabled={!customerId}>
              إعادة المحاولة
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        title={`مشتريات ${customer?.name || 'العميل'}`}
        description="سجل مستقل لجميع فواتير العميل والمدفوع والمتبقي لكل عملية"
        actions={
          <div className="flex flex-wrap gap-2">
            <Button variant="secondary" onClick={() => navigate('/app/customers')}>
              <ArrowRight className="h-4 w-4" />
              رجوع للعملاء
            </Button>
            <Button variant="primary" onClick={() => navigate('/app/sales')}>
              <Plus className="h-4 w-4" />
              بيع جديد
            </Button>
          </div>
        }
      />

      <div
        className="mb-5 grid-cols-1 sm:grid-cols-3"
        style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(190px, 1fr))', gap: '10px' }}
      >
        <StatCard
          title="إجمالي المشتريات"
          value={<span className="numeric-metric">{formatMoney(totalPurchases)}</span>}
          icon={DollarSign}
          subtitle="قيمة مشتريات العميل"
          variant="featured"
          compact
        />
        <StatCard
          title="إجمالي المدفوع"
          value={<span className="numeric-metric">{formatMoney(totalPaid)}</span>}
          icon={Calendar}
          subtitle="ما تم تحصيله من المشتريات"
          variant="success"
          compact
        />
        <StatCard
          title="الرصيد المستحق"
          value={<span className="numeric-metric">{formatMoney(outstanding)}</span>}
          icon={AlertTriangle}
          subtitle="المبلغ المتبقي على العميل"
          variant="warning"
          compact
        />
      </div>

      <Card className="overflow-hidden">
        <CardContent className="p-0">
          <div className="flex flex-col gap-5 border-b border-border bg-surface-elevated px-5 py-5 md:flex-row md:items-end md:justify-between">
            <div className="min-w-0">
              <div className="mb-2 flex items-center gap-2 text-xs font-semibold text-primary">
                <ReceiptText className="h-4 w-4" />
                <span>سجل الفواتير</span>
              </div>
              <div className="flex flex-wrap items-center gap-3">
                <h2 className="flex items-center gap-2 text-lg font-bold text-text-primary"><FileText className="h-5 w-5 text-primary" /> فواتير العميل</h2>
                <span className="rounded-full border border-primary/20 bg-primary/8 px-2.5 py-1 text-xs font-semibold text-primary">{total} فاتورة</span>
              </div>
              <p className="mt-2 max-w-[36rem] text-sm text-text-muted">افتح أي فاتورة لمراجعة المنتجات والمدفوعات والتفاصيل.</p>
            </div>
            <div className="relative w-full md:w-80 md:max-w-[40%]">
              <Search className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
              <Input value={search} onChange={(event) => { setSearch(event.target.value); setPage(1); }} placeholder="ابحث برقم الفاتورة..." className="h-10 bg-surface pr-9" />
            </div>
          </div>

            {salesQuery.isLoading || customerQuery.isLoading || ledgerQuery.isLoading || debtsQuery.isLoading ? (
            <div className="p-12 text-center text-text-muted">جاري تحميل مشتريات العميل...</div>
          ) : sales.length === 0 ? (
            <div className="p-12 text-center text-text-muted">لا توجد مشتريات مطابقة</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-[760px] w-full text-sm" dir="rtl">
                <thead className="border-b border-border bg-surface-muted text-right text-xs font-semibold text-text-muted">
                  <tr>
                    <th className="px-5 py-3.5">الفاتورة</th>
                    <th className="px-5 py-3.5"><span className="inline-flex items-center gap-1.5"><CalendarDays className="h-3.5 w-3.5" />التاريخ</span></th>
                    <th className="px-5 py-3.5"><span className="inline-flex items-center gap-1.5"><CircleDollarSign className="h-3.5 w-3.5" />الإجمالي</span></th>
                    <th className="px-5 py-3.5">المدفوع</th>
                    <th className="px-5 py-3.5">المتبقي</th>
                    <th className="px-5 py-3.5">الحالة</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {sales.map((sale: any) => {
                    const totalAmount = Number(sale.total_amount ?? 0);
                    const debt = customerDebts.find((item: any) => String(item.sale_id || item.reference_id || '') === String(sale.id));
                    const remaining = Math.max(Number(debt?.remaining_amount ?? sale.remaining_amount ?? (totalAmount - Number(sale.paid_amount ?? 0))), 0);
                    const paidAmount = Math.max(totalAmount - remaining, 0);
                    return (
                      <tr key={sale.id} className="group transition-colors hover:bg-surface-muted/70">
                        <td className="px-5 py-4 font-semibold text-text-primary">
                          <div className="flex min-w-0 items-center gap-3">
                            <span className="min-w-0 truncate font-mono text-xs font-semibold tracking-wide text-text-primary">{sale.invoice_number || sale.id}</span>
                            <Button type="button" variant="ghost" size="icon" tableAction className="text-primary" onClick={() => setSelectedSaleId(String(sale.id))} aria-label={`عرض الفاتورة ${sale.invoice_number || sale.id}`} title="عرض الفاتورة">
                              <Eye className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        </td>
                        <td className="whitespace-nowrap px-5 py-4 text-text-secondary">{formatDate(sale.sale_date || sale.created_at)}</td>
                        <td className="whitespace-nowrap px-5 py-4 font-bold text-text-primary">{formatMoney(totalAmount)}</td>
                        <td className="whitespace-nowrap px-5 py-4 font-medium text-green-600">{formatMoney(paidAmount)}</td>
                        <td className="whitespace-nowrap px-5 py-4 font-semibold text-red-600">{formatMoney(remaining)}</td>
                        <td className="px-5 py-4"><Badge variant={remaining > 0 ? 'danger' : 'success'}>{remaining > 0 ? 'متبقي عليها مبلغ' : 'مدفوعة بالكامل'}</Badge></td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
              <PaginationControls page={page} pageSize={pageSize} total={total} onPageChange={setPage} isLoading={salesQuery.isFetching} />
            </div>
          )}
        </CardContent>
      </Card>

      <Modal isOpen={Boolean(selectedSaleId)} onClose={() => setSelectedSaleId(null)} title="فاتورة البيع" variant="modern" size="xl">
        {saleDetailsQuery.isLoading ? (
          <div className="p-10 text-center text-text-muted">جاري تحميل الفاتورة...</div>
        ) : saleDetailsQuery.isError || !invoiceData ? (
          <div className="flex flex-col items-center gap-3 p-10 text-center text-red-600">
            <p>تعذر تحميل تفاصيل الفاتورة</p>
            <Button type="button" variant="secondary" onClick={() => { void saleDetailsQuery.refetch(); }}>
              إعادة المحاولة
            </Button>
          </div>
        ) : (
          <SalesInvoice saleData={invoiceData} onClose={() => setSelectedSaleId(null)} />
        )}
      </Modal>
    </div>
  );
}
