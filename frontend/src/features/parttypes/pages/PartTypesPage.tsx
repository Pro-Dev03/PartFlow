import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { partTypesApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { Select } from '../../../design-system/components/select';
import { PageHeader } from '../../../design-system/components/page-header';
import { Modal } from '../../../design-system/components/modal';
import { ConfirmDialog } from '../../../design-system/components/confirm-dialog';
import { getButtonSize } from '../../../config/button-sizes';
import { useLayout } from '../../../contexts/LayoutContext';
import { toast } from 'sonner';
import { compressPartTypeImage, getPartTypeImage, setPartTypeImage } from '../../../services/localPartTypeImages';
import { 
  Plus, 
  Edit, 
  Trash2, 
  Monitor,
  Cpu,
  HardDrive,
  Zap,
  Box,
  Thermometer,
  Package,
  Power,
  PowerOff,
  ImagePlus,
  Upload,
  X
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

const partTypeIconOptions = [
  { value: 'box', icon: Box },
  { value: 'monitor', icon: Monitor },
  { value: 'cpu', icon: Cpu },
  { value: 'hard-drive', icon: HardDrive },
  { value: 'zap', icon: Zap },
  { value: 'thermometer', icon: Thermometer },
  { value: 'package', icon: Package },
];

const partTypeColorOptions = ['#14b8a6', '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899', '#6B7280'];

function PartTypeImageField({ value, onChange }: { value?: string; onChange: (value?: string) => void }) {
  const handleChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file || !file.type.startsWith('image/')) return;
    try {
      onChange(await compressPartTypeImage(file));
    } catch {
      toast.error('تعذر تجهيز الصورة');
    }
  };

  return (
    <div>
      <label className="block text-sm font-medium text-text mb-2">صورة نوع القطعة</label>
      <div className="flex items-center gap-3 rounded-xl border border-dashed border-[var(--border-default)] bg-[var(--bg-surface-elevated)] p-3">
        <label className="group relative flex h-[88px] w-[104px] shrink-0 cursor-pointer flex-col items-center justify-center overflow-hidden rounded-xl border border-dashed border-[var(--color-primary-30)] bg-[var(--color-primary-08)] transition-all hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-10)]">
          {value ? <img src={value} alt="معاينة نوع القطعة" className="h-full w-full object-cover" /> : <><span className="mb-1 flex h-9 w-9 items-center justify-center rounded-full bg-[var(--color-primary-15)] text-[var(--color-primary)]"><ImagePlus className="h-5 w-5" /></span><span className="text-[11px] font-semibold text-[var(--text-primary)]">رفع صورة</span></>}
          {value && <span className="absolute inset-0 flex items-center justify-center bg-black/50 text-xs font-semibold text-white opacity-0 transition-opacity group-hover:opacity-100"><Upload className="me-1.5 h-3.5 w-3.5" />استبدال</span>}
          <input type="file" accept="image/*" onChange={handleChange} hidden />
        </label>
        <div className="flex min-w-0 flex-1 flex-col gap-2">
          {value && <button type="button" onClick={() => onChange(undefined)} className="inline-flex w-fit items-center gap-1.5 text-xs text-[var(--color-danger)] hover:underline"><X className="w-3.5 h-3.5" />إزالة الصورة</button>}
          <span className="truncate text-[10px] text-text-secondary">JPG أو PNG أو WEBP، حتى 5MB</span>
        </div>
      </div>
    </div>
  );
}

