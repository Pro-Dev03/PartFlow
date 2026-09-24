import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowRight } from 'lucide-react';
import { purchasesApi, suppliersApi } from '../../../services/api/endpoints';
import { PageHeader } from '../../../design-system/components/page-header';
import { Button } from '../../../design-system/components/button';
import { Card, CardContent } from '../../../design-system/components/card';
import { Badge } from '../../../design-system/components/badge';
import { SupplierInvoiceModal } from '../components/SupplierInvoiceModal';
import { useState } from 'react';
import { formatStoreDate } from '../../../utils/store-time';

export function PurchaseDetailsPage() {
  const [supplierInvoiceOpen, setSupplierInvoiceOpen] = useState(false);
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
        title="تفاصيل عملية الشراء"
        description={purchase?.invoice_number || 'معلومات العملية والقطع المرتبطة بها'}
        actions={
          <div className="flex gap-2">
            <Button variant="primary" onClick={() => setSupplierInvoiceOpen(true)} disabled={!purchase}>عرض فاتورة التاجر</Button>
            <Button variant="secondary" onClick={() => navigate('/app/purchases')}>
              <ArrowRight className="w-4 h-4" />
              رجوع
            </Button>
          </div>
        }
      />
      {isLoading && <div className="p-12 text-center">جاري التحميل...</div>}
      {isError && <div className="p-12 text-center text-red-500">تعذر تحميل تفاصيل الشراء</div>}
      {purchase && (
        <Card>
          <CardContent className="p-6 space-y-6">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div><span className="text-sm text-text-muted">التاجر</span><p>{purchase.supplier?.name || purchase.supplier_name || '-'}</p></div>
              <div><span className="text-sm text-text-muted">الحالة</span><p><Badge>{purchase.status}</Badge></p></div>
              <div><span className="text-sm text-text-muted">الإجمالي</span><p>₪{Number(purchase.total_amount || 0).toLocaleString('en-US')}</p></div>
              <div><span className="text-sm text-text-muted">تاريخ فاتورة المورد</span><p>{purchase.purchase_date ? formatStoreDate(purchase.purchase_date, 'en-US') : '-'}</p></div>
              <div><span className="text-sm text-text-muted">رقم فاتورة المورد</span><p>{purchase.invoice_number || '-'}</p></div>
              <div><span className="text-sm text-text-muted">المدفوع</span><p>₪{Number(purchase.paid_amount || 0).toLocaleString('en-US')}</p></div>
              <div><span className="text-sm text-text-muted">المتبقي</span><p>₪{Math.max(0, Number(purchase.total_amount || 0) - Number(purchase.paid_amount || 0)).toLocaleString('en-US')}</p></div>
            </div>
            <div className="rounded border border-border bg-surface-muted p-4">
              <h2 className="font-semibold mb-3">ملخص حساب التاجر</h2>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div><span className="text-text-muted">المستحق الأصلي لهذه الفاتورة</span><p className="font-semibold">₪{(Number(purchase.total_amount || 0) - Number(purchase.paid_amount || 0)).toLocaleString('en-US')}</p></div>
                <div><span className="text-text-muted">رصيد مرتجعات التاجر</span><p className="font-semibold">-₪{supplierReturnCredits.toLocaleString('en-US')}</p></div>
                <div><span className="text-text-muted">دفعات التاجر</span><p className="font-semibold">₪{supplierPayments.toLocaleString('en-US')}</p></div>
                <div><span className="text-text-muted">صافي رصيد التاجر</span><p className="font-semibold">₪{supplierNetBalance.toLocaleString('en-US')}</p></div>
              </div>
              <p className="mt-3 text-xs text-text-muted">Purchase تاريخية بالقيمة الإجمالية؛ صافي رصيد المورد يحسب المرتجعات والدفعات على مستوى الحساب.</p>
            </div>
            <div>
              <h2 className="font-semibold mb-3">تفاصيل العناصر</h2>
              <div className="overflow-x-auto rounded border border-border">
                <div className="min-w-[640px]">
                  <div className="grid grid-cols-4 gap-4 border-b border-border bg-surface-muted px-4 py-3 text-sm font-semibold text-text-muted">
                    <span>المنتج</span>
                    <span>الكمية في الفاتورة</span>
                    <span>سعر الوحدة</span>
                    <span>إجمالي العنصر</span>
                  </div>
                  {items.map((item: any) => {
                    const quantity = Number(item.quantity || 0);
                    const unitCost = Number(item.unit_cost || 0);
                    const itemTotal = quantity * unitCost;

                    return (
                      <div key={item.id} className="grid grid-cols-4 gap-4 border-b border-border px-4 py-3 last:border-b-0">
                        <span>{item.product_name || item.product?.name || 'قطعة'}</span>
                        <span>{quantity.toLocaleString('en-US')}</span>
                        <span>₪{unitCost.toLocaleString('en-US')}</span>
                        <span className="font-semibold">₪{itemTotal.toLocaleString('en-US')}</span>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
      <SupplierInvoiceModal
        isOpen={supplierInvoiceOpen}
        onClose={() => setSupplierInvoiceOpen(false)}
        purchase={purchase}
        supplier={purchase?.supplier}
        items={items}
      />
    </div>
  );
}
