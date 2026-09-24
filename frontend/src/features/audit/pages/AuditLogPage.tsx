import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ClipboardList, Clock3, Search, ShieldAlert, ShieldCheck } from 'lucide-react';
import { Badge } from '../../../design-system/components/badge';
import { Button } from '../../../design-system/components/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Input } from '../../../design-system/components/input';
import { PageHeader } from '../../../design-system/components/page-header';
import { auditApi } from '../../../services/api/endpoints';
import { useDebounce } from '../../../hooks/useDebounce';
import { formatStoreDateTime } from '../../../utils/store-time';

const actionLabels: Record<string, string> = {
  DELETE: 'حذف سجل',
  CREATE: 'إنشاء سجل',
  UPDATE: 'تعديل سجل',
  CREATE_SALE: 'إنشاء عملية بيع',
  DELETE_SALE: 'حذف عملية بيع',
  CREATE_PURCHASE: 'إنشاء عملية شراء',
  DELETE_PURCHASE: 'حذف عملية شراء',
  DELETE_INVENTORY: 'حذف عنصر مخزون',
  ADD_PAYMENT: 'تسجيل دفعة',
};

const entityLabels: Record<string, string> = {
  sale: 'بيع',
  sales: 'بيع',
  purchase: 'شراء',
  purchases: 'شراء',
  return: 'مرتجع',
  customer: 'عميل',
  supplier: 'تاجر',
  product: 'منتج',
  inventory: 'مخزون',
  inventory_item: 'عنصر مخزون',
  payment: 'دفعة',
  expense: 'مصروف',
  barcode: 'باركود',
};

const entityFilters = [
  { value: '', label: 'الكل' },
  { value: 'sale', label: 'المبيعات' },
  { value: 'purchase', label: 'المشتريات' },
  { value: 'inventory', label: 'المخزون' },
  { value: 'customer', label: 'العملاء' },
  { value: 'supplier', label: 'التجار' },
  { value: 'return', label: 'المرتجعات' },
];

export function AuditLogPage() {
  const [search, setSearch] = useState('');
  const [entityType, setEntityType] = useState('');
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const debouncedSearch = useDebounce(search, 250);

  useEffect(() => setPage(1), [debouncedSearch, entityType]);

  const auditQuery = useQuery({
    queryKey: ['audit-log', page, pageSize, debouncedSearch, entityType],
    queryFn: () => auditApi.list({
      page,
      per_page: pageSize,
      ...(debouncedSearch ? { search: debouncedSearch } : {}),
      ...(entityType ? { entity_type: entityType } : {}),
    }),
  });

  const logs = (auditQuery.data?.data || []) as any[];
  const total = Number(auditQuery.data?.meta?.total || 0);
  const pages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <div className="space-y-5">
      <PageHeader
        title="سجل النظام"
        description="سجل تدقيق للعمليات، بما فيها أحداث الحذف، من دون الاحتفاظ بنسخ مؤرشفة من السجلات التجارية."
      />

      <Card>
        <CardContent className="space-y-4 p-4">
          <div className="relative max-w-[36rem]">
            <Search className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
            <Input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="ابحث في وصف العملية أو المستخدم أو المعرّف" className="pr-10" />
          </div>
          <div className="flex flex-wrap gap-2" role="group" aria-label="تصفية سجل النظام حسب القسم">
            {entityFilters.map((filter) => (
              <Button key={filter.value || 'all'} size="sm" variant={entityType === filter.value ? 'primary' : 'secondary'} onClick={() => setEntityType(filter.value)}>
                {filter.label}
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between gap-3">
            <CardTitle className="flex items-center gap-2"><ClipboardList className="h-5 w-5" />سجل التدقيق</CardTitle>
            <span className="text-xs text-text-muted">{total} سجل</span>
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          {auditQuery.isLoading && <p className="p-8 text-center text-text-muted">جار تحميل السجلات...</p>}
          {auditQuery.isError && <p role="alert" className="rounded-lg border border-danger/30 bg-danger/10 p-4 text-sm text-danger">تعذر تحميل سجل النظام.</p>}
          {!auditQuery.isLoading && !auditQuery.isError && logs.length === 0 && <p className="p-8 text-center text-text-muted">لا توجد سجلات مطابقة.</p>}
          {logs.map((log) => {
            const isFailure = String(log.status || '').toLowerCase() === 'failure';
            const action = String(log.action || '').toUpperCase();
            const entity = String(log.entity_type || '').toLowerCase();
            return (
              <article key={log.id} className="flex min-w-0 gap-3 rounded-xl border border-border bg-surface p-4">
                <span className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-lg ${isFailure ? 'bg-danger/10 text-danger' : 'bg-primary/10 text-primary'}`}>
                  {isFailure ? <ShieldAlert className="h-4 w-4" /> : <ShieldCheck className="h-4 w-4" />}
                </span>
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <p className="font-semibold text-text-primary">{log.description || actionLabels[action] || log.action || 'عملية مسجلة'}</p>
                    <Badge variant={isFailure ? 'destructive' : 'secondary'}>{isFailure ? 'فشل' : 'نجاح'}</Badge>
                  </div>
                  <p className="mt-1 text-sm text-text-secondary">
                    {actionLabels[action] || log.action || 'عملية'} · {entityLabels[entity] || log.entity_type || 'سجل'} · {log.entity_id || '-'}
                  </p>
                  <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-text-muted">
                    <span className="inline-flex items-center gap-1"><Clock3 className="h-3.5 w-3.5" />{log.created_at ? formatStoreDateTime(log.created_at) : '-'}</span>
                    <span>{log.user_name || 'المستخدم'}</span>
                  </div>
                </div>
              </article>
            );
          })}
          {!auditQuery.isError && pages > 1 && (
            <div className="flex items-center justify-between border-t border-border pt-3">
              <Button variant="secondary" size="sm" disabled={page <= 1 || auditQuery.isFetching} onClick={() => setPage((current) => Math.max(1, current - 1))}>السابق</Button>
              <span className="text-xs text-text-muted">صفحة {page} من {pages}</span>
              <Button variant="secondary" size="sm" disabled={page >= pages || auditQuery.isFetching} onClick={() => setPage((current) => Math.min(pages, current + 1))}>التالي</Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
