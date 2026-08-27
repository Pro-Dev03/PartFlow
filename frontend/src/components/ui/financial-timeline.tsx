import { Card, CardContent, CardHeader, CardTitle } from './card';
import { ShoppingCart, DollarSign, RotateCcw, RefreshCw, AlertCircle } from 'lucide-react';

export interface LedgerEntry {
  id: string;
  transaction_type?: string;
  type?: string;
  amount?: number;
  balance?: number;
  previous_balance?: number;
  balance_after?: number;
  description?: string;
  created_at?: string;
  date?: string;
  status?: string;
}

export interface FinancialTransaction {
  id: string;
  type: string;
  amount: number;
  balance_after: number;
  date: string;
  description: string;
  status: string;
}

interface FinancialTimelineProps {
  transactions?: FinancialTransaction[];
  entries?: LedgerEntry[];
  loading?: boolean;
  showBalance?: boolean;
  currency?: string;
}

export function FinancialTimeline({
  transactions,
  entries,
  loading,
  showBalance = true,
  currency = '₪',
}: FinancialTimelineProps) {
  const normalizedTransactions = ((transactions && transactions.length > 0 ? transactions : (entries || []).map((entry) => {
    const type = entry.transaction_type || entry.type || 'other';
    const amount = Number(entry.amount ?? 0);
    const balanceAfter = Number(entry.balance_after ?? entry.balance ?? entry.previous_balance ?? 0);
    const dateValue = entry.date || entry.created_at || new Date().toISOString();
    return {
      id: String(entry.id || `${type}-${dateValue}`),
      type,
      amount,
      balance_after: balanceAfter,
      date: dateValue,
      description: entry.description || 'حركة مالية',
      status: entry.status || 'completed',
    } satisfies FinancialTransaction;
  })) || []) as FinancialTransaction[];
  const getTransactionIcon = (type: string) => {
    switch (type) {
      case 'sale':
      case 'debit':
        return ShoppingCart;
      case 'payment':
      case 'credit':
        return DollarSign;
      case 'return':
        return RotateCcw;
      case 'refund':
        return RefreshCw;
      default:
        return AlertCircle;
    }
  };

  const getTransactionColor = (type: string) => {
    switch (type) {
      case 'sale':
      case 'debit':
        return 'var(--color-info)';
      case 'payment':
      case 'credit':
        return 'var(--color-success)';
      case 'return':
      case 'refund':
        return 'var(--color-warning)';
      default:
        return 'var(--color-text-secondary)';
    }
  };

  const getTransactionBackground = (type: string) => {
    switch (type) {
      case 'sale':
      case 'debit':
        return 'var(--color-info-10)';
      case 'payment':
      case 'credit':
        return 'var(--color-success-10)';
      case 'return':
      case 'refund':
        return 'var(--color-warning-10)';
      default:
        return 'var(--color-surface-elevated)';
    }
  };

  const formatAmount = (amount: number, type: string) => {
    const isCredit = type === 'payment' || type === 'credit' || type === 'refund';
    return `${isCredit ? '+' : '-'}${currency}${amount.toLocaleString()}`;
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ar-EG', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>السجل المالي</CardTitle>
        </CardHeader>
        <CardContent>
          <div style={{
            padding: 'var(--spacing-6)',
            textAlign: 'center',
            color: 'var(--text-secondary)',
            minHeight: '120px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}>
            <p>جاري تحميل السجل المالي...</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (normalizedTransactions.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>السجل المالي</CardTitle>
        </CardHeader>
        <CardContent>
          <div style={{ 
            padding: 'var(--spacing-6)', 
            textAlign: 'center',
            color: 'var(--text-secondary)'
          }}>
            <p>لا توجد حركات مالية مسجلة</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>السجل المالي</CardTitle>
      </CardHeader>
      <CardContent>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
          {normalizedTransactions.map((transaction, index) => {
            const Icon = getTransactionIcon(transaction.type);
            const color = getTransactionColor(transaction.type);
            const background = getTransactionBackground(transaction.type);
            
            return (
              <div
                key={transaction.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 'var(--spacing-3)',
                  padding: 'var(--spacing-4)',
                  borderRadius: 'var(--radius-md)',
                  background: index === 0 ? 'var(--color-primary-08)' : 'var(--bg-surface-elevated)',
                  border: index === 0 ? '1px solid var(--color-primary-20)' : '1px solid var(--border-default)',
                  transition: '180ms ease'
                }}
              >
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0"
                  style={{ background }}
                >
                  <Icon className="w-5 h-5" style={{ color }} />
                </div>
                
                <div style={{ flex: 1 }}>
                  <p style={{ 
                    fontSize: 'var(--font-size-secondary)', 
                    fontWeight: 'var(--font-weight-medium)', 
                    color: 'var(--text-primary)' 
                  }}>
                    {transaction.description}
                  </p>
                  <p style={{ 
                    fontSize: 'var(--font-size-caption)', 
                    color: 'var(--text-secondary)',
                    marginTop: '2px'
                  }}>
                    {formatDate(transaction.date)}
                  </p>
                </div>
                
                <div style={{ textAlign: 'right' }}>
                  <p style={{ 
                    fontSize: 'var(--font-size-secondary)', 
                    fontWeight: 'var(--font-weight-semibold)', 
                    color 
                  }}>
                    {formatAmount(transaction.amount, transaction.type)}
                  </p>
                  {showBalance && (
                    <p style={{ 
                      fontSize: 'var(--font-size-caption)', 
                      color: 'var(--text-secondary)',
                      marginTop: '2px'
                    }}>
                      الرصيد: {currency}{Number(transaction.balance_after ?? 0).toLocaleString()}
                    </p>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}