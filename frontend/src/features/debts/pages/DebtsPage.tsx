import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { debtsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { 
  DollarSign, 
  Search, 
  AlertTriangle,
  Calendar,
  TrendingUp,
  CreditCard,
  Sparkles,
  TrendingDown,
  Bell,
  AlertCircle,
  Zap,
  Clock
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
    if (daysOverdue < 0) return { label: 'غير مستحق', variant: 'success' as const };
    if (daysOverdue <= 7) return { label: 'حديث', variant: 'success' as const };
    if (daysOverdue <= 30) return { label: '1-30 يوم', variant: 'warning' as const };
    if (daysOverdue <= 60) return { label: '31-60 يوم', variant: 'danger' as const };
    return { label: '+60 يوم', variant: 'danger' as const };
  };

  return (
    <div className="space-y-md">
      {/* Page Header - Futuristic + Attention-focused */}
      <PageHeader
        eyebrow="Debt Intelligence"
        title={t('debts.title')}
        description="إدارة ديون العملاء والمتابعة مع تنبيهات ذكية"
        actions={
          <div className="flex items-center gap-sm">
            <Button variant="secondary" className="gap-2">
              <Bell className="w-4 h-4" />
              إشعارات
            </Button>
            <Button variant="secondary" className="gap-2">
              <Zap className="w-4 h-4" />
              تحديث
            </Button>
          </div>
        }
      />

      {/* AI Debt Insight - Futuristic + Attention-focused */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-cyan" />
            AI Debt Insight
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-md">
            <div className="w-8 h-8 rounded-sm bg-red/10 flex items-center justify-center flex-shrink-0">
              <AlertCircle className="w-4 h-4 text-red" />
            </div>
            <div>
              <p className="text-small font-semibold text-text">ديون حرجة تحتاج انتباه فوري</p>
              <p className="text-tiny text-text-muted mt-1">
                لديك 3 ديون متأخرة أكثر من 60 يوم. الإجمالي: ₪{overdueAmount.toLocaleString()}
              </p>
              <Button variant="ghost" size="sm" className="mt-2 text-red">
                عرض التفاصيل ←
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Stats Cards - Futuristic + Attention-focused */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard 
          title={t('debts.totalDebts')} 
          value={`₪${totalDebts.toLocaleString()}`} 
          icon={DollarSign}
          subtitle="إجمالي الديون"
          variant="default"
          trend="+5%"
          trendUp={false}
        />
        <StatCard 
          title={t('debts.overdue')} 
          value={`₪${overdueAmount.toLocaleString()}`} 
          icon={AlertTriangle}
          subtitle="ديون متأخرة"
          variant={overdueAmount > 0 ? 'danger' : 'success'}
          trend={overdueAmount > 0 ? '+3' : null}
          trendUp={false}
        />
        <StatCard 
          title={t('debts.dueSoon')} 
          value={`₪${dueSoonAmount.toLocaleString()}`} 
          icon={Clock}
          subtitle="ديون قريب الاستحقاق"
          variant={dueSoonAmount > 0 ? 'warning' : 'success'}
          trend={dueSoonAmount > 0 ? '+2' : null}
          trendUp={false}
        />
        <StatCard 
          title={t('debts.paid')} 
          value={`₪${(totalDebts - overdueAmount - dueSoonAmount).toLocaleString()}`} 
          icon={TrendingUp}
          subtitle="المدفوع"
          variant="success"
          trend="+15%"
          trendUp={true}
        />
      </div>

      {/* Overdue Debts Alert - Integration with Debt Worker - Futuristic + Attention-focused */}
      {overdueDebts.length > 0 && (
        <Card variant="warning">
          <CardContent className="p-lg">
            <div className="flex items-center gap-md">
              <AlertTriangle className="w-5 h-5 text-yellow" />
              <div className="flex-1">
                <p className="text-small font-semibold text-text">
                  تنبيه: لديك {overdueDebts.length} ديون متأخرة
                </p>
                <p className="text-tiny text-text-muted">
                  إجمالي المبلغ المتأخر: ₪{overdueAmount.toLocaleString()}
                </p>
              </div>
              <Button variant="primary" size="sm" className="gap-2">
                <Zap className="w-4 h-4" />
                اتخاذ إجراء
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Search - Futuristic + Attention-focused */}
      <Card>
        <CardContent className="p-lg">
          <div className="relative">
            <Search className="absolute inset-y-0 end-3 w-4 h-4 text-cyan" />
            <Input
              placeholder="بحث عن عميل..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pe-10"
            />
          </div>
        </CardContent>
      </Card>

      {/* Debts Table - Futuristic + Attention-focused */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <DollarSign className="w-5 h-5 text-cyan" />
            قائمة الديون
          </CardTitle>
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
                  <TableHead>العميل</TableHead>
                  <TableHead>المبلغ الكلي</TableHead>
                  <TableHead>المدفوع</TableHead>
                  <TableHead>المتبقي</TableHead>
                  <TableHead>تاريخ الاستحقاق</TableHead>
                  <TableHead>أيام التأخير</TableHead>
                  <TableHead>فئة التقادم</TableHead>
                  <TableHead>الحالة</TableHead>
                  <TableHead className="text-start">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredDebts.map((debt: any) => {
                  const daysOverdue = getDaysOverdue(debt.dueDate);
                  const isOverdue = daysOverdue > 0;
                  const agingCategory = getAgingCategory(daysOverdue);
                  
                  return (
                    <TableRow key={debt.id}>
                      <TableCell className="font-medium text-text">
                        {debt.customer?.name}
                      </TableCell>
                      <TableCell className="text-small text-text">₪{debt.amount.toLocaleString()}</TableCell>
                      <TableCell className="text-small text-green">
                        ₪{debt.paidAmount.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-small font-bold text-text">
                        ₪{debt.remainingAmount.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-small text-text-muted">
                        {new Date(debt.dueDate).toLocaleDateString('ar-SA')}
                      </TableCell>
                      <TableCell>
                        {isOverdue ? (
                          <span className="text-tiny text-red font-medium">
                            {daysOverdue} يوم
                          </span>
                        ) : (
                          <span className="text-tiny text-text-muted">
                            {Math.abs(daysOverdue)} يوم
                          </span>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge variant={agingCategory.variant} size="sm">
                          {agingCategory.label}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant={
                          debt.status === 'paid' ? 'success' :
                          debt.status === 'overdue' ? 'danger' : 'warning'
                        } size="sm">
                          {debt.status === 'paid' ? 'مدفوع' :
                           debt.status === 'overdue' ? 'متأخر' : 'مستحق'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-start">
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
  subtitle?: string;
  variant?: 'default' | 'featured' | 'warning' | 'ai' | 'danger' | 'success' | 'info';
  trend?: string | null;
  trendUp?: boolean | null;
}

function StatCard({ title, value, icon: Icon, subtitle, variant = 'default', trend, trendUp }: StatCardProps) {
  return (
    <Card 
      variant={variant} 
      className="hover:border-border/22 hover:-translate-y-1 cursor-pointer"
      hoverable
    >
      <CardContent className="p-lg">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <p className="text-small text-text-muted">{title}</p>
            <p className="text-metric font-bold text-text mt-1">
              {value}
            </p>
            {subtitle && (
              <p className="text-tiny text-text-muted mt-1">
                {subtitle}
              </p>
            )}
            {trend && (
              <div className="flex items-center gap-1 mt-2">
                {trendUp ? (
                  <TrendingUp className="w-3 h-3 text-green" />
                ) : (
                  <TrendingDown className="w-3 h-3 text-red" />
                )}
                <span className={`text-tiny ${trendUp ? 'text-green' : 'text-red'}`}>
                  {trend}
                </span>
              </div>
            )}
          </div>
          <div className={`w-10 h-10 rounded-sm flex items-center justify-center ${
            variant === 'featured' ? 'bg-cyan/10' :
            variant === 'warning' ? 'bg-yellow/10' :
            variant === 'danger' ? 'bg-red/10' :
            variant === 'success' ? 'bg-green/10' :
            variant === 'info' ? 'bg-cyan/10' :
            variant === 'ai' ? 'bg-cyan/10' :
            'bg-cyan/10'
          }`}>
            <Icon className={`w-5 h-5 ${
              variant === 'featured' ? 'text-cyan' :
              variant === 'warning' ? 'text-yellow' :
              variant === 'danger' ? 'text-red' :
              variant === 'success' ? 'text-green' :
              variant === 'info' ? 'text-cyan' :
              variant === 'ai' ? 'text-cyan' :
              'text-cyan'
            }`} />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}