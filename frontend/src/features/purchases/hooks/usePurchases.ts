import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { purchasesApi, suppliersApi, productsApi } from '../../../services/api/endpoints';
import { Purchase, PurchaseItem, PurchaseFormData, PurchaseStats } from '../types/purchases.types';

export function usePurchases() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  const { data: purchasesData, isLoading } = useQuery({
    queryKey: ['purchases'],
    queryFn: () => purchasesApi.list({ page: 1, per_page: 100 }),
  });

  const { data: suppliersData, isLoading: suppliersLoading } = useQuery({
    queryKey: ['suppliers'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 100 }),
  });

  const purchases = (purchasesData?.data as Purchase[]) || [];
  const suppliers = (suppliersData?.data as any[]) || [];

  // Calculate stats
  const stats: PurchaseStats = {
    totalPurchases: purchases.length,
    pendingCount: purchases.filter((p) => p.status === 'pending').length,
    receivedCount: purchases.filter((p) => p.status === 'received').length,
    totalCost: purchases.reduce((sum, p) => sum + p.total_cost, 0),
  };

  // Filter purchases
  const filteredPurchases = purchases.filter((purchase) => {
    const matchesSearch =
      purchase.supplier?.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      purchase.id.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = !statusFilter || purchase.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  // Create purchase mutation
  const createPurchaseMutation = useMutation({
    mutationFn: (data: PurchaseFormData) => purchasesApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchases'] });
    },
    onError: (error) => {
      console.error('Error creating purchase:', error);
    },
  });

  // Receive purchase mutation
  const receivePurchaseMutation = useMutation({
    mutationFn: (purchaseId: string) => purchasesApi.receive(purchaseId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchases'] });
    },
    onError: (error) => {
      console.error('Error receiving purchase:', error);
    },
  });

  return {
    purchases,
    filteredPurchases,
    suppliers,
    stats,
    isLoading,
    suppliersLoading,
    searchQuery,
    setSearchQuery,
    statusFilter,
    setStatusFilter,
    createPurchaseMutation,
    receivePurchaseMutation,
  };
}