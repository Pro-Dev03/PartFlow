import { useState } from 'react'
import { clsx } from 'clsx'
import { OrganizationSettingsForm } from '../components/OrganizationSettingsForm'
import { UsersManagement } from '../components/UsersManagement'
import { NotificationSettingsForm } from '../components/NotificationSettingsForm'
import { SystemSettingsForm } from '../components/SystemSettingsForm'
import { IntegrationsSettings } from '../components/IntegrationsSettings'
import { BackupManagement } from '../components/BackupManagement'

type SettingsTab = 'organization' | 'users' | 'notifications' | 'system' | 'integrations' | 'backup'

export function SettingsPage() {
  const [activeTab, setActiveTab] = useState<SettingsTab>('organization')

  const tabs = [
    { id: 'organization' as SettingsTab, label: 'المؤسسة', icon: '🏢' },
    { id: 'users' as SettingsTab, label: 'المستخدمون', icon: '👥' },
    { id: 'notifications' as SettingsTab, label: 'الإشعارات', icon: '🔔' },
    { id: 'system' as SettingsTab, label: 'النظام', icon: '⚙️' },
    { id: 'integrations' as SettingsTab, label: 'التكاملات', icon: '🔗' },
    { id: 'backup' as SettingsTab, label: 'النسخ الاحتياطي', icon: '💾' },
  ]

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">الإعدادات</h1>
        <p className="text-muted">إدارة إعدادات النظام والمؤسسة</p>
      </div>

      {/* Tabs */}
      <div className="bg-surface rounded-lg">
        <div className="border-b border-border">
          <nav className="flex gap-4 px-6 overflow-x-auto">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={clsx(
                  'px-4 py-3 text-sm font-medium border-b-2 transition-colors whitespace-nowrap',
                  activeTab === tab.id
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted hover:text-text'
                )}
              >
                <span className="mr-2">{tab.icon}</span>
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Content */}
        <div className="p-6">
          {activeTab === 'organization' && <OrganizationSettingsForm />}
          {activeTab === 'users' && <UsersManagement />}
          {activeTab === 'notifications' && <NotificationSettingsForm />}
          {activeTab === 'system' && <SystemSettingsForm />}
          {activeTab === 'integrations' && <IntegrationsSettings />}
          {activeTab === 'backup' && <BackupManagement />}
        </div>
      </div>
    </div>
  )
}
