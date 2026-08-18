import { useState } from 'react';
import { Input } from '../ui/input';
import { Button } from '../ui/button';
import { Select } from '../ui/select';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';

interface ProductFormProps {
  onSubmit: (data: ProductFormData) => void;
  onCancel: () => void;
  initialData?: Partial<ProductFormData>;
}

export interface ProductFormData {
  name: string;
  sku: string;
  category: string;
  description?: string;
  buyingPrice: number;
  sellingPrice: number;
  condition: 'new' | 'used' | 'refurbished' | 'parts_only';
  warrantyPeriod?: number;
}

export function ProductForm({ onSubmit, onCancel, initialData }: ProductFormProps) {
  const [formData, setFormData] = useState<ProductFormData>({
    name: initialData?.name || '',
    sku: initialData?.sku || '',
    category: initialData?.category || '',
    description: initialData?.description || '',
    buyingPrice: initialData?.buyingPrice || 0,
    sellingPrice: initialData?.sellingPrice || 0,
    condition: initialData?.condition || 'new',
    warrantyPeriod: initialData?.warrantyPeriod || 12,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{initialData ? 'تعديل المنتج' : 'إضافة منتج جديد'}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="اسم المنتج"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
          />

          <Input
            label="رمز المنتج (SKU)"
            value={formData.sku}
            onChange={(e) => setFormData({ ...formData, sku: e.target.value })}
            required
          />

          <Select
            label="الفئة"
            value={formData.category}
            onChange={(e) => setFormData({ ...formData, category: e.target.value })}
            options={[
              { value: 'gpu', label: 'كروت شاشة' },
              { value: 'cpu', label: 'معالجات' },
              { value: 'ram', label: 'ذاكرة' },
              { value: 'storage', label: 'تخزين' },
              { value: 'motherboard', label: 'لوحات أم' },
              { value: 'power', label: 'مزودات طاقة' },
              { value: 'case', label: 'علب' },
              { value: 'other', label: 'أخرى' },
            ]}
            required
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              الوصف
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-800 dark:text-gray-100"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <Input
              label="سعر الشراء"
              type="number"
              value={formData.buyingPrice}
              onChange={(e) => setFormData({ ...formData, buyingPrice: parseFloat(e.target.value) || 0 })}
              required
            />

            <Input
              label="سعر البيع"
              type="number"
              value={formData.sellingPrice}
              onChange={(e) => setFormData({ ...formData, sellingPrice: parseFloat(e.target.value) || 0 })}
              required
            />
          </div>

          <Select
            label="الحالة"
            value={formData.condition}
            onChange={(e) => setFormData({ ...formData, condition: e.target.value as any })}
            options={[
              { value: 'new', label: 'جديد' },
              { value: 'used', label: 'مستعمل' },
              { value: 'refurbished', label: 'مجدد' },
              { value: 'parts_only', label: 'قطع غيار فقط' },
            ]}
            required
          />

          <Input
            label="فترة الضمان (شهور)"
            type="number"
            value={formData.warrantyPeriod}
            onChange={(e) => setFormData({ ...formData, warrantyPeriod: parseInt(e.target.value) || 0 })}
          />

          <div className="flex gap-3 justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={onCancel}
            >
              إلغاء
            </Button>
            <Button type="submit">
              {initialData ? 'حفظ التغييرات' : 'إضافة المنتج'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}