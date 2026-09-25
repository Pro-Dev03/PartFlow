import { useEffect, useState } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { formatStoreDate, parseBackendTimestamp } from '../../../utils/store-time';
import { PageHeader } from '../../../design-system/components/page-header';
import { useAuthStore } from '../../../stores/authStore';
import { authApi } from '../../../services/api/endpoints';
import {
  BadgeCheck,
  CalendarClock,
  ChevronLeft,
  Database,
  FileClock,
  Palette,
  ShieldCheck,
  Store,
  UserRound,
  WalletCards,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import './settings-page.css';

import { StoreSettings } from '../components/StoreSettings';
import { AppearanceSettings } from '../components/AppearanceSettings';
import { FinancialSettings } from '../components/FinancialSettings';
import { AuditSettings } from '../components/AuditSettings';
import { DatabaseSettings } from '../components/DatabaseSettings';
import { SubscriptionManagement } from '../components/SubscriptionManagement';
import { RegionalSettings } from '../components/RegionalSettings';
import { SubscriberSyncSettings } from '../components/SubscriberSyncSettings';

interface SettingsSection {
  id: string;
  label: string;
  description: string;
  icon: LucideIcon;
}

export function SettingsPage() {
  const { t } = useTranslation();
  const user = useAuthStore((state) => state.user);
  const [activeTab, setActiveTab] = useState('store');
  const [now, setNow] = useState(() => Date.now());
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 60_000);
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

  const expiryDate = user?.subscription_expires_at ? parseBackendTimestamp(user.subscription_expires_at) : null;
  const remainingMs = expiryDate ? expiryDate.getTime() - now : null;
  const remainingDays = remainingMs === null ? 0 : Math.max(0, Math.ceil(remainingMs / (1000 * 60 * 60 * 24)));
  const remainingText = remainingMs === null ? 'غير محدد' : `${remainingDays} يوم`;
  const displayName = user?.first_name || user?.name || 'مستخدم';
  const displayEmail = user?.email || 'غير متوفر';
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

  const tabs: SettingsSection[] = [
    {
      id: 'store',
      label: t('settings.store'),
      description: 'بيانات المتجر والمنطقة والمزامنة',
      icon: Store,
    },
    {
      id: 'financial',
      label: t('settings.financial'),
      description: 'العملة والضرائب والخصومات',
      icon: WalletCards,
    },
    {
      id: 'appearance',
      label: t('settings.appearance'),
      description: 'المظهر وطريقة عرض النظام',
      icon: Palette,
    },
    {
      id: 'audit',
      label: t('settings.audit'),
      description: 'سجل الإجراءات والاحتفاظ به',
      icon: FileClock,
    },
    ...(isAdmin && isOwner ? [{
      id: 'database',
      label: 'البيانات والاشتراك',
      description: 'إدارة الاتصال وقاعدة البيانات',
      icon: Database,
    }] : []),
  ];

  useEffect(() => {
    if (!(isAdmin && isOwner) && activeTab === 'database') {
      setActiveTab('store');
    }
  }, [activeTab, isAdmin, isOwner]);

  const activeSection = tabs.find((tab) => tab.id === activeTab) ?? tabs[0];
  const ActiveSectionIcon = activeSection.icon;

  return (
    <div className="settings-page" dir="rtl">
      <PageHeader
        title={t('settings.title')}
        description="إدارة حسابك واشتراكك وتفضيلات المتجر من مكان واحد"
      />

      <section className="settings-account-overview" aria-label="ملخص الحساب">
        <div className="settings-account-identity">
          <div className="settings-account-avatar" aria-hidden="true">
            <UserRound />
          </div>
          <div className="settings-account-copy">
            <span className="settings-overline">حساب المتجر</span>
            <h2 title={displayName}>{displayName}</h2>
            <p dir="ltr" title={displayEmail}>{displayEmail}</p>
          </div>
        </div>

        <div className="settings-account-detail">
          <span className="settings-detail-icon settings-detail-icon-status">
            <ShieldCheck aria-hidden="true" />
          </span>
          <div>
            <span className="settings-detail-label">حالة الاشتراك</span>
            <span className={`settings-status-pill ${isActiveSubscription ? 'is-active' : 'is-inactive'}`}>
              <span aria-hidden="true" />
              {subscriptionStatusLabel}
            </span>
          </div>
        </div>

        <div className="settings-account-detail">
          <span className="settings-detail-icon">
            <CalendarClock aria-hidden="true" />
          </span>
          <div className="settings-expiry-copy">
            <span className="settings-detail-label">تاريخ الانتهاء</span>
            <strong>
              {expiryDate
                ? formatStoreDate(expiryDate, 'ar-EG')
                : 'غير محدد'}
            </strong>
          </div>
        </div>

        <div className="settings-account-remaining">
          <div>
            <span className="settings-detail-label">الوقت المتبقي</span>
            <strong>{remainingText}</strong>
          </div>
          <BadgeCheck className="settings-remaining-mark" aria-hidden="true" />
        </div>
      </section>

      <div className="settings-workspace">
        <nav className="settings-section-nav" aria-label="أقسام الإعدادات">
          <div className="settings-nav-heading">
            <span>الإعدادات</span>
            <p>اختر القسم الذي تريد تعديله</p>
          </div>
          <div className="settings-nav-items">
            {tabs.map((tab) => {
              const Icon = tab.icon;
              const isSelected = activeTab === tab.id;

              return (
                <button
                  key={tab.id}
                  id={`settings-nav-${tab.id}`}
                  type="button"
                  className={`settings-nav-item${isSelected ? ' is-selected' : ''}`}
                  aria-current={isSelected ? 'page' : undefined}
                  aria-controls="settings-active-panel"
                  onClick={() => setActiveTab(tab.id)}
                >
                  <span className="settings-nav-icon"><Icon aria-hidden="true" /></span>
                  <span className="settings-nav-copy">
                    <strong>{tab.label}</strong>
                    <small>{tab.description}</small>
                  </span>
                  <ChevronLeft className="settings-nav-chevron" aria-hidden="true" />
                </button>
              );
            })}
          </div>
          <div className="settings-nav-footnote">
            <span className="settings-nav-footnote-icon"><ShieldCheck aria-hidden="true" /></span>
            <p>تُحفظ التغييرات في إعدادات المتجر وفق صلاحيات حسابك.</p>
          </div>
        </nav>

        <section
          id="settings-active-panel"
          className="settings-panel"
          aria-labelledby={`settings-panel-title-${activeTab}`}
        >
          <div className="settings-panel-heading">
            <span className="settings-panel-icon"><ActiveSectionIcon aria-hidden="true" /></span>
            <div>
              <span className="settings-overline">إدارة التفضيلات</span>
              <h2 id={`settings-panel-title-${activeTab}`}>{activeSection.label}</h2>
              <p>{activeSection.description}</p>
            </div>
          </div>

          <div className={`settings-panel-content settings-panel-content--${activeTab}`}>
            {activeTab === 'store' && (
              <div className="settings-card-stack">
                <StoreSettings />
                <RegionalSettings canManageRegionalSettings={isActiveSubscription} />
                {isActiveSubscription && <SubscriberSyncSettings canUpload={isAdmin} />}
              </div>
            )}
            {activeTab === 'financial' && <FinancialSettings />}
            {activeTab === 'appearance' && <AppearanceSettings />}
            {activeTab === 'audit' && <AuditSettings />}
            {activeTab === 'database' && isAdmin && isOwner && (
              <div className="settings-card-stack">
                <DatabaseSettings />
                <SubscriptionManagement />
              </div>
            )}
          </div>
        </section>
      </div>
    </div>
  );
}
