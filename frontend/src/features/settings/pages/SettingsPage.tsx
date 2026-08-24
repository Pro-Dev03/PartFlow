import { useState } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { PageHeader } from '../../../components/ui/page-header';
import { Button } from '../../../components/ui/button';
import {
  Settings as SettingsIcon,
  Store,
  Palette,
  Bell,
  Lock,
  FileText,
  Zap,
  DollarSign
} from 'lucide-react';

// Components
import { StoreSettings } from '../components/StoreSettings';
import { NotificationSettings } from '../components/NotificationSettings';
import { AppearanceSettings } from '../components/AppearanceSettings';
import { SecuritySettings } from '../components/SecuritySettings';
import { FinancialSettings } from '../components/FinancialSettings';
import { AuditSettings } from '../components/AuditSettings';

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
          {activeTab === 'financial' && <FinancialSettings />}
          {activeTab === 'appearance' && <AppearanceSettings />}
          {activeTab === 'notifications' && <NotificationSettings />}
          {activeTab === 'security' && <SecuritySettings />}
          {activeTab === 'audit' && <AuditSettings />}
        </div>
      </div>
    </div>
  );
}