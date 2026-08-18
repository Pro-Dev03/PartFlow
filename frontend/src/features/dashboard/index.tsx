import { Button } from '@components/ui/button'
import { Card, CardHeader } from '@components/ui/card'

export function Dashboard() {
  // Mock data matching partflow_demo.html with Arabic
  const stats = [
    { label: "مبيعات اليوم", value: '₪12,450', trend: '↑ 12.5% مقارنة بالأمس', color: 'text-success' },
    { label: 'الطلبات', value: '84', trend: '↑ 8.2% هذا الأسبوع', color: 'text-success' },
    { label: "ربح اليوم", value: '₪3,240', trend: '↑ 14.2% هذا الشهر', color: 'text-success' },
    { label: 'المستحقات', value: '₪5,820', trend: '⚠ 3 عملاء متأخرين', color: 'text-warning' },
  ]

  const recentSales = [
    { id: '1', product: 'RTX 4070', customer: 'أحمد محمد', amount: '₪2,790', time: 'منذ 5 دقائق' },
    { id: '2', product: 'RAM 32GB ×2', customer: 'سارة علي', amount: '₪1,100', time: 'منذ 15 دقيقة' },
    { id: '3', product: 'SSD 1TB', customer: 'خالد عبدالله', amount: '₪450', time: 'منذ 30 دقيقة' },
  ]

  const inventoryAlerts = [
    { id: '1', product: 'RTX 3060', stock: 1, minStock: 3, status: 'critical' },
    { id: '2', product: 'RAM 8GB DDR4', stock: 0, minStock: 5, status: 'out' },
    { id: '3', product: 'PSU 650W', stock: 2, minStock: 4, status: 'low' },
  ]

  const quickActions = [
    { id: '1', icon: '📷', label: 'مسح', action: 'scan' },
    { id: '2', icon: '💰', label: 'بيع', action: 'sale' },
    { id: '3', icon: '📦', label: 'منتج', action: 'product' },
    { id: '4', icon: '👤', label: 'عميل', action: 'customer' },
  ]

  return (
    <div className="p-8 max-w-[1500px] mx-auto direction-rtl animate-fade-in">
      {/* Header */}
      <div className="flex justify-between items-start mb-8 gap-5">
        <div className="flex-1">
          <div className="text-sm text-muted mb-2">
            {new Date().toLocaleDateString('ar-SA', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
          </div>
          <h1 className="text-4xl font-extrabold mb-2 text-text leading-tight">
            صباح الخير، أحمد 👋
          </h1>
          <p className="text-muted text-base">
            إليك ما يحدث في متجرك اليوم.
          </p>
        </div>
        <div className="flex gap-3 flex-shrink-0">
          <Button variant="outline">
            مسح باركود
          </Button>
          <Button variant="primary">
            + بيع جديد
          </Button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 mb-6">
        {stats.map((stat, index) => (
          <Card
            key={index}
            className="hover:shadow-md transition-shadow cursor-pointer"
          >
            <div className="p-6">
              <div className="text-sm text-muted font-medium mb-2">{stat.label}</div>
              <div className="text-3xl font-extrabold mb-2 text-text leading-none">{stat.value}</div>
              <div className={`text-sm font-medium ${stat.color}`}>{stat.trend}</div>
            </div>
          </Card>
        ))}
      </div>

      {/* Two Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column */}
        <div className="lg:col-span-2 space-y-6">
          {/* Sales Overview */}
          <Card>
            <CardHeader title="نظرة عامة على المبيعات" />
            <div className="p-6">
              <div className="h-60 flex items-center justify-center bg-muted-10 rounded-lg border-2 border-dashed border-border">
                <div className="text-center">
                  <div className="text-4xl mb-3">📈</div>
                  <div className="text-muted text-base">سيظهر رسم بياني للمبيعات هنا</div>
                </div>
              </div>
            </div>
          </Card>

          {/* Recent Sales */}
          <Card>
            <CardHeader title="المبيعات الأخيرة" />
            <div className="p-6">
              {recentSales.map((sale) => (
                <div
                  key={sale.id}
                  className="py-4 border-b border-border flex justify-between items-center hover:bg-muted-5 transition-colors last:border-0"
                >
                  <div className="flex-1">
                    <div className="font-semibold text-text text-base mb-1">{sale.product}</div>
                    <div className="text-sm text-muted">{sale.customer}</div>
                  </div>
                  <div className="text-left mr-4">
                    <div className="font-bold text-primary text-lg mb-1">{sale.amount}</div>
                    <div className="text-xs text-muted">{sale.time}</div>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </div>

        {/* Right Column */}
        <div className="space-y-6">
          {/* Inventory Alerts */}
          <Card>
            <CardHeader title="تنبيهات المخزون" />
            <div className="p-6 space-y-4">
              {inventoryAlerts.map((alert) => (
                <div
                  key={alert.id}
                  className={`p-4 rounded-lg border ${
                    alert.status === 'critical'
                      ? 'bg-danger-10 border-danger-30'
                      : alert.status === 'out'
                      ? 'bg-danger-10 border-danger-30'
                      : 'bg-warning-10 border-warning-30'
                  }`}
                >
                  <div className="flex justify-between items-center mb-2">
                    <span className="font-semibold text-text text-base">{alert.product}</span>
                    <span
                      className={`text-xs px-3 py-1 rounded-full font-semibold ${
                        alert.status === 'critical'
                          ? 'bg-danger text-white'
                          : alert.status === 'out'
                          ? 'bg-danger text-white'
                          : 'bg-warning text-white'
                      }`}
                    >
                      {alert.status === 'critical' ? '🔴 حرج' : alert.status === 'out' ? '❌ نفذ' : '🟡 منخفض'}
                    </span>
                  </div>
                  <div className="text-sm text-muted">
                    المخزون: {alert.stock} / الحد الأدنى: {alert.minStock}
                  </div>
                </div>
              ))}
            </div>
          </Card>

          {/* Quick Actions */}
          <Card>
            <CardHeader title="عمليات سريعة" />
            <div className="p-6">
              <div className="grid grid-cols-2 gap-3">
                {quickActions.map((action) => (
                  <button
                    key={action.id}
                    className="p-5 border border-border rounded-lg bg-surface hover:bg-muted-10 hover:border-primary transition-all cursor-pointer flex flex-col items-center gap-3"
                  >
                    <span className="text-3xl">{action.icon}</span>
                    <span className="text-sm font-semibold text-text">{action.label}</span>
                  </button>
                ))}
              </div>
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}

export default Dashboard