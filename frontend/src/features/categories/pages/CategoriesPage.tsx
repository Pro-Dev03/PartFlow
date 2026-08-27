import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { categoriesApi } from '../../../services/api/endpoints';
import { Card, CardContent } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { Modal } from '../../../components/ui/modal';
import { ConfirmDialog } from '../../../components/ui/confirm-dialog';
import { getButtonSize } from '../../../config/button-sizes';
import { toast } from 'sonner';
import { useLayout } from '../../../contexts/LayoutContext';

import {
  Plus,
  Edit,
  Trash2,
  Package,
  Tag,
  Monitor,
  Cpu,
  HardDrive,
  Zap,
  Box,
  Thermometer,
  Keyboard,
  Mouse,
  Headphones,
  Camera,
  Printer,
  Wifi,
  Shield,
  Wrench,
  Smartphone,
  Laptop,
  Speaker,
  Cable,
  Power,
  PowerOff
} from 'lucide-react';

const iconMap: Record<string, any> = {
  package: Package,
  tag: Tag,
  monitor: Monitor,
  cpu: Cpu,
  'hard-drive': HardDrive,
  zap: Zap,
  box: Box,
  thermometer: Thermometer,
  keyboard: Keyboard,
  mouse: Mouse,
  headphones: Headphones,
  camera: Camera,
  printer: Printer,
  wifi: Wifi,
  shield: Shield,
  wrench: Wrench,
  smartphone: Smartphone,
  laptop: Laptop,
  speaker: Speaker,
  cable: Cable,
};

const availableIcons = [
  { value: 'smartphone', label: 'هاتف', icon: Smartphone },
  { value: 'laptop', label: 'لابتوب', icon: Laptop },
  { value: 'monitor', label: 'شاشة', icon: Monitor },
  { value: 'cpu', label: 'معالج', icon: Cpu },
  { value: 'hard-drive', label: 'قرص صلب', icon: HardDrive },
  { value: 'camera', label: 'كاميرا', icon: Camera },
  { value: 'printer', label: 'طابعة', icon: Printer },
  { value: 'wifi', label: 'واي فاي', icon: Wifi },
  { value: 'headphones', label: 'سماعات', icon: Headphones },
  { value: 'speaker', label: 'مكبر صوت', icon: Speaker },
  { value: 'cable', label: 'كابل', icon: Cable },
  { value: 'keyboard', label: 'لوحة مفاتيح', icon: Keyboard },
  { value: 'mouse', label: 'فأرة', icon: Mouse },
  { value: 'package', label: 'صندوق', icon: Package },
  { value: 'tag', label: 'علامة', icon: Tag },
  { value: 'zap', label: 'طاقة', icon: Zap },
  { value: 'box', label: 'علبة', icon: Box },
  { value: 'thermometer', label: 'تبريد', icon: Thermometer },
  { value: 'shield', label: 'حماية', icon: Shield },
  { value: 'wrench', label: 'أدوات', icon: Wrench },
];

const colorOptions = [
  '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', 
  '#EC4899', '#06B6D4', '#6B7280', '#14B8A6', '#F97316'
];

