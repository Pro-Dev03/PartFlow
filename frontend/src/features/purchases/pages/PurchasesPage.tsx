import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { getButtonSize } from '../../../config/button-sizes';
import {
  Plus,
  Eye,
  Download,
  Printer,
  Check,
  Edit,
  Trash2,
  RotateCcw,
  DollarSign,
} from 'lucide-react';

// Custom hooks
import { usePurchases } from '../hooks/usePurchases';
import { purchasesApi } from '../../../services/api/endpoints';

// Components
import { PurchaseStats } from '../components/PurchaseStats';
import { PurchaseFilters } from '../components/PurchaseFilters';

// Types
import { Purchase } from '../types/purchases.types';

// SmartDelete utility (ARCHITECTURE-PRINCIPLES.md)
import { handleSmartDelete } from '../../../utils/smartDelete';
import { ConfirmDialog } from '../../../components/ui/confirm-dialog';
import { Modal } from '../../../components/ui/modal';
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
  const handleReversePurchase = async (purchaseId: string) => {
    try {
      const response = await purchasesApi.getUsedItemsInfo(purchaseId);
      const usedItems = Array.isArray(response?.data) ? response.data : [];
      const soldItems = usedItems.filter((item: any) => Number(item.sold_quantity || 0) > 0);
      if (soldItems.length > 0) {
        toast.error('لا يمكن عكس العملية لأن القطعة قد تم بيعها بالفعل.');
        return;
      }
    } catch (error) {
      console.error('Failed to check purchase items before reversal:', error);
      toast.error('تعذر التحقق من حالة القطعة. حاول مرة أخرى.');
      return;
    }
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
      reversed: { label: 'تم العكس', variant: 'destructive' },
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

      {/* Purchases Table */}
      <Card>
        <CardHeader>
          <div className="flex flex-col gap-md md:flex-row md:items-center md:justify-between">
            <CardTitle>
              {viewFilter === 'active'
                ? 'المشتريات الحالية'
                : viewFilter === 'received'
                  ? 'المشتريات المستلمة'
                  : viewFilter === 'archived'
                    ? 'أرشيف المشتريات'
                    : 'كل المشتريات'}
            </CardTitle>
            <div className="flex flex-wrap gap-sm" role="tablist" aria-label="عرض المشتريات">
              <Button
                variant={viewFilter === 'active' ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => {
                  setViewFilter('active');
                  setStatusFilter('');
                }}
                role="tab"
                aria-selected={viewFilter === 'active'}
              >
                الحالية
              </Button>
              <Button
                variant={viewFilter === 'received' ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => {
                  setViewFilter('received');
                  setStatusFilter('');
                }}
                role="tab"
                aria-selected={viewFilter === 'received'}
              >
                تم الاستلام
              </Button>
              <Button
                variant={viewFilter === 'archived' ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => {
                  setViewFilter('archived');
                  setStatusFilter('');
                }}
                role="tab"
                aria-selected={viewFilter === 'archived'}
              >
                الأرشيف
              </Button>
              <Button
                variant={viewFilter === 'all' ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => {
                  setViewFilter('all');
                  setStatusFilter('');
                }}
                role="tab"
                aria-selected={viewFilter === 'all'}
              >
                الكل
              </Button>
            </div>
          </div>
          <p className="mt-2 text-sm text-text-muted">
            {filteredPurchases.length} عملية شراء مطابقة
          </p>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan" />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>رقم الطلب</TableHead>
                  <TableHead>المورد</TableHead>
                  <TableHead>القطع</TableHead>
                  <TableHead>الضريبة</TableHead>
                  <TableHead>إجمالي التكلفة</TableHead>
                  <TableHead>المدفوع</TableHead>
                  <TableHead>المتبقي</TableHead>
                  <TableHead>التاريخ المتوقع</TableHead>
                  <TableHead>الحالة</TableHead>
                  <TableHead className="text-start">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredPurchases.map((purchase: any) => {
                  const normalizedStatus = purchase.status === 'completed' ? 'received' : purchase.status;
                  const statusBadge = getStatusBadge(purchase.status);
                  return (
                    <TableRow key={purchase.id}>
                      <TableCell className="font-medium">{purchase.invoice_number}</TableCell>
                      <TableCell>{purchase.supplier?.name || purchase.supplier_name}</TableCell>
                      <TableCell>{purchase.total_items || purchase.items?.length || 0} قطع</TableCell>
                      <TableCell>
                        {Number(purchase.tax_amount || 0) > 0
                          ? `₪${Number(purchase.tax_amount).toLocaleString()}`
                          : <Badge variant="outline">بدون ضريبة</Badge>}
                      </TableCell>
                      <TableCell>₪{purchase.total_amount?.toLocaleString()}</TableCell>
                      <TableCell className="text-green">₪{purchase.paid_amount?.toLocaleString()}</TableCell>
                      <TableCell>₪{purchase.remaining?.toLocaleString()}</TableCell>
                      <TableCell>
                        {purchase.expected_delivery_date
                          ? new Date(purchase.expected_delivery_date).toLocaleDateString('en-US')
                          : '-'
                        }
                      </TableCell>
                      <TableCell>
                        <Badge variant={statusBadge.variant}>
                          {statusBadge.label}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-start">
                        <div className="flex gap-2">
                          {Number(purchase.remaining || 0) > 0 && !['cancelled', 'reversed'].includes(normalizedStatus) && (
                            <Button
                              variant="ghost"
                              size="sm"
                              title="تسجيل دفعة"
                              aria-label="تسجيل دفعة"
                              onClick={(event) => {
                                event.stopPropagation();
                                setPurchaseToPay(purchase);
                                setPaymentAmount('');
                              }}
                              className="text-cyan-600 hover:text-cyan-700"
                            >
                              <DollarSign className="w-4 h-4" />
                            </Button>
                          )}

                          {normalizedStatus === 'pending' && (
                            <>
                              <Button
                                variant="ghost"
                                size="sm"
                                title="استلام البضاعة"
                                onClick={() => handleReceivePurchase(purchase.id)}
                                className="text-green-600 hover:text-green-700"
                              >
                                <Check className="w-4 h-4" />
                              </Button>
                            </>
                          )}

                          {(normalizedStatus === 'pending' || normalizedStatus === 'draft') && (
                            <>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => navigate(`/app/purchases/edit/${purchase.id}`)}
                                className="text-blue-600 hover:text-blue-700"
                              >
                                <Edit className="w-4 h-4" />
                              </Button>
                              {Number(purchase.paid_amount || 0) <= 0 && (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  title="حذف الطلب غير المدفوع"
                                  onClick={() => handleDeletePurchase(purchase.id)}
                                  className="text-red-600 hover:text-red-700"
                                >
                                  <Trash2 className="w-4 h-4" />
                                </Button>
                              )}
                            </>
                          )}
                          {normalizedStatus === 'received' && (
                            <>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={(event) => {
                                  event.stopPropagation();
                                  void handleReversePurchase(purchase.id);
                                }}
                                className="text-orange-600 hover:text-orange-700"
                                title="عكس العملية"
                              >
                                <RotateCcw className="w-4 h-4" />
                              </Button>
                            </>
                          )}
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={(event) => {
                              event.stopPropagation();
                              setPurchaseToView(purchase.id);
                            }}
                            title="عرض التفاصيل"
                          >
                            <Eye className="w-4 h-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
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
                    <span>{item.product_name || item.product?.name || 'قطعة'}</span>
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
        title="عكس عملية الشراء"
        message="أدخل سبب العكس في الحقل التالي ثم اضغط تأكيد."
        confirmText="تأكيد العكس"
        isLoading={reversePurchaseMutation.isPending}
        variant="danger"
      >
        <Input
          autoFocus
          value={reversalReason}
          onChange={(event) => setReversalReason(event.target.value)}
          placeholder="سبب عكس العملية"
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