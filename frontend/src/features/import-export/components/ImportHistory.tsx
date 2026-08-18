import { useState } from 'react'
import { clsx } from 'clsx'
import type { ImportJob } from '../types/import-export'

export function ImportHistory() {
  const [jobs, setJobs] = useState<ImportJob[]>([
    {
      id: '1',
      type: 'products',
      fileName: 'products_import.csv',
      fileSize: 245000,
      status: 'completed',
      totalRows: 150,
      processedRows: 148,
      failedRows: 2,
      errors: [
        { row: 45, field: 'price', message: 'Invalid price format', value: 'abc' },
        { row: 89, field: 'barcode', message: 'Duplicate barcode', value: '123456' },
      ],
      createdAt: '2026-08-17T10:30:00Z',
      completedAt: '2026-08-17T10:32:00Z',
    },
    {
      id: '2',
      type: 'customers',
      fileName: 'customers_update.xlsx',
      fileSize: 125000,
      status: 'completed',
      totalRows: 50,
      processedRows: 50,
      failedRows: 0,
      createdAt: '2026-08-16T14:20:00Z',
      completedAt: '2026-08-16T14:21:00Z',
    },
  ])
  const [selectedJob, setSelectedJob] = useState<ImportJob | null>(null)

  const typeLabels = {
    products: 'المنتجات',
    customers: 'العملاء',
    suppliers: 'الموردين',
    inventory: 'المخزون',
  }

  const statusColors = {
    pending: 'bg-warning-10 text-warning',
    processing: 'bg-info-10 text-info',
    completed: 'bg-success-10 text-success',
    failed: 'bg-danger-10 text-danger',
    partial: 'bg-warning-10 text-warning',
  }

  const statusLabels = {
    pending: 'معلق',
    processing: 'جاري',
    completed: 'مكتمل',
    failed: 'فشل',
    partial: 'جزئي',
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('ar-SA', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <div className="space-y-6">
      {selectedJob ? (
        <div>
          <button
            onClick={() => setSelectedJob(null)}
            className="text-muted hover:text-text mb-4 inline-flex items-center gap-2"
          >
            ← العودة للسجل
          </button>

          <div className="bg-surface rounded-lg p-6">
            <h3 className="text-lg font-semibold text-text mb-4">تفاصيل الاستيراد</h3>

            <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-6">
              <div>
                <p className="text-sm text-muted">نوع البيانات</p>
                <p className="font-medium text-text">{typeLabels[selectedJob.type]}</p>
              </div>
              <div>
                <p className="text-sm text-muted">اسم الملف</p>
                <p className="font-medium text-text">{selectedJob.fileName}</p>
              </div>
              <div>
                <p className="text-sm text-muted">الحجم</p>
                <p className="font-medium text-text">{formatSize(selectedJob.fileSize)}</p>
              </div>
              <div>
                <p className="text-sm text-muted">الحالة</p>
                <span className={clsx('px-2 py-1 rounded text-xs font-medium', statusColors[selectedJob.status])}>
                  {statusLabels[selectedJob.status]}
                </span>
              </div>
              <div>
                <p className="text-sm text-muted">إجمالي الصفوف</p>
                <p className="font-medium text-text">{selectedJob.totalRows}</p>
              </div>
              <div>
                <p className="text-sm text-muted">تمت المعالجة</p>
                <p className="font-medium text-text">{selectedJob.processedRows}</p>
              </div>
              <div>
                <p className="text-sm text-muted">فشلت</p>
                <p className="font-medium text-text">{selectedJob.failedRows}</p>
              </div>
              <div>
                <p className="text-sm text-muted">تاريخ البدء</p>
                <p className="font-medium text-text">{formatDate(selectedJob.createdAt)}</p>
              </div>
              <div>
                <p className="text-sm text-muted">تاريخ الانتهاء</p>
                <p className="font-medium text-text">{selectedJob.completedAt ? formatDate(selectedJob.completedAt) : '-'}</p>
              </div>
            </div>

            {selectedJob.errors && selectedJob.errors.length > 0 && (
              <div>
                <h4 className="font-medium text-text mb-3">الأخطاء</h4>
                <div className="bg-danger-10 border border-danger-30 rounded-lg p-4">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-danger-30">
                        <th className="px-3 py-2 text-right font-medium text-text">الصف</th>
                        <th className="px-3 py-2 text-right font-medium text-text">الحقل</th>
                        <th className="px-3 py-2 text-right font-medium text-text">الرسالة</th>
                        <th className="px-3 py-2 text-right font-medium text-text">القيمة</th>
                      </tr>
                    </thead>
                    <tbody>
                      {selectedJob.errors.map((error, index) => (
                        <tr key={index} className="border-b border-danger-30">
                          <td className="px-3 py-2 text-text">{error.row}</td>
                          <td className="px-3 py-2 text-text">{error.field}</td>
                          <td className="px-3 py-2 text-danger">{error.message}</td>
                          <td className="px-3 py-2 text-muted">{error.value || '-'}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-text">سجل الاستيراد</h3>
            <button className="text-sm text-primary hover:text-primary-600 underline">
              تصفية النتائج
            </button>
          </div>

          {jobs.length > 0 ? (
            <div className="bg-surface rounded-lg overflow-hidden">
              <table className="w-full">
                <thead className="bg-muted-10 border-b border-border">
                  <tr>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الملف</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">النوع</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الصفوف</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">التاريخ</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
                  </tr>
                </thead>
                <tbody>
                  {jobs.map((job) => (
                    <tr key={job.id} className="border-b border-border hover:bg-muted-5">
                      <td className="px-4 py-3">
                        <div>
                          <p className="font-medium text-text">{job.fileName}</p>
                          <p className="text-sm text-muted">{formatSize(job.fileSize)}</p>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-muted">
                        {typeLabels[job.type]}
                      </td>
                      <td className="px-4 py-3">
                        <span className={clsx('px-2 py-1 rounded text-xs font-medium', statusColors[job.status])}>
                          {statusLabels[job.status]}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-sm">
                        <div>
                          <p className="text-text">{job.processedRows}/{job.totalRows}</p>
                          {job.failedRows > 0 && (
                            <p className="text-danger">{job.failedRows} فشل</p>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3 text-muted text-sm">
                        {formatDate(job.createdAt)}
                      </td>
                      <td className="px-4 py-3">
                        <button
                          onClick={() => setSelectedJob(job)}
                          className="text-primary hover:text-primary-600 text-sm"
                        >
                          عرض التفاصيل
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-center py-12 bg-muted-10 rounded-lg">
              <p className="text-muted">لا يوجد سجل استيراد</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
