import { useState, useEffect, useCallback } from 'react';
import { Search, X, ChevronDown, ChevronUp, Clock, Filter, History } from 'lucide-react';
import { SearchInput } from '../../../components/ui/search-input';

interface AdvancedSearchProps {
  onSearch: (query: string, filters: SearchFilters) => void;
  customers?: Array<{ id: string; name: string; code?: string; phone?: string }>;
  loading?: boolean;
}

export interface SearchFilters {
  status?: 'all' | 'paid' | 'overdue' | 'partial';
  dateRange?: 'today' | 'week' | 'month' | 'custom';
  amountRange?: { min?: number; max?: number };
  searchType?: 'name' | 'code' | 'phone' | 'amount' | 'all';
}

export function AdvancedSearch({ onSearch, customers = [], loading = false }: AdvancedSearchProps) {
  const [query, setQuery] = useState('');
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [filters, setFilters] = useState<SearchFilters>({});
  const [searchHistory, setSearchHistory] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [suggestions, setSuggestions] = useState<Array<{ id: string; name: string; type: string }>>([]);

  // تحميل تاريخ البحث من localStorage
  useEffect(() => {
    const saved = localStorage.getItem('debtSearchHistory');
    if (saved) {
      setSearchHistory(JSON.parse(saved));
    }
  }, []);

  // حفظ تاريخ البحث
  const saveToHistory = useCallback((searchQuery: string) => {
    if (!searchQuery.trim()) return;
    
    setSearchHistory(prev => {
      const updated = [searchQuery, ...prev.filter(q => q !== searchQuery)].slice(0, 5);
      localStorage.setItem('debtSearchHistory', JSON.stringify(updated));
      return updated;
    });
  }, []);

  // اقتراحات أثناء الكتابة
  useEffect(() => {
    if (!query || query.length < 1) {
      setSuggestions([]);
      setShowSuggestions(false);
      return;
    }

    const queryLower = query.toLowerCase();

    const matched = customers
      .filter(customer => 
        customer.name && customer.name.toLowerCase().includes(queryLower) ||
        customer.code && customer.code.toLowerCase().includes(queryLower) ||
        customer.phone && customer.phone.includes(query)
      )
      .slice(0, 5)
      .map(customer => ({
        id: customer.id,
        name: customer.name,
        type: customer.code ? 'code' : 'name'
      }));

    setSuggestions(matched);
    setShowSuggestions(matched.length > 0);
  }, [query, customers]);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      if (query || Object.keys(filters).length > 0) {
        onSearch(query, filters);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [query, filters, onSearch]);

  const handleSearch = () => {
    saveToHistory(query);
    onSearch(query, filters);
    setShowSuggestions(false);
  };

  const clearSearch = () => {
    setQuery('');
    setFilters({});
    onSearch('', {});
  };

  const selectSuggestion = (suggestion: { name: string }) => {
    setQuery(suggestion.name);
    setShowSuggestions(false);
    saveToHistory(suggestion.name);
    onSearch(suggestion.name, filters);
  };

  const selectFromHistory = (historyQuery: string) => {
    setQuery(historyQuery);
    saveToHistory(historyQuery);
    onSearch(historyQuery, filters);
  };

  // تحليل البحث المتقدم
  const parseAdvancedQuery = (searchQuery: string) => {
    let processedQuery = searchQuery;
    
    // Amount operators: >100, <50, 100-500
    const amountMatch = searchQuery.match(/([<>])(\d+)/);
    if (amountMatch) {
      const operator = amountMatch[1];
      const value = parseFloat(amountMatch[2]);
      setFilters(prev => ({
        ...prev,
        amountRange: operator === '>' ? { min: value } : { max: value }
      }));
      processedQuery = searchQuery.replace(/[<>]\d+/, '').trim();
    }

    const rangeMatch = searchQuery.match(/(\d+)-(\d+)/);
    if (rangeMatch) {
      const min = parseFloat(rangeMatch[1]);
      const max = parseFloat(rangeMatch[2]);
      setFilters(prev => ({
        ...prev,
        amountRange: { min, max }
      }));
      processedQuery = searchQuery.replace(/\d+-\d+/, '').trim();
    }

    return processedQuery;
  };

  const handleQueryChange = (value: string) => {
    const processed = parseAdvancedQuery(value);
    setQuery(processed);
  };

  return (
    <div style={{
      border: '1px solid var(--card-border)',
      borderRadius: '12px',
      background: 'var(--card-bg)',
      boxShadow: 'var(--shadow-card)',
      transition: '200ms',
      padding: '12px',
    }}>
      <div style={{ padding: '10px' }}>
        {/* Main Search Input */}
        <div className="w-full">
          <div className="flex items-center gap-2 mb-1">
            <Search className="lucide lucide-search transition-all duration-300 text-text-muted/50 scale-100 w-4 h-4" />
            <span className="text-xs text-text-muted/50">
              {filters.searchType === 'name' ? 'ابحث بالاسم...' :
               filters.searchType === 'code' ? 'ابحث بالكود...' :
               filters.searchType === 'phone' ? 'ابحث برقم الهاتف...' :
               filters.searchType === 'amount' ? 'ابحث بالمبلغ...' :
               'ابحث بالاسم أو الكود أو الهاتف...'}
            </span>
          </div>
          
          <div className="relative w-full md:w-[500px] lg:w-[600px]">
            <SearchInput
              placeholder={filters.searchType === 'name' ? 'ابحث بالاسم...' :
               filters.searchType === 'code' ? 'ابحث بالكود...' :
               filters.searchType === 'phone' ? 'ابحث برقم الهاتف...' :
               filters.searchType === 'amount' ? 'ابحث بالمبلغ...' :
               'ابحث بالاسم أو الكود أو الهاتف...'}
              value={query}
              onChange={(e) => handleQueryChange(e.target.value)}
              onClear={clearSearch}
              onFocus={() => setShowSuggestions(true)}
              onBlur={() => {
                setTimeout(() => setShowSuggestions(false), 200);
              }}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  handleSearch();
                }
              }}
              size="sm"
            />
          </div>

          {/* Suggestions Dropdown */}
          {showSuggestions && suggestions.length > 0 && (
            <div style={{
              position: 'relative',
              marginTop: '4px',
              width: '100%',
              maxWidth: '600px',
              background: 'var(--bg-surface-elevated)',
              border: '1px solid var(--border-default)',
              borderRadius: '8px',
              boxShadow: '0 4px 12px rgba(0, 0, 0, 0.15)',
              zIndex: 1,
              maxHeight: '180px',
              overflowY: 'auto',
            }}>
              {suggestions.map((suggestion, index) => (
                <div
                  key={suggestion.id}
                  onClick={() => selectSuggestion(suggestion)}
                  style={{
                    padding: '8px 12px',
                    cursor: 'pointer',
                    borderBottom: index < suggestions.length - 1 ? '1px solid var(--border-subtle)' : 'none',
                    fontSize: '13px',
                    color: 'var(--text-primary)',
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.background = 'var(--bg-surface-hover)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = 'transparent';
                  }}
                >
                  {suggestion.name}
                </div>
              ))}
            </div>
          )}

          {/* Search History */}
          {searchHistory.length > 0 && !query && (
            <div style={{ marginTop: '12px' }}>
              <div className="flex items-center gap-2 mb-2">
                <History style={{ width: '14px', height: '14px', color: 'var(--text-muted)' }} />
                <span className="text-xs text-text-muted/50">آخر عمليات البحث</span>
              </div>
              <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
                {searchHistory.map((historyItem, index) => (
                  <button
                    key={index}
                    onClick={() => selectFromHistory(historyItem)}
                    style={{
                      padding: '6px 12px',
                      background: 'var(--bg-surface-elevated)',
                      border: '1px solid var(--border-subtle)',
                      borderRadius: '6px',
                      fontSize: '12px',
                      color: 'var(--text-secondary)',
                      cursor: 'pointer',
                      transition: 'all 0.15s ease',
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.background = 'var(--bg-surface-hover)';
                      e.currentTarget.style.borderColor = 'var(--border-primary)';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                      e.currentTarget.style.borderColor = 'var(--border-subtle)';
                    }}
                  >
                    {historyItem}
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Advanced Filters Toggle */}
        <button
          onClick={() => setShowAdvanced(!showAdvanced)}
          style={{
            marginTop: '10px',
            display: 'flex',
            alignItems: 'center',
            gap: '6px',
            background: 'transparent',
            border: 'none',
            color: 'var(--text-secondary)',
            fontSize: '12px',
            cursor: 'pointer',
            padding: '6px 10px',
            borderRadius: '6px',
            transition: 'all 0.15s ease',
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.background = 'var(--bg-surface-elevated)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = 'transparent';
          }}
        >
          <Filter style={{ width: '12px', height: '12px' }} />
          <span>خيارات متقدمة</span>
          {showAdvanced ? <ChevronUp style={{ width: '12px', height: '12px' }} /> : <ChevronDown style={{ width: '12px', height: '12px' }} />}
        </button>

        {/* Advanced Filters Panel */}
        {showAdvanced && (
          <div style={{
            marginTop: '10px',
            padding: '12px',
            background: 'var(--bg-surface-elevated)',
            borderRadius: '10px',
            border: '1px solid var(--border-subtle)',
          }}>
            {/* Search Type */}
            <div style={{ marginBottom: '10px' }}>
              <label style={{ 
                display: 'block', 
                fontSize: '11px', 
                fontWeight: '600', 
                color: 'var(--text-secondary)', 
                marginBottom: '6px' 
              }}>
                نوع البحث
              </label>
              <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
                {[
                  { value: 'all', label: 'الكل' },
                  { value: 'name', label: 'الاسم' },
                  { value: 'code', label: 'الكود' },
                  { value: 'phone', label: 'الهاتف' },
                  { value: 'amount', label: 'المبلغ' },
                ].map((type) => (
                  <button
                    key={type.value}
                    onClick={() => setFilters(prev => ({ ...prev, searchType: type.value as any }))}
                    style={{
                      padding: '5px 10px',
                      background: filters.searchType === type.value ? 'var(--primary)' : 'var(--bg-surface)',
                      border: filters.searchType === type.value ? '1px solid var(--primary)' : '1px solid var(--border-default)',
                      borderRadius: '6px',
                      fontSize: '12px',
                      color: filters.searchType === type.value ? '#FFFFFF' : 'var(--text-primary)',
                      cursor: 'pointer',
                      transition: 'all 0.15s ease',
                    }}
                  >
                    {type.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Status Filter */}
            <div style={{ marginBottom: '10px' }}>
              <label style={{ 
                display: 'block', 
                fontSize: '11px', 
                fontWeight: '600', 
                color: 'var(--text-secondary)', 
                marginBottom: '6px' 
              }}>
                حالة الدين
              </label>
              <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
                {[
                  { value: 'all', label: 'الكل' },
                  { value: 'paid', label: 'مدفوع' },
                  { value: 'overdue', label: 'متأخر' },
                  { value: 'partial', label: 'جزئي' },
                ].map((status) => (
                  <button
                    key={status.value}
                    onClick={() => setFilters(prev => ({ ...prev, status: status.value as any }))}
                    style={{
                      padding: '5px 10px',
                      background: filters.status === status.value ? 'var(--success)' : 'var(--bg-surface)',
                      border: filters.status === status.value ? '1px solid var(--success)' : '1px solid var(--border-default)',
                      borderRadius: '6px',
                      fontSize: '12px',
                      color: filters.status === status.value ? '#FFFFFF' : 'var(--text-primary)',
                      cursor: 'pointer',
                      transition: 'all 0.15s ease',
                    }}
                  >
                    {status.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Date Range Filter */}
            <div style={{ marginBottom: '10px' }}>
              <label style={{ 
                display: 'block', 
                fontSize: '11px', 
                fontWeight: '600', 
                color: 'var(--text-secondary)', 
                marginBottom: '6px' 
              }}>
                الفترة الزمنية
              </label>
              <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
                {[
                  { value: 'all', label: 'الكل' },
                  { value: 'today', label: 'اليوم' },
                  { value: 'week', label: 'هذا الأسبوع' },
                  { value: 'month', label: 'هذا الشهر' },
                ].map((range) => (
                  <button
                    key={range.value}
                    onClick={() => setFilters(prev => ({ ...prev, dateRange: range.value as any }))}
                    style={{
                      padding: '5px 10px',
                      background: filters.dateRange === range.value ? 'var(--info)' : 'var(--bg-surface)',
                      border: filters.dateRange === range.value ? '1px solid var(--info)' : '1px solid var(--border-default)',
                      borderRadius: '6px',
                      fontSize: '12px',
                      color: filters.dateRange === range.value ? '#FFFFFF' : 'var(--text-primary)',
                      cursor: 'pointer',
                      transition: 'all 0.15s ease',
                    }}
                  >
                    {range.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Amount Range Filter */}
            <div>
              <label style={{ 
                display: 'block', 
                fontSize: '11px', 
                fontWeight: '600', 
                color: 'var(--text-secondary)', 
                marginBottom: '6px' 
              }}>
                نطاق المبلغ
              </label>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <input
                  type="number"
                  placeholder="من"
                  value={filters.amountRange?.min || ''}
                  onChange={(e) => setFilters(prev => ({
                    ...prev,
                    amountRange: { ...prev.amountRange, min: e.target.value ? parseFloat(e.target.value) : undefined }
                  }))}
                  style={{
                    flex: 1,
                    padding: '6px 10px',
                    background: 'var(--bg-surface)',
                    border: '1px solid var(--border-default)',
                    borderRadius: '6px',
                    fontSize: '12px',
                    color: 'var(--text-primary)',
                  }}
                />
                <span style={{ color: 'var(--text-muted)' }}>-</span>
                <input
                  type="number"
                  placeholder="إلى"
                  value={filters.amountRange?.max || ''}
                  onChange={(e) => setFilters(prev => ({
                    ...prev,
                    amountRange: { ...prev.amountRange, max: e.target.value ? parseFloat(e.target.value) : undefined }
                  }))}
                  style={{
                    flex: 1,
                    padding: '6px 10px',
                    background: 'var(--bg-surface)',
                    border: '1px solid var(--border-default)',
                    borderRadius: '6px',
                    fontSize: '12px',
                    color: 'var(--text-primary)',
                  }}
                />
              </div>
            </div>

            {/* Clear Filters Button */}
            <button
              onClick={() => setFilters({})}
              style={{
                marginTop: '10px',
                width: '100%',
                padding: '8px',
                background: 'var(--bg-surface)',
                border: '1px solid var(--border-default)',
                borderRadius: '6px',
                fontSize: '12px',
                color: 'var(--text-secondary)',
                cursor: 'pointer',
                transition: 'all 0.15s ease',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'var(--bg-surface-hover)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'var(--bg-surface)';
              }}
            >
              مسح الفلاتر
            </button>
          </div>
        )}
      </div>
    </div>
  );
}