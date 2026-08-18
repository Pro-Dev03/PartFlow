import { useState, useCallback, useEffect } from 'react'
import { clsx } from 'clsx'
import { Search, X, Filter, ChevronDown } from 'lucide-react'

export interface SearchFilter {
  field: string
  operator: 'contains' | 'equals' | 'startsWith' | 'endsWith' | 'greaterThan' | 'lessThan'
  value: string
}

export interface SearchResult {
  id: string
  type: 'product' | 'customer' | 'inventory' | 'sale' | 'supplier'
  title: string
  subtitle: string
  icon: string
  data: any
}

interface AdvancedSearchProps {
  onSearch: (query: string, filters: SearchFilter[]) => Promise<SearchResult[]>
  placeholder?: string
  filters?: Array<{
    field: string
    label: string
    type: 'text' | 'number' | 'date' | 'select'
    options?: Array<{ value: string; label: string }>
  }>
  className?: string
}

export function AdvancedSearch({
  onSearch,
  placeholder = 'ابحث عن منتج، عميل، فاتورة...',
  filters = [],
  className
}: AdvancedSearchProps) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [showFilters, setShowFilters] = useState(false)
  const [activeFilters, setActiveFilters] = useState<SearchFilter[]>([])
  const [selectedIndex, setSelectedIndex] = useState(-1)

  const debouncedSearch = useCallback(
    debounce(async (searchQuery: string, searchFilters: SearchFilter[]) => {
      if (!searchQuery && searchFilters.length === 0) {
        setResults([])
        return
      }

      setIsLoading(true)
      try {
        const searchResults = await onSearch(searchQuery, searchFilters)
        setResults(searchResults)
        setSelectedIndex(-1)
      } catch (error) {
        console.error('Search error:', error)
        setResults([])
      } finally {
        setIsLoading(false)
      }
    }, 300),
    [onSearch]
  )

  useEffect(() => {
    debouncedSearch(query, activeFilters)
  }, [query, activeFilters, debouncedSearch])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex(prev => 
        prev < results.length - 1 ? prev + 1 : prev
      )
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelectedIndex(prev => prev > 0 ? prev - 1 : -1)
    } else if (e.key === 'Enter' && selectedIndex >= 0) {
      e.preventDefault()
      handleResultClick(results[selectedIndex])
    } else if (e.key === 'Escape') {
      setQuery('')
      setResults([])
      setSelectedIndex(-1)
    }
  }

  const handleResultClick = (result: SearchResult) => {
    // Handle navigation based on result type
    console.log('Selected result:', result)
    setQuery('')
    setResults([])
    setSelectedIndex(-1)
  }

  const addFilter = () => {
    if (filters.length > 0) {
      setActiveFilters([
        ...activeFilters,
        { field: filters[0].field, operator: 'contains', value: '' }
      ])
    }
  }

  const updateFilter = (index: number, updates: Partial<SearchFilter>) => {
    setActiveFilters(prev =>
      prev.map((filter, i) => (i === index ? { ...filter, ...updates } : filter))
    )
  }

  const removeFilter = (index: number) => {
    setActiveFilters(prev => prev.filter((_, i) => i !== index))
  }

  const clearAll = () => {
    setQuery('')
    setActiveFilters([])
    setResults([])
    setSelectedIndex(-1)
  }

  return (
    <div className={clsx('relative', className)}>
      {/* Search Input */}
      <div className="relative">
        <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted" />
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          className="w-full pr-10 pl-10 py-2.5 bg-surface border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
        />
        <div className="absolute left-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
          {filters.length > 0 && (
            <button
              onClick={() => setShowFilters(!showFilters)}
              className={clsx(
                'p-1.5 rounded-md transition-colors',
                showFilters ? 'bg-primary-10 text-primary' : 'hover:bg-muted-10 text-muted'
              )}
              title="إضافة فلاتر"
            >
              <Filter className="w-4 h-4" />
            </button>
          )}
          {(query || activeFilters.length > 0) && (
            <button
              onClick={clearAll}
              className="p-1.5 rounded-md hover:bg-muted-10 text-muted"
              title="مسح الكل"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>
      </div>

      {/* Filters Panel */}
      {showFilters && (
        <div className="mt-2 p-3 bg-surface border border-border rounded-lg">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium">فلاتر البحث</span>
            <button
              onClick={addFilter}
              className="text-sm text-primary hover:underline"
            >
              + إضافة فلتر
            </button>
          </div>
          
          {activeFilters.map((filter, index) => (
            <div key={index} className="flex items-center gap-2 mb-2">
              <select
                value={filter.field}
                onChange={(e) => updateFilter(index, { field: e.target.value })}
                className="px-2 py-1.5 bg-muted-10 border border-border rounded text-sm"
              >
                {filters.map(f => (
                  <option key={f.field} value={f.field}>{f.label}</option>
                ))}
              </select>
              
              <select
                value={filter.operator}
                onChange={(e) => updateFilter(index, { operator: e.target.value as any })}
                className="px-2 py-1.5 bg-muted-10 border border-border rounded text-sm"
              >
                <option value="contains">يحتوي</option>
                <option value="equals">يساوي</option>
                <option value="startsWith">يبدأ بـ</option>
                <option value="endsWith">ينتهي بـ</option>
                <option value="greaterThan">أكبر من</option>
                <option value="lessThan">أقل من</option>
              </select>
              
              <input
                type="text"
                value={filter.value}
                onChange={(e) => updateFilter(index, { value: e.target.value })}
                className="flex-1 px-2 py-1.5 bg-muted-10 border border-border rounded text-sm"
                placeholder="القيمة"
              />
              
              <button
                onClick={() => removeFilter(index)}
                className="p-1.5 hover:bg-muted-10 rounded text-muted"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          ))}
          
          {activeFilters.length === 0 && (
            <p className="text-sm text-muted">لا توجد فلاتر مفعلة</p>
          )}
        </div>
      )}

      {/* Search Results */}
      {results.length > 0 && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-surface border border-border rounded-lg shadow-lg max-h-96 overflow-y-auto z-50">
          {isLoading ? (
            <div className="p-4 text-center text-muted">جاري البحث...</div>
          ) : (
            results.map((result, index) => (
              <button
                key={result.id}
                onClick={() => handleResultClick(result)}
                className={clsx(
                  'w-full px-4 py-3 flex items-center gap-3 text-right transition-colors',
                  index === selectedIndex ? 'bg-primary-10' : 'hover:bg-muted-10'
                )}
              >
                <span className="text-2xl">{result.icon}</span>
                <div className="flex-1">
                  <div className="font-medium">{result.title}</div>
                  <div className="text-sm text-muted">{result.subtitle}</div>
                </div>
                <div className="text-xs text-muted bg-muted-10 px-2 py-1 rounded">
                  {result.type}
                </div>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  )
}

// Debounce utility
function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout | null = null
  return (...args: Parameters<T>) => {
    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}