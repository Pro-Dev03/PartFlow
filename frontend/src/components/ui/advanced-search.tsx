import { useState } from 'react';
import { Search, Filter, X, ChevronDown } from 'lucide-react';
import { Button } from './button';
import { Input } from './input';
import { Badge } from './badge';
import { cn } from '../../lib/utils/helpers';

interface FilterOption {
  value: string;
  label: string;
}

interface FilterConfig {
  key: string;
  label: string;
  type: 'select' | 'text' | 'date' | 'number';
  options?: FilterOption[];
}

interface AdvancedSearchProps {
  onSearch: (query: string, filters: Record<string, any>) => void;
  filters?: FilterConfig[];
  placeholder?: string;
  className?: string;
}

export function AdvancedSearch({
  onSearch,
  filters = [],
  placeholder = 'بحث...',
  className,
}: AdvancedSearchProps) {
  const [query, setQuery] = useState('');
  const [activeFilters, setActiveFilters] = useState<Record<string, any>>({});
  const [showFilters, setShowFilters] = useState(false);

  const handleSearch = () => {
    onSearch(query, activeFilters);
  };

  const handleFilterChange = (key: string, value: any) => {
    setActiveFilters(prev => {
      const newFilters = { ...prev };
      if (value === '' || value === null || value === undefined) {
        delete newFilters[key];
      } else {
        newFilters[key] = value;
      }
      return newFilters;
    });
  };

  const clearFilter = (key: string) => {
    handleFilterChange(key, '');
  };

  const clearAll = () => {
    setQuery('');
    setActiveFilters({});
    onSearch('', {});
  };

  const activeFilterCount = Object.keys(activeFilters).length;

  return (
    <div className={cn('space-y-4', className)}>
      {/* Search Bar */}
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <Input
            placeholder={placeholder}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            className="pr-10"
          />
        </div>
        <Button onClick={handleSearch}>
          <Search className="w-4 h-4" />
        </Button>
        {filters.length > 0 && (
          <Button
            variant="outline"
            onClick={() => setShowFilters(!showFilters)}
            className={cn(showFilters && 'bg-gray-100 dark:bg-gray-800')}
          >
            <Filter className="w-4 h-4" />
            {activeFilterCount > 0 && (
              <Badge variant="secondary" className="mr-2">
                {activeFilterCount}
              </Badge>
            )}
          </Button>
        )}
        {(query || activeFilterCount > 0) && (
          <Button variant="ghost" onClick={clearAll}>
            <X className="w-4 h-4" />
          </Button>
        )}
      </div>

      {/* Advanced Filters */}
      {showFilters && filters.length > 0 && (
        <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-medium text-gray-900 dark:text-gray-100">
              فلاتر متقدمة
            </h3>
            <Button variant="ghost" size="sm" onClick={clearAll}>
              مسح الكل
            </Button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filters.map((filter) => (
              <div key={filter.key}>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  {filter.label}
                </label>
                {filter.type === 'select' && filter.options ? (
                  <select
                    value={activeFilters[filter.key] || ''}
                    onChange={(e) => handleFilterChange(filter.key, e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:text-gray-100"
                  >
                    <option value="">الكل</option>
                    {filter.options.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                ) : filter.type === 'text' ? (
                  <Input
                    value={activeFilters[filter.key] || ''}
                    onChange={(e) => handleFilterChange(filter.key, e.target.value)}
                    placeholder={filter.label}
                  />
                ) : filter.type === 'number' ? (
                  <Input
                    type="number"
                    value={activeFilters[filter.key] || ''}
                    onChange={(e) => handleFilterChange(filter.key, parseFloat(e.target.value) || '')}
                    placeholder={filter.label}
                  />
                ) : filter.type === 'date' ? (
                  <Input
                    type="date"
                    value={activeFilters[filter.key] || ''}
                    onChange={(e) => handleFilterChange(filter.key, e.target.value)}
                  />
                ) : null}
              </div>
            ))}
          </div>

          {/* Active Filters */}
          {activeFilterCount > 0 && (
            <div className="flex flex-wrap gap-2 pt-2 border-t border-gray-200 dark:border-gray-700">
              {Object.entries(activeFilters).map(([key, value]) => {
                const filter = filters.find(f => f.key === key);
                if (!filter) return null;

                let displayValue = value;
                if (filter.type === 'select' && filter.options) {
                  const option = filter.options.find(opt => opt.value === value);
                  displayValue = option?.label || value;
                }

                return (
                  <Badge key={key} variant="secondary" className="gap-1">
                    {filter.label}: {displayValue}
                    <button
                      onClick={() => clearFilter(key)}
                      className="ml-1 hover:text-red-500"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </Badge>
                );
              })}
            </div>
          )}

          <div className="flex gap-2 justify-end">
            <Button variant="outline" onClick={() => setShowFilters(false)}>
              إلغاء
            </Button>
            <Button onClick={handleSearch}>
              تطبيق الفلاتر
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}