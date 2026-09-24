import { useEffect, useState } from 'react'
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
  Trash2,
  Plus,
  MessageCircle
} from 'lucide-react'
import { cn } from '../../../../utils'
import { PaymentAllocation, PaymentMethod } from '../../types/pos.types'
import { Modal } from '../../../../design-system/components/modal'

interface SplitPayment {
  id: string
  method: PaymentMethod
  amount: number
  checkNumber?: string
  checkBank?: string
  checkDate?: string
}

interface AdvancedPaymentPanelProps {
  paymentMethod: PaymentMethod
  setPaymentMethod: (method: PaymentMethod) => void
  paidAmount: string
  setPaidAmount: (value: string) => void
  total: number
  isProcessing: boolean
  onCheckout: () => void
  checkoutBlockedReason?: string
  onPaymentAllocationsChange?: (allocations: PaymentAllocation[]) => void
  electronicPaymentsEnabled?: boolean
  enabledElectronicMethods?: string[]
  selectedCustomer?: string
  customerBalance?: number
  customerCreditLimit?: number
  onQuickCustomerCreate?: () => void
  selectedCustomerName?: string
  storeName?: string
  installmentWhatsAppNumber?: string
  installmentWhatsAppMessage?: string
  installmentMonths: number
  setInstallmentMonths: (months: number) => void
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
  { id: 'installment', label: 'تقسيط', icon: Calendar, color: '#0ea5e9' },
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

let splitPaymentSequence = 0

const createSplitPaymentId = () => {
  if (typeof globalThis.crypto?.randomUUID === 'function') {
    return globalThis.crypto.randomUUID()
  }
  splitPaymentSequence += 1
  return `split-${Date.now()}-${splitPaymentSequence}`
}

export function AdvancedPaymentPanel({
  paymentMethod,
  setPaymentMethod,
  paidAmount,
  setPaidAmount,
  total,
  isProcessing,
  onCheckout,
  checkoutBlockedReason,
  onPaymentAllocationsChange,
  electronicPaymentsEnabled = false,
  enabledElectronicMethods = [],
  selectedCustomer,
  customerBalance = 0,
  customerCreditLimit,
  onQuickCustomerCreate,
  selectedCustomerName,
  storeName = 'PartFlow',
  installmentWhatsAppNumber,
  installmentWhatsAppMessage,
  installmentMonths,
  setInstallmentMonths,
}: AdvancedPaymentPanelProps) {
  const [paid, setPaid] = useState(parseFloat(paidAmount) || 0)
  const [splitPayments, setSplitPayments] = useState<SplitPayment[]>([])
  const [isPaymentMethodModalOpen, setIsPaymentMethodModalOpen] = useState(false)
  const [isSplitMode, setIsSplitMode] = useState(false)
  const [selectedElectronic, setSelectedElectronic] = useState<string | null>(null)
  const [installments, setInstallments] = useState(1)
  const [customInstallments, setCustomInstallments] = useState('')
  const [deferredDate, setDeferredDate] = useState('')
  const [checkNumber, setCheckNumber] = useState('')
  const [checkBank, setCheckBank] = useState('')
  const [checkDate, setCheckDate] = useState('')

  const remaining = total - paid
  const isCreditSale = paymentMethod === 'credit'
  const isInstallmentSale = paymentMethod === 'installment'
  const projectedDebt = customerBalance + Math.max(0, remaining)
  const willExceedCreditLimit = isCreditSale && Boolean(selectedCustomer)
    && Number(customerCreditLimit) > 0
    && projectedDebt > Number(customerCreditLimit)
  const isCreditSaleWithoutCustomer = (isCreditSale || isInstallmentSale) && !selectedCustomer
  const isCreditAdvanceMissing = isCreditSale && paidAmount.trim() === ''
  const selectedPaymentLabel = PAYMENT_METHODS.find((method) => method.id === paymentMethod)?.label ?? 'اختر طريقة الدفع'
  const normalizedAgentPhone = (installmentWhatsAppNumber || '').replace(/[^0-9]/g, '')
  const canSendInstallmentWhatsApp = isInstallmentSale && Boolean(selectedCustomer && normalizedAgentPhone)

  const sendInstallmentWhatsApp = () => {
    if (!canSendInstallmentWhatsApp) return
    const defaultMessage = `*طلب تقسيط جديد - {store_name}*\n\nالسلام عليكم،\nنرجو متابعة طلب التقسيط التالي:\n\n*اسم العميل:* {customer_name}\n*إجمالي الفاتورة:* ₪{total}\n*مدة التقسيط:* {months} أشهر\n*قيمة القسط التقريبية:* ₪{installment}\n\nيرجى تأكيد تسجيل الطلب ومتابعته.\n\nمع التحية،\n{store_name}`
    const messageTemplate = (installmentWhatsAppMessage?.trim() || defaultMessage)
      .replace(/\\+n/g, '\n')
      .replace(/[\u{1F000}-\u{1FAFF}\u{2300}-\u{27BF}]/gu, '')
      .replace(/\uFE0F/gu, '')
      .replace(/�/gu, '')
    const message = messageTemplate
      .replaceAll('{store_name}', storeName)
      .replaceAll('{customer_name}', selectedCustomerName || 'غير محدد')
      .replaceAll('{total}', total.toFixed(2))
      .replaceAll('{months}', String(installmentMonths))
      .replaceAll('{installment}', (total / installmentMonths).toFixed(2))
    const whatsappUrl = new URL(`https://wa.me/${normalizedAgentPhone}`)
    whatsappUrl.searchParams.set('text', message)
    window.open(whatsappUrl.toString(), '_blank', 'noopener,noreferrer')
  }

  useEffect(() => {
    setPaid(parseFloat(paidAmount) || 0)
  }, [paidAmount])

  useEffect(() => {
    const allocations: PaymentAllocation[] = isSplitMode
      ? splitPayments.map((payment) => ({
          amount: payment.amount,
          method: payment.method,
          ...(payment.method === 'checks' ? { check_number: payment.checkNumber, bank_name: payment.checkBank, check_date: payment.checkDate } : {}),
        }))
      : [{
          amount: paid,
          method: paymentMethod,
          ...(paymentMethod === 'checks' ? { check_number: checkNumber, bank_name: checkBank, check_date: checkDate } : {}),
        }]
    onPaymentAllocationsChange?.(allocations.filter((allocation) => allocation.amount > 0))
  }, [checkBank, checkDate, checkNumber, isSplitMode, paid, paymentMethod, splitPayments, onPaymentAllocationsChange])

  useEffect(() => {
    if (paymentMethod !== 'cash') return

    if (total > 0) {
      setPaid(total)
      setPaidAmount(total.toFixed(2))
    } else {
      setPaid(0)
      setPaidAmount('')
    }
  }, [paymentMethod, setPaidAmount, total])

  const addSplitPayment = () => {
    setSplitPayments((current) => {
      const remainingForSplit = total - current.reduce((sum, payment) => sum + payment.amount, 0)
      if (remainingForSplit <= 0) return current

      return [...current, {
        id: createSplitPaymentId(),
        method: 'cash',
        amount: remainingForSplit,
      }]
    })
  }

  const removeSplitPayment = (id: string) => {
    setSplitPayments(splitPayments.filter(p => p.id !== id))
  }

  const updateSplitPayment = (
    id: string,
    field: 'method' | 'amount' | 'checkNumber' | 'checkBank' | 'checkDate',
    value: string | number,
  ) => {
    setSplitPayments(splitPayments.map(p => {
      if (p.id !== id) return p
      if (field === 'method') return { ...p, method: value as PaymentMethod }
      if (field === 'amount') return { ...p, amount: Number(value) }
      return { ...p, [field]: String(value) }
    }))
  }

  const hasIncompleteCheck = (paymentMethod === 'checks' && !isSplitMode && !checkNumber.trim()) ||
    (isSplitMode && splitPayments.some((payment) => payment.method === 'checks' && !payment.checkNumber?.trim()))
  const amountInCents = Math.round(total * 100)
  const paidInCents = Math.round(paid * 100)
  const splitPaidInCents = Math.round(splitPayments.reduce((sum, payment) => sum + payment.amount, 0) * 100)
  const checkoutDisabledReason = checkoutBlockedReason || (
    total <= 0 ? 'أضف منتجًا إلى السلة قبل إتمام البيع.' :
    isCreditSaleWithoutCustomer ? 'اختر عميلًا لإتمام البيع بالدين أو بالتقسيط.' :
    isCreditAdvanceMissing ? 'أدخل مبلغ الدفعة المقدمة.' :
    hasIncompleteCheck ? 'أدخل رقم الشيك قبل إتمام البيع.' :
    isSplitMode && splitPaidInCents < amountInCents ? 'أكمل توزيع مبلغ الفاتورة على طرق الدفع.' :
    !isSplitMode && ['cash', 'card', 'checks'].includes(paymentMethod) && paidInCents < amountInCents
      ? 'المبلغ المدفوع أقل من إجمالي الفاتورة.'
      : undefined
  )
  const isCheckoutDisabled = isProcessing || Boolean(checkoutDisabledReason)

  return (
    <div className="pos-advanced-payment-panel">
      {/* Payment Method Selection */}
      <div className="payment-section">
        <div className="payment-method-summary">
          <div>
            <label className="section-label">طريقة الدفع</label>
            <strong>{selectedPaymentLabel}</strong>
          </div>
          <button className="payment-method-open-btn" onClick={() => setIsPaymentMethodModalOpen(true)}>
            تغيير الطريقة
          </button>
        </div>
      </div>

      <Modal
        isOpen={isPaymentMethodModalOpen}
        onClose={() => setIsPaymentMethodModalOpen(false)}
        title="اختيار طريقة الدفع"
        size="md"
        variant="modern"
      >
        <div className="payment-method-modal-body">
          <p className="payment-method-modal-hint">اختر الطريقة المناسبة، وستظهر تفاصيلها مباشرة بعد الاختيار.</p>
          <div className="payment-methods-grid">
            {PAYMENT_METHODS.map(({ id, label, icon: Icon, color }) => (
              <button
                key={id}
                className={cn('payment-method-btn', paymentMethod === id && 'active')}
                onClick={() => {
                  setPaymentMethod(id)
                  setIsPaymentMethodModalOpen(false)
                  if (id !== 'credit' && id !== 'installment') {
                    setPaidAmount(total.toFixed(2))
                    setPaid(total)
                  } else {
                    setPaidAmount('')
                    setPaid(0)
                  }
                }}
                style={{ '--method-color': color } as React.CSSProperties}
              >
                <Icon className="w-4 h-4 method-icon" />
                <span>{label}</span>
              </button>
            ))}
          </div>
        </div>
      </Modal>

      {/* Electronic Payments */}
      {(paymentMethod === 'card' && electronicPaymentsEnabled && enabledElectronicMethods.length > 0) && (
        <div className="payment-section">
          <label className="section-label">الدفع الإلكتروني</label>
          <div className="electronic-methods">
            {ELECTRONIC_METHODS.filter(({ id }) => enabledElectronicMethods.includes(id)).map(({ id, label, icon: Icon }) => (
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
                {payment.method === 'checks' && (
                  <div className="split-check-details">
                    <input
                      type="text"
                      placeholder="رقم الشيك"
                      value={payment.checkNumber || ''}
                      onChange={(e) => updateSplitPayment(payment.id, 'checkNumber', e.target.value)}
                      className="check-input"
                    />
                    <input
                      type="text"
                      placeholder="اسم البنك"
                      value={payment.checkBank || ''}
                      onChange={(e) => updateSplitPayment(payment.id, 'checkBank', e.target.value)}
                      className="check-input"
                    />
                    <input
                      type="date"
                      value={payment.checkDate || ''}
                      onChange={(e) => updateSplitPayment(payment.id, 'checkDate', e.target.value)}
                      className="check-input"
                    />
                  </div>
                )}
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
      {(paymentMethod === 'checks' || (isSplitMode && splitPayments.some((payment) => payment.method === 'checks'))) && (
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

      {/* Installment payment */}
      {isInstallmentSale && (
        <div className="payment-section">
          <label className="section-label">مدة التقسيط</label>
          <div className="installments-selector">
            {[3, 6, 12].map((num) => (
              <button
                key={num}
                className={cn('installment-btn', installments === num && 'active')}
                onClick={() => {
                  setInstallments(num)
                  setCustomInstallments('')
                  setInstallmentMonths(num)
                }}
              >
                {num} أشهر
              </button>
            ))}
            <button
              className={cn('installment-btn', customInstallments !== '' && 'active')}
              onClick={() => {
                setInstallments(0)
                setCustomInstallments(String(installmentMonths))
              }}
            >
              مدة مخصصة
            </button>
          </div>
          {customInstallments !== '' || installments === 0 ? (
            <input
              type="number"
              min={1}
              max={120}
              value={customInstallments}
              placeholder="عدد الأشهر"
              onChange={(event) => {
                const value = event.target.value
                setCustomInstallments(value)
                const months = Number(value)
                if (months >= 1 && months <= 120) setInstallmentMonths(months)
              }}
              className="deferred-date-input"
            />
          ) : null}
          {installmentMonths > 0 && (
            <div className="installment-info">
              <Calendar className="w-4 h-4" />
              <span>{installmentMonths} أشهر - كل قسط تقريبًا: ₪{(total / installmentMonths).toFixed(2)}</span>
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
          <span>{isInstallmentSale ? 'التقسيط يتطلب تحديد العميل' : 'البيع بالدين يتطلب تحديد العميل'}</span>
        </div>
      )}

      {willExceedCreditLimit && (
        <div className="payment-warning warning" role="alert">
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
                دفع المبلغ كاملًا
              </button>
            </div>
          )}

          {isCreditAdvanceMissing && (
            <p className="payment-error-text">يرجى إدخال رقم للدفعة المقدمة</p>
          )}
          {remaining > 0 && paymentMethod === 'credit' && (
            <p className="payment-remaining">المتبقي كدين: ₪{remaining.toLocaleString()}</p>
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
        type="button"
        className={cn('checkout-btn', isCheckoutDisabled && 'disabled')}
        onClick={onCheckout}
        disabled={isCheckoutDisabled}
        aria-describedby={checkoutDisabledReason ? 'checkout-disabled-reason' : undefined}
        title={checkoutDisabledReason}
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
      {checkoutDisabledReason && !isProcessing && (
        <p id="checkout-disabled-reason" className="payment-checkout-hint" role="status">
          <AlertTriangle className="h-4 w-4 shrink-0" aria-hidden="true" />
          <span>{checkoutDisabledReason}</span>
        </p>
      )}
      {isInstallmentSale && (
        <button
          type="button"
          className="installment-whatsapp-btn"
          onClick={sendInstallmentWhatsApp}
          disabled={!canSendInstallmentWhatsApp}
          title={!selectedCustomer ? 'اختر العميل أولًا' : !normalizedAgentPhone ? 'اضبط رقم واتساب وكيل التقسيط من الإعدادات' : 'إرسال تفاصيل التقسيط إلى الوكيل عبر واتساب'}
        >
          <MessageCircle className="w-5 h-5" />
          <span>إرسال تفاصيل التقسيط عبر واتساب</span>
        </button>
      )}
    </div>
  )
}
