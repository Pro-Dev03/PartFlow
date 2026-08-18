import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { clsx } from 'clsx'
import { Card, CardHeader } from '@components/ui/card'
import { Button } from '@components/ui/button'
import { AdvancedChart } from '../../components/charts/AdvancedChart'

type ReportPeriod = 'today' | 'week' | 'month' | 'quarter' | 'year'
type ReportType = 'overview' | 'sales' | 'products' | 'customers' | 'inventory' | 'debts'

export function Reports() {
  const { t } = useTranslation()
  const [period, setPeriod] = useState<ReportPeriod>('month')
  const [activeReport, setActiveReport] = useState<ReportType>('overview')

  // Mock data - TODO: Replace with API calls
  const reportData = {
    overview: {
      sales: 74500,
      profit: 18240,
      expenses: 7300,
      netProfit: 10940,
      salesCount: 124,
      averageSale: 74500 / 124,
      topProducts: [
        { name: 'RTX 4070', sales: 15, revenue: 35250 },
        { name: 'RTX 3060', sales: 22, revenue: 27500 },
        { name: 'RAM 16GB', sales: 45, revenue: 11250 },
      ],
      slowMovingProducts: [
        { name: 'Old HDD 1TB', daysInStock: 120, value: 150 },
        { name: 'DDR3 RAM 8GB', daysInStock: 90, value: 80 },
      ],
      moneyLockedInInventory: 32400,
    },
    sales: {
      dailySales: [
        { date: '2024-08-01', amount: 2500 },
        { date: '2024-08-02', amount: 3200 },
        { date: '2024-08-03', amount: 2800 },
        { date: '2024-08-04', amount: 4100 },
        { date: '2024-08-05', amount: 3800 },
        { date: '2024-08-06', amount: 2900 },
        { date: '2024-08-07', amount: 3500 },
      ],
      paymentMethods: [
        { method: 'نقد', amount: 45000, percentage: 60 },
        { method: 'بطاقة', amount: 22500, percentage: 30 },
        { method: 'تحويل', amount: 7500, percentage: 10 },
      ],
    },
    products: {
      categorySales: [
        { category: 'GPUs', sales: 45000, percentage: 60 },
        { category: 'Memory', sales: 18750, percentage: 25 },
        { category: 'Storage', sales: 7500, percentage: 10 },
        { category: 'Other', sales: 3750, percentage: 5 },
      ],
      conditionSales: [
        { condition: 'جديد', sales: 52000, percentage: 70 },
        { condition: 'مستعمل', sales: 22500, percentage: 30 },
      ],
    },
    customers: {
      newCustomers: 12,
      returningCustomers: 45,
      topSpenders: [
        { name: 'أحمد محمد', spent: 8450 },
        { name: 'سارة علي', spent: 6200 },
        { name: 'خالد عبدالله', spent: 4800 },
      ],
    },
    inventory: {
      totalValue: 125000,
      lowStockItems: 4,
      outOfStockItems: 2,
      turnoverRate: 2.5,
    },
    debts: {
      totalDebts: 24800,
      overdueDebts: 7500,
      paidThisMonth: 12500,
      collectionRate: 85,
    },
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('ar-SA', {
      style: 'currency',
      currency: 'ILS',
      minimumFractionDigits: 0,
    }).format(amount)
  }

  const periodLabels = {
    today: 'اليوم',
    week: 'هذا الأسبوع',
    month: 'هذا الشهر',
    quarter: 'هذا الربع',
    year: 'هذه السنة',
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text">{t('reports.title')}</h1>
          <p className="text-muted mt-1">تحليلات ذكية لأداء المحل</p>
        </div>
        <div className="flex gap-2">
          <select
            value={period}
            onChange={(e) => setPeriod(e.target.value as ReportPeriod)}
            className="px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            {Object.entries(periodLabels).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
          <Button variant="primary">
            تصدير التقرير
          </Button>
        </div>
      </div>

      {/* Report Type Tabs */}
      <div className="flex gap-2 border-b border-border">
        {([
          { value: 'overview', label: 'نظرة عامة' },
          { value: 'sales', label: 'المبيعات' },
          { value: 'products', label: 'المنتجات' },
          { value: 'customers', label: 'العملاء' },
          { value: 'inventory', label: 'المخزون' },
          { value: 'debts', label: 'الديون' },
        ] as const).map((option) => (
          <button
            key={option.value}
            onClick={() => setActiveReport(option.value)}
            className={clsx(
              'px-4 py-3 text-sm font-medium border-b-2 transition-colors',
              activeReport === option.value
                ? 'border-primary text-primary'
                : 'border-transparent text-muted hover:text-text'
            )}
          >
            {option.label}
          </button>
        ))}
      </div>

      {/* Report Content */}
      {activeReport === 'overview' && (
        <div className="space-y-6">
          {/* Main Question */}
          <Card className="bg-gradient-to-br from-primary-500 to-primary-600 text-white">
            <div className="p-6">
              <h2 className="text-xl font-bold mb-2">كيف كان هذا الشهر؟</h2>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
                <div>
                  <p className="text-sm opacity-80">المبيعات</p>
                  <p className="text-2xl font-bold">{formatCurrency(reportData.overview.sales)}</p>
                </div>
                <div>
                  <p className="text-sm opacity-80">الربح</p>
                  <p className="text-2xl font-bold">{formatCurrency(reportData.overview.profit)}</p>
                </div>
                <div>
                  <p className="text-sm opacity-80">المصروفات</p>
                  <p className="text-2xl font-bold">{formatCurrency(reportData.overview.expenses)}</p>
                </div>
                <div>
                  <p className="text-sm opacity-80">صافي الربح</p>
                  <p className="text-2xl font-bold">{formatCurrency(reportData.overview.netProfit)}</p>
                </div>
              </div>
            </div>
          </Card>

          {/* Key Insights */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card>
              <CardHeader title="أكثر المنتجات مبيعاً" />
              <div className="p-4 space-y-3">
                {reportData.overview.topProducts.map((product, index) => (
                  <div key={index} className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <span className={clsx(
                        'w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold',
                        index === 0 ? 'bg-yellow-500 text-white' :
                        index === 1 ? 'bg-gray-400 text-white' :
                        index === 2 ? 'bg-orange-500 text-white' : 'bg-muted text-muted'
                      )}>
                        {index + 1}
                      </span>
                      <span className="font-medium text-text">{product.name}</span>
                    </div>
                    <div className="text-right">
                      <p className="font-bold text-text">{formatCurrency(product.revenue)}</p>
                      <p className="text-xs text-muted">{product.sales} عملية</p>
                    </div>
                  </div>
                ))}
              </div>
            </Card>

            <Card>
              <CardHeader title="المخزون الراكد" />
              <div className="p-4 space-y-3">
                {reportData.overview.slowMovingProducts.map((product, index) => (
                  <div key={index} className="flex items-center justify-between">
                    <div>
                      <p className="font-medium text-text">{product.name}</p>
                      <p className="text-xs text-warning">{product.daysInStock} يوم في المخزون</p>
                    </div>
                    <p className="font-bold text-warning">{formatCurrency(product.value)}</p>
                  </div>
                ))}
                <div className="pt-3 border-t border-border">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted">إجمالي الأموال المجمدة:</span>
                    <span className="font-bold text-danger">
                      {formatCurrency(reportData.overview.moneyLockedInInventory)}
                    </span>
                  </div>
                  <p className="text-xs text-muted mt-1">
                    💡 لديك {formatCurrency(reportData.overview.moneyLockedInInventory)} في مخزون لم يتحرك منذ أكثر من 90 يوم
                  </p>
                </div>
              </div>
            </Card>
          </div>

          {/* Sales Chart */}
          <Card>
            <CardHeader title="اتجاه المبيعات" />
            <div className="p-4">
              <AdvancedChart
                type="line"
                data={reportData.sales.dailySales.map(item => ({ name: item.date, value: item.amount }))}
                height={300}
              />
            </div>
          </Card>
        </div>
      )}

      {activeReport === 'sales' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card>
              <CardHeader title="اتجاه المبيعات" />
              <div className="p-4">
                <AdvancedChart
                  type="line"
                  data={reportData.sales.dailySales.map(item => ({ name: item.date, value: item.amount }))}
                  height={250}
                />
              </div>
            </Card>

            <Card>
              <CardHeader title="طرق الدفع" />
              <div className="p-4">
                <AdvancedChart
                  type="pie"
                  data={reportData.sales.paymentMethods.map(item => ({ name: item.method, value: item.amount }))}
                  height={250}
                />
              </div>
            </Card>
          </div>

          <Card>
            <CardHeader title="تفاصيل طرق الدفع" />
            <div className="p-4">
              <div className="space-y-3">
                {reportData.sales.paymentMethods.map((item) => (
                  <div key={item.method} className="flex items-center justify-between p-3 bg-muted-10 rounded-lg">
                    <div>
                      <p className="font-medium text-text">{item.method}</p>
                      <p className="text-sm text-muted">{item.percentage}% من المبيعات</p>
                    </div>
                    <p className="font-bold text-text">{formatCurrency(item.amount)}</p>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        </div>
      )}

      {activeReport === 'products' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card>
              <CardHeader title="المبيعات حسب التصنيف" />
              <div className="p-4">
                <AdvancedChart
                  type="bar"
                  data={reportData.products.categorySales.map(item => ({ name: item.category, value: item.sales }))}
                  height={250}
                />
              </div>
            </Card>

            <Card>
              <CardHeader title="جديد vs مستعمل" />
              <div className="p-4">
                <AdvancedChart
                  type="pie"
                  data={reportData.products.conditionSales.map(item => ({ name: item.condition, value: item.sales }))}
                  height={250}
                />
              </div>
            </Card>
          </div>

          <Card>
            <CardHeader title="تفاصيل التصنيفات" />
            <div className="p-4">
              <div className="space-y-3">
                {reportData.products.categorySales.map((item) => (
                  <div key={item.category} className="flex items-center justify-between p-3 bg-muted-10 rounded-lg">
                    <div>
                      <p className="font-medium text-text">{item.category}</p>
                      <p className="text-sm text-muted">{item.percentage}% من المبيعات</p>
                    </div>
                    <p className="font-bold text-text">{formatCurrency(item.sales)}</p>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        </div>
      )}

      {activeReport === 'customers' && (
        <div className="space-y-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard label="عملاء جدد" value={reportData.customers.newCustomers.toString()} icon="👤" color="primary" />
            <StatCard label="عملاء عائدون" value={reportData.customers.returningCustomers.toString()} icon="🔄" color="success" />
            <StatCard label="إجمالي العملاء" value={(reportData.customers.newCustomers + reportData.customers.returningCustomers).toString()} icon="👥" color="info" />
            <StatCard label="معدل العودة" value="78%" icon="📈" color="warning" />
          </div>

          <Card>
            <CardHeader title="أكثر العملاء إنفاقاً" />
            <div className="p-4">
              <div className="space-y-3">
                {reportData.customers.topSpenders.map((customer, index) => (
                  <div key={index} className="flex items-center justify-between p-3 bg-muted-10 rounded-lg">
                    <div className="flex items-center gap-3">
                      <span className={clsx(
                        'w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold',
                        index === 0 ? 'bg-yellow-500 text-white' :
                        index === 1 ? 'bg-gray-400 text-white' :
                        'bg-orange-500 text-white'
                      )}>
                        {index + 1}
                      </span>
                      <span className="font-medium text-text">{customer.name}</span>
                    </div>
                    <p className="font-bold text-success">{formatCurrency(customer.spent)}</p>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        </div>
      )}

      {activeReport === 'inventory' && (
        <div className="space-y-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard label="قيمة المخزون" value={formatCurrency(reportData.inventory.totalValue)} icon="📦" color="primary" />
            <StatCard label="منخفض المخزون" value={reportData.inventory.lowStockItems.toString()} icon="⚠️" color="warning" />
            <StatCard label="نفذت الكمية" value={reportData.inventory.outOfStockItems.toString()} icon="🔴" color="danger" />
            <StatCard label="معدل الدوران" value={reportData.inventory.turnoverRate + 'x'} icon="🔄" color="success" />
          </div>

          <Card>
            <CardHeader title="تحليل المخزون" />
            <div className="p-4">
              <div className="space-y-4">
                <div className="p-4 bg-success-10 rounded-lg border border-success-30">
                  <h4 className="font-semibold text-success mb-2">معدل دوران المخزون جيد</h4>
                  <p className="text-sm text-muted">
                    معدل الدوران الحالي {reportData.inventory.turnoverRate}x يعني أن المخزون يتبدل {reportData.inventory.turnoverRate} مرة في السنة
                  </p>
                </div>
                {reportData.inventory.lowStockItems > 0 && (
                  <div className="p-4 bg-warning-10 rounded-lg border border-warning-30">
                    <h4 className="font-semibold text-warning mb-2">تنبيه: منتجات منخفضة المخزون</h4>
                    <p className="text-sm text-muted">
                      لديك {reportData.inventory.lowStockItems} منتج وصلت للحد الأدنى، يرجى التزويد
                    </p>
                  </div>
                )}
                {reportData.inventory.outOfStockItems > 0 && (
                  <div className="p-4 bg-danger-10 rounded-lg border border-danger-30">
                    <h4 className="font-semibold text-danger mb-2">تنبيه: منتجات نفذت</h4>
                    <p className="text-sm text-muted">
                      لديك {reportData.inventory.outOfStockItems} منتج نفذت الكمية، قد تفقد مبيعات
                    </p>
                  </div>
                )}
              </div>
            </div>
          </Card>
        </div>
      )}

      {activeReport === 'debts' && (
        <div className="space-y-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard label="إجمالي الديون" value={formatCurrency(reportData.debts.totalDebts)} icon="💳" color="primary" />
            <StatCard label="ديون متأخرة" value={formatCurrency(reportData.debts.overdueDebts)} icon="🔴" color="danger" />
            <StatCard label="مدفوعة هذا الشهر" value={formatCurrency(reportData.debts.paidThisMonth)} icon="✅" color="success" />
            <StatCard label="معدل التحصيل" value={reportData.debts.collectionRate.toString() + '%'} icon="📈" color="info" />
          </div>

          <Card>
            <CardHeader title="تحليل الديون" />
            <div className="p-4">
              <div className="space-y-4">
                <div className="p-4 bg-success-10 rounded-lg border border-success-30">
                  <h4 className="font-semibold text-success mb-2">معدل تحصيل ممتاز</h4>
                  <p className="text-sm text-muted">
                    معدل التحصيل {reportData.debts.collectionRate}% ي高于 المتوسط، استمر في المتابعة المنتظمة
                  </p>
                </div>
                {reportData.debts.overdueDebts > 0 && (
                  <div className="p-4 bg-danger-10 rounded-lg border border-danger-30">
                    <h4 className="font-semibold text-danger mb-2">تنبيه: ديون متأخرة</h4>
                    <p className="text-sm text-muted">
                      لديك {formatCurrency(reportData.debts.overdueDebts)} ديون متأخرة، يرجى التواصل مع العملاء
                    </p>
                  </div>
                )}
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}

function StatCard({ label, value, icon, color }: { label: string; value: string; icon: string; color: 'primary' | 'success' | 'warning' | 'danger' | 'info' }) {
  const colorClasses = {
    primary: 'text-primary',
    success: 'text-success',
    warning: 'text-warning',
    danger: 'text-danger',
    info: 'text-info',
  }

  return (
    <div className="bg-surface rounded-lg p-4 border border-border">
      <div className="flex items-center gap-2 mb-2">
        <span className="text-2xl">{icon}</span>
        <span className="text-sm text-muted">{label}</span>
      </div>
      <p className={clsx('text-xl font-bold', colorClasses[color])}>{value}</p>
    </div>
  )
}

export default Reports
