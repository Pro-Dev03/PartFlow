import { useState } from 'react';
import { useTranslation as useTranslationHook } from '../../../hooks/useTranslation';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { getButtonSize } from '../../../config/button-sizes';
import { 
  DollarSign, 
  AlertTriangle,
  AlertCircle,
  Calendar,
  Zap,
  Eye,
  Bell,
  CheckCircle,
  Printer,
  MoreHorizontal
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

export function DebtsPage() {
  const { t } = useTranslationHook();
  const [paymentModalOpen, setPaymentModalOpen] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState<any>(null);
  const [paymentAmount, setPaymentAmount] = useState('');
  const [paymentMethod, setPaymentMethod] = useState('cash');
  const [selectedDebt, setSelectedDebt] = useState<any>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [successModalOpen, setSuccessModalOpen] = useState(false);
  const [lastPayment, setLastPayment] = useState<any>(null);

  // Custom hook
  const {
    debts,
    filteredDebts,
    stats,
    isLoading,
    refetch,
    setSearchQuery,
    setSearchFilters,
    recordPaymentMutation,
  } = useDebts();

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
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const due = new Date(dueDate);
    due.setHours(0, 0, 0, 0);
    
    // Handle invalid dates
    if (isNaN(due.getTime())) {
      return { category: 'FUTURE', label: 'موعد السداد غير محدد', variant: 'secondary' as const, days: 0 };
    }
    
    const diffTime = due.getTime() - today.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

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
      {/* Page Header */}
      <PageHeader
        eyebrow="Debt Intelligence"
        title={t('debts.title')}
        description="تتبع الديون وتحليل التقادم مع تنبيهات ذكية"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="secondary" size={getButtonSize('debts', 'headerActions')} onClick={() => void refetch()}>
              <Zap style={{ width: '16px', height: '16px', marginRight: '8px' }} />
              تحديث
            </Button>
            <Button variant="secondary" size={getButtonSize('debts', 'headerActions')} onClick={() => toast.info('لا توجد تنبيهات جديدة حالياً')}>
              <Bell style={{ width: '16px', height: '16px', marginRight: '8px' }} />
              تنبيهات
            </Button>
          </div>
        }
      />

      {/* AI Debt Insight */}
      <div className="premium-insight">
        <div className="premium-insight-icon"><AlertCircle className="h-3.5 w-3.5" /></div>
        <div className="premium-insight-copy">
          <p className="premium-insight-title">AI Debt Insight · قيد التطوير</p>
          <p className="premium-insight-text">ستوفر تحليلات ذكية للتقادم وتوصيات لمتابعة التحصيل.</p>
        </div>
        <Button variant="secondary" size={getButtonSize('debts', 'recommendation')} disabled className="premium-insight-action">قيد التطوير</Button>
      </div>

      {/* Stats Cards + Advanced Search - side by side on desktop */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '12px', marginBottom: '12px' }}>
        <DebtStats stats={stats} />
        <AdvancedSearch
          onSearch={handleSearchAdvanced}
          customers={debts.map(debt => debt.customer).filter(Boolean)}
          loading={isLoading}
        />
      </div>

      <div className="rounded-[12px] border border-border bg-surface shadow-[0_8px_18px_rgba(15,23,42,0.04)]">
        <div className="flex items-center justify-between gap-3 border-b border-border px-5 py-4">
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-4 w-4 text-primary" />
            <div>
              <h3 className="text-sm font-semibold text-text-primary">قائمة الديون</h3>
              <p className="mt-0.5 text-[11px] text-text-tertiary">جميع الديون المفتوحة والمتأخرة — {filteredDebts.length} سجل</p>
            </div>
          </div>
        </div>

        {isLoading ? (
          <div className="flex h-64 items-center justify-center">
            <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          </div>
        ) : filteredDebts.length === 0 ? (
          <div className="px-4 py-12 text-center">
            <DollarSign className="mx-auto mb-4 h-10 w-10 text-text-secondary" />
            <p className="text-sm text-text-secondary">لا توجد ديون</p>
            <p className="mt-1 text-[11px] text-text-tertiary">جميع الديون مدفوعة</p>
          </div>
        ) : (
          <div className="hidden md:block">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[18%]">العميل</TableHead>
                  <TableHead className="w-[12%] text-center">المبلغ</TableHead>
                  <TableHead className="w-[12%] text-center">المتبقي</TableHead>
                  <TableHead className="w-[14%]">موعد السداد</TableHead>
                  <TableHead className="w-[24%]">تصنيف العمر</TableHead>
                  <TableHead className="w-[10%]">الحالة</TableHead>
                  <TableHead className="w-[10%] text-end">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredDebts.map((debt: any) => {
                  const aging = getDebtAging(debt.dueDate, debt.status);
                  return (
                    <TableRow key={debt.id} className="cursor-pointer hover:bg-surface-elevated/30" onClick={() => handleViewDebt(debt)}>
                      <TableCell className="font-semibold text-text-primary">{debt.customer?.name}</TableCell>
                      <TableCell className="text-center font-medium text-text-primary">₪{debt.amount?.toLocaleString()}</TableCell>
                      <TableCell className="text-center font-medium text-danger">₪{debt.remainingAmount?.toLocaleString() || '0'}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2 text-text-secondary">
                          <Calendar className="h-3.5 w-3.5" />
                          {debt.dueDate ? new Date(debt.dueDate).toLocaleDateString('en-GB') : 'غير محدد'}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-col items-start gap-1.5">
                          <Badge variant={aging.variant} size="sm" title={`موعد السداد: ${debt.dueDate ? new Date(debt.dueDate).toLocaleDateString('ar-SA') : 'غير محدد'}`}>
                            {aging.label}
                          </Badge>
                          <span className="text-[11px] text-text-secondary">
                            {aging.category.startsWith('OVERDUE')
                              ? `متأخر فعلياً ${aging.days} ${aging.days === 1 ? 'يوم' : 'يوماً'}`
                              : aging.days > 0
                                ? `متبقي ${aging.days} ${aging.days === 1 ? 'يوم' : 'أيام'}`
                                : 'موعد اليوم'}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={aging.category.startsWith('OVERDUE') ? 'danger' : debt.status === 'partial' ? 'warning' : 'secondary'} size="sm">
                          {aging.category.startsWith('OVERDUE') ? 'متأخر' : debt.status === 'partial' ? 'جزئي' : 'معلق'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-end">
                        <div className="flex items-center justify-end gap-2">
                          <Button variant="primary" size="sm" onClick={(e) => { e.stopPropagation(); handleViewDebt(debt); }} className="h-8 px-2.5 text-[11px]" aria-label="عرض الدين" title="عرض الدين">
                            <Eye className="h-3.5 w-3.5" />
                          </Button>
                          <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); handleRecordPayment(debt.customer?.id, debt.customer?.name); }} className="text-text-secondary hover:text-text-primary" aria-label="تسجيل دفعة" title="تسجيل دفعة">
                            <DollarSign className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        )}

        <div className="block md:hidden">
          {isLoading ? (
            <div className="flex h-64 items-center justify-center"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
          ) : filteredDebts.length === 0 ? (
            <div className="px-4 py-12 text-center">
              <DollarSign className="mx-auto mb-4 h-10 w-10 text-text-secondary" />
              <p className="text-sm text-text-secondary">لا توجد ديون</p>
              <p className="mt-1 text-[11px] text-text-tertiary">جميع الديون مدفوعة</p>
            </div>
          ) : (
            <div className="grid gap-3 p-4">
              {filteredDebts.map((debt: any) => {
                const aging = getDebtAging(debt.dueDate, debt.status);
                return (
                  <div key={debt.id} className="rounded-xl border border-border bg-surface-elevated/25 p-4" onClick={() => handleViewDebt(debt)}>
                    <div className="mb-3 flex items-start justify-between gap-3">
                      <div>
                        <div className="font-semibold text-text-primary">{debt.customer?.name}</div>
                        <div className="mt-1 text-[11px] text-text-tertiary">{debt.dueDate ? new Date(debt.dueDate).toLocaleDateString('en-GB') : 'غير محدد'}</div>
                      </div>
                      <Badge variant={aging.category.startsWith('OVERDUE') ? 'danger' : debt.status === 'partial' ? 'warning' : 'secondary'} size="sm">
                        {aging.category.startsWith('OVERDUE') ? 'متأخر' : debt.status === 'partial' ? 'جزئي' : 'معلق'}
                      </Badge>
                    </div>

                    <div className="space-y-2 text-sm">
                      <div className="flex items-center justify-between gap-2"><span className="text-text-tertiary">المبلغ</span><span className="font-semibold text-text-primary">₪{debt.amount?.toLocaleString()}</span></div>
                      <div className="flex items-center justify-between gap-2"><span className="text-text-tertiary">المتبقي</span><span className="font-semibold text-danger">₪{debt.remainingAmount?.toLocaleString() || '0'}</span></div>
                      <div className="flex items-center justify-between gap-2"><span className="text-text-tertiary">التصنيف</span><Badge variant={aging.variant} size="sm">{aging.label}</Badge></div>
                    </div>

                    <div className="mt-3 flex items-center justify-between gap-2 border-t border-border pt-3">
                      <Button variant="primary" size="sm" onClick={(e) => { e.stopPropagation(); handleViewDebt(debt); }} className="h-8 px-3 text-[11px]">
                        <Eye className="h-3.5 w-3.5" />
                        عرض
                      </Button>
                      <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); handleRecordPayment(debt.customer?.id, debt.customer?.name); }} className="text-text-secondary hover:text-text-primary" aria-label="تسجيل دفعة" title="تسجيل دفعة">
                        <DollarSign className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
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
                  <Input value={selectedDebt.dueDate ? new Date(selectedDebt.dueDate).toLocaleDateString('en-GB') : 'غير محدد'} disabled />
                </div>
                <div>
                  <label className="text-small font-medium text-text mb-sm block">الحالة</label>
                  <Input value={selectedDebt.status === 'overdue' ? 'متأخر' : selectedDebt.status === 'partial' ? 'جزئي' : 'معلق'} disabled />
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
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
