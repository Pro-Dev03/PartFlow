import { useState, type FormEvent } from 'react';
import { useQueryClient } from '@tanstack/react-query';
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
  inline?: boolean;
}

export function InventoryQuickCreateModal({ mode, isOpen, onClose, onCreated, inline = false }: InventoryQuickCreateModalProps) {
  const queryClient = useQueryClient();
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

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const cleanName = name.trim();
    if (!cleanName) {
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
        ? await categoriesApi.create({ name: cleanName, description: description.trim(), icon: 'smartphone', color: '#3B82F6', is_active: true })
        : await suppliersApi.create({
            code: `SUP-${Math.random().toString(36).slice(2, 10).toUpperCase()}`,
            name: cleanName,
            phone: phone.trim(),
            notes: description.trim(),
            is_active: true,
          } satisfies SupplierFormData);
      const record = response?.data?.category ?? response?.data?.supplier ?? response?.data;
      if (!record?.id) throw new Error('missing created record');

      await queryClient.invalidateQueries({ queryKey: [mode === 'category' ? 'categories' : 'suppliers'] });
      toast.success(mode === 'category' ? 'تمت إضافة التصنيف' : 'تمت إضافة التاجر');
      onCreated({ id: record.id, name: record.name || cleanName });
      reset();
    } catch (error: any) {
      toast.error(error?.arabicMessage || error?.message || 'تعذر الحفظ');
    } finally {
      setIsSaving(false);
    }
  };

  const form = (
    <form onSubmit={(event) => { void handleSubmit(event); }} className="space-y-4">
      <div className="rounded-xl border border-primary/15 bg-primary/5 px-4 py-3 text-sm text-text-secondary">
        {mode === 'category'
          ? 'أضف التصنيف وسيظهر مباشرة ضمن خيارات المخزون.'
          : 'أضف التاجر وسيظهر مباشرة ضمن خيارات المخزون.'}
      </div>
      <Input autoFocus label={mode === 'category' ? 'اسم التصنيف' : 'اسم التاجر'} value={name} onChange={(event) => setName(event.target.value)} required />
      {mode === 'supplier' && (
        <Input label="رقم الهاتف" type="tel" value={phone} onChange={(event) => setPhone(event.target.value)} required />
      )}
      <Input label="ملاحظات" value={description} onChange={(event) => setDescription(event.target.value)} />
      <div className="flex justify-start gap-2">
        <Button type="button" variant="secondary" onClick={close}>رجوع</Button>
        <Button type="submit" variant="primary" disabled={isSaving}>
          {isSaving ? 'جارٍ الحفظ...' : 'حفظ'}
        </Button>
      </div>
    </form>
  );

  if (inline) {
    if (!isOpen) return null;
    return (
      <section className="rounded-2xl border border-border bg-surface p-4">
        <div className="mb-4 flex items-center justify-between gap-3">
          <h3 className="text-sm font-bold text-text-primary">{mode === 'category' ? 'إضافة تصنيف' : 'إضافة تاجر'}</h3>
          <Button type="button" variant="ghost" size="sm" onClick={close}>رجوع</Button>
        </div>
        {form}
      </section>
    );
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={close}
      title={mode === 'category' ? 'إضافة تصنيف' : 'إضافة تاجر'}
      variant="modern"
      size="sm"
      enableEnterNavigation={false}
    >
      {form}
    </Modal>
  );
}
