import { useState } from 'react'
import { clsx } from 'clsx'
import { ProductTimeline } from './ProductTimeline'

interface Product {
  id: string
  name: string
  barcode: string
  sku?: string
  category: string
  manufacturer: string
  model?: string
  condition: 'new' | 'used'
  grade?: 'A' | 'B' | 'C' | 'D'
  cost: number
  price: number
  stock: number
  minStock?: number
  location?: string
  serialNumber?: string
  warranty?: {
    enabled: boolean
    duration?: number
    type?: string
  }
  description?: string
  images?: string[]
  supplierId?: string
  supplierName?: string
  createdAt: string
  updatedAt: string
}

interface ProductDetailProps {
  product: Product
  onEdit?: () => void
  onDelete?: () => void
  className?: string
}

export function ProductDetail({ product, onEdit, onDelete, className }: ProductDetailProps) {
  const [activeTab, setActiveTab] = useState<'details' | 'timeline' | 'images'>('details')

  const conditionGrades = {
    A: { label: 'ممتاز', color: 'bg-success text-white' },
    B: { label: 'جيد جداً', color: 'bg-primary text-white' },
    C: { label: 'جيد', color: 'bg-warning text-white' },
    D: { label: 'مقبول', color: 'bg-danger text-white' },
  }

  const isLowStock = product.minStock && product.stock <= product.minStock
  const isOutOfStock = product.stock === 0

  return (
    <div className={clsx('bg-surface rounded-lg shadow-sm', className)}>
      {/* Header */}
      <div className="p-6 border-b border-border">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-2xl font-bold text-text">{product.name}</h1>
              <span
                className={clsx(
                  'px-2 py-1 rounded text-xs font-medium',
                  product.condition === 'new' ? 'bg-success-10 text-success' : 'bg-warning-10 text-warning'
                )}
              >
                {product.condition === 'new' ? 'جديد' : 'مستعمل'}
              </span>
              {product.grade && (
                <span className={clsx('px-2 py-1 rounded text-xs font-medium', conditionGrades[product.grade].color)}>
                  {conditionGrades[product.grade].label}
                </span>
              )}
              {isOutOfStock && (
                <span className="px-2 py-1 rounded text-xs font-medium bg-danger text-white">
                  نفذت الكمية
                </span>
              )}
              {isLowStock && !isOutOfStock && (
                <span className="px-2 py-1 rounded text-xs font-medium bg-warning text-white">
                  منخفض
                </span>
              )}
            </div>
            <div className="flex items-center gap-4 text-sm text-muted">
              <span>Barcode: {product.barcode}</span>
              {product.sku && <span>SKU: {product.sku}</span>}
              {product.serialNumber && <span>Serial: {product.serialNumber}</span>}
            </div>
          </div>
          <div className="flex gap-2">
            {onEdit && (
              <button
                onClick={onEdit}
                className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
              >
                تعديل
              </button>
            )}
            {onDelete && (
              <button
                onClick={onDelete}
                className="px-4 py-2 bg-danger text-white rounded-lg hover:bg-danger-600 transition-colors"
              >
                حذف
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <nav className="flex gap-4 px-6">
          {(['details', 'timeline', 'images'] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={clsx(
                'px-4 py-3 text-sm font-medium border-b-2 transition-colors',
                activeTab === tab
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted hover:text-text'
              )}
            >
              {tab === 'details' && 'التفاصيل'}
              {tab === 'timeline' && 'السجل'}
              {tab === 'images' && 'الصور'}
            </button>
          ))}
        </nav>
      </div>

      {/* Content */}
      <div className="p-6">
        {activeTab === 'details' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Basic Info */}
            <div className="space-y-4">
              <h3 className="font-semibold text-text">المعلومات الأساسية</h3>
              <div className="space-y-2">
                <InfoRow label="التصنيف" value={product.category} />
                <InfoRow label="الشركة المصنعة" value={product.manufacturer} />
                {product.model && <InfoRow label="الموديل" value={product.model} />}
                {product.location && <InfoRow label="الموقع" value={product.location} />}
                {product.supplierName && <InfoRow label="المورد" value={product.supplierName} />}
              </div>
            </div>

            {/* Pricing */}
            <div className="space-y-4">
              <h3 className="font-semibold text-text">التسعير</h3>
              <div className="space-y-2">
                <InfoRow label="سعر التكلفة" value={`${product.cost.toFixed(2)}`} />
                <InfoRow label="سعر البيع" value={`${product.price.toFixed(2)}`} />
                <InfoRow
                  label="هامش الربح"
                  value={`${(((product.price - product.cost) / product.cost) * 100).toFixed(1)}%`}
                />
              </div>
            </div>

            {/* Stock */}
            <div className="space-y-4">
              <h3 className="font-semibold text-text">المخزون</h3>
              <div className="space-y-2">
                <InfoRow label="الكمية الحالية" value={product.stock.toString()} />
                {product.minStock && <InfoRow label="الحد الأدنى" value={product.minStock.toString()} />}
                {isLowStock && (
                  <div className="text-sm text-warning">
                    ⚠️ الكمية منخفضة، يرجى التزويد
                  </div>
                )}
              </div>
            </div>

            {/* Warranty */}
            {product.warranty?.enabled && (
              <div className="space-y-4">
                <h3 className="font-semibold text-text">الضمان</h3>
                <div className="space-y-2">
                  {product.warranty.duration && (
                    <InfoRow label="المدة" value={`${product.warranty.duration} شهر`}
                    />
                  )}
                  {product.warranty.type && <InfoRow label="النوع" value={product.warranty.type} />}
                </div>
              </div>
            )}

            {/* Description */}
            {product.description && (
              <div className="space-y-4 md:col-span-2">
                <h3 className="font-semibold text-text">الوصف</h3>
                <p className="text-muted">{product.description}</p>
              </div>
            )}

            {/* Dates */}
            <div className="space-y-4 md:col-span-2">
              <h3 className="font-semibold text-text">التواريخ</h3>
              <div className="space-y-2">
                <InfoRow label="تاريخ الإنشاء" value={product.createdAt} />
                <InfoRow label="آخر تحديث" value={product.updatedAt} />
              </div>
            </div>
          </div>
        )}

        {activeTab === 'timeline' && (
          <ProductTimeline
            events={[
              {
                id: '1',
                type: 'purchase',
                date: product.createdAt,
                description: 'إضافة المنتج',
                user: 'النظام',
              },
            ]}
          />
        )}

        {activeTab === 'images' && (
          <div>
            {product.images && product.images.length > 0 ? (
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {product.images.map((image, index) => (
                  <img
                    key={index}
                    src={image}
                    alt={`${product.name} - ${index + 1}`}
                    className="w-full h-32 object-cover rounded-lg border border-border"
                  />
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-muted">
                لا توجد صور للمنتج
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between py-2 border-b border-border">
      <span className="text-muted">{label}</span>
      <span className="font-medium text-text">{value}</span>
    </div>
  )
}
