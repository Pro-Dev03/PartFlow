import { useState } from 'react'
import { clsx } from 'clsx'
import type { NotificationSettings } from '../types/settings'

export function NotificationSettingsForm() {
  const [settings, setSettings] = useState<Partial<NotificationSettings>>({
    email: {
      lowStock: true,
      overdueDebts: true,
      warrantyExpiring: true,
      purchaseAlerts: false,
      salesReports: true,
      systemUpdates: true,
    },
    push: {
      lowStock: true,
      overdueDebts: true,
      warrantyExpiring: true,
      purchaseAlerts: true,
      salesReports: false,
      systemUpdates: true,
    },
    inApp: {
      lowStock: true,
      overdueDebts: true,
      warrantyExpiring: true,
      purchaseAlerts: true,
      salesReports: true,
      systemUpdates: true,
    },
  })
  const [saving, setSaving] = useState(false)

  // TODO: Fetch current notification settings from API

  const handleSave = async () => {
    setSaving(true)
    try {
      // TODO: Save settings to API
      console.log('Saving notification settings:', settings)
    } catch (error) {
      console.error('Failed to save settings:', error)
    } finally {
      setSaving(false)
    }
  }

  const toggleSetting = (category: keyof NotificationSettings, setting: string) => {
    setSettings({
      ...settings,
      [category]: {
        ...settings[category],
        [setting]: !settings[category]?.[setting as keyof typeof settings[category]],
      },
    })
  }

  const notificationTypes = [
    { key: 'lowStock', label: 'انخفاض المخزون', description: 'إشعار عند وصول المنتج للحد الأدنى' },
    { key: 'overdueDebts', label: 'ديون متأخرة', description: 'إشعار عند تأخر دفعات العملاء' },
    { key: 'warrantyExpiring', label: 'انتهاء الضمان', description: 'إشعار قبل انتهاء ضمان المنتجات' },
    { key: 'purchaseAlerts', label: 'تنبيهات المشتريات', description: 'إشعارات تتعلق بالمشتريات الجديدة' },
    { key: 'salesReports', label: 'تقارير المبيعات', description: 'ملخصات دورية للمبيعات' },
    { key: 'systemUpdates', label: 'تحديثات النظام', description: 'إشعارات حول تحديثات النظام' },
  ]

  const categories = [
    { key: 'email' as const, label: 'البريد الإلكتروني', icon: '📧' },
    { key: 'push' as const, label: 'إشعارات الدفع', icon: '📱' },
    { key: 'inApp' as const, label: 'إشعارات التطبيق', icon: '🔔' },
  ]

  return (
    <div className="max-w-3xl space-y-6">
      {/* Header */}
      <div>
        <h3 className="text-lg font-semibold text-text">إعدادات الإشعارات</h3>
        <p className="text-sm text-muted">تحديد كيفية وأين استلام الإشعارات</p>
      </div>

      {/* Notification Settings Table */}
      <div className="bg-surface rounded-lg overflow-hidden">
        <table className="w-full">
          <thead className="bg-muted-10 border-b border-border">
            <tr>
              <th className="px-4 py-3 text-right text-sm font-medium text-muted">نوع الإشعار</th>
              {categories.map((cat) => (
                <th key={cat.key} className="px-4 py-3 text-center text-sm font-medium text-text">
                  <span className="mr-2">{cat.icon}</span>
                  {cat.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {notificationTypes.map((type) => (
              <tr key={type.key} className="border-b border-border">
                <td className="px-4 py-3">
                  <div>
                    <p className="font-medium text-text">{type.label}</p>
                    <p className="text-sm text-muted">{type.description}</p>
                  </div>
                </td>
                {categories.map((cat) => (
                  <td key={cat.key} className="px-4 py-3 text-center">
                    <label className="inline-flex items-center justify-center">
                      <input
                        type="checkbox"
                        checked={settings[cat.key]?.[type.key as keyof typeof settings[typeof cat.key]] || false}
                        onChange={() => toggleSetting(cat.key, type.key)}
                        className="rounded border-border w-5 h-5"
                      />
                    </label>
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Quick Actions */}
      <div className="flex gap-3">
        <button
          onClick={() => {
            // Enable all
            const allEnabled = {
              email: { lowStock: true, overdueDebts: true, warrantyExpiring: true, purchaseAlerts: true, salesReports: true, systemUpdates: true },
              push: { lowStock: true, overdueDebts: true, warrantyExpiring: true, purchaseAlerts: true, salesReports: true, systemUpdates: true },
              inApp: { lowStock: true, overdueDebts: true, warrantyExpiring: true, purchaseAlerts: true, salesReports: true, systemUpdates: true },
            }
            setSettings(allEnabled)
          }}
          className="px-4 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors text-sm"
        >
          تفعيل الكل
        </button>
        <button
          onClick={() => {
            // Disable all
            const allDisabled = {
              email: { lowStock: false, overdueDebts: false, warrantyExpiring: false, purchaseAlerts: false, salesReports: false, systemUpdates: false },
              push: { lowStock: false, overdueDebts: false, warrantyExpiring: false, purchaseAlerts: false, salesReports: false, systemUpdates: false },
              inApp: { lowStock: false, overdueDebts: false, warrantyExpiring: false, purchaseAlerts: false, salesReports: false, systemUpdates: false },
            }
            setSettings(allDisabled)
          }}
          className="px-4 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors text-sm"
        >
          تعطيل الكل
        </button>
      </div>

      {/* Actions */}
      <div className="flex justify-end pt-4 border-t border-border">
        <button
          onClick={handleSave}
          disabled={saving}
          className={clsx(
            'px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors',
            saving && 'opacity-50 cursor-not-allowed'
          )}
        >
          {saving ? 'جاري الحفظ...' : 'حفظ التغييرات'}
        </button>
      </div>
    </div>
  )
}
