import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { acquisitionsApi, inventoryApi, partTypesApi } from '../../../services/api/endpoints';
import { Card, CardContent } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { Badge } from '../../../design-system/components/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../design-system/components/table';
import { PageHeader } from '../../../design-system/components/page-header';
import { StatCard } from '../../../design-system/components/stat-card';
import { Select } from '../../../design-system/components/select';
import { ArrowRight, Package, ShoppingCart, TrendingUp, LayoutGrid, List } from 'lucide-react';
import { formatPrice } from '../../../utils';
import { getPartTypeImage } from '../../../services/localPartTypeImages';
import { PaginationControls } from '../../../design-system/components/pagination-controls';

export function UsedPartsStockPage() {
  const navigate = useNavigate();
  const [selectedPartType, setSelectedPartType] = useState('');
  const [layoutMode, setLayoutMode] = useState<'cards' | 'table'>('cards');
  const [page, setPage] = useState(1);
  const pageSize = 10;
  const { data, isLoading } = useQuery({
    queryKey: ['inventory', 'used-stock', page, pageSize],
    queryFn: () => inventoryApi.list({ page, per_page: pageSize, condition: 'USED', status: 'AVAILABLE' }),
  });

  useEffect(() => {
    setPage(1);
  }, [selectedPartType]);
  const { data: acquisitionsData, isLoading: isLoadingAcquisitions } = useQuery({
    queryKey: ['acquisitions', 'used-stock'],
    queryFn: () => acquisitionsApi.list({ page: 1, per_page: 100, type: 'CUSTOMER' }),
  });
  const { data: partTypesData, isLoading: isLoadingPartTypes } = useQuery({
    queryKey: ['part-types'],
    queryFn: () => partTypesApi.list(),
  });

  const partTypes = Array.isArray(partTypesData?.data) ? partTypesData.data : [];

  const acquisitionItems = (Array.isArray(acquisitionsData?.data)
    ? acquisitionsData.data
    : acquisitionsData?.data?.items || []
  ).flatMap((acquisition: any) => acquisition.items || []);

  const items = (Array.isArray(data?.data) ? data.data : data?.data?.items || [])
    .filter((item: any) => String(item.condition || '').toUpperCase() === 'USED')
    .filter((item: any) => String(item.status || '').toUpperCase() === 'AVAILABLE')
    .filter((item: any) => !selectedPartType || item.part_type_id === selectedPartType);
  const usedPurchaseValue = items.reduce((total: number, item: any) => total + Number(item.purchase_cost || 0), 0);
  const usedSellingValue = items.reduce((total: number, item: any) => total + Number(item.selling_price || 0), 0);

  const getStatusLabel = (status: string) => {
    switch (String(status || '').toUpperCase()) {
      case 'AVAILABLE': return 'متاح للبيع';
      case 'SOLD': return 'مباع';
      case 'DAMAGED': return 'تالف';
      case 'RETURNED': return 'مرتجع';
      default: return status || 'غير محدد';
    }
  };

  return (
    <div>
      <PageHeader
        eyebrow="Used Parts Stock"
        title="مخزون القطع المستعملة"
        description="القطع المستعملة المتاحة وغير المباعة"
        actions={
          <div className="flex items-center gap-2">
            <Button variant="secondary" onClick={() => navigate('/app/usedparts')}>
              <ArrowRight className="w-4 h-4" />
              رجوع
            </Button>
            <Button variant="secondary" onClick={() => navigate('/app/usedparts')}>شراء قطعة مستعملة</Button>
            <Select
              value={selectedPartType}
              onChange={(e) => setSelectedPartType(e.target.value)}
              loading={isLoadingPartTypes}
              options={[
                { value: '', label: 'جميع أنواع القطع' },
                ...partTypes.map((partType: any) => ({ value: partType.id, label: partType.name_ar })),
              ]}
              emptyMessage="لا توجد أنواع قطع"
              size="sm"
              className="min-w-48"
            />
            <Button variant={layoutMode === 'cards' ? 'primary' : 'secondary'} onClick={() => setLayoutMode('cards')} aria-label="عرض البطاقات" title="عرض البطاقات">
              <LayoutGrid className="w-4 h-4" />
              بطاقات
            </Button>
            <Button variant={layoutMode === 'table' ? 'primary' : 'secondary'} onClick={() => setLayoutMode('table')} aria-label="عرض الجدول" title="عرض الجدول">
              <List className="w-4 h-4" />
              جدول
            </Button>
          </div>
        }
      />
      {isLoading || isLoadingAcquisitions ? (
        <div className="flex justify-center p-12">جاري التحميل...</div>
      ) : (
        <div className="space-y-5">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <StatCard title="القطع المستعملة المتاحة" value={items.length} icon={Package} subtitle="قطع مستعملة غير مباعة" variant="featured" />
            <StatCard title="قيمة الشراء" value={formatPrice(usedPurchaseValue)} icon={Package} subtitle="تكلفة القطع المستعملة فقط" />
            <StatCard title="قيمة البيع" value={formatPrice(usedSellingValue)} icon={TrendingUp} subtitle="قيمة القطع المستعملة المتاحة" variant="success" />
          </div>
          {layoutMode === 'table' ? (
            <div className="overflow-x-auto rounded-xl border border-border bg-surface">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>الصورة</TableHead>
                    <TableHead>القطعة</TableHead>
                    <TableHead>النوع</TableHead>
                    <TableHead>حالة المخزون</TableHead>
                    <TableHead>سعر الشراء</TableHead>
                    <TableHead>سعر البيع</TableHead>
                    <TableHead>الإجراء</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((item: any) => (
                    <TableRow key={item.id}>
                      <TableCell>
                        {getPartTypeImage(String(item.part_type_id)) ? (
                          <div style={{ width: '64px', height: '48px', overflow: 'hidden', borderRadius: '6px' }}>
                            <img src={getPartTypeImage(String(item.part_type_id))} alt={item.product_name || 'نوع القطعة'} style={{ width: '64px', height: '48px', maxWidth: '64px', maxHeight: '48px', objectFit: 'cover', display: 'block' }} />
                          </div>
                        ) : <Package className="h-8 w-8 text-text-tertiary" />}
                      </TableCell>
                      <TableCell className="font-medium">{item.product_name || item.product?.name || 'قطعة مستعملة'}</TableCell>
                      <TableCell>{partTypes.find((partType: any) => partType.id === item.part_type_id)?.name_ar || '-'}</TableCell>
                      <TableCell><Badge variant="success">{getStatusLabel(item.status)}</Badge></TableCell>
                      <TableCell>{formatPrice(Number(item.purchase_cost || 0))}</TableCell>
                      <TableCell className="font-semibold text-cyan">{formatPrice(Number(item.selling_price || 0))}</TableCell>
                      <TableCell>
                        <Button size="sm" variant="primary" onClick={() => navigate('/app/sales', { state: { usedPart: { id: item.product_id || item.product?.id, inventoryItemId: item.id, name: item.product_name || 'قطعة مستعملة', price: item.selling_price, stock: 1, isTradeIn: true } } })}>
                          <ShoppingCart className="w-3 h-3" /> بيع
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {items.map((item: any) => (
              <Card key={item.id}>
                <CardContent className="p-4 space-y-3">
                  <div className="flex justify-between items-start gap-3">
                    {getPartTypeImage(String(item.part_type_id)) && (
                      <img
                        src={getPartTypeImage(String(item.part_type_id))}
                        alt={item.product_name || 'نوع القطعة'}
                        style={{ width: '64px', height: '64px', objectFit: 'cover', display: 'block', borderRadius: '10px', flex: '0 0 64px' }}
                      />
                    )}
                    <h3 className="font-semibold">{item.product_name || item.product?.name || 'قطعة مستعملة'}</h3>
                    <Badge variant={String(item.status || '').toUpperCase() === 'AVAILABLE' ? 'success' : 'secondary'}>
                      {getStatusLabel(item.status)}
                    </Badge>
                  </div>
                  <p className="text-sm text-gray-400">سعر الشراء: {formatPrice(Number(item.purchase_cost || 0))}</p>
                  <p className="text-sm text-cyan">سعر البيع: {formatPrice(Number(item.selling_price || 0))}</p>
                  <div className="flex gap-2">
                    <Button size="sm" variant="primary" className="flex-1" disabled={String(item.status || '').toUpperCase() !== 'AVAILABLE'} onClick={() => navigate('/app/sales', { state: { usedPart: { id: item.product_id || item.product?.id, inventoryItemId: item.id, name: item.product_name || 'قطعة مستعملة', price: item.selling_price, stock: 1, isTradeIn: true } } })}>
                      <ShoppingCart className="w-3 h-3" /> بيع
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
          )}
          {items.length > 0 && (
            <Card>
              <PaginationControls
                page={page}
                pageSize={pageSize}
                total={Number(data?.meta?.total || items.length)}
                onPageChange={setPage}
                isLoading={isLoading}
              />
            </Card>
          )}
        </div>
      )}
    </div>
  );
}
