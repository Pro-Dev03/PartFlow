import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { Select } from '../../../design-system/components/select';
import { Palette, Save, Moon, Sun } from 'lucide-react';
import { useUIStore } from '../../../stores/uiStore';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { settingsApi } from '../../../services/api/endpoints';
import { toast } from 'sonner';

export function AppearanceSettings() {
  const queryClient = useQueryClient();
  const theme = useUIStore((state) => state.theme);
  const setTheme = useUIStore((state) => state.setTheme);
  const [appearanceSettings, setAppearanceSettings] = useState({
    language: 'ar',
    fontSize: 'medium',
    posProductsPerPage: 12,
    posProductViewMode: 'cards',
  });
  const { data: posProductsPerPageSetting, isLoading: isProductsPerPageLoading } = useQuery({
    queryKey: ['settings', 'pos_products_per_page'],
    queryFn: () => settingsApi.getSetting('pos_products_per_page'),
    retry: false,
  });
  const { data: posProductViewModeSetting, isLoading: isProductViewModeLoading } = useQuery({
    queryKey: ['settings', 'pos_product_view_mode'],
    queryFn: () => settingsApi.getSetting('pos_product_view_mode'),
    retry: false,
  });
  const saveAppearanceMutation = useMutation({
    mutationFn: async () => {
      await settingsApi.updateSetting('pos_products_per_page', String(appearanceSettings.posProductsPerPage));
      return settingsApi.updateSetting('pos_product_view_mode', appearanceSettings.posProductViewMode);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings', 'pos_products_per_page'] });
      queryClient.invalidateQueries({ queryKey: ['settings', 'pos_product_view_mode'] });
      toast.success('تم حفظ إعدادات العرض بنجاح');
    },
    onError: () => toast.error('تعذر حفظ إعدادات العرض'),
  });

  useEffect(() => {
    const value = Number(posProductsPerPageSetting?.data?.value);
    if (Number.isInteger(value) && value >= 4 && value <= 48) {
      setAppearanceSettings((settings) => ({ ...settings, posProductsPerPage: value }));
    }
  }, [posProductsPerPageSetting]);

  useEffect(() => {
    const value = posProductViewModeSetting?.data?.value;
    if (value === 'cards' || value === 'list') {
      setAppearanceSettings((settings) => ({ ...settings, posProductViewMode: value }));
    }
  }, [posProductViewModeSetting]);

  return (
    <Card className="overflow-hidden">
      <CardHeader className="relative mb-0 flex-col items-stretch gap-4 border-b border-border/70 bg-gradient-to-l from-violet-500/[0.08] via-surface to-surface px-5 py-5 sm:px-6">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3">
            <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-violet-600 text-white shadow-[0_8px_18px_rgba(124,58,237,0.2)]">
              <Palette className="h-5 w-5" />
            </span>
            <div>
              <CardTitle className="text-base sm:text-lg">إعدادات المظهر</CardTitle>
              <p className="mt-1 text-xs leading-5 text-text-secondary sm:text-sm">خصّص طريقة عرض PartFlow بما يناسب أسلوب عملك.</p>
            </div>
          </div>
          <span className="inline-flex shrink-0 items-center gap-1.5 rounded-full bg-violet-500/10 px-2.5 py-1 text-xs font-semibold text-violet-700">
            <span className="h-1.5 w-1.5 rounded-full bg-violet-500" />
            تخصيص
          </span>
        </div>
      </CardHeader>
      <CardContent className="space-y-5 px-5 py-5 sm:px-6">
        <div className="rounded-xl border border-border bg-surface/60 p-4 sm:p-5">
          <div className="mb-4 flex items-center gap-2">
            <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-violet-500/10 text-xs font-bold text-violet-700">1</span>
            <div>
              <h4 className="font-semibold text-text-primary">نمط الواجهة</h4>
              <p className="text-xs text-text-secondary">اختر المظهر الذي يناسب الإضاءة ووقت العمل.</p>
            </div>
          </div>
          <div className="flex flex-col gap-4 rounded-xl border border-border/70 bg-surface p-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h4 className="font-semibold text-text-primary">الوضع الليلي</h4>
            <p className="mt-1 text-xs text-text-secondary">تغيير مظهر التطبيق</p>
          </div>
          <div className="flex gap-2">
            <Button
              variant={theme === 'dark' ? 'primary' : 'secondary'}
              onClick={() => setTheme('dark')}
              className="gap-2"
            >
              <Moon className="w-4 h-4" />
              داكن
            </Button>
            <Button
              variant={theme === 'light' ? 'primary' : 'secondary'}
              onClick={() => setTheme('light')}
              className="gap-2"
            >
              <Sun className="w-4 h-4" />
              فاتح
            </Button>
          </div>
        </div>
        </div>
        <div className="grid gap-4 rounded-xl border border-border bg-surface/60 p-4 sm:grid-cols-2 sm:p-5">
          <div>
          <Select
            label="اللغة"
            value={appearanceSettings.language}
            onChange={(e) => setAppearanceSettings({ ...appearanceSettings, language: e.target.value })}
            options={[
              { value: 'ar', label: 'العربية' },
              { value: 'en', label: 'English' },
            ]}
            emptyMessage="لا توجد لغات"
          />
          </div>
          <div>
          <Select
            label="حجم الخط"
            value={appearanceSettings.fontSize}
            onChange={(e) => setAppearanceSettings({ ...appearanceSettings, fontSize: e.target.value })}
            options={[
              { value: 'small', label: 'صغير' },
              { value: 'medium', label: 'متوسط' },
              { value: 'large', label: 'كبير' },
            ]}
            emptyMessage="لا توجد أحجام"
          />
          </div>
        </div>
        <div className="rounded-xl border border-border bg-surface/60 p-4 sm:p-5">
          <div className="mb-4 flex items-center gap-2">
            <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-violet-500/10 text-xs font-bold text-violet-700">2</span>
            <div>
              <h4 className="font-semibold text-text-primary">تخصيص نقطة البيع</h4>
              <p className="text-xs text-text-secondary">حدد كثافة المنتجات وطريقة عرضها في شاشة البيع.</p>
            </div>
          </div>
        <Input
          label="عدد منتجات نقطة البيع في الصفحة"
          type="number"
          min="4"
          max="48"
          step="1"
          value={appearanceSettings.posProductsPerPage}
          onChange={(event) => setAppearanceSettings({
            ...appearanceSettings,
            posProductsPerPage: Math.min(48, Math.max(4, Number(event.target.value) || 12)),
          })}
        />
        <p className="text-xs text-text-secondary">اختر عدد البطاقات الظاهرة في صفحة نقطة البيع، من 4 إلى 48.</p>
        <Select
          label="طريقة عرض منتجات نقطة البيع"
          value={appearanceSettings.posProductViewMode}
          onChange={(event) => setAppearanceSettings({ ...appearanceSettings, posProductViewMode: event.target.value })}
          options={[
            { value: 'cards', label: 'بطاقات' },
            { value: 'list', label: 'قائمة مضغوطة' },
          ]}
        />
        </div>
        <div className="flex justify-end border-t border-border/70 pt-5">
        <Button
          variant="primary"
          className="gap-2"
          onClick={() => saveAppearanceMutation.mutate()}
          disabled={saveAppearanceMutation.isPending || isProductsPerPageLoading || isProductViewModeLoading}
        >
          <Save className="w-4 h-4" />
          حفظ التغييرات
        </Button>
        </div>
      </CardContent>
    </Card>
  );
}
