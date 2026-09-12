import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { useNavigate } from 'react-router-dom';
import { useIsMobile } from '../../../hooks/useIsMobile';
import { cn } from '../../../utils';
import { dashboardApi, debtsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { StatCardSkeleton } from '../../../components/ui/skeleton';
import { SalesChart } from '../../../components/charts/SalesChart';
import { DashboardMetrics } from '../components/DashboardMetrics';
import { AttentionSection } from '../components/AttentionSection';
import { SmartActions } from '../components/SmartActions';
import { InventoryDistribution } from '../../../components/dashboard/InventoryDistribution';
import { getButtonSize } from '../../../config/button-sizes';
import { DashboardStats } from '../../../types/api';
import {
  ShoppingCart,
  DollarSign,
  AlertTriangle,
  Clock,
  Activity,
  RotateCcw,
  Plus,
  Package,
  TrendingUp,
} from 'lucide-react';

function getWelcomeKey() {
  const hour = new Date().getHours();

  if (hour >= 5 && hour < 12) return 'dashboard.welcomeMorning';
  if (hour >= 17 && hour < 22) return 'dashboard.welcomeEvening';
  return 'dashboard.welcomeNight';
}

export function DashboardPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isMobile = useIsMobile();
  const welcomeMessage = t(getWelcomeKey());
  const { data: dashboardData, isLoading, error } = useQuery({
    queryKey: ['dashboard'],
    queryFn: () => dashboardApi.getStats(),
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
          eyebrow={t('dashboard.title')}
          title={welcomeMessage}
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
  const lowStockItems = lowStockItemsData?.data || [];
  const overdueDebtItems = overdueDebtsData?.data || [];
  const debtRows = Array.isArray(debtsData?.data) ? debtsData.data : [];
  const unpaidDebtItems = debtRows
    .filter((debt: any) => Number(debt.remaining_amount ?? debt.remainingAmount ?? 0) > 0)
    .map((debt: any) => {
      const dueDate = debt.due_date || debt.dueDate;
      const dueDay = dueDate ? String(dueDate).slice(0, 10) : '';
      const today = new Date();
      const todayDay = [today.getFullYear(), today.getMonth() + 1, today.getDate()]
        .map((part) => String(part).padStart(2, '0'))
        .join('-');
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
      .filter((debt: any) => Number(debt.remaining_amount ?? debt.remainingAmount ?? 0) > 0)
      .map((debt: any) => debt.customer_id || debt.customer?.id)
      .filter(Boolean)
  ).size;
  return (
    <div>
      {/* Page Header with Today's Summary */}
      <PageHeader
        eyebrow={t('dashboard.title')}
        title={welcomeMessage}
        description={t('dashboard.subtitle')}
        actions={
          <div className={cn(
            "flex gap-3",
            isMobile ? "flex-col w-full" : ""
          )}>
            <Button 
              variant="secondary" 
              size={getButtonSize('customers', 'headerActions')} 
              onClick={() => navigate('/app/sales')} 
              className={cn("gap-2", isMobile ? "w-full" : "")}
            >
              <ShoppingCart className="w-4 h-4" />
              <span>{t('dashboard.newSale')}</span>
            </Button>
            <Button 
              variant="secondary" 
              size={getButtonSize('customers', 'headerActions')} 
              onClick={() => navigate('/app/inventory')} 
              className={cn("gap-2", isMobile ? "w-full" : "")}
            >
              <Plus className="w-4 h-4" />
              <span>{t('dashboard.addProduct')}</span>
            </Button>
          </div>
        }
      />

      {/* Priority 1: Attention Section - يحتاج انتباهك */}
      <div className="dashboard-s-flow" style={{ marginTop: 'var(--spacing-6)' }}>
        <AttentionSection
          lowStockCount={stats?.lowStockCount as number}
          overdueDebtsCount={stats?.overdueDebts as number}
          lowStockItems={lowStockItems}
          overdueDebtItems={overdueDebtItems}
          unpaidDebtsCount={unpaidDebtorCount}
          unpaidDebtItems={unpaidDebtItems}
        />
        <SmartActions
          lowStockCount={stats?.lowStockCount as number}
          overdueDebtsCount={unpaidDebtorCount}
        />
      </div>

      {/* Priority 3: Today's Performance - أداء اليوم */}
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <DashboardMetrics 
         stats={stats}
        />
      </div>

        {/* Secondary: Charts Grid - الأداء والتوزيع */}
        <div style={{ 
          marginTop: 'var(--spacing-6)',
          display: 'grid', 
          gridTemplateColumns: isMobile ? '1fr' : 'minmax(0, 1.65fr) minmax(0, 0.9fr)', 
          gap: 'var(--spacing-4)' 
        }}>
          <Card variant="open">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <TrendingUp className="w-5 h-5" style={{ color: 'var(--color-success)' }} />
                {t('dashboard.performance')}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {stats?.salesChart && stats.salesChart.length > 0 ? (
                <SalesChart data={stats.salesChart} />
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
              <CardTitle className="flex items-center gap-2">
                <Package className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
                {t('dashboard.trends')}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {stats?.inventoryDistribution && stats.inventoryDistribution.data?.length > 0 ? (
                <InventoryDistribution
                  totalValue={stats.inventoryDistribution.totalValue}
                  totalItems={stats.inventoryDistribution.totalItems}
                  data={stats.inventoryDistribution.data}
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

      {/* Secondary: Activity & Notifications - النشاط الأخير */}
      <div style={{ 
        marginTop: 'var(--spacing-6)',
        display: 'grid', 
        gridTemplateColumns: isMobile ? '1fr' : 'minmax(0, 1fr) minmax(0, 1fr)', 
        gap: 'var(--spacing-4)' 
      }}>
        {/* Recent Activity */}
        <Card variant="open">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
              {t('dashboard.recentActivity')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
              {dashboardData?.data?.recent_activity?.length > 0 ? (
                dashboardData?.data?.recent_activity?.map((activity: any) => (
                  <ActivityItem
                    key={activity.id}
                    type={activity.type}
                    title={activity.title}
                    description={activity.description}
                    amount={activity.amount}
                    time={activity.time}
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

    </div>
  );
}

// Activity Item Component
function ActivityItem({ type, title, description, amount, time, status }: any) {
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
        <p style={{ fontSize: 'var(--font-size-caption)', color: 'var(--text-secondary)' }}>{description}</p>
      </div>
      <div style={{ textAlign: 'right' }}>
        <p style={{ fontSize: 'var(--font-size-secondary)', fontWeight: 'var(--font-weight-medium)', color: 'var(--text-primary)' }}>{amount}</p>
        <p style={{ fontSize: 'var(--font-size-caption)', color: 'var(--text-secondary)' }}>{time}</p>
      </div>
      <Badge variant={status === 'completed' ? 'success' : 'warning'} size="sm">
        {status === 'completed' ? 'مكتمل' : status}
      </Badge>
    </div>
  );
}
