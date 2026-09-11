import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { useNavigate } from 'react-router-dom';
import { returnsApi } from '../../../services/api/endpoints';
import { Card, CardContent } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { SearchInput } from '../../../components/ui/search-input';
import { PageHeader } from '../../../components/ui/page-header';
import { Select } from '../../../components/ui/select';
import { Badge } from '../../../components/ui/badge';
import { Modal } from '../../../components/ui/modal';
import { ConfirmDialog } from '../../../components/ui/confirm-dialog';
import { 
  RotateCcw, 
  Plus, 
  Eye,
  AlertTriangle,
  CheckCircle,
  DollarSign,
  TrendingDown,
  Filter,
  XCircle,
  RefreshCw,
  Truck,
  Edit,
  Trash2
} from 'lucide-react';
import { toast } from 'sonner';

interface ReturnItem {
  id: string;
  product_name: string;
  quantity_returned: number;
  unit_price: number;
  total_refund_amount: number;
  returned_condition: string;
  resolution: string;
}

interface Return {
  id: string;
  return_number: string;
  reference_number: string;
  sale_id?: string;
  customer_id?: string;
  customer_name?: string;
  return_date: string;
  return_type: string;
  status: string;
  total_refund_amount: number;
  refund_method: string;
  reason: string;
  item_condition_after_return: string;
  items?: ReturnItem[];
  created_at: string;
  updated_at: string;
}

