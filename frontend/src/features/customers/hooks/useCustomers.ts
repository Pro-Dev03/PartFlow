import { useState, useMemo, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { customersApi, dashboardApi } from '../../../services/api/endpoints';
import { Customer, CustomerFormData, SortConfig } from '../types/customers.types';
import { useDebounce } from '../../../hooks/useDebounce';

export function useCustomers() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(1);
  const pageSize = 10;
  const debouncedSearchQuery = useDebounce(searchQuery, 300);
  const [sortConfig, setSortConfig] = useState<SortConfig>({ key: '', direction: null });
  const normalizedSearchQuery = debouncedSearchQuery.trim();

  // Fetch customers with debounce search for scalability
  const { data: customersData, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['customers', normalizedSearchQuery, page, sortConfig.key, sortConfig.direction],
    queryFn: () => {
      const sortBy = sortConfig.key === 'totalPurchases' ? 'total_purchases' : sortConfig.key;
      const sortParams = sortConfig.direction && sortBy
        ? { sort_by: sortBy as 'name' | 'total_purchases', sort_order: sortConfig.direction }
        : {};
      if (normalizedSearchQuery) {
        return customersApi.list({
          page,
          per_page: pageSize,
          search: normalizedSearchQuery,
          is_active: true,
          ...sortParams,
        });
      } else {
        return customersApi.list({ page, per_page: pageSize, is_active: true, ...sortParams });
      }
    },
    enabled: true,
  });

  const { data: dashboardData } = useQuery({
    queryKey: ['dashboard'],
    queryFn: () => dashboardApi.getStats(),
    staleTime: 60_000,
  });

  useEffect(() => {
    setPage(1);
  }, [debouncedSearchQuery, sortConfig.key, sortConfig.direction]);

  const handleSearchQueryChange = (value: string) => {
    setSearchQuery(value);
    setPage(1);
  };

  // The backend uses snake_case for persisted fields while the customer UI
  // historically used camelCase summary fields. Normalize both shapes here
  // so cards, sorting, exports and statistics all use the same real values.
  const customers = useMemo(() => {
    const rows = Array.isArray(customersData?.data) ? customersData.data : [];
    return rows.map((row: any) => ({
      ...row,
      name: String(row.name ?? ''),
      code: String(row.code ?? ''),
      phone: String(row.phone ?? ''),
      notes: row.notes ?? row.customer_notes ?? '',
      debt_reason: row.debt_reason ?? row.debtReason ?? row.reason ?? '',
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
      queryClient.invalidateQueries({ queryKey: ['debts'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
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
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
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
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      toast.success('تمت أرشفة العميل بنجاح');
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
        toast.error('فشلت أرشفة العميل');
      }
    },
  });

  const filteredCustomers = customers;
  const dashboardStats = dashboardData?.data ?? dashboardData;
  const stats = useMemo(() => ({
    totalCustomers: Number(customersData?.meta?.total ?? customersData?.total ?? 0),
    activeCustomers: Number(dashboardStats?.activeCustomers ?? dashboardStats?.active_customers ?? 0),
    customersWithDebt: Number(dashboardStats?.outstandingDebtorCount ?? dashboardStats?.outstanding_debtor_count ?? 0),
    totalOutstanding: Number(dashboardStats?.outstandingDebts ?? dashboardStats?.outstanding_debts ?? 0),
  }), [customersData, dashboardStats]);

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
    isError,
    error,
    refetch,
    stats,
    
    // State
    searchQuery,
    setSearchQuery: handleSearchQueryChange,
    sortConfig,
    setSortConfig,
    
    // Mutations
    createMutation,
    updateMutation,
    deleteMutation,
    
    // Helpers
    handleSort,
    page,
    pageSize,
    total: Number(customersData?.meta?.total ?? customersData?.total ?? customersData?.data?.length ?? 0),
    setPage,
  };
}
