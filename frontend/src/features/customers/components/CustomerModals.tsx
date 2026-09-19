import { useState, lazy, Suspense, useEffect } from 'react';
import { Modal } from '../../../design-system/components/modal';
import { Button } from '../../../design-system/components/button';
import { CustomerForm, type CustomerFormData } from '../../../components/forms/CustomerForm';
import { getButtonSize } from '../../../config/button-sizes';
import { Customer } from '../types/customers.types';
import { User, DollarSign, ShoppingCart, Calendar, CreditCard, FileText, History, AlertTriangle, RotateCcw } from 'lucide-react';
import { LedgerEntry } from '../../../design-system/components/financial-timeline';
import { customersApi, salesApi } from '../../../services/api/endpoints';
import { UsedPartsInvoice } from '../../../components/invoice/UsedPartsInvoice';
import { useNavigate } from 'react-router-dom';

// Lazy load heavy FinancialTimeline component
const FinancialTimeline = lazy(() => import('../../../design-system/components/financial-timeline').then(m => ({ default: m.FinancialTimeline })));

interface CustomerModalsProps {
  isModalOpen: boolean;
  setIsModalOpen: (open: boolean) => void;
  isViewModalOpen: boolean;
  setIsViewModalOpen: (open: boolean) => void;
  editingCustomer: Customer | null;
  setEditingCustomer: (customer: Customer | null) => void;
  selectedCustomer: Customer | null;
  setSelectedCustomer: (customer: Customer | null) => void;
  onSubmit: (data: CustomerFormData) => void;
}

