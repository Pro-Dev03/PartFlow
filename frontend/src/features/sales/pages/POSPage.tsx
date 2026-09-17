import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useLocation } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { useToast } from '../../../hooks/useToast';
import { useUIStore } from '../../../stores/uiStore';
import { Modal } from '../../../components/ui/modal';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PartFlowLogo } from '../../../components/branding/PartFlowLogo';
import {
  productsApi,
  salesApi,
  customersApi,
  barcodeApi,
  inventoryApi,
  partTypesApi,
  categoriesApi,
  settingsApi,
  paymentTransactionsApi,
} from '../../../services/api/endpoints';
import { UsedPartsInvoice } from '../../../components/invoice/UsedPartsInvoice';
import { playScanSound } from '../../../hooks/useBarcodeContext';
import { useDebounce } from '../../../hooks/useDebounce';
import {
  Search,
  Printer,
  Pause,
  Plus,
  Package,
  ScanLine,
  Zap,
  X,
  ArrowRight,
  Wifi,
  ShoppingCart,
  Trash2,
  UserRound,
  ChevronDown,
  Info,
} from 'lucide-react';

// Modern Components
import { ModernProductGrid } from '../components/modern/ModernProductGrid';
import { ModernCartPanel } from '../components/modern/ModernCartPanel';
import { AdvancedPaymentPanel } from '../components/modern/AdvancedPaymentPanel';

// Hooks
import { normalizePosPrice, useCart } from '../hooks/useCart';
import { usePayment } from '../hooks/usePayment';
import { buildManualProductPayload } from '../utils/manualProductPayload';
import { getPartTypeImage } from '../../../services/localPartTypeImages';
import { getCategoryImage } from '../../../services/localCategoryImages';

