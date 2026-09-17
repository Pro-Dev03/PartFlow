import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { EmptyState } from '../../../components/ui/empty-state';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { getButtonSize } from '../../../config/button-sizes';
import {
  Plus,
  ShoppingCart,
  Eye,
  Download,
  Printer,
  Check,
  Edit,
  Trash2,
  RotateCcw,
  DollarSign,
  UserRound,
  Inbox,
} from 'lucide-react';

// Custom hooks
import { usePurchases } from '../hooks/usePurchases';
import { categoriesApi, purchasesApi } from '../../../services/api/endpoints';

// Components
import { PurchaseStats } from '../components/PurchaseStats';
import { PurchaseFilters } from '../components/PurchaseFilters';

// Types
import { Purchase } from '../types/purchases.types';

// SmartDelete utility (ARCHITECTURE-PRINCIPLES.md)
import { handleSmartDelete } from '../../../utils/smartDelete';
import { ConfirmDialog } from '../../../components/ui/confirm-dialog';
import { Modal } from '../../../components/ui/modal';
import { PaginationControls } from '../../../components/ui/pagination-controls';
import { toast } from 'sonner';

export function PurchasesPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  // Custom hook
  const {
    purchases,
    filteredPurchases,
    stats,
    isLoading,
    searchQuery,
    setSearchQuery,
    statusFilter,
    setStatusFilter,
    viewFilter,
    setViewFilter,
    receivePurchaseMutation,
    deletePurchaseMutation,
    reversePurchaseMutation,
    page,
    pageSize,
    total,
    setPage,
  } = usePurchases();
  const [purchaseToReceive, setPurchaseToReceive] = useState<string | null>(null);
  const [purchaseToPay, setPurchaseToPay] = useState<any | null>(null);
  const [paymentAmount, setPaymentAmount] = useState('');
  const [purchaseToReverse, setPurchaseToReverse] = useState<string | null>(null);
  const [reversalReason, setReversalReason] = useState('');
  const [purchaseToView, setPurchaseToView] = useState<string | null>(null);
  const { data: purchaseDetailsData, isLoading: purchaseDetailsLoading } = useQuery({
    queryKey: ['purchase', purchaseToView],
    queryFn: () => purchasesApi.get(purchaseToView || ''),
    enabled: Boolean(purchaseToView),
  });
  const purchaseDetails = purchaseDetailsData?.data?.purchase || purchaseDetailsData?.purchase;
  const purchaseItems = purchaseDetailsData?.data?.items || purchaseDetailsData?.items || [];
  const purchaseDetailsSupplier =
    purchaseDetails?.supplier ||
    purchaseDetailsData?.data?.supplier ||
    purchaseDetailsData?.supplier;
  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });
  const categories = (categoriesData?.data || categoriesData || []) as Array<{ id: string; name: string }>;
  const purchaseDetailsRemaining =
    purchaseDetails?.remaining ??
    purchaseDetailsData?.data?.remaining ??
    purchaseDetailsData?.remaining ??
    Math.max(0, Number(purchaseDetails?.total_amount || 0) - Number(purchaseDetails?.paid_amount || 0));
  const paymentMutation = useMutation({
    mutationFn: ({ id, amount }: { id: string; amount: number }) =>
      purchasesApi.addPayment(id, { amount, paymentMethod: 'cash' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchases'] });
      setPurchaseToPay(null);
      setPaymentAmount('');
      toast.success('تم تسجيل دفعة الشراء وتحديث الرصيد');
    },
    onError: () => toast.error('تعذر تسجيل دفعة الشراء'),
  });

  // Handle receive purchase
  const handleReceivePurchase = (purchaseId: string) => {
    setPurchaseToReceive(purchaseId);
  };

  // Handle delete purchase with SmartDelete (ARCHITECTURE-PRINCIPLES.md)
  const handleDeletePurchase = async (purchaseId: string) => {
    await handleSmartDelete(
      async () => {
        const result = await deletePurchaseMutation.mutateAsync(purchaseId);
        return result;
      },
      {
        confirmationMessage: 'هل أنت متأكد من حذف هذا الشراء؟',
        showConfirmationDialog: async (message) => window.confirm(message),
        onSuccess: (result) => {
          // Refresh the list or navigate
          console.log('Delete successful:', result);
        },
        onBlocked: (result) => {
          console.log('Delete blocked:', result);
          // Could show a modal with details here
        },
        onError: (error) => {
          console.error('Delete error:', error);
        }
      }
    );
  };

  // Handle reverse purchase
  const handleReversePurchase = (purchaseId: string) => {
    setReversalReason('');
    setPurchaseToReverse(purchaseId);
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
      draft: { label: 'مسودة', variant: 'outline' },
      pending: { label: 'قيد الانتظار', variant: 'secondary' },
      ordered: { label: 'تم الطلب', variant: 'outline' },
      received: { label: 'تم الاستلام', variant: 'default' },
      completed: { label: 'تم الاستلام', variant: 'default' },
      cancelled: { label: 'ملغي', variant: 'destructive' },
      reversed: { label: 'ملغاة', variant: 'danger' },
      partially_received: { label: 'استلام جزئي', variant: 'secondary' },
    };
    return variants[status] || { label: status, variant: 'default' };
  };

  const handleExport = () => {
    const dataToExport = purchases.map((purchase: Purchase) => ({
      'التاريخ': purchase.purchase_date,
      'المورد': purchase.supplier?.name || purchase.supplier_name,
      'الحالة': getStatusBadge(purchase.status).label,
      'الضريبة': Number(purchase.tax_amount || 0),
      'التكلفة': purchase.total_amount
    }));
    exportToCSV(dataToExport, `purchases-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    const dataToPrint = purchases.map((purchase: Purchase) => ({
      'التاريخ': purchase.purchase_date,
      'المورد': purchase.supplier?.name || purchase.supplier_name,
      'الحالة': getStatusBadge(purchase.status).label,
      'الضريبة': Number(purchase.tax_amount || 0),
      'التكلفة': purchase.total_amount
    }));
    printTable(dataToPrint, ['التاريخ', 'المورد', 'الحالة', 'التكلفة'], 'تقرير المشتريات');
  };

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Purchase Management"
        title={t('purchases.title')}
        description="إدارة المشتريات والطلبات"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="primary" size={getButtonSize('customers', 'headerActions')} className="gap-2" onClick={() => navigate('/app/purchases/create')}>
              <Plus className="w-4 h-4" />
              {t('purchases.newPurchase')}
            </Button>
            <Button variant="secondary" size={getButtonSize('customers', 'headerActions')} onClick={handleExport} className="gap-2">
              <Download className="w-4 h-4" />
              تصدير
            </Button>
            <Button variant="secondary" size={getButtonSize('customers', 'headerActions')} onClick={handlePrint} className="gap-2">
              <Printer className="w-4 h-4" />
              طباعة
            </Button>
          </div>
        }
      />

      {/* Stats Cards */}
      <PurchaseStats stats={stats} />

      {/* Search and Filters */}
      <PurchaseFilters
        searchQuery={searchQuery}
        setSearchQuery={setSearchQuery}
        statusFilter={statusFilter}
        setStatusFilter={setStatusFilter}
      />

      <div className="rounded-[16px] border border-[var(--border-default)] bg-[var(--bg-surface)] shadow-[0_10px_28px_rgba(15,23,42,0.05)]">
        <div className="flex flex-col gap-3 border-b border-[var(--border-subtle)] px-5 py-4 md:flex-row md:items-center md:justify-between">
          <div className="flex items-center gap-3">
            <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
              <ShoppingCart className="h-4 w-4" />
            </span>
            <div>
            <h3 className="text-sm font-extrabold text-[var(--text-primary)]">
              {viewFilter === 'active'
                ? 'المشتريات الحالية'
                : viewFilter === 'received'
                  ? 'المشتريات المستلمة'
                  : viewFilter === 'archived'
                    ? 'أرشيف المشتريات'
                    : 'كل المشتريات'}
            </h3>
            <p className="mt-0.5 text-[11px] font-medium text-[var(--text-muted)]">{filteredPurchases.length} عملية شراء مطابقة</p>
            </div>
          </div>

          <div className="flex flex-wrap gap-2" role="tablist" aria-label="عرض المشتريات">
            {['active', 'received', 'archived', 'all'].map((tab) => (
              <Button
                key={tab}
                variant={viewFilter === tab ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => {
                  setViewFilter(tab as 'active' | 'received' | 'archived' | 'all');
                  setStatusFilter('');
                }}
                role="tab"
                aria-selected={viewFilter === tab}
              >
                {tab === 'active' ? 'الحالية' : tab === 'received' ? 'تم الاستلام' : tab === 'archived' ? 'الأرشيف' : 'الكل'}
              </Button>
            ))}
          </div>
        </div>

        {isLoading ? (
          <div className="flex h-64 items-center justify-center">
            <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          </div>
        ) : filteredPurchases.length === 0 ? (
          <EmptyState
            icon={<Inbox className="h-5 w-5" />}
            title="لا توجد مشتريات"
            description="لم يتم العثور على عمليات شراء مطابقة للفلاتر الحالية"
          />
        ) : (
          <div className="hidden overflow-x-auto md:block">
            <Table className="min-w-[1100px]">
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[12%]">رقم الطلب</TableHead>
                  <TableHead className="w-[18%]">المورد</TableHead>
                  <TableHead className="w-[9%]">القطع</TableHead>
                  <TableHead className="w-[11%] text-center">الضريبة</TableHead>
                  <TableHead className="w-[12%] text-center">التكلفة</TableHead>
                  <TableHead className="w-[10%] text-center">المدفوع</TableHead>
                  <TableHead className="w-[10%] text-center">المتبقي</TableHead>
                  <TableHead className="w-[12%]">التاريخ</TableHead>
                  <TableHead className="w-[10%]">الحالة</TableHead>
                  <TableHead className="w-[6%] text-end">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody className="divide-y divide-[var(--border-subtle)]" dir="rtl">
                {filteredPurchases.map((purchase: any) => {
                  const normalizedStatus = purchase.status === 'completed' ? 'received' : purchase.status;
                  const statusBadge = getStatusBadge(purchase.status);
                  return (
                    <TableRow key={purchase.id} dir="rtl" className="transition-all duration-200 hover:bg-[var(--color-primary-05)] [&>td]:h-[76px]">
                      <TableCell className="font-mono text-xs font-extrabold text-[var(--primary)]">{purchase.invoice_number}</TableCell>
                      <TableCell>
                        <div className="flex min-w-0 items-center gap-3">
                          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl border border-[var(--border-subtle)] bg-[var(--color-primary-10)] text-[var(--primary)] shadow-sm">
                            <UserRound className="h-4 w-4" />
                          </div>
                          <div className="min-w-0">
                            <div className="truncate font-black text-[var(--text-primary)]">{purchase.supplier?.name || purchase.supplier_name || 'مورد غير محدد'}</div>
                            <div className="mt-0.5 text-[11px] font-medium text-[var(--text-tertiary)]">مورد مسجل</div>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="font-semibold text-[var(--text-secondary)]">{purchase.total_items || purchase.items?.length || 0} قطع</TableCell>
                      <TableCell className="text-center">
                        {Number(purchase.tax_amount || 0) > 0 ? `₪${Number(purchase.tax_amount).toLocaleString()}` : <Badge variant="outline" size="sm">بدون ضريبة</Badge>}
                      </TableCell>
                      <TableCell className="text-center font-black text-[var(--text-primary)]">₪{purchase.total_amount?.toLocaleString() || '0'}</TableCell>
                      <TableCell className="text-center font-black text-[var(--color-success)]">₪{purchase.paid_amount?.toLocaleString() || '0'}</TableCell>
                      <TableCell className="text-center"><Badge variant={Number(purchase.remaining || 0) > 0 ? 'warning' : 'success'} size="sm" className="min-w-[82px] justify-center rounded-full">₪{purchase.remaining?.toLocaleString() || '0'}</Badge></TableCell>
                      <TableCell className="font-semibold text-[var(--text-secondary)]">{purchase.expected_delivery_date ? new Date(purchase.expected_delivery_date).toLocaleDateString('en-US') : '-'}</TableCell>
                      <TableCell><Badge variant={statusBadge.variant} size="sm" className="whitespace-nowrap rounded-full">{statusBadge.label}</Badge></TableCell>
                      <TableCell className="text-end">
                        <div className="flex items-center justify-end gap-1">
                          {Number(purchase.remaining || 0) > 0 && !['cancelled', 'reversed'].includes(normalizedStatus) && (
                            <Button variant="ghost" size="icon" onClick={(event) => { event.stopPropagation(); setPurchaseToPay(purchase); setPaymentAmount(''); }} className="text-text-secondary hover:text-text-primary" title="تسجيل دفعة" aria-label="تسجيل دفعة">
                              <DollarSign className="h-4 w-4" />
                            </Button>
                          )}
                          {normalizedStatus === 'pending' && (
                            <Button variant="ghost" size="icon" onClick={() => handleReceivePurchase(purchase.id)} className="text-success hover:text-success" title="استلام البضاعة" aria-label="استلام البضاعة">
                              <Check className="h-4 w-4" />
                            </Button>
                          )}
                          {(normalizedStatus === 'pending' || normalizedStatus === 'draft') && (
                            <>
                              <Button variant="ghost" size="icon" onClick={() => navigate(`/app/purchases/edit/${purchase.id}`)} className="text-text-secondary hover:text-text-primary" title="تعديل" aria-label="تعديل">
                                <Edit className="h-4 w-4" />
                              </Button>
                              {Number(purchase.paid_amount || 0) <= 0 && (
                                <Button variant="ghost" size="icon" onClick={() => handleDeletePurchase(purchase.id)} className="text-danger hover:text-danger" title="حذف الطلب غير المدفوع" aria-label="حذف الطلب غير المدفوع">
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              )}
                            </>
                          )}
                          {normalizedStatus === 'received' && (
                            <Button variant="ghost" size="icon" onClick={(event) => { event.stopPropagation(); navigate(`/app/purchases/edit/${purchase.id}`); }} className="text-text-secondary hover:text-text-primary" title="تعديل الشراء وتحديث المخزون" aria-label="تعديل الشراء وتحديث المخزون">
                              <Edit className="h-4 w-4" />
                            </Button>
                          )}
                          {normalizedStatus === 'received' && (
                            <Button variant="ghost" size="icon" onClick={(event) => { event.stopPropagation(); void handleReversePurchase(purchase.id); }} className="text-warning hover:text-warning" title="إلغاء عملية الشراء" aria-label="إلغاء عملية الشراء">
                              <RotateCcw className="h-4 w-4" />
                            </Button>
                          )}
                          <Button variant="ghost" size="icon" onClick={(event) => { event.stopPropagation(); setPurchaseToView(purchase.id); }} className="text-text-secondary hover:text-text-primary" title="عرض التفاصيل" aria-label="عرض التفاصيل">
                            <Eye className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        )}

        <PaginationControls page={page} pageSize={pageSize} total={total} onPageChange={setPage} isLoading={isLoading} />

        <div className="md:hidden">
          {isLoading ? (
            <div className="flex h-64 items-center justify-center"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
          ) : (
            <div className="grid gap-3 p-4">
              {filteredPurchases.map((purchase: any) => {
                const normalizedStatus = purchase.status === 'completed' ? 'received' : purchase.status;
                const statusBadge = getStatusBadge(purchase.status);
                return (
                  <div key={purchase.id} className="rounded-xl border border-border bg-surface-elevated/25 p-4">
                    <div className="mb-3 flex items-start justify-between gap-3">
                      <div>
                        <div className="font-semibold text-text-primary">{purchase.invoice_number}</div>
                        <div className="mt-1 text-[11px] text-text-tertiary">{purchase.supplier?.name || purchase.supplier_name}</div>
                      </div>
                      <Badge variant={statusBadge.variant} size="sm" className="whitespace-nowrap">{statusBadge.label}</Badge>
                    </div>

                    <div className="space-y-2 text-sm">
                      <div className="flex items-center justify-between gap-2"><span className="text-text-tertiary">القطع</span><span className="font-medium text-text-secondary">{purchase.total_items || purchase.items?.length || 0}</span></div>
                      <div className="flex items-center justify-between gap-2"><span className="text-text-tertiary">التكلفة</span><span className="font-semibold text-text-primary">₪{purchase.total_amount?.toLocaleString()}</span></div>
                      <div className="flex items-center justify-between gap-2"><span className="text-text-tertiary">المتبقي</span><span className="font-semibold text-text-secondary">₪{purchase.remaining?.toLocaleString()}</span></div>
                    </div>

                    <div className="mt-3 flex items-center justify-end gap-2 border-t border-border pt-3">
                      {Number(purchase.remaining || 0) > 0 && !['cancelled', 'reversed'].includes(normalizedStatus) && (
                        <Button variant="ghost" size="icon" onClick={() => { setPurchaseToPay(purchase); setPaymentAmount(''); }} className="text-text-secondary hover:text-text-primary" title="تسجيل دفعة" aria-label="تسجيل دفعة">
                          <DollarSign className="h-4 w-4" />
                        </Button>
                      )}
                      {normalizedStatus === 'pending' && (
                        <Button variant="ghost" size="icon" onClick={() => handleReceivePurchase(purchase.id)} className="text-success hover:text-success" title="استلام البضاعة" aria-label="استلام البضاعة">
                          <Check className="h-4 w-4" />
                        </Button>
                      )}
                      {(normalizedStatus === 'pending' || normalizedStatus === 'draft') && Number(purchase.paid_amount || 0) <= 0 && (
                        <Button variant="ghost" size="icon" onClick={() => handleDeletePurchase(purchase.id)} className="text-danger hover:text-danger" title="حذف الطلب غير المدفوع" aria-label="حذف الطلب غير المدفوع">
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                      {normalizedStatus === 'received' && (
                        <Button variant="ghost" size="icon" onClick={() => { void handleReversePurchase(purchase.id); }} className="text-warning hover:text-warning" title="إلغاء عملية الشراء" aria-label="إلغاء عملية الشراء">
                          <RotateCcw className="h-4 w-4" />
                        </Button>
                      )}
                      <Button variant="ghost" size="icon" onClick={() => setPurchaseToView(purchase.id)} aria-label="عرض تفاصيل الشراء" title="عرض تفاصيل الشراء">
                        <Eye className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
      <ConfirmDialog
        isOpen={purchaseToReceive !== null}
        onClose={() => setPurchaseToReceive(null)}
        onConfirm={() => {
          if (purchaseToReceive) {
            receivePurchaseMutation.mutate(purchaseToReceive);
          }
          setPurchaseToReceive(null);
        }}
        title="تأكيد استلام البضاعة"
        message="سيتم إنشاء عناصر المخزون تلقائياً عند الاستلام. هل تريد المتابعة؟"
        confirmText="استلام البضاعة"
        isLoading={receivePurchaseMutation.isPending}
        variant="warning"
      />
      <Modal
        isOpen={purchaseToView !== null}
        onClose={() => setPurchaseToView(null)}
        title="تفاصيل عملية الشراء"
        size="xl"
      >
        {purchaseDetailsLoading ? (
          <div className="p-10 text-center">جاري تحميل التفاصيل...</div>
        ) : purchaseDetails ? (
          <div className="space-y-6">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div><span className="text-sm text-text-muted">رقم الطلب</span><p>{purchaseDetails.invoice_number || '-'}</p></div>
              <div><span className="text-sm text-text-muted">المورد</span><p>{purchaseDetailsSupplier?.name || purchaseDetails?.supplier_name || '-'}</p></div>
              <div><span className="text-sm text-text-muted">الحالة</span><p><Badge>{purchaseDetails.status}</Badge></p></div>
              <div><span className="text-sm text-text-muted">التاريخ</span><p>{purchaseDetails.purchase_date ? new Date(purchaseDetails.purchase_date).toLocaleDateString('en-US') : '-'}</p></div>
              <div><span className="text-sm text-text-muted">الضريبة</span><p>{Number(purchaseDetails.tax_amount || 0) > 0 ? `₪${Number(purchaseDetails.tax_amount).toLocaleString('en-US')}` : 'بدون ضريبة'}</p></div>
              <div><span className="text-sm text-text-muted">الإجمالي</span><p>₪{Number(purchaseDetails.total_amount || 0).toLocaleString('en-US')}</p></div>
              <div><span className="text-sm text-text-muted">المدفوع</span><p>₪{Number(purchaseDetails.paid_amount || 0).toLocaleString('en-US')}</p></div>
              <div><span className="text-sm text-text-muted">المتبقي</span><p>₪{Number(purchaseDetailsRemaining).toLocaleString('en-US')}</p></div>
            </div>
            <div>
              <h2 className="font-semibold mb-3">القطع</h2>
              <div className="space-y-2">
                {purchaseItems.map((item: any) => (
                  <div key={item.id} className="flex justify-between border-b border-border py-2">
                    <div>
                      <div>{item.product_name || item.product?.name || 'قطعة'}</div>
                      <div className="text-xs text-text-muted">
                        التصنيف: {categories.find((category) => category.id === item.category_id)?.name || 'بدون تصنيف'}
                      </div>
                    </div>
                    <span>{item.quantity} × ₪{Number(item.unit_cost || 0).toLocaleString('en-US')}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        ) : (
          <div className="p-10 text-center text-red-500">تعذر تحميل تفاصيل الشراء</div>
        )}
      </Modal>
      <ConfirmDialog
        isOpen={purchaseToReverse !== null}
        onClose={() => setPurchaseToReverse(null)}
        onConfirm={() => {
          const finalReason = reversalReason.trim() || 'بدون سبب';
          if (purchaseToReverse) {
            reversePurchaseMutation.mutate({
              purchaseId: purchaseToReverse,
              reason: finalReason,
            });
          }
          setReversalReason('');
          setPurchaseToReverse(null);
        }}
        title="إلغاء عملية الشراء"
        message="أدخل سبب إلغاء عملية الشراء في الحقل التالي ثم اضغط تأكيد."
        confirmText="تأكيد الإلغاء"
        isLoading={reversePurchaseMutation.isPending}
        variant="danger"
      >
        <Input
          autoFocus
          value={reversalReason}
          onChange={(event) => setReversalReason(event.target.value)}
          placeholder="سبب إلغاء عملية الشراء"
        />
      </ConfirmDialog>
      <Modal
        isOpen={purchaseToPay !== null}
        onClose={() => { setPurchaseToPay(null); setPaymentAmount(''); }}
        title="تسجيل دفعة للشراء"
        size="sm"
      >
        {purchaseToPay && (
          <div className="space-y-4">
            <div className="rounded-lg bg-surface-muted p-3 text-sm">
              <p className="font-medium">{purchaseToPay.invoice_number}</p>
              <p className="text-text-muted">الإجمالي: ₪{Number(purchaseToPay.total_amount || 0).toLocaleString('en-US')}</p>
              <p className="text-text-muted">المتبقي: ₪{Number(purchaseToPay.remaining || 0).toLocaleString('en-US')}</p>
            </div>
            <Input
              type="number"
              min="0.01"
              max={Number(purchaseToPay.remaining || 0)}
              step="0.01"
              value={paymentAmount}
              onChange={(event) => setPaymentAmount(event.target.value)}
              placeholder="مبلغ الدفعة الجزئية أو الكاملة"
            />
            <div className="flex justify-end gap-3">
              <Button variant="secondary" onClick={() => { setPurchaseToPay(null); setPaymentAmount(''); }}>إلغاء</Button>
              <Button
                onClick={() => paymentMutation.mutate({ id: purchaseToPay.id, amount: Number(paymentAmount) })}
                disabled={!Number(paymentAmount) || Number(paymentAmount) <= 0 || Number(paymentAmount) > Number(purchaseToPay.remaining || 0) || paymentMutation.isPending}
              >
                {paymentMutation.isPending ? 'جاري التسجيل...' : 'تسجيل الدفعة'}
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}