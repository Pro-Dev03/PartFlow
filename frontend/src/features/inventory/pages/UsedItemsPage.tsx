import { useState } from 'react'
import { clsx } from 'clsx'
import { UsedItemCard } from '../components/UsedItemCard'
import { EmptyState } from '@/components/feedback'

type GradeFilter = 'all' | 'A' | 'B' | 'C' | 'D'
type SortOption = 'name' | 'price' | 'grade' | 'date'

export function UsedItemsPage() {
  const [gradeFilter, setGradeFilter] = useState<GradeFilter>('all')
  const [sortBy, setSortBy] = useState<SortOption>('name')
  const [searchQuery, setSearchQuery] = useState('')

  // TODO: Fetch used items from API
  const usedItems: any[] = []

  const filteredItems = usedItems.filter(item => {
    if (gradeFilter !== 'all' && item.grade !== gradeFilter) return false
    if (searchQuery && !item.name.toLowerCase().includes(searchQuery.toLowerCase())) return false
    return true
  }).sort((a, b) => {
    switch (sortBy) {
      case 'name':
        return a.name.localeCompare(b.name)
      case 'price':
        return a.price - b.price
      case 'grade':
        const gradeOrder = { A: 0, B: 1, C: 2, D: 3 }
        return (gradeOrder[a.grade as keyof typeof gradeOrder] || 99) - 
               (gradeOrder[b.grade as keyof typeof gradeOrder] || 99)
      case 'date':
        return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
      default:
        return 0
    }
  })

  const handleViewDetails = (productId: string) => {
    // TODO: Navigate to product details
    console.log('View details:', productId)
  }

  const handleAddInspection = (productId: string) => {
    // TODO: Open inspection form
    console.log('Add inspection:', productId)
  }

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">المنتجات المستعملة</h1>
        <p className="text-muted">إدارة وفحص المنتجات المستعملة</p>
      </div>

      {/* Filters */}
      <div className="bg-surface rounded-lg p-4 mb-6 space-y-4">
        {/* Search */}
        <div>
          <input
            type="text"
            placeholder="بحث عن منتج..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
        </div>

        {/* Grade Filter */}
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted self-center">التقييم:</span>
          {(['all', 'A', 'B', 'C', 'D'] as GradeFilter[]).map((grade) => (
            <button
              key={grade}
              onClick={() => setGradeFilter(grade)}
              className={clsx(
                'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
                gradeFilter === grade
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted hover:bg-muted-80'
              )}
            >
              {grade === 'all' ? 'الكل' : grade}
            </button>
          ))}
        </div>

        {/* Sort */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted">ترتيب حسب:</span>
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as SortOption)}
            className="px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="name">الاسم</option>
            <option value="price">السعر</option>
            <option value="grade">التقييم</option>
            <option value="date">التاريخ</option>
          </select>
        </div>
      </div>

      {/* Items Grid */}
      {filteredItems.length > 0 ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filteredItems.map((item) => (
            <UsedItemCard
              key={item.id}
              product={item}
              onViewDetails={handleViewDetails}
              onAddInspection={handleAddInspection}
            />
          ))}
        </div>
      ) : (
        <EmptyState
          icon="📦"
          title="لا توجد منتجات مستعملة"
          description="ابدأ بإضافة منتجات مستعملة للمخزون"
          actionLabel="إضافة منتج"
          onAction={() => {/* TODO: Open add product modal */}}
        />
      )}
    </div>
  )
}

export default UsedItemsPage
