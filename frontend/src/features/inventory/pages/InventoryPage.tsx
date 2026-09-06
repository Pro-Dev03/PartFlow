import { useState, useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { useNavigate, useLocation } from 'react-router-dom';
import { cn } from '../../../utils';
import { PageHeader } from '../../../components/ui/page-header';
import { Button } from '../../../components/ui/button';
import { Modal } from '../../../components/ui/modal';
import { Input } from '../../../components/ui/input';
import { getButtonSize } from '../../../config/button-sizes';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { Plus, Download, Printer, Package, PackageOpen } from 'lucide-react';

// Custom hooks
import { useInventory } from '../hooks/useInventory';
import { useIsMobile } from '../../../hooks/useIsMobile';

// Components
import { InventoryStats } from '../components/InventoryStats';
import { InventoryFilters } from '../components/InventoryFilters';
import { InventoryScanner } from '../components/InventoryScanner';
import { InventoryList } from '../components/InventoryList';
import { InventoryModals } from '../components/InventoryModals';
import { InventoryLedger } from '../../../components/ui/inventory-ledger';
import type { InventoryMovement } from '../../../components/ui/inventory-ledger';
import { ConfirmDialog } from '../../../components/ui/confirm-dialog';

// Types
import { ViewMode, ItemInputMethodType, Product } from '../types/inventory.types';
import { inventoryApi } from '../../../services/api/endpoints';
import { toast } from 'sonner';
import { getLocalProductImage, setLocalProductImage } from '../../../services/localProductImages';

interface InventoryMovementResponse {
  id: string;
  movement_type: InventoryMovement['type'];
  quantity: number;
  before_quantity: number;
  after_quantity: number;
  reference_type?: string;
  reference_id?: string;
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
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isCreatingProduct, setIsCreatingProduct] = useState(false);
  const [inputMethod, setInputMethod] = useState<ItemInputMethodType>('barcode');
  const [isCameraScannerOpen, setIsCameraScannerOpen] = useState(false);
  const [barcodeInput, setBarcodeInput] = useState('');
  const [showInventoryLedger, setShowInventoryLedger] = useState(false);
  const [inventoryMovements, setInventoryMovements] = useState<InventoryMovement[]>([]);
  const [inventoryLedgerLoading, setInventoryLedgerLoading] = useState(false);
  const [inventoryLedgerError, setInventoryLedgerError] = useState<string | null>(null);
  const [inventoryLedgerProduct, setInventoryLedgerProduct] = useState<Product | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [productToDelete, setProductToDelete] = useState<string | null>(null);
  const [minimumStockProduct, setMinimumStockProduct] = useState<Product | null>(null);
  const [minimumStockValue, setMinimumStockValue] = useState('0');

  // Custom hook
  const {
    inventoryItems,
    filteredProducts,
    filteredInventoryItems,
    productsLoading,
    inventoryLoading,
    searchQuery,
    setSearchQuery,
    sortConfig,
    setSortConfig,
    filters,
    setFilters,
    deleteProductMutation,
    createProductMutation,
    updateProductMutation,
    updateMinimumStockMutation,
    lookupProduct,
  } = useInventory();

  // Handle edit product from navigation state
  useEffect(() => {
    if (location.state?.editProduct) {
      setSelectedProduct(location.state.editProduct);
      setIsEditModalOpen(true);
      // Clear the state to prevent reopening on refresh
      navigate(location.pathname, { replace: true, state: null });
    }
  }, [location.state, navigate, location.pathname]);

  useEffect(() => {
    const search = new URLSearchParams(location.search).get('search') || '';
    setSearchQuery(search);
  }, [location.search, setSearchQuery]);

  const handleClearSearch = () => {
    setSearchQuery('');
  };

  const handleExport = () => {
    const dataToExport = filteredProducts.map((product: Product) => ({
      'الاسم': product.name,
      'SKU': product.sku,
      'التصنيف': product.category_name || product.category || '-',
      'السعر': product.sellingPrice,
      'المخزون': product.stock,
      'الحالة': product.condition
    }));
    exportToCSV(dataToExport, `inventory-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    const dataToPrint = filteredProducts.map((product: Product) => ({
      'الاسم': product.name,
      'SKU': product.sku,
      'التصنيف': product.category_name || product.category || '-',
      'السعر': product.sellingPrice,
      'المخزون': product.stock,
      'الحالة': product.condition
    }));
    printTable(dataToPrint, ['الاسم', 'SKU', 'السعر', 'المخزون', 'الحالة'], 'تقرير المخزون');
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
      image_url: product.image_url || getLocalProductImage(product.id),
      min_stock_level: Number((product as Record<string, unknown>).min_stock_level ?? 0),
    };
    setSelectedProduct(mappedProduct);
    setIsViewModalOpen(true);
  };

  const handleAddPurchase = (product: Product) => {
    navigate('/app/purchases/create', {
      state: {
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
      image_url: product.image_url || getLocalProductImage(product.id),
      min_stock_level: Number((product as Record<string, unknown>).min_stock_level ?? 0),
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

  const handleSaveProduct = (productData: Product) => {
    if (selectedProduct && selectedProduct.id) {
      // Update existing product - map to API field names
      const apiData = {
        name: productData.name,
        sku: productData.sku || `SKU-${Date.now()}`,
        selling_price: productData.sellingPrice,
        cost_price: productData.costPrice,
        stock: productData.stock,
        condition: productData.condition,
        category_id: productData.category_id,
        barcode: productData.barcode,
      };
      setLocalProductImage(selectedProduct.id, productData.image_url || null);
      updateProductMutation.mutate({ id: selectedProduct.id, data: apiData });
      setIsEditModalOpen(false);
      setSelectedProduct(null);
      setIsCreatingProduct(false);
    } else {
      // Add new product - map to API field names
      const apiData = {
        name: productData.name,
        sku: productData.sku || `SKU-${Date.now()}`,
        selling_price: productData.sellingPrice,
        cost_price: productData.costPrice,
        stock: productData.stock,
        condition: productData.condition,
        category_id: productData.category_id,
        barcode: productData.barcode,
      };
      createProductMutation.mutate(apiData, {
        onSuccess: async (response: any) => {
          const product = response?.data?.product ?? response?.data;
          if (product?.id && productData.image_url) {
            setLocalProductImage(product.id, productData.image_url);
          }
          const quantity = Math.max(0, Math.floor(Number(productData.stock) || 0));
          if (!product?.id || quantity === 0) {
            return;
          }

          const inventoryItem = {
            product_id: product.id,
            condition: productData.condition === 'used' ? 'USED' : 'NEW',
            purchase_cost: Number(productData.costPrice) || 0,
            selling_price: Number(productData.sellingPrice) || 0,
            status: 'AVAILABLE' as const,
            notes: 'إضافة منتج بدون فاتورة',
          };
          try {
            await Promise.all(Array.from({ length: quantity }, () => inventoryApi.create(inventoryItem)));
            await queryClient.invalidateQueries({ queryKey: ['inventory'] });
            await queryClient.invalidateQueries({ queryKey: ['products'] });
            toast.success('تمت إضافة المنتج والكمية بدون فاتورة');
          } catch (error) {
            console.error('Failed to create no-invoice inventory quantity:', error);
            toast.error('تم إنشاء المنتج لكن تعذرت إضافة الكمية للمخزون');
          }
        },
      });
      setIsEditModalOpen(false);
      setSelectedProduct(null);
      setIsCreatingProduct(false);
    }
  };

  const handleBarcodeScan = async (e: React.FormEvent) => {
    e.preventDefault();
    if (barcodeInput.trim()) {
      const product = await lookupProduct(barcodeInput.trim());
      
      if (product) {
        setSelectedProduct(product);
        setIsCreatingProduct(false);
        setIsEditModalOpen(true);
        setBarcodeInput('');
      } else {
        // Product not found, open modal for new product
        setSelectedProduct({
          id: '',
          name: '',
          sku: '',
          sellingPrice: 0,
          costPrice: 0,
          stock: 0,
          condition: 'new',
        });
        setIsCreatingProduct(true);
        setIsEditModalOpen(true);
        setBarcodeInput('');
      }
    }
  };

  const handleCameraScan = (barcode: string) => {
    setBarcodeInput(barcode);
    // Create a proper event object for the barcode scan
    const event = new Event('submit', { bubbles: true, cancelable: true }) as unknown as React.FormEvent<HTMLFormElement>;
    handleBarcodeScan(event);
  };

  const handleManualAdd = () => {
    setSelectedProduct({
      id: '',
      name: '',
      sku: '',
      sellingPrice: 0,
      costPrice: 0,
      stock: 0,
      condition: 'new',
    });
    setIsCreatingProduct(true);
    setIsEditModalOpen(true);
  };

  const handleRecommendationClick = (action: string) => {
    if (action === 'search_intel') {
      setSearchQuery('Intel');
    }
  };

  const handleRefresh = () => {
    setSortConfig({ key: '', direction: null });
    setFilters([]);
  };

  const handleViewInventoryLedger = async (productId: string) => {
    const product = filteredProducts.find((item) => item.id === productId) ?? null;
    const inventoryItem = (inventoryItems as any[]).find((item: any) =>
      item.product_id === productId ||
      item.product?.id === productId ||
      item.productId === productId ||
      item.product_name === product?.name
    );
    const itemId = inventoryItem?.id || productId;

    setInventoryLedgerLoading(true);
    setInventoryLedgerError(null);
    setShowInventoryLedger(true);
    setInventoryLedgerProduct(product);

    try {
      const response = await inventoryApi.movements<{ movements?: InventoryMovementResponse[] }>(itemId);
      const movements = response.data?.movements ?? [];
      setInventoryMovements(movements.map((movement) => ({
        id: movement.id,
        type: movement.movement_type,
        quantity: movement.quantity,
        beforeQuantity: movement.before_quantity,
        afterQuantity: movement.after_quantity,
        referenceType: movement.reference_type,
        referenceId: movement.reference_id,
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
        eyebrow="Inventory Intelligence"
        title={t('inventory.title')}
        description="إدارة المخزون والقطع مع تحليلات فورية"
        actions={
          <div className={cn(
            "flex gap-2",
            isMobile ? "flex-col w-full" : ""
          )}>
            <Button 
              variant="secondary" 
              size={getButtonSize('inventory', 'headerActions')}
              className={cn(isMobile ? "w-full" : "")}
              onClick={handleManualAdd}
            >
              <Plus className="w-3.5 h-3.5 me-1.5" />
              {t('inventory.addItem')}
            </Button>
            <Button 
              variant="secondary" 
              size={getButtonSize('inventory', 'headerActions')} 
              onClick={handleExport}
              className={cn(isMobile ? "w-full" : "")}
            >
              <Download className="w-3.5 h-3.5 me-1.5" />
              تصدير
            </Button>
            <Button 
              variant="secondary" 
              size={getButtonSize('inventory', 'headerActions')} 
              onClick={handlePrint}
              className={cn(isMobile ? "w-full" : "")}
            >
              <Printer className="w-3.5 h-3.5 me-1.5" />
              طباعة
            </Button>
          </div>
        }
      />

      {/* Inventory Stats */}
      <InventoryStats 
        products={filteredProducts}
        inventoryItems={inventoryItems}
        onRecommendationClick={handleRecommendationClick}
        isMobile={isMobile}
      />

      {/* Inventory Scanner */}
      <InventoryScanner
        inputMethod={inputMethod}
        setInputMethod={setInputMethod}
        barcodeInput={barcodeInput}
        setBarcodeInput={setBarcodeInput}
        onBarcodeScan={handleBarcodeScan}
        onCameraScan={handleCameraScan}
        onManualAdd={handleManualAdd}
        isCameraScannerOpen={isCameraScannerOpen}
        onCameraOpen={() => setIsCameraScannerOpen(true)}
        onCameraClose={() => setIsCameraScannerOpen(false)}
      />

      {/* Inventory Filters */}
      <InventoryFilters
        searchQuery={searchQuery}
        setSearchQuery={setSearchQuery}
        onClearSearch={handleClearSearch}
        filters={filters}
        setFilters={setFilters}
        sortConfig={sortConfig}
        setSortConfig={setSortConfig}
        onRefresh={handleRefresh}
        isMobile={isMobile}
      />

      {/* View Toggle */}
      <div className="flex gap-2" style={{ marginBottom: '16px' }}>
        <Button
          variant={viewMode === 'products' ? 'primary' : 'secondary'}
          onClick={() => setViewMode('products')}
        >
          <Package className="w-3 h-3 me-1.5" />
          {t('products.title')}
        </Button>
        <Button
          variant={viewMode === 'items' ? 'primary' : 'secondary'}
          onClick={() => setViewMode('items')}
        >
          <PackageOpen className="w-3 h-3 me-1.5" />
          {t('inventory.items')}
        </Button>
        <Button
          variant={showInventoryLedger ? 'primary' : 'secondary'}
          onClick={() => {
            if (showInventoryLedger) {
              handleCloseInventoryLedger();
            } else {
              // Show ledger for first product as example
              if (filteredProducts.length > 0) {
                handleViewInventoryLedger(filteredProducts[0].id);
              }
            }
          }}
        >
          📊
          سجل الحركات
        </Button>
      </div>

      {/* Inventory List */}
      <InventoryList
        viewMode={viewMode}
        filteredProducts={filteredProducts}
        filteredInventoryItems={filteredInventoryItems}
        productsLoading={productsLoading}
        inventoryLoading={inventoryLoading}
        searchQuery={searchQuery}
        onViewProduct={handleViewProduct}
              onAddPurchase={handleAddPurchase}
        onEditProduct={handleEditProduct}
        onEditMinimumStock={handleEditMinimumStock}
        onDeleteProduct={handleDeleteProduct}
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
      />

      {/* Inventory Ledger - Conditionally rendered */}
      {showInventoryLedger && (
        <div style={{ marginTop: '24px' }}>
          <InventoryLedger
            movements={inventoryMovements}
            title="سجل حركات المخزون"
            currentStock={inventoryLedgerProduct?.stock}
            isLoading={inventoryLedgerLoading}
            error={inventoryLedgerError}
          />
        </div>
      )}

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
        message="هل أنت متأكد من حذف هذا المنتج؟ هذا الإجراء لا يمكن التراجع عنه."
        confirmText="حذف المنتج"
        cancelText="إلغاء"
        variant="danger"
        isLoading={deleteProductMutation.isPending}
      />
    </div>
  );
}