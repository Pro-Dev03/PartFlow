import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { clsx } from 'clsx'
import { EmptyState } from '../../components/feedback/EmptyState'
import { Button } from '@components/ui/button'
import { Card, CardHeader } from '@components/ui/card'

interface Debt {
  id: string
  customerId: string
  customerName: string
  customerPhone: string
  amount: number
  paidAmount: number
  remainingAmount: number
  dueDate: string
  lastPaymentDate?: string
  status: 'current' | 'overdue' | 'paid'
  notes?: string
  createdAt: string
}

type FilterType = 'all' | 'current' | 'overdue' | 'paid'
type SortOption = 'amount' | 'due_date' | 'customer' | 'date'

export function Debts() {
  const { t } = useTranslation()
  const [filter, setFilter] = useState<FilterType>('all')
  const [sort, setSort] = useState<SortOption>('due_date')
  const [search, setSearch] = useState('')
  const [selectedDebt, setSelectedDebt] = useState<Debt | null>(null)
  const [showPaymentModal, setShowPaymentModal] = useState(false)
  const [paymentAmount, setPaymentAmount] = useState('')

  // Mock data - TODO: Replace with API calls
  const debts: Debt[] = [
    {
      id: '1',
      customerId: '1',
      customerName: 'أحمد محمد',
      customerPhone: '050-1234567',
      amount: 2000,
      paidAmount: 750,
      remainingAmount: 1250,
      dueDate: '2024-08-20',
      lastPaymentDate: '2024-08-16',
      status: 'current',
      notes: 'دفعة شهرية',
      createdAt: '2024-07-15',
    },
    {
      id: '2',
      customerId: '3',
      customerName: 'خالد عبدالله',
      customerPhone: '055-9876543',
      amount: 1500,
      paidAmount: 750,
      remainingAmount: 750,
      dueDate: '2024-08-10',
      lastPaymentDate: '2024-07-25',
      status: 'overdue',
      notes: 'ديون متأخرة - يحتاج متابعة',
      createdAt: '2024-06-20',
    },
    {
      id: '3',
      customerId: '2',
      customerName: 'سارة علي',
      customerPhone: '050-7654321',
      amount: 800,
      paidAmount: 800,
      remainingAmount: 0,
      dueDate: '2024-08-05',
      lastPaymentDate: '2024-08-05',
      status: 'paid',
      createdAt: '2024-07-01',
    },
  ]

  const filteredDebts = debts.filter((debt) => {
    // Search filter
    if (search && !debt.customerName.toLowerCase().includes(search.toLowerCase()) && 
        !debt.customerPhone.includes(search)) {
      return false
    }

    // Status filter
    switch (filter) {
      case 'current':
        return debt.status === 'current'
      case 'overdue':
        return debt.status === 'overdue'
      case 'paid':
        return debt.status === 'paid'
      default:
        return true
    }
  }).sort((a, b) => {
    switch (sort) {
      case 'amount':
        return b.remainingAmount - a.remainingAmount
      case 'due_date':
        return new Date(a.dueDate).getTime() - new Date(b.dueDate).getTime()
      case 'customer':
        return a.customerName.localeCompare(b.customerName)
      case 'date':
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      default:
        return 0
    }
  })

  const handlePayment = () => {
    if (selectedDebt && paymentAmount) {
      const amount = parseFloat(paymentAmount)
      console.log('Processing payment:', amount, 'for debt:', selectedDebt.id)
      // TODO: Implement payment API call
      setShowPaymentModal(false)
      setPaymentAmount('')
      setSelectedDebt(null)
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('ar-SA', {
      style: 'currency',
      currency: 'ILS',
      minimumFractionDigits: 0,
    }).format(amount)
  }

  const isOverdue = (debt: Debt) => {
    return debt.status === 'overdue' || (debt.status !== 'paid' && new Date(debt.dueDate) < new Date())
  }

  const getDaysUntilDue = (dueDate: string) => {
    const today = new Date()
    const due = new Date(dueDate)
    const diffTime = due.getTime() - today.getTime()
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
    return diffDays
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text">{t('debts.title')}</h1>
          <p className="text-muted mt-1">{filteredDebts.length} دين</p>
        </div>
        <Button variant="primary">
          + {t('debts.addDebt')}
        </Button>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="إجمالي الديون" value={formatCurrency(debts.reduce((sum, d) => sum + d.remainingAmount, 0))} icon="💳" color="primary" />
        <StatCard label="ديون مستحقة" value={debts.filter(d => d.status === 'current').length.toString()} icon="📋" color="warning" />
        <StatCard label="ديون متأخرة" value={debts.filter(d => d.status === 'overdue').length.toString()} icon="🔴" color="danger" />
        <StatCard label="مدفوعة" value={debts.filter(d => d.status === 'paid').length.toString()} icon="✅" color="success" />
      </div>

      {/* Filters */}
      <div className="bg-surface rounded-lg p-4 space-y-4">
        {/* Search */}
        <div>
          <input
            type="text"
            placeholder="بحث بالاسم أو الهاتف..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
        </div>

        {/* Filter Buttons */}
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted self-center">الفلتر:</span>
          {([
            { value: 'all', label: 'الكل' },
            { value: 'current', label: 'مستحقة' },
            { value: 'overdue', label: 'متأخرة' },
            { value: 'paid', label: 'مدفوعة' },
          ] as const).map((option) => (
            <button
              key={option.value}
              onClick={() => setFilter(option.value)}
              className={clsx(
                'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
                filter === option.value
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted hover:bg-muted-80'
              )}
            >
              {option.label}
            </button>
          ))}
        </div>

        {/* Sort */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted">ترتيب حسب:</span>
          <select
            value={sort}
            onChange={(e) => setSort(e.target.value as SortOption)}
            className="px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="amount">المبلغ المتبقي</option>
            <option value="due_date">تاريخ الاستحقاق</option>
            <option value="customer">اسم العميل</option>
            <option value="date">تاريخ الإنشاء</option>
          </select>
        </div>
      </div>

      {/* Debts Table */}
      <Card>
        <CardHeader title={t('debts.list')} />
        {filteredDebts.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-muted-10 border-b border-border">
                <tr>
                  <th className="px-4 py-3 text-right text-sm font-medium text-muted">العميل</th>
                  <th className="px-4 py-3 text-right text-sm font-medium text-muted">المبلغ الكلي</th>
                  <th className="px-4 py-3 text-right text-sm font-medium text-muted">المدفوع</th>
                  <th className="px-4 py-3 text-right text-sm font-medium text-muted">المتبقي</th>
                  <th className="px-4 py-3 text-right text-sm font-medium text-muted">تاريخ الاستحقاق</th>
                  <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                  <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
                </tr>
              </thead>
              <tbody>
                {filteredDebts.map((debt) => (
                  <tr 
                    key={debt.id} 
                    className={clsx(
                      'border-b border-border hover:bg-muted-5',
                      isOverdue(debt) && 'bg-danger-5'
                    )}
                  >
                    <td className="px-4 py-3">
                      <div>
                        <p className="font-medium text-text">{debt.customerName}</p>
                        <p className="text-sm text-muted">{debt.customerPhone}</p>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-muted">{formatCurrency(debt.amount)}</td>
                    <td className="px-4 py-3 text-success">{formatCurrency(debt.paidAmount)}</td>
                    <td className="px-4 py-3">
                      <span className={clsx(
                        'font-bold',
                        debt.remainingAmount > 0 ? 'text-warning' : 'text-success'
                      )}>
                        {formatCurrency(debt.remainingAmount)}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <div>
                        <p className="text-muted">{debt.dueDate}</p>
                        {debt.status !== 'paid' && (
                          <p className={clsx(
                            'text-xs',
                            getDaysUntilDue(debt.dueDate) < 0 ? 'text-danger' : 
                            getDaysUntilDue(debt.dueDate) <= 7 ? 'text-warning' : 'text-muted'
                          )}>
                            {getDaysUntilDue(debt.dueDate) < 0 ? `متأخر ${Math.abs(getDaysUntilDue(debt.dueDate))} يوم` :
                             getDaysUntilDue(debt.dueDate) === 0 ? 'يستحق اليوم' :
                             getDaysUntilDue(debt.dueDate) <= 7 ? `بعد ${getDaysUntilDue(debt.dueDate)} أيام` :
                             `بعد ${getDaysUntilDue(debt.dueDate)} يوم`}
                          </p>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className={clsx(
                        'px-2 py-1 rounded text-xs font-medium',
                        debt.status === 'paid' ? 'bg-success-10 text-success' :
                        debt.status === 'overdue' ? 'bg-danger-10 text-danger' :
                        'bg-warning-10 text-warning'
                      )}>
                        {debt.status === 'paid' ? 'مدفوع' :
                         debt.status === 'overdue' ? 'متأخر' : 'مستحق'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex gap-2">
                        {debt.status !== 'paid' && (
                          <button
                            onClick={() => {
                              setSelectedDebt(debt)
                              setShowPaymentModal(true)
                            }}
                            className="text-primary hover:text-primary-600 text-sm"
                          >
                            تسديد
                          </button>
                        )}
                        <button className="text-muted hover:text-text text-sm">
                          عرض
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <EmptyState
            icon="💳"
            title="لا توجد ديون"
            description="لا توجد ديون مطابقة للفلاتر الحالية"
            actionLabel="إضافة دين"
            onAction={() => {}}
          />
        )}
      </Card>

      {/* Payment Modal */}
      {showPaymentModal && selectedDebt && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="bg-surface rounded-lg shadow-xl max-w-md w-full">
            <div className="p-6">
              <h2 className="text-xl font-bold text-text mb-4">تسديد دفعة</h2>
              
              <div className="bg-muted-10 rounded-lg p-4 mb-4">
                <div className="flex justify-between mb-2">
                  <span className="text-muted">العميل:</span>
                  <span className="font-medium text-text">{selectedDebt.customerName}</span>
                </div>
                <div className="flex justify-between mb-2">
                  <span className="text-muted">المبلغ الكلي:</span>
                  <span className="font-medium text-text">{formatCurrency(selectedDebt.amount)}</span>
                </div>
                <div className="flex justify-between mb-2">
                  <span className="text-muted">المدفوع سابقاً:</span>
                  <span className="font-medium text-success">{formatCurrency(selectedDebt.paidAmount)}</span>
                </div>
                <div className="flex justify-between border-t border-border pt-2">
                  <span className="text-muted font-medium">المتبقي:</span>
                  <span className="font-bold text-warning">{formatCurrency(selectedDebt.remainingAmount)}</span>
                </div>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-text mb-2">
                    مبلغ الدفعة
                  </label>
                  <input
                    type="number"
                    value={paymentAmount}
                    onChange={(e) => setPaymentAmount(e.target.value)}
                    placeholder="أدخل المبلغ"
                    max={selectedDebt.remainingAmount}
                    className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                    required
                  />
                </div>

                <div className="flex gap-2">
                  <Button
                    variant="primary"
                    className="flex-1"
                    onClick={handlePayment}
                    disabled={!paymentAmount || parseFloat(paymentAmount) <= 0}
                  >
                    تأكيد الدفعة
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => {
                      setShowPaymentModal(false)
                      setPaymentAmount('')
                      setSelectedDebt(null)
                    }}
                  >
                    إلغاء
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function StatCard({ label, value, icon, color }: { label: string; value: string; icon: string; color: 'primary' | 'success' | 'warning' | 'danger' }) {
  const colorClasses = {
    primary: 'text-primary',
    success: 'text-success',
    warning: 'text-warning',
    danger: 'text-danger',
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

export default Debts
