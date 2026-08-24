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
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '14px' }}
         className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="مبيعات اليوم"
        value={`₪${(stats?.todaySales as number)?.toLocaleString() || 0}`}
        icon={ShoppingCart}
        trend="+12.4%"
        trendUp={true}
        subtitle="مقارنة بالأمس"
        variant="featured"
      />
      <StatCard
        title="الربح اليوم"
        value={`₪${(stats?.todayProfit as number)?.toLocaleString() || 0}`}
        icon={DollarSign}
        trend="+15%"
        trendUp={true}
        subtitle="هامش الربح: 24%"
      />
      <StatCard
        title="الديون المستحقة"
        value={`₪${(stats?.outstandingDebts as number)?.toLocaleString() || 0}`}
        icon={AlertTriangle}
        trend="+5%"
        trendUp={false}
        subtitle={`${(stats?.activeCustomers as number) || 0} عميل`}
        variant="warning"
      />
      <StatCard
        title="المخزون المنخفض"
        value={(stats?.lowStockCount as number) || 0}
        icon={Package}
        trend={null}
        trendUp={null}
        subtitle="يحتاج انتباه"
        variant="danger"
      />
    </div>
  );
}