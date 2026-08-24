import { Card, CardContent, CardHeader, CardTitle } from './card';
import { Badge } from './badge';
import { cn } from '../../utils';

interface Movement {
  id: string;
  date: string;
  type: 'PURCHASE' | 'SALE' | 'RETURN' | 'ADJUSTMENT' | 'TRANSFER' | 'DAMAGE' | 'REPAIR' | 'RESERVATION' | 'RELEASE';
  quantity: number;
  beforeQuantity: number;
  afterQuantity: number;
  referenceType?: string;
  referenceId?: string;
  reason?: string;
  createdBy?: string;
}

interface InventoryLedgerProps {
  movements: Movement[];
  title?: string;
  currentStock?: number;
}

export function InventoryLedger({ movements, title = 'سجل حركات المخزون', currentStock }: InventoryLedgerProps) {
  const getMovementIcon = (type: Movement['type']) => {
    switch (type) {
      case 'PURCHASE':
        return '📥';
      case 'SALE':
        return '🛒';
      case 'RETURN':
        return '↩️';
      case 'ADJUSTMENT':
        return '⚙️';
      case 'TRANSFER':
        return '📦';
      case 'DAMAGE':
        return '⚠️';
      case 'REPAIR':
        return '🔧';
      case 'RESERVATION':
        return '🔒';
      case 'RELEASE':
        return '🔓';
      default:
        return '📋';
    }
  };

  const getMovementColor = (type: Movement['type']) => {
    switch (type) {
      case 'PURCHASE':
      case 'RETURN':
        return 'text-green';
      case 'SALE':
      case 'DAMAGE':
        return 'text-red';
      case 'ADJUSTMENT':
      case 'TRANSFER':
        return 'text-yellow';
      case 'REPAIR':
        return 'text-blue';
      case 'RESERVATION':
      case 'RELEASE':
        return 'text-purple';
      default:
        return 'text-text';
    }
  };

  const getMovementBadge = (type: Movement['type']) => {
    switch (type) {
      case 'PURCHASE':
        return { label: 'شراء', variant: 'success' as const };
      case 'SALE':
        return { label: 'بيع', variant: 'danger' as const };
      case 'RETURN':
        return { label: 'مرتجع', variant: 'warning' as const };
      case 'ADJUSTMENT':
        return { label: 'تعديل', variant: 'secondary' as const };
      case 'TRANSFER':
        return { label: 'نقل', variant: 'info' as const };
      case 'DAMAGE':
        return { label: 'تالف', variant: 'danger' as const };
      case 'REPAIR':
        return { label: 'إصلاح', variant: 'info' as const };
      case 'RESERVATION':
        return { label: 'حجز', variant: 'warning' as const };
      case 'RELEASE':
        return { label: 'إلغاء الحجز', variant: 'secondary' as const };
      default:
        return { label: 'أخرى', variant: 'default' as const };
    }
  };

  return (
    <Card>
      <CardHeader>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <CardTitle>{title}</CardTitle>
          {currentStock !== undefined && (
            <div style={{ textAlign: 'left' }}>
              <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginBottom: '2px' }}>
                المخزون الحالي
              </p>
              <p style={{ fontSize: '16px', fontWeight: '700', color: 'var(--text-primary)' }}>
                {currentStock}
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
            {movements.map((movement, index) => {
              const badge = getMovementBadge(movement.type);
              return (
                <div key={movement.id} style={{ position: 'relative' }}>
                  {/* Timeline Dot */}
                  <div style={{
                    position: 'absolute',
                    left: '-20px',
                    top: '4px',
                    width: '10px',
                    height: '10px',
                    borderRadius: '50%',
                    background: movement.type === 'PURCHASE' || movement.type === 'RETURN' ? 'var(--success)' :
                               movement.type === 'SALE' || movement.type === 'DAMAGE' ? 'var(--danger)' :
                               movement.type === 'ADJUSTMENT' || movement.type === 'TRANSFER' ? 'var(--warning)' :
                               movement.type === 'REPAIR' ? 'var(--info)' :
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
                        <span style={{ fontSize: '16px' }}>{getMovementIcon(movement.type)}</span>
                        <Badge variant={badge.variant} style={{ fontSize: '10px' }}>
                          {badge.label}
                        </Badge>
                        <span style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>
                          {new Date(movement.date).toLocaleDateString('ar-SA')}
                        </span>
                      </div>
                      <p style={{ fontSize: '13px', color: 'var(--text-primary)', marginBottom: '2px' }}>
                        {movement.reason || badge.label}
                      </p>
                      {movement.referenceType && (
                        <p style={{ fontSize: '11px', color: 'var(--text-tertiary)' }}>
                          المرجع: {movement.referenceType} #{movement.referenceId}
                        </p>
                      )}
                      <div style={{ display: 'flex', gap: '12px', marginTop: '4px', fontSize: '11px', color: 'var(--text-tertiary)' }}>
                        <span>قبل: {movement.beforeQuantity}</span>
                        <span>→</span>
                        <span>بعد: {movement.afterQuantity}</span>
                      </div>
                    </div>
                    <div style={{ textAlign: 'right', marginRight: '12px' }}>
                      <p style={{
                        fontSize: '14px',
                        fontWeight: '600',
                        color: movement.type === 'PURCHASE' || movement.type === 'RETURN' ? 'var(--success)' :
                               movement.type === 'SALE' || movement.type === 'DAMAGE' ? 'var(--danger)' :
                               'var(--text-primary)'
                      }}>
                        {movement.type === 'PURCHASE' || movement.type === 'RETURN' ? '+' : '-'}
                        {movement.quantity}
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
