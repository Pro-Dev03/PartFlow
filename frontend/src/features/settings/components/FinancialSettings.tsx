import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { DollarSign, Save } from 'lucide-react';

export function FinancialSettings() {
  const [financialSettings, setFinancialSettings] = useState({
    currency: 'ILS',
    taxRate: 17,
    profitMargin: 30,
    discountEnabled: true,
    maxDiscount: 15,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <DollarSign className="w-5 h-5 text-cyan" />
          إعدادات المالية
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <div>
          <label className="block text-sm font-medium text-text-primary mb-2">العملة</label>
          <select
            value={financialSettings.currency}
            onChange={(e) => setFinancialSettings({ ...financialSettings, currency: e.target.value })}
            className="w-full p-3 border border-border rounded-lg bg-surface text-text-primary"
          >
            <option value="ILS">شيكل إسرائيلي (₪)</option>
            <option value="USD">دولار أمريكي ($)</option>
            <option value="EUR">يورو (€)</option>
          </select>
        </div>
        <Input
          label="نسبة الضريبة (%)"
          type="number"
          value={financialSettings.taxRate}
          onChange={(e) => setFinancialSettings({ ...financialSettings, taxRate: Number(e.target.value) })}
        />
        <Input
          label="هامش الربح الافتراضي (%)"
          type="number"
          value={financialSettings.profitMargin}
          onChange={(e) => setFinancialSettings({ ...financialSettings, profitMargin: Number(e.target.value) })}
        />
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">السماح بالخصومات</h4>
            <p className="text-small text-text-secondary">تفعيل نظام الخصومات</p>
          </div>
          <Button
            variant={financialSettings.discountEnabled ? 'primary' : 'secondary'}
            onClick={() => setFinancialSettings({ ...financialSettings, discountEnabled: !financialSettings.discountEnabled })}
          >
            {financialSettings.discountEnabled ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        {financialSettings.discountEnabled && (
          <Input
            label="الحد الأقصى للخصم (%)"
            type="number"
            value={financialSettings.maxDiscount}
            onChange={(e) => setFinancialSettings({ ...financialSettings, maxDiscount: Number(e.target.value) })}
          />
        )}
        <Button variant="primary" className="gap-2">
          <Save className="w-4 h-4" />
          حفظ التغييرات
        </Button>
      </CardContent>
    </Card>
  );
}
