import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Plus } from 'lucide-react';

interface InventoryScannerProps {
  barcodeInput: string;
  setBarcodeInput: (value: string) => void;
  onBarcodeScan: (e: React.FormEvent) => void;
  onManualAdd: () => void;
}

export function InventoryScanner({
  barcodeInput,
  setBarcodeInput,
  onBarcodeScan,
  onManualAdd,
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
          <form onSubmit={onBarcodeScan}>
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
          <Button variant="primary" onClick={onManualAdd} className="mt-4 w-full pf-barcode-action">
            <Plus className="w-4 h-4 mr-2" />
            إضافة قطعة يدوياً
          </Button>
        </CardContent>
      </Card>

    </>
  );
}