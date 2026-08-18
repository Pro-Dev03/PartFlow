import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardHeader } from '@components/ui/card'
import { Button } from '@components/ui/button'
import { ProductSearch } from './components/ProductSearch'
import { CartSummary } from './components/CartSummary'
import { CheckoutModal } from './components/CheckoutModal'
import { useCart } from './hooks/useSales'
import { CartItem } from './types/sales.types'
import { clsx } from 'clsx'

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
  const [invoiceNumber, setInvoiceNumber] = useState('10482')

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
    // Increment invoice number
    setInvoiceNumber(prev => (parseInt(prev) + 1).toString())
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
      {/* Header - Professional POS */}
      <div className="flex items-center justify-between bg-surface border border-border rounded-xl p-4">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <span className="text-2xl">🛒</span>
            <div>
              <h1 className="text-xl font-bold text-primary">PartFlow POS</h1>
              <p className="text-xs text-muted">Invoice #{invoiceNumber}</p>
            </div>
          </div>
        </div>
        <Button onClick={() => clearCart()} disabled={cart.length === 0} variant="outline" size="sm">
          {t('common.reset')}
        </Button>
      </div>

      {/* Professional POS Layout - Split Screen */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - Product Selection */}
        <div className="lg:col-span-2 space-y-4">
          {/* Search Section */}
          <Card>
            <div className="p-4">
              <div className="relative">
                <input
                  type="text"
                  placeholder="🔍 Scan / Search products..."
                  className="w-full px-4 py-3 border border-border rounded-xl bg-surface-variant focus:ring-2 focus:ring-primary focus:border-transparent text-text"
                />
                <button className="absolute right-3 top-1/2 -translate-y-1/2 p-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors">
                  📷
                </button>
              </div>
              
              {/* Quick Categories */}
              <div className="flex gap-2 mt-3">
                {['GPUs', 'RAM', 'Storage', 'PSU', 'All'].map((category) => (
                  <button
                    key={category}
                    className="px-3 py-1.5 rounded-lg border border-border text-sm text-text-secondary hover:bg-surface-variant hover:text-text transition-colors"
                  >
                    {category}
                  </button>
                ))}
              </div>
            </div>
          </Card>

          {/* Cart Items - Professional Table */}
          {cart.length > 0 && (
            <Card>
              <CardHeader 
                title="CURRENT SALE" 
                description={`${itemCount} items`}
              />
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-muted-10 border-b border-border">
                    <tr>
                      <th className="px-4 py-3 text-right text-sm font-medium text-muted">Product</th>
                      <th className="px-4 py-3 text-right text-sm font-medium text-muted">Price</th>
                      <th className="px-4 py-3 text-right text-sm font-medium text-muted">Qty</th>
                      <th className="px-4 py-3 text-right text-sm font-medium text-muted">Total</th>
                      <th className="px-4 py-3 text-right text-sm font-medium text-muted">Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    {cart.map((item) => (
                      <tr key={item.id} className="border-b border-border hover:bg-muted-5">
                        <td className="px-4 py-3">
                          <div>
                            <p className="font-medium text-text">{item.productName}</p>
                            <p className="text-xs text-muted">
                              {item.barcode} • {item.condition === 'used' ? 'مستعمل' : 'جديد'}
                            </p>
                            {item.serialNumber && (
                              <p className="text-xs text-primary font-mono">SN: {item.serialNumber}</p>
                            )}
                          </div>
                        </td>
                        <td className="px-4 py-3 font-medium text-text">{formatCurrency(item.price)}</td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => handleUpdateQuantity(item.id, item.quantity - 1)}
                              className="w-8 h-8 rounded-lg bg-surface border border-border flex items-center justify-center hover:bg-surface-variant transition-colors"
                              disabled={item.quantity <= 1}
                            >
                              -
                            </button>
                            <span className="w-8 text-center font-medium">{item.quantity}</span>
                            <button
                              onClick={() => handleUpdateQuantity(item.id, item.quantity + 1)}
                              className="w-8 h-8 rounded-lg bg-surface border border-border flex items-center justify-center hover:bg-surface-variant transition-colors"
                              disabled={item.quantity >= item.availableStock}
                            >
                              +
                            </button>
                          </div>
                        </td>
                        <td className="px-4 py-3 font-bold text-primary">{formatCurrency(item.price * item.quantity)}</td>
                        <td className="px-4 py-3">
                          <Button
                            variant="danger"
                            size="sm"
                            onClick={() => handleRemoveFromCart(item.id)}
                          >
                            {t('common.delete')}
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </Card>
          )}
        </div>

        {/* Right Column - Cart Summary & Actions */}
        <div className="lg:col-span-1">
          <Card className="sticky top-4">
            <CardHeader title="CART SUMMARY" />
            <div className="p-4 space-y-4">
              {/* Cart Items Preview */}
              {cart.length > 0 && (
                <div className="space-y-2 mb-4 max-h-48 overflow-y-auto">
                  {cart.map((item) => (
                    <div key={item.id} className="flex justify-between items-center p-2 bg-muted-10 rounded-lg">
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-text truncate">{item.productName}</p>
                        <p className="text-xs text-muted">×{item.quantity}</p>
                      </div>
                      <p className="text-sm font-bold text-primary">{formatCurrency(item.price * item.quantity)}</p>
                    </div>
                  ))}
                </div>
              )}

              {/* Totals */}
              <div className="border-t border-border pt-4 space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted">Subtotal</span>
                  <span className="font-medium text-text">{formatCurrency(cartTotal)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted">Discount</span>
                  <span className="font-medium text-success">{formatCurrency(0)}</span>
                </div>
                <div className="flex justify-between text-lg font-bold border-t border-border pt-2">
                  <span className="text-text">TOTAL</span>
                  <span className="text-primary">{formatCurrency(cartTotal)}</span>
                </div>
              </div>

              {/* Payment Methods */}
              <div className="pt-4 border-t border-border">
                <p className="text-sm font-medium text-text mb-3">Payment Method</p>
                <div className="grid grid-cols-3 gap-2">
                  <Button
                    variant="outline"
                    className="flex flex-col items-center gap-1 h-16"
                    onClick={() => {}}
                  >
                    <span className="text-xl">💵</span>
                    <span className="text-xs">Cash</span>
                  </Button>
                  <Button
                    variant="outline"
                    className="flex flex-col items-center gap-1 h-16"
                    onClick={() => {}}
                  >
                    <span className="text-xl">💳</span>
                    <span className="text-xs">Card</span>
                  </Button>
                  <Button
                    variant="outline"
                    className="flex flex-col items-center gap-1 h-16"
                    onClick={() => {}}
                  >
                    <span className="text-xl">📱</span>
                    <span className="text-xs">Transfer</span>
                  </Button>
                </div>
              </div>

              {/* Complete Sale Button */}
              <Button
                variant="primary"
                className="w-full h-14 text-lg font-bold"
                onClick={handleCheckout}
                disabled={cart.length === 0 || isProcessing}
                isLoading={isProcessing}
              >
                {isProcessing ? 'Processing...' : 'COMPLETE SALE'}
              </Button>

              {/* Customer Selection */}
              <div className="pt-2">
                <select className="w-full px-3 py-2 border border-border rounded-lg bg-surface-variant focus:ring-2 focus:ring-primary focus:border-transparent text-text">
                  <option value="">Customer: Cash Customer</option>
                  <option value="1">أحمد محمد</option>
                  <option value="2">سارة علي</option>
                  <option value="3">خالد عبدالله</option>
                </select>
              </div>
            </div>
          </Card>
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

export default Sales