import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './card';
import { Badge } from './badge';
import { Modal } from './modal';

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
  const [selectedMovement, setSelectedMovement] = useState<InventoryMovement | null>(null);
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

  const selectedBadge = selectedMovement ? getMovementBadge(selectedMovement.type) : null;
  const selectedDescription = selectedMovement && selectedBadge
    ? selectedMovement.type === 'SALE' || selectedMovement.type === 'REVERSE_SALE'
      ? `${selectedBadge.label} إلى ${selectedMovement.customerName || 'عميل نقدي'}`
      : selectedMovement.reason || selectedBadge.label
    : '';

  return (
    <>
      <Card className="inventory-ledger-card">
        <CardHeader className="inventory-ledger-header">
          <div className="inventory-ledger-title-block">
            <CardTitle>{title}</CardTitle>
            <span className="inventory-ledger-product">{title.includes(':') ? title.split(':').slice(1).join(':').trim() : ''}</span>
          </div>
          <div className="inventory-ledger-summary">
            {currentStock !== undefined && <div><span>المخزون الحالي</span><strong>{currentStock}</strong></div>}
            <div><span>الحركات</span><strong>{movements.length}</strong></div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading && <div className="py-8 text-center text-sm text-text-muted" role="status">جار تحميل حركات المخزون...</div>}
          {!isLoading && error && <div className="py-8 text-center text-sm text-danger" role="alert">{error}</div>}
          {!isLoading && !error && movements.length === 0 && <div className="py-8 text-center text-sm text-text-muted">لا توجد حركات مخزون لهذا المنتج حتى الآن.</div>}
          {!isLoading && !error && movements.length > 0 && (
            <div className="inventory-ledger-list">
              {movements.map((movement) => {
                const badge = getMovementBadge(movement.type);
                const description = movement.type === 'SALE' || movement.type === 'REVERSE_SALE'
                  ? `${badge.label} إلى ${movement.customerName || 'عميل نقدي'}`
                  : movement.reason || badge.label;
                const hasBalance = !(movement.beforeQuantity === 0 && movement.afterQuantity === 0 && movement.quantity !== 0);
                return (
                  <button key={movement.id} type="button" className="inventory-ledger-row" onClick={() => setSelectedMovement(movement)}>
                    <span className="inventory-ledger-row-marker" aria-hidden="true">{getMovementIcon(movement.type)}</span>
                    <span className="inventory-ledger-row-main">
                      <span className="inventory-ledger-row-heading"><Badge variant={badge.variant} size="sm">{badge.label}</Badge><strong>{description}</strong></span>
                      <span className="inventory-ledger-row-date">{formatMovementDate(movement.date)}{movement.invoiceNumber ? ` · الفاتورة: ${movement.invoiceNumber}` : ''}</span>
                    </span>
                    <span className="inventory-ledger-impact">
                      <strong className={movement.quantity > 0 ? 'positive' : movement.quantity < 0 ? 'negative' : ''}>{movement.quantity > 0 ? '+' : ''}{movement.quantity}</strong>
                      <span>{hasBalance ? `${movement.beforeQuantity} → ${movement.afterQuantity}` : 'الرصيد غير مسجل'}</span>
                    </span>
                  </button>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      <Modal isOpen={selectedMovement !== null} onClose={() => setSelectedMovement(null)} title="تفاصيل الحركة" size="sm">
        {selectedMovement && selectedBadge && (
          <div className="inventory-ledger-details">
            <div className="inventory-ledger-details-heading"><span className="inventory-ledger-row-marker">{getMovementIcon(selectedMovement.type)}</span><div><Badge variant={selectedBadge.variant} size="sm">{selectedBadge.label}</Badge><strong>{selectedDescription}</strong></div></div>
            <dl>
              <div><dt>التاريخ والوقت</dt><dd>{formatMovementDate(selectedMovement.date)}</dd></div>
              <div><dt>الكمية قبل</dt><dd>{selectedMovement.beforeQuantity}</dd></div>
              <div><dt>التغيير</dt><dd>{selectedMovement.quantity > 0 ? '+' : ''}{selectedMovement.quantity}</dd></div>
              <div><dt>الكمية بعد</dt><dd>{selectedMovement.afterQuantity}</dd></div>
              {selectedMovement.invoiceNumber && <div><dt>الفاتورة</dt><dd>{selectedMovement.invoiceNumber}</dd></div>}
              {selectedMovement.referenceType && <div><dt>المرجع</dt><dd>{getReferenceLabel(selectedMovement.referenceType)}{selectedMovement.referenceId ? ` #${selectedMovement.referenceId}` : ''}</dd></div>}
              {selectedMovement.customerName && <div><dt>العميل</dt><dd>{selectedMovement.customerName}</dd></div>}
              {selectedMovement.createdBy && <div><dt>بواسطة</dt><dd>{selectedMovement.createdBy}</dd></div>}
            </dl>
          </div>
        )}
      </Modal>
    </>
  );
}
