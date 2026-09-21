import { useState, useEffect } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { useNavigate, useLocation } from 'react-router-dom';
import { cn } from '../../../utils';
import { PageHeader } from '../../../design-system/components/page-header';
import { Button } from '../../../design-system/components/button';
import { Modal } from '../../../design-system/components/modal';
import { Input } from '../../../design-system/components/input';
import { getButtonSize } from '../../../config/button-sizes';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { FileText, Plus, Package, PackageOpen, LayoutGrid, List } from 'lucide-react';

// Custom hooks
import { useInventory } from '../hooks/useInventory';
import { useIsMobile } from '../../../hooks/useIsMobile';

// Components
import { InventoryStats } from '../components/InventoryStats';
import { InventoryFilters } from '../components/InventoryFilters';
import { InventoryList } from '../components/InventoryList';
import { StockAlertCards } from '../components/StockAlertCards';
import { InventoryModals } from '../components/InventoryModals';
import { InventoryEntryModal } from '../components/InventoryEntryModal';
import { InventoryLedger } from '../../../design-system/components/inventory-ledger';
import type { InventoryMovement } from '../../../design-system/components/inventory-ledger';
import { ConfirmDialog } from '../../../design-system/components/confirm-dialog';
import { ReportActions } from '../../../design-system/components/report-actions';

// Types
import { ViewMode, Product } from '../types/inventory.types';
import { categoriesApi, inventoryApi, productsApi } from '../../../services/api/endpoints';
import { toast } from 'sonner';
import { getLocalProductImage } from '../../../services/localProductImages';
import { getCategoryImage } from '../../../services/localCategoryImages';
import { generateSku } from '../../../utils/sku';
import { InventoryQuickCreateModal } from '../components/InventoryQuickCreateModal';
import { BulkProductImportModal } from '../components/BulkProductImportModal';
import { CreatePurchasePage } from '../../purchases/pages/CreatePurchasePage';
import { SupplierInvoiceModal } from '../../purchases/components/SupplierInvoiceModal';

interface InventoryMovementResponse {
  id: string;
  movement_type: InventoryMovement['type'];
  quantity: number;
  before_quantity: number;
  after_quantity: number;
  reference_type?: string;
  reference_id?: string;
  customer_name?: string;
  invoice_number?: string;
  reason?: string;
  created_by?: string;
  created_at: string;
}

