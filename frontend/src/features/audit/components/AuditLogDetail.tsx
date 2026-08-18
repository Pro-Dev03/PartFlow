import { useState } from 'react'
import { clsx } from 'clsx'
import type { AuditLog } from '../types/audit'

interface AuditLogDetailProps {
  log: AuditLog
  onClose: () => void
  className?: string
}

export function AuditLogDetail({ log, onClose, className }: AuditLogDetailProps) {
  const [showChanges, setShowChanges] = useState(false)

  const entityTypeLabels = {
    product: 'منتج',
    customer: 'عميل',
    sale: 'بيع',
    purchase: 'مشتراة',
    expense: 'مصروف',
    supplier: 'مورد',
    inventory: 'مخزون',
    settings: 'إعدادات',
    user: 'مستخدم',
  }

  return (
    <div className={clsx('bg-surface rounded-lg shadow-sm', className)}>
      {/* Header */}
      <div className="p-6 border-b border-border flex items-center justify-between">
        <h2 className="text-xl font-bold text-text">تفاصيل السجل</h2>
        <button
          onClick={onClose}
          className="text-muted hover:text-text p-2 rounded-lg hover:bg-muted-10"
        >
          ✕
        </button>
      </div>

      {/* Content */}
      <div className="p-6 space-y-6">
        {/* Basic Info */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <InfoRow label="معرف السجل" value={log.id} />
          <InfoRow label="المستخدم" value={log.userName} />
          <InfoRow label="الإجراء" value={log.action} />
          <InfoRow
            label="الكيان"
            value={`${entityTypeLabels[log.entityType as keyof typeof entityTypeLabels] || log.entityType} ${log.entityName ? `(${log.entityName})` : ''}`}
          />
          <InfoRow label="التاريخ" value={log.timestamp} />
          <InfoRow label="الحالة" value={log.status === 'success' ? 'نجح' : 'فشل'} />
          {log.ipAddress && <InfoRow label="عنوان IP" value={log.ipAddress} />}
        </div>

        {/* Error Message */}
        {log.errorMessage && (
          <div className="bg-danger-10 border border-danger-30 rounded-lg p-4">
            <p className="text-sm font-medium text-danger mb-1">رسالة الخطأ:</p>
            <p className="text-sm text-danger">{log.errorMessage}</p>
          </div>
        )}

        {/* Changes */}
        {log.changes && log.changes.length > 0 && (
          <div>
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold text-text">التغييرات</h3>
              <button
                onClick={() => setShowChanges(!showChanges)}
                className="text-sm text-primary hover:text-primary-80"
              >
                {showChanges ? 'إخفاء' : 'عرض'}
              </button>
            </div>
            {showChanges && (
              <div className="space-y-2">
                {log.changes.map((change, index) => (
                  <div key={index} className="bg-muted-10 rounded-lg p-3">
                    <p className="text-sm font-medium text-text mb-2">{change.field}</p>
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <span className="text-muted">القديم:</span>
                        <span className="text-danger line-through ml-2">
                          {change.oldValue?.toString() || '-'}
                        </span>
                      </div>
                      <div>
                        <span className="text-muted">الجديد:</span>
                        <span className="text-success ml-2">
                          {change.newValue?.toString() || '-'}
                        </span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Old Values */}
        {log.oldValues && Object.keys(log.oldValues).length > 0 && (
          <div>
            <h3 className="font-semibold text-text mb-4">القيم القديمة</h3>
            <div className="bg-muted-10 rounded-lg p-4">
              <pre className="text-sm text-muted overflow-x-auto">
                {JSON.stringify(log.oldValues, null, 2)}
              </pre>
            </div>
          </div>
        )}

        {/* New Values */}
        {log.newValues && Object.keys(log.newValues).length > 0 && (
          <div>
            <h3 className="font-semibold text-text mb-4">القيم الجديدة</h3>
            <div className="bg-muted-10 rounded-lg p-4">
              <pre className="text-sm text-muted overflow-x-auto">
                {JSON.stringify(log.newValues, null, 2)}
              </pre>
            </div>
          </div>
        )}

        {/* User Agent */}
        {log.userAgent && (
          <div>
            <h3 className="font-semibold text-text mb-2">معلومات المتصفح</h3>
            <p className="text-sm text-muted bg-muted-10 p-3 rounded-lg">{log.userAgent}</p>
          </div>
        )}
      </div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between py-2 border-b border-border">
      <span className="text-muted">{label}</span>
      <span className="font-medium text-text">{value}</span>
    </div>
  )
}