export function ReturnsPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [returnTypeFilter, setReturnTypeFilter] = useState('');
  const [refundMethodFilter, setRefundMethodFilter] = useState('');
  const [editingReturn, setEditingReturn] = useState<Return | null>(null);
  const [returnToDelete, setReturnToDelete] = useState<Return | null>(null);
  const [editReason, setEditReason] = useState('');
  const [editRefundMethod, setEditRefundMethod] = useState('CASH');
  const [editCondition, setEditCondition] = useState('READY_FOR_SALE');


  const { data: returnsData, isLoading } = useQuery({
    queryKey: ['returns', statusFilter, returnTypeFilter, refundMethodFilter, searchQuery],
    queryFn: () => returnsApi.list({ 
      page: 1, 
      per_page: 100,
      status: statusFilter,
      return_type: returnTypeFilter,
      refund_method: refundMethodFilter,
      search: searchQuery
    }),
  });

  const returns = (returnsData?.data as Return[]) || [];

  const { data: salesReturnsAnalysis } = useQuery({
    queryKey: ['sales-returns-analysis'],
    queryFn: () => returnsApi.getSalesReturnsAnalysis(),
  });

  const { data: statistics } = useQuery({
    queryKey: ['returns-statistics'],
    queryFn: () => returnsApi.getStatistics(),
  });

  const handleClearSearch = () => {
    setSearchQuery('');
    setStatusFilter('');
    setReturnTypeFilter('');
    setRefundMethodFilter('');
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, { label: string; variant: 'success' | 'warning' | 'danger' | 'secondary' | 'info'; icon: any }> = {
      PENDING: { label: 'قيد الانتظار', variant: 'warning', icon: AlertTriangle },
      APPROVED: { label: 'موافق عليه', variant: 'success', icon: CheckCircle },
      PROCESSING: { label: 'قيد المعالجة', variant: 'info', icon: RefreshCw },
      COMPLETED: { label: 'مكتمل', variant: 'success', icon: CheckCircle },
      REJECTED: { label: 'مرفوض', variant: 'danger', icon: XCircle },
      CANCELLED: { label: 'ملغي', variant: 'secondary', icon: XCircle },
    };
    return variants[status] || { label: status, variant: 'secondary', icon: AlertTriangle };
  };

  const getReturnTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      FULL: 'مرتجع كامل',
      PARTIAL: 'مرتجع جزئي',
      QUANTITY_PARTIAL: 'مرتجع كمية جزئية',
    };
    return labels[type] || type;
  };

  const getRefundMethodLabel = (method: string) => {
    const labels: Record<string, string> = {
      CASH: 'نقدي',
      DEBT_ADJUSTMENT: 'تعديل دين',
      EXCHANGE: 'استبدال',
      BANK_TRANSFER: 'تحويل بنكي',
    };
    return labels[method] || method;
  };

  const getReasonLabel = (reason: string) => {
    const labels: Record<string, string> = {
      DEFECTIVE: 'منتج معطل',
      WRONG_ITEM: 'منتج خاطئ',
      COMPATIBILITY_ISSUE: 'مشكلة توافق',
      CUSTOMER_CHANGED_MIND: 'تغيير رأي العميل',
      DAMAGED: 'منتج تالف',
      WARRANTY: 'ضمان',
      INCORRECT_SPECIFICATION: 'مواصفات غير صحيحة',
      OTHER: 'أخرى',
    };
    return labels[reason] || reason;
  };

  const getConditionLabel = (condition: string) => {
    const labels: Record<string, string> = {
      READY_FOR_SALE: 'جاهز للبيع',
      NOT_FOR_SALE: 'غير قابل للبيع',
      RETURN_TO_SUPPLIER: 'إرجاع للمورد',
      SELLABLE: 'قابل للبيع',
      NEEDS_REPAIR: 'يحتاج إصلاح',
      NEEDS_INSPECTION: 'قيد المراجعة',
      DAMAGED: 'تالف',
      USED: 'مستعمل',
      REFURBISHED: 'مجدّد',
      SUPPLIER_RETURN: 'إرجاع للمورد',
      WRITE_OFF: 'شطب',
      PARTS: 'قطع غيار',
    };
    return labels[condition] || condition;
  };

  const handleViewDetails = (returnItem: Return) => {
    navigate(`/app/returns/${returnItem.id}`);
  };

  const completeReturnMutation = useMutation({
    mutationFn: (id: string) => returnsApi.complete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['returns'] });
      toast.success('تم إكمال المرتجع بنجاح!');
    },
    onError: (error) => {
      console.error('Failed to complete return:', error);
      toast.error('فشل إكمال المرتجع');
    },
  });

  const updateReturnMutation = useMutation({
    mutationFn: () => returnsApi.update(editingReturn!.id, {
      reason: editReason,
      refund_method: editRefundMethod,
      item_condition_after_return: editCondition,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['returns'] });
      setEditingReturn(null);
      toast.success('تم تعديل المرتجع بنجاح');
    },
    onError: () => toast.error('تعذر تعديل المرتجع في حالته الحالية'),
  });

  const deleteReturnMutation = useMutation({
    mutationFn: () => returnsApi.delete(returnToDelete!.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['returns'] });
      setReturnToDelete(null);
      toast.success('تم حذف المرتجع بنجاح');
    },
    onError: () => toast.error('لا يمكن حذف مرتجع تمت معالجته أو اعتماده'),
  });

  const canModifyReturn = (status: string) => ['PENDING', 'REJECTED'].includes(status);

  const openEditReturn = (returnItem: Return) => {
    setEditingReturn(returnItem);
    setEditReason(returnItem.reason);
    setEditRefundMethod(returnItem.refund_method);
    setEditCondition(returnItem.item_condition_after_return);
  };

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Returns Management"
        title={t('returns.title') || 'المرتجعات'}
        description="إدارة المرتجعات والاسترجاع مع تتبع كامل للمنتجات والماليات"
        actions={
          <div className="flex gap-2">
            <Button
              variant="primary"
              className="gap-2"
              onClick={() => navigate('/app/returns/create')}
            >
              <Plus className="w-4 h-4" />
              مرتجع جديد
            </Button>
            <Button variant="secondary" className="gap-2" onClick={() => navigate('/app/supplier-returns')}>
              <Truck className="w-4 h-4" />
              مرتجع للمورد
            </Button>
          </div>
        }
      />

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-cyan/10">
                <RotateCcw className="w-5 h-5 text-cyan" />
              </div>
              <div>
                <p className="text-sm text-gray-400">إجمالي المرتجعات</p>
                <p className="text-2xl font-bold">{statistics?.total_returns || returns.length}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-yellow/10">
                <AlertTriangle className="w-5 h-5 text-yellow" />
              </div>
              <div>
                <p className="text-sm text-gray-400">قيد الانتظار</p>
                <p className="text-2xl font-bold">
                  {statistics?.pending_returns || returns.filter((r) => r.status === 'PENDING').length}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-green/10">
                <DollarSign className="w-5 h-5 text-green" />
              </div>
              <div>
                <p className="text-sm text-gray-400">قيمة المرتجعات</p>
                <p className="text-2xl font-bold">
                  ₪{(statistics?.total_refunded ?? returns.filter((r) => r.status === 'COMPLETED').reduce((sum, r) => sum + r.total_refund_amount, 0)).toLocaleString()}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-blue/10">
                <TrendingDown className="w-5 h-5 text-blue" />
              </div>
              <div>
                <p className="text-sm text-gray-400">صافي المبيعات</p>
                <p className="text-2xl font-bold">
                  ₪{(salesReturnsAnalysis?.data?.[0]?.net_sales || 0).toLocaleString()}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Search and Filters */}
      <Card className="mb-4">
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-3 items-start md:items-center">
            <div className="flex-1 w-full">
              <SearchInput
                placeholder="بحث برقم المرتجع أو العميل..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onClear={handleClearSearch}
                size="sm"
                className="w-full"
              />
            </div>
            <div className="flex gap-2 w-full md:w-auto">
              <Select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                size="sm"
                options={[
                  { value: '', label: 'كل الحالات' },
                  { value: 'PENDING', label: 'قيد الانتظار' },
                  { value: 'APPROVED', label: 'موافق عليه' },
                  { value: 'PROCESSING', label: 'قيد المعالجة' },
                  { value: 'COMPLETED', label: 'مكتمل' },
                  { value: 'REJECTED', label: 'مرفوض' },
                ]}
                className="w-40"
              />
              <Select
                value={returnTypeFilter}
                onChange={(e) => setReturnTypeFilter(e.target.value)}
                size="sm"
                options={[
                  { value: '', label: 'كل الأنواع' },
                  { value: 'FULL', label: 'مرتجع كامل' },
                  { value: 'PARTIAL', label: 'مرتجع جزئي' },
                  { value: 'QUANTITY_PARTIAL', label: 'كمية جزئية' },
                ]}
                className="w-40"
              />
              <Button
                variant="secondary"
                size="sm"
                onClick={handleClearSearch}
                className="h-10 px-4"
              >
                <Filter className="w-4 h-4 mr-1" />
                مسح
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Returns Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : returns.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center">
            <RotateCcw className="w-12 h-12 mx-auto mb-4 text-gray-400" />
            <p className="text-gray-400">لا توجد مرتجعات</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {returns.map((returnItem) => {
            const statusBadge = getStatusBadge(returnItem.status);
            const StatusIcon = statusBadge.icon;
            return (
              <Card key={returnItem.id} className="hover:shadow-md transition-shadow">
                <CardContent className="p-4">
                  <div className="flex justify-between items-start mb-3">
                    <div>
                      <h3 className="font-semibold text-lg mb-1">{returnItem.return_number}</h3>
                      <p className="text-xs text-gray-400">{returnItem.reference_number}</p>
                    </div>
                    <Badge variant={statusBadge.variant} className="gap-1">
                      <StatusIcon className="w-3 h-3" />
                      {statusBadge.label}
                    </Badge>
                  </div>

                  <div className="space-y-2 mb-3">
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">العميل:</span>
                      <span className="font-medium">{returnItem.customer_name || '-'}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">التاريخ:</span>
                      <span>{new Date(returnItem.return_date).toLocaleDateString('ar-SA')}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">النوع:</span>
                      <span>{getReturnTypeLabel(returnItem.return_type)}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">طريقة الاسترجاع:</span>
                      <span>{getRefundMethodLabel(returnItem.refund_method)}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">السبب:</span>
                      <span>{getReasonLabel(returnItem.reason)}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">حالة القطعة:</span>
                      <span>{getConditionLabel(returnItem.item_condition_after_return)}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">قيمة الاسترجاع:</span>
                      <span className="font-bold text-green">₪{returnItem.total_refund_amount.toLocaleString()}</span>
                    </div>
                  </div>

                  <div className="flex gap-2">
                    <Button
                      variant="secondary"
                      size="sm"
                      className="flex-1"
                      onClick={() => handleViewDetails(returnItem)}
                    >
                      <Eye className="w-4 h-4 mr-1" />
                      التفاصيل
                    </Button>
                    {returnItem.status === 'APPROVED' && (
                      <Button
                        variant="success"
                        size="sm"
                        onClick={() => completeReturnMutation.mutate(returnItem.id)}
                        disabled={completeReturnMutation.isPending}
                      >
                        <CheckCircle className="w-4 h-4 mr-1" />
                        إكمال
                      </Button>
                    )}
                    {canModifyReturn(returnItem.status) && (
                      <>
                        <Button variant="outline" size="sm" tableAction onClick={() => openEditReturn(returnItem)} aria-label="تعديل المرتجع" title="تعديل المرتجع">
                          <Edit className="w-4 h-4" />
                        </Button>
                        <Button variant="danger" size="sm" tableAction onClick={() => setReturnToDelete(returnItem)} aria-label="حذف المرتجع" title="حذف المرتجع">
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </>
                    )}
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      <Modal isOpen={Boolean(editingReturn)} onClose={() => setEditingReturn(null)} title="تعديل المرتجع" size="md">
        <form className="space-y-4" onSubmit={(event) => { event.preventDefault(); updateReturnMutation.mutate(); }}>
          <div><label className="mb-2 block text-sm font-medium text-text-secondary">سبب المرتجع</label><Select value={editReason} onChange={(event) => setEditReason(event.target.value)} options={[{ value: 'DEFECTIVE', label: 'منتج معطل' }, { value: 'WRONG_ITEM', label: 'منتج خاطئ' }, { value: 'CUSTOMER_CHANGED_MIND', label: 'تغيير رأي العميل' }, { value: 'DAMAGED', label: 'تالف' }, { value: 'OTHER', label: 'أخرى' }]} /></div>
          <div><label className="mb-2 block text-sm font-medium text-text-secondary">حالة المنتج</label><Select value={editCondition} onChange={(event) => setEditCondition(event.target.value)} options={[{ value: 'READY_FOR_SALE', label: 'جاهز للبيع' }, { value: 'NOT_FOR_SALE', label: 'غير قابل للبيع' }, { value: 'RETURN_TO_SUPPLIER', label: 'إرجاع للمورد' }, { value: 'NEEDS_REPAIR', label: 'يحتاج إصلاح' }]} /></div>
          <div><label className="mb-2 block text-sm font-medium text-text-secondary">طريقة رد المبلغ</label><Select value={editRefundMethod} onChange={(event) => setEditRefundMethod(event.target.value)} options={[{ value: 'CASH', label: 'نقدي' }, { value: 'DEBT_ADJUSTMENT', label: 'تعديل الدين' }]} /></div>
          <div className="flex justify-end gap-3"><Button type="button" variant="secondary" onClick={() => setEditingReturn(null)}>إلغاء</Button><Button type="submit" variant="primary" disabled={updateReturnMutation.isPending || !editReason}>{updateReturnMutation.isPending ? 'جاري الحفظ...' : 'حفظ التعديل'}</Button></div>
        </form>
      </Modal>

      <ConfirmDialog
        isOpen={Boolean(returnToDelete)}
        onClose={() => setReturnToDelete(null)}
        onConfirm={() => deleteReturnMutation.mutate()}
        title="حذف المرتجع"
        message={`هل تريد حذف المرتجع «${returnToDelete?.return_number || ''}»؟ لا يمكن التراجع عن هذا الإجراء.`}
        confirmText="حذف المرتجع"
        isLoading={deleteReturnMutation.isPending}
      />
    </div>
  );
}