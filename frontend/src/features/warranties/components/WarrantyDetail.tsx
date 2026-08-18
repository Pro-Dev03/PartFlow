import { useState } from 'react'
import { clsx } from 'clsx'
import type { Warranty, WarrantyClaim } from '../types/warranty'

interface WarrantyDetailProps {
  warranty: Warranty
  claims: WarrantyClaim[]
  onEdit?: () => void
  onDelete?: () => void
  onFileClaim?: () => void
  onExtend?: () => void
  onCancel?: () => void
  className?: string
}

export function WarrantyDetail({
  warranty,
  claims,
  onEdit,
  onDelete,
  onFileClaim,
  onExtend,
  onCancel,
  className,
}: WarrantyDetailProps) {
  const [activeTab, setActiveTab] = useState<'details' | 'claims' | 'terms'>('details')

  const statusColors = {
    active: 'bg-success-10 text-success',
    expired: 'bg-danger-10 text-danger',
    claimed: 'bg-warning-10 text-warning',
    cancelled: 'bg-muted-10 text-muted',
  }

  const statusLabels = {
    active: 'نشط',
    expired: 'منتهي',
    claimed: 'مطالب به',
    cancelled: 'ملغي',
  }

  const typeLabels = {
    manufacturer: 'الشركة المصنعة',
    seller: 'البائع',
    extended: 'ممتد',
  }

  const daysRemaining = Math.ceil((new Date(warranty.endDate).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
  const isExpiringSoon = daysRemaining > 0 && daysRemaining <= 30
  const isExpired = daysRemaining <= 0
  const isActive = warranty.status === 'active'

  return (
    <div className={clsx('bg-surface rounded-lg shadow-sm', className)}>
      {/* Header */}
      <div className="p-6 border-b border-border">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-2xl font-bold text-text">ضمان #{warranty.id.slice(-6)}</h1>
              <span className={clsx('px-3 py-1 rounded-full text-sm font-medium', statusColors[warranty.status])}>
                {statusLabels[warranty.status]}
              </span>
              {isExpiringSoon && isActive && (
                <span className="px-2 py-1 rounded text-xs font-medium bg-warning text-white">
                  ينتهي قريباً
                </span>
              )}
            </div>
            <div className="flex items-center gap-4 text-sm text-muted">
              <span>📦 {warranty.productName}</span>
              {warranty.customerName && <span>👤 {warranty.customerName}</span>}
            </div>
          </div>
          <div className="flex gap-2">
            {isActive && onFileClaim && (
              <button
                onClick={onFileClaim}
                className="px-4 py-2 bg-warning text-white rounded-lg hover:bg-warning-600 transition-colors"
              >
                تقديم مطالبة
              </button>
            )}
            {isActive && onExtend && (
              <button
                onClick={onExtend}
                className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
              >
                تمديد
              </button>
            )}
            {isActive && onCancel && (
              <button
                onClick={onCancel}
                className="px-4 py-2 border border-border rounded-lg hover:bg-surface-80 transition-colors"
              >
                إلغاء
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

      {/* Duration Alert */}
      {isActive && (
        <div className={clsx('px-6 py-3 border-b border-border', isExpiringSoon ? 'bg-warning-10' : 'bg-success-10')}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className={isExpiringSoon ? 'text-warning' : 'text-success'}>
                {isExpiringSoon ? '⚠️' : '✅'}
              </span>
              <span className={clsx('font-medium', isExpiringSoon ? 'text-warning' : 'text-success')}>
                {isExpiringSoon ? 'ينتهي قريباً' : 'نشط'}
              </span>
            </div>
            <span className={clsx('text-lg font-bold', isExpiringSoon ? 'text-warning' : 'text-success')}>
              {isExpired ? 'منتهي' : `${daysRemaining} يوم متبقي`}
            </span>
          </div>
        </div>
      )}

      {/* Info Cards */}
      <div className="p-6 border-b border-border">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <InfoCard label="نوع الضمان" value={typeLabels[warranty.type]} icon="📋" />
          <InfoCard label="المدة" value={`${warranty.duration} يوم`} icon="📅" />
          <InfoCard label="تاريخ البدء" value={warranty.startDate} icon="🚀" />
          <InfoCard label="تاريخ الانتهاء" value={warranty.endDate} icon="🏁" />
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <nav className="flex gap-4 px-6">
          {(['details', 'claims', 'terms'] as const).map((tab) => (
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
              {tab === 'claims' && 'المطالبات'}
              {tab === 'terms' && 'الشروط'}
            </button>
          ))}
        </nav>
      </div>

      {/* Content */}
      <div className="p-6">
        {activeTab === 'details' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <h3 className="font-semibold text-text">معلومات الضمان</h3>
              <div className="space-y-2">
                <InfoRow label="رقم الضمان" value={warranty.id} />
                <InfoRow label="المنتج" value={warranty.productName} />
                {warranty.customerName && <InfoRow label="العميل" value={warranty.customerName} />}
                <InfoRow label="الحالة" value={statusLabels[warranty.status]} />
              </div>
            </div>

            <div className="space-y-4">
              <h3 className="font-semibold text-text">التواريخ</h3>
              <div className="space-y-2">
                <InfoRow label="تاريخ الإنشاء" value={warranty.createdAt} />
                <InfoRow label="آخر تحديث" value={warranty.updatedAt} />
              </div>
            </div>

            {warranty.notes && (
              <div className="space-y-4 md:col-span-2">
                <h3 className="font-semibold text-text">ملاحظات</h3>
                <p className="text-muted bg-muted-10 p-4 rounded-lg">{warranty.notes}</p>
              </div>
            )}
          </div>
        )}

        {activeTab === 'claims' && (
          <div>
            {claims.length > 0 ? (
              <div className="space-y-3">
                {claims.map((claim) => (
                  <div key={claim.id} className="bg-muted-10 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium text-text">مطالبة #{claim.id.slice(-6)}</span>
                      <span className="text-sm text-muted">{claim.claimedAt}</span>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-muted mb-2">
                      <span>السبب: {claim.reason}</span>
                      <ClaimStatusBadge status={claim.status} />
                    </div>
                    {claim.resolution && (
                      <p className="text-sm text-muted">الحل: {claim.resolution}</p>
                    )}
                    {claim.notes && (
                      <p className="text-sm text-muted mt-2">{claim.notes}</p>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-muted">
                لا توجد مطالبات
              </div>
            )}
          </div>
        )}

        {activeTab === 'terms' && (
          <div>
            {warranty.terms ? (
              <div className="bg-muted-10 p-4 rounded-lg">
                <p className="text-muted whitespace-pre-wrap">{warranty.terms}</p>
              </div>
            ) : (
              <div className="text-center py-8 text-muted">
                لا توجد شروط محددة
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

function InfoCard({ label, value, icon }: { label: string; value: string; icon: string }) {
  return (
    <div className="bg-muted-10 rounded-lg p-4">
      <div className="flex items-center gap-2 mb-2">
        <span className="text-xl">{icon}</span>
        <span className="text-sm text-muted">{label}</span>
      </div>
      <p className="text-sm font-medium text-text">{value}</p>
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

function ClaimStatusBadge({ status }: { status: WarrantyClaim['status'] }) {
  const statusColors = {
    pending: 'bg-warning-10 text-warning',
    approved: 'bg-primary-10 text-primary',
    rejected: 'bg-danger-10 text-danger',
    completed: 'bg-success-10 text-success',
  }

  const statusLabels = {
    pending: 'معلق',
    approved: 'موافق',
    rejected: 'مرفوض',
    completed: 'مكتمل',
  }

  return (
    <span className={clsx('px-2 py-1 rounded text-xs font-medium', statusColors[status])}>
      {statusLabels[status]}
    </span>
  )
}
