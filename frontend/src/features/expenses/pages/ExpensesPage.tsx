import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { expensesApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { 
  DollarSign, 
  Search, 
  Plus, 
  Filter,
  Edit,
  Trash2,
  Calendar,
  TrendingUp
} from 'lucide-react';

export function ExpensesPage() {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');

  const { data: expensesData, isLoading } = useQuery({
    queryKey: ['expenses'],
    queryFn: () => expensesApi.list(),
  });

  const expenses = (expensesData?.data as any[]) || [];

  const filteredExpenses = expenses.filter((expense: any) => {
    const matchesSearch = 
      expense.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      expense.category.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = !categoryFilter || expense.category === categoryFilter;
    return matchesSearch && matchesCategory;
  });

  const thisMonthExpenses = expenses.filter((e: any) => {
    const expenseDate = new Date(e.date);
    const now = new Date();
    return expenseDate.getMonth() === now.getMonth() && 
           expenseDate.getFullYear() === now.getFullYear();
  });

  const thisMonthTotal = thisMonthExpenses.reduce((sum: number, e: any) => sum + e.amount, 0);

  const categoryTotals = expenses.reduce((acc: Record<string, number>, expense: any) => {
    acc[expense.category] = (acc[expense.category] || 0) + expense.amount;
    return acc;
  }, {});

  const categories = [
    { value: '', label: 'كل الفئات' },
    { value: 'rent', label: 'الإيجار' },
    { value: 'salaries', label: 'الرواتب' },
    { value: 'utilities', label: 'المرافق' },
    { value: 'supplies', label: 'المستلزمات' },
    { value: 'maintenance', label: 'الصيانة' },
    { value: 'marketing', label: 'التسويق' },
    { value: 'shipping', label: 'الشحن' },
    { value: 'other', label: 'أخرى' },
  ];

  const getCategoryLabel = (category: string) => {
    const cat = categories.find(c => c.value === category);
    return cat?.label || category;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {t('expenses.title')}
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            إدارة المصروفات والميزانية
          </p>
        </div>
        <Button className="gap-2">
          <Plus className="w-4 h-4" />
          {t('expenses.addExpense')}
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard 
          title={t('expenses.thisMonth')} 
          value={`₪${thisMonthTotal.toLocaleString()}`} 
          icon={Calendar} 
        />
        <StatCard 
          title="الإيجار" 
          value={`₪${(categoryTotals.rent || 0).toLocaleString()}`} 
          icon={DollarSign} 
        />
        <StatCard 
          title="الرواتب" 
          value={`₪${(categoryTotals.salaries || 0).toLocaleString()}`} 
          icon={DollarSign} 
        />
        <StatCard 
          title="المرافق" 
          value={`₪${(categoryTotals.utilities || 0).toLocaleString()}`} 
          icon={DollarSign} 
        />
      </div>

      {/* Category Breakdown */}
      <Card>
        <CardHeader>
          <CardTitle>توزيع المصروفات حسب الفئة</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {Object.entries(categoryTotals).map(([category, total]) => (
              <div key={category} className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  {getCategoryLabel(category)}
                </p>
                <p className="text-xl font-bold text-gray-900 dark:text-gray-100 mt-1">
                  ₪{(total as number).toLocaleString()}
                </p>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

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
                value={categoryFilter}
                onChange={(e) => setCategoryFilter(e.target.value)}
                options={categories}
              />
              <Button variant="outline" className="gap-2">
                <Filter className="w-4 h-4" />
                {t('common.filter')}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Expenses Table */}
      <Card>
        <CardHeader>
          <CardTitle>سجل المصروفات</CardTitle>
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
                  <TableHead>التاريخ</TableHead>
                  <TableHead>الفئة</TableHead>
                  <TableHead>الوصف</TableHead>
                  <TableHead>المبلغ</TableHead>
                  <TableHead>متكرر</TableHead>
                  <TableHead>الإيصال</TableHead>
                  <TableHead className="text-left">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredExpenses.map((expense: any) => (
                  <TableRow key={expense.id}>
                    <TableCell>
                      {new Date(expense.date).toLocaleDateString('ar-SA')}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {getCategoryLabel(expense.category)}
                      </Badge>
                    </TableCell>
                    <TableCell>{expense.description}</TableCell>
                    <TableCell className="font-bold">
                      ₪{expense.amount.toLocaleString()}
                    </TableCell>
                    <TableCell>
                      {expense.recurring ? (
                        <Badge variant="secondary">
                          {expense.recurringPeriod === 'monthly' ? 'شهري' :
                           expense.recurringPeriod === 'weekly' ? 'أسبوعي' :
                           expense.recurringPeriod === 'yearly' ? 'سنوي' : 'نعم'}
                        </Badge>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {expense.receipt ? (
                        <Button variant="ghost" size="sm">
                          عرض
                        </Button>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </TableCell>
                    <TableCell className="text-left">
                      <div className="flex gap-2">
                        <Button variant="ghost" size="sm">
                          <Edit className="w-4 h-4" />
                        </Button>
                        <Button variant="ghost" size="sm">
                          <Trash2 className="w-4 h-4 text-red-500" />
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
  value: string;
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