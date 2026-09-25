import { useEffect, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { settingsApi } from '../../../services/api/endpoints';
import { useAuthStore } from '../../../stores/authStore';
import { RefreshCw, Wifi, Check, X, HardDrive, Trash2, Download, Upload } from 'lucide-react';
import { toast } from 'sonner';
import { getCloudApiUrl, getLocalApiUrl, setCloudApiUrl } from '../../../lib/config/app';

const getHealthUrl = (baseUrl: string) => {
  try {
    const url = new URL(baseUrl);
    url.pathname = '/health';
    return url.toString();
  } catch {
    return `${baseUrl.replace(/\/api\/v1\/?$/, '')}/health`;
  }
};

export function DatabaseSettings() {
  const queryClient = useQueryClient();
  const [confirmationText, setConfirmationText] = useState('');
  const [showConfirmation, setShowConfirmation] = useState(false);

  const [localApiUrl] = useState(getLocalApiUrl());
  const [isTestingConnection, setIsTestingConnection] = useState(false);
  const [isTestingCloudConnection, setIsTestingCloudConnection] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<'unknown' | 'testing' | 'connected' | 'failed'>('unknown');
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const [cloudUrl, setCloudUrl] = useState(getCloudApiUrl());
  const [isSavingCloudUrl, setIsSavingCloudUrl] = useState(false);
  const [isBackingUp, setIsBackingUp] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);
  const isDesktop = typeof window !== 'undefined' && Boolean(window.partflowDesktop);

  const createLocalBackup = async () => {
    const backup = window.partflowDesktop?.database?.backup;
    if (!backup) {
      toast.error('النسخ الاحتياطي متاح من تطبيق PartFlow المكتبي.');
      return;
    }
    setIsBackingUp(true);
    try {
      const result = await backup();
      if (!result.canceled && result.filePath) {
        toast.success('تم إنشاء النسخة الاحتياطية والتحقق من سلامتها.', { description: result.filePath });
      }
    } catch (error: any) {
      toast.error(`تعذر إنشاء النسخة الاحتياطية: ${error?.message || 'حدث خطأ غير متوقع'}`);
    } finally {
      setIsBackingUp(false);
    }
  };

  const restoreLocalBackup = async () => {
    const restore = window.partflowDesktop?.database?.restore;
    if (!restore) {
      toast.error('استعادة قاعدة البيانات متاحة من تطبيق PartFlow المكتبي.');
      return;
    }
    setIsRestoring(true);
    try {
      const result = await restore();
      if (!result.canceled) {
        void queryClient.clear();
        toast.success('تمت استعادة قاعدة البيانات. سيُعاد تحميل النظام الآن.', {
          description: result.recoveryPath ? `نسخة الرجوع محفوظة في: ${result.recoveryPath}` : undefined,
        });
        window.setTimeout(() => window.location.reload(), 900);
      }
    } catch (error: any) {
      toast.error(`تعذرت استعادة قاعدة البيانات: ${error?.message || 'حدث خطأ غير متوقع'}`);
    } finally {
      setIsRestoring(false);
    }
  };

  const saveCloudUrl = async () => {
    const normalized = cloudUrl.trim().replace(/\/+$/, '');
    let parsed: URL;
    try {
      parsed = new URL(normalized);
    } catch {
      toast.error('رابط الخادم السحابي غير صالح.');
      return;
    }
    if (parsed.protocol !== 'https:') {
      toast.error('يجب أن يبدأ رابط الخادم السحابي بـ https://');
      return;
    }
    if (normalized !== getCloudApiUrl()) {
      toast.error('رابط API السحابي مثبت على خادم Render المعتمد.');
      return;
    }

    setIsSavingCloudUrl(true);
    try {
      const response = await fetch(getHealthUrl(normalized), { headers: { Accept: 'application/json' } });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      setCloudApiUrl(normalized);
      setCloudUrl(normalized);
      toast.success('تم حفظ رابط الخادم السحابي والتحقق من اتصاله.');
    } catch (error: any) {
      toast.error(`تعذر الاتصال بالخادم السحابي: ${error?.message || 'تحقق من الرابط'}`);
    } finally {
      setIsSavingCloudUrl(false);
    }
  };

  const testCloudConnection = async () => {
    setIsTestingCloudConnection(true);
    try {
      const response = await fetch(getHealthUrl(getCloudApiUrl()), { headers: { Accept: 'application/json' } });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      toast.success('اتصال Render السحابي يعمل.');
    } catch (error: any) {
      toast.error(`تعذر الاتصال بـ Render: ${error?.message || 'تحقق من الاتصال'}`);
    } finally {
      setIsTestingCloudConnection(false);
    }
  };

  const validateSubscriptionMutation = useMutation({
    mutationFn: async () => {
      // Refresh through the auth store so the rotated cloud refresh token and
      // access token are persisted consistently. Calling the endpoint directly
      // would revoke the old token on the server while leaving it in storage.
      await useAuthStore.getState().refreshToken();
      return true;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries();
      toast.success('تم التحقق من اشتراكك بنجاح. لا توجد مشكلة في صلاحية الحساب.');
    },
    onError: (error: any) => {
      const message = error?.response?.data?.error || error?.message || 'تعذر التحقق من الاشتراك';
      toast.error(message);
    },
  });

  const syncMutation = useMutation({
    mutationFn: () => settingsApi.syncCloudData(),
    onSuccess: () => {
      void queryClient.invalidateQueries();
      toast.success('تم تنزيل أحدث نسخة من السحابة إلى SQLite المحلية.');
    },
    onError: (error: any) => {
      toast.error('فشل المزامنة السحابية: ' + (error?.message || 'تحقق من اتصال الإنترنت'));
    },
  });

  const pushSyncMutation = useMutation({
    mutationFn: () => settingsApi.syncLocalDataToCloud(),
    onSuccess: (result: any) => {
      void queryClient.invalidateQueries();
      const summary = result?.data ?? result;
      const processed = Number(summary?.processed || 0);
      const failed = Number(summary?.failed || 0);
      toast.success(`تمت مزامنة ${processed} عملية محلية إلى السحابة${failed ? `، وفشلت ${failed}` : ''}.`);
    },
    onError: (error: any) => {
      toast.error('فشل رفع التغييرات المحلية: ' + (error?.message || 'تحقق من اتصال الإنترنت وصلاحية الحساب'));
    },
  });

  const deleteCloudDataMutation = useMutation({
    mutationFn: () => settingsApi.deleteCloudData(confirmationText),
    onSuccess: () => {
      setConfirmationText('');
      setShowConfirmation(false);
      void queryClient.invalidateQueries();
      toast.success('تم حذف البيانات التشغيلية والسجلات التاريخية السحابية نهائيًا. بقي حساب المالك والإعدادات محفوظين.');
    },
    onError: (error: any) => {
      toast.error(error?.message || 'فشل حذف بيانات السحابة');
    },
  });

  const deleteLocalDataMutation = useMutation({
    mutationFn: () => settingsApi.deleteAllData(confirmationText),
    onSuccess: () => {
      setConfirmationText('');
      setShowConfirmation(false);
      void queryClient.invalidateQueries();
      toast.success('تم حذف البيانات التشغيلية والسجلات التاريخية المحلية نهائيًا. بقيت بنية SQLite وإعدادات الحساب محفوظة.');
    },
    onError: (error: any) => {
      toast.error(error?.message || 'فشل حذف البيانات المحلية');
    },
  });

  // Test local backend connection
  const testLocalConnection = async () => {
    setIsTestingConnection(true);
    setConnectionStatus('testing');
    setConnectionError(null);
    
    try {
      const response = await fetch(getHealthUrl(localApiUrl), {
        method: 'GET',
        headers: {
          'Accept': 'application/json',
        },
      });
      
      if (response.ok) {
        setConnectionStatus('connected');
        setConnectionError(null);
        toast.success('✓ متصل بقاعدة البيانات المحلية بنجاح');
      } else {
        setConnectionStatus('failed');
        const errorText = await response.text();
        setConnectionError(`HTTP ${response.status}: ${errorText || 'Unknown error'}`);
        toast.error(`✗ فشل الاتصال: ${response.status}`);
      }
    } catch (error: any) {
      setConnectionStatus('failed');
      const errorMsg = error?.message || 'فشل الاتصال بقاعدة البيانات المحلية';
      setConnectionError(errorMsg);
      toast.error('✗ ' + errorMsg);
    } finally {
      setIsTestingConnection(false);
    }
  };

  useEffect(() => {
    if (isDesktop) void testLocalConnection();
  }, [isDesktop]);

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Wifi className="w-5 h-5" />
            اتصال النظام
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-gray-600">
            تسجيل الدخول والتحقق من الاشتراك وجميع عمليات المتجر تتم عبر Render. قاعدة SQLite على الجهاز نسخة محلية احتياطية فقط.
          </p>
          <div className="flex items-center justify-between gap-3 pt-2">
            <p className="text-xs text-gray-500">
              API السحابي: <span dir="ltr" className="font-mono">{getCloudApiUrl()}</span>
            </p>
            <div className="flex flex-wrap items-center gap-2">
              <Button
                variant="secondary"
                className="min-h-10"
                onClick={() => validateSubscriptionMutation.mutate()}
                disabled={validateSubscriptionMutation.isPending}
              >
                {validateSubscriptionMutation.isPending ? 'جارِ التحقق...' : 'تحقق من الاشتراك'}
              </Button>
              <Button variant="secondary" className="min-h-10" onClick={() => void testCloudConnection()} disabled={isTestingCloudConnection}>
                {isTestingCloudConnection ? 'جارِ فحص Render...' : 'فحص اتصال Render'}
              </Button>
              {isDesktop && (
                <>
                  <Button
                    variant="secondary"
                    className="min-h-10"
                    onClick={() => syncMutation.mutate()}
                    disabled={syncMutation.isPending || typeof navigator !== 'undefined' && !navigator.onLine}
                  >
                    {syncMutation.isPending ? 'جارِ التنزيل...' : 'تنزيل نسخة SQLite'}
                  </Button>
                  <Button
                    variant="secondary"
                    className="min-h-10"
                    onClick={() => pushSyncMutation.mutate()}
                    disabled={pushSyncMutation.isPending || typeof navigator !== 'undefined' && !navigator.onLine}
                  >
                    {pushSyncMutation.isPending ? 'جارِ رفع الاستيراد...' : 'استيراد سجلات قديمة (مسؤول فقط)'}
                  </Button>
                </>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {isDesktop && (
        <Card className="border-amber-200 bg-amber-50/50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <HardDrive className="w-5 h-5 text-amber-600" />
              قاعدة البيانات المحلية
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-gray-600">
              قاعدة SQLite تخص هذا الجهاز وتُستخدم للنسخ الاحتياطي المحلي. عمليات المتجر الجديدة تعتمد على Render.
            </p>

            <div className="rounded-lg bg-white p-3 border border-amber-200">
              <div className="text-xs text-gray-500 mb-1">الموقع المحلي</div>
              <div className="font-mono text-sm text-gray-800 break-all">{localApiUrl}</div>
            </div>

            <div className="rounded-lg bg-white p-3 border border-amber-200 space-y-2">
              <div className="text-xs text-gray-500">رابط الخادم السحابي</div>
              <div className="flex flex-col gap-2 sm:flex-row">
                <input
                  value={cloudUrl}
                  onChange={(event) => setCloudUrl(event.target.value)}
                  className="min-h-10 flex-1 rounded-md border border-gray-300 px-3 text-sm"
                  dir="ltr"
                  inputMode="url"
                  aria-label="رابط الخادم السحابي"
                />
                <Button
                  variant="secondary"
                  className="min-h-10"
                  onClick={() => void saveCloudUrl()}
                  disabled={isSavingCloudUrl}
                >
                  {isSavingCloudUrl ? 'جارِ التحقق...' : 'حفظ الرابط'}
                </Button>
              </div>
              <p className="text-xs text-gray-500">يُستخدم هذا الرابط للمصادقة والتحقق والمزامنة السحابية.</p>
            </div>

            {connectionStatus !== 'unknown' && (
              <div className={`p-3 rounded-lg flex items-center gap-2 ${
                connectionStatus === 'connected' ? 'bg-green-50 border border-green-200' :
                connectionStatus === 'testing' ? 'bg-blue-50 border border-blue-200' :
                'bg-red-50 border border-red-200'
              }`}>
                {connectionStatus === 'connected' && (
                  <>
                    <Check className="w-5 h-5 text-green-600 flex-shrink-0" />
                    <span className="text-sm font-medium text-green-800">✓ قاعدة البيانات المحلية جاهزة</span>
                  </>
                )}
                {connectionStatus === 'testing' && (
                  <>
                    <RefreshCw className="w-5 h-5 text-blue-600 animate-spin flex-shrink-0" />
                    <span className="text-sm font-medium text-blue-800">جارِ الفحص...</span>
                  </>
                )}
                {connectionStatus === 'failed' && (
                  <>
                    <X className="w-5 h-5 text-red-600 flex-shrink-0" />
                    <div className="flex-1">
                      <div className="text-sm font-medium text-red-800">⚠️ مشكلة في البيانات المحلية</div>
                      {connectionError && <div className="text-xs text-red-700 mt-1">{connectionError}</div>}
                    </div>
                  </>
                )}
              </div>
            )}

            <Button
              variant="secondary"
              className="w-full"
              onClick={testLocalConnection}
              disabled={isTestingConnection}
            >
              {isTestingConnection ? (
                <>
                  <RefreshCw className="w-4 h-4 ml-2 animate-spin" />
                  جارِ الفحص...
                </>
              ) : (
                <>
                  <HardDrive className="w-4 h-4 ml-2" />
                  فحص قاعدة البيانات
                </>
              )}
            </Button>

            {window.partflowDesktop?.database && (
              <div className="flex flex-col gap-2 sm:flex-row">
                <Button
                  variant="secondary"
                  className="min-h-10 flex-1"
                  onClick={() => void createLocalBackup()}
                  disabled={isBackingUp || isRestoring}
                >
                  <Download className="ml-2 h-4 w-4" />
                  {isBackingUp ? 'جارٍ إنشاء النسخة والتحقق منها...' : 'إنشاء نسخة احتياطية'}
                </Button>
                <Button
                  variant="secondary"
                  className="min-h-10 flex-1"
                  onClick={() => void restoreLocalBackup()}
                  disabled={isBackingUp || isRestoring}
                >
                  <Upload className="ml-2 h-4 w-4" />
                  {isRestoring ? 'جارٍ التحقق والاستعادة...' : 'استعادة نسخة احتياطية'}
                </Button>
              </div>
            )}

            <div className="p-3 bg-amber-100 rounded-lg border border-amber-300">
              <p className="text-xs text-amber-900">
                ⚠️ <strong>تنبيه:</strong> الحذف النهائي يزيل السجلات التشغيلية والتاريخية. تأكّد من أخذ نسخة احتياطية قبل المتابعة.
              </p>
            </div>

            <div className="space-y-3 rounded-lg border border-red-200 bg-red-50 p-4">
              <div className="flex items-center gap-2 text-red-800">
                <Trash2 className="h-5 w-5" />
                <h3 className="font-semibold">تنظيف البيانات والسجلات التاريخية</h3>
              </div>
              <p className="text-sm text-red-800">
                اختر النطاق بعناية. سيحذف الخيار المحدد جميع البيانات التشغيلية والسجلات التاريخية مثل المبيعات والمشتريات والمخزون والعملاء والديون والمصروفات، مع إبقاء حساب المالك والإعدادات وبنية قاعدة البيانات محفوظة.
              </p>
              {!showConfirmation ? (
                <Button
                  variant="destructive"
                  className="w-full"
                  onClick={() => setShowConfirmation(true)}
                >
                  <Trash2 className="ml-2 h-4 w-4" />
                  فتح خيارات التنظيف النهائي
                </Button>
              ) : (
                <div className="space-y-2">
                  <label htmlFor="cloud-delete-confirmation" className="text-sm font-medium text-red-900">
                    اكتب DELETE ALL DATA للتأكيد
                  </label>
                  <input
                    id="cloud-delete-confirmation"
                    value={confirmationText}
                    onChange={(event) => setConfirmationText(event.target.value)}
                    className="w-full rounded-md border border-red-300 bg-white px-3 py-2 text-sm"
                    autoComplete="off"
                    spellCheck={false}
                  />
                  <div className="flex flex-wrap gap-2">
                    <Button
                      variant="destructive"
                      onClick={() => deleteCloudDataMutation.mutate()}
                      disabled={confirmationText !== 'DELETE ALL DATA' || deleteCloudDataMutation.isPending || deleteLocalDataMutation.isPending}
                    >
                      {deleteCloudDataMutation.isPending ? 'جارِ الحذف...' : 'حذف نهائي من السحابة'}
                    </Button>
                    {isDesktop && (
                      <Button
                        variant="destructive"
                        onClick={() => deleteLocalDataMutation.mutate()}
                        disabled={confirmationText !== 'DELETE ALL DATA' || deleteCloudDataMutation.isPending || deleteLocalDataMutation.isPending}
                      >
                        {deleteLocalDataMutation.isPending ? 'جارِ الحذف...' : 'حذف نهائي من SQLite المحلية'}
                      </Button>
                    )}
                    <Button
                      variant="secondary"
                      onClick={() => {
                        setConfirmationText('');
                        setShowConfirmation(false);
                      }}
                      disabled={deleteCloudDataMutation.isPending}
                    >
                      إلغاء
                    </Button>
                  </div>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}


    </div>
  );
}
