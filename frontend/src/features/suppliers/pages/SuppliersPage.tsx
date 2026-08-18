import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { suppliersApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { 
  Truck, 
  Search, 
  Plus, 
  Filter,
  Eye,
  Edit,
  Phone,
  Mail,
  DollarSign
} from 'lucide-react';

export function SuppliersPage() {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');

  const { data: suppliersData, isLoading } = useQuery({
    queryKey: ['suppliers'],
    queryFn: () => suppliersApi.list(),
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

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {t('suppliers.title')}
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            إدارة الموردين والمشتريات
          </p>
        </div>
        <Button className="gap-2">
          <Plus className="w-4 h-4" />
          {t('suppliers.addSupplier')}
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title="إجمالي الموردين" value={totalSuppliers} icon={Truck} />
        <StatCard title="إجمالي المشتريات" value={`₪${totalPurchases.toLocaleString()}`} icon={DollarSign} />
        <StatCard title="المدفوع" value={`₪${totalPaid.toLocaleString()}`} icon={DollarSign} />
        <StatCard title="المستحق" value={`₪${totalOutstanding.toLocaleString()}`} icon={DollarSign} />
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
              <Button variant="outline" className="gap-2">
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
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
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
                  <TableRow key={supplier.id}>
                    <TableCell className="font-medium">{supplier.name}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Phone className="w-4 h-4 text-gray-400" />
                        {supplier.phone}
                      </div>
                    </TableCell>
                    <TableCell>
                      {supplier.email ? (
                        <div className="flex items-center gap-2">
                          <Mail className="w-4 h-4 text-gray-400" />
                          {supplier.email}
                        </div>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                    <TableCell>₪{supplier.totalPurchases?.toLocaleString() || 0}</TableCell>
                    <TableCell className="text-green-600">₪{supplier.paidAmount?.toLocaleString() || 0}</TableCell>
                    <TableCell>
                      <Badge variant={supplier.outstanding > 0 ? 'destructive' : 'default'}>
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
                        <Button variant="ghost" size="sm">
                          <Eye className="w-4 h-4" />
                        </Button>
                        <Button variant="ghost" size="sm">
                          <Edit className="w-4 h-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
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