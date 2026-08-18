import { useState } from 'react'
import { clsx } from 'clsx'
import type { BackupInfo } from '../types/settings'

export function BackupManagement() {
  const [backups, setBackups] = useState<BackupInfo[]>([
    {
      id: '1',
      name: 'backup-2026-08-17-daily',
      size: 24500000,
      createdAt: '2026-08-17T02:00:00Z',
      type: 'automatic',
      status: 'completed',
    },
    {
      id: '2',
      name: 'backup-2026-08-16-daily',
      size: 24480000,
      createdAt: '2026-08-16T02:00:00Z',
      type: 'automatic',
      status: 'completed',
    },
    {
      id: '3',
      name: 'manual-backup-2026-08-15',
      size: 24520000,
      createdAt: '2026-08-15T14:30:00Z',
      type: 'manual',
      status: 'completed',
    },
  ])
  const [creatingBackup, setCreatingBackup] = useState(false)
  const [restoringBackup, setRestoringBackup] = useState<string | null>(null)

  const handleCreateBackup = async () => {
    setCreatingBackup(true)
    try {
      // TODO: Create backup via API
      console.log('Creating backup...')
    } catch (error) {
      console.error('Failed to create backup:', error)
    } finally {
      setCreatingBackup(false)
    }
  }

  const handleRestoreBackup = async (backupId: string) => {
    if (!confirm('هل أنت متأكد من استعادة هذا النسخة الاحتياطية؟ سيتم استبدال جميع البيانات الحالية.')) {
      return
    }

    setRestoringBackup(backupId)
    try {
      // TODO: Restore backup via API
      console.log('Restoring backup:', backupId)
    } catch (error) {
      console.error('Failed to restore backup:', error)
    } finally {
      setRestoringBackup(null)
    }
  }

  const handleDeleteBackup = async (backupId: string) => {
    if (!confirm('هل أنت متأكد من حذف هذا النسخة الاحتياطية؟')) {
      return
    }

    try {
      // TODO: Delete backup via API
      setBackups(backups.filter(b => b.id !== backupId))
    } catch (error) {
      console.error('Failed to delete backup:', error)
    }
  }

  const handleDownloadBackup = (backupId: string) => {
    // TODO: Download backup file
    console.log('Downloading backup:', backupId)
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

  const statusColors = {
    completed: 'bg-success-10 text-success',
    failed: 'bg-danger-10 text-danger',
    in_progress: 'bg-warning-10 text-warning',
  }

  const statusLabels = {
    completed: 'مكتمل',
    failed: 'فشل',
    in_progress: 'جاري...',
  }

  const typeLabels = {
    automatic: 'تلقائي',
    manual: 'يدوي',
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-text">النسخ الاحتياطي</h3>
          <p className="text-sm text-muted">إدارة نسخ البيانات الاحتياطية</p>
        </div>
        <button
          onClick={handleCreateBackup}
          disabled={creatingBackup}
          className={clsx(
            'px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors',
            creatingBackup && 'opacity-50 cursor-not-allowed'
          )}
        >
          {creatingBackup ? 'جاري الإنشاء...' : 'إنشاء نسخة احتياطية'}
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <div className="bg-surface rounded-lg p-4 border border-border">
          <p className="text-sm text-muted mb-1">إجمالي النسخ</p>
          <p className="text-2xl font-bold text-text">{backups.length}</p>
        </div>
        <div className="bg-surface rounded-lg p-4 border border-border">
          <p className="text-sm text-muted mb-1">آخر نسخة</p>
          <p className="text-2xl font-bold text-text">
            {backups.length > 0 ? formatDate(backups[0].createdAt) : '-'}
          </p>
        </div>
        <div className="bg-surface rounded-lg p-4 border border-border">
          <p className="text-sm text-muted mb-1">المساحة المستخدمة</p>
          <p className="text-2xl font-bold text-text">
            {formatSize(backups.reduce((total, b) => total + b.size, 0))}
          </p>
        </div>
      </div>

      {/* Backups List */}
      <div className="bg-surface rounded-lg overflow-hidden">
        <table className="w-full">
          <thead className="bg-muted-10 border-b border-border">
            <tr>
              <th className="px-4 py-3 text-right text-sm font-medium text-muted">الاسم</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-muted">التاريخ</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحجم</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-muted">النوع</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
            </tr>
          </thead>
          <tbody>
            {backups.map((backup) => (
              <tr key={backup.id} className="border-b border-border hover:bg-muted-5">
                <td className="px-4 py-3">
                  <p className="font-medium text-text">{backup.name}</p>
                </td>
                <td className="px-4 py-3 text-muted text-sm">
                  {formatDate(backup.createdAt)}
                </td>
                <td className="px-4 py-3 text-muted">
                  {formatSize(backup.size)}
                </td>
                <td className="px-4 py-3">
                  <span className="px-2 py-1 rounded text-xs font-medium bg-muted-10 text-muted">
                    {typeLabels[backup.type]}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={clsx('px-2 py-1 rounded text-xs font-medium', statusColors[backup.status])}>
                    {statusLabels[backup.status]}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleDownloadBackup(backup.id)}
                      className="text-primary hover:text-primary-600 text-sm"
                      disabled={backup.status !== 'completed'}
                    >
                      تحميل
                    </button>
                    <button
                      onClick={() => handleRestoreBackup(backup.id)}
                      className="text-success hover:text-success-600 text-sm"
                      disabled={backup.status !== 'completed' || restoringBackup !== null}
                    >
                      {restoringBackup === backup.id ? 'جاري...' : 'استعادة'}
                    </button>
                    <button
                      onClick={() => handleDeleteBackup(backup.id)}
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

      {/* Import Backup */}
      <div className="bg-surface rounded-lg p-6 border border-border">
        <h4 className="font-medium text-text mb-4">استيراد نسخة احتياطية</h4>
        <div className="flex items-center gap-4">
          <input
            type="file"
            accept=".zip,.sql"
            className="flex-1 px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
          <button
            onClick={() => {/* TODO: Handle import */}}
            className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
          >
            استيراد
          </button>
        </div>
        <p className="text-sm text-muted mt-2">
          يدعم ملفات .zip و .sql
        </p>
      </div>
    </div>
  )
}
