import { Modal } from '../../../design-system/components/modal';
import { Input } from '../../../design-system/components/input';
import { Select } from '../../../design-system/components/select';
import { Button } from '../../../design-system/components/button';
import { Product } from '../types/inventory.types';
import { Package, Plus, Sparkles, Tag, DollarSign } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { categoriesApi, settingsApi, suppliersApi } from '../../../services/api/endpoints';
import { calculateSuggestedSellingPrice, DEFAULT_PROFIT_MARGIN } from '../../../utils/pricing';
import { formatPrice } from '../../../utils/helpers';
import { clearBarcodeLookupFields, lookupProductByBarcode } from '../../../lib/productBarcodeLookup';
import { toast } from 'sonner';

interface InventoryModalsProps {
  isViewModalOpen: boolean;
  setIsViewModalOpen: (open: boolean) => void;
  isEditModalOpen: boolean;
  setIsEditModalOpen: (open: boolean) => void;
  isCreatingProduct: boolean;
  selectedProduct: Product | null;
  setSelectedProduct: (product: Product | null) => void;
  onSaveProduct: (productData: Product) => Promise<boolean>;
  onSaveProductAndReturnToSales?: (productData: Product) => Promise<void>;
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
  onSaveProductAndReturnToSales,
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
  const { data: suppliersData } = useQuery({
    queryKey: ['suppliers', 'inventory-product-edit'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 100, is_active: true }),
  });

  const categories = (categoriesData?.data as unknown) as any[] || [];
  const suppliersPayload = suppliersData?.data as any;
  const suppliers = (Array.isArray(suppliersPayload)
    ? suppliersPayload
    : suppliersPayload?.suppliers || (suppliersData as any)?.suppliers || [])
    .map((supplier: any) => ({
      id: supplier.id,
      name: supplier.name || supplier.supplier_name || 'تاجر بدون اسم',
    }));
  const taxRateValue = Number(taxSetting?.data?.value);
  const taxRate = Number.isFinite(taxRateValue) && taxRateValue >= 0 ? taxRateValue : 0;
  const configuredMargin = Number(marginSetting?.data?.value);
  const profitMargin = Number.isFinite(configuredMargin) && configuredMargin >= 0 && configuredMargin < 100
    ? configuredMargin
    : DEFAULT_PROFIT_MARGIN;
  const [productStep, setProductStep] = useState<1 | 2 | 3 | 4>(1);

  const goToNextStep = () => {
    setProductStep((currentStep) => {
      if (currentStep === 1) return 2;
      if (currentStep === 2) return 3;
      if (currentStep === 3) return 4;
      return currentStep;
    });
  };

  const canAdvanceToNextStep = () => {
    if (productStep === 1) return Boolean(selectedProduct?.category_id);
    if (productStep === 2) return Boolean(selectedProduct?.name?.trim() && (selectedProduct.costPrice ?? 0) > 0);
    if (productStep === 3) return Boolean((selectedProduct?.sellingPrice ?? 0) > 0);
    return Boolean(selectedProduct);
  };

  useEffect(() => {
    if (!isEditModalOpen || !isCreatingProduct) return;
    setProductStep(1);
    requestAnimationFrame(() => {
      document.querySelector<HTMLElement>('.product-create-stage input:not([disabled]), .product-create-stage select:not([disabled])')?.focus();
    });
  }, [isEditModalOpen, isCreatingProduct]);

  useEffect(() => {
    if (!isEditModalOpen || !isCreatingProduct || !selectedProduct?.barcode) return;

    const barcode = selectedProduct.barcode.trim();
    if (!barcode) return;

    let active = true;
    setSelectedProduct((prev) => prev ? clearBarcodeLookupFields(prev) : prev);
    const timer = window.setTimeout(() => {
      void (async () => {
        const match = await lookupProductByBarcode(barcode);
        if (!active) return;
        if (!match) {
          toast.info('لم يتم العثور على بيانات لهذا الباركود. يمكنك إكمال الإدخال يدويًا.');
          return;
        }

        setSelectedProduct((prev) => {
          if (!prev) return prev;
          return {
            ...prev,
            name: prev.name || match.name || prev.name,
            barcode: match.barcode || prev.barcode,
            image_url: prev.image_url || match.image || prev.image_url,
            category: prev.category || match.category || prev.category,
          };
        });
      })();
    }, 300);

    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [isEditModalOpen, isCreatingProduct, selectedProduct?.barcode]);

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
          <div className="product-edit-fields space-y-md">
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
                <label className="text-small font-medium text-text mb-sm block">التاجر</label>
                <Input value={selectedProduct.supplier_name || 'غير محدد'} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">سعر التكلفة</label>
                <Input value={formatPrice(selectedProduct.costPrice || 0)} disabled />
              </div>
              <div>
                <label className="text-small font-medium text-text mb-sm block">سعر البيع قبل الضريبة</label>
                <Input value={formatPrice(selectedProduct.sellingPrice || 0)} disabled />
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
                  value={formatPrice((Number(selectedProduct.sellingPrice) || 0) * (1 + taxRate / 100))}
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
        size="xl"
        autoFocus={true}
        enableEnterNavigation={true}
        className="modal-custom-style inventory-product-edit-modal"
        style={{
          borderRadius: '24px',
          overflow: 'hidden',
          background: 'var(--bg-surface)',
          border: '1px solid var(--border-primary)',
          boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(255, 255, 255, 0.05) inset, 0 0 40px rgba(99, 102, 241, 0.1)'
        }}
      >
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
                       sellingPrice: calculateSuggestedSellingPrice(costPrice, profitMargin),
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
                <div>
                  <label style={{
                    fontSize: '12px',
                    fontWeight: '600',
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    display: 'block',
                    letterSpacing: '0.2px'
                  }}>
                    الحد الأدنى للمخزون
                  </label>
                  <Input
                    type="number"
                    value={selectedProduct.min_stock_level ?? 0}
                    onChange={(e) => setSelectedProduct({ ...selectedProduct, min_stock_level: Math.max(0, Number(e.target.value) || 0) })}
                    placeholder="3"
                    min="0"
                    step="1"
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      borderRadius: '10px'
                    }}
                  />
                </div>
                <div>
                  <label style={{ fontSize: '12px', fontWeight: '600', color: 'var(--text-secondary)', marginBottom: '8px', display: 'block' }}>
                    سعر البيع قبل الضريبة (₪)
                    <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                  </label>
                  <Input
                    type="number"
                    value={selectedProduct.sellingPrice || ''}
                    onChange={(e) => setSelectedProduct({ ...selectedProduct, sellingPrice: Number(e.target.value) })}
                    placeholder="0.00"
                    min="0"
                    step="1"
                  />
                </div>
                <div>
                  <label style={{ fontSize: '12px', fontWeight: '600', color: 'var(--text-secondary)', marginBottom: '8px', display: 'block' }}>
                    الكمية الحالية
                    <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                  </label>
                  <Input
                    type="number"
                    value={selectedProduct.stock}
                    onChange={(e) => setSelectedProduct({ ...selectedProduct, stock: Number(e.target.value) })}
                    placeholder="0"
                    min="0"
                  />
                </div>
                <div>
                  <label style={{ fontSize: '12px', fontWeight: '600', color: 'var(--text-secondary)', marginBottom: '8px', display: 'block' }}>
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
                <div>
                  <label style={{ fontSize: '12px', fontWeight: '600', color: 'var(--text-secondary)', marginBottom: '8px', display: 'block' }}>
                    المورد
                  </label>
                  <Select
                    value={selectedProduct.supplier_id || ''}
                    onChange={(e) => {
                      const supplier = suppliers.find((item: any) => String(item.id) === e.target.value);
                      setSelectedProduct({
                        ...selectedProduct,
                        supplier_id: e.target.value,
                        supplier_name: supplier?.name || '',
                      });
                    }}
                    options={[
                      { value: '', label: 'بدون مورد' },
                      ...suppliers.map((supplier: any) => ({ value: supplier.id, label: supplier.name })),
                    ]}
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
              {onSaveProductAndReturnToSales && (
                <Button type="button" variant="success" size="sm" onClick={() => { void onSaveProductAndReturnToSales(selectedProduct); }}>
                  حفظ والعودة لنقطة البيع
                </Button>
              )}
            </div>
          </div>
        ) : (
          <div
            className="product-create-stage space-y-md"
            onKeyDown={(event) => {
              if (event.key !== 'Enter' || event.shiftKey) return;
              const target = event.target as HTMLElement;
              const isTextField = ['INPUT', 'SELECT', 'TEXTAREA'].includes(target.tagName);
              if (!isTextField) return;

              const isSubmitTrigger = target instanceof HTMLButtonElement || target.type === 'submit';
              if (isSubmitTrigger) return;

              event.preventDefault();
              event.stopPropagation();

              if (productStep === 1) {
                if (!selectedProduct?.category_id) return;
                goToNextStep();
              } else if (productStep === 2) {
                if (!selectedProduct?.name?.trim() || (selectedProduct.costPrice ?? 0) <= 0) return;
                goToNextStep();
              } else if (productStep === 3) {
                if ((selectedProduct?.sellingPrice ?? 0) <= 0) return;
                goToNextStep();
              } else if (selectedProduct) {
                onSaveProduct(selectedProduct);
              }
            }}
          >
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2" aria-label={`مرحلة المنتج ${productStep} من 4`}>
                {[1, 2, 3, 4].map((step) => <span key={step} className={`h-1.5 rounded-full transition-all duration-150 ${step === productStep ? 'w-6 bg-primary' : 'w-1.5 bg-border'}`} />)}
              </div>
              {productStep > 1 && <Button type="button" variant="ghost" size="sm" onClick={() => setProductStep((step) => Math.max(1, step - 1) as 1 | 2 | 3 | 4)}>رجوع</Button>}
            </div>

            {productStep === 2 && (
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
                    value={selectedProduct?.name || ''}
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
                        sellingPrice: calculateSuggestedSellingPrice(costPrice, profitMargin),
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
                    value={selectedProduct?.sku || ''}
                    placeholder="مثال: CPU-001"
                    onChange={(e) => setSelectedProduct((prev: Product | null) => ({ ...prev, sku: e.target.value } as Product))}
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
                    display: 'block'
                  }}>
                    الباركود
                  </label>
                  <Input
                    value={selectedProduct?.barcode || ''}
                    placeholder="امسح أو أدخل الباركود"
                    autoFocus={isCreatingProduct}
                    onChange={(e) => setSelectedProduct((prev: Product | null) => ({ ...prev, barcode: e.target.value } as Product))}
                  />
                </div>
              </div>
            </div>
            )}

            {productStep === 1 && (<>
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
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', marginTop: '16px' }}>
              <Button
                type="button"
                variant="primary"
                size="sm"
                disabled={!selectedProduct?.category_id}
                onClick={goToNextStep}
              >
                التالي
              </Button>
            </div>
            </>)}

            {productStep === 2 && (
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', marginTop: '16px' }}>
                <Button
                  type="button"
                  variant="primary"
                  size="sm"
                  disabled={!canAdvanceToNextStep()}
                  onClick={goToNextStep}
                >
                  التالي
                </Button>
              </div>
            )}

            {productStep === 3 && (<>
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
                    step="1"
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
                    الكمية الافتتاحية
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
                  <p style={{ margin: '5px 0 0', color: 'var(--text-tertiary)', fontSize: '11px' }}>
                    تُضاف مباشرة إلى المخزون دون إنشاء فاتورة شراء
                  </p>
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
                    الحد الأدنى للمخزون
                  </label>
                  <Input
                    type="number"
                    value={selectedProduct?.min_stock_level ?? 0}
                    onChange={(e) => setSelectedProduct((prev: Product | null) => ({
                      ...prev,
                      min_stock_level: Math.max(0, Number(e.target.value) || 0),
                    } as Product))}
                    placeholder="3"
                    min="0"
                    step="1"
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      borderRadius: '10px'
                    }}
                  />
                </div>
              </div>
            </div>
            </>)}

            {productStep === 3 && (
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', marginTop: '16px' }}>
                <Button
                  type="button"
                  variant="primary"
                  size="sm"
                  disabled={!canAdvanceToNextStep()}
                  onClick={goToNextStep}
                >
                  التالي
                </Button>
              </div>
            )}

            {productStep === 4 && (<>
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
            </>)}
          </div>
        )}
      </Modal>
    </>
  );
}