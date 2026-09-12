import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Activity, ArrowLeft, ArrowRight, Clock, DollarSign, RotateCcw, ShoppingCart } from 'lucide-react';
import { dashboardApi } from '../../../services/api/endpoints';
import { PageHeader } from '../../../components/ui/page-header';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Badge } from '../../../components/ui/badge';

const PAGE_SIZE = 10;

const statusLabels: Record<string, string> = {
  completed: 'مكتمل',
  pending: 'قيد الانتظار',
  reversed: 'تم عكس العملية',
  cancelled: 'ملغي',
  received: 'مستلم',
  ordered: 'تم الطلب',
  draft: 'مسودة',
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

function formatActivityTime(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString('ar');
}

export function ActivityPage() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [activityType, setActivityType] = useState('');
  const { data, isLoading, error } = useQuery({
    queryKey: ['dashboard-activity', page, activityType],
    queryFn: () => dashboardApi.getActivity({
      page,
      per_page: PAGE_SIZE,
      type: activityType || undefined,
    }),
    staleTime: 30000,
  });

  const activityPage = data?.data;
  const items = activityPage?.items || [];
  const totalPages = activityPage?.total_pages || 0;

  const changeType = (value: string) => {
    setActivityType(value);
    setPage(1);
  };

  return (
    <div>
      <PageHeader
        eyebrow="السجل"
        title="كل النشاط"
        description="سجل المبيعات والمشتريات مرتبًا من الأحدث إلى الأقدم"
        actions={(
          <Button type="button" variant="secondary" onClick={() => navigate(-1)} className="gap-2">
            <ArrowRight className="w-4 h-4" />
            رجوع
          </Button>
        )}
      />

      <Card variant="open" style={{ marginTop: 'var(--spacing-6)' }}>
        <CardHeader>
          <div className="flex items-center justify-between gap-3 flex-wrap">
            <CardTitle className="flex items-center gap-2">
              <Clock className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
              سجل العمليات
            </CardTitle>
            <select
              value={activityType}
              onChange={(event) => changeType(event.target.value)}
              aria-label="تصفية نوع النشاط"
              style={{
                minWidth: '150px',
                padding: '8px 10px',
                borderRadius: 'var(--radius-md)',
                border: '1px solid var(--border-default)',
                background: 'var(--bg-surface)',
                color: 'var(--text-primary)',
              }}
            >
              <option value="">كل العمليات</option>
              <option value="sale">المبيعات</option>
              <option value="purchase">المشتريات</option>
            </select>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-center text-text-muted py-8">جارِ تحميل النشاط...</p>
          ) : error ? (
            <p className="text-center text-text-muted py-8">تعذر تحميل سجل النشاط</p>
          ) : items.length === 0 ? (
            <p className="text-center text-text-muted py-8">لا توجد عمليات مطابقة</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
              {items.map((item: any) => {
                const Icon = getActivityIcon(item.type);
                const normalizedStatus = String(item.status || '').toLowerCase();
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
                    <div style={{ flex: 1 }}>
                      <p style={{ fontWeight: 'var(--font-weight-medium)', color: 'var(--text-primary)' }}>{item.title}</p>
                      <p style={{ fontSize: 'var(--font-size-caption)', color: 'var(--text-secondary)' }}>{item.description}</p>
                    </div>
                    <div style={{ textAlign: 'right' }}>
                      <p style={{ fontWeight: 'var(--font-weight-medium)', color: 'var(--text-primary)' }}>₪{Number(item.amount || 0).toLocaleString('en-US')}</p>
                      <p style={{ fontSize: 'var(--font-size-caption)', color: 'var(--text-secondary)' }}>{formatActivityTime(item.time)}</p>
                    </div>
                    <Badge variant={normalizedStatus === 'completed' ? 'success' : 'warning'} size="sm">
                      {getStatusLabel(item.status)}
                    </Badge>
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
    </div>
  );
}
