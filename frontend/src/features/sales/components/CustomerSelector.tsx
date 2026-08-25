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
    <div style={{
      background: 'var(--bg-surface-elevated)',
      border: '1px solid var(--border-default)',
      borderRadius: '12px',
      padding: '16px',
      transition: 'all 0.3s ease'
    }}
    onMouseEnter={(e) => {
      e.currentTarget.style.borderColor = 'var(--border-primary)';
      e.currentTarget.style.boxShadow = '0 4px 12px rgba(0, 0, 0, 0.1)';
    }}
    onMouseLeave={(e) => {
      e.currentTarget.style.borderColor = 'var(--border-default)';
      e.currentTarget.style.boxShadow = 'none';
    }}>
      <div style={{ 
        display: 'flex', 
        alignItems: 'center', 
        gap: '8px',
        marginBottom: '12px'
      }}>
        <User className="w-4 h-4" style={{ color: 'var(--text-secondary)' }} />
        <span style={{ 
          fontSize: '13px', 
          fontWeight: '600',
          color: 'var(--text-primary)'
        }}>
          العميل
        </span>
      </div>
      <Select
        value={selectedCustomer}
        onChange={(e) => setSelectedCustomer(e.target.value)}
        loading={customersLoading}
        options={[
          { value: '', label: 'عميل نقدي' },
          ...customers.map((c: any) => ({ value: c.id, label: c.name })),
        ]}
        emptyMessage="لا يوجد عملاء"
        style={{ borderRadius: '8px' }}
      />
    </div>
  );
}