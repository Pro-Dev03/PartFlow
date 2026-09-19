import { ArrowLeft, PackagePlus, Plus, ShoppingCart, Tag, Truck, Warehouse } from 'lucide-react';
import { Modal } from '../../../design-system/components/modal';
import { Button } from '../../../design-system/components/button';

interface InventoryEntryModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAddProduct: () => void;
  onAddCategory: () => void;
  onAddSupplier: () => void;
  onAddExistingStock: () => void;
  onCreatePurchase: () => void;
}

export function InventoryEntryModal({
  isOpen,
  onClose,
  onAddProduct,
  onAddCategory,
  onAddSupplier,
  onAddExistingStock,
  onCreatePurchase,
}: InventoryEntryModalProps) {
  const choose = (action: () => void) => {
    onClose();
    action();
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="إضافة إلى منظومة المخزون" size="lg">
      <div className="flex flex-col gap-3">
        <div className="rounded-xl border border-[var(--color-primary-15)] bg-[var(--color-primary-08)] px-5 py-4">
          <p className="text-base font-semibold text-text-primary">كيف تريد إضافة المخزون؟</p>
          <p className="mt-1.5 text-sm leading-6 text-text-secondary">
            ابدأ من المخزون، وسنحافظ على السياق أثناء انتقالك بين التصنيف والمنتج والمورد والشراء.
          </p>
        </div>

        <button
          type="button"
          onClick={() => choose(onAddCategory)}
          className="pf-entry-option group order-2"
        >
          <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
            <Tag className="h-5 w-5" />
          </span>
          <span className="min-w-0 flex-1">
            <span className="mb-1.5 flex flex-wrap items-center gap-2 text-base">
              <span className="rounded-md bg-[var(--color-primary-10)] px-1.5 py-0.5 text-[10px] font-bold text-[var(--primary)]">02</span>
              <span className="font-semibold text-text-primary">إضافة تصنيف</span>
            </span>
            <span className="block text-sm leading-6 text-text-secondary">جهّز التصنيف قبل تعريف المنتج.</span>
          </span>
          <ArrowLeft className="h-4 w-4 shrink-0 text-text-secondary" />
        </button>

        <button
          type="button"
          onClick={() => choose(onCreatePurchase)}
            className="pf-entry-option group order-4 border-primary/30 bg-primary/5 shadow-sm"
        >
          <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
            <ShoppingCart className="h-5 w-5" />
          </span>
          <span className="min-w-0 flex-1">
            <span className="mb-1.5 flex flex-wrap items-center gap-2 text-base">
              <span className="rounded-md bg-[var(--color-primary-10)] px-1.5 py-0.5 text-[10px] font-bold text-[var(--primary)]">04</span>
              <span className="font-semibold text-text-primary">شراء جديد</span>
            </span>
            <span className="block text-sm leading-6 text-text-secondary">إنشاء فاتورة، تحديث رصيد المورد، واستلام الكمية.</span>
          </span>
          <ArrowLeft className="h-4 w-4 shrink-0 text-text-secondary transition-transform group-hover:-translate-x-1 group-hover:text-[var(--primary)]" />
        </button>

        <button
          type="button"
          onClick={() => choose(onAddSupplier)}
          className="pf-entry-option group order-3"
        >
          <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
            <Truck className="h-5 w-5" />
          </span>
          <span className="min-w-0 flex-1">
            <span className="mb-1.5 flex flex-wrap items-center gap-2 text-base">
                <span className="rounded-md bg-primary/15 px-1.5 py-0.5 text-[10px] font-bold text-primary">03</span>
              <span className="font-semibold text-text-primary">إضافة مورد</span>
            </span>
            <span className="block text-sm leading-6 text-text-secondary">أضف موردًا ليظهر مباشرة في عمليات الشراء.</span>
          </span>
          <ArrowLeft className="h-4 w-4 shrink-0 text-text-secondary" />
        </button>

        <button
          type="button"
          onClick={() => choose(onAddExistingStock)}
          className="pf-entry-option group order-5"
        >
          <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
            <Warehouse className="h-5 w-5" />
          </span>
          <span className="min-w-0 flex-1">
            <span className="mb-1.5 flex flex-wrap items-center gap-2 text-base">
              <span className="rounded-md bg-[var(--color-primary-10)] px-1.5 py-0.5 text-[10px] font-bold text-[var(--primary)]">05</span>
              <span className="font-semibold text-text-primary">تسجيل بضاعة موجودة</span>
            </span>
            <span className="block text-sm leading-6 text-text-secondary">إضافة كمية موجودة فعليًا دون إنشاء فاتورة شراء.</span>
          </span>
          <ArrowLeft className="h-4 w-4 shrink-0 text-text-secondary transition-transform group-hover:-translate-x-1 group-hover:text-[var(--primary)]" />
        </button>

        <button
          type="button"
          onClick={() => choose(onAddProduct)}
          className="pf-entry-option group order-1"
        >
          <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
            <PackagePlus className="h-5 w-5" />
          </span>
          <span className="min-w-0 flex-1">
            <span className="mb-1.5 flex flex-wrap items-center gap-2 text-base">
                <span className="rounded-md bg-[var(--color-primary-10)] px-1.5 py-0.5 text-[10px] font-bold text-[var(--primary)]">01</span>
              <span className="font-semibold text-text-primary">إنشاء صنف جديد</span>
            </span>
            <span className="block text-sm leading-6 text-text-secondary">تعريف المنتج أولًا، ثم إضافة رصيده عند الحاجة.</span>
          </span>
          <ArrowLeft className="h-4 w-4 shrink-0 text-text-secondary transition-transform group-hover:-translate-x-1 group-hover:text-[var(--primary)]" />
        </button>

        <div className="flex justify-end pt-2">
          <Button variant="secondary" onClick={onClose}>إلغاء</Button>
        </div>
      </div>
    </Modal>
  );
}
