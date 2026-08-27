import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
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
    receivePurchaseMutation,
    deletePurchaseMutation,
    reversePurchaseMutation,
  } = usePurchases();
  const [purchaseToReceive, setPurchaseToReceive] = useState<string | null>(null);
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
      'التكلفة': purchase.total_amount
    }));
    exportToCSV(dataToExport, `purchases-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    const dataToPrint = purchases.map((purchase: Purchase) => ({
      'التاريخ': purchase.purchase_date,
      'المورد': purchase.supplier?.name || purchase.supplier_name,
      'الحالة': getStatusBadge(purchase.status).label,
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
          <CardTitle>قائمة المشتريات</CardTitle>
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
                  const statusBadge = getStatusBadge(purchase.status);
                  return (
                    <TableRow key={purchase.id}>
                      <TableCell className="font-medium">{purchase.invoice_number}</TableCell>
                      <TableCell>{purchase.supplier?.name || purchase.supplier_name}</TableCell>
                      <TableCell>{purchase.total_items || purchase.items?.length || 0} قطع</TableCell>
                      <TableCell>₪{purchase.total_amount?.toLocaleString()}</TableCell>
                      <TableCell className="text-green">₪{purchase.paid_amount?.toLocaleString()}</TableCell>
                      <TableCell>₪{purchase.remaining?.toLocaleString()}</TableCell>
                      <TableCell>
                        {purchase.expected_delivery_date
                          ? new Date(purchase.expected_delivery_date).toLocaleDateString('ar-SA')
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
                          {purchase.status === 'pending' && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleReceivePurchase(purchase.id)}
                              className="text-green-600 hover:text-green-700"
                            >
                              <Check className="w-4 h-4" />
                            </Button>
                          )}

                          {(purchase.status === 'pending' || purchase.status === 'draft') && (
                            <>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => navigate(`/app/purchases/edit/${purchase.id}`)}
                                className="text-blue-600 hover:text-blue-700"
                              >
                                <Edit className="w-4 h-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleDeletePurchase(purchase.id)}
                                className="text-red-600 hover:text-red-700"
                              >
                                <Trash2 className="w-4 h-4" />
                              </Button>
                            </>
                          )}
                          {purchase.status === 'received' && (
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
              <div><span className="text-sm text-text-muted">المورد</span><p>{purchaseDetails.supplier?.name || purchaseDetails.supplier_name || '-'}</p></div>
              <div><span className="text-sm text-text-muted">الحالة</span><p><Badge>{purchaseDetails.status}</Badge></p></div>
              <div><span className="text-sm text-text-muted">التاريخ</span><p>{purchaseDetails.purchase_date ? new Date(purchaseDetails.purchase_date).toLocaleDateString('en-US') : '-'}</p></div>
              <div><span className="text-sm text-text-muted">الإجمالي</span><p>₪{Number(purchaseDetails.total_amount || 0).toLocaleString('en-US')}</p></div>
              <div><span className="text-sm text-text-muted">المدفوع</span><p>₪{Number(purchaseDetails.paid_amount || 0).toLocaleString('en-US')}</p></div>
              <div><span className="text-sm text-text-muted">المتبقي</span><p>₪{Number(purchaseDetails.remaining || 0).toLocaleString('en-US')}</p></div>
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
    </div>
  );
}