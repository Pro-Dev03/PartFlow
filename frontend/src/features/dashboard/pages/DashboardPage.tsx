import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { useNavigate } from 'react-router-dom';
import { useIsMobile } from '../../../hooks/useIsMobile';
import { cn } from '../../../utils';
import { dashboardApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { StatCardSkeleton } from '../../../components/ui/skeleton';
import { SalesChart } from '../../../components/charts/SalesChart';
import { DashboardMetrics } from '../components/DashboardMetrics';
import { AttentionSection } from '../components/AttentionSection';
import { SmartActions } from '../components/SmartActions';
import { AIInsight } from '../components/AIInsight';
import { SmartAlerts } from '../../../components/notifications/smart-alerts';
import { InventoryDistribution } from '../../../components/dashboard/InventoryDistribution';
import { getButtonSize } from '../../../config/button-sizes';
import { DashboardStats } from '../../../types/api';
import {
  ShoppingCart,
  DollarSign,
  AlertTriangle,
  Clock,
  Activity,
  Bell,
  RotateCcw,
  Plus,
  Package,
  TrendingUp,
  Sparkles,
} from 'lucide-react';

export function DashboardPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isMobile = useIsMobile();

  const { data: dashboardData, isLoading, error } = useQuery({
    queryKey: ['dashboard'],
    queryFn: () => dashboardApi.getStats(),
    refetchInterval: 120000,
    staleTime: 60000,
  });

  // Use aggregation tables for faster dashboard (ARCHITECTURE-PRINCIPLES.md)
  const { data: dailySalesData } = useQuery({
    queryKey: ['daily-sales-summary'],
    queryFn: () => dashboardApi.getDailySalesSummary(),
    refetchInterval: 300000, // 5 minutes
    staleTime: 180000, // 3 minutes
  });

  const { data: dailyInventoryData } = useQuery({
    queryKey: ['daily-inventory-summary'],
    queryFn: () => dashboardApi.getDailyInventorySummary(),
    refetchInterval: 300000,
    staleTime: 180000,
  });

  const { data: dailyDebtData } = useQuery({
    queryKey: ['daily-debt-summary'],
    queryFn: () => dashboardApi.getDailyDebtSummary(),
    refetchInterval: 300000,
    staleTime: 180000,
  });

  const { data: dailyProfitData } = useQuery({
    queryKey: ['daily-profit-summary'],
    queryFn: () => dashboardApi.getDailyProfitSummary(),
    refetchInterval: 300000,
    staleTime: 180000,
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

  if (isLoading) {
    return (
      <div>
        <PageHeader
          eyebrow={t('dashboard.title')}
          title={t('dashboard.welcome')}
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

  const stats = dashboardData as DashboardStats;

  return (
    <div>
      {/* Page Header with Today's Summary */}
      <PageHeader
        eyebrow={t('dashboard.title')}
        title={t('dashboard.welcome')}
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
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <AttentionSection
          lowStockCount={stats?.lowStockCount as number}
          overdueDebtsCount={stats?.overdueDebtsCount as number}
          lowStockItems={lowStockItemsData?.data || []}
          overdueDebtItems={overdueDebtsData?.data || []}
        />
      </div>

      {/* Priority 2: Smart Actions - العمليات اليومية */}
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <SmartActions 
          lowStockCount={stats?.lowStockCount as number}
          overdueDebtsCount={stats?.overdueDebtsCount as number}
        />
      </div>

      {/* Priority 3: Today's Performance - أداء اليوم */}
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <DashboardMetrics 
          stats={stats}
          dailySales={dailySalesData?.data}
          dailyInventory={dailyInventoryData?.data}
          dailyDebt={dailyDebtData?.data}
          dailyProfit={dailyProfitData?.data}
        />
      </div>

      {/* Priority 4: AI Insights - الرؤى الذكية */}
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <AIInsight />
      </div>

        {/* Secondary: Charts Grid - الأداء والتوزيع */}
        <div style={{ 
          marginTop: 'var(--spacing-6)',
          display: 'grid', 
          gridTemplateColumns: isMobile ? '1fr' : 'minmax(0, 1.65fr) minmax(0, 0.9fr)', 
          gap: 'var(--spacing-4)' 
        }}>
          <Card>
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

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Package className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
                {t('dashboard.trends')}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {stats?.inventoryDistribution ? (
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
        <Card>
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

        {/* Notifications */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
              {t('notifications.title')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div style={{ padding: 'var(--spacing-6)', textAlign: 'center' }}>
              <p style={{ fontSize: 'var(--font-size-body)', fontWeight: 'var(--font-weight-semibold)', color: 'var(--text-secondary)' }}>
                قيد التطوير
              </p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tertiary: Smart Alerts */}
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <Card style={{
          background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.05) 0%, rgba(236, 72, 153, 0.05) 100%)',
          border: '1px solid rgba(99, 102, 241, 0.2)'
        }}>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Sparkles className="w-5 h-5" style={{ color: 'var(--color-primary)' }} />
              {t('notifications.smartAlerts')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <SmartAlerts alerts={[]} />
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
