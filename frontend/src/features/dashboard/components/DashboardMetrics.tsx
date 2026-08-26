import { StatCard } from '../../../components/ui/stat-card';
import { ShoppingCart, DollarSign, AlertTriangle, Package } from 'lucide-react';
import { DailySalesSummary, DailyInventorySummary, DailyDebtSummary, DailyProfitSummary } from '../../../types/api';

interface DashboardMetricsProps {
  stats?: {
    todaySales?: number;
    todayProfit?: number;
    outstandingDebts?: number;
    activeCustomers?: number;
    lowStockCount?: number;
  };
  // Aggregation data from ARCHITECTURE-PRINCIPLES.md
  dailySales?: DailySalesSummary;
  dailyInventory?: DailyInventorySummary;
  dailyDebt?: DailyDebtSummary;
  dailyProfit?: DailyProfitSummary;
}

export function DashboardMetrics({ stats, dailySales, dailyInventory, dailyDebt, dailyProfit }: DashboardMetricsProps) {
  // Use aggregation data when available (ARCHITECTURE-PRINCIPLES.md)
  const todaySales = dailySales?.total_revenue || stats?.todaySales || 0;
  const todayProfit = dailyProfit?.net_profit || stats?.todayProfit || 0;
  const profitMargin = dailyProfit?.profit_margin || stats?.profitMargin;
  const outstandingDebts = dailyDebt?.total_debt || stats?.outstandingDebts || 0;
  const activeCustomers = dailySales?.total_customers || stats?.activeCustomers || 0;
  const lowStockCount = dailyInventory?.low_stock_count || stats?.lowStockCount || 0;

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '12px' }}
         className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="مبيعات اليوم"
        value={<span className="numeric-metric">₪{todaySales.toLocaleString()}</span>}
        icon={ShoppingCart}
        trend={stats?.salesTrend}
        trendUp={stats?.salesTrendUp}
        subtitle="مقارنة بالأمس"
        variant="featured"
      />
      <StatCard
        title="الربح اليوم"
        value={<span className="numeric-metric">₪{todayProfit.toLocaleString()}</span>}
        icon={DollarSign}
        trend={stats?.profitTrend}
        trendUp={stats?.profitTrendUp}
        subtitle={profitMargin ? `هامش الربح: ${profitMargin}%` : undefined}
      />
      <StatCard
        title="الديون المستحقة"
        value={<span className="numeric-metric">₪{outstandingDebts.toLocaleString()}</span>}
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