import { StatCard } from '../../../components/ui/stat-card';
import { DollarSign, AlertTriangle, Calendar, Users } from 'lucide-react';
import { DebtStats } from '../types/debts.types';

interface DebtStatsProps {
  stats: DebtStats;
}

export function DebtStats({ stats }: DebtStatsProps) {
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '16px' }}
         className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="إجمالي الديون"
        value={<span className="numeric-metric">₪{stats.totalDebt.toLocaleString()}</span>}
        icon={DollarSign}
        subtitle="المبلغ الكلي"
        variant="featured"
      />
      <StatCard
        title="المسدد"
        value={<span className="numeric-metric">₪{stats.paidAmount.toLocaleString()}</span>}
        icon={Calendar}
        subtitle="تم السداد"
        variant="success"
      />
      <StatCard
        title="المتبقي"
        value={<span className="numeric-metric">₪{stats.remainingAmount.toLocaleString()}</span>}
        icon={AlertTriangle}
        subtitle="لم يسدد"
        variant="warning"
      />
      <StatCard
        title="العملاء المدينين"
        value={<span className="numeric-quantity">{stats.customerCount}</span>}
        icon={Users}
        subtitle="عملاء"
        variant="default"
      />
    </div>
  );
}
