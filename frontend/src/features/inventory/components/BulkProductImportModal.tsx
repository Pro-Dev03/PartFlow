import { useMemo, useState } from 'react';
import { Modal } from '../../../design-system/components/modal';
import { Button } from '../../../design-system/components/button';
import { toast } from 'sonner';
import { categoriesApi, productsApi } from '../../../services/api/endpoints';
import { useQuery } from '@tanstack/react-query';
import { clearBarcodeLookupFields, lookupProductByBarcode } from '../../../lib/productBarcodeLookup';

interface BulkProductImportModalProps {
  isOpen: boolean;
  onClose: () => void;
  onImported?: () => void;
}

interface BulkProductRow {
  id: number;
  name: string;
  barcode: string;
  costPrice: string;
  sellingPrice: string;
}

const createEmptyRow = (id: number): BulkProductRow => ({
  id,
  name: '',
  barcode: '',
  costPrice: '',
  sellingPrice: '',
});

export function BulkProductImportModal({ isOpen, onClose, onImported }: BulkProductImportModalProps) {
  const [rows, setRows] = useState<BulkProductRow[]>([createEmptyRow(1)]);
  const [categoryId, setCategoryId] = useState('');
  const [saving, setSaving] = useState(false);
  const [searchingByBarcode, setSearchingByBarcode] = useState<Record<number, boolean>>({});

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
    enabled: isOpen,
  });

  const categories = useMemo(() => (categoriesData?.data ?? []) as Array<{ id: string; name: string }>, [categoriesData]);

  const updateRow = (id: number, field: 'name' | 'barcode' | 'costPrice' | 'sellingPrice', value: string) => {
    setRows((currentRows) =>
      currentRows.map((row) => (row.id === id ? { ...row, [field]: value } : row)),
    );
  };

  const autoFillRowFromBarcode = async (rowId: number, barcode: string) => {
    const trimmed = barcode.trim();
    if (!trimmed) return;

    setRows((currentRows) => currentRows.map((row) => (
      row.id === rowId ? clearBarcodeLookupFields(row) : row
    )));
    setSearchingByBarcode((current) => ({ ...current, [rowId]: true }));
    try {
      const match = await lookupProductByBarcode(trimmed);
      if (!match) {
        toast.info('لم يتم العثور على بيانات لهذا الباركود. يمكنك إكمال الإدخال يدويًا.');
        return;
      }

      setRows((currentRows) =>
        currentRows.map((row) => {
          if (row.id !== rowId || row.barcode.trim() !== trimmed) return row;
          return {
            ...row,
            name: row.name || match.name || '',
            barcode: match.barcode || row.barcode,
          };
        }),
      );

      if (match.category && !categoryId) {
        const matchedCategory = categories.find((category) => category.name.toLowerCase() === match.category?.toLowerCase());
        if (matchedCategory) {
          setCategoryId(matchedCategory.id);
        }
      }

      toast.success(`تمت تعبئة البيانات من مصادر مجانية موثوقة: ${match.name}`);
    } catch {
      toast.error('تعذر جلب بيانات المنتج، يمكنك متابعة الإدخال اليدوي.');
    } finally {
      setSearchingByBarcode((current) => ({ ...current, [rowId]: false }));
    }
  };

  const addRow = () => {
    setRows((currentRows) => [...currentRows, createEmptyRow(Date.now() + Math.random())]);
  };

  const removeRow = (id: number) => {
    setRows((currentRows) => {
      if (currentRows.length === 1) {
        return [createEmptyRow(Date.now())];
      }
      return currentRows.filter((row) => row.id !== id);
    });
  };

  const parseRows = () => {
    return rows
      .filter((row) => row.name.trim() || row.barcode || row.costPrice || row.sellingPrice)
      .map((row) => {
        const name = row.name.trim();
        const barcode = row.barcode.trim();
        const cost = Number(row.costPrice.replace(/,/g, '').trim());
        const selling = Number(row.sellingPrice.replace(/,/g, '').trim());

        return {
          name,
          sku: '',
          barcode: barcode || '',
          category_id: categoryId || undefined,
          cost_price: Number.isFinite(cost) ? cost : 0,
          selling_price: Number.isFinite(selling) ? selling : 0,
        };
      });
  };

  const handleImport = async () => {
    const items = parseRows();
    if (!categoryId) {
      toast.error('يرجى اختيار تصنيف أولاً');
      return;
    }
    if (!items.length) {
      toast.error('أضف منتج واحد على الأقل');
      return;
    }

    const invalid = items.some((item) => !item.name || item.cost_price <= 0 || item.selling_price <= 0);
    if (invalid) {
      toast.error('كل منتج يحتاج اسم، سعر شراء، وسعر بيع صحيح');
      return;
    }

    setSaving(true);
    try {
      const response = await productsApi.bulkCreate({ items });
      const payload = response?.data ?? response;
      const created = payload?.created_count ?? payload?.created?.length ?? 0;
      const failed = payload?.failed ?? [];
      if (failed.length > 0) {
        toast.error(`تمت إضافة ${created} منتج، ورفضت ${failed.length} صفوف بسبب بيانات غير صالغة.`);
      } else {
        toast.success(`تمت إضافة ${created} منتج بنجاح`);
      }
      onImported?.();
      onClose();
      setRows([createEmptyRow(1)]);
      setCategoryId('');
    } catch (error: any) {
      toast.error(error?.arabicMessage || error?.message || 'تعذر استيراد المنتجات');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="إضافة منتجات متعددة" size="lg">
      <div className="space-y-4">
        <div className="rounded-xl border border-primary/15 bg-primary/5 px-4 py-3 text-sm text-text-secondary">
          أسهل طريقة: املأ اسم المنتج، سعر الشراء، سعر البيع، والباركود اختياريًا في كل سطر.
        </div>
        <div>
          <label className="mb-2 block text-sm font-medium text-text-primary">التصنيف</label>
          <select
            value={categoryId}
            onChange={(event) => setCategoryId(event.target.value)}
            className="pf-select-control w-full rounded-xl border border-border bg-surface px-3 py-2"
          >
            <option value="">اختر التصنيف...</option>
            {categories.map((category) => (
              <option key={category.id} value={category.id}>{category.name}</option>
            ))}
          </select>
        </div>

        <div className="space-y-3">
          {rows.map((row, index) => (
            <div key={row.id} className="grid grid-cols-1 gap-2 rounded-xl border border-border bg-surface p-3 md:grid-cols-[1.4fr_1.1fr_1.1fr_1.1fr_auto]">
              <input
                value={row.name}
                onChange={(event) => updateRow(row.id, 'name', event.target.value)}
                placeholder={`اسم المنتج ${index + 1}`}
                className="w-full rounded-lg border border-border bg-white px-3 py-2 text-sm"
              />
              <div className="flex gap-2">
                <input
                  value={row.barcode}
                  onChange={(event) => updateRow(row.id, 'barcode', event.target.value)}
                  onBlur={() => {
                    if (row.barcode.trim()) {
                      void autoFillRowFromBarcode(row.id, row.barcode);
                    }
                  }}
                  placeholder="باركود (اختياري)"
                  className="w-full rounded-lg border border-border bg-white px-3 py-2 text-sm"
                />
                <Button
                  type="button"
                  variant="secondary"
                  size="sm"
                  onClick={() => {
                    if (row.barcode.trim()) {
                      void autoFillRowFromBarcode(row.id, row.barcode);
                    }
                  }}
                  disabled={searchingByBarcode[row.id] || !row.barcode.trim()}
                >
                  {searchingByBarcode[row.id] ? '...' : 'بحث'}
                </Button>
              </div>
              <input
                value={row.costPrice}
                onChange={(event) => updateRow(row.id, 'costPrice', event.target.value)}
                placeholder="سعر الشراء"
                type="number"
                min="0"
                step="0.01"
                className="w-full rounded-lg border border-border bg-white px-3 py-2 text-sm"
              />
              <input
                value={row.sellingPrice}
                onChange={(event) => updateRow(row.id, 'sellingPrice', event.target.value)}
                placeholder="سعر البيع"
                type="number"
                min="0"
                step="0.01"
                className="w-full rounded-lg border border-border bg-white px-3 py-2 text-sm"
              />
              <button
                type="button"
                onClick={() => removeRow(row.id)}
                className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600"
              >
                حذف
              </button>
            </div>
          ))}
        </div>

        <div className="flex items-center justify-between gap-2">
          <button
            type="button"
            onClick={addRow}
            className="rounded-lg border border-primary/30 bg-primary/5 px-3 py-2 text-sm font-medium text-primary"
          >
            + إضافة سطر
          </button>
          <span className="text-xs text-text-secondary">SKU يضاف تلقائياً إذا كان فارغًا، والباركود اختياري</span>
        </div>

        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose}>إلغاء</Button>
          <Button variant="primary" onClick={() => { void handleImport(); }} disabled={saving || !categoryId}>
            {saving ? 'جاري الإضافة...' : 'إضافة المنتجات'}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
