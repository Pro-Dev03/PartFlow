import { useMutation, useQueryClient } from '@tanstack/react-query';
import { CloudCog, Download, Upload, WifiOff } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '../../../design-system/components/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { settingsApi } from '../../../services/api/endpoints';
import { getConnectionMode } from '../../../lib/config/app';

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
      const processed = Number(result?.processed || 0);
      const failed = Number(result?.failed || 0);
      toast.success(`تم رفع ${processed} تغيير محلي إلى السحابة${failed ? `، وفشل ${failed}` : ''}.`);
    },
    onError: (error: any) => {
      toast.error(error?.message || 'تعذر رفع التغييرات المحلية إلى السحابة.');
    },
  });

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
            <div className="text-xs text-slate-500">مصدر العمليات الحالي</div>
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
            {downloadMutation.isPending ? 'جارٍ التنزيل...' : 'تنزيل من السحابة'}
          </Button>
          <Button
            variant="secondary"
            className="min-h-11 justify-center gap-2"
            onClick={() => uploadMutation.mutate()}
            disabled={isBusy || isOffline}
          >
            <Upload className="h-4 w-4" />
            {uploadMutation.isPending ? 'جارٍ الرفع...' : 'رفع التغييرات للسحابة'}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
