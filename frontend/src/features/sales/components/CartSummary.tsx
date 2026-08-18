import { Card, CardHeader, CardContent } from '@components/ui/card'
import { Button } from '@components/ui/button'
import { useTranslation } from 'react-i18next'
import { CartItem } from '../types/sales.types'

interface CartSummaryProps {
  items: CartItem[]
  subtotal: number
  tax: number
  discount: number
  total: number
  onCheckout: () => void
  isLoading?: boolean
}

export function CartSummary({ 
  items, 
  subtotal, 
  tax, 
  discount, 
  total, 
  onCheckout,
  isLoading 
}: CartSummaryProps) {
  const { t } = useTranslation()

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('ar-SA', {
      style: 'currency',
      currency: 'ILS',
      minimumFractionDigits: 0,
    }).format(amount)
  }

  return (
    <Card className="h-full flex flex-col">
      <CardHeader title={t('sales.cart')} />
      <CardContent className="flex-1 space-y-4">
        {/* Items Summary */}
        <div className="space-y-2">
          {items.map((item) => (
            <div key={item.id} className="flex justify-between items-center text-sm">
              <div className="flex-1">
                <p className="font-medium text-text">{item.productName}</p>
                <p className="text-xs text-muted">
                  {item.condition === 'used' ? 'مستعمل' : 'جديد'} × {item.quantity}
                </p>
              </div>
              <p className="font-medium text-text">
                {formatCurrency(item.price * item.quantity)}
              </p>
            </div>
          ))}
        </div>

        {/* Totals */}
        <div className="border-t border-border pt-4 space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-muted">{t('sales.subtotal')}</span>
            <span className="font-medium">{formatCurrency(subtotal)}</span>
          </div>
          
          {tax > 0 && (
            <div className="flex justify-between text-sm">
              <span className="text-muted">{t('sales.tax')}</span>
              <span className="font-medium">{formatCurrency(tax)}</span>
            </div>
          )}
          
          {discount > 0 && (
            <div className="flex justify-between text-sm">
              <span className="text-muted">{t('sales.discount')}</span>
              <span className="font-medium text-success">-{formatCurrency(discount)}</span>
            </div>
          )}
          
          <div className="flex justify-between text-lg font-bold border-t border-border pt-2">
            <span>{t('sales.grandTotal')}</span>
            <span className="text-primary">{formatCurrency(total)}</span>
          </div>
        </div>

        {/* Checkout Button */}
        <Button 
          className="w-full" 
          size="lg"
          onClick={onCheckout}
          disabled={items.length === 0 || isLoading}
        >
          {isLoading && (
            <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          )}
          {isLoading ? t('common.loading') : t('sales.checkout')}
        </Button>
      </CardContent>
    </Card>
  )
}