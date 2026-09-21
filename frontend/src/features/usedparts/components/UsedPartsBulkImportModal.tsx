import { useMemo, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Modal } from '../../../design-system/components/modal';
import { Button } from '../../../design-system/components/button';
import { toast } from 'sonner';
import { inventoryApi, partTypesApi, productsApi } from '../../../services/api/endpoints';
import { Check, Download, Plus, Tag, Trash2, Upload, X } from 'lucide-react';

interface UsedPartsBulkImportModalProps {
  isOpen: boolean;
  onClose: () => void;
  onImported: () => void;
}

interface UsedPartRow {
  id: number;
  barcode: string;
  serialNumber: string;
}

const createRow = (id: number): UsedPartRow => ({ id, barcode: '', serialNumber: '' });

export function UsedPartsBulkImportModal({ isOpen, onClose, onImported }: UsedPartsBulkImportModalProps) {
  const [productId, setProductId] = useState('');
  const [partTypeId, setPartTypeId] = useState('');
  const [purchaseCost, setPurchaseCost] = useState('0');
  const [sellingPrice, setSellingPrice] = useState('0');
  const [rows, setRows] = useState<UsedPartRow[]>([createRow(1)]);
  const [saving, setSaving] = useState(false);
  const csvRef = useRef<HTMLInputElement>(null);

  const { data: productsData } = useQuery({
    queryKey: ['products', 'used-parts-bulk-import'],
    queryFn: () => productsApi.list({ page: 1, per_page: 1000 }),
    enabled: isOpen,
  });
  const { data: partTypesData } = useQuery({
    queryKey: ['part-types', 'used-parts-bulk-import'],
    queryFn: () => partTypesApi.list(),
    enabled: isOpen,
  });

  const products = useMemo(() => {
    const list = (productsData?.data?.products ?? []) as any[];
    return list.filter((product) => String(product.condition || '').toUpperCase() === 'USED' || String(product.sku || '').startsWith('USED-'));
  }, [productsData]);
  const partTypes = useMemo(() => Array.isArray(partTypesData?.data) ? partTypesData.data as any[] : [], [partTypesData]);

  const updateRow = (id: number, field: keyof UsedPartRow, value: string) => {
    setRows((current) => current.map((row) => row.id === id ? { ...row, [field]: value } : row));
  };

  const parseCsv = async (file: File) => {
    const lines = (await file.text()).split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
    const values = lines.slice(lines[0]?.toLowerCase().includes('barcode') ? 1 : 0)
      .map((line) => line.split(/[;,]/).map((value) => value.trim()))
      .filter((columns) => columns[0]);
    setRows(values.map((columns, index) => ({
      id: Date.now() + index,
      barcode: columns[0],
      serialNumber: columns[1] || '',
    })));
    toast.success(`تم تحميل ${values.length} قطعة للمراجعة`);
  };

  const downloadTemplate = () => {
    const blob = new Blob(['barcode,serial_number\nUSED-001,\nUSED-002,\n'], { type: 'text/csv;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'partflow-used-parts-template.csv';
    link.click();
    URL.revokeObjectURL(url);
  };

  const handleImport = async () => {
    if (!productId || !partTypeId) {
      toast.error('اختر المنتج ونوع القطعة أولًا');
      return;
    }
    const barcodes = rows.map((row) => row.barcode.trim()).filter(Boolean);
    if (!barcodes.length || new Set(barcodes.map((value) => value.toLowerCase())).size !== barcodes.length) {
      toast.error('أدخل باركودات غير مكررة');
      return;
    }
    setSaving(true);
    try {
      await inventoryApi.createBulkUsedStock({
        product_id: productId,
        barcodes,
        business_date: new Date().toISOString().slice(0, 10),
        part_type_id: partTypeId,
        purchase_cost: Number(purchaseCost) || 0,
        selling_price: Number(sellingPrice) || 0,
      });
      toast.success(`تمت إضافة ${barcodes.length} قطعة مستعملة بنجاح`);
      onImported();
      onClose();
      setRows([createRow(1)]);
    } catch (error: any) {
      toast.error(error?.arabicMessage || error?.message || 'تعذر استيراد القطع المستعملة');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="إضافة منتجات متعددة" size="xl" variant="modern">
      <div className="space-y-4" style={{ maxHeight: '72vh', overflowY: 'auto', padding: '2px 4px 4px' }}>
        <div className="rounded-2xl border border-primary/15 bg-gradient-to-l from-primary/10 via-white to-white px-5 py-4" style={{ boxShadow: '0 8px 24px rgba(15, 23, 42, 0.05)' }}>
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
            بيانات القطعة المستعملة المشتركة
          </label>
          <div className="grid gap-3 md:grid-cols-2">
            <label className="space-y-1 text-xs font-semibold text-text-secondary">
              <span>المنتج المستعمل</span>
              <select aria-label="المنتج" value={productId} onChange={(event) => setProductId(event.target.value)} className="pf-select-control w-full rounded-xl border border-border bg-surface px-3 py-3 text-sm font-normal">
                <option value="">اختر المنتج المستعمل...</option>
                {products.map((product) => <option key={product.id} value={product.id}>{product.name}</option>)}
              </select>
            </label>
            <label className="space-y-1 text-xs font-semibold text-text-secondary">
              <span>نوع القطعة</span>
              <select aria-label="نوع القطعة" value={partTypeId} onChange={(event) => setPartTypeId(event.target.value)} className="pf-select-control w-full rounded-xl border border-border bg-surface px-3 py-3 text-sm font-normal">
                <option value="">اختر نوع القطعة...</option>
                {partTypes.map((type) => <option key={type.id} value={type.id}>{type.name_ar}</option>)}
              </select>
            </label>
            <label className="space-y-1 text-xs font-semibold text-text-secondary">
              <span>سعر الشراء للوحدة</span>
              <input aria-label="سعر الشراء" value={purchaseCost} onChange={(event) => setPurchaseCost(event.target.value)} type="number" min="0" placeholder="0" className="w-full rounded-lg border border-border bg-white px-3 py-2 text-sm font-normal" style={{ minHeight: 42, borderRadius: 10 }} />
            </label>
            <label className="space-y-1 text-xs font-semibold text-text-secondary">
              <span>سعر البيع للوحدة</span>
              <input aria-label="سعر البيع" value={sellingPrice} onChange={(event) => setSellingPrice(event.target.value)} type="number" min="0" placeholder="0" className="w-full rounded-lg border border-border bg-white px-3 py-2 text-sm font-normal" style={{ minHeight: 42, borderRadius: 10 }} />
            </label>
          </div>
        </div>
        <div className="space-y-3">
          {rows.map((row, index) => (
            <div key={row.id} className="rounded-2xl border border-border bg-white p-4 shadow-sm transition-shadow hover:shadow-md" style={{ display: 'grid', gridTemplateColumns: 'minmax(150px, 1fr) minmax(150px, 1fr) auto', gap: '12px', alignItems: 'center', borderInlineStart: '4px solid rgba(37, 99, 235, 0.25)' }}>
              <input aria-label={`باركود القطعة ${index + 1}`} value={row.barcode} onChange={(event) => updateRow(row.id, 'barcode', event.target.value)} placeholder={`باركود القطعة ${index + 1}`} className="w-full rounded-lg border border-border bg-white px-3 py-2 text-sm" style={{ minHeight: 42, borderRadius: 10 }} />
              <input aria-label={`رقم تسلسلي ${index + 1}`} value={row.serialNumber} onChange={(event) => updateRow(row.id, 'serialNumber', event.target.value)} placeholder="رقم تسلسلي اختياري" className="w-full rounded-lg border border-border bg-white px-3 py-2 text-sm" style={{ minHeight: 42, borderRadius: 10 }} />
              <button type="button" onClick={() => setRows((current) => current.length === 1 ? [createRow(1)] : current.filter((item) => item.id !== row.id))} aria-label={`حذف القطعة ${index + 1}`} title="حذف القطعة" className="inline-flex min-h-10 items-center justify-center gap-1 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-xs font-medium text-red-600 hover:bg-red-100" style={{ minWidth: 42, minHeight: 42, borderRadius: 10 }}><Trash2 size={15} /><span className="hidden lg:inline">حذف</span></button>
            </div>
          ))}
        </div>
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-dashed border-primary/30 bg-primary/5 p-3">
          <input ref={csvRef} type="file" accept=".csv,text/csv" className="hidden" onChange={(event) => { const file = event.target.files?.[0]; if (file) void parseCsv(file); event.target.value = ''; }} />
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="secondary" onClick={() => csvRef.current?.click()}><Upload size={16} /> استيراد CSV</Button>
            <Button type="button" variant="secondary" onClick={downloadTemplate}><Download size={16} /> تحميل نموذج CSV</Button>
          </div>
          <button type="button" onClick={() => setRows((current) => [...current, createRow(Date.now() + current.length)])} className="inline-flex min-h-10 items-center gap-2 rounded-xl border border-primary/25 bg-primary/8 px-4 py-2 text-sm font-semibold text-primary" style={{ minHeight: 42, borderRadius: 12, border: '1px solid #bfdbfe', background: '#eff6ff', color: '#2563eb' }}><Plus size={17} /> إضافة سطر</button>
        </div>
        <div className="flex justify-end gap-2 border-t border-border pt-3">
          <Button variant="secondary" onClick={onClose}><X size={16} /> إلغاء</Button>
          <Button variant="primary" onClick={() => { void handleImport(); }} disabled={saving}>{!saving && <Check size={16} />}{saving ? 'جاري الإضافة...' : 'إضافة المنتجات'}</Button>
        </div>
      </div>
    </Modal>
  );
}
