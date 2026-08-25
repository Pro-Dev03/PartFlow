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
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '12px' }}
         className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="مبيعات اليوم"
        value={<span className="numeric-metric">₪{(stats?.todaySales as number)?.toLocaleString() || 0}</span>}
        icon={ShoppingCart}
        trend={stats?.salesTrend}
        trendUp={stats?.salesTrendUp}
        subtitle="مقارنة بالأمس"
        variant="featured"
      />
      <StatCard
        title="الربح اليوم"
        value={<span className="numeric-metric">₪{(stats?.todayProfit as number)?.toLocaleString() || 0}</span>}
        icon={DollarSign}
        trend={stats?.profitTrend}
        trendUp={stats?.profitTrendUp}
        subtitle={stats?.profitMargin ? `هامش الربح: ${stats.profitMargin}%` : undefined}
      />
      <StatCard
        title="الديون المستحقة"
        value={<span className="numeric-metric">₪{(stats?.outstandingDebts as number)?.toLocaleString() || 0}</span>}
        icon={AlertTriangle}
        trend={stats?.debtsTrend}
        trendUp={stats?.debtsTrendUp}
        subtitle={<span className="numeric-quantity">{(stats?.activeCustomers as number) || 0} عميل</span>}
        variant="warning"
      />
      <StatCard
        title="المخزون المنخفض"
        value={<span className="numeric-quantity">{(stats?.lowStockCount as number) || 0}</span>}
        icon={Package}
        trend={null}
        trendUp={null}
        subtitle="يحتاج انتباه"
        variant="danger"
      />
    </div>
  );
}