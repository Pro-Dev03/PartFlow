import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { expenseCategoriesApi, expensesApi } from '../../../services/api/endpoints';
import type { ExpenseCategory, ExpenseCategoryCreateRequest } from '../../../services/api/types';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { PageHeader } from '../../../design-system/components/page-header';
import { StatCard } from '../../../design-system/components/stat-card';
import { Select } from '../../../design-system/components/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../design-system/components/table';
import { PaginationControls } from '../../../design-system/components/pagination-controls';
import { Badge } from '../../../design-system/components/badge';
import { EmptyState } from '../../../design-system/components/empty-state';
import { Modal } from '../../../design-system/components/modal';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { ReportActions } from '../../../design-system/components/report-actions';
import { formatStoreDate, getStoreDateKey, getStoreToday } from '../../../utils/store-time';
import { 
  Search, 
  Plus, 
  Filter,
  Edit,
  Trash2,
  CheckCircle2,
  Calendar,
  ReceiptText,
  Inbox
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
  const dateKey = getExpenseDateKey(value);
  return dateKey ? formatStoreDate(dateKey, 'en-GB') : '-';
}

function getExpenseDateKey(value: string) {
  const normalized = String(value ?? '').trim();
  const match = normalized.match(/^(\d{4})-(\d{2})-(\d{2})(?:$|[T\s])/);
  if (!match) return null;

  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const leapYear = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
  const daysInMonth = [31, leapYear ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
  if (month < 1 || month > 12 || day < 1 || day > daysInMonth[month - 1]) return null;

  return getStoreDateKey(normalized);
}

export function isExpenseInCurrentMonth(value: string, referenceDate = new Date()) {
  const datePart = getExpenseDateKey(value)?.slice(0, 7);
  if (!datePart) return false;
  const currentMonth = getStoreDateKey(referenceDate)?.slice(0, 7) || '';
  return datePart === currentMonth;
}

function getExpenseCategoryErrorMessage(error: unknown) {
  const message = error instanceof Error ? error.message : '';
  if (message.toLowerCase().includes('expense category already exists')) {
    return 'فئة المصروف موجودة بالفعل';
  }
  return message || 'تعذر حفظ الفئة';
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
  const [page, setPage] = useState(1);
  const pageSize = 10;
  const [newExpense, setNewExpense] = useState({
    description: '',
    amount: '',
    category: '',
    date: getStoreToday(),
    recurring: false,
    recurringPeriod: 'monthly',
  });

  const { data: expensesData, isLoading } = useQuery({
    queryKey: ['expenses', page, pageSize],
    queryFn: () => expensesApi.list({ page, per_page: pageSize }),
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
      setNewExpense({ description: '', amount: '', category: '', date: getStoreToday(), recurring: false, recurringPeriod: 'monthly' });
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

  const nestedExpenseRows = (expensesData?.data as { data?: unknown } | undefined)?.data;
  const expenseRows = Array.isArray(expensesData)
    ? expensesData
    : Array.isArray(expensesData?.data)
    ? expensesData.data
    : Array.isArray(nestedExpenseRows)
      ? nestedExpenseRows
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

  const totalExpenses = Number(expensesData?.meta?.total || expensesData?.data?.total || expenses.length);

  useEffect(() => {
    setPage(1);
  }, [searchQuery, categoryFilter]);

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

  const getExpenseReportRows = (expenseRows: any[]) => expenseRows.map((expense: any) => ({
      'التاريخ': expense.date,
      'الوصف': expense.description,
      'الفئة': getCategoryLabel(expense.category),
      'المبلغ': expense.amount
    }));

  const handleExport = () => {
    exportToCSV(getExpenseReportRows(filteredExpenses), `expenses-${getStoreToday()}`);
  };

  const handlePrint = () => {
    printTable(getExpenseReportRows(filteredExpenses), ['التاريخ', 'الوصف', 'الفئة', 'المبلغ'], 'تقرير المصروفات');
  };

  const loadAllExpenses = async () => {
    const response = await expensesApi.list({ page: 1, per_page: 1000, ...(searchQuery ? { search: searchQuery } : {}) });
    const allExpenses = (((response as any)?.data ?? []) as any[]).map(normalizeExpenseForDisplay);
    return allExpenses.filter((expense) => !categoryFilter || expense.category_id === categoryFilter || expense.category === categoryFilter);
  };

  const handleExportAll = async () => {
    exportToCSV(getExpenseReportRows(await loadAllExpenses()), `expenses-all-${getStoreToday()}`);
  };

  const handlePrintAll = async () => {
    printTable(getExpenseReportRows(await loadAllExpenses()), ['التاريخ', 'الوصف', 'الفئة', 'المبلغ'], 'تقرير كل المصروفات');
  };

  const handleEdit = (expense: any) => {
    setEditingExpenseId(expense.id);
    setNewExpense({
      description: expense.description || '',
      amount: String(expense.amount ?? ''),
      category: expense.category_id || '',
      date: getStoreDateKey(expense.date) || getStoreToday(),
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
        title={t('expenses.title')}
        description="إدارة المصروفات والميزانية"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="primary" className="gap-2" onClick={() => {
              setEditingExpenseId(null);
              setNewExpense({ description: '', amount: '', category: expenseCategories[0]?.id || '', date: getStoreToday(), recurring: false, recurringPeriod: 'monthly' });
              setIsAddModalOpen(true);
            }}>
              <Plus className="w-4 h-4" />
              {t('expenses.addExpense')}
            </Button>
            <ReportActions onExportCurrent={handleExport} onPrintCurrent={handlePrint} onExportAll={() => { void handleExportAll(); }} onPrintAll={() => { void handlePrintAll(); }} />
          </div>
        }
      />

{/* Stats Cards - Futuristic + Clean */}
      <div className="unified-stats-grid supplier-stats grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-md">
        <StatCard 
          title={t('expenses.thisMonth')} 
          value={`₪${thisMonthTotal.toLocaleString()}`} 
          icon={Calendar}
          variant="featured"
          size="sm"
        />
      </div>

      {/* Category Breakdown */}
      <Card className="expense-category-card">
        <CardHeader className="px-4 py-3">
          <CardTitle>توزيع المصروفات حسب الفئة</CardTitle>
        </CardHeader>
        <CardContent className="px-4 py-3">
          <div className="flex flex-wrap items-center gap-2">
            {Object.entries(categoryTotals).map(([category, total]) => (
              <div key={category} className="flex min-w-[180px] flex-1 items-center justify-between gap-3 rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-surface-elevated)] px-3 py-2.5 shadow-[0_4px_14px_rgba(15,23,42,0.04)] transition-all duration-200 hover:-translate-y-0.5 hover:border-[var(--color-primary-25)] hover:shadow-[0_8px_20px_rgba(15,23,42,0.08)]">
                <p className="text-small text-text-muted">
                  {getCategoryLabel(category)}
                </p>
                <p className="text-sm font-bold text-text">
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
        size="lg"
        headerStyle={{ background: 'var(--bg-surface-2)', borderBottom: '1px solid var(--border-subtle)' }}
      >
        <form
          className="space-y-4"
          style={{
            background: 'transparent',
            padding: '24px',
            width: '100%',
            maxWidth: '100%',
            boxSizing: 'border-box',
          }}
          onSubmit={(event) => {
            event.preventDefault();
            const amount = Number(newExpense.amount);
            if (!newExpense.description.trim() || !Number.isInteger(amount) || amount <= 0) return;
            if (editingExpenseId) {
              updateExpenseMutation.mutate();
            } else {
              createExpenseMutation.mutate();
            }
          }}
        >
          <div style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}>
            <Input
              label="وصف المصروف"
              fullWidth
              value={newExpense.description}
              onChange={(event) => setNewExpense((current) => ({ ...current, description: event.target.value }))}
              placeholder="مثال: فاتورة كهرباء المحل"
              className="w-full min-w-0"
              style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}
              required
            />
          </div>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2" style={{ minWidth: 0 }}>
            <div className="sm:col-span-2" style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}>
              <Input
                label="المبلغ"
                fullWidth
                type="number"
                min="1"
                step="1"
                value={newExpense.amount}
                onChange={(event) => setNewExpense((current) => ({ ...current, amount: event.target.value }))}
                placeholder="0"
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
                fullWidth
                value={newExpense.category}
                onChange={(event) => setNewExpense((current) => ({ ...current, category: event.target.value }))}
                options={expenseCategories.map((category) => ({ value: category.id, label: category.name }))}
                disabled={isLoadingExpenseCategories || expenseCategories.length === 0}
                className="w-full min-w-0"
                style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}
              />
            </div>
            <div style={{ width: '100%', maxWidth: '100%', minWidth: 0, boxSizing: 'border-box' }}>
              <Input
                label="التاريخ"
                fullWidth
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
          <div className="flex justify-end gap-3 border-t border-border pt-4">
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
              {getExpenseCategoryErrorMessage(createCategoryMutation.error)}
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
      <Card className="rounded-[16px] border-[var(--border-default)] shadow-[0_10px_28px_rgba(15,23,42,0.05)]">
        <CardHeader className="border-b border-[var(--border-subtle)] px-5 py-4">
          <div className="flex items-center gap-3">
            <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
              <ReceiptText className="h-4 w-4" />
            </span>
            <div>
              <CardTitle className="text-sm font-extrabold text-[var(--text-primary)]">سجل المصروفات</CardTitle>
              <p className="mt-0.5 text-[11px] font-medium text-[var(--text-muted)]">متابعة المصروفات وحالات اعتمادها</p>
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="flex h-64 items-center justify-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan" />
            </div>
          ) : filteredExpenses.length === 0 ? (
            <EmptyState
              icon={<Inbox className="h-5 w-5" />}
              title="لا توجد مصروفات"
              description={searchQuery || categoryFilter ? 'لم يتم العثور على مصروفات مطابقة للفلاتر الحالية' : 'أضف أول مصروف لبدء متابعة ميزانية المتجر'}
              size="sm"
            />
          ) : (
            <>
              <div className="overflow-x-auto">
              <Table className="min-w-[760px]">
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
              <TableBody className="divide-y divide-[var(--border-subtle)]" dir="rtl">
                {filteredExpenses.map((expense: any) => (
                  <TableRow key={expense.id} dir="rtl" className="transition-colors duration-200 hover:bg-[var(--color-primary-05)] [&>td]:h-[68px]">
                    <TableCell className="font-semibold text-[var(--text-secondary)]">
                      {formatExpenseDate(expense.date)}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="rounded-full">
                        {getCategoryLabel(expense.category)}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-bold text-[var(--text-primary)]">{expense.description || '-'}</TableCell>
                    <TableCell className="font-black text-[var(--primary)]">
                      ₪{Number(expense.amount || 0).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <Badge variant={expense.status === 'approved' ? 'success' : expense.status === 'rejected' ? 'danger' : 'warning'} className="rounded-full">
                        {expense.status === 'approved' ? 'معتمد' : expense.status === 'rejected' ? 'مرفوض' : 'قيد الانتظار'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {expense.recurring ? (
                        <Badge variant="secondary" className="rounded-full">
                          {expense.recurringPeriod === 'monthly' ? 'شهري' :
                           expense.recurringPeriod === 'weekly' ? 'أسبوعي' :
                           expense.recurringPeriod === 'yearly' ? 'سنوي' : 'نعم'}
                        </Badge>
                      ) : (
                        <span className="text-[var(--text-muted)]">-</span>
                      )}
                    </TableCell>
                    <TableCell className="text-start">
                      <div className="flex gap-sm">
                        {expense.status === 'pending' && (
                          <Button
                            variant="ghost"
                            size="sm"
                            tableAction
                            onClick={() => approveExpenseMutation.mutate(expense.id)}
                            disabled={approveExpenseMutation.isPending}
                            aria-label="اعتماد المصروف"
                            title="اعتماد المصروف"
                          >
                            <CheckCircle2 className="w-4 h-4 text-green-500" />
                          </Button>
                        )}
                        <Button variant="ghost" size="sm" tableAction onClick={() => handleEdit(expense)} aria-label="تعديل المصروف">
                          <Edit className="w-4 h-4" />
                        </Button>
                        <Button variant="ghost" size="sm" tableAction onClick={() => handleDelete(expense)} disabled={deleteExpenseMutation.isPending} aria-label="حذف المصروف">
                          <Trash2 className="w-4 h-4 text-red" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
              </Table>
              </div>
              <PaginationControls page={page} pageSize={pageSize} total={totalExpenses} onPageChange={setPage} isLoading={isLoading} />
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
