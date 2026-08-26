import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { settingsApi } from '../../../services/api/endpoints';
import { Trash2, AlertTriangle, Shield, RefreshCw } from 'lucide-react';

export function DatabaseSettings() {
  const [confirmationText, setConfirmationText] = useState('');
  const [showConfirmation, setShowConfirmation] = useState(false);

  const deleteAllDataMutation = useMutation({
    mutationFn: () => settingsApi.deleteAllData(),
    onSuccess: () => {
      alert('تم حذف جميع البيانات بنجاح! سيتم إعادة تحميل الصفحة.');
      window.location.reload();
    },
    onError: (error: any) => {
      console.error('Failed to delete all data:', error);
      alert('فشل حذف البيانات: ' + (error.response?.data?.error || error.message));
    },
  });

  const handleDeleteAllData = () => {
    if (confirmationText === 'احذف جميع البيانات') {
      deleteAllDataMutation.mutate();
    } else {
      alert('النص المدخل غير صحيح. يجب كتابة "احذف جميع البيانات" للتأكيد.');
    }
  };

  return (
    <div className="space-y-6">
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
                <h4 className="font-semibold text-red-800 mb-2">حذف جميع البيانات</h4>
                <p className="text-sm text-red-700 mb-3">
                  هذا الإجراء سيحذف جميع البيانات من قاعدة البيانات بشكل نهائي، بما في ذلك:
                </p>
                <ul className="text-sm text-red-700 list-disc list-inside space-y-1 mb-3">
                  <li>جميع المنتجات والمخزون</li>
                  <li>جميع الموردين والعملاء</li>
                  <li>جميع المبيعات والمشتريات</li>
                  <li>جميع الديون والمدفوعات</li>
                  <li>جميع الإشعارات والسجلات</li>
                </ul>
                <p className="text-xs text-red-600 font-semibold">
                  ⚠️ هذا الإجراء لا يمكن التراجع عنه!
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
              حذف جميع البيانات
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