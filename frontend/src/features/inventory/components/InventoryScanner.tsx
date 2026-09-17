import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PackagePlus, Plus, ScanLine } from 'lucide-react';

interface InventoryScannerProps {
  barcodeInput: string;
  setBarcodeInput: (value: string) => void;
  onBarcodeScan: (e: React.FormEvent) => void;
  onAddInventory: () => void;
}

export function InventoryScanner({
  barcodeInput,
  setBarcodeInput,
  onBarcodeScan,
  onAddInventory,
}: InventoryScannerProps) {
  return (
    <>
      <Card className="pf-scanner-card">
        <CardHeader>
          <div className="pf-scanner-heading">
            <div className="pf-scanner-icon" aria-hidden="true">
              <PackagePlus className="h-5 w-5" />
            </div>
            <div>
              <CardTitle className="pf-scanner-title">إضافة مخزون</CardTitle>
              <p className="pf-scanner-description">سجّل بضاعة جديدة أو ابحث عن قطعة بالباركود</p>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={onBarcodeScan} className="pf-scanner-lookup">
            <div className="pf-scanner-label-row">
              <span className="pf-scanner-label">بحث سريع بالباركود</span>
              <ScanLine className="h-4 w-4" aria-hidden="true" />
            </div>
            <div className="pf-barcode-row">
              <div className="pf-barcode-input">
                <Input
                  aria-label="بحث بالباركود"
                  placeholder="امسح أو أدخل الباركود..."
                  value={barcodeInput}
                  onChange={(e) => setBarcodeInput(e.target.value)}
                  className="pf-barcode-field"
                />
              </div>
              <Button type="submit" variant="primary" className="pf-barcode-submit">
                بحث
              </Button>
            </div>
          </form>
          <Button variant="primary" onClick={onAddInventory} className="pf-scanner-add-button">
            <Plus className="h-4 w-4" />
            إضافة مخزون جديد
          </Button>
        </CardContent>
      </Card>

    </>
  );
}