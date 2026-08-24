import { StatCard } from '../../../components/ui/stat-card';
import { DollarSign, TrendingUp, Database, Target } from 'lucide-react';

export function ReportStats() {
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '14px' }}
         className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard 
        title="إجمالي المبيعات" 
        value="₪45,230" 
        icon={DollarSign}
        subtitle="للفترة المحددة"
        variant="featured"
        trend="+15.3%"
        trendUp={true}
      />
      <StatCard 
        title="إجمالي الأرباح" 
        value="₪12,450" 
        icon={TrendingUp}
        subtitle="هامش الربح"
        variant="success"
        trend="+8.7%"
        trendUp={true}
      />
      <StatCard 
        title="عدد المعاملات" 
        value="234" 
        icon={Database}
        subtitle="عمليات بيع"
        variant="default"
        trend="+12.1%"
        trendUp={true}
      />
      <StatCard 
        title="متوسط الطلب" 
        value="₪193" 
        icon={Target}
        subtitle="لكل معاملة"
        variant="info"
        trend="+5.4%"
        trendUp={true}
      />
    </div>
  );
}
