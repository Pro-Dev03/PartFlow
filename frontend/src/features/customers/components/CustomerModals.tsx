import { Modal } from '../../../components/ui/modal';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { CustomerForm, type CustomerFormData } from '../../../components/forms/CustomerForm';
import { FinancialTimeline, LedgerEntry } from '../../../components/ui/financial-timeline';
import { getButtonSize } from '../../../config/button-sizes';
import { Customer } from '../types/customers.types';
import { User, Sparkles } from 'lucide-react';

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
            {/* Customer Info */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '14px' }}>
              <div>
                <label className="text-small font-medium text-text mb-sm block">الاسم</label>
                <Input value={selectedCustomer.name || ''} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">الهاتف</label>
                <Input value={selectedCustomer.phone || ''} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">البريد الإلكتروني</label>
                <Input value={selectedCustomer.email || '-'} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">إجمالي المشتريات</label>
                <Input value={`₪${selectedCustomer.totalPurchases?.toLocaleString() || 0}`} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">الديون</label>
                <Input value={`₪${selectedCustomer.outstanding?.toLocaleString() || 0}`} disabled />
              </div>
            </div>

            {/* Financial Timeline */}
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