import { useState, useEffect, useCallback, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useLocation } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { useToast } from '../../../hooks/useToast';
import { Modal } from '../../../components/ui/modal';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { productsApi, salesApi, customersApi, barcodeApi, inventoryApi, partTypesApi, categoriesApi } from '../../../services/api/endpoints';
import { UsedPartsInvoice } from '../../../components/invoice/UsedPartsInvoice';
import { playScanSound } from '../../../hooks/useBarcodeContext';
import { useDebounce } from '../../../hooks/useDebounce';
import { Zap, Printer, Pause, Trash2, Plus } from 'lucide-react';

// Custom hooks
import { normalizePosPrice, useCart } from '../hooks/useCart';
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
import { Product, InventoryItem, PartType } from '../../../types/models';
import type { CustomerCreateRequest } from '../../../services/api/types';
import type { PosCartProduct } from '../hooks/useCart';

interface HeldSale {
  id: string;
  items: PosCartProduct[];
  created_at?: string;
}

export function POSPage() {
  const { t } = useTranslation();
  const location = useLocation();
  const queryClient = useQueryClient();
  const toast = useToast();

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

  useEffect(() => {
    const usedPart = (location.state as { usedPart?: any } | null)?.usedPart;
    if (!usedPart) return;

    addToCart(usedPart);
    window.history.replaceState({}, document.title, window.location.href);
  }, [addToCart, location.state]);

  // Local state
  const [barcodeInput, setBarcodeInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [customerSearchQuery, setCustomerSearchQuery] = useState('');
  const [selectedCustomer, setSelectedCustomer] = useState('');
  const [selectedCustomerOption, setSelectedCustomerOption] = useState<{ id: string; name: string }>();
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [soundEnabled] = useState(true);
  const [quickAddMode, setQuickAddMode] = useState(false);
  const [inputMethod, setInputMethod] = useState<'barcode' | 'camera'>('barcode');
  const [isCameraScannerOpen, setIsCameraScannerOpen] = useState(false);
  const [isInvoiceModalOpen, setIsInvoiceModalOpen] = useState(false);
  const [lastSaleData, setLastSaleData] = useState<InvoiceData | null>(null);
  const lastSaleDataRef = useRef<InvoiceData | null>(null);
  
  // Quick customer creation modal (SALES-PHILOSOPHY.md)
  const [isQuickCustomerModalOpen, setIsQuickCustomerModalOpen] = useState(false);
  const [customerBalance, setCustomerBalance] = useState(0);
  const [customerCreditLimit, setCustomerCreditLimit] = useState<number | undefined>();
  const [quickCustomerName, setQuickCustomerName] = useState('');
  const [quickCustomerPhone, setQuickCustomerPhone] = useState('');
  const [quickCustomerCreditLimit, setQuickCustomerCreditLimit] = useState('');
  const [productPage, setProductPage] = useState(1);
  const { data: heldSalesData } = useQuery({
    queryKey: ['held-sales'],
    queryFn: () => salesApi.listHeld(),
  });
  const heldSales = (((heldSalesData?.data as unknown) as Array<Record<string, unknown>> | undefined) || [])
    .map((held): HeldSale | null => {
      const items = typeof held.items === 'string' ? JSON.parse(held.items) : held.items;
      if (!held.id || !Array.isArray(items)) return null;
      return {
        id: String(held.id),
        items: items as PosCartProduct[],
        created_at: typeof held.created_at === 'string' ? held.created_at : undefined,
      };
    })
    .filter((held): held is HeldSale => held !== null);
  const [isManualProductOpen, setIsManualProductOpen] = useState(false);
  const [manualProduct, setManualProduct] = useState({ name: '', price: '', quantity: '1', barcode: '' });
  const [unknownBarcode, setUnknownBarcode] = useState('');
  const [isHeldSalesOpen, setIsHeldSalesOpen] = useState(false);

  // Debounce search queries for better performance
  const debouncedSearchQuery = useDebounce(searchQuery, 300);
  const debouncedCustomerSearchQuery = useDebounce(customerSearchQuery, 300);

  // Fetch data with debounce search for scalability
  useEffect(() => {
    setProductPage(1);
  }, [debouncedSearchQuery, selectedCategory]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'F4') {
        event.preventDefault();
        handleHoldSale();
      } else if (event.key === 'F2') {
        event.preventDefault();
        document.querySelector<HTMLInputElement>('input[placeholder*="بحث"]')?.focus();
      } else if (event.key === 'F8') {
        event.preventDefault();
        document.querySelector<HTMLButtonElement>('.pos-checkout-button:not(:disabled)')?.click();
      } else if (event.key === 'Escape') {
        setUnknownBarcode('');
        setIsManualProductOpen(false);
      } else if (event.key === 'Delete' && cart.length > 0 &&
        !(event.target as HTMLElement).matches('input, textarea, select')) {
        removeFromCart(cart[cart.length - 1].barcode);
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [cart, removeFromCart]);

  const { data: productsData } = useQuery({
    queryKey: ['products', debouncedSearchQuery, selectedCategory, productPage],
    queryFn: () => {
      return productsApi.list({
        page: productPage,
        per_page: 16,
        search: debouncedSearchQuery,
        category_id: selectedCategory || undefined,
      });
    },
    enabled: true, // Always enabled, but will refetch when search changes
  });

  const { data: customersData, isLoading: customersLoading } = useQuery({
    queryKey: ['customers', debouncedCustomerSearchQuery],
    queryFn: () => {
      if (debouncedCustomerSearchQuery) {
        // Search mode - use API search when query exists
        return customersApi.list({
          page: 1,
          per_page: 50,
          search: debouncedCustomerSearchQuery
        });
      } else {
        // Initial load - fetch limited results for performance
        return customersApi.list({ page: 1, per_page: 50 });
      }
    },
    enabled: true,
  });

  const { data: inventoryData } = useQuery({
    queryKey: ['inventory'],
    queryFn: () => inventoryApi.list({ page: 1, per_page: 50 }),
  });

  const { data: partTypesData } = useQuery({
    queryKey: ['part-types'],
    queryFn: () => partTypesApi.list(),
  });

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });

  const products = (productsData?.data?.products as unknown) as Product[] || [];
  const productTotal = productsData?.data?.total ?? products.length;
  const customers = (customersData?.data as unknown) as Customer[] || [];
  const inventoryItems = (inventoryData?.data?.items as unknown) as InventoryItem[] || [];
  const partTypes = (partTypesData?.data as unknown) as PartType[] || [];
  const categories = (categoriesData?.data as unknown) as Category[] || [];

  const getAvailableStockCount = useCallback((productId: string) => {
    if (!productId) return 0;

    return inventoryItems.filter((item: any) => {
      if (String(item.product_id) !== String(productId)) {
        return false;
      }

      const status = String(item.status || '').toUpperCase();
      return !['SOLD', 'RESERVED', 'DAMAGED', 'IN_REPAIR', 'RETURNED', 'FOR_PARTS', 'ARCHIVED'].includes(status);
    }).length;
  }, [inventoryItems]);

  const canAddProductToCart = useCallback((productId: string, productName: string, quantity: number = 1) => {
    const currentQuantity = cart.filter((item) => String(item.id) === String(productId)).reduce((sum, item) => sum + item.quantity, 0);
    const available = getAvailableStockCount(productId);

    if (available <= 0 || currentQuantity + quantity > available) {
      const label = productName || 'هذا النوع';
      toast.error(`هذا النوع قد نفذ من المخزون: ${label}`, 'مخزون غير كافٍ', 4000);
      return false;
    }

    return true;
  }, [cart, getAvailableStockCount, toast]);

  const createCustomerMutation = useMutation({
    mutationFn: (data: CustomerCreateRequest) => customersApi.create(data),
    onSuccess: (response) => {
      const payload = response?.data as any;
      const createdCustomer = payload?.customer ?? payload;
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      if (createdCustomer?.id) {
        setSelectedCustomer(String(createdCustomer.id));
        setSelectedCustomerOption({ id: String(createdCustomer.id), name: createdCustomer.name });
        setCustomerBalance(0);
        setCustomerCreditLimit(createdCustomer.credit_limit);
      }
      setIsQuickCustomerModalOpen(false);
      setQuickCustomerName('');
      setQuickCustomerPhone('');
      setQuickCustomerCreditLimit('');
    },
    onError: (error) => {
      console.error('Customer creation failed:', error);
      toast.error('تعذر إنشاء العميل. تحقق من البيانات وحاول مرة أخرى.', 'فشل إنشاء العميل', 4000);
    },
  });

  const holdSaleMutation = useMutation({
    mutationFn: () => salesApi.hold(cart),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['held-sales'] });
      clearCart();
      resetPayment();
      toast.success('تم تعليق البيع ويمكن استكماله لاحقًا', 'تم التعليق');
    },
  });

  const deleteHeldSaleMutation = useMutation({
    mutationFn: (id: string) => salesApi.deleteHeld(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['held-sales'] }),
  });

  const createProductMutation = useMutation({
    mutationFn: () => productsApi.create({
      name: manualProduct.name.trim(),
      selling_price: Number(manualProduct.price) || 0,
      barcode: manualProduct.barcode.trim() || undefined,
      condition: 'new',
      stock: Number(manualProduct.quantity) || 1,
    }),
    onSuccess: (response) => {
      const product = (response?.data as any)?.product ?? response?.data;
      if (product?.id) {
        const barcode = product.barcode || product.id;
        addToCart({ ...product, barcode }, Number(manualProduct.quantity) || 1);
      }
      queryClient.invalidateQueries({ queryKey: ['products'] });
      setManualProduct({ name: '', price: '', quantity: '1', barcode: '' });
      setIsManualProductOpen(false);
    },
  });

  // Create sale mutation
  const createSaleMutation = useMutation({
    mutationFn: (data: SaleRequest) => salesApi.create(data),
    onSuccess: (response: SaleResponse) => {
      queryClient.invalidateQueries({ queryKey: ['sales'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      
      if (lastSaleDataRef.current) {
        const sale = response?.data?.sale ?? response?.data ?? response?.sale;
        setLastSaleData({
          ...lastSaleDataRef.current,
          id: sale?.id || response?.id || lastSaleDataRef.current.id,
        });
        setIsInvoiceModalOpen(true);
      }
      
      clearCart();
      setBarcodeInput('');
      setPaidAmount('');
      setSelectedCustomer('');
      resetPayment();
      setProcessing(false);
    },
    onError: (error: any) => {
      const message = error?.response?.error?.message || error?.arabicMessage || error?.message || 'فشل إتمام البيع';
      const normalizedMessage = String(message).toLowerCase();

      if (normalizedMessage.includes('insufficient stock') || normalizedMessage.includes('نفذ') || normalizedMessage.includes('مخزون')) {
        toast.error('هذا النوع قد نفذ من المخزون', 'مخزون غير كافٍ', 4000);
      } else {
        toast.error(message, 'فشل إتمام البيع', 4000);
      }

      console.error('Sale failed:', error);
      setProcessing(false);
    },
  });

  // Handlers
  const handleClearSearch = () => {
    setSearchQuery('');
  };

  const handleHoldSale = () => {
    if (cart.length === 0) {
      toast.error('أضف منتجًا إلى السلة قبل تعليق البيع', 'السلة فارغة');
      return;
    }
    holdSaleMutation.mutate();
  };

  const handleResumeSale = (held: HeldSale) => {
    if (!held) return;
    clearCart();
    held.items.forEach((item) => addToCart(item, item.quantity));
    deleteHeldSaleMutation.mutate(held.id);
    setIsHeldSalesOpen(false);
  };

  const handleBarcodeScan = async (e: React.FormEvent) => {
    e.preventDefault();
    if (barcodeInput.trim()) {
      try {
        const response = await barcodeApi.lookupProduct(barcodeInput.trim());
        const product = response as Product;
        
        if (product && product.id) {
          if (!canAddProductToCart(product.id, product.name, 1)) {
            return;
          }

          addToCart({
            id: product.id,
            name: product.name,
            barcode: barcodeInput.trim(),
            price: normalizePosPrice(product.sellingPrice, (product as Product & { selling_price?: number }).selling_price),
            stock: product.stock,
            condition: product.condition,
            purchaseCost: product.costPrice || product.cost_price || 0,
          });
        } else {
          if (soundEnabled) {
            playScanSound(false);
          }
          setUnknownBarcode(barcodeInput.trim());
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
          if (!canAddProductToCart(product.id, product.name, 1)) {
            return;
          }

          addToCart({
            id: product.id,
            name: product.name,
            barcode: barcodeInput.trim(),
            price: normalizePosPrice(product.sellingPrice, (product as Product & { selling_price?: number }).selling_price),
            stock: product.stock,
            condition: product.condition,
            purchaseCost: product.cost_price || product.costPrice || 0,
          });
        } else {
          setUnknownBarcode(barcodeInput.trim());
        }
      }
    }
  };

  const handleProductSelect = useCallback((product: any) => {
    if (!canAddProductToCart(product.id, product.name, 1)) {
      return;
    }
    addToCart(product);
  }, [addToCart, canAddProductToCart]);

  const handleCameraScan = (barcode: string) => {
    setBarcodeInput(barcode);
    // Create a proper event object for the barcode scan
    const event = new Event('submit', { bubbles: true, cancelable: true }) as unknown as React.FormEvent;
    handleBarcodeScan(event);
  };

  const addTradeInToCart = (inventoryItem: InventoryItem) => {
    const partType = partTypes.find((pt: PartType) => pt.id === inventoryItem.part_type_id);
    addToCart({
      id: inventoryItem.id,
      name: inventoryItem.product_name || inventoryItem.product?.name,
      barcode: inventoryItem.serial_number || inventoryItem.id,
      price: normalizePosPrice(inventoryItem.selling_price),
      stock: 1,
      condition: inventoryItem.condition,
      purchaseCost: inventoryItem.purchase_cost,
      isTradeIn: true,
      partType: partType?.name_ar,
      partTypeColor: partType?.color,
      grade: inventoryItem.grade,
    });
  };

  const handleCheckout = useCallback(() => {
    if (cart.length === 0) return;
    if (paymentMethod === 'credit' && paidAmount.trim() === '') return;

    const exhaustedItems = cart.reduce((items, item) => {
      const availableStock = item.isTradeIn ? Number(item.stock ?? 1) : getAvailableStockCount(String(item.id));
      if (availableStock <= 0 || item.quantity > availableStock) {
        items.push(item.name);
      }
      return items;
    }, [] as string[]);

    if (exhaustedItems.length > 0) {
      const uniqueItems = [...new Set(exhaustedItems)];
      const message = uniqueItems.length > 1
        ? `هذه الأنواع قد نفذت من المخزون: ${uniqueItems.slice(0, 2).join(', ')}`
        : `هذا النوع قد نفذ من المخزون: ${uniqueItems[0]}`;
      toast.error(message, 'مخزون غير كافٍ', 4000);
      return;
    }

    setProcessing(true);

    const saleData = {
      customer_id: selectedCustomer || null,
      items: cart.map(item => ({
        product_id: item.id,
        quantity: item.quantity,
        unit_price: item.price
      })),
      payment_method: paymentMethod,
      payment_amount: parseFloat(paidAmount) || 0
    };

    const invoiceData: InvoiceData = {
      id: 'pending',
      customerName: selectedCustomer 
        ? (customers.find((c: any) => String(c.id) === String(selectedCustomer))?.name
          ?? selectedCustomerOption?.name)
        : '',
      customerPhone: selectedCustomer 
        ? customers.find((c: any) => String(c.id) === String(selectedCustomer))?.phone
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
    lastSaleDataRef.current = invoiceData;
    createSaleMutation.mutate(saleData);
  }, [cart, selectedCustomer, paymentMethod, paidAmount, total, customers, calculateRemaining, setProcessing, createSaleMutation, getAvailableStockCount, toast]);

  // Quick customer creation handler (SALES-PHILOSOPHY.md)
  const handleQuickCustomerCreate = () => {
    setIsQuickCustomerModalOpen(true);
  };

  // Handle customer selection to load balance info (SALES-PHILOSOPHY.md)
  const handleCustomerChange = (customerId: string) => {
    setSelectedCustomer(customerId);
    if (customerId) {
      const customer = customers.find((c: any) => String(c.id) === String(customerId));
      if (customer) {
        setSelectedCustomerOption({ id: String(customer.id), name: customer.name });
        setCustomerBalance(customer.balance || 0);
        setCustomerCreditLimit(customer.credit_limit);
      }
    } else {
      setSelectedCustomerOption(undefined);
      setCustomerBalance(0);
      setCustomerCreditLimit(undefined);
    }
  };

  return (
    <div>
      <header className="pos-cashier-header">
        <div>
          <span className="pos-cashier-brand">PARTFLOW POS</span>
          <h1>{t('sales.posTitle')}</h1>
        </div>
        <div className="pos-cashier-meta">
          <span>Cashier 01</span>
          <time>{new Intl.DateTimeFormat('ar', { hour: '2-digit', minute: '2-digit' }).format(new Date())}</time>
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
      </header>

      {/* Cashier layout: scanner first, products beside the cart on desktop. */}
      <div className="pos-cashier-shell">
        <div className="pos-cashier-scanner">
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
        </div>
        <div className="pos-cashier-grid">
          <div className="pos-cashier-products">
            <CategoryFilter
            categories={categories}
            hasMore={productPage * 16 < productTotal}
            onLoadMore={() => setProductPage((page) => page + 1)}
            selectedCategory={selectedCategory}
            onCategorySelect={setSelectedCategory}
            />
            <ProductSearch
            products={products}
            searchQuery={searchQuery}
            setSearchQuery={setSearchQuery}
            onClearSearch={handleClearSearch}
            quickAddMode={quickAddMode}
            onProductClick={handleProductSelect}
            selectedCategory={selectedCategory}
            categories={categories}
            hasMore={productPage * 16 < productTotal}
            onLoadMore={() => setProductPage((page) => page + 1)}
            />
          </div>
          <div className="pos-cashier-cart flex flex-col gap-4">
          <CustomerSelector
            selectedCustomer={selectedCustomer}
            setSelectedCustomer={handleCustomerChange}
            customers={customers}
            selectedCustomerOption={selectedCustomerOption}
            customersLoading={customersLoading}
            customerSearchQuery={customerSearchQuery}
            setCustomerSearchQuery={setCustomerSearchQuery}
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
            selectedCustomer={selectedCustomer}
            customerBalance={customerBalance}
            customerCreditLimit={customerCreditLimit}
            onQuickCustomerCreate={handleQuickCustomerCreate}
          />
          <TradeInItemsSection
            inventoryItems={inventoryItems}
            partTypes={partTypes}
            onTradeInClick={addTradeInToCart}
          />
          </div>
        </div>
        <div className="pos-cashier-footer">
          <Button variant="secondary" size="sm" onClick={() => setIsManualProductOpen(true)}>
            <Plus className="h-4 w-4" /> إضافة يدويًا
          </Button>
          <Button variant="secondary" size="sm" onClick={handleHoldSale}>
            <Pause className="h-4 w-4" /> تعليق البيع
          </Button>
          <Button variant="ghost" size="sm" onClick={() => { clearCart(); resetPayment(); }}>
            <Trash2 className="h-4 w-4" /> مسح السلة
          </Button>
          {heldSales.length > 0 && (
            <Button variant="ghost" size="sm" onClick={() => setIsHeldSalesOpen(true)}>
              استكمال بيع معلّق ({heldSales.length})
            </Button>
          )}
        </div>

        <Modal
          isOpen={isHeldSalesOpen}
          onClose={() => setIsHeldSalesOpen(false)}
          title="المبيعات المعلّقة"
          variant="modern"
          size="md"
        >
          <div className="flex flex-col gap-3">
            {heldSales.map((held, index) => (
              <Button key={held.id} variant="secondary" className="justify-between" onClick={() => handleResumeSale(held)}>
                <span>بيع معلّق #{index + 1}</span>
                <span>{held.items.length} منتجات</span>
              </Button>
            ))}
          </div>
        </Modal>

        <Modal
          isOpen={isManualProductOpen}
          onClose={() => setIsManualProductOpen(false)}
          title="إضافة منتج يدويًا"
          variant="modern"
          size="md"
        >
          <div className="flex flex-col gap-3">
            <Input autoFocus placeholder="اسم المنتج *" value={manualProduct.name}
              onChange={(e) => setManualProduct((current) => ({ ...current, name: e.target.value }))} />
            <Input type="number" placeholder="سعر البيع" value={manualProduct.price}
              onChange={(e) => setManualProduct((current) => ({ ...current, price: e.target.value }))} />
            <Input placeholder="الباركود (اختياري)" value={manualProduct.barcode}
              onChange={(e) => setManualProduct((current) => ({ ...current, barcode: e.target.value }))} />
          <Input type="number" min={1} placeholder="الكمية" value={manualProduct.quantity}
            onChange={(e) => setManualProduct((current) => ({ ...current, quantity: e.target.value }))} />
            {createProductMutation.isError && <p className="text-sm text-red-500">تعذر إنشاء المنتج، تحقق من البيانات.</p>}
            <Button variant="primary" disabled={!manualProduct.name.trim() || createProductMutation.isPending}
              onClick={() => createProductMutation.mutate()}>
              {createProductMutation.isPending ? 'جارٍ الحفظ...' : 'حفظ وإضافة للسلة'}
            </Button>
          </div>
        </Modal>

        <Modal
          isOpen={Boolean(unknownBarcode)}
          onClose={() => setUnknownBarcode('')}
          title="المنتج غير موجود"
          variant="modern"
          size="sm"
        >
          <div className="flex flex-col gap-3">
            <p className="text-sm text-text-secondary">الباركود: <strong>{unknownBarcode}</strong></p>
            <Button variant="secondary" onClick={() => {
              setSearchQuery(unknownBarcode);
              setUnknownBarcode('');
            }}>البحث يدويًا</Button>
            <Button variant="primary" onClick={() => {
              setManualProduct((current) => ({ ...current, barcode: unknownBarcode }));
              setUnknownBarcode('');
              setIsManualProductOpen(true);
            }}>إضافة منتج جديد</Button>
            <Button variant="ghost" onClick={() => setUnknownBarcode('')}>إلغاء</Button>
          </div>
        </Modal>
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

      {/* Quick Customer Creation Modal (SALES-PHILOSOPHY.md) */}
      <Modal
        isOpen={isQuickCustomerModalOpen}
        onClose={() => setIsQuickCustomerModalOpen(false)}
        title="إضافة عميل سريع"
        variant="modern"
        size="md"
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <div>
            <label className="text-small font-medium text-text mb-sm block">
              الاسم *
            </label>
            <Input
              placeholder="أدخل اسم العميل..."
              autoFocus
              value={quickCustomerName}
              onChange={(event) => setQuickCustomerName(event.target.value)}
            />
          </div>
          <div>
            <label className="text-small font-medium text-text mb-sm block">
              رقم الهاتف
            </label>
            <Input
              type="tel"
              placeholder="05xxxxxxxx"
              value={quickCustomerPhone}
              onChange={(event) => setQuickCustomerPhone(event.target.value)}
            />
          </div>
          <div>
            <label className="text-small font-medium text-text mb-sm block">
              حد الدين
            </label>
            <Input
              type="number"
              placeholder="₪0"
              value={quickCustomerCreditLimit}
              onChange={(event) => setQuickCustomerCreditLimit(event.target.value)}
            />
          </div>
          {createCustomerMutation.isError && (
            <p className="text-sm text-red-500">
              تعذر إنشاء العميل. تحقق من البيانات وحاول مرة أخرى.
            </p>
          )}
          <div className="flex gap-sm justify-end mt-4">
            <Button
              variant="secondary"
              onClick={() => setIsQuickCustomerModalOpen(false)}
            >
              إلغاء
            </Button>
            <Button
              variant="primary"
              disabled={createCustomerMutation.isPending || !quickCustomerName.trim()}
              onClick={() => createCustomerMutation.mutate({
                name: quickCustomerName.trim(),
                phone: quickCustomerPhone.trim() || undefined,
                credit_limit: Number(quickCustomerCreditLimit) || 0,
              })}
            >
              {createCustomerMutation.isPending ? 'جارٍ الحفظ...' : 'حفظ ومتابعة البيع'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}