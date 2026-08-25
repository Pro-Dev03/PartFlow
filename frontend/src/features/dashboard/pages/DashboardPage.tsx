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
import { DashboardMetrics } from '../components/DashboardMetrics';
import { AttentionSection } from '../components/AttentionSection';
import { SmartActions } from '../components/SmartActions';
import { AIInsight } from '../components/AIInsight';
import { SmartAlerts } from '../../../components/notifications/smart-alerts';
import { InventoryDistribution } from '../../../components/dashboard/InventoryDistribution';
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
    refetchInterval: 120000,
    staleTime: 60000,
  });

  const { data: notificationsData } = useQuery({
    queryKey: ['notifications'],
    queryFn: () => notificationsApi.list({ page: 1, per_page: 20 }),
    refetchInterval: 180000,
    staleTime: 120000,
  });

  const { data: unreadCountData } = useQuery({
    queryKey: ['notifications-unread'],
    queryFn: () => notificationsApi.getUnreadCount(),
    refetchInterval: 120000,
    staleTime: 60000,
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

      {/* Attention Section - لما يحتاج انتباهك */}
      <div style={{ marginTop: '16px' }}>
        <AttentionSection
          lowStockCount={stats?.lowStockCount as number}
          overdueDebtsCount={stats?.overdueDebts as number}
        />
      </div>

      {/* Smart Actions - العمليات اليومية */}
      <div style={{ marginTop: '16px' }}>
        <SmartActions />
      </div>

      {/* Key Metrics Cards - المؤشرات الرئيسية */}
      <div style={{ marginTop: '16px' }}>
        <DashboardMetrics stats={stats} />
      </div>

        {/* Charts Grid - الأداء والتوزيع */}
        <div style={{ 
          marginTop: '16px',
          display: 'grid', 
          gridTemplateColumns: isMobile ? '1fr' : 'minmax(0, 1.65fr) minmax(0, 0.9fr)', 
          gap: '16px' 
        }}>
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

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Package className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
                {t('dashboard.trends')}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <InventoryDistribution
                totalValue={125000}
                totalItems={342}
                data={[
                  { name: 'كروت شاشة', count: 85, value: 45000, color: '#0ea5e9', status: 'good' },
                  { name: 'معالجات', count: 62, value: 38000, color: '#10b981', status: 'good' },
                  { name: 'ذاكرة', count: 98, value: 22000, color: '#f59e0b', status: 'low' },
                  { name: 'تخزين', count: 45, value: 15000, color: '#ef4444', status: 'low' },
                  { name: 'أخرى', count: 52, value: 5000, color: '#64748b', status: 'critical' },
                ]}
              />
            </CardContent>
          </Card>
        </div>

      {/* Activity & Notifications - النشاط الأخير */}
      <div style={{ 
        marginTop: '16px',
        display: 'grid', 
        gridTemplateColumns: isMobile ? '1fr' : 'minmax(0, 1fr) minmax(0, 1fr)', 
        gap: '16px' 
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

        {/* Notifications */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
              {t('notifications.title')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div style={{ padding: '24px', textAlign: 'center' }}>
              <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--text-secondary)' }}>
                قيد التطوير
              </p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* AI Insights */}
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

      {/* Smart Alerts */}
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
      <div className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'var(--color-primary-10)' }}>
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
