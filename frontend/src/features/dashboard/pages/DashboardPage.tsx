import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import { cn } from '../../../utils';
import { dashboardApi, notificationsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { StatCardSkeleton } from '../../../components/ui/skeleton';
import { SalesChart } from '../../../components/charts/SalesChart';
import { CategoryChart } from '../../../components/charts/CategoryChart';
import { DashboardMetrics } from '../components/DashboardMetrics';
import { AttentionSection } from '../components/AttentionSection';
import { SmartActions } from '../components/SmartActions';
import { AIInsight } from '../components/AIInsight';
import { SmartAlerts } from '../../../components/notifications/smart-alerts';
import { getButtonSize } from '../../../config/button-sizes';
import {
  ShoppingCart,
  DollarSign,
  AlertTriangle,
  Clock,
  Activity,
  Bell,
  ArrowUpRight,
  RotateCcw,
  Plus,
  Package,
  Shield,
  TrendingUp,
  Target,
  Zap,
  XCircle,
  Calendar,
  Users,
  BarChart3,
  Sparkles,
  Sun,
  Moon,
  Coffee
} from 'lucide-react';

export function DashboardPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [isMobile, setIsMobile] = useState(false);
  
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768);
    };
    
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  const { data: dashboardData, isLoading, error } = useQuery({
    queryKey: ['dashboard'],
    queryFn: () => dashboardApi.getStats(),
    refetchInterval: 120000, // Reduced from 30s to 2min
    staleTime: 60000, // Increased from 10s to 1min
  });

  const { data: notificationsData } = useQuery({
    queryKey: ['notifications'],
    queryFn: () => notificationsApi.list({ page: 1, per_page: 20 }),
    refetchInterval: 180000, // Reduced from 60s to 3min
    staleTime: 120000, // Increased from 30s to 2min
  });

  const { data: unreadCountData } = useQuery({
    queryKey: ['notifications-unread'],
    queryFn: () => notificationsApi.getUnreadCount(),
    refetchInterval: 120000, // Reduced from 30s to 2min
    staleTime: 60000, // Increased from 15s to 1min
  });

  if (isLoading) {
    return (
      <div>
        <PageHeader
          eyebrow={t('dashboard.title')}
          title={t('dashboard.welcome')}
          description={t('dashboard.subtitle')}
        />
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '14px' }}>
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

  const stats = dashboardData as any;

  return (
    <div>
      {/* Page Header with Actions */}
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

      {/* "يحتاج انتباهك" Section - أهم قسم في Dashboard */}
      <div style={{ marginTop: '16px' }}>
        <AttentionSection
          lowStockCount={stats?.lowStockCount as number}
          overdueDebtsCount={stats?.overdueDebts as number}
        />
      </div>

      {/* Smart Actions - العمليات اليومية الأكثر استخداماً */}
      <div style={{ marginTop: '16px' }}>
        <SmartActions />
      </div>

      {/* Key Metrics Cards */}
      <DashboardMetrics stats={stats} />

      {/* Content Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : 'minmax(0, 1.65fr) minmax(0, 0.9fr)', gap: '16px' }}
           className="grid-cols-1 md:grid-cols-2">
        {/* Sales Chart */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="w-5 h-5" style={{ color: 'var(--color-success)' }} />
              {t('dashboard.performance')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <SalesChart data={[
              { name: 'السبت', sales: 4500, profit: 1200 },
              { name: 'الأحد', sales: 5200, profit: 1500 },
              { name: 'الاثنين', sales: 4800, profit: 1300 },
              { name: 'الثلاثاء', sales: 6100, profit: 1800 },
              { name: 'الأربعاء', sales: 5900, profit: 1700 },
              { name: 'الخميس', sales: 7200, profit: 2100 },
              { name: 'الجمعة', sales: 6800, profit: 1900 },
            ]} />
          </CardContent>
        </Card>

        {/* Category Chart */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Package className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
              {t('dashboard.trends')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <CategoryChart data={[
              { name: 'كروت شاشة', value: 35, color: '#38bdf8' },
              { name: 'معالجات', value: 25, color: '#34d399' },
              { name: 'ذاكرة', value: 20, color: '#fbbf24' },
              { name: 'تخزين', value: 15, color: '#fb7185' },
              { name: 'أخرى', value: 5, color: '#a78bfa' },
            ]} />
          </CardContent>
        </Card>
      </div>

      {/* AI Insights Card */}
      {dashboardData?.data?.insights && dashboardData.data.insights.length > 0 && (
        <div style={{ marginTop: '16px' }}>
          {dashboardData.data.insights.map((insight: any, index: number) => (
            <AIInsight
              key={index}
              title={insight.title}
              description={insight.description}
              actionLabel={t('dashboard.viewDetails')}
              onAction={() => navigate(insight.link || '/app/inventory')}
            />
          ))}
        </div>
      )}

      {/* Smart Alerts Section - التنبيهات الذكية */}
      <div style={{ marginTop: '16px' }}>
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

      {/* Recent Activity Section */}
      <div style={{ marginTop: '16px' }}>
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
              {t('dashboard.recentActivity')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
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

      {/* Recent Notifications Section */}
      <div style={{ marginTop: '16px' }}>
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
              {t('notifications.title')}
              {unreadCountData && typeof unreadCountData === 'object' && 'data' in unreadCountData && (unreadCountData as any).data > 0 && (
                <Badge variant="danger" className="text-xs">
                  {(unreadCountData as any).data}
                </Badge>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
              {(notificationsData?.data as any[])?.slice(0, 5).map((notification: any, index: number) => (
                <div
                  key={index}
                  style={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    gap: '14px',
                    padding: '18px',
                    borderRadius: '10px',
                    border: notification.is_read
                      ? '1px solid var(--border-default)'
                      : '1px solid var(--border-primary)',
                    background: notification.is_read
                      ? 'var(--bg-surface)'
                      : 'rgba(99, 102, 241, 0.05)'
                  }}
                >
                  <div style={{ flexShrink: 0 }}>
                    {getNotificationIcon(notification.type)}
                  </div>
                  <div style={{ flex: 1 }}>
                    <p style={{ fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>
                      {notification.title}
                    </p>
                    <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                      {notification.message}
                    </p>
                    <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '8px' }}>
                      {new Date(notification.created_at).toLocaleString('ar-SA')}
                    </p>
                  </div>
                </div>
              ))}
              {(!notificationsData?.data || (notificationsData.data as any[]).length === 0) && (
                <p style={{ fontSize: '13px', color: 'var(--text-secondary)', textAlign: 'center', padding: '16px' }}>
                  {t('notifications.noNotifications')}
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
      gap: '14px',
      padding: '18px',
      borderRadius: '10px',
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
      <div className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'rgba(99, 102, 241, 0.1)' }}>
        <Icon className="w-5 h-5" style={{ color: 'var(--color-primary)' }} />
      </div>
      <div style={{ flex: 1 }}>
        <p style={{ fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>{title}</p>
        <p style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>{description}</p>
      </div>
      <div style={{ textAlign: 'right' }}>
        <p style={{ fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>{amount}</p>
        <p style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>{time}</p>
      </div>
      <Badge variant={status === 'completed' ? 'success' : 'warning'} className="text-xs">
        {status === 'completed' ? 'مكتمل' : status}
      </Badge>
    </div>
  );
}

// Notification Icon Helper
function getNotificationIcon(type: string) {
  switch (type) {
    case 'debt':
      return <AlertTriangle className="w-5 h-5" style={{ color: 'var(--color-warning)' }} />;
    case 'inventory':
      return <Package className="w-5 h-5" style={{ color: 'var(--color-info)' }} />;
    default:
      return <Bell className="w-5 h-5" style={{ color: 'var(--color-info)' }} />;
  }
}


