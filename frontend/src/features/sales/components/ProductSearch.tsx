import { useState } from 'react'
import { Input } from '../../../components/ui/Input'
import { Card } from '../../../components/ui/Card'
import { Button } from '../../../components/ui/Button'
import { useTranslation } from 'react-i18next'
import { useSearchProducts } from '../hooks/useSales'
import { CartItem } from '../types/sales.types'

interface ProductSearchProps {
  onAddToCart: (item: CartItem) => void
}

export function ProductSearch({ onAddToCart }: ProductSearchProps) {
  const { t } = useTranslation()
  const [searchQuery, setSearchQuery] = useState('')
  const { data: products, isLoading } = useSearchProducts(searchQuery)

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('ar-SA', {
      style: 'currency',
      currency: 'ILS',
      minimumFractionDigits: 0,
    }).format(amount)
  }

  return (
    <div className="space-y-4">
      <Input
        placeholder={t('common.search') + '...'}
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        leftIcon="🔍"
        autoFocus
      />

      {searchQuery.length > 2 && (
        <Card className="max-h-96 overflow-y-auto scrollbar-thin">
          {isLoading ? (
            <div className="p-4 space-y-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="flex items-center gap-3">
                  <div className="w-12 h-12 bg-neutral-200 rounded animate-pulse" />
                  <div className="flex-1 space-y-2">
                    <div className="h-4 bg-neutral-200 rounded animate-pulse w-3/4" />
                    <div className="h-3 bg-neutral-200 rounded animate-pulse w-1/2" />
                  </div>
                </div>
              ))}
            </div>
          ) : products && products.length > 0 ? (
            <div className="divide-y divide-border">
              {products.map((product) => (
                <div key={product.id} className="p-4 hover:bg-background transition-colors">
                  <div className="flex items-start gap-3">
                    <div className="w-12 h-12 rounded-lg bg-primary-100 flex items-center justify-center flex-shrink-0">
                      <span className="text-2xl">📦</span>
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex-1">
                          <h4 className="font-medium text-text">{product.productName}</h4>
                          <p className="text-sm text-muted">
                            {product.barcode} • {product.condition === 'used' ? 'مستعمل' : 'جديد'}
                          </p>
                          {product.serialNumber && (
                            <p className="text-xs text-muted">SN: {product.serialNumber}</p>
                          )}
                        </div>
                        <div className="text-right">
                          <p className="font-bold text-primary">{formatCurrency(product.price)}</p>
                          <p className="text-xs text-muted">
                            المتوفر: {product.availableStock}
                          </p>
                        </div>
                      </div>
                      <Button
                        variant="primary"
                        size="sm"
                        className="mt-2"
                        onClick={() => onAddToCart(product)}
                        disabled={product.availableStock === 0}
                      >
                        {t('sales.addToCart')}
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="p-8 text-center text-muted">
              <span className="text-4xl">🔍</span>
              <p className="mt-2">{t('barcode.notFound')}</p>
            </div>
          )}
        </Card>
      )}
    </div>
  )
}