export function PartTypesPage() {
  const queryClient = useQueryClient();
  const { setFullWidth } = useLayout();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [selectedPartType, setSelectedPartType] = useState<any>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [partTypeToDelete, setPartTypeToDelete] = useState<string | null>(null);
  const [newPartType, setNewPartType] = useState({
    name_ar: '',
    name_en: '',
    icon: 'box',
    color: '#14b8a6',
    sort_order: 0
  });
  const [newPartTypeImage, setNewPartTypeImage] = useState<string>();

  useEffect(() => {
    setFullWidth(true);
    return () => setFullWidth(false);
  }, [setFullWidth]);

  const { data: partTypesData, isLoading } = useQuery({
    queryKey: ['part-types'],
    queryFn: () => partTypesApi.list(),
  });

  const partTypes = (partTypesData?.data as any[]) || [];

  const createMutation = useMutation({
    mutationFn: (data: any) => partTypesApi.create(data),
    onSuccess: (response) => {
      const createdPartType = (response?.data as any)?.part_type ?? response?.data;
      if (createdPartType?.id && newPartTypeImage) {
        setPartTypeImage(createdPartType.id, newPartTypeImage);
      }
      queryClient.invalidateQueries({ queryKey: ['part-types'] });
      setIsCreateModalOpen(false);
      setNewPartType({ name_ar: '', name_en: '', icon: 'box', color: '#14b8a6', sort_order: 0 });
      setNewPartTypeImage(undefined);
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

  const toggleActiveMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) => partTypesApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['part-types'] });
      toast.success('تم تحديث حالة نوع القطعة');
    },
    onError: () => toast.error('فشل تحديث حالة نوع القطعة'),
  });

  const handleCreate = () => {
    const normalizedNameAr = newPartType.name_ar.trim();
    if (!normalizedNameAr) {
      toast.error('يرجى ملء جميع الحقول المطلوبة');
      return;
    }

    const payload = {
      ...newPartType,
      name_ar: normalizedNameAr,
      name_en: (newPartType.name_en || normalizedNameAr).trim() || normalizedNameAr,
    };

    createMutation.mutate(payload);
  };

  const handleUpdate = () => {
    if (!selectedPartType?.name_ar) {
      toast.error('يرجى ملء جميع الحقول المطلوبة');
      return;
    }
    const normalizedNameAr = selectedPartType.name_ar.trim();
    const normalizedNameEn = (selectedPartType.name_en || normalizedNameAr).trim() || normalizedNameAr;
    const updatedPartType = {
      ...selectedPartType,
      name_ar: normalizedNameAr,
      name_en: normalizedNameEn,
    };
    setSelectedPartType(updatedPartType);
    setPartTypeImage(updatedPartType.id, updatedPartType.image_url || null);
    updateMutation.mutate({
      id: updatedPartType.id,
      data: updatedPartType
    });
  };

  const handleDelete = (id: string) => {
    setPartTypeToDelete(id);
    setDeleteDialogOpen(true);
  };

  const handleToggleActive = (partType: any) => {
    toggleActiveMutation.mutate({
      id: partType.id,
      data: {
        name_ar: partType.name_ar,
        name_en: partType.name_en,
        icon: partType.icon || 'box',
        color: partType.color || '#14b8a6',
        is_active: !partType.is_active,
        sort_order: partType.sort_order || 0,
      },
    });
  };

  const handleConfirmDelete = () => {
    if (partTypeToDelete) {
      deleteMutation.mutate(partTypeToDelete);
      setPartTypeImage(partTypeToDelete, null);
      setDeleteDialogOpen(false);
      setPartTypeToDelete(null);
    }
  };

  const getIconComponent = (iconName: string) => {
    const IconComponent = iconMap[iconName] || Box;
    return <IconComponent style={{ width: '20px', height: '20px' }} />;
  };

  return (
    <div>
      <PageHeader
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
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '20px' }}>
              {partTypes.map((partType: any) => (
                <div
                  key={partType.id}
                  className="transition-all duration-300"
                  style={{
                    position: 'relative',
                    opacity: partType.is_active ? 1 : 0.6,
                    transform: 'translateY(0)'
                  }}
                >
                  <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: '4px', background: `linear-gradient(90deg, ${partType.color || '#14b8a6'}, ${partType.color || '#14b8a6'}dd)`, borderRadius: '12px 12px 0 0', opacity: partType.is_active ? 1 : 0.3 }} />
                  <Card style={{ background: 'var(--card-bg)', border: `1px solid ${partType.is_active ? 'var(--card-border)' : 'rgba(107, 114, 128, 0.3)'}`, borderRadius: '12px', overflow: 'hidden', paddingTop: '8px' }}>
                    <CardContent style={{ padding: '20px' }}>
                      <div className="flex items-start justify-between mb-4">
                        <div className="flex items-center gap-4">
                          <div style={{ width: '72px', height: '72px', flex: '0 0 72px', overflow: 'hidden', borderRadius: '14px', display: 'flex', alignItems: 'center', justifyContent: 'center', background: `linear-gradient(135deg, ${partType.color || '#14b8a6'}15, ${partType.color || '#14b8a6'}08)`, border: `2px solid ${partType.color || '#14b8a6'}25`, boxShadow: `0 4px 12px ${partType.color || '#14b8a6'}15` }}>
                            {partType.image_url || getPartTypeImage(partType.id) ? (
                              <img src={partType.image_url || getPartTypeImage(partType.id)} alt={partType.name_ar} style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
                            ) : <span style={{ color: partType.color || '#14b8a6' }}>{getIconComponent(partType.icon)}</span>}
                          </div>
                          <div>
                            <h3 className="text-base font-bold text-text-primary mb-1">{partType.name_ar}</h3>
                            <p className="text-xs text-text-secondary">{partType.name_en}</p>
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center justify-between gap-2">
                        <button type="button" onClick={() => handleToggleActive(partType)} className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors" style={{ background: partType.is_active ? 'rgba(16, 185, 129, 0.1)' : 'rgba(107, 114, 128, 0.1)', border: partType.is_active ? '1px solid rgba(16, 185, 129, 0.2)' : '1px solid rgba(107, 114, 128, 0.2)', color: partType.is_active ? 'var(--color-success)' : 'var(--text-secondary)' }} title={partType.is_active ? 'تعطيل نوع القطعة' : 'تفعيل نوع القطعة'}>
                          {partType.is_active ? <PowerOff className="w-3.5 h-3.5" /> : <Power className="w-3.5 h-3.5" />}
                          {partType.is_active ? 'نشط' : 'غير نشط'}
                        </button>
                        <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => {
                          setSelectedPartType({ ...partType, image_url: getPartTypeImage(partType.id) });
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
                    </CardContent>
                  </Card>
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
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">اسم نوع القطعة *</label>
            <Input
              value={newPartType.name_ar}
              onChange={(e) => setNewPartType({ ...newPartType, name_ar: e.target.value, name_en: e.target.value })}
              placeholder="مثال: كرت شاشة"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">الأيقونة</label>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(8, 1fr)', gap: '8px' }}>
              {partTypeIconOptions.map(({ value, icon: Icon }) => (
                <button
                  key={value}
                  type="button"
                  onClick={() => setNewPartType({ ...newPartType, icon: value })}
                  style={{
                    padding: '8px',
                    borderRadius: '8px',
                    border: newPartType.icon === value ? '2px solid var(--color-primary)' : '1px solid var(--border-default)',
                    background: newPartType.icon === value ? 'rgba(99, 102, 241, 0.1)' : 'var(--bg-surface)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    cursor: 'pointer'
                  }}
                  aria-label={value}
                >
                  <Icon className="w-5 h-5" style={{ color: 'var(--text-primary)' }} />
                </button>
              ))}
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">اللون</label>
            <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
              {partTypeColorOptions.map((color) => (
                <button
                  key={color}
                  type="button"
                  onClick={() => setNewPartType({ ...newPartType, color })}
                  style={{
                    width: '32px',
                    height: '32px',
                    borderRadius: '50%',
                    background: color,
                    border: newPartType.color === color ? '3px solid var(--color-primary)' : '2px solid var(--border-default)',
                    cursor: 'pointer'
                  }}
                  aria-label={color}
                />
              ))}
            </div>
          </div>
          <PartTypeImageField value={newPartTypeImage} onChange={setNewPartTypeImage} />
          <div className="flex justify-end gap-3 pt-4">
            <Button
              variant="secondary"
              onClick={() => setIsCreateModalOpen(false)}
            >
              إلغاء
            </Button>
            <Button variant="primary" onClick={handleCreate}>
              {createMutation.isPending ? 'جاري الإضافة...' : 'إضافة'}
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
              <label className="block text-sm font-medium text-text mb-2">اسم نوع القطعة *</label>
              <Input
                value={selectedPartType.name_ar || ''}
                onChange={(e) => setSelectedPartType({ ...selectedPartType, name_ar: e.target.value, name_en: e.target.value })}
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
            <PartTypeImageField
              value={selectedPartType.image_url}
              onChange={(image_url) => setSelectedPartType({ ...selectedPartType, image_url })}
            />
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

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        isOpen={deleteDialogOpen}
        onClose={() => {
          setDeleteDialogOpen(false);
          setPartTypeToDelete(null);
        }}
        onConfirm={handleConfirmDelete}
        title="حذف نوع القطعة"
        message="هل أنت متأكد من حذف هذا النوع؟ هذا الإجراء لا يمكن التراجع عنه."
        confirmText="حذف النوع"
        cancelText="إلغاء"
        variant="danger"
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}