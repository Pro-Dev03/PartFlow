import { Card, CardContent, CardHeader, CardTitle } from './card';
import { Badge } from './badge';

export interface InventoryMovement {
  id: string;
  date: string;
  type: 'PURCHASE' | 'SALE' | 'RETURN' | 'ADJUSTMENT' | 'TRANSFER' | 'DAMAGE' | 'REPAIR' | 'RESERVATION' | 'RELEASE' |
    'REVERSE_PURCHASE' | 'REVERSE_SALE' | 'REVERSE_RETURN';
  quantity: number;
  beforeQuantity: number;
  afterQuantity: number;
  referenceType?: string;
  referenceId?: string;
  customerName?: string;
  invoiceNumber?: string;
  reason?: string;
  createdBy?: string;
}

interface InventoryLedgerProps {
  movements: InventoryMovement[];
  title?: string;
  currentStock?: number;
  isLoading?: boolean;
  error?: string | null;
}

export function InventoryLedger({
  movements,
  title = 'سجل حركات المخزون',
  currentStock,
  isLoading = false,
  error = null,
}: InventoryLedgerProps) {
  const getReferenceLabel = (referenceType?: string) => {
    const labels: Record<string, string> = {
      PURCHASE: 'شراء',
      SALE: 'بيع',
      RETURN: 'مرتجع',
      SUPPLIER_RETURN: 'مرتجع مورد',
      ADJUSTMENT: 'تعديل مخزون',
      TRANSFER: 'نقل مخزون',
    };
    return labels[String(referenceType || '').toUpperCase()] || referenceType || 'غير محدد';
  };

  const formatMovementDate = (value: string) => {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleString('ar-SA', {
      dateStyle: 'medium',
      timeStyle: 'short',
    });
  };

  const getMovementIcon = (type: InventoryMovement['type']) => {
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

  const getMovementBadge = (type: InventoryMovement['type']) => {
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
      case 'REVERSE_PURCHASE':
        return { label: 'إلغاء الشراء', variant: 'danger' as const };
      case 'REVERSE_SALE':
        return { label: 'إلغاء عملية البيع', variant: 'success' as const };
      case 'REVERSE_RETURN':
        return { label: 'إلغاء المرتجع', variant: 'secondary' as const };
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
          <div style={{ textAlign: 'left' }}>
            <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginBottom: '2px' }}>
              عدد الحركات
            </p>
            <p style={{ fontSize: '16px', fontWeight: '700', color: 'var(--text-primary)' }}>
              {movements.length}
            </p>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        {isLoading && (
          <div className="py-8 text-center text-sm text-text-muted" role="status">
            جار تحميل حركات المخزون...
          </div>
        )}
        {!isLoading && error && (
          <div className="py-8 text-center text-sm text-danger" role="alert">
            {error}
          </div>
        )}
        {!isLoading && !error && movements.length === 0 && (
          <div className="py-8 text-center text-sm text-text-muted">
            لا توجد حركات مخزون لهذا المنتج حتى الآن.
          </div>
        )}
        {!isLoading && !error && movements.length > 0 && (
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
            {movements.map((movement) => {
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
                          {formatMovementDate(movement.date)}
                        </span>
                      </div>
                      <p style={{ fontSize: '13px', color: 'var(--text-primary)', marginBottom: '2px' }}>
                        {movement.type === 'SALE' || movement.type === 'REVERSE_SALE'
                          ? `${badge.label} إلى ${movement.customerName || 'عميل نقدي'}`
                          : movement.reason || badge.label}
                      </p>
                      {movement.invoiceNumber && (
                        <p style={{ fontSize: '11px', color: 'var(--text-tertiary)' }}>
                          الفاتورة: {movement.invoiceNumber}
                        </p>
                      )}
                      {movement.referenceType && (
                        <p style={{ fontSize: '11px', color: 'var(--text-tertiary)' }}>
                          المرجع: {getReferenceLabel(movement.referenceType)}{movement.referenceId ? ` #${movement.referenceId}` : ''}
                        </p>
                      )}
                      {movement.beforeQuantity === 0 && movement.afterQuantity === 0 && movement.quantity !== 0 ? (
                        <div style={{ marginTop: '4px', fontSize: '11px', color: 'var(--text-tertiary)' }}>
                          الرصيد قبل وبعد الحركة غير مسجل لهذه الحركة القديمة
                        </div>
                      ) : (
                        <div style={{ display: 'flex', gap: '12px', marginTop: '4px', fontSize: '11px', color: 'var(--text-tertiary)' }}>
                          <span>قبل: {movement.beforeQuantity}</span>
                          <span>→</span>
                          <span>بعد: {movement.afterQuantity}</span>
                        </div>
                      )}
                    </div>
                    <div style={{ textAlign: 'right', marginRight: '12px' }}>
                      <p style={{
                        fontSize: '14px',
                        fontWeight: '600',
                           color: movement.quantity > 0 ? 'var(--success)' : movement.quantity < 0 ? 'var(--danger)' : 'var(--text-primary)'
                      }}>
                        {movement.quantity > 0 ? '+' : ''}{movement.quantity}
                      </p>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
        )}
      </CardContent>
    </Card>
  );
}
