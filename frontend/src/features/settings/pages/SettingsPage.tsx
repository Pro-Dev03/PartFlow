import { useEffect, useState } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { PageHeader } from '../../../components/ui/page-header';
import { Button } from '../../../components/ui/button';
import { useAuthStore } from '../../../stores/authStore';
import {
  Store,
  Palette,
  Bell,
  FileText,
  DollarSign,
  Trash2,
  User,
  ShieldCheck,
  CalendarClock
} from 'lucide-react';

// Components
import { StoreSettings } from '../components/StoreSettings';
import { NotificationSettings } from '../components/NotificationSettings';
import { AppearanceSettings } from '../components/AppearanceSettings';
import { FinancialSettings } from '../components/FinancialSettings';
import { AuditSettings } from '../components/AuditSettings';
import { DatabaseSettings } from '../components/DatabaseSettings';

export function SettingsPage() {
  const { t } = useTranslation();
  const user = useAuthStore((state) => state.user);
  const [activeTab, setActiveTab] = useState('store');
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);

  const expiryDate = user?.subscription_expires_at ? new Date(user.subscription_expires_at) : null;
  const remainingMs = expiryDate ? expiryDate.getTime() - now : null;
  const remainingDays = remainingMs === null ? 0 : Math.max(0, Math.ceil(remainingMs / (1000 * 60 * 60 * 24)));
  const remainingText = remainingMs === null ? 'غير محدد' : `${remainingDays} يوم`;
  const displayName = user?.first_name || user?.name || 'مستخدم';
  const displayEmail = user?.email || 'غير متوفر';
  const displayPhone = user?.phone || 'غير متوفر';
  const subscriptionStatus = (user?.subscription_status || 'active').toLowerCase();
  const isActiveSubscription = subscriptionStatus === 'active' || subscriptionStatus === 'trial';

  const tabs = [
    { id: 'store', label: t('settings.store'), icon: Store },
    { id: 'financial', label: t('settings.financial'), icon: DollarSign },
    { id: 'appearance', label: t('settings.appearance'), icon: Palette },
    { id: 'notifications', label: t('settings.notifications'), icon: Bell },
    { id: 'audit', label: t('settings.audit'), icon: FileText },
    { id: 'database', label: 'قاعدة البيانات', icon: Trash2 },
  ];

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="System Control"
        title={t('settings.title')}
        description="إدارة إعدادات النظام"
      />

      <div className="mb-6 grid grid-cols-1 xl:grid-cols-3 gap-4">
        <div className="rounded-2xl border border-border bg-surface p-5 shadow-sm">
          <div className="mb-4 flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <User className="h-5 w-5" />
            </div>
            <div>
              <p className="text-xs text-muted-foreground">بيانات المستخدم</p>
              <h3 className="text-lg font-bold text-foreground">{displayName}</h3>
            </div>
          </div>

          <div className="space-y-3 text-sm text-foreground/90">
            <div className="flex items-center justify-between gap-3">
              <span className="text-muted-foreground">البريد</span>
              <span className="font-medium">{displayEmail}</span>
            </div>
            <div className="flex items-center justify-between gap-3">
              <span className="text-muted-foreground">الهاتف</span>
              <span className="font-medium">{displayPhone}</span>
            </div>
          </div>
        </div>

        <div className="rounded-2xl border border-border bg-surface p-5 shadow-sm">
          <div className="mb-4 flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-600">
              <ShieldCheck className="h-5 w-5" />
            </div>
            <div>
              <p className="text-xs text-muted-foreground">حالة الاشتراك</p>
              <h3 className="text-lg font-bold text-foreground">{isActiveSubscription ? 'نشط' : 'منتهي'}</h3>
            </div>
          </div>

          <div className="space-y-3 text-sm text-foreground/90">
            <div className="flex items-center justify-between gap-3">
              <span className="text-muted-foreground">الحالة</span>
              <span className="font-medium">{subscriptionStatus}</span>
            </div>
            <div className="flex items-center justify-between gap-3">
              <span className="text-muted-foreground">تاريخ الانتهاء</span>
              <span className="font-medium">{expiryDate ? expiryDate.toLocaleDateString('ar-EG') : 'غير محدد'}</span>
            </div>
          </div>
        </div>

        <div className="rounded-2xl border border-border bg-surface p-5 shadow-sm">
          <div className="mb-4 flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-amber-500/10 text-amber-600">
              <CalendarClock className="h-5 w-5" />
            </div>
            <div>
              <p className="text-xs text-muted-foreground">الوقت المتبقي</p>
              <h3 className="text-lg font-bold text-foreground">{remainingText}</h3>
            </div>
          </div>

          <div className="space-y-3 text-sm text-foreground/90">
            <div className="flex items-center justify-between gap-3">
              <span className="text-muted-foreground">تحديث مباشر</span>
              <span className="font-medium">{isActiveSubscription ? 'يتم التحديث تلقائياً' : 'الاشتراك منتهي'}</span>
            </div>
            <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className={`h-full rounded-full ${isActiveSubscription ? 'bg-emerald-500' : 'bg-red-500'}`}
                style={{ width: `${expiryDate && remainingMs !== null ? Math.max(0, Math.min(100, (remainingMs / (expiryDate.getTime() - Date.now() + remainingMs + 1)) * 100)) : 100}%` }}
              />
            </div>
          </div>
        </div>
      </div>

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
          {activeTab === 'financial' && <FinancialSettings />}
          {activeTab === 'appearance' && <AppearanceSettings />}
          {activeTab === 'notifications' && <NotificationSettings />}
          {activeTab === 'audit' && <AuditSettings />}
          {activeTab === 'database' && <DatabaseSettings />}
        </div>
      </div>
    </div>
  );
}
