import { useState } from 'react'
import { clsx } from 'clsx'

type FilterType = 'all' | 'new' | 'used' | 'low_stock' | 'out_of_stock'
type SortOption = 'name' | 'stock' | 'price' | 'category' | 'date'

interface InventoryFiltersProps {
  onFilterChange: (filter: FilterType) => void
  onSortChange: (sort: SortOption) => void
  onSearchChange: (search: string) => void
  onCategoryChange: (category: string) => void
  categories: string[]
  className?: string
}

export function InventoryFilters({
  onFilterChange,
  onSortChange,
  onSearchChange,
  onCategoryChange,
  categories,
  className,
}: InventoryFiltersProps) {
  const [filter, setFilter] = useState<FilterType>('all')
  const [sort, setSort] = useState<SortOption>('name')
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('all')

  const handleFilterChange = (newFilter: FilterType) => {
    setFilter(newFilter)
    onFilterChange(newFilter)
  }

  const handleSortChange = (newSort: SortOption) => {
    setSort(newSort)
    onSortChange(newSort)
  }

  const handleSearchChange = (newSearch: string) => {
    setSearch(newSearch)
    onSearchChange(newSearch)
  }

  const handleCategoryChange = (newCategory: string) => {
    setCategory(newCategory)
    onCategoryChange(newCategory)
  }

  return (
    <div className={clsx('bg-surface rounded-lg p-4 space-y-4', className)}>
      {/* Search */}
      <div>
        <input
          type="text"
          placeholder="بحث بالاسم، الباركود، أو SKU..."
          value={search}
          onChange={(e) => handleSearchChange(e.target.value)}
          className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
        />
      </div>

      {/* Filter Buttons */}
      <div className="flex flex-wrap gap-2">
        <span className="text-sm text-muted self-center">الفلتر:</span>
        {([
          { value: 'all', label: 'الكل' },
          { value: 'new', label: 'جديد' },
          { value: 'used', label: 'مستعمل' },
          { value: 'low_stock', label: 'منخفض' },
          { value: 'out_of_stock', label: 'نفذت' },
        ] as const).map((option) => (
          <button
            key={option.value}
            onClick={() => handleFilterChange(option.value)}
            className={clsx(
              'px-3 py-1 rounded-lg text-sm font-medium transition-colors',
              filter === option.value
                ? 'bg-primary text-white'
                : 'bg-muted text-muted hover:bg-muted-80'
            )}
          >
            {option.label}
          </button>
        ))}
      </div>

      {/* Category Filter */}
      {categories.length > 0 && (
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted">التصنيف:</span>
          <select
            value={category}
            onChange={(e) => handleCategoryChange(e.target.value)}
            className="px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="all">الكل</option>
            {categories.map((cat) => (
              <option key={cat} value={cat}>
                {cat}
              </option>
            ))}
          </select>
        </div>
      )}

      {/* Sort */}
      <div className="flex items-center gap-2">
        <span className="text-sm text-muted">ترتيب حسب:</span>
        <select
          value={sort}
          onChange={(e) => handleSortChange(e.target.value as SortOption)}
          className="px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
        >
          <option value="name">الاسم</option>
          <option value="stock">المخزون</option>
          <option value="price">السعر</option>
          <option value="category">التصنيف</option>
          <option value="date">التاريخ</option>
        </select>
      </div>
    </div>
  )
}
