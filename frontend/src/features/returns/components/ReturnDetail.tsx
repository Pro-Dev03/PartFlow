import { useState } from 'react'
import { clsx } from 'clsx'
import type { Return } from '../types/return'

interface ReturnDetailProps {
  returnRequest: Return
  onApprove?: () => void
  onReject?: () => void
  onComplete?: () => void
  onEdit?: () => void
  onDelete?: () => void
  className?: string
}

export function ReturnDetail({
  returnRequest,
  onApprove,
  onReject,
  onComplete,
  onEdit,
  onDelete,
  className,
}: ReturnDetailProps) {
  const [activeTab, setActiveTab] = useState<'details' | 'items' | 'inspection'>('details')

  const statusColors = {
    pending: 'bg-warning-10 text-warning',
    approved: 'bg-primary-10 text-primary',
    rejected: 'bg-danger-10 text-danger',
    completed: 'bg-success-10 text-success',
  }

  const statusLabels = {
    pending: 'معلق',
    approved: 'موافق عليه',
    rejected: 'مرفوض',
    completed: 'مكتمل',
  }

  const canProcess = returnRequest.status === 'pending'
  const canComplete = returnRequest.status === 'approved'

  return (
    <div className={clsx('bg-surface rounded-lg shadow-sm', className)}>
      {/* Header */}
      <div className="p-6 border-b border-border">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-2xl font-bold text-text">مرتج #{returnRequest.id.slice(-6)}</h1>
              <span className={clsx('px-3 py-1 rounded-full text-sm font-medium', statusColors[returnRequest.status])}>
                {statusLabels[returnRequest.status]}
              </span>
            </div>
            <div className="flex items-center gap-4 text-sm text-muted">
              <span>👤 {returnRequest.customerName}</span>
              <span>🛒 من بيع #{returnRequest.saleId.slice(-6)}</span>
              <span>📅 {returnRequest.requestedAt}</span>
            </div>
          </div>
          <div className="flex gap-2">
            {canProcess && onApprove && (
              <button
                onClick={onApprove}
                className="px-4 py-2 bg-success text-white rounded-lg hover:bg-success-600 transition-colors"
              >
                موافقة
              </button>
            )}
            {canProcess && onReject && (
              <button
                onClick={onReject}
                className="px-4 py-2 bg-danger text-white rounded-lg hover:bg-danger-600 transition-colors"
              >
                رفض
              </button>
            )}
            {canComplete && onComplete && (
              <button
                onClick={onComplete}
                className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
              >
                إكمال
              </button>
            )}
            {onEdit && (
              <button
                onClick={onEdit}
                className="px-4 py-2 border border-border rounded-lg hover:bg-surface-80 transition-colors"
              >
                تعديل
              </button>
            )}
            {onDelete && (
              <button
                onClick={onDelete}
                className="px-4 py-2 bg-danger text-white rounded-lg hover:bg-danger-600 transition-colors"
              >
                حذف
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Amount Summary */}
      <div className="p-6 border-b border-border">
        <div className="grid grid-cols-3 gap-4">
          <AmountCard label="إجمالي المبلغ" amount={returnRequest.totalAmount} />
          <AmountCard label="مبلغ الاسترداد" amount={returnRequest.refundAmount} color="primary" />
          <AmountCard label="عدد البنود" amount={returnRequest.items.length} color="info" />
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <nav className="flex gap-4 px-6">
          {(['details', 'items', 'inspection'] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={clsx(
                'px-4 py-3 text-sm font-medium border-b-2 transition-colors',
                activeTab === tab
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted hover:text-text'
              )}
            >
              {tab === 'details' && 'التفاصيل'}
              {tab === 'items' && 'البنود'}
              {tab === 'inspection' && 'الفحص'}
            </button>
          ))}
        </nav>
      </div>

      {/* Content */}
      <div className="p-6">
        {activeTab === 'details' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <h3 className="font-semibold text-text">معلومات المرتج</h3>
              <div className="space-y-2">
                <InfoRow label="رقم المرتج" value={returnRequest.id} />
                <InfoRow label="رقم البيع" value={returnRequest.saleId} />
                <InfoRow label="العميل" value={returnRequest.customerName} />
                <InfoRow label="الحالة" value={statusLabels[returnRequest.status]} />
              </div>
            </div>

            <div className="space-y-4">
              <h3 className="font-semibold text-text">التواريخ</h3>
              <div className="space-y-2">
                <InfoRow label="تاريخ الطلب" value={returnRequest.requestedAt} />
                <InfoRow label="تاريخ المعالجة" value={returnRequest.processedAt || 'لم تتم المعالجة'} />
                {returnRequest.processedBy && (
                  <InfoRow label="معالج بواسطة" value={returnRequest.processedBy} />
                )}
              </div>
            </div>

            <div className="space-y-4 md:col-span-2">
              <h3 className="font-semibold text-text">سبب المرتج</h3>
              <p className="text-muted bg-muted-10 p-4 rounded-lg">{returnRequest.reason}</p>
            </div>

            {returnRequest.notes && (
              <div className="space-y-4 md:col-span-2">
                <h3 className="font-semibold text-text">ملاحظات</h3>
                <p className="text-muted bg-muted-10 p-4 rounded-lg">{returnRequest.notes}</p>
              </div>
            )}
          </div>
        )}

        {activeTab === 'items' && (
          <div>
            {returnRequest.items.length > 0 ? (
              <div className="space-y-3">
                {returnRequest.items.map((item, index) => (
                  <div key={index} className="bg-muted-10 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium text-text">{item.productName}</span>
                      <span className="text-sm text-muted">الكمية: {item.quantity}</span>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-muted mb-2">
                      <span>السعر: {item.price.toFixed(2)}</span>
                      <span>الإجمالي: {item.totalAmount.toFixed(2)}</span>
                    </div>
                    <div className="flex items-center gap-4 text-sm">
                      <span className="text-muted">الحالة:</span>
                      <span className={clsx(
                        'px-2 py-1 rounded text-xs font-medium',
                        item.condition === 'good' ? 'bg-success-10 text-success' :
                        item.condition === 'damaged' ? 'bg-warning-10 text-warning' :
                        'bg-danger-10 text-danger'
                      )}>
                        {item.condition === 'good' ? 'جيد' : item.condition === 'damaged' ? 'تالف' : 'معيب'}
                      </span>
                      <span className="text-muted">سبب: {item.reason}</span>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-muted">
                لا توجد بنود
              </div>
            )}
          </div>
        )}

        {activeTab === 'inspection' && (
          <div className="text-center py-8 text-muted">
            نتائج الفحص ستظهر هنا
          </div>
        )}
      </div>
    </div>
  )
}

function AmountCard({ label, amount, color = 'primary' }: { label: string; amount: number; color?: 'success' | 'warning' | 'primary' | 'info' }) {
  const colorClasses = {
    success: 'text-success',
    warning: 'text-warning',
    primary: 'text-primary',
    info: 'text-info',
  }

  return (
    <div className="bg-muted-10 rounded-lg p-4 text-center">
      <p className="text-sm text-muted mb-1">{label}</p>
      <p className={clsx('text-xl font-bold', colorClasses[color])}>{typeof amount === 'number' ? amount.toFixed(2) : amount}</p>
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
