import { useState } from 'react'
import { clsx } from 'clsx'
import type { ExportJob } from '../types/import-export'

export function ExportHistory() {
  const [jobs, setJobs] = useState<ExportJob[]>([
    {
      id: '1',
      type: 'products',
      format: 'excel',
      filters: { category: 'all' },
      status: 'completed',
      fileName: 'products_export_2026-08-17.xlsx',
      fileSize: 1250000,
      downloadUrl: '/downloads/products_export_2026-08-17.xlsx',
      createdAt: '2026-08-17T09:15:00Z',
      completedAt: '2026-08-17T09:16:00Z',
    },
    {
      id: '2',
      type: 'sales',
      format: 'pdf',
      filters: { dateRange: { from: '2026-08-01', to: '2026-08-17' } },
      status: 'completed',
      fileName: 'sales_report_august_2026.pdf',
      fileSize: 245000,
      downloadUrl: '/downloads/sales_report_august_2026.pdf',
      createdAt: '2026-08-16T16:30:00Z',
      completedAt: '2026-08-16T16:31:00Z',
    },
  ])
  const [selectedJob, setSelectedJob] = useState<ExportJob | null>(null)

  const typeLabels = {
    products: 'المنتجات',
    customers: 'العملاء',
    suppliers: 'الموردين',
    sales: 'المبيعات',
    purchases: 'المشتريات',
    inventory: 'المخزون',
    reports: 'التقارير',
  }

  const formatLabels = {
    csv: 'CSV',
    excel: 'Excel',
    pdf: 'PDF',
  }

  const statusColors = {
    pending: 'bg-warning-10 text-warning',
    processing: 'bg-info-10 text-info',
    completed: 'bg-success-10 text-success',
    failed: 'bg-danger-10 text-danger',
  }

  const statusLabels = {
    pending: 'معلق',
    processing: 'جاري',
    completed: 'مكتمل',
    failed: 'فشل',
  }

  const formatSize = (bytes?: number) => {
    if (!bytes) return '-'
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

  const handleDownload = (job: ExportJob) => {
    if (job.downloadUrl) {
      // TODO: Trigger download
      console.log('Downloading:', job.downloadUrl)
    }
  }

  const handleDelete = (jobId: string) => {
    if (!confirm('هل أنت متأكد من حذف هذا التصدير؟')) {
      return
    }
    setJobs(jobs.filter(job => job.id !== jobId))
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
            <h3 className="text-lg font-semibold text-text mb-4">تفاصيل التصدير</h3>

            <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-6">
              <div>
                <p className="text-sm text-muted">نوع البيانات</p>
                <p className="font-medium text-text">{typeLabels[selectedJob.type]}</p>
              </div>
              <div>
                <p className="text-sm text-muted">الصيغة</p>
                <p className="font-medium text-text">{formatLabels[selectedJob.format]}</p>
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
                <p className="text-sm text-muted">تاريخ الإنشاء</p>
                <p className="font-medium text-text">{formatDate(selectedJob.createdAt)}</p>
              </div>
              {selectedJob.completedAt && (
                <div>
                  <p className="text-sm text-muted">تاريخ الانتهاء</p>
                  <p className="font-medium text-text">{formatDate(selectedJob.completedAt)}</p>
                </div>
              )}
            </div>

            {selectedJob.filters && Object.keys(selectedJob.filters).length > 0 && (
              <div className="mb-6">
                <h4 className="font-medium text-text mb-3">الفلاتر المستخدمة</h4>
                <div className="bg-muted-10 rounded-lg p-4">
                  <pre className="text-sm text-muted overflow-x-auto">
                    {JSON.stringify(selectedJob.filters, null, 2)}
                  </pre>
                </div>
              </div>
            )}

            <div className="flex gap-3">
              {selectedJob.status === 'completed' && selectedJob.downloadUrl && (
                <button
                  onClick={() => handleDownload(selectedJob)}
                  className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
                >
                  تحميل الملف
                </button>
              )}
              <button
                onClick={() => handleDelete(selectedJob.id)}
                className="px-4 py-2 bg-danger text-white rounded-lg hover:bg-danger-600 transition-colors"
              >
                حذف
              </button>
            </div>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-text">سجل التصدير</h3>
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
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الصيغة</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحجم</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">التاريخ</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
                  </tr>
                </thead>
                <tbody>
                  {jobs.map((job) => (
                    <tr key={job.id} className="border-b border-border hover:bg-muted-5">
                      <td className="px-4 py-3">
                        <p className="font-medium text-text">{job.fileName}</p>
                      </td>
                      <td className="px-4 py-3 text-muted">
                        {typeLabels[job.type]}
                      </td>
                      <td className="px-4 py-3 text-muted">
                        {formatLabels[job.format]}
                      </td>
                      <td className="px-4 py-3">
                        <span className={clsx('px-2 py-1 rounded text-xs font-medium', statusColors[job.status])}>
                          {statusLabels[job.status]}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-muted">
                        {formatSize(job.fileSize)}
                      </td>
                      <td className="px-4 py-3 text-muted text-sm">
                        {formatDate(job.createdAt)}
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex gap-2">
                          {job.status === 'completed' && job.downloadUrl && (
                            <button
                              onClick={() => handleDownload(job)}
                              className="text-primary hover:text-primary-600 text-sm"
                            >
                              تحميل
                            </button>
                          )}
                          <button
                            onClick={() => setSelectedJob(job)}
                            className="text-muted hover:text-text text-sm"
                          >
                            تفاصيل
                          </button>
                          <button
                            onClick={() => handleDelete(job.id)}
                            className="text-danger hover:text-danger-600 text-sm"
                          >
                            حذف
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-center py-12 bg-muted-10 rounded-lg">
              <p className="text-muted">لا يوجد سجل تصدير</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
