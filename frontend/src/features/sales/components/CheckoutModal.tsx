import { useState } from 'react'
import { Card, CardHeader, CardContent, CardFooter } from '../../../components/ui/Card'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { Badge } from '../../../components/ui/Badge'
import { useTranslation } from 'react-i18next'
import { useCustomers, usePaymentMethods, useCreateSale } from '../hooks/useSales'
import { CartItem, PaymentMethod, Customer } from '../types/sales.types'

interface CheckoutModalProps {
  items: CartItem[]
  total: number
  onClose: () => void
  onSuccess: () => void
}

export function CheckoutModal({ items, total, onClose, onSuccess }: CheckoutModalProps) {
  const { t } = useTranslation()
  const { data: customers } = useCustomers()
  const { data: paymentMethods } = usePaymentMethods()
  const createSale = useCreateSale()

  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null)
  const [selectedPayment, setSelectedPayment] = useState<PaymentMethod | null>(null)
  const [paidAmount, setPaidAmount] = useState(total.toString())
  const [notes, setNotes] = useState('')

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('ar-SA', {
      style: 'currency',
      currency: 'ILS',
      minimumFractionDigits: 0,
    }).format(amount)
  }

  const handleCheckout = async () => {
    if (!selectedPayment) return

    const sale = {
      customerId: selectedCustomer?.id,
      customerName: selectedCustomer?.name,
      items,
      subtotal: total,
      tax: 0,
      discount: 0,
      total,
      paymentMethod: selectedPayment.value,
      paymentStatus: 'paid' as const,
      paidAmount: parseFloat(paidAmount),
      remainingAmount: total - parseFloat(paidAmount),
      notes,
    }

    try {
      await createSale.mutateAsync(sale)
      onSuccess()
    } catch (error) {
      console.error('Checkout failed:', error)
    }
  }

  const remaining = total - parseFloat(paidAmount || '0')

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <Card className="w-full max-w-2xl max-h-[90vh] overflow-y-auto scrollbar-thin">
        <CardHeader title={t('sales.checkout')} />
        <CardContent className="space-y-6">
          {/* Customer Selection */}
          <div>
            <label className="block text-sm font-medium text-text mb-2">
              {t('sales.customer')}
            </label>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
              <button
                onClick={() => setSelectedCustomer(null)}
                className={`p-3 rounded-lg border transition-all ${
                  !selectedCustomer
                    ? 'border-primary bg-primary-50'
                    : 'border-border hover:border-primary-300'
                }`}
              >
                <div className="flex items-center gap-2">
                  <span className="text-lg">💵</span>
                  <span className="font-medium">{t('sales.cash')}</span>
                </div>
              </button>
              {customers?.map((customer) => (
                <button
                  key={customer.id}
                  onClick={() => setSelectedCustomer(customer)}
                  className={`p-3 rounded-lg border transition-all ${
                    selectedCustomer?.id === customer.id
                      ? 'border-primary bg-primary-50'
                      : 'border-border hover:border-primary-300'
                  }`}
                >
                  <div className="text-left">
                    <p className="font-medium text-text">{customer.name}</p>
                    <p className="text-xs text-muted">{customer.phone}</p>
                    {customer.outstandingBalance > 0 && (
                      <Badge variant="danger" size="sm" className="mt-1">
                        {formatCurrency(customer.outstandingBalance)} {t('debts.title')}
                      </Badge>
                    )}
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Payment Method */}
          <div>
            <label className="block text-sm font-medium text-text mb-2">
              {t('sales.paymentMethod')}
            </label>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
              {paymentMethods?.map((method) => (
                <button
                  key={method.id}
                  onClick={() => setSelectedPayment(method)}
                  className={`p-3 rounded-lg border transition-all ${
                    selectedPayment?.id === method.id
                      ? 'border-primary bg-primary-50'
                      : 'border-border hover:border-primary-300'
                  }`}
                >
                  <div className="flex flex-col items-center gap-1">
                    <span className="text-2xl">{method.icon}</span>
                    <span className="text-sm font-medium">{method.name}</span>
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Amount */}
          <div>
            <label className="block text-sm font-medium text-text mb-2">
              {t('common.amount')}
            </label>
            <Input
              type="number"
              value={paidAmount}
              onChange={(e) => setPaidAmount(e.target.value)}
              leftIcon="₪"
            />
            {remaining > 0 && (
              <p className="text-sm text-danger mt-1">
                {t('debts.remaining')}: {formatCurrency(remaining)}
              </p>
            )}
          </div>

          {/* Notes */}
          <div>
            <label className="block text-sm font-medium text-text mb-2">
              {t('common.notes')}
            </label>
            <textarea
              className="w-full px-3 py-2 bg-background border border-border rounded-lg text-text placeholder:text-muted focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all duration-150 resize-none"
              rows={3}
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder={t('common.notes')}
            />
          </div>

          {/* Summary */}
          <div className="bg-surface p-4 rounded-lg">
            <div className="flex justify-between items-center">
              <span className="text-lg font-medium">{t('sales.grandTotal')}</span>
              <span className="text-2xl font-bold text-primary">{formatCurrency(total)}</span>
            </div>
          </div>
        </CardContent>
        <CardFooter>
          <div className="flex gap-3 w-full">
            <Button variant="secondary" onClick={onClose} className="flex-1">
              {t('common.cancel')}
            </Button>
            <Button
              onClick={handleCheckout}
              disabled={!selectedPayment || createSale.isLoading}
              className="flex-1"
            >
              {createSale.isLoading ? t('common.loading') : t('sales.checkout')}
            </Button>
          </div>
        </CardFooter>
      </Card>
    </div>
  )
}