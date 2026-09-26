import { useState } from 'react';
import { ArrowLeft, PackagePlus, Plus, ShoppingCart, Tag, Truck, Warehouse } from 'lucide-react';
import { Modal } from '../../../design-system/components/modal';
import { Button } from '../../../design-system/components/button';
import { InventoryQuickCreateModal } from './InventoryQuickCreateModal';

interface InventoryEntryModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAddProduct: () => void;
  onCreatePurchase: () => void;
  onBulkImport?: () => void;
  onAddCurrentStock: () => void;
  onAddUsedStock: () => void;
}

export function InventoryEntryModal({
  isOpen,
  onClose,
  onAddProduct,
  onCreatePurchase,
  onBulkImport,
  onAddCurrentStock,
  onAddUsedStock,
}: InventoryEntryModalProps) {
  const [screen, setScreen] = useState<'main' | 'category' | 'supplier'>('main');

  const handleClose = () => {
    setScreen('main');
    onClose();
  };

  const choose = (action: () => void) => {
    handleClose();
    action();
  };

  const quickActionClass =
    'group flex min-h-14 items-center gap-2.5 rounded-xl border border-border bg-surface px-3 py-2.5 text-start text-sm font-semibold text-text-primary transition-colors hover:border-primary/50 hover:bg-primary/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary';

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title={screen === 'category' ? 'تصنيف جديد' : screen === 'supplier' ? 'تاجر جديد' : 'إضافة إلى المخزون'}
      size="lg"
      autoFocus={false}
    >
      <div dir="rtl" className="space-y-4">
        {screen !== 'main' ? (
          <InventoryQuickCreateModal
            mode={screen}
            isOpen
            inline
            onClose={() => setScreen('main')}
            onCreated={() => setScreen('main')}
          />
        ) : (
          <>
        <div className="rounded-xl border border-[var(--color-primary-15)] bg-[var(--color-primary-08)] px-4 py-3">
          <p className="text-sm font-semibold text-text-primary">اختر طريقة الإدخال</p>
          <p className="mt-1 text-xs leading-5 text-text-secondary">
            أضف صنفًا مباشرة، أو سجّل فاتورة شراء لتحديث الكميات وحسابات التاجر.
          </p>
        </div>

        <div className="grid gap-3 sm:grid-cols-2">
          <section className="rounded-2xl border border-[var(--color-primary-20)] bg-[var(--color-primary-05)] p-4">
            <div className="mb-3 flex items-center gap-3">
              <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
                <PackagePlus className="h-5 w-5" />
              </span>
              <div className="min-w-0">
                <h3 className="text-sm font-bold text-text-primary">إضافة أصناف</h3>
                <p className="mt-0.5 text-xs text-text-secondary">صنف واحد أو قائمة أصناف</p>
              </div>
            </div>
            <div className={onBulkImport ? 'grid grid-cols-2 gap-2' : ''}>
              <button
                type="button"
                onClick={() => choose(onAddProduct)}
                className="flex min-h-11 w-full items-center justify-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm font-semibold text-white transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2"
              >
                <Plus className="h-4 w-4" />
                صنف واحد
              </button>
              {onBulkImport && (
                <button
                  type="button"
                  onClick={() => choose(onBulkImport)}
                  className="flex min-h-11 w-full items-center justify-center gap-2 rounded-lg border border-border bg-surface px-2 py-2 text-xs font-semibold text-text-primary transition-colors hover:border-primary/50 hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
                >
                  <PackagePlus className="h-4 w-4 shrink-0" />
                  عدة أصناف
                </button>
              )}
            </div>
          </section>

          <button
            type="button"
            onClick={() => choose(onCreatePurchase)}
            className="group flex min-h-32 w-full items-center gap-3 rounded-2xl border border-border bg-surface p-4 text-start transition-colors hover:border-primary/50 hover:bg-primary/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
          >
            <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
              <ShoppingCart className="h-5 w-5" />
            </span>
            <span className="min-w-0 flex-1">
              <span className="block text-sm font-bold text-text-primary">تسجيل شراء</span>
              <span className="mt-1 block text-xs leading-5 text-text-secondary">
                استلام الكمية وتحديث رصيد التاجر
              </span>
            </span>
            <ArrowLeft className="h-4 w-4 shrink-0 text-text-secondary transition-transform group-hover:-translate-x-0.5 group-hover:text-primary" />
          </button>

          <section className="rounded-2xl border border-border bg-surface p-4 sm:col-span-2">
            <div className="mb-3 flex items-center gap-3">
              <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
                <Warehouse className="h-5 w-5" />
              </span>
              <div className="min-w-0">
                <h3 className="text-sm font-bold text-text-primary">إضافة مخزون موجود دون فاتورة</h3>
                <p className="mt-0.5 text-xs text-text-secondary">سجّل رصيدًا حاليًا أو قطعًا مستعملة مع الباركود والتكلفة</p>
              </div>
            </div>
            <div className="grid gap-2 sm:grid-cols-2">
              <Button type="button" variant="secondary" onClick={() => choose(onAddCurrentStock)}>
                مخزون منتج موجود
              </Button>
              <Button type="button" variant="secondary" onClick={() => choose(onAddUsedStock)}>
                مخزون مستعمل
              </Button>
            </div>
          </section>

          <div className="sm:col-span-2">
            <p className="mb-2 text-xs font-semibold text-text-secondary">إعدادات سريعة</p>
            <div className="grid grid-cols-2 gap-2">
              <button type="button" onClick={() => setScreen('category')} className={quickActionClass}>
                <Tag className="h-4 w-4 shrink-0 text-primary" />
                <span className="min-w-0 flex-1">تصنيف جديد</span>
                <Plus className="h-4 w-4 shrink-0 text-text-secondary" />
              </button>
              <button type="button" onClick={() => setScreen('supplier')} className={quickActionClass}>
                <Truck className="h-4 w-4 shrink-0 text-primary" />
                <span className="min-w-0 flex-1">تاجر جديد</span>
                <Plus className="h-4 w-4 shrink-0 text-text-secondary" />
              </button>
            </div>
          </div>
        </div>

        <div className="flex justify-start border-t border-border pt-3">
          <Button variant="secondary" onClick={handleClose}>إلغاء</Button>
        </div>
          </>
        )}
      </div>
    </Modal>
  );
}
