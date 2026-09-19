import { useEffect, useState } from 'react';
import { Globe2, Save, ShieldAlert } from 'lucide-react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '../../../design-system/components/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Select } from '../../../design-system/components/select';
import { settingsApi } from '../../../services/api/endpoints';
import { RegionalProfile } from '../../../types/regional';
import { getDeviceTimezone, setRegionalProfile } from '../../../utils/store-time';
import { toast } from 'sonner';

interface RegionalSettingsProps {
  canManageRegionalSettings?: boolean;
}

export function RegionalSettings({ canManageRegionalSettings = false }: RegionalSettingsProps) {
  const queryClient = useQueryClient();
  const [selectedCountry, setSelectedCountry] = useState('IL');
  const [timezone, setTimezone] = useState('');
  const { data, isLoading } = useQuery({
    queryKey: ['settings', 'regional'],
    queryFn: () => settingsApi.getRegionalSettings(),
  });
  const profile = data?.data?.profile as RegionalProfile | undefined;
  const countries = (data?.data?.countries || []) as RegionalProfile[];

  useEffect(() => {
    if (profile) {
      setSelectedCountry(profile.country_code);
      setTimezone(profile.timezone);
      setRegionalProfile(profile);
    }
  }, [profile]);

  const updateMutation = useMutation({
    mutationFn: () => settingsApi.updateRegionalSettings({ country_code: selectedCountry, timezone }),
    onSuccess: (response) => {
      const nextProfile = response.data?.profile as RegionalProfile | undefined;
      if (nextProfile) setRegionalProfile(nextProfile);
      void queryClient.invalidateQueries({ queryKey: ['settings', 'regional'] });
          toast.success('تم حفظ المنطقة الزمنية للمتجر');
    },
    onError: () => toast.error('تعذر حفظ إعدادات الدولة والمنطقة الزمنية'),
  });

  return (
    <Card className="overflow-hidden">
      <CardHeader className="relative mb-0 flex-col items-stretch gap-4 border-b border-border/70 bg-gradient-to-l from-emerald-500/[0.08] via-surface to-surface px-5 py-5 sm:px-6">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3">
            <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-600 text-white shadow-[0_8px_18px_rgba(16,185,129,0.2)]">
              <Globe2 className="h-5 w-5" />
            </span>
            <div>
              <CardTitle className="text-base sm:text-lg">الدولة والمنطقة</CardTitle>
              <p className="mt-1 text-xs leading-5 text-text-secondary sm:text-sm">تحكم في توقيت المتجر وطريقة عرض التواريخ.</p>
            </div>
          </div>
          <span className="inline-flex shrink-0 items-center gap-1.5 rounded-full bg-emerald-500/10 px-2.5 py-1 text-xs font-semibold text-emerald-700">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
            مضبوط
          </span>
        </div>
      </CardHeader>
      <CardContent className="space-y-5 px-5 py-5 sm:px-6">
        <div className="rounded-xl border border-border bg-surface/60 p-4 sm:p-5">
          <div className="mb-4">
            <h4 className="font-semibold text-text-primary">إعدادات المتجر الإقليمية</h4>
            <p className="mt-1 text-xs leading-5 text-text-secondary">تُستخدم هذه القيم في الفواتير والتقارير والديون.</p>
          </div>
          <div className="flex flex-col gap-2 rounded-xl border border-emerald-500/20 bg-emerald-500/[0.04] p-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="text-xs text-text-secondary">المنطقة الزمنية الرسمية</p>
              <strong className="mt-1 block text-sm font-semibold text-text-primary" dir="ltr">{timezone || getDeviceTimezone()}</strong>
            </div>
            <Globe2 className="h-5 w-5 shrink-0 text-emerald-600" />
          </div>
          {profile && (
            <div className="mt-3 grid gap-2 sm:grid-cols-2">
              <div className="flex items-center justify-between gap-3 rounded-lg border border-border/60 bg-surface p-3"><span className="text-xs text-text-secondary">المصدر</span><strong className="text-xs text-text-primary">Store Timezone</strong></div>
              <div className="flex items-center justify-between gap-3 rounded-lg border border-border/60 bg-surface p-3"><span className="text-xs text-text-secondary">تنسيق التاريخ</span><strong className="text-xs text-text-primary">يوم/شهر/سنة</strong></div>
              <div className="flex items-center justify-between gap-3 rounded-lg border border-border/60 bg-surface p-3"><span className="text-xs text-text-secondary">نظام الوقت</span><strong className="text-xs text-text-primary">{profile.time_format}</strong></div>
              <div className="flex items-center justify-between gap-3 rounded-lg border border-border/60 bg-surface p-3"><span className="text-xs text-text-secondary">العملة</span><strong className="text-xs text-text-primary">{profile.currency}</strong></div>
            </div>
          )}
          <p className="mt-4 text-xs leading-5 text-text-secondary">
            تم اكتشافها تلقائيًا من الجهاز عند إعداد المتجر، وتبقى ثابتة عند تغيير الجهاز.
          </p>
        </div>
        {canManageRegionalSettings ? (
          <details className="rounded-lg border border-border p-3">
            <summary className="flex cursor-pointer items-center gap-2 text-sm font-medium text-text-primary">
              <ShieldAlert className="h-4 w-4 text-amber-500" /> تغيير إداري متقدم
            </summary>
            <div className="mt-3 space-y-3">
              <p className="text-xs text-text-secondary">سيؤثر تغيير المنطقة الزمنية على اليوم والتقارير والديون والدفعات والمرتجعات والربح والتدفق النقدي.</p>
              <Select
                label="الدولة والمنطقة الزمنية"
                value={selectedCountry}
                disabled={isLoading || updateMutation.isPending}
                onChange={(event) => {
                  const nextCountry = countries.find((country) => country.country_code === event.target.value);
                  setSelectedCountry(event.target.value);
                  if (nextCountry) setTimezone(nextCountry.timezone);
                }}
                options={countries.map((country) => ({
                  value: country.country_code,
                  label: `${country.country_name} (${country.timezone})`,
                }))}
              />
              <Button variant="primary" className="gap-2" onClick={() => {
                if (window.confirm('تغيير المنطقة الزمنية سيؤثر على الحسابات والتقارير التاريخية. هل تريد المتابعة؟')) updateMutation.mutate();
              }} disabled={isLoading || updateMutation.isPending}>
                <Save className="w-4 h-4" />
                {updateMutation.isPending ? 'جارٍ الحفظ...' : 'تأكيد تغيير المنطقة الزمنية'}
              </Button>
            </div>
          </details>
        ) : (
          <div className="flex items-start gap-2 rounded-lg border border-amber-200/70 bg-amber-50/70 p-3 text-xs leading-5 text-amber-950">
            <ShieldAlert className="mt-0.5 h-4 w-4 shrink-0 text-amber-700" />
            <span>تغيير المنطقة الزمنية متاح لحساب المدير فقط.</span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
