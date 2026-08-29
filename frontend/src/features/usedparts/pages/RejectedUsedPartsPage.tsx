import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { inventoryApi, inspectionsApi } from '../../../services/api/endpoints';
import { Card, CardContent } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Badge } from '../../../components/ui/badge';
import { PageHeader } from '../../../components/ui/page-header';
import { ArrowRight, Package, XCircle } from 'lucide-react';

export function RejectedUsedPartsPage() {
  const navigate = useNavigate();
  const { data, isLoading } = useQuery({
    queryKey: ['inventory', 'used-rejected'],
    queryFn: () => inventoryApi.list({ page: 1, per_page: 100 }),
  });
  const { data: inspectionsData } = useQuery({
    queryKey: ['inspections', 'used-rejected'],
    queryFn: () => inspectionsApi.list({ page: 1, per_page: 100 }),
  });

  const inventoryItems = Array.isArray(data?.data) ? data.data : data?.data?.items || [];
  const inspections = Array.isArray(inspectionsData?.data)
    ? inspectionsData.data
    : inspectionsData?.data?.items || [];
  const rejectedParts = inventoryItems.filter((item: any) =>
    String(item.condition || '').toUpperCase() === 'USED' &&
    String(item.status || '').toUpperCase() === 'DAMAGED'
  );

  return (
    <div>
      <PageHeader
        eyebrow="Rejected Used Parts"
        title="القطع المرفوضة"
        description="القطع المستعملة التي لم تُقبل بعد الفحص وأسباب الرفض"
        actions={
          <Button variant="secondary" onClick={() => navigate('/app/usedparts')}>
            <ArrowRight className="w-4 h-4" />
            رجوع
          </Button>
        }
      />
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      ) : rejectedParts.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center">
            <Package className="w-12 h-12 mx-auto mb-4 text-gray-400" />
            <p className="text-gray-400">لا توجد قطع مرفوضة أو تالفة</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4">
          {rejectedParts.map((item: any) => {
            const inspection = inspections.find((entry: any) => entry.inventory_item_id === item.id);
            return (
              <Card key={item.id} className="border border-red-500/30 bg-red-500/5">
                <CardContent className="p-4">
                  <div className="flex items-start justify-between gap-2 mb-3">
                    <h2 className="font-semibold text-[var(--text-primary)]">
                      {item.product_name || item.product?.name || 'قطعة بدون اسم'}
                    </h2>
                    <Badge variant="danger"><XCircle className="w-3 h-3 ml-1" />مرفوضة</Badge>
                  </div>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between gap-2">
                      <span className="text-[var(--text-muted)]">الرقم التسلسلي</span>
                      <span>{item.serial_number || 'غير مسجل'}</span>
                    </div>
                    <div className="flex justify-between gap-2">
                      <span className="text-[var(--text-muted)]">تكلفة الشراء</span>
                      <span>₪{Number(item.purchase_cost ?? item.unit_cost ?? 0).toFixed(2)}</span>
                    </div>
                    <div className="rounded-lg bg-red-500/10 p-3">
                      <p className="text-xs font-semibold text-red-500 mb-1">سبب عدم القبول</p>
                      <p className="text-[var(--text-secondary)]">
                        {inspection?.notes || item.notes || 'لم يتم تسجيل سبب الرفض في بيانات الفحص'}
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
