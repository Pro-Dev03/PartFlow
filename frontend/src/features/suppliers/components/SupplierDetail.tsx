import { useState } from 'react'
import { clsx } from 'clsx'
import { SupplierTimeline } from './SupplierTimeline'
import type { Supplier, SupplierStats } from '../types/supplier'

interface SupplierDetailProps {
  supplier: Supplier
  stats: SupplierStats
  onEdit?: () => void
  onDelete?: () => void
  onAddPayment?: () => void
  onCreatePurchase?: () => void
  className?: string
}

export function SupplierDetail({
  supplier,
  stats,
  onEdit,
  onDelete,
  onAddPayment,
  onCreatePurchase,
  className,
}: SupplierDetailProps) {
  const [activeTab, setActiveTab] = useState<'details' | 'timeline' | 'purchases'>('details')

  const hasOutstandingBalance = supplier.outstandingBalance > 0
  const isOverdue = supplier.outstandingBalance > 0 // TODO: Add actual overdue logic

  return (
    <div className={clsx('bg-surface rounded-lg shadow-sm', className)}>
      {/* Header */}
      <div className="p-6 border-b border-border">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-2xl font-bold text-text">{supplier.name}</h1>
              {!supplier.isActive && (
                <span className="px-2 py-1 rounded text-xs font-medium bg-muted text-muted">
                  غير نشط
                </span>
              )}
            </div>
            <div className="flex items-center gap-4 text-sm text-muted">
              <span>📱 {supplier.phone}</span>
              {supplier.email && <span>✉️ {supplier.email}</span>}
            </div>
          </div>
          <div className="flex gap-2">
            {onCreatePurchase && (
              <button
                onClick={onCreatePurchase}
                className="px-4 py-2 bg-success text-white rounded-lg hover:bg-success-600 transition-colors"
              >
                إنشاء مشتراة
              </button>
            )}
            {hasOutstandingBalance && onAddPayment && (
              <button
                onClick={onAddPayment}
                className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
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

      {/* Outstanding Balance Alert */}
      {hasOutstandingBalance && (
        <div className={clsx('px-6 py-3 border-b border-border', isOverdue ? 'bg-danger-10' : 'bg-warning-10')}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className={isOverdue ? 'text-danger' : 'text-warning'}>
                {isOverdue ? '⚠️' : '📋'}
              </span>
              <span className={clsx('font-medium', isOverdue ? 'text-danger' : 'text-warning')}>
                {isOverdue ? 'رصيد متأخر' : 'رصيد مستحق'}
              </span>
            </div>
            <span className={clsx('text-lg font-bold', isOverdue ? 'text-danger' : 'text-warning')}>
              {supplier.outstandingBalance.toFixed(2)}
            </span>
          </div>
        </div>
      )}

      {/* Stats Cards */}
      <div className="p-6 border-b border-border">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <StatCard
            label="إجمالي المشتريات"
            value={stats.totalPurchases.toString()}
            icon="🛒"
          />
          <StatCard
            label="إجمالي المبالغ"
            value={stats.totalAmount.toFixed(2)}
            icon="💰"
          />
          <StatCard
            label="متوسط قيمة الشراء"
            value={stats.averagePurchaseValue.toFixed(2)}
            icon="📊"
          />
          <StatCard
            label="المدفوعات في الوقت"
            value={`${stats.paymentHistory.onTime}%`}
            icon="✅"
          />
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <nav className="flex gap-4 px-6">
          {(['details', 'timeline', 'purchases'] as const).map((tab) => (
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
              {tab === 'timeline' && 'السجل'}
              {tab === 'purchases' && 'المشتريات'}
            </button>
          ))}
        </nav>
      </div>

      {/* Content */}
      <div className="p-6">
        {activeTab === 'details' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <h3 className="font-semibold text-text">معلومات الاتصال</h3>
              <div className="space-y-2">
                <InfoRow label="الهاتف" value={supplier.phone} />
                {supplier.email && <InfoRow label="البريد الإلكتروني" value={supplier.email} />}
                {supplier.address && <InfoRow label="العنوان" value={supplier.address} />}
                {supplier.city && <InfoRow label="المدينة" value={supplier.city} />}
              </div>
            </div>

            <div className="space-y-4">
              <h3 className="font-semibold text-text">معلومات إضافية</h3>
              <div className="space-y-2">
                <InfoRow label="تاريخ التسجيل" value={supplier.createdAt} />
                <InfoRow label="آخر شراء" value={supplier.lastPurchaseDate || 'لا يوجد'} />
                <InfoRow
                  label="الحالة"
                  value={supplier.isActive ? 'نشط' : 'غير نشط'}
                />
              </div>
            </div>

            {supplier.notes && (
              <div className="space-y-4 md:col-span-2">
                <h3 className="font-semibold text-text">ملاحظات</h3>
                <p className="text-muted bg-muted-10 p-4 rounded-lg">{supplier.notes}</p>
              </div>
            )}
          </div>
        )}

        {activeTab === 'timeline' && (
          <SupplierTimeline
            events={[
              {
                id: '1',
                type: 'purchase',
                date: supplier.createdAt,
                description: 'تسجيل المورد',
                user: 'النظام',
              },
            ]}
          />
        )}

        {activeTab === 'purchases' && (
          <div className="text-center py-8 text-muted">
            قائمة المشتريات ستظهر هنا
          </div>
        )}
      </div>
    </div>
  )
}

function StatCard({ label, value, icon }: { label: string; value: string; icon: string }) {
  return (
    <div className="bg-muted-10 rounded-lg p-4">
      <div className="flex items-center gap-2 mb-2">
        <span className="text-2xl">{icon}</span>
        <span className="text-sm text-muted">{label}</span>
      </div>
      <p className="text-xl font-bold text-text">{value}</p>
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
