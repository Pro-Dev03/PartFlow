import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Store, Save } from 'lucide-react';

export function StoreSettings() {
  const [storeSettings, setStoreSettings] = useState({
    storeName: 'PartFlow Store',
    lowStockThreshold: 10,
    allowDebt: true,
    maxDebtAmount: 5000,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Store className="w-5 h-5 text-cyan" />
          إعدادات المتجر
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <Input
          label="اسم المتجر"
          value={storeSettings.storeName}
          onChange={(e) => setStoreSettings({ ...storeSettings, storeName: e.target.value })}
        />
        <Input
          label="حد المخزون المنخفض"
          type="number"
          value={storeSettings.lowStockThreshold}
          onChange={(e) => setStoreSettings({ ...storeSettings, lowStockThreshold: Number(e.target.value) })}
        />
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">السماح بالديون</h4>
            <p className="text-small text-text-secondary">السماح للعملاء بالشراء على الحساب</p>
          </div>
          <Button
            variant={storeSettings.allowDebt ? 'primary' : 'secondary'}
            onClick={() => setStoreSettings({ ...storeSettings, allowDebt: !storeSettings.allowDebt })}
          >
            {storeSettings.allowDebt ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <Input
          label="الحد الأقصى للديون"
          type="number"
          value={storeSettings.maxDebtAmount}
          onChange={(e) => setStoreSettings({ ...storeSettings, maxDebtAmount: Number(e.target.value) })}
        />
        <Button variant="primary" className="gap-2">
          <Save className="w-4 h-4" />
          حفظ التغييرات
        </Button>
      </CardContent>
    </Card>
  );
}