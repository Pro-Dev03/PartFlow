import { useState } from 'react';
import { Modal } from '../../../components/ui/modal';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { TradeInFormData } from '../types/pos.types';

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
      size="sm"
    >
      <div className="space-y-4">
        {/* الزبون */}
        <div>
          <label className="block text-sm font-medium text-text mb-2">الزبون</label>
          <div className="flex gap-2 mb-2">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                checked={!isCustomerManual}
                onChange={() => setIsCustomerManual(false)}
                className="w-4 h-4"
              />
              <span>اختر من القائمة</span>
            </label>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                checked={isCustomerManual}
                onChange={() => setIsCustomerManual(true)}
                className="w-4 h-4"
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
            />
          ) : (
            <Input
              type="text"
              value={tradeInCustomerManual}
              onChange={(e) => setTradeInCustomerManual(e.target.value)}
              placeholder="أدخل اسم الزبون..."
            />
          )}
        </div>

        {/* المنتج */}
        <div>
          <label className="block text-sm font-medium text-text mb-2">المنتج (اختياري إذا اخترت نوع القطعة)</label>
          <div className="flex gap-2 mb-2">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                checked={!isProductManual}
                onChange={() => setIsProductManual(false)}
                className="w-4 h-4"
              />
              <span>اختر من القائمة</span>
            </label>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                checked={isProductManual}
                onChange={() => setIsProductManual(true)}
                className="w-4 h-4"
              />
              <span>اكتب يدوياً</span>
            </label>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                checked={tradeInProduct === '' && tradeInProductManual === ''}
                onChange={() => {
                  setIsProductManual(false);
                  setTradeInProduct('');
                  setTradeInProductManual('');
                }}
                className="w-4 h-4"
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
            />
          ) : (
            <Input
              type="text"
              value={tradeInProductManual}
              onChange={(e) => setTradeInProductManual(e.target.value)}
              placeholder="أدخل اسم المنتج..."
            />
          )}
        </div>

        {/* نوع القطعة */}
        <div>
          <label className="block text-sm font-medium text-text mb-2">نوع القطعة *</label>
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
          />
        </div>

        {/* السعر */}
        <div>
          <label className="block text-sm font-medium text-text mb-2">السعر (كم دفعت للزبون) *</label>
          <Input
            type="number"
            value={tradeInPrice}
            onChange={(e) => setTradeInPrice(e.target.value)}
            placeholder="أدخل السعر..."
          />
        </div>

        {/* مواصفات إضافية (سيتم تطويرها لاحقاً) */}
        {tradeInPartType && (
          <div className="p-3 bg-gray-800 rounded-lg">
            <p className="text-sm text-gray-400">المواصفات التفصيلية (قريباً)</p>
            <p className="text-xs text-gray-500">سيتم إضافة المواصفات التفصيلية لكل نوع قطعة قريباً</p>
          </div>
        )}

        <div className="flex justify-end gap-3 pt-4">
          <Button variant="secondary" onClick={handleCancel}>
            إلغاء
          </Button>
          <Button variant="primary" onClick={handleSubmit}>
            شراء
          </Button>
        </div>
      </div>
    </Modal>
  );
}