export function InventoryPage() {
  const queryClient = useQueryClient();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const isMobile = useIsMobile();
  const [viewMode, setViewMode] = useState<ViewMode>('products');
  const [layoutMode, setLayoutMode] = useState<'cards' | 'table'>('cards');
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [returnToSalesAfterSave, setReturnToSalesAfterSave] = useState(false);
  const [isCreatingProduct, setIsCreatingProduct] = useState(false);
  const [isInventoryEntryModalOpen, setIsInventoryEntryModalOpen] = useState(false);
  const [isBulkImportOpen, setIsBulkImportOpen] = useState(false);
  const [isPurchaseWorkflowOpen, setIsPurchaseWorkflowOpen] = useState(false);
  const [completedPurchase, setCompletedPurchase] = useState<any | null>(null);
  const [quickCreateMode, setQuickCreateMode] = useState<'category' | 'supplier' | null>(null);
  const [isProductCategoryPickerOpen, setIsProductCategoryPickerOpen] = useState(false);
  const [pendingProductCategoryId, setPendingProductCategoryId] = useState('');
  const [showInventoryLedger, setShowInventoryLedger] = useState(false);
  const [inventoryMovements, setInventoryMovements] = useState<InventoryMovement[]>([]);
  const [inventoryLedgerLoading, setInventoryLedgerLoading] = useState(false);
  const [inventoryLedgerError, setInventoryLedgerError] = useState<string | null>(null);
  const [inventoryLedgerProduct, setInventoryLedgerProduct] = useState<Product | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [productToDelete, setProductToDelete] = useState<string | null>(null);
  const [inventoryItemToDelete, setInventoryItemToDelete] = useState<string | null>(null);
  const [minimumStockProduct, setMinimumStockProduct] = useState<Product | null>(null);
  const [minimumStockValue, setMinimumStockValue] = useState('0');
  const [reportLoading, setReportLoading] = useState(false);

  // Custom hook
  const {
    inventoryItems,
    filteredProducts,
    filteredInventoryItems,
    productsLoading,
    inventoryLoading,
    inventoryStockMap,
    searchQuery,
    setSearchQuery,
    sortConfig,
    setSortConfig,
    filters,
    setFilters,
    refetch,
    deleteProductMutation,
    archiveProductMutation,
    deleteInventoryItemMutation,
    createProductMutation,
    updateProductMutation,
    updateMinimumStockMutation,
    lookupProduct,
    productPage,
    inventoryPage,
    pageSize,
    productTotal,
    inventoryTotal,
    setProductPage,
    setInventoryPage,
  } = useInventory();
  const supplierOnly = filters.some((filter) => filter.key === 'supplier_only' && String(filter.value).toLowerCase() === 'true');
  const manualOnly = filters.some((filter) => filter.key === 'manual_only' && String(filter.value).toLowerCase() === 'true');
  const inventorySource = manualOnly ? 'manual' : supplierOnly ? 'supplier' : 'all';
  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });
  const inventoryCategories = (categoriesData?.data as Array<{ id: string; name: string }>) || [];

  // Handle edit product from navigation state
  useEffect(() => {
    if (location.state?.editProduct) {
      setSelectedProduct(location.state.editProduct);
      setReturnToSalesAfterSave(Boolean(location.state.returnToSalesAfterSave));
      setIsEditModalOpen(true);
      // Clear the state to prevent reopening on refresh
      navigate(location.pathname, { replace: true, state: null });
    }
  }, [location.state, navigate, location.pathname]);

  useEffect(() => {
    const categoryId = location.state?.createProductCategoryId;
    if (!categoryId) return;
    setSelectedProduct({
      id: '',
      name: '',
      sku: generateSku(),
      sellingPrice: 0,
      costPrice: 0,
      stock: 0,
      condition: 'new',
      category_id: categoryId,
      barcode: '',
    });
    setIsCreatingProduct(true);
    setIsEditModalOpen(true);
    navigate(location.pathname, { replace: true, state: null });
  }, [location.pathname, location.state, navigate]);

  useEffect(() => {
    const search = new URLSearchParams(location.search).get('search') || '';
    setSearchQuery(search);
  }, [location.search, setSearchQuery]);

  const handleClearSearch = () => {
    setSearchQuery('');
  };

  const handleBarcodeScan = async (barcode: string) => {
    const product = await lookupProduct(barcode);
    if (!product) {
      toast.error('لم يتم العثور على منتج بهذا الباركود');
      setSearchQuery('');
      return false;
    }

    setViewMode('products');
    setSearchQuery(product.sku || product.name || barcode);
    return true;
  };

  const handleGeneralInventoryToggle = () => {
    setFilters(filters.filter((filter) => filter.key !== 'supplier_only' && filter.key !== 'manual_only'));
    setViewMode('products');
    setProductPage(1);
    setInventoryPage(1);
    void queryClient.invalidateQueries({ queryKey: ['products'] });
    void queryClient.invalidateQueries({ queryKey: ['inventory'] });
  };

  const handleSupplierInventoryToggle = () => {
    if (supplierOnly) {
      setViewMode('products');
      void refetch();
      return;
    }
    setFilters([...filters.filter((filter) => filter.key !== 'supplier_only' && filter.key !== 'manual_only'), { key: 'supplier_only', value: 'true' }]);
    setViewMode('products');
  };

  const handleManualInventoryToggle = () => {
    if (manualOnly) {
      setViewMode('products');
      void refetch();
      return;
    }
    setFilters([...filters.filter((filter) => filter.key !== 'supplier_only' && filter.key !== 'manual_only'), { key: 'manual_only', value: 'true' }]);
    setViewMode('products');
  };

  const getReportRows = (products: Product[]) => products.map((product: Product) => ({
      'الاسم': product.name,
      'SKU': product.sku,
      'التصنيف': product.category_name || product.category || '-',
      'السعر قبل الضريبة': product.sellingPrice,
      'المخزون': product.stock,
      'الحالة': product.condition
    }));

  const loadAllReportProducts = async (): Promise<Product[]> => {
    const categoryFilter = filters.find((filter) => filter.key === 'category_id');
    const supplierFilter = filters.find((filter) => filter.key === 'supplier_id');
    const purchaseDateFrom = filters.find((filter) => filter.key === 'purchase_date_from');
    const purchaseDateTo = filters.find((filter) => filter.key === 'purchase_date_to');
    const minPurchaseCost = filters.find((filter) => filter.key === 'min_purchase_cost');
    const maxPurchaseCost = filters.find((filter) => filter.key === 'max_purchase_cost');
    const inventoryFilters = {
      page: 1,
      per_page: 1000,
      ...(supplierOnly ? { supplier_only: 'true' } : {}),
      ...(manualOnly ? { manual_only: 'true' } : {}),
      ...(!supplierOnly && !manualOnly ? { exclude_condition: 'USED' } : {}),
      ...(searchQuery ? { search: searchQuery } : {}),
      ...(categoryFilter?.value ? { category_id: categoryFilter.value } : {}),
      ...(supplierFilter?.value ? { supplier_id: supplierFilter.value } : {}),
      ...(purchaseDateFrom?.value ? { purchase_date_from: purchaseDateFrom.value } : {}),
      ...(purchaseDateTo?.value ? { purchase_date_to: purchaseDateTo.value } : {}),
      ...(minPurchaseCost?.value ? { min_purchase_cost: Number(minPurchaseCost.value) } : {}),
      ...(maxPurchaseCost?.value ? { max_purchase_cost: Number(maxPurchaseCost.value) } : {}),
    };

    if (supplierOnly || manualOnly || supplierFilter || purchaseDateFrom || purchaseDateTo || minPurchaseCost || maxPurchaseCost) {
      const [productsResponse, inventoryResponse] = await Promise.all([
        productsApi.list({ page: 1, per_page: 1000, ...(searchQuery ? { search: searchQuery } : {}), ...(categoryFilter?.value ? { category_id: categoryFilter.value } : {}) }),
        inventoryApi.listWithSupplier(inventoryFilters),
      ]);
      const productsById = new Map(
        (((productsResponse as any)?.data?.products ?? []) as Product[]).map((product) => [String(product.id), product]),
      );
      const grouped = new Map<string, Product>();
      for (const item of (((inventoryResponse as any)?.data?.items ?? []) as any[])) {
        const productId = String(item.product_id ?? item.product?.id ?? '').trim();
        if (!productId) continue;
        const baseProduct = productsById.get(productId) ?? item.product ?? {};
        const quantity = Number(item.available_quantity ?? item.current_quantity ?? item.stock ?? item.quantity ?? 1);
        const existing = grouped.get(productId);
        if (existing) {
          existing.stock += Number.isFinite(quantity) ? Math.max(0, quantity) : 0;
          continue;
        }
        grouped.set(productId, {
          ...baseProduct,
          id: productId,
          name: String(baseProduct.name ?? item.product_name ?? '-'),
          sku: String(baseProduct.sku ?? item.item_code ?? item.barcode ?? productId),
          sellingPrice: Number(baseProduct.sellingPrice ?? baseProduct.selling_price ?? item.selling_price ?? item.price ?? 0),
          stock: Number.isFinite(quantity) ? Math.max(0, quantity) : 0,
          condition: String(baseProduct.condition ?? item.condition ?? ''),
          category: baseProduct.category ?? item.category_name,
          category_name: baseProduct.category_name ?? item.category_name,
        });
      }
      return Array.from(grouped.values());
    }

    const response = await productsApi.list({
      page: 1,
      per_page: 1000,
      ...(searchQuery ? { search: searchQuery } : {}),
      ...(categoryFilter?.value ? { category_id: categoryFilter.value } : {}),
    });
    return (((response as any)?.data?.products ?? []) as Product[]);
  };

  const runReportAction = async (action: 'export' | 'print', allResults: boolean) => {
    if (reportLoading) return;
    setReportLoading(true);
    try {
      const products = allResults ? await loadAllReportProducts() : filteredProducts;
      if (products.length === 0) {
        toast.error(action === 'export' ? 'لا توجد منتجات لتصديرها' : 'لا توجد منتجات لطباعتها');
        return;
      }
      const rows = getReportRows(products);
      if (action === 'export') {
        exportToCSV(rows, `inventory-${allResults ? 'all-' : ''}${new Date().toISOString().split('T')[0]}`);
        toast.success(`تم تصدير ${products.length} منتج بنجاح`);
      } else {
        printTable(rows, ['الاسم', 'SKU', 'السعر قبل الضريبة', 'المخزون', 'الحالة'], 'تقرير المخزون');
        toast.success(`تم تجهيز تقرير ${products.length} منتج للطباعة`);
      }
    } catch (error) {
      console.error('Inventory report failed:', error);
      toast.error('تعذر تجهيز تقرير المخزون');
    } finally {
      setReportLoading(false);
    }
  };

  const handleProductsView = () => {
    setViewMode('products');
    if (supplierOnly || manualOnly) {
      void refetch();
    }
  };

  const handleItemsView = () => {
    setViewMode('items');
    if (supplierOnly || manualOnly) {
      void refetch();
    }
  };

  const handleViewProduct = (product: Product) => {
    // Map API response to local Product type with proper field names
    const mappedProduct: Product = {
      id: product.id,
      name: product.name,
      sku: product.sku,
      sellingPrice: product.sellingPrice || Number((product as Record<string, unknown>).selling_price) || 0,
      costPrice: product.costPrice || ('cost_price' in product ? (product as Record<string, unknown>).cost_price as number : 0),
      stock: product.stock,
      condition: product.condition,
      category: product.category,
      category_id: product.category_id,
      price: product.price,
      barcode: product.barcode,
      image_url: product.image_url || getLocalProductImage(product.id) || (product.category_id ? getCategoryImage(product.category_id) : undefined),
      min_stock_level: Number((product as Record<string, unknown>).min_stock_level ?? 0),
      supplier_id: String((product as Record<string, unknown>).supplier_id ?? (product as Record<string, unknown>).preferred_supplier_id ?? (product as Record<string, any>).supplier?.id ?? ''),
      supplier_name: String((product as Record<string, unknown>).supplier_name ?? (product as Record<string, any>).supplier?.name ?? ''),
    };
    setSelectedProduct(mappedProduct);
    setIsViewModalOpen(true);
  };

  const handleAddPurchase = (product: Product) => {
    navigate('/app/purchases/create', {
      state: {
        supplierId: product.supplier_id || undefined,
        product: {
          id: product.id,
          name: product.name,
          cost_price: product.costPrice || Number((product as Record<string, unknown>).cost_price) || 0,
          selling_price: product.sellingPrice || Number((product as Record<string, unknown>).selling_price) || 0,
        },
      },
    });
  };

  const handleEditProduct = (product: Product) => {
    // Map API response to local Product type with proper field names
    const mappedProduct: Product = {
      id: product.id,
      name: product.name,
      sku: product.sku,
      sellingPrice: product.sellingPrice || Number((product as Record<string, unknown>).selling_price) || 0,
      costPrice: product.costPrice || ('cost_price' in product ? (product as Record<string, unknown>).cost_price as number : 0),
      stock: product.stock,
      condition: product.condition,
      category: product.category,
      category_id: product.category_id,
      price: product.price,
      barcode: product.barcode,
      image_url: product.image_url || getLocalProductImage(product.id) || (product.category_id ? getCategoryImage(product.category_id) : undefined),
      min_stock_level: Number((product as Record<string, unknown>).min_stock_level ?? 0),
      supplier_id: String((product as Record<string, unknown>).supplier_id ?? (product as Record<string, unknown>).preferred_supplier_id ?? (product as Record<string, any>).supplier?.id ?? ''),
      supplier_name: String((product as Record<string, unknown>).supplier_name ?? (product as Record<string, any>).supplier?.name ?? ''),
    };
    setSelectedProduct(mappedProduct);
    setIsCreatingProduct(false);
    setIsEditModalOpen(true);
  };

  const handleDeleteProduct = (productId: string) => {
    setProductToDelete(productId);
    setDeleteDialogOpen(true);
  };

  const handleEditMinimumStock = (product: Product) => {
    setMinimumStockProduct(product);
    setMinimumStockValue(String(product.min_stock_level ?? 0));
  };

  const handleSaveMinimumStock = () => {
    if (!minimumStockProduct) return;
    const value = Math.max(0, Math.floor(Number(minimumStockValue) || 0));
    updateMinimumStockMutation.mutate({ id: minimumStockProduct.id, minStockLevel: value }, {
      onSuccess: () => {
        setMinimumStockProduct(null);
        setMinimumStockValue('0');
      },
    });
  };

  const handleConfirmDelete = () => {
    if (productToDelete) {
      deleteProductMutation.mutate(productToDelete);
      setDeleteDialogOpen(false);
      setProductToDelete(null);
    }
  };

  const handleConfirmInventoryItemDelete = () => {
    if (inventoryItemToDelete) {
      deleteInventoryItemMutation.mutate(inventoryItemToDelete);
      setDeleteDialogOpen(false);
      setInventoryItemToDelete(null);
    }
  };

  const handleSaveProduct = async (productData: Product): Promise<boolean> => {
    if (!productData.name?.trim()) {
      toast.error('أدخل اسم المنتج');
      return false;
    }

    if (!productData.sku?.trim()) {
      toast.error('أدخل رمز المنتج SKU');
      return false;
    }

    if (!Number.isFinite(productData.costPrice) || productData.costPrice <= 0) {
      toast.error('أدخل سعر تكلفة أكبر من صفر');
      return false;
    }

    if (!Number.isFinite(productData.sellingPrice) || productData.sellingPrice <= 0) {
      toast.error('أدخل سعر بيع أكبر من صفر');
      return false;
    }

    const barcode = productData.barcode?.trim() || '';
    if (!selectedProduct?.id && barcode) {
      const existingProduct = await lookupProduct(barcode);
      if (existingProduct) {
        toast.error('هذا الباركود مسجل لمنتج موجود بالفعل');
        return false;
      }
    }

    if (selectedProduct && selectedProduct.id) {
      // Update existing product - map to API field names
      const apiData = {
        name: productData.name,
        sku: productData.sku || generateSku(),
        selling_price: productData.sellingPrice,
        cost_price: productData.costPrice,
        min_stock_level: Number(productData.min_stock_level) || 0,
        condition: productData.condition,
        category_id: productData.category_id || null,
        preferred_supplier_id: productData.supplier_id || null,
        barcode: productData.barcode,
      };
      try {
        const currentStockResponse = await productsApi.getStock(selectedProduct.id);
        const currentStock = Number(currentStockResponse?.data?.total_stock ?? 0);
        const requestedStock = Math.max(0, Math.floor(Number(productData.stock) || 0));
        await updateProductMutation.mutateAsync({ id: selectedProduct.id, data: apiData });
        if (requestedStock !== currentStock) {
          await inventoryApi.adjustProductQuantity(
            selectedProduct.id,
            requestedStock,
            'تعديل الكمية الحالية من شاشة المخزون',
          );
        }
        await queryClient.invalidateQueries({ queryKey: ['inventory'] });
        await queryClient.invalidateQueries({ queryKey: ['inventory', 'stats'] });
        await queryClient.invalidateQueries({ queryKey: ['products'] });
        await queryClient.invalidateQueries({ queryKey: ['products', 'inventory-stats'] });
        toast.success('تم تحديث المنتج والكمية بنجاح');
        setIsEditModalOpen(false);
        setSelectedProduct(null);
        setIsCreatingProduct(false);
        setReturnToSalesAfterSave(false);
        return true;
      } catch (error: any) {
        console.error('Update product or quantity failed:', error);
        toast.error(error?.arabicMessage || error?.message || 'تعذر تحديث المنتج أو الكمية');
        return false;
      }
    } else {
      // Add new product - map to API field names
      const apiData = {
        name: productData.name,
        sku: productData.sku || generateSku(),
        selling_price: productData.sellingPrice,
        cost_price: productData.costPrice,
        min_stock_level: Number(productData.min_stock_level) || 0,
        stock: productData.stock,
        condition: productData.condition,
        category_id: productData.category_id || null,
        barcode: productData.barcode,
      };
      createProductMutation.mutate(apiData, {
        onSuccess: async (response: any) => {
          const product = response?.data?.product ?? response?.data;
          setIsEditModalOpen(false);
          setSelectedProduct(null);
          setIsCreatingProduct(false);
          const quantity = Math.max(0, Math.floor(Number(productData.stock) || 0));
          if (!product?.id) {
            return;
          }

          await queryClient.invalidateQueries({ queryKey: ['inventory'] });
          await queryClient.invalidateQueries({ queryKey: ['products'] });
          if (quantity === 0) {
            toast.success('تمت إضافة المنتج بنجاح');
            return;
          }

          const openingStock = {
            product_id: product.id,
            mode: 'quantity' as const,
            quantity,
            business_date: new Date().toISOString().slice(0, 10),
            condition: productData.condition === 'used' ? 'USED' : 'NEW',
            purchase_cost: Number(productData.costPrice) || 0,
            selling_price: Number(productData.sellingPrice) || 0,
            notes: 'إضافة منتج بدون فاتورة',
          };
          try {
            await inventoryApi.createOpeningStock(openingStock);
            await queryClient.invalidateQueries({ queryKey: ['inventory'] });
            await queryClient.invalidateQueries({ queryKey: ['products'] });
            toast.success('تمت إضافة المنتج والكمية بدون فاتورة');
          } catch (error) {
            console.error('Failed to create no-invoice inventory quantity:', error);
            await queryClient.invalidateQueries({ queryKey: ['inventory'] });
            await queryClient.invalidateQueries({ queryKey: ['products'] });
            toast.error('تم إنشاء المنتج لكن تعذرت إضافة الكمية للمخزون');
          }
        },
      });
      return false;
    }
  };

  const handleSaveProductAndReturnToSales = async (productData: Product) => {
    const saved = await handleSaveProduct(productData);
    if (saved) navigate('/app/sales');
  };

  const handleManualAdd = () => {
    setPendingProductCategoryId('');
    setIsProductCategoryPickerOpen(true);
  };

  const openProductWithCategory = (categoryId: string) => {
    setSelectedProduct({
      id: '',
      name: '',
      sku: generateSku(),
      sellingPrice: 0,
      costPrice: 0,
      stock: 0,
      condition: 'new',
      category_id: categoryId,
    });
    setIsProductCategoryPickerOpen(false);
    setIsCreatingProduct(true);
    setIsEditModalOpen(true);
  };

  const handleCreatePurchase = () => {
    setIsInventoryEntryModalOpen(false);
    setIsPurchaseWorkflowOpen(true);
  };

  const handlePurchaseWorkflowComplete = async (response?: any) => {
    const completed = response?.purchase ? response : response?.data?.purchase ? response.data : response?.data || response;
    setCompletedPurchase(completed?.purchase ? completed : { purchase: completed });
    setIsPurchaseWorkflowOpen(false);
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['products'] }),
      queryClient.invalidateQueries({ queryKey: ['inventory'] }),
      queryClient.invalidateQueries({ queryKey: ['purchases'] }),
      queryClient.invalidateQueries({ queryKey: ['dashboard'] }),
    ]);
    await refetch();
  };

  const handleStockAlertClick = (alert: 'out_of_stock' | 'low_stock') => {
    const isActive = filters.some((filter) => filter.key === alert);
    setSearchQuery('');
    setFilters(isActive ? [] : [{ key: alert, value: 'true' }]);
    setProductPage(1);
    setInventoryPage(1);
    setViewMode('products');
  };

  const handleRefresh = () => {
    setSortConfig({ key: '', direction: null });
    setFilters([]);
    void refetch();
  };

  const handleViewInventoryLedger = async (productId: string) => {
    const product = filteredProducts.find((item) => item.id === productId) ?? null;
    const productInventoryItems = (inventoryItems as any[]).filter((item: any) =>
      item.product_id === productId ||
      item.product?.id === productId ||
      item.productId === productId ||
      item.product_name === product?.name
    );
    const itemIds = [...new Set(
      productInventoryItems
        .map((item: any) => item.id)
        .filter((id): id is string => typeof id === 'string' && id.trim().length > 0)
    )];

    setInventoryLedgerLoading(true);
    setInventoryLedgerError(null);
    setShowInventoryLedger(true);
    setInventoryLedgerProduct(product);

    try {
      if (itemIds.length === 0) {
        setInventoryMovements([]);
        setInventoryLedgerError('لا توجد عناصر مخزون فردية لهذا المنتج لعرض سجل الحركات.');
        return;
      }

      const responses = await Promise.all(itemIds.map(async (itemId) => {
        try {
          return await inventoryApi.movements<{ movements?: InventoryMovementResponse[] }>(itemId);
        } catch (error: any) {
          const message = error?.message || '';
          if (error?.status === 404 || message.includes('inventory item not found')) {
            return { data: { movements: [] } } as any;
          }
          throw error;
        }
      }));
      const movements = responses
        .flatMap((response) => response.data?.movements ?? [])
        .sort((left, right) => new Date(right.created_at).getTime() - new Date(left.created_at).getTime());
      setInventoryMovements(movements.map((movement) => ({
        id: movement.id,
        type: movement.movement_type,
        quantity: movement.quantity,
        beforeQuantity: movement.before_quantity,
        afterQuantity: movement.after_quantity,
        referenceType: movement.reference_type,
        referenceId: movement.reference_id,
        customerName: movement.customer_name,
        invoiceNumber: movement.invoice_number,
        reason: movement.reason,
        createdBy: movement.created_by,
        date: movement.created_at,
      })));
    } catch (error) {
      console.error('Error loading inventory movements:', error);
      setInventoryMovements([]);
      setInventoryLedgerError('تعذر تحميل سجل حركات المخزون. حاول مرة أخرى.');
    } finally {
      setInventoryLedgerLoading(false);
    }
  };

  const handleCloseInventoryLedger = () => {
    setShowInventoryLedger(false);
    setInventoryMovements([]);
    setInventoryLedgerProduct(null);
    setInventoryLedgerError(null);
  };

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        title={t('inventory.title')}
        description="إدارة المخزون والقطع مع تحليلات فورية"
        actions={
          <div className={cn(
            "flex gap-2",
            isMobile ? "flex-col w-full" : ""
          )}>
            <Button
              variant="primary"
              size={getButtonSize('inventory', 'headerActions')}
              className={cn(isMobile ? "w-full" : "")}
              onClick={() => setIsInventoryEntryModalOpen(true)}
            >
              <Plus className="w-3.5 h-3.5 me-1.5" />
              إضافة
            </Button>
            <div className={cn('flex items-center gap-1 rounded-[8px] border border-border bg-surface-muted p-1', isMobile ? 'w-full' : '')} aria-label="مصدر المخزون">
              <Button
                variant={inventorySource === 'all' ? 'primary' : 'ghost'}
                size={getButtonSize('inventory', 'headerActions')}
                className={cn(isMobile ? 'flex-1' : '')}
                onClick={handleGeneralInventoryToggle}
              >
                <Package className="w-3.5 h-3.5 me-1.5" />
                الكل
              </Button>
              <Button
                variant={inventorySource === 'manual' ? 'primary' : 'ghost'}
                size={getButtonSize('inventory', 'headerActions')}
                className={cn(isMobile ? 'flex-1' : '')}
                onClick={handleManualInventoryToggle}
              >
                <PackageOpen className="w-3.5 h-3.5 me-1.5" />
                المضاف يدويًا
              </Button>
              <Button
                variant={inventorySource === 'supplier' ? 'primary' : 'ghost'}
                size={getButtonSize('inventory', 'headerActions')}
                className={cn(isMobile ? 'flex-1' : '')}
                onClick={handleSupplierInventoryToggle}
              >
                <Package className="w-3.5 h-3.5 me-1.5" />
                مشتريات التجار
              </Button>
            </div>
            <ReportActions
              onExportCurrent={() => { void runReportAction('export', false); }}
              onPrintCurrent={() => { void runReportAction('print', false); }}
              onExportAll={() => { void runReportAction('export', true); }}
              onPrintAll={() => { void runReportAction('print', true); }}
              loading={reportLoading}
            />
          </div>
        }
      />

      {/* Inventory Stats */}
      <InventoryStats 
        products={filteredProducts}
        inventoryItems={inventoryItems}
        supplierOnly={supplierOnly}
        manualOnly={manualOnly}
        isMobile={isMobile}
      />

      <InventoryEntryModal
        isOpen={isInventoryEntryModalOpen}
        onClose={() => setIsInventoryEntryModalOpen(false)}
        onAddProduct={handleManualAdd}
        onAddCategory={() => setQuickCreateMode('category')}
        onAddSupplier={() => setQuickCreateMode('supplier')}
        onCreatePurchase={handleCreatePurchase}
        onBulkImport={() => setIsBulkImportOpen(true)}
      />

      <BulkProductImportModal
        isOpen={isBulkImportOpen}
        onClose={() => setIsBulkImportOpen(false)}
        onImported={async () => {
          await Promise.all([
            queryClient.invalidateQueries({ queryKey: ['products'] }),
            queryClient.invalidateQueries({ queryKey: ['inventory'] }),
            queryClient.invalidateQueries({ queryKey: ['dashboard'] }),
          ]);
        }}
      />

      <Modal
        isOpen={isProductCategoryPickerOpen}
        onClose={() => setIsProductCategoryPickerOpen(false)}
        title="اختر تصنيف المنتج"
        variant="modern"
        size="sm"
      >
        <div className="space-y-4">
          <p className="text-sm text-text-secondary">المنتج جزء من منظومة التصنيف. اختر تصنيفًا قبل إدخال بياناته.</p>
          <select
            autoFocus
            value={pendingProductCategoryId}
            onChange={(event) => setPendingProductCategoryId(event.target.value)}
            className="pf-select-control w-full rounded-xl border border-border bg-surface px-3"
          >
            <option value="">اختر التصنيف...</option>
            {inventoryCategories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}
          </select>
          <div className="flex items-center justify-between gap-2">
            <Button variant="ghost" onClick={() => setQuickCreateMode('category')}>+ إضافة تصنيف</Button>
            <div className="flex gap-2">
              <Button variant="secondary" onClick={() => setIsProductCategoryPickerOpen(false)}>إلغاء</Button>
              <Button variant="primary" disabled={!pendingProductCategoryId} onClick={() => openProductWithCategory(pendingProductCategoryId)}>متابعة للمنتج</Button>
            </div>
          </div>
        </div>
      </Modal>

      <InventoryQuickCreateModal
        mode={quickCreateMode || 'category'}
        isOpen={quickCreateMode !== null}
        onClose={() => setQuickCreateMode(null)}
        onCreated={(record) => {
          void queryClient.invalidateQueries({ queryKey: ['categories'] });
          void queryClient.invalidateQueries({ queryKey: ['suppliers'] });
          if (quickCreateMode === 'category' && isProductCategoryPickerOpen) {
            openProductWithCategory(record.id);
          }
          setQuickCreateMode(null);
        }}
      />

      {isPurchaseWorkflowOpen && (
        <CreatePurchasePage
          isOpen
          onClose={() => setIsPurchaseWorkflowOpen(false)}
          onComplete={handlePurchaseWorkflowComplete}
        />
      )}

      <SupplierInvoiceModal
        isOpen={completedPurchase !== null}
        onClose={() => setCompletedPurchase(null)}
        purchase={completedPurchase?.purchase}
        supplier={completedPurchase?.supplier}
        items={completedPurchase?.items || []}
      />

      {/* Inventory Filters */}
      <InventoryFilters
        searchQuery={searchQuery}
        setSearchQuery={setSearchQuery}
        onClearSearch={handleClearSearch}
        onBarcodeScan={handleBarcodeScan}
        filters={filters}
        setFilters={setFilters}
        sortConfig={sortConfig}
        setSortConfig={setSortConfig}
        onRefresh={handleRefresh}
        isMobile={isMobile}
      />

      {/* View Toggle */}
      <div className="inventory-command-bar" style={{ marginBottom: '16px' }}>
        <div className="inventory-command-group">
          <Button
            variant={viewMode === 'products' ? 'primary' : 'secondary'}
            size="sm"
            onClick={handleProductsView}
            className="inventory-command-button"
          >
            <Package className="h-3.5 w-3.5" />
            {t('products.title')}
          </Button>
          <Button
            variant={viewMode === 'items' ? 'primary' : 'secondary'}
            size="sm"
            onClick={handleItemsView}
            className="inventory-command-button"
          >
            <PackageOpen className="h-3.5 w-3.5" />
            {t('inventory.items')}
          </Button>
          <Button
            variant={showInventoryLedger ? 'primary' : 'secondary'}
            size="sm"
            onClick={() => {
              if (showInventoryLedger) {
                handleCloseInventoryLedger();
              } else if (filteredProducts.length > 0) {
                handleViewInventoryLedger(filteredProducts[0].id);
              }
            }}
            className="inventory-command-button"
          >
            <FileText className="h-3.5 w-3.5" />
            سجل الحركات
          </Button>
        </div>
        <div className="inventory-command-group inventory-layout-actions" aria-label="طريقة عرض المخزون">
          <Button
            variant={layoutMode === 'cards' ? 'primary' : 'secondary'}
            size="sm"
            onClick={() => setLayoutMode('cards')}
            aria-label="عرض البطاقات"
            title="عرض البطاقات"
            className="inventory-icon-button"
          >
            <LayoutGrid className="h-4 w-4" />
          </Button>
          <Button
            variant={layoutMode === 'table' ? 'primary' : 'secondary'}
            size="sm"
            onClick={() => setLayoutMode('table')}
            aria-label="عرض الجدول"
            title="عرض الجدول"
            className="inventory-icon-button"
          >
            <List className="h-4 w-4" />
          </Button>
        </div>
        <StockAlertCards
          products={filteredProducts}
          inventoryStockMap={inventoryStockMap}
          onAlertClick={handleStockAlertClick}
          activeAlert={filters.find((filter) => filter.key === 'out_of_stock' || filter.key === 'low_stock')?.key as 'out_of_stock' | 'low_stock' | undefined}
        />
      </div>

      {/* Inventory List */}
      <InventoryList
        viewMode={viewMode}
        filteredProducts={filteredProducts}
        filteredInventoryItems={filteredInventoryItems}
        productsLoading={productsLoading}
        inventoryLoading={inventoryLoading}
        inventoryStockMap={inventoryStockMap}
        searchQuery={searchQuery}
        onViewProduct={handleViewProduct}
              onAddPurchase={handleAddPurchase}
        onEditProduct={handleEditProduct}
        onEditMinimumStock={handleEditMinimumStock}
        onDeleteProduct={handleDeleteProduct}
        onArchiveProduct={(productId) => archiveProductMutation.mutate(productId)}
        onDeleteInventoryItem={(itemId) => { setInventoryItemToDelete(itemId); setDeleteDialogOpen(true); }}
        onClearSearch={handleClearSearch}
        onReorderFromSupplier={(supplierId, productName) => {
          // Navigate to purchases page with pre-filled supplier
          navigate('/app/purchases', { state: { supplierId, productName } });
        }}
        onViewInvoice={(supplierId) => {
          // Navigate to purchases page filtered by supplier
          navigate('/app/purchases', { state: { supplierId } });
        }}
        onViewInventoryLedger={handleViewInventoryLedger}
        pagination={viewMode === 'products'
          ? { page: productPage, pageSize, total: productTotal, onPageChange: setProductPage }
          : { page: inventoryPage, pageSize, total: inventoryTotal, onPageChange: setInventoryPage }}
        layoutMode={layoutMode}
        supplierOnly={supplierOnly}
      />

      {/* Inventory Ledger - Conditionally rendered */}
      <div className={cn(
        'mt-6 grid gap-4',
        showInventoryLedger && 'xl:grid-cols-[minmax(0,2fr)_minmax(280px,1fr)]'
      )}>
        {showInventoryLedger ? (
          <InventoryLedger
            movements={inventoryMovements}
            title={`سجل حركات المخزون: ${inventoryLedgerProduct?.name || 'المنتج المحدد'}`}
            currentStock={inventoryLedgerProduct?.stock}
            isLoading={inventoryLedgerLoading}
            error={inventoryLedgerError}
          />
        ) : null}
      </div>

      {/* Inventory Modals */}
      <InventoryModals
        isViewModalOpen={isViewModalOpen}
        setIsViewModalOpen={setIsViewModalOpen}
        isEditModalOpen={isEditModalOpen}
        setIsEditModalOpen={setIsEditModalOpen}
        isCreatingProduct={isCreatingProduct}
        selectedProduct={selectedProduct}
        setSelectedProduct={setSelectedProduct}
        onSaveProduct={handleSaveProduct}
        onSaveProductAndReturnToSales={returnToSalesAfterSave ? handleSaveProductAndReturnToSales : undefined}
      />

      <Modal
        isOpen={Boolean(minimumStockProduct)}
        onClose={() => setMinimumStockProduct(null)}
        title="تعديل الحد الأدنى للمخزون"
        size="sm"
      >
        <div className="space-y-4">
          <p className="text-sm text-text-secondary">{minimumStockProduct?.name || 'المنتج'}</p>
          <Input
            label="الحد الأدنى للمخزون"
            type="number"
            min="0"
            step="1"
            value={minimumStockValue}
            onChange={(event) => setMinimumStockValue(event.target.value)}
          />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setMinimumStockProduct(null)}>إلغاء</Button>
            <Button variant="primary" onClick={handleSaveMinimumStock} disabled={updateMinimumStockMutation.isPending}>
              {updateMinimumStockMutation.isPending ? 'جاري الحفظ...' : 'حفظ'}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        isOpen={deleteDialogOpen}
        onClose={() => {
          setDeleteDialogOpen(false);
          setProductToDelete(null);
        }}
        onConfirm={handleConfirmDelete}
        title="حذف المنتج"
        message="سيُحذف المنتج نهائيًا مع تنظيف المبيعات والمشتريات والمرتجعات المرتبطة به من البطاقات والتقارير. هل تريد المتابعة؟"
        confirmText="حذف المنتج"
        cancelText="إلغاء"
        variant="danger"
        isLoading={deleteProductMutation.isPending}
      />
      <ConfirmDialog
        isOpen={Boolean(inventoryItemToDelete)}
        onClose={() => setInventoryItemToDelete(null)}
        onConfirm={handleConfirmInventoryItemDelete}
        title="حذف عنصر المخزون"
        message="سيُحذف العنصر إذا لم يرتبط بتاريخ، أو سيُؤرشف مع حفظ الحركة المحاسبية إذا كان مرتبطًا."
        confirmText="تأكيد الحذف"
        isLoading={deleteInventoryItemMutation.isPending}
      />
    </div>
  );
}
