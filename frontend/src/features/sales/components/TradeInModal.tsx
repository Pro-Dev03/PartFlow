import { useState } from 'react';
import { Modal } from '../../../components/ui/modal';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { TradeInFormData } from '../types/pos.types';
import { ShoppingCart, Recycle } from 'lucide-react';

interface TradeInModalProps {
  isOpen: boolean;
  onClose: () => void;
  customers: any[];
  products: any[];
  partTypes: any[];
  customersLoading: boolean;
  productsLoading: boolean;
  partTypesLoading: boolean;
  onSubmit: (data: TradeInFormData) => Promise<void>;
}

export function TradeInModal({
  isOpen,
  onClose,
  customers,
  products,
  partTypes,
  customersLoading,
  productsLoading,
  partTypesLoading,
  onSubmit,
}: TradeInModalProps) {
  const [isCustomerManual, setIsCustomerManual] = useState(false);
  const [tradeInCustomer, setTradeInCustomer] = useState('');
  const [tradeInCustomerManual, setTradeInCustomerManual] = useState('');
  
  const [isProductManual, setIsProductManual] = useState(false);
  const [tradeInProduct, setTradeInProduct] = useState('');
  const [tradeInProductManual, setTradeInProductManual] = useState('');
  
  const [tradeInPartType, setTradeInPartType] = useState('');
  const [tradeInPrice, setTradeInPrice] = useState('');
  const [tradeInSpecifications, setTradeInSpecifications] = useState<any[]>([]);

  const handleSubmit = async () => {
    const customerValue = isCustomerManual ? tradeInCustomerManual : tradeInCustomer;
    const productValue = isProductManual ? tradeInProductManual : tradeInProduct;

    if (!customerValue || !tradeInPrice || !tradeInPartType) {
      alert('يرجى ملء جميع الحقول المطلوبة');
      return;
    }

    const data: TradeInFormData = {
      customerId: isCustomerManual ? undefined : tradeInCustomer,
      customerName: isCustomerManual ? tradeInCustomerManual : undefined,
      productId: (isProductManual || !productValue) ? undefined : tradeInProduct,
      productName: isProductManual ? tradeInProductManual : undefined,
      partTypeId: tradeInPartType,
      purchaseCost: parseFloat(tradeInPrice) * 100, // تحويل للسنت
      specifications: tradeInSpecifications,
    };

    await onSubmit(data);

    // Reset form
    setTradeInCustomer('');
    setTradeInCustomerManual('');
    setIsCustomerManual(false);
    setTradeInProduct('');
    setTradeInProductManual('');
    setIsProductManual(false);
    setTradeInPrice('');
    setTradeInPartType('');
    setTradeInSpecifications([]);
    onClose();
  };

  const handleCancel = () => {
    setTradeInCustomer('');
    setTradeInCustomerManual('');
    setIsCustomerManual(false);
    setTradeInProduct('');
    setTradeInProductManual('');
    setIsProductManual(false);
    setTradeInPrice('');
    setTradeInPartType('');
    setTradeInSpecifications([]);
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleCancel}
      title="شراء قطع مستعملة"
      variant="modern"
      size="lg"
      style={{
        borderRadius: '24px',
        overflow: 'hidden',
        background: 'var(--bg-surface)',
        border: '1px solid var(--border-primary)',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(255, 255, 255, 0.05) inset, 0 0 40px rgba(99, 102, 241, 0.1)'
      }}
    >
      <div className="space-y-5">
        {/* الزبون */}
        <div>
          <label style={{ 
            fontSize: '12px', 
            fontWeight: '600', 
            color: 'var(--text-secondary)',
            marginBottom: '8px',
            display: 'block',
            letterSpacing: '0.2px'
          }}>
            الزبون
            <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
          </label>
          <div style={{ display: 'flex', gap: '12px', marginBottom: '12px' }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', color: 'var(--text-primary)', cursor: 'pointer' }}>
              <input
                type="radio"
                checked={!isCustomerManual}
                onChange={() => setIsCustomerManual(false)}
                style={{ width: '16px', height: '16px', accentColor: 'var(--primary)' }}
              />
              <span>اختر من القائمة</span>
            </label>
            <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', color: 'var(--text-primary)', cursor: 'pointer' }}>
              <input
                type="radio"
                checked={isCustomerManual}
                onChange={() => setIsCustomerManual(true)}
                style={{ width: '16px', height: '16px', accentColor: 'var(--primary)' }}
              />
              <span>اكتب يدوياً</span>
            </label>
          </div>
          {!isCustomerManual ? (
            <Select
              value={tradeInCustomer}
              onChange={(e) => setTradeInCustomer(e.target.value)}
              loading={customersLoading}
              options={[
                { value: '', label: 'اختر الزبون...' },
                ...customers.map((c) => ({ value: c.id, label: c.name })),
              ]}
              emptyMessage="لا يوجد عملاء"
              style={{ borderRadius: '10px' }}
            />
          ) : (
            <Input
              type="text"
              value={tradeInCustomerManual}
              onChange={(e) => setTradeInCustomerManual(e.target.value)}
              placeholder="أدخل اسم الزبون..."
              style={{ borderRadius: '10px' }}
            />
          )}
        </div>

        {/* المنتج */}
        <div>
          <label style={{ 
            fontSize: '12px', 
            fontWeight: '600', 
            color: 'var(--text-secondary)',
            marginBottom: '8px',
            display: 'block',
            letterSpacing: '0.2px'
          }}>
            المنتج (اختياري إذا اخترت نوع القطعة)
          </label>
          <div style={{ display: 'flex', gap: '12px', marginBottom: '12px' }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', color: 'var(--text-primary)', cursor: 'pointer' }}>
              <input
                type="radio"
                checked={!isProductManual}
                onChange={() => setIsProductManual(false)}
                style={{ width: '16px', height: '16px', accentColor: 'var(--primary)' }}
              />
              <span>اختر من القائمة</span>
            </label>
            <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', color: 'var(--text-primary)', cursor: 'pointer' }}>
              <input
                type="radio"
                checked={isProductManual}
                onChange={() => setIsProductManual(true)}
                style={{ width: '16px', height: '16px', accentColor: 'var(--primary)' }}
              />
              <span>اكتب يدوياً</span>
            </label>
            <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', color: 'var(--text-primary)', cursor: 'pointer' }}>
              <input
                type="radio"
                checked={tradeInProduct === '' && tradeInProductManual === ''}
                onChange={() => {
                  setIsProductManual(false);
                  setTradeInProduct('');
                  setTradeInProductManual('');
                }}
                style={{ width: '16px', height: '16px', accentColor: 'var(--primary)' }}
              />
              <span>بدون منتج</span>
            </label>
          </div>
          {!isProductManual ? (
            <Select
              value={tradeInProduct}
              onChange={(e) => setTradeInProduct(e.target.value)}
              loading={productsLoading}
              options={[
                { value: '', label: 'اختر المنتج...' },
                ...products.map((p) => ({ value: p.id, label: p.name })),
              ]}
              emptyMessage="لا يوجد منتجات"
              style={{ borderRadius: '10px' }}
            />
          ) : (
            <Input
              type="text"
              value={tradeInProductManual}
              onChange={(e) => setTradeInProductManual(e.target.value)}
              placeholder="أدخل اسم المنتج..."
              style={{ borderRadius: '10px' }}
            />
          )}
        </div>

        {/* نوع القطعة */}
        <div>
          <label style={{ 
            fontSize: '12px', 
            fontWeight: '600', 
            color: 'var(--text-secondary)',
            marginBottom: '8px',
            display: 'block',
            letterSpacing: '0.2px'
          }}>
            نوع القطعة
            <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
          </label>
          <Select
            value={tradeInPartType}
            onChange={(e) => {
              setTradeInPartType(e.target.value);
              // Reset specifications when part type changes
              setTradeInSpecifications([]);
            }}
            loading={partTypesLoading}
            options={[
              { value: '', label: 'اختر نوع القطعة...' },
              ...partTypes.map((pt: any) => ({ value: pt.id, label: pt.name_ar })),
            ]}
            emptyMessage="لا يوجد أنواع قطع"
            style={{ borderRadius: '10px' }}
          />
        </div>

        {/* السعر */}
        <div>
          <label style={{ 
            fontSize: '12px', 
            fontWeight: '600', 
            color: 'var(--text-secondary)',
            marginBottom: '8px',
            display: 'block',
            letterSpacing: '0.2px'
          }}>
            السعر (كم دفعت للزبون)
            <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
          </label>
          <Input
            type="number"
            value={tradeInPrice}
            onChange={(e) => setTradeInPrice(e.target.value)}
            placeholder="أدخل السعر..."
            style={{ borderRadius: '10px' }}
          />
        </div>

        {/* مواصفات إضافية (سيتم تطويرها لاحقاً) */}
        {tradeInPartType && (
          <div style={{ 
            padding: '12px 16px', 
            background: 'var(--bg-surface-elevated)', 
            borderRadius: '12px',
            border: '1px solid var(--border-subtle)'
          }}>
            <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}>المواصفات التفصيلية (قريباً)</p>
            <p style={{ fontSize: '11px', color: 'var(--text-muted)' }}>سيتم إضافة المواصفات التفصيلية لكل نوع قطعة قريباً</p>
          </div>
        )}

        <div style={{ 
          display: 'flex', 
          gap: '12px', 
          justifyContent: 'flex-end',
          paddingTop: '24px',
          borderTop: '1px solid var(--border-subtle)',
          marginTop: '16px'
        }}>
          <button
            onClick={handleCancel}
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
            onClick={handleSubmit}
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
            <ShoppingCart className="w-4 h-4" style={{ position: 'relative', zIndex: 1 }} />
            <span style={{ position: 'relative', zIndex: 1 }}>شراء</span>
          </button>
        </div>
      </div>
    </Modal>
  );
}