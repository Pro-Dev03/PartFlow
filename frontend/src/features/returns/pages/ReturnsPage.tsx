import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { returnsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { SearchInput } from '../../../components/ui/search-input';
import { PageHeader } from '../../../components/ui/page-header';
import { StatCard } from '../../../components/ui/stat-card';
import { Select } from '../../../components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { getButtonSize } from '../../../config/button-sizes';
import { 
  RotateCcw, 
  Plus, 
  Eye,
  ShoppingCart,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Package,
  Download,
  Printer
} from 'lucide-react';

export function ReturnsPage() {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  const handleClearSearch = () => {
    setSearchQuery('');
  };

  const { data: returnsData, isLoading } = useQuery({
    queryKey: ['returns'],
    queryFn: () => returnsApi.list({ page: 1, per_page: 100 }),
  });

  const returns = (returnsData?.data as any[]) || [];

  const filteredReturns = returns.filter((returnItem: any) => {
    const matchesSearch = 
      returnItem.sale?.invoiceNumber?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      returnItem.customer?.name?.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = !statusFilter || returnItem.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const getStatusBadge = (status: string) => {
    const variants: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline'; icon: any }> = {
      pending: { label: 'قيد الانتظار', variant: 'secondary', icon: AlertTriangle },
      approved: { label: 'موافق عليه', variant: 'default', icon: CheckCircle },
      rejected: { label: 'مرفوض', variant: 'destructive', icon: XCircle },
      completed: { label: 'مكتمل', variant: 'outline', icon: CheckCircle },
    };
    return variants[status] || { label: status, variant: 'default', icon: AlertTriangle };
  };

  const handleExport = () => {
    const dataToExport = returns.map((returnItem: any) => ({
      'التاريخ': returnItem.date,
      'العميل': returnItem.customer,
      'المنتج': returnItem.product,
      'الحالة': getStatusBadge(returnItem.status).label,
      'السبب': returnItem.reason
    }));
    exportToCSV(dataToExport, `returns-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    const dataToPrint = returns.map((returnItem: any) => ({
      'التاريخ': returnItem.date,
      'العميل': returnItem.customer,
      'المنتج': returnItem.product,
      'الحالة': getStatusBadge(returnItem.status).label,
      'السبب': returnItem.reason
    }));
    printTable(dataToPrint, ['التاريخ', 'العميل', 'المنتج', 'الحالة', 'السبب'], 'تقرير المرتجعات');
  };

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Returns Management"
        title={t('returns.title')}
        description="إدارة المرتجعات والاسترجاع"
        actions={
          <div className="flex gap-sm">
            <Button variant="primary" size={getButtonSize('returns', 'headerActions')} className="gap-2">
              <Plus className="w-4 h-4" />
              {t('returns.newReturn')}
            </Button>
            <Button variant="secondary" size={getButtonSize('returns', 'headerActions')} onClick={handleExport} className="gap-2">
              <Download className="w-4 h-4" />
              تصدير
            </Button>
            <Button variant="secondary" size={getButtonSize('returns', 'headerActions')} onClick={handlePrint} className="gap-2">
              <Printer className="w-4 h-4" />
              طباعة
            </Button>
          </div>
        }
      />

      {/* Stats Cards - Futuristic + Clean */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-md">
        <StatCard title="إجمالي المرتجعات" value={returns.length} icon={RotateCcw} variant="featured" />
        <StatCard title="قيد الانتظار" value={returns.filter((r: any) => r.status === 'pending').length} icon={AlertTriangle} variant="warning" />
        <StatCard title="موافق عليه" value={returns.filter((r: any) => r.status === 'approved').length} icon={CheckCircle} variant="success" />
        <StatCard title="مكتمل" value={returns.filter((r: any) => r.status === 'completed').length} icon={Package} variant="default" />
      </div>

      {/* Search and Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base font-medium">البحث والتصفية</CardTitle>
        </CardHeader>
        <CardContent className="p-lg pt-0">
          <div className="flex flex-col md:flex-row gap-md items-start md:items-center">
            <div className="flex-1 w-full">
              <SearchInput
                placeholder="بحث برقم الفاتورة أو العميل..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onClear={handleClearSearch}
                size="md"
                className="w-full"
              />
            </div>
            <div className="w-full md:w-auto min-w-[200px]">
              <Select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                size="md"
                options={[
                  { value: '', label: 'كل الحالات' },
                  { value: 'pending', label: 'قيد الانتظار' },
                  { value: 'approved', label: 'موافق عليه' },
                  { value: 'rejected', label: 'مرفوض' },
                  { value: 'completed', label: 'مكتمل' },
                ]}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Returns Table */}
      <Card>
        <CardHeader>
          <CardTitle>سجل المرتجعات</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan" />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>رقم الفاتورة</TableHead>
                  <TableHead>العميل</TableHead>
                  <TableHead>القطعة</TableHead>
                  <TableHead>السبب</TableHead>
                  <TableHead>مبلغ الاسترجاع</TableHead>
                  <TableHead>يتطلب فحص</TableHead>
                  <TableHead>الحالة</TableHead>
                  <TableHead>التاريخ</TableHead>
                  <TableHead className="text-start">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredReturns.map((returnItem: any) => {
                  const statusBadge = getStatusBadge(returnItem.status);
                  const StatusIcon = statusBadge.icon;
                  return (
                    <TableRow key={returnItem.id}>
                      <TableCell className="font-medium">
                        {returnItem.sale?.invoiceNumber}
                      </TableCell>
                      <TableCell>{returnItem.customer?.name}</TableCell>
                      <TableCell>{returnItem.item?.product?.name}</TableCell>
                      <TableCell>{returnItem.reason}</TableCell>
                      <TableCell className="font-bold">
                        ₪{returnItem.refundAmount?.toLocaleString()}
                      </TableCell>
                      <TableCell>
                        {returnItem.inspectionRequired ? (
                          <Badge variant="danger">نعم</Badge>
                        ) : (
                          <Badge variant="default">لا</Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge variant={statusBadge.variant} className="gap-1">
                          <StatusIcon className="w-3 h-3" />
                          {statusBadge.label}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {new Date(returnItem.createdAt).toLocaleDateString('ar-SA')}
                      </TableCell>
                      <TableCell className="text-start">
                        <Button variant="ghost" size={getButtonSize('returns', 'tableAction')}>
                          <Eye className="w-4 h-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}