import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { clsx } from 'clsx'
import { CustomerDetail } from './components/CustomerDetail'
import { EmptyState } from '../../components/feedback/EmptyState'
import { Button } from '@components/ui/button'
import { Card, CardHeader } from '@components/ui/card'
import type { Customer, CustomerStats } from './types/customer'

type FilterType = 'all' | 'active' | 'inactive' | 'has_debt' | 'vip'
type SortOption = 'name' | 'total_spent' | 'purchases' | 'date'

export function Customers() {
  const { t } = useTranslation()
  const [filter, setFilter] = useState<FilterType>('all')
  const [sort, setSort] = useState<SortOption>('name')
  const [search, setSearch] = useState('')
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null)
  const [showMobileView, setShowMobileView] = useState(false)

  // Mock data - TODO: Replace with API calls
  const customers: Customer[] = [
    {
      id: '1',
      name: 'أحمد محمد',
      phone: '050-1234567',
      email: 'ahmed@example.com',
      address: 'شارع الملك فهد، الرياض',
      city: 'الرياض',
      notes: 'عميل VIP، يفضل الدفع نقداً',
      createdAt: '2024-01-15',
      updatedAt: '2024-08-18',
      totalPurchases: 15,
      totalSpent: 8450,
      outstandingBalance: 1250,
      lastPurchaseDate: '2024-08-16',
      isActive: true,
    },
    {
      id: '2',
      name: 'سارة علي',
      phone: '050-7654321',
      email: 'sara@example.com',
      address: 'حي النخيل، جدة',
      city: 'جدة',
      notes: 'تهتم بالقطع المستعملة',
      createdAt: '2024-02-20',
      updatedAt: '2024-08-15',
      totalPurchases: 8,
      totalSpent: 3200,
      outstandingBalance: 0,
      lastPurchaseDate: '2024-08-10',
      isActive: true,
    },
    {
      id: '3',
      name: 'خالد عبدالله',
      phone: '055-9876543',
      email: 'khaled@example.com',
      address: 'حي الملز، الدمام',
      city: 'الدمام',
      notes: 'ديون متأخرة - يحتاج متابعة',
      createdAt: '2024-03-10',
      updatedAt: '2024-08-12',
      totalPurchases: 5,
      totalSpent: 1800,
      outstandingBalance: 750,
      lastPurchaseDate: '2024-07-20',
      isActive: true,
    },
  ]

  const customerStats: Record<string, CustomerStats> = {
    '1': {
      totalPurchases: 15,
      totalSpent: 8450,
      averagePurchaseValue: 8450 / 15,
      outstandingBalance: 1250,
      paymentHistory: {
        onTime: 80,
        late: 15,
        overdue: 5,
      },
    },
    '2': {
      totalPurchases: 8,
      totalSpent: 3200,
      averagePurchaseValue: 3200 / 8,
      outstandingBalance: 0,
      paymentHistory: {
        onTime: 95,
        late: 5,
        overdue: 0,
      },
    },
    '3': {
      totalPurchases: 5,
      totalSpent: 1800,
      averagePurchaseValue: 1800 / 5,
      outstandingBalance: 750,
      paymentHistory: {
        onTime: 60,
        late: 20,
        overdue: 20,
      },
    },
  }

  const filteredCustomers = customers.filter((customer) => {
    // Search filter
    if (search && !customer.name.toLowerCase().includes(search.toLowerCase()) && 
        !customer.phone.includes(search) && 
        !(customer.email && customer.email.toLowerCase().includes(search.toLowerCase()))) {
      return false
    }

    // Status filter
    switch (filter) {
      case 'active':
        return customer.isActive
      case 'inactive':
        return !customer.isActive
      case 'has_debt':
        return customer.outstandingBalance > 0
      case 'vip':
        return customer.totalSpent > 5000
      default:
        return true
    }
  }).sort((a, b) => {
    switch (sort) {
      case 'name':
        return a.name.localeCompare(b.name)
      case 'total_spent':
        return b.totalSpent - a.totalSpent
      case 'purchases':
        return b.totalPurchases - a.totalPurchases
      case 'date':
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      default:
        return 0
    }
  })

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('ar-SA', {
      style: 'currency',
      currency: 'ILS',
      minimumFractionDigits: 0,
    }).format(amount)
  }

  const hasDebt = (customer: Customer) => customer.outstandingBalance > 0
  const isVip = (customer: Customer) => customer.totalSpent > 5000

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text">{t('customers.title')}</h1>
          <p className="text-muted mt-1">{filteredCustomers.length} عميل</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setShowMobileView(!showMobileView)}>
            {showMobileView ? '📊' : '📱'}
          </Button>
          <Button variant="primary">
            + {t('customers.addCustomer')}
          </Button>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="إجمالي العملاء" value={customers.length} icon="👥" color="primary" />
        <StatCard label="نشطون" value={customers.filter(c => c.isActive).length} icon="✅" color="success" />
        <StatCard label="لديهم ديون" value={customers.filter(c => hasDebt(c)).length} icon="💳" color="warning" />
        <StatCard label="VIP" value={customers.filter(c => isVip(c)).length} icon="⭐" color="info" />
      </div>

      {/* Filters */}
      <div className="bg-surface rounded-lg p-4 space-y-4">
        {/* Search */}
        <div>
          <input
            type="text"
            placeholder="بحث بالاسم، الهاتف، أو البريد..."
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
            { value: 'active', label: 'نشط' },
            { value: 'inactive', label: 'غير نشط' },
            { value: 'has_debt', label: 'لديهم ديون' },
            { value: 'vip', label: 'VIP' },
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
            <option value="name">الاسم</option>
            <option value="total_spent">إجمالي الإنفاق</option>
            <option value="purchases">عدد المشتريات</option>
            <option value="date">تاريخ التسجيل</option>
          </select>
        </div>
      </div>

      {/* Customers Table/Desktop View */}
      {!showMobileView && (
        <Card>
          <CardHeader title={t('customers.list')} />
          {filteredCustomers.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-muted-10 border-b border-border">
                  <tr>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">العميل</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الهاتف</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">المشتريات</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الإنفاق</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الرصيد</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredCustomers.map((customer) => (
                    <tr 
                      key={customer.id} 
                      className="border-b border-border hover:bg-muted-5 cursor-pointer"
                      onClick={() => setSelectedCustomer(customer)}
                    >
                      <td className="px-4 py-3">
                        <div>
                          <div className="flex items-center gap-2">
                            <p className="font-medium text-text">{customer.name}</p>
                            {isVip(customer) && (
                              <span className="text-xs">⭐</span>
                            )}
                          </div>
                          {customer.email && (
                            <p className="text-sm text-muted">{customer.email}</p>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3 text-muted">{customer.phone}</td>
                      <td className="px-4 py-3 font-medium text-text">{customer.totalPurchases}</td>
                      <td className="px-4 py-3 font-medium text-text">{formatCurrency(customer.totalSpent)}</td>
                      <td className="px-4 py-3">
                        <span className={clsx(
                          'font-medium',
                          hasDebt(customer) ? 'text-warning' : 'text-success'
                        )}>
                          {formatCurrency(customer.outstandingBalance)}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={clsx(
                          'px-2 py-1 rounded text-xs font-medium',
                          customer.isActive ? 'bg-success-10 text-success' : 'bg-muted-10 text-muted'
                        )}>
                          {customer.isActive ? 'نشط' : 'غير نشط'}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex gap-2">
                          <button className="text-primary hover:text-primary-600 text-sm">
                            عرض
                          </button>
                          {hasDebt(customer) && (
                            <button className="text-warning hover:text-warning-600 text-sm">
                              تسديد
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState
              icon="👥"
              title="لا يوجد عملاء"
              description="لا يوجد عملاء مطابقين للفلاتر الحالية"
              actionLabel="إضافة عميل"
              onAction={() => {}}
            />
          )}
        </Card>
      )}

      {/* Mobile Cards View */}
      {showMobileView && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredCustomers.map((customer) => (
            <Card 
              key={customer.id}
              className="cursor-pointer hover:shadow-md transition-shadow"
              onClick={() => setSelectedCustomer(customer)}
            >
              <div className="p-4">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold text-text">{customer.name}</h3>
                      {isVip(customer) && (
                        <span className="text-xs">⭐</span>
                      )}
                    </div>
                    <p className="text-sm text-muted">{customer.phone}</p>
                  </div>
                  <span className={clsx(
                    'px-2 py-1 rounded text-xs font-medium',
                    customer.isActive ? 'bg-success-10 text-success' : 'bg-muted-10 text-muted'
                  )}>
                    {customer.isActive ? 'نشط' : 'غير نشط'}
                  </span>
                </div>
                
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted">المشتريات:</span>
                    <span className="font-medium text-text">{customer.totalPurchases}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted">الإنفاق:</span>
                    <span className="font-medium text-text">{formatCurrency(customer.totalSpent)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted">الرصيد:</span>
                    <span className={clsx(
                      'font-medium',
                      hasDebt(customer) ? 'text-warning' : 'text-success'
                    )}>
                      {formatCurrency(customer.outstandingBalance)}
                    </span>
                  </div>
                </div>

                <div className="flex gap-2 mt-4">
                  <Button 
                    variant="primary" 
                    size="sm" 
                    className="flex-1"
                    onClick={(e) => { e.stopPropagation(); }}
                  >
                    عرض
                  </Button>
                  {hasDebt(customer) && (
                    <Button 
                      variant="warning" 
                      size="sm"
                      onClick={(e) => { e.stopPropagation(); }}
                    >
                      تسديد
                    </Button>
                  )}
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Customer Detail Drawer */}
      {selectedCustomer && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="bg-surface rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-auto">
            <div className="sticky top-0 bg-surface border-b border-border p-4 flex items-center justify-between">
              <h2 className="text-xl font-bold text-text">تفاصيل العميل</h2>
              <button
                onClick={() => setSelectedCustomer(null)}
                className="p-2 hover:bg-muted-10 rounded-lg transition-colors"
              >
                ✕
              </button>
            </div>
            <div className="p-6">
              <CustomerDetail
                customer={selectedCustomer}
                stats={customerStats[selectedCustomer.id] || {
                  totalPurchases: 0,
                  totalSpent: 0,
                  averagePurchaseValue: 0,
                  outstandingBalance: 0,
                  paymentHistory: { onTime: 0, late: 0, overdue: 0 },
                }}
                onEdit={() => {}}
                onDelete={() => {}}
                onAddPayment={() => {}}
                onCreateSale={() => {}}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function StatCard({ label, value, icon, color }: { label: string; value: string | number; icon: string; color: 'primary' | 'success' | 'warning' | 'danger' | 'info' }) {
  const colorClasses = {
    primary: 'text-primary',
    success: 'text-success',
    warning: 'text-warning',
    danger: 'text-danger',
    info: 'text-info',
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

export default Customers
