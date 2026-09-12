import { StatCard } from '../../../components/ui/stat-card';
import { useNavigate } from 'react-router-dom';
import { Button } from '../../../components/ui/button';
import { getButtonSize } from '../../../config/button-sizes';
import { Users, UserPlus, DollarSign, Shield, Sparkles } from 'lucide-react';
import { CustomerStats } from '../types/customers.types';

interface CustomerStatsProps {
  stats: CustomerStats;
}

export function CustomerStats({ stats }: CustomerStatsProps) {
  const navigate = useNavigate();
  const inactiveCustomers = Math.max(0, stats.totalCustomers - stats.activeCustomers);
  const hasDebtInsight = stats.customersWithDebt > 0 && stats.totalOutstanding > 0;
  const debtCustomerLabel = stats.customersWithDebt === 1 ? 'عميل لديه' : 'عملاء لديهم';
  const insight = hasDebtInsight
    ? {
        title: 'تحليل العملاء · متابعة التحصيل',
        text: `لديك ${stats.customersWithDebt} ${debtCustomerLabel} ديون مستحقة بقيمة ₪${stats.totalOutstanding.toLocaleString()}. يُفضّل متابعة التحصيل.`,
        action: 'عرض الديون',
        onClick: () => navigate('/app/debts'),
      }
    : {
        title: 'تحليل العملاء · فرصة تنشيط',
        text: inactiveCustomers > 0
          ? `يوجد ${inactiveCustomers} عميل لم يسجلوا مشتريات بعد. جرّب التواصل معهم لتحويلهم إلى عملاء نشطين.`
          : 'جميع العملاء المسجلين نشطون. استمر في متابعة المبيعات والعلاقات.',
        action: 'عرض العملاء',
        onClick: () => navigate('/app/customers'),
      };

  return (
    <>
      <div className="premium-insight">
        <div className="premium-insight-icon"><Sparkles className="h-3.5 w-3.5" /></div>
        <div className="premium-insight-copy">
          <p className="premium-insight-title">{insight.title}</p>
          <p className="premium-insight-text">{insight.text}</p>
        </div>
        <Button
          variant="secondary"
          size={getButtonSize('customers', 'recommendation')}
          onClick={insight.onClick}
          className="premium-insight-action"
        >
          {insight.action}
        </Button>
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