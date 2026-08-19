import { useState } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { PageHeader } from '../../../components/ui/page-header';
import { 
  Settings as SettingsIcon, 
  Building2, 
  Users, 
  Shield,
  Store,
  Globe,
  Palette,
  Bell,
  Lock,
  FileText,
  Save,
  Sparkles,
  Sliders,
  Zap
} from 'lucide-react';

export function SettingsPage() {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('organization');

  const tabs = [
    { id: 'organization', label: t('settings.organization'), icon: Building2 },
    { id: 'users', label: t('settings.users'), icon: Users },
    { id: 'permissions', label: t('settings.permissions'), icon: Shield },
    { id: 'store', label: t('settings.store'), icon: Store },
    { id: 'localization', label: t('settings.localization'), icon: Globe },
    { id: 'appearance', label: t('settings.appearance'), icon: Palette },
    { id: 'notifications', label: t('settings.notifications'), icon: Bell },
    { id: 'security', label: t('settings.security'), icon: Lock },
    { id: 'audit', label: t('settings.audit'), icon: FileText },
  ];

  return (
    <div className="space-y-md">
      {/* Page Header - Futuristic + Minimal */}
      <PageHeader
        eyebrow="System Control"
        title={t('settings.title')}
        description="إدارة إعدادات النظام والمؤسسة"
        actions={
          <div className="flex items-center gap-sm">
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
          {activeTab === 'organization' && <OrganizationSettings />}
          {activeTab === 'users' && <UsersSettings />}
          {activeTab === 'permissions' && <PermissionsSettings />}
          {activeTab === 'store' && <StoreSettings />}
          {activeTab === 'localization' && <LocalizationSettings />}
          {activeTab === 'appearance' && <AppearanceSettings />}
          {activeTab === 'notifications' && <NotificationSettings />}
          {activeTab === 'security' && <SecuritySettings />}
          {activeTab === 'audit' && <AuditSettings />}
        </div>
      </div>
    </div>
  );
}

function OrganizationSettings() {
  const { t } = useTranslation();
  const [formData, setFormData] = useState({
    name: 'PartFlow Store',
    email: 'store@partflow.com',
    phone: '+966 50 123 4567',
    address: 'الرياض، المملكة العربية السعودية',
    taxNumber: '300123456700003',
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Building2 className="w-5 h-5 text-cyan" />
          إعدادات المؤسسة
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <Input
          label={t('settings.storeName')}
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
        />
        <Input
          label={t('settings.storeEmail')}
          type="email"
          value={formData.email}
          onChange={(e) => setFormData({ ...formData, email: e.target.value })}
        />
        <Input
          label={t('settings.storePhone')}
          value={formData.phone}
          onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
        />
        <Input
          label={t('settings.storeAddress')}
          value={formData.address}
          onChange={(e) => setFormData({ ...formData, address: e.target.value })}
        />
        <Input
          label="الرقم الضريبي"
          value={formData.taxNumber}
          onChange={(e) => setFormData({ ...formData, taxNumber: e.target.value })}
        />
        <Button variant="primary" className="gap-2">
          <Save className="w-4 h-4" />
          حفظ التغييرات
        </Button>
      </CardContent>
    </Card>
  );
}

function UsersSettings() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Users className="w-5 h-5 text-cyan" />
          إدارة المستخدمين
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-text-muted">
          <Users className="w-12 h-12 mx-auto mb-md opacity-50 text-cyan" />
          <p className="text-small">قائمة المستخدمين</p>
        </div>
      </CardContent>
    </Card>
  );
}

function PermissionsSettings() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="w-5 h-5 text-cyan" />
          إدارة الصلاحيات
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-text-muted">
          <Shield className="w-12 h-12 mx-auto mb-md opacity-50 text-cyan" />
          <p className="text-small">إعدادات الصلاحيات</p>
        </div>
      </CardContent>
    </Card>
  );
}

function StoreSettings() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Store className="w-5 h-5 text-cyan" />
          إعدادات المتجر
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-text-muted">
          <Store className="w-12 h-12 mx-auto mb-md opacity-50 text-cyan" />
          <p className="text-small">إعدادات المتجر</p>
        </div>
      </CardContent>
    </Card>
  );
}

function LocalizationSettings() {
  const { t, languages, currentLanguage, changeLanguage } = useTranslation();

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Globe className="w-5 h-5 text-cyan" />
          اللغة والتوطين
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <div>
          <label className="text-small font-medium text-text mb-sm block">
            {t('settings.language')}
          </label>
          <Select
            value={currentLanguage}
            onChange={(e) => changeLanguage(e.target.value as 'ar' | 'en')}
            options={languages.map(lang => ({
              value: lang.code,
              label: `${lang.nativeName} (${lang.flag})`
            }))}
          />
        </div>
        <div>
          <label className="text-small font-medium text-text mb-sm block">
            {t('settings.timezone')}
          </label>
          <Select
            value="Asia/Riyadh"
            onChange={() => {}}
            options={[
              { value: 'Asia/Riyadh', label: 'الرياض (GMT+3)' },
              { value: 'Asia/Dubai', label: 'دبي (GMT+4)' },
              { value: 'Asia/Cairo', label: 'القاهرة (GMT+2)' },
            ]}
          />
        </div>
        <div>
          <label className="text-small font-medium text-text mb-sm block">
            {t('settings.currency')}
          </label>
          <Select
            value="SAR"
            onChange={() => {}}
            options={[
              { value: 'SAR', label: 'ريال سعودي (₪)' },
              { value: 'AED', label: 'درهم إماراتي (د.إ)' },
              { value: 'EGP', label: 'جنيه مصري (ج.م)' },
            ]}
          />
        </div>
      </CardContent>
    </Card>
  );
}

function AppearanceSettings() {
  const { t } = useTranslation();
  const [theme, setTheme] = useState('light');

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
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bell className="w-5 h-5 text-cyan" />
          إعدادات الإشعارات
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-text-muted">
          <Bell className="w-12 h-12 mx-auto mb-md opacity-50 text-cyan" />
          <p className="text-small">إعدادات الإشعارات</p>
        </div>
      </CardContent>
    </Card>
  );
}

function SecuritySettings() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Lock className="w-5 h-5 text-cyan" />
          الأمان
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-text-muted">
          <Lock className="w-12 h-12 mx-auto mb-md opacity-50 text-cyan" />
          <p className="text-small">إعدادات الأمان</p>
        </div>
      </CardContent>
    </Card>
  );
}

function AuditSettings() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileText className="w-5 h-5 text-cyan" />
          سجل التدقيق
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-text-muted">
          <FileText className="w-12 h-12 mx-auto mb-md opacity-50 text-cyan" />
          <p className="text-small">سجل التدقيق</p>
        </div>
      </CardContent>
    </Card>
  );
}