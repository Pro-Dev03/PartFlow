import { useState } from 'react'
import { clsx } from 'clsx'
import { EmptyState } from '@/components/feedback'
import type { Expense, ExpenseSummary } from '../types/expense'

type CategoryFilter = 'all' | string
type SortOption = 'date' | 'amount' | 'category'

export function ExpensesPage() {
  const [categoryFilter, setCategoryFilter] = useState<CategoryFilter>('all')
  const [sortBy, setSortBy] = useState<SortOption>('date')
  const [searchQuery, setSearchQuery] = useState('')

  // TODO: Fetch expenses from API
  const expenses: Expense[] = []
  const summary: ExpenseSummary = {
    totalExpenses: 0,
    thisMonth: 0,
    byCategory: [],
    monthlyTrend: [],
  }

  const categories = [
    { id: 'rent', name: 'إيجار', color: '#3b82f6' },
    { id: 'utilities', name: 'مرافق', color: '#10b981' },
    { id: 'salaries', name: 'رواتب', color: '#f59e0b' },
    { id: 'shipping', name: 'شحن', color: '#ef4444' },
    { id: 'maintenance', name: 'صيانة', color: '#8b5cf6' },
    { id: 'supplies', name: 'مستلزمات', color: '#06b6d4' },
    { id: 'other', name: 'أخرى', color: '#64748b' },
  ]

  const filteredExpenses = expenses.filter(expense => {
    if (categoryFilter !== 'all' && expense.categoryId !== categoryFilter) return false
    if (searchQuery && !expense.description.toLowerCase().includes(searchQuery.toLowerCase())) return false
    return true
  }).sort((a, b) => {
    switch (sortBy) {
      case 'date':
        return new Date(b.date).getTime() - new Date(a.date).getTime()
      case 'amount':
        return b.amount - a.amount
      case 'category':
        return a.categoryName.localeCompare(b.categoryName)
      default:
        return 0
    }
  })

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">المصروفات</h1>
        <p className="text-muted">إدارة وتتبع المصروفات</p>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <SummaryCard
          label="إجمالي المصروفات"
          value={summary.totalExpenses.toFixed(2)}
          icon="💸"
        />
        <SummaryCard
          label="هذا الشهر"
          value={summary.thisMonth.toFixed(2)}
          icon="📅"
        />
        <SummaryCard
          label="عدد الفئات"
          value={categories.length.toString()}
          icon="📁"
        />
        <SummaryCard
          label="متوسط شهري"
          value={(summary.totalExpenses / 12).toFixed(2)}
          icon="📊"
        />
      </div>

      {/* Category Breakdown */}
      {summary.byCategory.length > 0 && (
        <div className="bg-surface rounded-lg p-6 mb-6 border border-border">
          <h3 className="font-semibold text-text mb-4">توزيع المصروفات حسب الفئة</h3>
          <div className="space-y-3">
            {summary.byCategory.map((category) => (
              <div key={category.categoryId} className="flex items-center gap-4">
                <div className="flex-1">
                  <div className="flex justify-between mb-1">
                    <span className="text-sm font-medium text-text">{category.categoryName}</span>
                    <span className="text-sm text-muted">{category.percentage.toFixed(0)}%</span>
                  </div>
                  <div className="w-full bg-muted-20 rounded-full h-2">
                    <div
                      className="h-2 rounded-full transition-all"
                      style={{
                        width: `${category.percentage}%`,
                        backgroundColor: categories.find(c => c.id === category.categoryId)?.color || '#64748b',
                      }}
                    />
                  </div>
                </div>
                <span className="text-sm font-medium text-text w-24 text-left">
                  {category.amount.toFixed(2)}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="bg-surface rounded-lg p-4 mb-6 space-y-4">
        {/* Search */}
        <div>
          <input
            type="text"
            placeholder="بحث عن مصروف..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
        </div>

        {/* Category Filter */}
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted self-center">الفئة:</span>
          <button
            onClick={() => setCategoryFilter('all')}
            className={clsx(
              'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
              categoryFilter === 'all'
                ? 'bg-primary text-white'
                : 'bg-muted text-muted hover:bg-muted-80'
            )}
          >
            الكل
          </button>
          {categories.map((category) => (
            <button
              key={category.id}
              onClick={() => setCategoryFilter(category.id)}
              className={clsx(
                'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
                categoryFilter === category.id
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted hover:bg-muted-80'
              )}
            >
              {category.name}
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
            <option value="category">الفئة</option>
          </select>
        </div>
      </div>

      {/* Expenses List */}
      {filteredExpenses.length > 0 ? (
        <div className="bg-surface rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted-10 border-b border-border">
              <tr>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">التاريخ</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الفئة</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الوصف</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المبلغ</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {filteredExpenses.map((expense) => (
                <tr key={expense.id} className="border-b border-border hover:bg-muted-5">
                  <td className="px-4 py-3 text-muted">{expense.date}</td>
                  <td className="px-4 py-3">
                    <span className="px-2 py-1 rounded text-xs font-medium" style={{
                      backgroundColor: `${categories.find(c => c.id === expense.categoryId)?.color}20`,
                      color: categories.find(c => c.id === expense.categoryId)?.color,
                    }}>
                      {expense.categoryName}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div>
                      <p className="font-medium text-text">{expense.description}</p>
                      {expense.notes && (
                        <p className="text-sm text-muted">{expense.notes}</p>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3 font-medium text-text">{expense.amount.toFixed(2)}</td>
                  <td className="px-4 py-3">
                    <button className="text-primary hover:text-primary-600 text-sm">
                      عرض
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <EmptyState
          icon="💸"
          title="لا توجد مصروفات"
          description="لا توجد مصروفات مطابقة للفلاتر الحالية"
          actionLabel="إضافة مصروف"
          onAction={() => {/* TODO: Open add expense modal */}}
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
