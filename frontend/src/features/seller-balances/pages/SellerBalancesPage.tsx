import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { acquisitionsApi } from '../../../services/api/endpoints';
import { Card, CardContent } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { SearchInput } from '../../../design-system/components/search-input';
import { PageHeader } from '../../../design-system/components/page-header';
import { Badge } from '../../../design-system/components/badge';
import { Modal } from '../../../design-system/components/modal';
import {
  DollarSign,
  TrendingUp,
  User,
  CheckCircle,
} from 'lucide-react';
import { toast } from 'sonner';

interface SellerBalance {
  customer_id?: string;
  customer_name: string;
  total_acquired: number;
  total_paid: number;
  balance: number;
  transaction_count: number;
  last_transaction?: string;
}

export function SellerBalancesPage() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedSeller, setSelectedSeller] = useState<SellerBalance | null>(null);
  const [isPaymentModalOpen, setIsPaymentModalOpen] = useState(false);
  const [paymentAmount, setPaymentAmount] = useState('');

  const { data: balancesData, isLoading } = useQuery({
    queryKey: ['seller-balances'],
    queryFn: () => acquisitionsApi.getSellerBalances(),
  });

  const balances = Array.isArray(balancesData?.data) ? balancesData.data : [];

  // Filter sellers
  const filteredBalances = balances.filter((balance: SellerBalance) => {
    return !searchQuery || 
      balance.customer_name?.toLowerCase().includes(searchQuery.toLowerCase());
  });

  const createPaymentMutation = useMutation({
    mutationFn: ({ customerId, amount }: { customerId: string; amount: number }) =>
      acquisitionsApi.createSellerBalancePayment(customerId, { amount }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['seller-balances'] });
      queryClient.invalidateQueries({ queryKey: ['acquisitions'] });
      toast.success('تم تسجيل الدفعة بنجاح!');
      setIsPaymentModalOpen(false);
      setPaymentAmount('');
      setSelectedSeller(null);
    },
    onError: (error) => {
      console.error('Payment failed:', error);
      toast.error('فشل تسجيل الدفعة');
    },
  });

  const handleMakePayment = () => {
    if (!selectedSeller || !selectedSeller.customer_id || !paymentAmount) {
      toast.error('يرجى إدخال مبلغ الدفعة');
      return;
    }

    const amount = parseFloat(paymentAmount);
    if (!Number.isFinite(amount) || amount <= 0 || amount > selectedSeller.balance) {
      toast.error(`يجب أن يكون المبلغ بين 0.01 و ₪${selectedSeller.balance.toFixed(2)}`);
      return;
    }
    createPaymentMutation.mutate({ 
      customerId: selectedSeller.customer_id,
      amount 
    });
  };

  const getBalanceStatus = (balance: number) => {
    if (balance > 0) return { variant: 'warning' as const, label: 'مستحق للبائع' };
    if (balance < 0) return { variant: 'info' as const, label: 'رصيد زائد للمتجر' };
    return { variant: 'success' as const, label: 'متوازن' };
  };

  return (
    <div>
      <PageHeader
        eyebrow="Seller Balances"
        title="رصيد البائعين"
        description="تتبع المدفوعات المستحقة للبائعين"
      />

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-cyan/10">
                <User className="w-5 h-5 text-cyan" />
              </div>
              <div>
                <p className="text-sm text-gray-400">إجمالي البائعين</p>
                <p className="text-2xl font-bold">{balances.length}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-green/10">
                <TrendingUp className="w-5 h-5 text-green" />
              </div>
              <div>
                <p className="text-sm text-gray-400">إجمالي المشتريات</p>
                <p className="text-2xl font-bold">
                  ₪{balances.reduce((sum: number, b: SellerBalance) => sum + b.total_acquired, 0).toFixed(2)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-blue/10">
                <CheckCircle className="w-5 h-5 text-blue" />
              </div>
              <div>
                <p className="text-sm text-gray-400">إجمالي المدفوعات</p>
                <p className="text-2xl font-bold">
                  ₪{balances.reduce((sum: number, b: SellerBalance) => sum + b.total_paid, 0).toFixed(2)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-orange/10">
                <DollarSign className="w-5 h-5 text-orange" />
              </div>
              <div>
                <p className="text-sm text-gray-400">المستحق للبائعين</p>
                <p className="text-2xl font-bold">
                  ₪{balances.reduce((sum: number, b: SellerBalance) => sum + Math.max(b.balance, 0), 0).toFixed(2)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Search */}
      <Card className="mb-4 border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
        <CardContent className="p-4">
          <SearchInput
            placeholder="بحث عن بائع..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onClear={() => setSearchQuery('')}
            size="sm"
            className="w-full"
          />
        </CardContent>
      </Card>

      {/* Sellers List */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : filteredBalances.length === 0 ? (
        <Card className="border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
          <CardContent className="p-12 text-center">
            <User className="w-12 h-12 mx-auto mb-4 text-[var(--text-muted)]" />
            <p className="text-[var(--text-secondary)]">لا يوجد بائعين</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {filteredBalances.map((balance: SellerBalance) => {
            const status = getBalanceStatus(balance.balance);
            return (
              <Card key={balance.customer_id || balance.customer_name} className="border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm hover:shadow-md transition-shadow">
                <CardContent className="p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <div className="p-2 rounded-lg bg-cyan/10">
                          <User className="w-5 h-5 text-cyan" />
                        </div>
                        <div>
                          <h3 className="font-semibold text-[var(--text-primary)]">
                            {balance.customer_name}
                          </h3>
                          <p className="text-xs text-[var(--text-muted)]">
                            {balance.transaction_count} عملية
                          </p>
                        </div>
                      </div>
                      
                      <div className="grid grid-cols-3 gap-4 mt-3">
                        <div>
                          <p className="text-xs text-[var(--text-secondary)]">إجمالي المشتريات</p>
                          <p className="text-lg font-bold text-[var(--text-primary)]">
                            ₪{balance.total_acquired.toFixed(2)}
                          </p>
                        </div>
                        <div>
                          <p className="text-xs text-[var(--text-secondary)]">إجمالي المدفوعات</p>
                          <p className="text-lg font-bold text-[var(--text-primary)]">
                            ₪{balance.total_paid.toFixed(2)}
                          </p>
                        </div>
                        <div>
                          <p className="text-xs text-[var(--text-secondary)]">الرصيد</p>
                          <p className="text-lg font-bold" style={{
                            color: balance.balance > 0 ? 'var(--color-warning)' : 
                                   balance.balance < 0 ? 'var(--color-danger)' : 
                                   'var(--color-success)'
                          }}>
                            ₪{Math.abs(balance.balance).toFixed(2)}
                          </p>
                        </div>
                      </div>
                    </div>

                    <div className="flex flex-col items-end gap-2 mr-4">
                      <Badge variant={status.variant}>{status.label}</Badge>
                      {balance.balance > 0 && (
                        <Button
                          variant="primary"
                          size="sm"
                          onClick={() => {
                            setSelectedSeller(balance);
                            setIsPaymentModalOpen(true);
                          }}
                        >
                          <DollarSign className="w-4 h-4 mr-1" />
                          دفع
                        </Button>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* Payment Modal */}
      <Modal
        isOpen={isPaymentModalOpen}
        onClose={() => setIsPaymentModalOpen(false)}
        title="تسجيل دفعة"
        variant="modern"
        size="md"
      >
        <div className="space-y-4">
          {selectedSeller && (
            <div style={{
              padding: '16px',
              background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.05) 0%, rgba(34, 211, 238, 0.05) 100%)',
              borderRadius: '12px',
              border: '1px solid rgba(99, 102, 241, 0.2)'
            }}>
              <p className="text-sm text-[var(--text-secondary)]">البائع</p>
              <p className="font-semibold text-[var(--text-primary)]">{selectedSeller.customer_name}</p>
              <p className="text-sm text-[var(--text-secondary)] mt-2">الرصيد الحالي</p>
              <p className="font-bold text-lg" style={{
                color: selectedSeller.balance > 0 ? 'var(--color-warning)' : 
                       selectedSeller.balance < 0 ? 'var(--color-danger)' : 
                       'var(--color-success)'
              }}>
                ₪{Math.abs(selectedSeller.balance).toFixed(2)}
              </p>
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-[var(--text-primary)] mb-2">
              مبلغ الدفعة
            </label>
            <Input
              type="number"
              value={paymentAmount}
              onChange={(e) => setPaymentAmount(e.target.value)}
              placeholder="أدخل المبلغ..."
              step="0.01"
              min="0.01"
              max={selectedSeller?.balance}
            />
            {selectedSeller && (
              <p className="mt-2 text-xs text-[var(--text-secondary)]">
                الحد الأقصى للدفعة: ₪{selectedSeller.balance.toFixed(2)}
              </p>
            )}
          </div>

          <div className="flex gap-3 justify-end pt-4 border-t border-[var(--border-subtle)]">
            <Button
              variant="secondary"
              onClick={() => setIsPaymentModalOpen(false)}
            >
              إلغاء
            </Button>
            <Button
              variant="primary"
              onClick={handleMakePayment}
              disabled={!paymentAmount || createPaymentMutation.isPending}
            >
              {createPaymentMutation.isPending ? 'جاري التسجيل...' : 'تسجيل الدفعة'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
