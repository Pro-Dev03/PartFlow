import { useState } from 'react'
import { clsx } from 'clsx'
import { DebtDetail } from '../components/DebtDetail'
import { EmptyState } from '@/components/feedback'
import type { Debt, DebtPayment, DebtSummary } from '../types/debt'

type StatusFilter = 'all' | 'pending' | 'partial' | 'paid' | 'overdue'
type SortOption = 'dueDate' | 'amount' | 'customer'

export function DebtsPage() {
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [sortBy, setSortBy] = useState<SortOption>('dueDate')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedDebt, setSelectedDebt] = useState<Debt | null>(null)

  // TODO: Fetch debts from API
  const debts: Debt[] = []
  const summary: DebtSummary = {
    totalDebts: 0,
    totalAmount: 0,
    totalPaid: 0,
    totalRemaining: 0,
    overdueAmount: 0,
    overdueCount: 0,
    dueThisWeek: 0,
    dueThisMonth: 0,
  }

  const filteredDebts = debts.filter(debt => {
    if (statusFilter !== 'all' && debt.status !== statusFilter) return false
    if (searchQuery && !debt.customerName.toLowerCase().includes(searchQuery.toLowerCase())) return false
    return true
  }).sort((a, b) => {
    switch (sortBy) {
      case 'dueDate':
        return new Date(a.dueDate).getTime() - new Date(b.dueDate).getTime()
      case 'amount':
        return b.amount - a.amount
      case 'customer':
        return a.customerName.localeCompare(b.customerName)
      default:
        return 0
    }
  })

  const handleViewDebt = (debt: Debt) => {
    setSelectedDebt(debt)
  }

  const handleAddPayment = () => {
    // TODO: Open payment modal
    console.log('Add payment')
  }

  const handleSendReminder = () => {
    // TODO: Send reminder
    console.log('Send reminder')
  }

  if (selectedDebt) {
    return (
      <div className="container mx-auto p-6">
        <button
          onClick={() => setSelectedDebt(null)}
          className="text-muted hover:text-text mb-4 inline-flex items-center gap-2"
        >
          ← العودة للديون
        </button>
        <DebtDetail
          debt={selectedDebt}
          payments={[]} // TODO: Fetch payments
          onAddPayment={handleAddPayment}
          onSendReminder={handleSendReminder}
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
        <h1 className="text-2xl font-bold text-text mb-2">الديون</h1>
        <p className="text-muted">إدارة ومتابعة الديون والمدفوعات</p>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <SummaryCard
          label="إجمالي الديون"
          value={summary.totalDebts.toString()}
          icon="📋"
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
          label="المتأخر"
          value={summary.overdueAmount.toFixed(2)}
          icon="🚨"
          color={summary.overdueCount > 0 ? 'danger' : 'success'}
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
          {(['all', 'pending', 'partial', 'paid', 'overdue'] as StatusFilter[]).map((status) => (
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
              {status === 'partial' && 'جزئي'}
              {status === 'paid' && 'مدفوع'}
              {status === 'overdue' && 'متأخر'}
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
            <option value="dueDate">تاريخ الاستحقاق</option>
            <option value="amount">المبلغ</option>
            <option value="customer">العميل</option>
          </select>
        </div>
      </div>

      {/* Debts List */}
      {filteredDebts.length > 0 ? (
        <div className="bg-surface rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted-10 border-b border-border">
              <tr>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">العميل</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المبلغ</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المدفوع</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المتبقي</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الاستحقاق</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {filteredDebts.map((debt) => (
                <tr key={debt.id} className="border-b border-border hover:bg-muted-5">
                  <td className="px-4 py-3">
                    <div>
                      <p className="font-medium text-text">{debt.customerName}</p>
                      <p className="text-sm text-muted">{debt.customerPhone}</p>
                    </div>
                  </td>
                  <td className="px-4 py-3 font-medium text-text">{debt.amount.toFixed(2)}</td>
                  <td className="px-4 py-3 text-success">{debt.paidAmount.toFixed(2)}</td>
                  <td className="px-4 py-3 font-medium text-warning">{debt.remainingAmount.toFixed(2)}</td>
                  <td className="px-4 py-3 text-muted">{debt.dueDate}</td>
                  <td className="px-4 py-3">
                    <StatusBadge status={debt.status} />
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => handleViewDebt(debt)}
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
          icon="📋"
          title="لا توجد ديون"
          description="لا توجد ديون مطابقة للفلاتر الحالية"
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

function StatusBadge({ status }: { status: Debt['status'] }) {
  const statusColors = {
    pending: 'bg-warning-10 text-warning',
    partial: 'bg-primary-10 text-primary',
    paid: 'bg-success-10 text-success',
    overdue: 'bg-danger-10 text-danger',
  }

  const statusLabels = {
    pending: 'معلق',
    partial: 'جزئي',
    paid: 'مدفوع',
    overdue: 'متأخر',
  }

  return (
    <span className={clsx('px-2 py-1 rounded-full text-xs font-medium', statusColors[status])}>
      {statusLabels[status]}
    </span>
  )
}

export default DebtsPage
