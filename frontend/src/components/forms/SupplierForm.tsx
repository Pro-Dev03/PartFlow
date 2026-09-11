import { useState } from 'react';
import { Input } from '../ui/input';
import { Button } from '../ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';

export interface SupplierFormData {
  code: string;
  name: string;
  phone: string;
  email?: string;
  address?: string;
  notes?: string;
  credit_limit?: number;
  is_active?: boolean;
}

interface SupplierFormProps {
  onSubmit: (data: SupplierFormData) => void;
  onCancel: () => void;
  initialData?: Partial<SupplierFormData>;
}

const generateSupplierCode = () => `SUP-${Math.random().toString(36).slice(2, 10).toUpperCase()}`;

export function SupplierForm({ onSubmit, onCancel, initialData }: SupplierFormProps) {
  const [formData, setFormData] = useState<SupplierFormData>({
    code: initialData?.code || generateSupplierCode(),
    name: initialData?.name || '',
    phone: initialData?.phone || '',
    email: initialData?.email || '',
    address: initialData?.address || '',
    notes: initialData?.notes || '',
    credit_limit: initialData?.credit_limit || 0,
    is_active: initialData?.is_active !== undefined ? initialData.is_active : true,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{initialData ? 'تعديل المورد' : 'إضافة مورد جديد'}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="كود المورد"
            value={formData.code}
            onChange={(e) => setFormData({ ...formData, code: e.target.value })}
            required
          />
          <Input
            label="الاسم"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
          />

          <Input
            label="رقم الهاتف"
            type="tel"
            value={formData.phone}
            onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
            required
          />

          <Input
            label="البريد الإلكتروني"
            type="email"
            value={formData.email}
            onChange={(e) => setFormData({ ...formData, email: e.target.value })}
          />

          <Input
            label="العنوان"
            value={formData.address}
            onChange={(e) => setFormData({ ...formData, address: e.target.value })}
          />

          <Input
            label="حد الائتمان"
            type="number"
            value={formData.credit_limit || 0}
            onChange={(e) => setFormData({ ...formData, credit_limit: parseFloat(e.target.value) || 0 })}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              ملاحظات
            </label>
            <textarea
              value={formData.notes}
              onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-800 dark:text-gray-100"
            />
          </div>

          <div className="flex gap-3 justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={onCancel}
            >
              إلغاء
            </Button>
            <Button type="submit">
              {initialData ? 'حفظ التغييرات' : 'إضافة المورد'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
