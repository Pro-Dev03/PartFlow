import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Apple, Check, CreditCard, KeyRound, PlugZap, QrCode, Save, ShieldCheck, Smartphone } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { Select } from '../../../design-system/components/select';
import { settingsApi } from '../../../services/api/endpoints';
import { toast } from 'sonner';

type PaymentMethod = 'card' | 'apple_pay' | 'google_pay' | 'qr_code';

const paymentOptions: Array<{ id: PaymentMethod; label: string; description: string; icon: typeof CreditCard }> = [
  { id: 'card', label: 'بطاقة بنكية', description: 'Visa و Mastercard وغيرها', icon: CreditCard },
  { id: 'apple_pay', label: 'Apple Pay', description: 'لعملاء أجهزة Apple', icon: Apple },
  { id: 'google_pay', label: 'Google Pay', description: 'لعملاء أجهزة Android', icon: Smartphone },
  { id: 'qr_code', label: 'الدفع عبر QR', description: 'رمز دفع من مزودك', icon: QrCode },
];

const settingKeys = [
  'electronic_payments_enabled',
  'payment_provider',
  'payment_environment',
  'payment_public_key',
  'payment_secret_key',
  'payment_merchant_id',
  'payment_terminal_id',
  'payment_webhook_url',
  'payment_webhook_secret',
  'payment_methods',
] as const;

