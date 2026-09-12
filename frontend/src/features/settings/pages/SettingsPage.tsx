import { useEffect, useState } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { PageHeader } from '../../../components/ui/page-header';
import { Button } from '../../../components/ui/button';
import { useAuthStore } from '../../../stores/authStore';
import { authApi } from '../../../services/api/endpoints';
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
import { SubscriptionManagement } from '../components/SubscriptionManagement';

export function SettingsPage() {
  const { t } = useTranslation();
  const user = useAuthStore((state) => state.user);
  const [activeTab, setActiveTab] = useState('store');
  const [now, setNow] = useState(Date.now());
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    let mounted = true;
    void authApi.checkAdminAccess()
      .then((result) => {
        if (mounted) setIsAdmin(Boolean(result?.is_admin));
      })
      .catch(() => {
        if (mounted) setIsAdmin(false);
      });

    return () => {
      mounted = false;
    };
  }, []);

  const expiryDate = user?.subscription_expires_at ? new Date(user.subscription_expires_at) : null;
  const remainingMs = expiryDate ? expiryDate.getTime() - now : null;
  const remainingDays = remainingMs === null ? 0 : Math.max(0, Math.ceil(remainingMs / (1000 * 60 * 60 * 24)));
  const remainingText = remainingMs === null ? 'غير محدد' : `${remainingDays} يوم`;
  const displayName = user?.first_name || user?.name || 'مستخدم';
  const displayEmail = user?.email || 'غير متوفر';
  const displayPhone = user?.phone || 'غير متوفر';
  const isOwner = user?.email?.trim().toLowerCase() === 'owner@partflow.com';
  const subscriptionStatus = (user?.subscription_status || 'active').toLowerCase();
  const isActiveSubscription = subscriptionStatus === 'active' || subscriptionStatus === 'trial';
  const subscriptionStatusLabel = subscriptionStatus === 'active'
    ? 'نشط'
    : subscriptionStatus === 'trial'
      ? 'تجريبي'
      : subscriptionStatus === 'expired'
        ? 'منتهي'
        : subscriptionStatus === 'canceled' || subscriptionStatus === 'cancelled'
          ? 'موقوف'
          : subscriptionStatus;

  const tabs = [
    { id: 'store', label: t('settings.store'), icon: Store },
    { id: 'financial', label: t('settings.financial'), icon: DollarSign },
    { id: 'appearance', label: t('settings.appearance'), icon: Palette },
    { id: 'notifications', label: t('settings.notifications'), icon: Bell },
    { id: 'audit', label: t('settings.audit'), icon: FileText },
    ...(isAdmin ? [{ id: 'database', label: 'قاعدة البيانات', icon: Trash2 }] : []),
  ];

  useEffect(() => {
    if (!isAdmin && activeTab === 'database') {
      setActiveTab('store');
    }
  }, [activeTab, isAdmin, isOwner]);

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="مركز التحكم"
        title={t('settings.title')}
        description="تحكم في حسابك واشتراكك وإعدادات متجرك من مكان واحد"
      />

      <div
        className="mb-6 grid gap-4"
        dir="rtl"
        style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: '16px' }}
      >
        <div
          className="settings-summary-card user min-w-0 p-5"
          style={{
            borderColor: 'color-mix(in srgb, var(--primary) 18%, var(--border-default))',
          }}
        >
          <div className="mb-5 flex items-center gap-3">
            <div className="settings-summary-icon flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <User className="h-5 w-5" />
            </div>
            <div>
              <p className="text-xs font-medium text-muted-foreground">بيانات المستخدم</p>
              <h3 className="mt-1 text-lg font-bold text-foreground">{displayName}</h3>
            </div>
          </div>

          <div className="space-y-3 text-sm">
            <div className="border-t border-border/60 pt-3">
              <span className="block text-xs text-muted-foreground">البريد الإلكتروني</span>
              <span dir="ltr" className="mt-1 block min-w-0 truncate text-right font-semibold text-foreground">{displayEmail}</span>
            </div>
            <div className="border-t border-border/60 pt-3">
              <span className="text-muted-foreground">الهاتف</span>
              <span dir="ltr" className="mt-1 block truncate text-right font-semibold text-foreground">{displayPhone}</span>
            </div>
          </div>
        </div>

        <div
          className="settings-summary-card subscription min-w-0 p-5"
          style={{
            borderColor: 'rgba(16, 185, 129, 0.2)',
          }}
        >
          <div className="mb-5 flex items-center gap-3">
            <div className="settings-summary-icon flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-600">
              <ShieldCheck className="h-5 w-5" />
            </div>
            <div>
              <p className="text-xs font-medium text-muted-foreground">حالة الاشتراك</p>
              <h3 className="mt-1 text-lg font-bold text-foreground">{isActiveSubscription ? 'نشط' : 'منتهي'}</h3>
            </div>
          </div>

          <div className="space-y-3 text-sm">
            <div className="border-t border-border/60 pt-3">
              <span className="block text-xs text-muted-foreground">الحالة الحالية</span>
              <span className="mt-1 inline-flex rounded-full bg-emerald-500/10 px-2.5 py-1 text-xs font-bold text-emerald-700">{subscriptionStatusLabel}</span>
            </div>
            <div className="border-t border-border/60 pt-3">
              <span className="block text-xs text-muted-foreground">تاريخ الانتهاء</span>
              <span className="mt-1 block truncate font-semibold text-foreground">{expiryDate ? new Intl.DateTimeFormat('ar-EG', { dateStyle: 'medium', timeZone: 'UTC' }).format(expiryDate) : 'غير محدد'}</span>
            </div>
          </div>
        </div>

        <div
          className="settings-summary-card time min-w-0 p-5"
          style={{
            borderColor: 'rgba(245, 158, 11, 0.24)',
          }}
        >
          <div className="mb-5 flex items-center gap-3">
            <div className="settings-summary-icon flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-amber-500/10 text-amber-600">
              <CalendarClock className="h-5 w-5" />
            </div>
            <div>
              <p className="text-xs font-medium text-muted-foreground">الوقت المتبقي</p>
              <h3 className="mt-1 text-2xl font-bold text-foreground">{remainingText}</h3>
            </div>
          </div>

          <div className="space-y-3 text-sm">
            <div className="border-t border-border/60 pt-3">
              <span className="block text-xs text-muted-foreground">حالة التحديث</span>
              <span className="mt-1 block font-semibold text-foreground">{isActiveSubscription ? 'يتم التحديث تلقائياً' : 'الاشتراك منتهي'}</span>
            </div>
            <div className="h-2.5 w-full overflow-hidden rounded-full bg-muted">
              <div
                className={`h-full rounded-full ${isActiveSubscription ? 'bg-emerald-500' : 'bg-red-500'}`}
                style={{ width: `${expiryDate && remainingMs !== null ? Math.max(0, Math.min(100, (remainingMs / (expiryDate.getTime() - Date.now() + remainingMs + 1)) * 100)) : 100}%` }}
              />
            </div>
          </div>
        </div>
      </div>

      <div
        className="grid gap-5"
        style={{ gridTemplateColumns: 'minmax(180px, 220px) minmax(0, 1fr)', gap: '20px' }}
      >
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
        <div className="min-w-0">
          {activeTab === 'store' && <StoreSettings />}
          {activeTab === 'financial' && <FinancialSettings />}
          {activeTab === 'appearance' && <AppearanceSettings />}
          {activeTab === 'notifications' && <NotificationSettings />}
          {activeTab === 'audit' && <AuditSettings />}
          {activeTab === 'database' && isAdmin && (
            <div className="space-y-6">
              <DatabaseSettings />
              {isAdmin && <SubscriptionManagement />}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
