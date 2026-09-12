import { StatCard } from '../../../components/ui/stat-card';
import { useNavigate } from 'react-router-dom';
import { ShoppingCart, DollarSign, AlertTriangle, Package, RotateCcw } from 'lucide-react';

interface DashboardMetricsProps {
  stats?: {
    todaySales?: number;
    todayProfit?: number;
    todaySupplierReturns?: number;
    todayCollected?: number;
    todayDebtCollected?: number;
    todaySupplierPaid?: number;
    todayExpenses?: number;
    todayCashDifference?: number;
    outstandingDebts?: number;
    activeCustomers?: number;
    lowStockCount?: number;
  };
}

export function DashboardMetrics({ stats }: DashboardMetricsProps) {
  const navigate = useNavigate();
  // Dashboard stats are the authoritative live source. Aggregation endpoints may
  // contain stale snapshots while their background refresh is pending.
  const todaySales = Number(stats?.todaySales ?? 0);
  const todayProfit = Number(stats?.todayProfit ?? 0);
  const todaySupplierReturns = Number(stats?.todaySupplierReturns ?? 0);
  const todayCollected = Number(stats?.todayCollected ?? 0);
  const todayDebtCollected = Number(stats?.todayDebtCollected ?? 0);
  const todaySupplierPaid = Number(stats?.todaySupplierPaid ?? 0);
  const todayExpenses = Number(stats?.todayExpenses ?? 0);
  const todayCashDifference = Number(stats?.todayCashDifference ?? 0);
  const profitMargin = stats?.profitMargin;
  const outstandingDebts = Number(stats?.outstandingDebts ?? 0);
  const activeCustomers = Number(stats?.activeCustomers ?? 0);
  // The attention card and this metric must use the same live product count.
  const lowStockCount = stats?.lowStockCount ?? 0;

  // Keep currency output stable when the API returns decimal or negative values.
  const formatCurrency = (value: number) =>
    Number.isFinite(value)
      ? value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })
      : '0.00';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-5)' }}>
      <section aria-labelledby="dashboard-core-metrics">
        <h2 id="dashboard-core-metrics" className="mb-3 text-sm font-semibold text-text-primary">
          المؤشرات الأساسية
        </h2>
           <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '12px' }}
             className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-5">
          <StatCard
            title="صافي مبيعات اليوم"
            value={<span className="numeric-metric">₪{formatCurrency(todaySales)}</span>}
            icon={ShoppingCart}
            trend={stats?.salesTrend}
            trendUp={stats?.salesTrendUp}
            subtitle="بعد خصم مرتجعات العملاء"
            variant="featured"
            onClick={() => navigate('/app/reports?report=net-sales')}
          />
          <StatCard
            title="صافي ربح اليوم"
            value={<span className="numeric-metric">₪{formatCurrency(todayProfit)}</span>}
            icon={DollarSign}
            trend={stats?.profitTrend}
            trendUp={stats?.profitTrendUp}
            subtitle={profitMargin
              ? `بعد الخصم - هامش الربح: ${profitMargin}%`
              : 'بعد خصم تكلفة المنتجات والمصاريف والمرتجعات'}
            onClick={() => navigate('/app/reports?report=profit')}
          />
          <StatCard
            title="الديون المستحقة"
            value={<span className="numeric-metric">₪{formatCurrency(outstandingDebts)}</span>}
            icon={AlertTriangle}
            trend={stats?.debtsTrend}
            trendUp={stats?.debtsTrendUp}
            subtitle={<span className="numeric-quantity">{activeCustomers} عميل</span>}
            variant="warning"
            onClick={() => navigate('/app/debts')}
          />
          <StatCard
            title="المخزون المنخفض"
            value={<span className="numeric-quantity">{lowStockCount}</span>}
            icon={Package}
            trend={null}
            trendUp={null}
            subtitle="يحتاج انتباه"
            variant="danger"
            onClick={() => navigate('/app/inventory')}
          />
          <StatCard
            title="مرتجعات الموردين اليوم"
            value={<span className="numeric-metric">₪{formatCurrency(todaySupplierReturns)}</span>}
            icon={RotateCcw}
            subtitle="قيمة المرتجعات المكتملة"
            variant="warning"
            onClick={() => navigate('/app/supplier-returns')}
          />
        </div>
      </section>

      <section aria-labelledby="dashboard-financial-operations">
        <h2 id="dashboard-financial-operations" className="mb-3 text-sm font-semibold text-text-primary">
          العمليات المالية اليوم
        </h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '12px' }}
             className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
          <StatCard
            title="تحصيل ديون العملاء اليوم"
            value={<span className="numeric-metric">₪{formatCurrency(todayDebtCollected)}</span>}
            icon={DollarSign}
            subtitle="المبالغ المدفوعة لسداد الديون"
            variant="success"
            onClick={() => navigate('/app/debts')}
          />
          <StatCard
            title="المدفوع للموردين اليوم"
            value={<span className="numeric-metric">₪{formatCurrency(todaySupplierPaid)}</span>}
            icon={Package}
            subtitle="دفعات خرجت للموردين"
            variant="warning"
            onClick={() => navigate('/app/purchases')}
          />
          <StatCard
            title="المصروفات المدفوعة اليوم"
            value={<span className="numeric-metric">₪{formatCurrency(todayExpenses)}</span>}
            icon={DollarSign}
            subtitle="مصروفات التشغيل المسجلة"
            variant="warning"
            onClick={() => navigate('/app/expenses')}
          />
          <StatCard
            title="حصيلة حركة الأموال اليوم"
            value={<span className="numeric-metric">₪{formatCurrency(todayCashDifference)}</span>}
            icon={todayCashDifference >= 0 ? DollarSign : AlertTriangle}
            subtitle="ما دخل إلى المتجر ناقص ما خرج منه"
            variant={todayCashDifference >= 0 ? 'success' : 'danger'}
          />
        </div>
      </section>
    </div>
  );
}
