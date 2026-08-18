import { useState } from 'react'
import { clsx } from 'clsx'
import { FormField, FormSection, FormActions } from '../../../components/forms'

function CollapsibleSection({ label, optional, children }: { label: string; optional?: boolean; children: React.ReactNode }) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div className="border border-border dark:border-border-dark rounded-lg overflow-hidden">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-4 py-3 flex items-center justify-between bg-surface dark:bg-surface-dark hover:bg-surface-elevated dark:hover:bg-surface-elevated transition-colors"
      >
        <span className="font-medium text-text dark:text-text-dark">
          {label}
          {optional && <span className="text-text-secondary dark:text-text-dim text-sm mr-2">(اختياري)</span>}
        </span>
        <span className="text-text-secondary dark:text-text-dim">
          {isOpen ? '▼' : '▶'}
        </span>
      </button>
      {isOpen && (
        <div className="p-4">
          {children}
        </div>
      )}
    </div>
  )
}

interface InspectionData {
  productId: string
  inspector: string
  date: string
  powerTest: 'passed' | 'failed' | 'skipped'
  temperatureTest: 'passed' | 'failed' | 'skipped'
  performanceTest: 'passed' | 'failed' | 'skipped'
  portsTest: 'passed' | 'failed' | 'skipped'
  visualInspection: 'passed' | 'failed' | 'skipped'
  serialVerification: 'passed' | 'failed' | 'skipped'
  overallResult: 'passed' | 'failed'
  notes?: string
  photos?: string[]
}

interface InspectionFormProps {
  onSubmit: (data: InspectionData) => void
  onCancel?: () => void
  initialData?: Partial<InspectionData>
  className?: string
  productId?: string
}

export function InspectionForm({ onSubmit, onCancel, initialData, className, productId }: InspectionFormProps) {
  const [formData, setFormData] = useState<InspectionData>({
    productId: productId || initialData?.productId || '',
    inspector: initialData?.inspector || '',
    date: initialData?.date || new Date().toISOString().split('T')[0],
    powerTest: initialData?.powerTest || 'skipped',
    temperatureTest: initialData?.temperatureTest || 'skipped',
    performanceTest: initialData?.performanceTest || 'skipped',
    portsTest: initialData?.portsTest || 'skipped',
    visualInspection: initialData?.visualInspection || 'skipped',
    serialVerification: initialData?.serialVerification || 'skipped',
    overallResult: initialData?.overallResult || 'passed',
    notes: initialData?.notes || '',
    photos: initialData?.photos || [],
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit(formData)
  }

  const testOptions = [
    { value: 'passed', label: 'نجح', color: 'bg-success text-white' },
    { value: 'failed', label: 'فشل', color: 'bg-danger text-white' },
    { value: 'skipped', label: 'تخطي', color: 'bg-muted text-text' },
  ] as const

  return (
    <form onSubmit={handleSubmit} className={clsx('space-y-6', className)}>
      <FormSection title="معلومات الفحص" description="بيانات أساسية عن عملية الفحص">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormField label="اسم المفتش" required>
            <input
              type="text"
              value={formData.inspector}
              onChange={(e) => setFormData({ ...formData, inspector: e.target.value })}
              className="w-full px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              required
            />
          </FormField>
          <FormField label="تاريخ الفحص" required>
            <input
              type="date"
              value={formData.date}
              onChange={(e) => setFormData({ ...formData, date: e.target.value })}
              className="w-full px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              required
            />
          </FormField>
        </div>
      </FormSection>

      <FormSection title="الاختبارات الفنية" description="نتائج الاختبارات المختلفة">
        <div className="space-y-4">
          <TestRow
            label="اختبار الطاقة"
            value={formData.powerTest}
            onChange={(value) => setFormData({ ...formData, powerTest: value as any })}
            options={testOptions}
          />
          <TestRow
            label="اختبار الحرارة"
            value={formData.temperatureTest}
            onChange={(value) => setFormData({ ...formData, temperatureTest: value as any })}
            options={testOptions}
          />
          <TestRow
            label="اختبار الأداء"
            value={formData.performanceTest}
            onChange={(value) => setFormData({ ...formData, performanceTest: value as any })}
            options={testOptions}
          />
          <TestRow
            label="اختبار المنافذ"
            value={formData.portsTest}
            onChange={(value) => setFormData({ ...formData, portsTest: value as any })}
            options={testOptions}
          />
        </div>
      </FormSection>

      <CollapsibleSection label="الفحص البصري" optional>
        <TestRow
          label="الفحص البصري"
          value={formData.visualInspection}
          onChange={(value) => setFormData({ ...formData, visualInspection: value as any })}
          options={testOptions}
        />
      </CollapsibleSection>

      <CollapsibleSection label="التحقق من الرقم التسلسلي" optional>
        <TestRow
          label="التحقق من الرقم التسلسلي"
          value={formData.serialVerification}
          onChange={(value) => setFormData({ ...formData, serialVerification: value as any })}
          options={testOptions}
        />
      </CollapsibleSection>

      <FormSection title="النتيجة النهائية">
        <div className="flex gap-4">
          {(['passed', 'failed'] as const).map((result) => (
            <label key={result} className="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                name="overallResult"
                value={result}
                checked={formData.overallResult === result}
                onChange={(e) => setFormData({ ...formData, overallResult: e.target.value as any })}
                className="w-4 h-4 text-primary"
              />
              <span
                className={clsx(
                  'px-3 py-1 rounded-lg text-sm font-medium',
                  result === 'passed' ? 'bg-success text-white' : 'bg-danger text-white'
                )}
              >
                {result === 'passed' ? 'نجح' : 'فشل'}
              </span>
            </label>
          ))}
        </div>
      </FormSection>

      <ProgressiveDisclosure label="ملاحظات إضافية" optional>
        <FormField label="ملاحظات">
          <textarea
            value={formData.notes}
            onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
            rows={4}
            className="w-full px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            placeholder="أضف أي ملاحظات عن حالة المنتج..."
          />
        </FormField>
      </ProgressiveDisclosure>

      <FormActions>
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 border border-border rounded-lg hover:bg-surface-80 transition-colors"
          >
            إلغاء
          </button>
        )}
        <button
          type="submit"
          className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
        >
          حفظ الفحص
        </button>
      </FormActions>
    </form>
  )
}

function TestRow({
  label,
  value,
  onChange,
  options,
}: {
  label: string
  value: string
  onChange: (value: string) => void
  options: readonly { value: string; label: string; color: string }[]
}) {
  return (
    <div className="flex items-center justify-between py-3 border-b border-border">
      <span className="font-medium text-text">{label}</span>
      <div className="flex gap-2">
        {options.map((option) => (
          <button
            key={option.value}
            type="button"
            onClick={() => onChange(option.value)}
            className={clsx(
              'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
              value === option.value ? option.color : 'bg-muted text-muted hover:bg-muted-80'
            )}
          >
            {option.label}
          </button>
        ))}
      </div>
    </div>
  )
}
