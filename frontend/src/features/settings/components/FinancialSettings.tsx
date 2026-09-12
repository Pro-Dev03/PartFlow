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
  const { data: currencySetting } = useQuery({
    queryKey: ['settings', 'currency'],
    queryFn: () => settingsApi.getSetting('currency'),
    retry: false,
  });
  const { data: discountsEnabledSetting } = useQuery({
    queryKey: ['settings', 'discounts_enabled'],
    queryFn: () => settingsApi.getSetting('discounts_enabled'),
    retry: false,
  });
  const updateSetting = async (key: string, value: string, label: string) => {
    try {
      return await settingsApi.updateSetting(key, value);
    } catch (error) {
      const apiError = error as { response?: { error?: { message?: string } }; message?: string };
      const reason = apiError.response?.error?.message || apiError.message || 'خطأ غير معروف';
      throw new Error(`${label}: ${reason}`);
    }
  };
  const updateMarginMutation = useMutation({
    mutationFn: async () => {
      await updateSetting('currency', financialSettings.currency, 'العملة');
      await updateSetting('discounts_enabled', String(financialSettings.discountEnabled), 'السماح بالخصومات');
      await updateSetting('tax_rate', String(financialSettings.taxRate), 'نسبة الضريبة');
      await updateSetting('max_discount_rate', String(financialSettings.maxDiscount), 'الحد الأقصى للخصم');
      return updateSetting('default_profit_margin', String(financialSettings.profitMargin), 'نسبة الربح المقترحة');
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings', 'tax_rate'] });
      queryClient.invalidateQueries({ queryKey: ['settings', 'max_discount_rate'] });
      queryClient.invalidateQueries({ queryKey: ['settings', 'default_profit_margin'] });
      queryClient.invalidateQueries({ queryKey: ['settings', 'currency'] });
      queryClient.invalidateQueries({ queryKey: ['settings', 'discounts_enabled'] });
      localStorage.setItem('partflow-currency', financialSettings.currency);
      toast.success('تم حفظ إعدادات المالية بنجاح');
    },
    onError: (error) => toast.error(error instanceof Error ? error.message : 'تعذر حفظ الإعدادات المالية'),
  });

  useEffect(() => {
    const taxRate = Number(taxSetting?.data?.value);
    const maxDiscount = Number(discountSetting?.data?.value);
    const value = Number(marginSetting?.data?.value);
    const currency = currencySetting?.data?.value;
    const discountsEnabled = discountsEnabledSetting?.data?.value;
    setFinancialSettings((current) => ({
      ...current,
      ...(currency === 'ILS' || currency === 'USD' || currency === 'EUR' ? { currency } : {}),
      ...(discountsEnabled === 'true' || discountsEnabled === 'false' ? { discountEnabled: discountsEnabled === 'true' } : {}),
      ...(Number.isFinite(taxRate) && taxRate >= 0 && taxRate <= 100 ? { taxRate } : {}),
      ...(Number.isFinite(maxDiscount) && maxDiscount >= 0 && maxDiscount <= 100 ? { maxDiscount } : {}),
      ...(Number.isFinite(value) && value >= 0 && value < 100 ? { profitMargin: value } : {}),
    }));
  }, [currencySetting, discountsEnabledSetting, discountSetting, marginSetting, taxSetting]);

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
          <label className="mb-2 block text-sm font-medium text-text-primary">العملة الأساسية</label>
          <select
            value={financialSettings.currency}
            onChange={(event) => setFinancialSettings({ ...financialSettings, currency: event.target.value })}
            className="w-full rounded-lg border border-border bg-surface p-3 text-text-primary"
          >
            <option value="ILS">شيكل إسرائيلي (₪)</option>
            <option value="USD">دولار أمريكي ($)</option>
            <option value="EUR">يورو (€)</option>
          </select>
          <p className="mt-2 text-xs text-text-secondary">الشيكل هو العملة الافتراضية، ويمكن اختيار عملة أخرى عند الحاجة.</p>
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
          min="0"
          max="99.99"
          step="0.01"
          value={financialSettings.profitMargin}
          onChange={(e) => setFinancialSettings({ ...financialSettings, profitMargin: Number(e.target.value) })}
        />
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">السماح بالخصومات</h4>
            <p className="text-small text-text-secondary">السماح للموظفين بتطبيق خصم على الفاتورة ضمن الحد المحدد أدناه.</p>
          </div>
          <Button
            variant={financialSettings.discountEnabled ? 'primary' : 'secondary'}
            onClick={() => setFinancialSettings({ ...financialSettings, discountEnabled: !financialSettings.discountEnabled })}
          >
            {financialSettings.discountEnabled ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        {financialSettings.discountEnabled && (
          <div>
            <Input
              label="الحد الأقصى للخصم (%)"
              type="number"
              min="0"
              max="100"
              step="0.01"
              value={financialSettings.maxDiscount}
              onChange={(e) => setFinancialSettings({ ...financialSettings, maxDiscount: Number(e.target.value) })}
            />
            <p className="mt-2 text-xs text-text-secondary">
              اكتب هنا أكبر خصم تريد السماح به للزبون. مثال: إذا كتبت 10% وكانت الفاتورة 1,000 ₪، يمكن للموظف خصم 100 ₪ فقط، ويدفع الزبون 900 ₪.
            </p>
          </div>
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
