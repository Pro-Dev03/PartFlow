import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { debtsApi } from '../../../services/api/endpoints';
import { Debt, DebtStats } from '../types/debts.types';
import { SearchFilters } from '../components/AdvancedSearch';

export function useDebts() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [searchFilters, setSearchFilters] = useState<SearchFilters>({});
  const { data: overdueCustomersData, isLoading, refetch } = useQuery({
    queryKey: ['debts'],
    queryFn: () => debtsApi.list({ page: 1, per_page: 100 }),
  });

  const recordPaymentMutation = useMutation({
    mutationFn: ({ customerId, amount, method }: { customerId: string; amount: number; method: string }) =>
      debtsApi.recordPayment(customerId, { amount, method }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['debts'] });
      queryClient.invalidateQueries({ queryKey: ['customers'] });
    },
  });

  const rawDebts = (overdueCustomersData?.data as any[]) || [];
  const isFlatDebtResponse = rawDebts.some((entry) => entry?.customer_id && !entry?.debts);
  const overdueCustomers = isFlatDebtResponse
    ? rawDebts.map((debt: any) => ({
        id: debt.customer_id,
        name: debt.customer_name,
        code: debt.customer_code || '',
        phone: debt.customer_phone || '',
        debts: [debt],
      }))
    : rawDebts;

  // Transform customer data to debt entries for display
  const debts = overdueCustomers.flatMap((customer: any) => {
    const customerDebts = customer.debts?.length
      ? customer.debts
      : customer.current_balance > 0
        ? [{
            id: `customer-balance-${customer.id}`,
            customer_id: customer.id,
            amount: customer.current_balance + (customer.paid_amount || 0),
            paid_amount: customer.paid_amount || 0,
            remaining_amount: customer.current_balance,
            due_date: new Date().toISOString(),
            status: 'overdue',
            created_at: new Date().toISOString(),
          }]
        : [];

    return customerDebts.map((debt: any) => ({
      ...debt,
      dueDate: debt.due_date, // Map due_date to dueDate for consistency
      remainingAmount: debt.remaining_amount, // Map remaining_amount to remainingAmount
      customer: {
        id: customer.id,
        name: customer.name,
        code: customer.code,
        phone: customer.phone,
      },
    }));
  });

  // Calculate stats
  const stats: DebtStats = {
    totalDebt: debts.reduce((sum, debt) => sum + (debt.amount || 0), 0),
    paidAmount: debts.reduce(
      (sum, debt) => sum + Math.max(0, (debt.amount || 0) - (debt.remaining_amount || 0)),
      0
    ),
    remainingAmount: debts.reduce((sum, debt) => sum + (debt.remaining_amount || 0), 0),
    overdueCount: debts.filter((debt) => debt.status === 'overdue').length,
    customerCount: overdueCustomers.length,
  };

  // Helper function to check date range
  const isInDateRange = (dateString: string, range: string) => {
    if (!dateString) return false;
    const date = new Date(dateString);
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    
    const diffTime = date.getTime() - today.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

    switch (range) {
      case 'today':
        return diffDays === 0;
      case 'week':
        return diffDays >= 0 && diffDays <= 7;
      case 'month':
        return diffDays >= 0 && diffDays <= 30;
      default:
        return true;
    }
  };

  // Filter debts with advanced filters
  const filteredDebts = debts.filter((debt: Debt) => {
    const customerName = debt.customer?.name || '';
    const customerCode = debt.customer?.code || '';
    const customerPhone = debt.customer?.phone || '';
    const debtAmount = debt.amount || 0;
    const debtStatus = debt.status || '';
    const debtDueDate = debt.dueDate || '';

    // Apply search query based on search type
    let matchesQuery = true;
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      switch (searchFilters.searchType) {
        case 'name':
          matchesQuery = customerName.toLowerCase().includes(query);
          break;
        case 'code':
          matchesQuery = customerCode.toLowerCase().includes(query);
          break;
        case 'phone':
          matchesQuery = customerPhone.includes(query);
          break;
        case 'amount':
          matchesQuery = debtAmount.toString().includes(query);
          break;
        default:
          matchesQuery = 
            customerName.toLowerCase().includes(query) ||
            customerCode.toLowerCase().includes(query) ||
            customerPhone.includes(query);
      }
    }

    // Apply status filter
    let matchesStatus = true;
    if (searchFilters.status && searchFilters.status !== 'all') {
      matchesStatus = debtStatus === searchFilters.status;
    }

    // Apply date range filter
    let matchesDateRange = true;
    if (searchFilters.dateRange && searchFilters.dateRange !== 'all') {
      matchesDateRange = isInDateRange(debtDueDate, searchFilters.dateRange);
    }

    // Apply amount range filter
    let matchesAmountRange = true;
    if (searchFilters.amountRange) {
      if (searchFilters.amountRange.min !== undefined) {
        matchesAmountRange = matchesAmountRange && debtAmount >= searchFilters.amountRange.min;
      }
      if (searchFilters.amountRange.max !== undefined) {
        matchesAmountRange = matchesAmountRange && debtAmount <= searchFilters.amountRange.max;
      }
    }

    return matchesQuery && matchesStatus && matchesDateRange && matchesAmountRange;
  });

  return {
    debts,
    filteredDebts,
    stats,
    isLoading,
    refetch,
    searchQuery,
    setSearchQuery,
    searchFilters,
    setSearchFilters,
    recordPaymentMutation,
  };
}
