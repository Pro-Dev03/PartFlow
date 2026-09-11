import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { returnsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { toast } from 'sonner';
import { 
  ArrowRight,
  RotateCcw,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Package,
  DollarSign,
  RefreshCw,
  History,
  Calculator
} from 'lucide-react';

interface ReturnItem {
  id: string;
  product_name: string;
  quantity_returned: number;
  original_quantity?: number;
  unit_price: number;
  total_refund_amount: number;
  original_condition: string;
  returned_condition: string;
  condition_notes?: string;
  resolution?: string;
  inventory_status: string;
  inspection_date?: string;
  inspection_result?: string;
  inspection_notes?: string;
  original_cost?: number;
  repair_cost: number;
  serial_number?: string;
  barcode?: string;
}

interface Return {
  id: string;
  return_number: string;
  reference_number: string;
  sale_id?: string;
  sale_invoice?: string;
  purchase_id?: string;
  customer_id?: string;
  customer_name?: string;
  return_date: string;
  return_type: string;
  status: string;
  total_refund_amount: number;
  refund_method: string;
  refund_date?: string;
  refund_reference?: string;
  debt_id?: string;
  debt_adjustment: number;
  reason: string;
  reason_detail?: string;
  item_condition_after_return: string;
  is_warranty_claim: boolean;
  warranty_id?: string;
  warranty_valid_until?: string;
  created_by?: string;
  processed_by?: string;
  approved_by?: string;
  approved_at?: string;
  notes?: string;
  internal_notes?: string;
  items?: ReturnItem[];
  created_at: string;
  updated_at: string;
}

export function ReturnDetailsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: returnData, isLoading } = useQuery({
    queryKey: ['return', id],
    queryFn: () => returnsApi.get(id || ''),
    enabled: !!id,
  });

  const returnItem = returnData as Return;

  const completeReturnMutation = useMutation({
    mutationFn: (returnId: string) => returnsApi.complete(returnId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['return', id] });
      queryClient.invalidateQueries({ queryKey: ['returns'] });
      toast.success('تم إكمال المرتجع بنجاح!');
    },
    onError: (error) => {
      console.error('Failed to complete return:', error);
      toast.error('فشل إكمال المرتجع');
    },
  });

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

  const getResolutionLabel = (resolution?: string) => {
    const labels: Record<string, string> = {
      RESTOCK: 'إعادة للمخزون',
      REPAIR: 'إصلاح',
      SUPPLIER_RETURN: 'إرجاع للمورد',
      WRITE_OFF: 'شطب',
      PARTS: 'تفكيك لقطع غيار',
      REPLACEMENT: 'استبدال',
    };
    return resolution ? labels[resolution] || resolution : 'قيد الانتظار';
  };

  const getInventoryStatusLabel = (status: string) => {
    const labels: Record<string, string> = {
      RETURNED: 'مرتجع',
      INSPECTION: 'قيد الفحص',
      REPAIRING: 'قيد الإصلاح',
      RESTOCKED: 'معاد للمخزون',
      SUPPLIER_RETURNED: 'مرجع للمورد',
      WRITTEN_OFF: 'مشطوب',
      DISMANTLED: 'مفكك',
    };
    return labels[status] || status;
  };

  const getReasonLabel = (reason: string) => {
    const labels: Record<string, string> = {
      DEFECTIVE: 'معطل',
      WRONG_ITEM: 'منتج خاطئ',
      COMPATIBILITY_ISSUE: 'مشكلة توافق',
      CUSTOMER_CHANGED_MIND: 'تغيير رأي العميل',
      DAMAGED: 'تالف',
      WARRANTY: 'ضمان',
      INCORRECT_SPECIFICATION: 'مواصفات خاطئة',
      OTHER: 'أخرى',
    };
    return labels[reason] || reason;
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (!returnItem) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-400">المرتجع غير موجود</p>
      </div>
    );
  }

  const statusBadge = getStatusBadge(returnItem.status);
  const StatusIcon = statusBadge.icon;

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Return Details"
        title="تفاصيل المرتجع"
        description="معلومات كاملة عن المرتجع والمنتجات المرتجعة"
        actions={
          <Button variant="secondary" onClick={() => navigate(-1)}>
            <ArrowRight className="w-4 h-4 mr-1" />
            رجوع
          </Button>
        }
      />

      {/* Main Return Information */}
      <Card className="mb-4">
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <RotateCcw className="w-6 h-6" />
              <span>{returnItem.return_number}</span>
            </div>
            <Badge variant={statusBadge.variant} className="gap-1">
              <StatusIcon className="w-3 h-3" />
              {statusBadge.label}
            </Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <p className="text-sm text-gray-400">رقم المرجع</p>
              <p className="font-semibold">{returnItem.reference_number}</p>
            </div>
            <div>
              <p className="text-sm text-gray-400">التاريخ</p>
              <p className="font-semibold">{new Date(returnItem.return_date).toLocaleDateString('ar-SA')}</p>
            </div>
            <div>
              <p className="text-sm text-gray-400">نوع المرتجع</p>
              <p className="font-semibold">{getReturnTypeLabel(returnItem.return_type)}</p>
            </div>
            <div>
              <p className="text-sm text-gray-400">العميل</p>
              <p className="font-semibold">{returnItem.customer_name || '-'}</p>
            </div>
            <div>
              <p className="text-sm text-gray-400">رقم الفاتورة</p>
              <p className="font-semibold">{returnItem.sale_invoice || '-'}</p>
            </div>
            <div>
              <p className="text-sm text-gray-400">طريقة الاسترجاع</p>
              <p className="font-semibold">{getRefundMethodLabel(returnItem.refund_method)}</p>
            </div>
          </div>

          <div className="mt-4 pt-4 border-t">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-gray-400">سبب المرتجع</p>
                <p className="font-semibold">{getReasonLabel(returnItem.reason)}</p>
                {returnItem.reason_detail && (
                  <p className="text-sm text-gray-500 mt-1">{returnItem.reason_detail}</p>
                )}
              </div>
              <div>
                <p className="text-sm text-gray-400">حالة القطعة بعد المرتجع</p>
                <p className="font-semibold">{getConditionLabel(returnItem.item_condition_after_return)}</p>
              </div>
            </div>
          </div>

          {returnItem.is_warranty_claim && (
            <div className="mt-4 p-3 bg-blue-50 rounded-lg border border-blue-200">
              <div className="flex items-center gap-2 text-blue-700">
                <ClipboardCheck className="w-4 h-4" />
                <span className="font-semibold">مطالبة ضمان</span>
              </div>
              {returnItem.warranty_valid_until && (
                <p className="text-sm text-blue-600 mt-1">
                  صالح حتى: {new Date(returnItem.warranty_valid_until).toLocaleDateString('ar-SA')}
                </p>
              )}
            </div>
          )}

          <div className="mt-4 p-4 bg-green-50 rounded-lg border border-green-200">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <DollarSign className="w-5 h-5 text-green-600" />
                <span className="text-sm text-gray-600">قيمة الاسترجاع</span>
              </div>
              <span className="text-2xl font-bold text-green-600">
                ₪{returnItem.total_refund_amount.toLocaleString()}
              </span>
            </div>
          </div>

          {returnItem.debt_adjustment !== 0 && (
            <div className="mt-4 p-4 bg-orange-50 rounded-lg border border-orange-200">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Calculator className="w-5 h-5 text-orange-600" />
                  <span className="text-sm text-gray-600">تعديل الدين</span>
                </div>
                <span className="text-xl font-bold text-orange-600">
                  ₪{returnItem.debt_adjustment.toLocaleString()}
                </span>
              </div>
            </div>
          )}


          {returnItem.notes && (
            <div className="mt-4">
              <p className="text-sm text-gray-400 mb-1">ملاحظات</p>
              <p className="text-sm">{returnItem.notes}</p>
            </div>
          )}

          {returnItem.internal_notes && (
            <div className="mt-4 p-3 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-400 mb-1">ملاحظات داخلية</p>
              <p className="text-sm text-gray-600">{returnItem.internal_notes}</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Returned Items */}
      <Card className="mb-4">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Package className="w-5 h-5" />
            <span>المنتجات المرتجعة</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {returnItem.items && returnItem.items.length > 0 ? (
            <div className="space-y-4">
              {returnItem.items.map((item) => (
                <div key={item.id} className="p-4 border rounded-lg hover:shadow-md transition-shadow">
                  <div className="flex justify-between items-start mb-3">
                    <div className="flex-1">
                      <h4 className="font-semibold text-lg">{item.product_name}</h4>
                      {item.serial_number && (
                        <p className="text-sm text-gray-400">SN: {item.serial_number}</p>
                      )}
                      {item.barcode && (
                        <p className="text-sm text-gray-400">Barcode: {item.barcode}</p>
                      )}
                    </div>
                    <div className="text-left">
                      <p className="text-lg font-bold text-green">
                        ₪{item.total_refund_amount.toLocaleString()}
                      </p>
                      <p className="text-sm text-gray-400">
                        {item.quantity_returned} × ₪{item.unit_price.toLocaleString()}
                      </p>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
                    <div>
                      <p className="text-gray-400">الكمية المرتجعة</p>
                      <p className="font-medium">{item.quantity_returned}</p>
                    </div>
                    <div>
                      <p className="text-gray-400">الكمية الأصلية</p>
                      <p className="font-medium">{item.original_quantity || '-'}</p>
                    </div>
                    <div>
                      <p className="text-gray-400">الحالة الأصلية</p>
                      <p className="font-medium">{item.original_condition || '-'}</p>
                    </div>
                    <div>
                      <p className="text-gray-400">الحالة المرتجعة</p>
                      <p className="font-medium">{item.returned_condition}</p>
                    </div>
                  </div>

                  <div className="mt-3 grid grid-cols-2 md:grid-cols-3 gap-3 text-sm">
                    <div>
                      <p className="text-gray-400">الحالة الحالية</p>
                      <Badge variant="secondary">{getInventoryStatusLabel(item.inventory_status)}</Badge>
                    </div>
                    <div>
                      <p className="text-gray-400">القرار</p>
                      <Badge variant={item.resolution ? 'success' : 'warning'}>
                        {getResolutionLabel(item.resolution)}
                      </Badge>
                    </div>
                    <div>
                      <p className="text-gray-400">تكلفة الإصلاح</p>
                      <p className="font-medium">₪{item.repair_cost.toLocaleString()}</p>
                    </div>
                  </div>

                  {item.condition_notes && (
                    <div className="mt-2 text-sm">
                      <p className="text-gray-400">ملاحظات الحالة</p>
                      <p>{item.condition_notes}</p>
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <p className="text-gray-400 text-center py-8">لا توجد منتجات مرتجعة</p>
          )}
        </CardContent>
      </Card>

      {/* Timeline */}
      <Card className="mb-4">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <History className="w-5 h-5" />
            <span>سجل العمليات</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex gap-3">
              <div className="flex flex-col items-center">
                <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
                <div className="w-0.5 h-full bg-gray-200"></div>
              </div>
              <div className="flex-1 pb-4">
                <p className="font-medium">إنشاء المرتجع</p>
                <p className="text-sm text-gray-400">{new Date(returnItem.created_at).toLocaleString('ar-SA')}</p>
              </div>
            </div>

            {returnItem.approved_at && (
              <div className="flex gap-3">
                <div className="flex flex-col items-center">
                  <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                  <div className="w-0.5 h-full bg-gray-200"></div>
                </div>
                <div className="flex-1 pb-4">
                  <p className="font-medium">موافقة على المرتجع</p>
                  <p className="text-sm text-gray-400">{new Date(returnItem.approved_at).toLocaleString('ar-SA')}</p>
                </div>
              </div>
            )}

            {returnItem.refund_date && (
              <div className="flex gap-3">
                <div className="flex flex-col items-center">
                  <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                  <div className="w-0.5 h-full bg-gray-200"></div>
                </div>
                <div className="flex-1 pb-4">
                  <p className="font-medium">تنفيذ الاسترجاع</p>
                  <p className="text-sm text-gray-400">{new Date(returnItem.refund_date).toLocaleString('ar-SA')}</p>
                </div>
              </div>
            )}

            <div className="flex gap-3">
              <div className="flex flex-col items-center">
                <div className="w-3 h-3 bg-gray-300 rounded-full"></div>
              </div>
              <div className="flex-1">
                <p className="font-medium">آخر تحديث</p>
                <p className="text-sm text-gray-400">{new Date(returnItem.updated_at).toLocaleString('ar-SA')}</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Actions */}
      <Card>
        <CardContent className="p-4">
          <div className="flex gap-3 justify-end">
            <Button
              variant="secondary"
              onClick={() => navigate(-1)}
            >
              رجوع
            </Button>
            {returnItem.status === 'APPROVED' && (
              <Button
                variant="success"
                onClick={() => completeReturnMutation.mutate(returnItem.id)}
                disabled={completeReturnMutation.isPending}
              >
                <CheckCircle className="w-4 h-4 mr-1" />
                {completeReturnMutation.isPending ? 'جاري الإكمال...' : 'إكمال المرتجع'}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

    </div>
  );
}