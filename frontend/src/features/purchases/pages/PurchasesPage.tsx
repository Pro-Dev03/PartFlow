import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { purchasesApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { 
  ShoppingCart, 
  Search, 
  Plus, 
  Filter,
  Eye,
  Truck,
  Package,
  Calendar
} from 'lucide-react';

export function PurchasesPage() {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  const { data: purchasesData, isLoading } = useQuery({
    queryKey: ['purchases'],
    queryFn: () => purchasesApi.list(),
  });

  const purchases = (purchasesData?.data as any[]) || [];

  const filteredPurchases = purchases.filter((purchase: any) => {
    const matchesSearch = 
      purchase.supplier?.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      purchase.id.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = !statusFilter || purchase.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const getStatusBadge = (status: string) => {
    const variants: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
      pending: { label: 'قيد الانتظار', variant: 'secondary' },
      ordered: { label: 'تم الطلب', variant: 'outline' },
      received: { label: 'تم الاستلام', variant: 'default' },
      cancelled: { label: 'ملغي', variant: 'destructive' },
    };
    return variants[status] || { label: status, variant: 'default' };
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {t('purchases.title')}
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            إدارة المشتريات والطلبات
          </p>
        </div>
        <Button className="gap-2">
          <Plus className="w-4 h-4" />
          {t('purchases.newPurchase')}
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title="إجمالي المشتريات" value={purchases.length} icon={ShoppingCart} />
        <StatCard title="قيد الانتظار" value={purchases.filter((p: any) => p.status === 'pending').length} icon={Calendar} />
        <StatCard title="تم الاستلام" value={purchases.filter((p: any) => p.status === 'received').length} icon={Package} />
        <StatCard title="إجمالي التكلفة" value={`₪${(purchases as any[]).reduce((sum: number, p: any) => sum + p.totalCost, 0).toLocaleString()}`} icon={Truck} />
      </div>

      {/* Search and Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute inset-y-0 right-3 w-4 h-4 text-gray-400" />
              <Input
                placeholder={t('common.search')}
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
                  { value: 'ordered', label: 'تم الطلب' },
                  { value: 'received', label: 'تم الاستلام' },
                  { value: 'cancelled', label: 'ملغي' },
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

      {/* Purchases Table */}
      <Card>
        <CardHeader>
          <CardTitle>قائمة المشتريات</CardTitle>
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
                  <TableHead>رقم الطلب</TableHead>
                  <TableHead>المورد</TableHead>
                  <TableHead>القطع</TableHead>
                  <TableHead>إجمالي التكلفة</TableHead>
                  <TableHead>المدفوع</TableHead>
                  <TableHead>المتبقي</TableHead>
                  <TableHead>التاريخ المتوقع</TableHead>
                  <TableHead>الحالة</TableHead>
                  <TableHead className="text-left">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredPurchases.map((purchase: any) => {
                  const statusBadge = getStatusBadge(purchase.status);
                  return (
                    <TableRow key={purchase.id}>
                      <TableCell className="font-medium">{purchase.id}</TableCell>
                      <TableCell>{purchase.supplier?.name}</TableCell>
                      <TableCell>{purchase.items?.length || 0} قطع</TableCell>
                      <TableCell>₪{purchase.totalCost?.toLocaleString()}</TableCell>
                      <TableCell className="text-green-600">₪{purchase.paidAmount?.toLocaleString()}</TableCell>
                      <TableCell>₪{purchase.remainingAmount?.toLocaleString()}</TableCell>
                      <TableCell>
                        {purchase.expectedDate 
                          ? new Date(purchase.expectedDate).toLocaleDateString('ar-SA')
                          : '-'
                        }
                      </TableCell>
                      <TableCell>
                        <Badge variant={statusBadge.variant}>
                          {statusBadge.label}
                        </Badge>
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
  value: number | string;
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