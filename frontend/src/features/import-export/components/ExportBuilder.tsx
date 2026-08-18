import { useState } from 'react'
import { clsx } from 'clsx'
import type { ExportJob } from '../types/import-export'

export function ExportBuilder() {
  const [exportType, setExportType] = useState<'products' | 'customers' | 'suppliers' | 'sales' | 'purchases' | 'inventory' | 'reports'>('products')
  const [format, setFormat] = useState<'csv' | 'excel' | 'pdf'>('csv')
  const [filters, setFilters] = useState<Record<string, any>>({})
  const [dateRange, setDateRange] = useState({ from: '', to: '' })
  const [selectedFields, setSelectedFields] = useState<string[]>([])
  const [exporting, setExporting] = useState(false)

  const exportTypes = [
    { value: 'products', label: 'المنتجات', icon: '📦' },
    { value: 'customers', label: 'العملاء', icon: '👥' },
    { value: 'suppliers', label: 'الموردين', icon: '🚚' },
    { value: 'sales', label: 'المبيعات', icon: '💰' },
    { value: 'purchases', label: 'المشتريات', icon: '🛒' },
    { value: 'inventory', label: 'المخزون', icon: '📋' },
    { value: 'reports', label: 'التقارير', icon: '📊' },
  ]

  const formats = [
    { value: 'csv', label: 'CSV', icon: '📄' },
    { value: 'excel', label: 'Excel', icon: '📊' },
    { value: 'pdf', label: 'PDF', icon: '📕' },
  ]

  const availableFields = {
    products: ['الاسم', 'الباركود', 'السعر', 'التكلفة', 'المخزون', 'الفئة', 'المورد', 'الحالة'],
    customers: ['الاسم', 'الهاتف', 'البريد', 'العنوان', 'إجمالي المشتريات', 'الرصيد'],
    suppliers: ['الاسم', 'الهاتف', 'البريد', 'العنوان', 'إجمالي المشتريات', 'الرصيد'],
    sales: ['رقم البيع', 'التاريخ', 'العميل', 'الإجمالي', 'طريقة الدفع', 'الحالة'],
    purchases: ['رقم المشتراة', 'التاريخ', 'المورد', 'الإجمالي', 'الحالة'],
    inventory: ['المنتج', 'الكمية', 'الموقع', 'الحالة', 'آخر تحديث'],
    reports: ['التاريخ', 'النوع', 'الإجمالي', 'التفاصيل'],
  }

  const handleFieldToggle = (field: string) => {
    setSelectedFields(prev =>
      prev.includes(field)
        ? prev.filter(f => f !== field)
        : [...prev, field]
    )
  }

  const handleSelectAllFields = () => {
    setSelectedFields(availableFields[exportType])
  }

  const handleClearAllFields = () => {
    setSelectedFields([])
  }

  const handleExport = async () => {
    setExporting(true)
    try {
      // TODO: Create export job
      const exportData: Partial<ExportJob> = {
        type: exportType,
        format,
        filters: {
          ...filters,
          dateRange,
        },
        status: 'pending',
        fileName: `${exportType}_${new Date().toISOString().split('T')[0]}.${format}`,
      }
      console.log('Export data:', exportData)
    } catch (error) {
      console.error('Export failed:', error)
    } finally {
      setExporting(false)
    }
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div>
        <h2 className="text-xl font-semibold text-text mb-2">تصدير البيانات</h2>
        <p className="text-muted">اختر البيانات التي تريد تصديرها</p>
      </div>

      {/* Export Type */}
      <div>
        <label className="block text-sm font-medium text-text mb-3">نوع البيانات</label>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {exportTypes.map((type) => (
            <button
              key={type.value}
              onClick={() => {
                setExportType(type.value as any)
                setSelectedFields([])
              }}
              className={clsx(
                'p-3 rounded-lg border-2 transition-colors text-center',
                exportType === type.value
                  ? 'border-primary bg-primary-5'
                  : 'border-border hover:border-primary-50'
              )}
            >
              <div className="text-2xl mb-1">{type.icon}</div>
              <div className="text-sm font-medium text-text">{type.label}</div>
            </button>
          ))}
        </div>
      </div>

      {/* Format */}
      <div>
        <label className="block text-sm font-medium text-text mb-3">الصيغة</label>
        <div className="flex gap-3">
          {formats.map((fmt) => (
            <button
              key={fmt.value}
              onClick={() => setFormat(fmt.value as any)}
              className={clsx(
                'flex-1 p-3 rounded-lg border-2 transition-colors text-center',
                format === fmt.value
                  ? 'border-primary bg-primary-5'
                  : 'border-border hover:border-primary-50'
              )}
            >
              <div className="text-2xl mb-1">{fmt.icon}</div>
              <div className="text-sm font-medium text-text">{fmt.label}</div>
            </button>
          ))}
        </div>
      </div>

      {/* Date Range */}
      <div>
        <label className="block text-sm font-medium text-text mb-3">نطاق التاريخ (اختياري)</label>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm text-muted mb-1">من</label>
            <input
              type="date"
              value={dateRange.from}
              onChange={(e) => setDateRange({ ...dateRange, from: e.target.value })}
              className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            />
          </div>
          <div>
            <label className="block text-sm text-muted mb-1">إلى</label>
            <input
              type="date"
              value={dateRange.to}
              onChange={(e) => setDateRange({ ...dateRange, to: e.target.value })}
              className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            />
          </div>
        </div>
      </div>

      {/* Field Selection */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <label className="block text-sm font-medium text-text">الحقول المطلوبة</label>
          <div className="flex gap-2">
            <button
              onClick={handleSelectAllFields}
              className="text-sm text-primary hover:text-primary-600 underline"
            >
              تحديد الكل
            </button>
            <button
              onClick={handleClearAllFields}
              className="text-sm text-muted hover:text-text underline"
            >
              إلغاء الكل
            </button>
          </div>
        </div>
        <div className="bg-surface rounded-lg border border-border p-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
            {availableFields[exportType].map((field) => (
              <label key={field} className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedFields.includes(field)}
                  onChange={() => handleFieldToggle(field)}
                  className="rounded border-border"
                />
                <span className="text-sm text-text">{field}</span>
              </label>
            ))}
          </div>
        </div>
      </div>

      {/* Export Button */}
      <div className="flex justify-end">
        <button
          onClick={handleExport}
          disabled={exporting || selectedFields.length === 0}
          className={clsx(
            'px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors',
            (exporting || selectedFields.length === 0) && 'opacity-50 cursor-not-allowed'
          )}
        >
          {exporting ? 'جاري التصدير...' : 'تصدير'}
        </button>
      </div>
    </div>
  )
}
