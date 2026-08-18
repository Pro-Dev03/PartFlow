import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { returnsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { 
  RotateCcw, 
  Search, 
  Plus, 
  Filter,
  Eye,
  ShoppingCart,
  AlertTriangle,
  CheckCircle,
  XCircle
} from 'lucide-react';

export function ReturnsPage() {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  const { data: returnsData, isLoading } = useQuery({
    queryKey: ['returns'],
    queryFn: () => returnsApi.list(),
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

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {t('returns.title')}
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            إدارة المرتجعات والاسترجاع
          </p>
        </div>
        <Button className="gap-2">
          <Plus className="w-4 h-4" />
          {t('returns.newReturn')}
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title="إجمالي المرتجعات" value={returns.length} icon={RotateCcw} />
        <StatCard title="قيد الانتظار" value={returns.filter((r: any) => r.status === 'pending').length} icon={AlertTriangle} />
        <StatCard title="موافق عليه" value={returns.filter((r: any) => r.status === 'approved').length} icon={CheckCircle} />
        <StatCard title="مكتمل" value={returns.filter((r: any) => r.status === 'completed').length} icon={CheckCircle} />
      </div>

      {/* Search and Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute inset-y-0 right-3 w-4 h-4 text-gray-400" />
              <Input
                placeholder="بحث برقم الفاتورة أو العميل..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pr-10"
              />
            </div>
            <div className="flex gap-2">
              <Select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                options={[
                  { value: '', label: 'كل الحالات' },
                  { value: 'pending', label: 'قيد الانتظار' },
                  { value: 'approved', label: 'موافق عليه' },
                  { value: 'rejected', label: 'مرفوض' },
                  { value: 'completed', label: 'مكتمل' },
                ]}
              />
              <Button variant="outline" className="gap-2">
                <Filter className="w-4 h-4" />
                {t('common.filter')}
              </Button>
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
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
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
                  <TableHead className="text-left">الإجراءات</TableHead>
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
                          <Badge variant="destructive">نعم</Badge>
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
                      <TableCell className="text-left">
                        <Button variant="ghost" size="sm">
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

interface StatCardProps {
  title: string;
  value: number;
  icon: any;
}

function StatCard({ title, value, icon: Icon }: StatCardProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-gray-500 dark:text-gray-400">{title}</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">
              {value}
            </p>
          </div>
          <Icon className="w-5 h-5 text-gray-400" />
        </div>
      </CardContent>
    </Card>
  );
}