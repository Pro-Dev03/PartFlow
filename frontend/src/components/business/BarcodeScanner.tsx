import { useState, useEffect, useRef } from 'react';
import { useTranslation } from '../../hooks/useTranslation';
import { 
  BarcodeContext, 
  getBarcodeContextLabel, 
  getSuggestedAction,
  playScanSound 
} from '../../hooks/useBarcodeContext';
import { barcodeApi } from '../../services/api/endpoints';
import { Dialog } from '../ui/dialog';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Card, CardContent } from '../ui/card';
import { 
  Scan, 
  X, 
  Camera, 
  Keyboard,
  CheckCircle,
  AlertCircle,
  Package,
  ShoppingCart,
  Box,
  ArrowRightLeft,
  RotateCcw,
  Volume2,
  VolumeX
} from 'lucide-react';

interface BarcodeScannerProps {
  isOpen: boolean;
  onClose: () => void;
  onScanComplete?: (result: any) => void;
  context?: BarcodeContext;
}

export function BarcodeScanner({ isOpen, onClose, onScanComplete, context = BarcodeContext.LOOKUP }: BarcodeScannerProps) {
  const { t } = useTranslation();
  const [mode, setMode] = useState<'camera' | 'keyboard'>('keyboard');
  const [barcode, setBarcode] = useState('');
  const [isScanning, setIsScanning] = useState(false);
  const [scanResult, setScanResult] = useState<any>(null);
  const [error, setError] = useState('');
  const [soundEnabled, setSoundEnabled] = useState(true);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isOpen]);

  const handleManualSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!barcode.trim()) return;

    await processBarcode(barcode.trim());
  };

  const processBarcode = async (code: string) => {
    setIsScanning(true);
    setError('');
    setScanResult(null);

    try {
      const response = await barcodeApi.scan(code);
      const result = response.data;
      
      setScanResult(result);
      
      // Play success sound
      if (soundEnabled) {
        playScanSound(true);
      }
      
      if (onScanComplete) {
        onScanComplete({
          ...(result || {}),
          context,
          suggestedAction: getSuggestedAction(context, result),
        });
      }

      // Auto-close after successful scan
      setTimeout(() => {
        onClose();
        setBarcode('');
        setScanResult(null);
      }, 1500);
    } catch (err: any) {
      setError(err.message || 'فشل مسح الباركود');
      
      // Play error sound
      if (soundEnabled) {
        playScanSound(false);
      }
    } finally {
      setIsScanning(false);
    }
  };

  const handleCreateProduct = () => {
    // Navigate to product creation page
    window.location.href = `/app/products/new?barcode=${barcode}`;
    onClose();
  };

  const getActionButtons = () => {
    const suggestedAction = getSuggestedAction(context, scanResult);
    
    switch (suggestedAction) {
      case 'addToCart':
        return (
          <Button 
            className="w-full" 
            onClick={() => {
              if (onScanComplete) {
                onScanComplete({ ...scanResult, action: 'addToCart' });
              }
              onClose();
            }}
          >
            <ShoppingCart className="w-4 h-4 me-2" />
            إضافة للسلة
          </Button>
        );
      case 'outOfStock':
        return (
          <Button 
            className="w-full" 
            variant="danger"
            disabled
          >
            <AlertCircle className="w-4 h-4 me-2" />
            نفذ المخزون
          </Button>
        );
      case 'showDetails':
        return (
          <Button 
            className="w-full" 
            onClick={() => {
              if (onScanComplete) {
                onScanComplete({ ...scanResult, action: 'showDetails' });
              }
              onClose();
            }}
          >
            <Box className="w-4 h-4 me-2" />
            عرض التفاصيل
          </Button>
        );
      case 'releaseReservation':
        return (
          <Button 
            className="w-full" 
            variant="outline"
            onClick={() => {
              if (onScanComplete) {
                onScanComplete({ ...scanResult, action: 'releaseReservation' });
              }
              onClose();
            }}
          >
            <RotateCcw className="w-4 h-4 me-2" />
            إلغاء الحجز
          </Button>
        );
      case 'addToPurchase':
        return (
          <Button 
            className="w-full" 
            onClick={() => {
              if (onScanComplete) {
                onScanComplete({ ...scanResult, action: 'addToPurchase' });
              }
              onClose();
            }}
          >
            <ArrowRightLeft className="w-4 h-4 me-2" />
            إضافة لطلب الشراء
          </Button>
        );
      case 'processReturn':
        return (
          <Button 
            className="w-full" 
            onClick={() => {
              if (onScanComplete) {
                onScanComplete({ ...scanResult, action: 'processReturn' });
              }
              onClose();
            }}
          >
            <RotateCcw className="w-4 h-4 me-2" />
            معالجة المرتجع
          </Button>
        );
      default:
        return (
          <Button 
            className="w-full" 
            onClick={() => {
              if (onScanComplete) {
                onScanComplete({ ...scanResult, action: 'showDetails' });
              }
              onClose();
            }}
          >
            <Package className="w-4 h-4 me-2" />
            عرض التفاصيل
          </Button>
        );
    }
  };

  if (!isOpen) return null;

  return (
    <Dialog open={isOpen} onClose={onClose}>
      <Card className="w-full max-w-md">
        <CardContent className="p-6">
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">
                {t('scanner.title')}
              </h2>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                السياق: {getBarcodeContextLabel(context)}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button 
                variant="ghost" 
                size="sm" 
                onClick={() => setSoundEnabled(!soundEnabled)}
                title={soundEnabled ? 'إيقاف الصوت' : 'تشغيل الصوت'}
              >
                {soundEnabled ? (
                  <Volume2 className="w-4 h-4" />
                ) : (
                  <VolumeX className="w-4 h-4" />
                )}
              </Button>
              <Button variant="ghost" size="sm" onClick={onClose}>
                <X className="w-4 h-4" />
              </Button>
            </div>
          </div>

          {/* Mode Toggle */}
          <div className="flex gap-2 mb-6">
            <Button
              variant={mode === 'keyboard' ? 'primary' : 'outline'}
              className="flex-1 gap-2"
              onClick={() => setMode('keyboard')}
            >
              <Keyboard className="w-4 h-4" />
              {t('scanner.manualInput')}
            </Button>
            <Button
              variant={mode === 'camera' ? 'primary' : 'outline'}
              className="flex-1 gap-2"
              onClick={() => setMode('camera')}
            >
              <Camera className="w-4 h-4" />
              {t('scanner.cameraScanner')}
            </Button>
          </div>

          {/* Scan Result */}
          {scanResult ? (
            <div className="space-y-4">
              <div className="flex items-center gap-3 p-4 bg-green-50 dark:bg-green-900/20 rounded-lg border border-green-200 dark:border-green-800">
                <CheckCircle className="w-5 h-5 text-green-600" />
                <div>
                  <p className="font-medium text-green-900 dark:text-green-100">
                    {t('scanner.barcodeFound')}
                  </p>
                  <p className="text-sm text-green-700 dark:text-green-300">
                    {scanResult.name}
                  </p>
                </div>
              </div>

              {getActionButtons()}
            </div>
          ) : (
            <>
              {/* Error State */}
              {error && (
                <div className="flex items-center gap-3 p-4 bg-red-50 dark:bg-red-900/20 rounded-lg border border-red-200 dark:border-red-800 mb-4">
                  <AlertCircle className="w-5 h-5 text-red-600" />
                  <div className="flex-1">
                    <p className="font-medium text-red-900 dark:text-red-100">
                      {t('scanner.barcodeNotFound')}
                    </p>
                    <p className="text-sm text-red-700 dark:text-red-300">
                      {error}
                    </p>
                  </div>
                </div>
              )}

              {/* Manual Input */}
              {mode === 'keyboard' && (
                <form onSubmit={handleManualSubmit} className="space-y-4">
                  <div>
                    <label className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 block">
                      {t('scanner.scanBarcode')}
                    </label>
                    <Input
                      ref={inputRef}
                      value={barcode}
                      onChange={(e) => setBarcode(e.target.value)}
                      placeholder={t('scanner.scanBarcode')}
                      autoFocus
                    />
                  </div>
                  <Button
                    type="submit"
                    className="w-full gap-2"
                    isLoading={isScanning}
                    disabled={!barcode.trim()}
                  >
                    <Scan className="w-4 h-4" />
                    {t('scanner.scan')}
                  </Button>
                </form>
              )}

              {/* Camera Scanner */}
              {mode === 'camera' && (
                <div className="space-y-4">
                  <div className="aspect-video bg-gray-100 dark:bg-gray-800 rounded-lg flex items-center justify-center">
                    <div className="text-center">
                      <Camera className="w-12 h-12 mx-auto mb-2 text-gray-400" />
                      <p className="text-sm text-gray-500 dark:text-gray-400">
                        {t('scanner.scanning')}
                      </p>
                    </div>
                  </div>
                  <Button
                    className="w-full gap-2"
                    onClick={() => {
                      // Simulate camera scan
                      processBarcode('BC-' + Date.now());
                    }}
                  >
                    <Scan className="w-4 h-4" />
                    {t('scanner.scan')}
                  </Button>
                </div>
              )}

              {/* Create Product Option */}
              {error && (
                <Button
                  variant="outline"
                  className="w-full gap-2 mt-4"
                  onClick={handleCreateProduct}
                >
                  <Package className="w-4 h-4" />
                  {t('scanner.createProduct')}
                </Button>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </Dialog>
  );
}