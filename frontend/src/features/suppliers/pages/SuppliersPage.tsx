import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { suppliersApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { LoadingSpinner } from '../../../components/ui/loading-spinner';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { getButtonSize } from '../../../config/button-sizes';
import { 
  Truck, 
  Search, 
  Plus, 
  Filter,
  Eye,
  Edit,
  Phone,
  Mail,
  DollarSign,
  Sparkles,
  Target,
  TrendingUp,
  AlertTriangle,
  Download,
  Printer,
  Package,
  ChevronDown,
  ChevronUp
} from 'lucide-react';

export function SuppliersPage() {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedSupplierId, setExpandedSupplierId] = useState<string | null>(null);

  const { data: suppliersData, isLoading } = useQuery({
    queryKey: ['suppliers'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 100 }),
  });

  const { data: supplierInventory } = useQuery({
    queryKey: ['supplier-inventory', expandedSupplierId],
    queryFn: () => suppliersApi.getSupplierInventory(expandedSupplierId!),
    enabled: !!expandedSupplierId,
  });

  const suppliers = (suppliersData?.data as any[]) || [];

  const filteredSuppliers = suppliers.filter((supplier: any) =>
    supplier.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    supplier.phone.includes(searchQuery)
  );

  const totalSuppliers = suppliers.length;
  const totalPurchases = suppliers.reduce((sum: number, s: any) => sum + s.totalPurchases, 0);
  const totalPaid = suppliers.reduce((sum: number, s: any) => sum + s.paidAmount, 0);
  const totalOutstanding = suppliers.reduce((sum: number, s: any) => sum + s.outstanding, 0);

  const handleExport = () => {
    const dataToExport = filteredSuppliers.map((supplier: any) => ({
      'الاسم': supplier.name,
      'الهاتف': supplier.phone,
      'البريد': supplier.email || '-',
      'المشتريات': supplier.totalPurchases,
      'المدفوع': supplier.paidAmount,
      'المستحق': supplier.outstanding
    }));
    exportToCSV(dataToExport, `suppliers-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    const dataToPrint = filteredSuppliers.map((supplier: any) => ({
      'الاسم': supplier.name,
      'الهاتف': supplier.phone,
      'البريد': supplier.email || '-',
      'المشتريات': supplier.totalPurchases,
      'المدفوع': supplier.paidAmount,
      'المستحق': supplier.outstanding
    }));
    printTable(dataToPrint, ['الاسم', 'الهاتف', 'البريد', 'المشتريات', 'المدفوع', 'المستحق'], 'تقرير الموردين');
  };

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Supplier Hub"
        title={t('suppliers.title')}
        description="إدارة الموردين والمشتريات مع رؤى ذكية"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="primary" size={getButtonSize('suppliers', 'headerActions')} className="gap-2">
              <Plus className="w-4 h-4" />
              {t('suppliers.addSupplier')}
            </Button>
            <Button variant="secondary" size={getButtonSize('suppliers', 'headerActions')} onClick={handleExport} className="gap-2">
              <Download className="w-4 h-4" />
              تصدير
            </Button>
            <Button variant="secondary" size={getButtonSize('suppliers', 'headerActions')} onClick={handlePrint} className="gap-2">
              <Printer className="w-4 h-4" />
              طباعة
            </Button>
          </div>
        }
      />

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '14px' }}
           className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard title="إجمالي الموردين" value={totalSuppliers} icon={Truck} />
        <StatCard title="إجمالي المشتريات" value={`₪${totalPurchases.toLocaleString()}`} icon={DollarSign} />
        <StatCard title="المدفوع" value={`₪${totalPaid.toLocaleString()}`} icon={DollarSign} />
        <StatCard title="المستحق" value={`₪${totalOutstanding.toLocaleString()}`} icon={DollarSign} />
      </div>

      {/* AI Insights */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Sparkles style={{ width: '20px', height: '20px', color: '#22d3ee' }} />
            AI Insights - الموردين
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
            <div style={{ display: 'flex', gap: '14px' }}>
              <div style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'rgba(34, 211, 238, 0.1)',
                flexShrink: 0
              }}>
                <Target style={{ width: '16px', height: '16px', color: '#22d3ee' }} />
              </div>
              <div>
                <p style={{ fontSize: '13px', fontWeight: '600', color: '#f1f7ff' }}>فرصة تحسين التوريد</p>
                <p style={{ fontSize: '11px', color: '#8290a7', marginTop: '4px' }}>
                  المورد "إلكترونيات المتقدمة" يقدم أسعاراً أقل 15% من المنافسين مع جودة مماثلة. يُنصح بزيادة حجم التعامل.
                </p>
              </div>
            </div>
            <div style={{ display: 'flex', gap: '14px' }}>
              <div style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'rgba(251, 191, 36, 0.1)',
                flexShrink: 0
              }}>
                <AlertTriangle style={{ width: '16px', height: '16px', color: '#fbbf24' }} />
              </div>
              <div>
                <p style={{ fontSize: '13px', fontWeight: '600', color: '#f1f7ff' }}>تنبيه تأخير التوريد</p>
                <p style={{ fontSize: '11px', color: '#8290a7', marginTop: '4px' }}>
                  المورد "شاشات المستقبل" تأخر 3 مرات هذا الشهر. يُنصح بالبحث عن بدائل أو تقييم العقد.
                </p>
              </div>
            </div>
            <div style={{ display: 'flex', gap: '14px' }}>
              <div style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'rgba(52, 211, 153, 0.1)',
                flexShrink: 0
              }}>
                <TrendingUp style={{ width: '16px', height: '16px', color: '#34d399' }} />
              </div>
              <div>
                <p style={{ fontSize: '13px', fontWeight: '600', color: '#f1f7ff' }}>أداء ممتاز</p>
                <p style={{ fontSize: '11px', color: '#8290a7', marginTop: '4px' }}>
                  المورد "بطاريات القوة" حقق 100% من مواعيد التسليم هذا الربع. يُنصح بتجديد العقد تلقائياً.
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Search and Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute inset-y-0 right-3 w-4 h-4 text-text-muted" />
              <Input
                placeholder={t('common.search')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pr-10"
              />
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size={getButtonSize('suppliers', 'headerActions')} className="gap-2">
                <Filter className="w-4 h-4" />
                {t('common.filter')}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Suppliers Table */}
      <Card>
        <CardHeader>
          <CardTitle>قائمة الموردين</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <LoadingSpinner size="md" color="cyan" />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>الاسم</TableHead>
                  <TableHead>الهاتف</TableHead>
                  <TableHead>البريد الإلكتروني</TableHead>
                  <TableHead>إجمالي المشتريات</TableHead>
                  <TableHead>المدفوع</TableHead>
                  <TableHead>المستحق</TableHead>
                  <TableHead>آخر شراء</TableHead>
                  <TableHead className="text-left">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredSuppliers.map((supplier: any) => (
                  <>
                    <TableRow key={supplier.id}>
                      <TableCell className="font-medium">{supplier.name}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Phone className="w-4 h-4 text-text-muted" />
                          {supplier.phone}
                        </div>
                      </TableCell>
                      <TableCell>
                        {supplier.email ? (
                          <div className="flex items-center gap-2">
                            <Mail className="w-4 h-4 text-text-muted" />
                            {supplier.email}
                          </div>
                        ) : (
                          '-'
                        )}
                      </TableCell>
                      <TableCell>₪{supplier.totalPurchases?.toLocaleString() || 0}</TableCell>
                      <TableCell className="text-green">₪{supplier.paidAmount?.toLocaleString() || 0}</TableCell>
                      <TableCell>
                        <Badge variant={supplier.outstanding > 0 ? 'danger' : 'default'}>
                          ₪{supplier.outstanding?.toLocaleString() || 0}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {supplier.lastPurchase
                          ? new Date(supplier.lastPurchase).toLocaleDateString('ar-SA')
                          : '-'
                        }
                      </TableCell>
                      <TableCell className="text-left">
                        <div className="flex gap-2">
                          <Button
                            variant="ghost"
                            size={getButtonSize('suppliers', 'iconAction')}
                            onClick={() => setExpandedSupplierId(expandedSupplierId === supplier.id ? null : supplier.id)}
                          >
                            {expandedSupplierId === supplier.id ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
                          </Button>
                          <Button variant="ghost" size={getButtonSize('suppliers', 'iconAction')}>
                            <Eye className="w-3.5 h-3.5" />
                          </Button>
                          <Button variant="ghost" size={getButtonSize('suppliers', 'iconAction')}>
                            <Edit className="w-3.5 h-3.5" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                    {expandedSupplierId === supplier.id && (
                      <TableRow>
                        <TableCell colSpan={8} className="p-0">
                          <div className="p-4 bg-surface/30">
                            <div className="flex items-center gap-2 mb-4">
                              <Package className="w-4 h-4 text-cyan" />
                              <h4 className="font-semibold text-text">بضاعة المورد - {supplier.name}</h4>
                            </div>
                            {supplierInventory && supplierInventory.data && supplierInventory.data.length > 0 ? (
                              <Table>
                                <TableHeader>
                                  <TableRow>
                                    <TableHead>المنتج</TableHead>
                                    <TableHead>المستلمة</TableHead>
                                    <TableHead>المتاحة</TableHead>
                                    <TableHead>المباعة</TableHead>
                                    <TableHead>المحجوزة</TableHead>
                                    <TableHead>التالفة</TableHead>
                                    <TableHead>متوسط التكلفة</TableHead>
                                    <TableHead>متوسط السعر</TableHead>
                                  </TableRow>
                                </TableHeader>
                                <TableBody>
                                  {supplierInventory.data.map((item: any) => (
                                    <TableRow key={item.product_id}>
                                      <TableCell className="font-medium">{item.product_name}</TableCell>
                                      <TableCell>{item.total_received}</TableCell>
                                      <TableCell className="text-green-600">{item.available}</TableCell>
                                      <TableCell className="text-red-600">{item.sold}</TableCell>
                                      <TableCell className="text-yellow-600">{item.reserved}</TableCell>
                                      <TableCell className="text-orange-600">{item.damaged}</TableCell>
                                      <TableCell>₪{item.avg_cost?.toFixed(2) || 0}</TableCell>
                                      <TableCell>₪{item.avg_price?.toFixed(2) || 0}</TableCell>
                                    </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            ) : (
                              <div className="text-center py-8 text-text-muted">
                                لا توجد بضاعة من هذا المورد
                              </div>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    )}
                  </>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

interface StatCardProps {
  title: string;
  value: number | string;
  icon: any;
}

function StatCard({ title, value, icon: Icon }: StatCardProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-small text-text-muted">{title}</p>
            <p className="text-h2 font-bold text-text mt-1">
              {value}
            </p>
          </div>
          <Icon className="w-5 h-5 text-text-muted" />
        </div>
      </CardContent>
    </Card>
  );
}