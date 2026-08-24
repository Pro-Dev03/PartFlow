import { Card, CardContent, CardHeader, CardTitle } from './card';
import { Badge } from './badge';
import { cn } from '../../utils';

interface TimelineItem {
  id: string;
  date: string;
  type: 'sale' | 'payment' | 'return' | 'refund' | 'adjustment';
  amount: number;
  description: string;
  balance?: number;
}

interface FinancialTimelineProps {
  items: TimelineItem[];
  title?: string;
  currentBalance?: number;
}

export function FinancialTimeline({ items, title = 'السجل المالي', currentBalance }: FinancialTimelineProps) {
  const getTimelineIcon = (type: TimelineItem['type']) => {
    switch (type) {
      case 'sale':
        return '🛒';
      case 'payment':
        return '💰';
      case 'return':
        return '↩️';
      case 'refund':
        return '💸';
      case 'adjustment':
        return '⚙️';
      default:
        return '📋';
    }
  };

  const getTimelineColor = (type: TimelineItem['type']) => {
    switch (type) {
      case 'sale':
        return 'text-green';
      case 'payment':
        return 'text-blue';
      case 'return':
        return 'text-yellow';
      case 'refund':
        return 'text-red';
      case 'adjustment':
        return 'text-gray';
      default:
        return 'text-text';
    }
  };

  const getTimelineBadge = (type: TimelineItem['type']) => {
    switch (type) {
      case 'sale':
        return { label: 'بيع', variant: 'success' as const };
      case 'payment':
        return { label: 'دفعة', variant: 'info' as const };
      case 'return':
        return { label: 'مرتجع', variant: 'warning' as const };
      case 'refund':
        return { label: 'استرجاع', variant: 'danger' as const };
      case 'adjustment':
        return { label: 'تعديل', variant: 'secondary' as const };
      default:
        return { label: 'أخرى', variant: 'default' as const };
    }
  };

  return (
    <Card>
      <CardHeader>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <CardTitle>{title}</CardTitle>
          {currentBalance !== undefined && (
            <div style={{ textAlign: 'left' }}>
              <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginBottom: '2px' }}>
                الرصيد الحالي
              </p>
              <p style={{ fontSize: '16px', fontWeight: '700', color: 'var(--text-primary)' }}>
                ₪{currentBalance.toLocaleString()}
              </p>
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <div style={{ position: 'relative', paddingLeft: '24px' }}>
          {/* Timeline Line */}
          <div style={{
            position: 'absolute',
            left: '8px',
            top: '8px',
            bottom: '8px',
            width: '2px',
            background: 'var(--border-default)',
            borderRadius: '1px'
          }} />

          {/* Timeline Items */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {items.map((item, index) => {
              const badge = getTimelineBadge(item.type);
              return (
                <div key={item.id} style={{ position: 'relative' }}>
                  {/* Timeline Dot */}
                  <div style={{
                    position: 'absolute',
                    left: '-20px',
                    top: '4px',
                    width: '10px',
                    height: '10px',
                    borderRadius: '50%',
                    background: item.type === 'sale' ? 'var(--success)' :
                               item.type === 'payment' ? 'var(--info)' :
                               item.type === 'return' ? 'var(--warning)' :
                               item.type === 'refund' ? 'var(--danger)' :
                               'var(--text-secondary)',
                    border: '2px solid var(--bg-surface)',
                    zIndex: 1
                  }} />

                  {/* Timeline Content */}
                  <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'flex-start',
                    padding: '12px',
                    background: 'var(--bg-surface-elevated)',
                    borderRadius: '8px',
                    border: '1px solid var(--border-default)'
                  }}>
                    <div style={{ flex: 1 }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                        <span style={{ fontSize: '16px' }}>{getTimelineIcon(item.type)}</span>
                        <Badge variant={badge.variant} style={{ fontSize: '10px' }}>
                          {badge.label}
                        </Badge>
                        <span style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>
                          {new Date(item.date).toLocaleDateString('ar-SA')}
                        </span>
                      </div>
                      <p style={{ fontSize: '13px', color: 'var(--text-primary)', marginBottom: '2px' }}>
                        {item.description}
                      </p>
                      {item.balance !== undefined && (
                        <p style={{ fontSize: '11px', color: 'var(--text-tertiary)' }}>
                          الرصيد: ₪{item.balance.toLocaleString()}
                        </p>
                      )}
                    </div>
                    <div style={{ textAlign: 'right', marginRight: '12px' }}>
                      <p style={{
                        fontSize: '14px',
                        fontWeight: '600',
                        color: item.type === 'sale' || item.type === 'return' ? 'var(--success)' :
                               item.type === 'payment' || item.type === 'refund' ? 'var(--danger)' :
                               'var(--text-primary)'
                      }}>
                        {item.type === 'sale' || item.type === 'return' ? '+' : '-'}
                        ₪{item.amount.toLocaleString()}
                      </p>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
