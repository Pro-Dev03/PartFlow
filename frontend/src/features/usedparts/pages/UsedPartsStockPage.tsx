import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { inventoryApi, inspectionsApi } from '../../../services/api/endpoints';
import { Card, CardContent } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Badge } from '../../../components/ui/badge';
import { PageHeader } from '../../../components/ui/page-header';
import { ArrowRight, Clock, ShoppingCart } from 'lucide-react';

export function UsedPartsStockPage() {
  const navigate = useNavigate();
  const { data, isLoading } = useQuery({
    queryKey: ['inventory', 'used-stock'],
    queryFn: () => inventoryApi.list({ page: 1, per_page: 100 }),
  });
  const { data: passedInspectionsData, isLoading: isLoadingInspections } = useQuery({
    queryKey: ['inspections', 'passed', 'used-stock'],
    queryFn: () => inspectionsApi.list({ page: 1, per_page: 100, status: 'passed' }),
  });

  const passedInspections = Array.isArray(passedInspectionsData?.data)
    ? passedInspectionsData.data
    : passedInspectionsData?.data?.items || [];
  const passedItemIds = new Set(
    passedInspections
      .map((inspection: any) => inspection.inventory_item_id)
      .filter(Boolean)
  );

  const items = (Array.isArray(data?.data) ? data.data : data?.data?.items || [])
    .filter((item: any) => String(item.condition || '').toUpperCase() === 'USED')
    .filter((item: any) => passedItemIds.has(item.id))
    .filter((item: any) => String(item.status || '').toUpperCase() === 'AVAILABLE');

  return (
    <div>
      <PageHeader
        eyebrow="Used Parts Stock"
        title="مخزون القطع المستعملة"
        description="القطع المستعملة التي اجتازت الفحص والمتاحة للبيع"
        actions={
          <div className="flex items-center gap-2">
            <Button variant="secondary" onClick={() => navigate('/app/usedparts')}>
              <ArrowRight className="w-4 h-4" />
              رجوع
            </Button>
            <Button variant="secondary" onClick={() => navigate('/app/usedparts')}>شراء قطعة مستعملة</Button>
          </div>
        }
      />
      {isLoading || isLoadingInspections ? (
        <div className="flex justify-center p-12">جاري التحميل...</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {items.map((item: any) => (
            <Card key={item.id}>
              <CardContent className="p-4 space-y-3">
                <div className="flex justify-between">
                  <h3 className="font-semibold">{item.product_name || item.product?.name || 'قطعة مستعملة'}</h3>
                  <Badge variant="success">متاح</Badge>
                </div>
                <p className="text-sm text-gray-400">سعر الشراء: ₪{Number(item.purchase_cost || 0).toFixed(2)}</p>
                <p className="text-sm text-cyan">سعر البيع: ₪{Number(item.selling_price || 0).toFixed(2)}</p>
                <div className="flex gap-2">
                  <Button size="sm" variant="secondary" className="flex-1" onClick={() => navigate(`/app/item-history/${item.id}`)}>
                    <Clock className="w-3 h-3" /> السجل
                  </Button>
                  <Button size="sm" variant="primary" className="flex-1" onClick={() => navigate('/app/sales', { state: { usedPart: { id: item.id, name: item.product_name || 'قطعة مستعملة', price: item.selling_price, stock: 1, isTradeIn: true } } })}>
                    <ShoppingCart className="w-3 h-3" /> بيع
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
