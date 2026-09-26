import { useEffect, useMemo, useState } from 'react';
import type { KeyboardEvent as ReactKeyboardEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  AlertCircle,
  ArrowDownLeft,
  ArrowUpRight,
  CalendarDays,
  Check,
  CreditCard,
  History,
  Minus,
  PackagePlus,
  Plus,
  Search,
  ShoppingBasket,
  Trash2,
  UserRound,
  Wallet,
  X,
} from 'lucide-react';
import { Modal } from '../../../design-system/components/modal';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { Badge } from '../../../design-system/components/badge';
import { customersApi, debtsApi, productsApi, salesApi } from '../../../services/api/endpoints';
import { useDebounce } from '../../../hooks/useDebounce';
import { formatStoreDate } from '../../../utils/store-time';
import { toast } from 'sonner';

type PaymentMethod = 'cash' | 'credit' | 'bank_transfer' | 'check';
type AdjustmentType = 'debit' | 'credit';
type ActionMode = 'payment' | 'products' | 'adjustment' | null;

interface DebtLineItem {
  product_id: string;
  product_name: string;
  quantity: number;
  unit_price: number;
  total_amount: number;
}

interface DebtHistoryEntry {
  id: string;
  sale_id?: string;
  reference_type?: string;
  invoice_number?: string;
  amount: number;
  paid_amount: number;
  remaining_amount: number;
  due_date?: string;
  status?: string;
  notes?: string;
  created_at?: string;
  items: DebtLineItem[];
}

interface DebtProduct {
  id: string;
  name: string;
  sku?: string;
  barcode?: string;
  selling_price?: number;
  stock?: number;
  current_quantity?: number;
}

interface ProductCartLine {
  product: DebtProduct;
  quantity: number;
}

interface LinkedAdjustmentProduct {
  id: string;
  name: string;
  quantity: number;
}

const linkedAdjustmentProductMarker = '[PARTFLOW_LINKED_PRODUCT_V1]';

const readAdjustmentDetails = (description: string) => {
  const markerIndex = description.lastIndexOf(linkedAdjustmentProductMarker);
  if (markerIndex < 0) return { reason: description.trim(), product: null as LinkedAdjustmentProduct | null };
  const reason = description.slice(0, markerIndex).trim();
  const metadataText = description.slice(markerIndex + linkedAdjustmentProductMarker.length).trim();
  try {
    const parsed = JSON.parse(metadataText) as LinkedAdjustmentProduct;
    if (!parsed?.id || !parsed?.name || !Number.isFinite(Number(parsed.quantity)) || Number(parsed.quantity) <= 0) {
      return { reason: description.trim(), product: null };
    }
    return { reason, product: parsed };
  } catch {
    return { reason: description.trim(), product: null };
  }
};

const normalizeSearchText = (value: unknown) => String(value ?? '')
  .toLocaleLowerCase()
  .replace(/[٠-٩]/g, (digit) => String('٠١٢٣٤٥٦٧٨٩'.indexOf(digit)))
  .trim();

interface DebtManagementModalProps {
  debt: any;
  isOpen: boolean;
  initialAction?: ActionMode;
  onClose: () => void;
  onRecordPayment: (input: {
    customerId: string;
    amount: number;
    method: PaymentMethod;
    reference: string;
    saleId?: string;
  }) => Promise<void>;
  onAdjustDebt: (input: {
    customerId: string;
    amount: number;
    type: AdjustmentType;
    reason: string;
    productId?: string;
    productQuantity?: number;
  }) => Promise<void>;
}

const unwrapResponse = (response: any) => response?.data?.data ?? response?.data ?? response;

const formatMoney = (amount: unknown) => `₪${Number(amount || 0).toLocaleString('en-US', { maximumFractionDigits: 2 })}`;

const formatDate = (date?: string) => {
  if (!date) return 'غير محدد';
  const formatted = formatStoreDate(date, 'ar-SA');
  return formatted === 'غير محدد' ? date : formatted;
};

