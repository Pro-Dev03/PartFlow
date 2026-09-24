import { useEffect, useRef, useState } from 'react'
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
  amount: string
  checkNumber?: string
  checkBank?: string
  checkDate?: string
}

type PaymentAmountTarget = { type: 'main' } | { type: 'split'; id: string }

const MAX_PAYMENT_AMOUNT_INTEGER_DIGITS = 9

function sanitizePaymentAmount(value: string): string {
  const westernDigits = value
    .replace(/[٠-٩]/g, (digit) => String(digit.charCodeAt(0) - '٠'.charCodeAt(0)))
    .replace(/[۰-۹]/g, (digit) => String(digit.charCodeAt(0) - '۰'.charCodeAt(0)))
  const normalized = westernDigits.replace(/,/g, '.').replace(/[^\d.]/g, '')
  const decimalIndex = normalized.indexOf('.')
  const hasDecimal = decimalIndex >= 0
  let integer = (hasDecimal ? normalized.slice(0, decimalIndex) : normalized)
    .replace(/^0+(?=\d)/, '')
    .slice(0, MAX_PAYMENT_AMOUNT_INTEGER_DIGITS)
  const fraction = hasDecimal
    ? normalized.slice(decimalIndex + 1).replace(/\./g, '').slice(0, 2)
    : ''

  if (hasDecimal && integer === '') integer = '0'
  if (!integer && !hasDecimal) return ''
  return hasDecimal ? integer + '.' + fraction : integer
}

function parsePaymentAmount(value: string): number {
  const parsed = Number.parseFloat(value)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0
}

function appendPaymentAmount(value: string, key: string): string {
  const current = sanitizePaymentAmount(value)
  if (key === '.') {
    if (current.includes('.')) return current
    return (current || '0') + '.'
  }
  const next = (current || '') + key
  return sanitizePaymentAmount(next)
}

