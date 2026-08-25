import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './card';
import { Badge } from './badge';
import { Button } from './button';
import { 
  ArrowUpRight, 
  ArrowDownRight, 
  DollarSign, 
  ShoppingCart, 
  RotateCcw,
  Calendar,
  Filter
} from 'lucide-react';

export interface LedgerEntry {
  id: string;
  transaction_type: string;
  amount: number;
  balance: number;
  previous_balance: number;
  description: string;
  created_at: string;
  reference_id?: string;
  reference_type?: string;
}

interface FinancialTimelineProps {
  entries: LedgerEntry[];
  loading?: boolean;
  onLoadMore?: () => void;
  showBalance?: boolean;
}

export function FinancialTimeline({ 
  entries, 
  loading = false, 
  onLoadMore,
  showBalance = true 
}: FinancialTimelineProps) {
  const [filter, setFilter] = useState<string>('all');

  const getTransactionIcon = (type: string) => {
    switch (type) {
      case 'SALE':
        return <ShoppingCart className="w-4 h-4" />;
      case 'PAYMENT':
        return <DollarSign className="w-4 h-4" />;
      case 'RETURN':
      case 'REFUND':
        return <RotateCcw className="w-4 h-4" />;
      default:
        return <DollarSign className="w-4 h-4" />;
    }
  };

  const getTransactionLabel = (type: string) => {
    switch (type) {
      case 'SALE':
        return 'بيع';
      case 'PAYMENT':
        return 'دفعة';
      case 'RETURN':
        return 'مرتجع';
      case 'REFUND':
        return 'استرجاع';
      case 'ADJUSTMENT':
        return 'تعديل';
      default:
        return type;
    }
  };

  const getTransactionVariant = (type: string): 'success' | 'danger' | 'warning' | 'secondary' => {
    switch (type) {
      case 'SALE':
        return 'danger'; // Customer owes money
      case 'PAYMENT':
        return 'success'; // Payment received
      case 'RETURN':
      case 'REFUND':
        return 'warning';
      default:
        return 'secondary';
    }
  };

  const filteredEntries = filter === 'all' 
    ? entries 
    : entries.filter(entry => entry.transaction_type === filter);

  const transactionTypes = ['all', ...Array.from(new Set(entries.map(e => e.transaction_type)))];

  return (
    <Card>
      <CardHeader>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Calendar className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
            السجل المالي
          </CardTitle>
          
          {transactionTypes.length > 1 && (
            <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
              <Filter className="w-4 h-4" style={{ color: 'var(--text-secondary)' }} />
              <select
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
                style={{
                  padding: '6px 12px',
                  borderRadius: '6px',
                  border: '1px solid var(--border-default)',
                  background: 'var(--bg-surface)',
                  color: 'var(--text-primary)',
                  fontSize: '13px'
                }}
              >
                {transactionTypes.map(type => (
                  <option key={type} value={type}>
                    {type === 'all' ? 'الكل' : getTransactionLabel(type)}
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: '40px' }}>
            <div style={{ 
              animation: 'spin 1s linear infinite', 
              borderRadius: '50%', 
              height: '32px', 
              width: '32px', 
              borderBottom: '2px solid var(--color-primary)' 
            }} />
          </div>
        ) : filteredEntries.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '40px 20px' }}>
            <DollarSign className="w-12 h-12 mx-auto mb-2" style={{ color: 'var(--text-secondary)' }} />
            <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              لا توجد حركات مالية
            </p>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {filteredEntries.map((entry, index) => (
              <div
                key={entry.id}
                style={{
                  display: 'flex',
                  alignItems: 'flex-start',
                  gap: '14px',
                  padding: '16px',
                  borderRadius: '10px',
                  border: '1px solid var(--border-default)',
                  background: 'var(--bg-surface)',
                  transition: '180ms ease'
                }}
              >
                <div style={{
                  width: '40px',
                  height: '40px',
                  borderRadius: '8px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: entry.amount > 0 
                    ? 'rgba(251, 113, 133, 0.1)' 
                    : 'rgba(52, 211, 153, 0.1)',
                  flexShrink: 0
                }}>
                  <div style={{ 
                    color: entry.amount > 0 
                      ? 'var(--color-danger)' 
                      : 'var(--color-success)' 
                  }}>
                    {getTransactionIcon(entry.transaction_type)}
                  </div>
                </div>
                
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '4px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <Badge variant={getTransactionVariant(entry.transaction_type)} style={{ fontSize: '11px' }}>
                        {getTransactionLabel(entry.transaction_type)}
                      </Badge>
                      <span style={{ fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>
                        {entry.description}
                      </span>
                    </div>
                    <span style={{ fontSize: '13px', color: 'var(--text-secondary)', whiteSpace: 'nowrap' }}>
                      {new Date(entry.created_at).toLocaleDateString('ar-SA')}
                    </span>
                  </div>
                  
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                      {entry.amount > 0 ? (
                        <ArrowUpRight className="w-3.5 h-3.5" style={{ color: 'var(--color-danger)' }} />
                      ) : (
                        <ArrowDownRight className="w-3.5 h-3.5" style={{ color: 'var(--color-success)' }} />
                      )}
                      <span 
                        className="numeric-price"
                        style={{ 
                          fontSize: '14px', 
                          fontWeight: '600',
                          color: entry.amount > 0 
                            ? 'var(--color-danger)' 
                            : 'var(--color-success)' 
                        }}
                      >
                        {entry.amount > 0 ? '+' : ''}₪{Math.abs(entry.amount).toLocaleString()}
                      </span>
                    </div>
                    
                    {showBalance && (
                      <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                        الرصيد: <span className="numeric-price" style={{ fontWeight: '500' }}>₪{entry.balance.toLocaleString()}</span>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
            
            {onLoadMore && (
              <div style={{ textAlign: 'center', marginTop: '8px' }}>
                <Button variant="secondary" onClick={onLoadMore}>
                  تحميل المزيد
                </Button>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
