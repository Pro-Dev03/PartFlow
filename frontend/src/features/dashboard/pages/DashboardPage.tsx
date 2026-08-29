import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { useNavigate } from 'react-router-dom';
import { useIsMobile } from '../../../hooks/useIsMobile';
import { cn } from '../../../utils';
import { dashboardApi, notificationsApi } from '../../../services/api/endpoints';
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

  const { data: notificationsData } = useQuery({
    queryKey: ['notifications', 'dashboard'],
    queryFn: () => notificationsApi.list(),
    staleTime: 30000,
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
  const notifications = Array.isArray(notificationsData?.data) ? notificationsData.data : [];
  const insight = lowStockItems.length > 0
    ? {
        title: 'مخزون يحتاج إعادة طلب',
        description: `يوجد ${lowStockItems.length} منتجًا تحت الحد الأدنى. ابدأ بإعادة التوريد قبل نفاد المخزون.`,
        actionLabel: 'مراجعة المخزون',
        onAction: () => navigate('/app/inventory'),
      }
    : overdueDebtItems.length > 0
      ? {
          title: 'دفعات متأخرة تحتاج متابعة',
          description: `يوجد ${overdueDebtItems.length} دينًا متأخرًا. متابعة التحصيل الآن تحافظ على السيولة.`,
          actionLabel: 'مراجعة الديون',
          onAction: () => navigate('/app/debts'),
        }
      : {
          title: 'الوضع مستقر',
          description: 'لا توجد تنبيهات مخزون أو ديون متأخرة حاليًا وفق آخر بيانات النظام.',
          actionLabel: 'عرض التقارير',
          onAction: () => navigate('/app/reports'),
        };

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
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <AttentionSection
          lowStockCount={stats?.lowStockCount as number}
          overdueDebtsCount={stats?.overdueDebts as number}
          lowStockItems={lowStockItems}
          overdueDebtItems={overdueDebtItems}
        />
      </div>

      {/* Priority 2: Smart Actions - العمليات اليومية */}
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <SmartActions 
          lowStockCount={stats?.lowStockCount as number}
          overdueDebtsCount={stats?.overdueDebts as number}
        />
      </div>

      {/* Priority 3: Today's Performance - أداء اليوم */}
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <DashboardMetrics 
         stats={stats}
        />
      </div>

      {/* Priority 4: AI Insights - الرؤى الذكية */}
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <AIInsight {...insight} />
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
            {notifications.length > 0 ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
                {notifications.slice(0, 5).map((notification: any) => (
                  <div key={notification.id} className="rounded-lg border border-border p-3">
                    <p className="text-sm font-medium text-text-primary">{notification.title || notification.message}</p>
                    {notification.title && notification.message && (
                      <p className="mt-1 text-xs text-text-secondary">{notification.message}</p>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <p className="py-4 text-center text-text-muted">لا توجد إشعارات فعلية حاليًا</p>
            )}
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
            <SmartAlerts alerts={[
              ...(lowStockItems.length > 0 ? [{
                id: 'dashboard-low-stock',
                type: 'warning' as const,
                title: 'مخزون منخفض',
                message: `${lowStockItems.length} منتجًا يحتاج إلى إعادة طلب.`,
                priority: 'high' as const,
                actionLabel: 'فتح المخزون',
                onAction: () => navigate('/app/inventory'),
              }] : []),
              ...(overdueDebtItems.length > 0 ? [{
                id: 'dashboard-overdue-debts',
                type: 'warning' as const,
                title: 'ديون متأخرة',
                message: `${overdueDebtItems.length} عميلًا لديه دفعة متأخرة.`,
                priority: 'urgent' as const,
                actionLabel: 'فتح الديون',
                onAction: () => navigate('/app/debts'),
              }] : []),
            ]} />
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
