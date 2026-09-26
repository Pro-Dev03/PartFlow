import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { CalendarRange, Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '../../../design-system/components/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { settingsApi } from '../../../services/api/endpoints';
import { getStoreMonthBounds, getStoreToday, getStoreTimezone } from '../../../utils/store-time';

type CleanupType = 'sales' | 'purchases' | 'expenses' | 'income' | 'returns' | 'supplier_returns' | 'debts' | 'payments' | 'products' | 'inventory_items' | 'held_sales' | 'inspections' | 'inventory_adjustments' | 'customers' | 'suppliers';
type CleanupTarget = 'cloud' | 'local';

function unwrap(response: any) {
  return response?.data?.data ?? response?.data ?? response;
}

export function HistoricalCleanupSettings() {
  const queryClient = useQueryClient();
  const today = getStoreToday();
  const monthBounds = getStoreMonthBounds();
  const isDesktop = typeof window !== 'undefined' && Boolean(window.partflowDesktop);
  const [type, setType] = useState<CleanupType>('sales');
  const [startDate, setStartDate] = useState(monthBounds.start);
  const [endDate, setEndDate] = useState(today);
  const [target, setTarget] = useState<CleanupTarget>(isDesktop ? 'local' : 'cloud');
  const [preview, setPreview] = useState<any | null>(null);
  const [confirming, setConfirming] = useState(false);
  const [lastResult, setLastResult] = useState<any | null>(null);

  const previewMutation = useMutation({
    mutationFn: () => settingsApi.previewHistoricalCleanup(type, startDate, endDate, target),
    onSuccess: (response: any) => {
      setPreview(unwrap(response));
      setConfirming(false);
    },
    onError: (error: any) => toast.error(error?.message || 'تعذرت معاينة السجلات المحددة.'),
  });

  const runMutation = useMutation({
    mutationFn: () => settingsApi.runHistoricalCleanup({
      type,
      start_date: startDate,
      end_date: endDate,
      candidate_ids: preview?.candidate_ids || [],
    }, target),
    onSuccess: (response: any) => {
      const result = unwrap(response);
      setPreview(null);
      setConfirming(false);
      setLastResult(result);
      void queryClient.invalidateQueries();
      toast.success(`اكتمل الحذف: ${Number(result?.deleted || 0)} سجل. تعذّر حذف ${Number(result?.blocked || 0) + Number(result?.failed || 0)} سجل؛ راجع سبب كل سجل أدناه.`);
    },
    onError: (error: any) => toast.error(error?.message || 'تعذر تنظيف السجلات المحددة.'),
  });

  const clearPreview = () => {
    setPreview(null);
    setConfirming(false);
  };

  return (
    <Card className="border-amber-200">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          <Trash2 className="h-5 w-5 text-amber-700" />
          تنظيف السجلات التجارية القديمة
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4" dir="rtl">
        <p className="text-sm text-slate-600">
          اختر نوع السجلات والفترة لمعاينتها. يعكس PartFlow آثار السجل المالية والمخزنية والسجلات التابعة ثم يحذفه نهائيًا داخل المعاملة المناسبة. إذا تعذر عكس سجل قديم بسبب نقص أو تعارض في بياناته، لن يُحذف وستظهر هويته وسبب التعذر أدناه. حذف الصفوف لا يضمن انخفاض حجم قاعدة PostgreSQL فورًا؛ تصبح المساحة قابلة لإعادة الاستخدام بعد الصيانة التلقائية.
        </p>
        {isDesktop && (
          <label className="flex max-w-md flex-col gap-2 text-sm font-medium text-slate-700">
            قاعدة البيانات
            <select
              className="min-h-10 rounded-md border border-slate-300 bg-white px-3"
              value={target}
              onChange={(event) => { setTarget(event.target.value as CleanupTarget); clearPreview(); }}
              disabled={previewMutation.isPending || runMutation.isPending}
            >
              <option value="cloud">متجر Render السحابي</option>
              <option value="local">قاعدة SQLite على هذا الجهاز</option>
            </select>
          </label>
        )}
        <div className="grid gap-3 sm:grid-cols-3">
          <label className="flex flex-col gap-2 text-sm font-medium text-slate-700">
            نوع السجلات
            <select
              className="min-h-10 rounded-md border border-slate-300 bg-white px-3"
              value={type}
              onChange={(event) => { setType(event.target.value as CleanupType); clearPreview(); setLastResult(null); }}
              disabled={previewMutation.isPending || runMutation.isPending}
            >
              <option value="sales">فواتير المبيعات</option>
              <option value="purchases">فواتير المشتريات</option>
              <option value="expenses">المصروفات</option>
              <option value="returns">مرتجعات العملاء</option>
              <option value="supplier_returns">مرتجعات الموردين</option>
              <option value="debts">الديون</option>
              <option value="payments">الدفعات</option>
              <option value="products">المنتجات ومخزونها وسجلاتها المرتبطة</option>
              <option value="inventory_items">عناصر المخزون وسجلات عملياتها المرتبطة</option>
              <option value="income">الإيرادات المستقلة</option>
              <option value="held_sales">المبيعات المعلّقة الخاصة بي</option>
              <option value="inspections">الفحوص وسجلات سير العمل المرتبطة</option>
              <option value="inventory_adjustments">تعديلات كمية المخزون</option>
              <option value="customers">العملاء وسجلاتهم المرتبطة</option>
              <option value="suppliers">الموردون وسجلاتهم المرتبطة</option>
            </select>
          </label>
          <label className="flex flex-col gap-2 text-sm font-medium text-slate-700">
            من تاريخ
            <input type="date" className="min-h-10 rounded-md border border-slate-300 bg-white px-3" value={startDate} max={endDate} onChange={(event) => { setStartDate(event.target.value); clearPreview(); }} />
          </label>
          <label className="flex flex-col gap-2 text-sm font-medium text-slate-700">
            إلى تاريخ
              <input type="date" className="min-h-10 rounded-md border border-slate-300 bg-white px-3" value={endDate} min={startDate} max={today} onChange={(event) => { setEndDate(event.target.value); clearPreview(); }} />
          </label>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button
            variant="secondary"
            onClick={() => previewMutation.mutate()}
            disabled={!startDate || !endDate || startDate > endDate || previewMutation.isPending || runMutation.isPending}
          >
            <CalendarRange className="me-2 h-4 w-4" />
            {previewMutation.isPending ? 'جارٍ الفحص...' : 'معاينة السجلات'}
          </Button>
          {preview && Number(preview.candidate_count || 0) > 0 && !confirming && (
            <Button variant="destructive" onClick={() => setConfirming(true)} disabled={runMutation.isPending}>
              حذف السجلات المحددة
            </Button>
          )}
          {confirming && (
            <>
              <Button variant="destructive" onClick={() => runMutation.mutate()} disabled={runMutation.isPending || !Array.isArray(preview?.candidate_ids) || preview.candidate_ids.length === 0 || Number(preview?.candidate_count || 0) > Number(preview?.max_batch || 0)}>
                {runMutation.isPending ? 'جارٍ التنظيف...' : 'تأكيد التنظيف'}
              </Button>
              <Button variant="secondary" onClick={() => setConfirming(false)} disabled={runMutation.isPending}>إلغاء</Button>
            </>
          )}
        </div>
        {preview && (
          <div className="space-y-2 rounded-lg border border-slate-200 bg-slate-50 p-4 text-sm">
            <p className="font-semibold text-slate-800">
              {preview.label}: {Number(preview.candidate_count || 0).toLocaleString()} سجل بين {preview.start_date} و{preview.end_date}
            </p>
            <p className="text-slate-600">المنطقة الزمنية المستخدمة: {preview.timezone || getStoreTimezone()}</p>
            <p className="text-slate-600">{preview.note}</p>
            {Number(preview.candidate_count || 0) > Number(preview.max_batch || 0) && (
              <p className="font-medium text-amber-800">الفترة أكبر من الحد الآمن ({Number(preview.max_batch).toLocaleString()} سجل). قلّص الفترة ثم أعد المعاينة.</p>
            )}
            {Number(preview.candidate_count || 0) === 0 && <p className="text-slate-600">لا توجد سجلات ضمن هذه الفترة.</p>}
          </div>
        )}
        {lastResult && Array.isArray(lastResult.issues) && lastResult.issues.length > 0 && (
          <div className="max-h-72 space-y-2 overflow-y-auto rounded-lg border border-amber-300 bg-amber-50 p-4 text-sm" role="status">
            <p className="font-semibold text-amber-950">تعذر حذف بعض السجلات. لم تُحذف هذه السجلات؛ عولج السبب ثم أعد المحاولة.</p>
            {lastResult.issues.map((issue: any, index: number) => (
              <div key={`${issue.id}-${index}`} className="rounded-md border border-amber-200 bg-white p-3">
                <p className="font-medium text-slate-900">المعرّف: <span dir="ltr" className="break-all">{issue.id || 'غير متاح'}</span></p>
                <p className="mt-1 text-slate-700">{issue.reason}</p>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
