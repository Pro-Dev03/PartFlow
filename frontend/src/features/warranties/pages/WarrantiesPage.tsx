import { useState } from 'react'
import { clsx } from 'clsx'
import { WarrantyDetail } from '../components/WarrantyDetail'
import { EmptyState } from '@/components/feedback'
import type { Warranty, WarrantySummary } from '../types/warranty'

type StatusFilter = 'all' | 'active' | 'expired' | 'claimed' | 'cancelled'
type SortOption = 'endDate' | 'startDate' | 'product' | 'customer'

export function WarrantiesPage() {
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [sortBy, setSortBy] = useState<SortOption>('endDate')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedWarranty, setSelectedWarranty] = useState<Warranty | null>(null)

  // TODO: Fetch warranties from API
  const warranties: Warranty[] = []
  const summary: WarrantySummary = {
    totalWarranties: 0,
    activeWarranties: 0,
    expiringSoon: 0,
    expiredWarranties: 0,
    claimedWarranties: 0,
    thisMonthExpiring: 0,
  }

  const filteredWarranties = warranties.filter(warranty => {
    if (statusFilter !== 'all' && warranty.status !== statusFilter) return false
    if (searchQuery && !warranty.productName.toLowerCase().includes(searchQuery.toLowerCase())) return false
    return true
  }).sort((a, b) => {
    switch (sortBy) {
      case 'endDate':
        return new Date(a.endDate).getTime() - new Date(b.endDate).getTime()
      case 'startDate':
        return new Date(b.startDate).getTime() - new Date(a.startDate).getTime()
      case 'product':
        return a.productName.localeCompare(b.productName)
      case 'customer':
        return (a.customerName || '').localeCompare(b.customerName || '')
      default:
        return 0
    }
  })

  const handleViewWarranty = (warranty: Warranty) => {
    setSelectedWarranty(warranty)
  }

  if (selectedWarranty) {
    return (
      <div className="container mx-auto p-6">
        <button
          onClick={() => setSelectedWarranty(null)}
          className="text-muted hover:text-text mb-4 inline-flex items-center gap-2"
        >
          ← العودة للضمانات
        </button>
        <WarrantyDetail
          warranty={selectedWarranty}
          claims={[]} // TODO: Fetch claims
          onEdit={() => {/* TODO: Open edit modal */}}
          onDelete={() => {/* TODO: Show delete confirmation */}}
          onFileClaim={() => {/* TODO: Open claim modal */}}
          onExtend={() => {/* TODO: Open extend modal */}}
          onCancel={() => {/* TODO: Show cancel confirmation */}}
        />
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">الضمانات</h1>
        <p className="text-muted">إدارة ومتابعة الضمانات</p>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <SummaryCard
          label="إجمالي الضمانات"
          value={summary.totalWarranties.toString()}
          icon="📋"
        />
        <SummaryCard
          label="نشط"
          value={summary.activeWarranties.toString()}
          icon="✅"
          color="success"
        />
        <SummaryCard
          label="ينتهي قريباً"
          value={summary.expiringSoon.toString()}
          icon="⚠️"
          color={summary.expiringSoon > 0 ? 'warning' : 'success'}
        />
        <SummaryCard
          label="منتهي"
          value={summary.expiredWarranties.toString()}
          icon="🏁"
          color={summary.expiredWarranties > 0 ? 'danger' : 'success'}
        />
      </div>

      {/* Filters */}
      <div className="bg-surface rounded-lg p-4 mb-6 space-y-4">
        {/* Search */}
        <div>
          <input
            type="text"
            placeholder="بحث عن منتج أو عميل..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
        </div>

        {/* Status Filter */}
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted self-center">الحالة:</span>
          {(['all', 'active', 'expired', 'claimed', 'cancelled'] as StatusFilter[]).map((status) => (
            <button
              key={status}
              onClick={() => setStatusFilter(status)}
              className={clsx(
                'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
                statusFilter === status
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted hover:bg-muted-80'
              )}
            >
              {status === 'all' && 'الكل'}
              {status === 'active' && 'نشط'}
              {status === 'expired' && 'منتهي'}
              {status === 'claimed' && 'مطالب به'}
              {status === 'cancelled' && 'ملغي'}
            </button>
          ))}
        </div>

        {/* Sort */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted">ترتيب حسب:</span>
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as SortOption)}
            className="px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="endDate">تاريخ الانتهاء</option>
            <option value="startDate">تاريخ البدء</option>
            <option value="product">المنتج</option>
            <option value="customer">العميل</option>
          </select>
        </div>
      </div>

      {/* Warranties List */}
      {filteredWarranties.length > 0 ? (
        <div className="bg-surface rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted-10 border-b border-border">
              <tr>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المنتج</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">العميل</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">النوع</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">تاريخ الانتهاء</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المدة</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {filteredWarranties.map((warranty) => {
                const daysRemaining = Math.ceil((new Date(warranty.endDate).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
                const isExpiringSoon = daysRemaining > 0 && daysRemaining <= 30
                const isExpired = daysRemaining <= 0

                return (
                  <tr key={warranty.id} className="border-b border-border hover:bg-muted-5">
                    <td className="px-4 py-3">
                      <div>
                        <p className="font-medium text-text">{warranty.productName}</p>
                        <p className="text-sm text-muted">#{warranty.id.slice(-6)}</p>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-muted">
                      {warranty.customerName || '-'}
                    </td>
                    <td className="px-4 py-3 text-muted">
                      {warranty.type === 'manufacturer' && 'الشركة المصنعة'}
                      {warranty.type === 'seller' && 'البائع'}
                      {warranty.type === 'extended' && 'ممتد'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <span className={clsx(
                          'text-sm',
                          isExpired ? 'text-danger' : isExpiringSoon ? 'text-warning' : 'text-muted'
                        )}>
                          {warranty.endDate}
                        </span>
                        {isExpiringSoon && !isExpired && (
                          <span className="text-xs text-warning">⚠️</span>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-muted">{warranty.duration} يوم</td>
                    <td className="px-4 py-3">
                      <StatusBadge status={warranty.status} />
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleViewWarranty(warranty)}
                        className="text-primary hover:text-primary-600 text-sm"
                      >
                        عرض التفاصيل
                      </button>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      ) : (
        <EmptyState
          icon="📋"
          title="لا توجد ضمانات"
          description="لا توجد ضمانات مطابقة للفلاتر الحالية"
          actionLabel="إضافة ضمان"
          onAction={() => {/* TODO: Open add warranty modal */}}
        />
      )}
    </div>
  )
}

function SummaryCard({ label, value, icon, color = 'primary' }: { label: string; value: string; icon: string; color?: 'success' | 'danger' | 'warning' | 'primary' }) {
  const colorClasses = {
    success: 'text-success',
    danger: 'text-danger',
    warning: 'text-warning',
    primary: 'text-primary',
  }

  return (
    <div className="bg-surface rounded-lg p-4 border border-border">
      <div className="flex items-center gap-2 mb-2">
        <span className="text-2xl">{icon}</span>
        <span className="text-sm text-muted">{label}</span>
      </div>
      <p className={clsx('text-xl font-bold', colorClasses[color])}>{value}</p>
    </div>
  )
}

function StatusBadge({ status }: { status: Warranty['status'] }) {
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

  return (
    <span className={clsx('px-2 py-1 rounded-full text-xs font-medium', statusColors[status])}>
      {statusLabels[status]}
    </span>
  )
}
