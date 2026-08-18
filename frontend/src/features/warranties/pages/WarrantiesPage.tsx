import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { warrantiesApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { 
  Shield, 
  Search, 
  Plus, 
  Filter,
  Eye,
  AlertTriangle,
  CheckCircle,
  Clock,
  Calendar
} from 'lucide-react';

export function WarrantiesPage() {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  const { data: warrantiesData, isLoading } = useQuery({
    queryKey: ['warranties'],
    queryFn: () => warrantiesApi.list(),
  });

  const warranties = (warrantiesData?.data as any[]) || [];

  const filteredWarranties = warranties.filter((warranty: any) => {
    const matchesSearch = 
      warranty.item?.product?.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      warranty.customer?.name?.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = !statusFilter || warranty.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const getDaysRemaining = (endDate: string) => {
    const today = new Date();
    const end = new Date(endDate);
    const diffTime = end.getTime() - today.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
  };

  const getWarrantyStatus = (endDate: string, currentStatus: string) => {
    if (currentStatus === 'claimed') return { label: 'تم المطالبة', variant: 'secondary' as const, icon: AlertTriangle };
    
    const daysRemaining = getDaysRemaining(endDate);
    if (daysRemaining < 0) return { label: 'منتهي', variant: 'destructive' as const, icon: AlertTriangle };
    if (daysRemaining <= 7) return { label: 'ينتهي قريباً', variant: 'outline' as const, icon: Clock };
    return { label: 'نشط', variant: 'default' as const, icon: CheckCircle };
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {t('warranties.title')}
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            إدارة الضمانات والمطالبات
          </p>
        </div>
        <Button className="gap-2">
          <Plus className="w-4 h-4" />
          إضافة ضمان
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard
          title="الضمانات النشطة"
          value={(warranties as any[]).filter((w: any) => w.status === 'active').length}
          icon={Shield}
        />
        <StatCard
          title="تنتهي قريباً"
          value={(warranties as any[]).filter((w: any) => {
            const days = getDaysRemaining(w.endDate);
            return w.status === 'active' && days > 0 && days <= 7;
          }).length}
          icon={Clock}
        />
        <StatCard
          title="منتهية"
          value={(warranties as any[]).filter((w: any) => w.status === 'expired').length}
          icon={AlertTriangle}
        />
        <StatCard
          title="تم المطالبة"
          value={(warranties as any[]).filter((w: any) => w.status === 'claimed').length}
          icon={CheckCircle} 
        />
      </div>

      {/* Search and Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute inset-y-0 right-3 w-4 h-4 text-gray-400" />
              <Input
                placeholder="بحث عن منتج أو عميل..."
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
                  { value: 'active', label: 'نشط' },
                  { value: 'expiring', label: 'ينتهي قريباً' },
                  { value: 'expired', label: 'منتهي' },
                  { value: 'claimed', label: 'تم المطالبة' },
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

      {/* Warranties Table */}
      <Card>
        <CardHeader>
          <CardTitle>سجل الضمانات</CardTitle>
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
                  <TableHead>المنتج</TableHead>
                  <TableHead>العميل</TableHead>
                  <TableHead>تاريخ البدء</TableHead>
                  <TableHead>تاريخ الانتهاء</TableHead>
                  <TableHead>المدة</TableHead>
                  <TableHead>الأيام المتبقية</TableHead>
                  <TableHead>الحالة</TableHead>
                  <TableHead>المطالبات</TableHead>
                  <TableHead className="text-left">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredWarranties.map((warranty: any) => {
                  const daysRemaining = getDaysRemaining(warranty.endDate);
                  const statusBadge = getWarrantyStatus(warranty.endDate, warranty.status);
                  const StatusIcon = statusBadge.icon;
                  
                  return (
                    <TableRow key={warranty.id}>
                      <TableCell className="font-medium">
                        {warranty.item?.product?.name}
                      </TableCell>
                      <TableCell>{warranty.customer?.name}</TableCell>
                      <TableCell>
                        {new Date(warranty.startDate).toLocaleDateString('ar-SA')}
                      </TableCell>
                      <TableCell>
                        {new Date(warranty.endDate).toLocaleDateString('ar-SA')}
                      </TableCell>
                      <TableCell>{warranty.duration} شهر</TableCell>
                      <TableCell>
                        {warranty.status === 'active' ? (
                          <span className={daysRemaining <= 7 ? 'text-orange-600 font-medium' : 'text-gray-600'}>
                            {daysRemaining} يوم
                          </span>
                        ) : (
                          <span className="text-gray-400">-</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge variant={statusBadge.variant} className="gap-1">
                          <StatusIcon className="w-3 h-3" />
                          {statusBadge.label}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">
                          {warranty.claims?.length || 0}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-left">
                        <div className="flex gap-2">
                          <Button variant="ghost" size="sm">
                            <Eye className="w-4 h-4" />
                          </Button>
                          {warranty.status === 'active' && (
                            <Button variant="primary" size="sm">
                              مطالبة
                            </Button>
                          )}
                        </div>
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