export function CategoriesPage() {
  const queryClient = useQueryClient();
  const { setFullWidth } = useLayout();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<any>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [categoryToDelete, setCategoryToDelete] = useState<{ id: string; name: string } | null>(null);
  const [newCategory, setNewCategory] = useState({
    name: '',
    description: '',
    icon: 'smartphone',
    color: '#3B82F6',
    is_active: true
  });

  useEffect(() => {
    setFullWidth(true);
    
    const style = document.createElement('style');
    style.textContent = `
      @keyframes pulse {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.5; }
      }
    `;
    document.head.appendChild(style);
    
    return () => {
      setFullWidth(false);
      document.head.removeChild(style);
    };
  }, [setFullWidth]);

  const getCategoryIcon = (category: any) => {
    if (typeof category.icon === 'string' && category.icon) {
      return iconMap[category.icon] || Smartphone;
    }
    return Smartphone;
  };

  const getCategoryColor = (category: any) => {
    if (typeof category.color === 'string' && category.color) {
      return category.color;
    }
    return '#3B82F6';
  };

  const { data: categoriesData, isLoading, error } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });

  if (error) {
    console.error('Error loading categories:', error);
  }

  const categories = (categoriesData?.data as any[]) || [];

  const createMutation = useMutation({
    mutationFn: (data: any) => {
      const apiData = {
        ...data,
        icon: data.icon ? data.icon : null,
        color: data.color ? data.color : null
      };
      return categoriesApi.create(apiData);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      setIsCreateModalOpen(false);
      setNewCategory({ name: '', description: '', icon: 'smartphone', color: '#3B82F6', is_active: true });
      toast.success('تم إضافة التصنيف بنجاح');
    },
    onError: (error: any) => {
      console.error('Create category error:', error);
      toast.error(`فشل إضافة التصنيف: ${error.message || error.arabicMessage || 'خطأ غير معروف'}`);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) => {
      const apiData = {
        ...data,
        icon: data.icon ? data.icon : null,
        color: data.color ? data.color : null
      };
      return categoriesApi.update(id, apiData);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      setIsEditModalOpen(false);
      setSelectedCategory(null);
      toast.success('تم تحديث التصنيف بنجاح');
    },
    onError: (error: any) => {
      console.error('Update category error:', error);
      toast.error(`فشل تحديث التصنيف: ${error.message || error.arabicMessage || 'خطأ غير معروف'}`);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => categoriesApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      queryClient.invalidateQueries({ queryKey: ['products'] }); // Invalidate products since their category_id might be NULL now
      toast.success('تم حذف التصنيف بنجاح. المنتجات المرتبطة به أصبحت بدون تصنيف.');
    },
    onError: (error: any) => {
      // Always refresh the list on delete attempt to handle the case where it was already deleted
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      queryClient.invalidateQueries({ queryKey: ['products'] });

      // If category is already deleted, show success message (this is expected behavior for double-clicks)
      if (error.message === 'category not found' || error.arabicMessage === 'التصنيف غير موجود' || error.status === 400) {
        toast.success('تم حذف التصنيف بنجاح.');
      } else {
        console.error('Delete category error:', error);
        toast.error(`فشل حذف التصنيف: ${error.message || error.arabicMessage || 'خطأ غير معروف'}`);
      }
    },
    retry: 0, // Prevent double-click by making the mutation non-retryable
  });

  const toggleActiveMutation = useMutation({
    mutationFn: ({ id, is_active }: { id: string; is_active: boolean }) => {
      return categoriesApi.update(id, { is_active });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      toast.success('تم تحديث حالة التصنيف بنجاح');
    },
    onError: (error: any) => {
      console.error('Toggle category active error:', error);
      toast.error(`فشل تحديث حالة التصنيف: ${error.message || error.arabicMessage || 'خطأ غير معروف'}`);
    },
  });

  const handleCreate = () => {
    if (!newCategory.name) {
      toast.error('يرجى إدخال اسم التصنيف');
      return;
    }
    createMutation.mutate(newCategory);
  };

  const handleUpdate = () => {
    if (!selectedCategory?.name) {
      toast.error('يرجى إدخال اسم التصنيف');
      return;
    }
    updateMutation.mutate({
      id: selectedCategory.id,
      data: selectedCategory
    });
  };

  const handleDelete = (id: string, categoryName: string) => {
    setCategoryToDelete({ id, name: categoryName });
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = () => {
    if (categoryToDelete) {
      deleteMutation.mutate(categoryToDelete.id);
      setDeleteDialogOpen(false);
      setCategoryToDelete(null);
    }
  };

  const handleToggleActive = (id: string, currentStatus: boolean) => {
    toggleActiveMutation.mutate({ id, is_active: !currentStatus });
  };

  const handleEdit = (category: any) => {
    setSelectedCategory({
      ...category,
      icon: category.icon || 'smartphone',
      color: category.color || '#3B82F6'
    });
    setIsEditModalOpen(true);
  };

  if (isLoading) {
    return (
      <div>
        <PageHeader
          title="التصنيفات"
          description="إدارة تصنيفات المنتجات"
        />
        <div className="flex items-center justify-center h-64">
          <div className="text-text-secondary">جاري التحميل...</div>
        </div>
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        title="التصنيفات"
        description="إدارة تصنيفات المنتجات"
      />

      <div style={{ 
        marginBottom: '16px',
        display: 'flex',
        justifyContent: 'flex-end',
        alignItems: 'center',
        gap: '12px'
      }}>
        <Button
          variant="primary"
          size={getButtonSize('categories', 'headerActions')}
          onClick={() => setIsCreateModalOpen(true)}
          className="gap-2"
        >
          <Plus className="w-4 h-4" />
          <span>إضافة تصنيف</span>
        </Button>
      </div>

      <div style={{ 
        display: 'grid', 
        gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', 
        gap: '20px'
      }}>
        {categories.map((category: any) => {
          const CategoryIcon = getCategoryIcon(category);
          const categoryColor = getCategoryColor(category);
          const opacity = category.is_active ? 1 : 0.6;
          
          return (
            <div
              key={category.id}
              style={{
                position: 'relative',
                opacity: opacity,
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
              }}
              onMouseEnter={(e) => {
                if (category.is_active) {
                  e.currentTarget.style.transform = 'translateY(-4px)';
                  e.currentTarget.style.boxShadow = `0 12px 40px ${categoryColor}25`;
                }
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = 'translateY(0)';
                e.currentTarget.style.boxShadow = 'none';
              }}
            >
              <div style={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                height: '4px',
                background: `linear-gradient(90deg, ${categoryColor}, ${categoryColor}dd)`,
                borderRadius: '12px 12px 0 0',
                opacity: category.is_active ? 1 : 0.3
              }} />
              
              <Card
                style={{
                  background: 'var(--card-bg)',
                  border: `1px solid ${category.is_active ? 'var(--card-border)' : 'rgba(107, 114, 128, 0.3)'}`,
                  borderRadius: '12px',
                  overflow: 'hidden',
                  paddingTop: '8px'
                }}
              >
                <CardContent style={{ padding: '20px' }}>
                  <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: '16px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
                      <div
                        style={{
                          width: '56px',
                          height: '56px',
                          borderRadius: '14px',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          background: `linear-gradient(135deg, ${categoryColor}15, ${categoryColor}08)`,
                          border: `2px solid ${categoryColor}25`,
                          boxShadow: `0 4px 12px ${categoryColor}15`,
                          transition: 'all 0.3s ease'
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.transform = 'scale(1.1)';
                          e.currentTarget.style.boxShadow = `0 8px 20px ${categoryColor}25`;
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.transform = 'scale(1)';
                          e.currentTarget.style.boxShadow = `0 4px 12px ${categoryColor}15`;
                        }}
                      >
                        <CategoryIcon style={{ 
                          width: '28px', 
                          height: '28px', 
                          color: categoryColor,
                          strokeWidth: 2
                        }} />
                      </div>
                      
                      <div>
                        <h3 style={{
                          fontSize: '16px',
                          fontWeight: '700',
                          color: 'var(--text-primary)',
                          marginBottom: '4px',
                          letterSpacing: '-0.3px'
                        }}>
                          {category.name}
                        </h3>
                        {category.description && (
                          <p style={{
                            fontSize: '12px',
                            color: 'var(--text-secondary)',
                            lineHeight: '1.4',
                            maxWidth: '140px'
                          }}>
                            {category.description}
                          </p>
                        )}
                      </div>
                    </div>
                    
                    <div style={{ display: 'flex', gap: '6px' }}>
                      <button
                        onClick={() => handleEdit(category)}
                        style={{
                          width: '36px',
                          height: '36px',
                          borderRadius: '10px',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          background: 'var(--bg-surface-elevated)',
                          border: '1px solid var(--border-subtle)',
                          color: 'var(--text-secondary)',
                          cursor: 'pointer',
                          transition: 'all 0.2s ease'
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.background = `${categoryColor}15`;
                          e.currentTarget.style.borderColor = categoryColor;
                          e.currentTarget.style.color = categoryColor;
                          e.currentTarget.style.transform = 'scale(1.05)';
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                          e.currentTarget.style.borderColor = 'var(--border-subtle)';
                          e.currentTarget.style.color = 'var(--text-secondary)';
                          e.currentTarget.style.transform = 'scale(1)';
                        }}
                      >
                        <Edit style={{ width: '16px', height: '16px' }} />
                      </button>
                      <button
                        onClick={() => handleToggleActive(category.id, category.is_active)}
                        style={{
                          width: '36px',
                          height: '36px',
                          borderRadius: '10px',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          background: 'var(--bg-surface-elevated)',
                          border: '1px solid var(--border-subtle)',
                          color: 'var(--text-secondary)',
                          cursor: 'pointer',
                          transition: 'all 0.2s ease'
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.background = category.is_active ? 'rgba(107, 114, 128, 0.1)' : 'rgba(16, 185, 129, 0.1)';
                          e.currentTarget.style.borderColor = category.is_active ? 'var(--text-secondary)' : 'var(--color-success)';
                          e.currentTarget.style.color = category.is_active ? 'var(--text-secondary)' : 'var(--color-success)';
                          e.currentTarget.style.transform = 'scale(1.05)';
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                          e.currentTarget.style.borderColor = 'var(--border-subtle)';
                          e.currentTarget.style.color = 'var(--text-secondary)';
                          e.currentTarget.style.transform = 'scale(1)';
                        }}
                        title={category.is_active ? 'تعطيل التصنيف' : 'تفعيل التصنيف'}
                      >
                        {category.is_active ? <PowerOff style={{ width: '16px', height: '16px' }} /> : <Power style={{ width: '16px', height: '16px' }} />}
                      </button>
                      <button
                        onClick={() => handleDelete(category.id, category.name)}
                        style={{
                          width: '36px',
                          height: '36px',
                          borderRadius: '10px',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          background: 'var(--bg-surface-elevated)',
                          border: '1px solid var(--border-subtle)',
                          color: 'var(--text-secondary)',
                          cursor: 'pointer',
                          transition: 'all 0.2s ease'
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.background = 'rgba(239, 68, 68, 0.1)';
                          e.currentTarget.style.borderColor = 'var(--color-danger)';
                          e.currentTarget.style.color = 'var(--color-danger)';
                          e.currentTarget.style.transform = 'scale(1.05)';
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                          e.currentTarget.style.borderColor = 'var(--border-subtle)';
                          e.currentTarget.style.color = 'var(--text-secondary)';
                          e.currentTarget.style.transform = 'scale(1)';
                        }}
                      >
                        <Trash2 style={{ width: '16px', height: '16px' }} />
                      </button>
                    </div>
                  </div>
                  
                  <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    <button
                      onClick={() => handleToggleActive(category.id, category.is_active)}
                      style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '6px',
                        padding: '6px 12px',
                        borderRadius: '8px',
                        fontSize: '12px',
                        fontWeight: '600',
                        letterSpacing: '0.2px',
                        background: category.is_active
                          ? 'rgba(16, 185, 129, 0.1)'
                          : 'rgba(107, 114, 128, 0.1)',
                        border: category.is_active
                          ? '1px solid rgba(16, 185, 129, 0.2)'
                          : '1px solid rgba(107, 114, 128, 0.2)',
                        color: category.is_active
                          ? 'var(--color-success)'
                          : 'var(--text-secondary)',
                        cursor: 'pointer',
                        transition: 'all 0.2s ease'
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.background = category.is_active ? 'rgba(107, 114, 128, 0.15)' : 'rgba(16, 185, 129, 0.15)';
                        e.currentTarget.style.borderColor = category.is_active ? 'var(--text-secondary)' : 'var(--color-success)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.background = category.is_active ? 'rgba(16, 185, 129, 0.1)' : 'rgba(107, 114, 128, 0.1)';
                        e.currentTarget.style.borderColor = category.is_active ? 'rgba(16, 185, 129, 0.2)' : 'rgba(107, 114, 128, 0.2)';
                      }}
                      title={category.is_active ? 'تعطيل التصنيف' : 'تفعيل التصنيف'}
                    >
                      <div style={{
                        width: '8px',
                        height: '8px',
                        borderRadius: '50%',
                        background: category.is_active
                          ? 'var(--color-success)'
                          : 'var(--text-secondary)',
                        animation: category.is_active ? 'pulse 2s infinite' : 'none'
                      }} />
                      {category.is_active ? 'نشط' : 'غير نشط'}
                    </button>
                  </div>
                </CardContent>
              </Card>
            </div>
          );
        })}
      </div>

      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="إضافة تصنيف جديد"
        variant="modern"
        size="sm"
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">
              اسم التصنيف *
            </label>
            <Input
              value={newCategory.name}
              onChange={(e) => setNewCategory({ ...newCategory, name: e.target.value })}
              placeholder="مثال: كروت شاشة"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">
              الوصف
            </label>
            <Input
              value={newCategory.description}
              onChange={(e) => setNewCategory({ ...newCategory, description: e.target.value })}
              placeholder="وصف قصير للتصنيف"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">
              الأيقونة
            </label>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(8, 1fr)', gap: '8px' }}>
              {availableIcons.map((icon) => (
                <button
                  key={icon.value}
                  type="button"
                  onClick={() => setNewCategory({ ...newCategory, icon: icon.value })}
                  style={{
                    padding: '8px',
                    borderRadius: '8px',
                    border: newCategory.icon === icon.value ? '2px solid var(--color-primary)' : '1px solid var(--border-default)',
                    background: newCategory.icon === icon.value ? 'rgba(99, 102, 241, 0.1)' : 'var(--bg-surface)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    cursor: 'pointer'
                  }}
                >
                  <icon.icon className="w-5 h-5" style={{ color: 'var(--text-primary)' }} />
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">
              اللون
            </label>
            <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
              {colorOptions.map((color) => (
                <button
                  key={color}
                  type="button"
                  onClick={() => setNewCategory({ ...newCategory, color })}
                  style={{
                    width: '32px',
                    height: '32px',
                    borderRadius: '50%',
                    background: color,
                    border: newCategory.color === color ? '3px solid var(--color-primary)' : '2px solid var(--border-default)',
                    cursor: 'pointer'
                  }}
                />
              ))}
            </div>
          </div>

          <div className="flex gap-3 justify-end mt-4">
            <Button
              variant="secondary"
              onClick={() => setIsCreateModalOpen(false)}
            >
              إلغاء
            </Button>
            <Button
              variant="primary"
              onClick={handleCreate}
              disabled={createMutation.isPending}
            >
              {createMutation.isPending ? 'جاري الإضافة...' : 'إضافة'}
            </Button>
          </div>
        </div>
      </Modal>

      <Modal
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        title="تعديل التصنيف"
        variant="modern"
        size="sm"
      >
        {selectedCategory && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            <div>
              <label className="block text-sm font-medium text-text-primary mb-2">
                اسم التصنيف *
              </label>
              <Input
                value={selectedCategory.name || ''}
                onChange={(e) => setSelectedCategory({ ...selectedCategory, name: e.target.value })}
                placeholder="مثال: كروت شاشة"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-text-primary mb-2">
                الوصف
              </label>
              <Input
                value={selectedCategory.description || ''}
                onChange={(e) => setSelectedCategory({ ...selectedCategory, description: e.target.value })}
                placeholder="وصف قصير للتصنيف"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-text-primary mb-2">
                الأيقونة
              </label>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(8, 1fr)', gap: '8px' }}>
                {availableIcons.map((icon) => (
                  <button
                    key={icon.value}
                    type="button"
                    onClick={() => setSelectedCategory({ ...selectedCategory, icon: icon.value })}
                    style={{
                      padding: '8px',
                      borderRadius: '8px',
                      border: (selectedCategory.icon || 'package') === icon.value ? '2px solid var(--color-primary)' : '1px solid var(--border-default)',
                      background: (selectedCategory.icon || 'package') === icon.value ? 'rgba(99, 102, 241, 0.1)' : 'var(--bg-surface)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      cursor: 'pointer'
                    }}
                  >
                    <icon.icon className="w-5 h-5" style={{ color: 'var(--text-primary)' }} />
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-text-primary mb-2">
                اللون
              </label>
              <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
                {colorOptions.map((color) => (
                  <button
                    key={color}
                    type="button"
                    onClick={() => setSelectedCategory({ ...selectedCategory, color })}
                    style={{
                      width: '32px',
                      height: '32px',
                      borderRadius: '50%',
                      background: color,
                      border: (selectedCategory.color || '#3B82F6') === color ? '3px solid var(--color-primary)' : '2px solid var(--border-default)',
                      cursor: 'pointer'
                    }}
                  />
                ))}
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-text-primary mb-2">
                الحالة
              </label>
              <div className="flex gap-3">
                <Button
                  variant={selectedCategory.is_active ? 'primary' : 'secondary'}
                  onClick={() => setSelectedCategory({ ...selectedCategory, is_active: true })}
                  className="flex-1"
                >
                  نشط
                </Button>
                <Button
                  variant={!selectedCategory.is_active ? 'primary' : 'secondary'}
                  onClick={() => setSelectedCategory({ ...selectedCategory, is_active: false })}
                  className="flex-1"
                >
                  غير نشط
                </Button>
              </div>
            </div>

            <div className="flex gap-3 justify-end mt-4">
              <Button
                variant="secondary"
                onClick={() => setIsEditModalOpen(false)}
              >
                إلغاء
              </Button>
              <Button
                variant="primary"
                onClick={handleUpdate}
                disabled={updateMutation.isPending}
              >
                {updateMutation.isPending ? 'جاري التحديث...' : 'تحديث'}
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
          setCategoryToDelete(null);
        }}
        onConfirm={handleConfirmDelete}
        title="حذف التصنيف"
        message={categoryToDelete 
          ? `هل أنت متأكد من حذف تصنيف "${categoryToDelete.name}"؟\n\nملاحظة: سيتم تحديث المنتجات المرتبطة بهذا التصنيف لتصبح بدون تصنيف.`
          : 'هل أنت متأكد من حذف هذا التصنيف؟'
        }
        confirmText="حذف التصنيف"
        cancelText="إلغاء"
        variant="danger"
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}