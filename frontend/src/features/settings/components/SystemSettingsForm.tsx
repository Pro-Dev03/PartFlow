import { useState } from 'react'
import { clsx } from 'clsx'
import type { SystemSettings } from '../types/settings'

export function SystemSettingsForm() {
  const [settings, setSettings] = useState<Partial<SystemSettings>>({
    lowStockThreshold: 5,
    defaultWarrantyDays: 30,
    allowNegativeStock: false,
    requireBarcode: false,
    autoBackup: true,
    backupFrequency: 'daily',
    retentionDays: 90,
    barcodePrefix: 'FNX',
    priceRounding: 'nearest',
    decimalPlaces: 2,
  })
  const [saving, setSaving] = useState(false)

  // TODO: Fetch current system settings from API

  const handleSave = async () => {
    setSaving(true)
    try {
      // TODO: Save settings to API
      console.log('Saving system settings:', settings)
    } catch (error) {
      console.error('Failed to save settings:', error)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="max-w-2xl space-y-6">
      {/* Header */}
      <div>
        <h3 className="text-lg font-semibold text-text">إعدادات النظام</h3>
        <p className="text-sm text-muted">ضبط سلوك وقواعد النظام</p>
      </div>

      {/* Inventory Settings */}
      <div className="space-y-4">
        <h4 className="font-medium text-text">إعدادات المخزون</h4>
        
        <div>
          <label className="block text-sm font-medium text-text mb-2">الحد الأدنى للمخزون</label>
          <input
            type="number"
            min="0"
            value={settings.lowStockThreshold}
            onChange={(e) => setSettings({ ...settings, lowStockThreshold: parseInt(e.target.value) })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
          <p className="text-sm text-muted mt-1">سيتم إرسال إشعار عندما يصل المخزون لهذا الحد</p>
        </div>

        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="allowNegativeStock"
            checked={settings.allowNegativeStock}
            onChange={(e) => setSettings({ ...settings, allowNegativeStock: e.target.checked })}
            className="rounded border-border"
          />
          <label htmlFor="allowNegativeStock" className="text-sm text-text">
            السماح بمخزون سالب
          </label>
        </div>
        <p className="text-sm text-muted mr-7">السماح ببيع منتجات حتى لو لم يكن متوفرًا في المخزون</p>

        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="requireBarcode"
            checked={settings.requireBarcode}
            onChange={(e) => setSettings({ ...settings, requireBarcode: e.target.checked })}
            className="rounded border-border"
          />
          <label htmlFor="requireBarcode" className="text-sm text-text">
            إلزام Barcode
          </label>
        </div>
        <p className="text-sm text-muted mr-7">يتطلب وجود Barcode لكل منتج</p>
      </div>

      {/* Warranty Settings */}
      <div className="space-y-4">
        <h4 className="font-medium text-text">إعدادات الضمان</h4>
        
        <div>
          <label className="block text-sm font-medium text-text mb-2">مدة الضمان الافتراضية (أيام)</label>
          <input
            type="number"
            min="0"
            value={settings.defaultWarrantyDays}
            onChange={(e) => setSettings({ ...settings, defaultWarrantyDays: parseInt(e.target.value) })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
        </div>
      </div>

      {/* Barcode Settings */}
      <div className="space-y-4">
        <h4 className="font-medium text-text">إعدادات Barcode</h4>
        
        <div>
          <label className="block text-sm font-medium text-text mb-2">بادئة Barcode</label>
          <input
            type="text"
            value={settings.barcodePrefix}
            onChange={(e) => setSettings({ ...settings, barcodePrefix: e.target.value })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            placeholder="مثال: FNX"
          />
          <p className="text-sm text-muted mt-1">البادئة المستخدمة لتوليد Barcodes داخليًا</p>
        </div>
      </div>

      {/* Pricing Settings */}
      <div className="space-y-4">
        <h4 className="font-medium text-text">إعدادات التسعير</h4>
        
        <div>
          <label className="block text-sm font-medium text-text mb-2">تقريب الأسعار</label>
          <select
            value={settings.priceRounding}
            onChange={(e) => setSettings({ ...settings, priceRounding: e.target.value as any })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="none">بدون تقريب</option>
            <option value="nearest">أقرب رقم</option>
            <option value="up">للأعلى</option>
            <option value="down">للأسفل</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-text mb-2">عدد المنازل العشرية</label>
          <input
            type="number"
            min="0"
            max="4"
            value={settings.decimalPlaces}
            onChange={(e) => setSettings({ ...settings, decimalPlaces: parseInt(e.target.value) })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
        </div>
      </div>

      {/* Backup Settings */}
      <div className="space-y-4">
        <h4 className="font-medium text-text">إعدادات النسخ الاحتياطي</h4>
        
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="autoBackup"
            checked={settings.autoBackup}
            onChange={(e) => setSettings({ ...settings, autoBackup: e.target.checked })}
            className="rounded border-border"
          />
          <label htmlFor="autoBackup" className="text-sm text-text">
            نسخ احتياطي تلقائي
          </label>
        </div>

        <div>
          <label className="block text-sm font-medium text-text mb-2">تكرار النسخ</label>
          <select
            value={settings.backupFrequency}
            onChange={(e) => setSettings({ ...settings, backupFrequency: e.target.value as any })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            disabled={!settings.autoBackup}
          >
            <option value="daily">يومي</option>
            <option value="weekly">أسبوعي</option>
            <option value="monthly">شهري</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-text mb-2">فترة الاحتفاظ (أيام)</label>
          <input
            type="number"
            min="1"
            value={settings.retentionDays}
            onChange={(e) => setSettings({ ...settings, retentionDays: parseInt(e.target.value) })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            disabled={!settings.autoBackup}
          />
          <p className="text-sm text-muted mt-1">سيتم حذف النسخ القديمة بعد هذه الفترة</p>
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
