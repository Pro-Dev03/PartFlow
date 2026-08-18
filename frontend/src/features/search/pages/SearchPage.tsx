import { useState } from 'react'
import { AdvancedSearch } from '../components/AdvancedSearch'
import type { SearchResult, SearchFilters } from '../types/search'

export function SearchPage() {
  const [selectedResult, setSelectedResult] = useState<SearchResult | null>(null)
  const [searchHistory, setSearchHistory] = useState<string[]>([])

  const handleSearch = (query: string, filters?: SearchFilters) => {
    // Add to search history
    if (query && !searchHistory.includes(query)) {
      setSearchHistory([query, ...searchHistory.slice(0, 9)])
    }
    // TODO: Perform actual search
    console.log('Searching:', query, filters)
  }

  const handleResultClick = (result: SearchResult) => {
    setSelectedResult(result)
    // TODO: Navigate to result URL
    console.log('Navigating to:', result.url)
  }

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">بحث متقدم</h1>
        <p className="text-muted">ابحث في جميع بيانات النظام</p>
      </div>

      {/* Search Component */}
      <div className="mb-8">
        <AdvancedSearch
          onSearch={handleSearch}
          onResultClick={handleResultClick}
          placeholder="ابحث عن منتجات، عملاء، مبيعات..."
        />
      </div>

      {/* Search History */}
      {searchHistory.length > 0 && (
        <div className="mb-8">
          <h3 className="text-lg font-semibold text-text mb-4">عمليات البحث الأخيرة</h3>
          <div className="flex flex-wrap gap-2">
            {searchHistory.map((query, index) => (
              <button
                key={index}
                onClick={() => {/* TODO: Restore search */}}
                className="px-3 py-2 bg-muted-10 rounded-lg hover:bg-muted-20 transition-colors text-sm text-text"
              >
                {query}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Selected Result Details */}
      {selectedResult && (
        <div className="bg-surface rounded-lg p-6 border border-border">
          <h3 className="text-lg font-semibold text-text mb-4">تفاصيل النتيجة</h3>
          <div className="space-y-3">
            <div className="flex justify-between">
              <span className="text-muted">العنوان:</span>
              <span className="font-medium text-text">{selectedResult.title}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted">النوع:</span>
              <span className="font-medium text-text">{selectedResult.type}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted">الوصف:</span>
              <span className="font-medium text-text">{selectedResult.subtitle}</span>
            </div>
            {selectedResult.details && (
              <div className="flex justify-between">
                <span className="text-muted">التفاصيل:</span>
                <span className="font-medium text-text">{selectedResult.details}</span>
              </div>
            )}
          </div>
          <button
            onClick={() => {/* TODO: Navigate to full result */}}
            className="mt-4 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
          >
            عرض الكامل
          </button>
        </div>
      )}
    </div>
  )
}
