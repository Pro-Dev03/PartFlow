import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { suppliersApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { LoadingSpinner } from '../../../components/ui/loading-spinner';
import { DataTable, Column } from '../../../components/tables/data-table';
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
  ChevronUp,
  MoreHorizontal,
  EyeOff,
  RefreshCw
} from 'lucide-react';

export function SuppliersPage() {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedSupplierId, setExpandedSupplierId] = useState<string | null>(null);

  const { data: suppliersData, isLoading, refetch } = useQuery({
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
  const totalPurchases = suppliers.reduce((sum: number, s: any) => sum + (s.totalPurchases || 0), 0);
  const totalPaid = suppliers.reduce((sum: number, s: any) => sum + (s.paidAmount || 0), 0);
  const totalOutstanding = suppliers.reduce((sum: number, s: any) => sum + (s.outstanding || 0), 0);

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

  const columns: Column<any>[] = [
    {
      key: 'name',
      title: 'الاسم',
      width: '20%',
      sortable: true,
      render: (supplier) => <span className="font-medium">{supplier.name}</span>
    },
    {
      key: 'phone',
      title: 'الهاتف',
      width: '18%',
      sortable: true,
      render: (supplier) => (
        <div className="flex items-center gap-2">
          <Phone className="w-4 h-4 text-text-muted" />
          {supplier.phone}
        </div>
      )
    },
    {
      key: 'email',
      title: 'البريد الإلكتروني',
      width: '22%',
      sortable: true,
      render: (supplier) => (
        supplier.email ? (
          <div className="flex items-center gap-2">
            <Mail className="w-4 h-4 text-text-muted" />
            {supplier.email}
          </div>
        ) : '-'
      )
    },
    {
      key: 'totalPurchases',
      title: 'إجمالي المشتريات',
      width: '14%',
      sortable: true,
      render: (supplier) => `₪${(supplier.totalPurchases || 0).toLocaleString()}`
    },
    {
      key: 'paidAmount',
      title: 'المدفوع',
      width: '12%',
      sortable: true,
      render: (supplier) => <span className="text-green">₪{(supplier.paidAmount || 0).toLocaleString()}</span>
    },
    {
      key: 'outstanding',
      title: 'المستحق',
      width: '12%',
      sortable: true,
      render: (supplier) => (
        <Badge variant={supplier.outstanding > 0 ? 'danger' : 'default'}>
          ₪{(supplier.outstanding || 0).toLocaleString()}
        </Badge>
      )
    },
    {
      key: 'lastPurchase',
      title: 'آخر شراء',
      width: '12%',
      sortable: true,
      render: (supplier) => supplier.lastPurchase
        ? new Date(supplier.lastPurchase).toLocaleDateString('ar-SA')
        : '-'
    },
    {
      key: 'actions',
      title: 'الإجراءات',
      width: '10%',
      render: (supplier) => (
        <div className="flex gap-2" onClick={(e) => e.stopPropagation()}>
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
      )
    }
  ];

  const renderExpanded = (supplier: any) => {
    const currentInventory = expandedSupplierId === supplier.id ? supplierInventory : null;

    return (
      <div className="p-4 bg-surface/30">
        <div className="flex items-center gap-2 mb-4">
          <Package className="w-4 h-4 text-cyan" />
          <h4 className="font-semibold text-text">بضاعة المورد - {supplier.name}</h4>
        </div>
        {currentInventory && currentInventory.data && currentInventory.data.length > 0 ? (
          <DataTable
            data={currentInventory.data}
            columns={[
              { key: 'product_name', title: 'المنتج', sortable: true, render: (item) => <span className="font-medium">{item.product_name}</span> },
              { key: 'total_received', title: 'المستلمة', sortable: true },
              { key: 'available', title: 'المتاحة', sortable: true, render: (item) => <span className="text-green-600">{item.available}</span> },
              { key: 'sold', title: 'المباعة', sortable: true, render: (item) => <span className="text-red-600">{item.sold}</span> },
              { key: 'reserved', title: 'المحجوزة', sortable: true, render: (item) => <span className="text-yellow-600">{item.reserved}</span> },
              { key: 'damaged', title: 'التالفة', sortable: true, render: (item) => <span className="text-orange-600">{item.damaged}</span> },
              { key: 'avg_cost', title: 'متوسط التكلفة', sortable: true, render: (item) => `₪${(item.avg_cost || 0).toFixed(2)}` },
              { key: 'avg_price', title: 'متوسط السعر', sortable: true, render: (item) => `₪${(item.avg_price || 0).toFixed(2)}` }
            ]}
            expandable={false}
          />
        ) : (
          <div className="text-center py-8 text-text-muted">
            لا توجد بضاعة من هذا المورد
          </div>
        )}
      </div>
    );
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
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-small text-text-muted">إجمالي الموردين</p>
                <p className="text-h2 font-bold text-text mt-1">
                  {totalSuppliers}
                </p>
              </div>
              <Truck className="w-5 h-5 text-text-muted" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-small text-text-muted">إجمالي المشتريات</p>
                <p className="text-h2 font-bold text-text mt-1">
                  ₪{totalPurchases.toLocaleString()}
                </p>
              </div>
              <DollarSign className="w-5 h-5 text-text-muted" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-small text-text-muted">المدفوع</p>
                <p className="text-h2 font-bold text-text mt-1">
                  ₪{totalPaid.toLocaleString()}
                </p>
              </div>
              <DollarSign className="w-5 h-5 text-text-muted" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-small text-text-muted">المستحق</p>
                <p className="text-h2 font-bold text-text mt-1">
                  ₪{totalOutstanding.toLocaleString()}
                </p>
              </div>
              <DollarSign className="w-5 h-5 text-text-muted" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* AI Insights */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Sparkles style={{ width: '20px', height: '20px', color: 'var(--color-primary)' }} />
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
                <Target style={{ width: '16px', height: '16px', color: 'var(--color-primary)' }} />
              </div>
              <div>
                <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>فرصة تحسين التوريد</p>
                <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
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
                <AlertTriangle style={{ width: '16px', height: '16px', color: 'var(--color-warning)' }} />
              </div>
              <div>
                <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>تنبيه تأخير التوريد</p>
                <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
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
                <TrendingUp style={{ width: '16px', height: '16px', color: 'var(--color-success)' }} />
              </div>
              <div>
                <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>أداء ممتاز</p>
                <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
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
            <DataTable
              data={filteredSuppliers}
              columns={columns}
              loading={isLoading}
              refreshable
              onRefresh={() => refetch()}
              onExport={handleExport}
              expandable
              renderExpanded={renderExpanded}
              empty={
                <div className="text-center py-8 text-text-muted">
                  لا يوجد موردين
                </div>
              }
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
