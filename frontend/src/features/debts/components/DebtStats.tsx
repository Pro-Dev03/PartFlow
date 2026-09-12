import { StatCard } from '../../../components/ui/stat-card';
import { DollarSign, AlertTriangle, Calendar, Users } from 'lucide-react';
import { DebtStats } from '../types/debts.types';

interface DebtStatsProps {
  stats: DebtStats;
}

export function DebtStats({ stats }: DebtStatsProps) {
  return (
    <div className="unified-stats-grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="إجمالي الديون الأصلية"
        value={<span className="numeric-metric">₪{stats.totalDebt.toLocaleString()}</span>}
        icon={DollarSign}
        subtitle="قيمة الديون الأصلية على العملاء"
        variant="featured"
      />
      <StatCard
        title="المحصّل"
        value={<span className="numeric-metric">₪{stats.paidAmount.toLocaleString()}</span>}
        icon={Calendar}
        subtitle="ما تم تحصيله من هذه الديون"
        variant="success"
      />
      <StatCard
        title="غير مسدد"
        value={<span className="numeric-metric">₪{stats.remainingAmount.toLocaleString()}</span>}
        icon={AlertTriangle}
        subtitle="المبلغ الذي ما زال على العملاء"
        variant="warning"
      />
      <StatCard
        title="العملاء المدينون"
        value={<span className="numeric-quantity">{stats.customerCount}</span>}
        icon={Users}
        subtitle="لديهم مبلغ غير مسدد"
        variant="default"
      />
    </div>
  );
}
