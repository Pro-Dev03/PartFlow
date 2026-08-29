import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { acquisitionsApi, inventoryApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import {
  Package,
  ShoppingCart,
  CheckCircle,
  AlertTriangle,
  ArrowRight,
  DollarSign,
  Barcode,
  TrendingUp,
  Clock,
  Settings,
  Wrench,
  Trash2
} from 'lucide-react';

interface HistoryEvent {
  id: string;
  event_type: string;
  event_date: string;
  description: string;
  details: any;
  amount?: number;
  status?: string;
}

export function ItemHistoryPage() {
  const { itemId } = useParams<{ itemId: string }>();
  const navigate = useNavigate();
  const [selectedTab, setSelectedTab] = useState<'timeline' | 'financials' | 'details'>('timeline');

  const { data: historyData, isLoading } = useQuery({
    queryKey: ['item-history', itemId],
    queryFn: () => acquisitionsApi.getItemHistory(itemId || ''),
    enabled: !!itemId,
  });
  const { data: inventoryData, isLoading: isLoadingInventory } = useQuery({
    queryKey: ['used-parts-history-index'],
    queryFn: () => inventoryApi.list({ page: 1, per_page: 100 }),
    enabled: !itemId,
  });

  const history = (historyData as any)?.data ?? historyData;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (!itemId) {
    const soldUsedItems = (Array.isArray(inventoryData?.data) ? inventoryData.data : inventoryData?.data?.items || [])
      .filter((item: any) =>
        String(item.condition || '').toUpperCase() === 'USED' &&
        String(item.status || '').toUpperCase() === 'SOLD'
      );

    return (
      <div>
        <PageHeader
          eyebrow="Used Parts History"
          title="تاريخ القطع المباعة"
          description="القطع المستعملة التي تم بيعها مع سجل كل قطعة"
          actions={<Button variant="secondary" onClick={() => navigate('/app/usedparts')}>رجوع</Button>}
        />
        {isLoadingInventory ? (
          <div className="flex items-center justify-center h-64">جاري التحميل...</div>
        ) : soldUsedItems.length === 0 ? (
          <Card>
            <CardContent className="p-12 text-center">
              <Clock className="w-12 h-12 mx-auto mb-4 text-gray-400" />
              <p className="text-gray-400">لا توجد قطع مستعملة مباعة</p>
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {soldUsedItems.map((item: any) => (
              <Card key={item.id}>
                <CardContent className="p-4 space-y-3">
                  <div className="flex items-start justify-between gap-2">
                    <h3 className="font-semibold text-[var(--text-primary)]">
                      {item.product_name || item.product?.name || 'قطعة مستعملة'}
                    </h3>
                    <Badge variant="secondary">مباعة</Badge>
                  </div>
                  <p className="text-sm text-[var(--text-secondary)]">
                    الرقم التسلسلي: {item.serial_number || 'غير متوفر'}
                  </p>
                  <Button className="w-full" variant="secondary" onClick={() => navigate(`/app/usedparts/item-history/${item.id}`)}>
                    عرض السجل الكامل
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    );
  }

  if (!history) {
    return (
      <Card>
        <CardContent className="p-12 text-center">
          <AlertTriangle className="w-12 h-12 mx-auto mb-4 text-gray-400" />
          <p className="text-gray-400">اختر قطعة لعرض سجلها الكامل</p>
          <p className="text-sm text-gray-500 mt-2">يمكنك فتح تاريخ أي قطعة من صفحة مخزون القطع المستعملة.</p>
        </CardContent>
      </Card>
    );
  }

  const getEventIcon = (eventType: string) => {
    switch (eventType) {
      case 'ACQUISITION':
        return <ShoppingCart className="w-5 h-5 text-cyan" />;
      case 'INSPECTION':
        return <CheckCircle className="w-5 h-5 text-green" />;
      case 'REPAIR':
        return <Wrench className="w-5 h-5 text-orange" />;
      case 'SALE':
        return <DollarSign className="w-5 h-5 text-green" />;
      case 'RETURN':
        return <ArrowRight className="w-5 h-5 text-red" />;
      case 'WRITE_OFF':
        return <Trash2 className="w-5 h-5 text-red" />;
      default:
        return <Clock className="w-5 h-5 text-gray-400" />;
    }
  };

  const getEventBadge = (status?: string) => {
    if (!status) return null;
    switch (status) {
      case 'PASSED':
        return <Badge variant="success">اجتاز</Badge>;
      case 'FAILED':
        return <Badge variant="danger">فشل</Badge>;
      case 'COMPLETED':
        return <Badge variant="success">مكتمل</Badge>;
      case 'PENDING':
        return <Badge variant="warning">قيد الانتظار</Badge>;
      default:
        return <Badge variant="secondary">{status}</Badge>;
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ar-EG', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const events: HistoryEvent[] = history.events || [];

  return (
    <div>
      <PageHeader
        eyebrow="Item History"
        title="تاريخ القطعة"
        description="سجل كامل للقطعة منذ دخولها المتجر"
        actions={
          <Button variant="secondary" onClick={() => navigate('/app/usedparts')}>
            رجوع
          </Button>
        }
      />

      {/* Item Overview */}
      <Card className="mb-4 border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
        <CardContent className="p-6">
          <div className="flex items-start gap-4">
            <div className="p-3 rounded-lg bg-cyan/10">
              <Package className="w-6 h-6 text-cyan" />
            </div>
            <div className="flex-1">
              <h2 className="text-xl font-bold text-[var(--text-primary)] mb-2">
                {history.item?.product_name || history.item?.name}
              </h2>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div>
                  <p className="text-[var(--text-secondary)]">الرقم التسلسلي</p>
                  <p className="text-[var(--text-primary)] font-medium">
                    {history.item?.serial_number || 'غير متوفر'}
                  </p>
                </div>
                <div>
                  <p className="text-[var(--text-secondary)]">الباركود</p>
                  <p className="text-[var(--text-primary)] font-medium flex items-center gap-1">
                    <Barcode className="w-4 h-4" />
                    {history.item?.barcode || 'غير متوفر'}
                  </p>
                </div>
                <div>
                  <p className="text-[var(--text-secondary)]">الحالة</p>
                  <p className="text-[var(--text-primary)] font-medium">
                    {history.item?.condition || 'غير محدد'}
                  </p>
                </div>
                <div>
                  <p className="text-[var(--text-secondary)]">التقييم</p>
                  <p className="text-[var(--text-primary)] font-medium">
                    {history.item?.grade || 'غير محدد'}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Financial Summary */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-red/10">
                <DollarSign className="w-5 h-5 text-red" />
              </div>
              <div>
                <p className="text-sm text-gray-400">سعر الشراء</p>
                <p className="text-2xl font-bold">
                  ₪{(history.item?.cost || 0).toFixed(2)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-green/10">
                <TrendingUp className="w-5 h-5 text-green" />
              </div>
              <div>
                <p className="text-sm text-gray-400">سعر البيع</p>
                <p className="text-2xl font-bold">
                  ₪{(history.item?.selling_price || 0).toFixed(2)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-purple/10">
                <DollarSign className="w-5 h-5 text-purple" />
              </div>
              <div>
                <p className="text-sm text-gray-400">الربح</p>
                <p className="text-2xl font-bold">
                  ₪{((history.item?.selling_price || 0) - (history.item?.cost || 0)).toFixed(2)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 mb-4">
        <Button
          variant={selectedTab === 'timeline' ? 'primary' : 'secondary'}
          onClick={() => setSelectedTab('timeline')}
        >
          <Clock className="w-4 h-4 mr-1" />
          السجل الزمني
        </Button>
        <Button
          variant={selectedTab === 'financials' ? 'primary' : 'secondary'}
          onClick={() => setSelectedTab('financials')}
        >
          <DollarSign className="w-4 h-4 mr-1" />
          المالية
        </Button>
        <Button
          variant={selectedTab === 'details' ? 'primary' : 'secondary'}
          onClick={() => setSelectedTab('details')}
        >
          <Settings className="w-4 h-4 mr-1" />
          التفاصيل
        </Button>
      </div>

      {/* Timeline Tab */}
      {selectedTab === 'timeline' && (
        <Card className="border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
          <CardHeader>
            <CardTitle>سجل الأحداث</CardTitle>
          </CardHeader>
          <CardContent>
            {events.length === 0 ? (
              <div className="text-center py-8 text-[var(--text-secondary)]">
                لا توجد أحداث مسجلة
              </div>
            ) : (
              <div className="space-y-4">
                {events.map((event, index) => (
                  <div key={event.id} className="flex gap-4 relative">
                    {/* Timeline line */}
                    {index !== events.length - 1 && (
                      <div className="absolute right-[19px] top-8 bottom-0 w-0.5 bg-[var(--border-subtle)]" />
                    )}
                    
                    {/* Event icon */}
                    <div className="relative z-10 flex-shrink-0">
                      <div className="w-10 h-10 rounded-full bg-[var(--bg-surface-elevated)] border-2 border-[var(--border-default)] flex items-center justify-center">
                        {getEventIcon(event.event_type)}
                      </div>
                    </div>

                    {/* Event content */}
                    <div className="flex-1 pb-4">
                      <div className="flex items-start justify-between mb-2">
                        <div>
                          <h4 className="font-semibold text-[var(--text-primary)]">
                            {event.description}
                          </h4>
                          <p className="text-sm text-[var(--text-secondary)]">
                            {formatDate(event.event_date)}
                          </p>
                        </div>
                        {getEventBadge(event.status)}
                      </div>

                      {event.details && (
                        <div className="mt-2 p-3 rounded-lg bg-[var(--bg-surface-elevated)] border border-[var(--border-subtle)]">
                          <div className="grid grid-cols-2 gap-2 text-sm">
                            {Object.entries(event.details).map(([key, value]) => (
                              <div key={key}>
                                <p className="text-[var(--text-secondary)]">{key}</p>
                                <p className="text-[var(--text-primary)] font-medium">
                                  {typeof value === 'number' && key.includes('price') || key.includes('cost')
                                    ? `₪${(value / 100).toFixed(2)}`
                                    : String(value)}
                                </p>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}

                      {event.amount && (
                        <div className="mt-2 flex items-center gap-2 text-sm">
                          <DollarSign className="w-4 h-4 text-green" />
                          <span className="text-[var(--text-primary)] font-medium">
                            ₪{(event.amount / 100).toFixed(2)}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Financials Tab */}
      {selectedTab === 'financials' && (
        <Card className="border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
          <CardHeader>
            <CardTitle>السجل المالي</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex justify-between items-center p-4 rounded-lg bg-[var(--bg-surface-elevated)] border border-[var(--border-subtle)]">
                <div>
                  <p className="text-sm text-[var(--text-secondary)]">تكلفة الشراء</p>
                  <p className="text-xl font-bold text-[var(--text-primary)]">
                    ₪{(history.item?.cost || 0).toFixed(2)}
                  </p>
                </div>
                <ShoppingCart className="w-5 h-5 text-cyan" />
              </div>

              {history.repair_costs && history.repair_costs.length > 0 && (
                <div className="space-y-2">
                  <p className="text-sm font-semibold text-[var(--text-primary)]">تكاليف الإصلاح</p>
                  {history.repair_costs.map((repair: any, index: number) => (
                    <div key={index} className="flex justify-between items-center p-3 rounded-lg bg-[var(--bg-surface-elevated)] border border-[var(--border-subtle)]">
                      <div>
                        <p className="text-sm text-[var(--text-primary)]">{repair.description}</p>
                        <p className="text-xs text-[var(--text-secondary)]">{formatDate(repair.date)}</p>
                      </div>
                      <div className="text-right">
                        <p className="text-lg font-bold text-[var(--text-primary)]">
                          ₪{repair.amount.toFixed(2)}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}

              <div className="flex justify-between items-center p-4 rounded-lg bg-[var(--bg-surface-elevated)] border border-[var(--border-subtle)]">
                <div>
                  <p className="text-sm text-[var(--text-secondary)]">سعر البيع</p>
                  <p className="text-xl font-bold text-[var(--text-primary)]">
                    ₪{(history.item?.selling_price || 0).toFixed(2)}
                  </p>
                </div>
                <DollarSign className="w-5 h-5 text-green" />
              </div>

              <div className="flex justify-between items-center p-4 rounded-lg bg-gradient-to-r from-green/10 to-cyan/10 border border-green/20">
                <div>
                  <p className="text-sm text-[var(--text-secondary)]">الربح الإجمالي</p>
                  <p className="text-2xl font-bold text-green">
                    ₪{((history.item?.selling_price || 0) - (history.item?.cost || 0) - (history.total_repair_costs || 0)).toFixed(2)}
                  </p>
                </div>
                <TrendingUp className="w-6 h-6 text-green" />
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Details Tab */}
      {selectedTab === 'details' && (
        <Card className="border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
          <CardHeader>
            <CardTitle>تفاصيل القطعة</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-3">
                <h4 className="font-semibold text-[var(--text-primary)]">معلومات أساسية</h4>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">اسم المنتج</span>
                    <span className="text-[var(--text-primary)]">{history.item?.product_name || '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">الرقم التسلسلي</span>
                    <span className="text-[var(--text-primary)]">{history.item?.serial_number || '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">الباركود</span>
                    <span className="text-[var(--text-primary)]">{history.item?.barcode || '-'}</span>
                  </div>
                </div>
              </div>

              <div className="space-y-3">
                <h4 className="font-semibold text-[var(--text-primary)]">الحالة والتقييم</h4>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">الحالة</span>
                    <span className="text-[var(--text-primary)]">{history.item?.condition || '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">التقييم</span>
                    <span className="text-[var(--text-primary)]">{history.item?.grade || '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">الحالة الحالية</span>
                    <Badge variant={history.item?.status === 'AVAILABLE' ? 'success' : 'secondary'}>
                      {history.item?.status || '-'}
                    </Badge>
                  </div>
                </div>
              </div>

              <div className="space-y-3">
                <h4 className="font-semibold text-[var(--text-primary)]">معلومات الشراء</h4>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">البائع</span>
                    <span className="text-[var(--text-primary)]">{history.item?.seller_name || '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">تاريخ الشراء</span>
                    <span className="text-[var(--text-primary)]">
                      {history.item?.acquisition_date ? formatDate(history.item.acquisition_date) : '-'}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">حالة الدفع</span>
                    <Badge variant={history.item?.payment_status === 'PAID' ? 'success' : 'warning'}>
                      {history.item?.payment_status || '-'}
                    </Badge>
                  </div>
                </div>
              </div>

              <div className="space-y-3">
                <h4 className="font-semibold text-[var(--text-primary)]">معلومات إضافية</h4>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">الموقع</span>
                    <span className="text-[var(--text-primary)]">{history.item?.location || '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">الملاحظات</span>
                    <span className="text-[var(--text-primary)]">{history.item?.notes || '-'}</span>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}