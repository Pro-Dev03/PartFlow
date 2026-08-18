import { useState } from 'react'
import { clsx } from 'clsx'
import type { Purchase } from '../types/purchase'

interface PurchaseDetailProps {
  purchase: Purchase
  onEdit?: () => void
  onDelete?: () => void
  onAddPayment?: () => void
  onPrint?: () => void
  className?: string
}

export function PurchaseDetail({
  purchase,
  onEdit,
  onDelete,
  onAddPayment,
  onPrint,
  className,
}: PurchaseDetailProps) {
  const [activeTab, setActiveTab] = useState<'details' | 'items' | 'payments'>('details')

  const statusColors = {
    paid: 'bg-success-10 text-success',
    partial: 'bg-primary-10 text-primary',
    pending: 'bg-warning-10 text-warning',
  }

  const statusLabels = {
    paid: 'مدفوع بالكامل',
    partial: 'مدفوع جزئياً',
    pending: 'معلق',
  }

  const isPartial = purchase.paymentStatus === 'partial'
  const isPending = purchase.paymentStatus === 'pending'
  const progressPercent = purchase.totalCost > 0 ? (purchase.paidAmount / purchase.totalCost) * 100 : 0

  return (
    <div className={clsx('bg-surface rounded-lg shadow-sm', className)}>
      {/* Header */}
      <div className="p-6 border-b border-border">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-2xl font-bold text-text">مشتراة #{purchase.id.slice(-6)}</h1>
              <span className={clsx('px-3 py-1 rounded-full text-sm font-medium', statusColors[purchase.paymentStatus])}>
                {statusLabels[purchase.paymentStatus]}
              </span>
            </div>
            <div className="flex items-center gap-4 text-sm text-muted">
              <span>🚚 {purchase.supplierName}</span>
              <span>📅 {purchase.createdAt}</span>
            </div>
          </div>
          <div className="flex gap-2">
            {onPrint && (
              <button
                onClick={onPrint}
                className="px-4 py-2 border border-border rounded-lg hover:bg-surface-80 transition-colors"
              >
                طباعة
              </button>
            )}
            {(isPartial || isPending) && onAddPayment && (
              <button
                onClick={onAddPayment}
                className="px-4 py-2 bg-success text-white rounded-lg hover:bg-success-600 transition-colors"
              >
                تسديد دفعة
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

      {/* Progress */}
      <div className="p-6 border-b border-border">
        <div className="mb-4">
          <div className="flex justify-between mb-2">
            <span className="text-sm text-muted">تقدم السداد</span>
            <span className="text-sm font-medium text-text">{progressPercent.toFixed(0)}%</span>
          </div>
          <div className="w-full bg-muted-20 rounded-full h-3">
            <div
              className={clsx(
                'h-3 rounded-full transition-all',
                purchase.paymentStatus === 'paid' ? 'bg-success' : 'bg-primary'
              )}
              style={{ width: `${progressPercent}%` }}
            />
          </div>
        </div>

        <div className="grid grid-cols-3 gap-4">
          <AmountCard label="إجمالي التكلفة" amount={purchase.totalCost} />
          <AmountCard label="المدفوع" amount={purchase.paidAmount} color="success" />
          <AmountCard label="المتبقي" amount={purchase.remainingAmount} color={isPending ? 'warning' : 'primary'} />
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <nav className="flex gap-4 px-6">
          {(['details', 'items', 'payments'] as const).map((tab) => (
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
              {tab === 'payments' && 'المدفوعات'}
            </button>
          ))}
        </nav>
      </div>

      {/* Content */}
      <div className="p-6">
        {activeTab === 'details' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <h3 className="font-semibold text-text">معلومات المشتراة</h3>
              <div className="space-y-2">
                <InfoRow label="رقم المشتراة" value={purchase.id} />
                <InfoRow label="المورد" value={purchase.supplierName} />
                <InfoRow label="حالة الدفع" value={statusLabels[purchase.paymentStatus]} />
              </div>
            </div>

            <div className="space-y-4">
              <h3 className="font-semibold text-text">التواريخ</h3>
              <div className="space-y-2">
                <InfoRow label="تاريخ الإنشاء" value={purchase.createdAt} />
                <InfoRow label="آخر تحديث" value={purchase.updatedAt} />
                <InfoRow label="أنشأ بواسطة" value={purchase.createdBy} />
              </div>
            </div>

            {purchase.notes && (
              <div className="space-y-4 md:col-span-2">
                <h3 className="font-semibold text-text">ملاحظات</h3>
                <p className="text-muted bg-muted-10 p-4 rounded-lg">{purchase.notes}</p>
              </div>
            )}
          </div>
        )}

        {activeTab === 'items' && (
          <div>
            {purchase.items.length > 0 ? (
              <div className="space-y-3">
                {purchase.items.map((item, index) => (
                  <div key={index} className="bg-muted-10 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium text-text">{item.productName}</span>
                      <span className="text-sm text-muted">الكمية: {item.quantity}</span>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-muted">
                      <span>سعر الوحدة: {item.cost.toFixed(2)}</span>
                      <span>الإجمالي: {item.totalCost.toFixed(2)}</span>
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

        {activeTab === 'payments' && (
          <div className="text-center py-8 text-muted">
            سجل المدفوعات سيظهر هنا
          </div>
        )}
      </div>
    </div>
  )
}

function AmountCard({ label, amount, color = 'primary' }: { label: string; amount: number; color?: 'success' | 'warning' | 'primary' }) {
  const colorClasses = {
    success: 'text-success',
    warning: 'text-warning',
    primary: 'text-primary',
  }

  return (
    <div className="bg-muted-10 rounded-lg p-4 text-center">
      <p className="text-sm text-muted mb-1">{label}</p>
      <p className={clsx('text-xl font-bold', colorClasses[color])}>{amount.toFixed(2)}</p>
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
