import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { AlertCircle, Archive, CheckCircle2, ChevronLeft, ChevronRight, ClipboardList, Clock3, CreditCard, History, Package, PackageOpen, ReceiptText, Search, ShoppingCart, Truck, UserRound, X } from 'lucide-react';
import { PageHeader } from '../../../components/ui/page-header';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Badge } from '../../../components/ui/badge';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { auditApi, customersApi, inventoryApi, purchasesApi, salesApi } from '../../../services/api/endpoints';

const formatDate = (value?: string) => {
  if (!value) return '-';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString('ar-SA');
};

const formatMoney = (value?: number) => `₪${Number(value || 0).toLocaleString('en-US')}`;

const getPurchaseStatus = (status?: string) => {
  switch (String(status || '').toLowerCase()) {
    case 'reversed':
      return 'معكوسة';
    case 'cancelled':
    case 'canceled':
      return 'ملغاة';
    default:
      return status || 'غير محدد';
  }
};

const translateAuditAction = (action?: string) => {
  const labels: Record<string, string> = {
    CREATE_SALE: 'إنشاء عملية بيع',
    CREATE_PURCHASE: 'إنشاء عملية شراء',
    ADD_PAYMENT: 'تسجيل دفعة',
    REVERSE_PURCHASE: 'عكس عملية شراء',
    CREATE_INVENTORY: 'إضافة عنصر للمخزون',
    UPDATE_INVENTORY: 'تعديل عنصر المخزون',
    ARCHIVE_INVENTORY: 'أرشفة عنصر المخزون',
    DELETE_INVENTORY: 'حذف عنصر المخزون',
    DELETE: 'حذف',
    UPDATE: 'تعديل',
    CREATE: 'إنشاء',
    ARCHIVE: 'أرشفة',
    RESTORE: 'استعادة',
  };
  return labels[String(action || '').toUpperCase()] || action || 'عملية مسجلة';
};

const translateEntityType = (entityType?: string) => {
  const labels: Record<string, string> = {
    sale: 'بيع',
    purchase: 'شراء',
    inventory: 'مخزون',
    product: 'منتج',
    payment: 'دفعة',
    customer: 'عميل',
    supplier: 'مورد',
  };
  return labels[String(entityType || '').toLowerCase()] || entityType || 'سجل';
};

const translateMovementType = (movementType?: string) => {
  const labels: Record<string, string> = {
    PURCHASE: 'شراء',
    SALE: 'بيع',
    RETURN: 'مرتجع',
    SUPPLIER_RETURN: 'مرتجع مورد',
    ADJUSTMENT: 'تعديل مخزون',
    TRANSFER: 'نقل مخزون',
    DAMAGE: 'تالف',
    REPAIR: 'إصلاح',
    RESERVATION: 'حجز',
    RELEASE: 'إلغاء حجز',
    REVERSE_PURCHASE: 'عكس عملية شراء',
    REVERSE_SALE: 'عكس عملية بيع',
    REVERSE_RETURN: 'عكس مرتجع',
  };
  return labels[String(movementType || '').toUpperCase()] || movementType || 'حركة';
};

