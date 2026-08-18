import { useState, useEffect, useRef } from 'react'
import { clsx } from 'clsx'
import type { SearchResult, SearchFilters, SearchSuggestion } from '../types/search'

interface AdvancedSearchProps {
  onSearch: (query: string, filters?: SearchFilters) => void
  onResultClick: (result: SearchResult) => void
  placeholder?: string
  className?: string
}

export function AdvancedSearch({ onSearch, onResultClick, placeholder = 'بحث...', className }: AdvancedSearchProps) {
  const [query, setQuery] = useState('')
  const [isOpen, setIsOpen] = useState(false)
  const [results, setResults] = useState<SearchResult[]>([])
  const [suggestions, setSuggestions] = useState<SearchSuggestion[]>([])
  const [filters, setFilters] = useState<SearchFilters>({})
  const [showFilters, setShowFilters] = useState(false)
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const searchRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // TODO: Implement actual search logic
  useEffect(() => {
    if (query.length > 2) {
      // Mock search results - replace with actual API call
      const mockResults: SearchResult[] = []
      setResults(mockResults)
      
      // Mock suggestions
      const mockSuggestions: SearchSuggestion[] = [
        { text: query, type: 'query' },
        { text: 'بحث في المنتجات', type: 'command', action: 'products' },
        { text: 'بحث في العملاء', type: 'command', action: 'customers' },
      ]
      setSuggestions(mockSuggestions)
    } else {
      setResults([])
      setSuggestions([])
    }
  }, [query])

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!isOpen) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setSelectedIndex(prev => (prev + 1) % (results.length + suggestions.length))
        break
      case 'ArrowUp':
        e.preventDefault()
        setSelectedIndex(prev => prev <= 0 ? results.length + suggestions.length - 1 : prev - 1)
        break
      case 'Enter':
        e.preventDefault()
        if (selectedIndex >= 0) {
          const totalItems = suggestions.length + results.length
          if (selectedIndex < suggestions.length) {
            handleSuggestionClick(suggestions[selectedIndex])
          } else {
            handleResultClick(results[selectedIndex - suggestions.length])
          }
        } else {
          handleSearch()
        }
        break
      case 'Escape':
        setIsOpen(false)
        break
    }
  }

  const handleSearch = () => {
    if (query.trim()) {
      onSearch(query, filters)
      setIsOpen(false)
    }
  }

  const handleResultClick = (result: SearchResult) => {
    onResultClick(result)
    setIsOpen(false)
    setQuery('')
  }

  const handleSuggestionClick = (suggestion: SearchSuggestion) => {
    if (suggestion.action) {
      setFilters({ ...filters, type: suggestion.action as any })
    } else {
      setQuery(suggestion.text)
    }
    inputRef.current?.focus()
  }

  const handleFilterChange = (key: keyof SearchFilters, value: any) => {
    setFilters({ ...filters, [key]: value })
  }

  const clearFilters = () => {
    setFilters({})
    setShowFilters(false)
  }

  const getTypeIcon = (type: SearchResult['type']) => {
    const icons = {
      product: '📦',
      customer: '👤',
      sale: '💰',
      purchase: '🛒',
      supplier: '🚚',
      expense: '💸',
    }
    return icons[type] || '📄'
  }

  const getTypeLabel = (type: SearchResult['type']) => {
    const labels = {
      product: 'منتج',
      customer: 'عميل',
      sale: 'بيع',
      purchase: 'مشتراة',
      supplier: 'مورد',
      expense: 'مصروف',
    }
    return labels[type] || type
  }

  return (
    <div ref={searchRef} className={clsx('relative', className)}>
      {/* Search Input */}
      <div className="relative">
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => setIsOpen(true)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          className={clsx(
            'w-full px-4 py-3 pr-12 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent',
            isOpen && 'ring-2 ring-primary border-primary'
          )}
        />
        <button
          onClick={() => setShowFilters(!showFilters)}
          className="absolute left-3 top-1/2 -translate-y-1/2 text-muted hover:text-text p-1"
          title="فلاتر البحث"
        >
          🔍
        </button>
        {query && (
          <button
            onClick={() => {
              setQuery('')
              setResults([])
              setSuggestions([])
            }}
            className="absolute left-10 top-1/2 -translate-y-1/2 text-muted hover:text-text p-1"
          >
            ✕
          </button>
        )}
      </div>

      {/* Filters Panel */}
      {showFilters && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-surface border border-border rounded-lg shadow-lg p-4 z-50">
          <div className="flex items-center justify-between mb-4">
            <h4 className="font-medium text-text">فلاتر البحث</h4>
            <button
              onClick={clearFilters}
              className="text-sm text-primary hover:text-primary-600"
            >
              مسح الكل
            </button>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-text mb-2">النوع</label>
              <select
                value={filters.type || 'all'}
                onChange={(e) => handleFilterChange('type', e.target.value)}
                className="w-full px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              >
                <option value="all">الكل</option>
                <option value="product">منتجات</option>
                <option value="customer">عملاء</option>
                <option value="sale">مبيعات</option>
                <option value="purchase">مشتريات</option>
                <option value="supplier">موردين</option>
                <option value="expense">مصروفات</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-text mb-2">نطاق التاريخ</label>
              <div className="grid grid-cols-2 gap-2">
                <input
                  type="date"
                  value={filters.dateRange?.from || ''}
                  onChange={(e) => handleFilterChange('dateRange', { ...filters.dateRange, from: e.target.value })}
                  className="px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                />
                <input
                  type="date"
                  value={filters.dateRange?.to || ''}
                  onChange={(e) => handleFilterChange('dateRange', { ...filters.dateRange, to: e.target.value })}
                  className="px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-text mb-2">الحالة</label>
              <input
                type="text"
                value={filters.status || ''}
                onChange={(e) => handleFilterChange('status', e.target.value)}
                placeholder="الحالة..."
                className="w-full px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              />
            </div>
          </div>

          <div className="flex justify-end mt-4 pt-4 border-t border-border">
            <button
              onClick={() => {
                handleSearch()
                setShowFilters(false)
              }}
              className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              تطبيق الفلاتر
            </button>
          </div>
        </div>
      )}

      {/* Search Results Dropdown */}
      {isOpen && (results.length > 0 || suggestions.length > 0) && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-surface border border-border rounded-lg shadow-lg max-h-96 overflow-y-auto z-50">
          {/* Suggestions */}
          {suggestions.length > 0 && (
            <div className="p-2 border-b border-border">
              <p className="text-xs text-muted mb-2 px-2">اقتراحات</p>
              {suggestions.map((suggestion, index) => (
                <button
                  key={index}
                  onClick={() => handleSuggestionClick(suggestion)}
                  className={clsx(
                    'w-full text-right px-3 py-2 rounded-lg hover:bg-muted-10 transition-colors flex items-center gap-2',
                    selectedIndex === index && 'bg-primary-10'
                  )}
                >
                  {suggestion.icon && <span>{suggestion.icon}</span>}
                  <span className="text-sm text-text">{suggestion.text}</span>
                  {suggestion.type === 'command' && (
                    <span className="text-xs text-muted mr-auto">↵</span>
                  )}
                </button>
              ))}
            </div>
          )}

          {/* Results */}
          {results.length > 0 && (
            <div className="p-2">
              <p className="text-xs text-muted mb-2 px-2">النتائج ({results.length})</p>
              {results.map((result, index) => (
                <button
                  key={result.id}
                  onClick={() => handleResultClick(result)}
                  className={clsx(
                    'w-full text-right px-3 py-3 rounded-lg hover:bg-muted-10 transition-colors flex items-start gap-3',
                    selectedIndex === suggestions.length + index && 'bg-primary-10'
                  )}
                >
                  <span className="text-xl mt-0.5">{getTypeIcon(result.type)}</span>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <p className="font-medium text-text truncate">{result.title}</p>
                      <span className="text-xs text-muted bg-muted-10 px-2 py-0.5 rounded">
                        {getTypeLabel(result.type)}
                      </span>
                    </div>
                    <p className="text-sm text-muted truncate">{result.subtitle}</p>
                    {result.details && (
                      <p className="text-xs text-muted mt-1 truncate">{result.details}</p>
                    )}
                  </div>
                </button>
              ))}
            </div>
          )}

          {/* No Results */}
          {query.length > 2 && results.length === 0 && suggestions.length === 0 && (
            <div className="p-6 text-center">
              <p className="text-muted">لا توجد نتائج</p>
              <p className="text-sm text-muted mt-1">جرب كلمات مختلفة أو فلاتر أقل تحديداً</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
