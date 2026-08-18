export interface SearchResult {
  id: string
  type: 'product' | 'customer' | 'sale' | 'purchase' | 'supplier' | 'expense'
  title: string
  subtitle: string
  details?: string
  url: string
  icon: string
  relevance: number
  metadata?: Record<string, any>
}

export interface SearchFilters {
  type?: 'all' | 'product' | 'customer' | 'sale' | 'purchase' | 'supplier' | 'expense'
  dateRange?: {
    from: string
    to: string
  }
  category?: string
  status?: string
}

export interface SearchSuggestion {
  text: string
  type: 'query' | 'command' | 'recent'
  icon?: string
  action?: string
}
