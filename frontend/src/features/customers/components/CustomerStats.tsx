import { StatCard } from '../../../components/ui/stat-card';
import { Button } from '../../../components/ui/button';
import { getButtonSize } from '../../../config/button-sizes';
import { Users, UserPlus, DollarSign, Shield, Heart } from 'lucide-react';
import { CustomerStats } from '../types/customers.types';

interface CustomerStatsProps {
  stats: CustomerStats;
  onRecommendationClick: () => void;
}

export function CustomerStats({ stats }: CustomerStatsProps) {
  return (
    <>
      {/* AI Customer Insight */}
      <div className="premium-insight">
        <div className="premium-insight-icon"><Heart className="h-3.5 w-3.5" /></div>
        <div className="premium-insight-copy">
          <p className="premium-insight-title">AI Customer Insight · قيد التطوير</p>
          <p className="premium-insight-text">ستوفر تحليلات ذكية للعملاء وتوصيات لتحسين العلاقات وزيادة المبيعات.</p>
        </div>
        <Button variant="secondary" size={getButtonSize('customers', 'recommendation')} disabled className="premium-insight-action">قيد التطوير</Button>
      </div>

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '10px' }}
           className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard 
          title="إجمالي العملاء" 
          value={stats.totalCustomers} 
          icon={Users}
          subtitle="العملاء المسجلين"
          variant="featured"
        />
        <StatCard 
          title="العملاء النشطين" 
          value={stats.activeCustomers} 
          icon={UserPlus}
          subtitle="قاموا بشراء"
          variant="default"
        />
        <StatCard 
          title="عملاء بديون" 
          value={stats.customersWithDebt} 
          icon={DollarSign}
          subtitle="ديون مستحقة"
          variant="warning"
        />
        <StatCard 
          title="إجمالي الديون" 
          value={`₪${(stats.totalOutstanding || 0).toLocaleString()}`} 
          icon={Shield}
          subtitle="المبالغ المستحقة"
          variant="danger"
        />
      </div>
    </>
  );
}