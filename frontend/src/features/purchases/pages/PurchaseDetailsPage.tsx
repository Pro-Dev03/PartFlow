import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowRight } from 'lucide-react';
import { purchasesApi } from '../../../services/api/endpoints';
import { PageHeader } from '../../../components/ui/page-header';
import { Button } from '../../../components/ui/button';
import { Card, CardContent } from '../../../components/ui/card';
import { Badge } from '../../../components/ui/badge';

export function PurchaseDetailsPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const { data, isLoading, isError } = useQuery({
    queryKey: ['purchase', id],
    queryFn: () => purchasesApi.get(id || ''),
    enabled: Boolean(id),
  });

  const purchase = data?.data?.purchase || data?.purchase;
  const items = data?.data?.items || data?.items || [];

  return (
    <div>
      <PageHeader
        eyebrow="Purchase Details"
        title="تفاصيل عملية الشراء"
        description={purchase?.invoice_number || 'معلومات العملية والقطع المرتبطة بها'}
        actions={
          <Button variant="secondary" onClick={() => navigate('/app/purchases')}>
            <ArrowRight className="w-4 h-4" />
            رجوع
          </Button>
        }
      />
      {isLoading && <div className="p-12 text-center">جاري التحميل...</div>}
      {isError && <div className="p-12 text-center text-red-500">تعذر تحميل تفاصيل الشراء</div>}
      {purchase && (
        <Card>
          <CardContent className="p-6 space-y-6">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div><span className="text-sm text-text-muted">المورد</span><p>{purchase.supplier?.name || purchase.supplier_name || '-'}</p></div>
              <div><span className="text-sm text-text-muted">الحالة</span><p><Badge>{purchase.status}</Badge></p></div>
              <div><span className="text-sm text-text-muted">الإجمالي</span><p>₪{Number(purchase.total_amount || 0).toLocaleString('en-US')}</p></div>
              <div><span className="text-sm text-text-muted">التاريخ</span><p>{purchase.purchase_date ? new Date(purchase.purchase_date).toLocaleDateString('en-US') : '-'}</p></div>
            </div>
            <div>
              <h2 className="font-semibold mb-3">القطع</h2>
              <div className="space-y-2">
                {items.map((item: any) => (
                  <div key={item.id} className="flex justify-between border-b border-border py-2">
                    <span>{item.product_name || item.product?.name || 'قطعة'}</span>
                    <span>{item.quantity} × ₪{Number(item.unit_cost || 0).toLocaleString('en-US')}</span>
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