// Types
import { InvoiceData, PaymentAllocation } from '../types/pos.types';
import { Product, InventoryItem, PartType } from '../../../types/models';
import type { CustomerCreateRequest, Customer, Category } from '../../../services/api/types';
import type { PosCartProduct } from '../hooks/useCart';
import type { SaleCreateRequest } from '../../../services/api/types';

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
  const { data: taxSetting } = useQuery({
    queryKey: ['settings', 'tax_rate'],
    queryFn: () => settingsApi.getSetting('tax_rate'),
    retry: false,
  });
  const { data: discountSetting } = useQuery({
    queryKey: ['settings', 'max_discount_rate'],
    queryFn: () => settingsApi.getSetting('max_discount_rate'),
    retry: false,
  });
  const { data: discountsEnabledSetting } = useQuery({
    queryKey: ['settings', 'discounts_enabled'],
    queryFn: () => settingsApi.getSetting('discounts_enabled'),
    retry: false,
  });
  const { data: posProductsPerPageSetting } = useQuery({
    queryKey: ['settings', 'pos_products_per_page'],
    queryFn: () => settingsApi.getSetting('pos_products_per_page'),
    retry: false,
  });
  const { data: posProductViewModeSetting } = useQuery({
    queryKey: ['settings', 'pos_product_view_mode'],
    queryFn: () => settingsApi.getSetting('pos_product_view_mode'),
    retry: false,
  });
  const { data: electronicPaymentsEnabledSetting } = useQuery({
    queryKey: ['settings', 'electronic_payments_enabled'],
    queryFn: () => settingsApi.getSetting('electronic_payments_enabled'),
    retry: false,
  });
  const { data: electronicPaymentMethodsSetting } = useQuery({
    queryKey: ['settings', 'payment_methods'],
    queryFn: () => settingsApi.getSetting('payment_methods'),
    retry: false,
  });
  const { data: electronicPaymentProviderSetting } = useQuery({
    queryKey: ['settings', 'payment_provider'],
    queryFn: () => settingsApi.getSetting('payment_provider'),
    retry: false,
  });
  const systemTaxRate = Number(taxSetting?.data?.value);
  const configuredMaxDiscount = Number(discountSetting?.data?.value);
  const maxDiscountRate = Number.isFinite(configuredMaxDiscount) && configuredMaxDiscount >= 0 && configuredMaxDiscount <= 100
    ? configuredMaxDiscount
    : 15;
  const discountsEnabledValue = String(discountsEnabledSetting?.data?.value ?? '').trim().toLowerCase();
  const discountsEnabled = !['false', '0', 'off', 'disabled'].includes(discountsEnabledValue);
  const electronicPaymentsEnabled = electronicPaymentsEnabledSetting?.data?.value === 'true';
  const electronicPaymentProvider = String(electronicPaymentProviderSetting?.data?.value ?? 'manual');
  let enabledElectronicMethods: string[] = [];
  try {
    const configuredMethods = JSON.parse(String(electronicPaymentMethodsSetting?.data?.value ?? '[]'));
    if (Array.isArray(configuredMethods)) enabledElectronicMethods = configuredMethods.filter((method): method is string => typeof method === 'string');
  } catch {
    enabledElectronicMethods = [];
  }

  const [taxExempt, setTaxExempt] = useState(false);
  const [discountRate, setDiscountRate] = useState(0);

  // Custom hooks
  const {
    cart,
    addToCart,
    removeFromCart,
    updateQuantity,
    clearCart,
    total,
    subtotal,
  } = useCart(true, taxExempt ? 0 : (Number.isFinite(systemTaxRate) ? systemTaxRate : 0));
  const effectiveTaxRate = Number.isFinite(systemTaxRate) && systemTaxRate > 0 ? systemTaxRate : 0;
  const priceWithTax = (price: number) => Math.round(price * (1 + effectiveTaxRate / 100) * 100) / 100;
  const displayCart = cart.map((item) => ({
    ...item,
    price: taxExempt ? item.price : priceWithTax(item.price),
    total: (taxExempt ? item.price : priceWithTax(item.price)) * item.quantity,
  }));
  const displaySubtotal = displayCart.reduce((sum, item) => sum + item.total, 0);
  const appliedDiscountRate = discountsEnabled
    ? Math.min(Math.max(Number(discountRate) || 0, 0), maxDiscountRate)
    : 0;
  const displayDiscount = Math.round(displaySubtotal * appliedDiscountRate) / 100;
  const displayTotal = Math.max(0, Math.round((displaySubtotal - displayDiscount) * 100) / 100);
  const {
    paymentMethod,
    setPaymentMethod,
    paidAmount,
    setPaidAmount,
    isProcessing,
    setProcessing,
    resetPayment,
  } = usePayment();

  // Handle used part from navigation
  const processedUsedPartNavigation = useRef<string | null>(null);
  useEffect(() => {
    const usedPart = (location.state as { usedPart?: PosCartProduct } | null)
      ?.usedPart;
    if (!usedPart) return;

    const navigationKey = `${location.key}:${usedPart.inventoryItemId ?? usedPart.id}`;
    if (processedUsedPartNavigation.current === navigationKey) return;
    processedUsedPartNavigation.current = navigationKey;

    addToCart(usedPart);
    window.history.replaceState({}, document.title, window.location.href);
  }, [addToCart, location.key, location.state]);

  // Local state
  const [searchQuery, setSearchQuery] = useState('');
  const [posSection, setPosSection] = useState<'products' | 'used'>('products');
  const [soldUsedPartIds, setSoldUsedPartIds] = useState<Set<string>>(() => new Set());
  const [optimisticSoldQuantities, setOptimisticSoldQuantities] = useState<Record<string, number>>({});
  const [customerSearchQuery, setCustomerSearchQuery] = useState('');
  const [selectedCustomer, setSelectedCustomer] = useState('');
  const [selectedCustomerOption, setSelectedCustomerOption] = useState<{
    id: string;
    name: string;
  }>();
  const [isCustomerMenuOpen, setIsCustomerMenuOpen] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [soundEnabled] = useState(true);
  const [isInvoiceModalOpen, setIsInvoiceModalOpen] = useState(false);
  const [lastSaleData, setLastSaleData] = useState<InvoiceData | null>(null);
  const lastSaleDataRef = useRef<InvoiceData | null>(null);
  const pendingExternalPaymentRef = useRef<string | null>(null);
  const [isQuickCustomerModalOpen, setIsQuickCustomerModalOpen] =
    useState(false);
  const [customerBalance, setCustomerBalance] = useState(0);
  const [customerCreditLimit, setCustomerCreditLimit] = useState<
    number | undefined
  >();
  const [paymentAllocations, setPaymentAllocations] = useState<PaymentAllocation[]>([]);
  const [quickCustomerName, setQuickCustomerName] = useState('');
  const [quickCustomerPhone, setQuickCustomerPhone] = useState('');
  const [productPage, setProductPage] = useState(1);
  const [usedPartPage, setUsedPartPage] = useState(1);
  const configuredProductsPerPage = Number(posProductsPerPageSetting?.data?.value);
  const productsPerPage = Number.isInteger(configuredProductsPerPage) && configuredProductsPerPage >= 4 && configuredProductsPerPage <= 48
    ? configuredProductsPerPage
    : 12;
  const productViewMode = posProductViewModeSetting?.data?.value === 'list' ? 'list' : 'cards';
  const [showProductDetails, setShowProductDetails] = useState(false);
  const [isManualProductOpen, setIsManualProductOpen] = useState(false);
  const [manualProduct, setManualProduct] = useState({
    name: '',
    price: '',
    quantity: '1',
    barcode: '',
  });
  const [unknownBarcode, setUnknownBarcode] = useState('');
  const [isHeldSalesOpen, setIsHeldSalesOpen] = useState(false);
  const { checkoutMode, setCheckoutMode } = useUIStore();

  // Held sales query
  const { data: heldSalesData } = useQuery({
    queryKey: ['held-sales'],
    queryFn: () => salesApi.listHeld(),
  });
  const heldSales = (
    (
      (heldSalesData?.data as unknown) as Array<Record<string, unknown>> | undefined
    ) || []
  )
    .map((held): HeldSale | null => {
      const items =
        typeof held.items === 'string' ? JSON.parse(held.items) : held.items;
      if (!held.id || !Array.isArray(items)) return null;
      return {
        id: String(held.id),
        items: items as PosCartProduct[],
        created_at:
          typeof held.created_at === 'string' ? held.created_at : undefined,
      };
    })
    .filter((held): held is HeldSale => held !== null);

  // Debounced search
  const debouncedSearchQuery = useDebounce(searchQuery, 300);
  const debouncedCustomerSearchQuery = useDebounce(customerSearchQuery, 300);

  useEffect(() => {
    setProductPage(1);
    setUsedPartPage(1);
  }, [debouncedSearchQuery, selectedCategory, productsPerPage]);

  // Keyboard shortcuts
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'F4') {
        event.preventDefault();
        handleHoldSale();
      } else if (event.key === 'F2') {
        event.preventDefault();
        document
          .querySelector<HTMLInputElement>('.pos-barcode-input')
          ?.focus();
      } else if (event.key === 'F8') {
        event.preventDefault();
        document
          .querySelector<HTMLButtonElement>('.checkout-btn:not(:disabled)')
          ?.click();
      } else if (event.key === 'Escape') {
        setUnknownBarcode('');
        setIsManualProductOpen(false);
      } else if (event.key === 'Delete' && cart.length > 0) {
        if (
          !(event.target as HTMLElement).matches('input, textarea, select')
        ) {
          removeFromCart(cart[cart.length - 1].barcode);
        }
      }
    };
  }, [cart, removeFromCart]);

  // Products query
  const { data: productsData, isLoading: productsLoading } = useQuery({
    queryKey: [
      'products',
      debouncedSearchQuery,
      selectedCategory,
      productPage,
    ],
    queryFn: () => {
      return productsApi.list({
        page: productPage,
        per_page: productsPerPage,
        search: debouncedSearchQuery,
        category_id: selectedCategory || undefined,
      });
    },
    enabled: posSection === 'products',
    staleTime: 0,
    refetchInterval: 5000,
    refetchOnWindowFocus: true,
  });

  // Customers query
  const { data: customersData, isLoading: customersLoading } = useQuery({
    queryKey: ['customers', debouncedCustomerSearchQuery],
    queryFn: () => {
      if (debouncedCustomerSearchQuery) {
        return customersApi.list({
          page: 1,
          per_page: 50,
          search: debouncedCustomerSearchQuery,
        });
      } else {
        return customersApi.list({ page: 1, per_page: 50 });
      }
    },
    enabled: true,
  });

  // Other queries
  const { data: inventoryData } = useQuery({
    queryKey: ['inventory'],
    queryFn: () => inventoryApi.listWithSupplier({ page: 1, per_page: 1000 }),
    staleTime: 0,
    refetchInterval: 5000,
    refetchOnWindowFocus: true,
  });

  useEffect(() => {
    setOptimisticSoldQuantities({});
  }, [inventoryData]);

  const { data: partTypesData } = useQuery({
    queryKey: ['part-types'],
    queryFn: () => partTypesApi.list(),
  });

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });

  const products =
    ((productsData?.data?.products as unknown) as Product[]) || [];
  const productTotal = productsData?.data?.total ?? products.length;
  const customers =
    ((customersData?.data as unknown) as Customer[]) || [];
  const inventoryItems =
    ((inventoryData?.data?.items as unknown) as InventoryItem[]) || [];
  const partTypes = ((partTypesData?.data as unknown) as PartType[]) || [];
  const categories = ((categoriesData?.data as unknown) as Category[]) || [];

  const productsWithInventoryFallback = useMemo(() => {
    const query = debouncedSearchQuery.trim().toLowerCase();
    const productsById = new Map(products.map((product) => [String(product.id), product]));

    inventoryItems.forEach((item: any) => {
      const productId = String(item.product_id || item.product?.id || item.productId || '').trim();
      if (!productId || productsById.has(productId)) return;

      const name = String(item.product_name || item.product?.name || '').trim();
      const itemCode = String(item.item_code || '').trim();
      const barcode = String(item.barcode || '').trim();
      const categoryId = String(item.category_id || '').trim();
      const searchableText = `${name} ${itemCode} ${barcode}`.toLowerCase();
      if (!name || (query && !searchableText.includes(query))) return;
      if (selectedCategory && categoryId !== selectedCategory) return;

      productsById.set(productId, {
        id: productId,
        name,
        sku: itemCode || barcode || productId,
        barcode: barcode || itemCode,
        category_id: categoryId || undefined,
        track_serial: false,
        track_individual: false,
        min_stock_level: 0,
        warranty_days: 0,
        created_at: String(item.created_at || ''),
        updated_at: String(item.updated_at || ''),
        selling_price: Number(item.product_selling_price ?? item.selling_price ?? item.price ?? 0),
        cost_price: Number(item.purchase_cost ?? 0),
      } as Product);
    });

    return Array.from(productsById.values());
  }, [debouncedSearchQuery, inventoryItems, products, selectedCategory]);

  const availableUsedParts = useMemo(() => {
    const query = debouncedSearchQuery.trim().toLowerCase();
    return inventoryItems.filter((item: any) => {
      if (String(item.condition || '').toUpperCase() !== 'USED') return false;
      if (String(item.status || '').toUpperCase() !== 'AVAILABLE') return false;
      if (soldUsedPartIds.has(String(item.id))) return false;
      if (!query) return true;
      const name = String(item.product_name || item.product?.name || '').toLowerCase();
      const serialNumber = String(item.serial_number || '').toLowerCase();
      const barcode = String(item.barcode || '').toLowerCase();
      return name.includes(query) || serialNumber.includes(query) || barcode.includes(query);
    });
  }, [debouncedSearchQuery, inventoryItems, soldUsedPartIds]);

  const usedPartTotalPages = Math.max(1, Math.ceil(availableUsedParts.length / productsPerPage));
  const visibleUsedParts = availableUsedParts.slice(
    (usedPartPage - 1) * productsPerPage,
    usedPartPage * productsPerPage,
  );

  useEffect(() => {
    if (usedPartPage > usedPartTotalPages) setUsedPartPage(usedPartTotalPages);
  }, [usedPartPage, usedPartTotalPages]);

  const handleUsedPartSelect = useCallback((item: any) => {
    const productId = String(item.product_id || item.product?.id || '');
    if (!productId) {
      toast.error('هذه القطعة غير مرتبطة بمنتج صالح', 'تعذر إضافة القطعة', 4000);
      return;
    }
    addToCart({
      id: productId,
      inventoryItemId: String(item.id),
      name: item.product_name || item.product?.name || 'قطعة مستعملة',
      barcode: item.barcode || item.serial_number || item.id,
      price: normalizePosPrice(item.selling_price, item.price),
      stock: 1,
      isTradeIn: true,
      purchaseCost: normalizePosPrice(item.purchase_cost),
      condition: item.condition,
      grade: item.grade,
      serialNumber: item.serial_number,
    });
  }, [addToCart, toast]);

  const productsWithStock = useMemo(() => {
    const stockMap = new Map<string, number>();
    const newStockProducts = new Set<string>();

    inventoryItems.forEach((item: any) => {
      const productId = String(item.product_id || item.product?.id || '').trim();
      if (!productId) return;

      const status = String(item.status || '').trim().toUpperCase();
      const condition = String(item.condition || '').trim().toUpperCase();
      if (['SOLD', 'RESERVED', 'DAMAGED', 'IN_REPAIR', 'RETURNED', 'FOR_PARTS', 'ARCHIVED'].includes(status)) {
        return;
      }

      if (condition === 'USED') return;
      newStockProducts.add(productId);

      const explicitStock = Number(
        item.available_quantity ??
        item.current_quantity ??
        item.stock ??
        item.quantity ??
        0
      );

      const calculatedStock = Number.isFinite(explicitStock) && explicitStock > 0 ? explicitStock : (status === 'AVAILABLE' ? 1 : 0);
      if (calculatedStock <= 0) return;

      const current = stockMap.get(productId) ?? 0;
      stockMap.set(productId, Math.max(current, calculatedStock));
    });

    return productsWithInventoryFallback.filter((product) => {
      const productId = String(product.id);
      const declaredStock = Number((product as any).stock ?? 0);
      return newStockProducts.has(productId) || (!inventoryItems.some((item: any) => String(item.product_id || item.product?.id || '') === productId) && declaredStock > 0);
    }).map((product) => {
      const rawProduct = product as Product & {
        categoryId?: string;
        category?: { id?: string };
      };
      const categoryId = rawProduct.category_id || rawProduct.categoryId || rawProduct.category?.id;
      const baseStock = Number(stockMap.get(String(product.id)) ?? (product as any).stock ?? 0);
      const stock = Math.max(0, baseStock - Number(optimisticSoldQuantities[String(product.id)] ?? 0));
      return {
        ...product,
        stock,
        category_name: categories.find((category) => String(category.id) === String(categoryId))?.name,
        category_image_url: categoryId ? getCategoryImage(String(categoryId)) : undefined,
      };
    }).filter((product) => Number(product.stock ?? 0) > 0);
  }, [productsWithInventoryFallback, products, inventoryItems, categories, optimisticSoldQuantities]);

  const visibleProducts = useMemo(
    () => productsWithStock.slice(0, productsPerPage),
    [productsPerPage, productsWithStock]
  );

  const getAvailableStockCount = useCallback(
    (productId: string) => {
      if (!productId) return 0;
      return inventoryItems.filter((item) => {
        if (String(item.product_id) !== String(productId)) return false;
        const status = String(item.status || '').toUpperCase();
        return ![
          'SOLD',
          'RESERVED',
          'DAMAGED',
          'IN_REPAIR',
          'RETURNED',
          'FOR_PARTS',
          'ARCHIVED',
        ].includes(status);
      }).length;
    },
    [inventoryItems]
  );

  const canAddProductToCart = useCallback(
    (productId: string, productName: string, quantity: number = 1) => {
      const currentQuantity = cart
        .filter((item) => String(item.id) === String(productId))
        .reduce((sum, item) => sum + item.quantity, 0);
      const available = getAvailableStockCount(productId);

      if (available <= 0 || currentQuantity + quantity > available) {
        const label = productName || 'هذا النوع';
        toast.error(
          `هذا النوع قد نفذ من المخزون: ${label}`,
          'مخزون غير كافٍ',
          4000
        );
        return false;
      }
      return true;
    },
    [cart, getAvailableStockCount, toast]
  );

  // Mutations
  const createCustomerMutation = useMutation({
    mutationFn: (data: CustomerCreateRequest) => customersApi.create(data),
    onSuccess: (response) => {
      const payload = response?.data as { customer?: Customer } | Customer;
      const createdCustomer = payload?.customer ?? payload;
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      if (createdCustomer?.id) {
        setSelectedCustomer(String(createdCustomer.id));
        setSelectedCustomerOption({
          id: String(createdCustomer.id),
          name: createdCustomer.name,
        });
        setCustomerBalance(0);
        setCustomerCreditLimit(createdCustomer.credit_limit);
      }
      setIsQuickCustomerModalOpen(false);
      setQuickCustomerName('');
      setQuickCustomerPhone('');
    },
    onError: (error) => {
      console.error('Customer creation failed:', error);
      toast.error(
        'تعذر إنشاء العميل. تحقق من البيانات وحاول مرة أخرى.',
        'فشل إنشاء العميل',
        4000
      );
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
    mutationFn: async () => {
      const payload = buildManualProductPayload({
        name: manualProduct.name,
        price: manualProduct.price,
        quantity: manualProduct.quantity,
        barcode: manualProduct.barcode,
      });
      return productsApi.create(payload);
    },
    onSuccess: async (response) => {
      const product = (
        response?.data as { product?: PosCartProduct } | PosCartProduct
      )?.product ?? response?.data;
      if (product?.id) {
        const barcode = product.barcode || product.id;
        const quantity = Math.max(1, Number(manualProduct.quantity) || 1);
        try {
          await inventoryApi.create({
            product_id: product.id,
            quantity,
            condition: 'NEW',
            purchase_cost: Number(product.cost_price) || 0,
            selling_price: Number(product.selling_price ?? product.price) || 0,
            status: 'AVAILABLE',
            notes: 'إضافة منتج يدوي من نقطة البيع',
          });
          addToCart({ ...product, barcode }, quantity);
        } catch (error) {
          console.error('Failed to create manual product inventory:', error);
          toast.error('تم إنشاء المنتج لكن تعذرت إضافة كميته للمخزون');
          return;
        }
      }
      queryClient.invalidateQueries({ queryKey: ['products'] });
      setManualProduct({ name: '', price: '', quantity: '1', barcode: '' });
      setIsManualProductOpen(false);
    },
    onError: () => {
      toast.error(
        'تعذر إضافة المنتج. تحقق من البيانات وحاول مرة أخرى.',
        'فشل إضافة المنتج',
        4000
      );
    },
  });

  const createSaleMutation = useMutation({
    mutationFn: async (data: SaleCreateRequest) => {
      return salesApi.create(data);
    },
    onSuccess: (response) => {
      const soldProductQuantities = cart.reduce<Record<string, number>>((quantities, item) => {
        if (!item.inventoryItemId) {
          const productId = String(item.id);
          quantities[productId] = (quantities[productId] ?? 0) + item.quantity;
        }
        return quantities;
      }, {});
      if (Object.keys(soldProductQuantities).length > 0) {
        setOptimisticSoldQuantities((current) => {
          const next = { ...current };
          for (const [productId, quantity] of Object.entries(soldProductQuantities)) {
            next[productId] = (next[productId] ?? 0) + quantity;
          }
          return next;
        });
      }
      const soldInventoryIds = cart
        .filter((item) => item.inventoryItemId)
        .map((item) => String(item.inventoryItemId));
      if (soldInventoryIds.length > 0) {
        setSoldUsedPartIds((current) => new Set([...current, ...soldInventoryIds]));
      }
      queryClient.invalidateQueries({ queryKey: ['sales'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      queryClient.invalidateQueries({ queryKey: ['debts'] });
      queryClient.invalidateQueries({ queryKey: ['reports'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['inventory', 'used-stock'] });

      if (lastSaleDataRef.current) {
        const sale = response?.data?.sale ?? response?.data ?? response?.sale;
        const persistedAllocations = response?.data?.payment_allocations ?? sale?.payment_allocations;
        setLastSaleData({
          ...lastSaleDataRef.current,
          id: sale?.id || response?.id || lastSaleDataRef.current.id,
          cashReceived: Number(sale?.cash_received ?? lastSaleDataRef.current.cashReceived ?? 0),
          changeAmount: Number(sale?.change_amount ?? lastSaleDataRef.current.changeAmount ?? 0),
          ...(Array.isArray(persistedAllocations) ? { paymentAllocations: persistedAllocations } : {}),
        });
        setIsInvoiceModalOpen(true);
      }

      pendingExternalPaymentRef.current = null;

      clearCart();
      setPaidAmount('');
      setSelectedCustomer('');
      setTaxExempt(false);
      resetPayment();
      setProcessing(false);
    },
    onError: (error: unknown) => {
      const externalPaymentID = pendingExternalPaymentRef.current;
      if (externalPaymentID) {
        pendingExternalPaymentRef.current = null;
        void paymentTransactionsApi.refund(externalPaymentID, {
          amount_minor: Math.round(displayTotal * 100),
          currency: 'ILS',
          idempotency_key: `sale-failure-refund-${externalPaymentID}`,
          reason: 'sale_creation_failed',
        }).catch((refundError) => console.error('Failed to compensate external payment:', refundError));
      }
      const apiError = error as {
        response?: { error?: { message?: string } };
        arabicMessage?: string;
        message?: string;
      };
      const message =
        apiError.response?.error?.message ||
        apiError.arabicMessage ||
        apiError.message ||
        'فشل إتمام البيع';
      const normalizedMessage = String(message).toLowerCase();
      if (
        normalizedMessage.includes('insufficient stock') ||
        normalizedMessage.includes('نفذ') ||
        normalizedMessage.includes('مخزون')
      ) {
        const exhaustedProductName = message.split(':').slice(1).join(':').trim();
        if (exhaustedProductName) {
          cart
            .filter((item) => item.name.trim() === exhaustedProductName)
            .forEach((item) => removeFromCart(item.barcode));
        }
        queryClient.invalidateQueries({ queryKey: ['inventory'] });
        queryClient.invalidateQueries({ queryKey: ['inventory', 'used-stock'] });
        toast.error(
          exhaustedProductName
            ? `تمت إزالة القطعة النافدة من السلة: ${exhaustedProductName}`
            : 'هذا النوع قد نفد من المخزون',
          'مخزون غير كافٍ',
          4000,
        );
      } else {
        toast.error(message, 'فشل إتمام البيع', 4000);
      }
      console.error('Sale failed:', error);
      setProcessing(false);
    },
  });

  // Handlers
  const handleClearSearch = () => setSearchQuery('');

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
    if (!searchQuery.trim()) return;

    if (posSection === 'used') {
      const scannedValue = searchQuery.trim().toLowerCase();
      const usedPart = inventoryItems.find((item: any) =>
        String(item.status || '').toUpperCase() === 'AVAILABLE' &&
        String(item.condition || '').toUpperCase() === 'USED' &&
        [item.barcode, item.serial_number, item.id].some(
          (value) => String(value || '').toLowerCase() === scannedValue
        )
      );
      if (usedPart) {
        handleUsedPartSelect(usedPart);
      } else {
        toast.error('لم يتم العثور على قطعة مستعملة بهذا الباركود', 'قطعة غير موجودة', 4000);
      }
      setSearchQuery('');
      return;
    }

    try {
      const response = await barcodeApi.lookupProduct(searchQuery.trim());
      const product = response.data as Product & { inventory_item_id?: string; serial_number?: string; inventory_item?: any };

      if (product && product.id) {
        if (!canAddProductToCart(product.id, product.name, 1)) return;
        addToCart({
          id: product.id,
          inventoryItemId: product.inventory_item_id || product.inventory_item?.id,
          serialNumber: product.serial_number || product.inventory_item?.serial_number,
          name: product.name,
          barcode: product.barcode || searchQuery.trim(),
          price: normalizePosPrice(
            product.sellingPrice,
            (product as Product & { selling_price?: number }).selling_price
          ),
          stock: product.stock,
          condition: product.condition,
          purchaseCost: product.costPrice || product.cost_price || 0,
        });
        setSearchQuery('');
      } else {
        if (soundEnabled) playScanSound(false);
        setUnknownBarcode(searchQuery.trim());
        setSearchQuery('');
      }
    } catch {
      if (soundEnabled) playScanSound(false);
      const product = products?.find(
        (p) =>
          p.sku === searchQuery.trim() || p.barcode === searchQuery.trim()
      );
      if (product) {
        if (!canAddProductToCart(product.id, product.name, 1)) return;
        addToCart({
          id: product.id,
          name: product.name,
          barcode: searchQuery.trim(),
          price: normalizePosPrice(
            product.sellingPrice,
            (product as Product & { selling_price?: number }).selling_price
          ),
          stock: product.stock,
          condition: product.condition,
          purchaseCost: product.cost_price || product.costPrice || 0,
        });
        setSearchQuery('');
      } else {
        setUnknownBarcode(searchQuery.trim());
        setSearchQuery('');
      }
    }
  };

  const handleProductSelect = useCallback(
    (product: PosCartProduct) => {
      if (!canAddProductToCart(product.id, product.name, 1)) return;
      addToCart(product);
    },
    [addToCart, canAddProductToCart]
  );

  const handleCheckout = useCallback(async () => {
    if (cart.length === 0) return;
    if (paymentMethod === 'credit' && paidAmount.trim() === '') return;

    let latestInventoryItems = inventoryItems;
    try {
      const latestInventory = await queryClient.fetchQuery({
        queryKey: ['inventory'],
        queryFn: () => inventoryApi.listWithSupplier({ page: 1, per_page: 1000 }),
        staleTime: 0,
      });
      latestInventoryItems = ((latestInventory?.data?.items as unknown) as InventoryItem[]) || [];
    } catch {
    }

    const availableStockCount = (productId: string) => latestInventoryItems.filter((item: any) => {
      if (String(item.product_id) !== String(productId)) return false;
      const status = String(item.status || '').toUpperCase();
      return !['SOLD', 'RESERVED', 'DAMAGED', 'IN_REPAIR', 'RETURNED', 'FOR_PARTS', 'ARCHIVED'].includes(status);
    }).length;

    const exhaustedItems = cart.reduce((items, item) => {
      const availableStock = item.inventoryItemId
        ? latestInventoryItems.some((inventoryItem: any) =>
          String(inventoryItem.id) === String(item.inventoryItemId) &&
          String(inventoryItem.status || '').toUpperCase() === 'AVAILABLE'
        ) ? 1 : 0
        : availableStockCount(String(item.id));
      if (availableStock <= 0 || item.quantity > availableStock) {
        items.push(item.name);
      }
      return items;
    }, [] as string[]);

    if (exhaustedItems.length > 0) {
      const uniqueItems = [...new Set(exhaustedItems)];
      const message =
        uniqueItems.length > 1
          ? `هذه الأنواع قد نفذت من المخزون: ${uniqueItems.slice(0, 2).join(', ')}`
          : `هذا النوع قد نفذ من المخزون: ${uniqueItems[0]}`;
      toast.error(message, 'مخزون غير كافٍ', 4000);
      return;
    }

    setProcessing(true);

    let paymentTransactionID: string | undefined;
    if (paymentMethod === 'card' && electronicPaymentsEnabled && electronicPaymentProvider !== 'manual') {
      const orderID = crypto.randomUUID();
      try {
        const createdResponse = await paymentTransactionsApi.create({
          order_id: orderID,
          amount_minor: Math.round(displayTotal * 100),
          currency: 'ILS',
          description: `PartFlow sale ${orderID}`,
          idempotency_key: `pos-${orderID}`,
        });
        const created = createdResponse?.data?.data ?? createdResponse?.data;
        if (!created?.id) throw new Error('لم يتم إنشاء معاملة الدفع');
        pendingExternalPaymentRef.current = String(created.id);
        if (created.checkout_url) window.open(created.checkout_url, '_blank', 'noopener,noreferrer');

        let verified: any = null;
        for (let attempt = 0; attempt < 60; attempt += 1) {
          await new Promise((resolve) => window.setTimeout(resolve, 2000));
          const verifiedResponse = await paymentTransactionsApi.verify(String(created.id));
          verified = verifiedResponse?.data?.data ?? verifiedResponse?.data;
          if (verified?.status === 'paid' || verified?.status === 'failed' || verified?.status === 'cancelled') break;
        }
        if (verified?.status !== 'paid') throw new Error('لم يتم تأكيد الدفع الإلكتروني من المزود');
        paymentTransactionID = String(created.id);
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'فشل الدفع الإلكتروني', 'فشل الدفع', 5000);
        setProcessing(false);
        return;
      }
    }

    const receivedAmount = parseFloat(paidAmount) || 0;
    const appliedPaymentAmount = paymentMethod === 'cash'
      ? Math.min(receivedAmount, displayTotal)
      : receivedAmount;
    const changeAmount = paymentMethod === 'cash'
      ? Math.max(0, receivedAmount - displayTotal)
      : 0;
    const saleData: SaleCreateRequest = {
      customer_id: selectedCustomer || undefined,
      items: cart.map((item) => ({
        product_id: item.id,
        quantity: item.quantity,
        unit_price: item.price,
        ...(item.inventoryItemId ? { inventory_item_id: item.inventoryItemId } : {}),
        ...(item.inventoryItemId ? { is_trade_in: true } : {}),
        ...(item.purchaseCost ? { purchase_cost: item.purchaseCost } : {}),
      })),
      payment_method: paymentMethod === 'credit' ? 'debt' : paymentMethod === 'checks' ? 'transfer' : paymentMethod,
      payment_amount: appliedPaymentAmount,
      ...(paymentAllocations.length > 0 ? { payment_allocations: paymentAllocations } : {}),
      ...(paymentTransactionID ? { payment_transaction_id: paymentTransactionID } : {}),
      ...(paymentMethod === 'cash' ? { cash_received: receivedAmount } : {}),
      total_amount: displayTotal,
      tax_exempt: taxExempt,
      ...(appliedDiscountRate > 0 ? { discount_type: 'percentage', discount_value: appliedDiscountRate } : {}),
    };

    const invoiceData: InvoiceData = {
      id: 'pending',
      customerName: selectedCustomer
        ? customers.find((c) => String(c.id) === String(selectedCustomer))
            ?.name ?? selectedCustomerOption?.name
        : '',
      customerPhone: selectedCustomer
        ? customers.find((c) => String(c.id) === String(selectedCustomer))
            ?.phone
        : undefined,
      saleDate: new Date().toISOString(),
      items: displayCart.map((item) => ({
        name: item.name,
        partType: item.partType,
        partTypeColor: item.partTypeColor,
        condition: item.condition,
        grade: item.grade,
        serialNumber: item.serialNumber,
        sellingPrice: item.price,
        quantity: item.quantity,
        total: item.total,
      })),
      subtotal: displaySubtotal,
      total: displayTotal,
      paidAmount: appliedPaymentAmount,
      cashReceived: paymentMethod === 'cash' ? receivedAmount : undefined,
      changeAmount,
      remaining: Math.max(0, displayTotal - appliedPaymentAmount),
      paymentMethod,
      paymentAllocations,
    };

    setLastSaleData(invoiceData);
    lastSaleDataRef.current = invoiceData;
    createSaleMutation.mutate(saleData);
  }, [
    cart,
    selectedCustomer,
    paymentMethod,
    paidAmount,
    total,
    subtotal,
    displayCart,
    displaySubtotal,
    displayTotal,
    appliedDiscountRate,
    taxExempt,
    paymentAllocations,
    electronicPaymentsEnabled,
    electronicPaymentProvider,
    customers,
    setProcessing,
    createSaleMutation,
    inventoryItems,
    queryClient,
    removeFromCart,
    toast,
  ]);

  // Quick customer creation handler
  const handleQuickCustomerCreate = useCallback(() => {
    setIsQuickCustomerModalOpen(true)
  }, [])

  // Quick sale handler - one click to complete cash sale
  const handleQuickSale = useCallback(() => {
    if (cart.length === 0) return
    
    const exhaustedItems = cart.reduce((items, item) => {
      const availableStock = item.isTradeIn
        ? Number(item.stock ?? 1)
        : getAvailableStockCount(String(item.id))
      if (availableStock <= 0 || item.quantity > availableStock) {
        items.push(item.name)
      }
      return items
    }, [] as string[])

    if (exhaustedItems.length > 0) {
      const uniqueItems = [...new Set(exhaustedItems)]
      const message =
        uniqueItems.length > 1
          ? `هذه الأنواع قد نفذت من المخزون: ${uniqueItems.slice(0, 2).join(', ')}`
          : `هذا النوع قد نفذ من المخزون: ${uniqueItems[0]}`
      toast.error(message, 'مخزون غير كافٍ', 4000)
      return
    }

    setProcessing(true)

    const saleData: SaleCreateRequest = {
      customer_id: selectedCustomer || undefined,
      items: cart.map((item) => ({
        product_id: item.id,
        quantity: item.quantity,
        unit_price: item.price,
        ...(item.inventoryItemId ? { inventory_item_id: item.inventoryItemId } : {}),
        ...(item.inventoryItemId ? { is_trade_in: true } : {}),
        ...(item.purchaseCost ? { purchase_cost: item.purchaseCost } : {}),
      })),
      payment_method: 'cash',
      payment_amount: displayTotal,
      total_amount: displayTotal,
      tax_exempt: taxExempt,
      ...(appliedDiscountRate > 0 ? { discount_type: 'percentage', discount_value: appliedDiscountRate } : {}),
    }

    const invoiceData: InvoiceData = {
      id: 'pending',
      customerName: selectedCustomer
        ? customers.find((c) => String(c.id) === String(selectedCustomer))
            ?.name ?? selectedCustomerOption?.name
        : '',
      customerPhone: selectedCustomer
        ? customers.find((c) => String(c.id) === String(selectedCustomer))
            ?.phone
        : undefined,
      saleDate: new Date().toISOString(),
      items: displayCart.map((item) => ({
        name: item.name,
        partType: item.partType,
        partTypeColor: item.partTypeColor,
        condition: item.condition,
        grade: item.grade,
        serialNumber: item.serialNumber,
        sellingPrice: item.price,
        quantity: item.quantity,
        total: item.total,
      })),
      subtotal: displaySubtotal,
      total: displayTotal,
      paidAmount: displayTotal,
      remaining: 0,
      paymentMethod: 'cash',
    }

    setLastSaleData(invoiceData)
    lastSaleDataRef.current = invoiceData
    createSaleMutation.mutate(saleData)
  }, [
    cart,
    selectedCustomer,
    total,
    subtotal,
    displayCart,
    displaySubtotal,
    displayTotal,
    appliedDiscountRate,
    taxExempt,
    customers,
    getAvailableStockCount,
    toast,
    createSaleMutation,
    selectedCustomerOption?.name,
  ])

  // Cart animation state
  const [cartBounce, setCartBounce] = useState(false)
  
  // Trigger cart animation when item added
  useEffect(() => {
    if (cart.length > 0) {
      setCartBounce(true)
      const timer = setTimeout(() => setCartBounce(false), 300)
      return () => clearTimeout(timer)
    }
  }, [cart.length]);

  const handleCustomerChange = (customerId: string) => {
    setSelectedCustomer(customerId);
    if (customerId) {
      const customer = customers.find(
        (c) => String(c.id) === String(customerId)
      );
      if (customer) {
        setSelectedCustomerOption({
          id: String(customer.id),
          name: customer.name,
        });
        setCustomerBalance(customer.current_balance ?? customer.balance ?? 0);
        setCustomerCreditLimit(customer.credit_limit);
      }
    } else {
      setSelectedCustomerOption(undefined);
      setCustomerBalance(0);
      setCustomerCreditLimit(undefined);
    }
  };

  return (
    <div className="pos-modern-container">
      {/* Modern Header */}
      <header className="pos-modern-header">
        <div className="pos-header-left">
          <button
            className="checkout-mode-toggle"
            onClick={() => setCheckoutMode(!checkoutMode)}
            title={checkoutMode ? 'عرض القائمة الجانبية' : 'توسيع قسم الدفع'}
          >
            <ArrowRight className="w-5 h-5" style={{
              transform: checkoutMode ? 'rotate(180deg)' : 'rotate(0deg)',
              transition: 'transform 0.3s ease'
            }} />
          </button>
          <div className="pos-logo">
            <PartFlowLogo size={36} priority />
          </div>
          <div className="pos-header-titles">
            <h1 className="pos-main-title">نقطة البيع</h1>
            <p className="pos-subtitle">{t('sales.posTitle')}</p>
          </div>
        </div>
        <div className="pos-header-right">
          <div className="pos-meta-badges">
            {/* Cart Count Badge */}
            <span 
              className={`pos-badge cart-count ${cartBounce ? 'bounce' : ''}`}
              style={{
                background: cart.length > 0 ? 'var(--color-primary)' : 'var(--bg-surface-elevated)',
                color: cart.length > 0 ? 'white' : 'var(--text-secondary)',
              }}
            >
              <ShoppingCart className="w-3.5 h-3.5" />
              <span className="cart-count-number">{cart.length}</span>
            </span>
            {/* Total Amount Badge - shows when cart has items */}
            {cart.length > 0 && (
              <span 
                className="pos-badge total-amount"
                style={{
                  background: 'var(--color-info-10)',
                  color: 'var(--color-info)',
                  border: '1px solid var(--color-info-20)',
                  fontWeight: 700,
                }}
              >
                ₪{total.toLocaleString()}
              </span>
            )}
            <span className="pos-badge cashier">
              <Wifi className="w-3.5 h-3.5 text-green-500" />
              <span>Cashier 01</span>
            </span>
            <span className="pos-badge time">
              {new Intl.DateTimeFormat('ar', {
                hour: '2-digit',
                minute: '2-digit',
              }).format(new Date())}
            </span>
          </div>
          <div className="pos-header-actions">
            {/* Quick Sale Button - appears when cart has items */}
            {cart.length > 0 && (
              <Button
                variant="primary"
                size="sm"
                onClick={handleQuickSale}
                disabled={isProcessing}
                className="gap-2 quick-sale-btn"
              >
                <Zap className="w-4 h-4" />
                <span>بيع سريع</span>
                <span className="quick-sale-total">₪{total.toLocaleString()}</span>
              </Button>
            )}
            {/* Clear Cart Button - appears when cart has items */}
            {cart.length > 0 && (
              <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                  clearCart();
                  resetPayment();
                  setPaymentAllocations([]);
                }}
                className="gap-2"
                style={{ padding: '0.375rem 0.5rem' }}
              >
                <Trash2 className="w-4 h-4" />
              </Button>
            )}
            <Button
              variant="secondary"
              size="sm"
              onClick={() => lastSaleData && setIsInvoiceModalOpen(true)}
              disabled={!lastSaleData}
              className="gap-2"
            >
              <Printer className="w-4 h-4" />
              <span>طباعة</span>
            </Button>
          </div>
        </div>
      </header>

      {/* Main POS Body */}
      <div className="pos-modern-body">
        {/* Products Section */}
        <main className="pos-products-area">
          <div className="flex gap-2 mb-3" role="tablist" aria-label="نوع المنتجات">
            <button
              type="button"
              role="tab"
              aria-selected={posSection === 'products'}
              className={`category-chip ${posSection === 'products' ? 'active' : ''}`}
              onClick={() => setPosSection('products')}
            >
              <Package className="w-3.5 h-3.5" />
              <span>المنتجات الجديدة</span>
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={posSection === 'used'}
              className={`category-chip ${posSection === 'used' ? 'active' : ''}`}
              onClick={() => setPosSection('used')}
            >
              <ShoppingCart className="w-3.5 h-3.5" />
              <span>القطع المستعملة</span>
            </button>
          </div>

          {/* Search & Barcode */}
          <form onSubmit={handleBarcodeScan} className="pos-search-bar">
            <div className="pos-barcode-form">
              <ScanLine className="barcode-icon" />
              <Input
                placeholder="امسح الباركود أو اكتب للبحث..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pos-barcode-input"
                autoFocus
              />
              {searchQuery && (
                <button
                  type="button"
                  className="barcode-clear"
                  onClick={() => {
                    setSearchQuery('');
                    handleClearSearch();
                  }}
                >
                  <X className="w-4 h-4" />
                </button>
              )}
            </div>
          </form>

          {posSection === 'products' && (
            <div className="pos-category-bar">
              <button
                className={`category-chip ${selectedCategory === null ? 'active' : ''}`}
                onClick={() => setSelectedCategory(null)}
              >
                <Package className="w-3.5 h-3.5" />
                <span>الكل</span>
              </button>
              {categories.map((cat) => (
                <button
                  key={cat.id}
                  className={`category-chip ${selectedCategory === cat.id ? 'active' : ''}`}
                  onClick={() => setSelectedCategory(cat.id)}
                >
                  {cat.name}
                </button>
              ))}
            </div>
          )}

          {/* Products Grid */}
          {posSection === 'products' ? (
            <ModernProductGrid
              products={visibleProducts}
              onProductClick={handleProductSelect}
              taxRate={effectiveTaxRate}
              taxExempt={taxExempt}
              showDetails={showProductDetails}
              onToggleDetails={() => setShowProductDetails((visible) => !visible)}
              currentPage={productPage}
              totalPages={Math.max(1, Math.ceil(productTotal / productsPerPage))}
              viewMode={productViewMode}
              onPreviousPage={() => setProductPage((page) => Math.max(1, page - 1))}
              onNextPage={() => setProductPage((page) => page + 1)}
              isLoading={productsLoading}
            />
          ) : (
            <>
              <div className="pos-product-view-toolbar">
                <button
                  type="button"
                  className={`pos-product-details-toggle ${showProductDetails ? 'active' : ''}`}
                  onClick={() => setShowProductDetails((visible) => !visible)}
                  aria-pressed={showProductDetails}
                  title={showProductDetails ? 'إخفاء تفاصيل القطعة' : 'عرض تفاصيل القطعة'}
                >
                  <Info className="h-4 w-4" aria-hidden="true" />
                  <span>{showProductDetails ? 'إخفاء التفاصيل' : 'عرض التفاصيل'}</span>
                </button>
              </div>
              <div className={productViewMode === 'list' ? 'pos-modern-products-list' : 'pos-modern-products-grid'}>
              {availableUsedParts.length === 0 ? (
                <div className="pos-modern-products-empty">
                  <ShoppingCart className="empty-icon" />
                  <p className="empty-title">لا توجد قطع مستعملة</p>
                  <p className="empty-subtitle">ابدأ بالبحث أو امسح باركود القطعة</p>
                </div>
              ) : availableUsedParts.map((item: any) => {
                const partType = partTypes.find((type) => type.id === item.part_type_id);
                const partName = item.product_name || item.product?.name || 'قطعة مستعملة';
                const price = Number(item.selling_price || 0);
                const partImage = getPartTypeImage(String(item.part_type_id));
                const partTypeLabel = partType
                  ? [partType.name_ar || partType.name, partType.name_en].filter(Boolean).join(' · ')
                  : 'نوع غير محدد';
                if (productViewMode === 'list') {
                  return (
                    <button
                      type="button"
                      key={item.id}
                      className="pos-modern-product-list-item"
                      onClick={() => handleUsedPartSelect(item)}
                    >
                      <span className="pos-modern-product-list-main min-w-0">
                        <span className="pos-modern-product-list-thumb">
                          {partImage ? <img src={partImage} alt="" /> : <ShoppingCart className="h-4 w-4" aria-hidden="true" />}
                        </span>
                        <span className="min-w-0">
                          <span className="pos-modern-product-list-name block">{partName}</span>
                          <span className="pos-modern-product-list-category block">
                            {partTypeLabel}
                          </span>
                          {showProductDetails && (item.serial_number || item.grade) && (
                            <span className="pos-modern-product-list-identifiers">
                              {item.serial_number && <span>تسلسلي: {item.serial_number}</span>}
                              {item.grade && <span>التقييم: {item.grade}</span>}
                            </span>
                          )}
                        </span>
                      </span>
                      <span className="pos-modern-product-list-stock">مستعمل · متوفر</span>
                      <span className="pos-modern-product-list-price">₪{price.toLocaleString()}</span>
                      <Plus className="h-4 w-4" aria-hidden="true" />
                    </button>
                  );
                }
                return (
                  <button
                    type="button"
                    key={item.id}
                    className="pos-modern-product-card"
                    onClick={() => handleUsedPartSelect(item)}
                  >
                    <div className="product-card-image">
                      {getPartTypeImage(String(item.part_type_id)) ? (
                        <img src={getPartTypeImage(String(item.part_type_id))} alt={partName} />
                      ) : (
                        <div className="product-card-placeholder"><ShoppingCart /></div>
                      )}
                      <div className="product-card-add">
                        <Plus className="w-5 h-5" />
                      </div>
                    </div>
                    <div className="product-card-info">
                      <h3 className="product-card-name">{partName}</h3>
                      <div className="product-card-meta">
                        <span className="product-card-condition used">مستعمل</span>
                        <span className="product-card-stock">متوفر</span>
                        {partType && <span className="product-card-status success">{partTypeLabel}</span>}
                      </div>
                      {showProductDetails && (item.serial_number || item.grade) && (
                        <div className="product-card-identifiers">
                          {item.serial_number && <span>تسلسلي: {item.serial_number}</span>}
                          {item.grade && <span>التقييم: {item.grade}</span>}
                        </div>
                      )}
                      <div className="product-card-price">
                        <span className="price-value">₪{price.toLocaleString()}</span>
                      </div>
                    </div>
                  </button>
                );
                })}
              </div>
              {availableUsedParts.length > 0 && (
                <div className="pos-modern-pagination" dir="rtl" aria-label="التنقل بين صفحات القطع المستعملة">
                  <button
                    type="button"
                    className="pos-modern-pagination-btn"
                    onClick={() => setUsedPartPage((page) => Math.max(1, page - 1))}
                    disabled={usedPartPage <= 1}
                    aria-label="الصفحة السابقة للقطع المستعملة"
                  >
                    السابق
                  </button>
                  <span className="pos-modern-pagination-status" aria-live="polite">
                    صفحة {usedPartPage} من {usedPartTotalPages}
                  </span>
                  <button
                    type="button"
                    className="pos-modern-pagination-btn"
                    onClick={() => setUsedPartPage((page) => Math.min(usedPartTotalPages, page + 1))}
                    disabled={usedPartPage >= usedPartTotalPages}
                    aria-label="الصفحة التالية للقطع المستعملة"
                  >
                    التالي
                  </button>
                </div>
              )}
            </>
          )}
        </main>

        {/* Cart & Payment Sidebar */}
        <aside className="pos-cart-sidebar">
          {/* Customer Selector */}
          <div className="pos-customer-bar">
            <div className="customer-select-wrapper">
              <button
                type="button"
                className={`customer-select-trigger ${isCustomerMenuOpen ? 'open' : ''}`}
                onClick={() => setIsCustomerMenuOpen((open) => !open)}
                aria-haspopup="listbox"
                aria-expanded={isCustomerMenuOpen}
              >
                <span className="customer-select-avatar" aria-hidden="true"><UserRound className="h-4 w-4" /></span>
                <span className="customer-select-copy">
                  <span className="customer-select-label">العميل</span>
                  <span className="customer-select-value">{selectedCustomerOption?.name || 'عميل عام'}</span>
                </span>
                <ChevronDown className="customer-select-chevron" aria-hidden="true" />
              </button>
              {isCustomerMenuOpen && (
                <div className="customer-select-menu" role="listbox" aria-label="اختيار العميل">
                  <button
                    type="button"
                    className={`customer-option ${!selectedCustomer ? 'selected' : ''}`}
                    onClick={() => { handleCustomerChange(''); setIsCustomerMenuOpen(false); }}
                    role="option"
                    aria-selected={!selectedCustomer}
                  >
                    <span className="customer-option-avatar guest" aria-hidden="true"><UserRound className="h-4 w-4" /></span>
                    <span className="customer-option-copy"><strong>عميل عام</strong><small>بيع مباشر بدون حساب عميل</small></span>
                  </button>
                  {customers.map((customer) => (
                    <button
                      key={customer.id}
                      type="button"
                      className={`customer-option ${String(customer.id) === String(selectedCustomer) ? 'selected' : ''}`}
                      onClick={() => { handleCustomerChange(String(customer.id)); setIsCustomerMenuOpen(false); }}
                      role="option"
                      aria-selected={String(customer.id) === String(selectedCustomer)}
                    >
                      <span className="customer-option-avatar" aria-hidden="true"><UserRound className="h-4 w-4" /></span>
                      <span className="customer-option-copy"><strong>{customer.name}</strong><small>حساب عميل</small></span>
                    </button>
                  ))}
                </div>
              )}
            </div>
            <button
              className="quick-customer-btn"
              onClick={handleQuickCustomerCreate}
            >
              <Plus className="w-4 h-4" />
            </button>
          </div>

          <div className="pos-invoice-controls">
            <label className="pos-tax-toggle">
              <input
                type="checkbox"
                checked={taxExempt}
                onChange={(event) => setTaxExempt(event.target.checked)}
                className="h-4 w-4 accent-cyan"
              />
              <span>معفى من الضريبة لهذه الفاتورة</span>
            </label>

            {discountsEnabled && (
              <div className="pos-discount-control">
                <Input
                  label="خصم على الفاتورة (%)"
                  type="number"
                  min="0"
                  max={maxDiscountRate}
                  step="0.01"
                  value={discountRate}
                  onChange={(event) => setDiscountRate(Math.min(Math.max(Number(event.target.value) || 0, 0), maxDiscountRate))}
                />
                <p className="pos-discount-help">
                  أقصى خصم: {maxDiscountRate}%
                </p>
              </div>
            )}
          </div>

          {/* Cart Panel */}
          <ModernCartPanel
            cart={displayCart}
            total={displayTotal}
            onUpdateQuantity={updateQuantity}
            onRemoveFromCart={removeFromCart}
            onClearCart={() => {
              clearCart();
              resetPayment();
            }}
          />

          {/* Payment Panel */}
          <AdvancedPaymentPanel
            paymentMethod={paymentMethod}
            setPaymentMethod={setPaymentMethod}
            paidAmount={paidAmount}
            setPaidAmount={setPaidAmount}
            total={displayTotal}
            isProcessing={isProcessing}
            onCheckout={handleCheckout}
            onPaymentAllocationsChange={setPaymentAllocations}
            onPaymentAllocationsChange={setPaymentAllocations}
            electronicPaymentsEnabled={electronicPaymentsEnabled}
            enabledElectronicMethods={enabledElectronicMethods}
            selectedCustomer={selectedCustomer}
            customerBalance={customerBalance}
            customerCreditLimit={customerCreditLimit}
            onQuickCustomerCreate={handleQuickCustomerCreate}
          />
        </aside>
      </div>

      {/* Footer Actions */}
      <footer className="pos-modern-footer">
        <Button
          variant="secondary"
          size="sm"
          onClick={() => setIsManualProductOpen(true)}
          className="gap-2"
        >
          <Plus className="h-4 w-4" />
          <span>إضافة يدوي</span>
        </Button>
        <Button variant="secondary" size="sm" onClick={handleHoldSale}>
          <Pause className="h-4 w-4" />
          <span>تعليق البيع <kbd className="footer-kbd">F4</kbd></span>
        </Button>
        {heldSales.length > 0 && (
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setIsHeldSalesOpen(true)}
          >
            <span>المبيعات المعلقة ({heldSales.length})</span>
          </Button>
        )}
        
        {/* Keyboard Shortcuts Help */}
        <div className="footer-shortcuts">
          <span className="shortcut-hint">
            <kbd>F2</kbd> بحث
          </span>
          <span className="shortcut-hint">
            <kbd>F8</kbd> دفع
          </span>
          <span className="shortcut-hint">
            <kbd>Del</kbd> حذف
          </span>
        </div>
      </footer>

      {/* Modals */}
      {/* Held Sales Modal */}
      <Modal
        isOpen={isHeldSalesOpen}
        onClose={() => setIsHeldSalesOpen(false)}
        title="المبيعات المعلّقة"
        variant="modern"
        size="md"
      >
        <div className="flex flex-col gap-3">
          {heldSales.map((held, index) => (
            <Button
              key={held.id}
              variant="secondary"
              className="justify-between"
              onClick={() => handleResumeSale(held)}
            >
              <span>بيع معلّق #{index + 1}</span>
              <span>{held.items.length} منتجات</span>
            </Button>
          ))}
        </div>
      </Modal>

      {/* Manual Product Modal */}
      <Modal
        isOpen={isManualProductOpen}
        onClose={() => setIsManualProductOpen(false)}
        title="إضافة منتج يدويًا"
        variant="modern"
        size="md"
      >
        <div className="flex flex-col gap-3">
          <Input
            autoFocus
            placeholder="اسم المنتج *"
            value={manualProduct.name}
            onChange={(e) =>
              setManualProduct((current) => ({
                ...current,
                name: e.target.value,
              }))
            }
          />
          <Input
            type="number"
            placeholder="سعر البيع"
            value={manualProduct.price}
            onChange={(e) =>
              setManualProduct((current) => ({
                ...current,
                price: e.target.value,
              }))
            }
          />
          <Input
            placeholder="الباركود (اختياري)"
            value={manualProduct.barcode}
            onChange={(e) =>
              setManualProduct((current) => ({
                ...current,
                barcode: e.target.value,
              }))
            }
          />
          <Input
            type="number"
            min={1}
            placeholder="الكمية"
            value={manualProduct.quantity}
            onChange={(e) =>
              setManualProduct((current) => ({
                ...current,
                quantity: e.target.value,
              }))
            }
          />
          {createProductMutation.isError && (
            <p className="text-sm text-red-500">
              تعذر إنشاء المنتج، تحقق من البيانات.
            </p>
          )}
          <Button
            variant="primary"
            disabled={!manualProduct.name.trim() || createProductMutation.isPending}
            onClick={() => createProductMutation.mutate()}
          >
            {createProductMutation.isPending
              ? 'جارٍ الحفظ...'
              : 'حفظ وإضافة للسلة'}
          </Button>
        </div>
      </Modal>

      {/* Unknown Barcode Modal */}
      <Modal
        isOpen={Boolean(unknownBarcode)}
        onClose={() => setUnknownBarcode('')}
        title="المنتج غير موجود"
        variant="modern"
        size="sm"
      >
        <div className="flex flex-col gap-3">
          <p className="text-sm text-text-secondary">
            الباركود: <strong>{unknownBarcode}</strong>
          </p>
          <Button
            variant="secondary"
            onClick={() => {
              setSearchQuery(unknownBarcode);
              setUnknownBarcode('');
            }}
          >
            البحث عن منتج
          </Button>
          <Button
            variant="primary"
            onClick={() => {
              setManualProduct((current) => ({
                ...current,
                barcode: unknownBarcode,
              }));
              setUnknownBarcode('');
              setIsManualProductOpen(true);
            }}
          >
            إنشاء وإضافة للمخزون
          </Button>
          <Button variant="ghost" onClick={() => setUnknownBarcode('')}>
            إلغاء
          </Button>
        </div>
      </Modal>

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

      {/* Quick Customer Creation Modal */}
      <Modal
        isOpen={isQuickCustomerModalOpen}
        onClose={() => setIsQuickCustomerModalOpen(false)}
        title="إضافة عميل سريع"
        variant="modern"
        size="md"
      >
        <div className="flex flex-col gap-4">
          <div>
            <label className="text-sm font-medium text-text-secondary block mb-2">
              الاسم *
            </label>
            <Input
              placeholder="أدخل اسم العميل..."
              autoFocus
              value={quickCustomerName}
              onChange={(e) => setQuickCustomerName(e.target.value)}
            />
          </div>
          <div>
            <label className="text-sm font-medium text-text-secondary block mb-2">
              رقم الهاتف
            </label>
            <Input
              type="tel"
              placeholder="05xxxxxxxx"
              value={quickCustomerPhone}
              onChange={(e) => setQuickCustomerPhone(e.target.value)}
            />
          </div>
          {createCustomerMutation.isError && (
            <p className="text-sm text-red-500">
              تعذر إنشاء العميل. تحقق من البيانات وحاول مرة أخرى.
            </p>
          )}
          <div className="flex gap-3 justify-end mt-4">
            <Button
              variant="secondary"
              onClick={() => setIsQuickCustomerModalOpen(false)}
            >
              إلغاء
            </Button>
            <Button
              variant="primary"
              disabled={
                createCustomerMutation.isPending || !quickCustomerName.trim()
              }
              onClick={() =>
                createCustomerMutation.mutate({
                  name: quickCustomerName.trim(),
                  phone: quickCustomerPhone.trim() || undefined,
                  credit_limit: 0,
                })
              }
            >
              {createCustomerMutation.isPending
                ? 'جارٍ الحفظ...'
                : 'حفظ ومتابعة البيع'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
