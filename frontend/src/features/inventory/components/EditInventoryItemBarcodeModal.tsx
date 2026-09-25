import { useEffect, useState } from 'react';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { Modal } from '../../../design-system/components/modal';
import type { InventoryItem } from '../types/inventory.types';

interface EditInventoryItemBarcodeModalProps {
  item: InventoryItem | null;
  isOpen: boolean;
  onClose: () => void;
  onSave: (barcode: string) => Promise<boolean>;
}

export function EditInventoryItemBarcodeModal({ item, isOpen, onClose, onSave }: EditInventoryItemBarcodeModalProps) {
  const [barcode, setBarcode] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    if (isOpen) setBarcode(String(item?.barcode || ''));
  }, [isOpen, item?.id, item?.barcode]);

  const handleSave = async () => {
    if (!item || isSaving) return;
    setIsSaving(true);
    try {
      await onSave(barcode.trim());
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="تعديل باركود العنصر" size="sm" autoFocus={false}>
      <div className="space-y-4">
        <p className="text-sm text-text-secondary">
          {item?.product_name || item?.product?.name || 'عنصر المخزون'}
        </p>
        <Input
          label="الباركود"
          value={barcode}
          onChange={(event) => setBarcode(event.target.value)}
          placeholder="امسح أو أدخل الباركود"
          maxLength={100}
          dir="ltr"
          autoComplete="off"
          aria-label="باركود عنصر المخزون"
        />
        <p className="text-xs text-text-tertiary">يمكن ترك الحقل فارغًا لإزالة الباركود من هذا العنصر.</p>
        <div className="flex justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onClose} disabled={isSaving}>إلغاء</Button>
          <Button type="button" data-next-action onClick={() => void handleSave()} isLoading={isSaving}>حفظ</Button>
        </div>
      </div>
    </Modal>
  );
}
