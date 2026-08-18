import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { useNavigate } from 'react-router-dom';
import { dashboardApi, notificationsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Skeleton, StatCardSkeleton } from '../../../components/ui/skeleton';
import { SalesChart } from '../../../components/charts/SalesChart';
import { CategoryChart } from '../../../components/charts/CategoryChart';
import {
  ShoppingCart,
  DollarSign,
  Package,
  AlertTriangle,
  Plus,
  TrendingUp,
  ArrowUpRight,
  ArrowDownRight,
  Users,
  Clock,
  BarChart3,
  Lightbulb,
  Activity,
  Zap,
  Bell
} from 'lucide-react';

export function DashboardPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const { data: dashboardData, isLoading, error } = useQuery({
    queryKey: ['dashboard'],
    queryFn: () => dashboardApi.getStats(),
    refetchInterval: 30000, // Refetch every 30 seconds
    staleTime: 10000, // Consider data fresh for 10 seconds
  });

  const { data: notificationsData } = useQuery({
    queryKey: ['notifications'],
    queryFn: () => notificationsApi.list(),
    refetchInterval: 60000, // Refetch every minute
    staleTime: 30000, // Consider data fresh for 30 seconds
  });

  const { data: unreadCountData } = useQuery({
    queryKey: ['notifications-unread'],
    queryFn: () => notificationsApi.getUnreadCount(),
    refetchInterval: 30000, // Refetch every 30 seconds
    staleTime: 15000, // Consider data fresh for 15 seconds
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <Skeleton className="h-8 w-48 mb-2" />
          <Skeleton className="h-4 w-64" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
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
          <AlertTriangle className="w-12 h-12 text-red-500 mx-auto mb-2" />
          <p className="text-gray-500 dark:text-gray-400">
            فشل تحميل البيانات
          </p>
        </div>
      </div>
    );
  }

  const stats = dashboardData as any;

  return (
    <div className="space-y-6">
      {/* Welcome Section */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
          {t('dashboard.welcome')}
        </h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">
          {t('dashboard.subtitle')}
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title={t('dashboard.todaySales')}
          value={`₪${(stats?.todaySales as number)?.toLocaleString() || 0}`}
          icon={ShoppingCart}
          trend="+12%"
          trendUp={true}
          subtitle="vs yesterday"
        />
        <StatCard
          title={t('dashboard.todayProfit')}
          value={`₪${(stats?.todayProfit as number)?.toLocaleString() || 0}`}
          icon={DollarSign}
          trend="+8%"
          trendUp={true}
          subtitle="vs yesterday"
        />
        <StatCard
          title={t('dashboard.inventoryValue')}
          value={`₪${(stats?.inventoryValue as number)?.toLocaleString() || 0}`}
          icon={Package}
          trend="-2%"
          trendUp={false}
          subtitle="total stock"
        />
        <StatCard
          title={t('dashboard.outstandingDebts')}
          value={`₪${(stats?.outstandingDebts as number)?.toLocaleString() || 0}`}
          icon={AlertTriangle}
          trend="+5%"
          trendUp={false}
          subtitle="customer debts"
        />
      </div>

      {/* Additional KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <KPICard
          title={t('dashboard.activeCustomers')}
          value={(stats?.activeCustomers as number) || 0}
          icon={Users}
          color="blue"
          trend="+3%"
          trendUp={true}
        />
        <KPICard
          title={t('dashboard.todayTransactions')}
          value={(stats?.todayTransactions as number) || 0}
          icon={Activity}
          color="green"
          trend="+15%"
          trendUp={true}
        />
        <KPICard
          title={t('dashboard.averageSale')}
          value={`₪${((stats?.todaySales as number) / (stats?.todayTransactions as number) || 0).toFixed(0)}`}
          icon={BarChart3}
          color="purple"
          trend="+5%"
          trendUp={true}
        />
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>المبيعات والأرباح</CardTitle>
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
              { name: 'كروت شاشة', value: 35, color: '#3b82f6' },
              { name: 'معالجات', value: 25, color: '#10b981' },
              { name: 'ذاكرة', value: 20, color: '#f59e0b' },
              { name: 'تخزين', value: 15, color: '#ef4444' },
              { name: 'أخرى', value: 5, color: '#8b5cf6' },
            ]} />
          </CardContent>
        </Card>
      </div>

      {/* Smart Insights Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Lightbulb className="w-5 h-5 text-yellow-500" />
            {t('dashboard.insights')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {(stats?.insights as any[])?.map((insight: any, index: number) => (
              <InsightItem key={index} insight={insight} />
            ))}
            {(!stats?.insights || (stats.insights as any[]).length === 0) && (
              <div className="flex items-center gap-3 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <Zap className="w-5 h-5 text-yellow-500" />
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {t('common.noData')}
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    تظهر الأفكار الذكية هنا عندما يكون هناك بيانات كافية
                  </p>
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Debt Alerts Section - Integration with Debt Worker */}
      {(stats?.alerts as any[])?.filter((alert: any) => alert.type === 'OVERDUE_DEBT').length > 0 && (
        <Card className="border-red-200 dark:border-red-800">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-red-600 dark:text-red-400">
              <AlertTriangle className="w-5 h-5" />
              تنبيهات الديون المتأخرة
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {(stats?.alerts as any[])?.filter((alert: any) => alert.type === 'OVERDUE_DEBT').map((alert: any, index: number) => (
                <AlertItem 
                  key={index} 
                  alert={alert} 
                  onAction={() => alert.actionUrl && navigate(alert.actionUrl)}
                />
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Alerts Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertTriangle className="w-5 h-5 text-orange-500" />
            {t('dashboard.alerts')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {(stats?.alerts as any[])?.map((alert: any, index: number) => (
              <AlertItem
                key={index}
                alert={alert}
                onAction={() => alert.actionUrl && navigate(alert.actionUrl)}
              />
            ))}
            {(!stats?.alerts || (stats.alerts as any[]).length === 0) && (
              <p className="text-sm text-gray-500 dark:text-gray-400">
                لا توجد تنبيهات حالياً
              </p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Recent Notifications Section - Integration with Worker Service */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bell className="w-5 h-5 text-blue-500" />
            الإشعارات الحديثة
            {unreadCountData && typeof unreadCountData === 'object' && 'data' in unreadCountData && (unreadCountData as any).data > 0 && (
              <span className="ml-2 px-2 py-1 text-xs bg-red-500 text-white rounded-full">
                {(unreadCountData as any).data}
              </span>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {(notificationsData?.data as any[])?.slice(0, 5).map((notification: any, index: number) => (
              <div
                key={index}
                className={`flex items-start gap-3 p-3 rounded-lg ${
                  notification.is_read
                    ? 'bg-gray-50 dark:bg-gray-800'
                    : 'bg-blue-50 dark:bg-blue-900/10 border-l-4 border-blue-500'
                }`}
              >
                <div className="flex-shrink-0">
                  {getNotificationIcon(notification.type)}
                </div>
                <div className="flex-1">
                  <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {notification.title}
                  </p>
                  <p className="text-xs text-gray-600 dark:text-gray-400 mt-1">
                    {notification.message}
                  </p>
                  <p className="text-xs text-gray-400 dark:text-gray-500 mt-2">
                    {new Date(notification.created_at).toLocaleString('ar-SA')}
                  </p>
                </div>
              </div>
            ))}
            {(!notificationsData?.data || (notificationsData.data as any[]).length === 0) && (
              <p className="text-sm text-gray-500 dark:text-gray-400">
                لا توجد إشعارات حديثة
              </p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Quick Actions */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
          {t('dashboard.quickActions')}
        </h2>
        <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
          <QuickActionButton
            icon={ShoppingCart}
            label={t('dashboard.newSale')}
            onClick={() => navigate('/pos')}
          />
          <QuickActionButton
            icon={Package}
            label={t('dashboard.addProduct')}
            onClick={() => navigate('/products/new')}
          />
          <QuickActionButton
            icon={Users}
            label={t('dashboard.addCustomer')}
            onClick={() => navigate('/customers/new')}
          />
          <QuickActionButton
            icon={DollarSign}
            label={t('dashboard.recordPayment')}
            onClick={() => navigate('/debts')}
          />
          <QuickActionButton
            icon={BarChart3}
            label={t('dashboard.viewReports')}
            onClick={() => navigate('/reports')}
          />
        </div>
      </div>

      {/* Recent Sales */}
      <Card>
        <CardHeader>
          <CardTitle>{t('dashboard.recentSales')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {(stats?.recentSales as any[] || []).slice(0, 5).map((sale: any, index: number) => (
              <div key={index} className="flex items-center justify-between py-2 border-b border-gray-100 dark:border-gray-800 last:border-0">
                <div className="flex-1">
                  <p className="font-medium text-gray-900 dark:text-gray-100">
                    {sale.productName}
                  </p>
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    {sale.customerName} • {sale.timeAgo}
                  </p>
                </div>
                <div className="text-end">
                  <p className="font-medium text-gray-900 dark:text-gray-100">
                    ₪{sale.amount?.toLocaleString()}
                  </p>
                  <p className="text-sm text-green-600">{sale.paymentMethod}</p>
                </div>
              </div>
            ))}
            {(!stats?.recentSales || (stats.recentSales as any[]).length === 0) && (
              <p className="text-sm text-gray-500 dark:text-gray-400">
                لا توجد مبيعات حديثة
              </p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Top Products */}
      <Card>
        <CardHeader>
          <CardTitle>{t('dashboard.topProducts')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {(stats?.topProducts as any[] || []).slice(0, 5).map((product: any, index: number) => (
              <div key={index} className="flex items-center justify-between py-2 border-b border-gray-100 dark:border-gray-800 last:border-0">
                <div className="flex-1">
                  <p className="font-medium text-gray-900 dark:text-gray-100">
                    {product.name}
                  </p>
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    {product.stock} قطعة متوفرة
                  </p>
                </div>
                <div className="text-end">
                  <p className="font-medium text-gray-900 dark:text-gray-100">
                    ₪{product.sellingPrice?.toLocaleString()}
                  </p>
                  <p className="text-sm text-green-600">
                    {product.sales || 0} مبيع
                  </p>
                </div>
              </div>
            ))}
            {(!stats?.topProducts || (stats.topProducts as any[]).length === 0) && (
              <p className="text-sm text-gray-500 dark:text-gray-400">
                لا توجد منتجات
              </p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

interface StatCardProps {
  title: string;
  value: string;
  icon: any;
  trend: string;
  trendUp: boolean;
  subtitle?: string;
}

function StatCard({ title, value, icon: Icon, trend, trendUp, subtitle }: StatCardProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <p className="text-sm text-gray-500 dark:text-gray-400">{title}</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">
              {value}
            </p>
            {subtitle && (
              <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">
                {subtitle}
              </p>
            )}
          </div>
          <div className={`flex items-center gap-1 text-sm ${trendUp ? 'text-green-600' : 'text-red-600'}`}>
            {trendUp ? <ArrowUpRight className="w-4 h-4" /> : <ArrowDownRight className="w-4 h-4" />}
            {trend}
          </div>
        </div>
        <div className="mt-4 flex items-center gap-2 text-gray-500 dark:text-gray-400">
          <Icon className="w-4 h-4" />
        </div>
      </CardContent>
    </Card>
  );
}

interface KPICardProps {
  title: string;
  value: number | string;
  icon: any;
  color: 'blue' | 'green' | 'purple' | 'orange' | 'red';
  trend: string;
  trendUp: boolean;
}

function KPICard({ title, value, icon: Icon, color, trend, trendUp }: KPICardProps) {
  const colorClasses = {
    blue: 'bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-400',
    green: 'bg-green-100 dark:bg-green-900 text-green-600 dark:text-green-400',
    purple: 'bg-purple-100 dark:bg-purple-900 text-purple-600 dark:text-purple-400',
    orange: 'bg-orange-100 dark:bg-orange-900 text-orange-600 dark:text-orange-400',
    red: 'bg-red-100 dark:bg-red-900 text-red-600 dark:text-red-400',
  };

  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <p className="text-sm text-gray-500 dark:text-gray-400">{title}</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">
              {value}
            </p>
          </div>
          <div className={`flex items-center gap-1 text-sm ${trendUp ? 'text-green-600' : 'text-red-600'}`}>
            {trendUp ? <ArrowUpRight className="w-4 h-4" /> : <ArrowDownRight className="w-4 h-4" />}
            {trend}
          </div>
        </div>
        <div className="mt-4 flex items-center gap-2">
          <div className={`p-2 rounded-lg ${colorClasses[color]}`}>
            <Icon className="w-4 h-4" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

interface InsightItemProps {
  insight: any;
}

function InsightItem({ insight }: InsightItemProps) {
  const getInsightIcon = (type: string) => {
    switch (type) {
      case 'STOCK_OPPORTUNITY':
        return <TrendingUp className="w-4 h-4 text-green-500" />;
      case 'PRICING_SUGGESTION':
        return <DollarSign className="w-4 h-4 text-blue-500" />;
      case 'CUSTOMER_PATTERN':
        return <Users className="w-4 h-4 text-purple-500" />;
      case 'SALES_TREND':
        return <Activity className="w-4 h-4 text-orange-500" />;
      default:
        return <Lightbulb className="w-4 h-4 text-yellow-500" />;
    }
  };

  const getInsightPriority = (priority: string) => {
    switch (priority) {
      case 'HIGH':
        return 'bg-red-100 dark:bg-red-900 text-red-600 dark:text-red-400';
      case 'MEDIUM':
        return 'bg-yellow-100 dark:bg-yellow-900 text-yellow-600 dark:text-yellow-400';
      case 'LOW':
        return 'bg-green-100 dark:bg-green-900 text-green-600 dark:text-green-400';
      default:
        return 'bg-gray-100 dark:bg-gray-900 text-gray-600 dark:text-gray-400';
    }
  };

  return (
    <div className="flex items-start gap-3 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
      <div className="flex-shrink-0 mt-1">
        {getInsightIcon(insight.type)}
      </div>
      <div className="flex-1">
        <div className="flex items-start justify-between gap-2">
          <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
            {insight.title}
          </p>
          <span className={`px-2 py-1 text-xs rounded-full ${getInsightPriority(insight.priority)}`}>
            {insight.priority}
          </span>
        </div>
        <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
          {insight.description}
        </p>
        {insight.actionable && (
          <Button variant="ghost" size="sm" className="p-0 h-auto mt-2 text-primary-600">
            اتخاذ إجراء
          </Button>
        )}
      </div>
    </div>
  );
}

interface AlertItemProps {
  alert: any;
  onAction?: () => void;
}

// Helper function for notification icons
function getNotificationIcon(type: string) {
  switch (type) {
    case 'debt_overdue':
      return <DollarSign className="w-4 h-4 text-red-500" />;
    case 'warranty_expiring':
      return <AlertTriangle className="w-4 h-4 text-yellow-500" />;
    case 'low_stock':
      return <Package className="w-4 h-4 text-orange-500" />;
    case 'daily_insights':
      return <Activity className="w-4 h-4 text-blue-500" />;
    default:
      return <Bell className="w-4 h-4 text-gray-500" />;
  }
}

function AlertItem({ alert, onAction }: AlertItemProps) {
  const getAlertIcon = (type: string) => {
    switch (type) {
      case 'LOW_STOCK':
        return <Package className="w-4 h-4 text-orange-500" />;
      case 'OVERDUE_DEBT':
        return <DollarSign className="w-4 h-4 text-red-500" />;
      case 'WARRANTY_EXPIRING':
        return <AlertTriangle className="w-4 h-4 text-yellow-500" />;
      default:
        return <AlertTriangle className="w-4 h-4 text-gray-500" />;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'high':
        return 'border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/10';
      case 'medium':
        return 'border-orange-200 dark:border-orange-800 bg-orange-50 dark:bg-orange-900/10';
      case 'low':
        return 'border-yellow-200 dark:border-yellow-800 bg-yellow-50 dark:bg-yellow-900/10';
      default:
        return 'border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800';
    }
  };

  return (
    <div className={`flex items-center justify-between p-3 rounded-lg border ${getSeverityColor(alert.severity)}`}>
      <div className="flex items-center gap-3">
        {getAlertIcon(alert.type)}
        <div className="flex-1">
          <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
            {alert.title}
          </p>
          <p className="text-xs text-gray-500 dark:text-gray-400">
            {alert.message}
          </p>
        </div>
      </div>
      {alert.actionUrl && (
        <Button variant="ghost" size="sm" onClick={onAction}>
          عرض
        </Button>
      )}
    </div>
  );
}

interface QuickActionButtonProps {
  icon: any;
  label: string;
  onClick: () => void;
}

function QuickActionButton({ icon: Icon, label, onClick }: QuickActionButtonProps) {
  return (
    <Button
      variant="outline"
      className="flex flex-col items-center gap-2 h-auto py-4"
      onClick={onClick}
    >
      <Icon className="w-5 h-5" />
      <span className="text-xs">{label}</span>
    </Button>
  );
}