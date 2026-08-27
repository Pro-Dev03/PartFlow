import { useRef } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { ItemInputMethod } from '../../../components/ui/item-input-method';
import { CameraScanner } from '../../../components/ui/camera-scanner';
import { Camera } from 'lucide-react';
import { useTranslation } from '../../../hooks/useTranslation';

interface BarcodeScannerProps {
  barcodeInput: string;
  setBarcodeInput: (value: string) => void;
  inputMethod: 'barcode' | 'camera';
  setInputMethod: (method: 'barcode' | 'camera') => void;
  onBarcodeScan: (e: React.FormEvent) => void;
  onCameraScan: (barcode: string) => void;
  onCameraOpen: () => void;
  isCameraScannerOpen: boolean;
  onCameraClose: () => void;
}

export function BarcodeScanner({
  barcodeInput,
  setBarcodeInput,
  inputMethod,
  setInputMethod,
  onBarcodeScan,
  onCameraScan,
  onCameraOpen,
  isCameraScannerOpen,
  onCameraClose,
}: BarcodeScannerProps) {
  const { t } = useTranslation();
  const barcodeInputRef = useRef<HTMLInputElement>(null);

  return (
    <>
      <Card className="pf-scanner-card">
        <CardHeader>
          <CardTitle className="pf-scanner-title">
            {t('sales.barcodeScanner')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ItemInputMethod 
            onMethodChange={setInputMethod}
            defaultMethod="barcode"
          />
          
          {inputMethod === 'barcode' && (
            <form onSubmit={onBarcodeScan} className="mt-4">
              <div
                className="pf-barcode-row"
              >
                <div className="pf-barcode-input">
                  <Input
                    ref={barcodeInputRef}
                    placeholder={t('sales.scanBarcode')}
                    value={barcodeInput}
                    onChange={(e) => setBarcodeInput(e.target.value)}
                    className="pf-barcode-field"
                    autoFocus
                  />
                </div>
                <Button
                  type="submit" 
                  variant="primary"
                  className="pf-barcode-submit pf-barcode-action"
                >
                  {t('common.add')}
                </Button>
              </div>
            </form>
          )}

          {inputMethod === 'camera' && (
            <div className="mt-4">
              <Button
                variant="primary"
                onClick={onCameraOpen}
                className="w-full pf-barcode-action"
              >
                <Camera className="w-4 h-4 mr-2" />
                فتح الكاميرا للمسح
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Camera Scanner Modal */}
      {isCameraScannerOpen && (
        <CameraScanner
          onScan={onCameraScan}
          onClose={onCameraClose}
          onError={(error) => console.error('Camera scan error:', error)}
        />
      )}
    </>
  );
}