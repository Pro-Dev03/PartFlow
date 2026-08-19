import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { useNavigate } from 'react-router-dom';
import { dashboardApi, notificationsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { Skeleton, StatCardSkeleton } from '../../../components/ui/skeleton';
import { SalesChart } from '../../../components/charts/SalesChart';
import { CategoryChart } from '../../../components/charts/CategoryChart';
import { LoadingSpinner } from '../../../components/ui/loading-spinner';
import {
  ShoppingCart,
  DollarSign,
  Package,
  AlertTriangle,
  Plus,
  ArrowUpRight,
  ArrowDownRight,
  Clock,
  Activity,
  AlertCircle,
  Shield,
  RotateCcw,
  Bell,
  Users,
  BarChart3,
  Sparkles,
  TrendingUp
} from 'lucide-react';

export function DashboardPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const { data: dashboardData, isLoading, error } = useQuery({
    queryKey: ['dashboard'],
    queryFn: () => dashboardApi.getStats(),
    refetchInterval: 30000,
    staleTime: 10000,
  });

  const { data: notificationsData } = useQuery({
    queryKey: ['notifications'],
    queryFn: () => notificationsApi.list(),
    refetchInterval: 60000,
    staleTime: 30000,
  });

  const { data: unreadCountData } = useQuery({
    queryKey: ['notifications-unread'],
    queryFn: () => notificationsApi.getUnreadCount(),
    refetchInterval: 30000,
    staleTime: 15000,
  });

  if (isLoading) {
    return (
      <div className="space-y-md">
        <PageHeader 
          eyebrow="لوحة التحكم"
          title="صباح الخير، صاحب المتجر"
          description="إليك نظرة عامة على أداء متجرك اليوم"
        />
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-md">
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
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <AlertTriangle className="w-12 h-12 text-danger mx-auto mb-2" />
          <p className="text-text-secondary">
            فشل تحميل البيانات
          </p>
        </div>
      </div>
    );
  }

  const stats = dashboardData as any;

  return (
    <div className="space-y-md">
      {/* Page Header with Actions - Futuristic + Visual */}
      <PageHeader
        eyebrow="Store Intelligence"
        title="صباح الخير، صاحب المتجر"
        description="إليك نظرة عامة على أداء متجرك اليوم"
        actions={
          <div className="flex items-center gap-sm">
            <Button variant="primary" onClick={() => navigate('/app/sales')} className="gap-2">
              <ShoppingCart className="w-4 h-4" />
              <span>بيع جديد</span>
            </Button>
            <Button variant="secondary" onClick={() => navigate('/app/inventory')} className="gap-2">
              <Plus className="w-4 h-4" />
              <span>إضافة منتج</span>
            </Button>
          </div>
        }
      />

      {/* Key Metrics Cards - Futuristic + Visual */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-md">
        <MetricCard
          title="مبيعات اليوم"
          value={`₪${(stats?.todaySales as number)?.toLocaleString() || 0}`}
          icon={ShoppingCart}
          trend="+12.4%"
          trendUp={true}
          subtitle="مقارنة بالأمس"
          variant="featured"
        />
        <MetricCard
          title="المعاملات"
          value={(stats?.todayTransactions as number) || 0}
          icon={Activity}
          trend="+8%"
          trendUp={true}
          subtitle="مقارنة بالأمس"
        />
        <MetricCard
          title="الديون المستحقة"
          value={`₪${(stats?.outstandingDebts as number)?.toLocaleString() || 0}`}
          icon={AlertTriangle}
          trend="+5%"
          trendUp={false}
          subtitle={`${(stats?.activeCustomers as number) || 0} عميل`}
          variant="warning"
        />
        <MetricCard
          title="المخزون المنخفض"
          value={(stats?.lowStockCount as number) || 0}
          icon={Package}
          trend={null}
          trendUp={null}
          subtitle="يحتاج انتباه"
          variant="danger"
        />
      </div>

      {/* AI Insights Card - Futuristic + Visual */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-cyan" />
            AI Insight
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-md">
            <div className="w-8 h-8 rounded-sm bg-cyan/10 flex items-center justify-center flex-shrink-0">
              <Sparkles className="w-4 h-4 text-cyan" />
            </div>
            <div>
              <p className="text-small font-semibold text-text">ارتفاع الطلب على RTX 3060</p>
              <p className="text-tiny text-text-muted mt-1">
                المبيعات زادت 34% هذا الشهر. المخزون الحالي قد يستمر لمدة 6 أيام تقريباً.
              </p>
              <Button variant="ghost" size="sm" className="mt-2 text-cyan">
                عرض التوصية ←
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Needs Attention Section - Futuristic + Visual */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertCircle className="w-5 h-5 text-yellow" />
            يحتاج انتباهك
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-sm">
            <AttentionItem
              icon={Package}
              title={`${(stats?.lowStockCount as number) || 0} منتج منخفض المخزون`}
              description="راجع المخزون الآن"
              action={() => navigate('/app/inventory')}
            />
            <AttentionItem
              icon={AlertTriangle}
              title={`${(stats?.overdueDebtsCount as number) || 0} ديون متأخرة`}
              description="راجع الديون"
              action={() => navigate('/app/debts')}
            />
            <AttentionItem
              icon={Shield}
              title={`${(stats?.expiringWarrantiesCount as number) || 0} ضمانات تنتهي قريباً`}
              description="راجع الضمانات"
              action={() => navigate('/app/warranties')}
            />
          </div>
        </CardContent>
      </Card>

      {/* Charts Section - Futuristic + Visual */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-md">
        <Card>
          <CardHeader>
            <CardTitle>نظرة عامة على المبيعات</CardTitle>
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
            <CardTitle>توزيع المبيعات حسب الفئة</CardTitle>
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

      {/* Recent Activity Section - Futuristic + Visual */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Clock className="w-5 h-5 text-cyan" />
            النشاط الأخير
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-sm">
            <ActivityItem
              type="sale"
              title="بيع #1234"
              description="مبيعات RTX 3060"
              amount="₪2,450"
              time="منذ 5 دقائق"
              status="completed"
            />
            <ActivityItem
              type="purchase"
              title="شراء #987"
              description="شراء كروت شاشة من المورد"
              amount="₪5,200"
              time="منذ 15 دقيقة"
              status="completed"
            />
            <ActivityItem
              type="payment"
              title="دفعة دين"
              description="دفعة من أحمد محمد"
              amount="₪1,000"
              time="منذ 30 دقيقة"
              status="completed"
            />
            <ActivityItem
              type="return"
              title="مرتجع #456"
              description="مرتجع منتج معيب"
              amount="₪850"
              time="منذ ساعة"
              status="completed"
            />
          </div>
        </CardContent>
      </Card>

      {/* Recent Notifications Section - Futuristic + Visual */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bell className="w-5 h-5 text-cyan" />
            الإشعارات الحديثة
            {unreadCountData && typeof unreadCountData === 'object' && 'data' in unreadCountData && (unreadCountData as any).data > 0 && (
              <Badge variant="danger" size="sm">
                {(unreadCountData as any).data}
              </Badge>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-sm">
            {(notificationsData?.data as any[])?.slice(0, 5).map((notification: any, index: number) => (
              <div
                key={index}
                className={`flex items-start gap-md p-lg rounded-sm border ${
                  notification.is_read
                    ? 'bg-surface border-border'
                    : 'bg-cyan/5 border-cyan/20'
                }`}
              >
                <div className="flex-shrink-0">
                  {getNotificationIcon(notification.type)}
                </div>
                <div className="flex-1">
                  <p className="text-small font-medium text-text">
                    {notification.title}
                  </p>
                  <p className="text-tiny text-text-muted mt-1">
                    {notification.message}
                  </p>
                  <p className="text-tiny text-text-muted mt-2">
                    {new Date(notification.created_at).toLocaleString('ar-SA')}
                  </p>
                </div>
              </div>
            ))}
            {(!notificationsData?.data || (notificationsData.data as any[]).length === 0) && (
              <p className="text-small text-text-muted text-center py-4">
                لا توجد إشعارات حديثة
              </p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Quick Actions - Futuristic + Visual */}
      <div>
        <h2 className="text-h3 font-semibold text-text mb-md">
          الإجراءات السريعة
        </h2>
        <div className="grid grid-cols-2 md:grid-cols-5 gap-sm">
          <QuickActionButton
            icon={ShoppingCart}
            label="بيع جديد"
            onClick={() => navigate('/app/sales')}
          />
          <QuickActionButton
            icon={Package}
            label="إضافة منتج"
            onClick={() => navigate('/app/inventory')}
          />
          <QuickActionButton
            icon={Users}
            label="إضافة عميل"
            onClick={() => navigate('/app/customers')}
          />
          <QuickActionButton
            icon={DollarSign}
            label="تسجيل دفعة"
            onClick={() => navigate('/app/debts')}
          />
          <QuickActionButton
            icon={BarChart3}
            label="التقارير"
            onClick={() => navigate('/app/reports')}
          />
        </div>
      </div>
    </div>
  );
}

// Unified Metric Card Component - Following Design System
function MetricCard({ title, value, icon: Icon, trend, trendUp, subtitle, variant = 'default' }: any) {
  return (
    <Card variant={variant} className="hover:border-border/22 hover:-translate-y-1 cursor-pointer" hoverable>
      <CardContent className="p-lg">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <p className="text-small text-text-muted mb-1">{title}</p>
            <p className="text-metric font-bold text-text mb-1">{value}</p>
            {subtitle && (
              <p className="text-tiny text-text-muted">{subtitle}</p>
            )}
          </div>
          <div className={`w-10 h-10 rounded-sm flex items-center justify-center ${
            variant === 'featured' ? 'bg-cyan/10' :
            variant === 'warning' ? 'bg-yellow/10' :
            variant === 'danger' ? 'bg-red/10' :
            variant === 'success' ? 'bg-green/10' :
            variant === 'info' ? 'bg-cyan/10' :
            'bg-cyan/10'
          }`}>
            <Icon className={`w-5 h-5 ${
              variant === 'featured' ? 'text-cyan' :
              variant === 'warning' ? 'text-yellow' :
              variant === 'danger' ? 'text-red' :
              variant === 'success' ? 'text-green' :
              variant === 'info' ? 'text-cyan' :
              'text-cyan'
            }`} />
          </div>
        </div>
        {trend && (
          <div className="flex items-center gap-1 mt-3">
            {trendUp ? (
              <ArrowUpRight className="w-4 h-4 text-green" />
            ) : (
              <ArrowDownRight className="w-4 h-4 text-red" />
            )}
            <span className={`text-small ${trendUp ? 'text-green' : 'text-red'}`}>
              {trend}
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// Attention Item Component - Following Design System
function AttentionItem({ icon: Icon, title, description, action }: any) {
  return (
    <div 
      className="flex items-center gap-md p-lg rounded-sm bg-surface hover:bg-surface-2 cursor-pointer transition-colors duration-normal" 
      onClick={action}
    >
      <div className="w-10 h-10 rounded-sm bg-yellow/10 flex items-center justify-center flex-shrink-0">
        <Icon className="w-5 h-5 text-yellow" />
      </div>
      <div className="flex-1">
        <p className="text-small font-medium text-text">{title}</p>
        <p className="text-tiny text-text-muted">{description}</p>
      </div>
      <ArrowUpRight className="w-4 h-4 text-text-muted" />
    </div>
  );
}

// Activity Item Component - Following Design System
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
    <div className="flex items-center gap-md p-lg rounded-sm border border-border hover:bg-surface-2 transition-colors duration-normal">
      <div className="w-10 h-10 rounded-sm bg-cyan/10 flex items-center justify-center flex-shrink-0">
        <Icon className="w-5 h-5 text-cyan" />
      </div>
      <div className="flex-1">
        <p className="text-small font-medium text-text">{title}</p>
        <p className="text-tiny text-text-muted">{description}</p>
      </div>
      <div className="text-right">
        <p className="text-small font-medium text-text">{amount}</p>
        <p className="text-tiny text-text-muted">{time}</p>
      </div>
      <Badge variant={status === 'completed' ? 'success' : 'warning'} size="sm">
        {status === 'completed' ? 'مكتمل' : status}
      </Badge>
    </div>
  );
}

// Quick Action Button Component - Following Design System
function QuickActionButton({ icon: Icon, label, onClick }: any) {
  return (
    <button
      onClick={onClick}
      className="flex flex-col items-center gap-sm p-lg rounded-sm border border-border hover:bg-surface-2 hover:border-cyan/30 transition-all duration-normal"
    >
      <div className="w-10 h-10 rounded-sm bg-cyan/10 flex items-center justify-center">
        <Icon className="w-5 h-5 text-cyan" />
      </div>
      <span className="text-tiny text-text">{label}</span>
    </button>
  );
}

// Notification Icon Helper
function getNotificationIcon(type: string) {
  switch (type) {
    case 'debt':
      return <AlertTriangle className="w-5 h-5 text-yellow" />;
    case 'inventory':
      return <Package className="w-5 h-5 text-cyan" />;
    case 'warranty':
      return <Shield className="w-5 h-5 text-cyan" />;
    default:
      return <Bell className="w-5 h-5 text-cyan" />;
  }
}