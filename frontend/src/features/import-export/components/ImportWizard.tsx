import { useState } from 'react'
import { clsx } from 'clsx'
import type { ImportJob, ImportPreview, FieldMapping } from '../types/import-export'

type ImportStep = 'upload' | 'preview' | 'mapping' | 'confirm' | 'processing' | 'complete'

export function ImportWizard() {
  const [currentStep, setCurrentStep] = useState<ImportStep>('upload')
  const [importType, setImportType] = useState<'products' | 'customers' | 'suppliers' | 'inventory'>('products')
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<ImportPreview | null>(null)
  const [fieldMapping, setFieldMapping] = useState<FieldMapping[]>([])
  const [importJob, setImportJob] = useState<ImportJob | null>(null)

  const importTypes = [
    { value: 'products', label: 'المنتجات', icon: '📦' },
    { value: 'customers', label: 'العملاء', icon: '👥' },
    { value: 'suppliers', label: 'الموردين', icon: '🚚' },
    { value: 'inventory', label: 'المخزون', icon: '📋' },
  ]

  const handleFileUpload = async (uploadedFile: File) => {
    setFile(uploadedFile)
    // TODO: Process file and generate preview
    setCurrentStep('preview')
  }

  const handleMappingConfirm = () => {
    setCurrentStep('confirm')
  }

  const handleImportStart = async () => {
    setCurrentStep('processing')
    try {
      // TODO: Start import job
      console.log('Starting import:', { importType, file, fieldMapping })
    } catch (error) {
      console.error('Import failed:', error)
      setCurrentStep('upload')
    }
  }

  const handleImportComplete = () => {
    setCurrentStep('complete')
  }

  const handleReset = () => {
    setCurrentStep('upload')
    setFile(null)
    setPreview(null)
    setFieldMapping([])
    setImportJob(null)
  }

  return (
    <div className="max-w-4xl mx-auto">
      {/* Step Indicator */}
      <div className="flex items-center justify-center mb-8">
        {['upload', 'preview', 'mapping', 'confirm', 'processing', 'complete'].map((step, index) => (
          <div key={step} className="flex items-center">
            <div
              className={clsx(
                'w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium transition-colors',
                currentStep === step
                  ? 'bg-primary text-white'
                  : ['upload', 'preview', 'mapping', 'confirm', 'processing', 'complete'].indexOf(currentStep) > index
                  ? 'bg-success text-white'
                  : 'bg-muted text-muted'
              )}
            >
              {['upload', 'preview', 'mapping', 'confirm', 'processing', 'complete'].indexOf(currentStep) > index ? '✓' : index + 1}
            </div>
            {index < 5 && (
              <div
                className={clsx(
                  'w-12 h-1 mx-2',
                  ['upload', 'preview', 'mapping', 'confirm', 'processing', 'complete'].indexOf(currentStep) > index ? 'bg-success' : 'bg-muted'
                )}
              />
            )}
          </div>
        ))}
      </div>

      {/* Step Content */}
      {currentStep === 'upload' && (
        <div className="space-y-6">
          <div>
            <h2 className="text-xl font-semibold text-text mb-2">اختر نوع البيانات</h2>
            <p className="text-muted">حدد نوع البيانات التي تريد استيرادها</p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {importTypes.map((type) => (
              <button
                key={type.value}
                onClick={() => setImportType(type.value as any)}
                className={clsx(
                  'p-4 rounded-lg border-2 transition-colors text-center',
                  importType === type.value
                    ? 'border-primary bg-primary-5'
                    : 'border-border hover:border-primary-50'
                )}
              >
                <div className="text-3xl mb-2">{type.icon}</div>
                <div className="font-medium text-text">{type.label}</div>
              </button>
            ))}
          </div>

          <div className="border-t border-border pt-6">
            <h3 className="font-medium text-text mb-4">رفع الملف</h3>
            <div className="border-2 border-dashed border-border rounded-lg p-8 text-center hover:border-primary transition-colors">
              <input
                type="file"
                accept=".csv,.xlsx,.xls"
                onChange={(e) => e.target.files?.[0] && handleFileUpload(e.target.files[0])}
                className="hidden"
                id="file-upload"
              />
              <label htmlFor="file-upload" className="cursor-pointer">
                <div className="text-4xl mb-4">📁</div>
                <p className="text-text mb-2">اسحب الملف هنا أو انقر للاختيار</p>
                <p className="text-sm text-muted">يدعم CSV و Excel (Max 10MB)</p>
              </label>
            </div>
          </div>
        </div>
      )}

      {currentStep === 'preview' && preview && (
        <div className="space-y-6">
          <div>
            <h2 className="text-xl font-semibold text-text mb-2">معاينة البيانات</h2>
            <p className="text-muted">راجع البيانات قبل الاستيراد</p>
          </div>

          <div className="bg-muted-10 rounded-lg p-4">
            <div className="grid grid-cols-3 gap-4 mb-4">
              <div>
                <p className="text-sm text-muted">إجمالي الصفوف</p>
                <p className="font-medium text-text">{preview.totalRows}</p>
              </div>
              <div>
                <p className="text-sm text-muted">الترميز</p>
                <p className="font-medium text-text">{preview.detectedEncoding}</p>
              </div>
              <div>
                <p className="text-sm text-muted">الفاصل</p>
                <p className="font-medium text-text">{preview.detectedDelimiter}</p>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border">
                    {preview.headers.map((header, index) => (
                      <th key={index} className="px-3 py-2 text-right font-medium text-text">
                        {header}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {preview.sampleRows.map((row, rowIndex) => (
                    <tr key={rowIndex} className="border-b border-border">
                      {row.map((cell, cellIndex) => (
                        <td key={cellIndex} className="px-3 py-2 text-muted">
                          {cell}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          <div className="flex justify-between">
            <button
              onClick={handleReset}
              className="px-4 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors"
            >
              إلغاء
            </button>
            <button
              onClick={() => setCurrentStep('mapping')}
              className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              التالي: تعيين الحقول
            </button>
          </div>
        </div>
      )}

      {currentStep === 'mapping' && (
        <div className="space-y-6">
          <div>
            <h2 className="text-xl font-semibold text-text mb-2">تعيين الحقول</h2>
            <p className="text-muted">طابق بين حقول الملف وحقول النظام</p>
          </div>

          <div className="bg-surface rounded-lg border border-border overflow-hidden">
            <table className="w-full">
              <thead className="bg-muted-10 border-b border-border">
                <tr>
                  <th className="px-4 py-3 text-right text-sm font-medium text-muted">حقل الملف</th>
                  <th className="px-4 py-3 text-right text-sm font-medium text-muted">حقل النظام</th>
                  <th className="px-4 py-3 text-right text-sm font-medium text-muted">مطلوب</th>
                </tr>
              </thead>
              <tbody>
                {fieldMapping.map((mapping, index) => (
                  <tr key={index} className="border-b border-border">
                    <td className="px-4 py-3 text-text">{mapping.sourceField}</td>
                    <td className="px-4 py-3">
                      <select className="w-full px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent">
                        <option value="">اختر حقل</option>
                        {/* TODO: Add target field options based on import type */}
                      </select>
                    </td>
                    <td className="px-4 py-3">
                      {mapping.required && (
                        <span className="text-danger text-sm">مطلوب</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="flex justify-between">
            <button
              onClick={() => setCurrentStep('preview')}
              className="px-4 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors"
            >
              السابق
            </button>
            <button
              onClick={handleMappingConfirm}
              className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              التالي: تأكيد
            </button>
          </div>
        </div>
      )}

      {currentStep === 'confirm' && (
        <div className="space-y-6">
          <div>
            <h2 className="text-xl font-semibold text-text mb-2">تأكيد الاستيراد</h2>
            <p className="text-muted">راجع ملخص الاستيراد قبل البدء</p>
          </div>

          <div className="bg-muted-10 rounded-lg p-6 space-y-4">
            <div className="flex justify-between">
              <span className="text-muted">نوع البيانات:</span>
              <span className="font-medium text-text">{importTypes.find(t => t.value === importType)?.label}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted">الملف:</span>
              <span className="font-medium text-text">{file?.name}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted">حجم الملف:</span>
              <span className="font-medium text-text">{(file?.size || 0) / 1024} KB</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted">عدد الصفوف:</span>
              <span className="font-medium text-text">{preview?.totalRows}</span>
            </div>
          </div>

          <div className="bg-warning-10 border border-warning-30 rounded-lg p-4">
            <p className="text-sm text-warning">
              ⚠️ سيتم إضافة البيانات الجديدة وتحديث البيانات الموجودة. تأكد من أخذ نسخة احتياطية قبل الاستمرار.
            </p>
          </div>

          <div className="flex justify-between">
            <button
              onClick={() => setCurrentStep('mapping')}
              className="px-4 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors"
            >
              السابق
            </button>
            <button
              onClick={handleImportStart}
              className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              بدء الاستيراد
            </button>
          </div>
        </div>
      )}

      {currentStep === 'processing' && (
        <div className="text-center py-12">
          <div className="text-6xl mb-4 animate-spin">⚙️</div>
          <h2 className="text-xl font-semibold text-text mb-2">جاري الاستيراد...</h2>
          <p className="text-muted">قد يستغرق هذا بضع دقائق</p>
        </div>
      )}

      {currentStep === 'complete' && (
        <div className="space-y-6 text-center">
          <div className="text-6xl mb-4">✅</div>
          <h2 className="text-xl font-semibold text-text mb-2">تم الاستيراد بنجاح!</h2>
          <p className="text-muted mb-6">
            تم استيراد {importJob?.processedRows} من {importJob?.totalRows} صف
          </p>

          {importJob?.failedRows && importJob.failedRows > 0 && (
            <div className="bg-warning-10 border border-warning-30 rounded-lg p-4 mb-6 text-left">
              <p className="text-sm text-warning mb-2">
                ⚠️ {importJob.failedRows} صفوف فشلت
              </p>
              <button className="text-sm text-warning underline">
                عرض التفاصيل
              </button>
            </div>
          )}

          <div className="flex justify-center gap-3">
            <button
              onClick={handleReset}
              className="px-4 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors"
            >
              استيراد المزيد
            </button>
            <button
              onClick={() => {/* TODO: Navigate to imported data */}}
              className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              عرض البيانات
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
