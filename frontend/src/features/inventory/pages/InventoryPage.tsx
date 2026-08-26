import { useState, useEffect } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { useNavigate, useLocation } from 'react-router-dom';
import { cn } from '../../../utils';
import { PageHeader } from '../../../components/ui/page-header';
import { Button } from '../../../components/ui/button';
import { getButtonSize } from '../../../config/button-sizes';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { handleSort, type SortConfig } from '../../../lib/table-utils';
import { Plus, Download, Printer, Package, PackageOpen } from 'lucide-react';

// Custom hooks
import { useInventory } from '../hooks/useInventory';

// Components
import { InventoryStats } from '../components/InventoryStats';
import { InventoryFilters } from '../components/InventoryFilters';
import { InventoryScanner } from '../components/InventoryScanner';
import { InventoryList } from '../components/InventoryList';
import { InventoryModals } from '../components/InventoryModals';
import { InventoryLedger } from '../../../components/ui/inventory-ledger';

// Types
import { ViewMode, ItemInputMethodType, Product } from '../types/inventory.types';

export function InventoryPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const [isMobile, setIsMobile] = useState(false);
  const [viewMode, setViewMode] = useState<ViewMode>('products');
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [inputMethod, setInputMethod] = useState<ItemInputMethodType>('barcode');
  const [isCameraScannerOpen, setIsCameraScannerOpen] = useState(false);
  const [barcodeInput, setBarcodeInput] = useState('');
  const [showInventoryLedger, setShowInventoryLedger] = useState(false);
  const [inventoryMovements, setInventoryMovements] = useState<any[]>([]);

  // Custom hook
  const {
    products,
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
    archiveProductMutation,
    createProductMutation,
    updateProductMutation,
    lookupProduct,
  } = useInventory();

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768);
    };
    
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  // Handle edit product from navigation state
  useEffect(() => {
    if (location.state?.editProduct) {
      setSelectedProduct(location.state.editProduct);
      setIsEditModalOpen(true);
      // Clear the state to prevent reopening on refresh
      navigate(location.pathname, { replace: true, state: null });
    }
  }, [location.state, navigate, location.pathname]);

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
      sellingPrice: product.sellingPrice,
      costPrice: product.costPrice || (product as any).cost_price || 0,
      stock: product.stock,
      condition: product.condition,
      category: product.category,
      category_id: product.category_id,
      price: product.price,
      barcode: product.barcode,
    };
    setSelectedProduct(mappedProduct);
    setIsViewModalOpen(true);
  };

  const handleEditProduct = (product: Product) => {
    // Map API response to local Product type with proper field names
    const mappedProduct: Product = {
      id: product.id,
      name: product.name,
      sku: product.sku,
      sellingPrice: product.sellingPrice,
      costPrice: product.costPrice || (product as any).cost_price || 0,
      stock: product.stock,
      condition: product.condition,
      category: product.category,
      category_id: product.category_id,
      price: product.price,
      barcode: product.barcode,
    };
    setSelectedProduct(mappedProduct);
    setIsEditModalOpen(true);
  };

  const handleDeleteProduct = (productId: string) => {
    if (window.confirm('هل أنت متأكد من حذف هذا المنتج؟')) {
      deleteProductMutation.mutate(productId);
    }
  };

  const handleSaveProduct = (productData: Product) => {
    if (selectedProduct && selectedProduct.id) {
      // Update existing product - map to API field names
      const apiData = {
        name: productData.name,
        sku: productData.sku,
        selling_price: productData.sellingPrice,
        cost_price: productData.costPrice,
        stock: productData.stock,
        condition: productData.condition,
        category_id: productData.category_id,
        barcode: productData.barcode,
      };
      updateProductMutation.mutate({ id: selectedProduct.id, data: apiData });
      setIsEditModalOpen(false);
      setSelectedProduct(null);
    } else {
      // Add new product - map to API field names
      const apiData = {
        name: productData.name,
        sku: productData.sku,
        selling_price: productData.sellingPrice,
        cost_price: productData.costPrice,
        stock: productData.stock,
        condition: productData.condition,
        category_id: productData.category_id,
        barcode: productData.barcode,
      };
      createProductMutation.mutate(apiData);
      setIsEditModalOpen(false);
      setSelectedProduct(null);
    }
  };

  const handleBarcodeScan = async (e: React.FormEvent) => {
    e.preventDefault();
    if (barcodeInput.trim()) {
      const product = await lookupProduct(barcodeInput.trim());
      
      if (product) {
        setSelectedProduct(product);
        setIsEditModalOpen(true);
        setBarcodeInput('');
      } else {
        // Product not found, open modal for new product
        setSelectedProduct(null);
        setIsEditModalOpen(true);
        setBarcodeInput('');
      }
    }
  };

  const handleCameraScan = (barcode: string) => {
    setBarcodeInput(barcode);
    handleBarcodeScan(new Event('submit') as any);
  };

  const handleManualAdd = () => {
    setSelectedProduct(null);
    setIsEditModalOpen(true);
  };

  const handleRecommendationClick = (action: string) => {
    if (action === 'search_intel') {
      setSearchQuery('Intel');
    } else if (action === 'filter_used') {
      if (filters.some(f => f.key === 'condition' && f.value === 'USED')) {
        setFilters(filters.filter(f => !(f.key === 'condition' && f.value === 'USED')));
      } else {
        setFilters([...filters, { key: 'condition', value: 'USED' }]);
      }
    }
  };

  const handleRefresh = () => {
    setSortConfig({ key: '', direction: null });
    setFilters([]);
  };

  const handleViewInventoryLedger = async (productId: string) => {
    try {
      // In a real implementation, this would call the API to get inventory movements
      // For now, we'll show a sample
      const sampleMovements = [
        {
          id: '1',
          date: new Date().toISOString(),
          type: 'PURCHASE',
          quantity: 10,
          beforeQuantity: 0,
          afterQuantity: 10,
          referenceType: 'purchase',
          referenceId: 'PO-001',
          reason: 'شراء جديد من المورد',
          createdBy: 'user'
        },
        {
          id: '2',
          date: new Date(Date.now() - 86400000).toISOString(),
          type: 'SALE',
          quantity: 3,
          beforeQuantity: 10,
          afterQuantity: 7,
          referenceType: 'sale',
          referenceId: 'SALE-001',
          reason: 'بيع للعميل',
          createdBy: 'user'
        }
      ];
      setInventoryMovements(sampleMovements);
      setShowInventoryLedger(true);
    } catch (error) {
      console.error('Error loading inventory movements:', error);
    }
  };

  const handleCloseInventoryLedger = () => {
    setShowInventoryLedger(false);
    setInventoryMovements([]);
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
        onEditProduct={handleEditProduct}
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
            currentStock={inventoryMovements.length > 0 ? inventoryMovements[0].afterQuantity : 0}
          />
        </div>
      )}

      {/* Inventory Modals */}
      <InventoryModals
        isViewModalOpen={isViewModalOpen}
        setIsViewModalOpen={setIsViewModalOpen}
        isEditModalOpen={isEditModalOpen}
        setIsEditModalOpen={setIsEditModalOpen}
        selectedProduct={selectedProduct}
        setSelectedProduct={setSelectedProduct}
        onSaveProduct={handleSaveProduct}
      />
    </div>
  );
}