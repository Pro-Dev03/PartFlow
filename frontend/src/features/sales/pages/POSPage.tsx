import { useState, useEffect, useRef, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { Badge } from '../../../components/ui/badge';
import { PageHeader } from '../../../components/ui/page-header';
import { productsApi, salesApi, customersApi, barcodeApi } from '../../../services/api/endpoints';
import { BarcodeContext, playScanSound } from '../../../hooks/useBarcodeContext';
import { 
  Scan, 
  Search, 
  ShoppingCart, 
  Trash2,
  CreditCard,
  DollarSign,
  Banknote,
  Printer,
  Send,
  Loader2,
  Plus,
  Minus,
  X,
  Zap,
  Clock,
  User,
  Sparkles,
  Rocket,
  Package
} from 'lucide-react';

interface CartItem {
  id: string;
  name: string;
  barcode: string;
  price: number;
  quantity: number;
  total: number;
  stock?: number;
}

export function POSPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const barcodeInputRef = useRef<HTMLInputElement>(null);
  
  const [cart, setCart] = useState<CartItem[]>([]);
  const [barcodeInput, setBarcodeInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCustomer, setSelectedCustomer] = useState('');
  const [paymentMethod, setPaymentMethod] = useState<'cash' | 'card' | 'bank_transfer' | 'credit'>('cash');
  const [paidAmount, setPaidAmount] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);
  const [soundEnabled, setSoundEnabled] = useState(true);
  const [quickAddMode, setQuickAddMode] = useState(false);

  // Fetch products for search
  const { data: productsData } = useQuery({
    queryKey: ['products'],
    queryFn: () => productsApi.list(),
  });

  // Fetch customers
  const { data: customersData } = useQuery({
    queryKey: ['customers'],
    queryFn: () => customersApi.list(),
  });

  const products = (productsData as unknown) as any[];
  const customers = (customersData as unknown) as any[];

  // Create sale mutation
  const createSaleMutation = useMutation({
    mutationFn: (data: any) => salesApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sales'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      setCart([]);
      setBarcodeInput('');
      setPaidAmount('');
      setSelectedCustomer('');
      setIsProcessing(false);
      // Focus back on barcode input for next sale
      setTimeout(() => barcodeInputRef.current?.focus(), 100);
    },
    onError: (error) => {
      console.error('Sale failed:', error);
      setIsProcessing(false);
    },
  });

  const addToCart = useCallback((item: any) => {
    const existingItem = cart.find((c) => c.barcode === item.barcode);
    if (existingItem) {
      setCart(cart.map((c) => 
        c.barcode === item.barcode 
          ? { ...c, quantity: c.quantity + 1, total: (c.quantity + 1) * c.price }
          : c
      ));
      // Play sound for quantity increase
      if (soundEnabled) {
        playScanSound(true);
      }
    } else {
      setCart([...cart, {
        id: item.id,
        name: item.name,
        barcode: item.barcode,
        price: item.price,
        quantity: 1,
        total: item.price,
        stock: item.stock,
      }]);
      // Play sound for new item
      if (soundEnabled) {
        playScanSound(true);
      }
    }
    // Clear barcode input and refocus
    setBarcodeInput('');
    setTimeout(() => barcodeInputRef.current?.focus(), 100);
  }, [cart, soundEnabled]);

  const removeFromCart = (barcode: string) => {
    setCart(cart.filter((item) => item.barcode !== barcode));
  };

  const updateQuantity = (barcode: string, quantity: number) => {
    if (quantity <= 0) {
      removeFromCart(barcode);
      return;
    }
    setCart(cart.map((item) => 
      item.barcode === barcode 
        ? { ...item, quantity, total: quantity * item.price }
        : item
    ));
  };

  const handleBarcodeScan = async (e: React.FormEvent) => {
    e.preventDefault();
    if (barcodeInput.trim()) {
      try {
        // استخدام API الحقيقي أولاً
        const response = await barcodeApi.lookup(barcodeInput.trim());
        const product = response as any;
        
        if (product && product.id) {
          addToCart({
            id: product.id,
            name: product.name,
            barcode: barcodeInput.trim(),
            price: product.sellingPrice,
            stock: product.stock,
          });
        } else {
          // عرض رسالة واضحة
          if (soundEnabled) {
            playScanSound(false);
          }
          // يمكن إضافة toast notification هنا
        }
      } catch (error) {
        console.error('Barcode lookup failed:', error);
        // Play error sound
        if (soundEnabled) {
          playScanSound(false);
        }
        // Fallback to product search by barcode (SKU)
        const product = products?.find((p: any) => 
          p.sku === barcodeInput.trim() || 
          p.barcode === barcodeInput.trim()
        );
        if (product) {
          addToCart({
            id: product.id,
            name: product.name,
            barcode: barcodeInput.trim(),
            price: product.sellingPrice,
            stock: product.stock,
          });
        }
      }
    }
  };

  const subtotal = cart.reduce((sum, item) => sum + item.total, 0);
  const tax = subtotal * 0.15; // 15% tax
  const total = subtotal + tax;
  const paid = parseFloat(paidAmount) || 0;
  const remaining = total - paid;

  const handleCheckout = useCallback(() => {
    if (cart.length === 0) return;

    setIsProcessing(true);

    const saleData = {
      customerId: selectedCustomer || undefined,
      items: cart.map(item => ({
        productId: item.id,
        quantity: item.quantity,
        price: item.price,
      })),
      paymentMethod,
      paidAmount: paymentMethod === 'credit' ? paid : total,
      notes: '',
    };

    createSaleMutation.mutate(saleData);
  }, [cart, selectedCustomer, paymentMethod, paid, total, createSaleMutation]);

  // Auto-focus barcode input on mount
  useEffect(() => {
    barcodeInputRef.current?.focus();
  }, []);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // F2 - Focus barcode input
      if (e.key === 'F2') {
        e.preventDefault();
        barcodeInputRef.current?.focus();
      }
      // F4 - Quick add mode
      if (e.key === 'F4') {
        e.preventDefault();
        setQuickAddMode(!quickAddMode);
      }
      // F9 - Complete sale
      if (e.key === 'F9' && cart.length > 0 && !isProcessing) {
        e.preventDefault();
        handleCheckout();
      }
      // Escape - Clear cart
      if (e.key === 'Escape' && cart.length > 0) {
        e.preventDefault();
        if (confirm('هل تريد مسح السلة؟')) {
          setCart([]);
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [cart, isProcessing, quickAddMode, handleCheckout]);

  // Filter products based on search
  const filteredProducts = (products || []).filter((product: any) =>
    product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    product.sku.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const paymentMethods = [
    { value: 'cash', label: 'نقد', icon: Banknote },
    { value: 'card', label: 'بطاقة', icon: CreditCard },
    { value: 'bank_transfer', label: 'تحويل', icon: DollarSign },
    { value: 'credit', label: 'دين', icon: CreditCard },
  ];

  return (
    <div className="h-[calc(100vh-8rem)]">
      {/* Page Header - Futuristic + Extremely Fast */}
      <PageHeader
        eyebrow="Express POS"
        title={t('sales.pos')}
        description="نقطة البيع فائقة السرعة مع مسح الباركود الفوري"
        actions={
          <div className="flex items-center gap-sm">
            <Button variant="secondary" className="gap-2">
              <Rocket className="w-4 h-4" />
              وضع سريع
            </Button>
            <Button variant="secondary" className="gap-2">
              <Zap className="w-4 h-4" />
              إعدادات
            </Button>
          </div>
        }
      />
      
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-md h-full">
        {/* Left Side - Product Selection */}
        <div className="lg:col-span-2 space-y-md">
          {/* Barcode Scanner - Futuristic + Extremely Fast */}
          <Card variant="featured">
            <CardContent className="p-lg">
              <form onSubmit={handleBarcodeScan} className="flex gap-md">
                <div className="flex-1 relative">
                  <Scan className="absolute inset-y-0 end-3 w-4 h-4 text-cyan" />
                  <Input
                    ref={barcodeInputRef}
                    placeholder={t('scanner.scanBarcode')}
                    value={barcodeInput}
                    onChange={(e) => setBarcodeInput(e.target.value)}
                    className="pe-10 border-cyan/30 focus:border-cyan"
                  />
                </div>
                <Button 
                  type="submit" 
                  className="gap-2"
                  variant={soundEnabled ? 'primary' : 'secondary'}
                  onClick={() => setSoundEnabled(!soundEnabled)}
                >
                  {soundEnabled ? (
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
                    </svg>
                  ) : (
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" />
                    </svg>
                  )}
                  {t('scanner.scan')}
                </Button>
              </form>
              
              {/* Keyboard shortcuts hint - Futuristic + Extremely Fast */}
              <div className="flex gap-md mt-2 text-tiny text-text-muted">
                <span className="flex items-center gap-1">
                  <kbd className="px-1.5 py-0.5 bg-cyan/10 text-cyan rounded-sm border border-cyan/20">F2</kbd>
                  <span>تركيز</span>
                </span>
                <span className="flex items-center gap-1">
                  <kbd className="px-1.5 py-0.5 bg-cyan/10 text-cyan rounded-sm border border-cyan/20">F9</kbd>
                  <span>إتمام</span>
                </span>
                <span className="flex items-center gap-1">
                  <kbd className="px-1.5 py-0.5 bg-cyan/10 text-cyan rounded-sm border border-cyan/20">ESC</kbd>
                  <span>مسح</span>
                </span>
              </div>
            </CardContent>
          </Card>

          {/* Product Search - Futuristic + Extremely Fast */}
          <Card>
            <CardContent className="p-lg">
              <div className="relative mb-md">
                <Search className="absolute inset-y-0 end-3 w-4 h-4 text-cyan" />
                <Input
                  placeholder={t('common.search')}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pe-10"
                />
              </div>

              {/* Product Grid */}
              <div className="grid grid-cols-2 md:grid-cols-3 gap-sm max-h-96 overflow-y-auto">
                {filteredProducts.slice(0, 12).map((product: any) => (
                  <Card
                    key={product.id}
                    className="cursor-pointer hover:border-cyan/30 hover:-translate-y-1 transition-all duration-normal"
                    onClick={() => addToCart({ 
                      id: product.id, 
                      name: product.name, 
                      barcode: product.sku || `BC-${product.id}`,
                      price: product.sellingPrice 
                    })}
                  >
                    <CardContent className="p-md">
                      <div className="aspect-square bg-surface-2 rounded-sm mb-sm flex items-center justify-center">
                        <Package className="w-8 h-8 text-cyan" />
                      </div>
                      <h3 className="text-small font-medium text-text line-clamp-2">
                        {product.name}
                      </h3>
                      <div className="flex items-center justify-between mt-sm">
                        <span className="text-small font-bold text-cyan">₪{product.sellingPrice?.toLocaleString()}</span>
                        <Badge variant={product.stock > 0 ? 'success' : 'danger'} size="sm">
                          {product.stock}
                        </Badge>
                      </div>
                    </CardContent>
                  </Card>
                ))}
                {filteredProducts.length === 0 && (
                  <div className="col-span-full text-center py-8 text-text-muted">
                    لا توجد منتجات
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Right Side - Cart */}
        <div className="space-y-md">
          <Card className="h-full flex flex-col">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <ShoppingCart className="w-5 h-5 text-cyan" />
                {t('sales.cart')}
                <Badge variant="secondary">{cart.length}</Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col">
              {/* Cart Items - Futuristic + Extremely Fast */}
              <div className="flex-1 overflow-y-auto space-y-sm mb-md">
                {cart.length === 0 ? (
                  <div className="text-center py-8 text-text-muted">
                    <ShoppingCart className="w-12 h-12 mx-auto mb-md opacity-50 text-cyan" />
                    <p className="text-small">السلة فارغة</p>
                  </div>
                ) : (
                  cart.map((item) => (
                    <div key={item.barcode} className="flex items-center gap-md p-md bg-surface-2 rounded-sm hover:bg-surface transition-colors duration-normal">
                      <div className="flex-1">
                        <p className="text-small font-medium text-text">
                          {item.name}
                        </p>
                        <p className="text-tiny text-text-muted">
                          {item.barcode}
                        </p>
                      </div>
                      <div className="flex items-center gap-sm">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => updateQuantity(item.barcode, item.quantity - 1)}
                        >
                          <Minus className="w-4 h-4" />
                        </Button>
                        <span className="w-8 text-center text-small font-medium text-text">{item.quantity}</span>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => updateQuantity(item.barcode, item.quantity + 1)}
                        >
                          <Plus className="w-4 h-4" />
                        </Button>
                      </div>
                      <div className="text-end">
                        <p className="text-small font-bold text-text">
                          ₪{item.total.toLocaleString()}
                        </p>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => removeFromCart(item.barcode)}
                      >
                        <Trash2 className="w-4 h-4 text-red" />
                      </Button>
                    </div>
                  ))
                )}
              </div>

              {/* Customer Selection */}
              <div className="space-y-sm mb-md">
                <Select
                  label="العميل"
                  value={selectedCustomer}
                  onChange={(e) => setSelectedCustomer(e.target.value)}
                  options={[
                    { value: '', label: 'عميل عابر' },
                    ...(customers || []).map((c: any) => ({ value: c.id, label: c.name })),
                  ]}
                />
              </div>

              {/* Payment Method - Futuristic + Extremely Fast */}
              <div className="space-y-sm mb-md">
                <label className="text-small font-medium text-text">
                  طريقة الدفع
                </label>
                <div className="grid grid-cols-2 gap-sm">
                  {paymentMethods.map((method) => {
                    const Icon = method.icon;
                    return (
                      <Button
                        key={method.value}
                        variant={paymentMethod === method.value ? 'primary' : 'secondary'}
                        onClick={() => setPaymentMethod(method.value as any)}
                        className="flex flex-col items-center gap-1 h-auto py-md"
                      >
                        <Icon className="w-4 h-4" />
                        <span className="text-tiny">{method.label}</span>
                      </Button>
                    );
                  })}
                </div>
              </div>

              {/* Payment Amount */}
              {paymentMethod === 'credit' && (
                <div className="space-y-sm mb-md">
                  <Input
                    label="المبلغ المدفوع"
                    type="number"
                    value={paidAmount}
                    onChange={(e) => setPaidAmount(e.target.value)}
                    placeholder="0.00"
                  />
                </div>
              )}

              {/* Totals - Futuristic + Extremely Fast */}
              <div className="border-t border-border pt-md space-y-sm">
                <div className="flex justify-between text-small">
                  <span className="text-text-muted">المجموع الفرعي</span>
                  <span className="text-small font-medium text-text">₪{subtotal.toLocaleString()}</span>
                </div>
                <div className="flex justify-between text-small">
                  <span className="text-text-muted">الضريبة (15%)</span>
                  <span className="text-small font-medium text-text">₪{tax.toLocaleString()}</span>
                </div>
                {paymentMethod === 'credit' && (
                  <>
                    <div className="flex justify-between text-small">
                      <span className="text-text-muted">المدفوع</span>
                      <span className="text-small font-medium text-green">₪{paid.toLocaleString()}</span>
                    </div>
                    <div className="flex justify-between text-small">
                      <span className="text-text-muted">المتبقي</span>
                      <span className="text-small font-medium text-yellow">₪{remaining.toLocaleString()}</span>
                    </div>
                  </>
                )}
                <div className="flex justify-between text-h3 font-bold pt-md border-t border-border">
                  <span className="text-text">المجموع</span>
                  <span className="text-cyan">₪{total.toLocaleString()}</span>
                </div>
              </div>

              {/* Action Buttons - Futuristic + Extremely Fast */}
              <div className="space-y-sm mt-md">
                <Button
                  variant="primary"
                  className="w-full"
                  size="lg"
                  disabled={cart.length === 0 || isProcessing}
                  onClick={handleCheckout}
                >
                  {isProcessing ? (
                    <>
                      <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                      جاري المعالجة...
                    </>
                  ) : (
                    <>
                      <Rocket className="w-4 h-4 mr-2" />
                      {t('sales.checkout')}
                    </>
                  )}
                </Button>
                <div className="grid grid-cols-2 gap-sm">
                  <Button variant="secondary" className="gap-2">
                    <Printer className="w-4 h-4" />
                    طباعة
                  </Button>
                  <Button variant="secondary" className="gap-2">
                    <Send className="w-4 h-4" />
                    إرسال
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}