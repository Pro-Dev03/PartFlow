/**
 * Table Utilities - أدوات الجداول (فلترة وترتيب)
 */

export type SortDirection = 'asc' | 'desc' | null;

export interface SortConfig {
  key: string;
  direction: SortDirection;
}

export interface FilterConfig {
  key: string;
  value: string;
  operator?: 'equals' | 'contains' | 'startsWith' | 'endsWith' | 'greaterThan' | 'lessThan';
}

// دالة الترتيب
export const sortData = <T extends Record<string, any>>(
  data: T[],
  sortConfig: SortConfig
): T[] => {
  if (!sortConfig.key || !sortConfig.direction) {
    return data;
  }

  return [...data].sort((a, b) => {
    const aValue = a[sortConfig.key];
    const bValue = b[sortConfig.key];

    if (aValue === bValue) return 0;

    const comparison = aValue < bValue ? -1 : 1;
    return sortConfig.direction === 'asc' ? comparison : -comparison;
  });
};

// دالة الفلترة
export const filterData = <T extends Record<string, any>>(
  data: T[],
  filters: FilterConfig[]
): T[] => {
  if (filters.length === 0) {
    return data;
  }

  return data.filter((item) => {
    return filters.every((filter) => {
      const value = item[filter.key];
      const filterValue = filter.value.toLowerCase();

      if (value === undefined || value === null) {
        return false;
      }

      const itemValue = String(value).toLowerCase();

      switch (filter.operator) {
        case 'equals':
          return itemValue === filterValue;
        case 'contains':
          return itemValue.includes(filterValue);
        case 'startsWith':
          return itemValue.startsWith(filterValue);
        case 'endsWith':
          return itemValue.endsWith(filterValue);
        case 'greaterThan':
          return Number(value) > Number(filterValue);
        case 'lessThan':
          return Number(value) < Number(filterValue);
        default:
          return itemValue.includes(filterValue);
      }
    });
  });
};

// دالة معالجة تغيير الترتيب
export const handleSort = (
  key: string,
  currentSort: SortConfig,
  onSortChange: (sort: SortConfig) => void
) => {
  let newDirection: SortDirection = 'asc';

  if (currentSort.key === key) {
    if (currentSort.direction === 'asc') {
      newDirection = 'desc';
    } else if (currentSort.direction === 'desc') {
      newDirection = null;
    }
  }

  onSortChange({ key, direction: newDirection });
};

// دالة معالجة تغيير الفلتر
export const handleFilter = (
  key: string,
  value: string,
  currentFilters: FilterConfig[],
  onFiltersChange: (filters: FilterConfig[]) => void,
  operator?: FilterConfig['operator']
) => {
  const existingFilterIndex = currentFilters.findIndex((f) => f.key === key);
  const newFilters = [...currentFilters];

  if (existingFilterIndex >= 0) {
    if (value === '') {
      newFilters.splice(existingFilterIndex, 1);
    } else {
      newFilters[existingFilterIndex] = { key, value, operator };
    }
  } else if (value !== '') {
    newFilters.push({ key, value, operator });
  }

  onFiltersChange(newFilters);
};

// دالة مسح جميع الفلاتر
export const clearAllFilters = (
  onFiltersChange: (filters: FilterConfig[]) => void
) => {
  onFiltersChange([]);
};

// دالة مسح الترتيب
export const clearSort = (
  onSortChange: (sort: SortConfig) => void
) => {
  onSortChange({ key: '', direction: null });
};