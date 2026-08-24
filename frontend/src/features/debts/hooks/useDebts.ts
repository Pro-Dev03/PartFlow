import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { debtsApi } from '../../../services/api/endpoints';
import { Debt, DebtStats } from '../types/debts.types';

export function useDebts() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');

  const { data: overdueCustomersData, isLoading } = useQuery({
    queryKey: ['debts'],
    queryFn: () => debtsApi.list({ page: 1, per_page: 100 }),
  });

  const recordPaymentMutation = useMutation({
    mutationFn: ({ customerId, amount }: { customerId: string; amount: number }) =>
      debtsApi.recordPayment(customerId, { amount }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['debts'] });
    },
  });

  const overdueCustomers = (overdueCustomersData?.data as any[]) || [];

  // Transform customer data to debt entries for display
  const debts = overdueCustomers.flatMap((customer: any) => {
    return (customer.debts || []).map((debt: any) => ({
      ...debt,
      customer: {
        id: customer.id,
        name: customer.name,
        code: customer.code,
      },
    }));
  });

  // Calculate stats
  const stats: DebtStats = {
    totalDebt: debts.reduce((sum, debt) => sum + (debt.amount || 0), 0),
    paidAmount: debts.reduce((sum, debt) => sum + ((debt.amount || 0) - (debt.remaining_amount || 0)), 0),
    remainingAmount: debts.reduce((sum, debt) => sum + (debt.remaining_amount || 0), 0),
    overdueCount: debts.filter((debt) => debt.status === 'overdue').length,
    customerCount: overdueCustomers.length,
  };

  // Filter debts
  const filteredDebts = debts.filter((debt: Debt) => {
    const customerName = debt.customer?.name || '';
    const customerCode = debt.customer?.code || '';
    return (
      customerName.toLowerCase().includes(searchQuery.toLowerCase()) ||
      customerCode.toLowerCase().includes(searchQuery.toLowerCase())
    );
  });

  return {
    debts,
    filteredDebts,
    stats,
    isLoading,
    searchQuery,
    setSearchQuery,
    recordPaymentMutation,
  };
}
