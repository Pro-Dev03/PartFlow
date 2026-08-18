import { useState } from 'react'
import { clsx } from 'clsx'
import { ReturnDetail } from '../components/ReturnDetail'
import { EmptyState } from '@/components/feedback'
import type { Return, ReturnSummary } from '../types/return'

type StatusFilter = 'all' | 'pending' | 'approved' | 'rejected' | 'completed'
type SortOption = 'date' | 'amount' | 'customer'

export function ReturnsPage() {
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [sortBy, setSortBy] = useState<SortOption>('date')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedReturn, setSelectedReturn] = useState<Return | null>(null)

  // TODO: Fetch returns from API
  const returns: Return[] = []
  const summary: ReturnSummary = {
    totalReturns: 0,
    pendingReturns: 0,
    approvedReturns: 0,
    rejectedReturns: 0,
    totalRefundAmount: 0,
    thisMonth: 0,
  }

  const filteredReturns = returns.filter(returnItem => {
    if (statusFilter !== 'all' && returnItem.status !== statusFilter) return false
    if (searchQuery && !returnItem.customerName.toLowerCase().includes(searchQuery.toLowerCase())) return false
    return true
  }).sort((a, b) => {
    switch (sortBy) {
      case 'date':
        return new Date(b.requestedAt).getTime() - new Date(a.requestedAt).getTime()
      case 'amount':
        return b.refundAmount - a.refundAmount
      case 'customer':
        return a.customerName.localeCompare(b.customerName)
      default:
        return 0
    }
  })

  const handleViewReturn = (returnItem: Return) => {
    setSelectedReturn(returnItem)
  }

  const handleApprove = () => {
    // TODO: Approve return
    console.log('Approve return')
  }

  const handleReject = () => {
    // TODO: Reject return
    console.log('Reject return')
  }

  const handleComplete = () => {
    // TODO: Complete return
    console.log('Complete return')
  }

  if (selectedReturn) {
    return (
      <div className="container mx-auto p-6">
        <button
          onClick={() => setSelectedReturn(null)}
          className="text-muted hover:text-text mb-4 inline-flex items-center gap-2"
        >
          ← العودة للمرتجعات
        </button>
        <ReturnDetail
          returnRequest={selectedReturn}
          onApprove={handleApprove}
          onReject={handleReject}
          onComplete={handleComplete}
          onEdit={() => {/* TODO: Open edit modal */}}
          onDelete={() => {/* TODO: Show delete confirmation */}}
        />
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">المرتجعات</h1>
        <p className="text-muted">إدارة طلبات المرتجعات</p>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <SummaryCard
          label="إجمالي المرتجعات"
          value={summary.totalReturns.toString()}
          icon="🔄"
        />
        <SummaryCard
          label="معلق"
          value={summary.pendingReturns.toString()}
          icon="⏳"
          color={summary.pendingReturns > 0 ? 'warning' : 'success'}
        />
        <SummaryCard
          label="إجمالي الاسترداد"
          value={summary.totalRefundAmount.toFixed(2)}
          icon="💰"
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
            placeholder="بحث عن عميل..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
        </div>

        {/* Status Filter */}
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted self-center">الحالة:</span>
          {(['all', 'pending', 'approved', 'rejected', 'completed'] as StatusFilter[]).map((status) => (
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
              {status === 'pending' && 'معلق'}
              {status === 'approved' && 'موافق'}
              {status === 'rejected' && 'مرفوض'}
              {status === 'completed' && 'مكتمل'}
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
            <option value="customer">العميل</option>
          </select>
        </div>
      </div>

      {/* Returns List */}
      {filteredReturns.length > 0 ? (
        <div className="bg-surface rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted-10 border-b border-border">
              <tr>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">العميل</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">البيع</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">مبلغ الاسترداد</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">السبب</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">التاريخ</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {filteredReturns.map((returnItem) => (
                <tr key={returnItem.id} className="border-b border-border hover:bg-muted-5">
                  <td className="px-4 py-3">
                    <div>
                      <p className="font-medium text-text">{returnItem.customerName}</p>
                      <p className="text-sm text-muted">{returnItem.items.length} بنود</p>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-muted">#{returnItem.saleId.slice(-6)}</td>
                  <td className="px-4 py-3 font-medium text-primary">{returnItem.refundAmount.toFixed(2)}</td>
                  <td className="px-4 py-3 text-muted truncate max-w-xs">{returnItem.reason}</td>
                  <td className="px-4 py-3 text-muted">{returnItem.requestedAt}</td>
                  <td className="px-4 py-3">
                    <StatusBadge status={returnItem.status} />
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => handleViewReturn(returnItem)}
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
          icon="🔄"
          title="لا توجد مرتجعات"
          description="لا توجد طلبات مرتجعات مطابقة للفلاتر الحالية"
          actionLabel="إنشاء طلب مرتج"
          onAction={() => {/* TODO: Open create return modal */}}
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

function StatusBadge({ status }: { status: Return['status'] }) {
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
    <span className={clsx('px-2 py-1 rounded-full text-xs font-medium', statusColors[status])}>
      {statusLabels[status]}
    </span>
  )
}
