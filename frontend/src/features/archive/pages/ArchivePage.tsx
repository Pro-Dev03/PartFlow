import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { AlertCircle, Archive, CheckCircle2, ChevronLeft, ChevronRight, ClipboardList, Clock3, CreditCard, Eye, History, Package, PackageOpen, ReceiptText, Search, ShoppingCart, Truck, UserRound, X } from 'lucide-react';
import { PageHeader } from '../../../design-system/components/page-header';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Badge } from '../../../design-system/components/badge';
import { StatCard } from '../../../design-system/components/stat-card';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../design-system/components/table';
import { auditApi, customersApi, inventoryApi, purchasesApi, returnsApi, salesApi, settingsApi, supplierReturnsApi } from '../../../services/api/endpoints';

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

const getReturnStatus = (status?: string) => {
  switch (String(status || '').toUpperCase()) {
    case 'COMPLETED': return 'مكتمل';
    case 'REJECTED': return 'مرفوض';
    case 'CANCELLED': return 'ملغى';
    default: return status || 'غير محدد';
  }
};

const getReturnType = (type?: string) => ({
  FULL: 'مرتجع كامل',
  PARTIAL: 'مرتجع جزئي',
  QUANTITY_PARTIAL: 'مرتجع كمية جزئية',
}[String(type || '').toUpperCase()] || type || 'غير محدد');

const getRefundMethod = (method?: string) => ({
  CASH: 'نقدي',
  DEBT_ADJUSTMENT: 'تعديل دين',
  EXCHANGE: 'استبدال',
  BANK_TRANSFER: 'تحويل بنكي',
}[String(method || '').toUpperCase()] || method || 'غير محدد');

const getReturnReason = (reason?: string) => ({
  DEFECTIVE: 'منتج معطل',
  WRONG_ITEM: 'منتج خاطئ',
  COMPATIBILITY_ISSUE: 'مشكلة توافق',
  CUSTOMER_CHANGED_MIND: 'تغيير رأي العميل',
  DAMAGED: 'منتج تالف',
  WARRANTY: 'ضمان',
  INCORRECT_SPECIFICATION: 'مواصفات غير صحيحة',
  OTHER: 'أخرى',
}[String(reason || '').toUpperCase()] || reason || 'غير محدد');

const getReturnCondition = (condition?: string) => ({
  READY_FOR_SALE: 'جاهز للبيع',
  NOT_FOR_SALE: 'غير قابل للبيع',
  RETURN_TO_SUPPLIER: 'إرجاع للتاجر',
  SUPPLIER_RETURN: 'إرجاع للتاجر',
  NEEDS_REPAIR: 'يحتاج إصلاح',
  DAMAGED: 'تالف',
  USED: 'مستعمل',
  REFURBISHED: 'مجدّد',
  WRITE_OFF: 'شطب',
  PARTS: 'قطع غيار',
}[String(condition || '').toUpperCase()] || condition || 'غير محدد');

const normalizeAuditEntityType = (entityType?: string) => {
  const normalized = String(entityType || '').trim().toLowerCase();
  if (!normalized) return 'unknown';
  if (normalized === 'inventory_item' || normalized === 'inventoryitem' || normalized === 'stock') return 'inventory';
  if (normalized === 'sale' || normalized === 'sales') return 'sale';
  if (normalized === 'purchase' || normalized === 'purchases') return 'purchase';
  return normalized;
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
  const normalized = normalizeAuditEntityType(entityType);
  const labels: Record<string, string> = {
    sale: 'بيع',
    purchase: 'شراء',
    inventory: 'مخزون',
    product: 'منتج',
    payment: 'دفعة',
    customer: 'عميل',
    supplier: 'تاجر',
  };
  return labels[normalized] || entityType || 'سجل';
};

