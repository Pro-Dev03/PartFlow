import { InventoryQuickCreateModal } from '../../inventory/components/InventoryQuickCreateModal';
import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { Select } from '../../../design-system/components/select';
import { Badge } from '../../../design-system/components/badge';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { PaginationControls } from '../../../design-system/components/pagination-controls';

import { Modal } from '../../../design-system/components/modal';
import { ConfirmDialog } from '../../../design-system/components/confirm-dialog';
import { Product, Supplier, Category } from '../../../types/models';
import {
  Search,
  Plus,
  Trash2,
  ShoppingCart,
  Truck,
  CheckCircle2,
  Package,
  FileText,
  Sparkles,
  Box,
  MoreHorizontal,
  ChevronRight,
  ImagePlus,
  Upload,
  Minus,
  X,
} from 'lucide-react';
import { suppliersApi, productsApi, categoriesApi, purchasesApi } from '../../../services/api/endpoints';
import { usePurchases } from '../hooks/usePurchases';
import { PurchaseItem } from '../types/purchases.types';
import { toast } from 'sonner';
import { settingsApi } from '../../../services/api/endpoints';
import { calculateSuggestedSellingPrice, DEFAULT_PROFIT_MARGIN } from '../../../utils/pricing';
import { generateSku } from '../../../utils/sku';
import { getStoreToday, storeDateToUTCISOString } from '../../../utils/store-time';
import { compressProductImage, setLocalProductImage } from '../../../services/localProductImages';

interface LineItem extends PurchaseItem {
  key: string;
  sku?: string;
  barcode?: string;
}

interface CreatePurchasePageProps {
  isOpen?: boolean;
  onClose?: () => void;
  onComplete?: (purchase?: any) => void | Promise<void>;
}