export function ElectronicPaymentSettings() {
  const queryClient = useQueryClient();
  const providerCatalog = useQuery({ queryKey: ['payment-providers'], queryFn: () => settingsApi.listPaymentProviders(), retry: false });
  const enabledSetting = useQuery({ queryKey: ['settings', 'electronic_payments_enabled'], queryFn: () => settingsApi.getSetting('electronic_payments_enabled'), retry: false });
  const providerSetting = useQuery({ queryKey: ['settings', 'payment_provider'], queryFn: () => settingsApi.getSetting('payment_provider'), retry: false });
  const environmentSetting = useQuery({ queryKey: ['settings', 'payment_environment'], queryFn: () => settingsApi.getSetting('payment_environment'), retry: false });
  const publicKeySetting = useQuery({ queryKey: ['settings', 'payment_public_key'], queryFn: () => settingsApi.getSetting('payment_public_key'), retry: false });
  const secretKeySetting = useQuery({ queryKey: ['settings', 'payment_secret_key'], queryFn: () => settingsApi.getSetting('payment_secret_key'), retry: false });
  const merchantIdSetting = useQuery({ queryKey: ['settings', 'payment_merchant_id'], queryFn: () => settingsApi.getSetting('payment_merchant_id'), retry: false });
  const terminalIdSetting = useQuery({ queryKey: ['settings', 'payment_terminal_id'], queryFn: () => settingsApi.getSetting('payment_terminal_id'), retry: false });
  const webhookUrlSetting = useQuery({ queryKey: ['settings', 'payment_webhook_url'], queryFn: () => settingsApi.getSetting('payment_webhook_url'), retry: false });
  const webhookSecretSetting = useQuery({ queryKey: ['settings', 'payment_webhook_secret'], queryFn: () => settingsApi.getSetting('payment_webhook_secret'), retry: false });
  const methodsSetting = useQuery({ queryKey: ['settings', 'payment_methods'], queryFn: () => settingsApi.getSetting('payment_methods'), retry: false });
  const settings = {
    electronic_payments_enabled: enabledSetting.data?.data?.value ?? '',
    payment_provider: providerSetting.data?.data?.value ?? '',
    payment_environment: environmentSetting.data?.data?.value ?? '',
    payment_public_key: publicKeySetting.data?.data?.value ?? '',
    payment_secret_key: secretKeySetting.data?.data?.value ?? '',
    payment_merchant_id: merchantIdSetting.data?.data?.value ?? '',
    payment_terminal_id: terminalIdSetting.data?.data?.value ?? '',
    payment_webhook_url: webhookUrlSetting.data?.data?.value ?? '',
    payment_webhook_secret: webhookSecretSetting.data?.data?.value ?? '',
    payment_methods: methodsSetting.data?.data?.value ?? '',
  };
  const providerOptions = providerCatalog.data?.data ?? [
    { name: 'cardcom', display_name: 'Cardcom', implemented: true },
  ];
  const [form, setForm] = useState({
    enabled: false,
    provider: 'manual',
    environment: 'test',
    publicKey: '',
    secretKey: '',
    merchantId: '',
    terminalId: '',
    webhookUrl: '',
    webhookSecret: '',
    methods: ['card'] as PaymentMethod[],
  });

  useEffect(() => {
    let methods: PaymentMethod[] = ['card'];
    try {
      const parsed = JSON.parse(String(settings.payment_methods || '[]'));
      if (Array.isArray(parsed)) methods = parsed.filter((value): value is PaymentMethod => ['card', 'apple_pay', 'google_pay', 'qr_code'].includes(value));
    } catch {
      methods = ['card'];
    }
    setForm((current) => ({
      ...current,
      enabled: settings.electronic_payments_enabled === 'true',
      provider: settings.payment_provider || 'manual',
      environment: settings.payment_environment || 'test',
      publicKey: settings.payment_public_key || '',
      secretKey: settings.payment_secret_key || '',
      merchantId: settings.payment_merchant_id || '',
      terminalId: settings.payment_terminal_id || '',
      webhookUrl: settings.payment_webhook_url || '',
      webhookSecret: settings.payment_webhook_secret || '',
      methods: methods.length > 0 ? methods : ['card'],
    }));
  }, [settings.electronic_payments_enabled, settings.payment_environment, settings.payment_methods, settings.payment_provider, settings.payment_public_key, settings.payment_secret_key, settings.payment_merchant_id, settings.payment_terminal_id, settings.payment_webhook_url, settings.payment_webhook_secret]);

  const saveMutation = useMutation({
    mutationFn: async () => {
      await settingsApi.updateSetting('electronic_payments_enabled', String(form.enabled));
      await settingsApi.updateSetting('payment_provider', form.provider);
      await settingsApi.updateSetting('payment_environment', form.environment);
      await settingsApi.updateSetting('payment_public_key', form.publicKey.trim());
      await settingsApi.updateSetting('payment_merchant_id', form.merchantId.trim());
      await settingsApi.updateSetting('payment_terminal_id', form.terminalId.trim());
      await settingsApi.updateSetting('payment_webhook_url', form.webhookUrl.trim());
      if (form.secretKey.trim() && form.secretKey !== '********') {
        await settingsApi.updateSetting('payment_secret_key', form.secretKey.trim());
      }
      if (form.webhookSecret.trim() && form.webhookSecret !== '********') {
        await settingsApi.updateSetting('payment_webhook_secret', form.webhookSecret.trim());
      }
      return settingsApi.updateSetting('payment_methods', JSON.stringify(form.methods));
    },
    onSuccess: () => {
      settingKeys.forEach((key) => void queryClient.invalidateQueries({ queryKey: ['settings', key] }));
      setForm((current) => ({ ...current, secretKey: '********', webhookSecret: '********' }));
      toast.success('تم حفظ إعدادات الدفع الإلكتروني');
    },
    onError: () => toast.error('تعذر حفظ إعدادات الدفع الإلكتروني'),
  });

  const connectionMutation = useMutation({
    mutationFn: () => settingsApi.testPaymentConnection(),
    onSuccess: () => toast.success('تم الاتصال بمزود الدفع بنجاح'),
    onError: () => toast.error('تعذر الاتصال. تحقق من بيانات Cardcom وبيئة الحساب.'),
  });

  const toggleMethod = (method: PaymentMethod) => {
    setForm((current) => ({
      ...current,
      methods: current.methods.includes(method)
        ? current.methods.filter((item) => item !== method)
        : [...current.methods, method],
    }));
  };

  return (
    <Card className="overflow-hidden">
      <CardHeader className="relative mb-0 flex-col items-stretch gap-5 border-b border-border/70 bg-gradient-to-l from-primary/[0.08] via-surface to-surface px-5 py-5 sm:px-6">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3">
            <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary text-white shadow-[0_8px_18px_rgba(37,99,235,0.2)]">
              <CreditCard className="h-5 w-5" />
            </span>
            <div>
              <CardTitle className="text-base sm:text-lg">إعدادات الدفع الإلكتروني</CardTitle>
              <p className="mt-1 text-xs leading-5 text-text-secondary sm:text-sm">اربط مزود الدفع وحدد الطرق المتاحة لعملائك.</p>
            </div>
          </div>
          <span className={`inline-flex shrink-0 items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold ${form.enabled ? 'bg-emerald-500/10 text-emerald-700' : 'bg-surface-elevated text-text-secondary'}`}>
            <span className={`h-1.5 w-1.5 rounded-full ${form.enabled ? 'bg-emerald-500' : 'bg-slate-400'}`} />
            {form.enabled ? 'جاهز للاستقبال' : 'غير مفعّل'}
          </span>
        </div>
        <div className="grid gap-2 sm:grid-cols-3">
          {['التفعيل', 'مزود الدفع', 'طرق الدفع'].map((step, index) => (
            <div key={step} className="flex items-center gap-2 text-xs text-text-secondary">
              <span className={`flex h-6 w-6 items-center justify-center rounded-full text-[11px] font-bold ${index === 0 && form.enabled ? 'bg-primary text-white' : 'bg-surface-elevated text-text-secondary'}`}>{index + 1}</span>
              <span>{step}</span>
              {index < 2 && <span className="hidden h-px flex-1 bg-border sm:block" />}
            </div>
          ))}
        </div>
      </CardHeader>
      <CardContent className="space-y-5 px-5 py-5 sm:px-6" dir="rtl">
        <div className={`flex flex-col gap-4 rounded-xl border p-4 transition-colors sm:flex-row sm:items-center sm:justify-between ${form.enabled ? 'border-emerald-500/30 bg-emerald-500/[0.04]' : 'border-border bg-surface-elevated'}`}>
          <div className="flex items-start gap-3">
            <span className={`mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg ${form.enabled ? 'bg-emerald-500/10 text-emerald-600' : 'bg-surface text-text-secondary'}`}>
              <ShieldCheck className="h-4 w-4" />
            </span>
            <div>
              <p className="mb-1 text-xs font-semibold text-primary">الخطوة 1</p>
              <h4 className="font-semibold text-text-primary">هل تريد استقبال الدفع الإلكتروني؟</h4>
              <p className="mt-1 text-sm leading-6 text-text-secondary">فعّل هذه الخاصية بعد تجهيز بيانات مزود الدفع واختيار الطرق المناسبة.</p>
            </div>
          </div>
          <Button className="sm:min-w-[112px]" variant={form.enabled ? 'success' : 'secondary'} onClick={() => setForm({ ...form, enabled: !form.enabled })} aria-pressed={form.enabled}>
            {form.enabled ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>

        <div className="rounded-xl border border-border bg-surface/60 p-4 sm:p-5">
          <div className="mb-4 flex items-center gap-2">
            <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-primary/10 text-xs font-bold text-primary">2</span>
            <div>
              <h4 className="font-semibold text-text-primary">اختيار مزود الدفع</h4>
              <p className="text-xs text-text-secondary">حدد بيئة الاتصال وطريقة معالجة العمليات.</p>
            </div>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
          <div>
            <Select
              label="مزود الدفع"
              value={form.provider}
              onChange={(event) => setForm({ ...form, provider: event.target.value })}
              helperText="اختر “تسجيل يدوي” إذا كان الموظف يؤكد الدفع من جهاز خارجي."
            >
              <option value="manual">تسجيل يدوي فقط</option>
              {providerOptions.map((provider: { name: string; display_name: string; implemented: boolean }) => (
                <option key={provider.name} value={provider.name} disabled={!provider.implemented}>
                  {provider.display_name}{provider.implemented ? '' : ' (غير متاح بعد)'}
                </option>
              ))}
            </Select>
          </div>
          <div>
            <Select
              label="بيئة الحساب"
              value={form.environment}
              onChange={(event) => setForm({ ...form, environment: event.target.value })}
              helperText="ابدأ بالاختبار، ثم انتقل إلى الإنتاج بعد نجاح التجربة."
              options={[
                { value: 'test', label: 'اختبار' },
                { value: 'live', label: 'إنتاج' },
              ]}
            />
          </div>
        </div>
        </div>

        {form.provider !== 'manual' && (
          <div className="rounded-xl border border-border bg-surface/60 p-4 sm:p-5">
            <div className="mb-4 flex items-center gap-2">
              <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-primary/10 text-xs font-bold text-primary">3</span>
              <div>
                <h4 className="font-semibold text-text-primary">بيانات الحساب</h4>
                <p className="text-xs text-text-secondary">تُحفظ المفاتيح السرية بشكل مخفي ولا تظهر بعد الحفظ.</p>
              </div>
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <Input label="المفتاح العام أو رقم الحساب" value={form.publicKey} onChange={(event) => setForm({ ...form, publicKey: event.target.value })} placeholder="Public key / Client ID" />
              <Input label="المفتاح السري" type="password" value={form.secretKey} onChange={(event) => setForm({ ...form, secretKey: event.target.value })} placeholder="يُحفظ بشكل مخفي" />
            </div>
            {form.provider === 'cardcom' && (
              <div className="mt-4 grid gap-4 md:grid-cols-2">
                <Input label="Merchant ID / رقم التاجر" value={form.merchantId} onChange={(event) => setForm({ ...form, merchantId: event.target.value })} />
                <Input label="Terminal ID / رقم الجهاز" value={form.terminalId} onChange={(event) => setForm({ ...form, terminalId: event.target.value })} />
                <Input label="Webhook URL" value={form.webhookUrl} onChange={(event) => setForm({ ...form, webhookUrl: event.target.value })} placeholder="https://example.com/api/v1/payment-webhooks/cardcom" />
                <Input label="Webhook Secret" type="password" value={form.webhookSecret} onChange={(event) => setForm({ ...form, webhookSecret: event.target.value })} placeholder="يُحفظ بشكل مخفي" />
              </div>
            )}
          </div>
        )}

        <div className="rounded-xl border border-border bg-surface/60 p-4 sm:p-5">
          <div className="mb-4 flex items-start gap-2">
            <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary"><KeyRound className="h-3.5 w-3.5" /></span>
            <div>
              <h4 className="font-semibold text-text-primary">طرق الدفع المتاحة</h4>
              <p className="mt-1 text-xs text-text-secondary">يمكن اختيار أكثر من طريقة، بشرط أن يدعمها مزود الدفع.</p>
            </div>
          </div>
          <div className="grid gap-3 sm:grid-cols-2">
            {paymentOptions.map(({ id, label, description, icon: Icon }) => (
              <label key={id} className={`group flex cursor-pointer items-center gap-3 rounded-xl border p-3.5 transition-all ${form.methods.includes(id) ? 'border-primary/45 bg-primary/[0.06] shadow-[0_4px_12px_rgba(37,99,235,0.08)]' : 'border-border bg-surface hover:border-primary/25 hover:bg-surface-elevated'}`}>
                <input className="sr-only" type="checkbox" checked={form.methods.includes(id)} onChange={() => toggleMethod(id)} />
                <span className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl transition-colors ${form.methods.includes(id) ? 'bg-primary text-white' : 'bg-surface-elevated text-text-secondary group-hover:text-primary'}`}>
                  {form.methods.includes(id) ? <Check className="h-4 w-4" /> : <Icon className="h-4 w-4" />}
                </span>
                <span className="min-w-0"><span className="block text-sm font-semibold text-text-primary">{label}</span><span className="mt-0.5 block text-xs text-text-secondary">{description}</span></span>
              </label>
            ))}
          </div>
        </div>

        <div className="flex items-start gap-3 rounded-xl border border-amber-200/70 bg-amber-50/70 p-4 text-xs leading-6 text-amber-950">
          <ShieldCheck className="mt-0.5 h-4 w-4 shrink-0 text-amber-700" />
          <span><strong className="font-semibold">تنبيه أمني:</strong> المفتاح السري لا يُعرض بعد حفظه. تفعيل الإنتاج لا يعني تنفيذ الدفع فعليًا قبل ربط مزود الدفع بالـBackend.</span>
        </div>

        <div className="flex flex-col-reverse gap-3 border-t border-border/70 pt-5 sm:flex-row sm:justify-end">
          {form.provider === 'cardcom' && (
            <Button variant="secondary" className="gap-2" onClick={() => connectionMutation.mutate()} disabled={connectionMutation.isPending}>
              <PlugZap className="h-4 w-4" />
              {connectionMutation.isPending ? 'جارِ اختبار الاتصال...' : 'اختبار الاتصال'}
            </Button>
          )}
          <Button variant="primary" className="gap-2 sm:min-w-[190px]" onClick={() => saveMutation.mutate()} disabled={saveMutation.isPending || form.methods.length === 0}>
            <Save className="h-4 w-4" />
            {saveMutation.isPending ? 'جارِ الحفظ...' : 'حفظ إعدادات الدفع'}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
