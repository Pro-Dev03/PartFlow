import { useEffect, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { authApi, settingsApi } from '../../../services/api/endpoints';
import { RefreshCw, Wifi, Check, X, HardDrive } from 'lucide-react';
import { toast } from 'sonner';
import { getLocalApiUrl } from '../../../lib/config/app';

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
  const [connectionStatus, setConnectionStatus] = useState<'unknown' | 'testing' | 'connected' | 'failed'>('unknown');
  const [connectionError, setConnectionError] = useState<string | null>(null);

  const validateSubscriptionMutation = useMutation({
    mutationFn: async () => {
      await authApi.refreshToken();
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
    void testLocalConnection();
  }, []);

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
            تسجيل الدخول والتحقق من الاشتراك يتمان عبر الخادم السحابي. عمليات المتجر تُحفظ في قاعدة SQLite المحلية، مع إمكانية المزامنة عبر الخادم السحابي عند الحاجة.
          </p>
          <div className="flex items-center justify-between gap-3 pt-2">
            <p className="text-xs text-gray-500">
              قاعدة البيانات التشغيلية: SQLite محلية
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
              <Button
                variant="secondary"
                className="min-h-10"
                onClick={() => syncMutation.mutate()}
                disabled={syncMutation.isPending || typeof navigator !== 'undefined' && !navigator.onLine}
              >
                {syncMutation.isPending ? 'جارِ المزامنة...' : 'مزامنة سحابية الآن'}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {(
        <Card className="border-amber-200 bg-amber-50/50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <HardDrive className="w-5 h-5 text-amber-600" />
              قاعدة البيانات المحلية
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-gray-600">
              يتم تخزين البيانات محلياً على جهازك في SQLite. يمكن التحقق والمزامنة مع الخادم السحابي عند الحاجة.
            </p>

            <div className="rounded-lg bg-white p-3 border border-amber-200">
              <div className="text-xs text-gray-500 mb-1">الموقع المحلي</div>
              <div className="font-mono text-sm text-gray-800 break-all">{localApiUrl}</div>
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

            <div className="p-3 bg-amber-100 rounded-lg border border-amber-300">
              <p className="text-xs text-amber-900">
                ⚠️ <strong>تنبيه:</strong> البيانات محلية فقط. تأكّد من نسخ البيانات احتياطياً قبل مسح التطبيق.
              </p>
            </div>
          </CardContent>
        </Card>
      )}


    </div>
  );
}
