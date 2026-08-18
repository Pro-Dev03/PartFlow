import { useState } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
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
  Save
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
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
          {t('settings.title')}
        </h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">
          إدارة إعدادات النظام والمؤسسة
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Sidebar Tabs */}
        <div className="space-y-2">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            return (
              <Button
                key={tab.id}
                variant={activeTab === tab.id ? 'primary' : 'ghost'}
                className="w-full justify-start gap-3"
                onClick={() => setActiveTab(tab.id)}
              >
                <Icon className="w-4 h-4" />
                {tab.label}
              </Button>
            );
          })}
        </div>

        {/* Content Area */}
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
        <CardTitle>إعدادات المؤسسة</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
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
        <Button className="gap-2">
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
        <CardTitle>إدارة المستخدمين</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-gray-500 dark:text-gray-400">
          <Users className="w-12 h-12 mx-auto mb-2 opacity-50" />
          <p>قائمة المستخدمين</p>
        </div>
      </CardContent>
    </Card>
  );
}

function PermissionsSettings() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>إدارة الصلاحيات</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-gray-500 dark:text-gray-400">
          <Shield className="w-12 h-12 mx-auto mb-2 opacity-50" />
          <p>إعدادات الصلاحيات</p>
        </div>
      </CardContent>
    </Card>
  );
}

function StoreSettings() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>إعدادات المتجر</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-gray-500 dark:text-gray-400">
          <Store className="w-12 h-12 mx-auto mb-2 opacity-50" />
          <p>إعدادات المتجر</p>
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
        <CardTitle>اللغة والتوطين</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <label className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 block">
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
          <label className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 block">
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
          <label className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 block">
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
        <CardTitle>المظهر</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <label className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 block">
            {t('settings.theme')}
          </label>
          <div className="grid grid-cols-3 gap-3">
            <Button
              variant={theme === 'light' ? 'primary' : 'outline'}
              onClick={() => setTheme('light')}
            >
              {t('settings.lightMode')}
            </Button>
            <Button
              variant={theme === 'dark' ? 'primary' : 'outline'}
              onClick={() => setTheme('dark')}
            >
              {t('settings.darkMode')}
            </Button>
            <Button
              variant={theme === 'system' ? 'primary' : 'outline'}
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
        <CardTitle>إعدادات الإشعارات</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-gray-500 dark:text-gray-400">
          <Bell className="w-12 h-12 mx-auto mb-2 opacity-50" />
          <p>إعدادات الإشعارات</p>
        </div>
      </CardContent>
    </Card>
  );
}

function SecuritySettings() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>الأمان</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-gray-500 dark:text-gray-400">
          <Lock className="w-12 h-12 mx-auto mb-2 opacity-50" />
          <p>إعدادات الأمان</p>
        </div>
      </CardContent>
    </Card>
  );
}

function AuditSettings() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>سجل التدقيق</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-gray-500 dark:text-gray-400">
          <FileText className="w-12 h-12 mx-auto mb-2 opacity-50" />
          <p>سجل التدقيق</p>
        </div>
      </CardContent>
    </Card>
  );
}