export function CreatePurchasePage({ isOpen = true, onClose, onComplete }: CreatePurchasePageProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const isEmbedded = Boolean(onComplete);
  const [selectedSupplier, setSelectedSupplier] = useState('');
  const [items, setItems] = useState<LineItem[]>([]);
  const [productSearchQuery, setProductSearchQuery] = useState('');
  const [productPage, setProductPage] = useState(1);
  const [selectedProductDraft, setSelectedProductDraft] = useState<any | null>(null);
  const [draftQuantity, setDraftQuantity] = useState('1');
  const [draftUnitCost, setDraftUnitCost] = useState('0');
  const [invoiceNumber, setInvoiceNumber] = useState('');
  const [purchaseDate, setPurchaseDate] = useState(getStoreToday());
  const [expectedDate, setExpectedDate] = useState('');
  const [notes, setNotes] = useState('');
  const [receiveImmediately, setReceiveImmediately] = useState(isEmbedded);
  const [initialPayment, setInitialPayment] = useState('');
  const [currentStep, setCurrentStep] = useState<1 | 2 | 3 | 4 | 5>(1);
  const [quickCreateMode, setQuickCreateMode] = useState<'category' | 'supplier' | null>(null);
  const [productToDelete, setProductToDelete] = useState<Product | null>(null);
  const [editingProductId, setEditingProductId] = useState<string | null>(null);
  const [expandedItemKey, setExpandedItemKey] = useState<string | null>(null);
  const [showUnsavedExitDialog, setShowUnsavedExitDialog] = useState(false);
  const invoiceNumberInputRef = useRef<HTMLInputElement>(null);
  const productSearchInputRef = useRef<HTMLInputElement>(null);
  const paymentInputRef = useRef<HTMLInputElement>(null);
  const draftQuantityInputRef = useRef<HTMLInputElement>(null);
  const draftUnitCostInputRef = useRef<HTMLInputElement>(null);
  const manualNameInputRef = useRef<HTMLInputElement>(null);
  const manualSkuInputRef = useRef<HTMLInputElement>(null);
  const manualBarcodeInputRef = useRef<HTMLInputElement>(null);
  const manualCategoryInputRef = useRef<HTMLSelectElement>(null);
  const manualCostInputRef = useRef<HTMLInputElement>(null);
  const manualSellingInputRef = useRef<HTMLInputElement>(null);
  const manualMinStockInputRef = useRef<HTMLInputElement>(null);
  const createPurchaseButtonRef = useRef<HTMLButtonElement>(null);
  const quantityInputRefs = useRef<Record<string, HTMLInputElement | null>>({});
  const unitCostInputRefs = useRef<Record<string, HTMLInputElement | null>>({});
  
  // Manual product addition states
  const [isManualProductModalOpen, setIsManualProductModalOpen] = useState(false);
  const [showManualProductDetails, setShowManualProductDetails] = useState(false);
  const [manualProductData, setManualProductData] = useState({
    name: '',
    sku: '',
    barcode: '',
    category_id: '',
    cost_price: '',
    selling_price: '',
    min_stock: '',
    description: '',
  });
  const [manualProductImage, setManualProductImage] = useState<string | null>(null);
  const { data: marginSetting } = useQuery({
    queryKey: ['settings', 'default_profit_margin'],
    queryFn: () => settingsApi.getSetting('default_profit_margin'),
    retry: false,
  });
  const suggestedMargin = Number(marginSetting?.data?.value);
  const profitMargin = Number.isFinite(suggestedMargin) && suggestedMargin >= 0 && suggestedMargin < 100
    ? suggestedMargin
    : DEFAULT_PROFIT_MARGIN;
  const { data: taxSetting } = useQuery({
    queryKey: ['settings', 'tax_rate'],
    queryFn: () => settingsApi.getSetting('tax_rate'),
    retry: false,
  });
  const taxRate = Math.max(0, Math.min(100, Number(taxSetting?.data?.value) || 0));

  const { data: suppliersData, isLoading: suppliersLoading } = useQuery({
    queryKey: ['suppliers'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 100 }),
  });

  const { data: productsData, isLoading: productsLoading } = useQuery({
    queryKey: ['products', productSearchQuery, productPage],
    queryFn: () => productsApi.list({ search: productSearchQuery, page: productPage, per_page: 10 }),
  });

  const productTotal = Number(productsData?.data?.total ?? 0);
  const productTotalPages = Math.max(1, Math.ceil(productTotal / 10));

  useEffect(() => {
    setProductPage(1);
  }, [productSearchQuery]);

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });

  const createProductMutation = useMutation({
    mutationFn: (data: any) => productsApi.create(data),
    onSuccess: (response) => {
      const newProduct = response.data;
      if (manualProductImage && newProduct?.id) {
        setLocalProductImage(newProduct.id, manualProductImage);
      }
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      
      // Automatically add the new product to the purchase items
      setItems((prev) => [
        ...prev,
        {
          key: `${newProduct.id}-${Date.now()}`,
          product_id: newProduct.id,
          product_name: newProduct.name,
          sku: newProduct.sku || '',
          barcode: newProduct.barcode || '',
          quantity: 1,
          unit_cost: newProduct.cost_price || 0,
          selling_price: newProduct.selling_price || 0,
          category_id: newProduct.category_id || '',
          condition: 'new',
        },
      ]);
      
      setIsManualProductModalOpen(false);
      setShowManualProductDetails(false);
      setManualProductData({
        name: '',
        sku: '',
        barcode: '',
        category_id: '',
        cost_price: '',
        selling_price: '',
        min_stock: '',
        description: '',
      });
      setManualProductImage(null);
      requestAnimationFrame(() => {
        productSearchInputRef.current?.focus();
        productSearchInputRef.current?.select();
      });
    },
  });

  const updateProductMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) => productsApi.update(id, data),
    onSuccess: (response) => {
      const updatedProduct = response.data?.product || response.data;
      if (updatedProduct?.id) {
        setLocalProductImage(updatedProduct.id, manualProductImage);
      }
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      setItems((prev) => prev.map((item) => item.product_id === updatedProduct.id
        ? {
            ...item,
            product_name: updatedProduct.name,
            unit_cost: Number(updatedProduct.cost_price) || 0,
            selling_price: Number(updatedProduct.selling_price) || 0,
            category_id: updatedProduct.category_id || '',
          }
        : item
      ));
      setEditingProductId(null);
      setIsManualProductModalOpen(false);
      setShowManualProductDetails(false);
      requestAnimationFrame(() => {
        productSearchInputRef.current?.focus();
        productSearchInputRef.current?.select();
      });
      toast.success('تم تحديث بيانات القطعة');
    },
    onError: () => toast.error('تعذر تحديث بيانات القطعة'),
  });

  const deleteProductMutation = useMutation({
    mutationFn: (productId: string) => productsApi.delete(productId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      queryClient.invalidateQueries({ queryKey: ['reports'] });
      toast.success('تم حذف المنتج بنجاح');
    },
    onError: (error) => {
      console.error('Error deleting product:', error);
      toast.error('حدث خطأ أثناء حذف المنتج');
    },
  });

  const searchedProducts = (productsData?.data?.products as Product[]) || [];
  const suppliers = (suppliersData?.data as Supplier[]) || [];
  const categories = (categoriesData?.data as Category[]) || [];

  const { createPurchaseMutation, receivePurchaseMutation } = usePurchases();

  const totalCost = useMemo(
    () => items.reduce((sum, item) => sum + item.quantity * item.unit_cost, 0),
    [items]
  );
  const taxAmount = totalCost * taxRate / 100;
  const totalWithTax = totalCost + taxAmount;

  const totalQuantity = useMemo(
    () => items.reduce((sum, item) => sum + item.quantity, 0),
    [items]
  );

  useEffect(() => {
    setInvoiceNumber(`PO-${Date.now()}`);
    setPurchaseDate(getStoreToday());
  }, []);

  useEffect(() => {
    const purchaseState = location.state as {
      supplierId?: string;
      product?: { id: string; name: string; cost_price?: number };
    } | null;
    const { product, supplierId } = purchaseState || {};

    if (supplierId) setSelectedSupplier(supplierId);
    if (product && items.length === 0) {
      setItems([{
        key: `${product.id}-${Date.now()}`,
        product_id: product.id,
        product_name: product.name,
        quantity: 1,
        unit_cost: Number(product.cost_price) || 0,
        selling_price: Number((product as any).selling_price) || 0,
        category_id: (product as any).category_id || '',
        condition: 'new',
      }]);
    }
    if (product || supplierId) navigate(location.pathname, { replace: true, state: null });
  }, [items.length, location.pathname, location.state, navigate]);

  const handleManualAdd = useCallback((product: any) => {
    setSelectedProductDraft(product);
    setDraftQuantity('1');
    setDraftUnitCost(String(product.cost_price ?? 0));
    setProductSearchQuery('');
    requestAnimationFrame(() => draftQuantityInputRef.current?.focus());
  }, []);

  const handleQuickProductSearch = useCallback(async () => {
    const query = productSearchQuery.trim();
    if (!query) return;

    try {
      const response = await productsApi.getByBarcode(query);
      const product = response.data?.product || response.data;
      if (product?.id) {
        handleManualAdd({ ...product, barcode: query });
        return;
      }
    } catch (error) {
      console.debug('Barcode lookup did not find a product:', error);
    }

    if (searchedProducts[0]) {
      handleManualAdd(searchedProducts[0]);
      return;
    }

    toast.error('لم يتم العثور على المنتج بهذا الباركود');
  }, [handleManualAdd, productSearchQuery, searchedProducts]);

  const handleAddDraftProduct = useCallback(() => {
    if (!selectedProductDraft) return;
    const quantity = Math.max(1, parseInt(draftQuantity, 10) || 0);
    const unitCost = Math.max(0, Number(draftUnitCost) || 0);
    if (unitCost <= 0) {
      toast.error('أدخل سعر شراء صحيح');
      draftUnitCostInputRef.current?.focus();
      return;
    }

    setItems((prev) => {
      const existingIndex = prev.findIndex((item) =>
        item.product_id === selectedProductDraft.id && item.barcode === (selectedProductDraft.barcode || ''),
      );
      if (existingIndex >= 0) {
        return prev.map((item, index) => index === existingIndex
          ? { ...item, quantity: item.quantity + quantity, unit_cost: unitCost }
          : item);
      }
      return [...prev, {
        key: `${selectedProductDraft.id}-${Date.now()}`,
        product_id: selectedProductDraft.id,
        product_name: selectedProductDraft.name,
        sku: selectedProductDraft.sku || '',
        barcode: selectedProductDraft.barcode || '',
        quantity,
        unit_cost: unitCost,
        selling_price: Number(selectedProductDraft.selling_price ?? selectedProductDraft.sellingPrice ?? 0),
        category_id: selectedProductDraft.category_id || '',
        condition: 'new',
      }];
    });
    setSelectedProductDraft(null);
    setDraftQuantity('1');
    setDraftUnitCost('0');
    requestAnimationFrame(() => {
      productSearchInputRef.current?.focus();
      productSearchInputRef.current?.select();
    });
  }, [draftQuantity, draftUnitCost, selectedProductDraft]);

  const handleRemoveItem = useCallback((key: string) => {
    setItems((prev) => prev.filter((item) => item.key !== key));
  }, []);

  const handleUpdateQuantity = useCallback((key: string, quantity: number) => {
    setItems((prev) => {
      if (quantity <= 0) return prev.filter((item) => item.key !== key);
      return prev.map((item) => (item.key === key ? { ...item, quantity } : item));
    });
  }, []);

  const handleUpdateUnitCost = useCallback((key: string, unit_cost: number) => {
    setItems((prev) =>
      prev.map((item) => (item.key === key ? { ...item, unit_cost: unit_cost || 0 } : item))
    );
  }, []);

  const handleUpdateCondition = useCallback(
    (key: string, condition: 'new' | 'used' | 'refurbished') => {
      setItems((prev) =>
        prev.map((item) => (item.key === key ? { ...item, condition } : item))
      );
    },
    []
  );

  const handleManualProductCreate = useCallback(async () => {
    if (!manualProductData.name.trim()) {
      toast.error('يرجى إدخال اسم المنتج');
      return;
    }
    const costPrice = parseFloat(manualProductData.cost_price.toString());
    if (!costPrice || costPrice <= 0) {
      toast.error('يرجى إدخال سعر التكلفة');
      return;
    }
    const sellingPrice = parseFloat(manualProductData.selling_price.toString());
    if (!sellingPrice || sellingPrice <= 0) {
      toast.error('يرجى إدخال سعر البيع');
      return;
    }

    const productData = {
      name: manualProductData.name,
      sku: manualProductData.sku || generateSku(),
      barcode: manualProductData.barcode || undefined,
      category_id: manualProductData.category_id || undefined,
      cost_price: costPrice,
      selling_price: sellingPrice,
      min_stock_level: parseInt(manualProductData.min_stock.toString()) || 0,
      description: manualProductData.description || undefined,
    };

    if (editingProductId) {
      updateProductMutation.mutate({ id: editingProductId, data: productData });
    } else {
      createProductMutation.mutate(productData);
    }
  }, [manualProductData, editingProductId, createProductMutation, updateProductMutation]);

  const focusManualField = (event: React.KeyboardEvent, next: HTMLElement | null) => {
    if (event.key !== 'Enter') return;
    event.preventDefault();
    next?.focus();
    if (next instanceof HTMLInputElement) next.select();
  };

  const handleCostPriceChange = (costPrice: string) => {
    setManualProductData((current) => ({
      ...current,
      cost_price: costPrice,
      selling_price: String(calculateSuggestedSellingPrice(Number(costPrice), profitMargin) || ''),
    }));
  };

  const handleCreatePurchase = useCallback(async () => {
    if (!selectedSupplier) {
      toast.error('يرجى اختيار التاجر');
      return;
    }
    if (items.length === 0) {
      toast.error('يرجى إضافة عناصر للشراء');
      return;
    }

    // Validate items before sending
    for (const item of items) {
      if (!item.product_id) {
        toast.error('يوجد عنصر بدون معرف منتج');
        return;
      }
      if (!item.quantity || item.quantity <= 0) {
        toast.error('يرجى التأكد من الكميات');
        return;
      }
      if (!item.unit_cost || item.unit_cost <= 0) {
        toast.error('يرجى التأكد من أسعار التكلفة');
        return;
      }
      if (!item.condition || !['new', 'used', 'refurbished'].includes(item.condition)) {
        toast.error('يرجى التأكد من حالة العناصر');
        return;
      }
    }

    const formData = {
      supplier_id: selectedSupplier,
      invoice_number: invoiceNumber || `PO-${Date.now()}`,
      purchase_date: storeDateToUTCISOString(purchaseDate) ?? new Date().toISOString(),
      expected_delivery_date: expectedDate ? storeDateToUTCISOString(expectedDate) ?? undefined : undefined,
      notes: notes || undefined,
      items: items.map((item) => ({
        product_id: item.product_id,
        quantity: parseInt(item.quantity.toString()) || 0,
        unit_cost: parseFloat(item.unit_cost.toString()) || 0,
        condition: item.condition || 'new',
      })),
    };

    const paymentAmount = Number(initialPayment);
    if (initialPayment && (!Number.isFinite(paymentAmount) || paymentAmount <= 0 || paymentAmount > totalWithTax)) {
      toast.error(`أدخل دفعة صحيحة بين ₪0.01 و ₪${totalWithTax.toLocaleString('en-US', { maximumFractionDigits: 2 })}`);
      return;
    }
    if (receiveImmediately && (!Number.isFinite(paymentAmount) || paymentAmount <= 0)) {
      toast.error('أدخل مبلغ الدفعة قبل استلام البضاعة وتحديث المخزون');
      return;
    }

    // Debug logging (can be removed in production)
    if (process.env.NODE_ENV === 'development') {
      console.log('Sending purchase data:', JSON.stringify(formData, null, 2));
    }

    createPurchaseMutation.mutate(formData, {
      onSuccess: async (response) => {
        let latestResponse = response;
        const purchaseId =
          response?.purchase?.id ||
          response?.data?.purchase?.id ||
          response?.data?.id ||
          response?.id;
        if (purchaseId && (paymentAmount > 0 || receiveImmediately)) {
          try {
            if (paymentAmount > 0) {
              latestResponse = await purchasesApi.addPayment(purchaseId, {
                amount: paymentAmount,
                paymentMethod: 'cash',
              });
            }
            if (receiveImmediately) {
              latestResponse = await receivePurchaseMutation.mutateAsync(purchaseId);
            }
            queryClient.invalidateQueries({ queryKey: ['purchases'] });
            queryClient.invalidateQueries({ queryKey: ['dashboard'] });
            queryClient.invalidateQueries({ queryKey: ['inventory'] });
            queryClient.invalidateQueries({ queryKey: ['products'] });
            queryClient.invalidateQueries({ queryKey: ['suppliers'] });
            toast.success(receiveImmediately ? 'تم إنشاء الشراء وتحديث المخزون بنجاح' : 'تم إنشاء الشراء بنجاح');
            if (onComplete) {
              await onComplete(latestResponse);
            } else {
              navigate('/app/inventory');
            }
          } catch (error) {
            const message = error instanceof Error ? error.message : 'خطأ غير معروف';
            toast.error(`تم إنشاء الشراء، لكن تعذر تسجيل الدفعة أو الاستلام: ${message}`);
          }
        } else {
          queryClient.invalidateQueries({ queryKey: ['purchases'] });
          queryClient.invalidateQueries({ queryKey: ['inventory'] });
          queryClient.invalidateQueries({ queryKey: ['products'] });
          queryClient.invalidateQueries({ queryKey: ['suppliers'] });
          toast.success('تم إنشاء الشراء بنجاح');
          if (onComplete) {
            await onComplete(latestResponse);
          } else {
            navigate('/app/inventory');
          }
        }
      },
    });
  }, [
    selectedSupplier,
    items,
    invoiceNumber,
    purchaseDate,
    expectedDate,
    notes,
    receiveImmediately,
    initialPayment,
    totalCost,
    createPurchaseMutation,
    receivePurchaseMutation,
    queryClient,
    navigate,
    onComplete,
  ]);

  const isSubmitting =
    createPurchaseMutation.isPending || receivePurchaseMutation.isPending;

  const hasUnsavedPurchase = Boolean(selectedSupplier || items.length || notes || initialPayment);
  const completeClosePurchase = useCallback(() => {
    setShowUnsavedExitDialog(false);
    if (onClose) {
      onClose();
      return;
    }
    navigate('/app/purchases');
  }, [navigate, onClose]);

  const closePurchase = useCallback(() => {
    if (hasUnsavedPurchase) {
      setShowUnsavedExitDialog(true);
      return;
    }
    completeClosePurchase();
  }, [completeClosePurchase, hasUnsavedPurchase]);

  const advanceStep = useCallback(() => {
    if (currentStep === 1) {
      if (!selectedSupplier) {
        toast.error('يرجى اختيار التاجر');
        return;
      }
      setCurrentStep(2);
    } else if (currentStep === 2) {
      setCurrentStep(3);
    } else if (currentStep === 3) {
      if (items.length === 0) {
        toast.error('أضف منتجًا واحدًا على الأقل');
        return;
      }
      setCurrentStep(4);
    } else if (currentStep === 4) {
      setCurrentStep(5);
    } else {
      void handleCreatePurchase();
    }
  }, [currentStep, selectedSupplier, items.length, handleCreatePurchase]);

  useEffect(() => {
    requestAnimationFrame(() => {
      if (currentStep === 3) productSearchInputRef.current?.focus();
      if (currentStep === 4) paymentInputRef.current?.focus();
      if (currentStep === 2) invoiceNumberInputRef.current?.focus();
      if (currentStep === 5) createPurchaseButtonRef.current?.focus();
    });
  }, [currentStep]);

  return (
    <>
    <Modal
      isOpen={isOpen}
      onClose={closePurchase}
      title="إنشاء شراء جديد"
      variant="modern"
      size={currentStep === 3 ? '2xl' : 'lg'}
      autoFocus
      enableEnterNavigation={false}
      className={currentStep === 3 ? 'max-w-[980px]' : 'max-w-[620px]'}
    >
      <div
        key={currentStep}
        className="space-y-5 pb-2"
        onKeyDown={(event) => {
          if (event.key !== 'Enter' || event.shiftKey || event.isComposing) return;
          const target = event.target as HTMLElement;
          if (currentStep === 3 && items.length > 0 && target.tagName !== 'BUTTON' && !['INPUT', 'SELECT', 'TEXTAREA'].includes(target.tagName)) {
            event.preventDefault();
            event.stopPropagation();
            advanceStep();
            return;
          }
          if (!['INPUT', 'SELECT', 'TEXTAREA'].includes(target.tagName) || target.tagName === 'TEXTAREA') return;
          event.preventDefault();
          event.stopPropagation();
          advanceStep();
        }}
      >
        <div className="flex items-center justify-between border-b border-border pb-3">
          <div className="flex items-center gap-2 text-xs text-text-muted" aria-label={`المرحلة ${currentStep} من 5`}>
            {[1, 2, 3, 4, 5].map((step) => <span key={step} className={`h-1.5 rounded-full transition-all duration-150 ${step === currentStep ? 'w-6 bg-primary' : 'w-1.5 bg-border'}`} />)}
          </div>
          <div className="flex items-center gap-1">
            {currentStep > 1 && <Button variant="ghost" size="sm" onClick={() => setCurrentStep((step) => Math.max(1, step - 1) as 1 | 2 | 3 | 4 | 5)} className="gap-1"><ChevronRight className="h-4 w-4" />رجوع</Button>}
            <Button variant="ghost" size="sm" onClick={closePurchase}>إلغاء</Button>
          </div>
        </div>

      <div className={currentStep === 3 ? 'mx-auto w-full max-w-[64rem]' : 'mx-auto w-full max-w-[42rem]'}>
        {/* Left Column - Main Form */}
        <div className="space-y-6">
          {/* Supplier and Invoice Info */}
          {currentStep === 1 && <Card className="purchase-supplier-card border-primary/15">
            <CardHeader className="purchase-supplier-card-header">
              <CardTitle className="purchase-supplier-card-title flex items-center gap-2">
                <Truck className="w-5 h-5 text-cyan" />
                اختيار التاجر
              </CardTitle>
              <span className="text-xs text-text-muted">اختر التاجر المرتبط بعملية الشراء</span>
            </CardHeader>
            <CardContent className="purchase-supplier-card-content">
              <div className="space-y-4">
                <div>
                  <div className="mb-2 flex items-center justify-between gap-2">
                    <label className="block text-sm font-medium text-text">التاجر *</label>
                    <Button type="button" variant="ghost" size="sm" onClick={() => setQuickCreateMode('supplier')}>+ إضافة تاجر</Button>
                  </div>
                  <Select
                      value={selectedSupplier}
                      onChange={(e) => setSelectedSupplier(e.target.value)}
                      loading={suppliersLoading}
                      options={[
                        { value: '', label: 'اختر التاجر...' },
                        ...suppliers.map((s) => ({ value: s.id, label: s.name })),
                      ]}
                        emptyMessage="لا يوجد تجار"
                    />
                  <p className="mt-1.5 text-xs text-text-muted">يمكنك إضافة تاجر من هنا دون مغادرة عملية الشراء.</p>
                </div>
              </div>
                <div className="mt-5 flex justify-end border-t border-border pt-4">
                  <Button type="button" variant="primary" onClick={() => advanceStep()}>متابعة إلى معلومات الفاتورة</Button>
                </div>
            </CardContent>
          </Card>}

          {currentStep === 2 && <Card className="purchase-invoice-card border-primary/15">
            <CardHeader className="purchase-invoice-card-header">
              <CardTitle className="purchase-invoice-card-title flex items-center gap-2">
                <FileText className="w-5 h-5 text-cyan" />
                معلومات الفاتورة
              </CardTitle>
              <span className="text-xs text-text-muted">بيانات فاتورة المورد وتواريخها</span>
            </CardHeader>
            <CardContent className="purchase-invoice-card-content">
              <div className="purchase-invoice-fields-grid grid gap-3">
                <div>
                  <label className="block text-sm font-medium text-text mb-2">رقم فاتورة المورد</label>
                  <Input
                    ref={invoiceNumberInputRef}
                    value={invoiceNumber}
                    onChange={(e) => setInvoiceNumber(e.target.value)}
                    placeholder="PO-..."
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-text mb-2">تاريخ فاتورة المورد</label>
                  <Input
                    type="date"
                    value={purchaseDate}
                    onChange={(e) => setPurchaseDate(e.target.value)}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-text mb-2">تاريخ الاستلام المتوقع</label>
                  <Input
                    type="date"
                    value={expectedDate}
                    onChange={(e) => setExpectedDate(e.target.value)}
                  />
                </div>
              </div>
              <div className="purchase-invoice-actions flex justify-end border-t border-border pt-3">
                <Button type="button" variant="primary" onClick={() => advanceStep()}>متابعة إلى المنتجات</Button>
              </div>
            </CardContent>
          </Card>}

          {/* Items Section */}
          {currentStep === 3 && <div className="space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-4">
              <CardTitle className="flex items-center gap-2">
                <ShoppingCart className="w-5 h-5 text-cyan" />
                إضافة القطع
              </CardTitle>
              <div className="text-left">
                <Badge variant="secondary">{items.length} منتج</Badge>
                <div className="mt-1 text-xs text-text-muted">{totalQuantity} قطعة إجمالاً</div>
              </div>
            </div>

            <div className="space-y-3">
              <label className="sr-only" htmlFor="purchase-product-search">ابحث عن منتج أو امسح الباركود</label>
              <div className="relative">
                <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                <Input
                  id="purchase-product-search"
                  ref={productSearchInputRef}
                  placeholder="ابحث عن منتج أو امسح الباركود..."
                  value={productSearchQuery}
                  onChange={(e) => setProductSearchQuery(e.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === 'Escape') {
                      event.preventDefault();
                      event.stopPropagation();
                      setProductSearchQuery('');
                      return;
                    }
                    if (event.key !== 'Enter') return;
                    event.preventDefault();
                    event.stopPropagation();
                    if (!productSearchQuery.trim() && items.length > 0) {
                      advanceStep();
                      return;
                    }
                    void handleQuickProductSearch();
                  }}
                  className="h-11 pr-10"
                  autoFocus
                />
              </div>

              {selectedProductDraft && (
                <div className="rounded-lg border border-primary/20 bg-primary/5 p-3">
                  <div className="mb-3 flex items-center gap-3">
                    <div className="min-w-0 flex-1 truncate text-sm font-semibold text-text-primary">{selectedProductDraft.name}</div>
                    <Badge variant="success" size="sm">جديد</Badge>
                  </div>
                  <div className="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-end gap-2">
                    <div>
                      <label className="mb-1 block text-xs text-text-muted">الكمية</label>
                      <Input ref={draftQuantityInputRef} type="number" min="1" step="1" value={draftQuantity} onChange={(event) => setDraftQuantity(event.target.value)} onKeyDown={(event) => { if (event.key !== 'Enter') return; event.preventDefault(); event.stopPropagation(); draftUnitCostInputRef.current?.focus(); draftUnitCostInputRef.current?.select(); }} className="w-full" />
                    </div>
                    <div>
                      <label className="mb-1 block text-xs text-text-muted">سعر الشراء</label>
                      <Input ref={draftUnitCostInputRef} type="number" min="0" step="0.01" value={draftUnitCost} onChange={(event) => setDraftUnitCost(event.target.value)} onKeyDown={(event) => { if (event.key !== 'Enter') return; event.preventDefault(); event.stopPropagation(); handleAddDraftProduct(); }} className="w-full" />
                    </div>
                    <Button type="button" variant="primary" onClick={handleAddDraftProduct}><Plus className="h-4 w-4" />إضافة</Button>
                  </div>
                </div>
              )}

              {productSearchQuery.trim() && searchedProducts.length > 0 ? (
                <div className="max-h-56 overflow-y-auto rounded-lg border border-border bg-surface-elevated">
                  <div className="border-b border-border px-3 py-2 text-xs font-medium text-text-muted">نتائج البحث</div>
                  {searchedProducts.map((product) => (
                    <button
                      key={product.id}
                      type="button"
                      className="flex w-full items-center justify-between border-b border-border p-3 text-right transition-colors last:border-b-0 hover:bg-primary/5"
                      onClick={() => handleManualAdd(product)}
                    >
                      <span className="min-w-0 truncate font-medium text-text">{product.name}</span>
                      <span className="flex shrink-0 items-center gap-3 text-sm">
                        <span className="font-semibold text-cyan">₪{product.cost_price?.toFixed(2) || '0.00'}</span>
                        <Plus className="h-4 w-4 text-primary" />
                      </span>
                    </button>
                  ))}
                  {productTotalPages > 1 && (
                    <PaginationControls
                      page={productPage}
                      pageSize={10}
                      total={productTotal}
                      onPageChange={setProductPage}
                      isLoading={productsLoading}
                    />
                  )}
                </div>
              ) : (
                <div className="rounded-lg border border-dashed border-border px-4 py-5 text-center text-sm text-text-muted">
                  {productSearchQuery ? (
                    <>
                      <div>لم يتم العثور على منتج بهذا البحث</div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="mt-1 text-primary"
                        onClick={() => {
                          setManualProductData((current) => ({ ...current, name: productSearchQuery, sku: generateSku() }));
                          setEditingProductId(null);
                          setManualProductImage(null);
                          setShowManualProductDetails(false);
                          setIsManualProductModalOpen(true);
                        }}
                      >
                        <Plus className="h-4 w-4" />
                        المنتج غير موجود؟ أضف منتجًا جديدًا
                      </Button>
                    </>
                  ) : 'ابدأ بكتابة اسم المنتج أو امسح الباركود'}
                </div>
              )}
            </div>

              <div className="flex justify-end border-t border-border pt-4">
                <Button type="button" variant="primary" onClick={() => advanceStep()}>متابعة إلى الدفع</Button>
              </div>
          </div>}

          {/* Added items cards */}
          {currentStep === 3 && items.length > 0 && (
            <div className="rounded-2xl border border-border bg-surface">
              <CardHeader className="border-b border-border pb-3">
                <div className="flex items-center justify-between gap-3">
                  <CardTitle>العناصر المضافة ({items.length})</CardTitle>
                  <span className="text-sm text-text-muted">₪{totalCost.toFixed(2)}</span>
                </div>
              </CardHeader>
              <CardContent className="p-0">
                <div className="divide-y divide-border">
                  {items.map((item) => (
                    <article key={item.key} className="px-4 py-3">
                      <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
                        <div className="min-w-[12rem] flex-1">
                          <div className="truncate font-semibold text-text-primary">{item.product_name}</div>
                        </div>
                        <div className="flex items-center gap-1 text-sm text-text-secondary">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 border border-border"
                            onClick={() => handleUpdateQuantity(item.key, item.quantity - 1)}
                            disabled={item.quantity <= 1}
                            aria-label={`تقليل كمية ${item.product_name}`}
                            title="تقليل الكمية"
                          >
                            <Minus className="h-3.5 w-3.5" />
                          </Button>
                          <span className="min-w-8 text-center font-semibold">×{item.quantity}</span>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 border border-border"
                            onClick={() => handleUpdateQuantity(item.key, item.quantity + 1)}
                            aria-label={`زيادة كمية ${item.product_name}`}
                            title="زيادة الكمية"
                          >
                            <Plus className="h-3.5 w-3.5" />
                          </Button>
                          <span className="text-text-muted">شراء ₪{Number(item.unit_cost).toFixed(2)}</span>
                        </div>
                        <span className="text-sm font-semibold text-text-primary">الإجمالي ₪{(item.quantity * item.unit_cost).toFixed(2)}</span>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={() => setExpandedItemKey((current) => current === item.key ? null : item.key)}
                          className="shrink-0"
                          aria-label={`تفاصيل ${item.product_name}`}
                          title="تعديل أو تفاصيل"
                        >
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </div>

                      {expandedItemKey === item.key && (
                        <div className="mt-3 grid grid-cols-1 gap-3 rounded-lg border border-border bg-surface-elevated/40 p-3 sm:grid-cols-2 lg:grid-cols-4">
                          <div>
                            <label className="mb-1 block text-xs text-text-muted">الكمية</label>
                            <Input
                              ref={(element) => { quantityInputRefs.current[item.key] = element; }}
                              type="number"
                              value={item.quantity}
                              onChange={(e) => handleUpdateQuantity(item.key, parseInt(e.target.value) || 0)}
                              className="w-full"
                              min="1"
                            />
                          </div>
                          <div>
                            <label className="mb-1 block text-xs text-text-muted">سعر الشراء</label>
                            <Input
                              ref={(element) => { unitCostInputRefs.current[item.key] = element; }}
                              type="number"
                              value={item.unit_cost}
                              onChange={(e) => handleUpdateUnitCost(item.key, parseFloat(e.target.value) || 0)}
                              className="w-full"
                              min="0"
                              step="0.01"
                            />
                          </div>
                          <div>
                            <label className="mb-1 block text-xs text-text-muted">الحالة</label>
                            <Select
                              value={item.condition}
                              onChange={(e) => handleUpdateCondition(item.key, e.target.value as 'new' | 'used' | 'refurbished')}
                              options={[
                                { value: 'new', label: 'جديد' },
                                { value: 'used', label: 'مستعمل' },
                                { value: 'refurbished', label: 'مجدد' },
                              ]}
                              className="w-full"
                            />
                          </div>
                          <div>
                            <label className="mb-1 block text-xs text-text-muted">سعر البيع (اختياري)</label>
                            <Input
                              type="number"
                              value={item.selling_price ?? 0}
                              onChange={(e) => setItems((prev) => prev.map((current) => current.key === item.key ? { ...current, selling_price: parseFloat(e.target.value) || 0 } : current))}
                              className="w-full"
                              min="0"
                              step="0.01"
                            />
                          </div>
                          <div className="flex items-center justify-between gap-2 sm:col-span-2 lg:col-span-4">
                            <div className="min-w-0 text-xs text-text-muted">
                              {item.sku && <div className="truncate">SKU: {item.sku}</div>}
                              {item.barcode && <div className="truncate">الباركود: {item.barcode}</div>}
                            </div>
                            <Select
                              value={item.category_id ?? ''}
                              onChange={(e) => setItems((prev) => prev.map((current) => current.key === item.key ? { ...current, category_id: e.target.value } : current))}
                              options={[
                                { value: '', label: 'بدون تصنيف' },
                                ...categories.map((category) => ({ value: category.id, label: category.name })),
                              ]}
                              className="max-w-[20rem]"
                              aria-label="التصنيف"
                            />
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              onClick={() => handleRemoveItem(item.key)}
                              className="text-red-600 hover:text-red-700"
                            >
                              <Trash2 className="h-4 w-4" />
                              حذف
                            </Button>
                          </div>
                        </div>
                      )}
                    </article>
                  ))}
                </div>
              </CardContent>
            </div>
          )}

        </div>

        {/* Right Column - Summary and Actions */}
        <div className="grid grid-cols-1 gap-6">
          {/* Quick Stats */}
          {currentStep === 5 && <Card className="border-primary/20 bg-primary/5">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Package className="w-5 h-5 text-cyan" />
                ملخص الطلب
              </CardTitle>
              <Badge variant={selectedSupplier ? 'success' : 'warning'}>
                {selectedSupplier ? 'المورد محدد' : 'اختر المورد'}
              </Badge>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-text-muted">عدد العناصر</span>
                <span className="font-semibold">{items.length}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-text-muted">إجمالي القطع</span>
                <span className="font-semibold">{totalQuantity}</span>
              </div>
                <div className="border-t border-border pt-4">
                <div className="flex justify-between items-center">
                  <span className="text-text-muted">الإجمالي قبل الضريبة</span>
                  <span className="font-semibold">₪{totalCost.toFixed(2)}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-text-muted">ضريبة المورد ({taxRate}%)</span>
                  <span className="font-semibold">₪{taxAmount.toFixed(2)}</span>
                </div>
                <div className="flex justify-between items-center border-t border-border pt-3">
                  <span className="text-text-muted">{taxRate > 0 ? 'الإجمالي شامل الضريبة' : 'الإجمالي'}</span>
                  <span className="text-2xl font-bold text-cyan">₪{totalWithTax.toFixed(2)}</span>
                </div>
                <div className="mt-3 grid grid-cols-2 gap-3 border-t border-border pt-3 text-sm">
                  <div className="flex justify-between"><span className="text-text-muted">المدفوع</span><span className="font-semibold">₪{(Number(initialPayment) || 0).toFixed(2)}</span></div>
                  <div className="flex justify-between"><span className="text-text-muted">المتبقي</span><span className="font-semibold">₪{Math.max(0, totalWithTax - (Number(initialPayment) || 0)).toFixed(2)}</span></div>
                </div>
                <div className="mt-3 space-y-2 border-t border-border pt-3">
                  {items.map((item) => (
                    <div key={item.key} className="flex items-center justify-between gap-3 text-sm">
                      <span className="min-w-0 truncate text-text-secondary">{item.product_name}</span>
                      <span className="shrink-0 text-text-muted">{item.quantity} × ₪{Number(item.unit_cost).toFixed(2)}</span>
                    </div>
                  ))}
                </div>
                {selectedSupplier && (() => {
                  const supplier = suppliers.find((entry) => entry.id === selectedSupplier);
                  return supplier ? (
                    <div className="mt-3 border-t border-border pt-3 text-sm text-text-secondary">
                      المورد: <span className="font-semibold text-text-primary">{supplier.name}</span>
                      {supplier.phone ? ` · ${supplier.phone}` : ''}
                    </div>
                  ) : null;
                })()}
                <div className="mt-3 border-t border-border pt-3">
                  <label className="mb-2 block text-sm font-medium text-text">ملاحظات (اختياري)</label>
                  <textarea
                    value={notes}
                    onChange={(e) => setNotes(e.target.value)}
                    rows={2}
                    placeholder="أضف ملاحظة إذا لزم الأمر"
                    className="w-full rounded-xl border border-border px-3 py-2 text-sm focus:border-cyan/50 focus:ring-2 focus:ring-cyan/30"
                  />
                </div>
                <div className="mt-4 flex flex-col gap-2 border-t border-border pt-4">
                  <Button ref={createPurchaseButtonRef} variant="primary" onClick={() => { void handleCreatePurchase(); }} disabled={isSubmitting} className="w-full gap-2" size="lg">
                    {isSubmitting ? 'جاري التنفيذ...' : (receiveImmediately ? 'إنشاء واستلام' : 'إنشاء الشراء')}
                  </Button>
                  <Button variant="secondary" onClick={closePurchase} className="w-full">إلغاء</Button>
                </div>
              </div>
            </CardContent>
          </Card>}

          {/* Receive Immediately */}
          {currentStep === 4 && <Card className="border-success/25 bg-success/5">
            <CardContent className="p-4">
              <div className="mb-4 grid grid-cols-3 gap-2 rounded-xl border border-border bg-surface/70 p-3 text-center">
                <div>
                  <div className="text-xs text-text-muted">إجمالي الشراء</div>
                  <div className="mt-1 font-bold text-text-primary">₪{totalWithTax.toFixed(2)}</div>
                </div>
                <div>
                  <div className="text-xs text-text-muted">المدفوع</div>
                  <div className="mt-1 font-bold text-success">₪{(Number(initialPayment) || 0).toFixed(2)}</div>
                </div>
                <div>
                  <div className="text-xs text-text-muted">المتبقي</div>
                  <div className="mt-1 font-bold text-warning">₪{Math.max(0, totalWithTax - (Number(initialPayment) || 0)).toFixed(2)}</div>
                </div>
              </div>
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={receiveImmediately}
                  onChange={(e) => setReceiveImmediately(e.target.checked)}
                  disabled={isEmbedded}
                  className="w-5 h-5 rounded border-gray-300 text-green focus:ring-green"
                />
                <div>
                  <div className="font-medium flex items-center gap-2">
                    <CheckCircle2 className="w-4 h-4 text-green" />
                    استلام مباشر بعد الدفعة
                  </div>
                  <div className="text-sm text-text-muted">
                    عند التفعيل، سيحاول النظام الاستلام بعد إنشاء الشراء، ويتطلب ذلك تسجيل دفعة مسبقة ولو كانت جزئية
                  </div>
                </div>
              </label>
              {(
                <div className="mt-4">
                  <label className="block text-sm font-medium text-text mb-2">
                    مبلغ الدفعة المسبقة {receiveImmediately ? '*' : '(اختياري)'}
                  </label>
                  <Input
                    ref={paymentInputRef}
                    type="number"
                    min="0.01"
                    max={totalWithTax}
                    step="0.01"
                    value={initialPayment}
                    onChange={(e) => setInitialPayment(e.target.value)}
                    onKeyDown={(event) => {
                      if (event.key !== 'Enter') return;
                      event.preventDefault();
                      event.stopPropagation();
                      setCurrentStep(5);
                    }}
                    placeholder={`أدخل دفعة كاملة أو جزئية (الإجمالي ₪${totalCost.toLocaleString('en-US', { maximumFractionDigits: 2 })})`}
                  />
                  <p className="mt-1 text-xs text-text-muted">
                    {receiveImmediately ? 'يجب تسجيل مبلغ أكبر من صفر قبل استلام البضاعة.' : 'يمكن تسجيل دفعة كاملة أو جزئية الآن.'}
                  </p>
                </div>
              )}
              <div className="mt-5 flex justify-end border-t border-border pt-4">
                <Button type="button" variant="primary" onClick={() => advanceStep()}>متابعة إلى المراجعة</Button>
              </div>
            </CardContent>
          </Card>}

        </div>
      </div>

      </div>
    </Modal>

      {/* Manual Product Creation Modal */}
      <Modal
        isOpen={isManualProductModalOpen}
        onClose={() => { setIsManualProductModalOpen(false); setEditingProductId(null); setShowManualProductDetails(false); }}
        title={editingProductId ? 'تعديل بيانات القطعة' : 'إضافة قطعة جديدة'}
        variant="modern"
        size="lg"
      >
        <div className="rounded-xl border border-border bg-surface p-3">
          {/* Basic Information */}
          <div style={{ 
            marginBottom: '12px',
            paddingBottom: '12px',
            borderBottom: '1px solid var(--border-subtle)'
          }}>
            <div className="manual-product-basic-grid grid gap-2">
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
                  ref={manualNameInputRef}
                  value={manualProductData.name}
                  onChange={(e) => setManualProductData({ ...manualProductData, name: e.target.value })}
                  onKeyDown={(event) => focusManualField(event, manualSkuInputRef.current)}
                  placeholder="أدخل اسم المنتج"
                  autoFocus
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
                  SKU
                </label>
                <Input 
                  ref={manualSkuInputRef}
                  value={manualProductData.sku}
                  readOnly
                  onKeyDown={(event) => focusManualField(event, manualBarcodeInputRef.current)}
                  placeholder="SKU-..."
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
                  الباركود
                </label>
                <Input 
                  ref={manualBarcodeInputRef}
                  value={manualProductData.barcode}
                  onChange={(e) => setManualProductData({ ...manualProductData, barcode: e.target.value })}
                  onKeyDown={(event) => focusManualField(event, manualCategoryInputRef.current)}
                  placeholder="أدخل الباركود"
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
                  التصنيف
                </label>
                <div className="flex items-center gap-1">
                  <Select
                    ref={manualCategoryInputRef}
                    value={manualProductData.category_id}
                    onChange={(e) => setManualProductData({ ...manualProductData, category_id: e.target.value })}
                    onKeyDown={(event) => focusManualField(event, manualCostInputRef.current)}
                    options={[
                      { value: '', label: 'اختر التصنيف...' },
                      ...categories.map((c) => ({ value: c.id, label: c.name })),
                    ]}
                    emptyMessage="لا يوجد تصنيفات"
                    className="min-w-0"
                  />
                  <Button type="button" variant="ghost" size="icon" className="h-9 w-9 shrink-0" onClick={() => setQuickCreateMode('category')} aria-label="إضافة تصنيف" title="إضافة تصنيف">
                    +
                  </Button>
                </div>
              </div>
            </div>
          </div>

          {showManualProductDetails && <div className="mb-4 border-t border-border pt-4">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div className="text-sm font-semibold text-text-primary">معلومات إضافية</div>
              <span className="text-xs text-text-muted">اختياري</span>
            </div>

            <div className="relative overflow-hidden rounded-2xl border border-border bg-surface-elevated p-2 shadow-sm">
              {manualProductImage ? (
                <div className="relative flex min-h-44 items-center justify-center overflow-hidden rounded-xl bg-surface">
                  <img
                    src={manualProductImage}
                    alt="معاينة صورة المنتج"
                    className="h-44 w-full object-contain"
                  />
                  <div className="absolute inset-x-3 bottom-3 flex items-center justify-between rounded-xl border border-white/20 bg-black/60 px-3 py-2 text-white backdrop-blur-sm">
                    <span className="flex items-center gap-2 text-xs font-medium">
                      <ImagePlus className="h-4 w-4" />
                      صورة جاهزة للحفظ
                    </span>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={() => setManualProductImage(null)}
                      className="h-8 w-8 text-white hover:bg-white/15 hover:text-white"
                      aria-label="حذف صورة المنتج"
                      title="حذف الصورة"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ) : (
                <label
                  htmlFor="product-image-upload"
                  className="group flex min-h-44 cursor-pointer flex-col items-center justify-center rounded-xl border-2 border-dashed border-primary/25 bg-primary/[0.03] px-6 text-center transition-colors hover:border-primary/50 hover:bg-primary/[0.07]"
                >
                  <span className="mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 text-primary transition-transform group-hover:scale-105">
                    <Upload className="h-6 w-6" />
                  </span>
                  <span className="text-sm font-semibold text-text-primary">ارفع صورة المنتج</span>
                  <span className="mt-1 text-xs text-text-muted">JPG أو PNG أو WebP</span>
                  <span className="mt-3 inline-flex items-center gap-1.5 rounded-lg border border-border bg-surface px-3 py-1.5 text-xs font-semibold text-text-secondary shadow-sm">
                    <ImagePlus className="h-3.5 w-3.5" />
                    اختيار صورة
                  </span>
                </label>
              )}
              <Input
                id="product-image-upload"
                type="file"
                accept="image/jpeg,image/png,image/webp"
                className="sr-only"
                onChange={async (event) => {
                  const file = event.target.files?.[0];
                  if (!file) return;
                  try {
                    setManualProductImage(await compressProductImage(file));
                  } catch (error) {
                    toast.error(error instanceof Error ? error.message : 'تعذر تجهيز الصورة');
                  }
                  event.target.value = '';
                }}
                aria-label="رفع صورة المنتج"
              />
            </div>
          </div>}

          <div className="border-t border-border pt-3">
            <div className="manual-product-pricing-grid grid gap-2">
              <div>
                <label style={{ 
                  fontSize: '12px', 
                  fontWeight: '600', 
                  color: 'var(--text-secondary)',
                  marginBottom: '8px',
                  display: 'block',
                  letterSpacing: '0.2px'
                }}>
                  سعر التكلفة
                  <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                </label>
                <Input
                  ref={manualCostInputRef}
                  type="number"
                  value={manualProductData.cost_price}
                  onChange={(e) => handleCostPriceChange(e.target.value)}
                  onKeyDown={(event) => focusManualField(event, manualSellingInputRef.current)}
                  placeholder="0.00"
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
                  سعر البيع
                  <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
                </label>
                <Input
                  ref={manualSellingInputRef}
                  type="number"
                  value={manualProductData.selling_price}
                  onChange={(e) => setManualProductData({ ...manualProductData, selling_price: e.target.value })}
                  onKeyDown={(event) => focusManualField(event, manualMinStockInputRef.current)}
                  placeholder="0.00"
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
                  ref={manualMinStockInputRef}
                  type="number"
                  min="0"
                  step="1"
                  value={manualProductData.min_stock}
                  onChange={(e) => setManualProductData({ ...manualProductData, min_stock: e.target.value })}
                  onKeyDown={(event) => {
                    if (event.key !== 'Enter') return;
                    event.preventDefault();
                    void handleManualProductCreate();
                  }}
                  placeholder="0"
                  aria-label="الحد الأدنى للمخزون"
                />
              </div>
            </div>
          </div>

          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="mt-4 w-full justify-start border-t border-border pt-4 text-text-secondary"
            onClick={() => setShowManualProductDetails((current) => !current)}
          >
            <Box className="h-4 w-4" />
            {showManualProductDetails ? 'إخفاء المعلومات الإضافية' : '+ معلومات إضافية'}
          </Button>

          {showManualProductDetails && (
            <div>
              <label className="mb-2 block text-xs font-semibold text-text-secondary">الوصف</label>
              <Input
                value={manualProductData.description}
                onChange={(e) => setManualProductData({ ...manualProductData, description: e.target.value })}
                placeholder="أدخل وصف المنتج"
              />
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-3 justify-end pt-4">
            <Button
              variant="secondary"
              onClick={() => { setIsManualProductModalOpen(false); setEditingProductId(null); setShowManualProductDetails(false); }}
              disabled={createProductMutation.isPending || updateProductMutation.isPending}
            >
              إلغاء
            </Button>
            <Button
              variant="primary"
              onClick={handleManualProductCreate}
              disabled={createProductMutation.isPending || updateProductMutation.isPending}
              className="gap-2"
            >
              {createProductMutation.isPending || updateProductMutation.isPending ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white" />
                  {editingProductId ? 'جاري الحفظ...' : 'جاري الإضافة...'}
                </>
              ) : (
                <>
                  <Sparkles className="w-4 h-4" />
                  {editingProductId ? 'حفظ التعديلات' : 'إضافة للشراء والمخزون'}
                </>
              )}
            </Button>
          </div>
        </div>
      </Modal>
      <ConfirmDialog
        isOpen={showUnsavedExitDialog}
        onClose={() => setShowUnsavedExitDialog(false)}
        onConfirm={completeClosePurchase}
        title="الخروج من عملية الشراء؟"
        message="لديك بيانات غير محفوظة. يمكنك متابعة التحرير أو الخروج دون حفظ هذه العملية."
        confirmText="خروج دون حفظ"
        cancelText="متابعة التحرير"
        variant="warning"
      />
      <ConfirmDialog
        isOpen={productToDelete !== null}
        onClose={() => setProductToDelete(null)}
        onConfirm={() => {
          if (productToDelete) {
            deleteProductMutation.mutate(productToDelete.id);
          }
          setProductToDelete(null);
        }}
        title="حذف المنتج"
        message="سيُحذف المنتج نهائيًا مع تنظيف المبيعات والمشتريات والمرتجعات المرتبطة به من البطاقات والتقارير. هل تريد المتابعة؟"
        confirmText="حذف المنتج"
        isLoading={deleteProductMutation.isPending}
        variant="danger"
      />
      <InventoryQuickCreateModal
        mode={quickCreateMode || 'category'}
        isOpen={quickCreateMode !== null}
        onClose={() => setQuickCreateMode(null)}
        onCreated={(record) => {
          if (quickCreateMode === 'supplier') {
            setSelectedSupplier(record.id);
          } else {
            setManualProductData((current) => ({ ...current, category_id: record.id }));
          }
          void queryClient.invalidateQueries({ queryKey: quickCreateMode === 'supplier' ? ['suppliers'] : ['categories'] });
          setQuickCreateMode(null);
        }}
      />
    </>
  );
}
