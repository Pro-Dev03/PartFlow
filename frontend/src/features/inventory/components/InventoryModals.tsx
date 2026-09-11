import { Modal } from '../../../components/ui/modal';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { Product } from '../types/inventory.types';
import { Package, Plus, Sparkles, Tag, DollarSign } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { categoriesApi, settingsApi } from '../../../services/api/endpoints';
import { calculateSuggestedSellingPrice, DEFAULT_PROFIT_MARGIN } from '../../../utils/pricing';
import { compressProductImage } from '../../../services/localProductImages';

interface InventoryModalsProps {
  isViewModalOpen: boolean;
  setIsViewModalOpen: (open: boolean) => void;
  isEditModalOpen: boolean;
  setIsEditModalOpen: (open: boolean) => void;
  isCreatingProduct: boolean;
  selectedProduct: Product | null;
  setSelectedProduct: (product: Product | null) => void;
  onSaveProduct: (productData: Product) => void;
}

function ProductImageField({
  value,
  onChange,
}: {
  value?: string;
  onChange: (value: string | undefined) => void;
}) {
  const handleChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    if (!file.type.startsWith('image/')) return;
    try {
      onChange(await compressProductImage(file));
    } catch {
      onChange(undefined);
    }
  };

  return (
    <div style={{ marginTop: '16px' }}>
      <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '8px' }}>
        صورة المنتج في نقطة البيع
      </label>
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
        <div style={{ width: '72px', height: '72px', borderRadius: '10px', overflow: 'hidden', background: 'var(--bg-surface-elevated)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          {value ? <img src={value} alt="معاينة المنتج" style={{ width: '100%', height: '100%', objectFit: 'cover' }} /> : <Package className="w-6 h-6" />}
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <label style={{ cursor: 'pointer', padding: '8px 12px', borderRadius: '8px', background: 'var(--primary)', color: 'var(--text-on-primary)', fontSize: '12px', fontWeight: 600 }}>
            رفع صورة
            <input type="file" accept="image/*" onChange={handleChange} hidden />
          </label>
          {value && <button type="button" onClick={() => onChange(undefined)} style={{ padding: '8px 12px', borderRadius: '8px', border: '1px solid var(--border-default)', background: 'transparent', color: 'var(--text-secondary)', fontSize: '12px' }}>حذف</button>}
        </div>
      </div>
    </div>
  );
}

