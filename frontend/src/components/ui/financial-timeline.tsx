import { Card, CardContent, CardHeader, CardTitle } from './card';
import { ShoppingCart, DollarSign, RotateCcw, RefreshCw, AlertCircle } from 'lucide-react';
import { formatStoreDateTime } from '../../utils/store-time';

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
  showTitle?: boolean;
  currency?: string;
}

export function FinancialTimeline({
  transactions,
  entries,
  loading,
  showBalance = true,
  showTitle = true,
  currency = '₪',
}: FinancialTimelineProps) {
  const normalizedTransactions = ((transactions && transactions.length > 0 ? transactions : (entries || []).map((entry) => {
    const type = String(entry.transaction_type || entry.type || 'other').toLowerCase();
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
  const normalizeTransactionType = (type: string) => {
    const normalizedType = String(type || '').trim().toLowerCase();
    if (normalizedType === 'sale') return 'sale';
    if (normalizedType === 'payment') return 'payment';
    if (normalizedType === 'debit') return 'debit';
    if (normalizedType === 'credit') return 'credit';
    if (normalizedType === 'refund') return 'refund';
    if (normalizedType === 'return') return 'return';
    return normalizedType;
  };
  const getTransactionIcon = (type: string) => {
    switch (normalizeTransactionType(type)) {
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
    switch (normalizeTransactionType(type)) {
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
    switch (normalizeTransactionType(type)) {
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
    const normalizedType = normalizeTransactionType(type);
    const formattedAmount = `${currency}${Math.abs(amount).toLocaleString()}`;
    if (normalizedType === 'payment' || normalizedType === 'credit') {
      return `المبلغ المدفوع: ${formattedAmount}`;
    }
    if (normalizedType === 'refund' || normalizedType === 'return') {
      return `قيمة المرتجع: ${formattedAmount}`;
    }
    if (normalizedType === 'sale' || normalizedType === 'debit') {
      return `قيمة البيع: ${formattedAmount}`;
    }
    return `المبلغ: ${formattedAmount}`;
  };

  const formatDescription = (description: string, type: string) => {
    const normalizedType = normalizeTransactionType(type);
    if (normalizedType === 'debit' || normalizedType === 'sale') {
      return description.replace(/^Sale:\s*/i, 'فاتورة بيع: ');
    }
    if (normalizedType === 'credit' || normalizedType === 'payment') {
      return description.replace(/^Payment:\s*cash$/i, 'دفعة نقدية').replace(/^Payment:\s*/i, 'دفعة: ');
    }
    return description;
  };

  const formatDate = (dateString: string) => {
    return formatStoreDateTime(dateString, 'ar-EG');
  };

  if (loading) {
    return (
      <Card>
        {showTitle && <CardHeader><CardTitle>السجل المالي</CardTitle></CardHeader>}
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
        {showTitle && <CardHeader><CardTitle>السجل المالي</CardTitle></CardHeader>}
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
      {showTitle && <CardHeader><CardTitle>السجل المالي</CardTitle></CardHeader>}
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
                    {formatDescription(transaction.description, transaction.type)}
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
                      الرصيد المستحق بعد الحركة: {currency}{Number(transaction.balance_after ?? 0).toLocaleString()}
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