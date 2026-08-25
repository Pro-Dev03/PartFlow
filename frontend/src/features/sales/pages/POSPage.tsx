import { useState, useEffect, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { PageHeader } from '../../../components/ui/page-header';
import { Modal } from '../../../components/ui/modal';
import { Button } from '../../../components/ui/button';
import { productsApi, salesApi, customersApi, barcodeApi, inventoryApi, partTypesApi, categoriesApi } from '../../../services/api/endpoints';
import { UsedPartsInvoice } from '../../../components/invoice/UsedPartsInvoice';
import { ItemInputMethodType } from '../../../components/ui/item-input-method';
import { playScanSound } from '../../../hooks/useBarcodeContext';
import { Zap, Printer } from 'lucide-react';

// Custom hooks
import { useCart } from '../hooks/useCart';
import { usePayment } from '../hooks/usePayment';

// Components
import { BarcodeScanner } from '../components/BarcodeScanner';
import { ProductSearch } from '../components/ProductSearch';
import { CustomerSelector } from '../components/CustomerSelector';
import { TradeInItemsSection } from '../components/TradeInItemsSection';
import { CartSection } from '../components/CartSection';
import { PaymentSection } from '../components/PaymentSection';
import { CategoryFilter } from '../components/CategoryFilter';

// Types
import { InvoiceData } from '../types/pos.types';

export function POSPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  
  // Mobile detection
  const [isMobile, setIsMobile] = useState(false);
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768);
    };
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  // Custom hooks
  const { cart, addToCart, removeFromCart, updateQuantity, clearCart, total } = useCart(true);
  const { 
    paymentMethod, 
    setPaymentMethod, 
    paidAmount, 
    setPaidAmount, 
    isProcessing, 
    setProcessing, 
    calculateRemaining,
    resetPayment 
  } = usePayment();

  // Local state
  const [barcodeInput, setBarcodeInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCustomer, setSelectedCustomer] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [soundEnabled, setSoundEnabled] = useState(true);
  const [quickAddMode, setQuickAddMode] = useState(false);
  const [inputMethod, setInputMethod] = useState<'barcode' | 'camera'>('barcode');
  const [isCameraScannerOpen, setIsCameraScannerOpen] = useState(false);
  const [isInvoiceModalOpen, setIsInvoiceModalOpen] = useState(false);
  const [lastSaleData, setLastSaleData] = useState<InvoiceData | null>(null);

  // Fetch data
  const { data: productsData, isLoading: productsLoading } = useQuery({
    queryKey: ['products'],
    queryFn: () => productsApi.list({ page: 1, per_page: 100 }),
  });

  const { data: customersData, isLoading: customersLoading } = useQuery({
    queryKey: ['customers'],
    queryFn: () => customersApi.list({ page: 1, per_page: 100 }),
  });

  const { data: inventoryData } = useQuery({
    queryKey: ['inventory'],
    queryFn: () => inventoryApi.list({ page: 1, per_page: 100 }),
  });

  const { data: partTypesData, isLoading: partTypesLoading } = useQuery({
    queryKey: ['part-types'],
    queryFn: () => partTypesApi.list(),
  });

  const { data: categoriesData, isLoading: categoriesLoading } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });

  const products = (productsData?.data?.products as unknown) as any[] || [];
  const customers = (customersData?.data as unknown) as any[] || [];
  const inventoryItems = (inventoryData?.data?.items as unknown) as any[] || [];
  const partTypes = (partTypesData?.data as unknown) as any[] || [];
  const categories = (categoriesData?.data as unknown) as any[] || [];

  // Create sale mutation
  const createSaleMutation = useMutation({
    mutationFn: (data: any) => salesApi.create(data),
    onSuccess: (response: any) => {
      queryClient.invalidateQueries({ queryKey: ['sales'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      
      if (lastSaleData) {
        setLastSaleData({
          ...lastSaleData,
          id: response?.id || response?.data?.id || lastSaleData.id,
        });
        setIsInvoiceModalOpen(true);
      }
      
      clearCart();
      setBarcodeInput('');
      setPaidAmount('');
      setSelectedCustomer('');
      resetPayment();
    },
    onError: (error) => {
      console.error('Sale failed:', error);
      setProcessing(false);
    },
  });

  // Handlers
  const handleClearSearch = () => {
    setSearchQuery('');
  };

  const handleBarcodeScan = async (e: React.FormEvent) => {
    e.preventDefault();
    if (barcodeInput.trim()) {
      try {
        const response = await barcodeApi.lookupProduct(barcodeInput.trim());
        const product = response as any;
        
        if (product && product.id) {
          addToCart({
            id: product.id,
            name: product.name,
            barcode: barcodeInput.trim(),
            price: product.sellingPrice,
            stock: product.stock,
            condition: product.condition,
            purchaseCost: product.costPrice,
          });
        } else {
          if (soundEnabled) {
            playScanSound(false);
          }
        }
      } catch (error) {
        console.error('Barcode lookup failed:', error);
        if (soundEnabled) {
          playScanSound(false);
        }
        const product = products?.find((p: any) => 
          p.sku === barcodeInput.trim() || 
          p.barcode === barcodeInput.trim()
        );
        if (product) {
          addToCart({
            id: product.id,
            name: product.name,
            barcode: barcodeInput.trim(),
            price: product.sellingPrice,
            stock: product.stock,
            condition: product.condition,
            purchaseCost: product.cost_price,
          });
        }
      }
    }
  };

  const handleCameraScan = (barcode: string) => {
    setBarcodeInput(barcode);
    handleBarcodeScan(new Event('submit') as any);
  };

  const handleManualAdd = () => {
    // Manual add is now handled through the barcode scanner input
  };

  const addTradeInToCart = (inventoryItem: any) => {
    const partType = partTypes.find((pt: any) => pt.id === inventoryItem.part_type_id);
    addToCart({
      id: inventoryItem.id,
      name: inventoryItem.product_name || inventoryItem.product?.name,
      barcode: inventoryItem.serial_number || inventoryItem.id,
      price: inventoryItem.selling_price / 100,
      stock: 1,
      condition: inventoryItem.condition,
      purchaseCost: inventoryItem.purchase_cost / 100,
      isTradeIn: true,
      partType: partType?.name_ar,
      partTypeColor: partType?.color,
      grade: inventoryItem.grade,
    });
  };

  const handleCheckout = useCallback(() => {
    if (cart.length === 0) return;

    setProcessing(true);

    const saleData = {
      customer_id: selectedCustomer || null,
      items: cart.map(item => ({
        product_id: item.id,
        quantity: item.quantity,
        price: item.price,
        is_trade_in: item.isTradeIn || false,
        purchase_cost: item.purchaseCost || 0
      })),
      payment_method: paymentMethod,
      paid_amount: parseFloat(paidAmount) || 0,
      total_amount: total
    };

    const invoiceData: InvoiceData = {
      id: 'pending',
      customerName: selectedCustomer 
        ? customers.find((c: any) => c.id === selectedCustomer)?.name || 'عميل نقدي'
        : 'عميل نقدي',
      customerPhone: selectedCustomer 
        ? customers.find((c: any) => c.id === selectedCustomer)?.phone 
        : undefined,
      saleDate: new Date().toISOString(),
      items: cart.map(item => ({
        name: item.name,
        partType: item.partType,
        partTypeColor: item.partTypeColor,
        condition: item.condition,
        grade: item.grade,
        sellingPrice: item.price,
        quantity: item.quantity,
        total: item.total,
      })),
      subtotal: total,
      total,
      paidAmount: parseFloat(paidAmount) || 0,
      remaining: calculateRemaining(total),
      paymentMethod,
    };

    setLastSaleData(invoiceData);
    createSaleMutation.mutate(saleData);
  }, [cart, selectedCustomer, paymentMethod, paidAmount, total, customers, calculateRemaining, setProcessing, createSaleMutation]);

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow={t('sales.pos')}
        title={t('sales.posTitle')}
        description={t('sales.posTitle')}
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setQuickAddMode(!quickAddMode)}
              className="gap-2"
            >
              <Zap className="w-4 h-4" />
              <span>{t('sales.quickAdd')}</span>
            </Button>

            <Button
              variant="secondary"
              size="sm"
              onClick={() => lastSaleData && setIsInvoiceModalOpen(true)}
              disabled={!lastSaleData}
              className="gap-2"
            >
              <Printer className="w-4 h-4" />
              <span>{t('sales.printInvoice')}</span>
            </Button>
          </div>
        }
      />

      {/* POS Layout */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 lg:gap-4">
        <div className="flex flex-col gap-4">
          <BarcodeScanner
            barcodeInput={barcodeInput}
            setBarcodeInput={setBarcodeInput}
            inputMethod={inputMethod}
            setInputMethod={setInputMethod}
            onBarcodeScan={handleBarcodeScan}
            onCameraScan={handleCameraScan}
            onCameraOpen={() => setIsCameraScannerOpen(true)}
            isCameraScannerOpen={isCameraScannerOpen}
            onCameraClose={() => setIsCameraScannerOpen(false)}
          />

          <CategoryFilter
            categories={categories}
            selectedCategory={selectedCategory}
            onCategorySelect={setSelectedCategory}
          />

          <ProductSearch
            products={products}
            searchQuery={searchQuery}
            setSearchQuery={setSearchQuery}
            onClearSearch={handleClearSearch}
            quickAddMode={quickAddMode}
            onProductClick={addToCart}
            selectedCategory={selectedCategory}
            categories={categories}
          />
        </div>

        <div className="flex flex-col gap-4">
          <CustomerSelector
            selectedCustomer={selectedCustomer}
            setSelectedCustomer={setSelectedCustomer}
            customers={customers}
            customersLoading={customersLoading}
          />

          <TradeInItemsSection
            inventoryItems={inventoryItems}
            partTypes={partTypes}
            onTradeInClick={addTradeInToCart}
          />

          <CartSection
            cart={cart}
            inventoryItems={inventoryItems}
            partTypes={partTypes}
            onUpdateQuantity={updateQuantity}
            onRemoveFromCart={removeFromCart}
          />

          <PaymentSection
            paymentMethod={paymentMethod}
            setPaymentMethod={setPaymentMethod}
            paidAmount={paidAmount}
            setPaidAmount={setPaidAmount}
            total={total}
            isProcessing={isProcessing}
            onCheckout={handleCheckout}
          />
        </div>
      </div>

      {/* Invoice Modal */}
      <Modal
        isOpen={isInvoiceModalOpen}
        onClose={() => setIsInvoiceModalOpen(false)}
        title="فاتورة البيع"
        variant="modern"
        size="xl"
      >
        {lastSaleData && (
          <UsedPartsInvoice
            saleData={lastSaleData}
            onPrint={() => window.print()}
            onClose={() => setIsInvoiceModalOpen(false)}
          />
        )}
      </Modal>
    </div>
  );
}