const translateMovementType = (movementType?: string) => {
  const labels: Record<string, string> = {
    PURCHASE: 'شراء',
    SALE: 'بيع',
    RETURN: 'مرتجع',
    SUPPLIER_RETURN: 'مرتجع تاجر',
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
  const [section, setSection] = useState<'inventory' | 'purchases' | 'returns' | 'audit'>('audit');
  const [selectedItem, setSelectedItem] = useState<any | null>(null);
  const [selectedReturn, setSelectedReturn] = useState<any | null>(null);
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
  const returnsQuery = useQuery({
    queryKey: ['archive', 'returns'],
    queryFn: () => returnsApi.list({ page: 1, per_page: 100 }),
  });
  const supplierReturnsQuery = useQuery({
    queryKey: ['archive', 'supplier-returns'],
    queryFn: () => supplierReturnsApi.list(),
  });
  const usersQuery = useQuery({
    queryKey: ['archive', 'users'],
    queryFn: () => settingsApi.getUsers({ page: 1, per_page: 100 }),
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
  const usersPayload = usersQuery.data?.data;
  const users = (Array.isArray(usersPayload) ? usersPayload : usersPayload?.users || []) as any[];
  const getUserName = (id?: string) => {
    if (!id) return 'غير مسجل';
    const user = users.find((item) => String(item.id) === String(id));
    return user ? `${user.first_name || ''} ${user.last_name || ''}`.trim() || user.email || id : id;
  };
  const customerReturnsPayload = returnsQuery.data?.data;
  const customerReturns = (Array.isArray(customerReturnsPayload) ? customerReturnsPayload : customerReturnsPayload?.returns || []) as any[];
  const supplierReturns = (supplierReturnsQuery.data?.data || []) as any[];
  const archivedReturns = [
    ...customerReturns
      .filter((item) => ['COMPLETED', 'REJECTED', 'CANCELLED'].includes(String(item.status || '').toUpperCase()))
      .map((item) => ({ ...item, return_kind: 'مرتجع عميل', display_amount: item.total_refund_amount, display_date: item.return_date || item.created_at, created_by_name: item.created_by_name || getUserName(item.created_by), processed_by_name: item.processed_by_name || getUserName(item.processed_by), approved_by_name: item.approved_by_name || getUserName(item.approved_by) })),
    ...supplierReturns
      .filter((item) => ['COMPLETED', 'REJECTED', 'CANCELLED'].includes(String(item.status || '').toUpperCase()))
      .map((item) => ({ ...item, return_kind: 'مرتجع تاجر', display_amount: item.refund_amount, display_date: item.created_at, created_by_name: item.created_by_name || getUserName(item.created_by), processed_by_name: item.processed_by_name || getUserName(item.processed_by), approved_by_name: item.approved_by_name || getUserName(item.approved_by) })),
  ];
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

  const filteredReturns = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    if (!query) return archivedReturns;
    return archivedReturns.filter((item) => [item.return_number, item.reason, item.return_kind, item.reference_number, item.purchase_id]
      .some((value) => String(value || '').toLowerCase().includes(query)));
  }, [archivedReturns, searchQuery]);

  const auditRows = useMemo(() => auditLogs.map((log) => {
    const entityType = normalizeAuditEntityType(log.entity_type);
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
        ? `التاجر: ${purchase.supplier?.name || purchase.supplier_name || 'غير محدد'}`
        : '';
    const amount = sale
      ? `الإجمالي: ${formatMoney(sale.total_amount)}`
      : purchase
        ? `الإجمالي: ${formatMoney(purchase.total_amount)} · المدفوع: ${formatMoney(purchase.paid_amount)}`
        : '';
    let saleDetails = '';
    if (entityType === 'sale') {
      try {
        const payload = JSON.parse(String(log.new_values || log.changes || '{}')) as any;
        const itemDetails = Array.isArray(payload.items)
          ? payload.items.map((item: any) => `${item.product_name || item.product_id || 'منتج'} × ${item.quantity} بسعر ${formatMoney(item.unit_price)}`).join('، ')
          : '';
        saleDetails = [
          itemDetails ? `الأصناف: ${itemDetails}` : '',
          payload.payment_method ? `الدفع: ${payload.payment_method}` : '',
          payload.discount ? `الخصم: ${formatMoney(payload.discount)}` : '',
          payload.tax ? `الضريبة: ${formatMoney(payload.tax)}` : '',
          payload.notes ? `ملاحظات: ${payload.notes}` : '',
        ].filter(Boolean).join(' · ');
      } catch {
        saleDetails = '';
      }
    }
    return {
      log,
      label: translateAuditAction(log.action),
      entityType,
      details: [translateEntityType(log.entity_type), reference, party, amount, saleDetails].filter(Boolean).join(' · '),
      searchText: [log.action, log.entity_type, log.description, log.user_name, reference, party, amount, saleDetails].join(' ').toLowerCase(),
    };
  }), [auditLogs, customers, purchases, sales]);

  const filteredAudit = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    if (!query) return auditRows;
    return auditRows.filter((row) => row.searchText.includes(query));
  }, [auditRows, searchQuery]);

  const visibleAudit = auditEntityFilter === 'all'
    ? filteredAudit
    : filteredAudit.filter((row) => normalizeAuditEntityType(row.entityType) === auditEntityFilter);
  const successfulAuditCount = visibleAudit.filter(({ log }) => log.status !== 'failure').length;
  const failedAuditCount = visibleAudit.filter(({ log }) => log.status === 'failure').length;
  const getAuditIcon = (entityType: string) => {
    if (entityType === 'sale') return <ShoppingCart className="h-5 w-5" />;
    if (entityType === 'purchase') return <Truck className="h-5 w-5" />;
    if (entityType === 'inventory') return <PackageOpen className="h-5 w-5" />;
    if (entityType === 'payment') return <CreditCard className="h-5 w-5" />;
    return <ClipboardList className="h-5 w-5" />;
  };

  const show = (value: typeof section) => section === value;
  const isLoading = section === 'inventory'
    ? archivedInventoryQuery.isLoading
    : section === 'purchases'
      ? purchasesQuery.isLoading
      : section === 'returns'
        ? returnsQuery.isLoading || supplierReturnsQuery.isLoading
        : auditQuery.isLoading || salesQuery.isLoading || purchasesQuery.isLoading || customersQuery.isLoading || usersQuery.isLoading;
  const totalArchivedValue = archivedInventory.reduce((sum, item) => sum + Number(item.selling_price || 0), 0);

  return (
    <div className="space-y-6">
      <PageHeader
        title="الأرشيف والسجل التاريخي"
        description="مرجع دقيق للعناصر والعمليات التي خرجت من التشغيل دون حذف تاريخها التجاري."
      />

      <div className="unified-stats-grid supplier-stats grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard
          title="عناصر المخزون المؤرشفة"
          value={archivedInventory.length}
          icon={Archive}
          subtitle="عناصر محفوظة في السجل التاريخي"
          variant="featured"
          size="sm"
        />
        <StatCard
          title="العمليات الملغاة أو المعكوسة"
          value={archivedPurchases.length}
          icon={ClipboardList}
          subtitle="عمليات خرجت من التشغيل"
          variant="warning"
          size="sm"
        />
        <StatCard
          title="قيمة البيع التاريخية للعناصر المؤرشفة"
          value={formatMoney(totalArchivedValue)}
          icon={Package}
          subtitle="القيمة المسجلة للعناصر المؤرشفة"
          variant="success"
          size="sm"
        />
      </div>

      <Card>
        <CardContent className="flex flex-col gap-3 p-4 md:flex-row md:items-center md:justify-between">
          <div className="flex min-w-0 flex-1 flex-col gap-2">
            <p className="text-xs font-bold text-text-primary">استعرض حسب النوع</p>
            <div className="flex flex-wrap gap-2" role="tablist" aria-label="أقسام الأرشيف">
            {([
              ['inventory', 'عناصر المخزون'],
              ['purchases', 'المشتريات'],
              ['returns', 'المرتجعات'],
              ['audit', 'سجل التدقيق'],
            ] as const).map(([value, label]) => (
              <Button key={value} size="sm" variant={section === value ? 'primary' : 'ghost'} className="gap-sm whitespace-nowrap px-3" role="tab" aria-selected={section === value} onClick={() => setSection(value)}><span className="text-tiny">{label}</span></Button>
            ))}
            </div>
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
              <TableHeader><TableRow><TableHead>العنصر</TableHead><TableHead>التاجر</TableHead><TableHead>التكلفة</TableHead><TableHead>سعر البيع</TableHead><TableHead>تاريخ الشراء</TableHead><TableHead>الحالة</TableHead><TableHead className="min-w-[72px] text-center">التفاصيل</TableHead></TableRow></TableHeader>
              <TableBody>
                {filteredInventory.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell><div className="font-semibold">{item.product_name || 'منتج غير معروف'}</div><div className="text-xs text-text-muted">{item.item_code || item.barcode || item.id}</div></TableCell>
                    <TableCell>{item.supplier_name || '-'}</TableCell>
                    <TableCell>{formatMoney(item.purchase_cost)}</TableCell>
                    <TableCell>{formatMoney(item.selling_price)}</TableCell>
                    <TableCell>{formatDate(item.purchase_date || item.created_at)}</TableCell>
                    <TableCell><Badge variant="secondary">مؤرشف</Badge></TableCell>
                    <TableCell className="min-w-[72px] text-center"><Button size="icon" variant="ghost" title="عرض سجل حركة العنصر" aria-label={`عرض سجل حركة ${item.product_name || item.id}`} onClick={() => setSelectedItem(item)}><History className="h-4 w-4" /></Button></TableCell>
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
              <TableHeader><TableRow><TableHead>رقم الفاتورة</TableHead><TableHead>التاجر</TableHead><TableHead>التكلفة</TableHead><TableHead>المدفوع</TableHead><TableHead>التاريخ</TableHead><TableHead>الحالة</TableHead></TableRow></TableHeader>
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

      {!isLoading && show('returns') && (
        <Card>
          <CardHeader><CardTitle className="flex items-center gap-2"><Archive className="h-5 w-5" />المرتجعات المؤرشفة</CardTitle></CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader><TableRow><TableHead>نوع المرتجع</TableHead><TableHead>رقم المرتجع</TableHead><TableHead>السبب</TableHead><TableHead>القيمة</TableHead><TableHead>التاريخ</TableHead><TableHead>الحالة</TableHead><TableHead className="min-w-[72px] text-center">التفاصيل</TableHead></TableRow></TableHeader>
              <TableBody>
                {filteredReturns.map((item) => (
                  <TableRow key={`${item.return_kind}-${item.id}`}>
                    <TableCell>{item.return_kind}</TableCell>
                    <TableCell className="font-semibold">{item.return_number || item.id}</TableCell>
                    <TableCell>{getReturnReason(item.reason)}</TableCell>
                    <TableCell>{formatMoney(item.display_amount)}</TableCell>
                    <TableCell>{formatDate(item.display_date)}</TableCell>
                    <TableCell><Badge variant="secondary">{getReturnStatus(item.status)}</Badge></TableCell>
                    <TableCell className="min-w-[72px] text-center"><Button size="icon" variant="ghost" title="عرض تفاصيل المرتجع" aria-label={`عرض تفاصيل المرتجع ${item.return_number || item.id}`} onClick={() => setSelectedReturn(item)}><Eye className="h-4 w-4" /></Button></TableCell>
                  </TableRow>
                ))}
                {filteredReturns.length === 0 && <TableRow><TableCell colSpan={7} className="p-8 text-center text-text-muted">لا توجد مرتجعات مؤرشفة مطابقة.</TableCell></TableRow>}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {section === 'returns' && selectedReturn && (
        <Card>
          <CardHeader><CardTitle className="flex items-center justify-between gap-3"><span>تفاصيل {selectedReturn.return_kind}: {selectedReturn.return_number || selectedReturn.id}</span><Button variant="ghost" size="icon" onClick={() => setSelectedReturn(null)} aria-label="إغلاق التفاصيل"><X className="h-4 w-4" /></Button></CardTitle></CardHeader>
          <CardContent className="space-y-5">
            <div className="grid gap-4 text-sm sm:grid-cols-2 lg:grid-cols-4">
              <div><span className="text-text-muted">العميل / التاجر</span><p className="font-medium">{selectedReturn.customer_name || selectedReturn.supplier_name || selectedReturn.supplier_id || '-'}</p></div>
              <div><span className="text-text-muted">الفاتورة المرتبطة</span><p className="font-medium">{selectedReturn.sale_invoice || selectedReturn.reference_number || selectedReturn.invoice_number || selectedReturn.purchase_id || selectedReturn.sale_id || '-'}</p></div>
              <div><span className="text-text-muted">نوع المرتجع</span><p>{getReturnType(selectedReturn.return_type) || selectedReturn.return_kind}</p></div>
              <div><span className="text-text-muted">الحالة</span><div className="mt-1"><Badge variant="secondary">{getReturnStatus(selectedReturn.status)}</Badge></div></div>
              <div><span className="text-text-muted">طريقة الاسترداد</span><p>{selectedReturn.refund_method ? getRefundMethod(selectedReturn.refund_method) : 'استرداد تاجر'}</p></div>
              <div><span className="text-text-muted">تاريخ المرتجع</span><p>{formatDate(selectedReturn.return_date || selectedReturn.display_date)}</p></div>
              <div><span className="text-text-muted">تاريخ الاسترداد</span><p>{formatDate(selectedReturn.refund_date)}</p></div>
              <div><span className="text-text-muted">مرجع الاسترداد</span><p>{selectedReturn.refund_reference || '-'}</p></div>
              <div><span className="text-text-muted">صاحب الحساب المنفذ</span><p>{selectedReturn.created_by_name}</p></div>
              <div><span className="text-text-muted">صاحب الحساب المعالج</span><p>{selectedReturn.processed_by_name}</p></div>
              <div><span className="text-text-muted">صاحب الحساب المعتمد</span><p>{selectedReturn.approved_by_name}</p></div>
              <div><span className="text-text-muted">آخر تحديث</span><p>{formatDate(selectedReturn.updated_at)}</p></div>
            </div>
            <div className="grid gap-4 text-sm sm:grid-cols-2">
              <div><span className="text-text-muted">سبب المرتجع</span><p>{getReturnReason(selectedReturn.reason)}</p></div>
              <div><span className="text-text-muted">حالة الصنف بعد المرتجع</span><p>{getReturnCondition(selectedReturn.item_condition_after_return)}</p></div>
              <div><span className="text-text-muted">المبلغ المسترد</span><p className="font-semibold">{formatMoney(selectedReturn.display_amount)}</p></div>
              <div><span className="text-text-muted">تعديل الدين / رصيد العميل</span><p>{formatMoney(selectedReturn.debt_adjustment || selectedReturn.customer_credit)}</p></div>
              <div><span className="text-text-muted">الملاحظات</span><p>{selectedReturn.notes || '-'}</p></div>
              <div><span className="text-text-muted">الملاحظات الداخلية</span><p>{selectedReturn.internal_notes || '-'}</p></div>
            </div>
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
              <div className="unified-stats-grid supplier-stats grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
                <StatCard
                  title="السجلات المعروضة"
                  value={visibleAudit.length}
                  icon={ClipboardList}
                  subtitle="السجلات المطابقة للفلاتر"
                  variant="featured"
                  size="sm"
                />
                <StatCard
                  title="ناجحة"
                  value={successfulAuditCount}
                  icon={CheckCircle2}
                  subtitle="عمليات اكتملت بنجاح"
                  variant="success"
                  size="sm"
                />
                <StatCard
                  title="فاشلة"
                  value={failedAuditCount}
                  icon={AlertCircle}
                  subtitle="عمليات تحتاج مراجعة"
                  variant="danger"
                  size="sm"
                />
              </div>
              <div className="flex flex-wrap gap-2 border-b border-border pb-3">
                {([['all', 'الكل'], ['sale', 'المبيعات'], ['purchase', 'المشتريات'], ['inventory', 'المخزون']] as const).map(([value, label]) => <Button key={value} size="sm" variant={auditEntityFilter === value ? 'primary' : 'ghost'} className="gap-sm whitespace-nowrap px-3" onClick={() => setAuditEntityFilter(value)}><span className="text-tiny">{label}</span></Button>)}
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
                <Button variant="ghost" size="sm" className="gap-sm" disabled={auditPage <= 1 || auditQuery.isFetching} onClick={() => setAuditPage((page) => Math.max(1, page - 1))}>
                  <ChevronRight className="h-4 w-4" />السابق
                </Button>
                <span className="text-xs font-medium text-text-secondary">صفحة {auditPage} من {auditTotalPages}</span>
                <Button variant="ghost" size="sm" className="gap-sm" disabled={auditPage >= auditTotalPages || auditQuery.isFetching} onClick={() => setAuditPage((page) => Math.min(auditTotalPages, page + 1))}>
                  التالي<ChevronLeft className="h-4 w-4" />
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {section === 'inventory' && selectedItem && (
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
