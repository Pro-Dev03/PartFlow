import { useTranslation } from 'react-i18next'
import { useDashboardStats, useDashboardAlerts, useQuickActions, useRecentActivity } from './hooks/useDashboard'
import { StatCard } from './components/StatCard'
import { AlertCard } from './components/AlertCard'
import { QuickActionCard } from './components/QuickActionCard'
import { Card, CardHeader } from '../../components/ui/Card'
import { Skeleton, TextSkeleton } from '../../components/ui/Skeleton'

export function Dashboard() {
  const { t } = useTranslation()
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: alerts, isLoading: alertsLoading } = useDashboardAlerts()
  const { data: quickActions, isLoading: quickActionsLoading } = useQuickActions()
  const { data: recentActivity, isLoading: recentActivityLoading } = useRecentActivity()

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('ar-SA', {
      style: 'currency',
      currency: 'ILS',
      minimumFractionDigits: 0,
    }).format(amount)
  }

  const formatRelativeTime = (date: Date) => {
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(diff / 3600000)
    const days = Math.floor(diff / 86400000)

    if (minutes < 1) return 'الآن'
    if (minutes < 60) return `منذ ${minutes} دقيقة`
    if (hours < 24) return `منذ ${hours} ساعة`
    return `منذ ${days} يوم`
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Welcome Section */}
      <div>
        <h1 className="text-2xl font-bold text-text">
          {t('dashboard.welcome')}
        </h1>
        <p className="text-muted mt-1">
          {new Date().toLocaleDateString('ar-SA', { 
            weekday: 'long', 
            year: 'numeric', 
            month: 'long', 
            day: 'numeric' 
          })}
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {statsLoading ? (
          <>
            <Skeleton variant="rectangular" height={120} />
            <Skeleton variant="rectangular" height={120} />
            <Skeleton variant="rectangular" height={120} />
            <Skeleton variant="rectangular" height={120} />
          </>
        ) : (
          <>
            <StatCard
              title={t('dashboard.sales')}
              value={formatCurrency(stats?.todaySales || 0)}
              icon="💰"
              trend={{ value: 12, isPositive: true }}
            />
            <StatCard
              title={t('dashboard.profit')}
              value={formatCurrency(stats?.todayProfit || 0)}
              icon="📈"
              trend={{ value: 8, isPositive: true }}
            />
            <StatCard
              title={t('dashboard.inventory')}
              value={formatCurrency(stats?.inventoryValue || 0)}
              icon="📦"
            />
            <StatCard
              title={t('dashboard.debts')}
              value={formatCurrency(stats?.outstandingDebts || 0)}
              icon="💳"
              trend={{ value: 5, isPositive: false }}
            />
          </>
        )}
      </div>

      {/* Two Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - Alerts & Quick Actions */}
        <div className="lg:col-span-2 space-y-6">
          {/* Alerts Section */}
          <Card>
            <CardHeader 
              title={t('dashboard.needsAttention')} 
              description={`${alerts?.length || 0} إجراءات تحتاج انتباهك`}
            />
            <div className="space-y-3">
              {alertsLoading ? (
                <>
                  <Skeleton variant="rectangular" height={80} />
                  <Skeleton variant="rectangular" height={80} />
                  <Skeleton variant="rectangular" height={80} />
                </>
              ) : alerts && alerts.length > 0 ? (
                alerts.map((alert) => (
                  <AlertCard key={alert.id} alert={alert} />
                ))
              ) : (
                <div className="text-center py-8 text-muted">
                  <span className="text-4xl">✓</span>
                  <p className="mt-2">لا توجد تنبيهات حالياً</p>
                </div>
              )}
            </div>
          </Card>

          {/* Quick Actions */}
          <Card>
            <CardHeader title={t('dashboard.quickActions')} />
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              {quickActionsLoading ? (
                <>
                  <Skeleton variant="rectangular" height={80} />
                  <Skeleton variant="rectangular" height={80} />
                  <Skeleton variant="rectangular" height={80} />
                  <Skeleton variant="rectangular" height={80} />
                </>
              ) : (
                quickActions?.map((action) => (
                  <QuickActionCard key={action.id} action={action} />
                ))
              )}
            </div>
          </Card>
        </div>

        {/* Right Column - Recent Activity */}
        <div className="space-y-6">
          <Card>
            <CardHeader title="النشاط الأخير" />
            <div className="space-y-4">
              {recentActivityLoading ? (
                <TextSkeleton lines={5} />
              ) : recentActivity && recentActivity.length > 0 ? (
                recentActivity.map((activity) => (
                  <div key={activity.id} className="flex items-start gap-3 pb-3 border-b border-border last:border-0">
                    <div className="flex-shrink-0 w-8 h-8 rounded-full bg-surface flex items-center justify-center">
                      {activity.type === 'sale' && '💰'}
                      {activity.type === 'payment' && '💳'}
                      {activity.type === 'purchase' && '🚚'}
                      {activity.type === 'return' && '↩️'}
                      {activity.type === 'inventory' && '📦'}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-text truncate">
                        {activity.title}
                      </p>
                      <p className="text-xs text-muted truncate">
                        {activity.description}
                      </p>
                      <div className="flex items-center gap-2 mt-1">
                        <span className="text-xs text-muted">
                          {formatRelativeTime(activity.timestamp)}
                        </span>
                        {activity.amount && (
                          <span className="text-xs font-medium text-primary">
                            {formatCurrency(activity.amount)}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                ))
              ) : (
                <div className="text-center py-8 text-muted">
                  <p>لا يوجد نشاط حديث</p>
                </div>
              )}
            </div>
          </Card>

          {/* Today's Summary */}
          <Card>
            <CardHeader title="ملخص اليوم" />
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted">المبيعات</span>
                <span className="font-medium">{stats?.todaySalesCount || 0} عملية</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted">العملاء الجدد</span>
                <span className="font-medium">{stats?.newCustomersCount || 0} عميل</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted">المنتجات المنخفضة</span>
                <span className="font-medium text-warning">{stats?.lowStockCount || 0} قطعة</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted">الديون المتأخرة</span>
                <span className="font-medium text-danger">{stats?.overdueDebtsCount || 0} دين</span>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}