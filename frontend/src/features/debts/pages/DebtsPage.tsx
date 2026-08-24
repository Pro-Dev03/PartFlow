import { useState } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { SearchInput } from '../../../components/ui/search-input';
import { PageHeader } from '../../../components/ui/page-header';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { Modal } from '../../../components/ui/modal';
import { getButtonSize } from '../../../config/button-sizes';
import { 
  DollarSign, 
  AlertTriangle,
  Calendar,
  Sparkles,
  Zap,
  Clock,
  X,
  Eye,
  TrendingUp
} from 'lucide-react';

// Custom hooks
import { useDebts } from '../hooks/useDebts';

// Components
import { DebtStats } from '../components/DebtStats';

// Types
import { Debt } from '../types/debts.types';

export function DebtsPage() {
  const { t } = useTranslation();
  const [paymentModalOpen, setPaymentModalOpen] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState<any>(null);
  const [paymentAmount, setPaymentAmount] = useState('');
  const [selectedDebt, setSelectedDebt] = useState<any>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);

  const handleClearSearch = () => {
    // Will be handled by the custom hook
  };

  // Custom hook
  const {
    debts,
    filteredDebts,
    stats,
    isLoading,
    searchQuery,
    setSearchQuery,
    recordPaymentMutation,
  } = useDebts();

  const handleRecordPayment = (customerId: string, customerName: string) => {
    setSelectedCustomer({ id: customerId, name: customerName });
    setPaymentAmount('');
    setPaymentModalOpen(true);
  };

  const handlePaymentSubmit = () => {
    if (paymentAmount && !isNaN(parseFloat(paymentAmount))) {
      recordPaymentMutation.mutate({
        customerId: selectedCustomer.id,
        amount: parseFloat(paymentAmount),
      });
      setPaymentModalOpen(false);
    }
  };

  const handleViewDebt = (debt: Debt) => {
    setSelectedDebt(debt);
    setIsViewModalOpen(true);
  };

  const handleReversePayment = (paymentId: string) => {
    if (window.confirm('هل أنت متأكد من عكس هذه الدفعة؟\n\nسيتم إنشاء سجل عكس الدفعة ولن يتم حذف الدفعة الأصلية.')) {
      // Implement reverse logic here
      console.log('Reverse payment:', paymentId);
    }
  };

  // Debt Aging System - تصنيف ديون حسب العمر
  const getDebtAging = (dueDate: string, status: string) => {
    const today = new Date();
    const due = new Date(dueDate);
    const diffTime = due.getTime() - today.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

    if (status === 'paid') {
      return { category: 'PAID', label: 'مدفوع', variant: 'success' as const, days: 0 };
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
      return { category: 'DUE_SOON', label: 'يستحق قريباً', variant: 'warning' as const, days: diffDays };
    } else if (diffDays <= 30) {
      return { category: 'CURRENT', label: 'مستحق', variant: 'success' as const, days: diffDays };
    } else {
      return { category: 'FUTURE', label: 'مستقبلي', variant: 'secondary' as const, days: diffDays };
    }
  };

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
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Debt Intelligence"
        title={t('debts.title')}
        description="تتبع الديون وتحليل التقادم مع تنبيهات ذكية"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="secondary" size={getButtonSize('debts', 'headerActions')}>
              <Zap style={{ width: '16px', height: '16px', marginRight: '8px' }} />
              تحديث
            </Button>
            <Button variant="secondary" size={getButtonSize('debts', 'headerActions')}>
              <Bell style={{ width: '16px', height: '16px', marginRight: '8px' }} />
              تنبيهات
            </Button>
          </div>
        }
      />

      {/* AI Debt Insight */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Sparkles style={{ width: '20px', height: '20px', color: '#22d3ee' }} />
            AI Debt Insight
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div style={{ display: 'flex', gap: '14px' }}>
            <div style={{
              width: '32px',
              height: '32px',
              borderRadius: '8px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: 'rgba(34, 211, 238, 0.1)',
              flexShrink: 0
            }}>
              <AlertCircle style={{ width: '16px', height: '16px', color: '#22d3ee' }} />
            </div>
            <div>
              <p style={{ fontSize: '13px', fontWeight: '600', color: '#f1f7ff' }}>خطر إفلاس محتمل</p>
              <p style={{ fontSize: '11px', color: '#8290a7', marginTop: '4px' }}>
                3 عملاء لديهم ديون متأخرة أكثر من 60 يوم. يُنصح باتخاذ إجراءات قانونية فورية.
              </p>
              <Button variant="secondary" size={getButtonSize('debts', 'recommendation')}>
                عرض التوصية ←
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Stats Cards */}
      <DebtStats stats={stats} />

      {/* Search */}
      <Card>
        <CardContent>
          <div style={{ padding: '18px' }}>
            <SearchInput
              placeholder="ابحث بالاسم أو الكود..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onClear={handleClearSearch}
              size="sm"
              className="w-full md:w-[500px] lg:w-[600px]"
            />
          </div>
        </CardContent>
      </Card>

      {/* Debts Table */}
      <Card>
        <CardHeader>
          <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <AlertTriangle style={{ width: '20px', height: '20px', color: '#22d3ee' }} />
            قائمة الديون
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '256px' }}>
              <div style={{ animation: 'spin 1s linear infinite', borderRadius: '50%', height: '32px', width: '32px', borderBottom: '2px solid #22d3ee' }} />
            </div>
          ) : filteredDebts.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '40px 20px' }}>
              <DollarSign style={{ width: '48px', height: '48px', color: '#8290a7', margin: '0 auto 16px' }} />
              <p style={{ fontSize: '13px', color: '#8290a7' }}>
                لا توجد ديون
              </p>
              <p style={{ fontSize: '11px', color: '#56647a', marginTop: '4px' }}>
                جميع الديون مدفوعة
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>العميل</TableHead>
                  <TableHead>المبلغ</TableHead>
                  <TableHead>المتبقي</TableHead>
                  <TableHead>تاريخ الاستحقاق</TableHead>
                  <TableHead>تصنيف العمر</TableHead>
                  <TableHead>الحالة</TableHead>
                  <TableHead style={{ textAlign: 'right' }}>الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredDebts.map((debt: any) => {
                  const aging = getDebtAging(debt.dueDate, debt.status);
                  
                  return (
                    <TableRow key={debt.id}>
                      <TableCell style={{ fontWeight: '500' }}>{debt.customer?.name}</TableCell>
                      <TableCell>₪{debt.amount?.toLocaleString()}</TableCell>
                      <TableCell>
                        <span style={{ color: '#fb7185', fontWeight: '500' }}>
                          ₪{debt.remainingAmount?.toLocaleString()}
                        </span>
                      </TableCell>
                      <TableCell>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                          <Calendar style={{ width: '14px', height: '14px', color: '#8290a7' }} />
                          {new Date(debt.dueDate).toLocaleDateString('ar-SA')}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={aging.variant} style={{ fontSize: '11px' }}>
                          {aging.label}
                          {aging.days > 0 && ` (${aging.days} يوم)`}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant={debt.status === 'overdue' ? 'danger' : 'warning'}>
                          {debt.status === 'overdue' ? 'متأخر' : 'معلق'}
                        </Badge>
                      </TableCell>
                      <TableCell style={{ textAlign: 'right' }}>
                        <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                          <Button
                            variant="ghost"
                            size={getButtonSize('debts', 'iconAction')}
                            onClick={() => handleViewDebt(debt)}
                          >
                            <Eye className="w-3.5 h-3.5" />
                          </Button>
                          <Button
                            variant="ghost"
                            size={getButtonSize('debts', 'iconAction')}
                            onClick={() => handleRecordPayment(debt.customer?.id, debt.customer?.name)}
                          >
                            <DollarSign className="w-3.5 h-3.5" />
                          </Button>
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

      {/* Payment Modal */}
      <Modal
        isOpen={paymentModalOpen}
        onClose={() => setPaymentModalOpen(false)}
        title="تسجيل دفعة"
      >
        <div className="space-y-md">
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
            />
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
              disabled={!paymentAmount || isNaN(parseFloat(paymentAmount))}
            >
              تسجيل الدفعة
            </Button>
          </div>
        </div>
      </Modal>

      {/* View Debt Modal */}
      <Modal
        isOpen={isViewModalOpen}
        onClose={() => setIsViewModalOpen(false)}
        title="تفاصيل الدين"
      >
        {selectedDebt && (
          <div className="space-y-md">
            <div>
              <label className="text-small font-medium text-text mb-sm block">العميل</label>
              <Input value={selectedDebt.customer?.name || ''} disabled />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">مبلغ الدين</label>
              <Input value={`₪${selectedDebt.amount?.toLocaleString() || 0}`} disabled />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">تاريخ الاستحقاق</label>
              <Input value={new Date(selectedDebt.dueDate).toLocaleDateString('ar-SA')} disabled />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">الحالة</label>
              <Input value={selectedDebt.status === 'overdue' ? 'متأخر' : 'معلق'} disabled />
            </div>
            <div className="flex gap-sm justify-end">
              <Button variant="secondary" size={getButtonSize('debts', 'modalAction')} onClick={() => setIsViewModalOpen(false)}>
                إغلاق
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
