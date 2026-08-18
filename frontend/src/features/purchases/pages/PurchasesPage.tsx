import { useState } from 'react'
import { clsx } from 'clsx'
import { PurchaseDetail } from '../components/PurchaseDetail'
import { EmptyState } from '@/components/feedback'
import type { Purchase, PurchaseSummary } from '../types/purchase'

type StatusFilter = 'all' | 'paid' | 'partial' | 'pending'
type SortOption = 'date' | 'amount' | 'supplier'

export function PurchasesPage() {
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [sortBy, setSortBy] = useState<SortOption>('date')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedPurchase, setSelectedPurchase] = useState<Purchase | null>(null)

  // TODO: Fetch purchases from API
  const purchases: Purchase[] = []
  const summary: PurchaseSummary = {
    totalPurchases: 0,
    totalAmount: 0,
    totalPaid: 0,
    totalRemaining: 0,
    pendingAmount: 0,
    thisMonth: 0,
  }

  const filteredPurchases = purchases.filter(purchase => {
    if (statusFilter !== 'all' && purchase.paymentStatus !== statusFilter) return false
    if (searchQuery && !purchase.supplierName.toLowerCase().includes(searchQuery.toLowerCase())) return false
    return true
  }).sort((a, b) => {
    switch (sortBy) {
      case 'date':
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      case 'amount':
        return b.totalCost - a.totalCost
      case 'supplier':
        return a.supplierName.localeCompare(b.supplierName)
      default:
        return 0
    }
  })

  const handleViewPurchase = (purchase: Purchase) => {
    setSelectedPurchase(purchase)
  }

  const handleAddPayment = () => {
    // TODO: Open payment modal
    console.log('Add payment')
  }

  if (selectedPurchase) {
    return (
      <div className="container mx-auto p-6">
        <button
          onClick={() => setSelectedPurchase(null)}
          className="text-muted hover:text-text mb-4 inline-flex items-center gap-2"
        >
          ← العودة للمشتريات
        </button>
        <PurchaseDetail
          purchase={selectedPurchase}
          onAddPayment={handleAddPayment}
          onEdit={() => {/* TODO: Open edit modal */}}
          onDelete={() => {/* TODO: Show delete confirmation */}}
          onPrint={() => {/* TODO: Print purchase */}}
        />
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">المشتريات</h1>
        <p className="text-muted">إدارة المشتريات من الموردين</p>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <SummaryCard
          label="إجمالي المشتريات"
          value={summary.totalPurchases.toString()}
          icon="🛒"
        />
        <SummaryCard
          label="إجمالي المبالغ"
          value={summary.totalAmount.toFixed(2)}
          icon="💰"
        />
        <SummaryCard
          label="المتبقي"
          value={summary.totalRemaining.toFixed(2)}
          icon="⚠️"
          color={summary.totalRemaining > 0 ? 'warning' : 'success'}
        />
        <SummaryCard
          label="هذا الشهر"
          value={summary.thisMonth.toFixed(2)}
          icon="📅"
        />
      </div>

      {/* Filters */}
      <div className="bg-surface rounded-lg p-4 mb-6 space-y-4">
        {/* Search */}
        <div>
          <input
            type="text"
            placeholder="بحث عن مورد..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
        </div>

        {/* Status Filter */}
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted self-center">الحالة:</span>
          {(['all', 'paid', 'partial', 'pending'] as StatusFilter[]).map((status) => (
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
              {status === 'paid' && 'مدفوع'}
              {status === 'partial' && 'جزئي'}
              {status === 'pending' && 'معلق'}
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
            <option value="date">التاريخ</option>
            <option value="amount">المبلغ</option>
            <option value="supplier">المورد</option>
          </select>
        </div>
      </div>

      {/* Purchases List */}
      {filteredPurchases.length > 0 ? (
        <div className="bg-surface rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted-10 border-b border-border">
              <tr>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المورد</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">التكلفة</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المدفوع</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المتبقي</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">التاريخ</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {filteredPurchases.map((purchase) => (
                <tr key={purchase.id} className="border-b border-border hover:bg-muted-5">
                  <td className="px-4 py-3">
                    <div>
                      <p className="font-medium text-text">{purchase.supplierName}</p>
                      <p className="text-sm text-muted">#{purchase.id.slice(-6)}</p>
                    </div>
                  </td>
                  <td className="px-4 py-3 font-medium text-text">{purchase.totalCost.toFixed(2)}</td>
                  <td className="px-4 py-3 text-success">{purchase.paidAmount.toFixed(2)}</td>
                  <td className="px-4 py-3 font-medium text-warning">{purchase.remainingAmount.toFixed(2)}</td>
                  <td className="px-4 py-3 text-muted">{purchase.createdAt}</td>
                  <td className="px-4 py-3">
                    <StatusBadge status={purchase.paymentStatus} />
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => handleViewPurchase(purchase)}
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
        <EmptyState
          icon="🛒"
          title="لا توجد مشتريات"
          description="لا توجد مشتريات مطابقة للفلاتر الحالية"
          actionLabel="إنشاء مشتراة جديدة"
          onAction={() => {/* TODO: Open create purchase modal */}}
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

function StatusBadge({ status }: { status: Purchase['paymentStatus'] }) {
  const statusColors = {
    paid: 'bg-success-10 text-success',
    partial: 'bg-primary-10 text-primary',
    pending: 'bg-warning-10 text-warning',
  }

  const statusLabels = {
    paid: 'مدفوع',
    partial: 'جزئي',
    pending: 'معلق',
  }

  return (
    <span className={clsx('px-2 py-1 rounded-full text-xs font-medium', statusColors[status])}>
      {statusLabels[status]}
    </span>
  )
}
