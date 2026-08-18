import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { debtsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { 
  DollarSign, 
  Search, 
  AlertTriangle,
  Calendar,
  TrendingUp,
  CreditCard
} from 'lucide-react';

export function DebtsPage() {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const queryClient = useQueryClient();

  const { data: overdueCustomersData, isLoading } = useQuery({
    queryKey: ['debts'],
    queryFn: () => debtsApi.list(),
  });

  const recordPaymentMutation = useMutation({
    mutationFn: ({ customerId, amount }: { customerId: string; amount: number }) =>
      debtsApi.recordPayment(customerId, { amount }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['debts'] });
    },
  });

  const handleRecordPayment = (customerId: string) => {
    const amount = prompt('أدخل مبلغ الدفعة:');
    if (amount && !isNaN(parseFloat(amount))) {
      recordPaymentMutation.mutate({
        customerId,
        amount: parseFloat(amount),
      });
    }
  };

  const overdueCustomers = (overdueCustomersData?.data as any[]) || [];

  // Transform customer data to debt entries for display
  const debts = overdueCustomers.flatMap((customer: any) => {
    return (customer.debts || []).map((debt: any) => ({
      ...debt,
      customer: {
        id: customer.id,
        name: customer.name,
        code: customer.code,
      },
    }));
  });

  const totalDebts = debts.reduce((sum: number, d: any) => sum + d.amount, 0);
  const overdueDebts = debts.filter((d: any) => d.status === 'overdue');
  const overdueAmount = overdueDebts.reduce((sum: number, d: any) => sum + d.remainingAmount, 0);
  const dueSoonDebts = debts.filter((d: any) => d.status === 'pending');
  const dueSoonAmount = dueSoonDebts.reduce((sum: number, d: any) => sum + d.remainingAmount, 0);

  const filteredDebts = debts.filter((debt: any) =>
    debt.customer?.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getDaysOverdue = (dueDate: string) => {
    const today = new Date();
    const due = new Date(dueDate);
    const diffTime = today.getTime() - due.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
  };

  const getAgingCategory = (daysOverdue: number) => {
    if (daysOverdue < 0) return { label: 'غير مستحق', variant: 'default' as const };
    if (daysOverdue <= 7) return { label: 'حديث', variant: 'success' as const };
    if (daysOverdue <= 30) return { label: '1-30 يوم', variant: 'warning' as const };
    if (daysOverdue <= 60) return { label: '31-60 يوم', variant: 'danger' as const };
    return { label: '+60 يوم', variant: 'destructive' as const };
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
          {t('debts.title')}
        </h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">
          إدارة ديون العملاء والمتابعة
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard 
          title={t('debts.totalDebts')} 
          value={`₪${totalDebts.toLocaleString()}`} 
          icon={DollarSign} 
        />
        <StatCard 
          title={t('debts.overdue')} 
          value={`₪${overdueAmount.toLocaleString()}`} 
          icon={AlertTriangle}
          variant="danger"
        />
        <StatCard 
          title={t('debts.dueSoon')} 
          value={`₪${dueSoonAmount.toLocaleString()}`} 
          icon={Calendar}
          variant="warning"
        />
        <StatCard 
          title={t('debts.paid')} 
          value={`₪${(totalDebts - overdueAmount - dueSoonAmount).toLocaleString()}`} 
          icon={TrendingUp}
          variant="success"
        />
      </div>

      {/* Overdue Debts Alert - Integration with Debt Worker */}
      {overdueDebts.length > 0 && (
        <Card className="border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/10">
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <AlertTriangle className="w-5 h-5 text-red-600 dark:text-red-400" />
              <div className="flex-1">
                <p className="font-medium text-red-900 dark:text-red-100">
                  تنبيه: لديك {overdueDebts.length} ديون متأخرة
                </p>
                <p className="text-sm text-red-700 dark:text-red-300">
                  إجمالي المبلغ المتأخر: ₪{overdueAmount.toLocaleString()}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Search */}
      <Card>
        <CardContent className="p-4">
          <div className="relative">
            <Search className="absolute inset-y-0 right-3 w-4 h-4 text-gray-400" />
            <Input
              placeholder="بحث عن عميل..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pr-10"
            />
          </div>
        </CardContent>
      </Card>

      {/* Debts Table */}
      <Card>
        <CardHeader>
          <CardTitle>قائمة الديون</CardTitle>
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
                  <TableHead>العميل</TableHead>
                  <TableHead>المبلغ الكلي</TableHead>
                  <TableHead>المدفوع</TableHead>
                  <TableHead>المتبقي</TableHead>
                  <TableHead>تاريخ الاستحقاق</TableHead>
                  <TableHead>أيام التأخير</TableHead>
                  <TableHead>فئة التقادم</TableHead>
                  <TableHead>الحالة</TableHead>
                  <TableHead className="text-left">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredDebts.map((debt: any) => {
                  const daysOverdue = getDaysOverdue(debt.dueDate);
                  const isOverdue = daysOverdue > 0;
                  const agingCategory = getAgingCategory(daysOverdue);
                  
                  return (
                    <TableRow key={debt.id}>
                      <TableCell className="font-medium">
                        {debt.customer?.name}
                      </TableCell>
                      <TableCell>₪{debt.amount.toLocaleString()}</TableCell>
                      <TableCell className="text-green-600">
                        ₪{debt.paidAmount.toLocaleString()}
                      </TableCell>
                      <TableCell className="font-bold">
                        ₪{debt.remainingAmount.toLocaleString()}
                      </TableCell>
                      <TableCell>
                        {new Date(debt.dueDate).toLocaleDateString('ar-SA')}
                      </TableCell>
                      <TableCell>
                        {isOverdue ? (
                          <span className="text-red-600 font-medium">
                            {daysOverdue} يوم
                          </span>
                        ) : (
                          <span className="text-gray-500">
                            {Math.abs(daysOverdue)} يوم
                          </span>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge variant={agingCategory.variant}>
                          {agingCategory.label}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant={
                          debt.status === 'paid' ? 'default' :
                          debt.status === 'overdue' ? 'destructive' : 'secondary'
                        }>
                          {debt.status === 'paid' ? 'مدفوع' :
                           debt.status === 'overdue' ? 'متأخر' : 'مستحق'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-left">
                        {debt.status !== 'paid' && (
                          <Button
                            variant="primary"
                            size="sm"
                            className="gap-2"
                            onClick={() => handleRecordPayment(debt.customer.id)}
                            disabled={recordPaymentMutation.isPending}
                          >
                            <CreditCard className="w-4 h-4" />
                            تسجيل دفعة
                          </Button>
                        )}
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
  value: string;
  icon: any;
  variant?: 'default' | 'danger' | 'warning' | 'success';
}

function StatCard({ title, value, icon: Icon, variant = 'default' }: StatCardProps) {
  const variantStyles = {
    default: 'text-gray-900 dark:text-gray-100',
    danger: 'text-red-600',
    warning: 'text-orange-600',
    success: 'text-green-600',
  };

  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-gray-500 dark:text-gray-400">{title}</p>
            <p className={`text-2xl font-bold mt-1 ${variantStyles[variant]}`}>
              {value}
            </p>
          </div>
          <Icon className={`w-5 h-5 ${
            variant === 'danger' ? 'text-red-500' :
            variant === 'warning' ? 'text-orange-500' :
            variant === 'success' ? 'text-green-500' :
            'text-gray-400'
          }`} />
        </div>
      </CardContent>
    </Card>
  );
}