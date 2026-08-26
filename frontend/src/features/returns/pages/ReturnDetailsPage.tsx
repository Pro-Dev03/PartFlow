import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { returnsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { Modal } from '../../../components/ui/modal';
import { 
  ArrowRight,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Package,
  Calendar,
  DollarSign,
  User,
  RefreshCw,
  ClipboardList,
  Scissors,
  Wrench,
  Trash2,
  ArrowLeft
} from 'lucide-react';

interface ReturnItem {
  id: string;
  product_name: string;
  quantity_returned: number;
  unit_price: number;
  total_refund_amount: number;
  returned_condition: string;
  resolution: string;
  inspection_required: boolean;
  inspection_result?: string;
  inspection_notes?: string;
  serial_number?: string;
  barcode?: string;
  original_condition?: string;
  condition_notes?: string;
  repair_cost?: number;
  inventory_status: string;
}

interface Return {
  id: string;
  return_number: string;
  reference_number: string;
  sale_id?: string;
  sale_invoice?: string;
  customer_id?: string;
  customer_name?: string;
  return_date: string;
  return_type: string;
  status: string;
  total_refund_amount: number;
  refund_method: string;
  refund_date?: string;
  refund_reference?: string;
  debt_adjustment?: number;
  customer_credit?: number;
  reason: string;
  reason_detail?: string;
  item_condition_after_return: string;
  is_warranty_claim: boolean;
  warranty_valid_until?: string;
  notes?: string;
  internal_notes?: string;
  created_at: string;
  updated_at: string;
  processed_by?: string;
  approved_by?: string;
  approved_at?: string;
}

export function ReturnDetailsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [isInspectionModalOpen, setIsInspectionModalOpen] = useState(false);
  const [selectedItem, setSelectedItem] = useState<ReturnItem | null>(null);
  const [inspectionResult, setInspectionResult] = useState('');
  const [inspectionNotes, setInspectionNotes] = useState('');
  const [resolution, setResolution] = useState('');
  const [repairCost, setRepairCost] = useState('');

  const { data: returnData, isLoading } = useQuery({
    queryKey: ['return-with-items', id],
    queryFn: () => returnsApi.getWithItems(id || ''),
    enabled: !!id,
  });

  const returnRecord = returnData?.return as Return;
  const items = (returnData?.items as ReturnItem[]) || [];

  const completeReturnMutation = useMutation({
    mutationFn: (returnId: string) => returnsApi.complete(returnId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['return-with-items', id] });
      queryClient.invalidateQueries({ queryKey: ['returns'] });
      alert('تم إكمال المرتجع بنجاح!');
    },
    onError: (error) => {
      console.error('Failed to complete return:', error);
      alert('فشل إكمال المرتجع');
    },
  });

  const reverseReturnMutation = useMutation({
    mutationFn: (returnId: string) => returnsApi.reverse(returnId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['return-with-items', id] });
      queryClient.invalidateQueries({ queryKey: ['returns'] });
      alert('تم عكس المرتجع بنجاح!');
      navigate('/returns');
    },
    onError: (error) => {
      console.error('Failed to reverse return:', error);
      alert('فشل عكس المرتجع');
    },
  });

  const processInspectionMutation = useMutation({
    mutationFn: ({ itemId, data }: { itemId: string; data: any }) => 
      returnsApi.processInspection(itemId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['return-with-items', id] });
      queryClient.invalidateQueries({ queryKey: ['returns'] });
      alert('تم معالجة الفحص بنجاح!');
      setIsInspectionModalOpen(false);
      setSelectedItem(null);
    },
    onError: (error) => {
      console.error('Failed to process inspection:', error);
      alert('فشل معالجة الفحص');
    },
  });

  const handleStartInspection = (item: ReturnItem) => {
    setSelectedItem(item);
    setInspectionResult('');
    setInspectionNotes('');
    setResolution('');
    setRepairCost('');
    setIsInspectionModalOpen(true);
  };

  const handleCompleteInspection = () => {
    if (!selectedItem) return;

    const data = {
      inspection_date: new Date().toISOString().split('T')[0],
      inspection_result: inspectionResult,
      inspection_notes: inspectionNotes,
      resolution: resolution,
      repair_cost: parseFloat(repairCost) * 100,
    };

    processInspectionMutation.mutate({ itemId: selectedItem.id, data });
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
      CREDIT: 'رصيد عميل',
      DEBT_ADJUSTMENT: 'تعديل دين',
      EXCHANGE: 'استبدال',
      BANK_TRANSFER: 'تحويل بنكي',
      STORE_CREDIT: 'رصيد المتجر',
    };
    return labels[method] || method;
  };

  const getConditionLabel = (condition: string) => {
    const labels: Record<string, string> = {
      SELLABLE: 'قابل للبيع',
      NEEDS_INSPECTION: 'يحتاج فحص',
      NEEDS_REPAIR: 'يحتاج إصلاح',
      DAMAGED: 'تالف',
      USED: 'مستعمل',
      REFURBISHED: 'مجدّد',
      SUPPLIER_RETURN: 'إرجاع للمورد',
      WRITE_OFF: 'شطب',
      PARTS: 'قطع غيار',
    };
    return labels[condition] || condition;
  };

  const getInventoryStatusLabel = (status: string) => {
    const labels: Record<string, string> = {
      RETURNED: 'مرتجع',
      INSPECTION: 'قيد الفحص',
      REPAIRING: 'قيد الإصلاح',
      RESTOCKED: 'أعيد للمخزون',
      SUPPLIER_RETURNED: 'أُرجع للمورد',
      WRITTEN_OFF: 'مشطوب',
      DISMANTLED: 'مفكك',
    };
    return labels[status] || status;
  };

  const getResolutionLabel = (resolution: string) => {
    const labels: Record<string, string> = {
      RESTOCK: 'إعادة للمخزون',
      REPAIR: 'إصلاح',
      SUPPLIER_RETURN: 'إرجاع للمورد',
      WRITE_OFF: 'شطب',
      PARTS: 'قطع غيار',
      REPLACEMENT: 'استبدال',
    };
    return labels[resolution] || resolution;
  };

  const getReturnedConditionLabel = (condition: string) => {
    const labels: Record<string, string> = {
      NEW: 'جديد',
      USED: 'مستعمل',
      DAMAGED: 'تالف',
      DEFECTIVE: 'معطل',
      OPEN_BOX: 'صندوق مفتوح',
      REFURBISHED: 'مجدّد',
    };
    return labels[condition] || condition;
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (!returnRecord) {
    return (
      <Card>
        <CardContent className="p-12 text-center">
          <AlertTriangle className="w-12 h-12 mx-auto mb-4 text-gray-400" />
          <p className="text-gray-400">المرتجع غير موجود</p>
        </CardContent>
      </Card>
    );
  }

  const statusBadge = getStatusBadge(returnRecord.status);
  const StatusIcon = statusBadge.icon;

  return (
    <div>
      <PageHeader
        eyebrow="Return Details"
        title={`تفاصيل المرتجع ${returnRecord.return_number}`}
        description="عرض تفاصيل كاملة للمرتجع والمنتجات المرتجعة"
        actions={
          <div className="flex gap-2">
            <Button
              variant="secondary"
              onClick={() => navigate('/returns')}
            >
              <ArrowLeft className="w-4 h-4 mr-1" />
              رجوع
            </Button>
            {returnRecord.status === 'APPROVED' && (
              <Button
                variant="success"
                onClick={() => completeReturnMutation.mutate(returnRecord.id)}
                disabled={completeReturnMutation.isPending}
              >
                <CheckCircle className="w-4 h-4 mr-1" />
                إكمال المرتجع
              </Button>
            )}
            {returnRecord.status === 'COMPLETED' && (
              <Button
                variant="danger"
                onClick={() => reverseReturnMutation.mutate(returnRecord.id)}
                disabled={reverseReturnMutation.isPending}
              >
                <RefreshCw className="w-4 h-4 mr-1" />
                عكس المرتجع
              </Button>
            )}
          </div>
        }
      />

      {/* Return Information */}
      <Card className="mb-4">
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span>معلومات المرتجع</span>
            <Badge variant={statusBadge.variant} className="gap-1">
              <StatusIcon className="w-3 h-3" />
              {statusBadge.label}
            </Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div className="space-y-2">
              <p className="text-sm text-gray-400">رقم المرتجع</p>
              <p className="font-semibold">{returnRecord.return_number}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-gray-400">رقم المرجع</p>
              <p className="font-semibold">{returnRecord.reference_number}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-gray-400">التاريخ</p>
              <p className="font-semibold">{new Date(returnRecord.return_date).toLocaleDateString('ar-SA')}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-gray-400">العميل</p>
              <p className="font-semibold">{returnRecord.customer_name || '-'}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-gray-400">رقم الفاتورة</p>
              <p className="font-semibold">{returnRecord.sale_invoice || '-'}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-gray-400">نوع المرتجع</p>
              <p className="font-semibold">{getReturnTypeLabel(returnRecord.return_type)}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-gray-400">السبب</p>
              <p className="font-semibold">{returnRecord.reason}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-gray-400">طريقة الاسترجاع</p>
              <p className="font-semibold">{getRefundMethodLabel(returnRecord.refund_method)}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-gray-400">حالة القطعة</p>
              <p className="font-semibold">{getConditionLabel(returnRecord.item_condition_after_return)}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-gray-400">قيمة الاسترجاع</p>
              <p className="font-semibold text-green">₪{returnRecord.total_refund_amount.toLocaleString()}</p>
            </div>
            {returnRecord.debt_adjustment && returnRecord.debt_adjustment !== 0 && (
              <div className="space-y-2">
                <p className="text-sm text-gray-400">تعديل الدين</p>
                <p className="font-semibold">₪{returnRecord.debt_adjustment.toLocaleString()}</p>
              </div>
            )}
            {returnRecord.customer_credit && returnRecord.customer_credit !== 0 && (
              <div className="space-y-2">
                <p className="text-sm text-gray-400">رصيد العميل</p>
                <p className="font-semibold">₪{returnRecord.customer_credit.toLocaleString()}</p>
              </div>
            )}
          </div>
          {returnRecord.reason_detail && (
            <div className="mt-4 p-3 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-400 mb-1">تفاصيل السبب</p>
              <p className="text-sm">{returnRecord.reason_detail}</p>
            </div>
          )}
          {returnRecord.notes && (
            <div className="mt-4 p-3 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-400 mb-1">ملاحظات</p>
              <p className="text-sm">{returnRecord.notes}</p>
            </div>
          )}
          {returnRecord.internal_notes && (
            <div className="mt-4 p-3 bg-yellow-50 rounded-lg border border-yellow-200">
              <p className="text-sm text-yellow-700 mb-1">ملاحظات داخلية</p>
              <p className="text-sm text-yellow-800">{returnRecord.internal_notes}</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Return Items */}
      <Card>
        <CardHeader>
          <CardTitle>المنتجات المرتجعة</CardTitle>
        </CardHeader>
        <CardContent>
          {items.length === 0 ? (
            <div className="text-center py-8 text-gray-400">
              لا توجد منتجات مرتجعة
            </div>
          ) : (
            <div className="space-y-4">
              {items.map((item) => (
                <div key={item.id} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex justify-between items-start mb-3">
                    <div className="flex-1">
                      <h4 className="font-semibold text-lg mb-1">{item.product_name}</h4>
                      <div className="flex gap-4 text-sm text-gray-400">
                        {item.serial_number && (
                          <span>SN: {item.serial_number}</span>
                        )}
                        {item.barcode && (
                          <span>Barcode: {item.barcode}</span>
                        )}
                      </div>
                    </div>
                    <Badge variant="secondary">
                      {getInventoryStatusLabel(item.inventory_status)}
                    </Badge>
                  </div>

                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-3">
                    <div>
                      <p className="text-xs text-gray-400">الكمية المرتجعة</p>
                      <p className="font-semibold">{item.quantity_returned}</p>
                    </div>
                    <div>
                      <p className="text-xs text-gray-400">السعر</p>
                      <p className="font-semibold">₪{item.unit_price.toLocaleString()}</p>
                    </div>
                    <div>
                      <p className="text-xs text-gray-400">الإجمالي</p>
                      <p className="font-semibold text-green">₪{item.total_refund_amount.toLocaleString()}</p>
                    </div>
                    <div>
                      <p className="text-xs text-gray-400">الحالة عند الإرجاع</p>
                      <p className="font-semibold">{getReturnedConditionLabel(item.returned_condition)}</p>
                    </div>
                  </div>

                  {item.condition_notes && (
                    <div className="mb-3 p-2 bg-gray-50 rounded text-sm">
                      <p className="text-gray-400 mb-1">ملاحظات الحالة</p>
                      <p>{item.condition_notes}</p>
                    </div>
                  )}

                  <div className="flex justify-between items-center mb-3">
                    <div className="flex gap-2">
                      {item.inspection_result && (
                        <Badge variant={item.inspection_result === 'PASSED' ? 'success' : item.inspection_result === 'FAILED' ? 'danger' : 'warning'}>
                          {item.inspection_result === 'PASSED' ? 'اجتاز الفحص' : item.inspection_result === 'FAILED' ? 'فشل الفحص' : 'قيد الفحص'}
                        </Badge>
                      )}
                      {item.resolution && (
                        <Badge variant="info">
                          {getResolutionLabel(item.resolution)}
                        </Badge>
                      )}
                    </div>
                    {item.repair_cost && item.repair_cost > 0 && (
                      <p className="text-sm text-orange">
                        تكلفة الإصلاح: ₪{(item.repair_cost / 100).toLocaleString()}
                      </p>
                    )}
                  </div>

                  {item.inspection_notes && (
                    <div className="mb-3 p-2 bg-blue-50 rounded text-sm">
                      <p className="text-blue-700 mb-1">ملاحظات الفحص</p>
                      <p className="text-blue-800">{item.inspection_notes}</p>
                    </div>
                  )}

                  {item.inspection_required && !item.inspection_result && (
                    <Button
                      variant="primary"
                      size="sm"
                      onClick={() => handleStartInspection(item)}
                      className="w-full"
                    >
                      <ClipboardList className="w-4 h-4 mr-1" />
                      بدء الفحص
                    </Button>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Inspection Modal */}
      <Modal
        isOpen={isInspectionModalOpen}
        onClose={() => setIsInspectionModalOpen(false)}
        title="فحص المنتج المرتجع"
        variant="modern"
        size="lg"
      >
        <div className="space-y-4">
          {selectedItem && (
            <div className="p-4 bg-gray-50 rounded-lg">
              <p className="font-semibold">{selectedItem.product_name}</p>
              <p className="text-sm text-gray-400">الكمية: {selectedItem.quantity_returned}</p>
            </div>
          )}

          <div>
            <label className="block text-sm font-medium mb-2">نتيجة الفحص</label>
            <select
              value={inspectionResult}
              onChange={(e) => setInspectionResult(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg"
            >
              <option value="">اختر النتيجة</option>
              <option value="PASSED">اجتاز الفحص</option>
              <option value="FAILED">فشل الفحص</option>
              <option value="PENDING">قيد الفحص</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">القرار</label>
            <select
              value={resolution}
              onChange={(e) => setResolution(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg"
            >
              <option value="">اختر القرار</option>
              <option value="RESTOCK">إعادة للمخزون</option>
              <option value="REPAIR">إصلاح</option>
              <option value="SUPPLIER_RETURN">إرجاع للمورد</option>
              <option value="WRITE_OFF">شطب</option>
              <option value="PARTS">قطع غيار</option>
              <option value="REPLACEMENT">استبدال</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">تكلفة الإصلاح (اختياري)</label>
            <input
              type="number"
              value={repairCost}
              onChange={(e) => setRepairCost(e.target.value)}
              placeholder="أدخل التكلفة"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg"
              step="0.01"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">ملاحظات الفحص</label>
            <textarea
              value={inspectionNotes}
              onChange={(e) => setInspectionNotes(e.target.value)}
              placeholder="أدخل ملاحظات الفحص..."
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg"
            />
          </div>

          <div className="flex gap-2 justify-end pt-4 border-t">
            <Button
              variant="secondary"
              onClick={() => setIsInspectionModalOpen(false)}
            >
              إلغاء
            </Button>
            <Button
              variant="primary"
              onClick={handleCompleteInspection}
              disabled={!inspectionResult || !resolution || processInspectionMutation.isPending}
            >
              {processInspectionMutation.isPending ? 'جاري المعالجة...' : 'حفظ النتيجة'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}