import { useState, useEffect } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
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

// Types
import { ViewMode, ItemInputMethodType, Product } from '../types/inventory.types';

export function InventoryPage() {
  const { t } = useTranslation();
  const [isMobile, setIsMobile] = useState(false);
  const [viewMode, setViewMode] = useState<ViewMode>('products');
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [inputMethod, setInputMethod] = useState<ItemInputMethodType>('barcode');
  const [isCameraScannerOpen, setIsCameraScannerOpen] = useState(false);
  const [barcodeInput, setBarcodeInput] = useState('');

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

  const handleClearSearch = () => {
    setSearchQuery('');
  };

  const handleExport = () => {
    const dataToExport = filteredProducts.map((product: Product) => ({
      'الاسم': product.name,
      'SKU': product.sku,
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
      'السعر': product.sellingPrice,
      'المخزون': product.stock,
      'الحالة': product.condition
    }));
    printTable(dataToPrint, ['الاسم', 'SKU', 'السعر', 'المخزون', 'الحالة'], 'تقرير المخزون');
  };

  const handleViewProduct = (product: Product) => {
    setSelectedProduct(product);
    setIsViewModalOpen(true);
  };

  const handleEditProduct = (product: Product) => {
    setSelectedProduct(product);
    setIsEditModalOpen(true);
  };

  const handleDeleteProduct = (productId: string) => {
    if (window.confirm('هل أنت متأكد من حذف هذا المنتج؟')) {
      deleteProductMutation.mutate(productId);
    }
  };

  const handleSaveProduct = (productData: Product) => {
    if (selectedProduct && selectedProduct.id) {
      // Update existing product
      updateProductMutation.mutate({ id: selectedProduct.id, data: productData });
      setIsEditModalOpen(false);
      setSelectedProduct(null);
    } else {
      // Add new product
      createProductMutation.mutate(productData);
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
      <div className="flex gap-2">
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
      />

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