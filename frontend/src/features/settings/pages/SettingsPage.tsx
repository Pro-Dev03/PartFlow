import { useState, useEffect } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { PageHeader } from '../../../components/ui/page-header';
import { settingsApi } from '../../../services/api/endpoints';
import {
  Settings as SettingsIcon,
  Store,
  Palette,
  Bell,
  Lock,
  FileText,
  Save,
  Sparkles,
  Sliders,
  Zap,
  DollarSign
} from 'lucide-react';

export function SettingsPage() {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('store');

  const tabs = [
    { id: 'store', label: t('settings.store'), icon: Store },
    { id: 'financial', label: t('settings.financial'), icon: DollarSign },
    { id: 'appearance', label: t('settings.appearance'), icon: Palette },
    { id: 'notifications', label: t('settings.notifications'), icon: Bell },
    { id: 'security', label: t('settings.security'), icon: Lock },
    { id: 'audit', label: t('settings.audit'), icon: FileText },
  ];

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="System Control"
        title={t('settings.title')}
        description="إدارة إعدادات النظام"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="secondary" className="gap-2">
              <Zap className="w-4 h-4" />
              تحديث
            </Button>
            <Button variant="primary" className="gap-2">
              <Save className="w-4 h-4" />
              حفظ
            </Button>
          </div>
        }
      />

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-md">
        {/* Sidebar Tabs - Futuristic + Minimal */}
        <div className="space-y-sm">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            return (
              <Button
                key={tab.id}
                variant={activeTab === tab.id ? 'primary' : 'ghost'}
                className="w-full justify-start gap-sm"
                onClick={() => setActiveTab(tab.id)}
              >
                <Icon className="w-4 h-4" />
                <span className="text-tiny">{tab.label}</span>
              </Button>
            );
          })}
        </div>

        {/* Content Area - Futuristic + Minimal */}
        <div className="lg:col-span-3">
          {activeTab === 'store' && <StoreSettings />}
          {activeTab === 'appearance' && <AppearanceSettings />}
          {activeTab === 'notifications' && <NotificationSettings />}
          {activeTab === 'security' && <SecuritySettings />}
          {activeTab === 'audit' && <AuditSettings />}
        </div>
      </div>
    </div>
  );
}

