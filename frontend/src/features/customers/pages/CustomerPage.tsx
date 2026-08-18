import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { CustomerDetail } from '../components/CustomerDetail'
import { EmptyState } from '@/components/feedback'
import type { Customer, CustomerStats } from '../types/customer'

export function CustomerPage() {
  const { id } = useParams<{ id: string }>()
  const [customer, setCustomer] = useState<Customer | null>(null)
  const [stats, setStats] = useState<CustomerStats | null>(null)
  const [loading, setLoading] = useState(true)

  // TODO: Fetch customer data from API
  // useEffect(() => {
  //   Promise.all([fetchCustomer(id), fetchCustomerStats(id)])
  //     .then(([customerData, statsData]) => {
  //       setCustomer(customerData)
  //       setStats(statsData)
  //     })
  //     .finally(() => setLoading(false))
  // }, [id])

  if (loading) {
    return <div className="p-8">جاري التحميل...</div>
  }

  if (!customer || !stats) {
    return (
      <EmptyState
        icon="👤"
        title="العميل غير موجود"
        description="لم يتم العثور على العميل المطلوب"
        actionLabel="العودة للعملاء"
        onAction={() => window.history.back()}
      />
    )
  }

  return (
    <div className="container mx-auto p-6">
      <div className="mb-6">
        <button
          onClick={() => window.history.back()}
          className="text-muted hover:text-text mb-4 inline-flex items-center gap-2"
        >
          ← العودة
        </button>
      </div>

      <CustomerDetail
        customer={customer}
        stats={stats}
        onEdit={() => {/* TODO: Open edit modal */}}
        onDelete={() => {/* TODO: Show delete confirmation */}}
        onAddPayment={() => {/* TODO: Open payment modal */}}
        onCreateSale={() => {/* TODO: Navigate to new sale with customer */}}
      />
    </div>
  )
}
