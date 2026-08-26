import { Modal } from '../../../components/ui/modal';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { CustomerForm, type CustomerFormData } from '../../../components/forms/CustomerForm';
import { FinancialTimeline, LedgerEntry } from '../../../components/ui/financial-timeline';
import { getButtonSize } from '../../../config/button-sizes';
import { Customer } from '../types/customers.types';
import { User, Sparkles, DollarSign, ShoppingCart, Calendar, CreditCard, FileText, History, AlertTriangle } from 'lucide-react';

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
  ledgerEntries?: LedgerEntry[];
  loadingLedger?: boolean;
}

export function CustomerModals({
  isModalOpen,
  setIsModalOpen,
  isViewModalOpen,
  setIsViewModalOpen,
  editingCustomer,
  setEditingCustomer,
  selectedCustomer,
  setSelectedCustomer,
  onSubmit,
  ledgerEntries = [],
  loadingLedger = false,
}: CustomerModalsProps) {
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
            <FinancialTimeline
              entries={ledgerEntries}
              loading={loadingLedger}
              showBalance={true}
            />

            <div className="flex gap-sm justify-end">
              <Button variant="secondary" size={getButtonSize('customers', 'modalAction')} onClick={() => setIsViewModalOpen(false)}>
                إغلاق
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </>
  );
}