function StoreSettings() {
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

function AppearanceSettings() {
  const { t } = useTranslation();
  const [theme, setTheme] = useState(() => {
    // Initialize from localStorage or default to dark
    const savedTheme = localStorage.getItem('theme');
    return savedTheme || 'dark';
  });

  useEffect(() => {
    // Apply theme to document
    const root = document.documentElement;
    if (theme === 'light') {
      root.classList.add('light');
      root.classList.remove('dark');
    } else {
      root.classList.remove('light');
      root.classList.add('dark');
    }
    // Save to localStorage
    localStorage.setItem('theme', theme);
  }, [theme]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Palette className="w-5 h-5 text-cyan" />
          المظهر
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <div>
          <label className="text-small font-medium text-text mb-sm block">
            {t('settings.theme')}
          </label>
          <div className="grid grid-cols-3 gap-sm">
            <Button
              variant={theme === 'light' ? 'primary' : 'secondary'}
              onClick={() => setTheme('light')}
            >
              {t('settings.lightMode')}
            </Button>
            <Button
              variant={theme === 'dark' ? 'primary' : 'secondary'}
              onClick={() => setTheme('dark')}
            >
              {t('settings.darkMode')}
            </Button>
            <Button
              variant={theme === 'system' ? 'primary' : 'secondary'}
              onClick={() => setTheme('system')}
            >
              {t('settings.systemMode')}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function NotificationSettings() {
  const [notificationSettings, setNotificationSettings] = useState({
    lowStockAlerts: true,
    debtPaymentAlerts: true,
    warrantyExpiryAlerts: true,
    dailyReports: false,
    emailNotifications: true,
    browserNotifications: true,
  });

  const toggleNotification = (key: string) => {
    setNotificationSettings({ ...notificationSettings, [key]: !notificationSettings[key as keyof typeof notificationSettings] });
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bell className="w-5 h-5 text-cyan" />
          إعدادات الإشعارات
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">تنبيهات المخزون المنخفض</h4>
            <p className="text-small text-text-secondary">إشعار عند وصول المنتج للحد الأدنى</p>
          </div>
          <Button
            variant={notificationSettings.lowStockAlerts ? 'primary' : 'secondary'}
            onClick={() => toggleNotification('lowStockAlerts')}
          >
            {notificationSettings.lowStockAlerts ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">تنبيهات دفع الديون</h4>
            <p className="text-small text-text-secondary">إشعار عند استحقاق دفعات الديون</p>
          </div>
          <Button
            variant={notificationSettings.debtPaymentAlerts ? 'primary' : 'secondary'}
            onClick={() => toggleNotification('debtPaymentAlerts')}
          >
            {notificationSettings.debtPaymentAlerts ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">تنبيهات انتهاء الضمان</h4>
            <p className="text-small text-text-secondary">إشعار عند قرب انتهاء ضمان المنتج</p>
          </div>
          <Button
            variant={notificationSettings.warrantyExpiryAlerts ? 'primary' : 'secondary'}
            onClick={() => toggleNotification('warrantyExpiryAlerts')}
          >
            {notificationSettings.warrantyExpiryAlerts ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">التقارير اليومية</h4>
            <p className="text-small text-text-secondary">إرسال ملخص يومي بالأنشطة</p>
          </div>
          <Button
            variant={notificationSettings.dailyReports ? 'primary' : 'secondary'}
            onClick={() => toggleNotification('dailyReports')}
          >
            {notificationSettings.dailyReports ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">إشعارات البريد الإلكتروني</h4>
            <p className="text-small text-text-secondary">استلام الإشعارات عبر البريد الإلكتروني</p>
          </div>
          <Button
            variant={notificationSettings.emailNotifications ? 'primary' : 'secondary'}
            onClick={() => toggleNotification('emailNotifications')}
          >
            {notificationSettings.emailNotifications ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">إشعارات المتصفح</h4>
            <p className="text-small text-text-secondary">استلام الإشعارات في المتصفح</p>
          </div>
          <Button
            variant={notificationSettings.browserNotifications ? 'primary' : 'secondary'}
            onClick={() => toggleNotification('browserNotifications')}
          >
            {notificationSettings.browserNotifications ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <Button variant="primary" className="gap-2">
          <Save className="w-4 h-4" />
          حفظ التغييرات
        </Button>
      </CardContent>
    </Card>
  );
}

function FinancialSettings() {
  const { t } = useTranslation();
  const [taxRate, setTaxRate] = useState(0);

  // Fetch current tax rate
  const { data: taxData, isLoading } = useQuery({
    queryKey: ['taxRate'],
    queryFn: () => settingsApi.getTaxRate(),
  });

  // Update tax rate
  const updateTaxMutation = useMutation({
    mutationFn: (newRate: number) => settingsApi.updateTaxRate(newRate),
    onSuccess: () => {
      alert('تم تحديث نسبة الضريبة بنجاح');
    },
    onError: () => {
      alert('فشل تحديث نسبة الضريبة');
    },
  });

  useEffect(() => {
    if (taxData?.data?.tax_rate !== undefined) {
      setTaxRate(taxData.data.tax_rate);
    }
  }, [taxData]);

  const handleSaveTaxRate = () => {
    if (taxRate < 0 || taxRate > 100) {
      alert('نسبة الضريبة يجب أن تكون بين 0 و 100');
      return;
    }
    updateTaxMutation.mutate(taxRate);
  };

  if (isLoading) {
    return <div>جاري التحميل...</div>;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <DollarSign className="w-5 h-5 text-cyan" />
          الإعدادات المالية
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <div className="p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary mb-2">نسبة الضريبة</h4>
            <p className="text-small text-text-secondary mb-4">نسبة الضريبة المئوية المطبقة على المبيعات</p>
          </div>
          <div className="flex items-center gap-2">
            <Input
              type="number"
              min="0"
              max="100"
              step="0.1"
              value={taxRate}
              onChange={(e) => setTaxRate(parseFloat(e.target.value) || 0)}
              className="w-24"
            />
            <span className="text-text">%</span>
          </div>
        </div>
        <Button
          variant="primary"
          className="gap-2"
          onClick={handleSaveTaxRate}
          disabled={updateTaxMutation.isPending}
        >
          <Save className="w-4 h-4" />
          {updateTaxMutation.isPending ? 'جاري الحفظ...' : 'حفظ التغييرات'}
        </Button>
      </CardContent>
    </Card>
  );
}

function SecuritySettings() {
  const [securitySettings, setSecuritySettings] = useState({
    twoFactorAuth: false,
    sessionTimeout: 30,
    passwordMinLength: 8,
    requireUppercase: true,
    requireNumbers: true,
    requireSpecialChars: true,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Lock className="w-5 h-5 text-cyan" />
          الأمان
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">المصادقة الثنائية</h4>
            <p className="text-small text-text-secondary">تأكيد الهوية بخطوتين</p>
          </div>
          <Button
            variant={securitySettings.twoFactorAuth ? 'primary' : 'secondary'}
            onClick={() => setSecuritySettings({ ...securitySettings, twoFactorAuth: !securitySettings.twoFactorAuth })}
          >
            {securitySettings.twoFactorAuth ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <Input
          label="مهلة الجلسة (دقيقة)"
          type="number"
          value={securitySettings.sessionTimeout}
          onChange={(e) => setSecuritySettings({ ...securitySettings, sessionTimeout: Number(e.target.value) })}
        />
        <Input
          label="الحد الأدنى لطول كلمة المرور"
          type="number"
          value={securitySettings.passwordMinLength}
          onChange={(e) => setSecuritySettings({ ...securitySettings, passwordMinLength: Number(e.target.value) })}
        />
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">تطلب أحرف كبيرة</h4>
            <p className="text-small text-text-secondary">كلمة المرور يجب أن تحتوي على أحرف كبيرة</p>
          </div>
          <Button
            variant={securitySettings.requireUppercase ? 'primary' : 'secondary'}
            onClick={() => setSecuritySettings({ ...securitySettings, requireUppercase: !securitySettings.requireUppercase })}
          >
            {securitySettings.requireUppercase ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">تطلب أرقام</h4>
            <p className="text-small text-text-secondary">كلمة المرور يجب أن تحتوي على أرقام</p>
          </div>
          <Button
            variant={securitySettings.requireNumbers ? 'primary' : 'secondary'}
            onClick={() => setSecuritySettings({ ...securitySettings, requireNumbers: !securitySettings.requireNumbers })}
          >
            {securitySettings.requireNumbers ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <Button variant="primary" className="gap-2">
          <Save className="w-4 h-4" />
          حفظ التغييرات
        </Button>
      </CardContent>
    </Card>
  );
}

function AuditSettings() {
  const [auditLogs, setAuditLogs] = useState<any[]>([]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileText className="w-5 h-5 text-cyan" />
          سجل التدقيق
        </CardTitle>
        <Button variant="secondary" className="gap-2">
          <FileText className="w-4 h-4" />
          تصدير السجل
        </Button>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border text-text-secondary">
                <th className="text-start p-3 font-medium">المستخدم</th>
                <th className="text-start p-3 font-medium">الإجراء</th>
                <th className="text-start p-3 font-medium">التوقيت</th>
                <th className="text-start p-3 font-medium">IP</th>
              </tr>
            </thead>
            <tbody>
              {auditLogs.map((log) => (
                <tr key={log.id} className="border-b border-border hover:bg-surface-elevated">
                  <td className="p-3">{log.user}</td>
                  <td className="p-3">
                    <span className="px-2 py-1 rounded-full text-xs bg-cyan/20 text-cyan">
                      {log.action}
                    </span>
                  </td>
                  <td className="p-3">{log.timestamp}</td>
                  <td className="p-3 font-mono text-xs">{log.ip}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  );
}