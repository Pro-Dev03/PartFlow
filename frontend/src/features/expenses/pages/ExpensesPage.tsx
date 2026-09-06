import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { expenseCategoriesApi, expensesApi } from '../../../services/api/endpoints';
import type { ExpenseCategory, ExpenseCategoryCreateRequest } from '../../../services/api/types';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { StatCard } from '../../../components/ui/stat-card';
import { Select } from '../../../components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { Modal } from '../../../components/ui/modal';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { 
  DollarSign, 
  Search, 
  Plus, 
  Filter,
  Edit,
  Trash2,
  CheckCircle2,
  Calendar,
  Download,
  Printer
} from 'lucide-react';

export function normalizeExpenseForDisplay(expense: any) {
  const description = (expense?.description || expense?.title || '').toString();
  const category = (expense?.category || expense?.category_name || expense?.categoryName || '').toString();
  const date = expense?.date || expense?.expense_date || '';
  const amount = Number(expense?.amount ?? 0);
  const recurring = typeof expense?.recurring === 'boolean'
    ? expense.recurring
    : Boolean(expense?.is_recurring);
  const recurringPeriod = expense?.recurringPeriod || expense?.recurring_period || '';
  return {
    ...expense,
    description,
    category,
    date,
    amount,
    recurring,
    recurringPeriod,
  };
}

export function formatExpenseDate(value: string) {
  const datePart = value?.slice(0, 10);
  if (!datePart || !/^\d{4}-\d{2}-\d{2}$/.test(datePart)) return '-';
  const [year, month, day] = datePart.split('-');
  return `${day}/${month}/${year}`;
}

export function isExpenseInCurrentMonth(value: string, referenceDate = new Date()) {
  const datePart = value?.slice(0, 7);
  if (!datePart || !/^\d{4}-\d{2}$/.test(datePart)) return false;
  const currentMonth = `${referenceDate.getFullYear()}-${String(referenceDate.getMonth() + 1).padStart(2, '0')}`;
  return datePart === currentMonth;
}

