import { StatCard } from '../../../design-system/components/stat-card';
import { useNavigate } from 'react-router-dom';
import { Users, UserPlus, DollarSign, Shield } from 'lucide-react';
import { CustomerStats } from '../types/customers.types';

interface CustomerStatsProps {
  stats: CustomerStats;
}

export function CustomerStats({ stats }: CustomerStatsProps) {
  const navigate = useNavigate();

  return (
    <>
      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '10px' }}
           className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard 
          title="إجمالي الزبائن" 
          value={stats.totalCustomers} 
          icon={Users}
          subtitle="الزبائن المسجلين"
          variant="featured"
        />
        <StatCard 
          title="الزبائن النشطين" 
          value={stats.activeCustomers} 
          icon={UserPlus}
          subtitle="قاموا بشراء"
          variant="default"
        />
        <StatCard 
          title="زبائن بديون" 
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