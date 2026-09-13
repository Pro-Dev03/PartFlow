import { useEffect, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { inventoryApi, partTypesApi, customersApi, productsApi, barcodeApi, acquisitionsApi } from '../../../services/api/endpoints';
import { Card, CardContent } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { SearchInput } from '../../../components/ui/search-input';
import { Select } from '../../../components/ui/select';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { Modal } from '../../../components/ui/modal';
import { ConfirmDialog } from '../../../components/ui/confirm-dialog';
import { PaginationControls } from '../../../components/ui/pagination-controls';
import { toast } from 'sonner';
import { playScanSound } from '../../../hooks/useBarcodeContext';
import { getPartTypeImage } from '../../../services/localPartTypeImages';
import {
  Plus,
  Layers,
  Package,
  Cpu,
  Keyboard,
  Barcode,
  Type,
  Camera,
  ShoppingCart,
  Trash2,
  Pencil,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Clock
} from 'lucide-react';

export function UsedPartsPage() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedPartType, setSelectedPartType] = useState<string>('');
  const [page, setPage] = useState(1);
  const pageSize = 10;
  
  // Acquisition modal state
  const [isAcquisitionModalOpen, setIsAcquisitionModalOpen] = useState(false);
  const [isCustomerManual, setIsCustomerManual] = useState(false);
  const [acquisitionCustomer, setAcquisitionCustomer] = useState('');
  const [acquisitionCustomerManual, setAcquisitionCustomerManual] = useState('');
  const [acquisitionProductName, setAcquisitionProductName] = useState('');
  const [acquisitionPartType, setAcquisitionPartType] = useState('');
  const [acquisitionCondition, setAcquisitionCondition] = useState('used');
  const [acquisitionGrade, setAcquisitionGrade] = useState('good');
  const [acquisitionPrice, setAcquisitionPrice] = useState('');
  const [acquisitionSellingPrice, setAcquisitionSellingPrice] = useState('');
  const [acquisitionSerialNumber, setAcquisitionSerialNumber] = useState('');
  const [acquisitionNotes, setAcquisitionNotes] = useState('');
  const [paymentStatus, setPaymentStatus] = useState('payable');
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [partToDelete, setPartToDelete] = useState<string | null>(null);
  const [partToEdit, setPartToEdit] = useState<any | null>(null);
  const [editPurchaseCost, setEditPurchaseCost] = useState('');
  const [editSellingPrice, setEditSellingPrice] = useState('');
  const [editProductName, setEditProductName] = useState('');
  const [editPartType, setEditPartType] = useState('');
  const [editCondition, setEditCondition] = useState('used');
  const [editGrade, setEditGrade] = useState('good');
  const [editSerialNumber, setEditSerialNumber] = useState('');
  const [editNotes, setEditNotes] = useState('');

  // Barcode scanner state
  const [barcodeInput, setBarcodeInput] = useState('');
  const [inputMethod, setInputMethod] = useState<'barcode' | 'manual' | 'camera'>('barcode');
  const [soundEnabled] = useState(true);

  const handleClearSearch = () => {
    setSearchQuery('');
  };

  const { data: inventoryData, isLoading } = useQuery({
    queryKey: ['inventory', 'used-parts', page, pageSize],
    queryFn: () => inventoryApi.list({ page, per_page: pageSize, condition: 'USED', status: 'AVAILABLE' }),
    staleTime: 0,
    refetchOnMount: 'always',
  });

  useEffect(() => {
    setPage(1);
  }, [searchQuery, selectedPartType]);

  const { data: partTypesData, error: partTypesError, isLoading: partTypesLoading } = useQuery({
    queryKey: ['part-types'],
    queryFn: () => partTypesApi.list(),
    retry: 1,
  });

  const { data: customersData, isLoading: customersLoading } = useQuery({
    queryKey: ['customers'],
    queryFn: () => customersApi.list({ page: 1, per_page: 100 }),
  });

  const { data: acquisitionsData } = useQuery({
    queryKey: ['acquisitions', 'used-parts-sellers'],
    queryFn: () => acquisitionsApi.list({ type: 'CUSTOMER', page: 1, per_page: 100 }),
  });

  const { data: productsData } = useQuery({
    queryKey: ['products', 'used-parts-edit'],
    queryFn: () => productsApi.list({ page: 1, per_page: 100 }),
  });

  const inventoryItems = Array.isArray(inventoryData?.data)
    ? inventoryData.data
    : Array.isArray(inventoryData?.data?.items)
      ? inventoryData.data.items
      : [];
  const partTypes = Array.isArray(partTypesData?.data) ? partTypesData.data : [];
  const customers = Array.isArray(customersData?.data) ? customersData.data : [];
  const products = Array.isArray(productsData?.data?.products) ? productsData.data.products : [];
  const productById = new Map(products.map((product: any) => [product.id, product]));
  const customerNames = new Map(customers.map((customer: any) => [customer.id, customer.name]));
  const sellerByInventoryItemId = new Map<string, string>();
  const acquisitions = Array.isArray(acquisitionsData?.data) ? acquisitionsData.data : [];
  acquisitions.forEach((acquisition: any) => {
    const sellerName = customerNames.get(acquisition.customer_id) || 'بائع غير معروف';
    (acquisition.items || []).forEach((acquisitionItem: any) => {
      if (acquisitionItem.inventory_item_id) {
        sellerByInventoryItemId.set(acquisitionItem.inventory_item_id, sellerName);
      }
    });
  });
  // Handle part types error gracefully
  if (partTypesError) {
    console.warn('Failed to load part types:', partTypesError);
  }

  // Acquisition mutation
  const createAcquisitionMutation = useMutation({
    mutationFn: (data: any) => acquisitionsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['acquisitions'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      toast.success('تم شراء القطعة المستعملة بنجاح!');
      handleCancelAcquisition();
    },
    onError: (error) => {
      console.error('Acquisition failed:', error);
      toast.error('فشل شراء القطعة المستعملة');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => inventoryApi.delete(id, { permanent: true }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      setIsDeleteDialogOpen(false);
      setPartToDelete(null);
      toast.success('تم حذف القطعة من المخزون');
    },
    onError: (error) => {
      console.error('Failed to delete used part:', error);
      toast.error('فشل حذف القطعة');
    },
  });

  const updateMutation = useMutation({
    mutationFn: async (data: any) => {
      await inventoryApi.update(data.id, data);
      if (data.product_update) {
            await productsApi.update(data.product_update.id, {
              ...data.product_update,
              name: data.product_name.trim(),
              sku: data.product_update.sku || `USED-${data.product_update.id}`,
              cost_price: Number(data.product_update.cost_price ?? data.purchase_cost),
              selling_price: Number(data.selling_price),
            });
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['reports'] });
      setPartToEdit(null);
      toast.success('تم تحديث القطعة بنجاح');
    },
    onError: () => toast.error('فشل تحديث القطعة'),
  });

  const handleDeletePart = (id: string) => {
    setPartToDelete(id);
    setIsDeleteDialogOpen(true);
  };

  const handleConfirmDelete = () => {
    if (partToDelete) {
      deleteMutation.mutate(partToDelete);
    }
  };

  const handleEditPart = (item: any) => {
    setPartToEdit(item);
    setEditProductName(item.product_name || item.product?.name || '');
    setEditPartType(item.part_type_id || '');
    setEditCondition(String(item.condition || 'used').toLowerCase());
    setEditGrade(String(item.grade || 'good').toLowerCase());
    setEditSerialNumber(item.serial_number || '');
    setEditPurchaseCost(String(item.purchase_cost ?? ''));
    setEditSellingPrice(String(item.selling_price ?? ''));
    setEditNotes(item.notes || '');
  };

  const handleSaveEdit = () => {
    if (!partToEdit) return;
    const purchaseCost = Number(editPurchaseCost);
    const sellingPrice = Number(editSellingPrice);
    if (!Number.isFinite(purchaseCost) || purchaseCost < 0 || !Number.isFinite(sellingPrice) || sellingPrice < 0) {
      toast.error('أدخل أسعارًا صحيحة');
      return;
    }
    updateMutation.mutate({
      id: partToEdit.id,
      product_id: partToEdit.product_id,
      product_name: editProductName,
      product_update: productById.get(partToEdit.product_id),
      part_type_id: editPartType || null,
      condition: editCondition,
      grade: editGrade,
      serial_number: editSerialNumber,
      purchase_cost: purchaseCost,
      selling_price: sellingPrice,
      notes: editNotes,
    });
  };

  // Barcode scan handler
  const handleBarcodeScan = async (e: React.FormEvent) => {
    e.preventDefault();
    if (barcodeInput.trim()) {
      try {
        const response = await barcodeApi.lookupProduct(barcodeInput.trim());
        const product = response as any;
        
        if (product && product.id) {
          // Open acquisition modal with product pre-filled
          setAcquisitionProductName(product.name || '');
          setIsAcquisitionModalOpen(true);
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
      }
      setBarcodeInput('');
    }
  };

  // Acquisition handlers
  const handleSubmitAcquisition = async () => {
    const customerValue = isCustomerManual ? acquisitionCustomerManual : acquisitionCustomer;

    if (!customerValue || !acquisitionProductName.trim() || !acquisitionPartType || !acquisitionPrice || !acquisitionSellingPrice) {
      toast.error('يرجى ملء جميع الحقول المطلوبة');
      return;
    }

    const productResponse = await productsApi.create({
      name: acquisitionProductName.trim(),
      sku: `USED-${Date.now()}`,
      cost_price: parseFloat(acquisitionPrice),
      selling_price: parseFloat(acquisitionSellingPrice),
      condition: 'used',
    });
    const createdProduct = (productResponse.data as any)?.product ?? productResponse.data;
    if (!createdProduct?.id) {
      toast.error('تعذر إنشاء القطعة بالاسم المدخل');
      return;
    }

    const data = {
      type: 'CUSTOMER',
      acquisition_date: new Date().toISOString(),
      customer_id: isCustomerManual ? undefined : acquisitionCustomer,
      customer_name: isCustomerManual ? acquisitionCustomerManual : undefined,
      items: [{
        product_id: createdProduct.id,
        part_type_id: acquisitionPartType,
        serial_number: acquisitionSerialNumber,
        condition: acquisitionCondition,
        grade: acquisitionGrade,
        unit_cost: parseFloat(acquisitionPrice),
        selling_price: parseFloat(acquisitionSellingPrice),
        notes: acquisitionNotes,
      }],
      payment_status: paymentStatus,
      notes: acquisitionNotes,
    };

    createAcquisitionMutation.mutate(data);
  };

  const handleCancelAcquisition = () => {
    setAcquisitionCustomer('');
    setAcquisitionCustomerManual('');
    setIsCustomerManual(false);
    setAcquisitionProductName('');
    setAcquisitionPartType('');
    setAcquisitionCondition('used');
    setAcquisitionGrade('good');
    setAcquisitionPrice('');
    setAcquisitionSellingPrice('');
    setAcquisitionSerialNumber('');
    setAcquisitionNotes('');
    setPaymentStatus('payable');
    setIsAcquisitionModalOpen(false);
  };

  const handleSellItem = (item: any) => {
    navigate('/app/sales', {
      state: {
        usedPart: {
          id: item.product_id || item.product?.id,
          inventoryItemId: item.id,
          name: item.product_name || item.product?.name || 'قطعة مستعملة',
          barcode: item.serial_number || item.barcode || item.id,
          price: item.selling_price,
          stock: 1,
          condition: item.condition,
          purchaseCost: item.purchase_cost,
          isTradeIn: true,
        },
      },
    });
  };

  // Filter used parts only
  const usedParts = inventoryItems.filter((item: any) => {
    const condition = String(item.condition || '').trim().toUpperCase();
    const status = String(item.status || '').trim().toUpperCase();
    return condition === 'USED' &&
      status === 'AVAILABLE';
  });
  // Financial totals cover only sellable stock; rejected/damaged parts never contribute.
  const sellableUsedParts = usedParts.filter(
    (item: any) => String(item.status || '').trim().toUpperCase() !== 'DAMAGED'
  );

  // Filter by search and part type
  const filteredParts = usedParts.filter((item: any) => {
    const matchesSearch = !searchQuery || 
      (item.product_name && item.product_name.toLowerCase().includes(searchQuery.toLowerCase())) ||
      (item.notes && item.notes.toLowerCase().includes(searchQuery.toLowerCase()));
    
    const matchesType = !selectedPartType || item.part_type_id === selectedPartType;
    
    return matchesSearch && matchesType;
  });
  const totalUsedParts = Number(inventoryData?.meta?.total || inventoryItems.length);

  const formatCurrency = (value: number) => `₪${new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(value)}`;

  const totalPurchaseValue = sellableUsedParts.reduce((sum: number, item: any) => {
    const cost = Number(item.purchase_cost ?? item.unit_cost ?? item.cost_price ?? item.cost ?? 0);
    return sum + cost;
  }, 0);

  const totalSellingValue = sellableUsedParts.reduce((sum: number, item: any) => {
    const price = Number(item.selling_price ?? item.price ?? item.sale_price ?? 0);
    return sum + price;
  }, 0);

  const estimatedProfit = totalSellingValue - totalPurchaseValue;

  const getIconComponent = (_iconName: string) => {
    return <Layers className="w-4 h-4" />;
  };

  const getPartType = (partTypeId: string) => {
    if (!partTypes || partTypes.length === 0) return null;
    return partTypes.find((pt: any) => pt.id === partTypeId);
  };

  return (
    <div>
      <PageHeader
        eyebrow="Used Parts Inventory"
        title="القطع المستعملة"
        description="إدارة القطع المستعملة والبيع"
        actions={
          <div className="flex flex-col gap-2">
            <Button variant="primary" onClick={() => setIsAcquisitionModalOpen(true)}>
              <Plus className="w-4 h-4" />
              شراء قطعة مستعملة
            </Button>
            <Button
              variant="secondary"
              onClick={() => navigate('/app/usedparts/stock')}
            >
              <Layers className="w-4 h-4" />
              مخزون القطع المستعملة
            </Button>
          </div>
        }
      />

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-cyan/10">
                <Layers className="w-5 h-5 text-cyan" />
              </div>
              <div>
                <p className="text-sm text-gray-400">إجمالي القطع</p>
                <p className="text-2xl font-bold">{usedParts.length}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-green/10">
                <Package className="w-5 h-5 text-green" />
              </div>
              <div>
                <p className="text-sm text-gray-400">قيمة البيع</p>
                <p className="text-2xl font-bold">{formatCurrency(totalSellingValue)}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-orange/10">
                <ShoppingCart className="w-5 h-5 text-orange" />
              </div>
              <div>
                <p className="text-sm text-gray-400">قيمة الشراء</p>
                <p className="text-2xl font-bold">{formatCurrency(totalPurchaseValue)}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-purple/10">
                <Cpu className="w-5 h-5 text-purple" />
              </div>
              <div>
                <p className="text-sm text-gray-400">الربح المتوقع</p>
                <p className="text-2xl font-bold text-green-500">{formatCurrency(estimatedProfit)}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Barcode Scanner Section */}
      <Card className="mb-4 border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
        <CardContent className="p-4">
          <div className="flex justify-between items-center mb-3">
            <h3 className="text-base font-semibold text-[var(--text-primary)]">
              ماسح الباركود
            </h3>
          </div>
          
          <div>
            <div className="w-full">
              <div className="flex items-center gap-2 mb-3">
                <Keyboard className="w-3.5 h-3.5 text-[var(--text-muted)]" />
                <span className="text-xs font-medium text-[var(--text-secondary)]">طريقة الإضافة</span>
              </div>
              
              <div className="grid grid-cols-3 gap-2">
                <button
                  onClick={() => setInputMethod('barcode')}
                  className={`flex flex-col items-center gap-1.5 p-2.5 rounded-lg border transition-all duration-200 ${
                    inputMethod === 'barcode'
                      ? 'border-[var(--color-primary)] bg-[var(--color-primary-08)] shadow-md'
                      : 'border-[var(--border-default)] bg-[var(--bg-surface)] hover:border-[var(--color-primary-20)] hover:shadow-sm'
                  }`}
                >
                  <div className={`p-1.5 rounded-md transition-all duration-200 ${
                    inputMethod === 'barcode' ? 'bg-[var(--color-primary-15)]' : 'bg-[var(--bg-surface-elevated)]'
                  }`}>
                    <Barcode className={`w-4 h-4 transition-all duration-200 ${
                      inputMethod === 'barcode' ? 'text-[var(--color-primary)]' : 'text-[var(--text-muted)]'
                    }`} />
                  </div>
                  <div className="text-center">
                    <span className={`text-xs font-medium block transition-all duration-200 ${
                      inputMethod === 'barcode' ? 'text-[var(--text-primary)]' : 'text-[var(--text-secondary)]'
                    }`}>مسح باركود</span>
                    <span className="text-[10px] block text-[var(--text-muted)]">استخدام ماسح الباركود</span>
                  </div>
                </button>
                
                <button
                  onClick={() => setInputMethod('manual')}
                  className={`flex flex-col items-center gap-1.5 p-2.5 rounded-lg border transition-all duration-200 ${
                    inputMethod === 'manual'
                      ? 'border-[var(--color-primary)] bg-[var(--color-primary-08)] shadow-md'
                      : 'border-[var(--border-default)] bg-[var(--bg-surface)] hover:border-[var(--color-primary-20)] hover:shadow-sm'
                  }`}
                >
                  <div className={`p-1.5 rounded-md transition-all duration-200 ${
                    inputMethod === 'manual' ? 'bg-[var(--color-primary-15)]' : 'bg-[var(--bg-surface-elevated)]'
                  }`}>
                    <Type className={`w-4 h-4 transition-all duration-200 ${
                      inputMethod === 'manual' ? 'text-[var(--color-primary)]' : 'text-[var(--text-muted)]'
                    }`} />
                  </div>
                  <div className="text-center">
                    <span className={`text-xs font-medium block transition-all duration-200 ${
                      inputMethod === 'manual' ? 'text-[var(--text-primary)]' : 'text-[var(--text-secondary)]'
                    }`}>إضافة يدوية</span>
                    <span className="text-[10px] block text-[var(--text-muted)]">إدخال البيانات يدوياً</span>
                  </div>
                </button>
                
                <button
                  onClick={() => setInputMethod('camera')}
                  className={`flex flex-col items-center gap-1.5 p-2.5 rounded-lg border transition-all duration-200 ${
                    inputMethod === 'camera'
                      ? 'border-[var(--color-primary)] bg-[var(--color-primary-08)] shadow-md'
                      : 'border-[var(--border-default)] bg-[var(--bg-surface)] hover:border-[var(--color-primary-20)] hover:shadow-sm'
                  }`}
                >
                  <div className={`p-1.5 rounded-md transition-all duration-200 ${
                    inputMethod === 'camera' ? 'bg-[var(--color-primary-15)]' : 'bg-[var(--bg-surface-elevated)]'
                  }`}>
                    <Camera className={`w-4 h-4 transition-all duration-200 ${
                      inputMethod === 'camera' ? 'text-[var(--color-primary)]' : 'text-[var(--text-muted)]'
                    }`} />
                  </div>
                  <div className="text-center">
                    <span className={`text-xs font-medium block transition-all duration-200 ${
                      inputMethod === 'camera' ? 'text-[var(--text-primary)]' : 'text-[var(--text-secondary)]'
                    }`}>كاميرا</span>
                    <span className="text-[10px] block text-[var(--text-muted)]">مسح عبر الكاميرا</span>
                  </div>
                </button>
              </div>
            </div>
            
            <form onSubmit={handleBarcodeScan} className="mt-3">
              <div className="pf-barcode-row flex gap-2 items-stretch w-full">
                <div className="pf-barcode-input min-w-0 flex-1 relative flex items-center">
                  <div className="w-full">
                    <Input
                      id="barcode-input"
                      value={barcodeInput}
                      onChange={(e) => setBarcodeInput(e.target.value)}
                      placeholder="مسح الباركود أو أدخل الرقم يدوياً"
                      className="h-10"
                    />
                  </div>
                </div>
                <Button type="submit" size="sm" className="shrink-0 whitespace-nowrap px-4">
                  إضافة
                </Button>
              </div>
            </form>
          </div>
        </CardContent>
      </Card>

      {/* Search and Filters */}
      <Card className="mb-4 border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
        <CardContent className="p-4">
          <div className="pf-search-row flex flex-col md:flex-row gap-3 items-stretch">
            <div className="min-w-0 flex-1 w-full max-w-xl">
              <SearchInput
                placeholder="بحث عن قطعة..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onClear={handleClearSearch}
                size="sm"
                className="w-full"
              />
            </div>
            <div className="flex gap-2">
              <Select
                value={selectedPartType}
                onChange={(e) => setSelectedPartType(e.target.value)}
                loading={partTypesLoading}
                options={[
                  { value: '', label: 'جميع الأنواع' },
                  ...partTypes.map((pt: any) => ({ value: pt.id, label: pt.name_ar })),
                ]}
                emptyMessage="لا يوجد أنواع"
                size="sm"
                className="w-48"
              />
              <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                  setSearchQuery('');
                  setSelectedPartType('');
                }}
                className="h-10 px-4"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="mr-1.5">
                  <path d="M3 6h18"></path>
                  <path d="M7 12h10"></path>
                  <path d="M10 18h4"></path>
                </svg>
                <span>مسح</span>
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Dedicated used-parts stock */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : filteredParts.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center">
            <Package className="w-12 h-12 mx-auto mb-4 text-gray-400" />
            <p className="text-gray-400">لا توجد قطع مستعملة متاحة</p>
          </CardContent>
        </Card>
      ) : (
        <div id="used-parts-stock" className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filteredParts.map((item: any) => {
            const partType = getPartType(item.part_type_id);
            return (
              <Card 
                key={item.id}
                className="hover:border-cyan-500 transition-colors cursor-pointer"
                style={{
                  background: partType ? `linear-gradient(135deg, ${partType.color}15 0%, transparent 100%)` : undefined
                }}
              >
                <CardContent className="p-3">
                  <div className="flex items-start justify-between mb-2">
                      {getPartTypeImage(String(item.part_type_id)) ? (
                        <img
                          src={getPartTypeImage(String(item.part_type_id))}
                          alt={item.product_name || 'نوع القطعة'}
                          style={{ width: '96px', height: '72px', objectFit: 'cover', display: 'block', borderRadius: '10px', flex: '0 0 96px' }}
                        />
                      ) : partType && (
                      <div
                        className="p-2 rounded-lg"
                        style={{ background: `${partType.color}30`, color: partType.color }}
                      >
                        {getIconComponent(partType.icon)}
                      </div>
                    )}
                  </div>
                  
                  <h3 className="font-semibold text-sm mb-1">
                    {item.product_name || item.product?.name || 'قطعة بدون اسم'}
                  </h3>
                  
                  {partType && (
                    <p className="text-xs text-gray-400 mb-2">{partType.name_ar}</p>
                  )}
                  
                  <div className="space-y-2 mb-3">
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">اشتريت من:</span>
                      <span className="text-[var(--text-primary)]">
                        {sellerByInventoryItemId.get(item.id) || 'غير محدد'}
                      </span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">اشتريت بـ:</span>
                      <span>₪{item.purchase_cost.toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between text-sm font-semibold">
                      <span className="text-gray-400">سعر البيع:</span>
                       <span className="text-cyan">₪{item.selling_price.toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">الربح:</span>
                      <span className={item.selling_price > item.purchase_cost ? 'text-green' : 'text-red'}>
                         ₪{(item.selling_price - item.purchase_cost).toFixed(2)}
                      </span>
                    </div>
                  </div>
                  
                  {item.notes && (
                    <p className="text-xs text-gray-500 mt-2 line-clamp-2">{item.notes}</p>
                  )}
                  
                  <div className="flex gap-2 mt-4">
                    <Button
                      variant="primary"
                      size="sm"
                      className="flex-1"
                      onClick={() => handleSellItem(item)}
                    >
                      بيع
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleEditPart(item)}
                      title="تعديل القطعة"
                      aria-label="تعديل القطعة"
                    >
                      <Pencil className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDeletePart(item.id)}
                      title="حذف القطعة"
                      aria-label="حذف القطعة"
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
      {filteredParts.length > 0 && (
        <Card className="mt-4">
          <PaginationControls
            page={page}
            pageSize={pageSize}
            total={totalUsedParts}
            onPageChange={setPage}
            isLoading={isLoading}
          />
        </Card>
      )}

      {/* Acquisition Modal */}
      <Modal
        isOpen={isAcquisitionModalOpen}
        onClose={handleCancelAcquisition}
        title="شراء قطعة مستعملة"
        variant="modern"
        size="lg"
        style={{
          borderRadius: '24px',
          overflow: 'hidden',
          background: 'var(--bg-surface)',
          border: '1px solid var(--border-primary)',
          boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(255, 255, 255, 0.05) inset, 0 0 40px rgba(99, 102, 241, 0.1)'
        }}
      >
        <div className="space-y-6">
          {/* Customer Section */}
          <div style={{
            padding: '20px',
            background: 'var(--bg-surface-elevated)',
            borderRadius: '16px',
            border: '1px solid var(--border-subtle)'
          }}>
            <label style={{
              fontSize: '14px',
              fontWeight: '600',
              color: 'var(--text-primary)',
              marginBottom: '12px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              البائع (الزبون)
              <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
            </label>
            <div style={{ display: 'flex', gap: '16px', marginBottom: '16px' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '10px', fontSize: '14px', color: 'var(--text-primary)', cursor: 'pointer', padding: '10px 20px', borderRadius: '10px', background: !isCustomerManual ? 'var(--color-primary-10)' : 'transparent', border: !isCustomerManual ? '1px solid var(--color-primary-20)' : '1px solid var(--border-subtle)', transition: 'all 0.2s ease' }}>
                <input
                  type="radio"
                  checked={!isCustomerManual}
                  onChange={() => setIsCustomerManual(false)}
                  style={{ width: '18px', height: '18px', accentColor: 'var(--primary)' }}
                />
                <span>اختر من القائمة</span>
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: '10px', fontSize: '14px', color: 'var(--text-primary)', cursor: 'pointer', padding: '10px 20px', borderRadius: '10px', background: isCustomerManual ? 'var(--color-primary-10)' : 'transparent', border: isCustomerManual ? '1px solid var(--color-primary-20)' : '1px solid var(--border-subtle)', transition: 'all 0.2s ease' }}>
                <input
                  type="radio"
                  checked={isCustomerManual}
                  onChange={() => setIsCustomerManual(true)}
                  style={{ width: '18px', height: '18px', accentColor: 'var(--primary)' }}
                />
                <span>اكتب يدوياً</span>
              </label>
            </div>
            {!isCustomerManual ? (
              <Select
                value={acquisitionCustomer}
                onChange={(e) => setAcquisitionCustomer(e.target.value)}
                loading={customersLoading}
                options={[
                  { value: '', label: 'اختر البائع...' },
                  ...customers.map((c) => ({ value: c.id, label: c.name })),
                ]}
                emptyMessage="لا يوجد عملاء"
                style={{ borderRadius: '12px', height: '48px' }}
              />
            ) : (
              <Input
                type="text"
                value={acquisitionCustomerManual}
                onChange={(e) => setAcquisitionCustomerManual(e.target.value)}
                placeholder="أدخل اسم البائع..."
                style={{ borderRadius: '12px', height: '48px' }}
              />
            )}
          </div>

          {/* Product Section */}
          <div style={{
            padding: '20px',
            background: 'var(--bg-surface-elevated)',
            borderRadius: '16px',
            border: '1px solid var(--border-subtle)'
          }}>
            <label style={{
              fontSize: '14px',
              fontWeight: '600',
              color: 'var(--text-primary)',
              marginBottom: '12px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              المنتج
              <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
            </label>
            <Input
              type="text"
              value={acquisitionProductName}
              onChange={(e) => setAcquisitionProductName(e.target.value)}
              placeholder="اكتب اسم القطعة..."
              style={{ borderRadius: '12px', height: '48px' }}
            />
            <label style={{
              fontSize: '14px',
              fontWeight: '600',
              color: 'var(--text-primary)',
              marginBottom: '12px',
              marginTop: '16px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              نوع القطعة
              <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
            </label>
            <Select
              value={acquisitionPartType}
              onChange={(e) => setAcquisitionPartType(e.target.value)}
              loading={partTypesLoading}
              options={[
                { value: '', label: 'اختر نوع القطعة...' },
                ...partTypes.map((pt: any) => ({ value: pt.id, label: pt.name_ar })),
              ]}
              emptyMessage="لا توجد أنواع قطع"
              style={{ borderRadius: '12px', height: '48px' }}
            />
          </div>

          {/* Item Details in Grid */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            {/* Condition */}
            <div style={{
              padding: '20px',
              background: 'var(--bg-surface-elevated)',
              borderRadius: '16px',
              border: '1px solid var(--border-subtle)'
            }}>
              <label style={{
                fontSize: '14px',
                fontWeight: '600',
                color: 'var(--text-primary)',
                marginBottom: '12px',
                display: 'block',
                letterSpacing: '0.2px'
              }}>
                الحالة
              </label>
              <Select
                value={acquisitionCondition}
                onChange={(e) => setAcquisitionCondition(e.target.value)}
                options={[
                  { value: 'used', label: 'مستعمل' },
                ]}
                style={{ borderRadius: '12px', height: '48px' }}
              />
            </div>

            {/* Grade */}
            <div style={{
              padding: '20px',
              background: 'var(--bg-surface-elevated)',
              borderRadius: '16px',
              border: '1px solid var(--border-subtle)'
            }}>
              <label style={{
                fontSize: '14px',
                fontWeight: '600',
                color: 'var(--text-primary)',
                marginBottom: '12px',
                display: 'block',
                letterSpacing: '0.2px'
              }}>
                التقييم
              </label>
              <Select
                value={acquisitionGrade}
                onChange={(e) => setAcquisitionGrade(e.target.value)}
                options={[
                  { value: 'excellent', label: 'ممتاز' },
                  { value: 'very_good', label: 'جيد جداً' },
                  { value: 'good', label: 'جيد' },
                  { value: 'fair', label: 'متوسط' },
                  { value: 'poor', label: 'ضعيف' },
                ]}
                style={{ borderRadius: '12px', height: '48px' }}
              />
            </div>
          </div>

          {/* Serial Number and prices in grid */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            {/* Serial Number */}
            <div style={{
              padding: '20px',
              background: 'var(--bg-surface-elevated)',
              borderRadius: '16px',
              border: '1px solid var(--border-subtle)'
            }}>
              <label style={{
                fontSize: '14px',
                fontWeight: '600',
                color: 'var(--text-primary)',
                marginBottom: '12px',
                display: 'block',
                letterSpacing: '0.2px'
              }}>
                الرقم التسلسلي (اختياري)
              </label>
              <Input
                type="text"
                value={acquisitionSerialNumber}
                onChange={(e) => setAcquisitionSerialNumber(e.target.value)}
                placeholder="أدخل الرقم التسلسلي..."
                style={{ borderRadius: '12px', height: '48px' }}
              />
            </div>

            {/* Purchase Price */}
            <div style={{
              padding: '20px',
              background: 'var(--bg-surface-elevated)',
              borderRadius: '16px',
              border: '1px solid var(--border-subtle)'
            }}>
              <label style={{
                fontSize: '14px',
                fontWeight: '600',
                color: 'var(--text-primary)',
                marginBottom: '12px',
                display: 'block',
                letterSpacing: '0.2px'
              }}>
                سعر الشراء
                <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
              </label>
              <Input
                type="number"
                value={acquisitionPrice}
                onChange={(e) => setAcquisitionPrice(e.target.value)}
                placeholder="أدخل السعر..."
                style={{ borderRadius: '12px', height: '48px' }}
              />
            </div>
          </div>

          <div style={{
            padding: '20px',
            background: 'var(--bg-surface-elevated)',
            borderRadius: '16px',
            border: '1px solid var(--border-subtle)'
          }}>
            <label style={{
              fontSize: '14px',
              fontWeight: '600',
              color: 'var(--text-primary)',
              marginBottom: '12px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              سعر البيع
              <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
            </label>
            <Input
              type="number"
              value={acquisitionSellingPrice}
              onChange={(e) => setAcquisitionSellingPrice(e.target.value)}
              placeholder="أدخل سعر البيع..."
              style={{ borderRadius: '12px', height: '48px' }}
            />
          </div>

          {/* Payment Status */}
          <div style={{
            padding: '20px',
            background: 'var(--bg-surface-elevated)',
            borderRadius: '16px',
            border: '1px solid var(--border-subtle)'
          }}>
            <label style={{
              fontSize: '14px',
              fontWeight: '600',
              color: 'var(--text-primary)',
              marginBottom: '12px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              حالة الدفع
            </label>
            <Select
              value={paymentStatus}
              onChange={(e) => setPaymentStatus(e.target.value)}
              options={[
                { value: 'paid', label: 'مدفوع' },
              ]}
              style={{ borderRadius: '12px', height: '48px' }}
            />
          </div>

          {/* Notes */}
          <div style={{
            padding: '20px',
            background: 'var(--bg-surface-elevated)',
            borderRadius: '16px',
            border: '1px solid var(--border-subtle)'
          }}>
            <label style={{
              fontSize: '14px',
              fontWeight: '600',
              color: 'var(--text-primary)',
              marginBottom: '12px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              ملاحظات
            </label>
            <Input
              type="text"
              value={acquisitionNotes}
              onChange={(e) => setAcquisitionNotes(e.target.value)}
              placeholder="أدخل ملاحظات..."
              style={{ borderRadius: '12px', height: '48px' }}
            />
          </div>

          <div style={{
            display: 'flex',
            gap: '16px',
            justifyContent: 'flex-end',
            paddingTop: '32px',
            borderTop: '1px solid var(--border-subtle)',
            marginTop: '24px'
          }}>
            <button
              onClick={handleCancelAcquisition}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '10px',
                minWidth: '120px',
                padding: '12px 24px',
                borderRadius: '12px',
                background: 'var(--bg-surface-elevated)',
                border: '1px solid var(--border-default)',
                color: 'var(--text-primary)',
                fontSize: '14px',
                fontWeight: '600',
                letterSpacing: '0.3px',
                cursor: 'pointer',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                boxShadow: '0 4px 20px rgba(0, 0, 0, 0.1), 0 1px 3px rgba(0, 0, 0, 0.05)',
                position: 'relative',
                overflow: 'hidden'
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                e.currentTarget.style.borderColor = 'var(--primary)';
                e.currentTarget.style.boxShadow = '0 8px 30px rgba(99, 102, 241, 0.2), 0 2px 8px rgba(0, 0, 0, 0.1)';
                e.currentTarget.style.transform = 'translateY(-2px) scale(1.02)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'var(--bg-surface-elevated)';
                e.currentTarget.style.borderColor = 'var(--border-default)';
                e.currentTarget.style.boxShadow = '0 4px 20px rgba(0, 0, 0, 0.1), 0 1px 3px rgba(0, 0, 0, 0.05)';
                e.currentTarget.style.transform = 'translateY(0) scale(1)';
              }}
            >
              إلغاء
            </button>
            <button
              onClick={handleSubmitAcquisition}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '10px',
                minWidth: '140px',
                padding: '12px 24px',
                borderRadius: '12px',
                background: 'var(--primary)',
                border: '1px solid var(--primary)',
                color: 'var(--text-on-primary)',
                fontSize: '14px',
                fontWeight: '600',
                letterSpacing: '0.3px',
                cursor: 'pointer',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                boxShadow: '0 4px 20px rgba(99, 102, 241, 0.3), 0 1px 3px rgba(99, 102, 241, 0.1)',
                position: 'relative',
                overflow: 'hidden'
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'var(--primary-hover)';
                e.currentTarget.style.borderColor = 'var(--primary-hover)';
                e.currentTarget.style.boxShadow = '0 8px 30px rgba(99, 102, 241, 0.4), 0 2px 8px rgba(99, 102, 241, 0.2)';
                e.currentTarget.style.transform = 'translateY(-2px) scale(1.02)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'var(--primary)';
                e.currentTarget.style.borderColor = 'var(--primary)';
                e.currentTarget.style.boxShadow = '0 4px 20px rgba(99, 102, 241, 0.3), 0 1px 3px rgba(99, 102, 241, 0.1)';
                e.currentTarget.style.transform = 'translateY(0) scale(1)';
              }}
            >
              <ShoppingCart className="w-5 h-5" style={{ position: 'relative', zIndex: 1 }} />
              <span style={{ position: 'relative', zIndex: 1 }}>شراء</span>
            </button>
          </div>
        </div>
      </Modal>

      <Modal
        isOpen={Boolean(partToEdit)}
        onClose={() => !updateMutation.isPending && setPartToEdit(null)}
        title="تعديل القطعة المستعملة"
        variant="modern"
        size="md"
      >
        <div className="space-y-4">
          <label className="block text-sm font-medium text-[var(--text-primary)]">
            اسم القطعة
            <Input type="text" value={editProductName} onChange={(event) => setEditProductName(event.target.value)} className="mt-2" />
          </label>
          <label className="block text-sm font-medium text-[var(--text-primary)]">
            نوع القطعة
            <Select
              value={editPartType}
              onChange={(event) => setEditPartType(event.target.value)}
              options={[
                { value: '', label: 'بدون نوع' },
                ...partTypes.map((partType: any) => ({ value: partType.id, label: partType.name_ar })),
              ]}
              className="mt-2"
            />
          </label>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <label className="block text-sm font-medium text-[var(--text-primary)]">
              الحالة
              <Select
                value={editCondition}
                onChange={(event) => setEditCondition(event.target.value)}
                options={[{ value: 'used', label: 'مستعمل' }, { value: 'refurbished', label: 'مجدد' }, { value: 'new', label: 'جديد' }]}
                className="mt-2"
              />
            </label>
            <label className="block text-sm font-medium text-[var(--text-primary)]">
              التقييم
              <Select
                value={editGrade}
                onChange={(event) => setEditGrade(event.target.value)}
                options={[
                  { value: 'excellent', label: 'ممتاز' },
                  { value: 'very_good', label: 'جيد جدًا' },
                  { value: 'good', label: 'جيد' },
                  { value: 'fair', label: 'متوسط' },
                  { value: 'poor', label: 'ضعيف' },
                ]}
                className="mt-2"
              />
            </label>
          </div>
          <label className="block text-sm font-medium text-[var(--text-primary)]">
            الرقم التسلسلي
            <Input type="text" value={editSerialNumber} onChange={(event) => setEditSerialNumber(event.target.value)} className="mt-2" />
          </label>
          <label className="block text-sm font-medium text-[var(--text-primary)]">
            سعر الشراء
            <Input
              type="number"
              min="0"
              step="0.01"
              value={editPurchaseCost}
              onChange={(event) => setEditPurchaseCost(event.target.value)}
              className="mt-2"
            />
          </label>
          <label className="block text-sm font-medium text-[var(--text-primary)]">
            سعر البيع
            <Input
              type="number"
              min="0"
              step="0.01"
              value={editSellingPrice}
              onChange={(event) => setEditSellingPrice(event.target.value)}
              className="mt-2"
            />
          </label>
          <label className="block text-sm font-medium text-[var(--text-primary)]">
            ملاحظات
            <Input
              type="text"
              value={editNotes}
              onChange={(event) => setEditNotes(event.target.value)}
              className="mt-2"
            />
          </label>
          <div className="flex justify-end gap-3 border-t border-[var(--border-subtle)] pt-4">
            <Button variant="secondary" onClick={() => setPartToEdit(null)} disabled={updateMutation.isPending}>
              إلغاء
            </Button>
            <Button variant="primary" onClick={handleSaveEdit} disabled={updateMutation.isPending}>
              {updateMutation.isPending ? 'جاري الحفظ...' : 'حفظ التعديل'}
            </Button>
          </div>
        </div>
      </Modal>

      <ConfirmDialog
        isOpen={isDeleteDialogOpen}
        onClose={() => {
          if (!deleteMutation.isPending) {
            setIsDeleteDialogOpen(false);
            setPartToDelete(null);
          }
        }}
        onConfirm={handleConfirmDelete}
        title="حذف القطعة"
        message="هل أنت متأكد من حذف هذه القطعة من المخزون؟ قد يتم أرشفتها بدل حذفها نهائيًا إذا كانت مرتبطة بعمليات سابقة."
        confirmText="حذف القطعة"
        cancelText="إلغاء"
        variant="danger"
        isLoading={deleteMutation.isPending}
      />

    </div>
  );
}