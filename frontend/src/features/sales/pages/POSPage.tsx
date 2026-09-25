import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { useToast } from '../../../hooks/useToast';
import { formatStoreDateTime } from '../../../utils/store-time';
import { useUIStore } from '../../../stores/uiStore';
import { Modal } from '../../../design-system/components/modal';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { PartFlowLogo } from '../../../components/branding/PartFlowLogo';
import { SalesInvoice } from '../../../components/invoice/SalesInvoice';
import {
  productsApi,
  salesApi,
  customersApi,
  barcodeApi,
  inventoryApi,
  categoriesApi,
  settingsApi,
  paymentTransactionsApi,
  posShiftsApi,
} from '../../../services/api/endpoints';
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
  ShoppingCart,
  Trash2,
  UserRound,
  ChevronDown,
  FilePlus2,
  Wallet,
  LockKeyhole,
  UnlockKeyhole,
  History,
  RotateCcw,
  Wifi,
  WifiOff,
  RefreshCw,
} from 'lucide-react';

// Modern Components
import { ModernProductGrid } from '../components/modern/ModernProductGrid';
import { ModernCartPanel } from '../components/modern/ModernCartPanel';
import { AdvancedPaymentPanel } from '../components/modern/AdvancedPaymentPanel';
import { PosCustomerSelector } from '../components/modern/PosCustomerSelector';
import './pos-page.css';

// Hooks
import { normalizePosPrice, useCart } from '../hooks/useCart';
import { usePayment } from '../hooks/usePayment';
import { buildManualProductPayload } from '../utils/manualProductPayload';
import { clearPosCheckoutAttempt, getPosCheckoutAttempt } from '../utils/posCheckoutAttempt';
import { createDefaultPosShift, resolvePosShiftState } from '../utils/posShift';
import { getCategoryImage } from '../../../services/localCategoryImages';

// Types
import { InvoiceData, PaymentAllocation } from '../types/pos.types';
import { Product, InventoryItem } from '../../../types/models';
import type { CustomerCreateRequest, Customer, Category } from '../../../services/api/types';
import type { PosCartProduct } from '../hooks/useCart';
import type { SaleCreateRequest } from '../../../services/api/types';

interface HeldSale {
  id: string;
  items: PosCartProduct[];
  created_at?: string;
}

interface PosShiftState {
  status: 'open' | 'closed';
  openedAt: string;
  openingCash: number;
  closedAt?: string;
  closingCash?: number;
  salesTotal: number;
  saleCount: number;
}

function normalizeRemoteShift(value: unknown): PosShiftState | null {
  if (!value || typeof value !== 'object') return null;
  const shift = value as Record<string, unknown>;
  if (shift.status !== 'open' && shift.status !== 'closed') return null;
  return resolvePosShiftState(value);
}

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === 'object' ? value as Record<string, unknown> : {};
}

function buildInvoiceFromSale(summary: unknown, details: unknown): InvoiceData {
  const summaryRow = asRecord(summary);
  const detailsRow = asRecord(details);
  const dataRow = asRecord(detailsRow.data ?? detailsRow);
  const sale = asRecord(dataRow.sale ?? dataRow);
  const customer = asRecord(sale.customer);
  const itemRows = Array.isArray(dataRow.items)
    ? dataRow.items
    : Array.isArray(sale.items)
      ? sale.items
      : [];
  const total = Number(sale.total_amount ?? summaryRow.total_amount ?? 0) || 0;
  const paidAmount = Number(sale.paid_amount ?? summaryRow.paid_amount ?? 0) || 0;
  const subtotal = Number(sale.subtotal ?? summaryRow.subtotal ?? total) || 0;
  const paymentMethod = String(sale.payment_method ?? summaryRow.payment_method ?? 'cash');

  return {
    id: String(sale.id ?? summaryRow.id ?? ''),
    invoiceNumber: String(sale.invoice_number ?? summaryRow.invoice_number ?? '') || undefined,
    customerName: String(customer.name ?? sale.customer_name ?? summaryRow.customer_name ?? 'عميل عام'),
    customerPhone: String(customer.phone ?? sale.customer_phone ?? summaryRow.customer_phone ?? '') || undefined,
    saleDate: String(sale.created_at ?? summaryRow.created_at ?? sale.sale_date ?? summaryRow.sale_date ?? new Date().toISOString()),
    items: itemRows.map((rawItem) => {
      const item = asRecord(rawItem);
      const quantity = Number(item.quantity ?? 1) || 1;
      const sellingPrice = Number(item.unit_price ?? item.selling_price ?? 0) || 0;
      return {
        name: String(item.product_name ?? item.name ?? 'منتج'),
        partType: String(item.part_type ?? item.partType ?? '') || undefined,
        partTypeColor: String(item.part_type_color ?? item.partTypeColor ?? '') || undefined,
        condition: String(item.condition ?? 'NEW'),
        grade: String(item.grade ?? '') || undefined,
        serialNumber: String(item.serial_number ?? item.serialNumber ?? '') || undefined,
        sellingPrice,
        quantity,
        total: Number(item.total_amount ?? item.total ?? sellingPrice * quantity) || 0,
      };
    }),
    subtotal,
    discountAmount: Number(sale.discount_amount ?? summaryRow.discount_amount ?? Math.max(0, subtotal - total)) || 0,
    taxAmount: Number(sale.tax_amount ?? summaryRow.tax_amount ?? 0) || 0,
    total,
    paidAmount,
    paymentStatus: String(sale.payment_status ?? summaryRow.payment_status ?? '') || undefined,
    notes: String(sale.notes ?? summaryRow.notes ?? '') || undefined,
    cashReceived: Number(sale.cash_received ?? summaryRow.cash_received ?? 0) || 0,
    changeAmount: Number(sale.change_amount ?? summaryRow.change_amount ?? 0) || 0,
    remaining: Math.max(0, Number(sale.remaining_amount ?? summaryRow.remaining_amount ?? total - paidAmount) || 0),
    paymentMethod: paymentMethod === 'debt' ? 'credit' : paymentMethod === 'transfer' ? 'checks' : paymentMethod as InvoiceData['paymentMethod'],
  };
}

