import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { ItemInputMethod, type ItemInputMethodType } from '../../../components/ui/item-input-method';
import { CameraScanner } from '../../../components/ui/camera-scanner';
import { Camera, Plus } from 'lucide-react';

interface InventoryScannerProps {
  inputMethod: ItemInputMethodType;
  setInputMethod: (method: ItemInputMethodType) => void;
  barcodeInput: string;
  setBarcodeInput: (value: string) => void;
  onBarcodeScan: (e: React.FormEvent) => void;
  onCameraScan: (barcode: string) => void;
  onManualAdd: () => void;
  isCameraScannerOpen: boolean;
  onCameraOpen: () => void;
  onCameraClose: () => void;
}

export function InventoryScanner({
  inputMethod,
  setInputMethod,
  barcodeInput,
  setBarcodeInput,
  onBarcodeScan,
  onCameraScan,
  onManualAdd,
  isCameraScannerOpen,
  onCameraOpen,
  onCameraClose,
}: InventoryScannerProps) {
  return (
    <>
      <Card className="pf-scanner-card">
        <CardHeader>
          <CardTitle className="pf-scanner-title">
            إضافة قطع للمخزون
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
                    placeholder="مسح أو أدخل الباركود..."
                    value={barcodeInput}
                    onChange={(e) => setBarcodeInput(e.target.value)}
                    className="pf-barcode-field"
                  />
                </div>
                <Button
                  type="submit" 
                  variant="primary"
                  className="pf-barcode-submit"
                >
                  بحث
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

          {inputMethod === 'manual' && (
            <div className="mt-4">
              <Button
                variant="primary"
                onClick={onManualAdd}
                className="w-full pf-barcode-action"
              >
                <Plus className="w-4 h-4 mr-2" />
                إضافة قطعة يدوياً
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