export function InventoryModals({
  isViewModalOpen,
  setIsViewModalOpen,
  isEditModalOpen,
  setIsEditModalOpen,
  isCreatingProduct,
  selectedProduct,
  setSelectedProduct,
  onSaveProduct,
}: InventoryModalsProps) {
  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });
  const { data: marginSetting } = useQuery({
    queryKey: ['settings', 'default_profit_margin'],
    queryFn: () => settingsApi.getSetting('default_profit_margin'),
    retry: false,
  });
  const { data: taxSetting } = useQuery({
    queryKey: ['settings', 'tax_rate'],
    queryFn: () => settingsApi.getSetting('tax_rate'),
    retry: false,
  });

  const categories = (categoriesData?.data as unknown) as any[] || [];
  const taxRateValue = Number(taxSetting?.data?.value);
  const taxRate = Number.isFinite(taxRateValue) && taxRateValue >= 0 ? taxRateValue : 0;
  const configuredMargin = Number(marginSetting?.data?.value);
  const profitMargin = Number.isFinite(configuredMargin) && configuredMargin >= 0 && configuredMargin < 100
    ? configuredMargin
    : DEFAULT_PROFIT_MARGIN;
  return (
    <>
      {/* View Product Modal */}
      <Modal
        isOpen={isViewModalOpen}
        onClose={() => setIsViewModalOpen(false)}
        title="تفاصيل المنتج"
        variant="modern"
        size="lg"
        autoFocus={false}
        enableEnterNavigation={false}
      >
        {selectedProduct && (
          <div className="space-y-md">
            <div style={{ 
              display: 'grid', 
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', 
              gap: '16px' 
            }}>
              <div>
                <label className="text-small font-medium text-text mb-sm block">الاسم</label>
                <Input value={selectedProduct.name || ''} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">SKU</label>
                <Input value={selectedProduct.sku || ''} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">سعر التكلفة</label>
                <Input value={`₪${selectedProduct.costPrice || 0}`} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">سعر البيع قبل الضريبة</label>
                <Input value={`₪${selectedProduct.sellingPrice || 0}`} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">الضريبة</label>
                <Input
                  value={taxRate > 0 ? `${taxRate}% - تضاف عند البيع` : '0% - بدون ضريبة'}
                  disabled
                />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">السعر النهائي للعميل</label>
                <Input
                  value={`₪${((Number(selectedProduct.sellingPrice) || 0) * (1 + taxRate / 100)).toFixed(2)}`}
                  disabled
                />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">المخزون</label>
                <Input value={selectedProduct.stock || 0} disabled />
              </div>
            </div>
            <div className="flex gap-sm justify-end">
              <button
                onClick={() => setIsViewModalOpen(false)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '8px',
                  minWidth: '100px',
                  padding: '10px 20px',
                  borderRadius: '12px',
                  background: 'var(--bg-surface-elevated)',
                  border: '1px solid var(--border-default)',
                  color: 'var(--text-primary)',
                  fontSize: '13px',
                  fontWeight: '600',
                  letterSpacing: '0.3px',
                  cursor: 'pointer',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  boxShadow: '0 4px 20px rgba(0, 0, 0, 0.1), 0 1px 3px rgba(0, 0, 0, 0.05)',
                  position: 'relative',
                  overflow: 'hidden'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                  e.currentTarget.style.borderColor = 'var(--primary)';
                  e.currentTarget.style.boxShadow = '0 8px 30px rgba(99, 102, 241, 0.2), 0 2px 8px rgba(0, 0, 0, 0.1)';
                  e.currentTarget.style.transform = 'translateY(-2px) scale(1.02)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                  e.currentTarget.style.borderColor = 'var(--border-default)';
                  e.currentTarget.style.boxShadow = '0 4px 20px rgba(0, 0, 0, 0.1), 0 1px 3px rgba(0, 0, 0, 0.05)';
                  e.currentTarget.style.transform = 'translateY(0) scale(1)';
                }}
              >
                إغلاق
              </button>
            </div>
          </div>
        )}
      </Modal>

      {/* Edit Product Modal */}
      <Modal
        isOpen={isEditModalOpen}
        onClose={() => {
          setIsEditModalOpen(false);
          setSelectedProduct(null);
        }}
        title={isCreatingProduct ? "إضافة منتج جديد" : "تعديل المنتج"}
        variant="modern"
        size="lg"
        autoFocus={true}
        enableEnterNavigation={true}
        className="modal-custom-style"
        style={{
          borderRadius: '24px',
          overflow: 'hidden',
          background: 'var(--bg-surface)',
          border: '1px solid var(--border-primary)',
          boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(255, 255, 255, 0.05) inset, 0 0 40px rgba(99, 102, 241, 0.1)'
        }}
      >
        {selectedProduct && (
          <ProductImageField
            value={selectedProduct.image_url}
            onChange={(image_url) => setSelectedProduct({ ...selectedProduct, image_url })}
          />
        )}
        {selectedProduct && !isCreatingProduct ? (
          <div className="space-y-md">
            {/* Basic Information Section */}
            <div style={{ 
              marginBottom: '20px',
              paddingBottom: '20px',
              borderBottom: '1px solid var(--border-subtle)'
            }}>
              <div style={{ 
                display: 'flex', 
                alignItems: 'center', 
                gap: '10px',
                marginBottom: '16px',
                padding: '10px 14px',
                background: 'var(--bg-surface-elevated)',
                borderRadius: '12px',
                border: '1px solid var(--border-subtle)'
              }}>
                <div style={{
                  width: '32px',
                  height: '32px',
                  borderRadius: '8px',
                  background: 'var(--primary)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: '0 4px 12px rgba(99, 102, 241, 0.3)'
                }}>
                  <Package className="w-4 h-4" style={{ color: 'var(--text-on-primary)' }} />
                </div>
                <h4 style={{ 
                  fontSize: '13px', 
                  fontWeight: '600', 
                  color: 'var(--text-primary)',
                  margin: 0,
                  letterSpacing: '0.2px'
                }}>
                  المعلومات الأساسية
                </h4>
              </div>
              
              <div style={{ 
                display: 'grid', 
                gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', 
                gap: '16px' 
              }}>
                <div>
                 <label style={{ 
                   fontSize: '12px', 
                   fontWeight: '600', 
                   color: 'var(--text-secondary)',
                   marginBottom: '8px',
                   display: 'block',
                   letterSpacing: '0.2px'
                 }}>
                   سعر التكلفة (₪)
                   <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                 </label>
                 <Input
                   type="number"
                   value={selectedProduct?.costPrice || ''}
                   onChange={(e) => {
                     const costPrice = Number(e.target.value);
                     setSelectedProduct((prev) => prev ? {
                       ...prev,
                       costPrice,
                       sellingPrice: prev.sellingPrice > 0
                         ? prev.sellingPrice
                         : calculateSuggestedSellingPrice(costPrice, profitMargin),
                     } : prev);
                   }}
                   placeholder="0.00"
                   min="0"
                   step="0.01"
                 />
               </div>
               <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    اسم المنتج
                    <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                  </label>
                  <Input 
                    value={selectedProduct.name}
                    onChange={(e) => setSelectedProduct({ ...selectedProduct, name: e.target.value })}
                    placeholder="أدخل اسم المنتج"
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      borderRadius: '10px'
                    }}
                  />
                </div>
                <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    رمز المنتج (SKU)
                  </label>
                  <Input 
                    value={selectedProduct.sku}
                    onChange={(e) => setSelectedProduct({ ...selectedProduct, sku: e.target.value })}
                    placeholder="مثال: CPU-001"
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      borderRadius: '10px'
                    }}
                  />
                </div>
              </div>
            </div>

            {/* Classification Section */}
            <div style={{ 
              marginBottom: '20px',
              paddingBottom: '20px',
              borderBottom: '1px solid var(--border-subtle)'
            }}>
              <div style={{ 
                display: 'flex', 
                alignItems: 'center', 
                gap: '10px',
                marginBottom: '16px',
                padding: '10px 14px',
                background: 'var(--bg-surface-elevated)',
                borderRadius: '12px',
                border: '1px solid var(--border-subtle)'
              }}>
                <div style={{
                  width: '32px',
                  height: '32px',
                  borderRadius: '8px',
                  background: 'var(--info)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: '0 4px 12px rgba(59, 130, 246, 0.3)'
                }}>
                  <Tag className="w-4 h-4" style={{ color: 'var(--text-on-primary)' }} />
                </div>
                <h4 style={{ 
                  fontSize: '13px', 
                  fontWeight: '600', 
                  color: 'var(--text-primary)',
                  margin: 0,
                  letterSpacing: '0.2px'
                }}>
                  التصنيف
                </h4>
              </div>
              
              <div style={{ 
                display: 'grid', 
                gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', 
                gap: '16px' 
              }}>
                <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    التصنيف
                  </label>
                  <Select
                    value={selectedProduct.category_id || ''}
                    onChange={(e) => setSelectedProduct({ ...selectedProduct, category_id: e.target.value })}
                    options={[
                      { value: '', label: 'بدون تصنيف' },
                      ...categories.map((cat: any) => ({ value: cat.id, label: cat.name }))
                    ]}
                  />
                </div>
              </div>
            </div>

            {/* Pricing & Inventory Section */}
            <div style={{ marginBottom: '20px' }}>
              <div style={{ 
                display: 'flex', 
                alignItems: 'center', 
                gap: '10px',
                marginBottom: '16px',
                padding: '10px 14px',
                background: 'var(--bg-surface-elevated)',
                borderRadius: '12px',
                border: '1px solid var(--border-subtle)'
              }}>
                <div style={{
                  width: '32px',
                  height: '32px',
                  borderRadius: '8px',
                  background: 'var(--success)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: '0 4px 12px rgba(34, 197, 94, 0.3)'
                }}>
                  <DollarSign className="w-4 h-4" style={{ color: 'var(--text-on-primary)' }} />
                </div>
                <h4 style={{ 
                  fontSize: '13px', 
                  fontWeight: '600', 
                  color: 'var(--text-primary)',
                  margin: 0,
                  letterSpacing: '0.2px'
                }}>
                  التسعير والمخزون
                </h4>
              </div>
              
              <div style={{ 
                display: 'grid', 
                gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', 
                gap: '16px' 
              }}>
                <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    سعر التكلفة (₪)
                  </label>
                  <Input 
                    type="number"
                    value={selectedProduct.costPrice || ''}
                    onChange={(e) => {
                      const costPrice = Number(e.target.value);
                      setSelectedProduct({
                        ...selectedProduct,
                        costPrice,
                        sellingPrice: selectedProduct.sellingPrice > 0
                          ? selectedProduct.sellingPrice
                          : calculateSuggestedSellingPrice(costPrice, profitMargin),
                      });
                    }}
                    placeholder="0.00"
                    min="0"
                    step="0.01"
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      borderRadius: '10px'
                    }}
                  />
                </div>
                <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    سعر البيع قبل الضريبة (₪)
                    <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                  </label>
                  <Input 
                    type="number"
                    value={selectedProduct.sellingPrice || ''}
                    onChange={(e) => setSelectedProduct({ ...selectedProduct, sellingPrice: Number(e.target.value) })}
                    placeholder="0.00"
                    min="0"
                    step="0.01"
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      borderRadius: '10px'
                    }}
                  />
                </div>
                <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    الكمية بدون فاتورة
                    <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                  </label>
                  <Input 
                    type="number"
                    value={selectedProduct.stock}
                    onChange={(e) => setSelectedProduct({ ...selectedProduct, stock: Number(e.target.value) })}
                    placeholder="0"
                    min="0"
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      borderRadius: '10px'
                    }}
                  />
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div style={{ 
              display: 'flex', 
              gap: '12px', 
              justifyContent: 'flex-end',
              paddingTop: '24px',
              borderTop: '1px solid var(--border-subtle)',
              marginTop: '16px'
            }}>
              <button
                onClick={() => {
                  setIsEditModalOpen(false);
                  setSelectedProduct(null);
                }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '8px',
                  minWidth: '100px',
                  padding: '10px 20px',
                  borderRadius: '12px',
                  background: 'var(--bg-surface-elevated)',
                  border: '1px solid var(--border-default)',
                  color: 'var(--text-primary)',
                  fontSize: '13px',
                  fontWeight: '600',
                  letterSpacing: '0.3px',
                  cursor: 'pointer',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  boxShadow: '0 4px 20px rgba(0, 0, 0, 0.1), 0 1px 3px rgba(0, 0, 0, 0.05)',
                  position: 'relative',
                  overflow: 'hidden'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                  e.currentTarget.style.borderColor = 'var(--primary)';
                  e.currentTarget.style.boxShadow = '0 8px 30px rgba(99, 102, 241, 0.2), 0 2px 8px rgba(0, 0, 0, 0.1)';
                  e.currentTarget.style.transform = 'translateY(-2px) scale(1.02)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                  e.currentTarget.style.borderColor = 'var(--border-default)';
                  e.currentTarget.style.boxShadow = '0 4px 20px rgba(0, 0, 0, 0.1), 0 1px 3px rgba(0, 0, 0, 0.05)';
                  e.currentTarget.style.transform = 'translateY(0) scale(1)';
                }}
              >
                إلغاء
              </button>
              <button
                onClick={() => onSaveProduct(selectedProduct)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '8px',
                  minWidth: '120px',
                  padding: '10px 20px',
                  borderRadius: '12px',
                  background: 'var(--primary)',
                  border: '1px solid var(--primary)',
                  color: 'var(--text-on-primary)',
                  fontSize: '13px',
                  fontWeight: '600',
                  letterSpacing: '0.3px',
                  cursor: 'pointer',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  boxShadow: '0 4px 20px rgba(99, 102, 241, 0.3), 0 1px 3px rgba(99, 102, 241, 0.1)',
                  position: 'relative',
                  overflow: 'hidden'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'var(--primary-hover)';
                  e.currentTarget.style.borderColor = 'var(--primary-hover)';
                  e.currentTarget.style.boxShadow = '0 8px 30px rgba(99, 102, 241, 0.4), 0 2px 8px rgba(99, 102, 241, 0.2)';
                  e.currentTarget.style.transform = 'translateY(-2px) scale(1.02)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'var(--primary)';
                  e.currentTarget.style.borderColor = 'var(--primary)';
                  e.currentTarget.style.boxShadow = '0 4px 20px rgba(99, 102, 241, 0.3), 0 1px 3px rgba(99, 102, 241, 0.1)';
                  e.currentTarget.style.transform = 'translateY(0) scale(1)';
                }}
              >
                <Sparkles className="w-4 h-4" />
                حفظ التغييرات
              </button>
            </div>
          </div>
        ) : (
          <div className="space-y-md">
            {/* Basic Information Section */}
            <div style={{ 
              marginBottom: '20px',
              paddingBottom: '20px',
              borderBottom: '1px solid var(--border-subtle)'
            }}>
              <div style={{ 
                display: 'flex', 
                alignItems: 'center', 
                gap: '10px',
                marginBottom: '16px',
                padding: '10px 14px',
                background: 'var(--bg-surface-elevated)',
                borderRadius: '12px',
                border: '1px solid var(--border-subtle)'
              }}>
                <div style={{
                  width: '32px',
                  height: '32px',
                  borderRadius: '8px',
                  background: 'var(--primary)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: '0 4px 12px rgba(99, 102, 241, 0.3)'
                }}>
                  <Package className="w-4 h-4" style={{ color: 'var(--text-on-primary)' }} />
                </div>
                <h4 style={{ 
                  fontSize: '13px', 
                  fontWeight: '600', 
                  color: 'var(--text-primary)',
                  margin: 0,
                  letterSpacing: '0.2px'
                }}>
                  المعلومات الأساسية
                </h4>
              </div>
              
              <div style={{ 
                display: 'grid', 
                gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', 
                gap: '16px' 
              }}>
                <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    اسم المنتج
                    <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                  </label>
                  <Input 
                    placeholder="أدخل اسم المنتج"
                    onChange={(e) => setSelectedProduct((prev) => prev ? { ...prev, name: e.target.value } : prev)}
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      borderRadius: '10px'
                    }}
                  />
                </div>
                <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    سعر التكلفة (₪)
                    <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                  </label>
                  <Input
                    type="number"
                    value={selectedProduct?.costPrice || ''}
                    onChange={(e) => {
                      const costPrice = Number(e.target.value);
                      setSelectedProduct((prev) => prev ? {
                        ...prev,
                        costPrice,
                        sellingPrice: prev.sellingPrice > 0
                          ? prev.sellingPrice
                          : calculateSuggestedSellingPrice(costPrice, profitMargin),
                      } : prev);
                    }}
                    placeholder="0.00"
                    min="0"
                    step="0.01"
                  />
                </div>
                <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    رمز المنتج (SKU)
                  </label>
                  <Input 
                    placeholder="مثال: CPU-001"
                    onChange={(e) => setSelectedProduct((prev: Product | null) => ({ ...prev, sku: e.target.value } as Product))}
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      borderRadius: '10px'
                    }}
                  />
                </div>
              </div>
            </div>

            {/* Classification Section */}
            <div style={{ 
              marginBottom: '20px',
              paddingBottom: '20px',
              borderBottom: '1px solid var(--border-subtle)'
            }}>
              <div style={{ 
                display: 'flex', 
                alignItems: 'center', 
                gap: '10px',
                marginBottom: '16px',
                padding: '10px 14px',
                background: 'var(--bg-surface-elevated)',
                borderRadius: '12px',
                border: '1px solid var(--border-subtle)'
              }}>
                <div style={{
                  width: '32px',
                  height: '32px',
                  borderRadius: '8px',
                  background: 'var(--info)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: '0 4px 12px rgba(59, 130, 246, 0.3)'
                }}>
                  <Tag className="w-4 h-4" style={{ color: 'var(--text-on-primary)' }} />
                </div>
                <h4 style={{ 
                  fontSize: '13px', 
                  fontWeight: '600', 
                  color: 'var(--text-primary)',
                  margin: 0,
                  letterSpacing: '0.2px'
                }}>
                  التصنيف
                </h4>
              </div>
              
              <div style={{ 
                display: 'grid', 
                gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', 
                gap: '16px' 
              }}>
                <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    التصنيف
                  </label>
                  <Select
                    value={selectedProduct?.category_id || ''}
                    onChange={(e) => setSelectedProduct((prev: Product | null) => ({ ...prev, category_id: e.target.value } as Product))}
                    options={[
                      { value: '', label: 'بدون تصنيف' },
                      ...categories.map((cat: any) => ({ value: cat.id, label: cat.name }))
                    ]}
                  />
                </div>
              </div>
            </div>

            {/* Pricing & Inventory Section */}
            <div style={{ marginBottom: '20px' }}>
              <div style={{ 
                display: 'flex', 
                alignItems: 'center', 
                gap: '10px',
                marginBottom: '16px',
                padding: '10px 14px',
                background: 'var(--bg-surface-elevated)',
                borderRadius: '12px',
                border: '1px solid var(--border-subtle)'
              }}>
                <div style={{
                  width: '32px',
                  height: '32px',
                  borderRadius: '8px',
                  background: 'var(--success)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: '0 4px 12px rgba(34, 197, 94, 0.3)'
                }}>
                  <DollarSign className="w-4 h-4" style={{ color: 'var(--text-on-primary)' }} />
                </div>
                <h4 style={{ 
                  fontSize: '13px', 
                  fontWeight: '600', 
                  color: 'var(--text-primary)',
                  margin: 0,
                  letterSpacing: '0.2px'
                }}>
                  التسعير والمخزون
                </h4>
              </div>
              
              <div style={{ 
                display: 'grid', 
                gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', 
                gap: '16px' 
              }}>
                <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    سعر البيع قبل الضريبة (₪)
                    <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                  </label>
                  <Input
                    type="number"
                  value={selectedProduct?.sellingPrice || ''}
                    placeholder="0.00"
                    min="0"
                    step="0.01"
                    onChange={(e) => setSelectedProduct((prev: Product | null) => ({ ...prev, sellingPrice: Number(e.target.value) } as Product))}
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      borderRadius: '10px'
                    }}
                  />
                </div>
                <div>
                  <label style={{ 
                    fontSize: '12px', 
                    fontWeight: '600', 
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    الكمية بدون فاتورة
                    <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                  </label>
                  <Input 
                    type="number"
                    placeholder="0"
                    min="0"
                    onChange={(e) => setSelectedProduct((prev: Product | null) => ({ ...prev, stock: Number(e.target.value) } as Product))}
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      borderRadius: '10px'
                    }}
                  />
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div style={{ 
              display: 'flex', 
              gap: '12px', 
              justifyContent: 'flex-end',
              paddingTop: '24px',
              borderTop: '1px solid var(--border-subtle)',
              marginTop: '16px'
            }}>
              <button
                onClick={() => {
                  setIsEditModalOpen(false);
                  setSelectedProduct(null);
                }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '8px',
                  minWidth: '100px',
                  padding: '10px 20px',
                  borderRadius: '12px',
                  background: 'var(--bg-surface-elevated)',
                  border: '1px solid var(--border-default)',
                  color: 'var(--text-primary)',
                  fontSize: '13px',
                  fontWeight: '600',
                  letterSpacing: '0.3px',
                  cursor: 'pointer',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  boxShadow: '0 4px 20px rgba(0, 0, 0, 0.1), 0 1px 3px rgba(0, 0, 0, 0.05)',
                  position: 'relative',
                  overflow: 'hidden'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                  e.currentTarget.style.borderColor = 'var(--primary)';
                  e.currentTarget.style.boxShadow = '0 8px 30px rgba(99, 102, 241, 0.2), 0 2px 8px rgba(0, 0, 0, 0.1)';
                  e.currentTarget.style.transform = 'translateY(-2px) scale(1.02)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                  e.currentTarget.style.borderColor = 'var(--border-default)';
                  e.currentTarget.style.boxShadow = '0 4px 20px rgba(0, 0, 0, 0.1), 0 1px 3px rgba(0, 0, 0, 0.05)';
                  e.currentTarget.style.transform = 'translateY(0) scale(1)';
                }}
              >
                إلغاء
              </button>
              <button
                onClick={() => onSaveProduct(selectedProduct as Product)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '8px',
                  minWidth: '120px',
                  padding: '10px 20px',
                  borderRadius: '12px',
                  background: 'var(--primary)',
                  border: '1px solid var(--primary)',
                  color: 'var(--text-on-primary)',
                  fontSize: '13px',
                  fontWeight: '600',
                  letterSpacing: '0.3px',
                  cursor: 'pointer',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  boxShadow: '0 4px 20px rgba(99, 102, 241, 0.3), 0 1px 3px rgba(99, 102, 241, 0.1)',
                  position: 'relative',
                  overflow: 'hidden'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'var(--primary-hover)';
                  e.currentTarget.style.borderColor = 'var(--primary-hover)';
                  e.currentTarget.style.boxShadow = '0 8px 30px rgba(99, 102, 241, 0.4), 0 2px 8px rgba(99, 102, 241, 0.2)';
                  e.currentTarget.style.transform = 'translateY(-2px) scale(1.02)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'var(--primary)';
                  e.currentTarget.style.borderColor = 'var(--primary)';
                  e.currentTarget.style.boxShadow = '0 4px 20px rgba(99, 102, 241, 0.3), 0 1px 3px rgba(99, 102, 241, 0.1)';
                  e.currentTarget.style.transform = 'translateY(0) scale(1)';
                }}
              >
                <Plus className="w-4 h-4" />
                إضافة المنتج
              </button>
            </div>
          </div>
        )}
      </Modal>
    </>
  );
}