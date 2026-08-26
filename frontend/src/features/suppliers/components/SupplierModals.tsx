import { useQuery } from '@tanstack/react-query';
import { Modal } from '../../../components/ui/modal';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { SupplierForm, type SupplierFormData } from '../../../components/forms/SupplierForm';
import { suppliersApi } from '../../../services/api/endpoints';
import { getButtonSize } from '../../../config/button-sizes';

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

  return (
    <>
      {/* Add/Edit Modal */}
      <Modal
        isOpen={isOpen}
        onClose={close}
        title={editingSupplier ? 'تعديل المورد' : 'إضافة مورد جديد'}
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
        title="تفاصيل المورد"
        variant="modern"
        size="xl"
      >
        {viewingSupplier && (
          <div className="space-y-md">
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '14px' }}>
              <div>
                <label className="text-small font-medium text-text mb-sm block">الاسم</label>
                <Input value={viewingSupplier.name || ''} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">الهاتف</label>
                <Input value={viewingSupplier.phone || ''} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">البريد الإلكتروني</label>
                <Input value={viewingSupplier.email || '-'} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">إجمالي المشتريات</label>
                <Input value={`₪${(viewingSupplier.totalPurchases || 0).toLocaleString()}`} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">المدفوع</label>
                <Input value={`₪${(viewingSupplier.paidAmount || 0).toLocaleString()}`} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">المستحق</label>
                <Input value={`₪${(viewingSupplier.outstanding || 0).toLocaleString()}`} disabled />
              </div>
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
                            {entry.created_at ? new Date(entry.created_at).toLocaleDateString('ar-SA') : '-'}
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
                <div className="text-center py-6 text-text-muted">لا توجد حركات لهذا المورد</div>
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
