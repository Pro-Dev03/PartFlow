import { StatCard } from '../../../design-system/components/stat-card';
import { DollarSign, AlertTriangle, Calendar, Users } from 'lucide-react';
import { DebtStats } from '../types/debts.types';

interface DebtStatsProps {
  stats: DebtStats;
}

export function DebtStats({ stats }: DebtStatsProps) {
  return (
    <div
      style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(190px, 1fr))', gap: '10px' }}
      className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4"
    >
      <StatCard
        title="إجمالي الديون الأصلية"
        value={`₪${stats.totalDebt.toLocaleString()}`}
        icon={DollarSign}
        subtitle="قيمة الديون الأصلية على العملاء"
        variant="featured"
        compact
      />
      <StatCard
        title="التحصيل"
        value={`₪${stats.paidAmount.toLocaleString()}`}
        icon={Calendar}
        subtitle="ما تم تحصيله من هذه الديون"
        variant="success"
        compact
      />
      <StatCard
        title="غير مسدد"
        value={`₪${stats.remainingAmount.toLocaleString()}`}
        icon={AlertTriangle}
        subtitle="المبلغ الذي ما زال على العملاء"
        variant="warning"
        compact
      />
      <StatCard
        title="العملاء المدينون"
        value={stats.customerCount}
        icon={Users}
        subtitle="لديهم مبلغ غير مسدد"
        variant="default"
        compact
      />
    </div>
  );
}
