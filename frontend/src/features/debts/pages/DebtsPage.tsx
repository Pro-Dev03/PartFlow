import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useTranslation as useTranslationHook } from '../../../hooks/useTranslation';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { PageHeader } from '../../../design-system/components/page-header';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../design-system/components/table';
import { PaginationControls } from '../../../design-system/components/pagination-controls';
import { Badge } from '../../../design-system/components/badge';
import { getButtonSize } from '../../../config/button-sizes';
import { 
  DollarSign, 
  AlertTriangle,
  Calendar,
  Eye,
  CheckCircle,
  UserRound,
  Printer,
  FileDown,
  ArrowRight,
  History
} from 'lucide-react';
import '../styles/success-modal.css';
import { printPaymentReceipt } from '../../../lib/export-utils';
import { toast } from 'sonner';

// SmartDelete utility (ARCHITECTURE-PRINCIPLES.md)

// Custom hooks
import { useDebts } from '../hooks/useDebts';

// Components
import { DebtStats } from '../components/DebtStats';
import { AdvancedSearch, SearchFilters } from '../components/AdvancedSearch';

// Types
import { Debt } from '../types/debts.types';
import { formatStoreDate, getStoreDateKey, parseBackendTimestamp } from '../../../utils/store-time';

export function DebtsPage() {
  const { t } = useTranslationHook();
  const location = useLocation();
  const navigate = useNavigate();
  const [paymentModalOpen, setPaymentModalOpen] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState<any>(null);
  const [paymentAmount, setPaymentAmount] = useState('');
  const [paymentMethod, setPaymentMethod] = useState('cash');
  const [selectedDebt, setSelectedDebt] = useState<any>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [successModalOpen, setSuccessModalOpen] = useState(false);
  const [lastPayment, setLastPayment] = useState<any>(null);
  const [debtTab, setDebtTab] = useState<'open' | 'paid' | 'all'>('open');

  // Custom hook
  const {
    debts,
    filteredDebts,
    stats,
    isLoading,
    setSearchQuery,
    setSearchFilters,
    recordPaymentMutation,
    page,
    pageSize,
    total,
    setPage,
  } = useDebts();

  const displayedDebts = filteredDebts.filter((debt: any) => {
    const isPaid = Number(debt.remainingAmount ?? debt.remaining_amount ?? 0) <= 0 || debt.status === 'paid';
    return debtTab === 'all' || (debtTab === 'paid' ? isPaid : !isPaid);
  });

  // Get current language for receipt
  const { currentLanguage } = useTranslationHook();

  const handleRecordPayment = (customerId: string, customerName: string) => {
    console.log('handleRecordPayment called', { customerId, customerName });
    const customerDebt = debts.find((debt: any) => debt.customer?.id === customerId);
    setSelectedCustomer({
      id: customerId,
      name: customerName,
      outstanding: customerDebt?.remainingAmount || 0,
    });
    setPaymentAmount('');
    setPaymentMethod('cash');
    setPaymentModalOpen(true);
  };

  useEffect(() => {
    const customerId = new URLSearchParams(location.search).get('customer_id');
    if (!customerId || isLoading || paymentModalOpen) return;

    const customerDebt = debts.find((debt: any) => debt.customer?.id === customerId);
    if (customerDebt) {
      handleRecordPayment(customerId, customerDebt.customer.name);
      navigate('/app/debts', { replace: true });
    }
  }, [debts, isLoading, location.search, navigate, paymentModalOpen]);

  const handlePaymentSubmit = () => {
    console.log('handlePaymentSubmit called', { paymentAmount, selectedCustomer, paymentMethod });
    if (paymentAmount && !isNaN(parseFloat(paymentAmount))) {
      const paymentAmountNum = parseFloat(paymentAmount);
      
      // Validate payment doesn't exceed outstanding balance (SALES-PHILOSOPHY.md)
      const outstandingAmount = selectedCustomer?.outstanding || 0;
      
      if (paymentAmountNum > outstandingAmount) {
        toast.error(`المبلغ أكبر من المبلغ المستحق (₪${outstandingAmount.toLocaleString()})`);
        return;
      }
      
      const paymentData = {
        customerId: selectedCustomer.id,
        amount: paymentAmountNum,
        method: paymentMethod,
      };
      
      recordPaymentMutation.mutate(paymentData, {
        onSuccess: () => {
          const now = new Date();
          const day = String(now.getDate()).padStart(2, '0');
          const month = String(now.getMonth() + 1).padStart(2, '0');
          const year = now.getFullYear();
          
          setLastPayment({
            ...paymentData,
            customerId: selectedCustomer.id,
            customerName: selectedCustomer.name,
            date: `${day}/${month}/${year}`,
          });
          setSuccessModalOpen(true);
          setPaymentModalOpen(false);
        },
      });
    }
  };

  const handleSearchAdvanced = (query: string, filters: SearchFilters) => {
    setSearchQuery(query);
    setSearchFilters(filters);
  };

  const handleViewDebt = (debt: Debt) => {
    console.log('handleViewDebt called', debt);
    setSelectedDebt(debt);
    console.log('Setting isViewModalOpen to true');
    setIsViewModalOpen(true);
    console.log('isViewModalOpen after set:', true);
  };

  const handlePrintReceipt = () => {
    if (!lastPayment) {
      toast.error('لا توجد بيانات للدفعة');
      return;
    }

    // استخدام الدالة الجديدة لطباعة الإيصال
    printPaymentReceipt({
      amount: lastPayment.amount,
      method: lastPayment.method,
      customerName: lastPayment.customerName,
      date: lastPayment.date,
      language: currentLanguage,
    });
  };

  // Debt Aging System - تصنيف ديون حسب العمر
  const getDebtAging = (dueDate: string, status: string) => {
    const due = parseBackendTimestamp(dueDate);
    if (!due) {
      return { category: 'FUTURE', label: 'موعد السداد غير محدد', variant: 'secondary' as const, days: 0 };
    }

    const dueKey = getStoreDateKey(due);
    const todayKey = getStoreDateKey(new Date());
    if (!dueKey || !todayKey) {
      return { category: 'FUTURE', label: 'موعد السداد غير محدد', variant: 'secondary' as const, days: 0 };
    }
    const toOrdinal = (value: string) => {
      const [year, month, day] = value.split('-').map(Number);
      return Date.UTC(year, month - 1, day) / (1000 * 60 * 60 * 24);
    };
    const diffDays = toOrdinal(dueKey) - toOrdinal(todayKey);

    if (status === 'paid') {
      return { category: 'PAID', label: 'تم السداد بالكامل', variant: 'success' as const, days: 0 };
    }

    if (diffDays < 0) {
      const overdueDays = Math.abs(diffDays);
      if (overdueDays <= 7) {
        return { category: 'OVERDUE_1_7', label: 'متأخر 1-7 أيام', variant: 'warning' as const, days: overdueDays };
      } else if (overdueDays <= 14) {
        return { category: 'OVERDUE_8_14', label: 'متأخر 8-14 يوم', variant: 'danger' as const, days: overdueDays };
      } else if (overdueDays <= 30) {
        return { category: 'OVERDUE_15_30', label: 'متأخر 15-30 يوم', variant: 'danger' as const, days: overdueDays };
      } else {
        return { category: 'OVERDUE_30_PLUS', label: 'متأخر أكثر من 30 يوم', variant: 'danger' as const, days: overdueDays };
      }
    }

    if (diffDays <= 7) {
      return { category: 'DUE_SOON', label: 'موعد السداد قريب', variant: 'warning' as const, days: diffDays };
    } else if (diffDays <= 30) {
      return { category: 'CURRENT', label: 'موعد السداد خلال هذا الشهر', variant: 'success' as const, days: diffDays };
    } else {
      return { category: 'FUTURE', label: 'موعد السداد لاحقًا', variant: 'secondary' as const, days: diffDays };
    }
  };


  return (
    <div>
      <PageHeader
        title={t('debts.title')}
        description="تتبع الديون وتحليل التقادم مع تنبيهات ذكية"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
              <Button variant="secondary" size={getButtonSize('debts', 'headerActions')} onClick={() => navigate('/app/customers')}>
                <ArrowRight style={{ width: '16px', height: '16px', marginRight: '8px' }} />
                العودة للزبائن
              </Button>
          </div>
        }
      />

      {/* Stats Cards + Advanced Search - side by side on desktop */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '12px', marginBottom: '12px', alignItems: 'start' }}>
        <DebtStats stats={stats} />
        <AdvancedSearch
          onSearch={handleSearchAdvanced}
          customers={debts.map(debt => debt.customer).filter(Boolean)}
          loading={isLoading}
        />
      </div>

      <div className="rounded-[16px] border border-[var(--border-default)] bg-[var(--bg-surface)] shadow-[0_10px_28px_rgba(15,23,42,0.05)]">
        <div className="flex items-center justify-between gap-3 border-b border-[var(--border-subtle)] px-5 py-4">
          <div className="flex items-center gap-2">
            <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
              <AlertTriangle className="h-4 w-4" />
            </span>
            <div>
              <h3 className="text-sm font-extrabold text-[var(--text-primary)]">قائمة الديون</h3>
                    <p className="mt-0.5 text-[11px] font-medium text-[var(--text-muted)]">{debtTab === 'open' ? 'الديون المفتوحة والمتأخرة' : debtTab === 'paid' ? 'الديون المسددة' : 'كل سجلات الديون'} · {displayedDebts.length} سجل</p>
            </div>
          </div>
                <div className="flex flex-wrap gap-2" role="tablist" aria-label="تصفية الديون">
                  {[
                    { value: 'open', label: 'مفتوحة' },
                    { value: 'paid', label: 'مسددة' },
                    { value: 'all', label: 'الكل' },
                  ].map((tab) => (
                    <Button
                      key={tab.value}
                      size="sm"
                      variant={debtTab === tab.value ? 'primary' : 'ghost'}
                      className="gap-2 whitespace-nowrap"
                      type="button"
                      role="tab"
                      aria-selected={debtTab === tab.value}
                      onClick={() => setDebtTab(tab.value as 'open' | 'paid' | 'all')}
                    >
                      {tab.label}
                    </Button>
                  ))}
                </div>
        </div>

        {isLoading ? (
          <div className="flex h-64 items-center justify-center">
            <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          </div>
        ) : displayedDebts.length === 0 ? (
          <div className="px-4 py-12 text-center">
            <DollarSign className="mx-auto mb-4 h-10 w-10 text-text-secondary" />
            <p className="text-sm text-text-secondary">{debtTab === 'paid' ? 'لا توجد ديون مسددة' : debtTab === 'all' ? 'لا توجد سجلات ديون' : 'لا توجد ديون مفتوحة'}</p>
            <p className="mt-1 text-[11px] text-text-tertiary">{debtTab === 'open' ? 'جميع الديون مسددة أو لا توجد ديون حالية' : 'جرّب تغيير نطاق العرض أو البحث'}</p>
          </div>
        ) : (
          <>
          <div className="customer-cards-grid gap-3 p-4">
            {displayedDebts.map((debt: any) => {
              const aging = getDebtAging(debt.dueDate, debt.status);
              const isPaid = Number(debt.remainingAmount ?? debt.remaining_amount ?? 0) <= 0 || debt.status === 'paid';
              return (
                <article
                  key={debt.id}
                  className="rounded-2xl border border-[var(--border-subtle)] bg-[var(--bg-surface)] p-4 shadow-[0_6px_18px_rgba(15,23,42,0.05)] transition-all duration-200 hover:-translate-y-0.5 hover:border-[var(--color-primary-20)] hover:shadow-[0_10px_24px_rgba(15,23,42,0.08)]"
                  onClick={() => handleViewDebt(debt)}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex min-w-0 items-center gap-2.5">
                      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-[var(--border-subtle)] bg-[var(--color-primary-10)] text-[var(--primary)]">
                        <UserRound className="h-4 w-4" />
                      </div>
                      <div className="min-w-0">
                        <h4 className="truncate text-sm font-black text-[var(--text-primary)]">{debt.customer?.name || 'عميل غير محدد'}</h4>
                        <p className="mt-0.5 truncate text-[10px] font-medium text-[var(--text-tertiary)]">{debt.invoiceNumber || 'فاتورة غير مرتبطة'}</p>
                      </div>
                    </div>
                    <Badge variant={aging.category === 'PAID' ? 'success' : aging.category.startsWith('OVERDUE') ? 'danger' : debt.status === 'partial' ? 'warning' : 'secondary'} size="sm" className="shrink-0 rounded-full">
                      {aging.category === 'PAID' ? 'مدفوع' : aging.category.startsWith('OVERDUE') ? 'متأخر' : debt.status === 'partial' ? 'جزئي' : 'معلق'}
                    </Badge>
                  </div>
                  <div className="mt-3 grid grid-cols-2 gap-2">
                    <div className="rounded-xl bg-[var(--bg-surface-elevated)] px-3 py-2"><span className="block text-[10px] text-[var(--text-muted)]">المبلغ</span><strong className="mt-0.5 block text-sm text-[var(--text-primary)]">₪{debt.amount?.toLocaleString() || '0'}</strong></div>
                    <div className="rounded-xl bg-[var(--bg-surface-elevated)] px-3 py-2"><span className="block text-[10px] text-[var(--text-muted)]">المتبقي</span><strong className="mt-0.5 block text-sm text-[var(--color-danger)]">₪{debt.remainingAmount?.toLocaleString() || '0'}</strong></div>
                  </div>
                  <div className="mt-3 flex items-center gap-2 text-[11px] text-[var(--text-secondary)]">
                    <Calendar className="h-3.5 w-3.5 text-[var(--text-muted)]" />
                    <span>{formatStoreDate(debt.dueDate, 'ar-SA')}</span>
                    <span className="text-[var(--text-muted)]">·</span>
                    <span className="truncate">{aging.days > 0 ? `متبقي ${aging.days} أيام` : aging.label}</span>
                  </div>
                  <div className="mt-3 flex justify-end gap-1.5 border-t border-[var(--border-subtle)] pt-2.5">
                    <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); handleViewDebt(debt); }} aria-label="عرض الدين" title="عرض الدين"><Eye className="h-3.5 w-3.5" /></Button>
                    {isPaid ? (
                      <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); handleViewDebt(debt); }} aria-label="سجل الدفعات" title="سجل الدفعات"><History className="h-4 w-4" /></Button>
                    ) : (
                      <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); handleRecordPayment(debt.customer?.id, debt.customer?.name); }} aria-label="تسجيل دفعة" title="تسجيل دفعة"><DollarSign className="h-4 w-4" /></Button>
                    )}
                  </div>
                </article>
              );
            })}
          </div>
          <div className="hidden md:block">
            <PaginationControls page={page} pageSize={pageSize} total={total} onPageChange={setPage} isLoading={isLoading} />
          </div>
          <div className="hidden overflow-x-auto">
            <Table className="min-w-[820px] table-fixed">
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[22%]">العميل / الفاتورة</TableHead>
                  <TableHead className="w-[11%] whitespace-nowrap text-center">المبلغ</TableHead>
                  <TableHead className="w-[12%] whitespace-nowrap text-center">المتبقي</TableHead>
                  <TableHead className="w-[15%] whitespace-nowrap">موعد السداد</TableHead>
                  <TableHead className="w-[20%]">تصنيف العمر</TableHead>
                  <TableHead className="w-[10%]">الحالة</TableHead>
                  <TableHead className="sticky left-0 z-10 w-[10%] bg-[var(--bg-surface)] text-end">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody className="divide-y divide-[var(--border-subtle)]" dir="rtl">
                {displayedDebts.map((debt: any) => {
                  const aging = getDebtAging(debt.dueDate, debt.status);
                  const isPaid = Number(debt.remainingAmount ?? debt.remaining_amount ?? 0) <= 0 || debt.status === 'paid';
                  return (
                    <TableRow key={debt.id} dir="rtl" className="cursor-pointer transition-all duration-200 hover:bg-[var(--color-primary-05)] [&>td]:h-[60px] [&>td]:py-2" onClick={() => handleViewDebt(debt)}>
                      <TableCell>
                        <div className="flex min-w-0 items-center gap-2">
                          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl border border-[var(--border-subtle)] bg-[var(--color-primary-10)] text-[var(--primary)] shadow-sm">
                            <UserRound className="h-4 w-4" />
                          </div>
                          <div className="min-w-0">
                            <div className="truncate font-black text-[var(--text-primary)]">{debt.customer?.name || 'عميل غير محدد'}</div>
                            <div className="mt-0.5 truncate text-[11px] font-medium text-[var(--text-tertiary)]">{debt.invoiceNumber || 'فاتورة غير مرتبطة'}</div>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="whitespace-nowrap px-3 text-center font-black text-[var(--text-primary)]">₪{debt.amount?.toLocaleString() || '0'}</TableCell>
                      <TableCell className="whitespace-nowrap px-3 text-center">
                        <Badge variant={Number(debt.remainingAmount || 0) > 0 ? 'danger' : 'success'} size="sm" className="min-w-[88px] justify-center rounded-full">
                          ₪{debt.remainingAmount?.toLocaleString() || '0'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2 font-semibold text-[var(--text-secondary)]">
                          <Calendar className="h-3.5 w-3.5 text-[var(--text-muted)]" />
                          {formatStoreDate(debt.dueDate, 'ar-SA')}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-col items-start gap-1.5">
                          <Badge variant={aging.variant} size="sm" className="rounded-full" title={`موعد السداد: ${formatStoreDate(debt.dueDate, 'ar-SA')}`}>
                            {aging.label}
                          </Badge>
                          <span className="text-[11px] font-medium text-[var(--text-secondary)]">
                            {aging.category === 'PAID'
                              ? 'تم السداد بالكامل'
                              : aging.category.startsWith('OVERDUE')
                              ? `متأخر فعلياً ${aging.days} ${aging.days === 1 ? 'يوم' : 'يوماً'}`
                              : aging.days > 0
                                ? `متبقي ${aging.days} ${aging.days === 1 ? 'يوم' : 'أيام'}`
                                : 'موعد اليوم'}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={aging.category === 'PAID' ? 'success' : aging.category.startsWith('OVERDUE') ? 'danger' : debt.status === 'partial' ? 'warning' : 'secondary'} size="sm" className="rounded-full">
                          {aging.category === 'PAID' ? 'مدفوع' : aging.category.startsWith('OVERDUE') ? 'متأخر' : debt.status === 'partial' ? 'جزئي' : 'معلق'}
                        </Badge>
                      </TableCell>
                      <TableCell className="sticky left-0 z-[1] bg-[var(--bg-surface)] text-end">
                        <div className="flex items-center justify-end gap-1">
                          <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); handleViewDebt(debt); }} aria-label="عرض الدين" title="عرض الدين">
                            <Eye className="h-3.5 w-3.5" />
                          </Button>
                          {isPaid ? (
                            <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); handleViewDebt(debt); }} aria-label="سجل الدفعات" title="سجل الدفعات">
                              <History className="h-4 w-4" />
                            </Button>
                          ) : (
                            <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); handleRecordPayment(debt.customer?.id, debt.customer?.name); }} className="text-text-secondary hover:text-text-primary" aria-label="تسجيل دفعة" title="تسجيل دفعة">
                              <DollarSign className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
            <PaginationControls page={page} pageSize={pageSize} total={total} onPageChange={setPage} isLoading={isLoading} />
          </div>
          <div className="debt-mobile-cards space-y-3 p-3">
            {displayedDebts.map((debt: any) => {
              const aging = getDebtAging(debt.dueDate, debt.status);
              const isPaid = Number(debt.remainingAmount ?? debt.remaining_amount ?? 0) <= 0 || debt.status === 'paid';
              return (
                <article
                  key={debt.id}
                  className="rounded-2xl border border-[var(--border-subtle)] bg-[var(--bg-surface)] p-4 shadow-[0_6px_18px_rgba(15,23,42,0.05)]"
                  onClick={() => handleViewDebt(debt)}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex min-w-0 items-center gap-3">
                      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl border border-[var(--border-subtle)] bg-[var(--color-primary-10)] text-[var(--primary)]">
                        <UserRound className="h-4 w-4" />
                      </div>
                      <div className="min-w-0">
                        <h4 className="truncate font-black text-[var(--text-primary)]">{debt.customer?.name || 'عميل غير محدد'}</h4>
                        <p className="truncate text-[11px] text-[var(--text-tertiary)]">{debt.invoiceNumber || 'فاتورة غير مرتبطة'}</p>
                      </div>
                    </div>
                    <Badge variant={aging.category === 'PAID' ? 'success' : aging.category.startsWith('OVERDUE') ? 'danger' : debt.status === 'partial' ? 'warning' : 'secondary'} size="sm" className="shrink-0 rounded-full">
                      {aging.category === 'PAID' ? 'مدفوع' : aging.category.startsWith('OVERDUE') ? 'متأخر' : debt.status === 'partial' ? 'جزئي' : 'معلق'}
                    </Badge>
                  </div>
                  <div className="mt-4 grid grid-cols-2 gap-3 text-sm">
                    <div className="rounded-xl bg-[var(--bg-surface-elevated)] p-3"><span className="block text-[11px] text-[var(--text-muted)]">المبلغ</span><strong className="mt-1 block text-[var(--text-primary)]">₪{debt.amount?.toLocaleString() || '0'}</strong></div>
                    <div className="rounded-xl bg-[var(--bg-surface-elevated)] p-3"><span className="block text-[11px] text-[var(--text-muted)]">المتبقي</span><strong className="mt-1 block text-[var(--color-danger)]">₪{debt.remainingAmount?.toLocaleString() || '0'}</strong></div>
                  </div>
                  <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-[var(--text-secondary)]">
                    <Calendar className="h-3.5 w-3.5 text-[var(--text-muted)]" />
                    <span>{formatStoreDate(debt.dueDate, 'ar-SA')}</span>
                    <span className="text-[var(--text-muted)]">·</span>
                    <span>{aging.days > 0 ? `متبقي ${aging.days} أيام` : aging.label}</span>
                  </div>
                  <div className="mt-4 flex justify-end gap-2 border-t border-[var(--border-subtle)] pt-3">
                    <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); handleViewDebt(debt); }} aria-label="عرض الدين" title="عرض الدين"><Eye className="h-3.5 w-3.5" /></Button>
                    {isPaid ? (
                      <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); handleViewDebt(debt); }} aria-label="سجل الدفعات" title="سجل الدفعات"><History className="h-4 w-4" /></Button>
                    ) : (
                      <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); handleRecordPayment(debt.customer?.id, debt.customer?.name); }} aria-label="تسجيل دفعة" title="تسجيل دفعة"><DollarSign className="h-4 w-4" /></Button>
                    )}
                  </div>
                </article>
              );
            })}
            <PaginationControls page={page} pageSize={pageSize} total={total} onPageChange={setPage} isLoading={isLoading} />
          </div>
          </>
        )}
      </div>

      {/* Payment Modal */}
      {paymentModalOpen && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.5)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 99999,
        }}>
          <div style={{
            backgroundColor: 'var(--bg-surface)',
            padding: '24px',
            borderRadius: '16px',
            maxWidth: '500px',
            width: '100%',
            margin: '16px',
            border: '1px solid var(--border-primary)',
            boxShadow: '0 20px 60px rgba(0, 0, 0, 0.3)',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
              <h3 style={{ fontSize: '18px', fontWeight: '600', color: 'var(--text-primary)' }}>تسجيل دفعة</h3>
              <button 
                onClick={() => setPaymentModalOpen(false)}
                style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '24px', color: 'var(--text-secondary)' }}
              >
                ×
              </button>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
              <div>
                <label className="text-small font-medium text-text mb-sm block">
                  العميل
                </label>
                <Input
                  value={selectedCustomer?.name || ''}
                  disabled
                />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">
                  مبلغ الدفعة
                </label>
                <Input
                  type="number"
                  placeholder="أدخل المبلغ..."
                  value={paymentAmount}
                  onChange={(e) => setPaymentAmount(e.target.value)}
                  autoFocus
                />
              </div>
              
              {/* Automatic balance calculation (SALES-PHILOSOPHY.md) */}
              {paymentAmount && !isNaN(parseFloat(paymentAmount)) && (
                <div style={{
                  padding: '12px',
                  background: 'linear-gradient(135deg, var(--color-primary-05) 0%, rgba(147, 51, 234, 0.05) 100%)',
                  border: '1px solid var(--color-primary-10)',
                  borderRadius: '8px'
                }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <span style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>الرصيد الجديد:</span>
                    <span style={{ fontSize: '18px', fontWeight: '700', color: 'var(--color-info)' }}>
                      ₪{Math.max(0, (selectedCustomer?.outstanding || 0) - parseFloat(paymentAmount)).toLocaleString()}
                    </span>
                  </div>
                </div>
              )}
              <div>
                <label className="text-small font-medium text-text mb-sm block">
                  طريقة الدفع
                </label>
                <select
                  value={paymentMethod}
                  onChange={(e) => setPaymentMethod(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '8px 12px',
                    borderRadius: '8px',
                    border: '1px solid var(--border-primary)',
                    backgroundColor: 'var(--bg-surface)',
                    color: 'var(--text-primary)',
                    fontSize: '14px',
                  }}
                >
                  <option value="cash">نقدي</option>
                  <option value="credit">بطاقة ائتمان</option>
                  <option value="bank_transfer">تحويل بنكي</option>
                  <option value="check">شيك</option>
                </select>
              </div>
              <div className="flex gap-sm justify-end">
                <Button
                  variant="secondary"
                  size={getButtonSize('debts', 'modalAction')}
                  onClick={() => setPaymentModalOpen(false)}
                >
                  إلغاء
                </Button>
                <Button
                  variant="primary"
                  size={getButtonSize('debts', 'modalAction')}
                  onClick={handlePaymentSubmit}
                  disabled={
                    recordPaymentMutation.isPending ||
                    !paymentAmount ||
                    isNaN(parseFloat(paymentAmount)) ||
                    parseFloat(paymentAmount) <= 0 ||
                    parseFloat(paymentAmount) > (selectedCustomer?.outstanding || 0)
                  }
                >
                  {recordPaymentMutation.isPending ? 'جارٍ التسجيل...' : 'تسجيل الدفعة'}
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* View Debt Modal */}
      {isViewModalOpen && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.5)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 99999,
        }}>
          <div style={{
            backgroundColor: 'var(--bg-surface)',
            padding: '24px',
            borderRadius: '16px',
            maxWidth: '500px',
            width: '100%',
            margin: '16px',
            border: '1px solid var(--border-primary)',
            boxShadow: '0 20px 60px rgba(0, 0, 0, 0.3)',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
              <h3 style={{ fontSize: '18px', fontWeight: '600', color: 'var(--text-primary)' }}>تفاصيل الدين</h3>
              <button 
                onClick={() => setIsViewModalOpen(false)}
                style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '24px', color: 'var(--text-secondary)' }}
              >
                ×
              </button>
            </div>
            {selectedDebt && (
              <div style={{ 
                display: 'grid', 
                gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', 
                gap: '16px' 
              }}>
                {((selectedDebt.debt_reason || selectedDebt.reason || selectedDebt.notes) && (
                  <div style={{ gridColumn: '1 / -1' }}>
                    <label className="text-small font-medium text-text mb-sm block">سبب الدين</label>
                    <div style={{
                      background: 'rgba(15, 23, 42, 0.02)',
                      border: '1px solid var(--border-default)',
                      borderRadius: '10px',
                      padding: '12px 14px',
                      fontSize: '13px',
                      lineHeight: '1.7',
                      color: 'var(--text-primary)',
                      whiteSpace: 'pre-wrap',
                    }}>
                      {selectedDebt.debt_reason || selectedDebt.reason || selectedDebt.notes || 'لا يوجد سبب محدد'}
                    </div>
                  </div>
                ))}
                <div>
                  <label className="text-small font-medium text-text mb-sm block">العميل</label>
                  <Input value={selectedDebt.customer?.name || ''} disabled />
                </div>
                <div>
                  <label className="text-small font-medium text-text mb-sm block">مبلغ الدين</label>
                  <div className="numeric-price" style={{ fontSize: '15px', fontWeight: '500', color: 'var(--text-primary)' }}>
                    ₪{selectedDebt.amount?.toLocaleString() || 0}
                  </div>
                </div>
                <div>
                  <label className="text-small font-medium text-text mb-sm block">موعد السداد</label>
                  <Input value={selectedDebt.dueDate ? formatStoreDate(selectedDebt.dueDate, 'ar-SA') : 'غير محدد'} disabled />
                </div>
                <div>
                  <label className="text-small font-medium text-text mb-sm block">الحالة</label>
                  <Input value={Number(selectedDebt.remaining_amount ?? selectedDebt.remainingAmount ?? 0) <= 0 ? 'مدفوع' : selectedDebt.status === 'overdue' ? 'متأخر' : selectedDebt.status === 'partial' ? 'جزئي' : 'معلق'} disabled />
                </div>
                <div className="flex gap-sm justify-end" style={{ gridColumn: '1 / -1' }}>
                  <Button variant="secondary" size={getButtonSize('debts', 'modalAction')} onClick={() => setIsViewModalOpen(false)}>
                    إغلاق
                  </Button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Success Modal */}
      {successModalOpen && lastPayment && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.6)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 99999,
          animation: 'fadeIn 0.3s ease-in-out',
          backdropFilter: 'blur(4px)',
        }}>
          <div style={{
            backgroundColor: 'var(--bg-surface)',
            padding: '32px',
            borderRadius: '16px',
            maxWidth: '440px',
            width: '100%',
            margin: '16px',
            border: '1px solid var(--border-primary)',
            boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(255, 255, 255, 0.05)',
            animation: 'slideUp 0.3s ease-out',
          }}>
            {/* Success Icon */}
            <div style={{
              display: 'flex',
              justifyContent: 'center',
              marginBottom: '20px',
            }}>
              <div style={{
                width: '72px',
                height: '72px',
                borderRadius: '50%',
                backgroundColor: 'rgba(34, 197, 94, 0.12)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                animation: 'scaleIn 0.5s ease-out',
                boxShadow: '0 4px 12px rgba(34, 197, 94, 0.2)',
              }}>
                <CheckCircle style={{ width: '40px', height: '40px', color: 'var(--color-success)' }} />
              </div>
            </div>

            {/* Success Message */}
            <div style={{ textAlign: 'center', marginBottom: '20px' }}>
              <h2 style={{
                fontSize: '22px',
                fontWeight: '700',
                color: 'var(--text-primary)',
                marginBottom: '6px',
                letterSpacing: '-0.5px',
              }}>
                تم تسجيل الدفعة بنجاح!
              </h2>
              <p style={{
                fontSize: '13px',
                color: 'var(--text-secondary)',
                marginBottom: '20px',
                lineHeight: '1.5',
              }}>
                تم تحديث بيانات الدين ورصيد العميل
              </p>
            </div>

            {/* Payment Details */}
            <div style={{
              backgroundColor: 'var(--bg-secondary)',
              borderRadius: '12px',
              padding: '18px',
              marginBottom: '20px',
              border: '1px solid var(--border-primary)',
            }}>
              <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(2, 1fr)',
                gap: '14px',
              }}>
                <div>
                  <label style={{
                    fontSize: '11px',
                    fontWeight: '600',
                    color: 'var(--text-secondary)',
                    marginBottom: '4px',
                    display: 'block',
                    textTransform: 'uppercase',
                    letterSpacing: '0.5px',
                  }}>
                    العميل
                  </label>
                  <div style={{
                    fontSize: '14px',
                    fontWeight: '600',
                    color: 'var(--text-primary)',
                  }}>
                    {lastPayment.customerName}
                  </div>
                </div>
                <div>
                  <label style={{
                    fontSize: '11px',
                    fontWeight: '600',
                    color: 'var(--text-secondary)',
                    marginBottom: '4px',
                    display: 'block',
                    textTransform: 'uppercase',
                    letterSpacing: '0.5px',
                  }}>
                    المبلغ
                  </label>
                  <div style={{
                    fontSize: '17px',
                    fontWeight: '700',
                    color: '#22c55e',
                  }}>
                    ₪{lastPayment.amount.toLocaleString()}
                  </div>
                </div>
                <div>
                  <label style={{
                    fontSize: '11px',
                    fontWeight: '600',
                    color: 'var(--text-secondary)',
                    marginBottom: '4px',
                    display: 'block',
                    textTransform: 'uppercase',
                    letterSpacing: '0.5px',
                  }}>
                    طريقة الدفع
                  </label>
                  <div style={{
                    fontSize: '13px',
                    fontWeight: '500',
                    color: 'var(--text-primary)',
                  }}>
                    {lastPayment.method === 'cash' ? 'نقدي' : 
                     lastPayment.method === 'credit' ? 'بطاقة ائتمان' : 
                     lastPayment.method === 'bank_transfer' ? 'تحويل بنكي' : 'شيك'}
                  </div>
                </div>
                <div>
                  <label style={{
                    fontSize: '11px',
                    fontWeight: '600',
                    color: 'var(--text-secondary)',
                    marginBottom: '4px',
                    display: 'block',
                    textTransform: 'uppercase',
                    letterSpacing: '0.5px',
                  }}>
                    التاريخ
                  </label>
                  <div style={{
                    fontSize: '13px',
                    fontWeight: '500',
                    color: 'var(--text-primary)',
                  }}>
                    {lastPayment.date}
                  </div>
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div style={{
              display: 'flex',
              gap: '10px',
              justifyContent: 'center',
              marginTop: '8px',
            }}>
              <button
                onClick={() => setSuccessModalOpen(false)}
                style={{
                  flex: 1,
                  minWidth: '0',
                  padding: '10px 16px',
                  fontSize: '13px',
                  fontWeight: '500',
                  color: '#f1f5f9',
                  backgroundColor: '#151d26',
                  border: '1px solid #24303b',
                  borderRadius: '8px',
                  cursor: 'pointer',
                  transition: 'all 0.15s ease',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  minHeight: '40px',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.transform = 'translateY(-1px)';
                  e.currentTarget.style.borderColor = '#10b981';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.transform = 'translateY(0)';
                  e.currentTarget.style.borderColor = '#24303b';
                }}
              >
                إغلاق
              </button>
              <button
                onClick={handlePrintReceipt}
                style={{
                  flex: 1,
                  minWidth: '0',
                  padding: '10px 16px',
                  fontSize: '13px',
                  fontWeight: '600',
                  color: '#FFFFFF',
                  backgroundColor: '#14b8a6',
                  border: '1px solid #14b8a6',
                  borderRadius: '8px',
                  cursor: 'pointer',
                  transition: 'all 0.15s ease',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '8px',
                  minHeight: '40px',
                  boxShadow: '0 0 24px rgba(34, 211, 238, 0.08)',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.transform = 'translateY(-1px)';
                  e.currentTarget.style.borderColor = '#06b6d4';
                  e.currentTarget.style.boxShadow = '0 0 30px rgba(34, 211, 238, 0.1)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.transform = 'translateY(0)';
                  e.currentTarget.style.borderColor = '#14b8a6';
                  e.currentTarget.style.boxShadow = '0 0 24px rgba(34, 211, 238, 0.08)';
                }}
              >
                <Printer style={{ width: '16px', height: '16px' }} />
                طباعة الإيصال
              </button>
              <button
                onClick={handlePrintReceipt}
                style={{
                  flex: 1,
                  minWidth: '0',
                  padding: '10px 16px',
                  fontSize: '13px',
                  fontWeight: '600',
                  color: '#0f172a',
                  backgroundColor: '#f8fafc',
                  border: '1px solid #cbd5e1',
                  borderRadius: '8px',
                  cursor: 'pointer',
                  transition: 'all 0.15s ease',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '8px',
                  minHeight: '40px',
                }}
                title="يفتح نافذة الطباعة، اختر حفظ كـ PDF"
              >
                <FileDown style={{ width: '16px', height: '16px' }} />
                حفظ PDF
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
