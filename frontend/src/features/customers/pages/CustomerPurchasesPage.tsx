import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowRight, FileText, Plus, Search } from 'lucide-react';
import { PageHeader } from '../../../components/ui/page-header';
import { Button } from '../../../components/ui/button';
import { Card, CardContent } from '../../../components/ui/card';
import { Input } from '../../../components/ui/input';
import { Badge } from '../../../components/ui/badge';
import { PaginationControls } from '../../../components/ui/pagination-controls';
import { salesApi, customersApi, debtsApi } from '../../../services/api/endpoints';
import { useDebounce } from '../../../hooks/useDebounce';

const formatMoney = (value: unknown) => `₪${Number(value || 0).toLocaleString('en-US')}`;

const formatDate = (value?: string) => {
  if (!value) return 'بدون تاريخ';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString('ar-SA');
};

export function CustomerPurchasesPage() {
  const navigate = useNavigate();
  const { customerId } = useParams();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
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
    queryKey: ['customer-purchases', customerId, page, pageSize],
    queryFn: () => salesApi.list({ customer_id: customerId, page, per_page: pageSize }),
    enabled: Boolean(customerId),
  });

  const customer = customerQuery.data?.data ?? customerQuery.data;
  const payload = salesQuery.data?.data ?? salesQuery.data;
  const sales = Array.isArray(payload) ? payload : (payload?.sales ?? []);
  const ledgerPayload = ledgerQuery.data?.data ?? ledgerQuery.data;
  const ledgerEntries = Array.isArray(ledgerPayload) ? ledgerPayload : [];
  const lastLedgerEntry = ledgerEntries[ledgerEntries.length - 1];
  const debtsPayload = debtsQuery.data?.data ?? debtsQuery.data;
  const customerDebts = Array.isArray(debtsPayload) ? debtsPayload : (debtsPayload?.debts ?? []);
  const total = Number(payload?.total ?? sales.length);
  const filteredSales = debouncedSearch
    ? sales.filter((sale: any) => String(sale.invoice_number || sale.id || '').toLowerCase().includes(debouncedSearch.toLowerCase()))
    : sales;
  const salesTotal = sales.reduce((sum: number, sale: any) => sum + Number(sale.total_amount ?? 0), 0);
  const customerTotalPurchases = Number(customer?.totalPurchases ?? customer?.total_purchases ?? 0);
  const totalPurchases = customerTotalPurchases > 0 ? customerTotalPurchases : salesTotal;
  const outstanding = Number(lastLedgerEntry?.balance ?? customer?.outstanding ?? customer?.current_balance ?? 0);
  const totalPaid = Math.max(totalPurchases - outstanding, 0);

  return (
    <div>
      <PageHeader
        eyebrow="Customer Purchases"
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

      <div className="mb-5 grid gap-3 sm:grid-cols-3">
        <Card><CardContent className="p-4"><span className="text-xs text-text-muted">إجمالي المشتريات</span><p className="mt-1 text-xl font-bold text-text-primary">{formatMoney(totalPurchases)}</p></CardContent></Card>
        <Card><CardContent className="p-4"><span className="text-xs text-text-muted">إجمالي المدفوع</span><p className="mt-1 text-xl font-bold text-green-600">{formatMoney(totalPaid)}</p></CardContent></Card>
        <Card><CardContent className="p-4"><span className="text-xs text-text-muted">الرصيد المستحق</span><p className="mt-1 text-xl font-bold text-red-600">{formatMoney(outstanding)}</p></CardContent></Card>
      </div>

      <Card>
        <CardContent className="p-0">
          <div className="flex flex-col gap-3 border-b border-border p-4 md:flex-row md:items-center md:justify-between">
            <div>
              <h2 className="flex items-center gap-2 font-bold text-text-primary"><FileText className="h-5 w-5 text-primary" /> فواتير العميل</h2>
              <p className="mt-1 text-xs text-text-muted">{total} فاتورة، ويمكن فتح أي فاتورة لعرض المنتجات والتفاصيل</p>
            </div>
            <div className="relative w-full md:w-72">
              <Search className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
              <Input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="ابحث برقم الفاتورة..." className="pr-9" />
            </div>
          </div>

            {salesQuery.isLoading || customerQuery.isLoading || ledgerQuery.isLoading || debtsQuery.isLoading ? (
            <div className="p-12 text-center text-text-muted">جاري تحميل مشتريات العميل...</div>
          ) : salesQuery.isError ? (
            <div className="p-12 text-center text-red-600">تعذر تحميل مشتريات العميل</div>
          ) : filteredSales.length === 0 ? (
            <div className="p-12 text-center text-text-muted">لا توجد مشتريات مطابقة</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-[820px] w-full text-sm" dir="rtl">
                <thead className="border-b border-border bg-surface-muted text-right text-xs text-text-muted">
                  <tr>
                    <th className="px-4 py-3">الفاتورة</th>
                    <th className="px-4 py-3">التاريخ</th>
                    <th className="px-4 py-3">الإجمالي</th>
                    <th className="px-4 py-3">المدفوع</th>
                    <th className="px-4 py-3">المتبقي</th>
                    <th className="px-4 py-3">الحالة</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {filteredSales.map((sale: any) => {
                    const totalAmount = Number(sale.total_amount ?? 0);
                    const debt = customerDebts.find((item: any) => String(item.sale_id || item.reference_id || '') === String(sale.id));
                    const remaining = Math.max(Number(debt?.remaining_amount ?? sale.remaining_amount ?? (totalAmount - Number(sale.paid_amount ?? 0))), 0);
                    const paidAmount = Math.max(totalAmount - remaining, 0);
                    return (
                      <tr key={sale.id} className="hover:bg-surface-muted">
                        <td className="px-4 py-4 font-semibold text-text-primary">{sale.invoice_number || sale.id}</td>
                        <td className="px-4 py-4 text-text-secondary">{formatDate(sale.sale_date || sale.created_at)}</td>
                        <td className="px-4 py-4 font-semibold">{formatMoney(totalAmount)}</td>
                        <td className="px-4 py-4 text-green-600">{formatMoney(paidAmount)}</td>
                        <td className="px-4 py-4 font-semibold text-red-600">{formatMoney(remaining)}</td>
                        <td className="px-4 py-4"><Badge variant={remaining > 0 ? 'danger' : 'success'}>{remaining > 0 ? 'متبقي عليها مبلغ' : 'مدفوعة بالكامل'}</Badge></td>
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
    </div>
  );
}
