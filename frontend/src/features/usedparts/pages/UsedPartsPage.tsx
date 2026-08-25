import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { inventoryApi, partTypesApi, customersApi, productsApi, barcodeApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { SearchInput } from '../../../components/ui/search-input';
import { Select } from '../../../components/ui/select';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { Modal } from '../../../components/ui/modal';
import { getButtonSize } from '../../../config/button-sizes';
import { TradeInFormData } from '../../sales/types/pos.types';
import { playScanSound } from '../../../hooks/useBarcodeContext';
import { 
  Plus, 
  Edit, 
  Trash2, 
  Search,
  Filter,
  Layers,
  Package,
  Cpu,
  Keyboard,
  Barcode,
  Type,
  Camera,
  ShoppingCart,
  RefreshCw
} from 'lucide-react';

export function UsedPartsPage() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedPartType, setSelectedPartType] = useState<string>('');
  
  // Trade-in modal state
  const [isTradeInModalOpen, setIsTradeInModalOpen] = useState(false);
  const [isCustomerManual, setIsCustomerManual] = useState(false);
  const [tradeInCustomer, setTradeInCustomer] = useState('');
  const [tradeInCustomerManual, setTradeInCustomerManual] = useState('');
  const [isProductManual, setIsProductManual] = useState(false);
  const [tradeInProduct, setTradeInProduct] = useState('');
  const [tradeInProductManual, setTradeInProductManual] = useState('');
  const [tradeInPartType, setTradeInPartType] = useState('');
  const [tradeInPrice, setTradeInPrice] = useState('');
  const [tradeInSpecifications, setTradeInSpecifications] = useState<any[]>([]);
  
  // Barcode scanner state
  const [barcodeInput, setBarcodeInput] = useState('');
  const [inputMethod, setInputMethod] = useState<'barcode' | 'manual' | 'camera'>('barcode');
  const [isCameraScannerOpen, setIsCameraScannerOpen] = useState(false);
  const [soundEnabled, setSoundEnabled] = useState(true);

  const handleClearSearch = () => {
    setSearchQuery('');
  };

  const { data: inventoryData, isLoading } = useQuery({
    queryKey: ['inventory'],
    queryFn: () => inventoryApi.list({ page: 1, per_page: 100 }),
  });

  const { data: partTypesData, error: partTypesError, isLoading: partTypesLoading } = useQuery({
    queryKey: ['part-types'],
    queryFn: () => partTypesApi.list(),
    retry: 1,
  });

  const { data: customersData, isLoading: customersLoading } = useQuery({
    queryKey: ['customers'],
    queryFn: () => customersApi.list({ page: 1, per_page: 100 }),
  });

  const { data: productsData, isLoading: productsLoading } = useQuery({
    queryKey: ['products'],
    queryFn: () => productsApi.list({ page: 1, per_page: 100 }),
  });

  const inventoryItems = (inventoryData?.data as any[]) || [];
  const partTypes = (partTypesData?.data as any[]) || [];
  const customers = (customersData?.data as unknown) as any[] || [];
  const products = (productsData?.data?.products as unknown) as any[] || [];

  // Handle part types error gracefully
  if (partTypesError) {
    console.warn('Failed to load part types:', partTypesError);
  }

  // Trade-in mutation
  const createTradeInMutation = useMutation({
    mutationFn: (data: TradeInFormData) => inventoryApi.createTradeIn(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      alert('تم شراء القطعة المستعملة بنجاح!');
      handleCancelTradeIn();
    },
    onError: (error) => {
      console.error('Trade-in failed:', error);
      alert('فشل شراء القطعة المستعملة');
    },
  });

  // Barcode scan handler
  const handleBarcodeScan = async (e: React.FormEvent) => {
    e.preventDefault();
    if (barcodeInput.trim()) {
      try {
        const response = await barcodeApi.lookupProduct(barcodeInput.trim());
        const product = response as any;
        
        if (product && product.id) {
          // Open trade-in modal with product pre-filled
          setTradeInProduct(product.id);
          setIsProductManual(false);
          setIsTradeInModalOpen(true);
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

  const handleCameraScan = (barcode: string) => {
    setBarcodeInput(barcode);
    handleBarcodeScan(new Event('submit') as any);
  };

  // Trade-in handlers
  const handleSubmitTradeIn = async () => {
    const customerValue = isCustomerManual ? tradeInCustomerManual : tradeInCustomer;
    const productValue = isProductManual ? tradeInProductManual : tradeInProduct;

    if (!customerValue || !tradeInPrice || !tradeInPartType) {
      alert('يرجى ملء جميع الحقول المطلوبة');
      return;
    }

    const data: TradeInFormData = {
      customerId: isCustomerManual ? undefined : tradeInCustomer,
      customerName: isCustomerManual ? tradeInCustomerManual : undefined,
      productId: (isProductManual || !productValue) ? undefined : tradeInProduct,
      productName: isProductManual ? tradeInProductManual : undefined,
      partTypeId: tradeInPartType,
      purchaseCost: parseFloat(tradeInPrice) * 100,
      specifications: tradeInSpecifications,
    };

    createTradeInMutation.mutate(data);
  };

  const handleCancelTradeIn = () => {
    setTradeInCustomer('');
    setTradeInCustomerManual('');
    setIsCustomerManual(false);
    setTradeInProduct('');
    setTradeInProductManual('');
    setIsProductManual(false);
    setTradeInPrice('');
    setTradeInPartType('');
    setTradeInSpecifications([]);
    setIsTradeInModalOpen(false);
  };

  // Filter used parts only
  const usedParts = inventoryItems.filter((item: any) => 
    item.condition === 'USED' && item.status === 'AVAILABLE'
  );

  // Filter by search and part type
  const filteredParts = usedParts.filter((item: any) => {
    const matchesSearch = !searchQuery || 
      (item.product_name && item.product_name.toLowerCase().includes(searchQuery.toLowerCase())) ||
      (item.notes && item.notes.toLowerCase().includes(searchQuery.toLowerCase()));
    
    const matchesType = !selectedPartType || item.part_type_id === selectedPartType;
    
    return matchesSearch && matchesType;
  });

  const getIconComponent = (iconName: string) => {
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
        description="إدارة القطع المستعملة ومواصفاتها"
        actions={
          <button
            onClick={() => setIsTradeInModalOpen(true)}
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '8px',
              padding: '10px 16px',
              borderRadius: '12px',
              background: 'linear-gradient(135deg, var(--primary) 0%, var(--primary-hover) 100%)',
              border: '1px solid var(--primary)',
              color: 'var(--text-on-primary)',
              fontSize: '13px',
              fontWeight: '600',
              letterSpacing: '0.2px',
              cursor: 'pointer',
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
              boxShadow: '0 4px 20px rgba(99, 102, 241, 0.3), 0 1px 3px rgba(99, 102, 241, 0.1)',
              position: 'relative',
              overflow: 'hidden'
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'linear-gradient(135deg, var(--primary-hover) 0%, var(--primary) 100%)';
              e.currentTarget.style.borderColor = 'var(--primary-hover)';
              e.currentTarget.style.boxShadow = '0 8px 30px rgba(99, 102, 241, 0.4), 0 2px 8px rgba(99, 102, 241, 0.2)';
              e.currentTarget.style.transform = 'translateY(-2px) scale(1.02)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'linear-gradient(135deg, var(--primary) 0%, var(--primary-hover) 100%)';
              e.currentTarget.style.borderColor = 'var(--primary)';
              e.currentTarget.style.boxShadow = '0 4px 20px rgba(99, 102, 241, 0.3), 0 1px 3px rgba(99, 102, 241, 0.1)';
              e.currentTarget.style.transform = 'translateY(0) scale(1)';
            }}
          >
            <RefreshCw className="w-4 h-4" style={{ position: 'relative', zIndex: 1 }} />
            <span style={{ position: 'relative', zIndex: 1 }}>شراء قطعة مستعملة</span>
          </button>
        }
      />

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3 mb-4">
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
                <p className="text-sm text-gray-400">قيمة المخزون</p>
                <p className="text-2xl font-bold">
                  ₪{usedParts.reduce((sum: number, item: any) => sum + (item.selling_price / 100), 0).toFixed(0)}
                </p>
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
                <p className="text-sm text-gray-400">أنواع القطع</p>
                <p className="text-2xl font-bold">{partTypes.length}</p>
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
              <div className="flex gap-2 items-stretch">
                <div className="flex-1 relative flex items-center">
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
                <Button type="submit" size="sm" className="px-4">
                  إضافة
                </Button>
              </div>
            </form>
          </div>
        </CardContent>
      </Card>

      {/* Search and Filters */}
      <Card className="mb-4" style={{
        background: 'linear-gradient(135deg, rgba(17, 24, 39, 0.6) 0%, rgba(17, 24, 39, 0.4) 100%)',
        border: '1px solid rgba(99, 102, 241, 0.15)',
        backdropFilter: 'blur(10px)'
      }}>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-3">
            <div className="flex-1">
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
              />
              <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                  setSearchQuery('');
                  setSelectedPartType('');
                }}
                style={{
                  height: '48px',
                  fontSize: '14px',
                  fontWeight: '600',
                  letterSpacing: '0.3px',
                  borderRadius: '8px',
                  background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.2) 0%, rgba(34, 211, 238, 0.2) 100%)',
                  border: '1px solid rgba(99, 102, 241, 0.3)',
                  color: 'var(--text-primary)',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'linear-gradient(135deg, rgba(99, 102, 241, 0.3) 0%, rgba(34, 211, 238, 0.3) 100%)';
                  e.currentTarget.style.borderColor = 'rgba(99, 102, 241, 0.5)';
                  e.currentTarget.style.transform = 'translateY(-1px)';
                  e.currentTarget.style.boxShadow = '0 4px 12px rgba(99, 102, 241, 0.2)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'linear-gradient(135deg, rgba(99, 102, 241, 0.2) 0%, rgba(34, 211, 238, 0.2) 100%)';
                  e.currentTarget.style.borderColor = 'rgba(99, 102, 241, 0.3)';
                  e.currentTarget.style.transform = 'translateY(0)';
                  e.currentTarget.style.boxShadow = 'none';
                }}
              >
                مسح
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Parts Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : filteredParts.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center">
            <Package className="w-12 h-12 mx-auto mb-4 text-gray-400" />
            <p className="text-gray-400 mb-4">لا توجد قطع مستعملة متاحة</p>
            <button
              onClick={() => setIsTradeInModalOpen(true)}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '8px',
                padding: '10px 16px',
                borderRadius: '12px',
                background: 'linear-gradient(135deg, var(--primary) 0%, var(--primary-hover) 100%)',
                border: '1px solid var(--primary)',
                color: 'var(--text-on-primary)',
                fontSize: '13px',
                fontWeight: '600',
                letterSpacing: '0.2px',
                cursor: 'pointer',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                boxShadow: '0 4px 20px rgba(99, 102, 241, 0.3), 0 1px 3px rgba(99, 102, 241, 0.1)',
                position: 'relative',
                overflow: 'hidden'
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'linear-gradient(135deg, var(--primary-hover) 0%, var(--primary) 100%)';
                e.currentTarget.style.borderColor = 'var(--primary-hover)';
                e.currentTarget.style.boxShadow = '0 8px 30px rgba(99, 102, 241, 0.4), 0 2px 8px rgba(99, 102, 241, 0.2)';
                e.currentTarget.style.transform = 'translateY(-2px) scale(1.02)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'linear-gradient(135deg, var(--primary) 0%, var(--primary-hover) 100%)';
                e.currentTarget.style.borderColor = 'var(--primary)';
                e.currentTarget.style.boxShadow = '0 4px 20px rgba(99, 102, 241, 0.3), 0 1px 3px rgba(99, 102, 241, 0.1)';
                e.currentTarget.style.transform = 'translateY(0) scale(1)';
              }}
            >
              <RefreshCw className="w-4 h-4" style={{ position: 'relative', zIndex: 1 }} />
              <span style={{ position: 'relative', zIndex: 1 }}>شراء قطعة مستعملة</span>
            </button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
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
                <CardContent className="p-4">
                  <div className="flex items-start justify-between mb-3">
                    {partType && (
                      <div
                        className="p-2 rounded-lg"
                        style={{ background: `${partType.color}30`, color: partType.color }}
                      >
                        {getIconComponent(partType.icon)}
                      </div>
                    )}
                    <Badge variant="success">متاح</Badge>
                  </div>
                  
                  <h3 className="font-semibold mb-1">
                    {item.product_name || item.product?.name || 'قطعة بدون اسم'}
                  </h3>
                  
                  {partType && (
                    <p className="text-sm text-gray-400 mb-2">{partType.name_ar}</p>
                  )}
                  
                  <div className="space-y-2 mb-3">
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">اشتريت بـ:</span>
                      <span>₪{(item.purchase_cost / 100).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between text-sm font-semibold">
                      <span className="text-gray-400">سعر البيع:</span>
                      <span className="text-cyan">₪{(item.selling_price / 100).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">الربح:</span>
                      <span className={item.selling_price > item.purchase_cost ? 'text-green' : 'text-red'}>
                        ₪{((item.selling_price - item.purchase_cost) / 100).toFixed(2)}
                      </span>
                    </div>
                  </div>
                  
                  {item.notes && (
                    <p className="text-xs text-gray-500 mt-2 line-clamp-2">{item.notes}</p>
                  )}
                  
                  <div className="flex gap-2 mt-4">
                    <Button variant="secondary" size="sm" className="flex-1">
                      <Edit className="w-3 h-3 mr-1" />
                      تعديل
                    </Button>
                    <Button variant="primary" size="sm" className="flex-1">
                      بيع
                    </Button>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* Trade-in Modal */}
      <Modal
        isOpen={isTradeInModalOpen}
        onClose={handleCancelTradeIn}
        title="شراء قطع مستعملة"
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
        <div className="space-y-5">
          {/* الزبون */}
          <div>
            <label style={{ 
              fontSize: '12px', 
              fontWeight: '600', 
              color: 'var(--text-secondary)',
              marginBottom: '8px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              الزبون
              <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
            </label>
            <div style={{ display: 'flex', gap: '12px', marginBottom: '12px' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', color: 'var(--text-primary)', cursor: 'pointer' }}>
                <input
                  type="radio"
                  checked={!isCustomerManual}
                  onChange={() => setIsCustomerManual(false)}
                  style={{ width: '16px', height: '16px', accentColor: 'var(--primary)' }}
                />
                <span>اختر من القائمة</span>
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', color: 'var(--text-primary)', cursor: 'pointer' }}>
                <input
                  type="radio"
                  checked={isCustomerManual}
                  onChange={() => setIsCustomerManual(true)}
                  style={{ width: '16px', height: '16px', accentColor: 'var(--primary)' }}
                />
                <span>اكتب يدوياً</span>
              </label>
            </div>
            {!isCustomerManual ? (
              <Select
                value={tradeInCustomer}
                onChange={(e) => setTradeInCustomer(e.target.value)}
                loading={customersLoading}
                options={[
                  { value: '', label: 'اختر الزبون...' },
                  ...customers.map((c) => ({ value: c.id, label: c.name })),
                ]}
                emptyMessage="لا يوجد عملاء"
                style={{ borderRadius: '10px' }}
              />
            ) : (
              <Input
                type="text"
                value={tradeInCustomerManual}
                onChange={(e) => setTradeInCustomerManual(e.target.value)}
                placeholder="أدخل اسم الزبون..."
                style={{ borderRadius: '10px' }}
              />
            )}
          </div>

          {/* المنتج */}
          <div>
            <label style={{ 
              fontSize: '12px', 
              fontWeight: '600', 
              color: 'var(--text-secondary)',
              marginBottom: '8px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              المنتج (اختياري إذا اخترت نوع القطعة)
            </label>
            <div style={{ display: 'flex', gap: '12px', marginBottom: '12px' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', color: 'var(--text-primary)', cursor: 'pointer' }}>
                <input
                  type="radio"
                  checked={!isProductManual}
                  onChange={() => setIsProductManual(false)}
                  style={{ width: '16px', height: '16px', accentColor: 'var(--primary)' }}
                />
                <span>اختر من القائمة</span>
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', color: 'var(--text-primary)', cursor: 'pointer' }}>
                <input
                  type="radio"
                  checked={isProductManual}
                  onChange={() => setIsProductManual(true)}
                  style={{ width: '16px', height: '16px', accentColor: 'var(--primary)' }}
                />
                <span>اكتب يدوياً</span>
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', color: 'var(--text-primary)', cursor: 'pointer' }}>
                <input
                  type="radio"
                  checked={tradeInProduct === '' && tradeInProductManual === ''}
                  onChange={() => {
                    setIsProductManual(false);
                    setTradeInProduct('');
                    setTradeInProductManual('');
                  }}
                  style={{ width: '16px', height: '16px', accentColor: 'var(--primary)' }}
                />
                <span>بدون منتج</span>
              </label>
            </div>
            {!isProductManual ? (
              <Select
                value={tradeInProduct}
                onChange={(e) => setTradeInProduct(e.target.value)}
                loading={productsLoading}
                options={[
                  { value: '', label: 'اختر المنتج...' },
                  ...products.map((p) => ({ value: p.id, label: p.name })),
                ]}
                emptyMessage="لا يوجد منتجات"
                style={{ borderRadius: '10px' }}
              />
            ) : (
              <Input
                type="text"
                value={tradeInProductManual}
                onChange={(e) => setTradeInProductManual(e.target.value)}
                placeholder="أدخل اسم المنتج..."
                style={{ borderRadius: '10px' }}
              />
            )}
          </div>

          {/* نوع القطعة */}
          <div>
            <label style={{ 
              fontSize: '12px', 
              fontWeight: '600', 
              color: 'var(--text-secondary)',
              marginBottom: '8px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              نوع القطعة
              <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
            </label>
            <Select
              value={tradeInPartType}
              onChange={(e) => {
                setTradeInPartType(e.target.value);
                setTradeInSpecifications([]);
              }}
              loading={partTypesLoading}
              options={[
                { value: '', label: 'اختر نوع القطعة...' },
                ...partTypes.map((pt: any) => ({ value: pt.id, label: pt.name_ar })),
              ]}
              emptyMessage="لا يوجد أنواع قطع"
              style={{ borderRadius: '10px' }}
            />
          </div>

          {/* السعر */}
          <div>
            <label style={{ 
              fontSize: '12px', 
              fontWeight: '600', 
              color: 'var(--text-secondary)',
              marginBottom: '8px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              السعر (كم دفعت للزبون)
              <span style={{ color: 'var(--danger)', marginRight: '4px' }}>*</span>
            </label>
            <Input
              type="number"
              value={tradeInPrice}
              onChange={(e) => setTradeInPrice(e.target.value)}
              placeholder="أدخل السعر..."
              style={{ borderRadius: '10px' }}
            />
          </div>

          {/* مواصفات إضافية */}
          {tradeInPartType && (
            <div style={{ 
              padding: '12px 16px', 
              background: 'var(--bg-surface-elevated)', 
              borderRadius: '12px',
              border: '1px solid var(--border-subtle)'
            }}>
              <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}>المواصفات التفصيلية (قريباً)</p>
              <p style={{ fontSize: '11px', color: 'var(--text-muted)' }}>سيتم إضافة المواصفات التفصيلية لكل نوع قطعة قريباً</p>
            </div>
          )}

          <div style={{ 
            display: 'flex', 
            gap: '12px', 
            justifyContent: 'flex-end',
            paddingTop: '24px',
            borderTop: '1px solid var(--border-subtle)',
            marginTop: '16px'
          }}>
            <button
              onClick={handleCancelTradeIn}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '8px',
                minWidth: '100px',
                padding: '10px 20px',
                borderRadius: '12px',
                background: 'var(--bg-surface-elevated)',
                border: '1px solid var(--border-default)',
                color: 'var(--text-primary)',
                fontSize: '13px',
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
              onClick={handleSubmitTradeIn}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '8px',
                minWidth: '120px',
                padding: '10px 20px',
                borderRadius: '12px',
                background: 'var(--primary)',
                border: '1px solid var(--primary)',
                color: 'var(--text-on-primary)',
                fontSize: '13px',
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
              <ShoppingCart className="w-4 h-4" style={{ position: 'relative', zIndex: 1 }} />
              <span style={{ position: 'relative', zIndex: 1 }}>شراء</span>
            </button>
          </div>
        </div>
      </Modal>
    </div>
  );
}