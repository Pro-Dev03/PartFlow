import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { customersApi } from '../../../services/api/endpoints';
import { Customer, CustomerFormData, SortConfig } from '../types/customers.types';

export function useCustomers() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [sortConfig, setSortConfig] = useState<SortConfig>({ key: '', direction: null });

  // Fetch customers
  const { data: customersData, isLoading } = useQuery({
    queryKey: ['customers'],
    queryFn: () => customersApi.list({ page: 1, per_page: 100 }),
  });

  const customers = (customersData?.data as Customer[]) || [];

  // Mutations
  const createMutation = useMutation({
    mutationFn: (data: CustomerFormData) => customersApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      toast.success('تم إضافة العميل بنجاح');
    },
    onError: (error) => {
      console.error('Create customer failed:', error);
      toast.error('فشل إضافة العميل');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: CustomerFormData }) =>
      customersApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      toast.success('تم تحديث العميل بنجاح');
    },
    onError: (error) => {
      console.error('Update customer failed:', error);
      toast.error('فشل تحديث العميل');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (customerId: string) => customersApi.delete(customerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      toast.success('تم حذف العميل بنجاح');
    },
    onError: (error) => {
      console.error('Delete customer failed:', error);
      toast.error('فشل حذف العميل');
    },
  });

  // Filter and sort logic
  const filteredCustomers = useMemo(() => {
    let result = [...customers];

    // Search filter
    if (searchQuery) {
      result = result.filter((customer: Customer) =>
        customer.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        customer.phone.includes(searchQuery)
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
    activeCustomers: customers.filter((c: Customer) => c.totalPurchases > 0).length,
    customersWithDebt: customers.filter((c: Customer) => c.outstanding > 0).length,
    totalOutstanding: customers.reduce((sum: number, c: Customer) => sum + c.outstanding, 0),
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