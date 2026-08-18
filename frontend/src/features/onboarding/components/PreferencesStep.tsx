import { useState } from 'react'
import { clsx } from 'clsx'

interface PreferencesStepProps {
  data: {
    businessHours: {
      sunday: { open: string; close: string; closed: boolean }
      monday: { open: string; close: string; closed: boolean }
      tuesday: { open: string; close: string; closed: boolean }
      wednesday: { open: string; close: string; closed: boolean }
      thursday: { open: string; close: string; closed: boolean }
      friday: { open: string; close: string; closed: boolean }
      saturday: { open: string; close: string; closed: boolean }
    }
    lowStockThreshold: number
    defaultWarrantyDays: number
  }
  onChange: (updates: any) => void
  onNext: () => void
  onBack: () => void
}

export function PreferencesStep({ data, onChange, onNext, onBack }: PreferencesStepProps) {
  const [errors, setErrors] = useState<Record<string, string>>({})

  const validate = () => {
    const newErrors: Record<string, string> = {}

    if (data.lowStockThreshold < 0) {
      newErrors.lowStockThreshold = 'الحد الأدنى يجب أن يكون رقمًا موجبًا'
    }

    if (data.defaultWarrantyDays < 0) {
      newErrors.defaultWarrantyDays = 'مدة الضمان يجب أن تكون رقمًا موجبًا'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleNext = () => {
    if (validate()) {
      onNext()
    }
  }

  const days = [
    { key: 'sunday', label: 'الأحد' },
    { key: 'monday', label: 'الاثنين' },
    { key: 'tuesday', label: 'الثلاثاء' },
    { key: 'wednesday', label: 'الأربعاء' },
    { key: 'thursday', label: 'الخميس' },
    { key: 'friday', label: 'الجمعة' },
    { key: 'saturday', label: 'السبت' },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold text-text mb-2">التفضيلات الأساسية</h2>
        <p className="text-muted">ضبط الإعدادات الأساسية للنظام</p>
      </div>

      <div className="space-y-6">
        {/* Business Hours */}
        <div>
          <h3 className="font-medium text-text mb-4">ساعات العمل</h3>
          <div className="space-y-3">
            {days.map((day) => (
              <div key={day.key} className="flex items-center gap-4">
                <span className="w-20 text-sm text-text">{day.label}</span>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={!data.businessHours[day.key as keyof typeof data.businessHours].closed}
                    onChange={(e) => {
                      onChange({
                        businessHours: {
                          ...data.businessHours,
                          [day.key]: {
                            ...data.businessHours[day.key as keyof typeof data.businessHours],
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
                  value={data.businessHours[day.key as keyof typeof data.businessHours].open}
                  onChange={(e) => {
                    onChange({
                      businessHours: {
                        ...data.businessHours,
                        [day.key]: {
                          ...data.businessHours[day.key as keyof typeof data.businessHours],
                          open: e.target.value,
                        },
                      },
                    })
                  }}
                  className="px-3 py-1 border border-border rounded-lg text-sm"
                  disabled={data.businessHours[day.key as keyof typeof data.businessHours].closed}
                />
                <span className="text-muted">-</span>
                <input
                  type="time"
                  value={data.businessHours[day.key as keyof typeof data.businessHours].close}
                  onChange={(e) => {
                    onChange({
                      businessHours: {
                        ...data.businessHours,
                        [day.key]: {
                          ...data.businessHours[day.key as keyof typeof data.businessHours],
                          close: e.target.value,
                        },
                      },
                    })
                  }}
                  className="px-3 py-1 border border-border rounded-lg text-sm"
                  disabled={data.businessHours[day.key as keyof typeof data.businessHours].closed}
                />
              </div>
            ))}
          </div>
        </div>

        {/* Inventory Settings */}
        <div>
          <h3 className="font-medium text-text mb-4">إعدادات المخزون</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-text mb-2">الحد الأدنى للمخزون</label>
              <input
                type="number"
                min="0"
                value={data.lowStockThreshold}
                onChange={(e) => onChange({ lowStockThreshold: parseInt(e.target.value) })}
                className={clsx(
                  'w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent',
                  errors.lowStockThreshold ? 'border-danger' : 'border-border'
                )}
              />
              {errors.lowStockThreshold && (
                <p className="text-sm text-danger mt-1">{errors.lowStockThreshold}</p>
              )}
              <p className="text-xs text-muted mt-1">سيتم إرسال إشعار عند الوصول لهذا الحد</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-text mb-2">مدة الضمان الافتراضية (أيام)</label>
              <input
                type="number"
                min="0"
                value={data.defaultWarrantyDays}
                onChange={(e) => onChange({ defaultWarrantyDays: parseInt(e.target.value) })}
                className={clsx(
                  'w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent',
                  errors.defaultWarrantyDays ? 'border-danger' : 'border-border'
                )}
              />
              {errors.defaultWarrantyDays && (
                <p className="text-sm text-danger mt-1">{errors.defaultWarrantyDays}</p>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="flex justify-between">
        <button
          onClick={onBack}
          className="px-6 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors"
        >
          السابق
        </button>
        <button
          onClick={handleNext}
          className="px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
        >
          التالي
        </button>
      </div>
    </div>
  )
}
