import { useState } from 'react'
import { clsx } from 'clsx'
import { SalesChart, InventoryChart, ProfitLossChart } from '../components/charts'
import type { ReportType, DateRange } from '../types/report'

export function ReportsPage() {
  const [activeReport, setActiveReport] = useState<ReportType>('sales')
  const [dateRange, setDateRange] = useState<DateRange>('month')

  // TODO: Fetch report data from API
  const salesData = [
    { date: '2024-01-01', sales: 10, revenue: 1500 },
    { date: '2024-01-02', sales: 15, revenue: 2250 },
    { date: '2024-01-03', sales: 8, revenue: 1200 },
    { date: '2024-01-04', sales: 20, revenue: 3000 },
    { date: '2024-01-05', sales: 12, revenue: 1800 },
    { date: '2024-01-06', sales: 18, revenue: 2700 },
    { date: '2024-01-07', sales: 25, revenue: 3750 },
  ]

  const inventoryData = [
    { category: 'إلكترونيات', count: 150, value: 45000 },
    { category: 'أجهزة منزلية', count: 80, value: 32000 },
    { category: 'أثاث', count: 45, value: 18000 },
    { category: 'ملابس', count: 200, value: 15000 },
    { category: 'أخرى', count: 60, value: 8000 },
  ]

  const profitLossData = [
    { month: '2024-01-01', revenue: 15000, expenses: 10000, profit: 5000 },
    { month: '2024-02-01', revenue: 18000, expenses: 11000, profit: 7000 },
    { month: '2024-03-01', revenue: 22000, expenses: 13000, profit: 9000 },
    { month: '2024-04-01', revenue: 20000, expenses: 12000, profit: 8000 },
    { month: '2024-05-01', revenue: 25000, expenses: 14000, profit: 11000 },
    { month: '2024-06-01', revenue: 28000, expenses: 15000, profit: 13000 },
  ]

  const reportTypes: { value: ReportType; label: string; icon: string }[] = [
    { value: 'sales', label: 'المبيعات', icon: '💰' },
    { value: 'inventory', label: 'المخزون', icon: '📦' },
    { value: 'customers', label: 'العملاء', icon: '👥' },
    { value: 'expenses', label: 'المصروفات', icon: '💸' },
    { value: 'profit_loss', label: 'الأرباح والخسائر', icon: '📊' },
  ]

  const dateRanges: { value: DateRange; label: string }[] = [
    { value: 'today', label: 'اليوم' },
    { value: 'week', label: 'هذا الأسبوع' },
    { value: 'month', label: 'هذا الشهر' },
    { value: 'quarter', label: 'هذا الربع' },
    { value: 'year', label: 'هذه السنة' },
    { value: 'custom', label: 'مخصص' },
  ]

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">التقارير</h1>
        <p className="text-muted">تحليل الأداء واتخاذ القرارات</p>
      </div>

      {/* Controls */}
      <div className="bg-surface rounded-lg p-4 mb-6 space-y-4">
        {/* Report Type Selector */}
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted self-center">نوع التقرير:</span>
          {reportTypes.map((type) => (
            <button
              key={type.value}
              onClick={() => setActiveReport(type.value)}
              className={clsx(
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2',
                activeReport === type.value
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted hover:bg-muted-80'
              )}
            >
              <span>{type.icon}</span>
              <span>{type.label}</span>
            </button>
          ))}
        </div>

        {/* Date Range Selector */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted">الفترة:</span>
          <select
            value={dateRange}
            onChange={(e) => setDateRange(e.target.value as DateRange)}
            className="px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            {dateRanges.map((range) => (
              <option key={range.value} value={range.value}>
                {range.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Report Content */}
      <div className="space-y-6">
        {activeReport === 'sales' && (
          <div className="space-y-6">
            {/* Summary Cards */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <SummaryCard label="إجمالي المبيعات" value="162" icon="🛒" />
              <SummaryCard label="إجمالي الإيرادات" value="16,200" icon="💰" />
              <SummaryCard label="متوسط قيمة البيع" value="100" icon="📊" />
              <SummaryCard label="أعلى منتج مبيعاً" value="iPhone 15" icon="🏆" />
            </div>

            {/* Charts */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <ChartCard title="المبيعات والإيرادات" subtitle="آخر 7 أيام">
                <SalesChart data={salesData} type="line" />
              </ChartCard>
              <ChartCard title="المبيعات حسب التصنيف" subtitle="توزيع المبيعات">
                <SalesChart data={salesData} type="bar" />
              </ChartCard>
            </div>
          </div>
        )}

        {activeReport === 'inventory' && (
          <div className="space-y-6">
            {/* Summary Cards */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <SummaryCard label="إجمالي المنتجات" value="535" icon="📦" />
              <SummaryCard label="إجمالي القيمة" value="118,000" icon="💰" />
              <SummaryCard label="منخفض المخزون" value="25" icon="⚠️" color="warning" />
              <SummaryCard label="نفذت الكمية" value="12" icon="🚨" color="danger" />
            </div>

            {/* Charts */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <ChartCard title="المخزون حسب التصنيف" subtitle="بالعدد">
                <InventoryChart data={inventoryData} metric="count" />
              </ChartCard>
              <ChartCard title="قيمة المخزون حسب التصنيف" subtitle="بالقيمة">
                <InventoryChart data={inventoryData} metric="value" />
              </ChartCard>
            </div>
          </div>
        )}

        {activeReport === 'customers' && (
          <div className="space-y-6">
            {/* Summary Cards */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <SummaryCard label="إجمالي العملاء" value="1,250" icon="👥" />
              <SummaryCard label="العملاء النشطين" value="980" icon="✅" color="success" />
              <SummaryCard label="إجمالي الديون" value="15,500" icon="📋" color="warning" />
              <SummaryCard label="أعلى عميل" value="أحمد محمد" icon="🏆" />
            </div>

            {/* Customer Acquisition Chart */}
            <ChartCard title="اكتساب العملاء" subtitle="عملاء جدد شهرياً">
              <SalesChart data={salesData} type="bar" />
            </ChartCard>
          </div>
        )}

        {activeReport === 'expenses' && (
          <div className="space-y-6">
            {/* Summary Cards */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <SummaryCard label="إجمالي المصروفات" value="75,000" icon="💸" />
              <SummaryCard label="هذا الشهر" value="12,500" icon="📅" />
              <SummaryCard label="أعلى فئة" value="الرواتب" icon="📊" />
              <SummaryCard label="متوسط شهري" value="10,000" icon="📈" />
            </div>

            {/* Expenses by Category */}
            <ChartCard title="المصروفات حسب الفئة" subtitle="توزيع المصروفات">
              <InventoryChart data={inventoryData} metric="value" />
            </ChartCard>
          </div>
        )}

        {activeReport === 'profit_loss' && (
          <div className="space-y-6">
            {/* Summary Cards */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <SummaryCard label="إجمالي الإيرادات" value="162,000" icon="💰" color="success" />
              <SummaryCard label="إجمالي المصروفات" value="75,000" icon="💸" color="danger" />
              <SummaryCard label="إجمالي الربح" value="87,000" icon="📈" color="primary" />
              <SummaryCard label="هامش الربح" value="53.7%" icon="📊" />
            </div>

            {/* Profit Loss Chart */}
            <ChartCard title="الأرباح والخسائر الشهرية" subtitle="آخر 6 أشهر">
              <ProfitLossChart data={profitLossData} />
            </ChartCard>
          </div>
        )}
      </div>
    </div>
  )
}

function SummaryCard({ label, value, icon, color = 'primary' }: { label: string; value: string; icon: string; color?: 'success' | 'danger' | 'warning' | 'primary' }) {
  const colorClasses = {
    success: 'text-success',
    danger: 'text-danger',
    warning: 'text-warning',
    primary: 'text-primary',
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

function ChartCard({ title, subtitle, children }: { title: string; subtitle: string; children: React.ReactNode }) {
  return (
    <div className="bg-surface rounded-lg p-6 border border-border">
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-text">{title}</h3>
        <p className="text-sm text-muted">{subtitle}</p>
      </div>
      {children}
    </div>
  )
}
