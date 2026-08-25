import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { partTypesApi, specificationsApi, typeSpecsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { PageHeader } from '../../../components/ui/page-header';
import { Modal } from '../../../components/ui/modal';
import { Badge } from '../../../components/ui/badge';
import { getButtonSize } from '../../../config/button-sizes';
import { toast } from 'sonner';
import { 
  Plus, 
  Edit, 
  Trash2, 
  Settings,
  Monitor,
  Cpu,
  HardDrive,
  Zap,
  Box,
  Thermometer,
  Package
} from 'lucide-react';

const iconMap: Record<string, any> = {
  monitor: Monitor,
  cpu: Cpu,
  'hard-drive': HardDrive,
  zap: Zap,
  box: Box,
  thermometer: Thermometer,
  package: Package,
};

export function PartTypesPage() {
  const queryClient = useQueryClient();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [selectedPartType, setSelectedPartType] = useState<any>(null);
  const [newPartType, setNewPartType] = useState({
    name_ar: '',
    name_en: '',
    icon: 'box',
    color: '#14b8a6',
    sort_order: 0
  });

  const { data: partTypesData, isLoading } = useQuery({
    queryKey: ['part-types'],
    queryFn: () => partTypesApi.list(),
  });

  const partTypes = (partTypesData?.data as any[]) || [];

  const createMutation = useMutation({
    mutationFn: (data: any) => partTypesApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['part-types'] });
      setIsCreateModalOpen(false);
      setNewPartType({ name_ar: '', name_en: '', icon: 'box', color: '#14b8a6', sort_order: 0 });
      toast.success('تم إضافة نوع القطعة بنجاح');
    },
    onError: () => {
      toast.error('فشل إضافة نوع القطعة');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) => partTypesApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['part-types'] });
      setIsEditModalOpen(false);
      setSelectedPartType(null);
      toast.success('تم تحديث نوع القطعة بنجاح');
    },
    onError: () => {
      toast.error('فشل تحديث نوع القطعة');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => partTypesApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['part-types'] });
      toast.success('تم حذف نوع القطعة بنجاح');
    },
    onError: () => {
      toast.error('فشل حذف نوع القطعة');
    },
  });

  const handleCreate = () => {
    if (!newPartType.name_ar || !newPartType.name_en) {
      toast.error('يرجى ملء جميع الحقول المطلوبة');
      return;
    }
    createMutation.mutate(newPartType);
  };

  const handleUpdate = () => {
    if (!selectedPartType?.name_ar || !selectedPartType?.name_en) {
      toast.error('يرجى ملء جميع الحقول المطلوبة');
      return;
    }
    updateMutation.mutate({
      id: selectedPartType.id,
      data: selectedPartType
    });
  };

  const handleDelete = (id: string) => {
    if (window.confirm('هل أنت متأكد من حذف هذا النوع؟')) {
      deleteMutation.mutate(id);
    }
  };

  const getIconComponent = (iconName: string) => {
    const IconComponent = iconMap[iconName] || Box;
    return <IconComponent style={{ width: '20px', height: '20px' }} />;
  };

  return (
    <div>
      <PageHeader
        eyebrow="Part Types Management"
        title="إدارة أنواع القطع"
        description="إدارة أنواع القطع المستعملة ومواصفاتها"
        actions={
          <Button
            variant="primary"
            size={getButtonSize('parttypes', 'headerActions')}
            onClick={() => setIsCreateModalOpen(true)}
          >
            <Plus className="w-3.5 h-3.5 me-1.5" />
            إضافة نوع جديد
          </Button>
        }
      />

      <Card>
        <CardHeader>
          <CardTitle>الأنواع المتاحة</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
          ) : partTypes.length === 0 ? (
            <div className="text-center py-12 text-gray-400">
              <Package className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>لا توجد أنواع قطع حالياً</p>
              <Button
                variant="primary"
                size={getButtonSize('parttypes', 'modalAction')}
                onClick={() => setIsCreateModalOpen(true)}
                className="mt-4"
              >
                إضافة نوع جديد
              </Button>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {partTypes.map((partType: any) => (
                <div
                  key={partType.id}
                  className="p-4 rounded-lg border border-gray-700 hover:border-cyan-500 transition-colors"
                  style={{
                    background: `linear-gradient(135deg, ${partType.color}20 0%, transparent 100%)`,
                    borderColor: partType.is_active ? partType.color : '#374151'
                  }}
                >
                  <div className="flex items-start justify-between mb-3">
                    <div
                      className="p-2 rounded-lg"
                      style={{ background: `${partType.color}30`, color: partType.color }}
                    >
                      {getIconComponent(partType.icon)}
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => {
                          setSelectedPartType(partType);
                          setIsEditModalOpen(true);
                        }}
                      >
                        <Edit className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(partType.id)}
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                  <h3 className="font-semibold text-lg mb-1">{partType.name_ar}</h3>
                  <p className="text-sm text-gray-400 mb-2">{partType.name_en}</p>
                  <Badge variant={partType.is_active ? 'success' : 'secondary'}>
                    {partType.is_active ? 'نشط' : 'معطل'}
                  </Badge>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Modal */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="إضافة نوع قطعة جديد"
        variant="modern"
        size="sm"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-text mb-2">الاسم بالعربية</label>
            <Input
              value={newPartType.name_ar}
              onChange={(e) => setNewPartType({ ...newPartType, name_ar: e.target.value })}
              placeholder="مثال: كرت شاشة"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-text mb-2">الاسم بالإنجليزية</label>
            <Input
              value={newPartType.name_en}
              onChange={(e) => setNewPartType({ ...newPartType, name_en: e.target.value })}
              placeholder="مثال: Graphics Card"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-text mb-2">الأيقونة</label>
            <select
              value={newPartType.icon}
              onChange={(e) => setNewPartType({ ...newPartType, icon: e.target.value })}
              className="w-full p-2 rounded bg-gray-800 border border-gray-700 text-white"
            >
              <option value="box">صندوق</option>
              <option value="monitor">شاشة</option>
              <option value="cpu">معالج</option>
              <option value="hard-drive">قرص صلب</option>
              <option value="zap">طاقة</option>
              <option value="thermometer">تبريد</option>
              <option value="package">قطعة</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-text mb-2">اللون</label>
            <Input
              type="color"
              value={newPartType.color}
              onChange={(e) => setNewPartType({ ...newPartType, color: e.target.value })}
              className="h-10"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-text mb-2">ترتيب العرض</label>
            <Input
              type="number"
              value={newPartType.sort_order}
              onChange={(e) => setNewPartType({ ...newPartType, sort_order: parseInt(e.target.value) || 0 })}
            />
          </div>
          <div className="flex justify-end gap-3 pt-4">
            <Button
              variant="secondary"
              onClick={() => setIsCreateModalOpen(false)}
            >
              إلغاء
            </Button>
            <Button variant="primary" onClick={handleCreate}>
              إضافة
            </Button>
          </div>
        </div>
      </Modal>

      {/* Edit Modal */}
      <Modal
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        title="تعديل نوع القطعة"
        variant="modern"
        size="sm"
      >
        {selectedPartType && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-text mb-2">الاسم بالعربية</label>
              <Input
                value={selectedPartType.name_ar || ''}
                onChange={(e) => setSelectedPartType({ ...selectedPartType, name_ar: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-text mb-2">الاسم بالإنجليزية</label>
              <Input
                value={selectedPartType.name_en || ''}
                onChange={(e) => setSelectedPartType({ ...selectedPartType, name_en: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-text mb-2">الأيقونة</label>
              <Select
                value={selectedPartType.icon || 'box'}
                onChange={(e) => setSelectedPartType({ ...selectedPartType, icon: e.target.value })}
                options={[
                  { value: 'box', label: 'صندوق' },
                  { value: 'monitor', label: 'شاشة' },
                  { value: 'cpu', label: 'معالج' },
                  { value: 'hard-drive', label: 'قرص صلب' },
                  { value: 'zap', label: 'طاقة' },
                  { value: 'thermometer', label: 'تبريد' },
                  { value: 'package', label: 'قطعة' },
                ]}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-text mb-2">اللون</label>
              <Input
                type="color"
                value={selectedPartType.color || '#3B82F6'}
                onChange={(e) => setSelectedPartType({ ...selectedPartType, color: e.target.value })}
                className="h-10"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-text mb-2">الحالة</label>
              <Select
                value={selectedPartType.is_active ? 'true' : 'false'}
                onChange={(e) => setSelectedPartType({ ...selectedPartType, is_active: e.target.value === 'true' })}
                options={[
                  { value: 'true', label: 'نشط' },
                  { value: 'false', label: 'معطل' },
                ]}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-text mb-2">ترتيب العرض</label>
              <Input
                type="number"
                value={selectedPartType.sort_order || 0}
                onChange={(e) => setSelectedPartType({ ...selectedPartType, sort_order: parseInt(e.target.value) || 0 })}
              />
            </div>
            <div className="flex justify-end gap-3 pt-4">
              <Button
                variant="secondary"
                onClick={() => setIsEditModalOpen(false)}
              >
                إلغاء
              </Button>
              <Button variant="primary" onClick={handleUpdate}>
                حفظ التغييرات
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}