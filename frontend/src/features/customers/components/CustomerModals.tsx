import { useState, lazy, Suspense, useEffect } from 'react';
import { Modal } from '../../../components/ui/modal';
import { Button } from '../../../components/ui/button';
import { CustomerForm, type CustomerFormData } from '../../../components/forms/CustomerForm';
import { getButtonSize } from '../../../config/button-sizes';
import { Customer } from '../types/customers.types';
import { User, DollarSign, ShoppingCart, Calendar, CreditCard, FileText, History, AlertTriangle } from 'lucide-react';
import { LedgerEntry } from '../../../components/ui/financial-timeline';
import { customersApi, salesApi } from '../../../services/api/endpoints';
import { UsedPartsInvoice } from '../../../components/invoice/UsedPartsInvoice';

// Lazy load heavy FinancialTimeline component
const FinancialTimeline = lazy(() => import('../../../components/ui/financial-timeline').then(m => ({ default: m.FinancialTimeline })));

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
        const formattedEntries = response.data.map((entry: any) => ({
          id: entry.id,
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
        id: saleDetails.invoice_number || sale.invoice_number || saleDetails.id,
        customerName: selectedCustomer?.name || '',
        customerPhone: selectedCustomer?.phone,
        saleDate: saleDetails.sale_date || sale.created_at || new Date().toISOString(),
        items: items.map((item: any) => ({
          name: item.product_name || item.name || 'منتج',
          condition: item.condition || 'NEW',
          grade: item.grade,
          sellingPrice: Number(item.unit_price ?? item.selling_price ?? 0),
          quantity: Number(item.quantity ?? 1),
          total: Number(item.total_amount ?? item.total ?? 0),
          partType: item.part_type || item.partType,
          partTypeColor: item.part_type_color || item.partTypeColor,
        })),
        subtotal: Number(saleDetails.subtotal ?? total),
        total,
        paidAmount,
        remaining: Math.max(total - paidAmount, 0),
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

              {/* Quick Actions (SALES-PHILOSOPHY.md) */}
              <div style={{ display: 'flex', gap: '8px', marginTop: '16px' }}>
                <Button
                  variant="primary"
                  size={getButtonSize('customers', 'modalAction')}
                  style={{ flex: 1 }}
                >
                  <DollarSign className="w-4 h-4 mr-2" />
                  تسجيل دفعة
                </Button>
                <Button
                  variant="secondary"
                  size={getButtonSize('customers', 'modalAction')}
                  style={{ flex: 1 }}
                  onClick={loadCustomerSales}
                >
                  <ShoppingCart className="w-4 h-4 mr-2" />
                  بيع جديد
                </Button>
                <Button
                  variant="secondary"
                  size={getButtonSize('customers', 'modalAction')}
                  style={{ flex: 1 }}
                >
                  <FileText className="w-4 h-4 mr-2" />
                  عرض المشتريات
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
            {customerSales.map((sale) => {
              const total = Number(sale.total_amount ?? 0);
              const paid = Number(sale.paid_amount ?? 0);
              return (
                <button
                  key={sale.id}
                  type="button"
                  onClick={() => openSaleInvoice(sale)}
                  disabled={loadingInvoice}
                  className="flex w-full items-center justify-between rounded-lg border border-border bg-surface p-4 text-right transition hover:border-primary hover:bg-surface-elevated disabled:opacity-60"
                >
                  <span>
                    <span className="block font-semibold text-text-primary">{sale.invoice_number || sale.id}</span>
                    <span className="mt-1 block text-sm text-text-secondary">
                      {sale.sale_date ? new Date(sale.sale_date).toLocaleDateString('ar-SA') : 'بدون تاريخ'}
                    </span>
                  </span>
                  <span>
                    <span className="block font-semibold text-text-primary">₪{total.toLocaleString()}</span>
                    <span className="mt-1 block text-sm text-text-secondary">المتبقي: ₪{Math.max(total - paid, 0).toLocaleString()}</span>
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
            onPrint={() => window.print()}
            onClose={() => setIsInvoiceOpen(false)}
          />
        )}
      </Modal>
    </>
  );
}