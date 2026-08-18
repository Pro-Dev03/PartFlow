import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { ProductDetail } from '../components/ProductDetail'
import { InspectionForm } from '../components/InspectionForm'
import { EmptyState } from '@/components/feedback'
import type { Product } from '../types/product'

export function ProductPage() {
  const { id } = useParams<{ id: string }>()
  const [product, setProduct] = useState<Product | null>(null)
  const [showInspection, setShowInspection] = useState(false)
  const [loading, setLoading] = useState(true)

  // TODO: Fetch product data from API
  // useEffect(() => {
  //   fetchProduct(id).then(setProduct).finally(() => setLoading(false))
  // }, [id])

  if (loading) {
    return <div className="p-8">جاري التحميل...</div>
  }

  if (!product) {
    return (
      <EmptyState
        icon="📦"
        title="المنتج غير موجود"
        description="لم يتم العثور على المنتج المطلوب"
        actionLabel="العودة للمخزون"
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

      <ProductDetail
        product={product}
        onEdit={() => {/* TODO: Open edit modal */}}
        onDelete={() => {/* TODO: Show delete confirmation */}}
      />

      {product.condition === 'used' && (
        <div className="mt-6">
          {!showInspection ? (
            <button
              onClick={() => setShowInspection(true)}
              className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              إضافة فحص جديد
            </button>
          ) : (
            <div className="bg-surface rounded-lg p-6">
              <h2 className="text-xl font-bold text-text mb-4">فحص المنتج</h2>
              <InspectionForm
                productId={product.id}
                onSubmit={(data) => {
                  // TODO: Submit inspection
                  console.log('Inspection submitted:', data)
                  setShowInspection(false)
                }}
                onCancel={() => setShowInspection(false)}
              />
            </div>
          )}
        </div>
      )}
    </div>
  )
}
