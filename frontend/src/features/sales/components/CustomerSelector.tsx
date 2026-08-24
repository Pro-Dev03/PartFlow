import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Select } from '../../../components/ui/select';
import { User } from 'lucide-react';

interface CustomerSelectorProps {
  selectedCustomer: string;
  setSelectedCustomer: (value: string) => void;
  customers: any[];
  customersLoading: boolean;
}

export function CustomerSelector({
  selectedCustomer,
  setSelectedCustomer,
  customers,
  customersLoading,
}: CustomerSelectorProps) {
  return (
    <Card style={{
      background: 'linear-gradient(135deg, rgba(168, 85, 247, 0.05) 0%, rgba(236, 72, 153, 0.05) 100%)',
      border: '1px solid rgba(168, 85, 247, 0.1)',
      backdropFilter: 'blur(10px)',
      transition: 'all 0.3s ease'
    }}
    onMouseEnter={(e) => {
      e.currentTarget.style.background = 'linear-gradient(135deg, rgba(168, 85, 247, 0.08) 0%, rgba(236, 72, 153, 0.08) 100%)';
      e.currentTarget.style.borderColor = 'rgba(168, 85, 247, 0.2)';
      e.currentTarget.style.transform = 'translateY(-2px)';
      e.currentTarget.style.boxShadow = '0 8px 25px rgba(168, 85, 247, 0.1)';
    }}
    onMouseLeave={(e) => {
      e.currentTarget.style.background = 'linear-gradient(135deg, rgba(168, 85, 247, 0.05) 0%, rgba(236, 72, 153, 0.05) 100%)';
      e.currentTarget.style.borderColor = 'rgba(168, 85, 247, 0.1)';
      e.currentTarget.style.transform = 'translateY(0)';
      e.currentTarget.style.boxShadow = 'none';
    }}>
      <CardHeader>
        <CardTitle style={{ 
          display: 'flex', 
          alignItems: 'center', 
          gap: '10px',
          fontSize: '15px',
          fontWeight: '600',
          color: 'var(--text-primary)'
        }}>
          <div style={{
            width: '36px',
            height: '36px',
            borderRadius: '10px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'linear-gradient(135deg, rgba(168, 85, 247, 0.15) 0%, rgba(236, 72, 153, 0.15) 100%)',
            border: '1px solid rgba(168, 85, 247, 0.25)'
          }}>
            <User className="w-5 h-5 text-purple-400" />
          </div>
          العميل
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Select
          value={selectedCustomer}
          onChange={(e) => setSelectedCustomer(e.target.value)}
          loading={customersLoading}
          options={[
            { value: '', label: 'عميل نقدي' },
            ...customers.map((c: any) => ({ value: c.id, label: c.name })),
          ]}
          emptyMessage="لا يوجد عملاء"
        />
      </CardContent>
    </Card>
  );
}