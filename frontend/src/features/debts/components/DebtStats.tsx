import { StatCard } from '../../../components/ui/stat-card';
import { DollarSign, AlertTriangle, Calendar, Users } from 'lucide-react';
import { DebtStats } from '../types/debts.types';

interface DebtStatsProps {
  stats: DebtStats;
}

export function DebtStats({ stats }: DebtStatsProps) {
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '14px' }}
         className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard 
        title="إجمالي الديون" 
        value={`₪${stats.totalDebt.toLocaleString()}`} 
        icon={DollarSign}
        subtitle="المبلغ الكلي"
        variant="featured"
      />
      <StatCard 
        title="المسدد" 
        value={`₪${stats.paidAmount.toLocaleString()}`} 
        icon={Calendar}
        subtitle="تم السداد"
        variant="success"
      />
      <StatCard 
        title="المتبقي" 
        value={`₪${stats.remainingAmount.toLocaleString()}`} 
        icon={AlertTriangle}
        subtitle="لم يسدد"
        variant="warning"
      />
      <StatCard 
        title="العملاء المدينين" 
        value={stats.customerCount} 
        icon={Users}
        subtitle="عملاء"
        variant="default"
      />
    </div>
  );
}
