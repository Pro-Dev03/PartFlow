import { useState } from 'react'
import { clsx } from 'clsx'
import type { IntegrationSettings } from '../types/settings'

export function IntegrationsSettings() {
  const [integrations, setIntegrations] = useState<IntegrationSettings[]>([
    {
      id: '1',
      name: 'Stripe',
      type: 'payment',
      isEnabled: false,
      config: {},
      status: 'disconnected',
    },
    {
      id: '2',
      name: 'PayPal',
      type: 'payment',
      isEnabled: false,
      config: {},
      status: 'disconnected',
    },
    {
      id: '3',
      name: 'WhatsApp Business',
      type: 'notification',
      isEnabled: false,
      config: {},
      status: 'disconnected',
    },
    {
      id: '4',
      name: 'Email Service',
      type: 'notification',
      isEnabled: true,
      config: {},
      status: 'active',
    },
  ])
  const [showConfigModal, setShowConfigModal] = useState(false)
  const [selectedIntegration, setSelectedIntegration] = useState<IntegrationSettings | null>(null)

  const handleToggleIntegration = (id: string) => {
    setIntegrations(integrations.map(int => 
      int.id === id 
        ? { ...int, isEnabled: !int.isEnabled, status: int.isEnabled ? 'disconnected' : 'active' }
        : int
    ))
  }

  const handleConfigure = (integration: IntegrationSettings) => {
    setSelectedIntegration(integration)
    setShowConfigModal(true)
  }

  const handleSync = (id: string) => {
    // TODO: Trigger sync
    console.log('Sync integration:', id)
  }

  const typeLabels = {
    payment: 'دفع',
    shipping: 'شحن',
    accounting: 'محاسبة',
    notification: 'إشعارات',
    custom: 'مخصص',
  }

  const typeIcons = {
    payment: '💳',
    shipping: '🚚',
    accounting: '📊',
    notification: '🔔',
    custom: '⚙️',
  }

  const statusColors = {
    active: 'bg-success-10 text-success',
    error: 'bg-danger-10 text-danger',
    disconnected: 'bg-muted-10 text-muted',
  }

  const statusLabels = {
    active: 'نشط',
    error: 'خطأ',
    disconnected: 'غير متصل',
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h3 className="text-lg font-semibold text-text">التكاملات الخارجية</h3>
        <p className="text-sm text-muted">ربط النظام بخدمات خارجية</p>
      </div>

      {/* Integrations List */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {integrations.map((integration) => (
          <div key={integration.id} className="bg-surface rounded-lg p-4 border border-border">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <span className="text-2xl">{typeIcons[integration.type]}</span>
                <div>
                  <h4 className="font-medium text-text">{integration.name}</h4>
                  <p className="text-sm text-muted">{typeLabels[integration.type]}</p>
                </div>
              </div>
              <span className={clsx('px-2 py-1 rounded text-xs font-medium', statusColors[integration.status])}>
                {statusLabels[integration.status]}
              </span>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted">الحالة</span>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={integration.isEnabled}
                    onChange={() => handleToggleIntegration(integration.id)}
                    className="sr-only peer"
                  />
                  <div className="w-11 h-6 bg-muted peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
                </label>
              </div>

              {integration.lastSync && (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted">آخر مزامنة</span>
                  <span className="text-sm text-text">{integration.lastSync}</span>
                </div>
              )}
            </div>

            <div className="flex gap-2 mt-4 pt-4 border-t border-border">
              <button
                onClick={() => handleConfigure(integration)}
                className="flex-1 px-3 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors text-sm"
              >
                إعداد
              </button>
              {integration.isEnabled && (
                <button
                  onClick={() => handleSync(integration.id)}
                  className="flex-1 px-3 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors text-sm"
                >
                  مزامنة
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Add Custom Integration */}
      <button className="w-full px-4 py-3 border-2 border-dashed border-border rounded-lg hover:border-primary hover:bg-primary-5 transition-colors text-muted hover:text-primary">
        + إضافة تكامل مخصص
      </button>

      {/* Configuration Modal */}
      {showConfigModal && selectedIntegration && (
        <div className="fixed inset-0 bg-black-50 flex items-center justify-center z-50">
          <div className="bg-surface rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-text mb-4">
              إعداد {selectedIntegration.name}
            </h3>
            <div className="space-y-4">
              {/* TODO: Add configuration fields based on integration type */}
              <div>
                <label className="block text-sm font-medium text-text mb-2">API Key</label>
                <input
                  type="password"
                  className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  placeholder="أدخل مفتاح API"
                />
              </div>
              {selectedIntegration.type === 'payment' && (
                <div>
                  <label className="block text-sm font-medium text-text mb-2">Secret Key</label>
                  <input
                    type="password"
                    className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                    placeholder="أدخل المفتاح السري"
                  />
                </div>
              )}
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => {
                  setShowConfigModal(false)
                  setSelectedIntegration(null)
                }}
                className="px-4 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors"
              >
                إلغاء
              </button>
              <button
                onClick={() => {
                  // TODO: Save configuration
                  setShowConfigModal(false)
                  setSelectedIntegration(null)
                }}
                className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
              >
                حفظ
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
