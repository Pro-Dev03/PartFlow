import { clsx } from 'clsx'
import type { OnboardingData } from '../types/onboarding'

interface CompletionStepProps {
  data: Partial<OnboardingData>
  onComplete: () => void
  onBack: () => void
}

export function CompletionStep({ data, onComplete, onBack }: CompletionStepProps) {
  const typeLabels = {
    computer_store: 'محل قطع حاسوب',
    electronics: 'إلكترونيات',
    repair: 'صيانة',
    trading: 'تجارة',
  }

  const currencyLabels = {
    ILS: '₪ ILS',
    USD: '$ USD',
    EUR: '€ EUR',
    GBP: '£ GBP',
  }

  const languageLabels = {
    ar: 'العربية',
    he: 'עברית',
    en: 'English',
  }

  return (
    <div className="space-y-6">
      <div className="text-center">
        <div className="text-6xl mb-4">🎉</div>
        <h2 className="text-xl font-semibold text-text mb-2">ممتاز! متجرك جاهز</h2>
        <p className="text-muted">راجع المعلومات ثم ابدأ استخدام النظام</p>
      </div>

      {/* Summary */}
      <div className="bg-muted-10 rounded-lg p-6 space-y-4">
        <h3 className="font-medium text-text">ملخص الإعدادات</h3>

        <div className="space-y-3">
          <div className="flex justify-between">
            <span className="text-muted">اسم المحل:</span>
            <span className="font-medium text-text">{data.organization?.name}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted">نوع النشاط:</span>
            <span className="font-medium text-text">
              {data.organization?.type && typeLabels[data.organization.type]}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted">العملة:</span>
            <span className="font-medium text-text">
              {data.organization?.currency && currencyLabels[data.organization.currency]}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted">اللغة:</span>
            <span className="font-medium text-text">
              {data.organization?.language && languageLabels[data.organization.language]}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted">اسم المدير:</span>
            <span className="font-medium text-text">{data.user?.name}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted">البريد الإلكتروني:</span>
            <span className="font-medium text-text">{data.user?.email}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted">الحد الأدنى للمخزون:</span>
            <span className="font-medium text-text">{data.preferences?.lowStockThreshold}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted">مدة الضمان الافتراضية:</span>
            <span className="font-medium text-text">{data.preferences?.defaultWarrantyDays} يوم</span>
          </div>
        </div>
      </div>

      {/* Next Steps */}
      <div className="bg-primary-5 rounded-lg p-6">
        <h3 className="font-medium text-text mb-3">الخطوات التالية</h3>
        <ul className="space-y-2 text-sm text-muted">
          <li className="flex items-center gap-2">
            <span className="text-primary">✓</span>
            إضافة منتجاتك الأولى
          </li>
          <li className="flex items-center gap-2">
            <span className="text-primary">✓</span>
            إضافة الموردين
          </li>
          <li className="flex items-center gap-2">
            <span className="text-primary">✓</span>
            إعداد العملاء
          </li>
          <li className="flex items-center gap-2">
            <span className="text-primary">✓</span>
            بدء عمليات البيع
          </li>
        </ul>
      </div>

      {/* Actions */}
      <div className="flex justify-between">
        <button
          onClick={onBack}
          className="px-6 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors"
        >
          السابق
        </button>
        <button
          onClick={onComplete}
          className="px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
        >
          بدء استخدام النظام
        </button>
      </div>
    </div>
  )
}