function backspacePaymentAmount(value: string): string {
  return sanitizePaymentAmount(value.slice(0, -1))
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
  onPaymentAllocationsChange?: (allocations: PaymentAllocation[], isSplitMode: boolean) => void
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
  const [paid, setPaid] = useState(parsePaymentAmount(paidAmount))
  const [splitPayments, setSplitPayments] = useState<SplitPayment[]>([])
  const [splitPaymentDraft, setSplitPaymentDraft] = useState<SplitPayment[]>([])
  const [isPaymentMethodModalOpen, setIsPaymentMethodModalOpen] = useState(false)
  const [isSplitPaymentModalOpen, setIsSplitPaymentModalOpen] = useState(false)
  const [isCreditOptionsModalOpen, setIsCreditOptionsModalOpen] = useState(false)
  const [isSplitMode, setIsSplitMode] = useState(false)
  const [selectedElectronic, setSelectedElectronic] = useState<string | null>(null)
  const [installments, setInstallments] = useState(1)
  const [customInstallments, setCustomInstallments] = useState('')
  const [deferredDate, setDeferredDate] = useState('')
  const [checkNumber, setCheckNumber] = useState('')
  const [checkBank, setCheckBank] = useState('')
  const [checkDate, setCheckDate] = useState('')
  const [activeAmountTarget, setActiveAmountTarget] = useState<PaymentAmountTarget>({ type: 'main' })
  const mainAmountEntryStarted = useRef(false)
  const splitAmountEntriesStarted = useRef(new Set<string>())

  const remaining = total - paid
  const isCreditSale = paymentMethod === 'credit'
  const isInstallmentSale = paymentMethod === 'installment'
  const projectedDebt = customerBalance + Math.max(0, remaining)
  const willExceedCreditLimit = !isSplitMode && isCreditSale && Boolean(selectedCustomer)
    && Number(customerCreditLimit) > 0
    && projectedDebt > Number(customerCreditLimit)
  const isCreditSaleWithoutCustomer = !isSplitMode && (isCreditSale || isInstallmentSale) && !selectedCustomer
  const isCreditAdvanceMissing = !isSplitMode && isCreditSale && paidAmount.trim() === ''
  const selectedPaymentLabel = PAYMENT_METHODS.find((method) => method.id === paymentMethod)?.label ?? 'اختر طريقة الدفع'
  const normalizedAgentPhone = (installmentWhatsAppNumber || '').replace(/[^0-9]/g, '')
  const canSendInstallmentWhatsApp = !isSplitMode && isInstallmentSale && Boolean(selectedCustomer && normalizedAgentPhone)

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
    setPaid(parsePaymentAmount(paidAmount))
  }, [paidAmount])

  useEffect(() => {
    const allocations: PaymentAllocation[] = isSplitMode
      ? splitPayments.map((payment) => ({
          amount: parsePaymentAmount(payment.amount),
          method: payment.method,
          ...(payment.method === 'checks' ? { check_number: payment.checkNumber, bank_name: payment.checkBank, check_date: payment.checkDate } : {}),
        }))
      : [{
          amount: paid,
          method: paymentMethod,
          ...(paymentMethod === 'checks' ? { check_number: checkNumber, bank_name: checkBank, check_date: checkDate } : {}),
        }]
    onPaymentAllocationsChange?.(allocations.filter((allocation) => allocation.amount > 0), isSplitMode)
  }, [checkBank, checkDate, checkNumber, isSplitMode, paid, paymentMethod, splitPayments, onPaymentAllocationsChange])

  useEffect(() => {
    if (paymentMethod !== 'cash') return

    mainAmountEntryStarted.current = false
    if (total > 0) {
      setPaid(total)
      setPaidAmount(total.toFixed(2))
    } else {
      setPaid(0)
      setPaidAmount('')
    }
  }, [paymentMethod, setPaidAmount, total])

  const addSplitPayment = () => {
    const remainingForSplit = Math.max(0, total - splitPaymentDraft.reduce((sum, payment) => sum + parsePaymentAmount(payment.amount), 0))
    const payment = {
      id: createSplitPaymentId(),
      method: 'cash' as const,
      amount: remainingForSplit.toFixed(2),
    }
    setSplitPaymentDraft((current) => [...current, payment])
    setActiveAmountTarget({ type: 'split', id: payment.id })
  }

  const removeSplitPayment = (id: string) => {
    const next = splitPaymentDraft.filter((payment) => payment.id !== id)
    setSplitPaymentDraft(next)
    splitAmountEntriesStarted.current.delete(id)
    if (activeAmountTarget.type === 'split' && activeAmountTarget.id === id) {
      setActiveAmountTarget(next[0] ? { type: 'split', id: next[0].id } : { type: 'main' })
    }
  }

  const updateSplitPayment = (
    id: string,
    field: 'method' | 'amount' | 'checkNumber' | 'checkBank' | 'checkDate',
    value: string | number,
  ) => {
    setSplitPaymentDraft((current) => current.map(p => {
      if (p.id !== id) return p
      if (field === 'method') return { ...p, method: value as PaymentMethod }
      if (field === 'amount') return { ...p, amount: sanitizePaymentAmount(String(value)) }
      return { ...p, [field]: String(value) }
    }))
  }

  const openSplitPaymentModal = () => {
    const currentPayments = isSplitMode ? splitPayments.map((payment) => ({ ...payment })) : []
    const initialPayments = currentPayments.length > 0
      ? currentPayments
      : [{
          id: createSplitPaymentId(),
          method: 'cash' as const,
          amount: total > 0 ? total.toFixed(2) : '',
        }]
    setSplitPaymentDraft(initialPayments)
    splitAmountEntriesStarted.current.clear()
    setActiveAmountTarget(initialPayments[0]
      ? { type: 'split', id: initialPayments[0].id }
      : { type: 'main' })
    setIsSplitPaymentModalOpen(true)
  }

  const closeSplitPaymentModal = () => {
    setIsSplitPaymentModalOpen(false)
    setActiveAmountTarget({ type: 'main' })
  }

  const confirmSplitPaymentModal = () => {
    if (splitPaymentDraft.length === 0) return
    setSplitPayments(splitPaymentDraft)
    setIsSplitMode(true)
    closeSplitPaymentModal()
  }

  const disableSplitMode = () => {
    setIsSplitMode(false)
    setSplitPayments([])
    setSplitPaymentDraft([])
    splitAmountEntriesStarted.current.clear()
    setActiveAmountTarget({ type: 'main' })
    setIsSplitPaymentModalOpen(false)
  }

  const updateMainAmount = (value: string, markAsEntered = true) => {
    const safeValue = sanitizePaymentAmount(value)
    if (markAsEntered) mainAmountEntryStarted.current = true
    setPaidAmount(safeValue)
    setPaid(parsePaymentAmount(safeValue))
  }

  const updateActiveAmount = (key: string | 'backspace' | 'clear') => {
    if (activeAmountTarget.type === 'main') {
      const base = mainAmountEntryStarted.current || key === 'backspace' ? paidAmount : ''
      const next = key === 'clear'
        ? ''
        : key === 'backspace'
          ? backspacePaymentAmount(base)
          : appendPaymentAmount(base, key)
      updateMainAmount(next)
      return
    }

    const target = splitPaymentDraft.find((payment) => payment.id === activeAmountTarget.id)
    if (!target) {
      setActiveAmountTarget({ type: 'main' })
      return
    }
    const base = splitAmountEntriesStarted.current.has(target.id) || key === 'backspace' ? target.amount : ''
    const next = key === 'clear'
      ? ''
      : key === 'backspace'
        ? backspacePaymentAmount(base)
        : appendPaymentAmount(base, key)
    splitAmountEntriesStarted.current.add(target.id)
    updateSplitPayment(target.id, 'amount', next)
  }

  const handlePhysicalAmountKey = (event: React.KeyboardEvent<HTMLDivElement>) => {
    const target = event.target as HTMLElement | null
    if (target?.closest('input, textarea, select, [contenteditable="true"]')) return
    if (!event.code.startsWith('Numpad')) return

    if (/^Numpad\d$/.test(event.code)) {
      event.preventDefault()
      event.stopPropagation()
      updateActiveAmount(event.code.slice(-1))
    } else if (event.code === 'NumpadDecimal') {
      event.preventDefault()
      event.stopPropagation()
      updateActiveAmount('.')
    }
  }

  const hasIncompleteCheck = (paymentMethod === 'checks' && !isSplitMode && !checkNumber.trim()) ||
    (isSplitMode && splitPayments.some((payment) => payment.method === 'checks' && !payment.checkNumber?.trim()))
  const amountInCents = Math.round(total * 100)
  const paidInCents = Math.round(paid * 100)
  const splitPaid = splitPayments.reduce((sum, payment) => sum + parsePaymentAmount(payment.amount), 0)
  const splitPaidInCents = Math.round(splitPaid * 100)
  const splitDraftPaid = splitPaymentDraft.reduce((sum, payment) => sum + parsePaymentAmount(payment.amount), 0)
  const splitDraftRemaining = total - splitDraftPaid
  const splitDraftIsBalanced = Math.round(splitDraftPaid * 100) === amountInCents
  const checkoutDisabledReason = checkoutBlockedReason || (
    total <= 0 ? 'أضف منتجًا إلى السلة قبل إتمام البيع.' :
    isCreditSaleWithoutCustomer ? 'اختر عميلًا لإتمام البيع بالدين أو بالتقسيط.' :
    isCreditAdvanceMissing ? 'أدخل مبلغ الدفعة المقدمة.' :
    hasIncompleteCheck ? 'أدخل رقم الشيك قبل إتمام البيع.' :
    isSplitMode && splitPaidInCents !== amountInCents
      ? splitPaidInCents > amountInCents ? 'مجموع الدفعات يتجاوز إجمالي الفاتورة.' : 'أكمل توزيع مبلغ الفاتورة على طرق الدفع.' :
    !isSplitMode && ['cash', 'card', 'checks'].includes(paymentMethod) && paidInCents < amountInCents
      ? 'المبلغ المدفوع أقل من إجمالي الفاتورة.'
      : undefined
  )
  const isCheckoutDisabled = isProcessing || Boolean(checkoutDisabledReason)

  return (
    <div
      className={cn('pos-advanced-payment-panel', isSplitMode && 'split-mode')}
      onKeyDownCapture={handlePhysicalAmountKey}
      aria-label="الدفع وإتمام البيع"
    >
      {/* Payment Method Selection */}
      <div className="payment-section">
        <div className="payment-method-summary">
          <div>
            <label className="section-label">طريقة الدفع</label>
            <strong>{isSplitMode ? 'تقسيم الدفع' : selectedPaymentLabel}</strong>
          </div>
          <div className="pos-payment-method-actions">
            {!isSplitMode && (
              <button type="button" className="payment-method-open-btn" onClick={() => setIsPaymentMethodModalOpen(true)}>
                تغيير الطريقة
              </button>
            )}
            {isCreditSale && !isSplitMode && (
              <button type="button" className="pos-credit-options-open-btn" onClick={() => setIsCreditOptionsModalOpen(true)}>
                خيارات الدين
              </button>
            )}
            <button type="button" className="pos-split-payment-open-btn" onClick={openSplitPaymentModal}>
              {isSplitMode ? 'تعديل التقسيم' : 'تقسيم الدفع'}
            </button>
          </div>
        </div>

        {isSplitMode && (
          <div className="pos-split-payment-summary" aria-label="ملخص تقسيم الدفع">
            <div className="pos-split-payment-summary-metrics">
                <div>
                  <span>طرق الدفع</span>
                  <strong>{splitPayments.length}</strong>
                </div>
                <div>
                  <span>الموزع</span>
                  <strong>₪{splitPaid.toLocaleString()}</strong>
                </div>
                <div>
                  <span>المتبقي</span>
                  <strong className={total - splitPaid < 0 ? 'is-overpaid' : splitPaidInCents === amountInCents ? 'is-balanced' : 'is-unbalanced'}>
                    ₪{(total - splitPaid).toLocaleString()}
                  </strong>
                </div>
            </div>
            <button type="button" className="pos-split-payment-disable-btn" onClick={disableSplitMode}>
              إلغاء التقسيم
            </button>
          </div>
        )}
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
                  mainAmountEntryStarted.current = false
                  setPaymentMethod(id)
                  setIsPaymentMethodModalOpen(false)
                  if (id !== 'credit' && id !== 'installment') {
                    setPaidAmount(total.toFixed(2))
                    setPaid(total)
                  } else {
                    setPaidAmount('')
                    setPaid(0)
                  }
                  if (id === 'credit') setIsCreditOptionsModalOpen(true)
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

      <Modal
        isOpen={isSplitPaymentModalOpen}
        onClose={closeSplitPaymentModal}
        title="تقسيم الدفع"
        size="xl"
        variant="modern"
        className="pos-split-payment-modal"
        aria-describedby="pos-split-payment-help"
      >
        <div className="pos-split-payment-dialog" dir="rtl">
          <p id="pos-split-payment-help" className="pos-split-payment-help">
            وزّع إجمالي الفاتورة على طرق الدفع، ثم أكّد التقسيم لحفظه.
          </p>

          <div className="pos-split-payment-overview" aria-live="polite">
            <div className="pos-split-payment-stat">
              <span>إجمالي الفاتورة</span>
              <strong>₪{total.toLocaleString()}</strong>
            </div>
            <div className="pos-split-payment-stat">
              <span>المبلغ الموزع</span>
              <strong>₪{splitDraftPaid.toLocaleString()}</strong>
            </div>
            <div className="pos-split-payment-stat">
              <span>المبلغ المتبقي</span>
              <strong className={splitDraftRemaining < 0 ? 'is-overpaid' : splitDraftIsBalanced ? 'is-balanced' : 'is-unbalanced'}>
                ₪{splitDraftRemaining.toLocaleString()}
              </strong>
            </div>
          </div>

          <div className="pos-split-payment-list" aria-label="طرق الدفع المقسمة">
            {splitPaymentDraft.map((payment, index) => (
              <section key={payment.id} className="pos-split-modal-row">
                <div className="pos-split-modal-row-header">
                  <strong>طريقة الدفع {index + 1}</strong>
                  <button
                    type="button"
                    className="remove-split-btn"
                    onClick={() => removeSplitPayment(payment.id)}
                    aria-label={`حذف طريقة الدفع ${index + 1}`}
                    title="حذف طريقة الدفع"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
                <div className="pos-split-modal-fields">
                  <label className="pos-split-modal-field">
                    <span>طريقة الدفع</span>
                    <select
                      value={payment.method}
                      onChange={(event) => updateSplitPayment(payment.id, 'method', event.target.value)}
                      className="pos-split-modal-input"
                      aria-label={`طريقة الدفع ${index + 1}`}
                    >
                      <option value="cash">نقدًا</option>
                      <option value="card">بطاقة</option>
                      <option value="checks">شيكات</option>
                    </select>
                  </label>
                  <label className="pos-split-modal-field">
                    <span>المبلغ</span>
                    <input
                      type="text"
                      inputMode="decimal"
                      value={payment.amount}
                      onFocus={(event) => {
                        setActiveAmountTarget({ type: 'split', id: payment.id })
                        if (!splitAmountEntriesStarted.current.has(payment.id)) event.currentTarget.select()
                      }}
                      onChange={(event) => {
                        splitAmountEntriesStarted.current.add(payment.id)
                        updateSplitPayment(payment.id, 'amount', event.target.value)
                      }}
                      className="pos-split-modal-input"
                      aria-label={`مبلغ طريقة الدفع ${index + 1}`}
                      maxLength={MAX_PAYMENT_AMOUNT_INTEGER_DIGITS + 3}
                    />
                  </label>
                </div>
                {payment.method === 'checks' && (
                  <div className="pos-split-modal-check-details">
                    <label className="pos-split-modal-field">
                      <span>رقم الشيك</span>
                      <input
                        type="text"
                        value={payment.checkNumber || ''}
                        onChange={(event) => updateSplitPayment(payment.id, 'checkNumber', event.target.value)}
                        className="pos-split-modal-input"
                      />
                    </label>
                    <label className="pos-split-modal-field">
                      <span>اسم البنك</span>
                      <input
                        type="text"
                        value={payment.checkBank || ''}
                        onChange={(event) => updateSplitPayment(payment.id, 'checkBank', event.target.value)}
                        className="pos-split-modal-input"
                      />
                    </label>
                    <label className="pos-split-modal-field">
                      <span>تاريخ الشيك</span>
                      <input
                        type="date"
                        value={payment.checkDate || ''}
                        onChange={(event) => updateSplitPayment(payment.id, 'checkDate', event.target.value)}
                        className="pos-split-modal-input"
                      />
                    </label>
                  </div>
                )}
              </section>
            ))}
          </div>

          <button type="button" className="add-split-btn pos-split-modal-add-btn" onClick={addSplitPayment}>
            <Plus className="w-4 h-4" />
            <span>إضافة طريقة دفع</span>
          </button>

          {splitPaymentDraft.length > 0 && (
            <NumericKeypad
              onInput={updateActiveAmount}
              disabled={isProcessing}
              targetLabel={activeAmountTarget.type === 'split' ? 'مبلغ الدفعة المحددة' : 'مبلغ الدفعة'}
            />
          )}

          <div className="pos-split-modal-actions">
            <button type="button" className="pos-split-modal-cancel-btn" onClick={closeSplitPaymentModal}>
              إلغاء
            </button>
            <button
              type="button"
              className="pos-split-modal-confirm-btn"
              onClick={confirmSplitPaymentModal}
              disabled={isProcessing || splitPaymentDraft.length === 0}
            >
              تأكيد التقسيم
            </button>
          </div>
        </div>
      </Modal>

      <Modal
        isOpen={isCreditOptionsModalOpen}
        onClose={() => setIsCreditOptionsModalOpen(false)}
        title="خيارات البيع بالدين"
        size="md"
        variant="modern"
        className="pos-credit-options-modal"
      >
        <div className="pos-credit-options-dialog" dir="rtl">
          <p className="pos-credit-options-help">حدّد الدفعة المقدمة وتاريخ الاستحقاق قبل إتمام البيع بالدين.</p>

          <div className="pos-credit-options-overview">
            <div>
              <span>إجمالي الفاتورة</span>
              <strong>₪{total.toLocaleString()}</strong>
            </div>
            <div>
              <span>الدفعة المقدمة</span>
              <strong>₪{paid.toLocaleString()}</strong>
            </div>
            <div>
              <span>{remaining < 0 ? 'المردود' : 'المتبقي كدين'}</span>
              <strong>₪{Math.abs(remaining).toLocaleString()}</strong>
            </div>
          </div>

          <label className="pos-credit-options-field">
            <span>تاريخ الاستحقاق</span>
            <input
              type="date"
              value={deferredDate}
              onChange={(event) => setDeferredDate(event.target.value)}
              className="pos-credit-options-input"
            />
          </label>

          <div className="payment-amount-section pos-credit-advance-section">
            <label className="payment-amount-label" htmlFor="pos-paid-amount">
              الدفعة المقدمة
            </label>
            <div className="payment-amount-input-wrapper">
              <input
                id="pos-paid-amount"
                type="text"
                inputMode="decimal"
                className="payment-amount-input"
                value={paidAmount}
                onFocus={(event) => {
                  setActiveAmountTarget({ type: 'main' })
                  if (!mainAmountEntryStarted.current) event.currentTarget.select()
                }}
                onChange={(event) => updateMainAmount(event.target.value)}
                placeholder="0"
                aria-label="الدفعة المقدمة"
                maxLength={MAX_PAYMENT_AMOUNT_INTEGER_DIGITS + 3}
              />
              <span className="payment-currency">₪</span>
            </div>
            <NumericKeypad onInput={updateActiveAmount} disabled={isProcessing} targetLabel="الدفعة المقدمة" />
            {isCreditAdvanceMissing && (
              <p className="payment-error-text" role="alert">أدخل مبلغ الدفعة المقدمة.</p>
            )}
            {remaining < 0 && (
              <p className="payment-change">المردود: ₪{Math.abs(remaining).toLocaleString()}</p>
            )}
          </div>

          {selectedCustomer && customerBalance > 0 && (
            <div className="customer-balance">
              <User className="w-4 h-4" />
              <span>رصيد العميل: ₪{customerBalance.toLocaleString()}</span>
            </div>
          )}

          {isCreditSaleWithoutCustomer && (
            <div className="payment-warning error" role="alert">
              <AlertTriangle className="w-4 h-4" />
              <span>البيع بالدين يتطلب تحديد العميل.</span>
            </div>
          )}

          {willExceedCreditLimit && (
            <div className="payment-warning warning" role="alert">
              <AlertTriangle className="w-4 h-4" />
              <span>تنبيه: هذا البيع سيتجاوز حد الدين المحدد.</span>
            </div>
          )}

          {isCreditSaleWithoutCustomer && onQuickCustomerCreate && (
            <button
              type="button"
              className="payment-quick-customer"
              onClick={() => {
                setIsCreditOptionsModalOpen(false)
                onQuickCustomerCreate?.()
              }}
            >
              <User className="w-4 h-4" />
              <span>إنشاء عميل سريع</span>
            </button>
          )}

          <div className="pos-credit-options-actions">
            <button type="button" className="pos-credit-options-close-btn" onClick={() => setIsCreditOptionsModalOpen(false)}>
              تم
            </button>
          </div>
        </div>
      </Modal>

      <div className="pos-payment-options-scroll">
      {/* Electronic Payments */}
      {(paymentMethod === 'card' && !isSplitMode && electronicPaymentsEnabled && enabledElectronicMethods.length > 0) && (
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

      {/* Check Details */}
      {paymentMethod === 'checks' && !isSplitMode && (
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
      {isInstallmentSale && !isSplitMode && (
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

      {/* Installment customer warning */}
      {isInstallmentSale && isCreditSaleWithoutCustomer && (
        <div className="payment-warning error">
          <AlertTriangle className="w-4 h-4" />
          <span>التقسيط يتطلب تحديد العميل</span>
        </div>
      )}

      {isInstallmentSale && !isSplitMode && (
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

      {/* Amount Input for Cash */}
      {paymentMethod === 'cash' && !isSplitMode && (
        <div className="payment-amount-section">
          <label className="payment-amount-label" htmlFor="pos-paid-amount">
            المبلغ المدفوع
          </label>
          <div className="payment-amount-input-wrapper">
            <input
              id="pos-paid-amount"
              type="text"
              inputMode="decimal"
              className="payment-amount-input"
              value={paidAmount}
              onFocus={(event) => {
                setActiveAmountTarget({ type: 'main' })
                if (!mainAmountEntryStarted.current) event.currentTarget.select()
              }}
              onChange={(e) => {
                updateMainAmount(e.target.value)
              }}
              placeholder="0"
              aria-label="المبلغ المدفوع"
              maxLength={MAX_PAYMENT_AMOUNT_INTEGER_DIGITS + 3}
              min={0}
            />
             <span className="payment-currency">₪</span>
          </div>
          <NumericKeypad
            onInput={updateActiveAmount}
            disabled={isProcessing}
          />
          
          {/* Quick Amount Buttons */}
          {paymentMethod === 'cash' && (
            <div className="quick-amounts">
              {[50, 100, 200, 500].map((amount) => (
              <button
                key={amount}
                className="quick-amount-btn"
                onClick={() => {
                  mainAmountEntryStarted.current = true
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
                  mainAmountEntryStarted.current = true
                  setPaidAmount(total.toFixed(2))
                  setPaid(total)
                }}
              >
                دفع المبلغ كاملًا
              </button>
            </div>
          )}

          {remaining > 0 && (
            <p className="payment-remaining">المتبقي للدفع: ₪{remaining.toLocaleString()}</p>
          )}
          {remaining < 0 && (
            <p className="payment-change">الباقي للعميل (المردود): ₪{Math.abs(remaining).toLocaleString()}</p>
          )}
        </div>
      )}

      <div className="pos-payment-checkout-footer">
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
      </div>
    </div>
  )
}

interface NumericKeypadProps {
  onInput: (key: string | 'backspace' | 'clear') => void
  disabled: boolean
  targetLabel?: string
}

const NUMERIC_KEYPAD_KEYS = ['7', '8', '9', '4', '5', '6', '1', '2', '3', '00', '0']

function NumericKeypad({ onInput, disabled, targetLabel }: NumericKeypadProps) {
  return (
    <div className="pos-numeric-keypad" role="group" aria-label="لوحة الأرقام للمبلغ المدفوع">
      <div className="pos-numeric-keypad-header">
        {targetLabel && <span>{targetLabel}</span>}
        <div className="pos-numeric-keypad-actions">
          <button
            type="button"
            className="pos-numeric-keypad-action pos-numeric-keypad-decimal"
            onMouseDown={(event) => event.preventDefault()}
            onClick={() => onInput('.')}
            disabled={disabled}
            aria-label="نقطة عشرية"
            title="نقطة عشرية"
          >
            .
          </button>
          <button
            type="button"
            className="pos-numeric-keypad-clear"
            onMouseDown={(event) => event.preventDefault()}
            onClick={() => onInput('clear')}
            disabled={disabled}
          >
            مسح المبلغ
          </button>
        </div>
      </div>
      <div className="pos-numeric-keypad-grid">
        {NUMERIC_KEYPAD_KEYS.map((key) => (
          <button
            key={key}
            type="button"
            className={key === '00' ? 'pos-numeric-key wide-key' : 'pos-numeric-key'}
            onMouseDown={(event) => event.preventDefault()}
            onClick={() => onInput(key)}
            disabled={disabled}
            aria-label={key}
          >
            {key}
          </button>
        ))}
        <button
          type="button"
          className="pos-numeric-key pos-numeric-key-backspace"
          onMouseDown={(event) => event.preventDefault()}
          onClick={() => onInput('backspace')}
          disabled={disabled}
          aria-label="حذف آخر رقم"
          title="حذف آخر رقم"
        >
          ⌫
        </button>
      </div>
    </div>
  )
}
