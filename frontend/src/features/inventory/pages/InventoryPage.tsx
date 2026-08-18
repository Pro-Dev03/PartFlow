import { useState } from 'react'
import { clsx } from 'clsx'
import { InventoryFilters } from '../components/InventoryFilters'
import { BulkActions } from '../components/BulkActions'
import { EmptyState } from '@/components/feedback'
import type { Product } from '../types/product'

type FilterType = 'all' | 'new' | 'used' | 'low_stock' | 'out_of_stock'
type SortOption = 'name' | 'stock' | 'price' | 'category' | 'date'

export function InventoryPage() {
  const [filter, setFilter] = useState<FilterType>('all')
  const [sort, setSort] = useState<SortOption>('name')
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('all')
  const [selectedProducts, setSelectedProducts] = useState<Set<string>>(new Set())

  // TODO: Fetch products from API
  const products: Product[] = []
  const categories: string[] = []

  const filteredProducts = products.filter(product => {
    // Filter by type
    if (filter === 'new' && product.condition !== 'new') return false
    if (filter === 'used' && product.condition !== 'used') return false
    if (filter === 'low_stock' && product.minStock && product.stock > product.minStock) return false
    if (filter === 'low_stock' && !product.minStock) return false
    if (filter === 'out_of_stock' && product.stock > 0) return false

    // Filter by category
    if (category !== 'all' && product.category !== category) return false

    // Filter by search
    if (search) {
      const searchLower = search.toLowerCase()
      const matchesSearch = 
        product.name.toLowerCase().includes(searchLower) ||
        product.barcode.toLowerCase().includes(searchLower) ||
        (product.sku && product.sku.toLowerCase().includes(searchLower))
      if (!matchesSearch) return false
    }

    return true
  }).sort((a, b) => {
    switch (sort) {
      case 'name':
        return a.name.localeCompare(b.name)
      case 'stock':
        return a.stock - b.stock
      case 'price':
        return a.price - b.price
      case 'category':
        return a.category.localeCompare(b.category)
      case 'date':
        return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
      default:
        return 0
    }
  })

  const handleSelectProduct = (productId: string) => {
    const newSelected = new Set(selectedProducts)
    if (newSelected.has(productId)) {
      newSelected.delete(productId)
    } else {
      newSelected.add(productId)
    }
    setSelectedProducts(newSelected)
  }

  const handleSelectAll = () => {
    if (selectedProducts.size === filteredProducts.length) {
      setSelectedProducts(new Set())
    } else {
      setSelectedProducts(new Set(filteredProducts.map(p => p.id)))
    }
  }

  const handleBulkAction = (action: string, data?: any) => {
    console.log('Bulk action:', action, data, 'on products:', Array.from(selectedProducts))
    // TODO: Execute bulk action
    setSelectedProducts(new Set())
  }

  const handleClearSelection = () => {
    setSelectedProducts(new Set())
  }

  const isLowStock = (product: Product) => 
    product.minStock && product.stock <= product.minStock && product.stock > 0

  const isOutOfStock = (product: Product) => product.stock === 0

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">المخزون</h1>
        <p className="text-muted">إدارة المنتجات والمخزون</p>
      </div>

      {/* Filters */}
      <InventoryFilters
        onFilterChange={setFilter}
        onSortChange={setSort}
        onSearchChange={setSearch}
        onCategoryChange={setCategory}
        categories={categories}
        className="mb-6"
      />

      {/* Bulk Actions */}
      {selectedProducts.size > 0 && (
        <BulkActions
          selectedCount={selectedProducts.size}
          onAction={handleBulkAction}
          onClearSelection={handleClearSelection}
          className="mb-6"
        />
      )}

      {/* Products Table */}
      {filteredProducts.length > 0 ? (
        <div className="bg-surface rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted-10 border-b border-border">
              <tr>
                <th className="px-4 py-3 w-10">
                  <input
                    type="checkbox"
                    checked={selectedProducts.size === filteredProducts.length && filteredProducts.length > 0}
                    onChange={handleSelectAll}
                    className="w-4 h-4 text-primary rounded"
                  />
                </th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المنتج</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الباركود</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">التصنيف</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المخزون</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">السعر</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {filteredProducts.map((product) => (
                <tr 
                  key={product.id} 
                  className={clsx(
                    'border-b border-border hover:bg-muted-5',
                    selectedProducts.has(product.id) && 'bg-primary-5'
                  )}
                >
                  <td className="px-4 py-3">
                    <input
                      type="checkbox"
                      checked={selectedProducts.has(product.id)}
                      onChange={() => handleSelectProduct(product.id)}
                      className="w-4 h-4 text-primary rounded"
                    />
                  </td>
                  <td className="px-4 py-3">
                    <div>
                      <p className="font-medium text-text">{product.name}</p>
                      {product.sku && (
                        <p className="text-sm text-muted">SKU: {product.sku}</p>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-muted">{product.barcode}</td>
                  <td className="px-4 py-3 text-muted">{product.category}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span
                        className={clsx(
                          'px-2 py-1 rounded text-xs font-medium',
                          product.condition === 'new' ? 'bg-success-10 text-success' : 'bg-warning-10 text-warning'
                        )}
                      >
                        {product.condition === 'new' ? 'جديد' : 'مستعمل'}
                      </span>
                      {product.grade && (
                        <span className="px-2 py-1 rounded text-xs font-medium bg-muted-10 text-muted">
                          {product.grade}
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span className={clsx(
                        'font-medium',
                        isOutOfStock(product) ? 'text-danger' : isLowStock(product) ? 'text-warning' : 'text-text'
                      )}>
                        {product.stock}
                      </span>
                      {isOutOfStock(product) && (
                        <span className="text-xs text-danger">نفذت</span>
                      )}
                      {isLowStock(product) && (
                        <span className="text-xs text-warning">منخفض</span>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3 font-medium text-text">{product.price.toFixed(2)}</td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => {/* TODO: Navigate to product details */}}
                      className="text-primary hover:text-primary-600 text-sm"
                    >
                      عرض
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <EmptyState
          icon="📦"
          title="لا توجد منتجات"
          description="لا توجد منتجات مطابقة للفلاتر الحالية"
          actionLabel="إضافة منتج"
          onAction={() => {/* TODO: Open add product modal */}}
        />
      )}
    </div>
  )
}
