import { StatCard } from '../../../components/ui/stat-card';
import { DollarSign, TrendingUp, Database, Target } from 'lucide-react';

interface ReportStatsProps {
  data?: any;
  loading?: boolean;
}

export function ReportStats({ data, loading }: ReportStatsProps) {
  // Calculate stats from actual data
  const calculateStats = () => {
    if (!data || !data.data || data.data.length === 0) {
      return {
        totalSales: 0,
        totalProfit: 0,
        totalTransactions: 0,
        averageOrder: 0,
      };
    }

    const items = data.data;
    const totalSales = items.reduce((sum: number, item: any) => sum + (item.value || item.amount || 0), 0);
    const totalProfit = items.reduce((sum: number, item: any) => sum + (item.profit || 0), 0);
    const totalTransactions = items.length;
    const averageOrder = totalTransactions > 0 ? totalSales / totalTransactions : 0;

    return {
      totalSales,
      totalProfit,
      totalTransactions,
      averageOrder,
    };
  };

  const stats = calculateStats();

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '14px' }}
         className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard 
        title="إجمالي المبيعات" 
        value={loading ? '...' : `₪${stats.totalSales.toLocaleString()}`}
        icon={DollarSign}
        subtitle="للفترة المحددة"
        variant="featured"
      />
      <StatCard 
        title="إجمالي الأرباح" 
        value={loading ? '...' : `₪${stats.totalProfit.toLocaleString()}`}
        icon={TrendingUp}
        subtitle="هامش الربح"
        variant="success"
      />
      <StatCard 
        title="عدد المعاملات" 
        value={loading ? '...' : stats.totalTransactions.toString()}
        icon={Database}
        subtitle="عمليات بيع"
        variant="default"
      />
      <StatCard 
        title="متوسط الطلب" 
        value={loading ? '...' : `₪${Math.round(stats.averageOrder).toLocaleString()}`}
        icon={Target}
        subtitle="لكل معاملة"
        variant="info"
      />
    </div>
  );
}
