import { clsx } from 'clsx'
import type { Product } from '../types/product'

interface UsedItemCardProps {
  product: Product
  onViewDetails: (productId: string) => void
  onAddInspection: (productId: string) => void
  className?: string
}

export function UsedItemCard({ product, onViewDetails, onAddInspection, className }: UsedItemCardProps) {
  const gradeColors = {
    A: 'bg-success text-white',
    B: 'bg-primary text-white',
    C: 'bg-warning text-white',
    D: 'bg-danger text-white',
  }

  const gradeLabels = {
    A: 'ممتاز',
    B: 'جيد جداً',
    C: 'جيد',
    D: 'مقبول',
  }

  const hasRecentInspection = product.updatedAt && 
    new Date(product.updatedAt) > new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)

  return (
    <div className={clsx('bg-surface rounded-lg shadow-sm border border-border overflow-hidden', className)}>
      {/* Image */}
      {product.images && product.images.length > 0 ? (
        <div className="relative h-48 bg-muted-10">
          <img
            src={product.images[0]}
            alt={product.name}
            className="w-full h-full object-cover"
          />
          <div className="absolute top-2 right-2 flex gap-2">
            {product.grade && (
              <span className={clsx('px-2 py-1 rounded text-xs font-medium', gradeColors[product.grade])}>
                {gradeLabels[product.grade]}
              </span>
            )}
            {hasRecentInspection && (
              <span className="px-2 py-1 rounded text-xs font-medium bg-success text-white">
                ✓ فحص حديث
              </span>
            )}
          </div>
        </div>
      ) : (
        <div className="h-48 bg-muted-10 flex items-center justify-center text-muted text-4xl">
          📦
        </div>
      )}

      {/* Content */}
      <div className="p-4">
        <h3 className="font-semibold text-text mb-1">{product.name}</h3>
        <p className="text-sm text-muted mb-2">{product.manufacturer}</p>
        
        <div className="flex items-center justify-between mb-3">
          <span className="text-lg font-bold text-primary">{product.price.toFixed(2)}</span>
          <span className="text-sm text-muted">الكمية: {product.stock}</span>
        </div>

        {product.serialNumber && (
          <p className="text-xs text-muted mb-3">Serial: {product.serialNumber}</p>
        )}

        <div className="flex gap-2">
          <button
            onClick={() => onViewDetails(product.id)}
            className="flex-1 px-3 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors text-sm"
          >
            التفاصيل
          </button>
          <button
            onClick={() => onAddInspection(product.id)}
            className="flex-1 px-3 py-2 border border-border rounded-lg hover:bg-surface-80 transition-colors text-sm"
          >
            فحص
          </button>
        </div>
      </div>
    </div>
  )
}
