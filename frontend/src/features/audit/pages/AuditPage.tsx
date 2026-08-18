import { useState } from 'react'
import { clsx } from 'clsx'
import { AuditLogTable } from '../components/AuditLogTable'
import { AuditLogDetail } from '../components/AuditLogDetail'
import { EmptyState } from '@/components/feedback'
import type { AuditLog, AuditSummary } from '../types/audit'

type EntityTypeFilter = 'all' | 'product' | 'customer' | 'sale' | 'purchase' | 'expense' | 'supplier' | 'inventory' | 'settings' | 'user'
type ActionFilter = 'all' | 'create' | 'update' | 'delete' | 'sale' | 'purchase' | 'login' | 'logout'
type SortOption = 'timestamp' | 'user' | 'action'

export function AuditPage() {
  const [entityTypeFilter, setEntityTypeFilter] = useState<EntityTypeFilter>('all')
  const [actionFilter, setActionFilter] = useState<ActionFilter>('all')
  const [sortBy, setSortBy] = useState<SortOption>('timestamp')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null)

  // TODO: Fetch audit logs from API
  const logs: AuditLog[] = []
  const summary: AuditSummary = {
    totalLogs: 0,
    todayLogs: 0,
    thisWeekLogs: 0,
    thisMonthLogs: 0,
    byAction: [],
    byUser: [],
    byEntityType: [],
  }

  const filteredLogs = logs.filter(log => {
    if (entityTypeFilter !== 'all' && log.entityType !== entityTypeFilter) return false
    if (actionFilter !== 'all' && log.action !== actionFilter) return false
    if (searchQuery && !log.userName.toLowerCase().includes(searchQuery.toLowerCase()) && 
        !log.action.toLowerCase().includes(searchQuery.toLowerCase())) return false
    return true
  }).sort((a, b) => {
    switch (sortBy) {
      case 'timestamp':
        return new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
      case 'user':
        return a.userName.localeCompare(b.userName)
      case 'action':
        return a.action.localeCompare(b.action)
      default:
        return 0
    }
  })

  const handleViewDetails = (log: AuditLog) => {
    setSelectedLog(log)
  }

  if (selectedLog) {
    return (
      <div className="container mx-auto p-6">
        <button
          onClick={() => setSelectedLog(null)}
          className="text-muted hover:text-text mb-4 inline-flex items-center gap-2"
        >
          ← العودة للسجل
        </button>
        <AuditLogDetail
          log={selectedLog}
          onClose={() => setSelectedLog(null)}
        />
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">سجل التدقيق</h1>
        <p className="text-muted">تتبع جميع العمليات والتغييرات في النظام</p>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <SummaryCard
          label="إجمالي السجلات"
          value={summary.totalLogs.toString()}
          icon="📋"
        />
        <SummaryCard
          label="اليوم"
          value={summary.todayLogs.toString()}
          icon="📅"
        />
        <SummaryCard
          label="هذا الأسبوع"
          value={summary.thisWeekLogs.toString()}
          icon="📆"
        />
        <SummaryCard
          label="هذا الشهر"
          value={summary.thisMonthLogs.toString()}
          icon="📊"
        />
      </div>

      {/* Filters */}
      <div className="bg-surface rounded-lg p-4 mb-6 space-y-4">
        {/* Search */}
        <div>
          <input
            type="text"
            placeholder="بحث عن مستخدم أو إجراء..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
        </div>

        {/* Entity Type Filter */}
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted self-center">الكيان:</span>
          <button
            onClick={() => setEntityTypeFilter('all')}
            className={clsx(
              'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
              entityTypeFilter === 'all'
                ? 'bg-primary text-white'
                : 'bg-muted text-muted hover:bg-muted-80'
            )}
          >
            الكل
          </button>
          {(['product', 'customer', 'sale', 'purchase', 'expense', 'supplier', 'inventory', 'settings', 'user'] as EntityTypeFilter[]).map((type) => (
            <button
              key={type}
              onClick={() => setEntityTypeFilter(type)}
              className={clsx(
                'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
                entityTypeFilter === type
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted hover:bg-muted-80'
              )}
            >
              {type === 'product' && 'منتج'}
              {type === 'customer' && 'عميل'}
              {type === 'sale' && 'بيع'}
              {type === 'purchase' && 'مشتراة'}
              {type === 'expense' && 'مصروف'}
              {type === 'supplier' && 'مورد'}
              {type === 'inventory' && 'مخزون'}
              {type === 'settings' && 'إعدادات'}
              {type === 'user' && 'مستخدم'}
            </button>
          ))}
        </div>

        {/* Action Filter */}
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted self-center">الإجراء:</span>
          {(['all', 'create', 'update', 'delete', 'sale', 'purchase', 'login', 'logout'] as ActionFilter[]).map((action) => (
            <button
              key={action}
              onClick={() => setActionFilter(action)}
              className={clsx(
                'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
                actionFilter === action
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted hover:bg-muted-80'
              )}
            >
              {action === 'all' && 'الكل'}
              {action === 'create' && 'إنشاء'}
              {action === 'update' && 'تعديل'}
              {action === 'delete' && 'حذف'}
              {action === 'sale' && 'بيع'}
              {action === 'purchase' && 'مشتراة'}
              {action === 'login' && 'دخول'}
              {action === 'logout' && 'خروج'}
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
            <option value="timestamp">التاريخ</option>
            <option value="user">المستخدم</option>
            <option value="action">الإجراء</option>
          </select>
        </div>
      </div>

      {/* Audit Logs Table */}
      {filteredLogs.length > 0 ? (
        <AuditLogTable
          logs={filteredLogs}
          onViewDetails={handleViewDetails}
        />
      ) : (
        <EmptyState
          icon="📋"
          title="لا يوجد سجلات"
          description="لا توجد سجلات مطابقة للفلاتر الحالية"
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
