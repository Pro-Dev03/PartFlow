import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { expensesApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { StatCard } from '../../../components/ui/stat-card';
import { Select } from '../../../components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { 
  DollarSign, 
  Search, 
  Plus, 
  Filter,
  Edit,
  Trash2,
  Calendar,
  Download,
  Printer
} from 'lucide-react';

export function ExpensesPage() {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');

  const { data: expensesData, isLoading } = useQuery({
    queryKey: ['expenses'],
    queryFn: () => expensesApi.list({ page: 1, per_page: 100 }),
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

  const handleExport = () => {
    const dataToExport = filteredExpenses.map((expense: any) => ({
      'التاريخ': expense.date,
      'الوصف': expense.description,
      'الفئة': getCategoryLabel(expense.category),
      'المبلغ': expense.amount
    }));
    exportToCSV(dataToExport, `expenses-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    const dataToPrint = filteredExpenses.map((expense: any) => ({
      'التاريخ': expense.date,
      'الوصف': expense.description,
      'الفئة': getCategoryLabel(expense.category),
      'المبلغ': expense.amount
    }));
    printTable(dataToPrint, ['التاريخ', 'الوصف', 'الفئة', 'المبلغ'], 'تقرير المصروفات');
  };

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Expense Management"
        title={t('expenses.title')}
        description="إدارة المصروفات والميزانية"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="primary" className="gap-2">
              <Plus className="w-4 h-4" />
              {t('expenses.addExpense')}
            </Button>
            <Button variant="secondary" onClick={handleExport} className="gap-2">
              <Download className="w-4 h-4" />
              تصدير
            </Button>
            <Button variant="secondary" onClick={handlePrint} className="gap-2">
              <Printer className="w-4 h-4" />
              طباعة
            </Button>
          </div>
        }
      />

      {/* Stats Cards - Futuristic + Clean */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-md">
        <StatCard 
          title={t('expenses.thisMonth')} 
          value={`₪${thisMonthTotal.toLocaleString()}`} 
          icon={Calendar}
          variant="featured"
        />
        <StatCard 
          title="الإيجار" 
          value={`₪${(categoryTotals.rent || 0).toLocaleString()}`} 
          icon={DollarSign}
          variant="default"
        />
        <StatCard 
          title="الرواتب" 
          value={`₪${(categoryTotals.salaries || 0).toLocaleString()}`} 
          icon={DollarSign}
          variant="default"
        />
        <StatCard 
          title="المرافق" 
          value={`₪${(categoryTotals.utilities || 0).toLocaleString()}`} 
          icon={DollarSign}
          variant="default"
        />
      </div>

      {/* Category Breakdown */}
      <Card>
        <CardHeader>
          <CardTitle>توزيع المصروفات حسب الفئة</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-md">
            {Object.entries(categoryTotals).map(([category, total]) => (
              <div key={category} className="p-lg bg-surface-2 rounded-sm">
                <p className="text-small text-text-muted">
                  {getCategoryLabel(category)}
                </p>
                <p className="text-metric font-bold text-text mt-1">
                  ₪{(total as number).toLocaleString()}
                </p>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Search and Filters */}
      <Card>
        <CardContent className="p-lg">
          <div className="flex flex-col md:flex-row gap-md">
            <div className="flex-1 relative">
              <Search className="absolute inset-y-0 end-3 w-4 h-4 text-cyan" />
              <Input
                placeholder={t('common.search')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pe-10"
              />
            </div>
            <div className="flex gap-sm">
              <Select
                value={categoryFilter}
                onChange={(e) => setCategoryFilter(e.target.value)}
                options={categories}
              />
              <Button variant="secondary" className="gap-2">
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
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan" />
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
                  <TableHead className="text-start">الإجراءات</TableHead>
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
                        <span className="text-text-muted">-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {expense.receipt ? (
                        <Button variant="ghost" size="sm">
                          عرض
                        </Button>
                      ) : (
                        <span className="text-text-muted">-</span>
                      )}
                    </TableCell>
                    <TableCell className="text-start">
                      <div className="flex gap-sm">
                        <Button variant="ghost" size="sm">
                          <Edit className="w-4 h-4" />
                        </Button>
                        <Button variant="ghost" size="sm">
                          <Trash2 className="w-4 h-4 text-red" />
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