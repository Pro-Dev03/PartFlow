import { useCallback, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { debtsApi } from '../../../services/api/endpoints';
import { Debt, DebtStats, DebtSummaryResponse } from '../types/debts.types';
import { SearchFilters } from '../components/AdvancedSearch';
import { addStoreDays, getStoreDateKey, getStoreToday } from '../../../utils/store-time';

export function useDebts(debtTab: 'open' | 'paid' | 'all' = 'open', customerId?: string) {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [searchFilters, setSearchFilters] = useState<SearchFilters>({});
  const [page, setPage] = useState(1);
  const pageSize = 10;
  const { data: overdueCustomersData, isLoading, isError, refetch } = useQuery({
    queryKey: ['debts', page, pageSize, searchQuery, searchFilters, debtTab, customerId],
    queryFn: () => {
      const today = getStoreToday();
      const dateRange = searchFilters.dateRange;
      const dueDateFrom = dateRange === 'today' || dateRange === 'week' || dateRange === 'month' ? today : undefined;
      const dueDateTo = dateRange === 'today'
        ? today
        : dateRange === 'week'
          ? addStoreDays(today, 7)
          : dateRange === 'month'
            ? addStoreDays(today, 30)
            : undefined;

      return debtsApi.list({
        page,
        per_page: pageSize,
        customer_id: customerId,
        search: searchQuery.trim() || undefined,
        search_type: searchFilters.searchType || 'all',
        status: searchFilters.status && searchFilters.status !== 'all' ? searchFilters.status : undefined,
        tab: debtTab,
        min_amount: searchFilters.amountRange?.min,
        max_amount: searchFilters.amountRange?.max,
        due_date_from: dueDateFrom,
        due_date_to: dueDateTo,
      });
    },
  });

  const updateSearchQuery = useCallback((value: string) => {
    setSearchQuery(value);
    setPage(1);
  }, []);
  const updateSearchFilters = useCallback((value: SearchFilters) => {
    setSearchFilters(value);
    setPage(1);
  }, []);

  const recordPaymentMutation = useMutation({
    mutationFn: ({ customerId, amount, method, reference }: { customerId: string; amount: number; method: 'cash' | 'credit' | 'bank_transfer' | 'check'; reference: string }) =>
      debtsApi.recordPayment(customerId, { amount, method, reference }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['debts'] });
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      queryClient.invalidateQueries({ queryKey: ['reports'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
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
            due_date: getStoreToday(),
            status: 'overdue',
            created_at: new Date().toISOString(),
          }]
        : [];

    return customerDebts.map((debt: any) => ({
      ...debt,
      notes: debt.notes ?? debt.debt_reason ?? debt.reason ?? debt.debtReason ?? '',
      debt_reason: debt.debt_reason ?? debt.debtReason ?? debt.reason ?? debt.notes ?? '',
      status: Number(debt.remaining_amount ?? debt.remainingAmount ?? 0) <= 0 ? 'paid' : debt.status,
      invoiceNumber: debt.invoice_number || debt.invoiceNumber || '',
      dueDate: debt.due_date, // Map due_date to dueDate for consistency
      remainingAmount: debt.remaining_amount, // Map remaining_amount to remainingAmount
      customer: {
        id: customer.id,
        name: customer.name,
        code: customer.code,
        phone: customer.phone,
        notes: customer.notes ?? customer.customer_notes ?? '',
      },
    }));
  });

  const summary = overdueCustomersData?.meta?.summary as DebtSummaryResponse | undefined;
  const stats: DebtStats = {
    totalDebt: Number(summary?.total_debt ?? debts.reduce((sum, debt) => sum + (debt.amount || 0), 0)),
    paidAmount: Number(summary?.paid_amount ?? debts.reduce(
      (sum, debt) => sum + Math.max(0, (debt.amount || 0) - (debt.remaining_amount || 0)),
      0
    )),
    remainingAmount: Number(summary?.remaining_amount ?? debts.reduce((sum, debt) => sum + (debt.remaining_amount || 0), 0)),
    overdueCount: debts.filter((debt) => debt.status === 'overdue').length,
    customerCount: Number(summary?.customer_count ?? new Set(
      debts
        .filter((debt) => Number(debt.remaining_amount || 0) > 0)
        .map((debt) => debt.customer?.id || debt.customer_id)
        .filter(Boolean)
    ).size),
  };

  // Helper function to check date range
  const isInDateRange = (dateString: string, range: string) => {
    if (!dateString) return false;
    const dueDate = getStoreDateKey(dateString);
    const today = getStoreToday();
    if (!dueDate || !today) return false;
    const toOrdinal = (dateKey: string) => {
      const [year, month, day] = dateKey.split('-').map(Number);
      return Date.UTC(year, month - 1, day) / (24 * 60 * 60 * 1000);
    };
    const diffDays = toOrdinal(dueDate) - toOrdinal(today);

    switch (range) {
      case 'today':
        return diffDays === 0;
      case 'week':
        return diffDays >= 0 && diffDays <= 7 && dueDate < addStoreDays(today, 8);
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
    isError,
    refetch,
    searchQuery,
    setSearchQuery: updateSearchQuery,
    searchFilters,
    setSearchFilters: updateSearchFilters,
    recordPaymentMutation,
    page,
    pageSize,
    total: Number(overdueCustomersData?.meta?.total || rawDebts.length),
    setPage,
  };
}
