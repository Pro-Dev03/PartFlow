import { useState } from 'react'
import { clsx } from 'clsx'

interface UserStepProps {
  data: {
    name: string
    email: string
    password: string
  }
  onChange: (updates: any) => void
  onNext: () => void
  onBack: () => void
}

export function UserStep({ data, onChange, onNext, onBack }: UserStepProps) {
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [showPassword, setShowPassword] = useState(false)

  const validate = () => {
    const newErrors: Record<string, string> = {}

    if (!data.name.trim()) {
      newErrors.name = 'الاسم مطلوب'
    }

    if (!data.email.trim()) {
      newErrors.email = 'البريد الإلكتروني مطلوب'
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(data.email)) {
      newErrors.email = 'البريد الإلكتروني غير صالح'
    }

    if (!data.password) {
      newErrors.password = 'كلمة المرور مطلوبة'
    } else if (data.password.length < 8) {
      newErrors.password = 'كلمة المرور يجب أن تكون 8 أحرف على الأقل'
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
        <h2 className="text-xl font-semibold text-text mb-2">إنشاء حساب المدير</h2>
        <p className="text-muted">أنشئ حسابك للوصول إلى النظام</p>
      </div>

      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-text mb-2">الاسم الكامل *</label>
          <input
            type="text"
            value={data.name}
            onChange={(e) => onChange({ name: e.target.value })}
            className={clsx(
              'w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent',
              errors.name ? 'border-danger' : 'border-border'
            )}
            placeholder="الاسم الكامل"
          />
          {errors.name && <p className="text-sm text-danger mt-1">{errors.name}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-text mb-2">البريد الإلكتروني *</label>
          <input
            type="email"
            value={data.email}
            onChange={(e) => onChange({ email: e.target.value })}
            className={clsx(
              'w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent',
              errors.email ? 'border-danger' : 'border-border'
            )}
            placeholder="email@example.com"
          />
          {errors.email && <p className="text-sm text-danger mt-1">{errors.email}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-text mb-2">كلمة المرور *</label>
          <div className="relative">
            <input
              type={showPassword ? 'text' : 'password'}
              value={data.password}
              onChange={(e) => onChange({ password: e.target.value })}
              className={clsx(
                'w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent pr-10',
                errors.password ? 'border-danger' : 'border-border'
              )}
              placeholder="كلمة المرور"
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-muted hover:text-text"
            >
              {showPassword ? '🙈' : '👁️'}
            </button>
          </div>
          {errors.password && <p className="text-sm text-danger mt-1">{errors.password}</p>}
          <p className="text-xs text-muted mt-1">يجب أن تكون 8 أحرف على الأقل</p>
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
