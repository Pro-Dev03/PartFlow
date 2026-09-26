import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Modal } from '../../../design-system/components/modal';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { SupplierForm, type SupplierFormData } from '../../../components/forms/SupplierForm';
import { suppliersApi } from '../../../services/api/endpoints';
import { getButtonSize } from '../../../config/button-sizes';
import { normalizeSupplier } from '../utils/supplier-normalization';
import { formatStoreDate } from '../../../utils/store-time';

interface SupplierModalsProps {
  isOpen: boolean;
  setIsOpen: (open: boolean) => void;
  editingSupplier: any | null;
  setEditingSupplier: (supplier: any | null) => void;
  onSubmit: (data: SupplierFormData) => void;
  isViewModalOpen: boolean;
  setIsViewModalOpen: (open: boolean) => void;
  viewingSupplier: any | null;
}

export function SupplierModals({
  isOpen,
  setIsOpen,
  editingSupplier,
  setEditingSupplier,
  onSubmit,
  isViewModalOpen,
  setIsViewModalOpen,
  viewingSupplier,
}: SupplierModalsProps) {
  const queryClient = useQueryClient();
  const [paymentAmount, setPaymentAmount] = useState('');
  const close = () => {
    setIsOpen(false);
    setEditingSupplier(null);
  };

  const { data: ledgerData, isLoading: ledgerLoading } = useQuery({
    queryKey: ['supplier-ledger', viewingSupplier?.id],
    queryFn: () => suppliersApi.ledger(viewingSupplier.id),
    enabled: isViewModalOpen && !!viewingSupplier?.id,
  });

  const ledger = (ledgerData?.data as any) || null;
  const entries: any[] = Array.isArray(ledger?.entries) ? ledger.entries : [];
  const normalizedViewingSupplier = viewingSupplier ? normalizeSupplier(viewingSupplier) : null;
  const paymentMutation = useMutation({
    mutationFn: () => suppliersApi.addPayment(viewingSupplier.id, { amount: Number(paymentAmount), method: 'cash' }),
    onSuccess: () => {
      setPaymentAmount('');
      void queryClient.invalidateQueries({ queryKey: ['supplier-ledger', viewingSupplier?.id] });
      void queryClient.invalidateQueries({ queryKey: ['suppliers'] });
    },
  });

  return (
    <>
      {/* Add/Edit Modal */}
      <Modal
        isOpen={isOpen}
        onClose={close}
        title={editingSupplier ? 'تعديل التاجر' : 'إضافة تاجر جديد'}
        variant="modern"
        size="lg"
      >
        <SupplierForm
          initialData={editingSupplier || undefined}
          onSubmit={onSubmit}
          onCancel={close}
        />
      </Modal>

      {/* View Modal */}
      <Modal
        isOpen={isViewModalOpen}
        onClose={() => setIsViewModalOpen(false)}
        title="تفاصيل التاجر"
        variant="modern"
        size="xl"
      >
        {normalizedViewingSupplier && (
          <div className="space-y-md">
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '14px' }}>
              <div>
                <label className="text-small font-medium text-text mb-sm block">الاسم</label>
                <Input value={normalizedViewingSupplier.name || ''} readOnly />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">الهاتف</label>
                <Input value={normalizedViewingSupplier.phone || ''} readOnly />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">البريد الإلكتروني</label>
                <Input value={normalizedViewingSupplier.email || '-'} readOnly />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">إجمالي المشتريات</label>
                <Input value={`₪${(normalizedViewingSupplier.totalPurchases || 0).toLocaleString()}`} readOnly />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">المدفوع</label>
                <Input value={`₪${(normalizedViewingSupplier.paidAmount || 0).toLocaleString()}`} readOnly />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">صافي المستحق</label>
                <Input value={`₪${(normalizedViewingSupplier.outstanding || 0).toLocaleString()}`} readOnly />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">الرصيد الدائن</label>
                <Input value={`₪${(normalizedViewingSupplier.creditBalance || 0).toLocaleString()}`} readOnly />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">رصيد مرتجعات التاجر</label>
                <Input value={`-₪${Number(ledger?.supplier_return_credits || 0).toLocaleString()}`} readOnly />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">دفعات التاجر</label>
                <Input value={`₪${Number(ledger?.supplier_payments || 0).toLocaleString()}`} readOnly />
              </div>
            </div>
            <p className="text-xs text-text-muted">صافي المستحق = المستحق الأصلي - Credits المرتجعات - دفعات التاجر.</p>
            <div className="flex items-end gap-2 border-t border-border pt-4">
              <Input type="number" min="0.01" step="0.01" value={paymentAmount} onChange={(event) => setPaymentAmount(event.target.value)} placeholder="مبلغ الدفعة" />
              <Button onClick={() => paymentMutation.mutate()} disabled={paymentMutation.isPending || Number(paymentAmount) <= 0 || Number(paymentAmount) > Number(ledger?.current_balance ?? normalizedViewingSupplier.outstanding ?? 0)}>
                تسجيل دفعة
              </Button>
            </div>

            <div>
              <h4 className="text-small font-semibold text-text mb-2">دفتر الحركات</h4>
              {ledgerLoading ? (
                <div className="text-center py-6 text-text-muted">جاري التحميل...</div>
              ) : entries.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="w-full text-small">
                    <thead>
                      <tr className="border-b border-border text-text-secondary">
                        <th className="text-start p-2">التاريخ</th>
                        <th className="text-start p-2">النوع</th>
                        <th className="text-start p-2">الوصف</th>
                        <th className="text-end p-2">المبلغ</th>
                        <th className="text-end p-2">الرصيد</th>
                      </tr>
                    </thead>
                    <tbody>
                      {entries.map((entry: any, idx: number) => (
                        <tr key={entry.id || idx} className="border-b border-border">
                          <td className="p-2 text-text-secondary">
                            {entry.created_at ? formatStoreDate(entry.created_at, 'ar-SA') : '-'}
                          </td>
                          <td className="p-2">
                            <span className={entry.entry_type === 'credit' || entry.type === 'credit' ? 'text-green' : 'text-danger'}>
                              {entry.entry_type || entry.type || '-'}
                            </span>
                          </td>
                          <td className="p-2 text-text-secondary">{entry.description || entry.reference || '-'}</td>
                          <td className="p-2 text-end text-text-primary">
                            ₪{((entry.amount || 0)).toLocaleString()}
                          </td>
                          <td className="p-2 text-end text-text-secondary">
                            ₪{((entry.balance || 0)).toLocaleString()}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-center py-6 text-text-muted">لا توجد حركات لهذا التاجر</div>
              )}
            </div>

            <div className="flex gap-sm justify-end">
              <Button
                variant="secondary"
                size={getButtonSize('suppliers', 'modalAction')}
                onClick={() => setIsViewModalOpen(false)}
              >
                إغلاق
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </>
  );
}
