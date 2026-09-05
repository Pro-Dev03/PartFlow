import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { DollarSign, Save } from 'lucide-react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { settingsApi } from '../../../services/api/endpoints';
import { toast } from 'sonner';
import { DEFAULT_PROFIT_MARGIN } from '../../../utils/pricing';

export function FinancialSettings() {
  const queryClient = useQueryClient();
  const [financialSettings, setFinancialSettings] = useState({
    currency: 'ILS',
    taxRate: 0,
    profitMargin: DEFAULT_PROFIT_MARGIN,
    discountEnabled: true,
    maxDiscount: 15,
  });
  const { data: taxSetting } = useQuery({
    queryKey: ['settings', 'tax_rate'],
    queryFn: () => settingsApi.getSetting('tax_rate'),
    retry: false,
  });
  const { data: discountSetting } = useQuery({
    queryKey: ['settings', 'max_discount_rate'],
    queryFn: () => settingsApi.getSetting('max_discount_rate'),
    retry: false,
  });
  const { data: marginSetting } = useQuery({
    queryKey: ['settings', 'default_profit_margin'],
    queryFn: () => settingsApi.getSetting('default_profit_margin'),
    retry: false,
  });
  const updateMarginMutation = useMutation({
    mutationFn: async () => {
      await settingsApi.updateSetting('tax_rate', String(financialSettings.taxRate));
      await settingsApi.updateSetting('max_discount_rate', String(financialSettings.maxDiscount));
      return settingsApi.updateSetting('default_profit_margin', String(financialSettings.profitMargin));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings', 'tax_rate'] });
      queryClient.invalidateQueries({ queryKey: ['settings', 'max_discount_rate'] });
      queryClient.invalidateQueries({ queryKey: ['settings', 'default_profit_margin'] });
      toast.success('تم حفظ إعدادات المالية بنجاح');
    },
    onError: () => toast.error('تعذر حفظ نسبة الربح المقترحة'),
  });

  useEffect(() => {
    const taxRate = Number(taxSetting?.data?.value);
    const maxDiscount = Number(discountSetting?.data?.value);
    const value = Number(marginSetting?.data?.value);
    setFinancialSettings((current) => ({
      ...current,
      ...(Number.isFinite(taxRate) && taxRate >= 0 && taxRate <= 100 ? { taxRate } : {}),
      ...(Number.isFinite(maxDiscount) && maxDiscount >= 0 && maxDiscount <= 100 ? { maxDiscount } : {}),
      ...(Number.isFinite(value) && value >= 0 && value < 100 ? { profitMargin: value } : {}),
    }));
  }, [discountSetting, marginSetting, taxSetting]);

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
          min="0"
          max="100"
          step="0.01"
          value={financialSettings.taxRate}
          onChange={(e) => setFinancialSettings({ ...financialSettings, taxRate: Number(e.target.value) })}
        />
        <Input
          label="نسبة الربح المقترحة عند إضافة منتج (%)"
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
            min="0"
            max="100"
            step="0.01"
            value={financialSettings.maxDiscount}
            onChange={(e) => setFinancialSettings({ ...financialSettings, maxDiscount: Number(e.target.value) })}
          />
        )}
        <Button
          variant="primary"
          className="gap-2"
          onClick={() => updateMarginMutation.mutate()}
          disabled={updateMarginMutation.isPending || financialSettings.taxRate < 0 || financialSettings.taxRate > 100 || financialSettings.maxDiscount < 0 || financialSettings.maxDiscount > 100 || financialSettings.profitMargin < 0 || financialSettings.profitMargin >= 100}
        >
          <Save className="w-4 h-4" />
          حفظ التغييرات
        </Button>
      </CardContent>
    </Card>
  );
}
