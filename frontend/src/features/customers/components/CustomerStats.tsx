import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { StatCard } from '../../../components/ui/stat-card';
import { Button } from '../../../components/ui/button';
import { getButtonSize } from '../../../config/button-sizes';
import { Users, UserPlus, DollarSign, Shield, Sparkles, Heart } from 'lucide-react';
import { CustomerStats } from '../types/customers.types';

interface CustomerStatsProps {
  stats: CustomerStats;
  onRecommendationClick: () => void;
}

export function CustomerStats({ stats }: CustomerStatsProps) {
  return (
    <>
      {/* AI Customer Insight */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Sparkles style={{ width: '20px', height: '20px', color: 'var(--primary)' }} />
            AI Customer Insight
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div style={{ display: 'flex', gap: '14px' }}>
            <div style={{
              width: '32px',
              height: '32px',
              borderRadius: '8px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: 'rgba(34, 211, 238, 0.1)',
              flexShrink: 0
            }}>
              <Heart style={{ width: '16px', height: '16px', color: 'var(--primary)' }} />
            </div>
            <div>
              <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>قيد التطوير</p>
              <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                هذه الميزة قيد التطوير حالياً. ستوفر تحليلات ذكية للعملاء وتوصيات لتحسين العلاقات وزيادة المبيعات.
              </p>
              <Button variant="secondary" size={getButtonSize('customers', 'recommendation')} disabled>
                قيد التطوير
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '14px' }}
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