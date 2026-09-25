import { useState } from 'react';
import { Input } from '../../design-system/components/input';
import { Button } from '../../design-system/components/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../design-system/components/card';
import { CustomerFormData } from '../../features/customers/types/customers.types';

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

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{initialData ? 'تعديل العميل' : 'إضافة عميل جديد'}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <Input
                label="كود العميل"
                value={formData.code}
                onChange={(e) => setFormData({ ...formData, code: e.target.value })}
              />
            </div>

            <div>
              <Input
                label="الاسم الكامل"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>

            <div>
              <Input
                label="رقم الهاتف"
                type="tel"
                value={formData.phone}
                onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                required
              />
            </div>

            <div>
              <Input
                label="البريد الإلكتروني"
                type="email"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              />
            </div>

            <div className="col-span-2 grid grid-cols-1 items-start gap-4 sm:grid-cols-2 xl:grid-cols-4">
              <Input
                label="العنوان"
                value={formData.address}
                onChange={(e) => setFormData({ ...formData, address: e.target.value })}
              />

              {!initialData && (
                <Input
                  label="مبلغ سابق مستحق على العميل"
                  type="number"
                  min="0"
                  step="0.01"
                  value={formData.opening_debt || ''}
                  onChange={(e) => setFormData({ ...formData, opening_debt: Number(e.target.value) || 0 })}
                  placeholder="0"
                />
              )}

              <div className="xl:col-span-2">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  ملاحظات العميل
                </label>
                <textarea
                  value={formData.notes || ''}
                  onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                  rows={2}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-800 dark:text-gray-100"
                />
              </div>

              <div className="xl:col-span-2">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  سبب الدين
                </label>
                <textarea
                  value={formData.debt_reason || ''}
                  onChange={(e) => setFormData({ ...formData, debt_reason: e.target.value })}
                  rows={2}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-800 dark:text-gray-100"
                  placeholder="اكتب سبب الدين أو المذكرة المالية للعميل"
                />
              </div>
            </div>
          </div>

          <div className="flex gap-3 justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={onCancel}
              disabled={isSubmitting}
            >
              إلغاء
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'جاري الحفظ...' : initialData ? 'حفظ التغييرات' : 'إضافة العميل'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