export function ExpensesPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [editingExpenseId, setEditingExpenseId] = useState<string | null>(null);
  const [expenseToDelete, setExpenseToDelete] = useState<any | null>(null);
  const [isAddCategoryModalOpen, setIsAddCategoryModalOpen] = useState(false);
  const [newCategoryName, setNewCategoryName] = useState('');
  const [newExpense, setNewExpense] = useState({
    description: '',
    amount: '',
    category: '',
    date: new Date().toISOString().split('T')[0],
    recurring: false,
    recurringPeriod: 'monthly',
  });

  const { data: expensesData, isLoading } = useQuery({
    queryKey: ['expenses'],
    queryFn: () => expensesApi.list({ page: 1, per_page: 100 }),
  });

  const { data: expenseCategoriesData, isLoading: isLoadingExpenseCategories } = useQuery({
    queryKey: ['expense-categories'],
    queryFn: () => expenseCategoriesApi.list({ page: 1, per_page: 100, is_active: true }),
  });

  const expenseCategories = (expenseCategoriesData?.data as ExpenseCategory[] | undefined) || [];

  const createCategoryMutation = useMutation({
    mutationFn: (data: ExpenseCategoryCreateRequest) => expenseCategoriesApi.create(data),
    onSuccess: async (response) => {
      const createdCategory = (response.data ?? response) as ExpenseCategory;
      await queryClient.invalidateQueries({ queryKey: ['expense-categories'] });
      if (createdCategory?.id) {
        setNewExpense((current) => ({ ...current, category: createdCategory.id }));
      }
      setNewCategoryName('');
      setIsAddCategoryModalOpen(false);
    },
  });

  useEffect(() => {
    const firstCategory = expenseCategories[0];
    if (!newExpense.category && firstCategory?.id) {
      setNewExpense((current) => ({ ...current, category: firstCategory.id }));
    }
  }, [expenseCategories, newExpense.category]);

  const createExpenseMutation = useMutation({
    mutationFn: () => expensesApi.create({
      category_id: newExpense.category,
      title: newExpense.description.trim(),
      description: newExpense.description.trim(),
      amount: Number(newExpense.amount),
      currency: 'ILS',
      expense_date: `${newExpense.date}T12:00:00Z`,
      payment_method: 'cash',
      is_recurring: newExpense.recurring,
      recurring_period: newExpense.recurring ? newExpense.recurringPeriod as 'daily' | 'weekly' | 'monthly' | 'yearly' : undefined,
    }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['expenses'] });
      void queryClient.invalidateQueries({ queryKey: ['reports'] });
      void queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      setIsAddModalOpen(false);
      setEditingExpenseId(null);
      setNewExpense({ description: '', amount: '', category: '', date: new Date().toISOString().split('T')[0], recurring: false, recurringPeriod: 'monthly' });
    },
  });

  const updateExpenseMutation = useMutation({
    mutationFn: () => expensesApi.update(editingExpenseId as string, {
      category_id: newExpense.category,
      title: newExpense.description.trim(),
      description: newExpense.description.trim(),
      amount: Number(newExpense.amount),
      currency: 'ILS',
      expense_date: `${newExpense.date}T12:00:00Z`,
      payment_method: 'cash',
      is_recurring: newExpense.recurring,
      recurring_period: newExpense.recurring ? newExpense.recurringPeriod as 'daily' | 'weekly' | 'monthly' | 'yearly' : undefined,
    }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['expenses'] });
      void queryClient.invalidateQueries({ queryKey: ['reports'] });
      setIsAddModalOpen(false);
      setEditingExpenseId(null);
    },
  });

  const deleteExpenseMutation = useMutation({
    mutationFn: (id: string) => expensesApi.delete(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['expenses'] });
      void queryClient.invalidateQueries({ queryKey: ['reports'] });
      void queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  const approveExpenseMutation = useMutation({
    mutationFn: (id: string) => expensesApi.approve(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['expenses'] });
      void queryClient.invalidateQueries({ queryKey: ['reports'] });
      void queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  const expenseRows = Array.isArray(expensesData)
    ? expensesData
    : Array.isArray(expensesData?.data)
    ? expensesData.data
    : Array.isArray((expensesData?.data as { data?: unknown } | undefined)?.data)
      ? (expensesData?.data as { data: any[] }).data
      : [];
  const expenses = expenseRows.map(normalizeExpenseForDisplay);

  const filteredExpenses = expenses.filter((expense: any) => {
    const searchText = `${expense.description || ''} ${expense.category || ''}`.toLowerCase();
    const matchesSearch = searchText.includes(searchQuery.toLowerCase());
    const matchesCategory = !categoryFilter
      || expense.category_id === categoryFilter
      || expense.category === categoryFilter;
    return matchesSearch && matchesCategory;
  });

  const thisMonthExpenses = expenses.filter((e: any) => isExpenseInCurrentMonth(e.date));

  const thisMonthTotal = thisMonthExpenses.reduce((sum: number, e: any) => sum + Number(e.amount || 0), 0);

  const categoryTotals = expenses.reduce((acc: Record<string, number>, expense: any) => {
    const categoryKey = String(expense.category || 'other');
    acc[categoryKey] = (acc[categoryKey] || 0) + Number(expense.amount || 0);
    return acc;
  }, {});

  const categories = [
    { value: '', label: 'كل الفئات' },
    ...expenseCategories.map((category) => ({ value: category.id, label: category.name })),
  ];

  const getCategoryLabel = (category: string) => {
    const cat = expenseCategories.find(c => c.id === category || c.name === category);
    return cat?.name || category;
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

  const handleEdit = (expense: any) => {
    setEditingExpenseId(expense.id);
    setNewExpense({
      description: expense.description || '',
      amount: String(expense.amount ?? ''),
      category: expense.category_id || '',
      date: expense.date?.slice(0, 10) || new Date().toISOString().split('T')[0],
      recurring: Boolean(expense.recurring),
      recurringPeriod: expense.recurringPeriod || 'monthly',
    });
    setIsAddModalOpen(true);
  };

  const handleDelete = (expense: any) => {
    setExpenseToDelete(expense);
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
            <Button variant="primary" className="gap-2" onClick={() => {
              setEditingExpenseId(null);
              setNewExpense({ description: '', amount: '', category: expenseCategories[0]?.id || '', date: new Date().toISOString().split('T')[0], recurring: false, recurringPeriod: 'monthly' });
              setIsAddModalOpen(true);
            }}>
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
      <Modal
        isOpen={isAddModalOpen}
        onClose={() => setIsAddModalOpen(false)}
        title={editingExpenseId ? 'تعديل مصروف' : 'إضافة مصروف'}
        size="xl"
        headerStyle={{ background: 'var(--bg-surface-2)' }}
        style={{ overflowX: 'hidden' }}
      >
        <form
          className="space-y-4"
          style={{
            background: 'var(--bg-surface-2)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '14px',
            padding: '20px',
            width: '100%',
            maxWidth: '100%',
            boxSizing: 'border-box',
          }}
          onSubmit={(event) => {
            event.preventDefault();
            const amount = Number(newExpense.amount);
            if (!newExpense.description.trim() || !Number.isFinite(amount) || amount <= 0) return;
            if (editingExpenseId) {
              updateExpenseMutation.mutate();
            } else {
              createExpenseMutation.mutate();
            }
          }}
        >
          <div style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}>
            <label className="mb-2 block text-sm font-medium text-text-secondary">وصف المصروف</label>
            <Input
              value={newExpense.description}
              onChange={(event) => setNewExpense((current) => ({ ...current, description: event.target.value }))}
              placeholder="مثال: فاتورة كهرباء المحل"
              className="w-full min-w-0"
              style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}
              required
            />
          </div>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3" style={{ minWidth: 0 }}>
            <div style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}>
              <label className="mb-2 block text-sm font-medium text-text-secondary">المبلغ</label>
              <Input
                type="number"
                min="0.01"
                step="0.01"
                value={newExpense.amount}
                onChange={(event) => setNewExpense((current) => ({ ...current, amount: event.target.value }))}
                placeholder="0.00"
                className="w-full min-w-0"
                style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}
                required
              />
            </div>
            <div style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}>
              <div className="mb-2 flex items-center justify-between gap-2">
                <label className="block text-sm font-medium text-text-secondary">الفئة</label>
                <Button
                  type="button"
                  variant="secondary"
                  className="gap-1 px-2 py-1 text-xs"
                  onClick={() => setIsAddCategoryModalOpen(true)}
                >
                  <Plus className="h-3 w-3" />
                  إضافة فئة
                </Button>
              </div>
              <Select
                value={newExpense.category}
                onChange={(event) => setNewExpense((current) => ({ ...current, category: event.target.value }))}
                options={expenseCategories.map((category) => ({ value: category.id, label: category.name }))}
                disabled={isLoadingExpenseCategories || expenseCategories.length === 0}
                className="w-full min-w-0"
                style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}
              />
            </div>
            <div style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}>
              <label className="mb-2 block text-sm font-medium text-text-secondary">التاريخ</label>
              <Input
                type="date"
                value={newExpense.date}
                onChange={(event) => setNewExpense((current) => ({ ...current, date: event.target.value }))}
                dir="ltr"
                className="w-full min-w-0"
                style={{ direction: 'ltr', textAlign: 'left', width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}
                inputMode="numeric"
                lang="en-CA"
                aria-label="تاريخ المصروف بصيغة YYYY-MM-DD"
                required
              />
            </div>
          </div>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2" style={{ minWidth: 0 }}>
            <label className="flex items-center gap-2 text-sm font-medium text-text-secondary">
              <input
                type="checkbox"
                checked={newExpense.recurring}
                onChange={(event) => setNewExpense((current) => ({ ...current, recurring: event.target.checked }))}
              />
              مصروف متكرر
            </label>
            {newExpense.recurring && (
              <Select
                value={newExpense.recurringPeriod}
                onChange={(event) => setNewExpense((current) => ({ ...current, recurringPeriod: event.target.value }))}
                options={[
                  { value: 'daily', label: 'يومي' },
                  { value: 'weekly', label: 'أسبوعي' },
                  { value: 'monthly', label: 'شهري' },
                  { value: 'yearly', label: 'سنوي' },
                ]}
              />
            )}
          </div>
          <div className="flex justify-end gap-3">
            <Button type="button" variant="secondary" onClick={() => setIsAddModalOpen(false)}>إلغاء</Button>
            <Button type="submit" variant="primary" disabled={createExpenseMutation.isPending || updateExpenseMutation.isPending}>
              {createExpenseMutation.isPending || updateExpenseMutation.isPending ? 'جاري الحفظ...' : editingExpenseId ? 'حفظ التعديل' : 'حفظ المصروف'}
            </Button>
          </div>
          {(createExpenseMutation.isError || updateExpenseMutation.isError) && (
            <p className="text-sm text-red-500" role="alert">
              {createExpenseMutation.error instanceof Error
                ? createExpenseMutation.error.message
                : updateExpenseMutation.error instanceof Error
                  ? updateExpenseMutation.error.message
                  : 'تعذر حفظ المصروف'}
            </p>
          )}
        </form>
      </Modal>

      <Modal
        isOpen={isAddCategoryModalOpen}
        onClose={() => setIsAddCategoryModalOpen(false)}
        title="إضافة فئة مصروف"
        size="sm"
      >
        <form
          className="space-y-4"
          onSubmit={(event) => {
            event.preventDefault();
            const name = newCategoryName.trim();
            if (!name) return;
            createCategoryMutation.mutate({ name, is_active: true });
          }}
        >
          <div>
            <label className="mb-2 block text-sm font-medium text-text-secondary">اسم الفئة</label>
            <Input
              value={newCategoryName}
              onChange={(event) => setNewCategoryName(event.target.value)}
              placeholder="مثال: صيانة الأجهزة"
              autoFocus
              required
            />
          </div>
          <div className="flex justify-end gap-3">
            <Button type="button" variant="secondary" onClick={() => setIsAddCategoryModalOpen(false)}>
              إلغاء
            </Button>
            <Button type="submit" variant="primary" disabled={createCategoryMutation.isPending}>
              {createCategoryMutation.isPending ? 'جاري الحفظ...' : 'حفظ الفئة'}
            </Button>
          </div>
          {createCategoryMutation.isError && (
            <p className="text-sm text-red-500" role="alert">
              {createCategoryMutation.error instanceof Error ? createCategoryMutation.error.message : 'تعذر حفظ الفئة'}
            </p>
          )}
        </form>
      </Modal>

      <Modal
        isOpen={Boolean(expenseToDelete)}
        onClose={() => setExpenseToDelete(null)}
        title="تأكيد حذف المصروف"
        size="sm"
      >
        <div className="space-y-5">
          <p className="text-sm text-text-secondary">
            هل تريد حذف المصروف «{expenseToDelete?.description || '-'}»؟ لا يمكن التراجع عن هذا الإجراء.
          </p>
          <div className="flex justify-end gap-3">
            <Button type="button" variant="secondary" onClick={() => setExpenseToDelete(null)}>
              إلغاء
            </Button>
            <Button
              type="button"
              variant="danger"
              disabled={deleteExpenseMutation.isPending}
              onClick={() => {
                if (expenseToDelete?.id) {
                  deleteExpenseMutation.mutate(expenseToDelete.id, {
                    onSuccess: () => setExpenseToDelete(null),
                  });
                }
              }}
            >
              {deleteExpenseMutation.isPending ? 'جاري الحذف...' : 'حذف المصروف'}
            </Button>
          </div>
          {deleteExpenseMutation.isError && (
            <p className="text-sm text-red-500" role="alert">
              {deleteExpenseMutation.error instanceof Error ? deleteExpenseMutation.error.message : 'تعذر حذف المصروف'}
            </p>
          )}
        </div>
      </Modal>

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
                  <TableHead>الحالة</TableHead>
                  <TableHead>متكرر</TableHead>
                  <TableHead className="text-start">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredExpenses.map((expense: any) => (
                  <TableRow key={expense.id}>
                    <TableCell>
                      {formatExpenseDate(expense.date)}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {getCategoryLabel(expense.category)}
                      </Badge>
                    </TableCell>
                    <TableCell>{expense.description || '-'}</TableCell>
                    <TableCell className="font-bold">
                      ₪{Number(expense.amount || 0).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <Badge variant={expense.status === 'approved' ? 'success' : expense.status === 'rejected' ? 'danger' : 'warning'}>
                        {expense.status === 'approved' ? 'معتمد' : expense.status === 'rejected' ? 'مرفوض' : 'قيد الانتظار'}
                      </Badge>
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
                    <TableCell className="text-start">
                      <div className="flex gap-sm">
                        {expense.status === 'pending' && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => approveExpenseMutation.mutate(expense.id)}
                            disabled={approveExpenseMutation.isPending}
                            aria-label="اعتماد المصروف"
                            title="اعتماد المصروف"
                          >
                            <CheckCircle2 className="w-4 h-4 text-green-500" />
                          </Button>
                        )}
                        <Button variant="ghost" size="sm" onClick={() => handleEdit(expense)} aria-label="تعديل المصروف">
                          <Edit className="w-4 h-4" />
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => handleDelete(expense)} disabled={deleteExpenseMutation.isPending} aria-label="حذف المصروف">
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