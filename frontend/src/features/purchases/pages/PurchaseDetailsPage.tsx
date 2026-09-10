import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowRight } from 'lucide-react';
import { purchasesApi, suppliersApi } from '../../../services/api/endpoints';
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
  const supplierId = purchase?.supplier_id || purchase?.supplier?.id;
  const { data: ledgerData } = useQuery({
    queryKey: ['supplier-ledger', supplierId],
    queryFn: () => suppliersApi.ledger(supplierId),
    enabled: Boolean(supplierId),
  });
  const ledger = ledgerData?.data || ledgerData;
  const supplierReturnCredits = Number(ledger?.supplier_return_credits || 0);
  const supplierPayments = Number(ledger?.supplier_payments || 0);
  const supplierNetBalance = Number(ledger?.current_balance || 0);

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
            <div className="rounded border border-border bg-surface-muted p-4">
              <h2 className="font-semibold mb-3">ملخص حساب المورد</h2>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div><span className="text-text-muted">المستحق الأصلي لهذه الفاتورة</span><p className="font-semibold">₪{(Number(purchase.total_amount || 0) - Number(purchase.paid_amount || 0)).toLocaleString('en-US')}</p></div>
                <div><span className="text-text-muted">رصيد مرتجعات المورد</span><p className="font-semibold">-₪{supplierReturnCredits.toLocaleString('en-US')}</p></div>
                <div><span className="text-text-muted">دفعات المورد</span><p className="font-semibold">₪{supplierPayments.toLocaleString('en-US')}</p></div>
                <div><span className="text-text-muted">صافي رصيد المورد</span><p className="font-semibold">₪{supplierNetBalance.toLocaleString('en-US')}</p></div>
              </div>
              <p className="mt-3 text-xs text-text-muted">Purchase تاريخية بالقيمة الإجمالية؛ صافي رصيد المورد يحسب المرتجعات والدفعات على مستوى الحساب.</p>
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