export function ArchivePage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [section, setSection] = useState<'all' | 'inventory' | 'purchases' | 'audit'>('all');
  const [selectedItem, setSelectedItem] = useState<any | null>(null);
  const [auditPage, setAuditPage] = useState(1);
  const [auditEntityFilter, setAuditEntityFilter] = useState<'all' | 'sale' | 'purchase' | 'inventory'>('all');
  const auditPerPage = 10;

  const archivedInventoryQuery = useQuery({
    queryKey: ['archive', 'inventory'],
    queryFn: () => inventoryApi.listArchived({ page: 1, per_page: 100 }),
  });
  const purchasesQuery = useQuery({
    queryKey: ['archive', 'purchases'],
    queryFn: () => purchasesApi.list({ page: 1, per_page: 100 }),
  });
  const salesQuery = useQuery({
    queryKey: ['archive', 'sales'],
    queryFn: () => salesApi.list({ page: 1, per_page: 100 }),
  });
  const customersQuery = useQuery({
    queryKey: ['archive', 'customers'],
    queryFn: () => customersApi.list({ page: 1, per_page: 100 }),
  });
  const auditQuery = useQuery({
    queryKey: ['archive', 'audit', auditPage],
    queryFn: () => auditApi.list({ page: auditPage, per_page: auditPerPage }),
  });
  const historyQuery = useQuery({
    queryKey: ['archive', 'inventory-history', selectedItem?.id],
    queryFn: () => inventoryApi.movements(selectedItem.id),
    enabled: Boolean(selectedItem?.id),
  });

  const archivedInventory = (archivedInventoryQuery.data?.data?.items || []) as any[];
  const purchasesPayload = purchasesQuery.data?.data;
  const purchases = (Array.isArray(purchasesPayload) ? purchasesPayload : purchasesPayload?.purchases || []) as any[];
  const archivedPurchases = purchases.filter((purchase) =>
    ['cancelled', 'canceled', 'reversed'].includes(String(purchase.status || '').toLowerCase())
  );
  const salesPayload = salesQuery.data?.data;
  const sales = (Array.isArray(salesPayload) ? salesPayload : salesPayload?.sales || []) as any[];
  const customersPayload = customersQuery.data?.data;
  const customers = (Array.isArray(customersPayload) ? customersPayload : customersPayload?.customers || []) as any[];
  const auditLogs = (auditQuery.data?.data || []) as any[];
  const auditTotal = Number(auditQuery.data?.meta?.total || 0);
  const auditTotalPages = Math.max(1, Math.ceil(auditTotal / auditPerPage));

  const filteredInventory = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    if (!query) return archivedInventory;
    return archivedInventory.filter((item) => [item.product_name, item.item_code, item.barcode, item.serial_number, item.supplier_name]
      .some((value) => String(value || '').toLowerCase().includes(query)));
  }, [archivedInventory, searchQuery]);

  const filteredPurchases = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    if (!query) return archivedPurchases;
    return archivedPurchases.filter((purchase) => [purchase.invoice_number, purchase.supplier_name, purchase.notes]
      .some((value) => String(value || '').toLowerCase().includes(query)));
  }, [archivedPurchases, searchQuery]);

  const auditRows = useMemo(() => auditLogs.map((log) => {
    const entityType = String(log.entity_type || '').toLowerCase();
    const entityId = String(log.entity_id || '');
    const sale = entityType === 'sale' ? sales.find((item) => String(item.id) === entityId) : undefined;
    const purchase = entityType === 'purchase' ? purchases.find((item: any) => String(item.id) === entityId) : undefined;
    const saleCustomerId = sale?.customer_id || sale?.customerId || sale?.customer?.id;
    const saleCustomer = sale?.customer?.name
      ? sale.customer.name
      : sale?.customer_name || customers.find((customer) => String(customer.id) === String(saleCustomerId))?.name;
    const reference = sale
      ? `الفاتورة: ${sale.invoice_number || sale.sale_number || entityId}`
      : purchase
        ? `الفاتورة: ${purchase.invoice_number || entityId}`
        : `المعرّف: ${entityId || '-'}`;
    const party = sale
      ? `العميل: ${saleCustomer || (saleCustomerId ? 'اسم العميل غير متاح' : 'بيع نقدي')}`
      : purchase
        ? `المورد: ${purchase.supplier?.name || purchase.supplier_name || 'غير محدد'}`
        : '';
    const amount = sale
      ? `الإجمالي: ${formatMoney(sale.total_amount)}`
      : purchase
        ? `الإجمالي: ${formatMoney(purchase.total_amount)} · المدفوع: ${formatMoney(purchase.paid_amount)}`
        : '';
    return {
      log,
      label: translateAuditAction(log.action),
      entityType,
      details: [translateEntityType(log.entity_type), reference, party, amount].filter(Boolean).join(' · '),
      searchText: [log.action, log.entity_type, log.description, log.user_name, reference, party, amount].join(' ').toLowerCase(),
    };
  }), [auditLogs, customers, purchases, sales]);

  const filteredAudit = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    if (!query) return auditRows;
    return auditRows.filter((row) => row.searchText.includes(query));
  }, [auditRows, searchQuery]);

  const visibleAudit = auditEntityFilter === 'all'
    ? filteredAudit
    : filteredAudit.filter((row) => row.entityType === auditEntityFilter);
  const successfulAuditCount = visibleAudit.filter(({ log }) => log.status !== 'failure').length;
  const failedAuditCount = visibleAudit.filter(({ log }) => log.status === 'failure').length;
  const getAuditIcon = (entityType: string) => {
    if (entityType === 'sale') return <ShoppingCart className="h-5 w-5" />;
    if (entityType === 'purchase') return <Truck className="h-5 w-5" />;
    if (entityType === 'inventory') return <PackageOpen className="h-5 w-5" />;
    if (entityType === 'payment') return <CreditCard className="h-5 w-5" />;
    return <ClipboardList className="h-5 w-5" />;
  };

  const show = (value: typeof section) => section === 'all' || section === value;
  const isLoading = archivedInventoryQuery.isLoading || purchasesQuery.isLoading || salesQuery.isLoading || customersQuery.isLoading;
  const totalArchivedValue = archivedInventory.reduce((sum, item) => sum + Number(item.selling_price || 0), 0);

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="الأرشيف والتاريخ"
        title="الأرشيف والسجل التاريخي"
        description="مرجع دقيق للعناصر والعمليات التي خرجت من التشغيل دون حذف تاريخها التجاري."
      />

      <div className="grid gap-4 md:grid-cols-3">
        <Card><CardContent className="flex items-center justify-between p-5"><div><p className="text-xs text-text-muted">عناصر المخزون المؤرشفة</p><p className="mt-1 text-2xl font-bold">{archivedInventory.length}</p></div><Archive className="h-6 w-6 text-text-muted" /></CardContent></Card>
        <Card><CardContent className="flex items-center justify-between p-5"><div><p className="text-xs text-text-muted">العمليات الملغاة أو المعكوسة</p><p className="mt-1 text-2xl font-bold">{archivedPurchases.length}</p></div><ClipboardList className="h-6 w-6 text-text-muted" /></CardContent></Card>
        <Card><CardContent className="flex items-center justify-between p-5"><div><p className="text-xs text-text-muted">قيمة البيع التاريخية للعناصر المؤرشفة</p><p className="mt-1 text-2xl font-bold">{formatMoney(totalArchivedValue)}</p></div><Package className="h-6 w-6 text-text-muted" /></CardContent></Card>
      </div>

      <Card>
        <CardContent className="flex flex-col gap-3 p-4 md:flex-row md:items-center md:justify-between">
          <div className="flex flex-wrap gap-2">
            {([
              ['all', 'الكل'],
              ['inventory', 'عناصر المخزون'],
              ['purchases', 'المشتريات'],
              ['audit', 'سجل التدقيق'],
            ] as const).map(([value, label]) => (
              <Button key={value} size="sm" variant={section === value ? 'primary' : 'secondary'} onClick={() => setSection(value)}>{label}</Button>
            ))}
          </div>
          <div className="relative w-full md:max-w-xs">
            <Search className="pointer-events-none absolute right-3 top-2.5 h-4 w-4 text-text-muted" />
            <Input value={searchQuery} onChange={(event) => setSearchQuery(event.target.value)} placeholder="ابحث في السجل التاريخي" className="pr-9" />
          </div>
        </CardContent>
      </Card>

      {isLoading && <Card><CardContent className="p-10 text-center text-text-muted">جار تحميل السجل التاريخي...</CardContent></Card>}

      {!isLoading && show('inventory') && (
        <Card>
          <CardHeader><CardTitle className="flex items-center gap-2"><Archive className="h-5 w-5" />عناصر المخزون المؤرشفة</CardTitle></CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader><TableRow><TableHead>العنصر</TableHead><TableHead>المورد</TableHead><TableHead>التكلفة</TableHead><TableHead>سعر البيع</TableHead><TableHead>تاريخ الشراء</TableHead><TableHead>الحالة</TableHead><TableHead>التفاصيل</TableHead></TableRow></TableHeader>
              <TableBody>
                {filteredInventory.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell><div className="font-semibold">{item.product_name || 'منتج غير معروف'}</div><div className="text-xs text-text-muted">{item.item_code || item.barcode || item.id}</div></TableCell>
                    <TableCell>{item.supplier_name || '-'}</TableCell>
                    <TableCell>{formatMoney(item.purchase_cost)}</TableCell>
                    <TableCell>{formatMoney(item.selling_price)}</TableCell>
                    <TableCell>{formatDate(item.purchase_date || item.created_at)}</TableCell>
                    <TableCell><Badge variant="secondary">مؤرشف</Badge></TableCell>
                    <TableCell><Button size="sm" variant="ghost" onClick={() => setSelectedItem(item)}><History className="ml-1 h-4 w-4" />سجل الحركة</Button></TableCell>
                  </TableRow>
                ))}
                {filteredInventory.length === 0 && <TableRow><TableCell colSpan={7} className="p-8 text-center text-text-muted">لا توجد عناصر مؤرشفة مطابقة.</TableCell></TableRow>}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {!isLoading && show('purchases') && (
        <Card>
          <CardHeader><CardTitle className="flex items-center gap-2"><ClipboardList className="h-5 w-5" />المشتريات الملغاة والمعكوسة</CardTitle></CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader><TableRow><TableHead>رقم الفاتورة</TableHead><TableHead>المورد</TableHead><TableHead>التكلفة</TableHead><TableHead>المدفوع</TableHead><TableHead>التاريخ</TableHead><TableHead>الحالة</TableHead></TableRow></TableHeader>
              <TableBody>
                {filteredPurchases.map((purchase) => (
                  <TableRow key={purchase.id}><TableCell className="font-semibold">{purchase.invoice_number}</TableCell><TableCell>{purchase.supplier?.name || purchase.supplier_name || '-'}</TableCell><TableCell>{formatMoney(purchase.total_amount)}</TableCell><TableCell>{formatMoney(purchase.paid_amount)}</TableCell><TableCell>{formatDate(purchase.purchase_date || purchase.created_at)}</TableCell><TableCell><Badge variant="secondary">{getPurchaseStatus(purchase.status)}</Badge></TableCell></TableRow>
                ))}
                {filteredPurchases.length === 0 && <TableRow><TableCell colSpan={6} className="p-8 text-center text-text-muted">لا توجد مشتريات مؤرشفة مطابقة.</TableCell></TableRow>}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {!isLoading && show('audit') && (
        <Card>
          <CardHeader>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <CardTitle className="flex items-center gap-2"><ClipboardList className="h-5 w-5" />سجل التدقيق</CardTitle>
              <span className="text-xs text-text-muted">{auditTotal} سجل محفوظ · الصفحة {auditPage} من {auditTotalPages}</span>
            </div>
          </CardHeader>
          <CardContent className="space-y-3">
            {auditQuery.isError && <p className="rounded-lg border border-warning/30 bg-warning/10 p-4 text-sm text-text-secondary">تعذر تحميل سجل التدقيق من مصدره الحالي. بيانات الأرشيف والمشتريات المعروضة أعلاه ما زالت مقروءة مباشرة من مصادرها الأصلية.</p>}
            {!auditQuery.isError && <>
              <div className="grid gap-3 sm:grid-cols-3">
                <div className="rounded-xl border border-border bg-surface-elevated/30 px-4 py-3"><div className="flex items-center gap-2 text-xs text-text-muted"><ClipboardList className="h-4 w-4" />السجلات المعروضة</div><p className="mt-1 text-xl font-bold text-text-primary">{visibleAudit.length}</p></div>
                <div className="rounded-xl border border-success/20 bg-success/5 px-4 py-3"><div className="flex items-center gap-2 text-xs text-success"><CheckCircle2 className="h-4 w-4" />ناجحة</div><p className="mt-1 text-xl font-bold text-text-primary">{successfulAuditCount}</p></div>
                <div className="rounded-xl border border-danger/20 bg-danger/5 px-4 py-3"><div className="flex items-center gap-2 text-xs text-danger"><AlertCircle className="h-4 w-4" />فاشلة</div><p className="mt-1 text-xl font-bold text-text-primary">{failedAuditCount}</p></div>
              </div>
              <div className="flex flex-wrap gap-2 border-b border-border pb-3">
                {([['all', 'الكل'], ['sale', 'المبيعات'], ['purchase', 'المشتريات'], ['inventory', 'المخزون']] as const).map(([value, label]) => <Button key={value} size="sm" variant={auditEntityFilter === value ? 'primary' : 'secondary'} onClick={() => setAuditEntityFilter(value)}>{label}</Button>)}
              </div>
              <div className="grid gap-3">
              {visibleAudit.map(({ log, label, details, entityType }) => (
                <article key={log.id} className="relative overflow-hidden rounded-xl border border-border bg-surface-elevated/30 p-4 transition-all hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-sm">
                  <div className="flex items-start gap-3">
                    <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${log.status === 'failure' ? 'bg-danger/10 text-danger' : 'bg-primary/10 text-primary'}`}>{getAuditIcon(entityType)}</div>
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-start justify-between gap-2">
                        <div><p className="font-semibold text-text-primary">{log.description || label}</p><p className="mt-1 text-xs text-text-secondary">{details}</p></div>
                        <Badge variant={log.status === 'failure' ? 'destructive' : 'secondary'}>{log.status === 'failure' ? 'فشل' : 'نجاح'}</Badge>
                      </div>
                      <div className="mt-3 flex flex-wrap gap-x-5 gap-y-2 text-[11px] text-text-muted">
                        <span className="inline-flex items-center gap-1"><Clock3 className="h-3.5 w-3.5" />{formatDate(log.created_at)}</span>
                        <span className="inline-flex items-center gap-1"><UserRound className="h-3.5 w-3.5" />{log.user_name || 'المستخدم'}</span>
                        <span className="inline-flex items-center gap-1"><ReceiptText className="h-3.5 w-3.5" />{log.id}</span>
                      </div>
                    </div>
                  </div>
                </article>
              ))}
              </div>
            </>}
            {!auditQuery.isError && visibleAudit.length === 0 && <p className="p-8 text-center text-text-muted">لا توجد سجلات تدقيق مطابقة.</p>}
            {!auditQuery.isError && auditTotalPages > 1 && (
              <div className="flex flex-wrap items-center justify-between gap-3 border-t border-border pt-4">
                <Button variant="secondary" size="sm" disabled={auditPage <= 1 || auditQuery.isFetching} onClick={() => setAuditPage((page) => Math.max(1, page - 1))}>
                  <ChevronRight className="ml-1 h-4 w-4" />السابق
                </Button>
                <span className="text-xs font-medium text-text-secondary">صفحة {auditPage} من {auditTotalPages}</span>
                <Button variant="secondary" size="sm" disabled={auditPage >= auditTotalPages || auditQuery.isFetching} onClick={() => setAuditPage((page) => Math.min(auditTotalPages, page + 1))}>
                  التالي<ChevronLeft className="mr-1 h-4 w-4" />
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {selectedItem && (
        <Card>
          <CardHeader><CardTitle className="flex items-center justify-between gap-3"><span>تفاصيل العنصر: {selectedItem.product_name || 'منتج غير معروف'}</span><Button variant="ghost" size="icon" onClick={() => setSelectedItem(null)} aria-label="إغلاق التفاصيل"><X className="h-4 w-4" /></Button></CardTitle></CardHeader>
          <CardContent>
            <div className="mb-5 grid gap-3 text-sm md:grid-cols-4"><div><span className="text-text-muted">المعرّف</span><p className="break-all">{selectedItem.id}</p></div><div><span className="text-text-muted">الباركود</span><p>{selectedItem.barcode || '-'}</p></div><div><span className="text-text-muted">الملاحظات</span><p>{selectedItem.notes || '-'}</p></div><div><span className="text-text-muted">آخر تحديث</span><p>{formatDate(selectedItem.updated_at)}</p></div></div>
            <div className="space-y-2"><h3 className="font-semibold">الحركات المرتبطة</h3>{historyQuery.isLoading ? <p className="text-sm text-text-muted">جار تحميل الحركات...</p> : ((historyQuery.data?.data?.movements || []) as any[]).map((movement) => <div key={movement.id} className="flex flex-wrap justify-between gap-2 border-b border-border py-2 text-sm"><span>{translateMovementType(movement.movement_type)}</span><span>{movement.reason || '-'}</span><span>{formatDate(movement.created_at)}</span><span>{movement.quantity ?? 0}</span></div>)}{!historyQuery.isLoading && !((historyQuery.data?.data?.movements || []) as any[]).length && <p className="text-sm text-text-muted">لا توجد حركات مسجلة لهذا العنصر.</p>}</div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
