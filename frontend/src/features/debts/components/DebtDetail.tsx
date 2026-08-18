import { useState } from 'react'
import { clsx } from 'clsx'
import type { Debt, DebtPayment } from '../types/debt'

interface DebtDetailProps {
  debt: Debt
  payments: DebtPayment[]
  onAddPayment?: () => void
  onEdit?: () => void
  onDelete?: () => void
  onSendReminder?: () => void
  className?: string
}

export function DebtDetail({
  debt,
  payments,
  onAddPayment,
  onEdit,
  onDelete,
  onSendReminder,
  className,
}: DebtDetailProps) {
  const [activeTab, setActiveTab] = useState<'details' | 'payments' | 'timeline'>('details')

  const statusColors = {
    pending: 'bg-warning-10 text-warning',
    partial: 'bg-primary-10 text-primary',
    paid: 'bg-success-10 text-success',
    overdue: 'bg-danger-10 text-danger',
  }

  const statusLabels = {
    pending: 'معلق',
    partial: 'مدفوع جزئياً',
    paid: 'مدفوع بالكامل',
    overdue: 'متأخر',
  }

  const isOverdue = debt.status === 'overdue'
  const isPaid = debt.status === 'paid'
  const progressPercent = debt.amount > 0 ? (debt.paidAmount / debt.amount) * 100 : 0

  return (
    <div className={clsx('bg-surface rounded-lg shadow-sm', className)}>
      {/* Header */}
      <div className="p-6 border-b border-border">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-2xl font-bold text-text">دين #{debt.id.slice(-6)}</h1>
              <span className={clsx('px-3 py-1 rounded-full text-sm font-medium', statusColors[debt.status])}>
                {statusLabels[debt.status]}
              </span>
            </div>
            <div className="flex items-center gap-4 text-sm text-muted">
              <span>👤 {debt.customerName}</span>
              <span>📱 {debt.customerPhone}</span>
            </div>
          </div>
          <div className="flex gap-2">
            {!isPaid && onAddPayment && (
              <button
                onClick={onAddPayment}
                className="px-4 py-2 bg-success text-white rounded-lg hover:bg-success-600 transition-colors"
              >
                تسديد دفعة
              </button>
            )}
            {isOverdue && onSendReminder && (
              <button
                onClick={onSendReminder}
                className="px-4 py-2 bg-warning text-white rounded-lg hover:bg-warning-600 transition-colors"
              >
                إرسال تذكير
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
                isPaid ? 'bg-success' : isOverdue ? 'bg-danger' : 'bg-primary'
              )}
              style={{ width: `${progressPercent}%` }}
            />
          </div>
        </div>

        <div className="grid grid-cols-3 gap-4">
          <AmountCard label="إجمالي المبلغ" amount={debt.amount} />
          <AmountCard label="المدفوع" amount={debt.paidAmount} color="success" />
          <AmountCard label="المتبقي" amount={debt.remainingAmount} color={isOverdue ? 'danger' : 'warning'} />
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <nav className="flex gap-4 px-6">
          {(['details', 'payments', 'timeline'] as const).map((tab) => (
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
              {tab === 'payments' && 'المدفوعات'}
              {tab === 'timeline' && 'السجل'}
            </button>
          ))}
        </nav>
      </div>

      {/* Content */}
      <div className="p-6">
        {activeTab === 'details' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <h3 className="font-semibold text-text">معلومات الدين</h3>
              <div className="space-y-2">
                <InfoRow label="رقم الدين" value={debt.id} />
                <InfoRow label="تاريخ الاستحقاق" value={debt.dueDate} />
                {debt.saleId && <InfoRow label="رقم البيع" value={debt.saleId} />}
              </div>
            </div>

            <div className="space-y-4">
              <h3 className="font-semibold text-text">التواريخ</h3>
              <div className="space-y-2">
                <InfoRow label="تاريخ الإنشاء" value={debt.createdAt} />
                <InfoRow label="آخر تحديث" value={debt.updatedAt} />
              </div>
            </div>

            {debt.notes && (
              <div className="space-y-4 md:col-span-2">
                <h3 className="font-semibold text-text">ملاحظات</h3>
                <p className="text-muted bg-muted-10 p-4 rounded-lg">{debt.notes}</p>
              </div>
            )}
          </div>
        )}

        {activeTab === 'payments' && (
          <div>
            {payments.length > 0 ? (
              <div className="space-y-3">
                {payments.map((payment) => (
                  <div key={payment.id} className="bg-muted-10 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium text-text">{payment.amount.toFixed(2)}</span>
                      <span className="text-sm text-muted">{payment.createdAt}</span>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-muted">
                      <span>الطريقة: {payment.method}</span>
                      {payment.reference && <span>المرجع: {payment.reference}</span>}
                      <span>بواسطة: {payment.createdBy}</span>
                    </div>
                    {payment.notes && (
                      <p className="text-sm text-muted mt-2">{payment.notes}</p>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-muted">
                لا توجد مدفوعات بعد
              </div>
            )}
          </div>
        )}

        {activeTab === 'timeline' && (
          <div className="text-center py-8 text-muted">
            سجل النشاطات سيظهر هنا
          </div>
        )}
      </div>
    </div>
  )
}

function AmountCard({ label, amount, color = 'primary' }: { label: string; amount: number; color?: 'success' | 'danger' | 'warning' | 'primary' }) {
  const colorClasses = {
    success: 'text-success',
    danger: 'text-danger',
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
