import { useState } from 'react'
import { 
  CreditCard, 
  Banknote, 
  Wallet, 
  User, 
  Loader2, 
  AlertTriangle, 
  CheckCircle2,
  Smartphone,
  QrCode,
  Calendar,
  Receipt,
  Trash2,
  Zap
} from 'lucide-react'
import { cn } from '../../../../utils'
import { PaymentMethod } from '../../types/pos.types'

interface SplitPayment {
  id: string
  method: PaymentMethod
  amount: number
}

interface AdvancedPaymentPanelProps {
  paymentMethod: PaymentMethod
  setPaymentMethod: (method: PaymentMethod) => void
  paidAmount: string
  setPaidAmount: (value: string) => void
  total: number
  isProcessing: boolean
  onCheckout: () => void
  selectedCustomer?: string
  customerBalance?: number
  customerCreditLimit?: number
  onQuickCustomerCreate?: () => void
}

const PAYMENT_METHODS: Array<{
  id: PaymentMethod
  label: string
  icon: typeof Banknote
  color: string
}> = [
  { id: 'cash', label: 'نقداً', icon: Banknote, color: '#10b981' },
  { id: 'card', label: 'بطاقة', icon: CreditCard, color: '#3b82f6' },
  { id: 'checks', label: 'شيكات', icon: Wallet, color: '#8b5cf6' },
  { id: 'credit', label: 'دين', icon: User, color: '#f59e0b' },
]

const ELECTRONIC_METHODS: Array<{
  id: string
  label: string
  icon: typeof Smartphone
}> = [
  { id: 'apple_pay', label: 'Apple Pay', icon: Smartphone },
  { id: 'google_pay', label: 'Google Pay', icon: Smartphone },
  { id: 'qr_code', label: 'رمز QR', icon: QrCode },
]

