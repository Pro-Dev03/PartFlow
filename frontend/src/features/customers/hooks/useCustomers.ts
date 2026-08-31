import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { customersApi } from '../../../services/api/endpoints';
import { Customer, CustomerFormData, SortConfig } from '../types/customers.types';
import { useDebounce } from '../../../hooks/useDebounce';

export function useCustomers() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const debouncedSearchQuery = useDebounce(searchQuery, 300);
  const [sortConfig, setSortConfig] = useState<SortConfig>({ key: '', direction: null });

  // Fetch customers with debounce search for scalability
  const { data: customersData, isLoading } = useQuery({
    queryKey: ['customers', debouncedSearchQuery],
    queryFn: () => {
      if (debouncedSearchQuery) {
        // Search mode - use API search when query exists
        return customersApi.list({
          page: 1,
          per_page: 50,
          search: debouncedSearchQuery
        });
      } else {
        // Initial load - fetch limited results for performance
        return customersApi.list({ page: 1, per_page: 50 });
      }
    },
    enabled: true, // Always enabled, but will refetch when search changes
  });

  // The backend uses snake_case for persisted fields while the customer UI
  // historically used camelCase summary fields. Normalize both shapes here
  // so cards, sorting, exports and statistics all use the same real values.
  const customers = useMemo(() => {
    const rows = Array.isArray(customersData?.data) ? customersData.data : [];
    return rows.map((row: any) => ({
      ...row,
      totalPurchases: Number(row.totalPurchases ?? row.total_purchases ?? 0),
      paidAmount: Number(row.paidAmount ?? row.paid_amount ?? 0),
      outstanding: Number(row.outstanding ?? row.current_balance ?? 0),
      is_active: row.is_active === undefined
        ? true
        : row.is_active === true || row.is_active === 1 || String(row.is_active).toLowerCase() === 'true',
      lastPurchase: row.lastPurchase ?? row.last_purchase ?? undefined,
    })) as Customer[];
  }, [customersData]);

  // Mutations
  const createMutation = useMutation({
    mutationFn: (data: CustomerFormData) => customersApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      toast.success('تم إضافة العميل بنجاح');
    },
    onError: (error: any) => {
      console.error('Create customer failed:', error);
      
      // Show specific error message based on error type
      if (error.arabicMessage) {
        toast.error(error.arabicMessage);
      } else if (error.response?.error?.message) {
        toast.error(error.response.error.message);
      } else if (error.message) {
        toast.error(error.message);
      } else {
        toast.error('فشل إضافة العميل');
      }
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: CustomerFormData }) =>
      customersApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      toast.success('تم تحديث العميل بنجاح');
    },
    onError: (error: any) => {
      console.error('Update customer failed:', error);
      
      // Show specific error message based on error type
      if (error.arabicMessage) {
        toast.error(error.arabicMessage);
      } else if (error.response?.error?.message) {
        toast.error(error.response.error.message);
      } else if (error.message) {
        toast.error(error.message);
      } else {
        toast.error('فشل تحديث العميل');
      }
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (customerId: string) => customersApi.delete(customerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      toast.success('تم حذف العميل بنجاح');
    },
    onError: (error: any) => {
      console.error('Delete customer failed:', error);
      
      // Show specific error message based on error type
      if (error.arabicMessage) {
        toast.error(error.arabicMessage);
      } else if (error.response?.error?.message) {
        toast.error(error.response.error.message);
      } else if (error.message) {
        toast.error(error.message);
      } else {
        toast.error('فشل حذف العميل');
      }
    },
  });

  // Filter and sort logic
  const filteredCustomers = useMemo(() => {
    let result = [...customers];

    // Search filter
    if (searchQuery) {
      result = result.filter((customer: Customer) =>
        customer.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        customer.phone.includes(searchQuery) ||
        customer.code.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    // Sort
    if (sortConfig.key && sortConfig.direction) {
      result.sort((a: Customer, b: Customer) => {
        const aValue = a[sortConfig.key as keyof Customer];
        const bValue = b[sortConfig.key as keyof Customer];
        
        if (aValue === bValue) return 0;
        
        const comparison = aValue < bValue ? -1 : 1;
        return sortConfig.direction === 'asc' ? comparison : -comparison;
      });
    }

    return result;
  }, [customers, searchQuery, sortConfig]);

  // Stats
  const stats = useMemo(() => ({
    totalCustomers: customers.length,
    activeCustomers: customers.filter((c: Customer) => c.is_active !== false && (c.totalPurchases || 0) > 0).length,
    customersWithDebt: customers.filter((c: Customer) => (c.outstanding || 0) > 0).length,
    totalOutstanding: customers.reduce((sum: number, c: Customer) => sum + (c.outstanding || 0), 0),
  }), [customers]);

  const handleSort = (key: string) => {
    let direction: 'asc' | 'desc' | null = 'asc';
    
    if (sortConfig.key === key) {
      if (sortConfig.direction === 'asc') {
        direction = 'desc';
      } else if (sortConfig.direction === 'desc') {
        direction = null;
      }
    }
    
    setSortConfig({ key, direction });
  };

  return {
    // Data
    customers,
    filteredCustomers,
    isLoading,
    stats,
    
    // State
    searchQuery,
    setSearchQuery,
    sortConfig,
    setSortConfig,
    
    // Mutations
    createMutation,
    updateMutation,
    deleteMutation,
    
    // Helpers
    handleSort,
  };
}
