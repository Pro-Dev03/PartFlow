import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { purchasesApi, suppliersApi } from '../../../services/api/endpoints';
import { Purchase, PurchaseFormData, PurchaseStats } from '../types/purchases.types';
import { SmartDeleteResult } from '../../../types/api'; // ARCHITECTURE-PRINCIPLES.md
import { toast } from 'sonner';

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

  const stats: PurchaseStats = {
    totalPurchases: purchases.length,
    pendingCount: purchases.filter((p) => p.status === 'pending' || p.status === 'draft').length,
    receivedCount: purchases.filter((p) => p.status === 'received' || p.status === 'partially_received').length,
    reversedCount: purchases.filter((p) => p.status === 'reversed').length,
    totalCost: purchases.filter((p) => p.status !== 'cancelled' && p.status !== 'reversed').reduce((sum, p) => sum + (p.total_amount || 0), 0),
  };

  const filteredPurchases = purchases.filter((purchase: any) => {
    const matchesSearch =
      (purchase.supplier_name || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
      purchase.id.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = !statusFilter || purchase.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const createPurchaseMutation = useMutation({
    mutationFn: (data: PurchaseFormData) => purchasesApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchases'] });
    },
    onError: (error: any) => {
      console.error('Error creating purchase:', error);
      if (process.env.NODE_ENV === 'development') {
        console.error('Error status:', error.status);
        console.error('Error response:', error.response);
        console.error('Error code:', error.code);
      }
      toast.error(`خطأ في إنشاء الشراء: ${error.message || error.arabicMessage || 'حدث خطأ غير معروف'}`);
    },
  });

  const receivePurchaseMutation = useMutation({
    mutationFn: (purchaseId: string) => purchasesApi.receive(purchaseId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchases'] });
    },
    onError: (error) => {
      console.error('Error receiving purchase:', error);
    },
  });

  const updatePurchaseMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) => purchasesApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchases'] });
    },
    onError: (error) => {
      console.error('Error updating purchase:', error);
    },
  });

  const deletePurchaseMutation = useMutation({
    mutationFn: async (purchaseId: string): Promise<SmartDeleteResult> => {
      const response = await purchasesApi.delete(purchaseId);
      return response.data as SmartDeleteResult; // ARCHITECTURE-PRINCIPLES.md
    },
    onSuccess: (result) => {
      if (result.can_proceed) {
        queryClient.invalidateQueries({ queryKey: ['purchases'] });
        queryClient.invalidateQueries({ queryKey: ['inventory'] });
      }
    },
    onError: (error) => {
      console.error('Error deleting purchase:', error);
      toast.error(`خطأ في حذف الشراء: ${error.message || 'حدث خطأ غير معروف'}`);
    },
  });

  const reversePurchaseMutation = useMutation({
    mutationFn: ({ purchaseId, reason }: { purchaseId: string; reason: string }) =>
      purchasesApi.reverse(purchaseId, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchases'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
    },
    onError: (error) => {
      console.error('Error reversing purchase:', error);
      if (error.message?.includes('some items have been sold')) {
        toast.error('لا يمكن عكس عملية الشراء لأن بعض القطع تم بيعها بالفعل.');
      } else {
        toast.error(`خطأ في عكس الشراء: ${error.message || 'حدث خطأ غير معروف'}`);
      }
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
    updatePurchaseMutation,
    deletePurchaseMutation,
    reversePurchaseMutation,
  };
}