import { useState } from 'react'
import { clsx } from 'clsx'

interface OrganizationStepProps {
  data: {
    name: string
    type: 'computer_store' | 'electronics' | 'repair' | 'trading'
    currency: 'ILS' | 'USD' | 'EUR' | 'GBP'
    language: 'ar' | 'he' | 'en'
  }
  onChange: (updates: any) => void
  onNext: () => void
}

export function OrganizationStep({ data, onChange, onNext }: OrganizationStepProps) {
  const [errors, setErrors] = useState<Record<string, string>>({})

  const validate = () => {
    const newErrors: Record<string, string> = {}

    if (!data.name.trim()) {
      newErrors.name = 'اسم المحل مطلوب'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleNext = () => {
    if (validate()) {
      onNext()
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold text-text mb-2">معلومات المؤسسة</h2>
        <p className="text-muted">أخبرنا عن متجرك لنقوم بتخصيص النظام لك</p>
      </div>

      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-text mb-2">اسم المحل *</label>
          <input
            type="text"
            value={data.name}
            onChange={(e) => onChange({ name: e.target.value })}
            className={clsx(
              'w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent',
              errors.name ? 'border-danger' : 'border-border'
            )}
            placeholder="مثال: محل الحاسوب الذكي"
          />
          {errors.name && <p className="text-sm text-danger mt-1">{errors.name}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-text mb-2">نوع النشاط</label>
          <select
            value={data.type}
            onChange={(e) => onChange({ type: e.target.value })}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="computer_store">محل قطع حاسوب</option>
            <option value="electronics">إلكترونيات</option>
            <option value="repair">صيانة</option>
            <option value="trading">تجارة</option>
          </select>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-text mb-2">العملة</label>
            <select
              value={data.currency}
              onChange={(e) => onChange({ currency: e.target.value })}
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
              value={data.language}
              onChange={(e) => onChange({ language: e.target.value })}
              className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value="ar">العربية</option>
              <option value="he">עברית</option>
              <option value="en">English</option>
            </select>
          </div>
        </div>
      </div>

      <div className="flex justify-end">
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
