import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { purchasesApi, suppliersApi } from '../../../services/api/endpoints';
import { Purchase, PurchaseFormData, PurchaseStats } from '../types/purchases.types';
import { SmartDeleteResult } from '../../../types/api'; // ARCHITECTURE-PRINCIPLES.md
import { toast } from 'sonner';
import { useDebounce } from '../../../hooks/useDebounce';

export function usePurchases() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [viewFilter, setViewFilter] = useState<'active' | 'received' | 'archived' | 'all'>('active');
  const debouncedSearchQuery = useDebounce(searchQuery, 250);

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
  const activePurchases = purchases.filter((p) => p.status !== 'cancelled' && p.status !== 'reversed');
  const receivedPurchases = activePurchases.filter((p) => p.status === 'received' || p.status === 'completed');
  const pendingPurchases = activePurchases.filter((p) => p.status !== 'received' && p.status !== 'completed');
  const untaxedPurchases = activePurchases.filter((p) => Number(p.tax_amount || 0) === 0);
  const taxedPurchases = activePurchases.filter((p) => Number(p.tax_amount || 0) > 0);

  const stats: PurchaseStats = {
    totalPurchases: activePurchases.length,
    pendingCount: purchases.filter((p) => p.status === 'pending' || p.status === 'draft').length,
    receivedCount: purchases.filter((p) => p.status === 'received' || p.status === 'completed' || p.status === 'partially_received').length,
    reversedCount: purchases.filter((p) => p.status === 'reversed').length,
    pendingCost: pendingPurchases.reduce((sum, p) => sum + Number(p.total_amount || 0), 0),
    receivedCost: receivedPurchases.reduce((sum, p) => sum + Number(p.total_amount || 0), 0),
    outstandingAmount: activePurchases.reduce(
      (sum, p) => sum + Math.max(0, Number(p.total_amount || 0) - Number(p.paid_amount || 0)),
      0
    ),
    untaxedCount: untaxedPurchases.length,
    untaxedCost: untaxedPurchases.reduce((sum, p) => sum + Number(p.total_amount || 0), 0),
    taxedCount: taxedPurchases.length,
    taxedCost: taxedPurchases.reduce((sum, p) => sum + Number(p.total_amount || 0), 0),
  };

  const filteredPurchases = purchases.filter((purchase: any) => {
    const isArchived = purchase.status === 'reversed' || purchase.status === 'cancelled';
    const isReceived = purchase.status === 'received' || purchase.status === 'completed';
    const matchesView =
      Boolean(statusFilter) ||
      viewFilter === 'all' ||
      (viewFilter === 'archived' && isArchived) ||
      (viewFilter === 'received' && isReceived) ||
      (viewFilter === 'active' && !isArchived && !isReceived);
    const matchesSearch =
      (purchase.invoice_number || '').toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
      (purchase.supplier?.name || purchase.supplier_name || '').toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
      purchase.id.toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
      (purchase.notes || '').toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
      (purchase.items || []).some((item: any) =>
        (item.product_name || '').toLowerCase().includes(debouncedSearchQuery.toLowerCase())
      );
    const normalizedStatus = purchase.status === 'completed' ? 'received' : purchase.status;
    const matchesStatus = !statusFilter || normalizedStatus === statusFilter;
    return matchesView && matchesSearch && matchesStatus;
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
    onError: (error: any) => {
      console.error('Error receiving purchase:', error);
      const message = error?.message || error?.response?.data?.error;
      toast.error(
        message === 'cannot receive purchase before recording a payment'
          ? 'لا يمكن استلام المشتريات قبل تسجيل دفعة واحدة على الأقل'
          : `خطأ في استلام المشتريات: ${message || 'حدث خطأ غير معروف'}`
      );
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
    viewFilter,
    setViewFilter,
    createPurchaseMutation,
    receivePurchaseMutation,
    updatePurchaseMutation,
    deletePurchaseMutation,
    reversePurchaseMutation,
  };
}