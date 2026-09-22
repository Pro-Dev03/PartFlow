import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { useNavigate } from 'react-router-dom';
import { useIsMobile } from '../../../hooks/useIsMobile';
import { dashboardApi, debtsApi, inventoryApi, productsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { PageHeader } from '../../../design-system/components/page-header';
import { Badge } from '../../../design-system/components/badge';
import { StatCardSkeleton } from '../../../design-system/components/skeleton';
import { SalesChart } from '../../../components/charts/SalesChart';
import { DashboardMetrics } from '../components/DashboardMetrics';
import { AttentionSection } from '../components/AttentionSection';
import { SmartActions } from '../components/SmartActions';
import { InventoryDistribution } from '../../../components/dashboard/InventoryDistribution';
import { DashboardStats } from '../../../types/api';
import { addStoreDays, formatStoreActivityDateTime, formatStoreDateTime, getStoreDateKey } from '../../../utils/store-time';
import {
  ShoppingCart,
  DollarSign,
  AlertTriangle,
  Clock,
  Activity,
  RotateCcw,
  Package,
  ArrowLeft,
} from 'lucide-react';

function formatDashboardActivityTime(value: unknown) {
  return value ? formatStoreDateTime(String(value), 'ar') : '-';
}

type PerformancePoint = {
  name: string;
  sales: number;
  profit: number;
};

function buildPerformanceChartData(points: PerformancePoint[], days: number) {
  const pointsByDate = new Map(points.map((point) => [point.name, point]));
  const endDate = getStoreDateKey(new Date()) || '';
  const startDate = addStoreDays(endDate, -days + 1);

  return Array.from({ length: days }, (_, index) => {
    const dateKey = addStoreDays(startDate, index);
    const point = pointsByDate.get(dateKey);
    return {
      name: dateKey,
      sales: Number(point?.sales ?? 0),
      profit: Number(point?.profit ?? 0),
    };
  });
}

export function DashboardPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isMobile = useIsMobile();
  const [chartRange, setChartRange] = useState<1 | 7 | 30 | 90>(7);
  const { data: dashboardData, isLoading, error } = useQuery({
    queryKey: ['dashboard'],
    queryFn: () => dashboardApi.getStats(),
    refetchInterval: 120000,
    staleTime: 60000,
  });

  const { data: inventoryData } = useQuery({
    queryKey: ['dashboard-inventory-distribution'],
    queryFn: () => inventoryApi.listWithSupplier({ page: 1, per_page: 1000, exclude_condition: 'USED' }),
    staleTime: 60000,
  });
  const { data: productsData } = useQuery({
    queryKey: ['dashboard-inventory-products'],
    queryFn: () => productsApi.list({ page: 1, per_page: 1000 }),
    staleTime: 60000,
  });

  const { data: recentActivityData } = useQuery({
    queryKey: ['dashboard-activity', 'recent'],
    queryFn: () => dashboardApi.getActivity({ page: 1, per_page: 5 }),
    refetchInterval: 120000,
    staleTime: 60000,
  });

  // Low stock items and overdue debts for attention section
  const { data: lowStockItemsData } = useQuery({
    queryKey: ['low-stock-items'],
    queryFn: () => dashboardApi.getLowStockItems(),
    refetchInterval: 300000,
    staleTime: 180000,
  });

  const { data: overdueDebtsData } = useQuery({
    queryKey: ['overdue-debts'],
    queryFn: () => dashboardApi.getOverdueDebts(),
    refetchInterval: 300000,
    staleTime: 180000,
  });

  const { data: debtsData } = useQuery({
    queryKey: ['debts'],
    queryFn: () => debtsApi.list({ page: 1, per_page: 100 }),
    refetchInterval: 300000,
    staleTime: 180000,
  });

  if (isLoading) {
    return (
      <div>
        <PageHeader
          title={t('dashboard.title')}
          description={t('dashboard.subtitle')}
        />
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '12px' }}>
          <StatCardSkeleton />
          <StatCardSkeleton />
          <StatCardSkeleton />
          <StatCardSkeleton />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '256px' }}>
        <div style={{ textAlign: 'center' }}>
          <AlertTriangle className="w-12 h-12 mx-auto mb-2" style={{ color: 'var(--color-danger)' }} />
          <p style={{ color: 'var(--text-secondary)' }}>
            {t('common.error')}
          </p>
        </div>
      </div>
    );
  }

  const stats = dashboardData?.data as DashboardStats | undefined;
  const inventoryItems = (inventoryData?.data as any)?.items || [];
  const availableProductIds = new Set<string>();
  const fallbackDistribution = Object.values(inventoryItems
    .filter((item: any) => String(item.status || '').toUpperCase() !== 'ARCHIVED')
    .reduce((groups: Record<string, { name: string; count: number; value: number; color: string; status: string }>, item: any) => {
      const status = String(item.status || 'UNKNOWN').toUpperCase();
      const presentation: Record<string, { name: string; color: string; health: string }> = {
        AVAILABLE: { name: 'متاح', color: '#10b981', health: 'good' },
        SOLD: { name: 'قطع مباعة', color: '#64748b', health: 'neutral' },
        RETURNED: { name: 'قطع مرتجعة', color: '#f59e0b', health: 'attention' },
        DAMAGED: { name: 'تالف/قيد الإصلاح', color: '#ef4444', health: 'critical' },
        IN_REPAIR: { name: 'تالف/قيد الإصلاح', color: '#ef4444', health: 'critical' },
      };
      const current = presentation[status] || { name: status, color: '#94a3b8', health: 'attention' };
      const group = groups[current.name] || { name: current.name, count: 0, value: 0, color: current.color, status: current.health };
      const productId = String(item.product_id || '');
      if (status === 'AVAILABLE' && productId) {
        if (availableProductIds.has(productId)) return groups;
        availableProductIds.add(productId);
      }
      const quantity = status === 'AVAILABLE'
        ? Number(item.available_quantity ?? item.current_quantity ?? item.quantity ?? 1)
        : 1;
      group.count += quantity;
      group.value += quantity * Number(item.selling_price || 0);
      groups[current.name] = group;
      return groups;
    }, {}));
  const inventoryProducts = (productsData?.data as any)?.products || [];
  const availableSummary = inventoryProducts.reduce((summary: { count: number; value: number }, product: any) => {
    if (String(product.condition || '').toUpperCase() === 'USED') return summary;
    const quantity = Number(product.current_quantity ?? product.stock ?? 0);
    if (quantity <= 0) return summary;
    summary.count += quantity;
    summary.value += quantity * Number(product.selling_price ?? product.sellingPrice ?? 0);
    return summary;
  }, { count: 0, value: 0 });
  const normalizedFallbackDistribution = fallbackDistribution.map((item: any) => item.name === 'متاح' && availableSummary.count > 0
    ? { ...item, count: availableSummary.count, value: availableSummary.value }
    : item);
  const inventoryDistribution = normalizedFallbackDistribution.length
    ? normalizedFallbackDistribution
    : (stats?.inventoryDistribution?.data || []);
  const recentActivities = recentActivityData?.data?.items || dashboardData?.data?.recent_activity || [];
  const chartData = buildPerformanceChartData(stats?.salesChart || [], chartRange);
  const activeCustomerCount = Number(stats?.activeCustomers ?? 0);
  const lowStockItems = lowStockItemsData?.data || [];
  const lowStockAlertCount = Number(stats?.lowStockCount ?? 0);
  const todayDay = getStoreDateKey(new Date()) || '';
  const overdueDebtItems = (overdueDebtsData?.data || []).filter((debt: any) => {
    const dueDate = debt.due_date || debt.dueDate;
    const dueDay = dueDate ? String(dueDate).slice(0, 10) : '';
    return Number(debt.remaining_amount ?? debt.remainingAmount ?? 0) > 0 && dueDay && dueDay < todayDay;
  });
  const overdueDebtCount = new Set(
    overdueDebtItems
      .map((debt: any) => debt.customer_id || debt.customerId || debt.customer?.id)
      .filter(Boolean)
  ).size;
  const debtRows = Array.isArray(debtsData?.data) ? debtsData.data : [];
  const unpaidDebtItems = debtRows
    .filter((debt: any) => {
      const dueDate = debt.due_date || debt.dueDate;
      const dueDay = dueDate ? String(dueDate).slice(0, 10) : '';
      return Number(debt.remaining_amount ?? debt.remainingAmount ?? 0) > 0 && dueDay && dueDay < todayDay;
    })
    .map((debt: any) => {
      const dueDate = debt.due_date || debt.dueDate;
      const dueDay = dueDate ? String(dueDate).slice(0, 10) : '';
      const daysOverdue = dueDay && dueDay < todayDay
        ? Math.floor((Date.parse(`${todayDay}T00:00:00`) - Date.parse(`${dueDay}T00:00:00`)) / (1000 * 60 * 60 * 24))
        : 0;
      return {
        id: String(debt.id),
        customer_name: debt.customer_name || debt.customer?.name || 'عميل',
        remaining_amount: Number(debt.remaining_amount ?? debt.remainingAmount ?? 0),
        due_date: dueDate || '',
        days_overdue: daysOverdue,
      };
    });
  const unpaidDebtorCount = new Set(
    debtRows
      .filter((debt: any) => {
        const dueDate = debt.due_date || debt.dueDate;
        const dueDay = dueDate ? String(dueDate).slice(0, 10) : '';
        return Number(debt.remaining_amount ?? debt.remainingAmount ?? 0) > 0 && dueDay && dueDay < todayDay;
      })
      .map((debt: any) => debt.customer_id || debt.customer?.id)
      .filter(Boolean)
  ).size;
  return (
    <div>
      {/* Page Header with Today's Summary */}
      <PageHeader
        title={t('dashboard.title')}
        description={t('dashboard.subtitle')}
      />

      {/* Priority 1: Attention Section - يحتاج انتباهك */}
      <div className="dashboard-s-flow" style={{ marginTop: 'var(--spacing-6)' }}>
        <AttentionSection
          lowStockCount={lowStockAlertCount}
          overdueDebtsCount={overdueDebtCount}
          lowStockItems={lowStockItems}
          overdueDebtItems={overdueDebtItems}
          unpaidDebtsCount={unpaidDebtorCount}
          unpaidDebtItems={unpaidDebtItems}
        />
        <SmartActions
          lowStockCount={lowStockAlertCount}
          overdueDebtsCount={unpaidDebtorCount}
        />
      </div>

      {/* Priority 3: Today's Performance - أداء اليوم */}
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <DashboardMetrics 
         stats={stats ? { ...stats, lowStockCount: lowStockAlertCount, activeCustomers: activeCustomerCount } : stats}
        />
      </div>

        {/* Secondary: Charts Grid - الأداء والتوزيع */}
        <div style={{
          marginTop: 'var(--spacing-6)',
          display: 'grid',
          gridTemplateColumns: isMobile ? '1fr' : 'minmax(0, 1.65fr) minmax(0, 0.9fr)',
          gap: 'var(--spacing-4)',
          alignItems: 'start'
        }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-4)' }}>
            <Card
              className="dashboard-performance-card"
              style={{ padding: 0 }}
            >
              <CardHeader className="dashboard-performance-header">
                <div className="flex w-full items-center justify-between gap-4 flex-wrap">
                  <div className="min-w-0">
                    <CardTitle>أداء المبيعات والأرباح</CardTitle>
                    <p className="mt-1 text-xs text-text-muted">مقارنة المبيعات مع الربح الإجمالي خلال الفترة المحددة</p>
                  </div>
                  <div className="dashboard-performance-range" role="group" aria-label="الفترة الزمنية للرسم البياني">
                    {([
                      { days: 1, label: 'يوم' },
                      { days: 7, label: '٧ أيام' },
                      { days: 30, label: '٣٠' },
                      { days: 90, label: '٩٠' },
                    ] as const).map(({ days, label }) => (
                      <Button
                        key={days}
                        type="button"
                        size="xs"
                        variant={chartRange === days ? 'primary' : 'ghost'}
                        className="dashboard-performance-range-button"
                        onClick={() => setChartRange(days)}
                      >
                        {label}
                      </Button>
                    ))}
                  </div>
                </div>
              </CardHeader>
              <CardContent className="dashboard-performance-content">
                  {chartData.length > 0 ? (
                    <SalesChart data={chartData} />
                ) : (
                  <div style={{ padding: 'var(--spacing-6)', textAlign: 'center' }}>
                    <p style={{ fontSize: 'var(--font-size-body)', color: 'var(--text-secondary)' }}>
                      لا توجد بيانات كافية للعرض
                    </p>
                  </div>
                )}
              </CardContent>
            </Card>

            <Card variant="open">
              <CardHeader>
                <div className="flex items-center justify-between gap-3">
                  <CardTitle className="flex items-center gap-2">
                    <Clock className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
                    {t('dashboard.recentActivity')}
                  </CardTitle>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="gap-1"
                    onClick={() => navigate('/app/activity')}
                  >
                    عرض كل النشاط
                    <ArrowLeft className="w-4 h-4" />
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
                  {recentActivities.length > 0 ? (
                    recentActivities.map((activity: any) => (
                      <ActivityItem
                        key={activity.id}
                        type={activity.type}
                        title={activity.title}
                        description={activity.description}
                        amount={activity.amount}
                        time={activity.type === 'sale'
                          ? formatStoreActivityDateTime(activity.time, activity.sale_date)
                          : formatDashboardActivityTime(activity.time)}
                        sellerName={activity.type === 'sale' ? activity.seller_name : undefined}
                        status={activity.status}
                      />
                    ))
                  ) : (
                    <p className="text-center text-text-muted py-4">
                      {t('dashboard.noRecentActivity')}
                    </p>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>

          <Card variant="open">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Package className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
                {t('dashboard.trends')}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {inventoryDistribution.length > 0 ? (
                <InventoryDistribution
                  totalValue={stats?.inventoryDistribution?.totalValue || fallbackDistribution.reduce((total: number, item: any) => total + item.value, 0)}
                  totalItems={stats?.inventoryDistribution?.totalItems || fallbackDistribution.reduce((total: number, item: any) => total + item.count, 0)}
                  data={inventoryDistribution as any}
                />
              ) : (
                <div style={{ padding: 'var(--spacing-6)', textAlign: 'center' }}>
                  <p style={{ fontSize: 'var(--font-size-body)', color: 'var(--text-secondary)' }}>
                    لا توجد بيانات كافية للعرض
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

    </div>
  );
}

// Activity Item Component
function ActivityItem({ type, title, description, amount, time, sellerName, status }: any) {
  const getIcon = () => {
    switch (type) {
      case 'sale': return ShoppingCart;
      case 'purchase': return DollarSign;
      case 'payment': return DollarSign;
      case 'return': return RotateCcw;
      default: return Activity;
    }
  };

  const Icon = getIcon();
  const statusLabels: Record<string, string> = {
    completed: 'مكتمل',
    pending: 'قيد الانتظار',
    reversed: 'إلغاء الشراء',
    cancelled: 'ملغي',
    received: 'مستلم',
    ordered: 'تم الطلب',
    draft: 'مسودة',
  };
  const normalizedStatus = String(status || '').toLowerCase();
  const statusLabel = statusLabels[normalizedStatus] || 'قيد المعالجة';

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--spacing-3)',
      padding: 'var(--spacing-4)',
      borderRadius: 'var(--radius-md)',
      border: '1px solid var(--border-default)',
      transition: '180ms ease'
    }}
    onMouseEnter={(e) => {
      e.currentTarget.style.background = 'var(--bg-surface-elevated)';
    }}
    onMouseLeave={(e) => {
      e.currentTarget.style.background = 'transparent';
    }}
    >
      <div className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'var(--color-primary-10)' }}>
        <Icon className="w-5 h-5" style={{ color: 'var(--color-primary)' }} />
      </div>
      <div style={{ flex: 1 }}>
        <p style={{ fontSize: 'var(--font-size-secondary)', fontWeight: 'var(--font-weight-medium)', color: 'var(--text-primary)' }}>{title}</p>
        <p style={{ fontSize: 'var(--font-size-caption)', color: 'var(--text-secondary)' }}>
          {type === 'sale' && sellerName ? `بواسطة: ${sellerName}` : description}
        </p>
      </div>
      <div style={{ textAlign: 'right' }}>
        <p style={{ fontSize: 'var(--font-size-secondary)', fontWeight: 'var(--font-weight-medium)', color: 'var(--text-primary)' }}>{amount}</p>
        <p style={{ fontSize: 'var(--font-size-caption)', color: 'var(--text-secondary)' }}>{time}</p>
      </div>
      <Badge variant={normalizedStatus === 'completed' ? 'success' : 'warning'} size="sm">
        {statusLabel}
      </Badge>
    </div>
  );
}
