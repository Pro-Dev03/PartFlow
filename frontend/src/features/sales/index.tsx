import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardHeader } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { ProductSearch } from './components/ProductSearch'
import { CartSummary } from './components/CartSummary'
import { CheckoutModal } from './components/CheckoutModal'
import { useCart } from './hooks/useSales'
import { CartItem } from './types/sales.types'

export function Sales() {
  const { t } = useTranslation()
  const {
    cart,
    addToCart,
    removeFromCart,
    updateQuantity,
    clearCart,
    cartTotal,
    cartProfit,
    itemCount,
  } = useCart()

  const [showCheckout, setShowCheckout] = useState(false)
  const [isProcessing, setIsProcessing] = useState(false)

  const handleAddToCart = (item: CartItem) => {
    addToCart(item)
  }

  const handleCheckout = () => {
    setShowCheckout(true)
  }

  const handleCheckoutSuccess = () => {
    setShowCheckout(false)
    clearCart()
    setIsProcessing(false)
    // Show success message
  }

  const handleRemoveFromCart = (itemId: string) => {
    removeFromCart(itemId)
  }

  const handleUpdateQuantity = (itemId: string, quantity: number) => {
    updateQuantity(itemId, quantity)
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('ar-SA', {
      style: 'currency',
      currency: 'ILS',
      minimumFractionDigits: 0,
    }).format(amount)
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text">{t('sales.title')}</h1>
          <p className="text-muted mt-1">{t('sales.newSale')}</p>
        </div>
        <Button onClick={() => clearCart()} disabled={cart.length === 0}>
          {t('common.reset')}
        </Button>
      </div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - Product Search */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader title={t('common.search')} />
            <div className="p-4">
              <ProductSearch onAddToCart={handleAddToCart} />
            </div>
          </Card>

          {/* Cart Items */}
          {cart.length > 0 && (
            <Card className="mt-6">
              <CardHeader 
                title={t('sales.cart')} 
                description={`${itemCount} ${t('inventory.products')}`}
              />
              <div className="divide-y divide-border">
                {cart.map((item) => (
                  <div key={item.id} className="p-4">
                    <div className="flex items-start gap-4">
                      <div className="w-16 h-16 rounded-lg bg-primary-100 flex items-center justify-center flex-shrink-0">
                        <span className="text-3xl">📦</span>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-start justify-between gap-2">
                          <div className="flex-1">
                            <h4 className="font-medium text-text">{item.productName}</h4>
                            <p className="text-sm text-muted">
                              {item.barcode} • {item.condition === 'used' ? 'مستعمل' : 'جديد'}
                            </p>
                            {item.serialNumber && (
                              <p className="text-xs text-muted">SN: {item.serialNumber}</p>
                            )}
                          </div>
                          <div className="text-right">
                            <p className="font-bold text-primary">{formatCurrency(item.price)}</p>
                            <p className="text-xs text-muted">
                              الربح: {formatCurrency(item.price - item.cost)}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-center justify-between mt-3">
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => handleUpdateQuantity(item.id, item.quantity - 1)}
                              className="w-8 h-8 rounded-lg bg-surface border border-border flex items-center justify-center hover:bg-background transition-colors"
                              disabled={item.quantity <= 1}
                            >
                              -
                            </button>
                            <span className="w-8 text-center font-medium">{item.quantity}</span>
                            <button
                              onClick={() => handleUpdateQuantity(item.id, item.quantity + 1)}
                              className="w-8 h-8 rounded-lg bg-surface border border-border flex items-center justify-center hover:bg-background transition-colors"
                              disabled={item.quantity >= item.availableStock}
                            >
                              +
                            </button>
                          </div>
                          <Button
                            variant="danger"
                            size="sm"
                            onClick={() => handleRemoveFromCart(item.id)}
                          >
                            {t('common.delete')}
                          </Button>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </Card>
          )}
        </div>

        {/* Right Column - Cart Summary */}
        <div className="lg:col-span-1">
          <CartSummary
            items={cart}
            subtotal={cartTotal}
            tax={0}
            discount={0}
            total={cartTotal}
            onCheckout={handleCheckout}
            isLoading={isProcessing}
          />
        </div>
      </div>

      {/* Checkout Modal */}
      {showCheckout && (
        <CheckoutModal
          items={cart}
          total={cartTotal}
          onClose={() => setShowCheckout(false)}
          onSuccess={handleCheckoutSuccess}
        />
      )}
    </div>
  )
}