import { useState } from 'react';
import { Input } from '../../design-system/components/input';
import { Button } from '../../design-system/components/button';
import { CustomerFormData } from '../../features/customers/types/customers.types';
import './CustomerForm.css';

const generateCustomerCode = () => `CUST-${Math.random().toString(36).slice(2, 10).toUpperCase()}`;

interface CustomerFormProps {
  onSubmit: (data: CustomerFormData) => void;
  onCancel: () => void;
  initialData?: Partial<CustomerFormData>;
  isSubmitting?: boolean;
}

export function CustomerForm({ onSubmit, onCancel, initialData, isSubmitting = false }: CustomerFormProps) {
  const [formData, setFormData] = useState<CustomerFormData>({
    code: initialData?.code || generateCustomerCode(),
    name: initialData?.name || '',
    phone: initialData?.phone || '',
    email: initialData?.email || '',
    address: initialData?.address || '',
    notes: initialData?.notes || '',
    debt_reason: initialData?.debt_reason || '',
    credit_limit: initialData?.credit_limit || 0,
    opening_debt: 0,
    is_active: initialData?.is_active !== undefined ? initialData.is_active : true,
  });

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    onSubmit(formData);
  };

  return (
    <form onSubmit={handleSubmit} className="customer-form">
      <div className="customer-form__grid">
        <Input
          label="كود العميل"
          value={formData.code}
          onChange={(event) => setFormData({ ...formData, code: event.target.value })}
        />

        <Input
          label="الاسم الكامل"
          value={formData.name}
          onChange={(event) => setFormData({ ...formData, name: event.target.value })}
          required
          autoComplete="name"
        />

        <Input
          label="رقم الهاتف"
          type="tel"
          value={formData.phone}
          onChange={(event) => setFormData({ ...formData, phone: event.target.value })}
          required
          autoComplete="tel"
        />

        <Input
          label="البريد الإلكتروني"
          type="email"
          value={formData.email}
          onChange={(event) => setFormData({ ...formData, email: event.target.value })}
          autoComplete="email"
        />

        <div className={initialData ? 'customer-form__field--wide' : undefined}>
          <Input
            label="العنوان"
            value={formData.address}
            onChange={(event) => setFormData({ ...formData, address: event.target.value })}
            autoComplete="street-address"
          />
        </div>

        {!initialData && (
          <Input
            label="مبلغ سابق مستحق على العميل"
            type="number"
            min="0"
            step="0.01"
            value={formData.opening_debt || ''}
            onChange={(event) => setFormData({ ...formData, opening_debt: Number(event.target.value) || 0 })}
            placeholder="0.00"
          />
        )}

        <div className="customer-form__textarea-field">
          <label className="customer-form__label" htmlFor="customer-notes">ملاحظات العميل</label>
          <textarea
            id="customer-notes"
            value={formData.notes || ''}
            onChange={(event) => setFormData({ ...formData, notes: event.target.value })}
            rows={3}
          />
        </div>

        <div className="customer-form__textarea-field">
          <label className="customer-form__label" htmlFor="customer-debt-reason">سبب الدين</label>
          <textarea
            id="customer-debt-reason"
            value={formData.debt_reason || ''}
            onChange={(event) => setFormData({ ...formData, debt_reason: event.target.value })}
            rows={3}
            placeholder="اكتب سبب الدين أو المذكرة المالية للعميل"
          />
        </div>
      </div>

      <div className="customer-form__actions">
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'جاري الحفظ...' : initialData ? 'حفظ التغييرات' : 'إضافة العميل'}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>
          إلغاء
        </Button>
      </div>
    </form>
  );
}
