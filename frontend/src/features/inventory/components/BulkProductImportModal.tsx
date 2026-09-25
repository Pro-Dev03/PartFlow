import { useMemo, useRef, useState } from 'react';
import { Modal } from '../../../design-system/components/modal';
import { Button } from '../../../design-system/components/button';
import { toast } from 'sonner';
import { categoriesApi, productsApi, suppliersApi } from '../../../services/api/endpoints';
import { useQuery } from '@tanstack/react-query';
import { clearBarcodeLookupFields, lookupProductByBarcode } from '../../../lib/productBarcodeLookup';
import { createProductCsvTemplate, parseProductCsv } from '../utils/productCsvImport';
import { inventoryApi } from '../../../services/api/endpoints';
import { Check, Download, Minus, Plus, Search, Tag, Trash2, Upload, X } from 'lucide-react';
import './bulk-product-import.css';

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
  quantity: string;
  minStockLevel: string;
  supplierId: string;
}

const createEmptyRow = (id: number): BulkProductRow => ({
  id,
  name: '',
  barcode: '',
  costPrice: '',
  sellingPrice: '',
  quantity: '0',
  minStockLevel: '0',
  supplierId: '',
});

export function BulkProductImportModal({ isOpen, onClose, onImported }: BulkProductImportModalProps) {
  const [rows, setRows] = useState<BulkProductRow[]>([createEmptyRow(1)]);
  const [categoryId, setCategoryId] = useState('');
  const [saving, setSaving] = useState(false);
  const [searchingByBarcode, setSearchingByBarcode] = useState<Record<number, boolean>>({});
  const csvInputRef = useRef<HTMLInputElement>(null);

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
    enabled: isOpen,
  });

  const categories = useMemo(() => (categoriesData?.data ?? []) as Array<{ id: string; name: string }>, [categoriesData]);
  const { data: suppliersData } = useQuery({
    queryKey: ['suppliers', 'bulk-import'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 100 }),
    enabled: isOpen,
  });
  const suppliers = useMemo(() => {
    const payload = suppliersData?.data;
    const list = Array.isArray(payload) ? payload : (payload?.suppliers ?? suppliersData?.suppliers ?? []);
    return (list as Array<{ id: string; name?: string; supplier_name?: string }>).map((supplier) => ({
      id: supplier.id,
      name: supplier.name || supplier.supplier_name || 'تاجر بدون اسم',
    }));
  }, [suppliersData]);

  const updateRow = (id: number, field: 'name' | 'barcode' | 'costPrice' | 'sellingPrice' | 'quantity' | 'minStockLevel' | 'supplierId', value: string) => {
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

  const handleCsvImport = async (file: File) => {
    try {
      const importedRows = parseProductCsv(await file.text());
      setRows(importedRows.map((row, index) => ({ ...row, id: Date.now() + index })));
      toast.success(`تم تحميل ${importedRows.length} منتجًا للمراجعة`);
    } catch (error: any) {
      toast.error(error?.message || 'تعذر قراءة ملف CSV');
    } finally {
      if (csvInputRef.current) csvInputRef.current.value = '';
    }
  };

  const downloadCsvTemplate = () => {
    const blob = new Blob([`\ufeff${createProductCsvTemplate()}`], { type: 'text/csv;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'partflow-products-template.csv';
    link.click();
    URL.revokeObjectURL(url);
  };

  const parseRows = (sourceRows: BulkProductRow[] = rows) => {
    return sourceRows
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
          quantity: Math.max(0, Math.floor(Number(row.quantity) || 0)),
          min_stock_level: Math.max(0, Math.floor(Number(row.minStockLevel) || 0)),
          preferred_supplier_id: row.supplierId || undefined,
        };
      });
  };

  const handleImport = async () => {
    const rowsForImport = rows.filter((row) => row.name.trim() || row.barcode || row.costPrice || row.sellingPrice);
    const items = parseRows(rowsForImport);
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
      const createdProducts: Array<{ id?: string }> = Array.isArray(payload?.created) ? payload.created : [];
      const created = Number(payload?.created_count ?? createdProducts.length) || 0;
      const failed: Array<{ index?: number }> = Array.isArray(payload?.failed) ? payload.failed : [];
      const failedIndexes = new Set(failed
        .map((failure) => Number(failure.index))
        .filter((index) => Number.isInteger(index) && index >= 0 && index < items.length));
      const successfulItems = items.filter((_, index) => !failedIndexes.has(index));
      const responseMappingIsValid = successfulItems.length === createdProducts.length;
      const stockUpdates = responseMappingIsValid
        ? createdProducts.map((product, index) => {
            const quantity = successfulItems[index]?.quantity ?? 0;
            if (!product.id || quantity <= 0) return Promise.resolve();
            return inventoryApi.adjustProductQuantity(product.id, quantity, 'bulk product import');
          })
        : [];
      const stockUpdateResults = await Promise.allSettled(stockUpdates);
      const failedStockUpdates = stockUpdateResults.filter((result) => result.status === 'rejected').length;
      const hasUnmappedRejectedRows = failed.length !== failedIndexes.size;
      const rowsToRetry = hasUnmappedRejectedRows
        ? rowsForImport
        : Array.from(failedIndexes, (index) => rowsForImport[index]).filter(Boolean);
      if (failed.length > 0) {
        toast.error(`تمت إضافة ${created} منتج، ورفضت ${failed.length} صفوف بسبب بيانات غير صالغة.`);
      } else {
        toast.success(`تمت إضافة ${created} منتج بنجاح`);
      }
      if (failed.length > 0) {
        setRows(rowsToRetry.length > 0 ? rowsToRetry : rowsForImport);
        if (created > 0) onImported?.();
        if (failedStockUpdates > 0 || !responseMappingIsValid) {
          toast.error('تم إنشاء المنتجات، لكن تعذر تحديث كمية المخزون لبعضها. راجع الكميات في صفحة المخزون.');
        }
        return;
      }
      if (failedStockUpdates > 0 || !responseMappingIsValid) {
        toast.error('تم إنشاء المنتجات، لكن تعذر تحديث كمية المخزون لبعضها. راجع الكميات في صفحة المخزون.');
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
    <Modal isOpen={isOpen} onClose={onClose} title="إضافة منتجات متعددة" size="xl" variant="modern">
      <div className="bulk-product-import-content space-y-4" style={{ padding: '2px 4px 4px' }}>
        <div className="bulk-product-import-intro rounded-2xl border border-primary/15 px-5 py-4">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-sm font-semibold text-text-primary">أضف مخزونك دفعة واحدة</p>
              <p className="mt-1 text-xs text-text-secondary">حمّل CSV أو أضف صفوفًا، ثم راجع البيانات قبل الحفظ.</p>
            </div>
            <span className="hidden rounded-full bg-primary/10 px-3 py-1 text-xs font-medium text-primary sm:inline-flex">Bulk Import</span>
          </div>
        </div>
        <div className="rounded-2xl border border-border bg-surface-muted/40 p-4">
          <label className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-text-secondary">
            <Tag size={15} />
            التصنيف المشترك
          </label>
          <div className="relative">
            <select
              value={categoryId}
              onChange={(event) => setCategoryId(event.target.value)}
              aria-label="التصنيف المشترك"
              className="pf-select-control w-full rounded-xl border border-border bg-surface px-3 py-3 text-sm"
            >
              <option value="">اختر التصنيف...</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>{category.name}</option>
              ))}
            </select>
          </div>
        </div>

        <div className="space-y-3">
          {rows.map((row, index) => (
            <div key={row.id} className="bulk-product-import-row rounded-2xl border border-border p-4 shadow-sm transition-shadow hover:shadow-md">
              <div className="contents">
                <span className="sr-only">المنتج {index + 1}</span>
              <input
                value={row.name}
                onChange={(event) => updateRow(row.id, 'name', event.target.value)}
                placeholder={`اسم المنتج ${index + 1}`}
                className="w-full rounded-lg border border-border bg-white px-3 py-2 text-sm"
                style={{ minHeight: 42, borderRadius: 10, border: '1px solid #e2e8f0', padding: '0 12px' }}
              />
              <div className="flex gap-2" style={{ minWidth: 0 }}>
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
                  style={{ minHeight: 42, minWidth: 0, borderRadius: 10, border: '1px solid #e2e8f0', padding: '0 12px' }}
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
                  style={{ minHeight: 42, minWidth: 48, borderRadius: 10, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 5 }}
                >
                  {searchingByBarcode[row.id] ? '...' : <><Search size={15} /> <span className="hidden sm:inline">بحث</span></>}
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
                style={{ minHeight: 42, borderRadius: 10, border: '1px solid #e2e8f0', padding: '0 12px' }}
              />
              <input
                value={row.sellingPrice}
                onChange={(event) => updateRow(row.id, 'sellingPrice', event.target.value)}
                placeholder="سعر البيع"
                type="number"
                min="0"
                step="0.01"
                className="w-full rounded-lg border border-border bg-white px-3 py-2 text-sm"
                style={{ minHeight: 42, borderRadius: 10, border: '1px solid #e2e8f0', padding: '0 12px' }}
              />
              <div className="flex items-center gap-1 rounded-lg border border-border bg-white p-1" aria-label="الكمية" style={{ minHeight: 42, borderRadius: 10, border: '1px solid #e2e8f0' }}>
                <button
                  type="button"
                  aria-label="إنقاص الكمية"
                  title="إنقاص الكمية"
                  className="grid h-8 w-8 place-items-center rounded-md text-text-secondary hover:bg-surface-muted"
                  style={{ width: 34, height: 34, borderRadius: 9, border: '1px solid #e2e8f0', background: '#f8fafc', color: '#64748b' }}
                  onClick={() => updateRow(row.id, 'quantity', String(Math.max(0, Number(row.quantity || 0) - 1)))}
                ><Minus size={15} /></button>
                <input
                  value={row.quantity ?? '0'}
                  onChange={(event) => updateRow(row.id, 'quantity', event.target.value.replace(/[^0-9]/g, ''))}
                  placeholder="الكمية"
                  type="number"
                  min="0"
                  step="1"
                  aria-label="الكمية"
                  className="min-w-0 w-full border-0 bg-transparent px-1 text-center text-sm outline-none"
                />
                <button
                  type="button"
                  aria-label="زيادة الكمية"
                  title="زيادة الكمية"
                  className="grid h-8 w-8 place-items-center rounded-md bg-primary/10 text-primary hover:bg-primary/20"
                  style={{ width: 34, height: 34, borderRadius: 9, border: '1px solid rgba(37, 99, 235, 0.2)', background: '#eff6ff', color: '#2563eb' }}
                  onClick={() => updateRow(row.id, 'quantity', String(Number(row.quantity || 0) + 1))}
                ><Plus size={15} /></button>
              </div>
              <div className="bulk-import-meta">
                <label className="flex items-center gap-2 text-xs text-text-secondary">
                  الحد الأدنى
                  <input
                    value={row.minStockLevel ?? '0'}
                    onChange={(event) => updateRow(row.id, 'minStockLevel', event.target.value.replace(/[^0-9]/g, ''))}
                    type="number"
                    min="0"
                    step="1"
                    aria-label="الحد الأدنى للمخزون"
                    className="w-full rounded-lg border border-border bg-white px-2 py-2 text-sm text-text-primary"
                  />
                </label>
                <label className="flex items-center gap-2 text-xs text-text-secondary">
                  التاجر <span className="text-[11px]">(اختياري)</span>
                  <select
                    value={row.supplierId ?? ''}
                    onChange={(event) => updateRow(row.id, 'supplierId', event.target.value)}
                    aria-label="التاجر الاختياري"
                    className="w-full rounded-lg border border-border bg-white px-2 py-2 text-sm text-text-primary"
                  >
                    <option value="">بدون تاجر</option>
                    {suppliers.map((supplier) => <option key={supplier.id} value={supplier.id}>{supplier.name}</option>)}
                  </select>
                </label>
              </div>
              <button
                type="button"
                onClick={() => removeRow(row.id)}
                aria-label={`حذف المنتج ${index + 1}`}
                title="حذف المنتج"
                className="inline-flex min-h-10 items-center justify-center gap-1 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-xs font-medium text-red-600 transition-colors hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-red-200"
                style={{ minWidth: 42, minHeight: 42, borderRadius: 10, border: '1px solid #fecaca', background: '#fff1f2', color: '#dc2626' }}
              >
                <Trash2 size={15} />
                <span className="hidden lg:inline">حذف</span>
              </button>
              </div>
            </div>
          ))}
        </div>

        <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-dashed border-primary/30 bg-primary/5 p-3">
          <input
            ref={csvInputRef}
            type="file"
            accept=".csv,text/csv"
            className="hidden"
            onChange={(event) => {
              const file = event.target.files?.[0];
              if (file) void handleCsvImport(file);
            }}
          />
          <Button type="button" variant="secondary" onClick={() => csvInputRef.current?.click()}>
            <Upload size={16} />
            استيراد CSV
          </Button>
          <Button type="button" variant="secondary" onClick={downloadCsvTemplate}>
            <Download size={16} />
            تحميل نموذج CSV
          </Button>
        </div>

        <div className="flex items-center justify-between gap-2 rounded-xl bg-surface-muted/50 px-3 py-2">
          <button
            type="button"
            onClick={addRow}
            className="inline-flex min-h-10 items-center gap-2 rounded-xl border border-primary/25 bg-primary/8 px-4 py-2 text-sm font-semibold text-primary transition-all hover:-translate-y-0.5 hover:bg-primary/15 focus:outline-none focus:ring-2 focus:ring-primary/20"
            style={{ minHeight: 42, borderRadius: 12, border: '1px solid #bfdbfe', background: '#eff6ff', color: '#2563eb', padding: '0 16px' }}
          >
            <Plus size={17} />
            إضافة سطر
          </button>
        </div>

        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose}>
            <X size={16} />
            إلغاء
          </Button>
          <Button variant="primary" onClick={() => { void handleImport(); }} disabled={saving || !categoryId}>
            {!saving && <Check size={16} />}
            {saving ? 'جاري الإضافة...' : 'إضافة المنتجات'}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