export function AdvancedPaymentPanel({
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
}: AdvancedPaymentPanelProps) {
  const [paid, setPaid] = useState(parseFloat(paidAmount) || 0)
  const [splitPayments, setSplitPayments] = useState<SplitPayment[]>([])
  const [isSplitMode, setIsSplitMode] = useState(false)
  const [selectedElectronic, setSelectedElectronic] = useState<string | null>(null)
  const [installments, setInstallments] = useState(1)
  const [deferredDate, setDeferredDate] = useState('')
  const [checkNumber, setCheckNumber] = useState('')
  const [checkBank, setCheckBank] = useState('')
  const [checkDate, setCheckDate] = useState('')

  const remaining = total - paid
  const isCreditSale = paymentMethod === 'credit'
  const isCreditSaleWithoutCustomer = isCreditSale && !selectedCustomer
  const isCreditAdvanceMissing = isCreditSale && paidAmount.trim() === ''

  const addSplitPayment = () => {
    const remainingForSplit = total - splitPayments.reduce((sum, p) => sum + p.amount, 0)
    if (remainingForSplit <= 0) return
    
    setSplitPayments([...splitPayments, {
      id: Date.now().toString(),
      method: 'cash',
      amount: remainingForSplit,
    }])
  }

  const removeSplitPayment = (id: string) => {
    setSplitPayments(splitPayments.filter(p => p.id !== id))
  }

  const updateSplitPayment = (id: string, field: 'method' | 'amount', value: string | number) => {
    setSplitPayments(splitPayments.map(p => {
      if (p.id !== id) return p
      if (field === 'method') return { ...p, method: value as PaymentMethod }
      return { ...p, amount: Number(value) }
    }))
  }

  const isCheckoutDisabled =
    isProcessing ||
    total === 0 ||
    (['cash', 'card', 'checks'].includes(paymentMethod) && 
     !isSplitMode && 
     paid < total) ||
    (isSplitMode && splitPayments.reduce((sum, p) => sum + p.amount, 0) < total) ||
    isCreditSaleWithoutCustomer ||
    isCreditAdvanceMissing

  return (
    <div className="pos-advanced-payment-panel">
      {/* Payment Method Selection */}
      <div className="payment-section">
        <label className="section-label">طريقة الدفع</label>
        <div className="payment-methods-grid">
          {PAYMENT_METHODS.map(({ id, label, icon: Icon, color }) => (
            <button
              key={id}
              className={cn('payment-method-btn', paymentMethod === id && 'active')}
              onClick={() => {
                setPaymentMethod(id)
                if (id !== 'credit') {
                  setPaidAmount(total.toFixed(2))
                  setPaid(total)
                } else {
                  setPaidAmount('')
                  setPaid(0)
                }
              }}
              style={{
                '--method-color': color,
              } as React.CSSProperties}
            >
              <Icon className="w-4 h-4 method-icon" />
              <span>{label}</span>
            </button>
          ))}
        </div>
      </div>

      {/* Electronic Payments */}
      {(paymentMethod === 'card') && (
        <div className="payment-section">
          <label className="section-label">الدفع الإلكتروني</label>
          <div className="electronic-methods">
            {ELECTRONIC_METHODS.map(({ id, label, icon: Icon }) => (
              <button
                key={id}
                className={cn('electronic-btn', selectedElectronic === id && 'active')}
                onClick={() => setSelectedElectronic(selectedElectronic === id ? null : id)}
              >
                <Icon className="w-4 h-4" />
                <span>{label}</span>
              </button>
            ))}
          </div>
          {selectedElectronic === 'qr_code' && (
            <div className="qr-placeholder">
              <QrCode className="w-16 h-16" />
              <span>امسح رمز QR للدفع</span>
            </div>
          )}
        </div>
      )}

      {/* Split Payment Mode */}
      <div className="payment-section">
        <div className="section-header">
          <label className="section-label">تقسيم الدفع</label>
          <button
            className={cn('toggle-btn', isSplitMode && 'active')}
            onClick={() => setIsSplitMode(!isSplitMode)}
          >
            {isSplitMode ? 'مفعل' : 'غير مفعل'}
          </button>
        </div>

        {isSplitMode && (
          <div className="split-payments">
            {splitPayments.map((payment) => (
              <div key={payment.id} className="split-payment-row">
                <select
                  value={payment.method}
                  onChange={(e) => updateSplitPayment(payment.id, 'method', e.target.value)}
                  className="split-method-select"
                >
                  <option value="cash">نقداً</option>
                  <option value="card">بطاقة</option>
                  <option value="checks">شيكات</option>
                </select>
                <input
                  type="number"
                  value={payment.amount}
                  onChange={(e) => updateSplitPayment(payment.id, 'amount', e.target.value)}
                  className="split-amount-input"
                  min={0}
                />
                <button
                  className="remove-split-btn"
                  onClick={() => removeSplitPayment(payment.id)}
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
            ))}
            <button className="add-split-btn" onClick={addSplitPayment}>
              <Plus className="w-4 h-4" />
              <span>إضافة طريقة دفع</span>
            </button>
            <div className="split-total">
              <span>المبلغ المقسم: ₪{splitPayments.reduce((sum, p) => sum + p.amount, 0).toLocaleString()}</span>
              <span className={cn('remaining', remaining > 0 && 'text-warning')}>
                المتبقي: ₪{(total - splitPayments.reduce((sum, p) => sum + p.amount, 0)).toLocaleString()}
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Check Details */}
      {paymentMethod === 'checks' && (
        <div className="payment-section">
          <label className="section-label">تفاصيل الشيك</label>
          <div className="check-details">
            <input
              type="text"
              placeholder="رقم الشيك"
              value={checkNumber}
              onChange={(e) => setCheckNumber(e.target.value)}
              className="check-input"
            />
            <input
              type="text"
              placeholder="اسم البنك"
              value={checkBank}
              onChange={(e) => setCheckBank(e.target.value)}
              className="check-input"
            />
            <input
              type="date"
              value={checkDate}
              onChange={(e) => setCheckDate(e.target.value)}
              className="check-input"
            />
          </div>
        </div>
      )}

      {/* Installments (for card) */}
      {paymentMethod === 'card' && (
        <div className="payment-section">
          <label className="section-label">الأقساط</label>
          <div className="installments-selector">
            {[1, 3, 6, 12].map((num) => (
              <button
                key={num}
                className={cn('installment-btn', installments === num && 'active')}
                onClick={() => setInstallments(num)}
              >
                {num === 1 ? 'دفعة واحدة' : `${num} أقساط`}
              </button>
            ))}
          </div>
          {installments > 1 && (
            <div className="installment-info">
              <Calendar className="w-4 h-4" />
              <span>كل قسط: ₪{(total / installments).toFixed(2)}</span>
            </div>
          )}
        </div>
      )}

      {/* Deferred Payment (for credit) */}
      {isCreditSale && (
        <div className="payment-section">
          <label className="section-label">تاريخ الاستحقاق</label>
          <input
            type="date"
            value={deferredDate}
            onChange={(e) => setDeferredDate(e.target.value)}
            className="deferred-date-input"
          />
          {selectedCustomer && customerBalance > 0 && (
            <div className="customer-balance">
              <User className="w-4 h-4" />
              <span>رصيد العميل: ₪{customerBalance.toLocaleString()}</span>
            </div>
          )}
        </div>
      )}

      {/* Credit Sale Warning */}
      {isCreditSaleWithoutCustomer && (
        <div className="payment-warning error">
          <AlertTriangle className="w-4 h-4" />
          <span>البيع بالدين يتطلب تحديد العميل</span>
        </div>
      )}

      {/* Quick Customer for Credit */}
      {isCreditSale && !selectedCustomer && onQuickCustomerCreate && (
        <button className="payment-quick-customer" onClick={onQuickCustomerCreate}>
          <User className="w-4 h-4" />
          <span>إنشاء عميل سريع</span>
        </button>
      )}

      {/* Amount Input for Cash/Credit */}
      {(paymentMethod === 'cash' || paymentMethod === 'credit') && !isSplitMode && (
        <div className="payment-amount-section">
          <label className="payment-amount-label">
            {paymentMethod === 'credit' ? 'الدفعة المقدمة' : 'المبلغ المدفوع'}
          </label>
          <div className="payment-amount-input-wrapper">
            <input
              type="number"
              className="payment-amount-input"
              value={paidAmount}
              onChange={(e) => {
                setPaidAmount(e.target.value)
                setPaid(parseFloat(e.target.value) || 0)
              }}
              placeholder="0"
              min={0}
            />
             <span className="payment-currency">₪</span>
          </div>
          
          {/* Quick Amount Buttons */}
          {paymentMethod === 'cash' && (
            <div className="quick-amounts">
              {[50, 100, 200, 500].map((amount) => (
                <button
                  key={amount}
                  className="quick-amount-btn"
                  onClick={() => {
                    setPaidAmount(String(amount))
                    setPaid(amount)
                  }}
                >
                  ₪{amount}
                </button>
              ))}
              <button
                className="quick-amount-btn exact"
                onClick={() => {
                  setPaidAmount(total.toFixed(2))
                  setPaid(total)
                }}
              >
                بالضبط
              </button>
            </div>
          )}

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
             <span className="checkout-total">₪{total.toLocaleString()}</span>
          </>
        )}
      </button>
    </div>
  )
}
