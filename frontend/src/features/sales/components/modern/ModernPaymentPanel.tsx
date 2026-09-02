import { CreditCard, Banknote, Wallet, User, Loader2, AlertTriangle, CheckCircle2 } from 'lucide-react';
import { cn } from '../../../../utils';
import { PaymentMethod } from '../../types/pos.types';

interface ModernPaymentPanelProps {
  paymentMethod: PaymentMethod;
  setPaymentMethod: (method: PaymentMethod) => void;
  paidAmount: string;
  setPaidAmount: (value: string) => void;
  total: number;
  isProcessing: boolean;
  onCheckout: () => void;
  selectedCustomer?: string;
  customerBalance?: number;
  customerCreditLimit?: number;
  onQuickCustomerCreate?: () => void;
}

const PAYMENT_METHODS: Array<{
  id: PaymentMethod;
  label: string;
  icon: typeof Banknote;
}> = [
  { id: 'cash', label: 'نقداً', icon: Banknote },
  { id: 'card', label: 'بطاقة', icon: CreditCard },
  { id: 'checks', label: 'شيكات', icon: Wallet },
  { id: 'credit', label: 'دين', icon: User },
];

export function ModernPaymentPanel({
  paymentMethod,
  setPaymentMethod,
  paidAmount,
  setPaidAmount,
  total,
  isProcessing,
  onCheckout,
  selectedCustomer,
  customerBalance = 0,
  customerCreditLimit,
  onQuickCustomerCreate,
}: ModernPaymentPanelProps) {
  const paid = parseFloat(paidAmount) || 0;
  const remaining = total - paid;
  const isCreditSale = paymentMethod === 'credit';
  const isCreditSaleWithoutCustomer = isCreditSale && !selectedCustomer;
  const isCreditAdvanceMissing = isCreditSale && paidAmount.trim() === '';
  const projectedDebt = customerBalance + remaining;
  const willExceedCreditLimit = customerCreditLimit && projectedDebt > customerCreditLimit;

  const isCheckoutDisabled =
    isProcessing ||
    total === 0 ||
    (['cash', 'card', 'checks'].includes(paymentMethod) && paid < total) ||
    isCreditSaleWithoutCustomer ||
    isCreditAdvanceMissing;

  return (
    <div className="pos-modern-payment-panel">
      {/* Payment Method Selection */}
      <div className="payment-methods">
        {PAYMENT_METHODS.map(({ id, label, icon: Icon }) => (
          <button
            key={id}
            className={cn(
              'payment-method-btn',
              paymentMethod === id && 'active'
            )}
            onClick={() => {
              setPaymentMethod(id);
              if (id === 'credit') {
                setPaidAmount('');
              } else {
                setPaidAmount(total.toFixed(2));
              }
            }}
          >
            <Icon className={cn('w-4 h-4 method-icon')} />
            <span>{label}</span>
          </button>
        ))}
      </div>

      {/* Warnings */}
      {isCreditSaleWithoutCustomer && (
        <div className="payment-warning error">
          <AlertTriangle className="w-4 h-4" />
          <span>البيع بالدين يتطلب تحديد العميل</span>
        </div>
      )}

      {isCreditSale && selectedCustomer && willExceedCreditLimit && (
        <div className="payment-warning warning">
          <AlertTriangle className="w-4 h-4" />
          <span>تنبيه: هذا البيع سيتجاوز حد الدين المحدد</span>
        </div>
      )}

      {/* Quick Customer for Credit */}
      {isCreditSale && !selectedCustomer && onQuickCustomerCreate && (
        <button className="payment-quick-customer" onClick={onQuickCustomerCreate}>
          <User className="w-4 h-4" />
          <span>إنشاء عميل سريع</span>
        </button>
      )}

      {/* Amount Input */}
      {(paymentMethod === 'cash' || paymentMethod === 'credit') && (
        <div className="payment-amount-section">
          <label className="payment-amount-label">
            {paymentMethod === 'credit' ? 'الدفعة المقدمة' : 'المبلغ المدفوع'}
          </label>
          <div className="payment-amount-input-wrapper">
            <input
              type="number"
              className="payment-amount-input"
              value={paidAmount}
              onChange={(e) => setPaidAmount(e.target.value)}
              placeholder="0"
              min={0}
            />
            <span className="payment-currency">₪</span>
          </div>
          {isCreditAdvanceMissing && (
            <p className="payment-error-text">يرجى إدخال رقم للدفعة المقدمة</p>
          )}
          {remaining > 0 && paymentMethod === 'cash' && (
            <p className="payment-remaining">المتبقي: ₪{remaining.toLocaleString()}</p>
          )}
          {remaining < 0 && (
            <p className="payment-change">المردود: ₪{Math.abs(remaining).toLocaleString()}</p>
          )}
        </div>
      )}

      {/* Quick Amount Buttons for Cash */}
      {paymentMethod === 'cash' && total > 0 && (
        <div className="quick-amounts">
          {[50, 100, 200, 500].map((amount) => (
            <button
              key={amount}
              className="quick-amount-btn"
              onClick={() => setPaidAmount(String(amount))}
            >
              ₪{amount}
            </button>
          ))}
          <button
            className="quick-amount-btn exact"
            onClick={() => setPaidAmount(total.toFixed(2))}
          >
            بالضبط
          </button>
        </div>
      )}

      {/* Checkout Button */}
      <button
        className={cn('checkout-btn', isCheckoutDisabled && 'disabled')}
        onClick={onCheckout}
        disabled={isCheckoutDisabled}
      >
        {isProcessing ? (
          <>
            <Loader2 className="w-5 h-5 animate-spin" />
            <span>جاري المعالجة...</span>
          </>
        ) : (
          <>
            <CheckCircle2 className="w-5 h-5" />
            <span>إتمام البيع</span>
            {total > 0 && <span className="checkout-total">₪{total.toLocaleString()}</span>}
          </>
        )}
      </button>
    </div>
  );
}
