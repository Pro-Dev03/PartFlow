import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { clsx } from 'clsx'
import { ProductDetail } from './components/ProductDetail'
import { InventoryFilters } from './components/InventoryFilters'
import { BulkActions } from './components/BulkActions'
import { QuickPreviewDrawer } from '../../components/tables/QuickPreviewDrawer'
import { EmptyState } from '../../components/feedback/EmptyState'
import { Button } from '@components/ui/button'
import { Card, CardHeader } from '@components/ui/card'
import type { Product } from './types/product'

type FilterType = 'all' | 'new' | 'used' | 'low_stock' | 'out_of_stock'
type SortOption = 'name' | 'stock' | 'price' | 'category' | 'date'
type ViewMode = 'table' | 'grid'

export function Inventory() {
  const { t } = useTranslation()
  const [filter, setFilter] = useState<FilterType>('all')
  const [sort, setSort] = useState<SortOption>('name')
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('all')
  const [selectedProducts, setSelectedProducts] = useState<Set<string>>(new Set())
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null)
  const [showMobileView, setShowMobileView] = useState(false)
  const [viewMode, setViewMode] = useState<ViewMode>('table') // Table view as default per report

  // Mock data - TODO: Replace with API calls
  const products: Product[] = [
    {
      id: '1',
      name: 'RTX 4070',
      barcode: '1234567890123',
      sku: 'GPU-RTX4070',
      category: 'GPUs',
      manufacturer: 'ASUS',
      model: 'RTX 4070 OC',
      condition: 'used',
      grade: 'A',
      cost: 1850,
      price: 2350,
      stock: 1,
      minStock: 2,
      location: 'B-03',
      serialNumber: 'SN123456789',
      warranty: { enabled: true, duration: 30, type: 'seller' },
      description: 'بطاقة رسوميات مستعملة بحالة ممتازة',
      supplierName: 'Tech Supplier',
      createdAt: '2024-01-15',
      updatedAt: '2024-08-18',
    },
    {
      id: '2',
      name: 'RTX 3060',
      barcode: '1234567890124',
      sku: 'GPU-RTX3060',
      category: 'GPUs',
      manufacturer: 'MSI',
      model: 'RTX 3060 Ventus 2X',
      condition: 'used',
      grade: 'B',
      cost: 900,
      price: 1250,
      stock: 2,
      minStock: 3,
      location: 'B-04',
      serialNumber: 'SN987654321',
      warranty: { enabled: true, duration: 30, type: 'seller' },
      createdAt: '2024-02-10',
      updatedAt: '2024-08-15',
    },
    {
      id: '3',
      name: 'RAM 16GB DDR4',
      barcode: '1234567890125',
      sku: 'RAM-16GB-DDR4',
      category: 'Memory',
      manufacturer: 'Corsair',
      model: 'Vengeance LPX',
      condition: 'new',
      cost: 180,
      price: 250,
      stock: 8,
      minStock: 5,
      location: 'A-01',
      createdAt: '2024-03-20',
      updatedAt: '2024-08-10',
    },
  ]

  const categories = ['GPUs', 'Memory', 'Storage', 'Motherboards', 'PSU', 'Cases']

  const filteredProducts = products.filter((product) => {
    // Search filter
    if (search && !product.name.toLowerCase().includes(search.toLowerCase()) && 
        !product.barcode.includes(search) && 
        !(product.sku && product.sku.toLowerCase().includes(search.toLowerCase()))) {
      return false
    }

    // Category filter
    if (category !== 'all' && product.category !== category) {
      return false
    }

    // Status filter
    switch (filter) {
      case 'new':
        return product.condition === 'new'
      case 'used':
        return product.condition === 'used'
      case 'low_stock':
        return product.minStock && product.stock <= product.minStock && product.stock > 0
      case 'out_of_stock':
        return product.stock === 0
      default:
        return true
    }
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
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      default:
        return 0
    }
  })

  const handleSelectProduct = (productId: string) => {
    const newSelection = new Set(selectedProducts)
    if (newSelection.has(productId)) {
      newSelection.delete(productId)
    } else {
      newSelection.add(productId)
    }
    setSelectedProducts(newSelection)
  }

  const handleSelectAll = () => {
    if (selectedProducts.size === filteredProducts.length) {
      setSelectedProducts(new Set())
    } else {
      setSelectedProducts(new Set(filteredProducts.map(p => p.id)))
    }
  }

  const handleBulkAction = (action: string, data?: any) => {
    console.log('Bulk action:', action, data)
    // TODO: Implement bulk actions
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('ar-SA', {
      style: 'currency',
      currency: 'ILS',
      minimumFractionDigits: 0,
    }).format(amount)
  }

  const isLowStock = (product: Product) => product.minStock && product.stock <= product.minStock && product.stock > 0
  const isOutOfStock = (product: Product) => product.stock === 0

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text">{t('inventory.title')}</h1>
          <p className="text-muted mt-1">{filteredProducts.length} منتج</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setShowMobileView(!showMobileView)}>
            {showMobileView ? '📊' : '📱'}
          </Button>
          <Button variant="primary">
            + {t('inventory.addProduct')}
          </Button>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="إجمالي المنتجات" value={products.length} icon="📦" color="primary" />
        <StatCard label="منخفض المخزون" value={products.filter(p => isLowStock(p)).length} icon="⚠️" color="warning" />
        <StatCard label="نفذت الكمية" value={products.filter(p => isOutOfStock(p)).length} icon="🔴" color="danger" />
        <StatCard label="قيمة المخزون" value={formatCurrency(products.reduce((sum, p) => sum + (p.cost * p.stock), 0))} icon="💰" color="success" />
      </div>

      {/* Filters */}
      <div className="flex items-center justify-between">
        <InventoryFilters
          onFilterChange={setFilter}
          onSortChange={setSort}
        onSearchChange={setSearch}
        onCategoryChange={setCategory}
        categories={categories}
      />
        
        {/* View Toggle - Grid / Table */}
        <div className="flex items-center gap-2 bg-muted-10 rounded-lg p-1">
          <button
            onClick={() => setViewMode('table')}
            className={clsx(
              'px-3 py-1.5 rounded-md text-sm font-medium transition-colors',
              viewMode === 'table' ? 'bg-surface text-text shadow-sm' : 'text-muted hover:text-text'
            )}
          >
            Table
          </button>
          <button
            onClick={() => setViewMode('grid')}
            className={clsx(
              'px-3 py-1.5 rounded-md text-sm font-medium transition-colors',
              viewMode === 'grid' ? 'bg-surface text-text shadow-sm' : 'text-muted hover:text-text'
            )}
          >
            Grid
          </button>
        </div>
      </div>

      {/* Bulk Actions */}
      {selectedProducts.size > 0 && (
        <BulkActions
          selectedCount={selectedProducts.size}
          onAction={handleBulkAction}
          onClearSelection={() => setSelectedProducts(new Set())}
        />
      )}

      {/* Products Table/Desktop View */}
      {!showMobileView && (
        <Card>
          <CardHeader title={t('inventory.products')} />
          {filteredProducts.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-muted-10 border-b border-border">
                  <tr>
                    <th className="px-4 py-3 text-right">
                      <input
                        type="checkbox"
                        checked={selectedProducts.size === filteredProducts.length}
                        onChange={handleSelectAll}
                        className="w-4 h-4 text-primary rounded"
                      />
                    </th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">المنتج</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">المخزون</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">التكلفة</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">السعر</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">الربح</th>
                    <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredProducts.map((product) => (
                    <tr 
                      key={product.id} 
                      className={clsx(
                        'border-b border-border hover:bg-muted-5 cursor-pointer',
                        selectedProducts.has(product.id) && 'bg-primary-5'
                      )}
                      onClick={() => setSelectedProduct(product)}
                    >
                      <td className="px-4 py-3" onClick={(e) => { e.stopPropagation(); handleSelectProduct(product.id) }}>
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
                          <p className="text-sm text-muted">{product.barcode}</p>
                          {product.serialNumber && (
                            <p className="text-xs text-muted">SN: {product.serialNumber}</p>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <span className={clsx(
                            'px-2 py-1 rounded text-xs font-medium',
                            product.condition === 'new' ? 'bg-success-10 text-success' : 'bg-warning-10 text-warning'
                          )}>
                            {product.condition === 'new' ? 'جديد' : 'مستعمل'}
                          </span>
                          {product.grade && (
                            <span className="px-2 py-1 rounded text-xs font-medium bg-info-10 text-info">
                              {product.grade}
                            </span>
                          )}
                          {isOutOfStock(product) && (
                            <span className="px-2 py-1 rounded text-xs font-medium bg-danger text-white">
                              نفذت
                            </span>
                          )}
                          {isLowStock(product) && !isOutOfStock(product) && (
                            <span className="px-2 py-1 rounded text-xs font-medium bg-warning text-white">
                              منخفض
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <span className={clsx(
                          'font-medium',
                          isOutOfStock(product) ? 'text-danger' : isLowStock(product) ? 'text-warning' : 'text-text'
                        )}>
                          {product.stock}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-muted">{formatCurrency(product.cost)}</td>
                      <td className="px-4 py-3 font-medium text-text">{formatCurrency(product.price)}</td>
                      <td className="px-4 py-3">
                        <span className={clsx(
                          'text-sm font-medium',
                          ((product.price - product.cost) / product.cost) > 0.2 ? 'text-success' : 'text-text'
                        )}>
                          {formatCurrency(product.price - product.cost)}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex gap-2">
                          <button
                            onClick={(e) => { e.stopPropagation(); setSelectedProduct(product) }}
                            className="text-primary hover:text-primary-600 text-sm"
                          >
                            عرض
                          </button>
                          <button className="text-muted hover:text-text text-sm">
                            تعديل
                          </button>
                        </div>
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
              onAction={() => {}}
            />
          )}
        </Card>
      )}

      {/* Mobile Cards View */}
      {showMobileView && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredProducts.map((product) => (
            <Card 
              key={product.id}
              className="cursor-pointer hover:shadow-md transition-shadow"
              onClick={() => setSelectedProduct(product)}
            >
              <div className="p-4">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1">
                    <h3 className="font-semibold text-text">{product.name}</h3>
                    <p className="text-sm text-muted">{product.barcode}</p>
                  </div>
                  <div className="flex gap-1">
                    <span className={clsx(
                      'px-2 py-1 rounded text-xs font-medium',
                      product.condition === 'new' ? 'bg-success-10 text-success' : 'bg-warning-10 text-warning'
                    )}>
                      {product.condition === 'new' ? 'جديد' : 'مستعمل'}
                    </span>
                    {isOutOfStock(product) && (
                      <span className="px-2 py-1 rounded text-xs font-medium bg-danger text-white">
                        نفذت
                      </span>
                    )}
                  </div>
                </div>
                
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted">المخزون:</span>
                    <span className={clsx(
                      'font-medium',
                      isOutOfStock(product) ? 'text-danger' : isLowStock(product) ? 'text-warning' : 'text-text'
                    )}>
                      {product.stock}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted">السعر:</span>
                    <span className="font-medium text-text">{formatCurrency(product.price)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted">الربح:</span>
                    <span className="font-medium text-success">{formatCurrency(product.price - product.cost)}</span>
                  </div>
                </div>

                <div className="flex gap-2 mt-4">
                  <Button 
                    variant="primary" 
                    size="sm" 
                    className="flex-1"
                    onClick={(e) => { e.stopPropagation(); }}
                  >
                    بيع
                  </Button>
                  <Button 
                    variant="outline" 
                    size="sm"
                    onClick={(e) => { e.stopPropagation(); setSelectedProduct(product) }}
                  >
                    التفاصيل
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Product Detail Drawer */}
      {selectedProduct && (
        <QuickPreviewDrawer
          isOpen={!!selectedProduct}
          onClose={() => setSelectedProduct(null)}
          title={selectedProduct.name}
        >
          <ProductDetail
            product={selectedProduct}
            onEdit={() => {}}
            onDelete={() => {}}
          />
        </QuickPreviewDrawer>
      )}
    </div>
  )
}

function StatCard({ label, value, icon, color }: { label: string; value: string | number; icon: string; color: 'primary' | 'success' | 'warning' | 'danger' }) {
  const colorClasses = {
    primary: 'text-primary',
    success: 'text-success',
    warning: 'text-warning',
    danger: 'text-danger',
  }

  return (
    <div className="bg-surface rounded-lg p-4 border border-border">
      <div className="flex items-center gap-2 mb-2">
        <span className="text-2xl">{icon}</span>
        <span className="text-sm text-muted">{label}</span>
      </div>
      <p className={clsx('text-xl font-bold', colorClasses[color])}>{value}</p>
    </div>
  )
}

export default Inventory
