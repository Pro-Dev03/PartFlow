import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { settingsApi } from '../../../services/api/endpoints';
import { Trash2, AlertTriangle, Shield, RefreshCw, Wifi, WifiOff } from 'lucide-react';
import { toast } from 'sonner';

export function DatabaseSettings() {
  const [operatingMode, setOperatingMode] = useState<'offline' | 'online'>(() => {
    return localStorage.getItem('partflow-operating-mode') === 'online' ? 'online' : 'offline';
  });
  const [resetTarget, setResetTarget] = useState<'offline' | 'online' | 'current'>('current');
  const [confirmationText, setConfirmationText] = useState('');
  const [showConfirmation, setShowConfirmation] = useState(false);

  const handleOperatingModeChange = (mode: 'offline' | 'online') => {
    const previousMode = operatingMode;
    setOperatingMode(mode);
    settingsApi.setOperatingMode(mode).then(() => {
      localStorage.setItem('partflow-operating-mode', mode);
      toast.success(mode === 'offline' ? 'تم تفعيل قاعدة البيانات المحلية.' : 'تم اختيار مزامنة البيانات السحابية.');
    }).catch((error: any) => {
      setOperatingMode(previousMode);
      toast.error('تعذر تهيئة قاعدة البيانات المحلية: ' + (error?.message || 'خطأ غير معروف'));
    });
  };

  const deleteAllDataMutation = useMutation({
    mutationFn: () => settingsApi.deleteAllData('احذف جميع البيانات', resetTarget),
    onSuccess: () => {
      toast.success('تم تصفير النظام مع الحفاظ على حساب المالك والإعدادات. سيتم إعادة تحميل الصفحة.');
      window.location.reload();
    },
    onError: (error: any) => {
      console.error('Failed to delete all data:', error);
      toast.error('فشل حذف البيانات: ' + (error.response?.data?.error || error.message));
    },
  });

  const handleDeleteAllData = () => {
    if (confirmationText === 'احذف جميع البيانات') {
      deleteAllDataMutation.mutate();
    } else {
      toast.error('النص المدخل غير صحيح. يجب كتابة "احذف جميع البيانات" للتأكيد.');
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Wifi className="w-5 h-5" />
            وضع التشغيل
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-gray-600">
            اختر طريقة تشغيل المتجر على هذا الجهاز. عند اختيار الأوفلاين ينشئ النظام قاعدة SQLite محلية.
          </p>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3" role="group" aria-label="وضع التشغيل">
            <Button
              variant={operatingMode === 'offline' ? 'primary' : 'secondary'}
              className="justify-center gap-2 min-h-11"
              onClick={() => handleOperatingModeChange('offline')}
              aria-pressed={operatingMode === 'offline'}
            >
              <WifiOff className="w-4 h-4" />
              أوفلاين فقط
            </Button>
            <Button
              variant={operatingMode === 'online' ? 'primary' : 'secondary'}
              className="justify-center gap-2 min-h-11"
              onClick={() => handleOperatingModeChange('online')}
              aria-pressed={operatingMode === 'online'}
            >
              <RefreshCw className="w-4 h-4" />
              مزامنة البيانات السحابية
            </Button>
          </div>
          <p className="text-xs text-gray-500">
            الوضع المحدد: {operatingMode === 'offline' ? 'أوفلاين فقط' : 'مزامنة البيانات السحابية'}
          </p>
        </CardContent>
      </Card>

      {/* Danger Zone */}
      <Card className="border-red-200 bg-red-50/50">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-red-600">
            <AlertTriangle className="w-5 h-5" />
            منطقة الخطر
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="p-4 bg-red-100 rounded-lg border border-red-200">
            <div className="flex items-start gap-3">
              <Shield className="w-5 h-5 text-red-600 mt-0.5" />
              <div>
                <h4 className="font-semibold text-red-800 mb-2">تصفير النظام</h4>
                <p className="text-sm text-red-700 mb-3">
                  هذا الإجراء يحذف بيانات التشغيل نهائياً ويُبقي حساب المالك وإعدادات النظام:
                </p>
                <div className="mb-3 grid grid-cols-1 sm:grid-cols-3 gap-2">
                  {(['current', 'online', 'offline'] as const).map((option) => (
                    <Button
                      key={option}
                      type="button"
                      variant={resetTarget === option ? 'primary' : 'secondary'}
                      className="justify-center min-h-10"
                      onClick={() => setResetTarget(option)}
                    >
                      {option === 'current' && 'الحالية'}
                      {option === 'online' && 'أونلاين'}
                      {option === 'offline' && 'أوفلاين'}
                    </Button>
                  ))}
                </div>
                <ul className="text-sm text-red-700 list-disc list-inside space-y-1 mb-3">
                  <li>جميع المنتجات والمخزون</li>
                  <li>جميع الموردين والعملاء والبيانات التشغيلية</li>
                  <li>جميع المبيعات والمشتريات</li>
                  <li>جميع الديون والمدفوعات</li>
                  <li>جميع الإشعارات والسجلات والأرشيفات</li>
                </ul>
                <p className="text-sm text-green-700 font-semibold mb-3">
                  {resetTarget === 'current' && 'سيتم تصفير قاعدة البيانات الحالية وفق الوضع المحدد.'}
                  {resetTarget === 'online' && 'سيتم تصفير قاعدة البيانات السحابية فقط.'}
                  {resetTarget === 'offline' && 'سيتم تصفير قاعدة البيانات المحلية فقط مع الاحتفاظ بوضع التشغيل.'}
                </p>
                <p className="text-xs text-red-600 font-semibold">
                  ⚠️ سيبقى حساب المالك وإعدادات النظام. هذا الإجراء لا يمكن التراجع عنه!
                </p>
              </div>
            </div>
          </div>

          {!showConfirmation ? (
            <Button
              variant="danger"
              className="w-full"
              onClick={() => setShowConfirmation(true)}
              disabled={deleteAllDataMutation.isPending}
            >
              <Trash2 className="w-4 h-4 ml-2" />
              تصفير النظام
            </Button>
          ) : (
            <div className="space-y-3 p-4 bg-white rounded-lg border border-red-200">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  للتأكيد، اكتب: <span className="font-mono bg-gray-100 px-2 py-1 rounded">احذف جميع البيانات</span>
                </label>
                <input
                  type="text"
                  value={confirmationText}
                  onChange={(e) => setConfirmationText(e.target.value)}
                  placeholder="اكتب النص للتأكيد..."
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500"
                />
              </div>
              <div className="flex gap-2">
                <Button
                  variant="secondary"
                  onClick={() => {
                    setShowConfirmation(false);
                    setConfirmationText('');
                  }}
                  disabled={deleteAllDataMutation.isPending}
                >
                  إلغاء
                </Button>
                <Button
                  variant="danger"
                  onClick={handleDeleteAllData}
                  disabled={deleteAllDataMutation.isPending || confirmationText !== 'احذف جميع البيانات'}
                  className="flex-1"
                >
                  {deleteAllDataMutation.isPending ? (
                    <>
                      <RefreshCw className="w-4 h-4 ml-2 animate-spin" />
                      جاري الحذف...
                    </>
                  ) : (
                    <>
                      <Trash2 className="w-4 h-4 ml-2" />
                      تأكيد الحذف النهائي
                    </>
                  )}
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Info Card */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">معلومات قاعدة البيانات</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 text-sm text-gray-600">
            <p>النظام الحالي: PostgreSQL</p>
            <p>الوضع: Development</p>
            <p>آخر نسخة احتياطية: غير متوفر</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}