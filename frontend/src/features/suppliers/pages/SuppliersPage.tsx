import { useState } from 'react'
import { clsx } from 'clsx'
import { SupplierDetail } from '../components/SupplierDetail'
import { EmptyState } from '@/components/feedback'
import type { Supplier, SupplierStats } from '../types/supplier'

export function SuppliersPage() {
  const [selectedSupplier, setSelectedSupplier] = useState<Supplier | null>(null)
  const [stats, setStats] = useState<SupplierStats | null>(null)
  const [loading, setLoading] = useState(false)

  // TODO: Fetch suppliers from API
  const suppliers: Supplier[] = []

  const handleViewSupplier = (supplier: Supplier) => {
    setSelectedSupplier(supplier)
    // TODO: Fetch supplier stats
    setStats({
      totalPurchases: supplier.totalPurchases,
      totalAmount: supplier.totalAmount,
      averagePurchaseValue: supplier.totalAmount / Math.max(supplier.totalPurchases, 1),
      outstandingBalance: supplier.outstandingBalance,
      paymentHistory: {
        onTime: 85,
        late: 10,
        overdue: 5,
      },
    })
  }

  if (selectedSupplier && stats) {
    return (
      <div className="container mx-auto p-6">
        <button
          onClick={() => setSelectedSupplier(null)}
          className="text-muted hover:text-text mb-4 inline-flex items-center gap-2"
        >
          ← العودة للموردين
        </button>
        <SupplierDetail
          supplier={selectedSupplier}
          stats={stats}
          onEdit={() => {/* TODO: Open edit modal */}}
          onDelete={() => {/* TODO: Show delete confirmation */}}
          onAddPayment={() => {/* TODO: Open payment modal */}}
          onCreatePurchase={() => {/* TODO: Navigate to new purchase with supplier */}}
        />
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">الموردون</h1>
        <p className="text-muted">إدارة الموردين والمشتريات</p>
      </div>

      {/* Suppliers List */}
      {suppliers.length > 0 ? (
        <div className="bg-surface rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted-10 border-b border-border">
              <tr>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المورد</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الهاتف</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجمالي المشتريات</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الرصيد المستحق</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">آخر شراء</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {suppliers.map((supplier) => (
                <tr key={supplier.id} className="border-b border-border hover:bg-muted-5">
                  <td className="px-4 py-3">
                    <div>
                      <p className="font-medium text-text">{supplier.name}</p>
                      {supplier.email && (
                        <p className="text-sm text-muted">{supplier.email}</p>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-muted">{supplier.phone}</td>
                  <td className="px-4 py-3 font-medium text-text">
                    {supplier.totalPurchases} ({supplier.totalAmount.toFixed(2)})
                  </td>
                  <td className="px-4 py-3">
                    <span className={clsx(
                      'font-medium',
                      supplier.outstandingBalance > 0 ? 'text-warning' : 'text-success'
                    )}>
                      {supplier.outstandingBalance.toFixed(2)}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-muted">
                    {supplier.lastPurchaseDate || 'لا يوجد'}
                  </td>
                  <td className="px-4 py-3">
                    <span className={clsx(
                      'px-2 py-1 rounded-full text-xs font-medium',
                      supplier.isActive ? 'bg-success-10 text-success' : 'bg-muted-10 text-muted'
                    )}>
                      {supplier.isActive ? 'نشط' : 'غير نشط'}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => handleViewSupplier(supplier)}
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
          icon="🚚"
          title="لا يوجد موردين"
          description="ابدأ بإضافة موردين للمتجر"
          actionLabel="إضافة مورد"
          onAction={() => {/* TODO: Open add supplier modal */}}
        />
      )}
    </div>
  )
}
