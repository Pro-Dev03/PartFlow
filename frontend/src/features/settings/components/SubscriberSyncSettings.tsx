import { useMutation, useQueryClient } from '@tanstack/react-query';
import { CheckCircle2, CloudCog, Download, Upload, WifiOff, XCircle } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '../../../design-system/components/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { settingsApi } from '../../../services/api/endpoints';
import { getConnectionMode } from '../../../lib/config/app';

type UploadOperation = {
  id: string;
  entity_type: string;
  entity_id: string;
  operation: string;
  status: 'processed' | 'failed' | 'conflict';
  error?: string;
};

type UploadResult = {
  processed: number;
  failed: number;
  operations: UploadOperation[];
};

const operationLabels: Record<string, string> = {
  create: 'إضافة',
  update: 'تعديل',
  delete: 'حذف',
};

const entityLabels: Record<string, string> = {
  customer: 'عميل',
  customers: 'عملاء',
  product: 'منتج',
  products: 'منتجات',
  sale: 'بيع',
  sales: 'مبيعات',
  purchase: 'شراء',
  purchases: 'مشتريات',
  debt: 'دين',
  debts: 'ديون',
};

export function SubscriberSyncSettings() {
  const queryClient = useQueryClient();
  const connectionMode = getConnectionMode();
  const isOffline = typeof navigator !== 'undefined' && !navigator.onLine;

  const downloadMutation = useMutation({
    mutationFn: () => settingsApi.syncCloudData(),
    onSuccess: () => {
      void queryClient.invalidateQueries();
      toast.success('تم تنزيل أحدث البيانات السحابية إلى البيانات المحلية.');
    },
    onError: (error: any) => {
      toast.error(error?.message || 'تعذر تنزيل البيانات من السحابة.');
    },
  });

  const uploadMutation = useMutation({
    mutationFn: () => settingsApi.syncLocalDataToCloud(),
    onSuccess: (result: any) => {
      void queryClient.invalidateQueries();
      const summary = result?.data ?? result;
      const processed = Number(summary?.processed || 0);
      const failed = Number(summary?.failed || 0);
      toast.success(`تم رفع ${processed} تغيير محلي إلى السحابة${failed ? `، وفشل ${failed}` : ''}.`);
    },
    onError: (error: any) => {
      toast.error(error?.message || 'تعذر رفع التغييرات المحلية إلى السحابة.');
    },
  });

  const uploadResult = uploadMutation.data as { data?: UploadResult } | undefined;
  const uploadSummary = uploadResult?.data;

  const isBusy = downloadMutation.isPending || uploadMutation.isPending;

  return (
    <Card className="border-sky-200 bg-sky-50/50">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          <CloudCog className="h-5 w-5 text-sky-600" />
          مزامنة البيانات محليًا وسحابيًا
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm leading-6 text-slate-600">
          احتفظ بنسخة محلية للعمل السريع، وحدثها من السحابة أو ارفع التغييرات المحلية عند توفر الإنترنت.
          يتم التحقق من الاشتراك قبل تنفيذ أي مزامنة.
        </p>

        <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-sky-200 bg-white p-3">
          <div>
            <div className="text-xs text-slate-500">مصدر البيانات الحالي</div>
            <div className="mt-1 text-sm font-semibold text-slate-800">
              {connectionMode === 'cloud' ? 'السحابة مباشرة' : 'SQLite المحلية'}
            </div>
          </div>
          {isOffline ? (
            <div className="flex items-center gap-2 text-xs font-semibold text-amber-700">
              <WifiOff className="h-4 w-4" />
              لا يوجد اتصال بالإنترنت
            </div>
          ) : (
            <div className="text-xs font-semibold text-emerald-700">جاهز للمزامنة</div>
          )}
        </div>

        <div className="grid gap-3 sm:grid-cols-2">
          <Button
            variant="secondary"
            className="min-h-11 justify-center gap-2"
            onClick={() => downloadMutation.mutate()}
            disabled={isBusy || isOffline}
          >
            <Download className="h-4 w-4" />
            {downloadMutation.isPending ? 'جارٍ تنزيل البيانات...' : 'تنزيل البيانات'}
          </Button>
          <Button
            variant="secondary"
            className="min-h-11 justify-center gap-2"
            onClick={() => uploadMutation.mutate()}
            disabled={isBusy || isOffline}
          >
            <Upload className="h-4 w-4" />
            {uploadMutation.isPending ? 'جارٍ رفع البيانات...' : 'رفع البيانات'}
          </Button>
        </div>

        {uploadSummary && (
          <div className="rounded-lg border border-slate-200 bg-white p-4" dir="rtl">
            <div className="flex items-start justify-between gap-3">
              <div>
                <h3 className="text-sm font-bold text-slate-800">تفاصيل آخر عملية رفع</h3>
                <p className="mt-1 text-xs text-slate-500">
                  {uploadSummary.processed === 0 && uploadSummary.failed === 0
                    ? 'لا توجد تغييرات محلية معلقة تحتاج إلى رفع.'
                    : `تمت معالجة ${uploadSummary.processed} عملية${uploadSummary.failed ? `، مع ${uploadSummary.failed} عملية تحتاج إلى مراجعة` : ''}.`}
                </p>
              </div>
              <span className="shrink-0 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-bold text-emerald-700">
                {uploadSummary.processed} ناجحة
              </span>
            </div>

            {uploadSummary.operations.length > 0 && (
              <div className="mt-3 space-y-2">
                {uploadSummary.operations.map((operation) => {
                  const isProcessed = operation.status === 'processed';
                  return (
                    <div key={operation.id} className="flex items-start gap-2 rounded-md border border-slate-100 bg-slate-50 p-2.5 text-xs">
                      {isProcessed ? <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" /> : <XCircle className="mt-0.5 h-4 w-4 shrink-0 text-rose-600" />}
                      <div className="min-w-0 flex-1">
                        <div className="font-semibold text-slate-700">
                          {operationLabels[operation.operation] || operation.operation} {entityLabels[operation.entity_type] || operation.entity_type}
                        </div>
                        <div className="mt-0.5 truncate text-slate-500" dir="ltr">المعرّف: {operation.entity_id}</div>
                        {!isProcessed && <div className="mt-1 text-rose-700">{operation.status === 'conflict' ? 'تعارض يحتاج إلى مراجعة' : operation.error || 'تعذر رفع العملية'}</div>}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