export function CustomerModals({
  isModalOpen,
  setIsModalOpen,
  isViewModalOpen,
  setIsViewModalOpen,
  editingCustomer,
  setEditingCustomer,
  selectedCustomer,
  onSubmit,
}: CustomerModalsProps) {
  const navigate = useNavigate();
  const [ledgerEntries, setLedgerEntries] = useState<LedgerEntry[]>([]);
  const [loadingLedger, setLoadingLedger] = useState(false);
  const [customerSales, setCustomerSales] = useState<any[]>([]);
  const [isSalesHistoryOpen, setIsSalesHistoryOpen] = useState(false);
  const [loadingSales, setLoadingSales] = useState(false);
  const [salesError, setSalesError] = useState('');
  const [isInvoiceOpen, setIsInvoiceOpen] = useState(false);
  const [invoiceData, setInvoiceData] = useState<any>(null);
  const [loadingInvoice, setLoadingInvoice] = useState(false);

  // Load ledger entries when viewing a customer
  useEffect(() => {
    if (isViewModalOpen && selectedCustomer) {
      loadCustomerLedger(selectedCustomer.id);
    }
  }, [isViewModalOpen, selectedCustomer]);

  const loadCustomerLedger = async (customerId: string) => {
    setLoadingLedger(true);
    try {
      const response = await customersApi.ledger(customerId, { page: 1, per_page: 50 });
      if (response && response.data) {
        const customerEntries = response.data.filter((entry: any) =>
          !entry.customer_id || String(entry.customer_id) === String(customerId),
        );
        const formattedEntries = customerEntries.map((entry: any) => ({
          id: entry.id,
          customer_id: entry.customer_id,
          type: entry.type || entry.transaction_type,
          transaction_type: entry.transaction_type,
          amount: entry.amount,
          balance: entry.balance,
          previous_balance: entry.previous_balance,
          description: entry.description,
          created_at: entry.created_at,
          reference_id: entry.reference_id,
          reference_type: entry.reference_type,
        }));
        setLedgerEntries(formattedEntries);
      } else {
        setLedgerEntries([]);
      }
    } catch (error) {
      console.error('Failed to load ledger entries:', error);
      setLedgerEntries([]);
    } finally {
      setLoadingLedger(false);
    }
  };

  const loadCustomerSales = async () => {
    if (!selectedCustomer) return;

    setIsSalesHistoryOpen(true);
    setLoadingSales(true);
    setSalesError('');
    try {
      const response = await salesApi.list({
        customer_id: selectedCustomer.id,
        page: 1,
        per_page: 100,
      });
      const payload = response?.data ?? response;
      const sales = Array.isArray(payload)
        ? payload
        : payload?.sales ?? payload?.data ?? [];
      setCustomerSales(Array.isArray(sales) ? sales : []);
    } catch (error) {
      console.error('Failed to load customer sales:', error);
      setCustomerSales([]);
      setSalesError('تعذر تحميل فواتير العميل. حاول مرة أخرى.');
    } finally {
      setLoadingSales(false);
    }
  };

  const openSaleInvoice = async (sale: any) => {
    setLoadingInvoice(true);
    try {
      const response = await salesApi.get(String(sale.id));
      const payload = response?.data ?? response;
      const saleDetails = payload?.sale ?? payload;
      const items = Array.isArray(payload?.items) ? payload.items : [];
      const total = Number(saleDetails.total_amount ?? sale.total_amount ?? 0);
      const paidAmount = Number(saleDetails.paid_amount ?? sale.paid_amount ?? 0);

      setInvoiceData({
        id: saleDetails.id,
        invoiceNumber: saleDetails.invoice_number || sale.invoice_number || saleDetails.id,
        customerName: selectedCustomer?.name || '',
        customerPhone: selectedCustomer?.phone,
        saleDate: saleDetails.sale_date || sale.created_at || new Date().toISOString(),
        items: items.map((item: any) => ({
          name: item.product_name || item.name || 'منتج',
          sku: item.sku,
          barcode: item.barcode,
          condition: item.condition || 'NEW',
          grade: item.grade,
          sellingPrice: Number(item.unit_price ?? item.selling_price ?? 0),
          quantity: Number(item.quantity ?? 1),
          total: Number(item.total_amount ?? item.total ?? 0),
          discountAmount: Number(item.discount_amount || 0),
          taxAmount: Number(item.tax_amount || 0),
          partType: item.part_type || item.partType,
          partTypeColor: item.part_type_color || item.partTypeColor,
        })),
        subtotal: Number(saleDetails.subtotal ?? total),
        discountAmount: Number(saleDetails.discount_amount || 0),
        taxAmount: Number(saleDetails.tax_amount || 0),
        total,
        paidAmount,
        remaining: Math.max(total - paidAmount, 0),
        paymentStatus: saleDetails.payment_status,
        cashReceived: Number(saleDetails.cash_received || 0),
        changeAmount: Number(saleDetails.change_amount || 0),
        notes: saleDetails.notes,
        paymentAllocations: Array.isArray(payload?.payment_allocations) ? payload.payment_allocations : [],
        paymentMethod: saleDetails.payment_method === 'debt'
          ? 'credit'
          : saleDetails.payment_method === 'transfer'
            ? 'bank_transfer'
            : saleDetails.payment_method || 'cash',
      });
      setIsSalesHistoryOpen(false);
      setIsInvoiceOpen(true);
    } catch (error) {
      console.error('Failed to load sale invoice:', error);
      setSalesError('تعذر تحميل تفاصيل الفاتورة. حاول مرة أخرى.');
    } finally {
      setLoadingInvoice(false);
    }
  };

  const purchaseEntries = ledgerEntries.filter((entry) => {
    const type = String(entry.transaction_type || entry.type || '').toLowerCase();
    return type === 'sale' || type === 'debit';
  });
  const paymentEntries = ledgerEntries.filter((entry) => {
    const type = String(entry.transaction_type || entry.type || '').toLowerCase();
    return type === 'payment' || type === 'credit';
  });
  const ledgerPurchases = purchaseEntries.reduce((sum, entry) => sum + Math.max(Number(entry.amount) || 0, 0), 0);
  const ledgerPayments = paymentEntries.reduce((sum, entry) => sum + Math.max(Number(entry.amount) || 0, 0), 0);
  const selectedPurchases = Math.max(Number(selectedCustomer?.totalPurchases ?? 0), 0);
  const selectedOutstanding = Math.max(Number(selectedCustomer?.outstanding ?? 0), 0);
  const lastLedgerBalance = Number(ledgerEntries[ledgerEntries.length - 1]?.balance);
  const outstanding = Number.isFinite(lastLedgerBalance)
    ? Math.max(lastLedgerBalance, 0)
    : selectedOutstanding;
  const totalPaid = ledgerPayments > 0
    ? ledgerPayments
    : Math.max(selectedPurchases - outstanding, 0);
  const totalPurchases = Math.max(selectedPurchases, ledgerPurchases, totalPaid + outstanding);
  const totalTransactions = ledgerEntries.length;
  const paymentRate = totalPurchases > 0
    ? Math.min(Math.max((totalPaid / totalPurchases) * 100, 0), 100)
    : 0;
  const lastLedgerEntry = ledgerEntries[ledgerEntries.length - 1];
  const accountReconciles = Math.abs(totalPurchases - totalPaid - outstanding) < 0.01;
  return (
    <>
      {/* Add/Edit Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setEditingCustomer(null);
        }}
        title={editingCustomer ? 'تعديل العميل' : 'إضافة عميل جديد'}
        variant="modern"
        size="lg"
      >
        <CustomerForm
          initialData={editingCustomer}
          onSubmit={onSubmit}
          onCancel={() => {
            setIsModalOpen(false);
            setEditingCustomer(null);
          }}
        />
      </Modal>

      {/* View Modal */}
      <Modal
        isOpen={isViewModalOpen}
        onClose={() => setIsViewModalOpen(false)}
        title="تفاصيل العميل"
        variant="modern"
        size="xl"
      >
        {selectedCustomer && (
          <div className="space-y-md">
            {/* Practical Customer Profile (SALES-PHILOSOPHY.md) */}
            <div style={{
              background: 'linear-gradient(135deg, var(--color-primary-05) 0%, rgba(147, 51, 234, 0.05) 100%)',
              border: '1px solid var(--color-primary-10)',
              borderRadius: '12px',
              padding: '20px',
              marginBottom: '16px'
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '16px' }}>
                <div style={{
                  width: '48px',
                  height: '48px',
                  borderRadius: '12px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: 'linear-gradient(135deg, var(--color-primary-15) 0%, rgba(147, 51, 234, 0.15) 100%)',
                  border: '1px solid var(--color-primary-25)'
                }}>
                  <User className="w-6 h-6" style={{ color: 'var(--color-primary)' }} />
                </div>
                <div>
                  <h3 style={{ fontSize: '18px', fontWeight: '700', color: 'var(--text-primary)', marginBottom: '4px' }}>
                    {selectedCustomer.name}
                  </h3>
                  <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                    {selectedCustomer.phone || 'لا يوجد رقم هاتف'}
                  </p>
                </div>
              </div>

              {/* Key Metrics (SALES-PHILOSOPHY.md) */}
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '12px' }}>
                <div style={{
                  background: 'var(--bg-surface)',
                  padding: '12px',
                  borderRadius: '8px',
                  border: '1px solid var(--border-default)'
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                    <DollarSign className="w-4 h-4" style={{ color: 'var(--color-danger)' }} />
                    <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>الرصيد المستحق</span>
                  </div>
                  <div style={{ fontSize: '20px', fontWeight: '700', color: 'var(--color-danger)' }}>
                    ₪{(selectedCustomer.outstanding || 0).toLocaleString()}
                  </div>
                </div>

                <div style={{
                  background: 'var(--bg-surface)',
                  padding: '12px',
                  borderRadius: '8px',
                  border: '1px solid var(--border-default)'
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                    <ShoppingCart className="w-4 h-4" style={{ color: 'var(--color-info)' }} />
                    <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>إجمالي المشتريات</span>
                  </div>
                  <div style={{ fontSize: '20px', fontWeight: '700', color: 'var(--color-info)' }}>
                    ₪{(selectedCustomer.totalPurchases || 0).toLocaleString()}
                  </div>
                </div>

                <div style={{
                  background: 'var(--bg-surface)',
                  padding: '12px',
                  borderRadius: '8px',
                  border: '1px solid var(--border-default)'
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                    <CreditCard className="w-4 h-4" style={{ color: 'var(--color-success)' }} />
                    <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>إجمالي المدفوع</span>
                  </div>
                  <div style={{ fontSize: '20px', fontWeight: '700', color: 'var(--color-success)' }}>
                    ₪{((selectedCustomer.totalPurchases || 0) - (selectedCustomer.outstanding || 0)).toLocaleString()}
                  </div>
                </div>

                <div style={{
                  background: 'var(--bg-surface)',
                  padding: '12px',
                  borderRadius: '8px',
                  border: '1px solid var(--border-default)'
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                    <Calendar className="w-4 h-4" style={{ color: 'var(--color-info)' }} />
                    <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>آخر شراء</span>
                  </div>
                  <div style={{ fontSize: '16px', fontWeight: '600', color: 'var(--text-primary)' }}>
                    {selectedCustomer.lastPurchase ? new Date(selectedCustomer.lastPurchase).toLocaleDateString('ar-SA') : 'لا يوجد'}
                  </div>
                </div>
              </div>

              <div style={{
                marginTop: '16px',
                background: 'var(--bg-surface)',
                border: '1px solid var(--border-default)',
                borderRadius: '8px',
                overflow: 'hidden'
              }}>
                <div style={{ padding: '12px 14px', borderBottom: '1px solid var(--border-default)' }}>
                  <h4 style={{ fontSize: '14px', fontWeight: '700', color: 'var(--text-primary)' }}>
                    تفاصيل الحساب المالي
                  </h4>
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))' }}>
                  <div style={{ padding: '12px 14px', borderBottom: '1px solid var(--border-default)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', gap: '12px' }}>
                      <span style={{ color: 'var(--text-secondary)', fontSize: '12px' }}>المشتريات</span>
                      <strong style={{ color: 'var(--color-info)', fontSize: '14px' }}>₪{totalPurchases.toLocaleString()}</strong>
                    </div>
                    <div style={{ marginTop: '5px', color: 'var(--text-secondary)', fontSize: '11px' }}>
                      {purchaseEntries.length} حركة بيع مسجلة
                    </div>
                  </div>
                  <div style={{ padding: '12px 14px', borderBottom: '1px solid var(--border-default)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', gap: '12px' }}>
                      <span style={{ color: 'var(--text-secondary)', fontSize: '12px' }}>المدفوع</span>
                      <strong style={{ color: 'var(--color-success)', fontSize: '14px' }}>₪{totalPaid.toLocaleString()}</strong>
                    </div>
                    <div style={{ marginTop: '5px', color: 'var(--text-secondary)', fontSize: '11px' }}>
                      {paymentEntries.length} دفعة مسجلة
                    </div>
                  </div>
                  <div style={{ padding: '12px 14px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', gap: '12px' }}>
                      <span style={{ color: 'var(--text-secondary)', fontSize: '12px' }}>المتبقي المستحق</span>
                      <strong style={{ color: 'var(--color-danger)', fontSize: '14px' }}>₪{outstanding.toLocaleString()}</strong>
                    </div>
                    <div style={{ marginTop: '5px', color: 'var(--text-secondary)', fontSize: '11px' }}>
                      المشتريات - المدفوع = الرصيد المستحق
                    </div>
                  </div>
                  <div style={{ padding: '12px 14px', borderTop: '1px solid var(--border-default)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', gap: '12px' }}>
                      <span style={{ color: 'var(--text-secondary)', fontSize: '12px' }}>نسبة السداد</span>
                      <strong style={{ color: 'var(--color-success)', fontSize: '14px' }}>{paymentRate.toFixed(0)}%</strong>
                    </div>
                    <div style={{ height: '6px', marginTop: '8px', background: 'var(--border-default)', borderRadius: '999px', overflow: 'hidden' }}>
                      <div style={{ width: `${paymentRate}%`, height: '100%', background: 'var(--color-success)', borderRadius: '999px' }} />
                    </div>
                  </div>
                  <div style={{ padding: '12px 14px', borderTop: '1px solid var(--border-default)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', gap: '12px' }}>
                      <span style={{ color: 'var(--text-secondary)', fontSize: '12px' }}>إجمالي الحركات</span>
                      <strong style={{ color: 'var(--text-primary)', fontSize: '14px' }}>{totalTransactions}</strong>
                    </div>
                    <div style={{ marginTop: '5px', color: 'var(--text-secondary)', fontSize: '11px' }}>
                      {lastLedgerEntry?.created_at
                        ? `آخر حركة: ${new Date(lastLedgerEntry.created_at).toLocaleDateString('ar-SA')}`
                        : 'لا توجد حركات محمّلة'}
                    </div>
                  </div>
                  <div style={{ padding: '12px 14px', borderTop: '1px solid var(--border-default)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', gap: '12px' }}>
                      <span style={{ color: 'var(--text-secondary)', fontSize: '12px' }}>حالة الحساب</span>
                      <strong style={{ color: accountReconciles ? 'var(--color-success)' : 'var(--color-warning)', fontSize: '14px' }}>
                        {accountReconciles ? 'متطابق' : 'يحتاج مراجعة'}
                      </strong>
                    </div>
                    <div style={{ marginTop: '5px', color: 'var(--text-secondary)', fontSize: '11px' }}>
                      يتم التحقق من إجمالي المشتريات والمدفوع والرصيد
                    </div>
                  </div>
                </div>
              </div>

              {/* Quick Actions (SALES-PHILOSOPHY.md) */}
              <div style={{ display: 'flex', gap: '8px', marginTop: '16px' }}>
                <Button
                  variant="primary"
                  size={getButtonSize('customers', 'modalAction')}
                  style={{ flex: 1 }}
                  onClick={() => navigate(`/app/debts?customer_id=${encodeURIComponent(selectedCustomer.id)}`)}
                >
                  <DollarSign className="w-4 h-4 mr-2" />
                  تسجيل دفعة
                </Button>
                <Button
                  variant="secondary"
                  size={getButtonSize('customers', 'modalAction')}
                  style={{ flex: 1 }}
                  onClick={() => navigate(`/app/customers/${selectedCustomer.id}/purchases`)}
                >
                  <FileText className="w-4 h-4 mr-2" />
                  عرض المشتريات
                </Button>
                <Button
                  variant="secondary"
                  size={getButtonSize('customers', 'modalAction')}
                  style={{ flex: 1 }}
                  onClick={() => navigate(`/app/returns/create?customer_id=${encodeURIComponent(selectedCustomer.id)}`)}
                >
                  <RotateCcw className="w-4 h-4 mr-2" />
                  إنشاء مرتجع
                </Button>
              </div>

              {/* Credit Limit Warning (SALES-PHILOSOPHY.md) */}
              {selectedCustomer.credit_limit && selectedCustomer.outstanding > selectedCustomer.credit_limit && (
                <div style={{
                  marginTop: '12px',
                  padding: '10px',
                  background: 'rgba(251, 191, 36, 0.1)',
                  border: '1px solid rgba(251, 191, 36, 0.3)',
                  borderRadius: '8px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px'
                }}>
                  <AlertTriangle className="w-4 h-4" style={{ color: 'var(--color-warning)' }} />
                  <span style={{ fontSize: '12px', color: 'var(--color-warning)', fontWeight: '500' }}>
                    تنبيه: العميل تجاوز حد الدين المحدد
                  </span>
                </div>
              )}
            </div>

            {/* Financial Timeline */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
              <History className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
              <h4 style={{ fontSize: '16px', fontWeight: '600', color: 'var(--text-primary)' }}>
                السجل المالي
              </h4>
            </div>
            <Suspense fallback={
              <div style={{ 
                display: 'flex', 
                justifyContent: 'center', 
                alignItems: 'center', 
                padding: '40px',
                minHeight: '200px'
              }}>
                <div className="animate-spin rounded-full border-2 border-text-muted/20 border-t-primary" 
                     style={{ width: '40px', height: '40px' }} />
              </div>
            }>
              <FinancialTimeline
                transactions={ledgerEntries.map((entry) => ({
                  id: entry.id,
                  type: entry.transaction_type || entry.type || 'other',
                  amount: Number(entry.amount ?? 0),
                  balance_after: Number(entry.balance ?? entry.previous_balance ?? 0),
                  date: entry.created_at || entry.date || new Date().toISOString(),
                  description: entry.description || 'حركة مالية',
                  status: entry.status || 'completed',
                }))}
                loading={loadingLedger}
                showBalance={true}
                showTitle={false}
              />
            </Suspense>

            <div className="flex gap-sm justify-end">
              <Button variant="secondary" size={getButtonSize('customers', 'modalAction')} onClick={() => setIsViewModalOpen(false)}>
                إغلاق
              </Button>
            </div>
          </div>
        )}
      </Modal>

      <Modal
        isOpen={isSalesHistoryOpen}
        onClose={() => setIsSalesHistoryOpen(false)}
        title={`فواتير ${selectedCustomer?.name || 'العميل'}`}
        variant="modern"
        size="xl"
      >
        {loadingSales ? (
          <div className="flex min-h-40 items-center justify-center">
            <div className="animate-spin rounded-full border-2 border-text-muted/20 border-t-primary" style={{ width: '36px', height: '36px' }} />
          </div>
        ) : salesError ? (
          <div className="rounded-lg border border-red-500/20 bg-red-500/10 p-4 text-center text-red-600">{salesError}</div>
        ) : customerSales.length === 0 ? (
          <div className="rounded-lg border border-border p-6 text-center text-text-secondary">لا توجد فواتير مسجلة لهذا العميل.</div>
        ) : (
          <div className="space-y-3">
            <div className="rounded-lg border border-border bg-surface-elevated px-4 py-3 text-sm text-text-secondary">
              إجمالي الفواتير المعروضة: <strong className="text-text-primary">{customerSales.length}</strong>
            </div>
            {customerSales.map((sale) => {
              const total = Number(sale.total_amount ?? 0);
              const paid = Number(sale.paid_amount ?? 0);
              const remaining = Math.max(total - paid, 0);
              return (
                <button
                  key={sale.id}
                  type="button"
                  onClick={() => openSaleInvoice(sale)}
                  disabled={loadingInvoice}
                  className="flex w-full items-center justify-between rounded-lg border border-border bg-surface p-4 text-right transition hover:border-primary hover:bg-surface-elevated disabled:opacity-60"
                >
                  <span className="min-w-0">
                    <span className="block font-semibold text-text-primary">{sale.invoice_number || sale.id}</span>
                    <span className="mt-1 block text-sm text-text-secondary">
                      {sale.sale_date ? new Date(sale.sale_date).toLocaleDateString('ar-SA') : 'بدون تاريخ'}
                    </span>
                  </span>
                  <span className="text-left">
                    <span className="block font-semibold text-text-primary">الإجمالي: ₪{total.toLocaleString()}</span>
                    <span className="mt-1 block text-sm text-green-600">المدفوع: ₪{paid.toLocaleString()}</span>
                    <span className={`mt-1 block text-sm ${remaining > 0 ? 'text-red-600' : 'text-text-secondary'}`}>
                      المتبقي: ₪{remaining.toLocaleString()}
                    </span>
                    <span className="mt-1 block text-xs text-text-secondary">
                      {remaining > 0 ? 'متبقي عليها مبلغ' : 'مدفوعة بالكامل'}
                    </span>
                  </span>
                </button>
              );
            })}
          </div>
        )}
      </Modal>

      <Modal
        isOpen={isInvoiceOpen}
        onClose={() => setIsInvoiceOpen(false)}
        title="فاتورة البيع"
        variant="modern"
        size="xl"
      >
        {invoiceData && (
          <UsedPartsInvoice
            saleData={invoiceData}
            onClose={() => setIsInvoiceOpen(false)}
          />
        )}
      </Modal>
    </>
  );
}