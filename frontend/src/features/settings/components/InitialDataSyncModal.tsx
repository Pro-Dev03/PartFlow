import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { RefreshCw, CheckCircle, AlertCircle } from 'lucide-react';
import { useInitialDataSync, isInitialSyncNeeded } from '../../../hooks/useInitialDataSync';

interface InitialDataSyncModalProps {
  isOpen: boolean;
  onComplete: () => void;
  userId?: string;
}

export function InitialDataSyncModal({ isOpen, onComplete, userId }: InitialDataSyncModalProps) {
  const [shouldSync, setShouldSync] = useState(false);

  // Trigger sync only when modal opens
  useEffect(() => {
    if (isOpen && isInitialSyncNeeded(userId)) {
      setShouldSync(true);
    }
  }, [isOpen, userId]);

  const { isLoading, isError, error, progress } = useInitialDataSync(
    shouldSync,
    () => {
      // Sync completed successfully
      setShouldSync(false);
      onComplete();
    },
    (err) => {
      // Sync failed
      console.error('Initial data sync failed:', err);
    },
    userId
  );

  if (!isOpen || !shouldSync) {
    return null;
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-center">تحميل البيانات من الخادم</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {isLoading && (
            <>
              <div className="flex justify-center">
                <RefreshCw className="w-12 h-12 text-blue-600 animate-spin" />
              </div>
              <div className="text-center">
                <p className="text-sm text-gray-600 mb-2">جارِ تحميل البيانات...</p>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${progress}%` }}
                  />
                </div>
                <p className="text-xs text-gray-500 mt-2">{progress}%</p>
              </div>
              <p className="text-xs text-center text-gray-500">
                يرجى الانتظار... لا تغلق التطبيق أثناء التحميل
              </p>
            </>
          )}

          {isError && error && (
            <>
              <div className="flex justify-center">
                <AlertCircle className="w-12 h-12 text-red-600" />
              </div>
              <div className="text-center">
                <p className="text-sm font-medium text-red-800 mb-2">فشل تحميل البيانات</p>
                <p className="text-xs text-red-700">{error.message}</p>
              </div>
              <p className="text-xs text-center text-gray-500">
                يتم إعادة المحاولة تلقائياً...
              </p>
            </>
          )}

          {!isLoading && !isError && progress === 100 && (
            <>
              <div className="flex justify-center">
                <CheckCircle className="w-12 h-12 text-green-600" />
              </div>
              <div className="text-center">
                <p className="text-sm font-medium text-green-800">
                  ✓ تم تحميل البيانات بنجاح
                </p>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
