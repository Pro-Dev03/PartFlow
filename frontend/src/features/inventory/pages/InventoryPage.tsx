import { useState } from 'react'

export function InventoryPage() {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState('all')
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')

  // Mock data for demonstration
  const products = [
    { id: '1', name: 'RTX 4070 Gaming OC', barcode: 'PF-NEW-000452', category: 'كروت شاشة', condition: 'new', stock: 5, price: 2790, image: '🎮' },
    { id: '2', name: 'RTX 3060 Ti Used', barcode: 'PF-USED-000285', category: 'كروت شاشة', condition: 'used', stock: 1, price: 1200, image: '🎮' },
    { id: '3', name: 'RAM 32GB DDR4', barcode: 'PF-NEW-000321', category: 'ذاكرة', condition: 'new', stock: 12, price: 550, image: '💾' },
    { id: '4', name: 'SSD 1TB NVMe', barcode: 'PF-USED-000398', category: 'تخزين', condition: 'used', stock: 3, price: 350, image: '💾' },
    { id: '5', name: 'Intel i7-13700K', barcode: 'PF-NEW-000410', category: 'معالجات', condition: 'new', stock: 8, price: 1800, image: '⚡' },
    { id: '6', name: 'PSU 750W Gold', barcode: 'PF-NEW-000089', category: 'طاقة', condition: 'new', stock: 2, price: 400, image: '🔌' },
    { id: '7', name: 'RAM 8GB DDR4', barcode: 'PF-USED-000156', category: 'ذاكرة', condition: 'used', stock: 0, price: 120, image: '💾' },
    { id: '8', name: 'Case RGB', barcode: 'PF-NEW-000205', category: 'هيكل', condition: 'new', stock: 4, price: 250, image: '🖥️' },
  ]

  const categories = ['all', 'كروت شاشة', 'ذاكرة', 'تخزين', 'معالجات', 'طاقة', 'هيكل']

  const filteredProducts = products.filter(product => {
    if (filter === 'new' && product.condition !== 'new') return false
    if (filter === 'used' && product.condition !== 'used') return false
    if (filter === 'low' && product.stock > 3) return false
    if (filter === 'out' && product.stock > 0) return false
    if (search && !product.name.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  const stats = [
    { label: 'إجمالي المنتجات', value: products.length, color: '#2563eb' },
    { label: 'جديد', value: products.filter(p => p.condition === 'new').length, color: '#10b981' },
    { label: 'مستعمل', value: products.filter(p => p.condition === 'used').length, color: '#f59e0b' },
    { label: 'نفذت الكمية', value: products.filter(p => p.stock === 0).length, color: '#ef4444' },
  ]

  return (
    <div style={{ padding: '30px', maxWidth: '1500px', direction: 'rtl' }}>
      {/* Header */}
      <div style={{ marginBottom: '32px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '24px' }}>
          <div>
            <h1 style={{ fontSize: '32px', fontWeight: 800, margin: '0 0 8px', color: '#111827' }}>
              المنتجات
            </h1>
            <p style={{ color: '#6b7280', margin: 0, fontSize: '15px' }}>
              إدارة المخزون والمنتجات
            </p>
          </div>
          <button style={{
            background: '#2563eb',
            color: '#fff',
            border: 'none',
            padding: '12px 24px',
            borderRadius: '8px',
            fontWeight: 600,
            cursor: 'pointer',
            fontSize: '14px',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            transition: 'all 0.2s'
          }}
          onMouseOver={(e) => e.currentTarget.style.backgroundColor = '#1d4ed8'}
          onMouseOut={(e) => e.currentTarget.style.backgroundColor = '#2563eb'}>
            <span style={{ fontSize: '18px' }}>+</span>
            إضافة منتج
          </button>
        </div>

        {/* Stats */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px' }}>
          {stats.map((stat, index) => (
            <div key={index} style={{
              background: '#fff',
              border: '1px solid #e5e7eb',
              borderRadius: '12px',
              padding: '20px',
              display: 'flex',
              alignItems: 'center',
              gap: '16px',
              boxShadow: '0 1px 3px rgba(0,0,0,0.05)'
            }}>
              <div style={{
                width: '48px',
                height: '48px',
                borderRadius: '12px',
                backgroundColor: `${stat.color}15`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '24px'
              }}>
                📊
              </div>
              <div>
                <div style={{ fontSize: '13px', color: '#6b7280', marginBottom: '4px' }}>{stat.label}</div>
                <div style={{ fontSize: '24px', fontWeight: 800, color: '#111827', lineHeight: 1 }}>{stat.value}</div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Filters & Search */}
      <div style={{ 
        background: '#fff',
        border: '1px solid #e5e7eb',
        borderRadius: '12px',
        padding: '20px',
        marginBottom: '24px',
        display: 'flex',
        gap: '16px',
        alignItems: 'center',
        flexWrap: 'wrap'
      }}>
        {/* Search */}
        <div style={{ flex: 1, minWidth: '200px' }}>
          <input
            type="text"
            placeholder="بحث عن منتج..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            style={{
              width: '100%',
              padding: '10px 16px',
              border: '1px solid #e5e7eb',
              borderRadius: '8px',
              fontSize: '14px',
              direction: 'rtl',
              outline: 'none'
            }}
            onFocus={(e) => e.target.style.borderColor = '#2563eb'}
            onBlur={(e) => e.target.style.borderColor = '#e5e7eb'}
          />
        </div>

        {/* Filter Buttons */}
        <div style={{ display: 'flex', gap: '8px' }}>
          {[
            { value: 'all', label: 'الكل' },
            { value: 'new', label: 'جديد' },
            { value: 'used', label: 'مستعمل' },
            { value: 'low', label: 'منخفض' },
            { value: 'out', label: 'نفذت' },
          ].map((f) => (
            <button
              key={f.value}
              onClick={() => setFilter(f.value)}
              style={{
                padding: '8px 16px',
                borderRadius: '6px',
                border: '1px solid',
                backgroundColor: filter === f.value ? '#2563eb' : '#fff',
                borderColor: filter === f.value ? '#2563eb' : '#e5e7eb',
                color: filter === f.value ? '#fff' : '#374151',
                fontSize: '14px',
                fontWeight: 500,
                cursor: 'pointer',
                transition: 'all 0.2s'
              }}
              onMouseOver={(e) => {
                if (filter !== f.value) {
                  e.currentTarget.style.backgroundColor = '#f9fafb'
                }
              }}
              onMouseOut={(e) => {
                if (filter !== f.value) {
                  e.currentTarget.style.backgroundColor = '#fff'
                }
              }}
            >
              {f.label}
            </button>
          ))}
        </div>

        {/* View Toggle */}
        <div style={{ display: 'flex', gap: '4px', background: '#f3f4f6', padding: '4px', borderRadius: '8px' }}>
          <button
            onClick={() => setViewMode('grid')}
            style={{
              padding: '8px 12px',
              borderRadius: '6px',
              border: 'none',
              backgroundColor: viewMode === 'grid' ? '#fff' : 'transparent',
              color: '#374151',
              fontSize: '16px',
              cursor: 'pointer',
              boxShadow: viewMode === 'grid' ? '0 1px 3px rgba(0,0,0,0.1)' : 'none'
            }}
          >
            ⊞
          </button>
          <button
            onClick={() => setViewMode('list')}
            style={{
              padding: '8px 12px',
              borderRadius: '6px',
              border: 'none',
              backgroundColor: viewMode === 'list' ? '#fff' : 'transparent',
              color: '#374151',
              fontSize: '16px',
              cursor: 'pointer',
              boxShadow: viewMode === 'list' ? '0 1px 3px rgba(0,0,0,0.1)' : 'none'
            }}
          >
            ☰
          </button>
        </div>
      </div>

      {/* Products Grid */}
      {viewMode === 'grid' ? (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '20px' }}>
          {filteredProducts.map((product) => (
            <div
              key={product.id}
              style={{
                background: '#fff',
                border: '1px solid #e5e7eb',
                borderRadius: '12px',
                padding: '20px',
                transition: 'all 0.2s',
                boxShadow: '0 1px 3px rgba(0,0,0,0.05)',
                cursor: 'pointer'
              }}
              onMouseOver={(e) => {
                e.currentTarget.style.transform = 'translateY(-4px)'
                e.currentTarget.style.boxShadow = '0 8px 16px rgba(0,0,0,0.1)'
              }}
              onMouseOut={(e) => {
                e.currentTarget.style.transform = 'translateY(0)'
                e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.05)'
              }}
            >
              {/* Product Image */}
              <div style={{
                width: '100%',
                height: '140px',
                backgroundColor: '#f9fafb',
                borderRadius: '8px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '48px',
                marginBottom: '16px'
              }}>
                {product.image}
              </div>

              {/* Product Info */}
              <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#111827', margin: '0 0 8px', lineHeight: 1.4 }}>
                {product.name}
              </h3>
              
              <div style={{ fontSize: '13px', color: '#6b7280', marginBottom: '12px' }}>
                {product.barcode}
              </div>

              {/* Tags */}
              <div style={{ display: 'flex', gap: '8px', marginBottom: '12px' }}>
                <span style={{
                  padding: '4px 10px',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  backgroundColor: product.condition === 'new' ? '#dcfce7' : '#fef3c7',
                  color: product.condition === 'new' ? '#166534' : '#92400e'
                }}>
                  {product.condition === 'new' ? 'جديد' : 'مستعمل'}
                </span>
                <span style={{
                  padding: '4px 10px',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  backgroundColor: '#f3f4f6',
                  color: '#4b5563'
                }}>
                  {product.category}
                </span>
              </div>

              {/* Stock & Price */}
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', paddingTop: '12px', borderTop: '1px solid #e5e7eb' }}>
                <div>
                  <div style={{ fontSize: '12px', color: '#6b7280', marginBottom: '2px' }}>المخزون</div>
                  <div style={{ 
                    fontSize: '16px', 
                    fontWeight: 700, 
                    color: product.stock === 0 ? '#ef4444' : product.stock <= 3 ? '#f59e0b' : '#111827'
                  }}>
                    {product.stock}
                  </div>
                </div>
                <div style={{ textAlign: 'left' }}>
                  <div style={{ fontSize: '12px', color: '#6b7280', marginBottom: '2px' }}>السعر</div>
                  <div style={{ fontSize: '18px', fontWeight: 700, color: '#2563eb' }}>
                    ₪{product.price}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        /* List View */
        <div style={{ background: '#fff', border: '1px solid #e5e7eb', borderRadius: '12px', overflow: 'hidden' }}>
          {filteredProducts.map((product, index) => (
            <div
              key={product.id}
              style={{
                padding: '20px',
                borderBottom: index < filteredProducts.length - 1 ? '1px solid #e5e7eb' : 'none',
                display: 'flex',
                alignItems: 'center',
                gap: '20px',
                transition: 'background-color 0.2s'
              }}
              onMouseOver={(e) => e.currentTarget.style.backgroundColor = '#f9fafb'}
              onMouseOut={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
            >
              <div style={{
                width: '60px',
                height: '60px',
                backgroundColor: '#f9fafb',
                borderRadius: '8px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '28px'
              }}>
                {product.image}
              </div>
              
              <div style={{ flex: 1 }}>
                <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#111827', margin: '0 0 4px' }}>
                  {product.name}
                </h3>
                <div style={{ fontSize: '13px', color: '#6b7280' }}>{product.barcode}</div>
              </div>

              <div style={{ display: 'flex', gap: '8px' }}>
                <span style={{
                  padding: '4px 10px',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  backgroundColor: product.condition === 'new' ? '#dcfce7' : '#fef3c7',
                  color: product.condition === 'new' ? '#166534' : '#92400e'
                }}>
                  {product.condition === 'new' ? 'جديد' : 'مستعمل'}
                </span>
                <span style={{
                  padding: '4px 10px',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  backgroundColor: '#f3f4f6',
                  color: '#4b5563'
                }}>
                  {product.category}
                </span>
              </div>

              <div style={{ textAlign: 'center', minWidth: '80px' }}>
                <div style={{ fontSize: '12px', color: '#6b7280', marginBottom: '4px' }}>المخزون</div>
                <div style={{ 
                  fontSize: '18px', 
                  fontWeight: 700, 
                  color: product.stock === 0 ? '#ef4444' : product.stock <= 3 ? '#f59e0b' : '#111827'
                }}>
                  {product.stock}
                </div>
              </div>

              <div style={{ textAlign: 'center', minWidth: '100px' }}>
                <div style={{ fontSize: '12px', color: '#6b7280', marginBottom: '4px' }}>السعر</div>
                <div style={{ fontSize: '18px', fontWeight: 700, color: '#2563eb' }}>
                  ₪{product.price}
                </div>
              </div>

              <button style={{
                padding: '8px 16px',
                borderRadius: '6px',
                border: '1px solid #e5e7eb',
                background: '#fff',
                color: '#374151',
                fontSize: '14px',
                fontWeight: 500,
                cursor: 'pointer',
                transition: 'all 0.2s'
              }}
              onMouseOver={(e) => {
                e.currentTarget.style.backgroundColor = '#f9fafb'
                e.currentTarget.style.borderColor = '#2563eb'
              }}
              onMouseOut={(e) => {
                e.currentTarget.style.backgroundColor = '#fff'
                e.currentTarget.style.borderColor = '#e5e7eb'
              }}>
                عرض
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Empty State */}
      {filteredProducts.length === 0 && (
        <div style={{
          textAlign: 'center',
          padding: '80px 20px',
          background: '#fff',
          border: '1px solid #e5e7eb',
          borderRadius: '12px'
        }}>
          <div style={{ fontSize: '64px', marginBottom: '16px' }}>📦</div>
          <h3 style={{ fontSize: '20px', fontWeight: 600, color: '#111827', margin: '0 0 8px' }}>
            لا توجد منتجات
          </h3>
          <p style={{ color: '#6b7280', marginBottom: '24px' }}>
            لا توجد منتجات مطابقة للفلاتر الحالية
          </p>
          <button style={{
            background: '#2563eb',
            color: '#fff',
            border: 'none',
            padding: '12px 24px',
            borderRadius: '8px',
            fontWeight: 600,
            cursor: 'pointer',
            fontSize: '14px'
          }}>
            إضافة منتج
          </button>
        </div>
      )}
    </div>
  )
}

export default InventoryPage