export function DebtManagementModal({ debt, isOpen, initialAction = null, onClose, onRecordPayment, onAdjustDebt }: DebtManagementModalProps) {
  const queryClient = useQueryClient();
  const customerId = String(debt?.customer?.id ?? debt?.customer_id ?? '');
  const customerName = String(debt?.customer?.name ?? debt?.customer_name ?? 'العميل');
  const [actionMode, setActionMode] = useState<ActionMode>(null);
  const [paymentAmount, setPaymentAmount] = useState('');
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('cash');
  const [paymentReference, setPaymentReference] = useState(() => crypto.randomUUID());
  const [saleIdempotencyKey, setSaleIdempotencyKey] = useState(() => `debt-sale-${crypto.randomUUID()}`);
  const [productSearch, setProductSearch] = useState('');
  const [productCart, setProductCart] = useState<ProductCartLine[]>([]);
  const [adjustmentAmount, setAdjustmentAmount] = useState('');
  const [adjustmentType, setAdjustmentType] = useState<AdjustmentType>('debit');
  const [adjustmentReason, setAdjustmentReason] = useState('');
  const [adjustmentProduct, setAdjustmentProduct] = useState<DebtProduct | null>(null);
  const [adjustmentProductQuantity, setAdjustmentProductQuantity] = useState('1');
  const [paymentSaleTarget, setPaymentSaleTarget] = useState<{ saleId: string; invoiceLabel: string; remaining: number } | null>(null);
  const [recordSearch, setRecordSearch] = useState('');
  const debouncedProductSearch = useDebounce(productSearch, 250);

  useEffect(() => {
    if (!isOpen) return;
    setActionMode(initialAction);
    setPaymentAmount('');
    setPaymentMethod('cash');
    setPaymentReference(crypto.randomUUID());
    setSaleIdempotencyKey(`debt-sale-${crypto.randomUUID()}`);
    setProductSearch('');
    setProductCart([]);
    setAdjustmentAmount('');
    setAdjustmentType('debit');
    setAdjustmentReason(`تصحيح رصيد - ${customerName}`);
    setAdjustmentProduct(null);
    setAdjustmentProductQuantity('1');
    setPaymentSaleTarget(null);
    setRecordSearch('');
  }, [customerId, customerName, initialAction, isOpen]);

  const customerQuery = useQuery({
    queryKey: ['debt-management-customer', customerId],
    queryFn: () => customersApi.get(customerId),
    enabled: isOpen && Boolean(customerId),
    staleTime: 15_000,
  });
  const historyQuery = useQuery({
    queryKey: ['debt-management-history', customerId],
    queryFn: () => debtsApi.getCustomerHistory(customerId),
    enabled: isOpen && Boolean(customerId),
  });
  const timelineQuery = useQuery({
    queryKey: ['debt-management-timeline', customerId],
    queryFn: () => customersApi.getFinancialTimeline(customerId),
    enabled: isOpen && Boolean(customerId),
  });
  const productsQuery = useQuery({
    queryKey: ['debt-management-products', debouncedProductSearch],
    queryFn: () => productsApi.list({
      page: 1,
      per_page: 12,
      search: debouncedProductSearch.trim(),
      in_stock_only: actionMode === 'products' ? true : undefined,
    }),
    enabled: isOpen && (actionMode === 'products' || actionMode === 'adjustment') && Boolean(debouncedProductSearch.trim()),
    staleTime: 15_000,
  });

  const historyPayload = unwrapResponse(historyQuery.data);
  const history = (Array.isArray(historyPayload)
    ? historyPayload
    : historyPayload?.debts ?? historyPayload?.items ?? []) as DebtHistoryEntry[];
  const customerPayload = unwrapResponse(customerQuery.data);
  const customer = customerPayload?.customer ?? customerPayload;
  const historyBalance = history.reduce((sum, entry) => sum + Math.max(0, Number(entry.remaining_amount ?? 0)), 0);
  const rawCustomerBalance = customer?.current_balance ?? customer?.currentBalance ?? customer?.balance;
  const parsedCustomerBalance = rawCustomerBalance == null ? Number.NaN : Number(rawCustomerBalance);
  const outstanding = Math.max(0, Number.isFinite(parsedCustomerBalance) ? parsedCustomerBalance : historyBalance);

  const timelinePayload = unwrapResponse(timelineQuery.data);
  const timeline = (Array.isArray(timelinePayload)
    ? timelinePayload
    : timelinePayload?.transactions ?? timelinePayload?.entries ?? []) as Array<Record<string, any>>;
  const sortedTimeline = useMemo(
    () => [...timeline].sort((left, right) => String(right.date ?? right.created_at ?? '').localeCompare(String(left.date ?? left.created_at ?? ''))),
    [timeline],
  );
  const normalizedRecordSearch = normalizeSearchText(recordSearch);
  const filteredHistory = useMemo(() => {
    if (!normalizedRecordSearch) return history;
    return history.filter((entry) => {
      const itemDetails = (entry.items ?? []).flatMap((item) => [
        item.product_name,
        item.product_id,
        item.quantity,
        item.unit_price,
        item.total_amount,
      ]);
      return normalizeSearchText([
        entry.invoice_number,
        entry.sale_id,
        entry.reference_type,
        entry.amount,
        entry.paid_amount,
        entry.remaining_amount,
        entry.due_date,
        entry.status,
        entry.notes,
        entry.created_at,
        ...itemDetails,
      ].join(' ')).includes(normalizedRecordSearch);
    });
  }, [history, normalizedRecordSearch]);
  const filteredTimeline = useMemo(() => {
    if (!normalizedRecordSearch) return sortedTimeline;
    return sortedTimeline.filter((entry) => {
      const description = String(entry.description ?? '');
      const adjustment = readAdjustmentDetails(description);
      return normalizeSearchText([
        description,
        adjustment.reason,
        adjustment.product?.name,
        adjustment.product?.id,
        adjustment.product?.quantity,
        entry.type,
        entry.transaction_type,
        entry.amount,
        entry.balance_after,
        entry.reference_id,
        entry.date,
        entry.created_at,
      ].join(' ')).includes(normalizedRecordSearch);
    });
  }, [normalizedRecordSearch, sortedTimeline]);

  const productsPayload = unwrapResponse(productsQuery.data);
  const products = (Array.isArray(productsPayload)
    ? productsPayload
    : productsPayload?.products ?? productsPayload?.items ?? []) as DebtProduct[];

  const subtotal = productCart.reduce((sum, line) => {
    const unitPrice = Number(line.product.selling_price ?? 0);
    return sum + unitPrice * line.quantity;
  }, 0);

  const invalidateBusinessData = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['debts'] }),
      queryClient.invalidateQueries({ queryKey: ['customers'] }),
      queryClient.invalidateQueries({ queryKey: ['sales'] }),
      queryClient.invalidateQueries({ queryKey: ['inventory'] }),
      queryClient.invalidateQueries({ queryKey: ['products'] }),
      queryClient.invalidateQueries({ queryKey: ['reports'] }),
      queryClient.invalidateQueries({ queryKey: ['dashboard'] }),
      queryClient.invalidateQueries({ queryKey: ['debt-management-customer', customerId] }),
      queryClient.invalidateQueries({ queryKey: ['debt-management-history', customerId] }),
      queryClient.invalidateQueries({ queryKey: ['debt-management-timeline', customerId] }),
    ]);
  };

  const saleMutation = useMutation({
    mutationFn: () => salesApi.create({
      customer_id: customerId,
      items: productCart.map(({ product, quantity }) => ({
        product_id: product.id,
        barcode: product.barcode,
        quantity,
        unit_price: Number(product.selling_price ?? 0),
      })),
      payment_method: 'debt',
      payment_amount: 0,
      total_amount: subtotal,
    }, saleIdempotencyKey),
    onSuccess: async () => {
      await invalidateBusinessData();
      setProductCart([]);
      setProductSearch('');
      setActionMode(null);
      setSaleIdempotencyKey(`debt-sale-${crypto.randomUUID()}`);
      toast.success('تمت إضافة السلع كفاتورة آجلة وتحديث المخزون والدين');
    },
    onError: (error: any) => {
      const message = String(error?.arabicMessage ?? error?.message ?? 'تعذر إنشاء الفاتورة الآجلة');
      toast.error(message.toLowerCase().includes('shift') || message.includes('وردية')
        ? 'افتح الوردية أولًا من نقطة البيع، ثم أعد إضافة السلع.'
        : message);
    },
  });

  const addProduct = (product: DebtProduct) => {
    if (!product.id || !Number.isFinite(Number(product.selling_price)) || Number(product.selling_price) < 0) {
      toast.error('لا يوجد سعر بيع صالح لهذا المنتج');
      return;
    }
    setProductCart((current) => {
      const existing = current.find((line) => line.product.id === product.id);
      if (existing) {
        return current.map((line) => line.product.id === product.id
          ? { ...line, quantity: line.quantity + 1 }
          : line);
      }
      return [...current, { product, quantity: 1 }];
    });
  };

  const handleProductSearchKeyDown = async (event: ReactKeyboardEvent<HTMLInputElement>) => {
    if (event.key !== 'Enter') return;
    event.preventDefault();
    const query = productSearch.trim();
    if (!query) return;
    const normalizedQuery = query.toLocaleLowerCase();
    const searchResultsAreCurrent = debouncedProductSearch.trim().toLocaleLowerCase() === normalizedQuery;
    const exactMatch = searchResultsAreCurrent
      ? products.find((product) => [product.barcode, product.sku]
        .some((value) => String(value ?? '').trim().toLocaleLowerCase() === normalizedQuery))
      : undefined;
    if (exactMatch) {
      addProduct(exactMatch);
      setProductSearch('');
      return;
    }
    if (searchResultsAreCurrent && products.length === 1) {
      addProduct(products[0]);
      setProductSearch('');
      return;
    }

    try {
      const barcodeResponse = await productsApi.getByBarcode(query);
      const barcodePayload = unwrapResponse(barcodeResponse);
      const product = barcodePayload?.product ?? barcodePayload;
      if (!product?.id) throw new Error('لم يتم العثور على سلعة بهذا الباركود');
      addProduct(product as DebtProduct);
      setProductSearch('');
    } catch (error: any) {
      toast.error(String(error?.arabicMessage ?? error?.message ?? 'لم يتم العثور على السلعة'));
    }
  };

  const changeQuantity = (productId: string, delta: number) => {
    setProductCart((current) => current
      .map((line) => line.product.id === productId
        ? { ...line, quantity: Math.max(0, line.quantity + delta) }
        : line)
      .filter((line) => line.quantity > 0));
  };

  const submitPayment = async () => {
    const amount = Number(paymentAmount);
    if (!customerId || !Number.isFinite(amount) || amount <= 0 || amount > paymentLimit + 0.000001) {
      toast.error('أدخل مبلغًا صالحًا لا يتجاوز الرصيد المستحق');
      return;
    }
    try {
      await onRecordPayment({ customerId, amount, method: paymentMethod, reference: paymentReference, saleId: paymentSaleTarget?.saleId });
      await invalidateBusinessData();
      setPaymentAmount('');
      setPaymentReference(crypto.randomUUID());
      setPaymentSaleTarget(null);
      setActionMode(null);
      toast.success('تم تسجيل الدفعة');
    } catch (error: any) {
      toast.error(String(error?.arabicMessage ?? error?.message ?? 'تعذر تسجيل الدفعة؛ بقي النموذج مفتوحًا'));
    }
  };

  const submitAdjustment = async () => {
    const amount = Number(adjustmentAmount);
    if (!customerId || !Number.isFinite(amount) || amount <= 0) {
      toast.error('أدخل مبلغًا صالحًا للتصحيح');
      return;
    }
    if (adjustmentType === 'credit' && amount > outstanding + 0.000001) {
      toast.error('مبلغ الخصم أكبر من الرصيد المستحق');
      return;
    }
    const productQuantity = adjustmentProduct ? Number(adjustmentProductQuantity) : undefined;
    if (adjustmentProduct && (!Number.isFinite(productQuantity) || Number(productQuantity) <= 0 || Number(productQuantity) > 1_000_000)) {
      toast.error('أدخل كمية صحيحة للسلعة المرتبطة');
      return;
    }
    try {
      await onAdjustDebt({
        customerId,
        amount,
        type: adjustmentType,
        reason: adjustmentReason.trim() || `تصحيح رصيد - ${customerName}`,
        productId: adjustmentProduct?.id,
        productQuantity,
      });
      await invalidateBusinessData();
      setAdjustmentAmount('');
      setAdjustmentProduct(null);
      setAdjustmentProductQuantity('1');
      setActionMode(null);
      toast.success(adjustmentType === 'debit' ? 'تمت زيادة الدين وتسجيلها في السجل' : 'تم خصم الدين وتسجيله في السجل');
    } catch (error: any) {
      toast.error(String(error?.arabicMessage ?? error?.message ?? 'تعذر تصحيح رصيد الدين'));
    }
  };

  const startPaymentForInvoice = (entry: DebtHistoryEntry, remaining: number) => {
    if (!entry.sale_id || remaining <= 0) return;
    setPaymentSaleTarget({
      saleId: entry.sale_id,
      invoiceLabel: entry.invoice_number || 'فاتورة بيع',
      remaining,
    });
    setPaymentAmount('');
    setPaymentReference(crypto.randomUUID());
    setActionMode('payment');
  };

  const toggleGeneralPayment = () => {
    if (actionMode === 'payment' && !paymentSaleTarget) {
      setActionMode(null);
      return;
    }
    setPaymentSaleTarget(null);
    setActionMode('payment');
  };

  const isLoadingHistory = historyQuery.isLoading || customerQuery.isLoading;
  const paymentLimit = paymentSaleTarget?.remaining ?? outstanding;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`إدارة دين ${customerName}`}
      size="2xl"
      className="max-h-[92vh] overflow-y-auto"
      autoFocus={false}
    >
      <div dir="rtl" className="space-y-4 p-4 sm:p-6">
        <section className="grid gap-3 sm:grid-cols-3">
          <div className="rounded-xl border border-border bg-surface-elevated p-3">
            <div className="flex items-center gap-2 text-xs text-text-muted"><UserRound className="h-4 w-4" /> العميل</div>
            <div className="mt-1 truncate font-bold text-text-primary">{customerName}</div>
            {customer?.phone || debt?.customer?.phone ? <div className="mt-1 text-xs text-text-muted">{customer?.phone ?? debt?.customer?.phone}</div> : null}
          </div>
          <div className="rounded-xl border border-border bg-surface-elevated p-3">
            <div className="flex items-center gap-2 text-xs text-text-muted"><Wallet className="h-4 w-4" /> الرصيد المستحق الآن</div>
            <div className="mt-1 text-xl font-black text-danger">{formatMoney(outstanding)}</div>
          </div>
          <div className="rounded-xl border border-border bg-surface-elevated p-3">
            <div className="flex items-center gap-2 text-xs text-text-muted"><History className="h-4 w-4" /> عدد فواتير الدين</div>
            <div className="mt-1 text-xl font-black text-text-primary">{history.length}</div>
          </div>
        </section>

        <section className="flex flex-wrap gap-2">
          <Button type="button" variant={actionMode === 'payment' && !paymentSaleTarget ? 'primary' : 'secondary'} onClick={toggleGeneralPayment} disabled={outstanding <= 0}>
            <CreditCard className="h-4 w-4" /> تسجيل دفعة
          </Button>
          <Button type="button" variant={actionMode === 'products' ? 'primary' : 'secondary'} onClick={() => setActionMode(actionMode === 'products' ? null : 'products')}>
            <PackagePlus className="h-4 w-4" /> إضافة سلعة على الحساب
          </Button>
          <Button type="button" variant={actionMode === 'adjustment' ? 'primary' : 'secondary'} onClick={() => setActionMode(actionMode === 'adjustment' ? null : 'adjustment')}>
            <ArrowUpRight className="h-4 w-4" /> تصحيح الرصيد
          </Button>
        </section>

        {actionMode === 'payment' && (
          <section className="space-y-3 rounded-xl border border-primary/20 bg-primary/5 p-4">
            <div className="flex items-start justify-between gap-3">
              <div>
                <h4 className="font-bold text-text-primary">{paymentSaleTarget ? `تسجيل دفعة للفاتورة ${paymentSaleTarget.invoiceLabel}` : 'تسجيل دفعة للعميل'}</h4>
                <p className="mt-1 text-xs text-text-muted">{paymentSaleTarget ? `سيُخصم المبلغ من هذه الفاتورة فقط. المتبقي: ${formatMoney(paymentSaleTarget.remaining)}` : `المبلغ الأقصى: ${formatMoney(outstanding)}`}</p>
              </div>
              {paymentSaleTarget ? <Button type="button" variant="ghost" size="icon" aria-label="إلغاء ربط الدفعة بالفاتورة" title="إلغاء ربط الدفعة بالفاتورة" onClick={() => setPaymentSaleTarget(null)}><X className="h-4 w-4" /></Button> : null}
            </div>
            <div className="grid gap-3 sm:grid-cols-2">
              <div>
                <label className="mb-1 block text-xs font-semibold text-text-secondary">مبلغ الدفعة</label>
                <Input type="number" min="0.01" max={paymentLimit} step="0.01" value={paymentAmount} onChange={(event) => setPaymentAmount(event.target.value)} placeholder="أدخل المبلغ" autoFocus />
              </div>
              <div>
                <label className="mb-1 block text-xs font-semibold text-text-secondary">طريقة الدفع</label>
                <select value={paymentMethod} onChange={(event) => setPaymentMethod(event.target.value as PaymentMethod)} className="h-10 w-full rounded-lg border border-border bg-surface px-3 text-sm text-text-primary">
                  <option value="cash">نقدي</option>
                  <option value="credit">بطاقة</option>
                  <option value="bank_transfer">تحويل بنكي</option>
                  <option value="check">شيك</option>
                </select>
              </div>
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="secondary" onClick={() => setActionMode(null)}>إلغاء</Button>
              <Button type="button" onClick={submitPayment} disabled={!paymentAmount || Number(paymentAmount) <= 0 || Number(paymentAmount) > paymentLimit}>حفظ الدفعة</Button>
            </div>
          </section>
        )}

        {actionMode === 'products' && (
          <section className="space-y-3 rounded-xl border border-primary/20 bg-primary/5 p-4">
            <div>
              <h4 className="font-bold text-text-primary">إضافة بضاعة كفاتورة آجلة</h4>
              <p className="mt-1 text-xs text-text-muted">تُسجّل العملية كبيع اعتيادي، ويخصم المخزون ويُضاف الدين تلقائيًا. تطبق ضريبة المتجر عند الحفظ.</p>
            </div>
            <div className="relative">
              <Search className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
              <Input value={productSearch} onChange={(event) => setProductSearch(event.target.value)} onKeyDown={handleProductSearchKeyDown} placeholder="ابحث باسم السلعة أو امسح الباركود" className="pr-9" autoComplete="off" />
            </div>

            {productSearch.trim() && (
              <div className="max-h-44 space-y-1 overflow-y-auto rounded-lg border border-border bg-surface p-1">
                {productsQuery.isLoading ? <div className="p-3 text-center text-xs text-text-muted">جارٍ البحث...</div>
                  : productsQuery.isError ? <div className="p-3 text-center text-xs text-danger">تعذر البحث عن السلع</div>
                    : products.length === 0 ? <div className="p-3 text-center text-xs text-text-muted">لا توجد سلعة متاحة بهذا الاسم أو الباركود</div>
                      : products.map((product) => (
                        <button key={product.id} type="button" onClick={() => addProduct(product)} className="flex w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-right hover:bg-surface-elevated">
                          <span className="min-w-0">
                            <span className="block truncate text-sm font-semibold text-text-primary">{product.name}</span>
                            <span className="block truncate text-[11px] text-text-muted">{product.barcode || product.sku || ''}</span>
                          </span>
                          <span className="shrink-0 text-sm font-bold text-primary">{formatMoney(product.selling_price)}</span>
                        </button>
                      ))}
              </div>
            )}

            {productCart.length > 0 ? (
              <div className="space-y-2">
                {productCart.map(({ product, quantity }) => (
                  <div key={product.id} className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-border bg-surface px-3 py-2">
                    <div className="min-w-0 flex-1">
                      <div className="truncate text-sm font-semibold text-text-primary">{product.name}</div>
                      <div className="text-xs text-text-muted">{formatMoney(product.selling_price)} × {quantity} = {formatMoney(Number(product.selling_price ?? 0) * quantity)}</div>
                    </div>
                    <div className="flex items-center gap-1">
                      <Button type="button" variant="ghost" size="icon" aria-label={`زيادة كمية ${product.name}`} onClick={() => changeQuantity(product.id, 1)}><Plus className="h-4 w-4" /></Button>
                      <span className="min-w-7 text-center text-sm font-bold">{quantity}</span>
                      <Button type="button" variant="ghost" size="icon" aria-label={`تقليل كمية ${product.name}`} onClick={() => changeQuantity(product.id, -1)}><Minus className="h-4 w-4" /></Button>
                      <Button type="button" variant="ghost" size="icon" aria-label={`حذف ${product.name} من الفاتورة`} onClick={() => setProductCart((current) => current.filter((line) => line.product.id !== product.id))}><Trash2 className="h-4 w-4 text-danger" /></Button>
                    </div>
                  </div>
                ))}
                <div className="flex items-center justify-between rounded-lg bg-surface-elevated px-3 py-2 font-bold text-text-primary">
                  <span>الإجمالي قبل ضريبة المتجر</span><span>{formatMoney(subtotal)}</span>
                </div>
                <div className="flex justify-end gap-2">
                  <Button type="button" variant="secondary" onClick={() => setProductCart([])}>تفريغ القائمة</Button>
                  <Button type="button" onClick={() => saleMutation.mutate()} disabled={saleMutation.isPending || productCart.length === 0} isLoading={saleMutation.isPending}>
                    <ShoppingBasket className="h-4 w-4" /> حفظ كفاتورة آجلة
                  </Button>
                </div>
              </div>
            ) : (
              <div className="rounded-lg border border-dashed border-border p-4 text-center text-xs text-text-muted">ابحث عن السلع واخترها لإضافتها إلى الفاتورة.</div>
            )}
          </section>
        )}

        {actionMode === 'adjustment' && (
          <section className="space-y-3 rounded-xl border border-warning/30 bg-warning/5 p-4">
            <div>
              <h4 className="font-bold text-text-primary">تصحيح رصيد الدين</h4>
              <p className="mt-1 text-xs text-text-muted">يُسجّل التصحيح مع السبب في حركة حساب العميل للمراجعة.</p>
            </div>
            <div className="grid gap-3 sm:grid-cols-2">
              <div>
                <label className="mb-1 block text-xs font-semibold text-text-secondary">نوع التصحيح</label>
                <select value={adjustmentType} onChange={(event) => setAdjustmentType(event.target.value as AdjustmentType)} className="h-10 w-full rounded-lg border border-border bg-surface px-3 text-sm text-text-primary">
                  <option value="debit">زيادة الدين</option>
                  <option value="credit">خصم من الدين</option>
                </select>
              </div>
              <div>
                <label className="mb-1 block text-xs font-semibold text-text-secondary">المبلغ</label>
                <Input type="number" min="0.01" max={adjustmentType === 'credit' ? outstanding : undefined} step="0.01" value={adjustmentAmount} onChange={(event) => setAdjustmentAmount(event.target.value)} placeholder="أدخل المبلغ" />
              </div>
            </div>
            <div>
              <label className="mb-1 block text-xs font-semibold text-text-secondary">سبب التصحيح</label>
              <Input value={adjustmentReason} onChange={(event) => setAdjustmentReason(event.target.value)} placeholder="اكتب سببًا واضحًا" />
            </div>
            <div className="space-y-2 rounded-lg border border-border bg-surface/70 p-3">
              <div>
                <label className="mb-1 block text-xs font-semibold text-text-secondary">سلعة مرتبطة (اختياري)</label>
                <p className="text-[11px] text-text-muted">تُحفظ السلعة وكميتها في تفاصيل التصحيح فقط، دون تغيير المخزون.</p>
              </div>
              {adjustmentProduct ? (
                <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg bg-surface-elevated p-3">
                  <div className="min-w-0">
                    <div className="truncate text-sm font-semibold text-text-primary">{adjustmentProduct.name}</div>
                    <div className="text-xs text-text-muted">{adjustmentProduct.sku || adjustmentProduct.barcode || 'سلعة مرتبطة بالتصحيح'}</div>
                  </div>
                  <div className="flex items-center gap-2">
                    <label htmlFor="adjustment-product-quantity" className="text-xs text-text-secondary">الكمية</label>
                    <Input id="adjustment-product-quantity" type="number" min="0.01" max="1000000" step="0.01" className="w-24" value={adjustmentProductQuantity} onChange={(event) => setAdjustmentProductQuantity(event.target.value)} />
                    <Button type="button" variant="ghost" size="icon" aria-label="إزالة السلعة المرتبطة" title="إزالة السلعة المرتبطة" onClick={() => { setAdjustmentProduct(null); setAdjustmentProductQuantity('1'); }}>
                      <Trash2 className="h-4 w-4 text-danger" />
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="space-y-2">
                  <div className="relative">
                    <Search className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
                    <Input value={productSearch} onChange={(event) => setProductSearch(event.target.value)} placeholder="ابحث باسم السلعة أو رمزها" className="pr-9" />
                  </div>
                  {productSearch.trim() ? (
                    <div className="max-h-40 space-y-1 overflow-y-auto">
                      {productsQuery.isFetching ? <div className="p-2 text-xs text-text-muted">جارٍ البحث عن السلع...</div>
                        : products.length === 0 ? <div className="p-2 text-xs text-text-muted">لا توجد سلع مطابقة.</div>
                          : products.map((product) => (
                            <button key={product.id} type="button" className="flex w-full items-center justify-between gap-3 rounded-lg border border-border bg-surface px-3 py-2 text-right hover:bg-surface-elevated" onClick={() => { setAdjustmentProduct(product); setAdjustmentProductQuantity('1'); setProductSearch(''); }}>
                              <span className="min-w-0 truncate text-sm font-medium text-text-primary">{product.name}</span>
                              <span className="shrink-0 text-xs text-text-muted">{product.sku || product.barcode || ''}</span>
                            </button>
                          ))}
                    </div>
                  ) : null}
                </div>
              )}
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="secondary" onClick={() => setActionMode(null)}>إلغاء</Button>
              <Button type="button" variant="warning" onClick={submitAdjustment} disabled={!adjustmentAmount || Number(adjustmentAmount) <= 0 || (adjustmentType === 'credit' && Number(adjustmentAmount) > outstanding) || (Boolean(adjustmentProduct) && (!adjustmentProductQuantity || Number(adjustmentProductQuantity) <= 0 || Number(adjustmentProductQuantity) > 1000000))}>حفظ التصحيح</Button>
            </div>
          </section>
        )}

        <section className="space-y-2 rounded-xl border border-border bg-surface-elevated/60 p-3">
          <label htmlFor="debt-record-search" className="block text-sm font-semibold text-text-primary">بحث في سجل العميل</label>
          <div className="relative">
            <Search className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
            <Input id="debt-record-search" value={recordSearch} onChange={(event) => setRecordSearch(event.target.value)} placeholder="ابحث في الفواتير والسلع والديون والدفعات وأسباب التصحيح" className="h-11 bg-surface pr-9" />
            {recordSearch ? <Button type="button" variant="ghost" size="icon" className="absolute left-1 top-1/2 -translate-y-1/2" aria-label="مسح البحث" title="مسح البحث" onClick={() => setRecordSearch('')}><X className="h-4 w-4" /></Button> : null}
          </div>
          {normalizedRecordSearch ? <div className="text-[11px] text-text-muted">{filteredHistory.length} فاتورة أو دين · {filteredTimeline.length} دفعة أو تعديل مطابق</div> : null}
        </section>

        <section className="space-y-3">
          <div className="flex items-center justify-between gap-3">
            <h4 className="flex items-center gap-2 font-bold text-text-primary"><ShoppingBasket className="h-4 w-4 text-primary" /> تفاصيل الديون والفواتير السابقة</h4>
            {historyQuery.isError && <Button type="button" variant="ghost" size="sm" onClick={() => void historyQuery.refetch()}>إعادة المحاولة</Button>}
          </div>
          {isLoadingHistory ? <div className="rounded-lg bg-surface-elevated p-5 text-center text-sm text-text-muted">جارٍ تحميل سجل الديون...</div>
            : historyQuery.isError ? <div role="alert" className="rounded-lg border border-danger/20 bg-danger/5 p-4 text-sm text-danger">تعذر تحميل تفاصيل الديون. أعد المحاولة.</div>
              : filteredHistory.length === 0 ? <div className="rounded-lg border border-dashed border-border p-4 text-sm text-text-muted">{normalizedRecordSearch ? 'لا توجد فواتير أو ديون تطابق البحث.' : 'لا توجد فواتير دين مرتبطة. راجع سجل الحركات أدناه لأي رصيد يدوي أو افتتاحي.'}</div>
                : <div className="max-h-64 space-y-3 overflow-y-auto pr-1">
                  {filteredHistory.map((entry) => {
                    const hasSale = Boolean(entry.sale_id);
                    const referenceType = String(entry.reference_type ?? '').toLowerCase();
                    const remaining = Math.max(0, Number(entry.remaining_amount ?? Number(entry.amount ?? 0) - Number(entry.paid_amount ?? 0)));
                    return (
                      <article key={entry.id} className="rounded-xl border border-border bg-surface p-3">
                        <div className="flex flex-wrap items-start justify-between gap-2">
                          <div className="min-w-0">
                            <div className="flex flex-wrap items-center gap-2">
                              <span className="font-bold text-text-primary">{hasSale ? entry.invoice_number || 'فاتورة بيع' : referenceType === 'opening_debt' ? 'رصيد افتتاحي' : referenceType === 'manual_adjustment' ? 'تعديل يدوي' : 'دين يدوي أو سجل قديم بلا فاتورة'}</span>
                              <Badge variant={remaining <= 0 ? 'success' : entry.status === 'partial' ? 'warning' : 'secondary'} size="sm">{remaining <= 0 ? 'مسدد' : entry.status === 'partial' ? 'مسدد جزئيًا' : 'مفتوح'}</Badge>
                            </div>
                            <div className="mt-1 flex items-center gap-1 text-xs text-text-muted"><CalendarDays className="h-3.5 w-3.5" /> {formatDate(entry.created_at)}</div>
                          </div>
                          <div className="grid grid-cols-3 gap-3 text-left text-xs">
                            <div><span className="block text-text-muted">الأصل</span><strong>{formatMoney(entry.amount)}</strong></div>
                            <div><span className="block text-text-muted">المدفوع</span><strong className="text-success">{formatMoney(entry.paid_amount)}</strong></div>
                            <div><span className="block text-text-muted">المتبقي</span><strong className="text-danger">{formatMoney(remaining)}</strong></div>
                          </div>
                        </div>
                        {entry.items?.length ? (
                          <div className="mt-3 divide-y divide-border rounded-lg bg-surface-elevated px-3">
                            {entry.items.map((item, index) => (
                              <div key={`${entry.id}-${item.product_id}-${index}`} className="flex items-center justify-between gap-3 py-2 text-xs">
                                <span className="min-w-0 truncate font-medium text-text-primary">{item.product_name || 'منتج'}</span>
                                <span className="shrink-0 text-text-muted">{item.quantity} × {formatMoney(item.unit_price)}</span>
                                <strong className="shrink-0">{formatMoney(item.total_amount)}</strong>
                              </div>
                            ))}
                          </div>
                        ) : <div className="mt-2 text-xs text-text-muted">{hasSale ? 'لم تعد تفاصيل السلع متاحة لهذه الفاتورة.' : entry.notes && !['opening_debt', 'manual_adjustment'].includes(referenceType) ? entry.notes : 'لا توجد سلع مرتبطة بهذا الرصيد.'}</div>}
                        {hasSale && entry.sale_id && remaining > 0 ? (
                          <div className="mt-3 flex justify-end">
                            <Button type="button" size="sm" variant="secondary" onClick={() => startPaymentForInvoice(entry, remaining)}>
                              <CreditCard className="h-4 w-4" /> تسجيل دفعة لهذه الفاتورة
                            </Button>
                          </div>
                        ) : null}
                      </article>
                    );
                  })}
                </div>}
        </section>

        <section className="space-y-3">
          <h4 className="flex items-center gap-2 font-bold text-text-primary"><History className="h-4 w-4 text-primary" /> سجل الدفعات والتعديلات</h4>
          {timelineQuery.isLoading ? <div className="rounded-lg bg-surface-elevated p-4 text-center text-xs text-text-muted">جارٍ تحميل الحركات...</div>
            : timelineQuery.isError ? <div className="flex items-center justify-between gap-2 rounded-lg border border-warning/30 bg-warning/5 p-3 text-xs text-text-secondary"><span className="flex items-center gap-2"><AlertCircle className="h-4 w-4" /> تعذر تحميل سجل الحركات</span><Button type="button" variant="ghost" size="sm" onClick={() => void timelineQuery.refetch()}>إعادة المحاولة</Button></div>
              : filteredTimeline.length === 0 ? <div className="rounded-lg bg-surface-elevated p-4 text-center text-xs text-text-muted">{normalizedRecordSearch ? 'لا توجد دفعات أو تعديلات تطابق البحث.' : 'لا توجد حركات مالية سابقة مسجلة.'}</div>
                : <div className="max-h-48 divide-y divide-border overflow-y-auto rounded-lg border border-border px-3">
                  {filteredTimeline.slice(0, 40).map((entry, index) => {
                    const isDebit = String(entry.type ?? '').toLowerCase() === 'debit';
                    const adjustmentDetails = readAdjustmentDetails(String(entry.description ?? ''));
                    return (
                      <div key={String(entry.id ?? index)} className="flex flex-wrap items-center justify-between gap-2 py-2 text-xs">
                        <div className="min-w-0 flex-1">
                          <div className="truncate font-medium text-text-primary">{adjustmentDetails.product ? adjustmentDetails.reason : entry.description || (isDebit ? 'إضافة دين' : 'دفعة أو خصم')}</div>
                          <div className="mt-0.5 text-text-muted">{formatDate(entry.date ?? entry.created_at)}</div>
                          {adjustmentDetails.product ? <div className="mt-1 flex flex-wrap gap-x-3 gap-y-1 rounded-md bg-surface-elevated px-2 py-1 text-text-secondary"><span>السلعة: <strong className="text-text-primary">{adjustmentDetails.product.name}</strong></span><span>الكمية: <strong className="text-text-primary">{adjustmentDetails.product.quantity}</strong></span></div> : null}
                        </div>
                        <span className={`inline-flex items-center gap-1 font-bold ${isDebit ? 'text-danger' : 'text-success'}`}>
                          {isDebit ? <ArrowUpRight className="h-3.5 w-3.5" /> : <ArrowDownLeft className="h-3.5 w-3.5" />}
                          {isDebit ? '+' : '−'}{formatMoney(entry.amount)}
                        </span>
                        {entry.balance_after != null ? <span className="text-text-muted">الرصيد {formatMoney(entry.balance_after)}</span> : null}
                      </div>
                    );
                  })}
                </div>}
        </section>

        <div className="flex justify-end border-t border-border pt-3">
          <Button type="button" variant="secondary" onClick={onClose}><Check className="h-4 w-4" /> إغلاق</Button>
        </div>
      </div>
    </Modal>
  );
}