export function POSPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
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
  const { data: installmentWhatsAppSetting } = useQuery({
    queryKey: ['settings', 'installment_whatsapp_number'],
    queryFn: () => settingsApi.getSetting('installment_whatsapp_number'),
    retry: false,
  });
  const { data: storeNameSetting } = useQuery({
    queryKey: ['settings', 'store_name'],
    queryFn: () => settingsApi.getSetting('store_name'),
    retry: false,
  });
  const { data: installmentWhatsAppMessageSetting } = useQuery({
    queryKey: ['settings', 'installment_whatsapp_message'],
    queryFn: () => settingsApi.getSetting('installment_whatsapp_message'),
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
  const installmentWhatsAppNumber = String(installmentWhatsAppSetting?.data?.value ?? '');
  const storeName = String(storeNameSetting?.data?.value ?? '').trim() || 'PartFlow';
  const installmentWhatsAppMessage = String(installmentWhatsAppMessageSetting?.data?.value ?? '');
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
  const displayCart = useMemo(() => cart.map((item) => {
    const price = taxExempt
      ? item.price
      : Math.round(item.price * (1 + effectiveTaxRate / 100) * 100) / 100;
    return { ...item, price, total: price * item.quantity };
  }), [cart, taxExempt, effectiveTaxRate]);
  const displaySubtotal = Math.round(subtotal * 100) / 100;
  const appliedDiscountRate = discountsEnabled
    ? Math.min(Math.max(Number(discountRate) || 0, 0), maxDiscountRate)
    : 0;
  const displayDiscount = Math.round(displaySubtotal * appliedDiscountRate) / 100;
  const displayTax = taxExempt
    ? 0
    : Math.round((displaySubtotal - displayDiscount) * effectiveTaxRate) / 100;
  const displayTotal = Math.max(0, Math.round((displaySubtotal - displayDiscount + displayTax) * 100) / 100);
  const {
    paymentMethod,
    setPaymentMethod,
    paidAmount,
    setPaidAmount,
    isProcessing,
    setProcessing,
    resetPayment,
  } = usePayment();
  const [installmentMonths, setInstallmentMonths] = useState(3);

  // Local state
  const [searchQuery, setSearchQuery] = useState('');
  const barcodeScanValueRef = useRef('');
  const barcodeScanTimerRef = useRef<number | null>(null);
  const processBarcodeValueRef = useRef<(value: string) => Promise<void>>(async () => undefined);
  const [posSection, setPosSection] = useState<'products'>('products');
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
  const [isSalesHistoryOpen, setIsSalesHistoryOpen] = useState(false);
  const [salesHistorySearch, setSalesHistorySearch] = useState('');
  const [showSalesCleanupConfirmation, setShowSalesCleanupConfirmation] = useState(false);
  const [salesCleanupConfirmationText, setSalesCleanupConfirmationText] = useState('');
  const [loadingHistoricalSaleId, setLoadingHistoricalSaleId] = useState<string | null>(null);
  const [lastSaleData, setLastSaleData] = useState<InvoiceData | null>(null);
  const lastSaleDataRef = useRef<InvoiceData | null>(null);
  const [isLoadingLastInvoice, setIsLoadingLastInvoice] = useState(false);
  const pendingExternalPaymentRef = useRef<string | null>(null);
  const [isQuickCustomerModalOpen, setIsQuickCustomerModalOpen] =
    useState(false);
  const [customerBalance, setCustomerBalance] = useState(0);
  const [customerCreditLimit, setCustomerCreditLimit] = useState<
    number | undefined
  >();
  const [paymentAllocations, setPaymentAllocations] = useState<PaymentAllocation[]>([]);
  const [isSplitPayment, setIsSplitPayment] = useState(false);
  const [paymentPanelVersion, setPaymentPanelVersion] = useState(0);
  const [quickCustomerName, setQuickCustomerName] = useState('');
  const [quickCustomerPhone, setQuickCustomerPhone] = useState('');
  const [productPage, setProductPage] = useState(1);
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
  const [isShiftModalOpen, setIsShiftModalOpen] = useState(false);
  const [shiftOpeningCash, setShiftOpeningCash] = useState('');
  const [shiftClosingCash, setShiftClosingCash] = useState('');
  const [shift, setShift] = useState<PosShiftState>(createDefaultPosShift);
  const [isBrowserOnline, setIsBrowserOnline] = useState(() =>
    typeof navigator === 'undefined' ? true : navigator.onLine,
  );
  const checkoutInFlightRef = useRef(false);
  const shiftActionInFlightRef = useRef(false);
  const checkoutMode = useUIStore((state) => state.checkoutMode);
  const setCheckoutMode = useUIStore((state) => state.setCheckoutMode);
  const initialCheckoutMode = useRef(checkoutMode);

  useEffect(() => {
    setCheckoutMode(true);

    return () => {
      setCheckoutMode(initialCheckoutMode.current);
    };
  }, [setCheckoutMode]);

  useEffect(() => {
    const updateOnlineStatus = () => setIsBrowserOnline(navigator.onLine);
    window.addEventListener('online', updateOnlineStatus);
    window.addEventListener('offline', updateOnlineStatus);
    return () => {
      window.removeEventListener('online', updateOnlineStatus);
      window.removeEventListener('offline', updateOnlineStatus);
    };
  }, []);

  // Held sales query
  const { data: heldSalesData, isLoading: heldSalesLoading, isError: heldSalesError, refetch: refetchHeldSales } = useQuery({
    queryKey: ['held-sales'],
    queryFn: () => salesApi.listHeld(),
    enabled: isHeldSalesOpen,
    staleTime: 15_000,
  });
  const {
    data: currentShiftData,
    isLoading: isShiftLoading,
    isError: isShiftError,
    isSuccess: isShiftLoaded,
    refetch: refetchShift,
  } = useQuery({
    queryKey: ['pos-shift', 'current'],
    queryFn: () => posShiftsApi.current(),
    retry: false,
    staleTime: 15_000,
  });

  useEffect(() => {
    if (!currentShiftData) return;
    const remoteShift = normalizeRemoteShift(currentShiftData.data);
    if (remoteShift) {
      setShift(remoteShift);
      return;
    }

    setShift(createDefaultPosShift());
  }, [currentShiftData]);
  const heldSales = (
    (
      (heldSalesData?.data as unknown) as Array<Record<string, unknown>> | undefined
    ) || []
  )
    .map((held): HeldSale | null => {
      let items: unknown;
      try {
        items = typeof held.items === 'string' ? JSON.parse(held.items) : held.items;
      } catch {
        return null;
      }
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
  const debouncedSalesHistorySearch = useDebounce(salesHistorySearch, 300);

  // Products query
  const {
    data: productsData,
    isLoading: productsLoading,
    isError: productsError,
    isSuccess: productsLoaded,
    refetch: refetchProducts,
  } = useQuery({
    queryKey: [
      'products',
      debouncedSearchQuery,
      selectedCategory,
      productPage,
      productsPerPage,
    ],
    queryFn: () => {
      return productsApi.list({
        page: productPage,
        per_page: productsPerPage,
        search: debouncedSearchQuery,
        category_id: selectedCategory || undefined,
        in_stock_only: true,
      });
    },
    enabled: posSection === 'products',
    staleTime: 30_000,
    refetchOnMount: 'always',
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
          is_active: true,
        });
      } else {
        return customersApi.list({ page: 1, per_page: 50, is_active: true });
      }
    },
    enabled: isCustomerMenuOpen,
    staleTime: 60_000,
  });

  // Other queries
  const {
    data: inventoryData,
    isError: inventoryError,
    isSuccess: inventoryLoaded,
    refetch: refetchInventory,
  } = useQuery({
    queryKey: ['inventory'],
    queryFn: () => inventoryApi.listWithSupplier({ page: 1, per_page: 1000 }),
    enabled: posSection === 'products',
    staleTime: 30_000,
    refetchOnMount: 'always',
    refetchOnWindowFocus: true,
  });

  useEffect(() => {
    setOptimisticSoldQuantities({});
  }, [inventoryData]);

  const { data: salesHistoryData, isLoading: salesHistoryLoading, isError: salesHistoryError, refetch: refetchSalesHistory } = useQuery({
    queryKey: ['sales', 'pos-history', debouncedSalesHistorySearch],
    queryFn: () => salesApi.list({
      page: 1,
      per_page: 20,
      search: debouncedSalesHistorySearch.trim() || undefined,
    }),
    enabled: isSalesHistoryOpen,
    staleTime: 15_000,
  });
  const cleanSalesHistoryMutation = useMutation({
    mutationFn: async () => {
      const response = await salesApi.cleanHistory();
      return response?.data ?? response;
    },
    onSuccess: async ({ deleted, blocked, total, failed }) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['sales'] }),
        queryClient.invalidateQueries({ queryKey: ['sales-available-for-return'] }),
        queryClient.invalidateQueries({ queryKey: ['sales-returns-analysis'] }),
        queryClient.invalidateQueries({ queryKey: ['inventory'] }),
        queryClient.invalidateQueries({ queryKey: ['products'] }),
        queryClient.invalidateQueries({ queryKey: ['customers'] }),
        queryClient.invalidateQueries({ queryKey: ['debts'] }),
        queryClient.invalidateQueries({ queryKey: ['overdue-debts'] }),
        queryClient.invalidateQueries({ queryKey: ['customer-purchases-debts'] }),
        queryClient.invalidateQueries({ queryKey: ['dashboard'] }),
        queryClient.invalidateQueries({ queryKey: ['dashboard-activity'] }),
        queryClient.invalidateQueries({ queryKey: ['reports'] }),
      ]);
      setShowSalesCleanupConfirmation(false);
      setSalesCleanupConfirmationText('');
      setSalesHistorySearch('');
      if (deleted === 0 && total === 0) {
        toast.info('لا توجد سجلات مبيعات لتنظيفها.', 'السجل فارغ');
      } else if (failed > 0) {
        toast.error(`حُذف ${deleted} سجلًا وتعذّر حذف ${blocked + failed} سجلًا. راجع هذه السجلات يدويًا.`, 'اكتمل التنظيف جزئيًا', 7000);
      } else {
        toast.success(`حُذف ${deleted} سجل مبيعات، وتعذّر حذف ${blocked} سجلًا بسبب ارتباط يمنع الحذف أو لأنه لم يعد موجودًا.`, 'اكتمل تنظيف السجل');
      }
    },
    onError: (error: any) => {
      setShowSalesCleanupConfirmation(true);
      toast.error(error instanceof Error ? error.message : 'تعذر تنظيف سجل المبيعات. حاول مجددًا.', 'فشل تنظيف المبيعات', 7000);
    },
  });

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });

  const products = useMemo(
    () => ((productsData?.data?.products as unknown) as Product[]) || [],
    [productsData],
  );
  const productTotal = productsData?.data?.total ?? products.length;
  const customers = useMemo(
    () => ((customersData?.data as unknown) as Customer[]) || [],
    [customersData],
  );
  const inventoryItems = useMemo(
    () => ((inventoryData?.data?.items as unknown) as InventoryItem[]) || [],
    [inventoryData],
  );
  const categories = useMemo(
    () => ((categoriesData?.data as unknown) as Category[]) || [],
    [categoriesData],
  );
  const isShiftOpen = isShiftLoaded && !isShiftLoading && !isShiftError && shift.status === 'open';
  const shiftStatusLabel = isShiftLoading
    ? 'جارٍ التحقق من الوردية'
    : isShiftError
      ? 'تعذر التحقق من الوردية'
      : isShiftOpen
        ? 'الوردية مفتوحة'
        : 'الوردية مغلقة';
  const connectionStatus = !isBrowserOnline || productsError || inventoryError || isShiftError
    ? 'offline'
    : productsLoaded && inventoryLoaded && isShiftLoaded
      ? 'online'
      : 'connecting';
  const salesHistoryPayload = salesHistoryData?.data ?? salesHistoryData;
  const historicalSales = (
    Array.isArray(salesHistoryPayload)
      ? salesHistoryPayload
      : salesHistoryPayload?.sales ?? salesHistoryPayload?.items ?? salesHistoryPayload?.data ?? []
  ) as Array<Record<string, any>>;
  const salesHistoryResultCount = Number(salesHistoryPayload?.total ?? historicalSales.length);

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
        sku: itemCode || productId,
        barcode: barcode || undefined,
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

  const cartQuantities = useMemo(() => {
    return cart.reduce<Record<string, number>>((quantities, item) => {
      const productId = String(item.id);
      quantities[productId] = (quantities[productId] ?? 0) + item.quantity;
      return quantities;
    }, {});
  }, [cart]);

  const productsWithStock = useMemo(() => {
    const stockMap = new Map<string, number>();
    const catalogProductIds = new Set(products.map((product) => String(product.id)));

    inventoryItems.forEach((item: any) => {
      const productId = String(item.product_id || item.product?.id || '').trim();
      if (!productId) return;

      const status = String(item.status || '').trim().toUpperCase();
      const condition = String(item.condition || '').trim().toUpperCase();
      if (condition === 'USED') return;

      const explicitStock = Number(
        item.available_quantity ??
        item.current_quantity ??
        item.stock ??
        item.quantity ??
        0
      );

      const unavailableStatus = ['SOLD', 'RESERVED', 'DAMAGED', 'IN_REPAIR', 'RETURNED', 'FOR_PARTS', 'ARCHIVED'].includes(status);
      if (unavailableStatus && !(Number.isFinite(explicitStock) && explicitStock > 0)) return;

      const calculatedStock = Number.isFinite(explicitStock) && explicitStock > 0 ? explicitStock : (status === 'AVAILABLE' ? 1 : 0);
      if (calculatedStock <= 0) return;

      const current = stockMap.get(productId) ?? 0;
      stockMap.set(productId, Math.max(current, calculatedStock));
    });

    return productsWithInventoryFallback.filter((product) => {
      const productId = String(product.id);
      if (catalogProductIds.has(productId)) {
        // The products API computes this from the authoritative inventory summary.
        // A paginated/stale inventory-items response must not hide catalog stock.
        return Number((product as any).current_quantity ?? 0) > 0;
      }
      const declaredStock = Number(
        (product as any).current_quantity ??
        (product as any).stock ??
        (product as any).stock_quantity ??
        0,
      );
      return declaredStock > 0 || Number(stockMap.get(productId) ?? 0) > 0;
    }).map((product) => {
      const rawProduct = product as Product & {
        categoryId?: string;
        category?: { id?: string };
      };
      const categoryId = rawProduct.category_id || rawProduct.categoryId || rawProduct.category?.id;
      const productId = String(product.id);
      const isCatalogProduct = catalogProductIds.has(productId);
      const baseStock = isCatalogProduct
        ? Number((product as any).current_quantity ?? 0)
        : Number(
          (product as any).current_quantity ??
          (product as any).stock ??
          (product as any).stock_quantity ??
          stockMap.get(productId) ??
          0,
        );
      const stock = Math.max(
        0,
        baseStock -
          Number(optimisticSoldQuantities[productId] ?? 0) -
          Number(cartQuantities[productId] ?? 0),
      );
      return {
        ...product,
        stock,
        current_quantity: baseStock,
        category_name: categories.find((category) => String(category.id) === String(categoryId))?.name,
        category_image_url: categoryId ? getCategoryImage(String(categoryId)) : undefined,
      };
    }).filter((product) => Number(product.stock ?? 0) > 0);
  }, [productsWithInventoryFallback, products, inventoryItems, categories, optimisticSoldQuantities, cartQuantities]);

  const visibleProducts = useMemo(
    () => productsWithStock.slice(0, productsPerPage),
    [productsPerPage, productsWithStock]
  );

  const getAvailableStockCount = useCallback(
    (productId: string) => {
      if (!productId) return 0;
      const productInventoryItems = inventoryItems.filter((item) => String(item.product_id) === String(productId));
      if (productInventoryItems.length === 0) {
        const product = productsWithStock.find((item) => String(item.id) === String(productId))
          ?? products.find((item) => String(item.id) === String(productId));
        return Number((product as any)?.current_quantity ?? product?.stock ?? 0);
      }

      const reportedQuantities = productInventoryItems
        .map((item: any) => Number(item.available_quantity ?? item.current_quantity))
        .filter((quantity: number) => Number.isFinite(quantity));
      if (reportedQuantities.length > 0) {
        return Math.max(...reportedQuantities, 0);
      }

      return productInventoryItems.filter((item) => {
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
    [inventoryItems, products, productsWithStock]
  );

  const canAddProductToCart = useCallback(
    (productId: string, productName: string, quantity: number = 1, reportedStock?: number) => {
      const currentQuantity = cart
        .filter((item) => String(item.id) === String(productId))
        .reduce((sum, item) => sum + item.quantity, 0);
      // Product lookup can return the current server stock before the inventory page is cached.
      const available = Math.max(getAvailableStockCount(productId), Number(reportedStock) || 0);

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
    mutationFn: async ({ data, idempotencyKey }: { data: SaleCreateRequest; idempotencyKey: string }) => {
      return salesApi.create(data, idempotencyKey);
    },
    onSuccess: (response) => {
      checkoutInFlightRef.current = false;
      clearPosCheckoutAttempt();
      setShift((current) => current.status === 'open'
        ? { ...current, salesTotal: current.salesTotal + displayTotal, saleCount: current.saleCount + 1 }
        : current);
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
      }
      queryClient.invalidateQueries({ queryKey: ['sales'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      queryClient.invalidateQueries({ queryKey: ['debts'] });
      queryClient.invalidateQueries({ queryKey: ['reports'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['products'] });

      if (lastSaleDataRef.current) {
        const sale = response?.data?.sale ?? response?.data ?? response?.sale;
        const persistedAllocations = response?.data?.payment_allocations ?? sale?.payment_allocations;
        setLastSaleData({
          ...lastSaleDataRef.current,
          id: sale?.id || response?.id || lastSaleDataRef.current.id,
          invoiceNumber: sale?.invoice_number || response?.invoice_number || lastSaleDataRef.current.invoiceNumber,
          saleDate: sale?.created_at || sale?.sale_date || lastSaleDataRef.current.saleDate,
          subtotal: Number(sale?.subtotal ?? lastSaleDataRef.current.subtotal),
          discountAmount: Number(sale?.discount_amount ?? lastSaleDataRef.current.discountAmount ?? 0),
          taxAmount: Number(sale?.tax_amount ?? lastSaleDataRef.current.taxAmount ?? 0),
          paidAmount: Number(sale?.paid_amount ?? lastSaleDataRef.current.paidAmount),
          remaining: Number(sale?.remaining_amount ?? lastSaleDataRef.current.remaining),
          paymentStatus: sale?.payment_status || lastSaleDataRef.current.paymentStatus,
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
      setSelectedCustomerOption(undefined);
      setCustomerBalance(0);
      setCustomerCreditLimit(undefined);
      setPaymentAllocations([]);
      setIsSplitPayment(false);
      setPaymentPanelVersion((version) => version + 1);
      setTaxExempt(false);
      resetPayment();
      setProcessing(false);
    },
    onError: (error: unknown) => {
      checkoutInFlightRef.current = false;
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

  const openShiftMutation = useMutation({
    mutationFn: (amount: number) => posShiftsApi.open(amount),
    onSuccess: (response) => {
      const remoteShift = normalizeRemoteShift(response?.data?.shift ?? response?.data ?? response);
      if (!remoteShift || remoteShift.status !== 'open') {
        toast.error('لم يتأكد الخادم من فتح الوردية. حدّث الحالة وحاول مجددًا.', 'تعذر فتح الوردية');
        return;
      }
      setShift(remoteShift);
      setShiftOpeningCash('');
      queryClient.setQueryData(['pos-shift', 'current'], { data: remoteShift });
      toast.success('تم فتح الوردية بنجاح', 'الوردية مفتوحة');
    },
    onError: (error: unknown) => {
      const message = (error as { message?: string })?.message;
      toast.error(message || 'تعذر فتح الوردية. تحقق من الاتصال وحاول مجددًا.', 'فشل فتح الوردية');
    },
    onSettled: () => {
      shiftActionInFlightRef.current = false;
      void queryClient.invalidateQueries({ queryKey: ['pos-shift', 'current'] });
    },
  });

  const closeShiftMutation = useMutation({
    mutationFn: (amount: number) => posShiftsApi.close(amount),
    onSuccess: (response) => {
      const remoteShift = normalizeRemoteShift(response?.data?.shift ?? response?.data ?? response);
      const closedShift = remoteShift?.status === 'closed'
        ? remoteShift
        : { ...shift, status: 'closed' as const, closedAt: new Date().toISOString(), closingCash: Number(shiftClosingCash) || 0 };
      setShift(closedShift);
      setShiftClosingCash('');
      queryClient.setQueryData(['pos-shift', 'current'], { data: closedShift });
      toast.success('تم إغلاق الوردية وتسوية الصندوق', 'تم الإغلاق');
    },
    onError: (error: unknown) => {
      const message = (error as { message?: string })?.message;
      toast.error(message || 'تعذر إغلاق الوردية. تحقق من الاتصال وحاول مجددًا.', 'فشل إغلاق الوردية');
    },
    onSettled: () => {
      shiftActionInFlightRef.current = false;
      void queryClient.invalidateQueries({ queryKey: ['pos-shift', 'current'] });
    },
  });

  const handlePrintLastInvoice = async () => {
    if (isLoadingLastInvoice) return;

    setIsLoadingLastInvoice(true);
    try {
      const listResponse = await salesApi.list({ page: 1, per_page: 1 });
      const listPayload = listResponse?.data ?? listResponse;
      const sales = Array.isArray(listPayload) ? listPayload : listPayload?.sales ?? listPayload?.data ?? [];
      const latestSale = sales[0];

      if (!latestSale?.id) {
        toast.error('لا توجد فاتورة مبيعات للطباعة.', 'لا توجد فاتورة', 4000);
        return;
      }

      const detailsResponse = await salesApi.get(String(latestSale.id));
      const invoiceData = buildInvoiceFromSale(latestSale, detailsResponse);

      setLastSaleData(invoiceData);
      lastSaleDataRef.current = invoiceData;
      setIsInvoiceModalOpen(true);
    } catch (error) {
      console.error('Failed to load latest sale invoice:', error);
      toast.error('تعذر تحميل آخر فاتورة. تحقق من الاتصال وحاول مرة أخرى.', 'فشل تحميل الفاتورة', 5000);
    } finally {
      setIsLoadingLastInvoice(false);
    }
  };

  const handleOpenHistoricalInvoice = async (sale: Record<string, any>) => {
    const saleId = String(sale.id ?? '').trim();
    if (!saleId || loadingHistoricalSaleId) return;
    setLoadingHistoricalSaleId(saleId);
    try {
      const detailsResponse = await salesApi.get(saleId);
      const invoiceData = buildInvoiceFromSale(sale, detailsResponse);
      setLastSaleData(invoiceData);
      lastSaleDataRef.current = invoiceData;
      setIsSalesHistoryOpen(false);
      setIsInvoiceModalOpen(true);
    } catch (error) {
      console.error('Failed to load historical sale invoice:', error);
      toast.error('تعذر تحميل الفاتورة السابقة. تحقق من الاتصال وحاول مجددًا.', 'فشل تحميل الفاتورة', 5000);
    } finally {
      setLoadingHistoricalSaleId(null);
    }
  };

  // Handlers
  const handleClearSearch = () => {
    setSearchQuery('');
    setProductPage(1);
  };

  const handleHoldSale = useCallback(() => {
    if (cart.length === 0) {
      toast.error('أضف منتجًا إلى السلة قبل تعليق البيع', 'السلة فارغة');
      return;
    }
    holdSaleMutation.mutate();
  }, [cart.length, holdSaleMutation, toast]);

  // Register cashier shortcuts after handleHoldSale is initialized.
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'F4') {
        event.preventDefault();
        handleHoldSale();
      } else if (event.key === 'F2') {
        event.preventDefault();
        document.querySelector<HTMLInputElement>('.pos-barcode-input')?.focus();
      } else if (event.key === 'F8') {
        event.preventDefault();
        document.querySelector<HTMLButtonElement>('.pos-advanced-payment-panel .checkout-btn:not(:disabled)')?.click();
      } else if (event.key === 'Escape') {
        setUnknownBarcode('');
        setIsManualProductOpen(false);
      } else if (event.key === 'Delete' && cart.length > 0) {
        if (!(event.target as HTMLElement).matches('input, textarea, select')) {
          removeFromCart(cart[cart.length - 1].barcode);
        }
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [cart, handleHoldSale, removeFromCart]);

  const handleResumeSale = (held: HeldSale) => {
    if (!held) return;
    clearCart();
    held.items.forEach((item) => addToCart(item, item.quantity));
    deleteHeldSaleMutation.mutate(held.id);
    setIsHeldSalesOpen(false);
  };

  const processBarcodeValue = async (barcodeValue: string) => {
    const normalizedBarcode = barcodeValue.trim();
    if (!normalizedBarcode) return;

    try {
      const response = await barcodeApi.lookupProduct(normalizedBarcode);
      let product = response.data as Product & { inventory_item_id?: string; serial_number?: string; inventory_item?: any; stock?: number };

      // Barcode resolution identifies the catalog record, but may not include its
      // aggregate stock. Fetch that count only when the lookup omitted it.
      if (product?.id && product.current_quantity == null && product.stock == null) {
        try {
          const catalogResponse = await productsApi.list({ page: 1, per_page: 20, search: normalizedBarcode });
          const catalogProducts = catalogResponse?.data?.products as Product[] | undefined;
          const normalizedCode = normalizedBarcode.toLocaleLowerCase();
          const catalogProduct = catalogProducts?.find((candidate) =>
            String(candidate.barcode || '').toLocaleLowerCase() === normalizedCode,
          );
          if (catalogProduct) {
            product = {
              ...catalogProduct,
              ...product,
              current_quantity: Number(catalogProduct.current_quantity ?? 0),
              stock: Number(catalogProduct.current_quantity ?? 0),
            };
          }
        } catch {
          // Continue with the barcode result; checkout still performs a fresh stock check.
        }
      }

      if (product && product.id) {
        const resolvedInventoryItem = product.inventory_item;
        const resolvedStatus = String(resolvedInventoryItem?.status ?? '').trim().toUpperCase();
        const resolvedCondition = String(resolvedInventoryItem?.condition ?? product.condition ?? '').trim().toUpperCase();
        if (resolvedInventoryItem && resolvedStatus !== 'AVAILABLE') {
          if (soundEnabled) playScanSound(false);
          setUnknownBarcode(normalizedBarcode);
          setSearchQuery('');
          return;
        }
        if (resolvedCondition === 'USED') {
          if (soundEnabled) playScanSound(false);
          setUnknownBarcode(normalizedBarcode);
          setSearchQuery('');
          return;
        }

        const reportedStock = Number(product.current_quantity ?? product.stock ?? 0);
        if (!canAddProductToCart(product.id, product.name, 1, reportedStock)) {
          navigate('/app/inventory', {
            state: {
              editProduct: {
                id: product.id,
                name: product.name,
                sku: product.sku,
                barcode: normalizedBarcode,
                sellingPrice: normalizePosPrice(product.sellingPrice, product.selling_price),
                costPrice: Number(product.costPrice ?? product.cost_price ?? 0),
                stock: Number(product.stock ?? product.current_quantity ?? 0),
                condition: 'new',
              },
              returnToSalesAfterSave: true,
            },
          });
          toast.info('تم فتح المنتج في المخزون لزيادة الكمية', 'المخزون نفد');
          return;
        }
        addToCart({
          id: product.id,
          inventoryItemId: product.inventory_item_id || product.inventory_item?.id,
          serialNumber: product.serial_number || product.inventory_item?.serial_number,
          name: product.name,
          barcode: normalizedBarcode,
          price: normalizePosPrice(
            product.sellingPrice,
            (product as Product & { selling_price?: number }).selling_price
          ),
          stock: reportedStock,
          condition: product.condition,
          purchaseCost: product.costPrice || product.cost_price || 0,
        });
        setSearchQuery('');
      } else {
        if (soundEnabled) playScanSound(false);
        setUnknownBarcode(normalizedBarcode);
        setSearchQuery('');
      }
    } catch {
      if (soundEnabled) playScanSound(false);
      const product = products?.find(
        (p) =>
          p.barcode === normalizedBarcode
      );
      if (product) {
        if (!canAddProductToCart(
          product.id,
          product.name,
          1,
          Number(product.current_quantity ?? product.stock ?? 0),
        )) return;
        addToCart({
          id: product.id,
          name: product.name,
          barcode: normalizedBarcode,
          price: normalizePosPrice(
            product.sellingPrice,
            (product as Product & { selling_price?: number }).selling_price
          ),
          stock: Number(product.current_quantity ?? product.stock ?? 0),
          condition: product.condition,
          purchaseCost: product.cost_price || product.costPrice || 0,
        });
        setSearchQuery('');
      } else {
        setUnknownBarcode(normalizedBarcode);
        setSearchQuery('');
      }
    }
  };

  processBarcodeValueRef.current = processBarcodeValue;

  const handleBarcodeScan = (e: React.FormEvent) => {
    e.preventDefault();
    const value = searchQuery.trim();
    if (!value) return;

    const looksLikeScannedCode = /^[a-z0-9-]{8,}$/i.test(value) && /\d/.test(value);
    if (looksLikeScannedCode) {
      void processBarcodeValue(value);
      return;
    }

    setProductPage(1);
  };

  useEffect(() => {
    const handleScannerKeyDown = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement | null;
      if (target && (target.matches('input, textarea, select, [contenteditable="true"]') || target.isContentEditable)) return;

      if (event.key === 'Enter') {
        const scannedValue = barcodeScanValueRef.current;
        barcodeScanValueRef.current = '';
        if (barcodeScanTimerRef.current !== null) window.clearTimeout(barcodeScanTimerRef.current);
        barcodeScanTimerRef.current = null;
        if (scannedValue.length >= 4) {
          event.preventDefault();
          void processBarcodeValueRef.current(scannedValue);
        }
        return;
      }

      if (event.key.length !== 1 || event.ctrlKey || event.altKey || event.metaKey) return;
      barcodeScanValueRef.current += event.key;
      if (barcodeScanTimerRef.current !== null) window.clearTimeout(barcodeScanTimerRef.current);
      barcodeScanTimerRef.current = window.setTimeout(() => {
        barcodeScanValueRef.current = '';
        barcodeScanTimerRef.current = null;
      }, 100);
    };

    window.addEventListener('keydown', handleScannerKeyDown);
    return () => {
      window.removeEventListener('keydown', handleScannerKeyDown);
      if (barcodeScanTimerRef.current !== null) window.clearTimeout(barcodeScanTimerRef.current);
    };
  }, []);

  const handleProductSelect = useCallback(
    (product: PosCartProduct) => {
      if (!canAddProductToCart(product.id, product.name, 1, Number(product.current_quantity ?? product.stock ?? 0))) return;
      addToCart(product);
    },
    [addToCart, canAddProductToCart]
  );

  const handleCheckout = useCallback(async () => {
    if (checkoutInFlightRef.current) return;
    if (cart.length === 0) return;
    if (!isShiftOpen) {
      toast.error('افتح الوردية أولًا قبل إتمام البيع', 'الوردية مغلقة');
      return;
    }
    if (!isSplitPayment && paymentMethod === 'credit' && paidAmount.trim() === '') return;
    const splitPaymentAmount = paymentAllocations.reduce((sum, allocation) => sum + Number(allocation.amount || 0), 0);
    const splitPaymentCents = Math.round(splitPaymentAmount * 100);
    const saleTotalCents = Math.round(displayTotal * 100);
    if (isSplitPayment && splitPaymentCents !== saleTotalCents) {
      toast.error(
        splitPaymentCents > saleTotalCents
          ? 'مجموع الدفعات يتجاوز إجمالي الفاتورة.'
          : 'أكمل توزيع إجمالي الفاتورة قبل إتمام البيع.',
        'راجع تقسيم الدفع',
      );
      return;
    }

    checkoutInFlightRef.current = true;
    setProcessing(true);
    const exhaustedItems = cart.reduce((items, item) => {
      // Use the stock snapshot already attached when adding the item. The sale API
      // remains the authoritative transactional check if inventory changed since.
      const availableStock = Number(item.stock);
      if (!item.inventoryItemId && Number.isFinite(availableStock) && item.quantity > availableStock) {
        items.push(item.name);
      }
      return items;
    }, [] as string[]);

    if (exhaustedItems.length > 0) {
      checkoutInFlightRef.current = false;
      setProcessing(false);
      const uniqueItems = [...new Set(exhaustedItems)];
      const message =
        uniqueItems.length > 1
          ? `هذه الأنواع قد نفذت من المخزون: ${uniqueItems.slice(0, 2).join(', ')}`
          : `هذا النوع قد نفذ من المخزون: ${uniqueItems[0]}`;
      toast.error(message, 'مخزون غير كافٍ', 4000);
      return;
    }

    const checkoutAttempt = getPosCheckoutAttempt();
    let paymentTransactionID: string | undefined;
    if (!isSplitPayment && paymentMethod === 'card' && electronicPaymentsEnabled && electronicPaymentProvider !== 'manual') {
      const orderID = checkoutAttempt.paymentOrderId;
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
        checkoutInFlightRef.current = false;
        toast.error(error instanceof Error ? error.message : 'فشل الدفع الإلكتروني', 'فشل الدفع', 5000);
        setProcessing(false);
        return;
      }
    }

    const receivedAmount = parseFloat(paidAmount) || 0;
    const splitPaymentMethods = [...new Set(paymentAllocations.map((allocation) => allocation.method))];
    const splitPrimaryMethod = splitPaymentMethods.length === 1
      ? splitPaymentMethods[0]
      : paymentAllocations.find((allocation) => allocation.method !== 'cash')?.method ?? 'cash';
    const effectivePaymentMethod = isSplitPayment ? splitPrimaryMethod : paymentMethod;
    const appliedPaymentAmount = isSplitPayment
      ? splitPaymentAmount
      : effectivePaymentMethod === 'cash'
      ? Math.min(receivedAmount, displayTotal)
      : receivedAmount;
    const splitCashReceived = isSplitPayment
      ? paymentAllocations
          .filter((allocation) => allocation.method === 'cash')
          .reduce((sum, allocation) => sum + Number(allocation.amount || 0), 0)
      : undefined;
    const cashReceived = isSplitPayment ? splitCashReceived : effectivePaymentMethod === 'cash' ? receivedAmount : undefined;
    const changeAmount = !isSplitPayment && effectivePaymentMethod === 'cash'
      ? Math.max(0, receivedAmount - displayTotal)
      : 0;
    const saleData: SaleCreateRequest = {
      customer_id: selectedCustomer || undefined,
      items: cart.map((item) => ({
        product_id: item.id,
        quantity: item.quantity,
        unit_price: item.price,
        ...(item.operationalBarcode ? { barcode: item.operationalBarcode } : {}),
        ...(item.inventoryItemId ? { inventory_item_id: item.inventoryItemId } : {}),
        ...(item.purchaseCost ? { purchase_cost: item.purchaseCost } : {}),
      })),
      payment_method: effectivePaymentMethod === 'credit' ? 'debt' : effectivePaymentMethod === 'checks' ? 'transfer' : effectivePaymentMethod,
      payment_amount: appliedPaymentAmount,
      ...(!isSplitPayment && effectivePaymentMethod === 'installment' ? { installment_months: installmentMonths } : {}),
      ...(paymentAllocations.length > 0 && (isSplitPayment || effectivePaymentMethod !== 'cash') ? { payment_allocations: paymentAllocations } : {}),
      ...(paymentTransactionID ? { payment_transaction_id: paymentTransactionID } : {}),
      ...(cashReceived !== undefined ? { cash_received: cashReceived } : {}),
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
        ...(item.operationalBarcode ? { barcode: item.operationalBarcode } : {}),
        sellingPrice: item.price,
        quantity: item.quantity,
        total: item.total,
      })),
      subtotal: displaySubtotal,
      discountAmount: displayDiscount,
      taxAmount: displayTax,
      total: displayTotal,
      paidAmount: appliedPaymentAmount,
      cashReceived,
      changeAmount,
      remaining: Math.max(0, displayTotal - appliedPaymentAmount),
      paymentMethod: effectivePaymentMethod,
      paymentAllocations,
    };

    setLastSaleData(invoiceData);
    lastSaleDataRef.current = invoiceData;
    createSaleMutation.mutate({ data: saleData, idempotencyKey: checkoutAttempt.idempotencyKey });
  }, [
    cart,
    selectedCustomer,
    paymentMethod,
    installmentMonths,
    paidAmount,
    displayCart,
    displaySubtotal,
    displayDiscount,
    displayTax,
    displayTotal,
    appliedDiscountRate,
    taxExempt,
    paymentAllocations,
    isSplitPayment,
    electronicPaymentsEnabled,
    electronicPaymentProvider,
    customers,
    selectedCustomerOption?.name,
    setProcessing,
    createSaleMutation,
    toast,
    isShiftOpen,
    checkoutInFlightRef,
  ]);

  // Quick customer creation handler
  const handleQuickCustomerCreate = useCallback(() => {
    setIsQuickCustomerModalOpen(true)
  }, [])

  // Reuse the regular checkout path so payment rules and stock validation stay identical.
  const handleQuickSale = useCallback(() => {
    document.querySelector<HTMLButtonElement>(
      '.pos-advanced-payment-panel .checkout-btn:not(:disabled)',
    )?.click();
  }, []);

  const handlePaymentAllocationsChange = useCallback((allocations: PaymentAllocation[], splitMode: boolean) => {
    setPaymentAllocations(allocations);
    setIsSplitPayment(splitMode);
  }, []);
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
      } else {
        setSelectedCustomerOption(undefined);
        setCustomerBalance(0);
        setCustomerCreditLimit(undefined);
      }
    } else {
      setSelectedCustomerOption(undefined);
      setCustomerBalance(0);
      setCustomerCreditLimit(undefined);
    }
  };

  const handleNewSale = () => {
    if (checkoutInFlightRef.current) return;
    if (cart.length > 0 && !window.confirm('سيتم مسح الفاتورة الحالية وبدء فاتورة جديدة. هل تريد المتابعة؟')) {
      return;
    }

    clearCart();
    resetPayment();
    setPaymentAllocations([]);
    setIsSplitPayment(false);
    setPaymentPanelVersion((version) => version + 1);
    setSelectedCustomer('');
    setSelectedCustomerOption(undefined);
    setDiscountRate(0);
    setTaxExempt(false);
  };

  const handleOpenShift = () => {
    if (shiftActionInFlightRef.current || openShiftMutation.isPending) return;
    const openingCash = Math.max(0, Number(shiftOpeningCash) || 0);
    if (isShiftLoading || isShiftError) {
      toast.error('تعذر التحقق من حالة الوردية. تحقق من الاتصال ثم أعد المحاولة.', 'حالة الوردية غير متاحة');
      return;
    }
    shiftActionInFlightRef.current = true;
    openShiftMutation.mutate(openingCash);
  };

  const handleCloseShift = () => {
    if (shiftActionInFlightRef.current || closeShiftMutation.isPending) return;
    const closingCash = Math.max(0, Number(shiftClosingCash) || 0);
    if (!isShiftOpen) {
      toast.error('لا توجد وردية مفتوحة لإغلاقها.', 'الوردية مغلقة');
      return;
    }
    shiftActionInFlightRef.current = true;
    closeShiftMutation.mutate(closingCash);
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
              className="pos-badge cart-count"
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
            <button
              type="button"
              className="pos-badge cashier"
              onClick={() => setIsShiftModalOpen(true)}
              title="إدارة الوردية"
            >
              {isShiftLoading || isShiftError
                ? <WifiOff className="w-3.5 h-3.5 text-amber-500" />
                : isShiftOpen
                ? <UnlockKeyhole className="w-3.5 h-3.5 text-green-500" />
                : <LockKeyhole className="w-3.5 h-3.5 text-amber-500" />}
              <span>{shiftStatusLabel}</span>
            </button>
                <span className="pos-badge time">
              {new Intl.DateTimeFormat('ar', {
                hour: '2-digit',
                minute: '2-digit',
                timeZone: 'Asia/Jerusalem',
              }).format(new Date())}
            </span>
            <span className={`pos-api-status ${connectionStatus}`} role="status" aria-live="polite">
              {connectionStatus === 'offline'
                ? <WifiOff className="h-3.5 w-3.5" aria-hidden="true" />
                : <Wifi className="h-3.5 w-3.5" aria-hidden="true" />}
              <span>{connectionStatus === 'online' ? 'متصل' : connectionStatus === 'connecting' ? 'جارٍ الاتصال' : 'الاتصال متعذر'}</span>
            </span>
          </div>
          <div className="pos-header-actions">
            {/* Quick Sale Button - appears when cart has items */}
            {cart.length > 0 && (
              <Button
                variant="primary"
                size="sm"
                onClick={handleQuickSale}
                disabled={isProcessing || !isShiftOpen}
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
              onClick={() => { setSalesHistorySearch(''); setIsSalesHistoryOpen(true); }}
              className="gap-2"
              title="البحث في المبيعات السابقة وإعادة طباعة الفاتورة"
            >
              <History className="w-4 h-4" />
              <span>المبيعات السابقة</span>
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => navigate('/app/returns')}
              className="gap-2"
              title="فتح المرتجعات"
            >
              <RotateCcw className="w-4 h-4" />
              <span>المرتجعات</span>
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={handlePrintLastInvoice}
              disabled={isLoadingLastInvoice}
              className="gap-2"
              title="طباعة آخر فاتورة مباعة"
            >
              <Printer className="w-4 h-4" />
              <span>{isLoadingLastInvoice ? 'جاري التحميل...' : 'طباعة آخر فاتورة'}</span>
            </Button>
          </div>
        </div>
      </header>

      {/* Main POS Body */}
      <div className="pos-modern-body">
        {/* Products Section */}
        <main className="pos-products-area">
          <div className="pos-category-bar" role="group" aria-label="قسم البيع">
            <button
              type="button"
              aria-pressed={posSection === 'products'}
              className={`category-chip ${posSection === 'products' ? 'active' : ''}`}
              onClick={() => setPosSection('products')}
            >
              <span>المنتجات</span>
            </button>
          </div>

          {/* Search & Barcode */}
          <form onSubmit={handleBarcodeScan} className="pos-search-bar">
            <div className="pos-barcode-form">
              <ScanLine className="barcode-icon" />
              <Input
                placeholder="امسح الباركود أو اكتب للبحث..."
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value);
                  setProductPage(1);
                }}
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

          {(productsError || inventoryError) && (
            <div className="pos-data-warning" role="alert">
              <span>{productsError ? 'تعذر تحديث المنتجات.' : 'تعذر تحديث بيانات المخزون؛ اعرض البيانات المتاحة وأعد المحاولة.'}</span>
              <button
                type="button"
                onClick={() => {
                  if (productsError) void refetchProducts();
                  if (inventoryError) void refetchInventory();
                }}
              >
                <RefreshCw className="h-3.5 w-3.5" aria-hidden="true" />
                <span>إعادة المحاولة</span>
              </button>
            </div>
          )}

          {posSection === 'products' && (
            <div className="pos-category-bar">
              <button
                type="button"
                className={`category-chip ${selectedCategory === null ? 'active' : ''}`}
                aria-pressed={selectedCategory === null}
                onClick={() => { setSelectedCategory(null); setProductPage(1); }}
              >
                <Package className="w-3.5 h-3.5" />
                <span>الكل</span>
              </button>
              {categories.map((cat) => (
                <button
                  type="button"
                  key={cat.id}
                  className={`category-chip ${selectedCategory === cat.id ? 'active' : ''}`}
                  aria-pressed={selectedCategory === cat.id}
                  onClick={() => { setSelectedCategory(cat.id); setProductPage(1); }}
                >
                  {cat.name}
                </button>
              ))}
            </div>
          )}

          {/* Products Grid */}
          {posSection === 'products' && <ModernProductGrid
              products={visibleProducts}
              onProductClick={handleProductSelect}
              taxRate={effectiveTaxRate}
              taxExempt={taxExempt}
              showDetails={false}
              currentPage={productPage}
              totalPages={Math.max(1, Math.ceil(productTotal / productsPerPage))}
              viewMode={productViewMode}
              onPreviousPage={() => setProductPage((page) => Math.max(1, page - 1))}
              onNextPage={() => setProductPage((page) => page + 1)}
              isLoading={productsLoading}
              hasSearch={Boolean(debouncedSearchQuery.trim())}
              onAddProduct={() => navigate('/app/inventory')}
            />}
        </main>

        {/* Cart and payment workspace */}
        <div className="pos-checkout-layout">
          <section className="pos-cart-column" aria-label="سلة البيع">
            <ModernCartPanel
              cart={displayCart}
              total={displayTotal}
              showTotals={false}
              onUpdateQuantity={updateQuantity}
              onRemoveFromCart={removeFromCart}
              onClearCart={() => {
                clearCart();
                resetPayment();
              }}
            />
          </section>

          <aside className="pos-payment-sidebar" aria-label="ملخص البيع والدفع">
            <div className="pos-payment-context">
            <PosCustomerSelector
              customers={customers}
              selectedCustomer={selectedCustomer}
              selectedCustomerOption={selectedCustomerOption}
              searchQuery={customerSearchQuery}
              isLoading={customersLoading}
              isOpen={isCustomerMenuOpen}
              onSearchChange={setCustomerSearchQuery}
              onSelect={(customerId) => {
                handleCustomerChange(customerId);
                setCustomerSearchQuery('');
                setIsCustomerMenuOpen(false);
              }}
              onToggle={() => setIsCustomerMenuOpen((open) => !open)}
              onCreate={handleQuickCustomerCreate}
            />

            <details className="pos-invoice-options">
              <summary>
                <span>خيارات الفاتورة</span>
                {(taxExempt || discountRate > 0) && <span className="pos-invoice-customized">مخصصة</span>}
              </summary>
              <div className="pos-invoice-options-content">
                <label className="pos-tax-toggle">
                  <input
                    type="checkbox"
                    checked={taxExempt}
                    onChange={(event) => setTaxExempt(event.target.checked)}
                    className="h-4 w-4 accent-cyan"
                  />
                  <span>إعفاء هذه الفاتورة من الضريبة</span>
                </label>

                {discountsEnabled && (
                  <div className="pos-discount-control">
                    <Input
                      label="خصم الفاتورة (%)"
                      type="number"
                      min="0"
                      max={maxDiscountRate}
                      step="0.01"
                      value={discountRate}
                      onChange={(event) => setDiscountRate(Math.min(Math.max(Number(event.target.value) || 0, 0), maxDiscountRate))}
                    />
                    <p className="pos-discount-help">الحد الأقصى للخصم: {maxDiscountRate}%</p>
                  </div>
                )}
              </div>
            </details>

            <section className="pos-sale-summary" aria-label="إجمالي الفاتورة" aria-live="polite">
              <div className="pos-sale-summary-row">
                <span>المجموع الفرعي</span>
                <span>₪{displaySubtotal.toLocaleString()}</span>
              </div>
              {displayDiscount > 0 && (
                <div className="pos-sale-summary-row discount">
                  <span>الخصم</span>
                  <span>−₪{displayDiscount.toLocaleString()}</span>
                </div>
              )}
              {displayTax > 0 && (
                <div className="pos-sale-summary-row">
                  <span>الضريبة</span>
                  <span>₪{displayTax.toLocaleString()}</span>
                </div>
              )}
              <div className="pos-sale-summary-total">
                <span>الإجمالي المطلوب</span>
                <strong>₪{displayTotal.toLocaleString()}</strong>
              </div>
            </section>
            </div>

            <AdvancedPaymentPanel
              key={paymentPanelVersion}
              paymentMethod={paymentMethod}
              setPaymentMethod={setPaymentMethod}
              paidAmount={paidAmount}
              setPaidAmount={setPaidAmount}
              total={displayTotal}
              isProcessing={isProcessing}
              onCheckout={handleCheckout}
              checkoutBlockedReason={isShiftOpen ? undefined : isShiftError ? 'تعذر التحقق من الوردية. أعد المحاولة قبل البيع.' : isShiftLoading ? 'جارٍ التحقق من الوردية.' : 'افتح الوردية أولًا قبل إتمام البيع.'}
              onPaymentAllocationsChange={handlePaymentAllocationsChange}
              electronicPaymentsEnabled={electronicPaymentsEnabled}
              enabledElectronicMethods={enabledElectronicMethods}
              selectedCustomer={selectedCustomer}
              customerBalance={customerBalance}
              customerCreditLimit={customerCreditLimit}
              onQuickCustomerCreate={handleQuickCustomerCreate}
              selectedCustomerName={customers.find((customer) => String(customer.id) === String(selectedCustomer))?.name ?? selectedCustomerOption?.name}
              storeName={storeName}
              installmentWhatsAppNumber={installmentWhatsAppNumber}
              installmentWhatsAppMessage={installmentWhatsAppMessage}
              installmentMonths={installmentMonths}
              setInstallmentMonths={setInstallmentMonths}
            />
          </aside>
        </div>
      </div>
      {/* Footer Actions */}
      <footer className="pos-modern-footer">
        <Button
          variant="primary"
          size="sm"
          onClick={handleNewSale}
          className="gap-2"
        >
          <FilePlus2 className="h-4 w-4" />
          <span>فاتورة جديدة</span>
        </Button>
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
        <Button
          variant="secondary"
          size="sm"
          onClick={() => setIsHeldSalesOpen(true)}
        >
          <span>المبيعات المعلقة ({heldSalesData ? heldSales.length : '…'})</span>
        </Button>
        
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
      {/* Cashier Shift Modal */}
      <Modal
        isOpen={isShiftModalOpen}
        onClose={() => setIsShiftModalOpen(false)}
        title="إدارة الوردية"
        variant="modern"
        size="md"
      >
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between rounded-lg border border-border-default p-3">
            <div className="flex items-center gap-2">
              <Wallet className="h-5 w-5" />
              <span>{shiftStatusLabel}</span>
            </div>
            <span className="text-sm text-text-secondary">
              {formatStoreDateTime(shift.openedAt, 'ar')}
            </span>
          </div>
          {(isShiftLoading || isShiftError) && (
            <div className="flex items-center justify-between gap-3 rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-sm" role={isShiftError ? 'alert' : 'status'}>
              <span>{isShiftLoading ? 'جارٍ تحميل حالة الوردية…' : 'تعذر تحميل حالة الوردية.'}</span>
              {isShiftError && (
                <Button variant="secondary" size="sm" onClick={() => void refetchShift()}>
                  <RefreshCw className="h-4 w-4" />
                  إعادة المحاولة
                </Button>
              )}
            </div>
          )}
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div><span className="text-text-secondary">رصيد البداية</span><strong className="block">₪{shift.openingCash.toLocaleString()}</strong></div>
            <div><span className="text-text-secondary">عدد المبيعات</span><strong className="block">{shift.saleCount}</strong></div>
            <div><span className="text-text-secondary">إجمالي المبيعات</span><strong className="block">₪{shift.salesTotal.toLocaleString()}</strong></div>
            <div><span className="text-text-secondary">النقد المتوقع</span><strong className="block">₪{(shift.openingCash + shift.salesTotal).toLocaleString()}</strong></div>
          </div>
          {isShiftOpen ? (
            <>
              <Input
                type="number"
                min="0"
                placeholder="النقد الفعلي عند الإغلاق"
                value={shiftClosingCash}
                onChange={(event) => setShiftClosingCash(event.target.value)}
              />
              <Button variant="primary" onClick={handleCloseShift} disabled={closeShiftMutation.isPending || isShiftLoading || isShiftError}>
                إغلاق الوردية
              </Button>
            </>
          ) : (
            <>
              <Input
                type="number"
                min="0"
                placeholder="رصيد بداية الوردية الجديدة"
                value={shiftOpeningCash}
                onChange={(event) => setShiftOpeningCash(event.target.value)}
              />
              <Button variant="primary" onClick={handleOpenShift} disabled={openShiftMutation.isPending || isShiftLoading || isShiftError}>
                فتح وردية جديدة
              </Button>
            </>
          )}
        </div>
      </Modal>

      {/* Held Sales Modal */}
      <Modal
        isOpen={isHeldSalesOpen}
        onClose={() => setIsHeldSalesOpen(false)}
        title="المبيعات المعلّقة"
        variant="modern"
        size="md"
      >
        {heldSalesLoading ? (
          <div className="flex items-center justify-center gap-2 py-6 text-text-secondary" role="status">
            <RefreshCw className="h-4 w-4 animate-spin" />
            <span>جارٍ تحميل المبيعات المعلقة…</span>
          </div>
        ) : heldSalesError ? (
          <div className="flex flex-col items-center gap-3 py-6 text-center" role="alert">
            <p>تعذر تحميل المبيعات المعلقة.</p>
            <Button variant="secondary" onClick={() => void refetchHeldSales()}>إعادة المحاولة</Button>
          </div>
        ) : heldSales.length === 0 ? (
          <div className="flex flex-col items-center gap-2 py-6 text-center text-text-secondary">
            <Pause className="h-8 w-8 opacity-50" />
            <p>لا توجد مبيعات معلقة</p>
            <p className="text-sm">استخدم «تعليق البيع» لحفظ الفاتورة والعودة إليها لاحقًا.</p>
          </div>
        ) : (
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
        )}
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

      <Modal
        isOpen={isSalesHistoryOpen}
        onClose={() => {
          if (!cleanSalesHistoryMutation.isPending) {
            setIsSalesHistoryOpen(false);
            setShowSalesCleanupConfirmation(false);
            setSalesCleanupConfirmationText('');
          }
        }}
        title="المبيعات السابقة"
        variant="modern"
        size="xl"
      >
        <div className="pos-sales-history">
          <div className="pos-history-intro">
            <div className="pos-history-intro-icon" aria-hidden="true">
              <History className="h-5 w-5" />
            </div>
            <div className="pos-history-intro-copy">
              <strong>سجل فواتير المبيعات</strong>
              <span>ابحث عن فاتورة سابقة لعرضها أو إعادة طباعتها</span>
            </div>
            <span className="pos-history-count" aria-live="polite">
              {Number.isFinite(salesHistoryResultCount) ? salesHistoryResultCount : historicalSales.length} فاتورة
            </span>
            <Button
              type="button"
              variant="danger"
              size="sm"
              className="pos-history-clean-button"
              onClick={() => {
                setSalesCleanupConfirmationText('');
                setShowSalesCleanupConfirmation(true);
              }}
              disabled={cleanSalesHistoryMutation.isPending}
              title="حذف كل سجلات المبيعات القابلة للحذف"
            >
              <Trash2 className="h-4 w-4" aria-hidden="true" />
              <span>تنظيف السجل</span>
            </Button>
          </div>
          {showSalesCleanupConfirmation && (
            <section className="pos-history-clean-confirm" aria-label="تأكيد تنظيف سجل المبيعات">
              <div className="pos-history-clean-copy">
                <strong>حذف كل سجلات المبيعات القابلة للحذف؟</strong>
                <p>
                  سيُعكس أثر المبيعات المحذوفة على المخزون وأرصدة العملاء. ستبقى الفواتير المرتبطة بتحصيلات أو مرتجعات كما هي، ولا يمكن التراجع عن الحذف.
                </p>
              </div>
              <label className="pos-history-clean-confirm-input">
                <span>اكتب «حذف المبيعات» للتأكيد</span>
                <input
                  value={salesCleanupConfirmationText}
                  onChange={(event) => setSalesCleanupConfirmationText(event.target.value)}
                  disabled={cleanSalesHistoryMutation.isPending}
                  autoComplete="off"
                  spellCheck={false}
                  aria-label="تأكيد حذف المبيعات"
                />
              </label>
              {cleanSalesHistoryMutation.isPending && (
                <p className="pos-history-clean-progress" role="status" aria-live="polite">
                  جارٍ تنظيف السجل على الخادم. لا تغلق هذه النافذة.
                </p>
              )}
              <div className="pos-history-clean-actions">
                <Button
                  type="button"
                  variant="danger"
                  size="sm"
                  onClick={() => cleanSalesHistoryMutation.mutate()}
                  disabled={salesCleanupConfirmationText !== 'حذف المبيعات' || cleanSalesHistoryMutation.isPending}
                >
                  {cleanSalesHistoryMutation.isPending ? 'جارٍ حذف السجلات...' : 'حذف المبيعات القابلة للحذف'}
                </Button>
                <Button
                  type="button"
                  variant="secondary"
                  size="sm"
                  onClick={() => {
                    setShowSalesCleanupConfirmation(false);
                    setSalesCleanupConfirmationText('');
                  }}
                  disabled={cleanSalesHistoryMutation.isPending}
                >
                  إلغاء
                </Button>
              </div>
            </section>
          )}
          <label className="pos-history-search">
            <Search className="h-4 w-4" aria-hidden="true" />
            <input
              type="search"
              value={salesHistorySearch}
              onChange={(event) => setSalesHistorySearch(event.target.value)}
              placeholder="ابحث برقم الفاتورة أو اسم العميل..."
              aria-label="البحث في المبيعات السابقة"
              autoFocus
            />
            {salesHistorySearch && (
              <button
                type="button"
                className="pos-history-search-clear"
                onClick={() => setSalesHistorySearch('')}
                aria-label="مسح البحث"
                title="مسح البحث"
              >
                <X className="h-4 w-4" aria-hidden="true" />
              </button>
            )}
          </label>
          {salesHistoryLoading ? (
            <div className="pos-history-state" role="status">جارٍ تحميل المبيعات...</div>
          ) : salesHistoryError ? (
            <div className="pos-history-state error" role="alert">
              <span>تعذر تحميل المبيعات السابقة.</span>
              <button type="button" onClick={() => void refetchSalesHistory()}>إعادة المحاولة</button>
            </div>
          ) : historicalSales.length === 0 ? (
            <div className="pos-history-state">لا توجد مبيعات مطابقة لبحثك.</div>
          ) : (
            <ul className="pos-history-list" aria-label="نتائج المبيعات السابقة">
              {historicalSales.map((sale, index) => {
                const saleId = String(sale.id ?? `sale-${index}`);
                const invoiceNumber = String(sale.invoice_number ?? saleId);
                const customerName = String(asRecord(sale.customer).name ?? sale.customer_name ?? 'عميل عام');
                const saleDateValue = sale.created_at ?? sale.sale_date;
                const saleDate = saleDateValue ? formatStoreDateTime(String(saleDateValue), 'ar') : '—';
                const saleTotal = Number(sale.total_amount ?? sale.total ?? 0) || 0;
                return (
                  <li key={saleId}>
                    <button
                      type="button"
                      className="pos-history-row"
                      onClick={() => void handleOpenHistoricalInvoice(sale)}
                      disabled={Boolean(loadingHistoricalSaleId) || cleanSalesHistoryMutation.isPending}
                      aria-label={`عرض فاتورة ${invoiceNumber}`}
                    >
                      <span className="pos-history-row-main">
                        <strong>فاتورة {invoiceNumber}</strong>
                        <small>{customerName} · {saleDate}</small>
                      </span>
                      <span className="pos-history-row-total">₪{saleTotal.toLocaleString()}</span>
                      {loadingHistoricalSaleId === saleId
                        ? <RefreshCw className="h-4 w-4 animate-spin" aria-label="جارٍ التحميل" />
                        : <Printer className="h-4 w-4" aria-hidden="true" />}
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      </Modal>

      <Modal
        isOpen={isInvoiceModalOpen && Boolean(lastSaleData)}
        onClose={() => setIsInvoiceModalOpen(false)}
        title="فاتورة البيع"
        variant="modern"
        size="xl"
      >
        {lastSaleData && (
          <SalesInvoice
            saleData={lastSaleData}
            onClose={() => setIsInvoiceModalOpen(false)}
          />
        )}
      </Modal>
    </div>
  );
}
