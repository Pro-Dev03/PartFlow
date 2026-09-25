import { useState } from 'react';
import { Modal } from '../../../design-system/components/modal';
import { Input } from '../../../design-system/components/input';
import { Button } from '../../../design-system/components/button';
import { categoriesApi, suppliersApi } from '../../../services/api/endpoints';
import type { SupplierFormData } from '../../../components/forms/SupplierForm';
import { toast } from 'sonner';

interface InventoryQuickCreateModalProps {
  mode: 'category' | 'supplier';
  isOpen: boolean;
  onClose: () => void;
  onCreated: (record: { id: string; name: string }) => void;
}

export function InventoryQuickCreateModal({ mode, isOpen, onClose, onCreated }: InventoryQuickCreateModalProps) {
  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');
  const [description, setDescription] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  const reset = () => {
    setName('');
    setPhone('');
    setDescription('');
  };

  const close = () => {
    reset();
    onClose();
  };

  const handleSubmit = async () => {
    if (!name.trim()) {
      toast.error(mode === 'category' ? 'يرجى إدخال اسم التصنيف' : 'يرجى إدخال اسم التاجر');
      return;
    }
    if (mode === 'supplier' && !phone.trim()) {
      toast.error('يرجى إدخال رقم هاتف التاجر');
      return;
    }

    setIsSaving(true);
    try {
      const response = mode === 'category'
        ? await categoriesApi.create({ name: name.trim(), description: description.trim(), icon: 'smartphone', color: '#3B82F6', is_active: true })
        : await suppliersApi.create({
            code: `SUP-${Math.random().toString(36).slice(2, 10).toUpperCase()}`,
            name: name.trim(),
            phone: phone.trim(),
            notes: description.trim(),
            is_active: true,
          } satisfies SupplierFormData);
      const record = response?.data?.category ?? response?.data?.supplier ?? response?.data;
      if (!record?.id) throw new Error('missing created record');
      toast.success(mode === 'category' ? 'تمت إضافة التصنيف' : 'تمت إضافة التاجر');
      onCreated({ id: record.id, name: record.name });
      reset();
    } catch (error: any) {
      toast.error(error?.arabicMessage || error?.message || 'تعذر الحفظ');
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={close}
      title={mode === 'category' ? 'إضافة تصنيف' : 'إضافة تاجر'}
      variant="modern"
      size="sm"
      enableEnterNavigation={false}
    >
      <div
        className="space-y-4"
        data-next-disabled={mode === 'category' ? true : undefined}
        onKeyDown={(event) => {
          if (event.key !== 'Enter' || event.shiftKey || event.isComposing) return;
          event.preventDefault();
          void handleSubmit();
        }}
      >
        <div className="rounded-xl border border-primary/15 bg-primary/5 px-4 py-3 text-sm text-text-secondary">
          {mode === 'category'
            ? 'أنشئ التصنيف الآن وسيتم اختياره تلقائيًا في نموذج المنتج.'
            : 'أنشئ التاجر الآن وسيتم إعادته تلقائيًا إلى عملية الشراء.'}
        </div>
        <Input autoFocus label={mode === 'category' ? 'اسم التصنيف' : 'اسم التاجر'} value={name} onChange={(event) => setName(event.target.value)} />
        {mode === 'supplier' && (
          <Input label="رقم الهاتف" type="tel" value={phone} onChange={(event) => setPhone(event.target.value)} />
        )}
        <Input label="ملاحظات" value={description} onChange={(event) => setDescription(event.target.value)} />
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={close}>إلغاء</Button>
          <Button variant="primary" onClick={() => { void handleSubmit(); }} disabled={isSaving}>
            {isSaving ? 'جاري الحفظ...' : 'حفظ والعودة'}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
