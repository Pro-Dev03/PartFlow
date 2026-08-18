import { useState } from 'react'
import { clsx } from 'clsx'
import type { OrganizationSettings } from '../types/settings'

export function OrganizationSettingsForm() {
  const [settings, setSettings] = useState<Partial<OrganizationSettings>>({
    name: '',
    type: 'computer_store',
    currency: 'ILS',
    language: 'ar',
    timezone: 'Asia/Jerusalem',
  })
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)

  // TODO: Fetch current settings from API

  const handleSave = async () => {
    setSaving(true)
    try {
      // TODO: Save settings to API
      console.log('Saving settings:', settings)
    } catch (error) {
      console.error('Failed to save settings:', error)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="max-w-2xl space-y-6">
      {/* Basic Information */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-text">معلومات المؤسسة</h3>
        
        <div>
          <label className="block text-sm font-medium text-text mb-2">اسم المحل</label>
          <input
            type="text"
            value={settings.name}
            onChange={(e) => setSettings({ ...settings, name: e.target.value })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            placeholder="مثال: محل الحاسوب الذكي"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text mb-2">نوع النشاط</label>
          <select
            value={settings.type}
            onChange={(e) => setSettings({ ...settings, type: e.target.value as any })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="computer_store">محل قطع حاسوب</option>
            <option value="electronics">إلكترونيات</option>
            <option value="repair">صيانة</option>
            <option value="trading">تجارة</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-text mb-2">العنوان</label>
          <input
            type="text"
            value={settings.address || ''}
            onChange={(e) => setSettings({ ...settings, address: e.target.value })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            placeholder="العنوان الكامل"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-text mb-2">الهاتف</label>
            <input
              type="tel"
              value={settings.phone || ''}
              onChange={(e) => setSettings({ ...settings, phone: e.target.value })}
              className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              placeholder="05X-XXXXXXX"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-text mb-2">البريد الإلكتروني</label>
            <input
              type="email"
              value={settings.email || ''}
              onChange={(e) => setSettings({ ...settings, email: e.target.value })}
              className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              placeholder="email@example.com"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-text mb-2">الرقم الضريبي</label>
          <input
            type="text"
            value={settings.taxId || ''}
            onChange={(e) => setSettings({ ...settings, taxId: e.target.value })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            placeholder="الرقم الضريبي (اختياري)"
          />
        </div>
      </div>

      {/* Localization */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-text">الإعدادات المحلية</h3>
        
        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-text mb-2">العملة</label>
            <select
              value={settings.currency}
              onChange={(e) => setSettings({ ...settings, currency: e.target.value as any })}
              className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value="ILS">₪ ILS</option>
              <option value="USD">$ USD</option>
              <option value="EUR">€ EUR</option>
              <option value="GBP">£ GBP</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-text mb-2">اللغة</label>
            <select
              value={settings.language}
              onChange={(e) => setSettings({ ...settings, language: e.target.value as any })}
              className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value="ar">العربية</option>
              <option value="he">עברית</option>
              <option value="en">English</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-text mb-2">التوقيت</label>
            <select
              value={settings.timezone}
              onChange={(e) => setSettings({ ...settings, timezone: e.target.value })}
              className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value="Asia/Jerusalem">Asia/Jerusalem</option>
              <option value="UTC">UTC</option>
              <option value="America/New_York">America/New_York</option>
              <option value="Europe/London">Europe/London</option>
            </select>
          </div>
        </div>
      </div>

      {/* Business Hours */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-text">ساعات العمل</h3>
        
        <div className="space-y-3">
          {[
            { key: 'sunday', label: 'الأحد' },
            { key: 'monday', label: 'الاثنين' },
            { key: 'tuesday', label: 'الثلاثاء' },
            { key: 'wednesday', label: 'الأربعاء' },
            { key: 'thursday', label: 'الخميس' },
            { key: 'friday', label: 'الجمعة' },
            { key: 'saturday', label: 'السبت' },
          ].map((day) => (
            <div key={day.key} className="flex items-center gap-4">
              <span className="w-24 text-sm text-text">{day.label}</span>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={!settings.businessHours?.[day.key as keyof typeof settings.businessHours]?.closed}
                  onChange={(e) => {
                    setSettings({
                      ...settings,
                      businessHours: {
                        ...settings.businessHours,
                        [day.key]: {
                          ...settings.businessHours?.[day.key as keyof typeof settings.businessHours],
                          closed: !e.target.checked,
                        },
                      },
                    })
                  }}
                  className="rounded border-border"
                />
                <span className="text-sm text-muted">مفتوح</span>
              </label>
              <input
                type="time"
                value={settings.businessHours?.[day.key as keyof typeof settings.businessHours]?.open || '09:00'}
                onChange={(e) => {
                  setSettings({
                    ...settings,
                    businessHours: {
                      ...settings.businessHours,
                      [day.key]: {
                        ...settings.businessHours?.[day.key as keyof typeof settings.businessHours],
                        open: e.target.value,
                      },
                    },
                  })
                }}
                className="px-3 py-1 border border-border rounded-lg text-sm"
                disabled={settings.businessHours?.[day.key as keyof typeof settings.businessHours]?.closed}
              />
              <span className="text-muted">-</span>
              <input
                type="time"
                value={settings.businessHours?.[day.key as keyof typeof settings.businessHours]?.close || '18:00'}
                onChange={(e) => {
                  setSettings({
                    ...settings,
                    businessHours: {
                      ...settings.businessHours,
                      [day.key]: {
                        ...settings.businessHours?.[day.key as keyof typeof settings.businessHours],
                        close: e.target.value,
                      },
                    },
                  })
                }}
                className="px-3 py-1 border border-border rounded-lg text-sm"
                disabled={settings.businessHours?.[day.key as keyof typeof settings.businessHours]?.closed}
              />
            </div>
          ))}
        </div>
      </div>

      {/* Actions */}
      <div className="flex justify-end gap-3 pt-4 border-t border-border">
        <button
          onClick={() => {/* TODO: Reset to original */}}
          className="px-4 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors"
        >
          إلغاء
        </button>
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
