import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { purchasesApi, suppliersApi } from '../../../services/api/endpoints';
import { Purchase, PurchaseFormData, PurchaseStats } from '../types/purchases.types';
import { SmartDeleteResult } from '../../../types/api'; // ARCHITECTURE-PRINCIPLES.md
import { toast } from 'sonner';
import { useDebounce } from '../../../hooks/useDebounce';
import { isReceivedPurchaseStatus, matchesPurchaseViewFilter, normalizePurchaseStatus } from '../utils/purchase-status';

export function usePurchases() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [viewFilter, setViewFilter] = useState<'active' | 'received' | 'archived' | 'all'>('active');
  const [page, setPage] = useState(1);
  const pageSize = 10;
  const debouncedSearchQuery = useDebounce(searchQuery, 250);

  const { data: purchasesData, isLoading } = useQuery({
    queryKey: ['purchases', page, pageSize],
    queryFn: () => purchasesApi.list({ page, per_page: pageSize }),
  });

  useEffect(() => {
    setPage(1);
  }, [debouncedSearchQuery, statusFilter, viewFilter]);

  const { data: suppliersData, isLoading: suppliersLoading } = useQuery({
    queryKey: ['suppliers'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 100 }),
  });

  const purchases = (purchasesData?.data as Purchase[]) || [];
  const suppliers = (suppliersData?.data as any[]) || [];
  const activePurchases = purchases.filter((p) => {
    const normalized = normalizePurchaseStatus(p.status);
    return normalized !== 'cancelled' && normalized !== 'reversed';
  });
  const receivedPurchases = activePurchases.filter((p) => isReceivedPurchaseStatus(p.status));
  const pendingPurchases = activePurchases.filter((p) => !isReceivedPurchaseStatus(p.status));
  const untaxedPurchases = activePurchases.filter((p) => Number(p.tax_amount || 0) === 0);
  const taxedPurchases = activePurchases.filter((p) => Number(p.tax_amount || 0) > 0);

  const stats: PurchaseStats = {
    totalPurchases: activePurchases.length,
    pendingCount: purchases.filter((p) => {
      const normalized = normalizePurchaseStatus(p.status);
      return normalized === 'pending' || normalized === 'draft';
    }).length,
    receivedCount: purchases.filter((p) => isReceivedPurchaseStatus(p.status)).length,
    reversedCount: purchases.filter((p) => normalizePurchaseStatus(p.status) === 'reversed').length,
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
    const normalizedStatus = normalizePurchaseStatus(purchase.status);
    const matchesView = matchesPurchaseViewFilter(normalizedStatus, viewFilter);
    const matchesSearch =
      (purchase.invoice_number || '').toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
      (purchase.supplier?.name || purchase.supplier_name || '').toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
      purchase.id.toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
      (purchase.notes || '').toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
      (purchase.items || []).some((item: any) =>
        (item.product_name || '').toLowerCase().includes(debouncedSearchQuery.toLowerCase())
      );
    const matchesStatus = !statusFilter || normalizePurchaseStatus(statusFilter) === normalizedStatus;
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
      return response as SmartDeleteResult; // API client returns the SmartDeleteResult body directly.
    },
    onSuccess: (result) => {
      if (result.can_proceed) {
        queryClient.invalidateQueries({ queryKey: ['purchases'] });
        queryClient.invalidateQueries({ queryKey: ['inventory'] });
        queryClient.invalidateQueries({ queryKey: ['reports'] });
        queryClient.invalidateQueries({ queryKey: ['dashboard'] });
        queryClient.invalidateQueries({ queryKey: ['dashboard-activity'] });
        queryClient.invalidateQueries({ queryKey: ['low-stock-items'] });
        queryClient.invalidateQueries({ queryKey: ['overdue-debts'] });
        queryClient.invalidateQueries({ queryKey: ['debts'] });
        queryClient.invalidateQueries({ queryKey: ['suppliers'] });
        queryClient.invalidateQueries({ queryKey: ['supplier-ledger'] });
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
        toast.error('لا يمكن إلغاء عملية الشراء لأن بعض القطع تم بيعها بالفعل.');
      } else {
        toast.error(`خطأ في إلغاء عملية الشراء: ${error.message || 'حدث خطأ غير معروف'}`);
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
    page,
    pageSize,
    total: Number(purchasesData?.meta?.total || purchases.length),
    setPage,
  };
}