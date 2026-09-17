import { useEffect, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { inventoryApi, partTypesApi, customersApi, productsApi, acquisitionsApi } from '../../../services/api/endpoints';
import { Card, CardContent } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { SearchInput } from '../../../components/ui/search-input';
import { Select } from '../../../components/ui/select';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { StatCard } from '../../../components/ui/stat-card';
import { Modal } from '../../../components/ui/modal';
import { OpeningStockModal } from '../../inventory/components/OpeningStockModal';
import { ConfirmDialog } from '../../../components/ui/confirm-dialog';
import { PaginationControls } from '../../../components/ui/pagination-controls';
import { toast } from 'sonner';
import { getPartTypeImage } from '../../../services/localPartTypeImages';
import {
  Plus,
  Layers,
  Package,
  Cpu,
  ShoppingCart,
  TrendingUp,
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
  const [showLowStockOnly, setShowLowStockOnly] = useState(false);
  const [page, setPage] = useState(1);
  const pageSize = 10;
  
  // Acquisition modal state
  const [isAcquisitionModalOpen, setIsAcquisitionModalOpen] = useState(false);
    const [isOpeningStockModalOpen, setIsOpeningStockModalOpen] = useState(false);
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

  const handleClearSearch = () => {
    setSearchQuery('');
  };

  const { data: inventoryData, isLoading } = useQuery({
    queryKey: ['inventory', 'used-parts', page, pageSize, showLowStockOnly],
    queryFn: () => inventoryApi.list({
      page: showLowStockOnly ? 1 : page,
      per_page: showLowStockOnly ? 1000 : pageSize,
      condition: 'USED',
      status: 'AVAILABLE',
    }),
    staleTime: 0,
    refetchOnMount: 'always',
  });

  const { data: usedPartsInsightData } = useQuery({
    queryKey: ['inventory', 'used-parts-insight'],
    queryFn: () => inventoryApi.list({ page: 1, per_page: 1000, condition: 'USED', status: 'AVAILABLE' }),
    staleTime: 60000,
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

  // Acquisition handlers
  const handleSubmitAcquisition = async () => {
    const customerValue = isCustomerManual ? acquisitionCustomerManual : acquisitionCustomer;

    if (!customerValue || !acquisitionProductName.trim() || !acquisitionPartType || !acquisitionPrice || !acquisitionSellingPrice) {
      toast.error('يرجى ملء جميع الحقول المطلوبة');
      return;
    }

    const serialNumber = acquisitionSerialNumber.trim();
    if (serialNumber) {
      const inventoryResponse = await inventoryApi.list({ page: 1, per_page: 1000 });
      const inventoryItems = Array.isArray(inventoryResponse?.data)
        ? inventoryResponse.data
        : Array.isArray(inventoryResponse?.data?.items) ? inventoryResponse.data.items : [];
      const serialExists = inventoryItems.some((item: any) =>
        String(item.serial_number || '').trim().toLowerCase() === serialNumber.toLowerCase(),
      );
      if (serialExists) {
        toast.error('الرقم التسلسلي موجود مسبقًا. استخدم رقمًا مختلفًا أو استخدم مسار المرتجع للقطعة نفسها.');
        return;
      }
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
    return condition === 'USED' && status === 'AVAILABLE';
  });
  const usedPartsForInsight = Array.isArray(usedPartsInsightData?.data)
    ? usedPartsInsightData.data
    : Array.isArray(usedPartsInsightData?.data?.items)
      ? usedPartsInsightData.data.items
      : usedParts;
  const usedStockByProduct = new Map<string, number>();
  usedPartsForInsight.forEach((item: any) => {
    const productId = String(item.product_id || item.product?.id || item.productId || '').trim();
    if (productId) {
      usedStockByProduct.set(productId, (usedStockByProduct.get(productId) || 0) + 1);
    }
  });
  const lowStockUsedProductIds = new Set(
    Array.from(usedStockByProduct.entries())
      .filter(([productId, stock]) => {
        const product = productById.get(productId) as any;
        const minimumStock = Math.max(1, Number(product?.min_stock_level) || 3);
        return stock <= minimumStock;
      })
      .map(([productId]) => productId),
  );
  const lowStockUsedCount = lowStockUsedProductIds.size;
  const sellableUsedParts = usedParts.filter(
    (item: any) => String(item.status || '').trim().toUpperCase() !== 'DAMAGED'
  );

  // Filter by search and part type
  const filteredParts = usedParts.filter((item: any) => {
    const productId = String(item.product_id || item.product?.id || item.productId || '').trim();
    if (showLowStockOnly && !lowStockUsedProductIds.has(productId)) return false;
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
        eyebrow="مخزون خاص"
        title="القطع المستعملة"
        description="سجّل القطع الموجودة، راقب قيمتها، وبعها من مكان واحد."
        actions={
          <div className="flex flex-col sm:flex-row gap-2 w-full sm:w-auto">
            <Button
              variant="primary"
              data-testid="open-used-opening-stock"
              onClick={() => setIsOpeningStockModalOpen(true)}
              className="min-h-11 shadow-[0_10px_24px_rgba(8,145,178,0.18)]"
            >
              <Plus className="w-4 h-4" />
              إضافة مخزون مستعمل موجود
            </Button>
            <Button variant="secondary" onClick={() => setIsAcquisitionModalOpen(true)} className="min-h-11">
              <ShoppingCart className="w-4 h-4" />
              شراء قطعة مستعملة
            </Button>
            <Button
              variant="secondary"
              onClick={() => navigate('/app/usedparts/stock')}
              className="min-h-11"
            >
              <Layers className="w-4 h-4" />
              عرض المخزون
            </Button>
          </div>
        }
      />

      <div className="premium-insight mb-3">
        <div className="premium-insight-icon"><Package className="h-3.5 w-3.5" /></div>
        <div className="premium-insight-copy">
          <p className="premium-insight-title">إضافة مخزون موجود</p>
          <p className="premium-insight-text">أضف القطع الموجودة لديك مباشرة إلى المخزون دون فاتورة شراء</p>
        </div>
        <Button
          type="button"
          variant="secondary"
          size="sm"
          data-testid="used-opening-stock-shortcut"
          onClick={() => setIsOpeningStockModalOpen(true)}
          className="premium-insight-action"
        >
          <Plus className="h-3.5 w-3.5" />
          إضافة
        </Button>
      </div>

      <div className="mb-4 grid gap-2" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))' }}>
        <StatCard title="إجمالي القطع" value={usedParts.length} icon={Layers} variant="featured" compact />
        <StatCard title="قيمة البيع" value={formatCurrency(totalSellingValue)} icon={ShoppingCart} variant="success" compact />
        <StatCard title="قيمة الشراء" value={formatCurrency(totalPurchaseValue)} icon={Package} variant="default" compact />
        <StatCard title="الربح المتوقع" value={formatCurrency(estimatedProfit)} icon={TrendingUp} variant="info" compact />
      </div>

      <OpeningStockModal
        isOpen={isOpeningStockModalOpen}
        onClose={() => setIsOpeningStockModalOpen(false)}
        onCreated={() => {
          queryClient.invalidateQueries({ queryKey: ['inventory'] });
          queryClient.invalidateQueries({ queryKey: ['products'] });
        }}
        stockType="used"
      />

      {/* Search and Filters */}
      <div className="mb-5 rounded-xl border border-[var(--border-default)] bg-[var(--bg-surface-elevated)] shadow-sm">
        <div className="p-4">
          <div className="mb-3 flex items-center justify-between gap-3">
            <div>
              <h2 className="text-sm font-bold text-[var(--text-primary)]">مخزونك الحالي</h2>
              <p className="mt-1 text-xs text-[var(--text-secondary)]">ابحث بالاسم أو صفِّ حسب نوع القطعة.</p>
            </div>
            <span className="hidden rounded-full border border-[var(--border-default)] px-3 py-1 text-xs text-[var(--text-secondary)] sm:inline-flex">
              {filteredParts.length} قطعة ظاهرة
            </span>
          </div>
          <div className="pf-search-row flex flex-col items-stretch gap-3 md:flex-row">
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
                  setShowLowStockOnly(false);
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
        </div>
      </div>

      {/* Dedicated used-parts stock */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : filteredParts.length === 0 ? (
        <Card className="border-dashed border-[var(--color-primary-25)] bg-[var(--bg-surface-elevated)]">
          <CardContent className="p-10 text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-[var(--color-primary-08)] text-[var(--color-primary)]">
              <Package className="h-7 w-7" />
            </div>
            <h3 className="font-semibold text-[var(--text-primary)]">لا توجد قطع مستعملة مطابقة</h3>
            <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-[var(--text-secondary)]">أضف قطعة مستعملة موجودة لديك أو غيّر البحث والفلاتر لعرض مخزون آخر.</p>
            <Button variant="primary" className="mt-5" onClick={() => setIsOpeningStockModalOpen(true)}>
              <Plus className="h-4 w-4" />
              إضافة مخزون مستعمل
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div id="used-parts-stock" className="overflow-hidden rounded-xl border border-[var(--border-default)] bg-[var(--bg-surface-elevated)]">
          {filteredParts.map((item: any) => {
            const partType = getPartType(item.part_type_id);
            return (
              <div
                key={item.id}
                className="group border-b border-[var(--border-default)] transition-colors last:border-b-0 hover:bg-[var(--bg-surface)]"
                style={{
                  borderRight: partType?.color ? `3px solid ${partType.color}55` : undefined,
                  display: 'grid',
                  gridTemplateColumns: 'minmax(180px, 1.4fr) minmax(120px, .8fr) repeat(3, minmax(90px, .55fr)) auto',
                  alignItems: 'center',
                  gap: '12px',
                  padding: '12px 16px',
                }}
              >
                <div className="flex min-w-0 items-center gap-3">
                  {getPartTypeImage(String(item.part_type_id)) ? (
                    <img
                      src={getPartTypeImage(String(item.part_type_id))}
                      alt=""
                      className="h-10 w-10 max-h-10 max-w-10 shrink-0 rounded-lg object-cover"
                      style={{ width: '40px', height: '40px', maxWidth: '40px', maxHeight: '40px' }}
                    />
                  ) : (
                    <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[var(--color-primary-08)] text-[var(--color-primary)]">
                      {getIconComponent(item.part_type_id)}
                    </div>
                  )}
                  <div className="min-w-0">
                    <h3 className="truncate text-sm font-semibold text-[var(--text-primary)]">{item.product_name || item.product?.name || 'قطعة بدون اسم'}</h3>
                    <p className="mt-0.5 truncate text-xs text-[var(--text-secondary)]">{partType?.name_ar || 'نوع غير محدد'}{item.notes ? ` · ${item.notes}` : ''}</p>
                  </div>
                </div>
                <div className="text-xs text-[var(--text-secondary)]">{sellerByInventoryItemId.get(item.id) || 'بدون مصدر'}</div>
                <div><span className="block text-[10px] text-[var(--text-secondary)]">شراء</span><span className="text-sm">₪{Number(item.purchase_cost || 0).toFixed(2)}</span></div>
                <div><span className="block text-[10px] text-[var(--text-secondary)]">بيع</span><span className="text-sm font-semibold text-cyan-500">₪{Number(item.selling_price || 0).toFixed(2)}</span></div>
                <div><span className="block text-[10px] text-[var(--text-secondary)]">هامش</span><span className={`text-sm font-semibold ${item.selling_price > item.purchase_cost ? 'text-emerald-500' : 'text-red-500'}`}>₪{(Number(item.selling_price || 0) - Number(item.purchase_cost || 0)).toFixed(2)}</span></div>
                <div className="flex items-center justify-end gap-1">
                  <Badge variant="success" className="hidden shrink-0 md:inline-flex">متاح</Badge>
                  <Button variant="primary" size="sm" onClick={() => handleSellItem(item)}>بيع</Button>
                  <Button variant="ghost" size="icon" onClick={() => handleEditPart(item)} title="تعديل القطعة" aria-label="تعديل القطعة"><Pencil className="h-4 w-4" /></Button>
                  <Button variant="ghost" size="icon" onClick={() => handleDeletePart(item.id)} title="حذف القطعة" aria-label="حذف القطعة"><Trash2 className="h-4 w-4" /></Button>
                </div>
              </div>
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