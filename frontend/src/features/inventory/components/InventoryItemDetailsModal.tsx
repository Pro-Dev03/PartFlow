import { Modal } from '../../../design-system/components/modal';
import type { InventoryItem } from '../types/inventory.types';
import { formatPrice, normalizeCurrencyValue } from '../../../utils';
import { formatStoreDate } from '../../../utils/store-time';

interface InventoryItemDetailsModalProps {
  item: InventoryItem | null;
  isOpen: boolean;
  onClose: () => void;
}

export function InventoryItemDetailsModal({ item, isOpen, onClose }: InventoryItemDetailsModalProps) {
  if (!item) return null;
  const fields = [
    { label: 'الباركود', value: <bdi dir="ltr">{item.barcode || '—'}</bdi> },
    { label: 'الرقم التسلسلي', value: <bdi dir="ltr">{item.serial_number || '—'}</bdi> },
    { label: 'الحالة', value: item.status || '—' },
    { label: 'التاجر', value: item.supplier_name || '—' },
    { label: 'تكلفة الشراء', value: formatPrice(normalizeCurrencyValue(item.purchase_cost ?? 0)) },
    { label: 'سعر البيع', value: formatPrice(normalizeCurrencyValue(item.selling_price ?? item.price ?? 0)) },
    { label: 'تاريخ الشراء', value: item.purchase_date ? formatStoreDate(item.purchase_date, 'ar-SA') : '—' },
  ];

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="تفاصيل عنصر المخزون" size="md" autoFocus={false}>
      <div className="space-y-4">
        <h4 className="text-base font-semibold text-text-primary">{item.product_name || item.product?.name || 'عنصر المخزون'}</h4>
        <dl className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          {fields.map(({ label, value }) => (
            <div key={label} className="min-w-0 rounded-xl border border-border bg-surface-muted/40 p-3">
              <dt className="text-xs text-text-secondary">{label}</dt>
              <dd className="mt-1 break-words text-sm font-medium text-text-primary">{value}</dd>
            </div>
          ))}
        </dl>
      </div>
    </Modal>
  );
}
