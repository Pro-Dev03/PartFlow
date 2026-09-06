import { StatCard } from '../../../components/ui/stat-card';
import { ShoppingCart, DollarSign, AlertTriangle, Package } from 'lucide-react';

interface DashboardMetricsProps {
  stats?: {
    todaySales?: number;
    todayProfit?: number;
    outstandingDebts?: number;
    activeCustomers?: number;
    lowStockCount?: number;
  };
}

export function DashboardMetrics({ stats }: DashboardMetricsProps) {
  // Dashboard stats are the authoritative live source. Aggregation endpoints may
  // contain stale snapshots while their background refresh is pending.
  const todaySales = Number(stats?.todaySales ?? 0);
  const todayProfit = Number(stats?.todayProfit ?? 0);
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
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '12px' }}
         className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="مبيعات اليوم"
        value={<span className="numeric-metric">₪{formatCurrency(todaySales)}</span>}
        icon={ShoppingCart}
        trend={stats?.salesTrend}
        trendUp={stats?.salesTrendUp}
        subtitle="مقارنة بالأمس"
        variant="featured"
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
      />
      <StatCard
        title="الديون المستحقة"
        value={<span className="numeric-metric">₪{formatCurrency(outstandingDebts)}</span>}
        icon={AlertTriangle}
        trend={stats?.debtsTrend}
        trendUp={stats?.debtsTrendUp}
        subtitle={<span className="numeric-quantity">{activeCustomers} عميل</span>}
        variant="warning"
      />
      <StatCard
        title="المخزون المنخفض"
        value={<span className="numeric-quantity">{lowStockCount}</span>}
        icon={Package}
        trend={null}
        trendUp={null}
        subtitle="يحتاج انتباه"
        variant="danger"
      />
    </div>
  );
}
